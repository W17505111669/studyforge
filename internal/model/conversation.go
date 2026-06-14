package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Conversation 对话会话模型
type Conversation struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"index;not null"`
	Title     string    `json:"title" gorm:"size:200"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Messages []ChatMessage `json:"messages,omitempty" gorm:"foreignKey:ConversationID"`
}

// BeforeCreate 创建前自动生成 UUID
func (c *Conversation) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// ChatMessage 对话消息模型
type ChatMessage struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	ConversationID string    `json:"conversation_id" gorm:"index;not null"`
	Role           string    `json:"role" gorm:"size:20;not null"` // "user" or "assistant"
	Content        string    `json:"content" gorm:"type:text;not null"`
	CreatedAt      time.Time `json:"created_at"`
}

// BeforeCreate 创建前自动生成 UUID
func (m *ChatMessage) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

// ConversationListItem 对话列表项（轻量，不含完整消息）
type ConversationListItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	MessageCount int      `json:"message_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
