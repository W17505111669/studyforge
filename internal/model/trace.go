package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LLMTrace LLM 调用追踪记录
type LLMTrace struct {
	ID           string        `json:"id" gorm:"primaryKey"`
	UserID       string        `json:"user_id" gorm:"index"`
	AgentName    string        `json:"agent_name" gorm:"size:50"`    // Analyst / QuizMaster / CardMaker / MapBuilder / Judge
	Model        string        `json:"model" gorm:"size:50"`         // qwen-max / qwen-plus
	InputTokens  int           `json:"input_tokens"`
	OutputTokens int           `json:"output_tokens"`
	TotalTokens  int           `json:"total_tokens"`
	Latency      time.Duration `json:"latency"`                      // 调用耗时
	LatencyMs    int64         `json:"latency_ms" gorm:"-"`          // 毫秒（JSON 序列化用）
	QualityScore  float64       `json:"quality_score"`                  // 0-10 分（Judge LLM 评分）
	JudgeComment  string        `json:"judge_comment" gorm:"size:1000"` // Judge 评语反馈
	PromptSummary string        `json:"prompt_summary" gorm:"size:200"` // Prompt 摘要（不存完整 prompt）
	Status       string        `json:"status" gorm:"size:20"`        // success / error
	ErrorMessage string        `json:"error_message" gorm:"size:500"`
	CreatedAt    time.Time     `json:"created_at"`
}

// BeforeCreate 创建前自动生成 UUID
func (t *LLMTrace) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	t.LatencyMs = t.Latency.Milliseconds()
	return nil
}

// DashboardMetrics Dashboard 聚合指标
type DashboardMetrics struct {
	TotalCalls      int64   `json:"total_calls"`
	AvgLatencyMs    float64 `json:"avg_latency_ms"`
	TotalTokens     int64   `json:"total_tokens"`
	AvgQualityScore float64 `json:"avg_quality_score"`
	AgentBreakdown  map[string]int64 `json:"agent_breakdown"` // 各 Agent 调用次数
}
