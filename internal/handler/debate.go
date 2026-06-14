package handler

import (
	"log"
	"net/http"

	"studyforge/internal/agent"
	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 多 Agent 辩论 ====================

// DebateRequest 辩论请求
type DebateRequest struct {
	Concept string `json:"concept" binding:"required,min=1,max=500"`
}

// StartDebate 发起多 Agent 辩论
// POST /api/debate
func (h *Handler) StartDebate(c *gin.Context) {
	userID := c.GetString("userID")

	var req DebateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供辩论主题（concept 字段，1-500 字）"})
		return
	}

	// 创建辩论编排器
	debater := agent.NewDebateOrchestrator(h.LLM)

	// 执行辩论
	output, inputTokens, outputTokens, err := debater.RunDebate(c.Request.Context(), req.Concept)
	if err != nil {
		log.Printf("辩论失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "辩论生成失败，请稍后重试"})
		return
	}

	// 记录 LLM Trace
	modelName := "unknown"
	if h.LLM != nil && h.LLM.ModelName != "" {
		modelName = h.LLM.ModelName
	}

	trace := model.LLMTrace{
		UserID:        userID,
		AgentName:     "Debate",
		Model:         modelName,
		InputTokens:   inputTokens,
		OutputTokens:  outputTokens,
		TotalTokens:   inputTokens + outputTokens,
		Status:        "success",
		PromptSummary: "multi-agent debate: " + req.Concept,
	}
	if err := h.DB.Create(&trace).Error; err != nil {
		log.Printf("保存辩论 Trace 失败: %v", err)
	}

	c.JSON(http.StatusOK, output)
}
