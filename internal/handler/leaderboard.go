package handler

import (
	"net/http"
	"sort"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 学习排行榜 ====================
// 积分规则：
//   卡片复习:      1 分/次
//   练习题作答:    3 分/次
//   答对奖励:      1 分/题 (is_correct=true)
//   上传材料:      8 分/份
//   番茄钟专注:    5 分/次 (已完成 work 类型)
//   笔记创建:      2 分/篇

// GetLeaderboard 获取学习排行榜
// GET /api/leaderboard?period=week|month
func (h *Handler) GetLeaderboard(c *gin.Context) {
	userID := c.GetString("userID")
	period := c.DefaultQuery("period", "week")

	now := time.Now()
	var start time.Time

	if period == "month" {
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	} else {
		period = "week"
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start = time.Date(now.Year(), now.Month(), now.Day()-(weekday-1), 0, 0, 0, 0, time.Local)
	}
	end := now

	type userScore struct {
		UserID   string
		Username string
		Nickname string
		Score    int64
		Detail   map[string]int64
	}

	// 加载所有用户
	var users []model.User
	h.DB.Select("id, username, nickname").Find(&users)

	scores := make(map[string]*userScore)
	for _, u := range users {
		scores[u.ID] = &userScore{
			UserID:   u.ID,
			Username: u.Username,
			Nickname: u.Nickname,
			Detail:   make(map[string]int64),
		}
	}

	// 1) 卡片复习次数 (cards.last_reviewed_at 在时间范围内)
	var cardRows []struct {
		UserID string
		Cnt    int64
	}
	h.DB.Raw(`
		SELECT user_id, COUNT(*) as cnt FROM cards
		WHERE last_reviewed_at >= ? AND last_reviewed_at <= ?
		GROUP BY user_id
	`, start, end).Scan(&cardRows)
	for _, r := range cardRows {
		if s, ok := scores[r.UserID]; ok {
			s.Score += r.Cnt * 1
			s.Detail["cards_reviewed"] = r.Cnt
		}
	}

	// 2) 练习题作答次数
	var quizRows []struct {
		UserID string
		Cnt    int64
	}
	h.DB.Raw(`
		SELECT user_id, COUNT(*) as cnt FROM quiz_attempts
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY user_id
	`, start, end).Scan(&quizRows)
	for _, r := range quizRows {
		if s, ok := scores[r.UserID]; ok {
			s.Score += r.Cnt * 3
			s.Detail["quizzes_completed"] = r.Cnt
		}
	}

	// 3) 答对奖励 (is_correct = true)
	var correctRows []struct {
		UserID string
		Cnt    int64
	}
	h.DB.Raw(`
		SELECT user_id, COUNT(*) as cnt FROM quiz_attempts
		WHERE created_at >= ? AND created_at <= ? AND is_correct = 1
		GROUP BY user_id
	`, start, end).Scan(&correctRows)
	for _, r := range correctRows {
		if s, ok := scores[r.UserID]; ok {
			s.Score += r.Cnt * 1
			s.Detail["correct_answers"] = r.Cnt
		}
	}

	// 4) 上传材料数
	var materialRows []struct {
		UserID string
		Cnt    int64
	}
	h.DB.Raw(`
		SELECT user_id, COUNT(*) as cnt FROM materials
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY user_id
	`, start, end).Scan(&materialRows)
	for _, r := range materialRows {
		if s, ok := scores[r.UserID]; ok {
			s.Score += r.Cnt * 8
			s.Detail["materials_uploaded"] = r.Cnt
		}
	}

	// 5) 番茄钟专注次数 (已完成)
	var pomodoroRows []struct {
		UserID string
		Cnt    int64
	}
	h.DB.Raw(`
		SELECT user_id, COUNT(*) as cnt FROM pomodoro_sessions
		WHERE started_at >= ? AND started_at <= ? AND completed = 1 AND type = 'work'
		GROUP BY user_id
	`, start, end).Scan(&pomodoroRows)
	for _, r := range pomodoroRows {
		if s, ok := scores[r.UserID]; ok {
			s.Score += r.Cnt * 5
			s.Detail["pomodoros_completed"] = r.Cnt
		}
	}

	// 6) 笔记创建数
	var noteRows []struct {
		UserID string
		Cnt    int64
	}
	h.DB.Raw(`
		SELECT user_id, COUNT(*) as cnt FROM notes
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY user_id
	`, start, end).Scan(&noteRows)
	for _, r := range noteRows {
		if s, ok := scores[r.UserID]; ok {
			s.Score += r.Cnt * 2
			s.Detail["notes_created"] = r.Cnt
		}
	}

	// 转为数组，只保留 score > 0 的用户
	var ranked []*userScore
	for _, s := range scores {
		if s.Score > 0 {
			ranked = append(ranked, s)
		}
	}

	// 按积分降序排序，同分按 user_id 排序保持稳定
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].Score != ranked[j].Score {
			return ranked[i].Score > ranked[j].Score
		}
		return ranked[i].Username < ranked[j].Username
	})

	// 取前 50 名，构建响应
	limit := 50
	if len(ranked) < limit {
		limit = len(ranked)
	}

	type LeaderboardEntry struct {
		Rank              int   `json:"rank"`
		UserID            string `json:"user_id"`
		Username          string `json:"username"`
		Nickname          string `json:"nickname"`
		Score             int64  `json:"score"`
		CardsReviewed     int64  `json:"cards_reviewed"`
		QuizzesCompleted  int64  `json:"quizzes_completed"`
		CorrectAnswers    int64  `json:"correct_answers"`
		MaterialsUploaded int64  `json:"materials_uploaded"`
		PomodorosDone     int64  `json:"pomodoros_completed"`
		NotesCreated      int64  `json:"notes_created"`
	}

	var topUsers []LeaderboardEntry
	for i := 0; i < limit; i++ {
		r := ranked[i]
		topUsers = append(topUsers, LeaderboardEntry{
			Rank:              i + 1,
			UserID:            r.UserID,
			Username:          r.Username,
			Nickname:          r.Nickname,
			Score:             r.Score,
			CardsReviewed:     r.Detail["cards_reviewed"],
			QuizzesCompleted:  r.Detail["quizzes_completed"],
			CorrectAnswers:    r.Detail["correct_answers"],
			MaterialsUploaded: r.Detail["materials_uploaded"],
			PomodorosDone:     r.Detail["pomodoros_completed"],
			NotesCreated:      r.Detail["notes_created"],
		})
	}

	// 当前用户排名
	currentUserRank := 0
	var currentUserScore int64
	found := false
	for i, r := range ranked {
		if r.UserID == userID {
			currentUserRank = i + 1
			currentUserScore = r.Score
			found = true
			break
		}
	}

	// 如果当前用户不在有积分的用户列表中
	if !found {
		currentUserScore = calcUserScore(h, userID, start, end)
		// 计算排名：分数严格大于当前用户的人数 + 1
		currentUserRank = 1
		for _, r := range ranked {
			if r.Score > currentUserScore {
				currentUserRank++
			}
		}
	}

	// 下一名所需的分数
	nextRankScore := currentUserScore
	if currentUserRank > 1 && currentUserRank <= len(ranked) {
		nextRankScore = ranked[currentUserRank-2].Score
	} else if currentUserRank > len(ranked) && len(ranked) > 0 {
		nextRankScore = ranked[len(ranked)-1].Score
	}

	// 当前用户的详细数据
	currentDetail := calcUserDetail(h, userID, start, end)

	c.JSON(http.StatusOK, gin.H{
		"users":            topUsers,
		"total_users":      len(ranked),
		"current_user": gin.H{
			"rank":              currentUserRank,
			"score":             currentUserScore,
			"cards_reviewed":    currentDetail["cards_reviewed"],
			"quizzes_completed": currentDetail["quizzes_completed"],
			"correct_answers":   currentDetail["correct_answers"],
			"materials_uploaded": currentDetail["materials_uploaded"],
			"pomodoros_completed": currentDetail["pomodoros_completed"],
			"notes_created":     currentDetail["notes_created"],
		},
		"next_rank_score": nextRankScore,
		"period":          period,
		"start":           start.Format("2006-01-02"),
		"end":             end.Format("2006-01-02"),
	})
}

