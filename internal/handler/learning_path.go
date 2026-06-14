package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"studyforge/internal/agent"
	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 学习路径规划 ====================

// pathCacheEntry 学习路径缓存条目
type pathCacheEntry struct {
	data      gin.H   // 缓存的响应数据
	createdAt time.Time
}

const pathCacheTTL = 30 * time.Minute // 缓存有效期

// LearningPathStep 前端友好的学习路径步骤
type LearningPathStep struct {
	Index            int      `json:"index"`              // 步骤序号（0-based）
	Title            string   `json:"title"`              // 步骤标题
	Description      string   `json:"description"`        // 步骤说明
	EstimatedMinutes int      `json:"estimated_minutes"`  // 预估时长（分钟）
	MaterialIDs      []string `json:"material_ids"`       // 关联材料 ID
	Prerequisites    []int    `json:"prerequisites"`      // 前置依赖步骤索引
	Difficulty       string   `json:"difficulty"`         // easy / medium / hard
	// 前端展示用的附加字段
	MaterialTitles []string `json:"material_titles"` // 关联材料标题
	CardCount      int64    `json:"card_count"`      // 关联材料的卡片总数
	ReviewedCount  int64    `json:"reviewed_count"`  // 已复习的卡片数
}

// GetLearningPath 获取学习路径规划（基于用户所有材料生成学习路线图）
// GET /api/learning-path?force=true（force=true 跳过缓存重新生成）
func (h *Handler) GetLearningPath(c *gin.Context) {
	userID := c.GetString("userID")
	forceRefresh := c.Query("force") == "true"

	// ===== 1. 检查缓存 =====
	if !forceRefresh {
		h.pathCacheMu.RLock()
		if entry, ok := h.pathCache[userID]; ok && time.Since(entry.createdAt) < pathCacheTTL {
			h.pathCacheMu.RUnlock()
			c.JSON(http.StatusOK, entry.data)
			return
		}
		h.pathCacheMu.RUnlock()
	}

	// 查询用户所有已完成分析的材料
	var materials []model.Material
	if err := h.DB.Where("user_id = ? AND status = ?", userID, "completed").
		Order("created_at ASC").
		Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询材料失败"})
		return
	}

	if len(materials) == 0 {
		resp := gin.H{
			"overview":       "暂无已分析的学习材料，请先上传并分析材料后再生成学习路径。",
			"total_hours":    0,
			"ordered_steps":  []LearningPathStep{},
			"material_count": 0,
		}
		c.JSON(http.StatusOK, resp)
		return
	}

	// 构建材料摘要上下文
	summary := buildPathSummary(materials)

	// ===== 2. 调用 PathPlanner Agent（带 90s 超时） =====
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()

	pathPlanner := agent.NewPathPlannerAgent(h.LLM)
	output, inputTokens, outputTokens, err := pathPlanner.Generate(ctx, summary)
	if err != nil {
		log.Printf("PathPlanner Agent 失败: %v", err)
		// 区分超时和其他错误
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "AI 生成超时，请稍后重试"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成学习路径失败，请稍后重试"})
		}
		return
	}

	// 记录 LLM Trace
	h.savePathTrace(userID, inputTokens, outputTokens, nil)

	// 构建材料标题和统计映射
	materialMap := make(map[string]*model.Material)
	for i := range materials {
		materialMap[materials[i].ID] = &materials[i]
	}

	// 将 Agent 输出转为前端友好格式
	steps := make([]LearningPathStep, 0, len(output.Steps))
	for i, step := range output.Steps {
		ls := LearningPathStep{
			Index:            i,
			Title:            step.Title,
			Description:      step.Description,
			EstimatedMinutes: step.EstimatedMinutes,
			MaterialIDs:      step.MaterialIDs,
			Prerequisites:    step.Prerequisites,
			Difficulty:       step.Difficulty,
		}

		// 补充材料标题和统计
		for _, mid := range step.MaterialIDs {
			if mat, ok := materialMap[mid]; ok {
				ls.MaterialTitles = append(ls.MaterialTitles, mat.Title)

				// 统计该材料的卡片数和已复习数
				var cardCount, reviewedCount int64
				h.DB.Model(&model.Card{}).Where("user_id = ? AND material_id = ?", userID, mid).Count(&cardCount)
				h.DB.Model(&model.Card{}).Where("user_id = ? AND material_id = ? AND review_count > 0", userID, mid).Count(&reviewedCount)
				ls.CardCount += cardCount
				ls.ReviewedCount += reviewedCount
			}
		}

		if ls.MaterialTitles == nil {
			ls.MaterialTitles = []string{}
		}
		if ls.Prerequisites == nil {
			ls.Prerequisites = []int{}
		}
		if ls.MaterialIDs == nil {
			ls.MaterialIDs = []string{}
		}

		steps = append(steps, ls)
	}

	resp := gin.H{
		"overview":       output.Overview,
		"total_hours":    output.TotalHours,
		"ordered_steps":  steps,
		"material_count": len(materials),
	}

	// ===== 3. 缓存结果 =====
	h.pathCacheMu.Lock()
	h.pathCache[userID] = &pathCacheEntry{data: resp, createdAt: time.Now()}
	h.pathCacheMu.Unlock()

	c.JSON(http.StatusOK, resp)
}

