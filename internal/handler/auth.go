package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"studyforge/internal/agent"
	"studyforge/internal/memory"
	"studyforge/internal/middleware"
	"studyforge/internal/model"
	"studyforge/internal/rag"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AgentProgress 单个 Agent 的执行进度
type AgentProgress struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "waiting", "running", "done", "error"
}

// MaterialAnalysisProgress 材料分析的完整进度
type MaterialAnalysisProgress struct {
	Agents   []AgentProgress `json:"agents"`
	Error    string          `json:"error,omitempty"`
	Started  time.Time       `json:"started"`
}

// Handler 聚合所有 HTTP 请求处理器
type Handler struct {
	DB           *gorm.DB
	Orchestrator *agent.Orchestrator
	Hub          *WSHub
	LLM          *agent.LLMClient
	VectorStore  *rag.VectorStore
	jwtSecret    string
	jwtExpire    int

	// 会话记忆：每个用户一个 ConversationMemory
	memoryMu    sync.RWMutex
	memories    map[string]*memoryEntry
	loadingKeys sync.Map // 防止同一 convID 并发加载 DB

	// 分析进度追踪（用于 WebSocket 断线时的轮询 fallback）
	analysisMu       sync.RWMutex
	analysisProgress map[string]*MaterialAnalysisProgress // key: materialID
	activeAnalyses   sync.Map                             // key: materialID → true，限制并发

	// RAG 配置
	chunkSize    int
	chunkOverlap int
	ragTopK      int

	// 学习路径缓存：每个用户 30 分钟内复用，避免重复 LLM 调用
	pathCacheMu sync.RWMutex
	pathCache   map[string]*pathCacheEntry
}

// NewHandler 创建 Handler 实例，注入所有依赖
func NewHandler(db *gorm.DB, jwtSecret string, jwtExpire int, apiKey, baseURL, modelName, embeddingModel string, vectorStore *rag.VectorStore, chunkSize, chunkOverlap, ragTopK int) *Handler {
	hub := NewWSHub()
	go hub.Run()

	llm := agent.NewLLMClient(apiKey, baseURL, modelName, embeddingModel)
	orch := agent.NewOrchestratorWithLLM(db, llm)

	// 将 VectorStore 注入到 Orchestrator（用于分析后自动 RAG 索引）
	orch.VectorStore = vectorStore

	h := &Handler{
		DB:               db,
		Orchestrator:     orch,
		Hub:              hub,
		LLM:              llm,
		VectorStore:      vectorStore,
		jwtSecret:        jwtSecret,
		jwtExpire:        jwtExpire,
		memories:         make(map[string]*memoryEntry),
		analysisProgress: make(map[string]*MaterialAnalysisProgress),
		pathCache:        make(map[string]*pathCacheEntry),
		chunkSize:        chunkSize,
		chunkOverlap:     chunkOverlap,
		ragTopK:          ragTopK,
	}

	// 定期清理过期的会话记忆（防止内存无限增长）
	go h.cleanupStaleMemories()

	// 设置 Orchestrator 的结果回调：通过 WebSocket 定向推送给触发分析的用户，并更新进度追踪
	orch.OnResult = func(result agent.AgentResult) {
		data, _ := json.Marshal(WSMessage{
			Type:      "agent_output",
			AgentName: result.AgentName,
			Content:   result,
			Timestamp: time.Now(),
		})
		hub.BroadcastToUser(result.UserID, data)

		// 更新分析进度（用于轮询 fallback）
		h.analysisMu.Lock()
		if prog, ok := h.analysisProgress[result.MaterialID]; ok {
			for i, a := range prog.Agents {
				if a.Name == result.AgentName {
					if result.Status == "success" {
						prog.Agents[i].Status = "done"
					} else {
						prog.Agents[i].Status = "error"
					}
					break
				}
			}
		}
		h.analysisMu.Unlock()
	}

	return h
}

// memoryEntry 包装 ConversationMemory 并追踪访问时间（atomic 防竞态）
type memoryEntry struct {
	mem        *memory.ConversationMemory
	lastAccess atomic.Int64 // UnixNano 时间戳，原子读写
}

// touchAccess 原子更新访问时间
func (e *memoryEntry) touchAccess() {
	e.lastAccess.Store(time.Now().UnixNano())
}

