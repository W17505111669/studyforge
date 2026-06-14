package handler

import (
	"fmt"
	"net/http"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== AI 学习建议 ====================

// Recommendation 单条学习建议
type Recommendation struct {
	Type        string `json:"type"`         // "review", "weak_card", "start_study", "weak_quiz", "daily_goal"
	Title       string `json:"title"`        // 建议标题
	Description string `json:"description"`  // 建议描述
	ActionURL   string `json:"action_url"`   // 跳转路由
	Priority    int    `json:"priority"`     // 1=最高 5=最低
	Icon        string `json:"icon"`         // 图标 emoji
	Badge       string `json:"badge"`        // 可选的角标数字（如待复习数量）
}

// GetRecommendations 获取 AI 学习建议（基于学习数据分析）
// GET /api/recommendations
func (h *Handler) GetRecommendations(c *gin.Context) {
	userID := c.GetString("userID")

	var recs []Recommendation

	// ===== 1. 间隔重复到期卡片（今日待复习） =====
	dueCount := h.countDueCards(userID)
	if dueCount > 0 {
		rec := Recommendation{
			Type:      "review",
			Title:     fmt.Sprintf("%d 张卡片待复习", dueCount),
			ActionURL: "/cards?due=true",
			Priority:  1,
			Icon:      "📖",
			Badge:     fmt.Sprintf("%d", dueCount),
		}
		if dueCount <= 5 {
			rec.Description = "今天有少量卡片需要复习，趁热打铁巩固记忆！"
		} else if dueCount <= 15 {
			rec.Description = "有不少卡片到期了，花 10 分钟快速过一遍吧。"
		} else {
			rec.Description = "积累了较多待复习卡片，建议分批复习，优先处理最旧的卡片。"
		}
		recs = append(recs, rec)
	}

	// ===== 2. 薄弱知识点（复习次数多但 ease_factor 低的卡片） =====
	weakCards := h.findWeakCards(userID)
	if len(weakCards) > 0 {
		rec := Recommendation{
			Type:        "weak_card",
			Title:       fmt.Sprintf("%d 个薄弱知识点", len(weakCards)),
			Description: fmt.Sprintf("「%s」等知识点复习较吃力，建议重点攻克。", weakCards[0].Concept),
			ActionURL:   "/cards?difficulty=hard",
			Priority:    2,
			Icon:        "🔥",
			Badge:       fmt.Sprintf("%d", len(weakCards)),
		}
		recs = append(recs, rec)
	}

	// ===== 3. 已分析但从未复习卡片的材料 =====
	unusedMaterials := h.findUnusedMaterials(userID)
	if len(unusedMaterials) > 0 {
		rec := Recommendation{
			Type:        "start_study",
			Title:       fmt.Sprintf("%d 份材料尚未学习", len(unusedMaterials)),
			Description: fmt.Sprintf("「%s」已分析完成，快去开始学习卡片吧！", unusedMaterials[0].Title),
			ActionURL:   fmt.Sprintf("/materials/%s", unusedMaterials[0].ID),
			Priority:    3,
			Icon:        "📚",
			Badge:       fmt.Sprintf("%d", len(unusedMaterials)),
		}
		recs = append(recs, rec)
	}

	// ===== 4. 薄弱题型（正确率低的题型） =====
	weakQuizType := h.findWeakQuizType(userID)
	if weakQuizType != nil {
		recs = append(recs, *weakQuizType)
	}

	// ===== 5. 每日学习鼓励（如果没有其他建议，或作为补充） =====
	if len(recs) < 3 {
		todayActive := h.checkTodayActivity(userID)
		if !todayActive {
			recs = append(recs, Recommendation{
				Type:        "daily_goal",
				Title:       "今天还没有学习哦",
				Description: "保持每日学习习惯，哪怕只复习 5 张卡片也好！",
				ActionURL:   "/study",
				Priority:    4,
				Icon:        "💪",
			})
		}
	}

	// 限制最多 5 条
	if len(recs) > 5 {
		recs = recs[:5]
	}

	// 空切片保护
	if recs == nil {
		recs = []Recommendation{}
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": recs,
		"count":           len(recs),
	})
}

// countDueCards 统计到期待复习卡片数量
func (h *Handler) countDueCards(userID string) int64 {
	now := time.Now()
	var count int64
	h.DB.Model(&model.Card{}).
		Where("user_id = ? AND (next_review_at IS NULL OR next_review_at <= ?)", userID, now).
		Count(&count)
	return count
}

