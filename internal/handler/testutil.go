package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"studyforge/internal/middleware"
	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB 创建 SQLite 内存测试数据库并自动迁移所有模型
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("测试数据库初始化失败: %v", err)
	}

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
	)
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}

	return db
}

// setupTestHandler 创建最小化的 Handler（仅 DB + JWT，无 LLM/VectorStore）
func setupTestHandler(db *gorm.DB) *Handler {
	hub := NewWSHub()
	go hub.Run()

	return &Handler{
		DB:               db,
		Hub:              hub,
		jwtSecret:        "test-secret-key-for-testing",
		jwtExpire:        24,
		memories:         make(map[string]*memoryEntry),
		analysisProgress: make(map[string]*MaterialAnalysisProgress),
	}
}

// setupTestRouter 搭建测试用的 Gin 路由（模拟 main.go 中的路由结构，含公开路由+认证路由）
func setupTestRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 公开路由
	api := r.Group("/api")
	{
		api.POST("/register", h.Register)
		api.POST("/login", h.Login)
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "StudyForge Pro"})
		})
	}

	// 认证路由
	auth := api.Group("")
	auth.Use(middleware.JWTAuth(h.jwtSecret))
	{
		auth.GET("/materials", h.ListMaterials)
		auth.GET("/materials/:id", h.GetMaterial)
		auth.POST("/materials", h.UploadMaterial)
		auth.DELETE("/materials/:id", h.DeleteMaterial)
	}

	return r
}

// createTestUser 在数据库中直接创建测试用户（bcrypt 加密密码）
func createTestUser(t *testing.T, db *gorm.DB, username, password string) *model.User {
	t.Helper()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("密码加密失败: %v", err)
	}

	user := model.User{
		Username: username,
		Email:    username + "@test.com",
		Password: string(hashedPassword),
		Nickname: "Test " + username,
	}

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}

	return &user
}

// getAuthToken 通过 Login 端点获取 JWT 认证 Token
func getAuthToken(t *testing.T, router *gin.Engine, username, password string) string {
	t.Helper()

	body, _ := json.Marshal(gin.H{"username": username, "password": password})
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("获取 token 失败，状态码: %d, 响应: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	token, ok := resp["token"].(string)
	if !ok || token == "" {
		t.Fatal("登录响应中缺少 token 字段")
	}

	return token
}

// authRequest 创建带 Bearer Token 的 HTTP 请求
func authRequest(t *testing.T, method, url, token string, body interface{}) *http.Request {
	t.Helper()

	var req *http.Request
	if body != nil {
		data, _ := json.Marshal(body)
		req = httptest.NewRequest(method, url, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

// TestMain 测试入口，设置 Gin 为测试模式
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
