package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"studyforge/internal/agent"
	"studyforge/internal/memory"
	"studyforge/internal/model"

	openai "github.com/sashabaranov/go-openai"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ==================== 知识卡片 ====================

// ExportCards 导出卡片为 Anki 兼容 CSV（HTML 富文本格式）
// GET /api/cards/export?material_id=xxx&difficulty=xxx
func (h *Handler) ExportCards(c *gin.Context) {
	userID := c.GetString("userID")
	materialID := c.Query("material_id")
	difficulty := c.Query("difficulty")

	query := h.DB.Where("user_id = ?", userID)
	if materialID != "" {
		query = query.Where("material_id = ?", materialID)
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}

	var cards []model.Card
	if err := query.Order("created_at DESC").Find(&cards).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询卡片失败"})
		return
	}

	if len(cards) == 0 {
		c.JSON(http.StatusOK, gin.H{"error": "没有可导出的卡片"})
		return
	}

	// 难度标签颜色和文本
	diffBadge := func(d string) string {
		switch d {
		case "easy":
			return `<span style="background:#d1fae5;color:#065f46;font-size:11px;padding:2px 8px;border-radius:4px;margin-left:8px;">简单</span>`
		case "medium":
			return `<span style="background:#fef3c7;color:#92400e;font-size:11px;padding:2px 8px;border-radius:4px;margin-left:8px;">中等</span>`
		case "hard":
			return `<span style="background:#fee2e2;color:#991b1b;font-size:11px;padding:2px 8px;border-radius:4px;margin-left:8px;">困难</span>`
		}
		return ""
	}

	var sb strings.Builder
	// UTF-8 BOM
	sb.WriteString("\xEF\xBB\xBF")
	// Anki 导入指令
	sb.WriteString("#separator:tab\n")
	sb.WriteString("#html:true\n")
	sb.WriteString("#tags column:4\n")
	sb.WriteString("Front\tBack\tNotes\tTags\n")

	for _, card := range cards {
		// ── 正面：概念名 + 难度标签 ──
		front := `<div style="font-size:20px;font-weight:bold;color:#1f2937;">` +
			htmlEsc(card.Concept) + diffBadge(card.Difficulty) + `</div>`

		// ── 背面：结构化详情 ──
		var backParts []string

		if card.Detail != "" {
			backParts = append(backParts,
				`<div style="margin-bottom:12px;">`+
					`<div style="font-size:13px;color:#6b7280;margin-bottom:4px;font-weight:600;">📖 详解</div>`+
					`<div style="font-size:14px;color:#374151;line-height:1.6;">`+htmlNL(card.Detail)+`</div>`+
					`</div>`)
		}

		if card.Formula != "" {
			backParts = append(backParts,
				`<div style="margin-bottom:12px;padding:10px 14px;background:#f0f9ff;border:1px solid #bae6fd;border-radius:8px;">`+
					`<div style="font-size:12px;color:#0369a1;margin-bottom:4px;font-weight:600;">📐 公式 / 代码</div>`+
					`<div style="font-size:14px;color:#1e40af;font-family:monospace;word-break:break-all;">`+htmlNL(card.Formula)+`</div>`+
					`</div>`)
		}

		if card.MemoryTip != "" {
			backParts = append(backParts,
				`<div style="padding:10px 14px;background:#fffbeb;border:1px solid #fde68a;border-radius:8px;">`+
					`<div style="font-size:12px;color:#b45309;margin-bottom:4px;font-weight:600;">💡 记忆技巧</div>`+
					`<div style="font-size:13px;color:#78350f;line-height:1.5;">`+htmlNL(card.MemoryTip)+`</div>`+
					`</div>`)
		}

		back := strings.Join(backParts, "")

		// 个人笔记
		note := strings.TrimSpace(card.UserNote)

		// 标签
		tags := strings.ReplaceAll(card.Tags, ",", ";")
		tags = strings.TrimSpace(tags)

		sb.WriteString(csvEscape(front) + "\t" + csvEscape(back) + "\t" + csvEscape(note) + "\t" + tags + "\n")
	}

	filename := fmt.Sprintf("studyforge_cards_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.String(http.StatusOK, sb.String())
}

// htmlEsc HTML 转义
func htmlEsc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// htmlNL HTML 转义 + 换行转 <br>
func htmlNL(s string) string {
	return strings.ReplaceAll(htmlEsc(s), "\n", "<br>")
}

// csvEscape 转义 CSV 字段中的特殊字符
func csvEscape(s string) string {
	if strings.ContainsAny(s, "\t\n\"") {
		s = strings.ReplaceAll(s, "\"", "\"\"")
		return "\"" + s + "\""
	}
	return s
}

// ListCards 获取用户的所有知识卡片（支持分页）
// GET /api/cards?material_id=xxx&difficulty=xxx&due=true&bookmarked=true&limit=20&offset=0
func (h *Handler) ListCards(c *gin.Context) {
	userID := c.GetString("userID")
	materialID := c.Query("material_id")
	difficulty := c.Query("difficulty")
	due := c.Query("due")
	bookmarked := c.Query("bookmarked")
	limit, offset := parsePagination(c)

	query := h.DB.Where("user_id = ?", userID)
	if materialID != "" {
		query = query.Where("material_id = ?", materialID)
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	if due == "true" {
		now := time.Now()
		query = query.Where("next_review_at IS NULL OR next_review_at <= ?", now)
	}
	if bookmarked == "true" {
		query = query.Where("is_bookmarked = ?", true)
	}

	var total int64
	query.Model(&model.Card{}).Count(&total)

	var cards []model.Card
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&cards).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询卡片失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": cards, "total": total, "limit": limit, "offset": offset})
}

// GetCard 获取单个卡片详情
// GET /api/cards/:id
func (h *Handler) GetCard(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")

	var card model.Card
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "卡片不存在"})
		return
	}

	c.JSON(http.StatusOK, card)
}

