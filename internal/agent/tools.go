package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"studyforge/internal/model"
	"studyforge/internal/rag"

	openai "github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

// ==================== Function Calling 工具定义 ====================

// ToolExecutor 工具执行器，持有所有必要依赖
type ToolExecutor struct {
	DB          *gorm.DB
	VectorStore *rag.VectorStore
	LLM         *LLMClient
	UserID      string // 当前用户 ID
}

// NewToolExecutor 创建工具执行器
func NewToolExecutor(db *gorm.DB, vs *rag.VectorStore, llm *LLMClient, userID string) *ToolExecutor {
	return &ToolExecutor{
		DB:          db,
		VectorStore: vs,
		LLM:         llm,
		UserID:      userID,
	}
}

// ToolDefinitions 返回 Agent 可用的工具函数定义
func ToolDefinitions() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "search_materials",
				Description: "在用户上传的学习材料中语义搜索相关内容，返回最相关的文本片段",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"query": {"type": "string", "description": "搜索查询内容"},
						"top_k": {"type": "integer", "description": "返回结果数量，默认 5"}
					},
					"required": ["query"]
				}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "create_quiz",
				Description: "根据指定知识点和难度生成练习题",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"topic": {"type": "string", "description": "知识点名称"},
						"difficulty": {"type": "string", "enum": ["easy", "medium", "hard"]},
						"count": {"type": "integer", "description": "题目数量", "default": 3}
					},
					"required": ["topic", "difficulty"]
				}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "generate_card",
				Description: "为指定知识点生成一张学习卡片",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"concept": {"type": "string", "description": "知识点/概念名称"}
					},
					"required": ["concept"]
				}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_user_stats",
				Description: "获取用户的学习统计信息，包括做题数量、正确率、已学习材料等",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {}
				}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "recommend_study_plan",
				Description: "根据用户的学习情况生成个性化复习计划",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"focus_areas": {
							"type": "array",
							"items": {"type": "string"},
							"description": "需要重点复习的知识领域"
						}
					}
				}`),
			},
		},
	}
}

// ExecuteTool 执行工具函数，根据工具名调用对应实现
func (te *ToolExecutor) ExecuteTool(ctx context.Context, toolName string, args json.RawMessage) (string, error) {
	switch toolName {
	case "search_materials":
		return te.executeSearchMaterials(ctx, args)
	case "create_quiz":
		return te.executeCreateQuiz(ctx, args)
	case "generate_card":
		return te.executeGenerateCard(ctx, args)
	case "get_user_stats":
		return te.executeGetUserStats(ctx, args)
	case "recommend_study_plan":
		return te.executeRecommendStudyPlan(ctx, args)
	default:
		return "", fmt.Errorf("未知工具: %s", toolName)
	}
}

// ===== 工具函数真实实现 =====

// executeSearchMaterials 使用 RAG 向量检索语义搜索学习材料
func (te *ToolExecutor) executeSearchMaterials(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Query string `json:"query"`
		TopK  int    `json:"top_k"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return `{"error": "参数解析失败"}`, nil
	}
	if params.TopK <= 0 {
		params.TopK = 5
	}

	// 调用 VectorStore 的一体化语义搜索
	if te.VectorStore == nil {
		return `{"results": [], "message": "向量检索服务未初始化"}`, nil
	}

	results, err := te.VectorStore.SemanticSearch(ctx, params.Query, te.UserID, params.TopK)
	if err != nil {
		log.Printf("RAG 搜索失败: %v", err)
		return fmt.Sprintf(`{"error": "搜索失败: %s"}`, err.Error()), nil
	}

	if len(results) == 0 {
		return `{"results": [], "message": "未找到相关内容，请先上传学习材料"}`, nil
	}

	// 构建结果 JSON
	type searchResult struct {
		Content    string  `json:"content"`
		Score      float32 `json:"score"`
		MaterialID string  `json:"material_id"`
	}
	output := struct {
		Results []searchResult `json:"results"`
		Query   string         `json:"query"`
		Count   int            `json:"count"`
	}{
		Query: params.Query,
		Count: len(results),
	}
	for _, r := range results {
		// 截断过长的内容（rune 感知，不会切断多字节 UTF-8 字符）
		content := r.Content
		runes := []rune(content)
		if len(runes) > 500 {
			content = string(runes[:500]) + "..."
		}
		output.Results = append(output.Results, searchResult{
			Content:    content,
			Score:      r.Score,
			MaterialID: r.MaterialID,
		})
	}

	data, _ := json.Marshal(output)
	return string(data), nil
}

