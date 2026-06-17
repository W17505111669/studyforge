package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExplainCache AI 概念解释缓存
type ExplainCache struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	UserID       string    `json:"user_id" gorm:"index;not null"`
	Concept      string    `json:"concept" gorm:"size:500;not null"`         // 查询的概念
	ConceptHash  string    `json:"concept_hash" gorm:"size:64;index"`        // 概念 hash（用于缓存去重）
	Explanation  string    `json:"explanation" gorm:"type:text"`             // 通俗解释
	Analogy      string    `json:"analogy" gorm:"type:text"`                 // 生活类比
	Example      string    `json:"example" gorm:"type:text"`                 // 具体例子
	RelatedConcepts string `json:"related_concepts" gorm:"type:text"`       // 关联知识点（JSON 数组字符串）
	CreatedAt    time.Time `json:"created_at"`
}

// BeforeCreate 创建前自动生成 UUID
func (e *ExplainCache) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

// ExplainRequest 概念解释请求
type ExplainRequest struct {
	Concept string `json:"concept" binding:"required,min=1,max=500"`
	Context string `json:"context,omitempty"` // 可选上下文（来源材料/卡片 ID）
}

// ExplainResponse 概念解释响应
type ExplainResponse struct {
	ID              string   `json:"id"`
	Concept         string   `json:"concept"`
	Explanation     string   `json:"explanation"`
	Analogy         string   `json:"analogy"`
	Example         string   `json:"example"`
	RelatedConcepts []string `json:"related_concepts"`
	Cached          bool     `json:"cached"` // 是否命中缓存
	CreatedAt       time.Time `json:"created_at"`
}
