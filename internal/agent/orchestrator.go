package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"studyforge/internal/eval"
	"studyforge/internal/model"
	"studyforge/internal/rag"

	"gorm.io/gorm"
)

// AgentResult Agent 执行结果
type AgentResult struct {
	AgentName    string          `json:"agent_name"`
	MaterialID   string          `json:"material_id"`          // 分析的材料 ID（用于进度追踪）
	Status       string          `json:"status"`               // "success", "error"
	Data         json.RawMessage `json:"data"`                 // 各 Agent 的输出数据
	Error        string          `json:"error,omitempty"`
	Duration     time.Duration   `json:"duration_ms"`
	QualityScore float64         `json:"quality_score,omitempty"`  // Judge 评分
	JudgeComment string          `json:"judge_comment,omitempty"`  // Judge 评语
	InputTokens  int             `json:"-"`                        // LLM 输入 token 数
	OutputTokens int             `json:"-"`                       // LLM 输出 token 数
	UserID       string          `json:"-"`                       // 触发分析的用户 ID（不序列化到 JSON，用于 WebSocket 定向推送）
}

// Orchestrator Agent 编排器，负责并发调度多个 Agent
type Orchestrator struct {
	Analyst     *AnalystAgent
	QuizMaster  *QuizMasterAgent
	CardMaker   *CardMakerAgent
	MapBuilder  *MapBuilderAgent
	DB          *gorm.DB
	LLM         *LLMClient
	Judge       *eval.Judge
	VectorStore *rag.VectorStore
	ModelName   string                    // 实际使用的模型名（用于 Trace）
	OnResult    func(result AgentResult)  // 结果回调（用于 WebSocket 推送）
}

// NewOrchestrator 创建编排器
func NewOrchestrator(db *gorm.DB, apiKey, baseURL, modelName, embeddingModel string) *Orchestrator {
	llm := NewLLMClient(apiKey, baseURL, modelName, embeddingModel)
	return NewOrchestratorWithLLM(db, llm)
}

// NewOrchestratorWithLLM 使用已有的 LLMClient 创建编排器
func NewOrchestratorWithLLM(db *gorm.DB, llm *LLMClient) *Orchestrator {
	return &Orchestrator{
		Analyst:    NewAnalystAgent(llm),
		QuizMaster: NewQuizMasterAgent(llm),
		CardMaker:  NewCardMakerAgent(llm),
		MapBuilder: NewMapBuilderAgent(llm),
		DB:         db,
		LLM:        llm,
		Judge:      eval.NewJudge(llm.Chat),
		ModelName:  llm.ModelName,
	}
}

