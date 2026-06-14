package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ==================== 材料管理 ====================

// UploadMaterial 上传学习材料
// POST /api/materials
func (h *Handler) UploadMaterial(c *gin.Context) {
	userID := c.GetString("userID")

	var req model.UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证：text 类型必须有 content，url 类型必须有 source_url
	if req.ContentType == "text" && req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文本类型必须提供 content"})
		return
	}
	if req.ContentType == "url" && req.SourceURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL 类型必须提供 source_url"})
		return
	}

	material := model.Material{
		UserID:      userID,
		Title:       req.Title,
		ContentType: req.ContentType,
		Content:     req.Content,
		SourceURL:   req.SourceURL,
		Tags:        req.Tags,
		Status:      "pending",
	}

	if err := h.DB.Create(&material).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建材料失败"})
		return
	}

	// 自动 RAG 索引（后台异步执行，不阻塞响应）
	if h.VectorStore != nil && material.Content != "" {
		go func() {
			indexCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			if err := h.VectorStore.IndexMaterial(indexCtx, material.ID, userID, material.Content, h.chunkSize, h.chunkOverlap); err != nil {
				fmt.Printf("材料 %s RAG 索引失败: %v\n", material.ID, err)
			}
		}()
	}

	c.JSON(http.StatusCreated, material)
}

// ListMaterials 获取用户的所有材料（支持分页 + 标签过滤）
// GET /api/materials?limit=20&offset=0&tag=Go
func (h *Handler) ListMaterials(c *gin.Context) {
	userID := c.GetString("userID")
	limit, offset := parsePagination(c)
	tagFilter := strings.TrimSpace(c.Query("tag"))

	query := h.DB.Model(&model.Material{}).Where("user_id = ?", userID)
	if tagFilter != "" {
		// 模糊匹配标签（逗号分隔字段中搜索）
		query = query.Where("tags LIKE ?", "%"+tagFilter+"%")
	}

	var total int64
	query.Count(&total)

	var materials []model.Material
	if err := query.Select("id, user_id, title, content_type, source_url, status, tags, analyzed_at, created_at, updated_at").
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询材料失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": materials, "total": total, "limit": limit, "offset": offset})
}

// GetMaterial 获取单个材料详情
// GET /api/materials/:id
func (h *Handler) GetMaterial(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")

	var material model.Material
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).
		Preload("Cards").
		Preload("Quizzes").
		First(&material).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "材料不存在"})
		return
	}

	c.JSON(http.StatusOK, material)
}

