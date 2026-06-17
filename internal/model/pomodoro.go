package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PomodoroSession 番茄钟学习会话记录
type PomodoroSession struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	UserID           string    `json:"user_id" gorm:"index;not null"`
	StartedAt        time.Time `json:"started_at" gorm:"not null"`
	EndedAt          time.Time `json:"ended_at"`
	DurationSeconds  int       `json:"duration_seconds" gorm:"not null"` // 实际持续秒数
	PlannedMinutes   int       `json:"planned_minutes" gorm:"not null"`  // 计划时长（分钟）
	Type             string    `json:"type" gorm:"size:20;not null"`     // work, short_break, long_break
	Completed        bool      `json:"completed" gorm:"default:false"`   // 是否完整完成
	CreatedAt        time.Time `json:"created_at"`
}

// BeforeCreate 创建前自动生成 UUID
func (p *PomodoroSession) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
