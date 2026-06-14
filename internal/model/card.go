package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Card 知识卡片模型
type Card struct {
	ID           string     `json:"id" gorm:"primaryKey"`
	UserID       string     `json:"user_id" gorm:"index;not null"`
	MaterialID   string     `json:"material_id" gorm:"index;not null"`
	Concept      string     `json:"concept" gorm:"size:200;not null"` // 核心概念（一句话）
	Detail       string     `json:"detail" gorm:"type:text"`          // 详细解释
	Formula      string     `json:"formula" gorm:"type:text"`         // 关键公式/代码（如有）
	MemoryTip    string     `json:"memory_tip" gorm:"type:text"`      // 记忆技巧
	Difficulty   string     `json:"difficulty" gorm:"size:10"`        // easy, medium, hard
	Tags         string     `json:"tags" gorm:"size:500"`             // 逗号分隔的标签
	ReviewCount  int        `json:"review_count" gorm:"default:0"`    // 累计复习次数
	NextReviewAt *time.Time `json:"next_review_at"`                   // 下次复习时间（nil=新卡片，未复习过）
	EaseFactor   float64    `json:"ease_factor" gorm:"default:2.5"`   // SM-2 难度因子（最低 1.3）
	IntervalDays int        `json:"interval_days" gorm:"default:0"`   // 当前复习间隔（天）
	LastReviewedAt *time.Time `json:"last_reviewed_at"`               // 上次复习时间
	IsBookmarked   bool       `json:"is_bookmarked" gorm:"default:false"` // 是否书签
	UserNote       string     `json:"user_note" gorm:"type:text"`         // 个人笔记
	CreatedAt      time.Time  `json:"created_at"`
}

// BeforeCreate 创建前自动生成 UUID
func (c *Card) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// ApplyReview 应用 SM-2 间隔重复算法更新卡片复习状态
// result: "mastered"（已掌握）或 "review"（需要再复习）
// 返回下次复习时间
func (c *Card) ApplyReview(result string) time.Time {
	now := time.Now()
	c.LastReviewedAt = &now
	c.ReviewCount++

	if result == "mastered" {
		// 已掌握：增加间隔，提高 ease factor
		if c.IntervalDays == 0 {
			c.IntervalDays = 1 // 首次：1 天
		} else if c.IntervalDays == 1 {
			c.IntervalDays = 3 // 第二次：3 天
		} else {
			c.IntervalDays = int(float64(c.IntervalDays) * c.EaseFactor)
		}
		c.EaseFactor += 0.1 // 掌握得好，难度因子上升
		if c.EaseFactor > 3.0 {
			c.EaseFactor = 3.0
		}
	} else {
		// 再复习：缩短间隔，降低 ease factor
		c.IntervalDays = 1 // 重置为 1 天
		c.EaseFactor -= 0.2
		if c.EaseFactor < 1.3 {
			c.EaseFactor = 1.3
		}
	}

	nextReview := now.AddDate(0, 0, c.IntervalDays)
	c.NextReviewAt = &nextReview
	return nextReview
}
