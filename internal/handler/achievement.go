package handler

import (
	"net/http"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 学习成就系统 ====================

// GetAchievements 获取用户的所有成就（含解锁状态和进度）
// GET /api/achievements
func (h *Handler) GetAchievements(c *gin.Context) {
	userID := c.GetString("userID")

	// 查询已解锁成就
	var unlocked []model.UserAchievement
	h.DB.Where("user_id = ?", userID).Find(&unlocked)
	unlockedMap := make(map[string]time.Time)
	for _, ua := range unlocked {
		unlockedMap[ua.AchievementID] = ua.UnlockedAt
	}

	// 一次性查询所有需要的统计数据
	stats := h.computeAchievementStats(userID)

	// 检查并解锁新成就
	newUnlocks := h.checkAndUnlock(userID, stats, unlockedMap)
	for k, v := range newUnlocks {
		unlockedMap[k] = v
	}

	// 构建响应
	var response []model.AchievementResponse
	for _, def := range model.AllAchievements {
		progress := stats[def.ID]
		if progress > def.Target {
			progress = def.Target
		}

		pct := float64(0)
		if def.Target > 0 {
			pct = float64(progress) / float64(def.Target) * 100
			if pct > 100 {
				pct = 100
			}
		}

		_, isUnlocked := unlockedMap[def.ID]
		var unlockedAt *time.Time
		if t, ok := unlockedMap[def.ID]; ok {
			unlockedAt = &t
		}

		response = append(response, model.AchievementResponse{
			ID:          def.ID,
			Name:        def.Name,
			Description: def.Description,
			Icon:        def.Icon,
			Category:    def.Category,
			Target:      def.Target,
			Unlocked:    isUnlocked,
			UnlockedAt:  unlockedAt,
			Progress:    progress,
			ProgressPct: pct,
		})
	}

	// 统计解锁数量
	unlockedCount := 0
	for _, r := range response {
		if r.Unlocked {
			unlockedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"achievements":    response,
		"total":           len(response),
		"unlocked_count":  unlockedCount,
		"categories":      model.AchievementCategories,
	})
}

// achievementStats 每个成就对应的当前进度值
type achievementStats = map[string]int

// computeAchievementStats 一次性计算所有成就所需的统计值
func (h *Handler) computeAchievementStats(userID string) achievementStats {
	stats := make(achievementStats)

	// === 批量查询基础计数 ===
	var materialCount, cardCount int64
	h.DB.Model(&model.Material{}).Where("user_id = ?", userID).Count(&materialCount)
	h.DB.Model(&model.Card{}).Where("user_id = ?", userID).Count(&cardCount)
	stats["first_upload"] = int(materialCount)
	stats["upload_5"] = int(materialCount)
	stats["upload_20"] = int(materialCount)
	stats["card_collector"] = int(cardCount)

	// === 答题统计 ===
	var totalAttempts, correctCount int64
	h.DB.Model(&model.QuizAttempt{}).Where("user_id = ?", userID).Count(&totalAttempts)
	h.DB.Model(&model.QuizAttempt{}).Where("user_id = ? AND is_correct = ?", userID, true).Count(&correctCount)
	stats["first_quiz"] = int(totalAttempts)
	stats["quiz_50"] = int(totalAttempts)
	stats["quiz_200"] = int(totalAttempts)

	// 正确率成就（至少 10 题）
	if totalAttempts >= 10 {
		stats["accuracy_80"] = int(float64(correctCount) / float64(totalAttempts) * 100)
	}

	// 满分通关检测：按 quiz_id 分组查找是否有连续全对的轮次
	// 简化：检查最近一次答题中是否有至少 3 题且全对（按创建时间相近的一批）
	var perfectRoundCount int64
	h.DB.Raw(`
		SELECT COUNT(*) FROM (
			SELECT qa.created_at, qa.quiz_id, qa.is_correct,
				ROW_NUMBER() OVER (ORDER BY qa.created_at) as rn
			FROM quiz_attempts qa
			WHERE qa.user_id = ?
		) ranked
		WHERE is_correct = 1
	`, userID).Scan(&perfectRoundCount)
	// 简化逻辑：如果用户有连续答对记录（3题以上），就算解锁
	// 更好的方式：按答题会话分组，但这里用简单方式
	if totalAttempts >= 3 && correctCount == totalAttempts {
		stats["perfect_round"] = 1
	} else {
		// 检查是否有至少 3 道连续答对的记录
		var attempts []model.QuizAttempt
		h.DB.Where("user_id = ?", userID).Order("created_at ASC").Find(&attempts)
		maxStreak := 0
		currentStreak := 0
		for _, a := range attempts {
			if a.IsCorrect {
				currentStreak++
				if currentStreak > maxStreak {
					maxStreak = currentStreak
				}
			} else {
				currentStreak = 0
			}
		}
		if maxStreak >= 3 {
			stats["perfect_round"] = 1
		}
	}

	// === 复习统计 ===
	var totalReviews struct{ Total int64 }
	h.DB.Model(&model.Card{}).Where("user_id = ?", userID).
		Select("COALESCE(SUM(review_count), 0) as total").
		Scan(&totalReviews)
	stats["first_review"] = int(totalReviews.Total)
	stats["review_30"] = int(totalReviews.Total)
	stats["review_100"] = int(totalReviews.Total)

	// === 对话统计 ===
	var chatMessages int64
	h.DB.Model(&model.ChatMessage{}).Where("conversation_id IN (?)",
		h.DB.Model(&model.Conversation{}).Select("id").Where("user_id = ?", userID),
	).Where("role = ?", "user").Count(&chatMessages)
	stats["first_chat"] = int(chatMessages)
	stats["chat_50"] = int(chatMessages)

	// === 图谱探险：是否有材料生成了图谱数据 ===
	var graphCount int64
	h.DB.Model(&model.Material{}).Where("user_id = ? AND graph_data != '' AND graph_data IS NOT NULL", userID).Count(&graphCount)
	if graphCount > 0 {
		stats["graph_view"] = 1
	}

	// === 导出卡片：通过 LLMTrace 或简单判断（有卡片即视为可能导出过）===
	// 简化：有卡片数据就认为解锁（因为导出是前端操作，没有独立日志）
	if cardCount > 0 {
		stats["card_export"] = 1
	}

	// === Agent 全种类调用 ===
	type AgentRow struct {
		AgentName string
	}
	var agentNames []AgentRow
	h.DB.Model(&model.LLMTrace{}).Where("user_id = ?", userID).
		Select("DISTINCT agent_name as agent_name").
		Scan(&agentNames)
	// 4 种核心 Agent：Analyst, CardMaker, QuizMaster, MapBuilder
	coreAgents := map[string]bool{
		"Analyst": false, "CardMaker": false, "QuizMaster": false, "MapBuilder": false,
	}
	for _, a := range agentNames {
		if _, ok := coreAgents[a.AgentName]; ok {
			coreAgents[a.AgentName] = true
		}
	}
	agentCount := 0
	for _, v := range coreAgents {
		if v {
			agentCount++
		}
	}
	stats["all_agents"] = agentCount

	// === 夜猫子：检查是否在 0-5 点有学习记录 ===
	var nightActivity int64
	h.DB.Model(&model.LLMTrace{}).Where("user_id = ?", userID).
		Where("CAST(strftime('%H', created_at) AS INTEGER) BETWEEN 0 AND 4").
		Count(&nightActivity)
	if nightActivity > 0 {
		stats["night_owl"] = 1
	}
	// 也检查答题记录的时间
	if stats["night_owl"] == 0 {
		var nightQuiz int64
		h.DB.Model(&model.QuizAttempt{}).Where("user_id = ?", userID).
			Where("CAST(strftime('%H', created_at) AS INTEGER) BETWEEN 0 AND 4").
			Count(&nightQuiz)
		if nightQuiz > 0 {
			stats["night_owl"] = 1
		}
	}

	return stats
}

// checkAndUnlock 检查并解锁新成就，返回新解锁的成就 map
func (h *Handler) checkAndUnlock(userID string, stats achievementStats, existingUnlocks map[string]time.Time) map[string]time.Time {
	newUnlocks := make(map[string]time.Time)
	now := time.Now()

	for _, def := range model.AllAchievements {
		// 已解锁的跳过
		if _, exists := existingUnlocks[def.ID]; exists {
			continue
		}

		progress := stats[def.ID]
		if progress >= def.Target {
			// 解锁成就
			ua := model.UserAchievement{
				UserID:        userID,
				AchievementID: def.ID,
				UnlockedAt:    now,
			}
			h.DB.Create(&ua)
			newUnlocks[def.ID] = now
		}
	}

	return newUnlocks
}
