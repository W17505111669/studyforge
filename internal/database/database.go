package database

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"studyforge/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// slowQueryThreshold 慢查询阈值（超过此时间打印警告）
const slowQueryThreshold = 200 * time.Millisecond

// Init 初始化 SQLite 数据库并自动迁移表结构
func Init(dbPath string) (*gorm.DB, error) {
	// 确保数据库文件目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 配置 GORM Logger：慢查询阈值 200ms，Warn 级别
	logLevel := logger.Warn
	if os.Getenv("DB_LOG_LEVEL") == "info" {
		logLevel = logger.Info
	}
	customLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             slowQueryThreshold,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// 连接 SQLite
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: customLogger,
	})
	if err != nil {
		return nil, err
	}

	// 迁移前清理旧版全量唯一索引（已改为部分唯一索引，避免空字符串冲突）
	preMigrationIndexCleanup(db)

	// 自动迁移所有模型
	log.Println("正在执行数据库迁移...")
	err = db.AutoMigrate(
		&model.User{},
		&model.Material{},
		&model.Card{},
		&model.Quiz{},
		&model.QuizAttempt{},
		&model.LLMTrace{},
		&model.Conversation{},
		&model.ChatMessage{},
		&model.UserAchievement{},
		&model.QuizMistake{},
		&model.Notification{},
		&model.PomodoroSession{},
		&model.LearningGoal{},
		&model.StudyStreak{},
		&model.DailyTask{},
		&model.Note{},
		&model.NoteFolder{},
		&model.Deck{},
		&model.DeckCard{},
		&model.Friendship{},
		&model.StudyGroup{},
		&model.GroupMember{},
		&model.GroupGoal{},
		&model.ExamSession{},
		&model.ExplainCache{},
	)
	if err != nil {
		return nil, err
	}
	log.Println("数据库迁移完成")

	// 创建复合索引（优化高频查询）
	if err := CreateIndexes(db); err != nil {
		log.Printf("⚠ 索引创建过程出错: %v", err)
		// 索引创建失败不阻止启动，仅记录警告
	}

	return db, nil
}
