package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Note 知识笔记
type Note struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	UserID     string    `json:"user_id" gorm:"index;not null"`
	Title      string    `json:"title" gorm:"size:200;not null;default:'无标题笔记'"`
	Content    string    `json:"content" gorm:"type:text"`                           // Markdown 内容
	MaterialID string    `json:"material_id,omitempty" gorm:"size:36;index"`         // 关联材料（可选）
	CardID     string    `json:"card_id,omitempty" gorm:"size:36;index"`             // 关联卡片（可选）
	FolderID   string    `json:"folder_id,omitempty" gorm:"size:36;index"`           // 所属文件夹（可选）
	Pinned     bool      `json:"pinned" gorm:"not null;default:false"`               // 是否置顶
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// BeforeCreate 创建前自动生成 UUID
func (n *Note) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	return nil
}

// NoteFolder 笔记文件夹
type NoteFolder struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"index;not null"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	Color     string    `json:"color" gorm:"size:20;not null;default:'#6366f1'"` // 文件夹颜色（hex）
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate 创建前自动生成 UUID
func (f *NoteFolder) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	return nil
}

// NoteFolderWithCount 文件夹 + 笔记计数（响应用）
type NoteFolderWithCount struct {
	NoteFolder
	NoteCount int `json:"note_count" gorm:"-"`
}

// FolderColors 预设文件夹颜色
var FolderColors = []string{
	"#6366f1", // indigo
	"#ec4899", // pink
	"#10b981", // emerald
	"#f59e0b", // amber
	"#3b82f6", // blue
	"#8b5cf6", // violet
	"#ef4444", // red
	"#06b6d4", // cyan
}