// ProcessMaterial 两阶段流水线分析学习材料
// Phase 1: Analyst 独享完整原文，提取结构化知识（摘要+知识点+关系）
// Phase 2: CardMaker/QuizMaster/MapBuilder 并发读取 Analyst 的精炼输出（短上下文，快速响应）
// 零信息丢失 + 大幅降低 Phase 2 延迟
func (o *Orchestrator) ProcessMaterial(ctx context.Context, material model.Material, userID string) error {
	log.Printf("开始分析材料: %s (ID: %s)", material.Title, material.ID)

	agentCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	var successCount, errorCount int

	// ========== 辅助函数：收集单个 Agent 结果 ==========
	collectResult := func(result AgentResult) {
		result.UserID = userID
		result.MaterialID = material.ID

		// Judge 质量评估
		if result.Status == "success" && o.Judge != nil {
			evalResult, err := o.Judge.Evaluate(agentCtx, material.Title, string(result.Data))
			if err != nil {
				log.Printf("Judge 评估 %s 失败: %v", result.AgentName, err)
			} else {
				result.QualityScore = evalResult.Overall
				result.JudgeComment = evalResult.Feedback
				log.Printf("Judge 评分 %s: %.1f 分 (相关性:%.1f 准确性:%.1f 完整性:%.1f) — %s",
					result.AgentName, evalResult.Overall,
					evalResult.Relevance, evalResult.Accuracy, evalResult.Completeness,
					evalResult.Feedback)
			}
		}

		// WebSocket 推送
		if o.OnResult != nil {
			o.OnResult(result)
		}

		// 持久化
		if result.Status == "success" {
			successCount++
			switch result.AgentName {
			case "Analyst":
				o.saveAnalysisData(material.ID, result.Data)
			case "CardMaker":
				o.saveCards(userID, material.ID, result.Data)
			case "QuizMaster":
				o.saveQuizzes(userID, material.ID, result.Data)
			case "MapBuilder":
				o.saveGraphData(material.ID, result.Data)
			}
		} else {
			errorCount++
		}

		o.saveTrace(userID, result)

		if result.Status == "error" {
			log.Printf("Agent %s 执行失败: %s", result.AgentName, result.Error)
		} else {
			log.Printf("Agent %s 完成，耗时 %v，质量分 %.1f", result.AgentName, result.Duration, result.QualityScore)
		}
	}

	// ========== Phase 1: Analyst 处理完整原文 ==========
	log.Printf("Phase 1: Analyst 分析完整原文 (%d 字)", len([]rune(material.Content)))
	start := time.Now()
	analystOutput, inputTokens, outputTokens, err := o.Analyst.Analyze(agentCtx, material.Content)
	analystResult := AgentResult{
		AgentName:    "Analyst",
		Duration:     time.Since(start),
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
	}
	if err != nil {
		analystResult.Status = "error"
		analystResult.Error = err.Error()
	} else {
		analystResult.Status = "success"
		if marshaled, err := json.Marshal(analystOutput); err != nil {
			log.Printf("Analyst 输出序列化失败: %v", err)
			analystResult.Status = "error"
			analystResult.Error = "输出序列化失败: " + err.Error()
		} else {
			analystResult.Data = marshaled
		}
	}
	collectResult(analystResult)

	// ========== 构建 Phase 2 的精炼上下文 ==========
	// 如果 Analyst 成功，用其结构化输出作为后续 Agent 的输入（~2000字 vs 原文可能 20000+字）
	// 如果 Analyst 失败，降级为截断原文
	var phase2Content string
	if analystResult.Status == "success" && analystOutput != nil {
		phase2Content = buildCompactContext(material.Title, analystOutput)
		log.Printf("Phase 2 上下文: %d 字（原文 %d 字，压缩率 %.0f%%）",
			len([]rune(phase2Content)), len([]rune(material.Content)),
			100.0*float64(len([]rune(phase2Content)))/float64(len([]rune(material.Content))))
	} else {
		log.Printf("Analyst 失败，Phase 2 降级使用截断原文")
		phase2Content = truncateContent(material.Content, 8000)
	}

	// ========== Phase 2: 三个 Agent 并发（短上下文，快速响应）==========
	log.Printf("Phase 2: CardMaker/QuizMaster/MapBuilder 并发执行")
	results := make(chan AgentResult, 3)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("CardMaker goroutine panic: %v", r)
				results <- AgentResult{AgentName: "CardMaker", Status: "error", Error: fmt.Sprintf("内部错误: %v", r)}
			}
		}()
		s := time.Now()
		data, in, out, e := o.CardMaker.Generate(agentCtx, phase2Content)
		r := AgentResult{AgentName: "CardMaker", Duration: time.Since(s), InputTokens: in, OutputTokens: out}
		if e != nil {
			r.Status = "error"
			r.Error = e.Error()
		} else {
			r.Status = "success"
			if m, err := json.Marshal(data); err != nil {
				r.Status = "error"
				r.Error = "输出序列化失败: " + err.Error()
			} else {
				r.Data = m
			}
		}
		results <- r
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("QuizMaster goroutine panic: %v", r)
				results <- AgentResult{AgentName: "QuizMaster", Status: "error", Error: fmt.Sprintf("内部错误: %v", r)}
			}
		}()
		s := time.Now()
		data, in, out, e := o.QuizMaster.Generate(agentCtx, phase2Content)
		r := AgentResult{AgentName: "QuizMaster", Duration: time.Since(s), InputTokens: in, OutputTokens: out}
		if e != nil {
			r.Status = "error"
			r.Error = e.Error()
		} else {
			r.Status = "success"
			if m, err := json.Marshal(data); err != nil {
				r.Status = "error"
				r.Error = "输出序列化失败: " + err.Error()
			} else {
				r.Data = m
			}
		}
		results <- r
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("MapBuilder goroutine panic: %v", r)
				results <- AgentResult{AgentName: "MapBuilder", Status: "error", Error: fmt.Sprintf("内部错误: %v", r)}
			}
		}()
		s := time.Now()
		data, in, out, e := o.MapBuilder.Generate(agentCtx, phase2Content)
		r := AgentResult{AgentName: "MapBuilder", Duration: time.Since(s), InputTokens: in, OutputTokens: out}
		if e != nil {
			r.Status = "error"
			r.Error = e.Error()
		} else {
			r.Status = "success"
			if m, err := json.Marshal(data); err != nil {
				r.Status = "error"
				r.Error = "输出序列化失败: " + err.Error()
			} else {
				r.Data = m
			}
		}
		results <- r
	}()

	// 收集 Phase 2 结果
	for i := 0; i < 3; i++ {
		collectResult(<-results)
	}

	// ========== 更新材料状态 ==========
	status := "completed"
	if successCount == 0 {
		status = "failed"
	} else if errorCount > 0 {
		status = "partial"
	}
	now := time.Now()
	if err := o.DB.Model(&model.Material{}).Where("id = ?", material.ID).Updates(map[string]interface{}{
		"status":      status,
		"analyzed_at": &now,
	}).Error; err != nil {
		log.Printf("更新材料状态失败: %v", err)
	}

	log.Printf("材料分析完成: %s (状态: %s, 成功: %d, 失败: %d)", material.Title, status, successCount, errorCount)

	if successCount == 0 {
		return fmt.Errorf("所有 Agent 均执行失败")
	}
	return nil
}