// ReviewCardRequest 复习反馈请求
type ReviewCardRequest struct {
	Result string `json:"result" binding:"required"` // "mastered" 或 "review"
}

// ReviewCard 提交卡片复习结果（简化 SM-2 间隔重复算法）
// POST /api/cards/:id/review
func (h *Handler) ReviewCard(c *gin.Context) {
	userID := c.GetString("userID")
	cardID := c.Param("id")

	var req ReviewCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Result != "mastered" && req.Result != "review" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "result 必须是 mastered 或 review"})
		return
	}

	var card model.Card
	if err := h.DB.Where("id = ? AND user_id = ?", cardID, userID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "卡片不存在"})
		return
	}

	// 使用 Card 模型的 ApplyReview 方法执行 SM-2 间隔重复算法
	card.ApplyReview(req.Result)

	h.DB.Save(&card)

	c.JSON(http.StatusOK, gin.H{
		"card_id":        card.ID,
		"review_count":   card.ReviewCount,
		"interval_days":  card.IntervalDays,
		"ease_factor":    card.EaseFactor,
		"next_review_at": card.NextReviewAt,
		"result":         req.Result,
	})
}

// ==================== 练习题 ====================

// NoteRequest 更新卡片笔记请求
type NoteRequest struct {
	Note string `json:"note"` // 笔记内容（可为空以清除）
}

// ToggleBookmark 切换卡片书签状态
// PUT /api/cards/:id/bookmark
func (h *Handler) ToggleBookmark(c *gin.Context) {
	userID := c.GetString("userID")
	cardID := c.Param("id")

	var card model.Card
	if err := h.DB.Where("id = ? AND user_id = ?", cardID, userID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "卡片不存在"})
		return
	}

	card.IsBookmarked = !card.IsBookmarked
	h.DB.Model(&card).Update("is_bookmarked", card.IsBookmarked)

	c.JSON(http.StatusOK, gin.H{
		"card_id":       card.ID,
		"is_bookmarked": card.IsBookmarked,
	})
}

// UpdateCardNote 更新卡片个人笔记
// PUT /api/cards/:id/note
func (h *Handler) UpdateCardNote(c *gin.Context) {
	userID := c.GetString("userID")
	cardID := c.Param("id")

	var req NoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var card model.Card
	if err := h.DB.Where("id = ? AND user_id = ?", cardID, userID).First(&card).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "卡片不存在"})
		return
	}

	card.UserNote = req.Note
	h.DB.Model(&card).Update("user_note", card.UserNote)

	c.JSON(http.StatusOK, gin.H{
		"card_id":   card.ID,
		"user_note": card.UserNote,
	})
}

// ListQuizzes 获取练习题列表（支持分页+智能推荐模式）
// GET /api/quizzes?material_id=xxx&difficulty=xxx&recommended=true&limit=20&offset=0
func (h *Handler) ListQuizzes(c *gin.Context) {
	userID := c.GetString("userID")
	materialID := c.Query("material_id")
	difficulty := c.Query("difficulty")
	recommended := c.Query("recommended")
	limit, offset := parsePagination(c)

	query := h.DB.Where("user_id = ?", userID)
	if materialID != "" {
		query = query.Where("material_id = ?", materialID)
	}

	// 智能推荐模式：根据用户水平推荐匹配难度的题目
	var recommendedDiffs []string
	var recommendReason string
	var isRecommended bool
	if recommended == "true" {
		isRecommended = true
		_, _, recommendedDiffs, recommendReason = h.analyzeUserLevel(userID)
		if len(recommendedDiffs) > 0 {
			query = query.Where("difficulty IN ?", recommendedDiffs)
		}
	} else if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}

	var total int64
	query.Model(&model.Quiz{}).Count(&total)

	var quizzes []model.Quiz
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&quizzes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询题目失败"})
		return
	}

	resp := gin.H{"data": quizzes, "total": total, "limit": limit, "offset": offset}
	if isRecommended {
		resp["recommended"] = true
		resp["recommended_difficulty"] = recommendedDiffs
		resp["recommended_reason"] = recommendReason
	}

	c.JSON(http.StatusOK, resp)
}

// ==================== 自适应出题 ====================

// analyzeUserLevel 分析用户最近 20 道题的正确率，返回推荐难度
func (h *Handler) analyzeUserLevel(userID string) (accuracy float64, totalAttempts int64, recommendedDiffs []string, reason string) {
	var attempts []model.QuizAttempt
	h.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(20).Find(&attempts)

	totalAttempts = int64(len(attempts))
	if totalAttempts == 0 {
		return 0, 0, []string{"easy", "medium", "hard"}, "暂无答题记录，推荐全部难度"
	}

	var correct int64
	for _, a := range attempts {
		if a.IsCorrect {
			correct++
		}
	}
	accuracy = float64(correct) / float64(totalAttempts) * 100

	if accuracy > 80 {
		recommendedDiffs = []string{"medium", "hard"}
		reason = fmt.Sprintf("最近 %d 题正确率 %.0f%%，推荐中高难度挑战", totalAttempts, accuracy)
	} else if accuracy < 50 {
		recommendedDiffs = []string{"easy", "medium"}
		reason = fmt.Sprintf("最近 %d 题正确率 %.0f%%，推荐从基础巩固", totalAttempts, accuracy)
	} else {
		recommendedDiffs = []string{"easy", "medium", "hard"}
		reason = fmt.Sprintf("最近 %d 题正确率 %.0f%%，推荐均衡练习", totalAttempts, accuracy)
	}

	return accuracy, totalAttempts, recommendedDiffs, reason
}

