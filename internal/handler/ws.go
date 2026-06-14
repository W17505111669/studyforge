package handler

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"studyforge/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ==================== WebSocket Hub ====================
// Hub 模式：管理所有 WebSocket 客户端连接，支持广播消息
// 这是 gorilla/websocket chat 示例的核心模式

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // 同源请求无 Origin 头
		}
		// 与 CORS 白名单保持一致
		allowed := map[string]bool{
			"http://localhost:5173":  true,
			"http://localhost:8080":  true,
			"http://127.0.0.1:5173": true,
			"http://127.0.0.1:8080": true,
		}
		// 支持环境变量扩展
		if extra := os.Getenv("CORS_ORIGINS"); extra != "" {
			for _, o := range strings.Split(extra, ",") {
				o = strings.TrimSpace(o)
				if o != "" {
					allowed[o] = true
				}
			}
		}
		return allowed[origin]
	},
}

// WSMessage WebSocket 消息结构
type WSMessage struct {
	Type          string      `json:"type"`                 // "agent_output", "analysis_complete", "error"
	AgentName     string      `json:"agent_name"`           // "Analyst" / "QuizMaster" / "CardMaker" / "MapBuilder"
	Content       interface{} `json:"content"`              // 消息内容
	Timestamp     time.Time   `json:"timestamp"`
	TargetUserID  string      `json:"-"`                    // 内部字段：定向推送给特定用户（不序列化到 JSON）
}

// WSClient 单个 WebSocket 客户端
type WSClient struct {
	hub    *WSHub
	conn   *websocket.Conn
	send   chan []byte
	userID string // 认证后的用户 ID，用于定向推送
}

// scopedMessage 定向推送消息（仅推送给特定用户）
type scopedMessage struct {
	userID string
	data   []byte
}

// WSHub WebSocket 连接管理中心
type WSHub struct {
	clients         map[*WSClient]bool
	broadcast       chan []byte
	broadcastScoped chan scopedMessage
	register        chan *WSClient
	unregister      chan *WSClient
	mu              sync.RWMutex
}

// NewWSHub 创建 WebSocket Hub
func NewWSHub() *WSHub {
	return &WSHub{
		clients:         make(map[*WSClient]bool),
		broadcast:       make(chan []byte),
		broadcastScoped: make(chan scopedMessage),
		register:        make(chan *WSClient),
		unregister:      make(chan *WSClient),
	}
}

// Run 启动 Hub 的消息循环（在单独的 goroutine 中运行）
func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket 客户端连接 (userID=%s)，当前在线: %d", client.userID, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.Lock()
			var deadClients []*WSClient
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					deadClients = append(deadClients, client)
				}
			}
			for _, client := range deadClients {
				close(client.send)
				delete(h.clients, client)
			}
			h.mu.Unlock()

		case scoped := <-h.broadcastScoped:
			h.mu.Lock()
			var deadClients []*WSClient
			for client := range h.clients {
				// 只推送给目标用户
				if client.userID != scoped.userID {
					continue
				}
				select {
				case client.send <- scoped.data:
				default:
					deadClients = append(deadClients, client)
				}
			}
			for _, client := range deadClients {
				close(client.send)
				delete(h.clients, client)
			}
			h.mu.Unlock()
		}
	}
}

// Broadcast 向所有连接的客户端广播消息（不区分用户）
func (h *WSHub) Broadcast(message []byte) {
	h.broadcast <- message
}

// BroadcastToUser 向指定用户的所有 WebSocket 连接推送消息
func (h *WSHub) BroadcastToUser(userID string, message []byte) {
	h.broadcastScoped <- scopedMessage{userID: userID, data: message}
}

// readPump 从 WebSocket 连接读取消息（用于心跳和客户端主动消息）
func (c *WSClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(4096)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// writePump 向 WebSocket 连接写入消息（含写超时防 goroutine 泄漏）
func (c *WSClient) writePump() {
	ticker := time.NewTicker(30 * time.Second) // 心跳间隔
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			// 发送心跳 ping（含写超时）
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// HandleWebSocket 处理 WebSocket 连接升级
// GET /ws?token=xxx
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// JWT 鉴权：从 query 参数获取 token 并验证
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少 token 参数"})
		return
	}

	userID, err := middleware.ValidateToken(token, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 无效或已过期"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败: %v", err)
		return
	}

	log.Printf("WebSocket 已认证连接: userID=%s", userID)

	client := &WSClient{
		hub:    h.Hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
	}

	h.Hub.register <- client

	// 启动读写 goroutine
	go client.writePump()
	go client.readPump()
}
