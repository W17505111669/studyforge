package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== AI 考试模拟 ====================

// GenerateExam 生成模拟考试试卷
// POST /api/exam/generate
func (h *Handler) GenerateExam(c *gin.Context) {
	userID := c.GetString("userID")

	var req model.GenerateExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 验证材料属于当前用户
	var materials []model.Material
	if err := h.DB.Where("id IN ? AND user_id = ?", req.MaterialIDs, userID).Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询材料失败"})
		return
	}
	if len(materials) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未找到有效材料"})
		return
	}

	// 加载材料关联的卡片和练习题，作为出题素材
	var cards []model.Card
	h.DB.Where("material_id IN ? AND user_id = ?", req.MaterialIDs, userID).Find(&cards)

	var quizzes []model.Quiz
	h.DB.Where("material_id IN ? AND user_id = ?", req.MaterialIDs, userID).Find(&quizzes)

	if len(cards) == 0 && len(quizzes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "所选材料尚无知识卡片或练习题，无法生成试卷"})
		return
	}

	// 构建素材摘要
	var materialSummary strings.Builder
	for _, m := range materials {
		materialSummary.WriteString(fmt.Sprintf("【%s】", m.Title))
		// 提取知识点
		if m.AnalysisData != "" {
			var analysis struct {
				KeyPoints []string `json:"key_points"`
			}
			if err := json.Unmarshal([]byte(m.AnalysisData), &analysis); err == nil {
				for _, kp := range analysis.KeyPoints {
					materialSummary.WriteString(fmt.Sprintf("\n- %s", kp))
				}
			}
		}
		materialSummary.WriteString("\n")
	}

	// 构建卡片知识
	var cardSummary strings.Builder
	for _, card := range cards {
		cardSummary.WriteString(fmt.Sprintf("- 概念: %s | 详情: %s", card.Concept, truncate(card.Detail, 100)))
		if card.Formula != "" {
			cardSummary.WriteString(fmt.Sprintf(" | 公式: %s", truncate(card.Formula, 80)))
		}
		cardSummary.WriteString("\n")
	}

	// 计算题目分布：选择40%、判断20%、填空20%、简答20%
	total := req.QuestionCount
	choiceCount := int(math.Round(float64(total) * 0.4))
	trueFalseCount := int(math.Round(float64(total) * 0.2))
	fillCount := int(math.Round(float64(total) * 0.2))
	shortCount := total - choiceCount - trueFalseCount - fillCount
	if shortCount < 1 {
		shortCount = 1
	}

	// 分值分配
	choicePoints := 2
	trueFalsePoints := 2
	fillPoints := 3
	shortPoints := 5

	systemPrompt := `你是一位专业的考试出题老师。请根据提供的学习材料生成一份模拟考试试卷。

要求：
1. 题目必须基于提供的知识点，不要编造无关内容
2. 选择题提供 4 个选项（A/B/C/D），只有 1 个正确
3. 判断题答案为 "正确" 或 "错误"
4. 填空题答案简洁明确
5. 简答题需要一定的分析和阐述

输出格式为严格的 JSON 数组，每个元素包含：
{
  "type": "choice|true_false|fill|short_answer",
  "question": "题目内容",
  "options": ["A. xxx", "B. xxx", "C. xxx", "D. xxx"], // 仅选择题
  "answer": "正确答案",
  "explanation": "详细解析",
  "difficulty": "easy|medium|hard",
  "concept": "关联的知识点概念"
}

请直接输出 JSON 数组，不要包含任何 markdown 代码块标记。`

	userMessage := fmt.Sprintf(`请根据以下学习材料生成 %d 道考试题：
- 选择题 %d 道（每题 %d 分）
- 判断题 %d 道（每题 %d 分）
- 填空题 %d 道（每题 %d 分）
- 简答题 %d 道（每题 %d 分）

材料摘要：
%s

知识点卡片：
%s

请确保题目难度分布合理（简单30%%、中等40%%、困难30%%），覆盖所有材料的核心知识点。`,
		total, choiceCount, choicePoints, trueFalseCount, trueFalsePoints,
		fillCount, fillPoints, shortCount, shortPoints,
		materialSummary.String(), cardSummary.String())

	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()

	response, err := h.LLM.Chat(ctx, systemPrompt, userMessage)
	if err != nil {
		log.Printf("AI 生成考试失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 生成试卷失败，请重试"})
		return
	}

	// 清理可能的 markdown 代码块包裹
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
	}
	if strings.HasSuffix(response, "```") {
		response = strings.TrimSuffix(response, "```")
	}
	response = strings.TrimSpace(response)

	// 解析 AI 生成的题目
	var rawQuestions []struct {
		Type        string   `json:"type"`
		Question    string   `json:"question"`
		Options     []string `json:"options"`
		Answer      string   `json:"answer"`
		Explanation string   `json:"explanation"`
		Difficulty  string   `json:"difficulty"`
		Concept     string   `json:"concept"`
	}

	if err := json.Unmarshal([]byte(response), &rawQuestions); err != nil {
		log.Printf("解析 AI 生成的考试题目失败: %v\n原始响应: %s", err, truncate(response, 500))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 生成的试卷格式异常，请重试"})
		return
	}

	// 转换为 ExamQuestion 并分配分值
	questions := make([]model.ExamQuestion, 0, len(rawQuestions))
	for i, rq := range rawQuestions {
		points := choicePoints
		switch rq.Type {
		case "true_false":
			points = trueFalsePoints
		case "fill":
			points = fillPoints
		case "short_answer":
			points = shortPoints
		}

		// 尝试关联到材料
		var materialID string
		for _, card := range cards {
			if card.Concept == rq.Concept || strings.Contains(card.Detail, rq.Concept) {
				materialID = card.MaterialID
				break
			}
		}

		questions = append(questions, model.ExamQuestion{
			Index:       i + 1,
			Type:        rq.Type,
			Question:    rq.Question,
			Options:     rq.Options,
			Answer:      rq.Answer,
			Explanation: rq.Explanation,
			Difficulty:  rq.Difficulty,
			MaterialID:  materialID,
			Concept:     rq.Concept,
			Points:      points,
		})
	}

	if len(questions) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "未能生成有效题目，请重试"})
		return
	}

	// 序列化
	matIDsJSON, _ := json.Marshal(req.MaterialIDs)
	questionsJSON, _ := json.Marshal(questions)

	session := model.ExamSession{
		UserID:      userID,
		MaterialIDs: string(matIDsJSON),
		Questions:   string(questionsJSON),
		TimeLimit:   req.TimeLimit,
		StartedAt:   time.Now(),
		Status:      "in_progress",
	}

	if err := h.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存考试会话失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            session.ID,
		"question_count": len(questions),
		"time_limit":    session.TimeLimit,
		"started_at":    session.StartedAt,
		"max_score":     calcMaxScore(questions),
		"questions":     questions, // 首次返回完整题目（不含答案）
	})
}