// GetDifficultyLevel 获取用户当前难度等级和推荐
// GET /api/quizzes/difficulty-level
func (h *Handler) GetDifficultyLevel(c *gin.Context) {
	userID := c.GetString("userID")

	accuracy, totalAttempts, recommendedDiffs, reason := h.analyzeUserLevel(userID)

	// 统计各难度答题数
	var diffStats []struct {
		Difficulty string `json:"difficulty"`
		Total      int64  `json:"total"`
		Correct    int64  `json:"correct"`
	}
	h.DB.Model(&model.QuizAttempt{}).
		Select("quizzes.difficulty, COUNT(*) as total, SUM(CASE WHEN quiz_attempts.is_correct THEN 1 ELSE 0 END) as correct").
		Joins("JOIN quizzes ON quizzes.id = quiz_attempts.quiz_id").
		Where("quiz_attempts.user_id = ?", userID).
		Group("quizzes.difficulty").
		Scan(&diffStats)

	// 统计各难度题目总数
	var quizCounts []struct {
		Difficulty string `json:"difficulty"`
		Count      int64  `json:"count"`
	}
	h.DB.Model(&model.Quiz{}).
		Select("difficulty, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("difficulty").
		Scan(&quizCounts)

	quizCountMap := map[string]int64{}
	for _, qc := range quizCounts {
		quizCountMap[qc.Difficulty] = qc.Count
	}

	c.JSON(http.StatusOK, gin.H{
		"accuracy":              accuracy,
		"total_attempts":        totalAttempts,
		"recommended_difficulty": recommendedDiffs,
		"reason":                reason,
		"difficulty_stats":      diffStats,
		"quiz_counts":           quizCountMap,
	})
}

// AnswerQuiz 提交答案
// POST /api/quizzes/:id/answer
func (h *Handler) AnswerQuiz(c *gin.Context) {
	userID := c.GetString("userID")
	quizID := c.Param("id")

	var req model.AnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找题目（限定当前用户的题目）
	var quiz model.Quiz
	if err := h.DB.Where("id = ? AND user_id = ?", quizID, userID).First(&quiz).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
		return
	}

	// 智能判断答案是否正确
	isCorrect := checkAnswer(req.Answer, quiz.Answer, quiz.Options, quiz.Type)

	attempt := model.QuizAttempt{
		UserID:    userID,
		QuizID:    quizID,
		Answer:    req.Answer,
		IsCorrect: isCorrect,
		HintsUsed: req.HintsUsed,
	}
	h.DB.Create(&attempt)

	// 答错时自动收集到错题本（upsert 防止重复）
	if !isCorrect {
		var existing model.QuizMistake
		err := h.DB.Where("user_id = ? AND quiz_id = ?", userID, quizID).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			mistake := model.QuizMistake{
				UserID:        userID,
				QuizID:        quizID,
				UserAnswer:    req.Answer,
				CorrectAnswer: quiz.Answer,
				MistakeAt:     time.Now(),
			}
			h.DB.Create(&mistake)
		} else if err == nil {
			// 已存在则更新答案和时间、重置复习状态
			h.DB.Model(&existing).Updates(map[string]interface{}{
				"user_answer":    req.Answer,
				"correct_answer": quiz.Answer,
				"mistake_at":     time.Now(),
				"reviewed":       false,
			})
		}
	}

	resp := gin.H{
		"is_correct":  isCorrect,
		"answer":      quiz.Answer,
		"explanation": quiz.Explanation,
	}

	// 答错时附带第1级提示（如果存在）
	if !isCorrect && quiz.Hint1 != "" {
		resp["hint_1"] = quiz.Hint1
	}

	c.JSON(http.StatusOK, resp)
}

// checkAnswer 智能判断答案是否正确
// 支持：字母对比（"B"）、选项文本匹配（"B. xxx"）、判断题归一化、以及直接文本对比
func checkAnswer(userAnswer, correctAnswer, optionsJSON, quizType string) bool {
	user := strings.TrimSpace(userAnswer)
	correct := strings.TrimSpace(correctAnswer)

	// 直接匹配
	if user == correct {
		return true
	}

	// 大小写不敏感匹配
	if strings.EqualFold(user, correct) {
		return true
	}

	// 判断题特殊处理：归一化 "正确"/"对"/"T"/"true" 和 "错误"/"错"/"F"/"false"
	if quizType == "judge" {
		normalizeJudge := func(s string) string {
			s = strings.TrimSpace(s)
			switch strings.ToUpper(s) {
			case "正确", "对", "T", "TRUE", "√", "YES":
				return "正确"
			case "错误", "错", "F", "FALSE", "×", "NO":
				return "错误"
			}
			return s
		}
		return normalizeJudge(user) == normalizeJudge(correct)
	}

	// 对于选择题：用户可能发送字母（"B"），正确答案可能是完整文本（"B. xxx"）
	// 或者反过来
	correctUpper := strings.ToUpper(correct)
	userUpper := strings.ToUpper(user)

	// 如果正确答案是单字母（A/B/C/D），检查用户答案是否以该字母开头的选项
	if len(correctUpper) == 1 && correctUpper[0] >= 'A' && correctUpper[0] <= 'Z' {
		// 用户发送了选项文本，检查该选项是否以正确字母开头
		if strings.HasPrefix(userUpper, correctUpper+".") || strings.HasPrefix(userUpper, correctUpper+" ") {
			return true
		}
	}

	// 如果用户答案是单字母，检查正确答案是否以该字母开头
	if len(userUpper) == 1 && userUpper[0] >= 'A' && userUpper[0] <= 'Z' {
		if strings.HasPrefix(correctUpper, userUpper+".") || strings.HasPrefix(correctUpper, userUpper+" ") {
			return true
		}
		// 尝试从选项 JSON 中找到对应字母的选项文本
		var options []string
		if err := json.Unmarshal([]byte(optionsJSON), &options); err == nil {
			idx := int(userUpper[0] - 'A')
			if idx >= 0 && idx < len(options) {
				// 用户选的选项文本 == 正确答案？
				if strings.TrimSpace(options[idx]) == correct {
					return true
				}
			}
		}
	}

	return false
}

