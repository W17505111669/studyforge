package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QuizMistake 错题记录
type QuizMistake struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	UserID       string    `json:"user_id" gorm:"index;not null"`
	QuizID       string    `json:"quiz_id" gorm:"index;not null"`
	UserAnswer   string    `json:"user_answer" gorm:"type:text"`
	CorrectAnswer string   `json:"correct_answer" gorm:"type:text"`
	MistakeAt    time.Time `json:"mistake_at"`
	Reviewed     bool      `json:"reviewed" gorm:"default:false"`

	// 关联（非数据库字段）
	Quiz *Quiz `json:"quiz,omitempty" gorm:"-"`
}

// BeforeCreate 创建前自动生成 UUID
func (qm *QuizMistake) BeforeCreate(tx *gorm.DB) error {
	if qm.ID == "" {
		qm.ID = uuid.New().String()
	}
	return nil
}
