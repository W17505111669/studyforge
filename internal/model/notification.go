package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Notification 用户通知
type Notification struct {
	ID        string     `json:"id" gorm:"primaryKey"`
	UserID    string     `json:"user_id" gorm:"index;not null"`
	Type      string     `json:"type" gorm:"size:50;not null"`    // review_reminder, analysis_complete, achievement_unlocked
	Title     string     `json:"title" gorm:"size:200;not null"`
	Body      string     `json:"body" gorm:"size:1000"`
	ActionURL string     `json:"action_url" gorm:"size:200"`       // 点击跳转的路由
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// BeforeCreate 创建前自动生成 UUID
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	return nil
}