// GetQuizHint 获取题目渐进式提示
// GET /api/quizzes/:id/hint?level=1|2|3
func (h *Handler) GetQuizHint(c *gin.Context) {
	userID := c.GetString("userID")
	quizID := c.Param("id")

	// 解析 level 参数（默认 1）
	levelStr := c.DefaultQuery("level", "1")
	level := 1
	switch levelStr {
	case "2":
		level = 2
	case "3":
		level = 3
	}

	// 查找题目（限定当前用户的题目）
	var quiz model.Quiz
	if err := h.DB.Where("id = ? AND user_id = ?", quizID, userID).First(&quiz).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
		return
	}

	// 根据 level 返回对应级别的提示（以及所有更低级别的提示）
	resp := gin.H{
		"quiz_id": quizID,
		"level":   level,
	}

	switch level {
	case 1:
		resp["hint_1"] = quiz.Hint1
	case 2:
		resp["hint_1"] = quiz.Hint1
		resp["hint_2"] = quiz.Hint2
	case 3:
		resp["hint_1"] = quiz.Hint1
		resp["hint_2"] = quiz.Hint2
		resp["hint_3"] = quiz.Hint3
	}

	// 标记哪些级别有内容
	hasHints := gin.H{
		"1": quiz.Hint1 != "",
		"2": quiz.Hint2 != "",
		"3": quiz.Hint3 != "",
	}
	resp["available"] = hasHints

	c.JSON(http.StatusOK, resp)
}

// ==================== 对话（多轮 + Function Calling + RAG） ====================

// ChatRequest 对话请求
type ChatRequest struct {
	Message        string `json:"message" binding:"required,max=10000"`
	MaterialID     string `json:"material_id"`      // 可选：关联材料上下文
	ConversationID string `json:"conversation_id"`  // 可选：指定对话会话（不传则自动创建）
}

// Chat 多轮对话接口（支持 Function Calling + RAG 检索增强）
// POST /api/chat
func (h *Handler) Chat(c *gin.Context) {
	userID := c.GetString("userID")

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// ===== 0. 获取或创建对话会话 =====
	convID := req.ConversationID
	if convID == "" {
		// 自动创建新对话
		conv := model.Conversation{
			UserID: userID,
			Title:  truncate(req.Message, 50),
		}
		h.DB.Create(&conv)
		convID = conv.ID
	} else {
		// 验证对话属于当前用户
		var conv model.Conversation
		if err := h.DB.Where("id = ? AND user_id = ?", convID, userID).First(&conv).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "对话不存在"})
			return
		}
		// 更新对话的更新时间
		h.DB.Model(&conv).Update("updated_at", time.Now())
	}

	// 保存用户消息到数据库
	h.DB.Create(&model.ChatMessage{
		ConversationID: convID,
		Role:           "user",
		Content:        req.Message,
	})

	// ===== 1. 获取/创建用户的会话记忆 =====
	mem := h.getOrCreateMemory(convID)

	// 如果是新加载的对话（内存中没有历史），从数据库加载
	if len(mem.GetShortTermMemory()) == 0 {
		h.loadMemoryFromDB(convID, mem)
	}

	// ===== 2. RAG 语义检索：从向量库中检索相关材料 =====
	ragContext := ""
	if h.VectorStore != nil && req.Message != "" {
		searchResults, err := h.VectorStore.SemanticSearch(ctx, req.Message, userID, h.ragTopK)
		if err == nil && len(searchResults) > 0 {
			for _, sr := range searchResults {
				ragContext += sr.Content + "\n\n"
			}
		}
	}

	// 如果指定了材料 ID，也加入该材料的原始内容
	if req.MaterialID != "" {
		var material model.Material
		if err := h.DB.Select("title, content").Where("id = ? AND user_id = ?", req.MaterialID, userID).First(&material).Error; err == nil {
			ragContext += fmt.Sprintf("【指定材料「%s」】\n%s\n", material.Title, truncate(material.Content, 2000))
		}
	}

	// ===== 3. 构建三层上下文（长期摘要 + RAG + 短期对话） =====
	fullContext := mem.BuildContext(ragContext)

	// ===== 4. 将用户消息加入短期记忆 =====
	mem.AddMessage("user", req.Message)

	// ===== 5. 使用 Function Calling 进行对话 =====
	systemPrompt := `你是 StudyForge Pro 的 AI 学习助手。你可以帮助用户：
1. 解释知识点和学习概念
2. 回答关于学习材料的问题（使用 search_materials 工具搜索相关内容）
3. 生成练习题（使用 create_quiz 工具）和知识卡片（使用 generate_card 工具）
4. 查看学习统计（使用 get_user_stats 工具）
5. 制定学习计划和复习建议（使用 recommend_study_plan 工具）

请用中文回答，语言通俗易懂，像给同学讲题一样。
如果用户的问题和学习材料相关，优先使用 search_materials 工具搜索相关内容后再回答。
当用户要求出题、生成卡片时，调用对应的工具函数。`

	// 构建消息历史
	messages := buildChatMessages(systemPrompt, fullContext, req.Message)

	// 创建 ToolExecutor
	toolExec := agent.NewToolExecutor(h.DB, h.VectorStore, h.LLM, userID)
	tools := agent.ToolDefinitions()

	// Function Calling 循环（最多 3 轮工具调用）
	var finalReply string
	maxToolRounds := 3

	for i := 0; i < maxToolRounds; i++ {
		resp, err := h.LLM.ChatWithTools(ctx, messages, tools)
		if err != nil {
			// Function Calling 失败，降级到简单对话
			reply, fallbackErr := h.LLM.Chat(ctx, systemPrompt, fullContext+"\n用户："+req.Message)
			if fallbackErr != nil {
				log.Printf("Chat fallback 失败: %v", fallbackErr)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 服务暂时不可用，请稍后重试"})
				return
			}
			finalReply = reply
			break
		}

		if len(resp.Choices) == 0 {
			finalReply = "抱歉，AI 未能生成回复，请重试。"
			break
		}

		choice := resp.Choices[0]

		// 如果 LLM 直接返回文本（没有调用工具），就是最终回复
		if choice.FinishReason == "stop" || len(choice.Message.ToolCalls) == 0 {
			finalReply = choice.Message.Content
			break
		}

		// LLM 请求调用工具
		messages = append(messages, choice.Message) // 添加 assistant 的 tool_call 消息

		for _, toolCall := range choice.Message.ToolCalls {
			// 执行工具
			result, execErr := toolExec.ExecuteTool(ctx, toolCall.Function.Name, json.RawMessage(toolCall.Function.Arguments))
			if execErr != nil {
				result = fmt.Sprintf(`{"error": "%s"}`, execErr.Error())
			}

			// 添加工具结果到消息历史
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: toolCall.ID,
			})
		}
	}

	if finalReply == "" {
		finalReply = "处理完成，但未生成回复。请重试。"
	}

	// ===== 6. 将 AI 回复加入短期记忆 =====
	mem.AddMessage("assistant", finalReply)
	mem.IncrementRound()
	h.maybeSummarize(mem)

	// 保存 AI 回复到数据库
	h.DB.Create(&model.ChatMessage{
		ConversationID: convID,
		Role:           "assistant",
		Content:        finalReply,
	})
	// 更新对话的更新时间
	h.DB.Model(&model.Conversation{}).Where("id = ?", convID).Update("updated_at", time.Now())

	c.JSON(http.StatusOK, gin.H{
		"reply":           finalReply,
		"user_id":         userID,
		"conversation_id": convID,
	})
}