// DeleteMaterial 删除材料
// DELETE /api/materials/:id
func (h *Handler) DeleteMaterial(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")

	var material model.Material
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&material).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "材料不存在"})
		return
	}

	// 事务内级联删除：材料 + 卡片 + 题目
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Material{}).Error; err != nil {
			return err
		}
		if err := tx.Where("material_id = ?", id).Delete(&model.Card{}).Error; err != nil {
			return err
		}
		if err := tx.Where("material_id = ?", id).Delete(&model.Quiz{}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Printf("删除材料失败 (id=%s): %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败，请稍后重试"})
		return
	}

	// 清理 RAG 向量数据
	if h.VectorStore != nil {
		h.VectorStore.DeleteMaterialVectors(id)
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// AnalyzeMaterial 触发 Agent 并发分析材料
// POST /api/materials/:id/analyze
func (h *Handler) AnalyzeMaterial(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")

	var material model.Material
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&material).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "材料不存在"})
		return
	}

	if material.Status == "analyzing" {
		c.JSON(http.StatusConflict, gin.H{"error": "材料正在分析中，请等待完成"})
		return
	}

	// 并发守卫：防止同一材料被重复触发分析 goroutine
	if _, alreadyRunning := h.activeAnalyses.LoadOrStore(material.ID, true); alreadyRunning {
		c.JSON(http.StatusConflict, gin.H{"error": "分析任务已在运行中"})
		return
	}

	// 更新状态为 analyzing
	if err := h.DB.Model(&material).Update("status", "analyzing").Error; err != nil {
		h.activeAnalyses.Delete(material.ID) // 清除并发守卫
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新材料状态失败"})
		return
	}

	// 初始化分析进度追踪（用于 WebSocket 断线时的轮询 fallback）
	h.analysisMu.Lock()
	h.analysisProgress[material.ID] = &MaterialAnalysisProgress{
		Agents: []AgentProgress{
			{Name: "Analyst", Status: "waiting"},
			{Name: "QuizMaster", Status: "waiting"},
			{Name: "CardMaker", Status: "waiting"},
			{Name: "MapBuilder", Status: "waiting"},
		},
		Started: time.Now(),
	}
	h.analysisMu.Unlock()

	// 启动 goroutine 执行 Agent 并发分析（不阻塞 HTTP 响应）
	go func() {
		// 确保无论如何都清除并发守卫
		defer h.activeAnalyses.Delete(material.ID)

		ctx := context.Background()
		err := h.Orchestrator.ProcessMaterial(ctx, material, userID)

		// 分析完成，更新最终进度并清理
		h.analysisMu.Lock()
		if prog, ok := h.analysisProgress[material.ID]; ok {
			if err != nil {
				prog.Error = err.Error()
			}
			// 将所有未完成的 Agent 标记为最终状态
			for i, a := range prog.Agents {
				if a.Status == "waiting" || a.Status == "running" {
					if err != nil {
						prog.Agents[i].Status = "error"
					} else {
						prog.Agents[i].Status = "done"
					}
				}
			}
		}
		h.analysisMu.Unlock()

		if err != nil {
			// 分析失败，通过 WebSocket 通知该用户
			errMsg, _ := json.Marshal(WSMessage{
				Type:    "error",
				Content: "分析失败: " + err.Error(),
			})
			h.Hub.BroadcastToUser(userID, errMsg)
		}

		// 分析完成通知（只发给触发分析的用户）
		doneMsg, _ := json.Marshal(WSMessage{
			Type:    "analysis_complete",
			Content: gin.H{"material_id": material.ID},
		})
		h.Hub.BroadcastToUser(userID, doneMsg)

		// 延迟清理进度数据（给轮询端点留足够时间获取最终状态）
		go func() {
			time.Sleep(30 * time.Second)
			h.analysisMu.Lock()
			delete(h.analysisProgress, material.ID)
			h.analysisMu.Unlock()
		}()
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "分析任务已启动，结果将通过 WebSocket 实时推送",
		"material_id": material.ID,
	})
}

