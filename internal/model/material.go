package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Material 学习材料模型
type Material struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	UserID      string    `json:"user_id" gorm:"index;not null"`
	Title       string    `json:"title" gorm:"size:200;not null"`
	ContentType string    `json:"content_type" gorm:"size:20"` // "text", "pdf", "url"
	Content     string    `json:"content" gorm:"type:text"`    // 原文内容
	SourceURL   string    `json:"source_url" gorm:"size:500"`  // 来源 URL（如有）
	Status      string    `json:"status" gorm:"size:20;default:'pending'"` // pending, analyzing, completed, failed
	Tags        string    `json:"tags" gorm:"size:500"` // 自定义标签，逗号分隔（如 "编程,Go,并发"）
	IsPublic    bool      `json:"is_public" gorm:"default:false"` // 是否公开到市场
	ShareCode   string    `json:"share_code,omitempty" gorm:"size:32;index;default:''"` // 分享码（公开时生成，非空时唯一）
	AnalyzedAt  *time.Time `json:"analyzed_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联
	Cards []Card `json:"cards,omitempty" gorm:"foreignKey:MaterialID"`
	Quizzes []Quiz `json:"quizzes,omitempty" gorm:"foreignKey:MaterialID"`
	GraphData    string `json:"graph_data,omitempty" gorm:"type:text"`    // JSON 格式的图谱数据
	AnalysisData string `json:"analysis_data,omitempty" gorm:"type:text"` // JSON 格式的分析师输出
}

// BeforeCreate 创建前自动生成 UUID
func (m *Material) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

// UploadRequest 上传材料请求
type UploadRequest struct {
	Title       string `json:"title" binding:"required"`
	ContentType string `json:"content_type" binding:"required,oneof=text url"`
	Content     string `json:"content"`
	SourceURL   string `json:"source_url"`
	Tags        string `json:"tags"` // 自定义标签，逗号分隔
}
