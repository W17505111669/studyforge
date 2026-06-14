package handler

import (
	"net/http"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ==================== 对话会话 CRUD ====================

// ListConversations 获取用户的所有对话（按更新时间倒序，含消息数量，支持分页）
// GET /api/conversations?limit=50&offset=0
func (h *Handler) ListConversations(c *gin.Context) {
	userID := c.GetString("userID")
	limit, offset := parsePagination(c)

	type Result struct {
		model.Conversation
		MessageCount int `json:"message_count"`
	}

	var total int64
	h.DB.Model(&model.Conversation{}).Where("user_id = ?", userID).Count(&total)

	var results []Result
	h.DB.Raw(`
		SELECT c.*, COUNT(m.id) as message_count
		FROM conversations c
		LEFT JOIN chat_messages m ON m.conversation_id = c.id
		WHERE c.user_id = ?
		GROUP BY c.id
		ORDER BY c.updated_at DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset).Scan(&results)

	c.JSON(http.StatusOK, gin.H{"data": results, "total": total, "limit": limit, "offset": offset})
}

// GetConversation 获取单个对话及其所有消息
// GET /api/conversations/:id
func (h *Handler) GetConversation(c *gin.Context) {
	userID := c.GetString("userID")
	convID := c.Param("id")

	var conv model.Conversation
	if err := h.DB.Where("id = ? AND user_id = ?", convID, userID).First(&conv).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "对话不存在"})
		return
	}

	// 加载消息（按时间顺序）
	var messages []model.ChatMessage
	h.DB.Where("conversation_id = ?", convID).Order("created_at ASC").Find(&messages)

	c.JSON(http.StatusOK, gin.H{
		"id":         conv.ID,
		"title":      conv.Title,
		"created_at": conv.CreatedAt,
		"updated_at": conv.UpdatedAt,
		"messages":   messages,
	})
}

// CreateConversationRequest 创建对话请求
type CreateConversationRequest struct {
	Title string `json:"title"`
}

// CreateConversation 创建一个新的对话会话
// POST /api/conversations
func (h *Handler) CreateConversation(c *gin.Context) {
	userID := c.GetString("userID")

	var req CreateConversationRequest
	// 允许空 body（title 可选）
	_ = c.ShouldBindJSON(&req)

	title := req.Title
	if title == "" {
		title = "新对话"
	}

	conv := model.Conversation{
		UserID: userID,
		Title:  title,
	}
	if err := h.DB.Create(&conv).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建对话失败"})
		return
	}

	c.JSON(http.StatusCreated, conv)
}

// UpdateConversationRequest 更新对话请求
type UpdateConversationRequest struct {
	Title string `json:"title" binding:"required"`
}

// UpdateConversation 更新对话标题
// PUT /api/conversations/:id
func (h *Handler) UpdateConversation(c *gin.Context) {
	userID := c.GetString("userID")
	convID := c.Param("id")

	var req UpdateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "标题不能为空"})
		return
	}

	var conv model.Conversation
	if err := h.DB.Where("id = ? AND user_id = ?", convID, userID).First(&conv).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "对话不存在"})
		return
	}

	conv.Title = req.Title
	h.DB.Save(&conv)

	c.JSON(http.StatusOK, conv)
}

// DeleteConversation 删除对话及其所有消息（事务保护）
// DELETE /api/conversations/:id
func (h *Handler) DeleteConversation(c *gin.Context) {
	userID := c.GetString("userID")
	convID := c.Param("id")

	var conv model.Conversation
	if err := h.DB.Where("id = ? AND user_id = ?", convID, userID).First(&conv).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "对话不存在"})
		return
	}

	// 事务内级联删除：消息 + 对话本体
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("conversation_id = ?", convID).Delete(&model.ChatMessage{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&conv).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除对话失败: " + err.Error()})
		return
	}

	// 清理内存中的会话记忆（key 是 convID，不是 userID）
	h.EvictConversationMemory(convID)

	c.JSON(http.StatusOK, gin.H{"message": "对话已删除"})
}