// ChatStream 流式对话接口（SSE 打字机效果 + Function Calling）
// GET /api/chat/stream?message=xxx&material_id=xxx&conversation_id=xxx
// 先通过非流式调用检测是否需要工具调用，若需要则执行工具后再流式输出最终回复
func (h *Handler) ChatStream(c *gin.Context) {
	userID := c.GetString("userID")
	message := c.Query("message")
	materialID := c.Query("material_id")
	conversationID := c.Query("conversation_id")

	if message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message 参数不能为空"})
		return
	}
	if len(message) > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "消息长度不能超过 10000 字符"})
		return
	}

	ctx := c.Request.Context()

	// ===== 0. 获取或创建对话会话 =====
	convID := conversationID
	if convID == "" {
		conv := model.Conversation{
			UserID: userID,
			Title:  truncate(message, 50),
		}
		h.DB.Create(&conv)
		convID = conv.ID
	} else {
		var conv model.Conversation
		if err := h.DB.Where("id = ? AND user_id = ?", convID, userID).First(&conv).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "对话不存在"})
			return
		}
		h.DB.Model(&conv).Update("updated_at", time.Now())
	}

	// 保存用户消息到数据库
	h.DB.Create(&model.ChatMessage{
		ConversationID: convID,
		Role:           "user",
		Content:        message,
	})

	// ===== 1. 构建上下文（与 Chat 相同逻辑）=====
	mem := h.getOrCreateMemory(convID)

	// 如果是新加载的对话（内存中没有历史），从数据库加载
	if len(mem.GetShortTermMemory()) == 0 {
		h.loadMemoryFromDB(convID, mem)
	}

	ragContext := ""
	if h.VectorStore != nil {
		searchResults, err := h.VectorStore.SemanticSearch(ctx, message, userID, h.ragTopK)
		if err == nil && len(searchResults) > 0 {
			for _, sr := range searchResults {
				ragContext += sr.Content + "\n\n"
			}
		}
	}

	if materialID != "" {
		var material model.Material
		if err := h.DB.Select("title, content").Where("id = ? AND user_id = ?", materialID, userID).First(&material).Error; err == nil {
			ragContext += fmt.Sprintf("【指定材料「%s」】\n%s\n", material.Title, truncate(material.Content, 2000))
		}
	}

	fullContext := mem.BuildContext(ragContext)
	mem.AddMessage("user", message)

	// ===== 2. 设置 SSE 响应头 =====
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器不支持流式响应"})
		return
	}

	// 发送 conversation_id 作为第一个 SSE 事件，让前端关联到正确的对话
	convIDData, _ := json.Marshal(map[string]string{"conversation_id": convID})
	fmt.Fprintf(c.Writer, "data: [CONV_ID]%s\n\n", string(convIDData))
	flusher.Flush()

	systemPrompt := `你是 StudyForge Pro 的 AI 学习助手。你可以帮助用户：
1. 解释知识点和学习概念
2. 回答关于学习材料的问题（使用 search_materials 工具搜索相关内容）
3. 生成练习题（使用 create_quiz 工具）和知识卡片（使用 generate_card 工具）
4. 查看学习统计（使用 get_user_stats 工具）
5. 制定学习计划和复习建议（使用 recommend_study_plan 工具）

请用中文回答，语言通俗易懂，像给同学讲题一样。
如果用户的问题和学习材料相关，优先使用 search_materials 工具搜索相关内容后再回答。
当用户要求出题、生成卡片时，调用对应的工具函数。`

	toolExec := agent.NewToolExecutor(h.DB, h.VectorStore, h.LLM, userID)
	tools := agent.ToolDefinitions()

	// ===== 3. 非流式调用检测是否需要工具调用 =====
	detectMessages := buildToolMessages(systemPrompt, fullContext, nil, message)
	detectResp, detectErr := h.LLM.ChatWithTools(ctx, detectMessages, tools)

	// 检测失败或无有效响应 → 降级到纯流式
	if detectErr != nil || detectResp == nil || len(detectResp.Choices) == 0 {
		log.Printf("ChatStream: 工具检测调用失败 (%v)，降级到纯流式", detectErr)
		h.streamDirectReply(c, ctx, systemPrompt, fullContext, message, mem, flusher, convID)
		return
	}

	choice := detectResp.Choices[0]

	// ===== 4. 无需工具调用 → 流式输出（重新发起流式请求以获得打字机效果）=====
	if choice.FinishReason == "stop" || len(choice.Message.ToolCalls) == 0 {
		h.streamDirectReply(c, ctx, systemPrompt, fullContext, message, mem, flusher, convID)
		return
	}

	// ===== 5. 需要工具调用 → 执行工具循环，然后流式输出最终回复 =====
	messages := detectMessages
	messages = append(messages, choice.Message) // 添加 assistant 的 tool_call 消息

	for round := 0; round < 3; round++ {
		// 执行本轮所有工具调用
		for _, toolCall := range choice.Message.ToolCalls {
			// 通知前端：工具开始执行
			tcData, _ := json.Marshal(map[string]string{
				"name": toolCall.Function.Name,
				"args": toolCall.Function.Arguments,
			})
			fmt.Fprintf(c.Writer, "data: [TOOL_CALL]%s\n\n", string(tcData))
			flusher.Flush()

			result, execErr := toolExec.ExecuteTool(ctx, toolCall.Function.Name, json.RawMessage(toolCall.Function.Arguments))
			if execErr != nil {
				result = fmt.Sprintf(`{"error": "%s"}`, execErr.Error())
			}

			// 通知前端：工具执行完成
			trData, _ := json.Marshal(map[string]string{
				"name":   toolCall.Function.Name,
				"result": truncate(result, 200),
			})
			fmt.Fprintf(c.Writer, "data: [TOOL_RESULT]%s\n\n", string(trData))
			flusher.Flush()

			// 添加工具结果到消息历史
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: toolCall.ID,
			})
		}

		// 再次调用 LLM，看是否需要更多工具调用或已得到最终回复
		resp, err := h.LLM.ChatWithTools(ctx, messages, tools)
		if err != nil || len(resp.Choices) == 0 {
			break
		}

		choice = resp.Choices[0]

		// 不再需要工具调用 → 跳出循环
		if choice.FinishReason == "stop" || len(choice.Message.ToolCalls) == 0 {
			break
		}

		// 继续下一轮工具调用
		messages = append(messages, choice.Message)
	}

	// ===== 6. 流式输出最终回复 =====
	// 基于完整工具对话历史发起流式请求，让 LLM 根据工具结果生成最终回复
	stream, streamErr := h.LLM.ChatStreamFromMessages(ctx, messages)
	if streamErr != nil {
		// 流式失败：使用工具循环中可能已获得的文本回复
		fallbackReply := choice.Message.Content
		if fallbackReply == "" {
			fallbackReply = "工具调用完成，但生成回复时出错。请重试。"
		}
		fmt.Fprintf(c.Writer, "data: %s\n\n", fallbackReply)
		fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
		flusher.Flush()
		mem.AddMessage("assistant", fallbackReply)
		mem.IncrementRound()
		h.maybeSummarize(mem)
		// 保存到数据库
		h.DB.Create(&model.ChatMessage{
			ConversationID: convID,
			Role:           "assistant",
			Content:        fallbackReply,
		})
		h.DB.Model(&model.Conversation{}).Where("id = ?", convID).Update("updated_at", time.Now())
		return
	}
	defer stream.Close()

	var fullReply strings.Builder
	for {
		response, err := stream.Recv()
		if err != nil {
			break
		}
		if len(response.Choices) == 0 {
			continue
		}
		delta := response.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		fullReply.WriteString(delta)
		fmt.Fprintf(c.Writer, "data: %s\n\n", delta)
		flusher.Flush()
	}

	fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
	flusher.Flush()

	finalReply := fullReply.String()
	if finalReply == "" {
		finalReply = choice.Message.Content
	}
	if finalReply == "" {
		finalReply = "处理完成，但未生成回复。请重试。"
	}

	mem.AddMessage("assistant", finalReply)
	mem.IncrementRound()
	h.maybeSummarize(mem)

	// 保存 AI 回复到数据库
	h.DB.Create(&model.ChatMessage{
		ConversationID: convID,
		Role:           "assistant",
		Content:        finalReply,
	})
	h.DB.Model(&model.Conversation{}).Where("id = ?", convID).Update("updated_at", time.Now())
}
// streamDirectReply 纯流式输出（无需工具调用时的快速路径）
func (h *Handler) streamDirectReply(c *gin.Context, ctx context.Context, systemPrompt, fullContext, userMessage string, mem *memory.ConversationMemory, flusher http.Flusher, convID string) {
	userMsg := userMessage
	if fullContext != "" {
		userMsg = fullContext + "\n用户：" + userMessage
	}

	stream, err := h.LLM.ChatStream(ctx, systemPrompt, userMsg)
	if err != nil {
		log.Printf("streamDirectReply 失败: %v", err)
		fmt.Fprintf(c.Writer, "data: AI 服务暂时不可用，请稍后重试\n\n")
		fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
		flusher.Flush()
		return
	}
	defer stream.Close()

	var fullReply strings.Builder
	for {
		response, err := stream.Recv()
		if err != nil {
			break
		}
		if len(response.Choices) == 0 {
			continue
		}
		delta := response.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		fullReply.WriteString(delta)
		fmt.Fprintf(c.Writer, "data: %s\n\n", delta)
		flusher.Flush()
	}

	fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
	flusher.Flush()
	mem.AddMessage("assistant", fullReply.String())
	mem.IncrementRound()
	h.maybeSummarize(mem)

	// 保存 AI 回复到数据库
	if convID != "" {
		h.DB.Create(&model.ChatMessage{
			ConversationID: convID,
			Role:           "assistant",
			Content:        fullReply.String(),
		})
		h.DB.Model(&model.Conversation{}).Where("id = ?", convID).Update("updated_at", time.Now())
	}
}