// buildPathSummary 构建给 PathPlanner Agent 的材料摘要文本
func buildPathSummary(materials []model.Material) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("以下是用户的 %d 份学习材料及其知识点摘要，请据此规划学习路径：\n\n", len(materials)))

	for i, mat := range materials {
		sb.WriteString(fmt.Sprintf("=== 材料 %d ===\n", i+1))
		sb.WriteString(fmt.Sprintf("ID: %s\n", mat.ID))
		sb.WriteString(fmt.Sprintf("标题: %s\n", mat.Title))
		sb.WriteString(fmt.Sprintf("上传时间: %s\n", mat.CreatedAt.Format("2006-01-02")))

		// 解析 analysis_data 获取知识点摘要
		if mat.AnalysisData != "" {
			var analysis agent.AnalystOutput
			if err := json.Unmarshal([]byte(mat.AnalysisData), &analysis); err == nil {
				if analysis.Summary != "" {
					sb.WriteString(fmt.Sprintf("摘要: %s\n", analysis.Summary))
				}
				if len(analysis.KeyPoints) > 0 {
					sb.WriteString("核心知识点:\n")
					for j, kp := range analysis.KeyPoints {
						if j >= 8 {
							sb.WriteString(fmt.Sprintf("  ...还有 %d 个知识点\n", len(analysis.KeyPoints)-8))
							break
						}
						sb.WriteString(fmt.Sprintf("  - %s（%s）\n", kp.Concept, kp.Difficulty))
					}
				}
				if len(analysis.Relationships) > 0 {
					sb.WriteString("概念关系:\n")
					for j, rel := range analysis.Relationships {
						if j >= 5 {
							sb.WriteString(fmt.Sprintf("  ...还有 %d 条关系\n", len(analysis.Relationships)-5))
							break
						}
						sb.WriteString(fmt.Sprintf("  - %s → %s: %s\n", rel.From, rel.To, rel.Type))
					}
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// savePathTrace 保存 PathPlanner 的 LLM 调用追踪
func (h *Handler) savePathTrace(userID string, inputTokens, outputTokens int, err error) {
	status := "success"
	errMsg := ""
	if err != nil {
		status = "error"
		errMsg = err.Error()
	}

	modelName := "unknown"
	if h.LLM != nil && h.LLM.ModelName != "" {
		modelName = h.LLM.ModelName
	}

	trace := model.LLMTrace{
		UserID:        userID,
		AgentName:     "PathPlanner",
		Model:         modelName,
		InputTokens:   inputTokens,
		OutputTokens:  outputTokens,
		TotalTokens:   inputTokens + outputTokens,
		Status:        status,
		ErrorMessage:  errMsg,
		PromptSummary: "learning path generation",
	}
	if err := h.DB.Create(&trace).Error; err != nil {
		log.Printf("保存 PathPlanner Trace 失败: %v", err)
	}
}
