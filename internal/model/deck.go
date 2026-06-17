package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Deck 卡片组（牌组）模型
type Deck struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	UserID      string    `json:"user_id" gorm:"index;not null"`
	Name        string    `json:"name" gorm:"size:200;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Tags        string    `json:"tags" gorm:"size:500"`
	CardCount   int       `json:"card_count" gorm:"default:0"`
	CollectCount int      `json:"collect_count" gorm:"default:0"`
	IsPublic    bool      `json:"is_public" gorm:"default:false"`
	ShareCode   string    `json:"share_code" gorm:"size:32;index;default:''"` // 非空时唯一（手动部分索引）
	CreatedAt   time.Time `json:"created_at"`
}

// BeforeCreate 创建前自动生成 UUID
func (d *Deck) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

// DeckCard 牌组-卡片关联表（多对多）
type DeckCard struct {
	ID       string `json:"id" gorm:"primaryKey"`
	DeckID   string `json:"deck_id" gorm:"index;not null"`
	CardID   string `json:"card_id" gorm:"index;not null"`
}

// BeforeCreate 创建前自动生成 UUID
func (dc *DeckCard) BeforeCreate(tx *gorm.DB) error {
	if dc.ID == "" {
		dc.ID = uuid.New().String()
	}
	return nil
}