// buildChatMessages 构建 OpenAI 格式的消息列表
func buildChatMessages(systemPrompt, context, userMessage string) []openai.ChatCompletionMessage {
	systemContent := systemPrompt
	if context != "" {
		systemContent += "\n\n--- 上下文信息 ---\n" + context
	}

	return []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemContent,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userMessage,
		},
	}
}

// buildToolMessages 构建带工具调用的消息列表，支持上下文和历史消息
func buildToolMessages(systemPrompt, context string, history []openai.ChatCompletionMessage, userMessage string) []openai.ChatCompletionMessage {
	systemContent := systemPrompt
	if context != "" {
		systemContent += "\n\n--- 上下文信息 ---\n" + context
	}

	msgs := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemContent,
		},
	}
	msgs = append(msgs, history...)
	msgs = append(msgs, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userMessage,
	})
	return msgs
}

// truncate 截断字符串到指定长度（rune 感知，不会切断多字节 UTF-8 字符）
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// maybeSummarize 检查是否需要生成对话长期摘要（每 10 轮触发一次）
// 异步执行，不阻塞当前响应。摘要会合并已有长期摘要 + 当前短期对话缓冲。
func (h *Handler) maybeSummarize(mem *memory.ConversationMemory) {
	if !mem.NeedsSummary() {
		return
	}

	existingSummary := mem.GetLongTermSummary()
	bufferText := mem.GetBufferText()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		prompt := "请将以下对话历史压缩为一段简洁的中文摘要（200字以内）。保留关键信息点、用户的学习兴趣、已讨论的知识主题和重要结论。"
		if existingSummary != "" {
			prompt += "\n\n已有的历史摘要（请在此基础上合并新内容）：\n" + existingSummary
		}
		prompt += "\n\n最近对话记录：\n" + bufferText

		summary, err := h.LLM.Chat(ctx, prompt, "请生成压缩摘要，只输出摘要内容，不要任何前缀或解释。")
		if err != nil {
			log.Printf("maybeSummarize: 生成对话摘要失败: %v", err)
			return
		}

		mem.UpdateSummary(strings.TrimSpace(summary))
		log.Printf("maybeSummarize: 对话摘要已更新（约 %d 字）", len([]rune(summary)))
	}()
}

