package handler

import (
	"fmt"
	"net/http"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 学习报告 ====================

// GetWeeklyReport 获取周学习报告
// GET /api/reports/weekly?date=2026-06-16
func (h *Handler) GetWeeklyReport(c *gin.Context) {
	userID := c.GetString("userID")

	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		date = time.Now()
	}

	// 计算本周一和上周一
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := time.Date(date.Year(), date.Month(), date.Day()-(weekday-1), 0, 0, 0, 0, time.Local)
	prevMonday := monday.AddDate(0, 0, -7)

	thisStart := monday
	thisEnd := monday.AddDate(0, 0, 7)
	prevStart := prevMonday
	prevEnd := monday

	thisStats := h.computeReportStats(thisStart, thisEnd, userID)
	prevStats := h.computeReportStats(prevStart, prevEnd, userID)
	change := computeChange(thisStats, prevStats)

	// 每日活动数据（用于柱状图）
	dailyActivity := h.getWeekDailyActivity(thisStart, userID)

	c.JSON(http.StatusOK, gin.H{
		"period": gin.H{
			"start": thisStart.Format("2006-01-02"),
			"end":   thisEnd.AddDate(0, 0, -1).Format("2006-01-02"),
		},
		"prev_period": gin.H{
			"start": prevStart.Format("2006-01-02"),
			"end":   prevEnd.AddDate(0, 0, -1).Format("2006-01-02"),
		},
		"this_week":       thisStats,
		"prev_week":       prevStats,
		"change":          change,
		"daily_activity":  dailyActivity,
	})
}

// GetMonthlyReport 获取月学习报告
// GET /api/reports/monthly?month=2026-06
func (h *Handler) GetMonthlyReport(c *gin.Context) {
	userID := c.GetString("userID")

	monthStr := c.DefaultQuery("month", time.Now().Format("2006-01"))
	month, err := time.ParseInLocation("2006-01", monthStr, time.Local)
	if err != nil {
		month = time.Now()
	}

	thisStart := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.Local)
	thisEnd := thisStart.AddDate(0, 1, 0)
	prevStart := thisStart.AddDate(0, -1, 0)
	prevEnd := thisStart

	thisStats := h.computeReportStats(thisStart, thisEnd, userID)
	prevStats := h.computeReportStats(prevStart, prevEnd, userID)
	change := computeChange(thisStats, prevStats)

	// 每日活动数据
	dailyActivity := h.getMonthDailyActivity(thisStart, userID)

	c.JSON(http.StatusOK, gin.H{
		"period": gin.H{
			"start": thisStart.Format("2006-01-02"),
			"end":   thisEnd.AddDate(0, 0, -1).Format("2006-01-02"),
		},
		"prev_period": gin.H{
			"start": prevStart.Format("2006-01-02"),
			"end":   prevEnd.AddDate(0, 0, -1).Format("2006-01-02"),
		},
		"this_month":      thisStats,
		"prev_month":      prevStats,
		"change":          change,
		"daily_activity":  dailyActivity,
	})
}

// reportStats 报告统计数据
type reportStats struct {
	CardsReviewed   int64   `json:"cards_reviewed"`
	MasteredCount   int64   `json:"mastered_count"`
	MasteryRate     float64 `json:"mastery_rate"`
	QuizzesDone     int64   `json:"quizzes_done"`
	AvgAccuracy     float64 `json:"avg_accuracy"`
	HintsUsed       int64   `json:"hints_used"`
	ChatMessages    int64   `json:"chat_messages"`
	NewMaterials    int64   `json:"new_materials"`
	ActiveDays      int     `json:"active_days"`
	LongestFocusMin int     `json:"longest_focus_min"`
}