// findWeakCards 查找薄弱知识点（复习次数 >= 2 但 ease_factor < 2.0 的卡片）
func (h *Handler) findWeakCards(userID string) []model.Card {
	var cards []model.Card
	h.DB.Where("user_id = ? AND review_count >= 2 AND ease_factor < 2.0", userID).
		Order("ease_factor ASC").
		Limit(5).
		Find(&cards)
	return cards
}

// findUnusedMaterials 查找已分析完成但卡片从未被复习过的材料（单条 SQL，无 N+1）
func (h *Handler) findUnusedMaterials(userID string) []model.Material {
	// 使用 LEFT JOIN 一次性找出有卡片但零复习的材料
	var materials []model.Material
	h.DB.Table("materials").
		Select("materials.*").
		Joins("INNER JOIN cards ON cards.material_id = materials.id").
		Where("materials.user_id = ? AND materials.status = ?", userID, "completed").
		Where("cards.user_id = ?", userID).
		Group("materials.id").
		Having("SUM(CASE WHEN cards.review_count > 0 THEN 1 ELSE 0 END) = 0").
		Limit(5).
		Find(&materials)
	return materials
}

// findWeakQuizType 查找正确率最低的题型
func (h *Handler) findWeakQuizType(userID string) *Recommendation {
	// 统计各题型的答题正确率
	type QuizTypeStats struct {
		Type        string
		Total       int64
		Correct     int64
	}

	// 先获取用户所有答题记录，按题目类型聚合
	var attempts []struct {
		QuizType  string
		IsCorrect bool
	}

	h.DB.Table("quiz_attempts").
		Select("quizzes.type as quiz_type, quiz_attempts.is_correct").
		Joins("JOIN quizzes ON quizzes.id = quiz_attempts.quiz_id").
		Where("quiz_attempts.user_id = ?", userID).
		Find(&attempts)

	if len(attempts) < 3 {
		return nil // 答题记录太少，不给建议
	}

	// 按题型聚合
	typeStats := make(map[string]*QuizTypeStats)
	for _, a := range attempts {
		if _, ok := typeStats[a.QuizType]; !ok {
			typeStats[a.QuizType] = &QuizTypeStats{Type: a.QuizType}
		}
		typeStats[a.QuizType].Total++
		if a.IsCorrect {
			typeStats[a.QuizType].Correct++
		}
	}

	// 找到正确率最低的题型（至少答过 2 题）
	var weakest *QuizTypeStats
	var lowestRate float64 = 1.0

	for _, stats := range typeStats {
		if stats.Total < 2 {
			continue
		}
		rate := float64(stats.Correct) / float64(stats.Total)
		if rate < lowestRate {
			lowestRate = rate
			s := *stats
			weakest = &s
		}
	}

	if weakest == nil || lowestRate >= 0.6 {
		return nil // 正确率都在 60% 以上，不需要建议
	}

	typeLabels := map[string]string{
		"choice":       "选择题",
		"fill":         "填空题",
		"short_answer": "简答题",
	}
	label := typeLabels[weakest.Type]
	if label == "" {
		label = weakest.Type
	}

	pct := int(lowestRate * 100)
	return &Recommendation{
		Type:        "weak_quiz",
		Title:       fmt.Sprintf("%s正确率偏低 (%d%%)", label, pct),
		Description: fmt.Sprintf("你在%s上的正确率仅为 %d%%，建议进行专项练习。", label, pct),
		ActionURL:   "/quiz",
		Priority:    3,
		Icon:        "🎯",
	}
}

// checkTodayActivity 检查今天是否有学习活动
func (h *Handler) checkTodayActivity(userID string) bool {
	today := time.Now().Format("2006-01-02")

	// 检查今天是否有答题记录
	var quizCount int64
	h.DB.Model(&model.QuizAttempt{}).
		Where("user_id = ? AND DATE(created_at) = ?", userID, today).
		Count(&quizCount)
	if quizCount > 0 {
		return true
	}

	// 检查今天是否有卡片复习
	var reviewCount int64
	h.DB.Model(&model.Card{}).
		Where("user_id = ? AND last_reviewed_at IS NOT NULL AND DATE(last_reviewed_at) = ?", userID, today).
		Count(&reviewCount)
	if reviewCount > 0 {
		return true
	}

	// 检查今天是否有对话
	var msgCount int64
	h.DB.Table("chat_messages").
		Joins("JOIN conversations ON conversations.id = chat_messages.conversation_id").
		Where("conversations.user_id = ? AND DATE(chat_messages.created_at) = ?", userID, today).
		Count(&msgCount)

	return msgCount > 0
}
