package database

import (
	"log"

	"gorm.io/gorm"
)

// IndexDefinition 索引定义
type IndexDefinition struct {
	Table  string
	Name   string
	SQL    string
}

// compositeIndexes 定义所有复合索引
// 这些索引针对项目中高频查询模式进行优化
var compositeIndexes = []IndexDefinition{
	// ===== share_code 部分唯一索引（非空时唯一）=====
	{
		Table: "materials",
		Name:  "idx_materials_share_code_unique",
		SQL:   "CREATE UNIQUE INDEX IF NOT EXISTS idx_materials_share_code_unique ON materials (share_code) WHERE share_code != ''",
	},
	{
		Table: "decks",
		Name:  "idx_decks_share_code_unique",
		SQL:   "CREATE UNIQUE INDEX IF NOT EXISTS idx_decks_share_code_unique ON decks (share_code) WHERE share_code != ''",
	},

	// ===== cards: 待复习查询、材料关联、书签过滤 =====
	{
		Table: "cards",
		Name:  "idx_cards_user_next_review",
		SQL:   "CREATE INDEX IF NOT EXISTS idx_cards_user_next_review ON cards (user_id, next_review_at)",
	},
	{
		Table: "cards",
		Name:  "idx_cards_user_material",
		SQL:   "CREATE INDEX IF NOT EXISTS idx_cards_user_material ON cards (user_id, material_id)",
	},
	{
		Table: "cards",
		Name:  "idx_cards_user_bookmarked",
		SQL:   "CREATE INDEX IF NOT EXISTS idx_cards_user_bookmarked ON cards (user_id, is_bookmarked)",
	},

	// ===== quizzes: 材料关联、时间排序 =====
	{
		Table: "quizzes",
		Name:  "idx_quizzes_user_material",
		SQL:   "CREATE INDEX IF NOT EXISTS idx_quizzes_user_material ON quizzes (user_id, material_id)",
	},
	{
		Table: "quizzes",
		Name:  "idx_quizzes_user_created",
		SQL:   "CREATE INDEX IF NOT EXISTS idx_quizzes_user_created ON quizzes (user_id, created_at)",
	},

	// ===== quiz_attempts: 答题历史查询 =====
	{
		Table: "quiz_attempts",
		Name:  "idx_quiz_attempts_user_quiz",
		SQL:   "CREATE INDEX IF NOT EXISTS idx_quiz_attempts_user_quiz ON quiz_attempts (user_id, quiz_id)",
	},
	{
		Table: "quiz_attempts",
		Name:  "idx_quiz_attempts_user_created",
		SQL:   "CREATE INDEX IF NOT EXISTS idx_quiz_attempts_user_created ON quiz_attempts (user_id, created_at)",
	},

	// ===== chat_messages: 对话消息按时间排序 =====
	{
		Table: "chat_messages",
		Name:  "idx_chat_messages_conv_created",
		SQL:   "CREATE INDEX IF NOT EXISTS idx_chat_messages_conv_created ON chat_messages (conversation_id, created_at)",
	},

	// ===== conversations: 用户对话列表按更新时间排序 =====
	{
		Table: "conversations",
		Name:  "idx_conversations_user_updated",
		SQL:   "CREATE INDEX IF NOT EXISTS idx_conversations_user_updated ON conversations (user_id, updated_at)",
	},

	// ===== materials: 状态过滤、时间排序 =====
	{
		Table: "materials",
		Name:  "idx_materials_user_status",
		SQL:   "CREATE INDEX IF NOT EXISTS idx_materials_user_status ON materials (user_id, status)",
	},
	{
		Table: "materials",
		Name:  "idx_materials_user_created",
		SQL:   "CREATE INDEX IF NOT EXISTS idx_materials_user_created ON materials (user_id, created_at)",
	},

	// ===== notifications: 已读/未读查询 =====
	{
		Table: "notifications",
		Name:  "idx_notifications_user_read",
		SQL:   "CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications (user_id, read_at)",
	},

	// ===== pomodoro_sessions: 专注统计按时间 =====
	{
		Table: "pomodoro_sessions",
		Name:  "idx_pomodoro_user_started",
		SQL:   "CREATE INDEX IF NOT EXISTS idx_pomodoro_user_started ON pomodoro_sessions (user_id, started_at)",
	},

	// ===== quiz_mistakes: 错题本已复习状态过滤 =====
	{
		Table: "quiz_mistakes",
		Name:  "idx_mistakes_user_reviewed",
		SQL:   "CREATE INDEX IF NOT EXISTS idx_mistakes_user_reviewed ON quiz_mistakes (user_id, reviewed)",
	},

	// ===== daily_tasks: 每日任务按日期查询 =====
	{
		Table: "daily_tasks",
		Name:  "idx_daily_tasks_user_date",
		SQL:   "CREATE INDEX IF NOT EXISTS idx_daily_tasks_user_date ON daily_tasks (user_id, task_date)",
	},

	// ===== learning_goals: 目标按状态过滤 =====
	{
		Table: "learning_goals",
		Name:  "idx_goals_user_status",
		SQL:   "CREATE INDEX IF NOT EXISTS idx_goals_user_status ON learning_goals (user_id, status)",
	},
}

// CreateIndexes 在 AutoMigrate 后创建复合索引
func CreateIndexes(db *gorm.DB) error {
	log.Println("正在创建复合索引...")

	created := 0
	for _, idx := range compositeIndexes {
		result := db.Exec(idx.SQL)
		if result.Error != nil {
			log.Printf("⚠ 创建索引 %s 失败: %v", idx.Name, result.Error)
			continue
		}
		created++
	}

	log.Printf("复合索引创建完成: %d/%d", created, len(compositeIndexes))
	return nil
}

// preMigrationIndexCleanup 在 AutoMigrate 前清理旧版全量唯一索引
// 这些旧索引在 share_code 列上使用了 FULL UNIQUE 约束，当多行 share_code 为空字符串时会冲突
// 现已替换为部分唯一索引 (WHERE share_code != '')
func preMigrationIndexCleanup(db *gorm.DB) {
	oldIndexes := []string{
		"idx_materials_share_code",
		"idx_decks_share_code",
	}
	for _, name := range oldIndexes {
		result := db.Exec("DROP INDEX IF EXISTS " + name)
		if result.Error != nil {
			log.Printf("⚠ 清理旧索引 %s 失败: %v", name, result.Error)
		}
	}
}
