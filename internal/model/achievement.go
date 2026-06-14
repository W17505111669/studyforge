package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserAchievement 用户成就解锁记录
type UserAchievement struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	UserID        string    `json:"user_id" gorm:"index;not null"`
	AchievementID string    `json:"achievement_id" gorm:"size:50;not null;index"`
	UnlockedAt    time.Time `json:"unlocked_at"`
}

// BeforeCreate 创建前自动生成 UUID
func (ua *UserAchievement) BeforeCreate(tx *gorm.DB) error {
	if ua.ID == "" {
		ua.ID = uuid.New().String()
	}
	return nil
}

// AchievementDef 成就定义（静态配置，不存储到数据库）
type AchievementDef struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Category    string `json:"category"` // learning, practice, review, exploration, special
	Target      int    `json:"target"`
}

// AchievementResponse API 返回的成就信息（含解锁状态和进度）
type AchievementResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Category    string    `json:"category"`
	Target      int       `json:"target"`
	Unlocked    bool      `json:"unlocked"`
	UnlockedAt  *time.Time `json:"unlocked_at,omitempty"`
	Progress    int       `json:"progress"`
	ProgressPct float64   `json:"progress_pct"`
}

// AllAchievements 所有成就定义（按类别分组）
var AllAchievements = []AchievementDef{
	// ===== 学习类 (learning) =====
	{ID: "first_upload", Name: "初学者", Description: "上传第一份学习材料", Icon: "📖", Category: "learning", Target: 1},
	{ID: "upload_5", Name: "求知若渴", Description: "累计上传 5 份学习材料", Icon: "📚", Category: "learning", Target: 5},
	{ID: "upload_20", Name: "学海无涯", Description: "累计上传 20 份学习材料", Icon: "🎓", Category: "learning", Target: 20},

	// ===== 练习类 (practice) =====
	{ID: "first_quiz", Name: "初试锋芒", Description: "回答第一道练习题", Icon: "✏️", Category: "practice", Target: 1},
	{ID: "quiz_50", Name: "勤学苦练", Description: "累计回答 50 道练习题", Icon: "📝", Category: "practice", Target: 50},
	{ID: "quiz_200", Name: "题海战术", Description: "累计回答 200 道练习题", Icon: "🏋️", Category: "practice", Target: 200},
	{ID: "perfect_round", Name: "满分通关", Description: "一轮练习全部答对（至少 3 题）", Icon: "💯", Category: "practice", Target: 1},

	// ===== 复习类 (review) =====
	{ID: "first_review", Name: "温故知新", Description: "首次复习知识卡片", Icon: "🔄", Category: "review", Target: 1},
	{ID: "review_30", Name: "持之以恒", Description: "累计复习 30 次知识卡片", Icon: "📅", Category: "review", Target: 30},
	{ID: "review_100", Name: "记忆大师", Description: "累计复习 100 次知识卡片", Icon: "🧠", Category: "review", Target: 100},

	// ===== 探索类 (exploration) =====
	{ID: "first_chat", Name: "求知问道", Description: "首次与 AI 助手对话", Icon: "💬", Category: "exploration", Target: 1},
	{ID: "chat_50", Name: "好学不倦", Description: "累计发送 50 条对话消息", Icon: "🗨️", Category: "exploration", Target: 50},
	{ID: "graph_view", Name: "图谱探险家", Description: "查看知识图谱（有图谱数据生成）", Icon: "🗺️", Category: "exploration", Target: 1},
	{ID: "card_export", Name: "知识整理师", Description: "导出过知识卡片", Icon: "📤", Category: "exploration", Target: 1},

	// ===== 特殊类 (special) =====
	{ID: "accuracy_80", Name: "学霸潜质", Description: "答题正确率达到 80%（至少 10 题）", Icon: "🌟", Category: "special", Target: 80},
	{ID: "card_collector", Name: "卡片收藏家", Description: "累计收集 50 张知识卡片", Icon: "🃏", Category: "special", Target: 50},
	{ID: "all_agents", Name: "全能学者", Description: "触发全部 4 种 Agent 调用", Icon: "🏆", Category: "special", Target: 4},
	{ID: "night_owl", Name: "夜猫子", Description: "在凌晨 0 点至 5 点之间学习", Icon: "🦉", Category: "special", Target: 1},
}

// AchievementCategories 成就类别定义（用于前端展示）
var AchievementCategories = map[string]string{
	"learning":    "学习",
	"practice":    "练习",
	"review":      "复习",
	"exploration": "探索",
	"special":     "特殊",
}
