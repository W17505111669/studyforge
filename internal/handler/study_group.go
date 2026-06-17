package handler

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ========== 小组 CRUD ==========

// CreateGroup 创建学习小组
func (h *Handler) CreateGroup(c *gin.Context) {
	userID := c.GetString("userID")

	var req model.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写小组名称"})
		return
	}

	if req.MaxMembers == 0 {
		req.MaxMembers = 20
	}

	group := model.StudyGroup{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		OwnerID:     userID,
		MaxMembers:  req.MaxMembers,
		IsPublic:    req.IsPublic,
		MemberCount: 1,
	}

	if err := h.DB.Create(&group).Error; err != nil {
		log.Printf("创建学习小组失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建小组失败"})
		return
	}

	// 创建者自动加入为 owner
	member := model.GroupMember{
		GroupID:  group.ID,
		UserID:   userID,
		Role:     "owner",
		JoinedAt: time.Now(),
	}
	if err := h.DB.Create(&member).Error; err != nil {
		log.Printf("创建小组成员记录失败: %v", err)
	}

	c.JSON(http.StatusCreated, gin.H{"group": group, "message": "小组创建成功"})
}

// ListGroups 列出小组（自己加入的 + 公开可加入的）
func (h *Handler) ListGroups(c *gin.Context) {
	userID := c.GetString("userID")
	filter := c.DefaultQuery("filter", "all") // all / my / public

	var groups []model.StudyGroup

	switch filter {
	case "my":
		// 查询用户加入的小组
		var memberIDs []string
		h.DB.Model(&model.GroupMember{}).Where("user_id = ?", userID).Pluck("group_id", &memberIDs)
		if len(memberIDs) > 0 {
			h.DB.Where("id IN ?", memberIDs).Order("updated_at DESC").Find(&groups)
		}
	case "public":
		// 查询所有公开小组
		h.DB.Where("is_public = ?", true).Order("created_at DESC").Limit(50).Find(&groups)
	default:
		// 查询自己加入的 + 公开小组（去重）
		var memberIDs []string
		h.DB.Model(&model.GroupMember{}).Where("user_id = ?", userID).Pluck("group_id", &memberIDs)

		query := h.DB.Where("is_public = ?", true)
		if len(memberIDs) > 0 {
			query = h.DB.Where("is_public = ? OR id IN ?", true, memberIDs)
		}
		query.Order("updated_at DESC").Limit(50).Find(&groups)
	}

	// 批量查询用户在各小组的角色
	groupInfos := make([]model.GroupInfo, 0, len(groups))
	for _, g := range groups {
		info := model.GroupInfo{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			OwnerID:     g.OwnerID,
			MaxMembers:  g.MaxMembers,
			IsPublic:    g.IsPublic,
			MemberCount: g.MemberCount,
			CreatedAt:   g.CreatedAt,
			IsOwner:     g.OwnerID == userID,
		}
		// 检查是否是成员
		var count int64
		h.DB.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", g.ID, userID).Count(&count)
		info.IsMember = count > 0
		groupInfos = append(groupInfos, info)
	}

	c.JSON(http.StatusOK, gin.H{"groups": groupInfos})
}

// GetGroup 获取小组详情
func (h *Handler) GetGroup(c *gin.Context) {
	userID := c.GetString("userID")
	groupID := c.Param("id")

	var group model.StudyGroup
	if err := h.DB.Where("id = ?", groupID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "小组不存在"})
		return
	}

	// 检查是否是成员
	var memberCount int64
	h.DB.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, userID).Count(&memberCount)
	isMember := memberCount > 0

	info := model.GroupInfo{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		OwnerID:     group.OwnerID,
		MaxMembers:  group.MaxMembers,
		IsPublic:    group.IsPublic,
		MemberCount: group.MemberCount,
		CreatedAt:   group.CreatedAt,
		IsOwner:     group.OwnerID == userID,
		IsMember:    isMember,
	}

	c.JSON(http.StatusOK, gin.H{"group": info})
}