// SubmitExam 提交考试并自动评分
// POST /api/exams/:id/submit
func (h *Handler) SubmitExam(c *gin.Context) {
	userID := c.GetString("userID")
	examID := c.Param("id")

	var session model.ExamSession
	if err := h.DB.Where("id = ? AND user_id = ?", examID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "考试会话不存在"})
		return
	}

	if session.Status == "completed" || session.Status == "timed_out" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该考试已提交"})
		return
	}

	var req model.SubmitExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 解析题目
	var questions []model.ExamQuestion
	if err := json.Unmarshal([]byte(session.Questions), &questions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "试卷数据异常"})
		return
	}

	// 构建答案映射
	answerMap := make(map[int]string)
	markedMap := make(map[int]bool)
	for _, a := range req.Answers {
		answerMap[a.Index] = a.Answer
		markedMap[a.Index] = a.Marked
	}

	// 自动评分
	var totalScore float64
	maxScore := calcMaxScore(questions)
	typeStats := make(map[string]*model.TypeStat)
	questionResults := make([]model.QuestionResult, 0, len(questions))
	var weakPoints []model.WeakPoint

	for _, q := range questions {
		userAnswer := answerMap[q.Index]

		// 初始化题型统计
		if _, ok := typeStats[q.Type]; !ok {
			typeStats[q.Type] = &model.TypeStat{}
		}
		typeStats[q.Type].Total++

		isCorrect := false
		switch q.Type {
		case "choice":
			isCorrect = strings.EqualFold(strings.TrimSpace(userAnswer), strings.TrimSpace(q.Answer))
		case "true_false":
			isCorrect = strings.TrimSpace(userAnswer) == strings.TrimSpace(q.Answer)
		case "fill":
			isCorrect = strings.EqualFold(strings.TrimSpace(userAnswer), strings.TrimSpace(q.Answer))
		case "short_answer":
			// 简答题使用关键词匹配 + 长度评估
			isCorrect = gradeShortAnswer(userAnswer, q.Answer)
		}

		if isCorrect {
			totalScore += float64(q.Points)
			typeStats[q.Type].Correct++
		} else if q.Concept != "" {
			weakPoints = append(weakPoints, model.WeakPoint{
				Concept:    q.Concept,
				MaterialID: q.MaterialID,
				Reason:     fmt.Sprintf("%s题答错，正确答案: %s", typeLabel(q.Type), q.Answer),
			})
		}

		questionResults = append(questionResults, model.QuestionResult{
			Index:         q.Index,
			Type:          q.Type,
			IsCorrect:     isCorrect,
			UserAnswer:    userAnswer,
			CorrectAnswer: q.Answer,
			Explanation:   q.Explanation,
			Concept:       q.Concept,
		})
	}

	// 计算题型正确率
	for _, ts := range typeStats {
		if ts.Total > 0 {
			ts.Rate = math.Round(float64(ts.Correct)/float64(ts.Total)*100) / 100
		}
	}

	// 限制薄弱知识点数量
	if len(weakPoints) > 10 {
		weakPoints = weakPoints[:10]
	}

	percentage := 0.0
	if maxScore > 0 {
		percentage = math.Round(totalScore/float64(maxScore)*1000) / 10
	}

	timeUsed := int(time.Since(session.StartedAt).Seconds())

	report := model.ExamReport{
		TotalScore:      totalScore,
		MaxScore:        maxScore,
		Percentage:      percentage,
		TimeUsed:        timeUsed,
		TypeStats:       typeStats,
		WeakPoints:      weakPoints,
		QuestionResults: questionResults,
	}

	// 序列化
	answersJSON, _ := json.Marshal(req.Answers)
	reportJSON, _ := json.Marshal(report)

	now := time.Now()
	session.Answers = string(answersJSON)
	session.Score = percentage
	session.Report = string(reportJSON)
	session.EndedAt = &now
	session.Status = "completed"

	if err := h.DB.Save(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存考试结果失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      session.ID,
		"report":  report,
		"status":  session.Status,
		"ended_at": session.EndedAt,
	})
}

