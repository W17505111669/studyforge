package handler

import (
	"net/http"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 好友系统 ====================

// SendFriendRequest 发送好友请求
// POST /api/friends/request
func (h *Handler) SendFriendRequest(c *gin.Context) {
	userID := c.GetString("userID")

	var req model.SendFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入用户名"})
		return
	}

	// 查找目标用户
	var targetUser model.User
	if err := h.DB.Where("username = ?", req.Username).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 不能加自己为好友
	if targetUser.ID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能添加自己为好友"})
		return
	}

	// 检查是否已存在好友关系（双向检查）
	var existing model.Friendship
	err := h.DB.Where(
		"(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, targetUser.ID, targetUser.ID, userID,
	).First(&existing).Error
	if err == nil {
		if existing.Status == "accepted" {
			c.JSON(http.StatusConflict, gin.H{"error": "已经是好友了"})
		} else {
			c.JSON(http.StatusConflict, gin.H{"error": "已有待处理的好友请求"})
		}
		return
	}

	// 创建好友请求
	friendship := model.Friendship{
		UserID:   userID,
		FriendID: targetUser.ID,
		Status:   "pending",
	}
	if err := h.DB.Create(&friendship).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送请求失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "好友请求已发送",
		"friend": gin.H{
			"id":       targetUser.ID,
			"username": targetUser.Username,
			"nickname": targetUser.Nickname,
		},
	})
}

// AcceptFriendRequest 接受好友请求
// PUT /api/friends/request/:id/accept
func (h *Handler) AcceptFriendRequest(c *gin.Context) {
	userID := c.GetString("userID")
	friendshipID := c.Param("id")

	var friendship model.Friendship
	if err := h.DB.Where("id = ? AND friend_id = ? AND status = 'pending'",
		friendshipID, userID).First(&friendship).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "请求不存在或已处理"})
		return
	}

	// 更新状态为已接受
	h.DB.Model(&friendship).Update("status", "accepted")

	// 获取发起方信息
	var friend model.User
	h.DB.Where("id = ?", friendship.UserID).First(&friend)

	c.JSON(http.StatusOK, gin.H{
		"message": "已添加为好友",
		"friend": gin.H{
			"id":       friend.ID,
			"username": friend.Username,
			"nickname": friend.Nickname,
		},
	})
}