// UpdateGroup 更新小组信息（仅组长）
func (h *Handler) UpdateGroup(c *gin.Context) {
	userID := c.GetString("userID")
	groupID := c.Param("id")

	var group model.StudyGroup
	if err := h.DB.Where("id = ?", groupID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "小组不存在"})
		return
	}

	if group.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "仅组长可修改小组信息"})
		return
	}

	var req model.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		updates["description"] = strings.TrimSpace(*req.Description)
	}
	if req.MaxMembers != nil {
		updates["max_members"] = *req.MaxMembers
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}

	if len(updates) > 0 {
		h.DB.Model(&group).Updates(updates)
	}

	h.DB.Where("id = ?", groupID).First(&group)
	c.JSON(http.StatusOK, gin.H{"group": group, "message": "更新成功"})
}

// DeleteGroup 删除小组（仅组长）
func (h *Handler) DeleteGroup(c *gin.Context) {
	userID := c.GetString("userID")
	groupID := c.Param("id")

	var group model.StudyGroup
	if err := h.DB.Where("id = ?", groupID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "小组不存在"})
		return
	}

	if group.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "仅组长可删除小组"})
		return
	}

	// 事务：删除小组 + 成员 + 目标
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", groupID).Delete(&model.GroupGoal{}).Error; err != nil {
			return err
		}
		if err := tx.Where("group_id = ?", groupID).Delete(&model.GroupMember{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&group).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Printf("删除小组失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "小组已删除"})
}

// ========== 成员管理 ==========

// JoinGroup 加入小组
func (h *Handler) JoinGroup(c *gin.Context) {
	userID := c.GetString("userID")
	groupID := c.Param("id")

	var group model.StudyGroup
	if err := h.DB.Where("id = ?", groupID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "小组不存在"})
		return
	}

	// 检查是否已经是成员
	var existingCount int64
	h.DB.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, userID).Count(&existingCount)
	if existingCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "你已是该小组成员"})
		return
	}

	// 检查人数上限
	if group.MemberCount >= group.MaxMembers {
		c.JSON(http.StatusBadRequest, gin.H{"error": "小组已满"})
		return
	}

	// 加入小组
	member := model.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		Role:     "member",
		JoinedAt: time.Now(),
	}
	if err := h.DB.Create(&member).Error; err != nil {
		log.Printf("加入小组失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加入失败"})
		return
	}

	// 更新成员计数
	h.DB.Model(&model.StudyGroup{}).Where("id = ?", groupID).Update("member_count", gorm.Expr("member_count + 1"))

	c.JSON(http.StatusOK, gin.H{"message": "成功加入小组"})
}

// LeaveGroup 离开小组
func (h *Handler) LeaveGroup(c *gin.Context) {
	userID := c.GetString("userID")
	groupID := c.Param("id")

	var group model.StudyGroup
	if err := h.DB.Where("id = ?", groupID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "小组不存在"})
		return
	}

	// 组长不能离开自己的小组（需要先转让或删除）
	if group.OwnerID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "组长不能离开小组，请先删除小组或转让组长"})
		return
	}

	result := h.DB.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&model.GroupMember{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "你不是该小组成员"})
		return
	}

	// 更新成员计数
	h.DB.Model(&model.StudyGroup{}).Where("id = ?", groupID).Update("member_count", gorm.Expr("CASE WHEN member_count > 0 THEN member_count - 1 ELSE 0 END"))

	c.JSON(http.StatusOK, gin.H{"message": "已离开小组"})
}

