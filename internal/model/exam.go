package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExamSession 考试会话模型
type ExamSession struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	UserID      string     `json:"user_id" gorm:"index;not null"`
	MaterialIDs string     `json:"material_ids" gorm:"type:text"` // JSON array of material IDs
	Questions   string     `json:"questions" gorm:"type:text"`    // JSON array of ExamQuestion
	Answers     string     `json:"answers" gorm:"type:text"`      // JSON array of ExamAnswer (submitted)
	TimeLimit   int        `json:"time_limit" gorm:"default:30"`  // minutes
	StartedAt   time.Time  `json:"started_at"`
	EndedAt     *time.Time `json:"ended_at"`
	Score       float64    `json:"score" gorm:"default:0"`        // 0-100
	Report      string     `json:"report" gorm:"type:text"`       // JSON report
	Status      string     `json:"status" gorm:"size:20;default:in_progress"` // in_progress, completed, timed_out
	CreatedAt   time.Time  `json:"created_at"`
}

// BeforeCreate 创建前自动生成 UUID
func (e *ExamSession) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

// ExamQuestion 考试题目结构（存储在 JSON 中）
type ExamQuestion struct {
	Index      int      `json:"index"`
	Type       string   `json:"type"`       // choice, true_false, fill, short_answer
	Question   string   `json:"question"`
	Options    []string `json:"options,omitempty"`    // 选择题选项
	Answer     string   `json:"answer"`               // 正确答案
	Explanation string  `json:"explanation"`           // 解析
	Difficulty string   `json:"difficulty"`            // easy, medium, hard
	MaterialID string   `json:"material_id,omitempty"` // 来源材料
	Concept    string   `json:"concept,omitempty"`     // 关联知识点
	Points     int      `json:"points"`                // 分值
	Marked     bool     `json:"marked,omitempty"`      // 用户标记为不确定
}

// ExamAnswer 用户考试作答
type ExamAnswer struct {
	Index  int    `json:"index"`
	Answer string `json:"answer"`
	Marked bool   `json:"marked"`
}

// ExamReport 考试报告
type ExamReport struct {
	TotalScore    float64              `json:"total_score"`
	MaxScore      int                  `json:"max_score"`
	Percentage    float64              `json:"percentage"`
	TimeUsed      int                  `json:"time_used_seconds"`
	TypeStats     map[string]*TypeStat `json:"type_stats"`
	WeakPoints    []WeakPoint          `json:"weak_points"`
	QuestionResults []QuestionResult   `json:"question_results"`
}

// TypeStat 题型统计
type TypeStat struct {
	Total   int     `json:"total"`
	Correct int     `json:"correct"`
	Rate    float64 `json:"rate"`
}

// WeakPoint 薄弱知识点
type WeakPoint struct {
	Concept    string `json:"concept"`
	MaterialID string `json:"material_id"`
	Reason     string `json:"reason"`
}

// QuestionResult 单题结果
type QuestionResult struct {
	Index       int    `json:"index"`
	Type        string `json:"type"`
	IsCorrect   bool   `json:"is_correct"`
	UserAnswer  string `json:"user_answer"`
	CorrectAnswer string `json:"correct_answer"`
	Explanation string `json:"explanation"`
	Concept     string `json:"concept"`
}

// GenerateExamRequest 生成考试请求
type GenerateExamRequest struct {
	MaterialIDs   []string `json:"material_ids" binding:"required"`
	QuestionCount int      `json:"question_count" binding:"required,min=5,max=100"`
	TimeLimit     int      `json:"time_limit" binding:"required,min=5,max=180"` // minutes
}

// SubmitExamRequest 提交考试请求
type SubmitExamRequest struct {
	Answers []ExamAnswer `json:"answers" binding:"required"`
}
