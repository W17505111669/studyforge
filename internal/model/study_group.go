package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StudyGroup 学习小组模型
type StudyGroup struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"size:500"`
	OwnerID     string    `json:"owner_id" gorm:"index;not null"`
	MaxMembers  int       `json:"max_members" gorm:"default:20"`
	IsPublic    bool      `json:"is_public" gorm:"default:true"`
	MemberCount int       `json:"member_count" gorm:"default:1"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (g *StudyGroup) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = uuid.New().String()
	}
	return nil
}

// GroupMember 小组成员关联表
type GroupMember struct {
	ID      string    `json:"id" gorm:"primaryKey"`
	GroupID string    `json:"group_id" gorm:"index;not null"`
	UserID  string    `json:"user_id" gorm:"index;not null"`
	Role    string    `json:"role" gorm:"size:20;not null;default:'member'"` // owner/member
	JoinedAt time.Time `json:"joined_at"`
}

func (m *GroupMember) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

// GroupGoal 小组目标
type GroupGoal struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	GroupID      string    `json:"group_id" gorm:"index;not null"`
	Type         string    `json:"type" gorm:"size:50;not null"` // review_cards/complete_quizzes/study_minutes
	TargetValue  int       `json:"target_value" gorm:"not null"`
	CurrentValue int       `json:"current_value" gorm:"default:0"`
	Deadline     time.Time `json:"deadline"`
	CreatedAt    time.Time `json:"created_at"`
}

func (g *GroupGoal) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = uuid.New().String()
	}
	return nil
}

// ========== 请求/响应 DTO ==========

// CreateGroupRequest 创建小组请求体
type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	MaxMembers  int    `json:"max_members" binding:"min=2,max=100"`
	IsPublic    bool   `json:"is_public"`
}

// UpdateGroupRequest 更新小组请求体
type UpdateGroupRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=2,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	MaxMembers  *int    `json:"max_members" binding:"omitempty,min=2,max=100"`
	IsPublic    *bool   `json:"is_public"`
}

// CreateGroupGoalRequest 创建小组目标请求体
type CreateGroupGoalRequest struct {
	Type        string `json:"type" binding:"required,oneof=review_cards complete_quizzes study_minutes"`
	TargetValue int    `json:"target_value" binding:"required,min=1,max=10000"`
	Deadline    string `json:"deadline" binding:"required"` // ISO 8601 date
}

// GroupInfo 小组详情响应（含成员信息）
type GroupInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	MaxMembers  int       `json:"max_members"`
	IsPublic    bool      `json:"is_public"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
	IsOwner     bool      `json:"is_owner"`
	IsMember    bool      `json:"is_member"`
}

// GroupMemberInfo 成员信息响应
type GroupMemberInfo struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	Nickname    string    `json:"nickname"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
	WeeklyCards int64     `json:"weekly_cards"`
	WeeklyQuiz  int64     `json:"weekly_quizzes"`
}

// GroupGoalInfo 目标信息响应（含进度百分比）
type GroupGoalInfo struct {
	ID           string    `json:"id"`
	GroupID      string    `json:"group_id"`
	Type         string    `json:"type"`
	TypeLabel    string    `json:"type_label"`
	TargetValue  int       `json:"target_value"`
	CurrentValue int       `json:"current_value"`
	Percent      int       `json:"percent"`
	Deadline     time.Time `json:"deadline"`
	CreatedAt    time.Time `json:"created_at"`
}

// GroupGoalTypeLabels 目标类型中文标签
var GroupGoalTypeLabels = map[string]string{
	"review_cards":    "复习卡片",
	"complete_quizzes": "完成练习",
	"study_minutes":   "专注学习(分钟)",
}
