package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ==================== 可观测性 Dashboard ====================

// GetMetrics 获取 Dashboard 聚合指标
// GET /api/dashboard/metrics
func (h *Handler) GetMetrics(c *gin.Context) {
	userID := c.GetString("userID")

	var totalCalls int64
	var totalTokens struct{ Total int64 }
	var avgLatency struct{ Avg float64 }
	var avgQuality struct{ Avg float64 }

	h.DB.Model(&model.LLMTrace{}).Where("user_id = ?", userID).Count(&totalCalls)

	h.DB.Model(&model.LLMTrace{}).Where("user_id = ?", userID).
		Select("COALESCE(SUM(total_tokens), 0) as total").
		Scan(&totalTokens)

	h.DB.Model(&model.LLMTrace{}).Where("user_id = ?", userID).
		Select("COALESCE(AVG(latency), 0) as avg").
		Scan(&avgLatency)
	avgLatencyMs := avgLatency.Avg / 1e6 // nanoseconds → milliseconds

	h.DB.Model(&model.LLMTrace{}).Where("user_id = ?", userID).
		Where("quality_score > 0").
		Select("COALESCE(AVG(quality_score), 0) as avg").
		Scan(&avgQuality)

	// 各 Agent 调用分布
	type AgentCount struct {
		AgentName string `json:"agent_name"`
		Count     int64  `json:"count"`
	}
	var agentBreakdown []AgentCount
	h.DB.Model(&model.LLMTrace{}).Where("user_id = ?", userID).
		Select("agent_name, COUNT(*) as count").
		Group("agent_name").
		Scan(&agentBreakdown)

	breakdownMap := make(map[string]int64)
	for _, ac := range agentBreakdown {
		breakdownMap[ac.AgentName] = ac.Count
	}

	// 各 Agent 平均质量分
	type AgentQuality struct {
		AgentName string  `json:"agent_name"`
		AvgScore  float64 `json:"avg_score"`
	}
	var agentQuality []AgentQuality
	h.DB.Model(&model.LLMTrace{}).Where("user_id = ?", userID).
		Where("quality_score > 0").
		Select("agent_name, AVG(quality_score) as avg_score").
		Group("agent_name").
		Scan(&agentQuality)

	qualityMap := make(map[string]float64)
	for _, aq := range agentQuality {
		qualityMap[aq.AgentName] = aq.AvgScore
	}

	// 各 Agent 平均延迟（latency 列存储为纳秒，转换为毫秒）
	type AgentLatency struct {
		AgentName string  `json:"agent_name"`
		AvgNs     float64 `json:"avg_ns"`
	}
	var agentLatency []AgentLatency
	h.DB.Model(&model.LLMTrace{}).Where("user_id = ?", userID).
		Select("agent_name, AVG(latency) as avg_ns").
		Group("agent_name").
		Scan(&agentLatency)

	latencyMap := make(map[string]float64)
	for _, al := range agentLatency {
		latencyMap[al.AgentName] = al.AvgNs / 1e6 // ns → ms
	}

	// 质量分布（按分数区间统计）
	type QualityBucket struct {
		Range string `json:"range"`
		Count int64  `json:"count"`
	}
	var qualityDist []QualityBucket
	h.DB.Model(&model.LLMTrace{}).Where("user_id = ? AND quality_score > 0", userID).
		Select(`
			CASE
				WHEN quality_score >= 8 THEN '8-10'
				WHEN quality_score >= 6 THEN '6-8'
				WHEN quality_score >= 3 THEN '3-6'
				ELSE '0-3'
			END as range,
			COUNT(*) as count
		`).
		Group("range").
		Order("range DESC").
		Scan(&qualityDist)

	distMap := make(map[string]int64)
	for _, qb := range qualityDist {
		distMap[qb.Range] = qb.Count
	}
	// 确保所有区间都存在
	for _, r := range []string{"0-3", "3-6", "6-8", "8-10"} {
		if _, ok := distMap[r]; !ok {
			distMap[r] = 0
		}
	}

	// 各 Agent 最近 Judge 评语（每个 Agent 取最新 3 条有评语的记录）
	type JudgeCommentRow struct {
		AgentName    string  `json:"agent_name"`
		QualityScore float64 `json:"quality_score"`
		JudgeComment string  `json:"judge_comment"`
		CreatedAt    string  `json:"created_at"`
	}
	var commentRows []JudgeCommentRow
	h.DB.Model(&model.LLMTrace{}).
		Where("user_id = ? AND judge_comment != '' AND quality_score > 0", userID).
		Select("agent_name, quality_score, judge_comment, created_at").
		Order("created_at DESC").
		Limit(20).
		Scan(&commentRows)

	// 按 Agent 分组
	commentsByAgent := make(map[string][]gin.H)
	for _, row := range commentRows {
		if len(commentsByAgent[row.AgentName]) < 3 {
			commentsByAgent[row.AgentName] = append(commentsByAgent[row.AgentName], gin.H{
				"score":      row.QualityScore,
				"comment":    row.JudgeComment,
				"created_at": row.CreatedAt,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_calls":          totalCalls,
		"avg_latency_ms":       avgLatencyMs,
		"total_tokens":         totalTokens.Total,
		"avg_quality_score":    avgQuality.Avg,
		"agent_breakdown":      breakdownMap,
		"agent_quality":        qualityMap,
		"agent_latency":        latencyMap,
		"quality_distribution": distMap,
		"judge_comments":       commentsByAgent,
	})
}

// GetDailyActivity 获取最近 30 天每日活动数据（用于学习日历热力图）
// GET /api/dashboard/activity
func (h *Handler) GetDailyActivity(c *gin.Context) {
	userID := c.GetString("userID")

	type DailyRow struct {
		Day   string `json:"day"`
		Count int64  `json:"count"`
	}

	var rows []DailyRow
	h.DB.Model(&model.LLMTrace{}).
		Where("user_id = ?", userID).
		Select("DATE(created_at) as day, COUNT(*) as count").
		Group("DATE(created_at)").
		Order("day DESC").
		Limit(30).
		Scan(&rows)

	// 补齐最近 30 天（没有数据的日期填 0）
	activityMap := make(map[string]int64)
	for _, r := range rows {
		activityMap[r.Day] = r.Count
	}

	now := time.Now()
	var result []gin.H
	for i := 29; i >= 0; i-- {
		day := now.AddDate(0, 0, -i).Format("2006-01-02")
		count := activityMap[day]
		result = append(result, gin.H{
			"date":  day,
			"count": count,
		})
	}

	c.JSON(http.StatusOK, result)
}

// GetCalendarHeatmap 获取全年学习日历热力图数据
// GET /api/stats/calendar?year=2026
func (h *Handler) GetCalendarHeatmap(c *gin.Context) {
	userID := c.GetString("userID")

	yearStr := c.DefaultQuery("year", time.Now().Format("2006"))
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2020 || year > 2100 {
		year = time.Now().Year()
	}

	startDate := fmt.Sprintf("%d-01-01", year)
	endDate := fmt.Sprintf("%d-12-31", year)

	type DailyRow struct {
		Day   string `json:"day"`
		Count int64  `json:"count"`
	}

	activityMap := make(map[string]int64)

	// 1. 聊天消息（通过 conversations 表关联 user_id）
	var chatRows []DailyRow
	h.DB.Raw(`
		SELECT DATE(cm.created_at) as day, COUNT(*) as count
		FROM chat_messages cm
		JOIN conversations c ON cm.conversation_id = c.id
		WHERE c.user_id = ? AND DATE(cm.created_at) BETWEEN ? AND ?
		GROUP BY DATE(cm.created_at)
	`, userID, startDate, endDate).Scan(&chatRows)
	for _, r := range chatRows {
		activityMap[r.Day] += r.Count
	}

	// 2. 练习题作答
	var quizRows []DailyRow
	h.DB.Model(&model.QuizAttempt{}).
		Where("user_id = ? AND DATE(created_at) BETWEEN ? AND ?", userID, startDate, endDate).
		Select("DATE(created_at) as day, COUNT(*) as count").
		Group("DATE(created_at)").
		Scan(&quizRows)
	for _, r := range quizRows {
		activityMap[r.Day] += r.Count
	}

	// 3. 卡片复习
	var cardRows []DailyRow
	h.DB.Model(&model.Card{}).
		Where("user_id = ? AND last_reviewed_at IS NOT NULL AND DATE(last_reviewed_at) BETWEEN ? AND ?", userID, startDate, endDate).
		Select("DATE(last_reviewed_at) as day, COUNT(*) as count").
		Group("DATE(last_reviewed_at)").
		Scan(&cardRows)
	for _, r := range cardRows {
		activityMap[r.Day] += r.Count
	}

	// 4. 材料上传
	var matRows []DailyRow
	h.DB.Model(&model.Material{}).
		Where("user_id = ? AND DATE(created_at) BETWEEN ? AND ?", userID, startDate, endDate).
		Select("DATE(created_at) as day, COUNT(*) as count").
		Group("DATE(created_at)").
		Scan(&matRows)
	for _, r := range matRows {
		activityMap[r.Day] += r.Count
	}

	// 生成全年 365/366 天数据
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.Local)

	var result []gin.H
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		day := d.Format("2006-01-02")
		count := activityMap[day]
		level := 0
		switch {
		case count >= 10:
			level = 4
		case count >= 6:
			level = 3
		case count >= 3:
			level = 2
		case count >= 1:
			level = 1
		}
		result = append(result, gin.H{
			"date":  day,
			"count": count,
			"level": level,
		})
	}

	c.JSON(http.StatusOK, result)
}

// ListTraces 获取最近的 LLM 调用追踪记录（支持分页）
// GET /api/dashboard/traces?limit=50&offset=0
func (h *Handler) ListTraces(c *gin.Context) {
	userID := c.GetString("userID")
	limit, offset := parsePagination(c)
	// traces 默认 50 条（parsePagination 默认 20）
	if c.Query("limit") == "" {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	var total int64
	h.DB.Model(&model.LLMTrace{}).Where("user_id = ?", userID).Count(&total)

	var traces []model.LLMTrace
	h.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&traces)

	for i := range traces {
		traces[i].LatencyMs = traces[i].Latency.Milliseconds()
	}

	c.JSON(http.StatusOK, gin.H{"data": traces, "total": total, "limit": limit, "offset": offset})
}

// parseLimit 简单辅助函数
func parseLimit(s string, n *int) (interface{}, error) {
	for _, c := range s {
		if c < '0' || c > '9' {
			return nil, gorm.ErrInvalidDB
		}
		*n = *n*10 + int(c-'0')
	}
	return nil, nil
}