// GetMyLeaderboardStats 获取当前用户的积分详情
// GET /api/leaderboard/me?period=week|month
func (h *Handler) GetMyLeaderboardStats(c *gin.Context) {
	userID := c.GetString("userID")
	period := c.DefaultQuery("period", "week")

	now := time.Now()
	var start time.Time

	if period == "month" {
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	} else {
		period = "week"
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start = time.Date(now.Year(), now.Month(), now.Day()-(weekday-1), 0, 0, 0, 0, time.Local)
	}

	detail := calcUserDetail(h, userID, start, now)
	score := detail["cards_reviewed"]*1 + detail["quizzes_completed"]*3 +
		detail["correct_answers"]*1 + detail["materials_uploaded"]*8 +
		detail["pomodoros_completed"]*5 + detail["notes_created"]*2

	c.JSON(http.StatusOK, gin.H{
		"score":               score,
		"cards_reviewed":      detail["cards_reviewed"],
		"quizzes_completed":   detail["quizzes_completed"],
		"correct_answers":     detail["correct_answers"],
		"materials_uploaded":  detail["materials_uploaded"],
		"pomodoros_completed": detail["pomodoros_completed"],
		"notes_created":       detail["notes_created"],
		"period":              period,
		"start":               start.Format("2006-01-02"),
	})
}

