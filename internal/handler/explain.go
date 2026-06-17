package handler

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== AI 概念解释器 ====================

// explainSystemPrompt LLM 系统提示词
const explainSystemPrompt = `你是一位优秀的学习导师，擅长用通俗易懂的方式解释复杂概念。
对于用户提出的概念，请按以下 JSON 格式返回解释（不要包含 markdown 代码块标记）：

{
  "explanation": "用2-3段通俗文字解释这个概念的核心含义，避免过多专业术语",
  "analogy": "用1-2个生活中的类比来帮助理解，让抽象概念变得具体",
  "example": "给出1-2个具体的应用例子或使用场景",
  "related_concepts": ["关联概念1", "关联概念2", "关联概念3"]
}

要求：
- 解释清晰简洁，适合初学者理解
- 类比要贴近日常生活
- 例子要具体可感知
- 关联概念列出 3-5 个相关的知识点名称
- 严格返回 JSON 格式，不要包含其他文字`

// ExplainConcept 解释概念
// POST /api/explain
func (h *Handler) ExplainConcept(c *gin.Context) {
	userID := c.GetString("userID")

	var req model.ExplainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供要解释的概念（concept 字段，1-500 字）"})
		return
	}

	// 计算概念 hash（用于缓存匹配）
	conceptHash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(req.Concept)))))

	// 检查缓存（同用户 + 相同概念 hash）
	var cached model.ExplainCache
	if err := h.DB.Where("user_id = ? AND concept_hash = ?", userID, conceptHash).First(&cached).Error; err == nil {
		// 缓存命中
		var relatedConcepts []string
		if cached.RelatedConcepts != "" {
			json.Unmarshal([]byte(cached.RelatedConcepts), &relatedConcepts)
		}
		c.JSON(http.StatusOK, model.ExplainResponse{
			ID:              cached.ID,
			Concept:         cached.Concept,
			Explanation:     cached.Explanation,
			Analogy:         cached.Analogy,
			Example:         cached.Example,
			RelatedConcepts: relatedConcepts,
			Cached:          true,
			CreatedAt:       cached.CreatedAt,
		})
		return
	}

	// 构建用户消息（包含可选上下文）
	userMessage := fmt.Sprintf("请解释以下概念：%s", req.Concept)
	if req.Context != "" {
		userMessage += fmt.Sprintf("\n\n学习上下文：%s", req.Context)
	}

	// 调用 LLM 生成解释
	output, inputTokens, outputTokens, err := h.LLM.ChatWithUsage(
		c.Request.Context(),
		explainSystemPrompt,
		userMessage,
	)
	if err != nil {
		log.Printf("概念解释 LLM 调用失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "概念解释生成失败，请稍后重试"})
		return
	}

	// 解析 LLM 输出（清理 markdown 代码块包裹）
	cleanOutput := extractExplainJSON(output)
	var parsed struct {
		Explanation     string   `json:"explanation"`
		Analogy         string   `json:"analogy"`
		Example         string   `json:"example"`
		RelatedConcepts []string `json:"related_concepts"`
	}
	if err := json.Unmarshal([]byte(cleanOutput), &parsed); err != nil {
		log.Printf("概念解释 JSON 解析失败: %v, 原始输出前200字: %.200s", err, cleanOutput)
		// 降级：将原始文本作为解释返回
		parsed.Explanation = output
		parsed.Analogy = ""
		parsed.Example = ""
		parsed.RelatedConcepts = []string{}
	}

	// 序列化关联概念
	relatedJSON, _ := json.Marshal(parsed.RelatedConcepts)

	// 存入缓存
	cacheEntry := model.ExplainCache{
		UserID:          userID,
		Concept:         req.Concept,
		ConceptHash:     conceptHash,
		Explanation:     parsed.Explanation,
		Analogy:         parsed.Analogy,
		Example:         parsed.Example,
		RelatedConcepts: string(relatedJSON),
	}
	if err := h.DB.Create(&cacheEntry).Error; err != nil {
		log.Printf("保存概念解释缓存失败: %v", err)
	}

	// 记录 LLM Trace
	modelName := "unknown"
	if h.LLM != nil && h.LLM.ModelName != "" {
		modelName = h.LLM.ModelName
	}
	trace := model.LLMTrace{
		UserID:        userID,
		AgentName:     "Explainer",
		Model:         modelName,
		InputTokens:   inputTokens,
		OutputTokens:  outputTokens,
		TotalTokens:   inputTokens + outputTokens,
		Status:        "success",
		PromptSummary: "explain concept: " + req.Concept,
	}
	if err := h.DB.Create(&trace).Error; err != nil {
		log.Printf("保存概念解释 Trace 失败: %v", err)
	}

	c.JSON(http.StatusOK, model.ExplainResponse{
		ID:              cacheEntry.ID,
		Concept:         req.Concept,
		Explanation:     parsed.Explanation,
		Analogy:         parsed.Analogy,
		Example:         parsed.Example,
		RelatedConcepts: parsed.RelatedConcepts,
		Cached:          false,
		CreatedAt:       cacheEntry.CreatedAt,
	})
}