// BatchAnalyzeMaterials 批量分析材料（并发处理 + WebSocket 逐个推送进度）
// POST /api/materials/batch-analyze
// body: { "ids": ["uuid1", "uuid2", ...] }
func (h *Handler) BatchAnalyzeMaterials(c *gin.Context) {
	userID := c.GetString("userID")

	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供要分析的材料 ID 列表"})
		return
	}

	if len(req.IDs) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "单次最多批量分析 20 个材料"})
		return
	}

	// 验证所有材料存在且属于当前用户
	var materials []model.Material
	if err := h.DB.Where("id IN ? AND user_id = ?", req.IDs, userID).Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询材料失败"})
		return
	}

	if len(materials) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到任何可分析的材料"})
		return
	}

	// 过滤掉正在分析中的材料
	var validMaterials []model.Material
	var skipped []string
	for _, m := range materials {
		if m.Status == "analyzing" {
			skipped = append(skipped, m.ID)
			continue
		}
		if _, alreadyRunning := h.activeAnalyses.LoadOrStore(m.ID, true); alreadyRunning {
			skipped = append(skipped, m.ID)
			continue
		}
		validMaterials = append(validMaterials, m)
	}

	if len(validMaterials) == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "所有材料都已在分析中"})
		return
	}

	total := len(validMaterials)

	// 更新所有有效材料的状态为 analyzing
	for i := range validMaterials {
		h.DB.Model(&validMaterials[i]).Update("status", "analyzing")
		// 初始化进度追踪
		h.analysisMu.Lock()
		h.analysisProgress[validMaterials[i].ID] = &MaterialAnalysisProgress{
			Agents: []AgentProgress{
				{Name: "Analyst", Status: "waiting"},
				{Name: "QuizMaster", Status: "waiting"},
				{Name: "CardMaker", Status: "waiting"},
				{Name: "MapBuilder", Status: "waiting"},
			},
			Started: time.Now(),
		}
		h.analysisMu.Unlock()
	}

	// 启动 goroutine 并发分析（信号量限制最多 2 个并发）
	go func() {
		sem := make(chan struct{}, 2) // 并发上限
		var wg sync.WaitGroup

		for i := range validMaterials {
			wg.Add(1)
			sem <- struct{}{} // 获取信号量

			go func(mat model.Material) {
				defer wg.Done()
				defer func() { <-sem }() // 释放信号量
				defer h.activeAnalyses.Delete(mat.ID)

				ctx := context.Background()
				err := h.Orchestrator.ProcessMaterial(ctx, mat, userID)

				// 更新进度
				h.analysisMu.Lock()
				if prog, ok := h.analysisProgress[mat.ID]; ok {
					if err != nil {
						prog.Error = err.Error()
					}
					for j, a := range prog.Agents {
						if a.Status == "waiting" || a.Status == "running" {
							if err != nil {
								prog.Agents[j].Status = "error"
							} else {
								prog.Agents[j].Status = "done"
							}
						}
					}
				}
				h.analysisMu.Unlock()

				// 通过 WebSocket 推送单个材料分析完成
				status := "completed"
				errMsg := ""
				if err != nil {
					status = "failed"
					errMsg = err.Error()
				}
				progressMsg, _ := json.Marshal(WSMessage{
					Type: "batch_analyze_progress",
					Content: gin.H{
						"material_id": mat.ID,
						"title":       mat.Title,
						"status":      status,
						"error":       errMsg,
						"completed":   0, // 前端用计数器自行累加
						"total":       total,
					},
				})
				h.Hub.BroadcastToUser(userID, progressMsg)

				// 延迟清理进度数据
				go func(matID string) {
					time.Sleep(30 * time.Second)
					h.analysisMu.Lock()
					delete(h.analysisProgress, matID)
					h.analysisMu.Unlock()
				}(mat.ID)
			}(validMaterials[i])
		}

		wg.Wait()

		// 全部完成通知
		doneMsg, _ := json.Marshal(WSMessage{
			Type: "batch_analyze_complete",
			Content: gin.H{
				"total":   total,
				"skipped": skipped,
			},
		})
		h.Hub.BroadcastToUser(userID, doneMsg)
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message":    fmt.Sprintf("批量分析已启动，共 %d 个材料", total),
		"total":      total,
		"skipped":    skipped,
		"started":    len(validMaterials),
	})
}

// BatchDeleteMaterials 批量删除材料（逐个事务删除含关联数据 + RAG 清理）
// DELETE /api/materials/batch
// body: { "ids": ["uuid1", "uuid2", ...] }
func (h *Handler) BatchDeleteMaterials(c *gin.Context) {
	userID := c.GetString("userID")

	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供要删除的材料 ID 列表"})
		return
	}

	if len(req.IDs) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "单次最多批量删除 50 个材料"})
		return
	}

	// 验证材料存在且属于当前用户
	var materials []model.Material
	if err := h.DB.Where("id IN ? AND user_id = ?", req.IDs, userID).Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询材料失败"})
		return
	}

	if len(materials) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到任何可删除的材料"})
		return
	}

	successCount := 0
	var failedIDs []string

	for _, mat := range materials {
		err := h.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("id = ? AND user_id = ?", mat.ID, userID).Delete(&model.Material{}).Error; err != nil {
				return err
			}
			if err := tx.Where("material_id = ?", mat.ID).Delete(&model.Card{}).Error; err != nil {
				return err
			}
			if err := tx.Where("material_id = ?", mat.ID).Delete(&model.Quiz{}).Error; err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			log.Printf("批量删除材料失败 (id=%s): %v", mat.ID, err)
			failedIDs = append(failedIDs, mat.ID)
			continue
		}
		// 清理 RAG 向量
		if h.VectorStore != nil {
			h.VectorStore.DeleteMaterialVectors(mat.ID)
		}
		successCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("成功删除 %d 个材料", successCount),
		"deleted":  successCount,
		"total":    len(materials),
		"failed":   failedIDs,
	})
}