// calcUserScore 计算指定用户在时间范围内的总积分
func calcUserScore(h *Handler, userID string, start, end time.Time) int64 {
	d := calcUserDetail(h, userID, start, end)
	return d["cards_reviewed"]*1 + d["quizzes_completed"]*3 +
		d["correct_answers"]*1 + d["materials_uploaded"]*8 +
		d["pomodoros_completed"]*5 + d["notes_created"]*2
}

// calcUserDetail 计算指定用户在时间范围内的各项积分明细
func calcUserDetail(h *Handler, userID string, start, end time.Time) map[string]int64 {
	detail := make(map[string]int64)

	var cardsReviewed int64
	h.DB.Model(&model.Card{}).
		Where("user_id = ? AND last_reviewed_at >= ? AND last_reviewed_at <= ?", userID, start, end).
		Count(&cardsReviewed)
	detail["cards_reviewed"] = cardsReviewed

	var quizzesCompleted int64
	h.DB.Table("quiz_attempts").
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, start, end).
		Count(&quizzesCompleted)
	detail["quizzes_completed"] = quizzesCompleted

	var correctAnswers int64
	h.DB.Table("quiz_attempts").
		Where("user_id = ? AND created_at >= ? AND created_at <= ? AND is_correct = 1",
			userID, start, end).
		Count(&correctAnswers)
	detail["correct_answers"] = correctAnswers

	var materialsUploaded int64
	h.DB.Model(&model.Material{}).
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, start, end).
		Count(&materialsUploaded)
	detail["materials_uploaded"] = materialsUploaded

	var pomodorosCompleted int64
	h.DB.Table("pomodoro_sessions").
		Where("user_id = ? AND started_at >= ? AND started_at <= ? AND completed = 1 AND type = 'work'",
			userID, start, end).
		Count(&pomodorosCompleted)
	detail["pomodoros_completed"] = pomodorosCompleted

	var notesCreated int64
	h.DB.Model(&model.Note{}).
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, start, end).
		Count(&notesCreated)
	detail["notes_created"] = notesCreated

	return detail
}
