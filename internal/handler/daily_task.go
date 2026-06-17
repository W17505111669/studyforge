package handler

import (
	"net/http"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ensureDailyTasks 惰性生成每日默认任务（每天首次访问时触发）
func (h *Handler) ensureDailyTasks(userID string, taskDate string) {
	// 检查今日是否已有任务
	var count int64
	h.DB.Model(&model.DailyTask{}).Where("user_id = ? AND task_date = ?", userID, taskDate).Count(&count)
	if count > 0 {
		return // 已有任务，无需生成
	}

	// 查询到期卡片数（用于 review_due_cards 任务的 target_count）
	now := time.Now()
	var dueCards int64
	h.DB.Model(&model.Card{}).Where(
		"user_id = ? AND (next_review_at IS NULL OR next_review_at <= ?)", userID, now,
	).Count(&dueCards)

	// 生成默认任务
	for _, def := range model.DefaultDailyTasks {
		target := def.TargetCount
		if def.Type == "review_due_cards" {
			target = int(dueCards)
			if target == 0 {
				target = 10 // 无到期卡片时默认 10 张
			}
		}

		task := model.DailyTask{
			UserID:       userID,
			TaskDate:     taskDate,
			Type:         def.Type,
			TargetCount:  target,
			CompletedCount: 0,
			IsCompleted:  false,
		}
		h.DB.Create(&task)
	}
}

// GetDailyTasks 获取指定日期的任务列表
func (h *Handler) GetDailyTasks(c *gin.Context) {
	userID := c.GetString("userID")

	// 默认今天，支持 ?date=2006-01-02
	taskDate := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	// 惰性生成默认任务
	h.ensureDailyTasks(userID, taskDate)

	// 惰性更新已完成计数（根据当天实际活动）
	h.updateDailyTaskProgress(userID, taskDate)

	var tasks []model.DailyTask
	h.DB.Where("user_id = ? AND task_date = ?", userID, taskDate).
		Order("created_at ASC").
		Find(&tasks)

	// 统计完成情况
	completedCount := 0
	for _, t := range tasks {
		if t.IsCompleted {
			completedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":           tasks,
		"total":           len(tasks),
		"completed_count": completedCount,
		"all_completed":   len(tasks) > 0 && completedCount == len(tasks),
		"task_date":       taskDate,
	})
}

// updateDailyTaskProgress 惰性更新每日任务的实际完成进度
func (h *Handler) updateDailyTaskProgress(userID string, taskDate string) {
	// 解析日期
	dateStart, err := time.Parse("2006-01-02", taskDate)
	if err != nil {
		return
	}
	dateEnd := dateStart.Add(24 * time.Hour)

	// 查询今天的活动统计
	// 1. 复习卡片数（last_reviewed_at 在今天范围内）
	var cardsReviewed int64
	h.DB.Model(&model.Card{}).Where(
		"user_id = ? AND last_reviewed_at >= ? AND last_reviewed_at < ?",
		userID, dateStart, dateEnd,
	).Count(&cardsReviewed)

	// 2. 完成练习题数
	var quizzesCompleted int64
	h.DB.Model(&model.QuizAttempt{}).Where(
		"user_id = ? AND created_at >= ? AND created_at < ?",
		userID, dateStart, dateEnd,
	).Count(&quizzesCompleted)

	// 3. 学习时长（分钟）- 从番茄钟统计
	var studySeconds int64
	h.DB.Model(&model.PomodoroSession{}).Where(
		"user_id = ? AND type = ? AND completed = ? AND started_at >= ? AND started_at < ?",
		userID, "work", true, dateStart, dateEnd,
	).Select("COALESCE(SUM(duration_seconds), 0)").Scan(&studySeconds)
	studyMinutes := int(studySeconds / 60)

	// 更新对应任务的 completed_count
	h.updateTaskField(userID, taskDate, "review_due_cards", int(cardsReviewed))
	h.updateTaskField(userID, taskDate, "complete_n_quizzes", int(quizzesCompleted))
	h.updateTaskField(userID, taskDate, "study_n_minutes", studyMinutes)
}

// updateTaskField 更新单个任务的完成计数和完成状态
func (h *Handler) updateTaskField(userID string, taskDate string, taskType string, completedCount int) {
	var task model.DailyTask
	result := h.DB.Where("user_id = ? AND task_date = ? AND type = ?", userID, taskDate, taskType).First(&task)
	if result.Error != nil {
		return
	}

	// 更新完成计数
	updates := map[string]interface{}{
		"completed_count": completedCount,
	}

	// 检查是否达到目标
	if completedCount >= task.TargetCount && task.TargetCount > 0 && !task.IsCompleted {
		now := time.Now()
		updates["is_completed"] = true
		updates["completed_at"] = &now
	} else if completedCount < task.TargetCount && task.IsCompleted {
		// 如果回退（如数据修正），重置完成状态
		updates["is_completed"] = false
		updates["completed_at"] = nil
	}

	h.DB.Model(&task).Updates(updates)
}

// ToggleDailyTask 切换任务完成状态
func (h *Handler) ToggleDailyTask(c *gin.Context) {
	userID := c.GetString("userID")
	taskID := c.Param("id")

	var task model.DailyTask
	result := h.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	if task.IsCompleted {
		// 取消完成
		h.DB.Model(&task).Updates(map[string]interface{}{
			"is_completed": false,
			"completed_at": nil,
		})
	} else {
		// 标记完成
		now := time.Now()
		h.DB.Model(&task).Updates(map[string]interface{}{
			"is_completed":   true,
			"completed_at":   &now,
			"completed_count": task.TargetCount, // 手动完成时直接设为目标值
		})
	}

	h.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task)
	c.JSON(http.StatusOK, task)
}

// UpdateDailyTask 自定义任务目标值
func (h *Handler) UpdateDailyTask(c *gin.Context) {
	userID := c.GetString("userID")
	taskID := c.Param("id")

	var req struct {
		TargetCount int `json:"target_count" binding:"required,min=1,max=9999"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	var task model.DailyTask
	result := h.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	updates := map[string]interface{}{
		"target_count": req.TargetCount,
	}

	// 如果新的目标值已经达到，自动标记完成
	if task.CompletedCount >= req.TargetCount && !task.IsCompleted {
		now := time.Now()
		updates["is_completed"] = true
		updates["completed_at"] = &now
	} else if task.CompletedCount < req.TargetCount && task.IsCompleted {
		updates["is_completed"] = false
		updates["completed_at"] = nil
	}

	h.DB.Model(&task).Updates(updates)
	h.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task)
	c.JSON(http.StatusOK, task)
}

// CreateDailyTask 创建自定义任务
func (h *Handler) CreateDailyTask(c *gin.Context) {
	userID := c.GetString("userID")

	var req struct {
		TaskDate    string `json:"task_date" binding:"required"`
		Type        string `json:"type" binding:"required,oneof=review_due_cards complete_n_quizzes study_n_minutes read_material upload_material"`
		TargetCount int    `json:"target_count" binding:"required,min=1,max=9999"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	task := model.DailyTask{
		UserID:        userID,
		TaskDate:      req.TaskDate,
		Type:          req.Type,
		TargetCount:   req.TargetCount,
		CompletedCount: 0,
		IsCompleted:   false,
	}

	if err := h.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// DeleteDailyTask 删除自定义任务
func (h *Handler) DeleteDailyTask(c *gin.Context) {
	userID := c.GetString("userID")
	taskID := c.Param("id")

	result := h.DB.Where("id = ? AND user_id = ?", taskID, userID).Delete(&model.DailyTask{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务已删除"})
}