// computeChange 计算两个周期之间的变化百分比
func computeChange(this, prev reportStats) map[string]gin.H {
	calc := func(thisVal, prevVal int64) gin.H {
		var pct float64
		if prevVal > 0 {
			pct = float64(thisVal-prevVal) / float64(prevVal) * 100
		} else if thisVal > 0 {
			pct = 100
		}
		dir := "same"
		if pct > 0 {
			dir = "up"
		} else if pct < 0 {
			dir = "down"
		}
		return gin.H{"value": thisVal, "prev_value": prevVal, "pct": pct, "direction": dir}
	}

	return map[string]gin.H{
		"cards_reviewed":   calc(this.CardsReviewed, prev.CardsReviewed),
		"quizzes_done":     calc(this.QuizzesDone, prev.QuizzesDone),
		"chat_messages":    calc(this.ChatMessages, prev.ChatMessages),
		"new_materials":    calc(this.NewMaterials, prev.NewMaterials),
		"active_days":      calc(int64(this.ActiveDays), int64(prev.ActiveDays)),
		"longest_focus":    calc(int64(this.LongestFocusMin), int64(prev.LongestFocusMin)),
	}
}

// computeReportStats 计算指定时间范围内的报告统计
func (h *Handler) computeReportStats(start, end time.Time, userID string) reportStats {
	startStr := start.Format("2006-01-02")
	endStr := end.Format("2006-01-02")
	var stats reportStats

	// 卡片复习数量 + 掌握率（按 last_reviewed_at 统计）
	type cardRow struct {
		Total   int64
		Mastered int64
	}
	var cr cardRow
	h.DB.Model(&model.Card{}).
		Where("user_id = ? AND last_reviewed_at IS NOT NULL AND DATE(last_reviewed_at) >= ? AND DATE(last_reviewed_at) < ?", userID, startStr, endStr).
		Select("COUNT(*) as total, COUNT(CASE WHEN ease_factor >= 2.5 AND review_count >= 3 THEN 1 END) as mastered").
		Scan(&cr)
	stats.CardsReviewed = cr.Total
	stats.MasteredCount = cr.Mastered
	if stats.CardsReviewed > 0 {
		stats.MasteryRate = float64(stats.MasteredCount) / float64(stats.CardsReviewed) * 100
	}

	// 练习题完成数 + 平均正确率 + 使用提示次数
	type quizRow struct {
		Total      int64
		Correct    int64
		HintsUsed  int64
	}
	var qr quizRow
	h.DB.Model(&model.QuizAttempt{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startStr, endStr).
		Select("COUNT(*) as total, COUNT(CASE WHEN is_correct = true THEN 1 END) as correct, COALESCE(SUM(hints_used), 0) as hints_used").
		Scan(&qr)
	stats.QuizzesDone = qr.Total
	stats.HintsUsed = qr.HintsUsed
	if stats.QuizzesDone > 0 {
		stats.AvgAccuracy = float64(qr.Correct) / float64(stats.QuizzesDone) * 100
	}

	// 对话消息数（通过 conversations 表 JOIN 关联 user_id）
	h.DB.Raw(`
		SELECT COUNT(*) FROM chat_messages cm
		JOIN conversations c ON cm.conversation_id = c.id
		WHERE c.user_id = ? AND cm.created_at >= ? AND cm.created_at < ?
	`, userID, startStr, endStr).Scan(&stats.ChatMessages)

	// 新上传材料数
	h.DB.Model(&model.Material{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startStr, endStr).
		Count(&stats.NewMaterials)

	// 学习活跃天数（合并 4 种活动源的 DATE，去重计数）
	activeDays := make(map[string]bool)
	type dateRow struct {
		Day string
	}
	var dates []dateRow

	// 聊天消息日
	h.DB.Raw(`
		SELECT DISTINCT DATE(cm.created_at) as day FROM chat_messages cm
		JOIN conversations c ON cm.conversation_id = c.id
		WHERE c.user_id = ? AND cm.created_at >= ? AND cm.created_at < ?
	`, userID, startStr, endStr).Scan(&dates)
	for _, d := range dates {
		activeDays[d.Day] = true
	}
	// 练习题作答日
	dates = nil
	h.DB.Model(&model.QuizAttempt{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startStr, endStr).
		Select("DISTINCT DATE(created_at) as day").
		Scan(&dates)
	for _, d := range dates {
		activeDays[d.Day] = true
	}
	// 卡片复习日
	dates = nil
	h.DB.Model(&model.Card{}).
		Where("user_id = ? AND last_reviewed_at IS NOT NULL AND last_reviewed_at >= ? AND last_reviewed_at < ?", userID, startStr, endStr).
		Select("DISTINCT DATE(last_reviewed_at) as day").
		Scan(&dates)
	for _, d := range dates {
		activeDays[d.Day] = true
	}
	// 材料上传日
	dates = nil
	h.DB.Model(&model.Material{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startStr, endStr).
		Select("DISTINCT DATE(created_at) as day").
		Scan(&dates)
	for _, d := range dates {
		activeDays[d.Day] = true
	}
	stats.ActiveDays = len(activeDays)

	// 最长专注时长（番茄钟最大单次 duration_seconds / 60）
	var longestSec int64
	h.DB.Model(&model.PomodoroSession{}).
		Where("user_id = ? AND type = 'work' AND completed = ? AND ended_at >= ? AND ended_at < ?", userID, true, start, end).
		Select("COALESCE(MAX(duration_seconds), 0)").
		Scan(&longestSec)
	stats.LongestFocusMin = int(longestSec / 60)

	return stats
}

// getWeekDailyActivity 获取一周内每天的活动计数（4 种来源聚合）
func (h *Handler) getWeekDailyActivity(monday time.Time, userID string) []gin.H {
	startStr := monday.Format("2006-01-02")
	endStr := monday.AddDate(0, 0, 7).Format("2006-01-02")
	return h.getDailyActivity(startStr, endStr, monday, 7, userID)
}

// getMonthDailyActivity 获取一月内每天的活动计数
func (h *Handler) getMonthDailyActivity(monthStart time.Time, userID string) []gin.H {
	days := daysInMonth(monthStart.Year(), monthStart.Month())
	startStr := monthStart.Format("2006-01-02")
	endStr := monthStart.AddDate(0, 1, 0).Format("2006-01-02")
	return h.getDailyActivity(startStr, endStr, monthStart, days, userID)
}

// getDailyActivity 通用每日活动查询
func (h *Handler) getDailyActivity(startStr, endStr string, start time.Time, days int, userID string) []gin.H {
	activityMap := make(map[string]int64)
	type dailyRow struct {
		Day   string
		Count int64
	}

	// 聊天消息
	var chatRows []dailyRow
	h.DB.Raw(`
		SELECT DATE(cm.created_at) as day, COUNT(*) as count FROM chat_messages cm
		JOIN conversations c ON cm.conversation_id = c.id
		WHERE c.user_id = ? AND cm.created_at >= ? AND cm.created_at < ?
		GROUP BY DATE(cm.created_at)
	`, userID, startStr, endStr).Scan(&chatRows)
	for _, r := range chatRows {
		activityMap[r.Day] += r.Count
	}

	// 练习题作答
	var quizRows []dailyRow
	h.DB.Model(&model.QuizAttempt{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startStr, endStr).
		Select("DATE(created_at) as day, COUNT(*) as count").
		Group("DATE(created_at)").
		Scan(&quizRows)
	for _, r := range quizRows {
		activityMap[r.Day] += r.Count
	}

	// 卡片复习
	var cardRows []dailyRow
	h.DB.Model(&model.Card{}).
		Where("user_id = ? AND last_reviewed_at IS NOT NULL AND last_reviewed_at >= ? AND last_reviewed_at < ?", userID, startStr, endStr).
		Select("DATE(last_reviewed_at) as day, COUNT(*) as count").
		Group("DATE(last_reviewed_at)").
		Scan(&cardRows)
	for _, r := range cardRows {
		activityMap[r.Day] += r.Count
	}

	// 材料上传
	var matRows []dailyRow
	h.DB.Model(&model.Material{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startStr, endStr).
		Select("DATE(created_at) as day, COUNT(*) as count").
		Group("DATE(created_at)").
		Scan(&matRows)
	for _, r := range matRows {
		activityMap[r.Day] += r.Count
	}

	// 生成每日数据
	dayNames := []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}
	var result []gin.H
	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i)
		day := d.Format("2006-01-02")
		label := day
		if days == 7 {
			label = dayNames[i]
		} else {
			label = fmt.Sprintf("%d日", d.Day())
		}
		result = append(result, gin.H{
			"date":  day,
			"label": label,
			"count": activityMap[day],
		})
	}

	return result
}

// daysInMonth 计算指定年月的天数
func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