// buildCompactContext 将 Analyst 的结构化输出转换为精炼的文本上下文
// 供 CardMaker/QuizMaster/MapBuilder 使用，替代冗长的原始材料
func buildCompactContext(title string, output *AnalystOutput) string {
	var sb strings.Builder
	sb.WriteString("【材料标题】" + title + "\n\n")
	sb.WriteString("【摘要】" + output.Summary + "\n\n")

	if len(output.KeyPoints) > 0 {
		sb.WriteString("【核心知识点】\n")
		for i, kp := range output.KeyPoints {
			sb.WriteString(fmt.Sprintf("%d. %s（%s）: %s\n", i+1, kp.Concept, kp.Difficulty, kp.Detail))
		}
		sb.WriteString("\n")
	}

	if len(output.Relationships) > 0 {
		sb.WriteString("【概念关系】\n")
		for _, rel := range output.Relationships {
			sb.WriteString(fmt.Sprintf("- %s → %s: %s\n", rel.From, rel.To, rel.Type))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("【重要性】" + output.Importance + "\n")
	return sb.String()
}

// saveTrace 保存 LLM 调用追踪记录
func (o *Orchestrator) saveTrace(userID string, result AgentResult) {
	modelName := o.ModelName
	if modelName == "" {
		modelName = "unknown"
	}

	trace := model.LLMTrace{
		UserID:        userID,
		AgentName:     result.AgentName,
		Model:         modelName,
		InputTokens:   result.InputTokens,
		OutputTokens:  result.OutputTokens,
		TotalTokens:   result.InputTokens + result.OutputTokens,
		Latency:       result.Duration,
		LatencyMs:     result.Duration.Milliseconds(),
		QualityScore:  result.QualityScore,
		JudgeComment:  result.JudgeComment,
		Status:        result.Status,
		ErrorMessage:  result.Error,
		PromptSummary: result.AgentName + " analysis",
	}
	if err := o.DB.Create(&trace).Error; err != nil {
		log.Printf("保存 LLM Trace 失败: %v", err)
	}
}

// saveCards 将 CardMaker 的输出批量保存为知识卡片
func (o *Orchestrator) saveCards(userID, materialID string, data json.RawMessage) {
	var output CardMakerOutput
	if err := json.Unmarshal(data, &output); err != nil {
		log.Printf("CardMaker 输出解析失败: %v", err)
		return
	}

	if len(output.Cards) == 0 {
		return
	}

	// 构建批量插入的切片
	cards := make([]model.Card, 0, len(output.Cards))
	for _, item := range output.Cards {
		cards = append(cards, model.Card{
			UserID:     userID,
			MaterialID: materialID,
			Concept:    item.Concept,
			Detail:     item.Detail,
			Formula:    item.Formula,
			MemoryTip:  item.MemoryTip,
			Difficulty: item.Difficulty,
			Tags:       item.Tags,
		})
	}

	// 批量插入（每批 50 条），减少 DB round-trip
	if err := o.DB.CreateInBatches(&cards, 50).Error; err != nil {
		log.Printf("批量保存卡片失败: %v", err)
		return
	}
	log.Printf("已批量保存 %d 张知识卡片", len(cards))
}

// saveQuizzes 将 QuizMaster 的输出批量保存为练习题
func (o *Orchestrator) saveQuizzes(userID, materialID string, data json.RawMessage) {
	var output QuizMasterOutput
	if err := json.Unmarshal(data, &output); err != nil {
		log.Printf("QuizMaster 输出解析失败: %v", err)
		return
	}

	if len(output.Quizzes) == 0 {
		return
	}

	// 构建批量插入的切片
	quizzes := make([]model.Quiz, 0, len(output.Quizzes))
	for _, item := range output.Quizzes {
		// Options 需要从 []string 序列化为 JSON 字符串
		optionsJSON, marshalErr := json.Marshal(item.Options)
		if marshalErr != nil {
			log.Printf("题目选项序列化失败: %v", marshalErr)
			optionsJSON = []byte("[]")
		}

		quizzes = append(quizzes, model.Quiz{
			UserID:      userID,
			MaterialID:  materialID,
			Question:    item.Question,
			Type:        item.Type,
			Difficulty:  item.Difficulty,
			Options:     string(optionsJSON),
			Answer:      item.Answer,
			Explanation: item.Explanation,
			Hint1:       item.Hint1,
			Hint2:       item.Hint2,
			Hint3:       item.Hint3,
		})
	}

	// 批量插入（每批 50 条），减少 DB round-trip
	if err := o.DB.CreateInBatches(&quizzes, 50).Error; err != nil {
		log.Printf("批量保存题目失败: %v", err)
		return
	}
	log.Printf("已批量保存 %d 道练习题", len(quizzes))
}

// saveGraphData 将 MapBuilder 的输出保存为知识图谱数据
func (o *Orchestrator) saveGraphData(materialID string, data json.RawMessage) {
	// data 已经是 JSON 格式的 nodes+edges，直接存为字符串
	graphJSON := string(data)
	if err := o.DB.Model(&model.Material{}).Where("id = ?", materialID).
		Update("graph_data", graphJSON).Error; err != nil {
		log.Printf("保存图谱数据失败: %v", err)
	}
	log.Printf("已保存知识图谱数据")
}

// saveAnalysisData 将 Analyst 的输出保存为分析数据（摘要、知识点、关系）
func (o *Orchestrator) saveAnalysisData(materialID string, data json.RawMessage) {
	analysisJSON := string(data)
	if err := o.DB.Model(&model.Material{}).Where("id = ?", materialID).
		Update("analysis_data", analysisJSON).Error; err != nil {
		log.Printf("保存分析数据失败: %v", err)
	}
	log.Printf("已保存分析数据")
}

// truncateContent 智能截断文本：保留头部 60% + 尾部 40%，砍掉中间冗余段落
// 开头通常含概述/定义，结尾含总结/要点，比简单截断更保留关键信息
func truncateContent(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}

	headLen := maxLen * 6 / 10 // 头部保留 60%
	tailLen := maxLen - headLen // 尾部保留 40%

	head := runes[:headLen]
	tail := runes[len(runes)-tailLen:]

	// 尝试在段落/换行处断开，避免截断到句子中间
	if idx := lastIndex(head, '\n'); idx > headLen*7/10 {
		head = head[:idx+1]
	}
	if idx := indexByte(tail, '\n'); idx >= 0 && idx < tailLen*3/10 {
		tail = tail[idx+1:]
	}

	return string(head) + "\n\n[...中间内容已省略，以下为结尾部分...]\n\n" + string(tail)
}

// lastIndex 在 rune 切片中从后往前查找字符
func lastIndex(runes []rune, ch rune) int {
	for i := len(runes) - 1; i >= 0; i-- {
		if runes[i] == ch {
			return i
		}
	}
	return -1
}

// indexByte 在 rune 切片中从前往后查找字符
func indexByte(runes []rune, ch rune) int {
	for i, r := range runes {
		if r == ch {
			return i
		}
	}
	return -1
}

// extractJSON 从 LLM 输出中提取 JSON 内容（去除 markdown 代码块包裹）
// LLM 常以 ```json\n{...}\n``` 形式输出，直接 json.Unmarshal 会失败
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return s
	}

	// 去掉 ```json 或 ``` 开头
	if strings.HasPrefix(s, "```") {
		// 跳过第一行（```json 或 ```）
		nl := strings.Index(s, "\n")
		if nl >= 0 {
			s = s[nl+1:]
		}
		// 去掉末尾的 ```
		if idx := strings.LastIndex(s, "\n```"); idx >= 0 {
			s = s[:idx]
		} else if strings.HasSuffix(s, "```") {
			s = s[:len(s)-3]
		}
		s = strings.TrimSpace(s)
	}

	// 如果仍不以 { 或 [ 开头，尝试找到第一个 JSON 对象/数组
	if len(s) > 0 && s[0] != '{' && s[0] != '[' {
		start := -1
		for i, ch := range s {
			if ch == '{' || ch == '[' {
				start = i
				break
			}
		}
		if start >= 0 {
			s = s[start:]
		}
	}

	// 从末尾截断到最后一个 } 或 ]
	if len(s) > 0 {
		lastClose := -1
		for i := len(s) - 1; i >= 0; i-- {
			if s[i] == '}' || s[i] == ']' {
				lastClose = i
				break
			}
		}
		if lastClose >= 0 && lastClose < len(s)-1 {
			s = s[:lastClose+1]
		}
	}

	return s
}