// executeCreateQuiz 调用 QuizMaster Agent 生成练习题并保存到数据库
func (te *ToolExecutor) executeCreateQuiz(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Topic      string `json:"topic"`
		Difficulty string `json:"difficulty"`
		Count      int    `json:"count"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return `{"error": "参数解析失败"}`, nil
	}
	if params.Count <= 0 {
		params.Count = 3
	}

	// 构建精简 prompt（系统提示 quizMasterPrompt 已包含完整出题规则，此处只传具体请求参数）
	quizPrompt := fmt.Sprintf(`请为以下知识点出题：

知识点: %s
指定难度偏好: %s
题目数量: %d 道

要求：
- 题目数量精确为 %d 道
- 指定难度为 %s 时，该难度的题目占比至少 60%%，其余为相邻难度
- 4 种题型混合（choice/fill/judge/short_answer 各至少 1 道）
- 解答题必须为 medium 或 hard 难度，禁止 easy 解答题
- 每道题必须生成三级提示（hint_1 方向提示、hint_2 关键线索、hint_3 接近答案）
- 严格按系统提示中的所有出题规则执行`,
		params.Topic, params.Difficulty, params.Count,
		params.Count, params.Difficulty)

	qm := NewQuizMasterAgent(te.LLM)
	result, _, _, err := qm.Generate(ctx, quizPrompt)
	if err != nil {
		return fmt.Sprintf(`{"error": "生成题目失败: %s"}`, err.Error()), nil
	}

	// 保存到数据库（含后置校验）
	for _, q := range result.Quizzes {
		// 校验 difficulty 合法性
		switch q.Difficulty {
		case "easy", "medium", "hard":
			// valid
		default:
			q.Difficulty = params.Difficulty
		}
		// 校验 type 合法性
		switch q.Type {
		case "choice", "fill", "judge", "short_answer":
			// valid
		default:
			q.Type = "choice"
		}
		// 解答题禁止 easy
		if q.Type == "short_answer" && q.Difficulty == "easy" {
			q.Difficulty = "medium"
		}
		optionsJSON, _ := json.Marshal(q.Options)
		quiz := model.Quiz{
			UserID:      te.UserID,
			Question:    q.Question,
			Type:        q.Type,
			Difficulty:  q.Difficulty,
			Options:     string(optionsJSON),
			Answer:      q.Answer,
			Explanation: q.Explanation,
			Hint1:       q.Hint1,
			Hint2:       q.Hint2,
			Hint3:       q.Hint3,
		}
		te.DB.Create(&quiz)
	}

	data, _ := json.Marshal(map[string]interface{}{
		"message": fmt.Sprintf("已生成 %d 道关于「%s」的 %s 难度题目", len(result.Quizzes), params.Topic, params.Difficulty),
		"count":   len(result.Quizzes),
	})
	return string(data), nil
}

// executeGenerateCard 调用 CardMaker Agent 生成知识卡片并保存到数据库
func (te *ToolExecutor) executeGenerateCard(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		Concept string `json:"concept"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return `{"error": "参数解析失败"}`, nil
	}

	cm := NewCardMakerAgent(te.LLM)
	result, _, _, err := cm.Generate(ctx, params.Concept)
	if err != nil {
		return fmt.Sprintf(`{"error": "生成卡片失败: %s"}`, err.Error()), nil
	}

	// 保存到数据库
	for _, c := range result.Cards {
		card := model.Card{
			UserID:     te.UserID,
			Concept:    c.Concept,
			Detail:     c.Detail,
			Formula:    c.Formula,
			MemoryTip:  c.MemoryTip,
			Difficulty: c.Difficulty,
			Tags:       c.Tags,
		}
		te.DB.Create(&card)
	}

	data, _ := json.Marshal(map[string]interface{}{
		"message": fmt.Sprintf("已为概念「%s」生成 %d 张知识卡片", params.Concept, len(result.Cards)),
		"count":   len(result.Cards),
	})
	return string(data), nil
}