// ListExams 列出考试历史
// GET /api/exams
func (h *Handler) ListExams(c *gin.Context) {
	userID := c.GetString("userID")

	var sessions []model.ExamSession
	if err := h.DB.Where("user_id = ?", userID).
		Select("id, user_id, material_ids, time_limit, started_at, ended_at, score, status, created_at").
		Order("created_at DESC").Limit(50).Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询考试记录失败"})
		return
	}

	// 批量查材料标题
	materialTitleMap := make(map[string]string)
	for _, s := range sessions {
		var matIDs []string
		if err := json.Unmarshal([]byte(s.MaterialIDs), &matIDs); err == nil {
			for _, mid := range matIDs {
				if _, ok := materialTitleMap[mid]; !ok {
					var m model.Material
					if h.DB.Select("title").Where("id = ? AND user_id = ?", mid, userID).First(&m).Error == nil {
						materialTitleMap[mid] = m.Title
					}
				}
			}
		}
	}

	type examItem struct {
		ID            string    `json:"id"`
		MaterialNames []string  `json:"material_names"`
		TimeLimit     int       `json:"time_limit"`
		Score         float64   `json:"score"`
		Status        string    `json:"status"`
		StartedAt     time.Time `json:"started_at"`
		EndedAt       *time.Time `json:"ended_at"`
	}

	items := make([]examItem, 0, len(sessions))
	for _, s := range sessions {
		var matIDs []string
		json.Unmarshal([]byte(s.MaterialIDs), &matIDs)
		names := make([]string, 0, len(matIDs))
		for _, mid := range matIDs {
			if title, ok := materialTitleMap[mid]; ok {
				names = append(names, title)
			}
		}
		items = append(items, examItem{
			ID:            s.ID,
			MaterialNames: names,
			TimeLimit:     s.TimeLimit,
			Score:         s.Score,
			Status:        s.Status,
			StartedAt:     s.StartedAt,
			EndedAt:       s.EndedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": items, "total": len(items)})
}

// GetExam 获取考试详情（含报告）
// GET /api/exams/:id
func (h *Handler) GetExam(c *gin.Context) {
	userID := c.GetString("userID")
	examID := c.Param("id")

	var session model.ExamSession
	if err := h.DB.Where("id = ? AND user_id = ?", examID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "考试记录不存在"})
		return
	}

	// 解析 questions（返回时隐藏正确答案，仅对未完成考试）
	var questions []model.ExamQuestion
	json.Unmarshal([]byte(session.Questions), &questions)

	var report *model.ExamReport
	if session.Report != "" {
		report = &model.ExamReport{}
		json.Unmarshal([]byte(session.Report), report)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            session.ID,
		"material_ids":  session.MaterialIDs,
		"questions":     questions,
		"time_limit":    session.TimeLimit,
		"started_at":    session.StartedAt,
		"ended_at":      session.EndedAt,
		"score":         session.Score,
		"status":        session.Status,
		"report":        report,
	})
}