// UploadFile 上传文件并提取文本内容
// POST /api/materials/upload
func (h *Handler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的文件"})
		return
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".pdf": true, ".docx": true, ".md": true, ".txt": true, ".markdown": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("不支持的文件格式: %s（支持 .pdf .docx .md .txt）", ext)})
		return
	}

	// 限制文件大小：20MB
	if file.Size > 20*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小不能超过 20MB"})
		return
	}

	// 保存上传文件到临时目录（净化文件名防止路径穿越）
	os.MkdirAll("data/uploads", 0755)
	safeFilename := filepath.Base(file.Filename)
	tmpPath := filepath.Join("data/uploads", safeFilename)
	if err := c.SaveUploadedFile(file, tmpPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败"})
		return
	}
	defer os.Remove(tmpPath) // 处理完后删除临时文件

	// 打开文件用于文本提取
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件打开失败"})
		return
	}
	defer f.Close()

	// 提取文本
	text, err := extractTextFromFile(f, tmpPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 用安全文件名生成默认标题（去掉扩展名）
	title := strings.TrimSuffix(safeFilename, ext)

	c.JSON(http.StatusOK, gin.H{
		"text":     text,
		"title":    title,
		"filename": safeFilename,
		"size":     file.Size,
	})
}

// GetMaterialStatus 获取材料分析状态（轮询端点，WebSocket 断线时的 fallback）
// GET /api/materials/:id/status
func (h *Handler) GetMaterialStatus(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")

	var material model.Material
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).First(&material).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "材料不存在"})
		return
	}

	response := gin.H{
		"status":        material.Status,
		"material_id":   material.ID,
		"completed":     material.Status != "analyzing",
	}

	// 如果正在分析中，返回各 Agent 的实时进度
	if material.Status == "analyzing" {
		h.analysisMu.RLock()
		if prog, ok := h.analysisProgress[id]; ok {
			completedCount := 0
			for _, a := range prog.Agents {
				if a.Status == "done" || a.Status == "error" {
					completedCount++
				}
			}
			response["agents"] = prog.Agents
			response["completed_count"] = completedCount
			response["total_agents"] = len(prog.Agents)
			response["error"] = prog.Error
		} else {
			// 状态是 analyzing 但无进度数据（可能是服务重启），返回默认值
			response["agents"] = []AgentProgress{
				{Name: "Analyst", Status: "running"},
				{Name: "QuizMaster", Status: "running"},
				{Name: "CardMaker", Status: "running"},
				{Name: "MapBuilder", Status: "running"},
			}
			response["completed_count"] = 0
			response["total_agents"] = 4
		}
		h.analysisMu.RUnlock()
	} else {
		// 非分析中状态，返回已完成的默认 Agent 列表
		response["agents"] = []AgentProgress{
			{Name: "Analyst", Status: "done"},
			{Name: "QuizMaster", Status: "done"},
			{Name: "CardMaker", Status: "done"},
			{Name: "MapBuilder", Status: "done"},
		}
		response["completed_count"] = 4
		response["total_agents"] = 4
	}

	c.JSON(http.StatusOK, response)
}

// GetTags 获取用户所有已用标签及计数
// GET /api/tags
func (h *Handler) GetTags(c *gin.Context) {
	userID := c.GetString("userID")

	var materials []model.Material
	h.DB.Select("tags").Where("user_id = ? AND tags != ''", userID).Find(&materials)

	tagCount := map[string]int{}
	for _, m := range materials {
		parts := strings.Split(m.Tags, ",")
		seen := map[string]bool{}
		for _, t := range parts {
			tag := strings.TrimSpace(t)
			if tag == "" || seen[tag] {
				continue
			}
			seen[tag] = true
			tagCount[tag]++
		}
	}

	type TagItem struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	var tags []TagItem
	for name, count := range tagCount {
		tags = append(tags, TagItem{Name: name, Count: count})
	}
	// 按使用次数降序排序
	for i := 0; i < len(tags); i++ {
		for j := i + 1; j < len(tags); j++ {
			if tags[j].Count > tags[i].Count {
				tags[i], tags[j] = tags[j], tags[i]
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

// parsePagination 解析分页参数，返回 limit 和 offset
// 默认 limit=20, offset=0，limit 上限 200
func parsePagination(c *gin.Context) (int, int) {
	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if limit > 200 {
		limit = 200
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	return limit, offset
}
