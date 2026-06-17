package handler

import (
	"log"
	"math"
	"net/http"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 学习目标 CRUD ====================

// CreateGoal 创建学习目标
func (h *Handler) CreateGoal(c *gin.Context) {
	userID := c.GetString("userID")

	var req struct {
		Type        string `json:"type" binding:"required,oneof=review_cards complete_quizzes study_minutes upload_materials"`
		TargetValue int    `json:"target_value" binding:"required,min=1,max=99999"`
		Period      string `json:"period" binding:"required,oneof=weekly monthly"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 计算周期开始/结束日期
	now := time.Now()
	var startDate, endDate time.Time
	loc := now.Location()

	if req.Period == "weekly" {
		// 本周：周一 ~ 周日
		weekday := now.Weekday()
		if weekday == time.Sunday {
			weekday = 7
		}
		monday := now.AddDate(0, 0, -int(weekday-time.Monday))
		startDate = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, loc)
		endDate = startDate.AddDate(0, 0, 6)
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, loc)
	} else {
		// 本月：1 日 ~ 月末
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		endDate = startDate.AddDate(0, 1, -1)
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, loc)
	}

	goal := model.LearningGoal{
		UserID:       userID,
		Type:         req.Type,
		TargetValue:  req.TargetValue,
		CurrentValue: 0,
		Period:       req.Period,
		StartDate:    startDate,
		EndDate:      endDate,
		Status:       "active",
	}

	if err := h.DB.Create(&goal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目标失败"})
		return
	}

	c.JSON(http.StatusCreated, goal)
}

// ListGoals 获取学习目标列表
func (h *Handler) ListGoals(c *gin.Context) {
	userID := c.GetString("userID")
	status := c.DefaultQuery("status", "active")

	var goals []model.LearningGoal
	query := h.DB.Where("user_id = ?", userID)
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("created_at DESC").Find(&goals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取目标失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"goals": goals,
		"total": len(goals),
	})
}

// UpdateGoal 更新学习目标
func (h *Handler) UpdateGoal(c *gin.Context) {
	userID := c.GetString("userID")
	goalID := c.Param("id")

	var req struct {
		TargetValue *int    `json:"target_value"`
		Status      *string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	var goal model.LearningGoal
	if err := h.DB.Where("id = ? AND user_id = ?", goalID, userID).First(&goal).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "目标不存在"})
		return
	}

	updates := map[string]interface{}{}
	if req.TargetValue != nil && *req.TargetValue > 0 {
		updates["target_value"] = *req.TargetValue
	}
	if req.Status != nil {
		s := *req.Status
		if s == "active" || s == "completed" || s == "failed" {
			updates["status"] = s
		}
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有可更新的字段"})
		return
	}

	if err := h.DB.Model(&goal).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新目标失败"})
		return
	}

	h.DB.Where("id = ? AND user_id = ?", goalID, userID).First(&goal)
	c.JSON(http.StatusOK, goal)
}

// DeleteGoal 删除学习目标
func (h *Handler) DeleteGoal(c *gin.Context) {
	userID := c.GetString("userID")
	goalID := c.Param("id")

	result := h.DB.Where("id = ? AND user_id = ?", goalID, userID).Delete(&model.LearningGoal{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除目标失败"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "目标不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "目标已删除"})
}

// GetGoalProgress 获取所有活跃目标的进度
func (h *Handler) GetGoalProgress(c *gin.Context) {
	userID := c.GetString("userID")
	now := time.Now()

	var goals []model.LearningGoal
	if err := h.DB.Where("user_id = ? AND status = ?", userID, "active").
		Order("created_at ASC").Find(&goals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取目标进度失败"})
		return
	}

	// 过滤过期目标（EndDate 已过且状态仍为 active 的标记为 failed）
	var activeGoals []model.LearningGoal
	for i := range goals {
		g := &goals[i]
		if now.After(g.EndDate) {
			// 周期已过，检查是否完成
			if g.CurrentValue >= g.TargetValue {
				h.DB.Model(g).Update("status", "completed")
				g.Status = "completed"
			} else {
				h.DB.Model(g).Update("status", "failed")
				g.Status = "failed"
			}
			// 过期的不计入活跃目标
			continue
		}
		activeGoals = append(activeGoals, *g)
	}

	// 构建进度响应
	type GoalProgress struct {
		model.LearningGoal
		Percent       float64 `json:"percent"`        // 完成百分比
		RemainingDays int     `json:"remaining_days"` // 剩余天数
		TypeLabel     string  `json:"type_label"`     // 类型中文标签
	}

	progress := make([]GoalProgress, 0, len(activeGoals))
	for _, g := range activeGoals {
		pct := float64(0)
		if g.TargetValue > 0 {
			pct = math.Min(100, float64(g.CurrentValue)/float64(g.TargetValue)*100)
		}
		remaining := int(math.Ceil(g.EndDate.Sub(now).Hours() / 24))
		if remaining < 0 {
			remaining = 0
		}
		progress = append(progress, GoalProgress{
			LearningGoal:  g,
			Percent:       math.Round(pct*10) / 10, // 保留一位小数
			RemainingDays: remaining,
			TypeLabel:     model.GoalTypeLabels[g.Type],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"goals": progress,
		"total": len(progress),
	})
}

// ==================== 异步更新目标进度 ====================

// IncrementGoalProgress 异步增加目标进度（在关键操作后调用）
// goalType: review_cards / complete_quizzes / study_minutes / upload_materials
// increment: 增加值
func (h *Handler) IncrementGoalProgress(userID string, goalType string, increment int) {
	if increment <= 0 {
		return
	}
	now := time.Now()

	// 查找当前周期内、该类型的活跃目标
	var goals []model.LearningGoal
	h.DB.Where("user_id = ? AND type = ? AND status = ? AND start_date <= ? AND end_date >= ?",
		userID, goalType, "active", now, now).Find(&goals)

	for i := range goals {
		g := &goals[i]
		newVal := g.CurrentValue + increment
		updates := map[string]interface{}{
			"current_value": newVal,
		}
		if newVal >= g.TargetValue {
			updates["status"] = "completed"
		}
		if err := h.DB.Model(g).Updates(updates).Error; err != nil {
			log.Printf("更新目标进度失败: goal=%s type=%s err=%v", g.ID, goalType, err)
		}
	}
}