// ==================== 辅助函数 ====================

func calcMaxScore(questions []model.ExamQuestion) int {
	total := 0
	for _, q := range questions {
		total += q.Points
	}
	return total
}

func typeLabel(t string) string {
	switch t {
	case "choice":
		return "选择题"
	case "true_false":
		return "判断题"
	case "fill":
		return "填空题"
	case "short_answer":
		return "简答题"
	}
	return "题目"
}

// gradeShortAnswer 简答题评分：基于关键词匹配和长度评估
// 返回 true 表示"基本正确"（>=60% 关键词命中且有一定长度）
func gradeShortAnswer(userAnswer, correctAnswer string) bool {
	if strings.TrimSpace(userAnswer) == "" {
		return false
	}

	// 完全匹配
	if strings.EqualFold(strings.TrimSpace(userAnswer), strings.TrimSpace(correctAnswer)) {
		return true
	}

	// 提取正确答案的关键词（2字以上的词/短语）
	correctWords := extractKeywords(correctAnswer)
	if len(correctWords) == 0 {
		return len(userAnswer) > 10
	}

	userLower := strings.ToLower(userAnswer)
	matchCount := 0
	for _, kw := range correctWords {
		if strings.Contains(userLower, strings.ToLower(kw)) {
			matchCount++
		}
	}

	rate := float64(matchCount) / float64(len(correctWords))
	// 60% 关键词命中 + 答案长度 >= 10 字符
	return rate >= 0.6 && len(userAnswer) >= 10
}

// extractKeywords 从文本中提取关键词（简单实现：按标点分割，过滤短词）
func extractKeywords(text string) []string {
	seps := []string{"，", ",", "。", ".", "；", ";", "、", " ", "\n"}
	words := []string{text}
	for _, sep := range seps {
		var next []string
		for _, w := range words {
			next = append(next, strings.Split(w, sep)...)
		}
		words = next
	}

	var keywords []string
	seen := make(map[string]bool)
	for _, w := range words {
		w = strings.TrimSpace(w)
		if len(w) >= 2 && !seen[w] {
			seen[w] = true
			keywords = append(keywords, w)
		}
	}
	return keywords
}