// getAccessTime 原子读取访问时间
func (e *memoryEntry) getAccessTime() time.Time {
	return time.Unix(0, e.lastAccess.Load())
}

// getOrCreateMemory 获取或创建用户的会话记忆（带加载锁防并发 DB 读取）
func (h *Handler) getOrCreateMemory(convID string) *memory.ConversationMemory {
	h.memoryMu.RLock()
	if entry, ok := h.memories[convID]; ok {
		entry.touchAccess()
		h.memoryMu.RUnlock()
		return entry.mem
	}
	h.memoryMu.RUnlock()

	h.memoryMu.Lock()
	// 双重检查
	if entry, ok := h.memories[convID]; ok {
		entry.touchAccess()
		h.memoryMu.Unlock()
		return entry.mem
	}
	mem := memory.NewConversationMemory(10)
	entry := &memoryEntry{mem: mem}
	entry.touchAccess()
	h.memories[convID] = entry
	h.memoryMu.Unlock()
	return mem
}

// loadMemoryFromDB 从数据库加载对话历史到内存中（带加载锁，同一 convID 只加载一次）
func (h *Handler) loadMemoryFromDB(convID string, mem *memory.ConversationMemory) {
	// 使用 sync.Map 的 LoadOrStore 保证同一 convID 只有一个 goroutine 加载
	if _, loaded := h.loadingKeys.LoadOrStore(convID, true); loaded {
		return // 另一个 goroutine 正在加载此对话
	}
	defer h.loadingKeys.Delete(convID)

	var messages []struct {
		Role    string
		Content string
	}
	h.DB.Model(&model.ChatMessage{}).
		Select("role, content").
		Where("conversation_id = ?", convID).
		Order("created_at ASC").
		Limit(20). // 最近 20 条消息（10 轮对话）
		Find(&messages)

	for _, msg := range messages {
		mem.AddMessage(msg.Role, msg.Content)
	}
}

// EvictConversationMemory 清除指定对话的内存缓存（对话删除时调用）
func (h *Handler) EvictConversationMemory(convID string) {
	h.memoryMu.Lock()
	delete(h.memories, convID)
	h.memoryMu.Unlock()
}

// cleanupStaleMemories 定期清理超过 30 分钟未访问的会话记忆
func (h *Handler) cleanupStaleMemories() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.memoryMu.Lock()
		now := time.Now()
		evicted := 0
		for convID, entry := range h.memories {
			if now.Sub(entry.getAccessTime()) > 30*time.Minute {
				delete(h.memories, convID)
				evicted++
			}
		}
		h.memoryMu.Unlock()
		if evicted > 0 {
			log.Printf("cleanupStaleMemories: 清理了 %d 个过期会话记忆", evicted)
		}
	}
}

// ==================== 认证相关 ====================

// Register 用户注册
// POST /api/register
func (h *Handler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		msg := parseValidationErr(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// 输入裁剪：去除用户名和邮箱两端空白，防止空白字符绕过唯一性检查
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	// 检查用户名是否已存在
	var count int64
	h.DB.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		return
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	user := model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	// 生成 Token
	token, err := middleware.GenerateToken(user.ID, user.Username, h.jwtSecret, h.jwtExpire)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 Token 失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"nickname": user.Nickname,
		},
	})
}

// Login 用户登录
// POST /api/login
func (h *Handler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入用户名和密码"})
		return
	}

	var user model.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 生成 Token
	token, err := middleware.GenerateToken(user.ID, user.Username, h.jwtSecret, h.jwtExpire)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 Token 失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"nickname": user.Nickname,
		},
	})
}

// parseValidationErr 将 gin 的绑定/验证错误翻译为用户友好的中文提示
func parseValidationErr(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "Username") && strings.Contains(msg, "min"):
		return "用户名至少需要 3 个字符"
	case strings.Contains(msg, "Username") && strings.Contains(msg, "required"):
		return "请输入用户名"
	case strings.Contains(msg, "Password") && strings.Contains(msg, "min"):
		return "密码至少需要 6 个字符"
	case strings.Contains(msg, "Password") && strings.Contains(msg, "required"):
		return "请输入密码"
	case strings.Contains(msg, "Email") && strings.Contains(msg, "email"):
		return "请输入有效的邮箱地址"
	case strings.Contains(msg, "Email") && strings.Contains(msg, "required"):
		return "请输入邮箱"
	default:
		return "请检查输入内容"
	}
}
