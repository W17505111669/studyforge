package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LearningGoal 学习目标
type LearningGoal struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	UserID       string    `json:"user_id" gorm:"index;not null"`
	Type         string    `json:"type" gorm:"size:30;not null"`          // review_cards, complete_quizzes, study_minutes, upload_materials
	TargetValue  int       `json:"target_value" gorm:"not null"`          // 目标值
	CurrentValue int       `json:"current_value" gorm:"not null;default:0"` // 当前进度
	Period       string    `json:"period" gorm:"size:10;not null"`        // weekly, monthly
	StartDate    time.Time `json:"start_date" gorm:"not null"`            // 周期开始日期
	EndDate      time.Time `json:"end_date" gorm:"not null"`              // 周期结束日期
	Status       string    `json:"status" gorm:"size:15;not null;default:active"` // active, completed, failed
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// BeforeCreate 创建前自动生成 UUID
func (g *LearningGoal) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = uuid.New().String()
	}
	return nil
}

// GoalTypeLabels 目标类型中文标签
var GoalTypeLabels = map[string]string{
	"review_cards":     "复习卡片",
	"complete_quizzes": "完成练习",
	"study_minutes":    "学习时长(分钟)",
	"upload_materials": "上传材料",
}

// ValidGoalTypes 合法的目标类型列表
var ValidGoalTypes = []string{"review_cards", "complete_quizzes", "study_minutes", "upload_materials"}

// ValidPeriods 合法的周期列表
var ValidPeriods = []string{"weekly", "monthly"}