// GetExplainHistory 获取解释历史
// GET /api/explain/history
func (h *Handler) GetExplainHistory(c *gin.Context) {
	userID := c.GetString("userID")
	limit, offset := parsePagination(c)
	q := c.Query("q")

	query := h.DB.Where("user_id = ?", userID)
	if q != "" {
		query = query.Where("concept LIKE ?", "%"+q+"%")
	}

	var total int64
	if err := query.Model(&model.ExplainCache{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "统计解释历史失败"})
		return
	}

	var items []model.ExplainCache
	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取解释历史失败"})
		return
	}

	// 转换为响应格式
	results := make([]model.ExplainResponse, 0, len(items))
	for _, item := range items {
		var relatedConcepts []string
		if item.RelatedConcepts != "" {
			json.Unmarshal([]byte(item.RelatedConcepts), &relatedConcepts)
		}
		results = append(results, model.ExplainResponse{
			ID:              item.ID,
			Concept:         item.Concept,
			Explanation:     item.Explanation,
			Analogy:         item.Analogy,
			Example:         item.Example,
			RelatedConcepts: relatedConcepts,
			Cached:          true,
			CreatedAt:       item.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   results,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// DeleteExplainCache 删除单条解释缓存
// DELETE /api/explain/:id
func (h *Handler) DeleteExplainCache(c *gin.Context) {
	userID := c.GetString("userID")
	explainID := c.Param("id")

	result := h.DB.Where("id = ? AND user_id = ?", explainID, userID).Delete(&model.ExplainCache{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除解释记录失败"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "解释记录不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// extractExplainJSON 从 LLM 输出中提取 JSON（去除 markdown 代码块包裹）
func extractExplainJSON(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return s
	}

	// 去掉 ```json 或 ``` 开头
	if strings.HasPrefix(s, "```") {
		nl := strings.Index(s, "\n")
		if nl >= 0 {
			s = s[nl+1:]
		}
		if idx := strings.LastIndex(s, "\n```"); idx >= 0 {
			s = s[:idx]
		} else if strings.HasSuffix(s, "```") {
			s = s[:len(s)-3]
		}
		s = strings.TrimSpace(s)
	}

	// 如果仍不以 { 或 [ 开头，尝试找到第一个 JSON 对象/数组
	if len(s) > 0 && s[0] != '{' && s[0] != '[' {
		for i, ch := range s {
			if ch == '{' || ch == '[' {
				s = s[i:]
				break
			}
		}
	}

	// 去掉末尾的非 JSON 尾部
	if len(s) > 0 {
		lastBrace := strings.LastIndexByte(s, '}')
		lastBracket := strings.LastIndexByte(s, ']')
		end := lastBrace
		if lastBracket > end {
			end = lastBracket
		}
		if end >= 0 {
			s = s[:end+1]
		}
	}

	return s
}
