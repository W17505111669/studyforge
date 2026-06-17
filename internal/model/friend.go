package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Friendship 好友关系模型
// 双向关系：A→B 发送请求，status=pending；B 接受后 status=accepted
type Friendship struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"index;not null"`    // 发起方
	FriendID  string    `json:"friend_id" gorm:"index;not null"`  // 接收方
	Status    string    `json:"status" gorm:"size:20;not null;default:'pending'"` // pending/accepted
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (f *Friendship) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	return nil
}

// SendFriendRequest 发送好友请求的请求体
type SendFriendRequest struct {
	Username string `json:"username" binding:"required,min=2,max=50"`
}

// FriendInfo 好友信息响应
type FriendInfo struct {
	ID              string    `json:"id"`               // friendship ID
	FriendID        string    `json:"friend_id"`
	FriendUsername   string    `json:"friend_username"`
	FriendNickname   string    `json:"friend_nickname"`
	Status          string    `json:"status"`
	LastActiveAt    *time.Time `json:"last_active_at"`   // 最近活跃时间
	WeeklyCards     int64     `json:"weekly_cards"`      // 本周复习卡片数
	WeeklyQuizzes   int64     `json:"weekly_quizzes"`    // 本周做题数
	WeeklyStreak    int64     `json:"weekly_streak"`     // 本周打卡天数
	CreatedAt       time.Time `json:"created_at"`
}

// PendingRequest 待处理的好友请求
type PendingRequest struct {
	ID           string    `json:"id"`            // friendship ID
	UserID       string    `json:"user_id"`       // 发起方 ID
	Username     string    `json:"username"`
	Nickname     string    `json:"nickname"`
	CreatedAt    time.Time `json:"created_at"`
}