// ==================== 知识图谱 ====================

// GetKnowledgeGraph 获取材料的知识图谱数据
// GET /api/graph/:material_id
func (h *Handler) GetKnowledgeGraph(c *gin.Context) {
	userID := c.GetString("userID")
	materialID := c.Param("material_id")

	var material model.Material
	if err := h.DB.Select("graph_data").Where("id = ? AND user_id = ?", materialID, userID).First(&material).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "材料不存在"})
		return
	}

	if material.GraphData == "" {
		c.JSON(http.StatusOK, gin.H{
			"nodes": []interface{}{},
			"edges": []interface{}{},
			"message": "图谱数据尚未生成，请先分析材料",
		})
		return
	}

	// GraphData 是 JSON 字符串，解析后返回
	var graphData interface{}
	if err := json.Unmarshal([]byte(material.GraphData), &graphData); err != nil {
		c.JSON(http.StatusOK, gin.H{"raw": material.GraphData})
		return
	}

	c.JSON(http.StatusOK, graphData)
}

// GetAllKnowledgeGraphs 合并用户所有材料的知识图谱
// GET /api/graph/all
func (h *Handler) GetAllKnowledgeGraphs(c *gin.Context) {
	userID := c.GetString("userID")

	var materials []model.Material
	if err := h.DB.Select("id, title, graph_data").Where("user_id = ? AND status = ? AND graph_data != ''", userID, "completed").Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	if len(materials) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"nodes":     []interface{}{},
			"edges":     []interface{}{},
			"materials": []interface{}{},
			"message":   "暂无已分析的材料",
		})
		return
	}

	// 材料颜色调色板
	materialColors := []string{"#6366f1", "#ec4899", "#10b981", "#f59e0b", "#3b82f6", "#8b5cf6", "#ef4444", "#06b6d4", "#84cc16", "#f97316"}

	type graphRaw struct {
		Nodes []map[string]interface{} `json:"nodes"`
		Edges []map[string]interface{} `json:"edges"`
	}

	allNodes := make([]map[string]interface{}, 0)
	allEdges := make([]map[string]interface{}, 0)
	materialList := make([]map[string]interface{}, 0)
	nameSet := make(map[string]bool) // 去重

	for i, mat := range materials {
		color := materialColors[i%len(materialColors)]
		matPrefix := mat.Title
		if len(matPrefix) > 6 {
			matPrefix = matPrefix[:6]
		}

		materialList = append(materialList, map[string]interface{}{
			"id":    mat.ID,
			"title": mat.Title,
			"color": color,
		})

		var g graphRaw
		if err := json.Unmarshal([]byte(mat.GraphData), &g); err != nil {
			continue
		}

		// 节点：添加材料前缀和颜色
		nameMapping := make(map[string]string) // old name -> prefixed name
		for _, node := range g.Nodes {
			origName, _ := node["name"].(string)
			if origName == "" {
				continue
			}
			prefixedName := matPrefix + "·" + origName
			nameMapping[origName] = prefixedName

			if nameSet[prefixedName] {
				continue // 去重
			}
			nameSet[prefixedName] = true

			node["name"] = prefixedName
			node["material_id"] = mat.ID
			node["material_title"] = mat.Title
			node["material_color"] = color
			node["category"] = fmt.Sprintf("[%s] %v", matPrefix, node["category"])
			allNodes = append(allNodes, node)
		}

		// 边：映射节点名称
		for _, edge := range g.Edges {
			src, _ := edge["source"].(string)
			tgt, _ := edge["target"].(string)
			if mappedSrc, ok := nameMapping[src]; ok {
				edge["source"] = mappedSrc
			}
			if mappedTgt, ok := nameMapping[tgt]; ok {
				edge["target"] = mappedTgt
			}
			edge["material_id"] = mat.ID
			allEdges = append(allEdges, edge)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes":     allNodes,
		"edges":     allEdges,
		"materials": materialList,
	})
}

