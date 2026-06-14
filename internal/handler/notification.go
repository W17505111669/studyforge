package handler

import (
	"log"
	"net/http"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 通知系统 ====================

// ListNotifications 获取用户通知列表（未读优先）
// GET /api/notifications
func (h *Handler) ListNotifications(c *gin.Context) {
	userID := c.GetString("userID")

	// 惰性触发：检查并生成待处理通知
	h.generatePendingNotifications(userID)

	var notifications []model.Notification
	if err := h.DB.Where("user_id = ?", userID).
		Order("CASE WHEN read_at IS NULL THEN 0 ELSE 1 END ASC, created_at DESC").
		Limit(50).
		Find(&notifications).Error; err != nil {
		log.Printf("查询通知失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取通知失败"})
		return
	}

	// 统计未读数
	var unreadCount int64
	h.DB.Model(&model.Notification{}).Where("user_id = ? AND read_at IS NULL", userID).Count(&unreadCount)

	c.JSON(http.StatusOK, gin.H{
		"data":          notifications,
		"unread_count":  unreadCount,
	})
}

// ReadNotification 标记单条通知为已读
// POST /api/notifications/:id/read
func (h *Handler) ReadNotification(c *gin.Context) {
	userID := c.GetString("userID")
	notifID := c.Param("id")

	now := time.Now()
	result := h.DB.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", notifID, userID).
		Update("read_at", &now)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "通知不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ReadAllNotifications 标记所有通知为已读
// POST /api/notifications/read-all
func (h *Handler) ReadAllNotifications(c *gin.Context) {
	userID := c.GetString("userID")

	now := time.Now()
	h.DB.Model(&model.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", &now)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetUnreadNotificationCount 获取未读通知计数（轻量端点，轮询用）
// GET /api/notifications/unread-count
func (h *Handler) GetUnreadNotificationCount(c *gin.Context) {
	userID := c.GetString("userID")

	var count int64
	h.DB.Model(&model.Notification{}).Where("user_id = ? AND read_at IS NULL", userID).Count(&count)

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// generatePendingNotifications 惰性检查并生成待处理通知
func (h *Handler) generatePendingNotifications(userID string) {
	// 1. 检查待复习卡片
	h.checkDueCardsNotification(userID)

	// 2. 检查最近完成的分析
	h.checkAnalysisCompleteNotification(userID)

	// 3. 检查最近解锁的成就
	h.checkAchievementNotification(userID)
}

// checkDueCardsNotification 检查是否有待复习卡片，生成复习提醒
func (h *Handler) checkDueCardsNotification(userID string) {
	var dueCount int64
	now := time.Now()
	h.DB.Model(&model.Card{}).
		Where("user_id = ? AND (next_review_at IS NULL OR next_review_at <= ?)", userID, now).
		Count(&dueCount)

	if dueCount <= 0 {
		return
	}

	// 检查今天是否已经发过复习提醒
	today := now.Format("2006-01-02")
	var existing int64
	h.DB.Model(&model.Notification{}).
		Where("user_id = ? AND type = ? AND DATE(created_at) = ?", userID, "review_reminder", today).
		Count(&existing)

	if existing > 0 {
		return
	}

	notif := model.Notification{
		UserID:    userID,
		Type:      "review_reminder",
		Title:     "复习提醒",
		Body:      "你还有卡片等待复习，及时复习能巩固记忆！",
		ActionURL: "/cards",
	}
	if err := h.DB.Create(&notif).Error; err != nil {
		log.Printf("创建复习提醒通知失败: %v", err)
	}
}

// checkAnalysisCompleteNotification 检查最近完成但未通知的分析
func (h *Handler) checkAnalysisCompleteNotification(userID string) {
	// 查找最近 1 小时内状态为 completed 的材料
	var materials []model.Material
	h.DB.Where("user_id = ? AND status = ? AND updated_at >= ?",
		userID, "completed", time.Now().Add(-1*time.Hour)).
		Find(&materials)

	for _, m := range materials {
		// 检查是否已为该材料生成过分析完成通知
		var existing int64
		h.DB.Model(&model.Notification{}).
			Where("user_id = ? AND type = ? AND action_url LIKE ?",
				userID, "analysis_complete", "%"+m.ID+"%").
			Count(&existing)

		if existing > 0 {
			continue
		}

		notif := model.Notification{
			UserID:    userID,
			Type:      "analysis_complete",
			Title:     "分析完成",
			Body:      "「" + m.Title + "」已分析完成，快去查看知识卡片和练习题吧！",
			ActionURL: "/materials/" + m.ID,
		}
		if err := h.DB.Create(&notif).Error; err != nil {
			log.Printf("创建分析完成通知失败: %v", err)
		}
	}
}

// checkAchievementNotification 检查最近解锁的成就
func (h *Handler) checkAchievementNotification(userID string) {
	// 查找最近 1 小时内解锁的成就
	var achievements []model.UserAchievement
	h.DB.Where("user_id = ? AND unlocked_at >= ?",
		userID, time.Now().Add(-1*time.Hour)).
		Find(&achievements)

	for _, a := range achievements {
		// 检查是否已为该成就生成过通知
		var existing int64
		h.DB.Model(&model.Notification{}).
			Where("user_id = ? AND type = ? AND body LIKE ?",
				userID, "achievement_unlocked", "%"+a.AchievementID+"%").
			Count(&existing)

		if existing > 0 {
			continue
		}

		// 查找成就定义获取名称
		achievementName := a.AchievementID
		for _, def := range model.AllAchievements {
			if def.ID == a.AchievementID {
				achievementName = def.Name
				break
			}
		}

		notif := model.Notification{
			UserID:    userID,
			Type:      "achievement_unlocked",
			Title:     "成就解锁！",
			Body:      "恭喜你解锁了成就「" + achievementName + "」！",
			ActionURL: "/",
		}
		if err := h.DB.Create(&notif).Error; err != nil {
			log.Printf("创建成就通知失败: %v", err)
		}
	}
}

// CreateNotification 创建通知的内部方法（供其他 handler 调用）
func (h *Handler) CreateNotification(userID, notifType, title, body, actionURL string) {
	notif := model.Notification{
		UserID:    userID,
		Type:      notifType,
		Title:     title,
		Body:      body,
		ActionURL: actionURL,
	}
	if err := h.DB.Create(&notif).Error; err != nil {
		log.Printf("创建通知失败: %v", err)
	}
}