// RejectFriendRequest 拒绝好友请求
// DELETE /api/friends/request/:id
func (h *Handler) RejectFriendRequest(c *gin.Context) {
	userID := c.GetString("userID")
	friendshipID := c.Param("id")

	result := h.DB.Where("id = ? AND friend_id = ? AND status = 'pending'",
		friendshipID, userID).Delete(&model.Friendship{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "请求不存在或已处理"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已拒绝好友请求"})
}

// RemoveFriend 删除好友
// DELETE /api/friends/:id
func (h *Handler) RemoveFriend(c *gin.Context) {
	userID := c.GetString("userID")
	friendID := c.Param("id")

	// 双向匹配删除
	result := h.DB.Where(
		"(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, friendID, friendID, userID,
	).Delete(&model.Friendship{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "好友关系不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已删除好友"})
}

// ListFriends 获取好友列表
// GET /api/friends
func (h *Handler) ListFriends(c *gin.Context) {
	userID := c.GetString("userID")

	// 查询所有已接受的好友关系（双向）
	var friendships []model.Friendship
	h.DB.Where(
		"((user_id = ? OR friend_id = ?) AND status = 'accepted')",
		userID, userID,
	).Find(&friendships)

	if len(friendships) == 0 {
		c.JSON(http.StatusOK, gin.H{"friends": []interface{}{}, "total": 0})
		return
	}

	// 收集好友 ID
	friendIDs := make([]string, 0, len(friendships))
	friendshipMap := make(map[string]model.Friendship)
	for _, f := range friendships {
		fid := f.FriendID
		if f.FriendID == userID {
			fid = f.UserID
		}
		friendIDs = append(friendIDs, fid)
		friendshipMap[fid] = f
	}

	// 批量查询好友用户信息
	var users []model.User
	h.DB.Where("id IN ?", friendIDs).Select("id, username, nickname, updated_at").Find(&users)

	// 本周起始时间
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := time.Date(now.Year(), now.Month(), now.Day()-(weekday-1), 0, 0, 0, 0, time.Local)

	// 批量查询本周学习数据
	type userCount struct {
		UserID string
		Cnt    int64
	}

	// 本周复习卡片数
	weeklyCards := make(map[string]int64)
	var cardRows []userCount
	h.DB.Raw(`SELECT user_id, COUNT(*) as cnt FROM cards
		WHERE user_id IN ? AND last_reviewed_at >= ?
		GROUP BY user_id`, friendIDs, weekStart).Scan(&cardRows)
	for _, r := range cardRows {
		weeklyCards[r.UserID] = r.Cnt
	}

	// 本周做题数
	weeklyQuizzes := make(map[string]int64)
	var quizRows []userCount
	h.DB.Raw(`SELECT user_id, COUNT(*) as cnt FROM quiz_attempts
		WHERE user_id IN ? AND created_at >= ?
		GROUP BY user_id`, friendIDs, weekStart).Scan(&quizRows)
	for _, r := range quizRows {
		weeklyQuizzes[r.UserID] = r.Cnt
	}

	// 本周打卡天数
	weeklyStreaks := make(map[string]int64)
	var streakRows []userCount
	h.DB.Raw(`SELECT user_id, COUNT(DISTINCT DATE(streak_date)) as cnt FROM study_streaks
		WHERE user_id IN ? AND streak_date >= ? AND is_active = 1 AND streak_date != '0001-01-01'
		GROUP BY user_id`, friendIDs, weekStart).Scan(&streakRows)
	for _, r := range streakRows {
		weeklyStreaks[r.UserID] = r.Cnt
	}

	// 构建响应
	var friends []model.FriendInfo
	for _, u := range users {
		f := friendshipMap[u.ID]
		var lastActive *time.Time
		// 用 updated_at 近似最近活跃时间
		if !u.UpdatedAt.IsZero() {
			lastActive = &u.UpdatedAt
		}
		friends = append(friends, model.FriendInfo{
			ID:             f.ID,
			FriendID:       u.ID,
			FriendUsername:  u.Username,
			FriendNickname:  u.Nickname,
			Status:         "accepted",
			LastActiveAt:   lastActive,
			WeeklyCards:    weeklyCards[u.ID],
			WeeklyQuizzes:  weeklyQuizzes[u.ID],
			WeeklyStreak:   weeklyStreaks[u.ID],
			CreatedAt:      f.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"friends": friends,
		"total":   len(friends),
	})
}

// GetFriendRequests 获取待处理的好友请求
// GET /api/friends/requests
func (h *Handler) GetFriendRequests(c *gin.Context) {
	userID := c.GetString("userID")

	// 查询发给当前用户的待处理请求
	var friendships []model.Friendship
	h.DB.Where("friend_id = ? AND status = 'pending'", userID).Find(&friendships)

	if len(friendships) == 0 {
		c.JSON(http.StatusOK, gin.H{"requests": []interface{}{}, "total": 0})
		return
	}

	// 收集发起方 ID
	requesterIDs := make([]string, 0, len(friendships))
	requestMap := make(map[string]model.Friendship)
	for _, f := range friendships {
		requesterIDs = append(requesterIDs, f.UserID)
		requestMap[f.UserID] = f
	}

	// 批量查询发起方用户信息
	var users []model.User
	h.DB.Where("id IN ?", requesterIDs).Select("id, username, nickname").Find(&users)

	var requests []model.PendingRequest
	for _, u := range users {
		f := requestMap[u.ID]
		requests = append(requests, model.PendingRequest{
			ID:        f.ID,
			UserID:    u.ID,
			Username:  u.Username,
			Nickname:  u.Nickname,
			CreatedAt: f.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"requests": requests,
		"total":    len(requests),
	})
}

// SearchUsers 搜索用户（用于添加好友）
// GET /api/friends/search?q=xxx
func (h *Handler) SearchUsers(c *gin.Context) {
	userID := c.GetString("userID")
	q := c.Query("q")
	if len(q) < 1 {
		c.JSON(http.StatusOK, gin.H{"users": []interface{}{}})
		return
	}

	// 搜索用户名或昵称匹配的用户（排除自己）
	var users []model.User
	h.DB.Where(
		"id != ? AND (username LIKE ? OR nickname LIKE ?)",
		userID, "%"+q+"%", "%"+q+"%",
	).Select("id, username, nickname").Limit(10).Find(&users)

	// 检查与当前用户的好友关系状态
	type userResult struct {
		ID           string `json:"id"`
		Username     string `json:"username"`
		Nickname     string `json:"nickname"`
		FriendStatus string `json:"friend_status"` // none/pending/accepted
	}

	var results []userResult
	for _, u := range users {
		status := "none"
		var existing model.Friendship
		err := h.DB.Where(
			"((user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?))",
			userID, u.ID, u.ID, userID,
		).First(&existing).Error
		if err == nil {
			status = existing.Status
		}
		results = append(results, userResult{
			ID:           u.ID,
			Username:     u.Username,
			Nickname:     u.Nickname,
			FriendStatus: status,
		})
	}

	if results == nil {
		results = []userResult{}
	}

	c.JSON(http.StatusOK, gin.H{"users": results})
}

// GetFriendCount 获取好友数量（含待处理请求数）
// GET /api/friends/count
func (h *Handler) GetFriendCount(c *gin.Context) {
	userID := c.GetString("userID")

	var friendCount int64
	h.DB.Model(&model.Friendship{}).
		Where("((user_id = ? OR friend_id = ?) AND status = 'accepted')", userID, userID).
		Count(&friendCount)

	var pendingCount int64
	h.DB.Model(&model.Friendship{}).
		Where("friend_id = ? AND status = 'pending'", userID).
		Count(&pendingCount)

	c.JSON(http.StatusOK, gin.H{
		"friend_count":  friendCount,
		"pending_count": pendingCount,
	})
}
