package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"studyforge/internal/agent"
	"studyforge/internal/database"
	"studyforge/internal/handler"
	"studyforge/internal/middleware"
	"studyforge/internal/rag"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置（支持 CONFIG_PATH 环境变量指定配置文件路径）
	cfgPath := GetConfigPath()
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n✘ 加载配置失败: %v\n", err)
		fmt.Fprintf(os.Stderr, "  配置文件路径: %s\n", cfgPath)
		fmt.Fprintf(os.Stderr, "  提示: 复制 config.yaml.example 为 config.yaml 并填入配置，或通过环境变量设置\n")
		fmt.Fprintf(os.Stderr, "  详见 .env.example 了解所有可用的环境变量\n\n")
		os.Exit(1)
	}

	// 校验必需配置项
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "\n✘ %v\n\n", err)
		os.Exit(1)
	}

	cfg.PrintSummary()

	// 初始化数据库
	db, err := database.Init(cfg.Database.Path)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 确保数据目录存在
	os.MkdirAll("data", 0755)

	// 初始化 LLM 客户端（用于 VectorStore 的 embedding）
	llmForRAG := agent.NewLLMClient(cfg.LLM.APIKey, cfg.LLM.BaseURL, cfg.LLM.Model, cfg.LLM.EmbeddingModel)

	// 初始化内存向量存储（RAG）
	vectorStore, err := rag.NewVectorStore(
		cfg.Qdrant.Host,
		cfg.Qdrant.Port,
		cfg.Qdrant.Collection,
		cfg.Qdrant.VectorSize,
		llmForRAG,
	)
	if err != nil {
		log.Fatalf("向量存储初始化失败: %v", err)
	}
	defer vectorStore.Close()

	log.Printf("RAG 向量存储初始化完成（内存模式）")

	// 创建 Gin 引擎
	r := gin.Default()

	// 禁止信任代理头（防止 X-Forwarded-For 伪造 IP 绕过限流）
	r.SetTrustedProxies(nil)

	// CORS 中间件
	r.Use(corsMiddleware())

	// 初始化管理器（注入所有依赖：DB、LLM、VectorStore、RAG 配置）
	h := handler.NewHandler(
		db,
		cfg.JWT.Secret,
		cfg.JWT.ExpireHours,
		cfg.LLM.APIKey,
		cfg.LLM.BaseURL,
		cfg.LLM.Model,
		cfg.LLM.EmbeddingModel,
		vectorStore,
		cfg.RAG.ChunkSize,
		cfg.RAG.ChunkOverlap,
		cfg.RAG.TopK,
	)

	// ========== IP 限流 ==========
	// 公开路由：严格限流（防暴力破解）
	publicLimiter := middleware.NewIPRateLimiter(5, 10)
	// 认证路由：宽松限流
	authLimiter := middleware.NewIPRateLimiter(20, 40)

	// ========== 公开路由 ==========
	api := r.Group("/api")
	api.Use(publicLimiter.Middleware())
	{
		// 认证
		api.POST("/register", h.Register)
		api.POST("/login", h.Login)

		// 健康检查
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "service": "StudyForge Pro"})
		})
	}

	// ========== 需要认证的路由 ==========
	auth := api.Group("")
	auth.Use(middleware.JWTAuth(cfg.JWT.Secret), authLimiter.Middleware())
	{
		// 材料管理
		auth.GET("/tags", h.GetTags)
		auth.POST("/materials", h.UploadMaterial)
		auth.POST("/materials/upload", h.UploadFile)
		auth.POST("/materials/batch-analyze", h.BatchAnalyzeMaterials)
		auth.DELETE("/materials/batch", h.BatchDeleteMaterials)
		auth.GET("/materials", h.ListMaterials)
		auth.GET("/materials/:id", h.GetMaterial)
		auth.DELETE("/materials/:id", h.DeleteMaterial)

		// 分析（触发 Agent 并发处理）
		auth.POST("/materials/:id/analyze", h.AnalyzeMaterial)
		auth.GET("/materials/:id/status", h.GetMaterialStatus)

		// 知识卡片
		auth.GET("/cards/export", h.ExportCards)
		auth.GET("/cards", h.ListCards)
		auth.GET("/cards/:id", h.GetCard)
		auth.POST("/cards/:id/review", h.ReviewCard)
		auth.PUT("/cards/:id/bookmark", h.ToggleBookmark)
		auth.PUT("/cards/:id/note", h.UpdateCardNote)

		// 练习题
		auth.GET("/quizzes/difficulty-level", h.GetDifficultyLevel)
		auth.GET("/quizzes", h.ListQuizzes)
		auth.GET("/quizzes/:id/hint", h.GetQuizHint)
		auth.POST("/quizzes/:id/answer", h.AnswerQuiz)

		// 错题本
		auth.GET("/mistakes/stats", h.GetMistakeStats)
		auth.GET("/mistakes", h.ListMistakes)
		auth.POST("/mistakes/consolidate", h.ConsolidatePractice)
		auth.POST("/mistakes/retry", h.RetryMistakes)
		auth.POST("/mistakes/batch-review", h.BatchReviewMistakes)
		auth.POST("/mistakes/:id/review", h.ReviewMistake)
		auth.DELETE("/mistakes/:id", h.DeleteMistake)

		// 对话（多轮对话 + Function Calling）
		auth.POST("/chat", h.Chat)

		// 对话流式输出（SSE 打字机效果）
		auth.GET("/chat/stream", h.ChatStream)

		// 对话会话管理
		auth.GET("/conversations", h.ListConversations)
		auth.GET("/conversations/:id", h.GetConversation)
		auth.POST("/conversations", h.CreateConversation)
		auth.PUT("/conversations/:id", h.UpdateConversation)
		auth.DELETE("/conversations/:id", h.DeleteConversation)

		// 全局搜索
		auth.GET("/search", h.GlobalSearch)

		// 学习成就
		auth.GET("/achievements", h.GetAchievements)

		// 学习建议
		auth.GET("/recommendations", h.GetRecommendations)

		// 学习路径规划
		auth.GET("/learning-path", h.GetLearningPath)

		// 多 Agent 辩论
		auth.POST("/debate", h.StartDebate)

		// 通知系统
		auth.GET("/notifications/unread-count", h.GetUnreadNotificationCount)
		auth.POST("/notifications/read-all", h.ReadAllNotifications)
		auth.GET("/notifications", h.ListNotifications)
		auth.POST("/notifications/:id/read", h.ReadNotification)

		// 知识图谱
		auth.GET("/graph/all", h.GetAllKnowledgeGraphs)
		auth.GET("/graph/:material_id", h.GetKnowledgeGraph)

		// 用户学习统计
		auth.GET("/stats", h.GetUserStats)
		auth.GET("/stats/calendar", h.GetCalendarHeatmap)

		// Dashboard / 可观测性
		auth.GET("/dashboard/metrics", h.GetMetrics)
		auth.GET("/dashboard/activity", h.GetDailyActivity)
		auth.GET("/dashboard/traces", h.ListTraces)

		// 示例数据
		auth.POST("/seed", h.SeedData)
	}

	// WebSocket 端点
	r.GET("/ws", h.HandleWebSocket)

	// ========== 静态文件（Vue 前端 + PWA） ==========

	// PWA 核心文件（从 public 目录直接提供，不缓存 SW）
	r.StaticFile("/manifest.json", "./web/public/manifest.json")
	r.GET("/sw.js", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Service-Worker-Allowed", "/")
		c.File("./web/public/sw.js")
	})
	r.StaticFile("/offline.html", "./web/public/offline.html")

	// PWA 图标
	r.StaticFile("/icon.svg", "./web/public/icon.svg")
	r.StaticFile("/icon-192.png", "./web/public/icon-192.png")
	r.StaticFile("/icon-512.png", "./web/public/icon-512.png")
	r.StaticFile("/apple-touch-icon.png", "./web/public/apple-touch-icon.png")
	r.StaticFile("/favicon-32.png", "./web/public/favicon-32.png")
	r.StaticFile("/favicon-16.png", "./web/public/favicon-16.png")

	// Vite 构建产物（JS/CSS 带 hash，长期缓存）
	r.Static("/assets", "./web/dist/assets")

	// SPA 路由 fallback: 非 API/WebSocket/静态资源 请求返回 index.html
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// API 和 WebSocket 路径返回 404
		if strings.HasPrefix(path, "/api/") || path == "/ws" {
			c.JSON(404, gin.H{"error": "endpoint not found"})
			return
		}
		// 尝试提供 public 目录下的静态文件（防路径穿越）
		publicDir := filepath.Join(".", "web", "public")
		filePath := filepath.Join(publicDir, filepath.Clean("/"+path))
		// 确保解析后的路径仍在 public 目录内
		absPublic, _ := filepath.Abs(publicDir)
		absFile, _ := filepath.Abs(filePath)
		if strings.HasPrefix(absFile, absPublic) {
			if info, err := os.Stat(absFile); err == nil && !info.IsDir() {
				c.File(absFile)
				return
			}
		}
		// 其他路径返回 SPA index.html（由 Vue Router 处理前端路由）
		c.File("./web/dist/index.html")
	})

	// 启动服务
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("StudyForge Pro 启动于 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// corsMiddleware 处理跨域请求（安全模式：仅允许白名单 Origin）
func corsMiddleware() gin.HandlerFunc {
	// 允许的 Origin 白名单（开发阶段包含本地地址）
	allowedOrigins := map[string]bool{
		"http://localhost:5173":  true, // Vite dev server
		"http://localhost:8080":  true, // 后端同源
		"http://127.0.0.1:5173": true,
		"http://127.0.0.1:8080": true,
	}

	// 生产环境可通过环境变量追加
	if extra := os.Getenv("CORS_ORIGINS"); extra != "" {
		for _, origin := range strings.Split(extra, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				allowedOrigins[origin] = true
			}
		}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			// 同源请求，无需 CORS 头
			c.Next()
			return
		}

		if allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			c.Header("Access-Control-Max-Age", "86400")
		}
		// 不在白名单的 Origin：不设置任何 CORS 头，浏览器会拒绝跨域请求

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