// ==================== 全局搜索 ====================

// SearchResultItem 单条搜索结果
type SearchResultItem struct {
	ID          string `json:"id"`
	Type        string `json:"type"`          // "material", "card", "quiz"
	Title       string `json:"title"`         // 主标题
	Subtitle    string `json:"subtitle"`      // 副标题/摘要
	MaterialID  string `json:"material_id"`   // 关联材料 ID（卡片和题目用）
	MaterialTitle string `json:"material_title,omitempty"` // 关联材料标题
}

// GlobalSearch 全局搜索（材料 + 卡片 + 练习题）
// GET /api/search?q=关键词
func (h *Handler) GlobalSearch(c *gin.Context) {
	userID := c.GetString("userID")
	q := strings.TrimSpace(c.Query("q"))

	if q == "" {
		c.JSON(http.StatusOK, gin.H{"total": 0, "query": "", "results": []interface{}{}})
		return
	}

	like := "%" + q + "%"
	var results []SearchResultItem

	// 搜索材料（title + content + tags）
	var materials []model.Material
	h.DB.Select("id, title, content_type, status, tags, created_at").
		Where("user_id = ? AND (title LIKE ? OR content LIKE ? OR tags LIKE ?)", userID, like, like, like).
		Order("created_at DESC").Limit(8).Find(&materials)
	for _, m := range materials {
		subtitle := "类型: " + m.ContentType + " | 状态: " + m.Status
		if m.Tags != "" {
			subtitle += " | 标签: " + m.Tags
		}
		results = append(results, SearchResultItem{
			ID: m.ID, Type: "material", Title: m.Title, Subtitle: subtitle, MaterialID: m.ID,
		})
	}

	// 搜索卡片（concept + detail + tags + memory_tip）
	var cards []model.Card
	h.DB.Where("user_id = ? AND (concept LIKE ? OR detail LIKE ? OR tags LIKE ? OR memory_tip LIKE ?)",
		userID, like, like, like, like).
		Order("created_at DESC").Limit(8).Find(&cards)

	// 批量查询卡片关联的材料标题
	materialTitles := map[string]string{}
	if len(cards) > 0 {
		var matIDs []string
		for _, card := range cards {
			matIDs = append(matIDs, card.MaterialID)
		}
		var mats []model.Material
		h.DB.Select("id, title").Where("id IN ?", matIDs).Find(&mats)
		for _, m := range mats {
			materialTitles[m.ID] = m.Title
		}
	}
	for _, card := range cards {
		subtitle := card.Detail
		if len([]rune(subtitle)) > 60 {
			subtitle = string([]rune(subtitle)[:60]) + "..."
		}
		if subtitle == "" {
			subtitle = card.Tags
		}
		results = append(results, SearchResultItem{
			ID: card.ID, Type: "card", Title: card.Concept, Subtitle: subtitle,
			MaterialID: card.MaterialID, MaterialTitle: materialTitles[card.MaterialID],
		})
	}

	// 搜索练习题（question）
	var quizzes []model.Quiz
	h.DB.Where("user_id = ? AND question LIKE ?", userID, like).
		Order("created_at DESC").Limit(8).Find(&quizzes)

	// 批量查询题目关联的材料标题
	if len(quizzes) > 0 {
		var matIDs []string
		for _, quiz := range quizzes {
			matIDs = append(matIDs, quiz.MaterialID)
		}
		for _, id := range matIDs {
			if _, exists := materialTitles[id]; !exists {
				var m model.Material
				if h.DB.Select("id, title").Where("id = ?", id).First(&m).Error == nil {
					materialTitles[id] = m.Title
				}
			}
		}
	}
	for _, quiz := range quizzes {
		subtitle := quiz.Type + " | " + quiz.Difficulty
		if quiz.Explanation != "" {
			exp := quiz.Explanation
			if len([]rune(exp)) > 40 {
				exp = string([]rune(exp)[:40]) + "..."
			}
			subtitle += " | " + exp
		}
		results = append(results, SearchResultItem{
			ID: quiz.ID, Type: "quiz", Title: quiz.Question, Subtitle: subtitle,
			MaterialID: quiz.MaterialID, MaterialTitle: materialTitles[quiz.MaterialID],
		})
	}

	// 截断题目主标题（可能很长）
	for i := range results {
		if results[i].Type == "quiz" && len([]rune(results[i].Title)) > 80 {
			results[i].Title = string([]rune(results[i].Title)[:80]) + "..."
		}
	}

	if results == nil {
		results = []SearchResultItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   len(results),
		"query":   q,
		"results": results,
	})
}

// ==================== 学习统计 ====================

// GetUserStats 获取用户学习统计
// GET /api/stats
func (h *Handler) GetUserStats(c *gin.Context) {
	userID := c.GetString("userID")

	var materialCount, cardCount, quizCount, correctCount, totalAttempts int64

	h.DB.Model(&model.Material{}).Where("user_id = ?", userID).Count(&materialCount)
	h.DB.Model(&model.Card{}).Where("user_id = ?", userID).Count(&cardCount)
	h.DB.Model(&model.Quiz{}).Where("user_id = ?", userID).Count(&quizCount)
	h.DB.Model(&model.QuizAttempt{}).Where("user_id = ?", userID).Count(&totalAttempts)
	h.DB.Model(&model.QuizAttempt{}).Where("user_id = ? AND is_correct = ?", userID, true).Count(&correctCount)

	accuracy := float64(0)
	if totalAttempts > 0 {
		accuracy = float64(correctCount) / float64(totalAttempts) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"material_count": materialCount,
		"card_count":     cardCount,
		"quiz_count":     quizCount,
		"total_attempts": totalAttempts,
		"correct_count":  correctCount,
		"accuracy":       accuracy,
	})
}
