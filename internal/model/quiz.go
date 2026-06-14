package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Quiz 练习题模型
type Quiz struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	UserID     string    `json:"user_id" gorm:"index;not null"`
	MaterialID string    `json:"material_id" gorm:"index;not null"`
	Question   string    `json:"question" gorm:"type:text;not null"`
	Type       string    `json:"type" gorm:"size:20"` // choice, fill, short_answer
	Difficulty string    `json:"difficulty" gorm:"size:10"` // easy, medium, hard
	Options    string    `json:"options" gorm:"type:text"`  // JSON 格式选项（选择题用）
	Answer     string    `json:"answer" gorm:"type:text"`
	Explanation string   `json:"explanation" gorm:"type:text"` // 解析
	Hint1      string    `json:"hint_1" gorm:"type:text"`      // 第1级提示：概念方向
	Hint2      string    `json:"hint_2" gorm:"type:text"`      // 第2级提示：关键线索
	Hint3      string    `json:"hint_3" gorm:"type:text"`      // 第3级提示：接近答案
	CreatedAt  time.Time `json:"created_at"`
}

// BeforeCreate 创建前自动生成 UUID
func (q *Quiz) BeforeCreate(tx *gorm.DB) error {
	if q.ID == "" {
		q.ID = uuid.New().String()
	}
	return nil
}

// QuizAttempt 用户答题记录
type QuizAttempt struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"index;not null"`
	QuizID    string    `json:"quiz_id" gorm:"index;not null"`
	Answer    string    `json:"answer" gorm:"type:text"`
	IsCorrect bool      `json:"is_correct"`
	HintsUsed int       `json:"hints_used" gorm:"default:0"` // 使用的提示数量
	CreatedAt time.Time `json:"created_at"`
}

// BeforeCreate 创建前自动生成 UUID
func (qa *QuizAttempt) BeforeCreate(tx *gorm.DB) error {
	if qa.ID == "" {
		qa.ID = uuid.New().String()
	}
	return nil
}

// AnswerRequest 答题请求
type AnswerRequest struct {
	Answer    string `json:"answer" binding:"required"`
	HintsUsed int    `json:"hints_used"` // 使用的提示数量
}
