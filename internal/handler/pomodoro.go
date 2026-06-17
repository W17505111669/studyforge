package handler

import (
	"net/http"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// StartPomodoro 开始一个番茄钟会话
func (h *Handler) StartPomodoro(c *gin.Context) {
	userID := c.GetString("userID")

	var req struct {
		Type           string `json:"type" binding:"required,oneof=work short_break long_break"`
		PlannedMinutes int    `json:"planned_minutes" binding:"required,min=1,max=120"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	session := model.PomodoroSession{
		UserID:           userID,
		StartedAt:        time.Now(),
		DurationSeconds:  0,
		PlannedMinutes:   req.PlannedMinutes,
		Type:             req.Type,
		Completed:        false,
	}

	if err := h.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建番茄钟失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":              session.ID,
		"started_at":      session.StartedAt,
		"planned_minutes": session.PlannedMinutes,
		"type":            session.Type,
	})
}

// EndPomodoro 结束番茄钟会话
func (h *Handler) EndPomodoro(c *gin.Context) {
	userID := c.GetString("userID")

	var req struct {
		SessionID string `json:"session_id" binding:"required"`
		Completed bool   `json:"completed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	var session model.PomodoroSession
	if err := h.DB.Where("id = ? AND user_id = ?", req.SessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "番茄钟会话不存在"})
		return
	}

	// 如果已结束，返回已有状态
	if !session.EndedAt.IsZero() {
		c.JSON(http.StatusOK, gin.H{
			"id":               session.ID,
			"completed":        session.Completed,
			"duration_seconds": session.DurationSeconds,
			"ended_at":         session.EndedAt,
		})
		return
	}

	now := time.Now()
	duration := int(now.Sub(session.StartedAt).Seconds())

	updates := map[string]interface{}{
		"ended_at":         now,
		"duration_seconds": duration,
		"completed":        req.Completed,
	}

	if err := h.DB.Model(&session).Where("id = ? AND user_id = ?", req.SessionID, userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新番茄钟失败"})
		return
	}

	// 异步更新学习目标进度（学习时长：按实际分钟数增加）
	if req.Completed && session.Type == "work" {
		minutes := duration / 60
		if minutes > 0 {
			go h.IncrementGoalProgress(userID, "study_minutes", minutes)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":               session.ID,
		"completed":        req.Completed,
		"duration_seconds": duration,
		"ended_at":         now,
	})
}

// GetPomodoroStats 获取番茄钟统计数据
func (h *Handler) GetPomodoroStats(c *gin.Context) {
	userID := c.GetString("userID")
	now := time.Now()

	// 今日开始时间
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// 本周一
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	weekStart := time.Date(now.Year(), now.Month(), now.Day()-int(weekday-time.Monday), 0, 0, 0, 0, now.Location())
	// 本月 1 日
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	type periodStats struct {
		TotalMinutes    int `json:"total_minutes"`
		CompletedCount  int `json:"completed_count"`
		SessionCount    int `json:"session_count"`
	}

	getStats := func(start time.Time) periodStats {
		var result periodStats

		// 工作类型统计
		type row struct {
			TotalSeconds     int64
			CompletedCount   int64
			SessionCount     int64
		}
		var r row

		h.DB.Model(&model.PomodoroSession{}).
			Select("COALESCE(SUM(duration_seconds), 0) as total_seconds, "+
				"COUNT(CASE WHEN completed = true THEN 1 END) as completed_count, "+
				"COUNT(*) as session_count").
			Where("user_id = ? AND type = 'work' AND ended_at IS NOT NULL AND ended_at >= ? AND completed = ?", userID, start, true).
			Scan(&r)

		result.TotalMinutes = int(r.TotalSeconds / 60)
		result.CompletedCount = int(r.CompletedCount)
		result.SessionCount = int(r.SessionCount)
		return result
	}

	today := getStats(todayStart)
	week := getStats(weekStart)
	month := getStats(monthStart)

	c.JSON(http.StatusOK, gin.H{
		"today": today,
		"week":  week,
		"month": month,
	})
}
