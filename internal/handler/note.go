package handler

import (
	"net/http"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 笔记 CRUD ====================

// CreateNote 创建笔记
func (h *Handler) CreateNote(c *gin.Context) {
	userID := c.GetString("userID")

	var req struct {
		Title      string `json:"title"`
		Content    string `json:"content"`
		MaterialID string `json:"material_id"`
		CardID     string `json:"card_id"`
		FolderID   string `json:"folder_id"`
		Pinned     bool   `json:"pinned"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	title := req.Title
	if title == "" {
		title = "无标题笔记"
	}

	note := model.Note{
		UserID:     userID,
		Title:      title,
		Content:    req.Content,
		MaterialID: req.MaterialID,
		CardID:     req.CardID,
		FolderID:   req.FolderID,
		Pinned:     req.Pinned,
	}

	if err := h.DB.Create(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建笔记失败"})
		return
	}

	c.JSON(http.StatusCreated, note)
}

// ListNotes 获取笔记列表（支持文件夹过滤 + 分页）
func (h *Handler) ListNotes(c *gin.Context) {
	userID := c.GetString("userID")
	folderID := c.Query("folder_id")
	limit, offset := parsePagination(c)

	query := h.DB.Where("user_id = ?", userID)

	if folderID != "" {
		if folderID == "__none__" {
			// 专门筛选"无文件夹"的笔记
			query = query.Where("folder_id = '' OR folder_id IS NULL")
		} else {
			query = query.Where("folder_id = ?", folderID)
		}
	}

	var total int64
	if err := query.Model(&model.Note{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "统计笔记数量失败"})
		return
	}

	var notes []model.Note
	if err := query.
		Order("pinned DESC, updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取笔记失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   notes,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetNote 获取单条笔记详情
func (h *Handler) GetNote(c *gin.Context) {
	userID := c.GetString("userID")
	noteID := c.Param("id")

	var note model.Note
	if err := h.DB.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "笔记不存在"})
		return
	}

	c.JSON(http.StatusOK, note)
}

// UpdateNote 更新笔记
func (h *Handler) UpdateNote(c *gin.Context) {
	userID := c.GetString("userID")
	noteID := c.Param("id")

	var req struct {
		Title      *string `json:"title"`
		Content    *string `json:"content"`
		MaterialID *string `json:"material_id"`
		CardID     *string `json:"card_id"`
		FolderID   *string `json:"folder_id"`
		Pinned     *bool   `json:"pinned"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	var note model.Note
	if err := h.DB.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "笔记不存在"})
		return
	}

	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.MaterialID != nil {
		updates["material_id"] = *req.MaterialID
	}
	if req.CardID != nil {
		updates["card_id"] = *req.CardID
	}
	if req.FolderID != nil {
		updates["folder_id"] = *req.FolderID
	}
	if req.Pinned != nil {
		updates["pinned"] = *req.Pinned
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有可更新的字段"})
		return
	}

	if err := h.DB.Model(&note).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新笔记失败"})
		return
	}

	h.DB.Where("id = ? AND user_id = ?", noteID, userID).First(&note)
	c.JSON(http.StatusOK, note)
}

// DeleteNote 删除笔记
func (h *Handler) DeleteNote(c *gin.Context) {
	userID := c.GetString("userID")
	noteID := c.Param("id")

	result := h.DB.Where("id = ? AND user_id = ?", noteID, userID).Delete(&model.Note{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除笔记失败"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "笔记不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "笔记已删除"})
}

// SearchNotes 全文搜索笔记（标题 + 内容）
func (h *Handler) SearchNotes(c *gin.Context) {
	userID := c.GetString("userID")
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	var notes []model.Note
	if err := h.DB.Where("user_id = ? AND (title LIKE ? OR content LIKE ?)",
		userID, "%"+q+"%", "%"+q+"%").
		Order("pinned DESC, updated_at DESC").
		Limit(50).
		Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索笔记失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  notes,
		"total": len(notes),
	})
}

// ==================== 文件夹 ====================

// ListNoteFolders 获取文件夹列表（含笔记计数）
func (h *Handler) ListNoteFolders(c *gin.Context) {
	userID := c.GetString("userID")

	var folders []model.NoteFolder
	if err := h.DB.Where("user_id = ?", userID).Order("created_at ASC").Find(&folders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件夹失败"})
		return
	}

	// 批量查询每个文件夹的笔记数量
	result := make([]model.NoteFolderWithCount, 0, len(folders))
	for _, f := range folders {
		var count int64
		h.DB.Model(&model.Note{}).Where("user_id = ? AND folder_id = ?", userID, f.ID).Count(&count)
		result = append(result, model.NoteFolderWithCount{
			NoteFolder: f,
			NoteCount:  int(count),
		})
	}

	// 额外统计"无文件夹"的笔记数量
	var noFolderCount int64
	h.DB.Model(&model.Note{}).Where("user_id = ? AND (folder_id = '' OR folder_id IS NULL)", userID).Count(&noFolderCount)

	c.JSON(http.StatusOK, gin.H{
		"folders":         result,
		"no_folder_count": noFolderCount,
	})
}

// CreateNoteFolder 创建文件夹
func (h *Handler) CreateNoteFolder(c *gin.Context) {
	userID := c.GetString("userID")

	var req struct {
		Name  string `json:"name" binding:"required,min=1,max=100"`
		Color string `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	color := req.Color
	if color == "" {
		color = "#6366f1"
	}

	folder := model.NoteFolder{
		UserID: userID,
		Name:   req.Name,
		Color:  color,
	}

	if err := h.DB.Create(&folder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文件夹失败"})
		return
	}

	c.JSON(http.StatusCreated, folder)
}

// UpdateNoteFolder 更新文件夹（重命名 / 换颜色）
func (h *Handler) UpdateNoteFolder(c *gin.Context) {
	userID := c.GetString("userID")
	folderID := c.Param("id")

	var req struct {
		Name  *string `json:"name"`
		Color *string `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	var folder model.NoteFolder
	if err := h.DB.Where("id = ? AND user_id = ?", folderID, userID).First(&folder).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件夹不存在"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有可更新的字段"})
		return
	}

	if err := h.DB.Model(&folder).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新文件夹失败"})
		return
	}

	h.DB.Where("id = ? AND user_id = ?", folderID, userID).First(&folder)
	c.JSON(http.StatusOK, folder)
}

// DeleteNoteFolder 删除文件夹（其中的笔记移到"无文件夹"）
func (h *Handler) DeleteNoteFolder(c *gin.Context) {
	userID := c.GetString("userID")
	folderID := c.Param("id")

	var folder model.NoteFolder
	if err := h.DB.Where("id = ? AND user_id = ?", folderID, userID).First(&folder).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件夹不存在"})
		return
	}

	// 将该文件夹下的笔记移到"无文件夹"
	h.DB.Model(&model.Note{}).
		Where("user_id = ? AND folder_id = ?", userID, folderID).
		Update("folder_id", "")

	// 删除文件夹
	if err := h.DB.Delete(&folder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除文件夹失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件夹已删除，笔记已移至未分类"})
}