// GetGroupMembers 获取小组成员列表（含本周学习数据）
func (h *Handler) GetGroupMembers(c *gin.Context) {
	groupID := c.Param("id")

	var members []model.GroupMember
	h.DB.Where("group_id = ?", groupID).Order("role DESC, joined_at ASC").Find(&members)

	// 获取本周起始时间（周一）
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := now.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)

	memberInfos := make([]model.GroupMemberInfo, 0, len(members))
	for _, m := range members {
		var user model.User
		h.DB.Where("id = ?", m.UserID).First(&user)

		info := model.GroupMemberInfo{
			UserID:   m.UserID,
			Username: user.Username,
			Nickname: user.Nickname,
			Role:     m.Role,
			JoinedAt: m.JoinedAt,
		}

		// 本周学习数据
		var weeklyCards int64
		h.DB.Model(&model.Card{}).Where("user_id = ? AND last_reviewed_at >= ?", m.UserID, weekStart).Count(&weeklyCards)
		info.WeeklyCards = weeklyCards

		var weeklyQuiz int64
		h.DB.Model(&model.QuizAttempt{}).Where("user_id = ? AND created_at >= ?", m.UserID, weekStart).Count(&weeklyQuiz)
		info.WeeklyQuiz = weeklyQuiz

		memberInfos = append(memberInfos, info)
	}

	c.JSON(http.StatusOK, gin.H{"members": memberInfos})
}

// ========== 小组目标 ==========

// CreateGroupGoal 创建小组目标（仅组长）
func (h *Handler) CreateGroupGoal(c *gin.Context) {
	userID := c.GetString("userID")
	groupID := c.Param("id")

	var group model.StudyGroup
	if err := h.DB.Where("id = ?", groupID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "小组不存在"})
		return
	}

	if group.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "仅组长可设定目标"})
		return
	}

	var req model.CreateGroupGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写完整的目标信息"})
		return
	}

	deadline, err := time.Parse("2006-01-02", req.Deadline)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "截止日期格式错误，请使用 YYYY-MM-DD"})
		return
	}

	goal := model.GroupGoal{
		GroupID:     groupID,
		Type:        req.Type,
		TargetValue: req.TargetValue,
		Deadline:    deadline,
	}

	if err := h.DB.Create(&goal).Error; err != nil {
		log.Printf("创建小组目标失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目标失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"goal": goal, "message": "目标已创建"})
}

// GetGroupGoals 获取小组目标（含进度）
func (h *Handler) GetGroupGoals(c *gin.Context) {
	groupID := c.Param("id")

	var goals []model.GroupGoal
	h.DB.Where("group_id = ?", groupID).Order("deadline ASC").Find(&goals)

	// 获取小组成员
	var memberIDs []string
	h.DB.Model(&model.GroupMember{}).Where("group_id = ?", groupID).Pluck("user_id", &memberIDs)

	goalInfos := make([]model.GroupGoalInfo, 0, len(goals))
	for _, g := range goals {
		info := model.GroupGoalInfo{
			ID:          g.ID,
			GroupID:     g.GroupID,
			Type:        g.Type,
			TypeLabel:   model.GroupGoalTypeLabels[g.Type],
			TargetValue: g.TargetValue,
			Deadline:    g.Deadline,
			CreatedAt:   g.CreatedAt,
		}

		// 计算当前进度（惰性聚合所有成员的贡献）
		if len(memberIDs) > 0 {
			info.CurrentValue = h.calcGroupGoalProgress(g.Type, memberIDs, g.CreatedAt)
		}

		if g.TargetValue > 0 {
			info.Percent = info.CurrentValue * 100 / g.TargetValue
			if info.Percent > 100 {
				info.Percent = 100
			}
		}

		goalInfos = append(goalInfos, info)
	}

	c.JSON(http.StatusOK, gin.H{"goals": goalInfos})
}

// DeleteGroupGoal 删除小组目标（仅组长）
func (h *Handler) DeleteGroupGoal(c *gin.Context) {
	userID := c.GetString("userID")
	groupID := c.Param("id")
	goalID := c.Param("goal_id")

	var group model.StudyGroup
	if err := h.DB.Where("id = ?", groupID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "小组不存在"})
		return
	}

	if group.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "仅组长可删除目标"})
		return
	}

	h.DB.Where("id = ? AND group_id = ?", goalID, groupID).Delete(&model.GroupGoal{})
	c.JSON(http.StatusOK, gin.H{"message": "目标已删除"})
}

