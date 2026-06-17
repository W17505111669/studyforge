package model

import (
	"time"
)

// StudyStreak 学习连续打卡记录
// streak_date 为零值时表示聚合统计行，非零时表示每日活动记录
type StudyStreak struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"index;not null"`
	StreakDate time.Time `json:"streak_date" gorm:"index"`

	// 每日活动记录（streak_date 非零时使用）
	PerDayCards    int `json:"per_day_cards" gorm:"not null;default:0"`
	PerDayQuizzes  int `json:"per_day_quizzes" gorm:"not null;default:0"`
	PerDayMessages int `json:"per_day_messages" gorm:"not null;default:0"`
	IsActive       bool `json:"is_active" gorm:"default:false"`

	// 聚合统计（streak_date 为零值时使用）
	CurrentStreak int `json:"current_streak" gorm:"not null;default:0"`
	LongestStreak int `json:"longest_streak" gorm:"not null;default:0"`
	TotalDays     int `json:"total_days" gorm:"not null;default:0"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StreakMilestones 打卡里程碑天数
var StreakMilestones = []int{7, 30, 100, 365}
