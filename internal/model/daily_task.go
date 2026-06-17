package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DailyTask 每日学习任务
type DailyTask struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	UserID        string    `json:"user_id" gorm:"index;not null"`
	TaskDate      string    `json:"task_date" gorm:"size:10;index;not null"` // 格式: 2006-01-02
	Type          string    `json:"type" gorm:"size:30;not null"`            // review_due_cards, complete_n_quizzes, study_n_minutes, read_material, upload_material
	TargetCount   int       `json:"target_count" gorm:"not null;default:1"`
	CompletedCount int      `json:"completed_count" gorm:"not null;default:0"`
	IsCompleted   bool      `json:"is_completed" gorm:"default:false"`
	CompletedAt   *time.Time `json:"completed_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// BeforeCreate 创建前自动生成 UUID
func (d *DailyTask) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

// TaskTypeLabels 任务类型中文标签
var TaskTypeLabels = map[string]string{
	"review_due_cards":    "复习到期卡片",
	"complete_n_quizzes":  "完成练习题",
	"study_n_minutes":     "学习时长(分钟)",
	"read_material":       "阅读材料",
	"upload_material":     "上传材料",
}

// TaskTypeIcons 任务类型图标（emoji）
var TaskTypeIcons = map[string]string{
	"review_due_cards":    "🃏",
	"complete_n_quizzes":  "✏️",
	"study_n_minutes":     "⏱️",
	"read_material":       "📖",
	"upload_material":     "📤",
}

// DefaultDailyTasks 默认每日任务模板
var DefaultDailyTasks = []struct {
	Type        string
	TargetCount int
}{
	{"review_due_cards", 0},   // 0 表示动态计算（到期卡片数）
	{"complete_n_quizzes", 5},
	{"study_n_minutes", 30},
}