// ========== 小组进度 ==========

// GetGroupProgress 获取小组目标进度概览
func (h *Handler) GetGroupProgress(c *gin.Context) {
	groupID := c.Param("id")

	// 获取小组成员
	var memberIDs []string
	h.DB.Model(&model.GroupMember{}).Where("group_id = ?", groupID).Pluck("user_id", &memberIDs)

	// 获取活跃目标
	var goals []model.GroupGoal
	h.DB.Where("group_id = ? AND deadline >= ?", groupID, time.Now()).Order("deadline ASC").Find(&goals)

	goalInfos := make([]model.GroupGoalInfo, 0, len(goals))
	for _, g := range goals {
		info := model.GroupGoalInfo{
			ID:          g.ID,
			GroupID:     g.GroupID,
			Type:        g.Type,
			TypeLabel:   model.GroupGoalTypeLabels[g.Type],
			TargetValue: g.TargetValue,
			Deadline:    g.Deadline,
			CreatedAt:   g.CreatedAt,
		}

		if len(memberIDs) > 0 {
			info.CurrentValue = h.calcGroupGoalProgress(g.Type, memberIDs, g.CreatedAt)
		}

		if g.TargetValue > 0 {
			info.Percent = info.CurrentValue * 100 / g.TargetValue
			if info.Percent > 100 {
				info.Percent = 100
			}
		}

		goalInfos = append(goalInfos, info)
	}

	// 成员贡献排行
	type MemberContribution struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Nickname string `json:"nickname"`
		Cards    int64  `json:"cards"`
		Quizzes  int64  `json:"quizzes"`
		Total    int64  `json:"total"`
	}

	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := now.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)

	contributions := make([]MemberContribution, 0, len(memberIDs))
	for _, uid := range memberIDs {
		var user model.User
		h.DB.Where("id = ?", uid).First(&user)

		var cards int64
		h.DB.Model(&model.Card{}).Where("user_id = ? AND last_reviewed_at >= ?", uid, weekStart).Count(&cards)

		var quizzes int64
		h.DB.Model(&model.QuizAttempt{}).Where("user_id = ? AND created_at >= ?", uid, weekStart).Count(&quizzes)

		contributions = append(contributions, MemberContribution{
			UserID:   uid,
			Username: user.Username,
			Nickname: user.Nickname,
			Cards:    cards,
			Quizzes:  quizzes,
			Total:    cards + quizzes,
		})
	}

	// 按总贡献降序排列
	sort.Slice(contributions, func(i, j int) bool {
		return contributions[i].Total > contributions[j].Total
	})

	c.JSON(http.StatusOK, gin.H{
		"goals":         goalInfos,
		"contributions": contributions,
	})
}

// ========== 辅助函数 ==========

// calcGroupGoalProgress 计算小组目标进度（聚合所有成员在目标创建后的活动量）
func (h *Handler) calcGroupGoalProgress(goalType string, memberIDs []string, since time.Time) int {
	var count int64

	switch goalType {
	case "review_cards":
		h.DB.Model(&model.Card{}).
			Where("user_id IN ? AND last_reviewed_at >= ?", memberIDs, since).
			Count(&count)
	case "complete_quizzes":
		h.DB.Model(&model.QuizAttempt{}).
			Where("user_id IN ? AND created_at >= ?", memberIDs, since).
			Count(&count)
	case "study_minutes":
		// 聚合番茄钟完成次数 * 平均时长
		var totalSeconds int64
		h.DB.Model(&model.PomodoroSession{}).
			Where("user_id IN ? AND created_at >= ? AND type = ? AND completed = ?", memberIDs, since, "work", true).
			Select("COALESCE(SUM(duration_seconds), 0)").
			Scan(&totalSeconds)
		count = totalSeconds / 60
	}

	return int(count)
}