// executeGetUserStats 从数据库查询真实的学习统计数据
func (te *ToolExecutor) executeGetUserStats(ctx context.Context, args json.RawMessage) (string, error) {
	var materialCount, cardCount, quizCount, correctCount, totalAttempts int64

	te.DB.Model(&model.Material{}).Where("user_id = ?", te.UserID).Count(&materialCount)
	te.DB.Model(&model.Card{}).Where("user_id = ?", te.UserID).Count(&cardCount)
	te.DB.Model(&model.Quiz{}).Where("user_id = ?", te.UserID).Count(&quizCount)
	te.DB.Model(&model.QuizAttempt{}).Where("user_id = ?", te.UserID).Count(&totalAttempts)
	te.DB.Model(&model.QuizAttempt{}).Where("user_id = ? AND is_correct = ?", te.UserID, true).Count(&correctCount)

	accuracy := float64(0)
	if totalAttempts > 0 {
		accuracy = float64(correctCount) / float64(totalAttempts) * 100
	}

	data, _ := json.Marshal(map[string]interface{}{
		"material_count": materialCount,
		"card_count":     cardCount,
		"quiz_count":     quizCount,
		"total_attempts": totalAttempts,
		"correct_count":  correctCount,
		"accuracy":       fmt.Sprintf("%.1f%%", accuracy),
	})
	return string(data), nil
}

// executeRecommendStudyPlan 根据学习数据生成个性化复习计划
func (te *ToolExecutor) executeRecommendStudyPlan(ctx context.Context, args json.RawMessage) (string, error) {
	var params struct {
		FocusAreas []string `json:"focus_areas"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		log.Printf("recommend_study_plan 参数解析失败: %v", err)
		// 解析失败时继续执行，使用空的 focus_areas
	}

	// 先获取用户的学习统计
	statsJSON, _ := te.executeGetUserStats(ctx, nil)

	// 获取用户的薄弱题目
	var wrongQuizzes []model.QuizAttempt
	te.DB.Where("user_id = ? AND is_correct = ?", te.UserID, false).
		Order("created_at DESC").Limit(10).Find(&wrongQuizzes)

	// 收集错题对应的题目信息
	var wrongTopics []string
	for _, attempt := range wrongQuizzes {
		var quiz model.Quiz
		if err := te.DB.Where("id = ?", attempt.QuizID).First(&quiz).Error; err == nil {
			wrongTopics = append(wrongTopics, quiz.Question)
		}
	}

	// 让 LLM 生成复习计划
	planPrompt := fmt.Sprintf(`请根据以下学习数据，为用户生成一份个性化的复习计划（中文，3-5 条建议）：

学习统计：%s
错题相关题目：%v
用户关注的重点领域：%v

请用 JSON 格式输出：{"plan": "复习计划内容", "suggestions": ["建议1", "建议2", ...]}`,
		statsJSON, wrongTopics, params.FocusAreas)

	reply, err := te.LLM.Chat(ctx, "你是一个学习规划助手，帮助用户制定高效的复习计划。", planPrompt)
	if err != nil {
		return `{"plan": "生成复习计划失败，请稍后重试"}`, nil
	}

	return reply, nil
}
