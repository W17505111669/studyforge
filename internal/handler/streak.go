package handler

import (
	"log"
	"net/http"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// getOrCreateStats 获取或创建用户的打卡统计行（streak_date 为零值）
func (h *Handler) getOrCreateStats(userID string) (*model.StudyStreak, error) {
	var stats model.StudyStreak
	zeroDate := time.Time{}
	err := h.DB.Where("user_id = ? AND streak_date = ?", userID, zeroDate).First(&stats).Error
	if err == gorm.ErrRecordNotFound {
		stats = model.StudyStreak{
			UserID:     userID,
			StreakDate: zeroDate,
		}
		if createErr := h.DB.Create(&stats).Error; createErr != nil {
			return nil, createErr
		}
		return &stats, nil
	}
	return &stats, err
}

// EnsureTodayStreak 惰性打卡：检查今日是否已打卡，如无则创建记录并更新连续天数
// 调用时机：用户任意关键 API 请求时异步触发
func (h *Handler) EnsureTodayStreak(userID string) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)

	// 1. 检查今天是否已有打卡记录
	var todayRecord model.StudyStreak
	err := h.DB.Where("user_id = ? AND streak_date = ?", userID, today).First(&todayRecord).Error
	if err == nil {
		// 今日已有记录，无需操作
		return
	}
	if err != gorm.ErrRecordNotFound {
		log.Printf("查询今日打卡记录失败: user=%s err=%v", userID, err)
		return
	}

	// 2. 检查昨天是否有活跃记录（决定是否延续连续天数）
	var yesterdayRecord model.StudyStreak
	h.DB.Where("user_id = ? AND streak_date = ? AND is_active = ?", userID, yesterday, true).First(&yesterdayRecord)

	// 3. 创建今日打卡记录
	todayRec := model.StudyStreak{
		UserID:     userID,
		StreakDate: today,
		IsActive:   true,
	}
	if createErr := h.DB.Create(&todayRec).Error; createErr != nil {
		log.Printf("创建今日打卡记录失败: user=%s err=%v", userID, createErr)
		return
	}

	// 4. 获取或创建聚合统计行
	stats, err := h.getOrCreateStats(userID)
	if err != nil {
		log.Printf("获取打卡统计失败: user=%s err=%v", userID, err)
		return
	}

	// 5. 计算新的连续天数
	var newStreak int
	if yesterdayRecord.ID != "" && yesterdayRecord.IsActive {
		newStreak = stats.CurrentStreak + 1
	} else {
		newStreak = 1
	}

	// 6. 更新聚合统计
	updates := map[string]interface{}{
		"current_streak": newStreak,
		"total_days":     stats.TotalDays + 1,
	}
	if newStreak > stats.LongestStreak {
		updates["longest_streak"] = newStreak
	}
	if err := h.DB.Model(stats).Updates(updates).Error; err != nil {
		log.Printf("更新打卡统计失败: user=%s err=%v", userID, err)
	}
}

// freezeStreak 惰性冻结连续天数：如果昨天有活动但今天还没有，将 current_streak 置零
func (h *Handler) freezeStreak(userID string) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)

	// 检查今天是否有活动
	var todayCount int64
	h.DB.Model(&model.StudyStreak{}).Where("user_id = ? AND streak_date = ? AND is_active = ?", userID, today, true).Count(&todayCount)
	if todayCount > 0 {
		return // 今天已有活动，不需要冻结
	}

	// 检查昨天是否有活动
	var yesterdayCount int64
	h.DB.Model(&model.StudyStreak{}).Where("user_id = ? AND streak_date = ? AND is_active = ?", userID, yesterday, true).Count(&yesterdayCount)
	if yesterdayCount > 0 {
		// 昨天有活动但今天没有 → 连续天数冻结
		h.DB.Model(&model.StudyStreak{}).Where("user_id = ? AND streak_date = ?", userID, time.Time{}).Update("current_streak", 0)
	}
}

// GetStreaks 获取打卡统计
func (h *Handler) GetStreaks(c *gin.Context) {
	userID := c.GetString("userID")

	// 惰性冻结：检查是否需要冻结连续天数
	h.freezeStreak(userID)

	// 惰性打卡：确保今日有记录（如果有活动的话，由其他 API 触发）
	// 这里只做读取，不主动创建

	// 获取聚合统计
	stats, err := h.getOrCreateStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取打卡统计失败"})
		return
	}

	// 获取最近 7 天的每日记录
	now := time.Now()
	weekAgo := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -6)
	var last7Days []model.StudyStreak
	h.DB.Where("user_id = ? AND streak_date >= ? AND streak_date > ?", userID, weekAgo, time.Time{}).
		Order("streak_date ASC").Find(&last7Days)

	// 构建 7 天完整数据（填充空白日期）
	sevenDaysMap := make(map[string]bool)
	for _, d := range last7Days {
		sevenDaysMap[d.StreakDate.Format("2006-01-02")] = d.IsActive
	}
	last7DaysData := make([]gin.H, 0, 7)
	for i := 6; i >= 0; i-- {
		d := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -i)
		dateStr := d.Format("2006-01-02")
		active, ok := sevenDaysMap[dateStr]
		last7DaysData = append(last7DaysData, gin.H{
			"date":   dateStr,
			"active": ok && active,
		})
	}

	// 构建里程碑数据
	milestones := make([]gin.H, 0, len(model.StreakMilestones))
	for _, m := range model.StreakMilestones {
		milestones = append(milestones, gin.H{
			"days":     m,
			"achieved": stats.LongestStreak >= m,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"current_streak":    stats.CurrentStreak,
		"longest_streak":    stats.LongestStreak,
		"total_days":        stats.TotalDays,
		"last_7_days":       last7DaysData,
		"streak_milestones": milestones,
	})
}

// IncrementStreakActivity 异步更新今日打卡活动计数
// 在 ReviewCard、AnswerQuiz、Chat 等关键操作中调用
func (h *Handler) IncrementStreakActivity(userID string, field string) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 查找今日记录
	var record model.StudyStreak
	err := h.DB.Where("user_id = ? AND streak_date = ?", userID, today).First(&record).Error
	if err != nil {
		// 今日无记录，先确保打卡
		h.EnsureTodayStreak(userID)
		// 重新查找
		if err2 := h.DB.Where("user_id = ? AND streak_date = ?", userID, today).First(&record).Error; err2 != nil {
			log.Printf("打卡后仍无法找到今日记录: user=%s err=%v", userID, err2)
			return
		}
	}

	// 增加对应字段计数
	if err := h.DB.Model(&record).Update(field, gorm.Expr(field+" + 1")).Error; err != nil {
		log.Printf("更新打卡活动计数失败: user=%s field=%s err=%v", userID, field, err)
	}
}
