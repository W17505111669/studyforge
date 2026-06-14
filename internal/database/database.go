package database

import (
	"log"
	"os"
	"path/filepath"

	"studyforge/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Init 初始化 SQLite 数据库并自动迁移表结构
func Init(dbPath string) (*gorm.DB, error) {
	// 确保数据库文件目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 连接 SQLite
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

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
	)
	if err != nil {
		return nil, err
	}

	log.Println("数据库迁移完成")
	return db, nil
}
