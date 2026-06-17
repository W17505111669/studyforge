package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 材料共享市场 ====================

// ToggleShare 切换材料公开状态
// PUT /api/materials/:id/share
func (h *Handler) ToggleShare(c *gin.Context) {
	userID := c.GetString("userID")
	materialID := c.Param("id")

	var material model.Material
	if err := h.DB.Where("id = ? AND user_id = ?", materialID, userID).First(&material).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "材料不存在"})
		return
	}

	if material.IsPublic {
		// 取消公开：清除 share_code
		if err := h.DB.Model(&material).Updates(map[string]interface{}{
			"is_public":   false,
			"share_code":  "",
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"is_public":  false,
			"share_code": "",
			"message":    "已取消分享",
		})
	} else {
		// 设为公开：生成 share_code
		code := generateShareCode()
		if err := h.DB.Model(&material).Updates(map[string]interface{}{
			"is_public":  true,
			"share_code": code,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"is_public":  true,
			"share_code": code,
			"message":    "已公开分享",
		})
	}
}

// ListMarketMaterials 浏览公开材料市场
// GET /api/market/materials?q=keyword&tag=Go&sort=latest|popular&limit=20&offset=0
func (h *Handler) ListMarketMaterials(c *gin.Context) {
	limit, offset := parsePagination(c)
	q := strings.TrimSpace(c.Query("q"))
	tag := strings.TrimSpace(c.Query("tag"))
	sortBy := c.DefaultQuery("sort", "latest")

	query := h.DB.Model(&model.Material{}).
		Select("id, user_id, title, tags, status, is_public, share_code, analyzed_at, created_at, updated_at").
		Where("is_public = ? AND status IN ?", true, []string{"completed", "partial"})

	// 关键词搜索（标题）
	if q != "" {
		query = query.Where("title LIKE ?", "%"+q+"%")
	}

	// 标签过滤
	if tag != "" {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
	}

	// 排序
	switch sortBy {
	case "popular":
		// 按收藏数排序（需要 JOIN collect 表，暂用 created_at 降序代替）
		query = query.Order("created_at DESC")
	default:
		query = query.Order("created_at DESC")
	}

	var total int64
	query.Count(&total)

	var materials []model.Material
	if err := query.Limit(limit).Offset(offset).Find(&materials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载市场材料失败"})
		return
	}

	// 构建响应（附加作者昵称和卡片数）
	type MarketItem struct {
		model.Material
		AuthorName string `json:"author_name"`
		CardCount  int64  `json:"card_count"`
	}

	// 批量查询作者和卡片数
	userIDs := make([]string, 0)
	matIDs := make([]string, 0)
	for _, m := range materials {
		userIDs = append(userIDs, m.UserID)
		matIDs = append(matIDs, m.ID)
	}

	// 批量查作者昵称
	authorMap := make(map[string]string)
	if len(userIDs) > 0 {
		var users []model.User
		h.DB.Select("id, nickname").Where("id IN ?", userIDs).Find(&users)
		for _, u := range users {
			name := u.Nickname
			if name == "" {
				name = u.Username
			}
			authorMap[u.ID] = name
		}
	}

	// 批量查卡片数
	cardCountMap := make(map[string]int64)
	if len(matIDs) > 0 {
		type MatCardCount struct {
			MaterialID string
			Count      int64
		}
		var counts []MatCardCount
		h.DB.Model(&model.Card{}).
			Select("material_id, COUNT(*) as count").
			Where("material_id IN ?", matIDs).
			Group("material_id").
			Find(&counts)
		for _, cc := range counts {
			cardCountMap[cc.MaterialID] = cc.Count
		}
	}

	items := make([]MarketItem, len(materials))
	for i, m := range materials {
		author := authorMap[m.UserID]
		if author == "" {
			author = "匿名用户"
		}
		items[i] = MarketItem{
			Material:   m,
			AuthorName: author,
			CardCount:  cardCountMap[m.ID],
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// PreviewMarketMaterial 预览公开材料（不暴露原始内容）
// GET /api/market/materials/:share_code
func (h *Handler) PreviewMarketMaterial(c *gin.Context) {
	shareCode := c.Param("share_code")
	if shareCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少分享码"})
		return
	}

	var material model.Material
	if err := h.DB.Where("share_code = ? AND is_public = ?", shareCode, true).First(&material).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "材料不存在或已取消分享"})
		return
	}

	// 查作者昵称
	authorName := "匿名用户"
	var author model.User
	if err := h.DB.Select("nickname, username").Where("id = ?", material.UserID).First(&author).Error; err == nil {
		authorName = author.Nickname
		if authorName == "" {
			authorName = author.Username
		}
	}

	// 查卡片数
	var cardCount int64
	h.DB.Model(&model.Card{}).Where("material_id = ?", material.ID).Count(&cardCount)

	// 查练习题数
	var quizCount int64
	h.DB.Model(&model.Quiz{}).Where("material_id = ?", material.ID).Count(&quizCount)

	// 解析 analysis_data 获取摘要（仅提取 summary）
	summary := ""
	keyPoints := []string{}
	if material.AnalysisData != "" {
		type AnalystOutput struct {
			Summary    string `json:"summary"`
			KeyPoints  []struct {
				Concept string `json:"concept"`
			} `json:"key_points"`
		}
		var ao AnalystOutput
		if err := json.Unmarshal([]byte(material.AnalysisData), &ao); err == nil {
			summary = ao.Summary
			for _, kp := range ao.KeyPoints {
				if kp.Concept != "" {
					keyPoints = append(keyPoints, kp.Concept)
				}
				if len(keyPoints) >= 10 {
					break
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           material.ID,
		"title":        material.Title,
		"tags":         material.Tags,
		"author_name":  authorName,
		"card_count":   cardCount,
		"quiz_count":   quizCount,
		"summary":      summary,
		"key_points":   keyPoints,
		"created_at":   material.CreatedAt,
		"share_code":   material.ShareCode,
	})
}

// CollectMarketMaterial 收藏公开材料到自己的库
// POST /api/market/materials/:share_code/collect
func (h *Handler) CollectMarketMaterial(c *gin.Context) {
	userID := c.GetString("userID")
	shareCode := c.Param("share_code")
	if shareCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少分享码"})
		return
	}

	// 查找原始材料
	var original model.Material
	if err := h.DB.Where("share_code = ? AND is_public = ?", shareCode, true).First(&original).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "材料不存在或已取消分享"})
		return
	}

	// 不能收藏自己的材料
	if original.UserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能收藏自己的材料"})
		return
	}

	// 深拷贝材料（不包含原始内容，保留标题/标签/分析数据）
	newMaterial := model.Material{
		UserID:       userID,
		Title:        original.Title + "（收藏）",
		ContentType:  original.ContentType,
		Content:      original.Content,
		SourceURL:    original.SourceURL,
		Status:       original.Status,
		Tags:         original.Tags,
		AnalysisData: original.AnalysisData,
		GraphData:    original.GraphData,
		AnalyzedAt:   original.AnalyzedAt,
	}

	if err := h.DB.Create(&newMaterial).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "收藏失败"})
		return
	}

	// 深拷贝卡片
	var cards []model.Card
	h.DB.Where("material_id = ?", original.ID).Find(&cards)
	if len(cards) > 0 {
		newCards := make([]model.Card, len(cards))
		for i, card := range cards {
			newCards[i] = model.Card{
				MaterialID:  newMaterial.ID,
				UserID:      userID,
				Concept:     card.Concept,
				Detail:      card.Detail,
				Formula:     card.Formula,
				MemoryTip:   card.MemoryTip,
				Difficulty:  card.Difficulty,
				Tags:        card.Tags,
			}
		}
		if err := h.DB.CreateInBatches(&newCards, 50).Error; err != nil {
			log.Printf("收藏卡片失败: %v", err)
		}
	}

	// 深拷贝练习题
	var quizzes []model.Quiz
	h.DB.Where("material_id = ?", original.ID).Find(&quizzes)
	if len(quizzes) > 0 {
		newQuizzes := make([]model.Quiz, len(quizzes))
		for i, quiz := range quizzes {
			newQuizzes[i] = model.Quiz{
				MaterialID:  newMaterial.ID,
				UserID:      userID,
				Question:    quiz.Question,
				Type:        quiz.Type,
				Options:     quiz.Options,
				Answer:      quiz.Answer,
				Explanation: quiz.Explanation,
				Difficulty:  quiz.Difficulty,
				Hint1:       quiz.Hint1,
				Hint2:       quiz.Hint2,
				Hint3:       quiz.Hint3,
			}
		}
		if err := h.DB.CreateInBatches(&newQuizzes, 50).Error; err != nil {
			log.Printf("收藏练习题失败: %v", err)
		}
	}

	// 异步 RAG 索引
	if h.VectorStore != nil && newMaterial.Content != "" {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			if err := h.VectorStore.IndexMaterial(ctx, newMaterial.ID, userID, newMaterial.Content, h.chunkSize, h.chunkOverlap); err != nil {
				log.Printf("收藏材料 RAG 索引失败: %v", err)
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "收藏成功",
		"material_id": newMaterial.ID,
		"card_count":  len(cards),
		"quiz_count":  len(quizzes),
	})
}

// GetMarketTags 获取市场所有公开材料的标签
// GET /api/market/tags
func (h *Handler) GetMarketTags(c *gin.Context) {
	var materials []model.Material
	h.DB.Select("tags").Where("is_public = ? AND tags != '' AND status IN ?", true, []string{"completed", "partial"}).Find(&materials)

	tagCount := make(map[string]int)
	for _, m := range materials {
		for _, t := range strings.Split(m.Tags, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tagCount[t]++
			}
		}
	}

	type TagItem struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	tags := make([]TagItem, 0, len(tagCount))
	for name, count := range tagCount {
		tags = append(tags, TagItem{Name: name, Count: count})
	}
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Count > tags[j].Count
	})

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

// generateShareCode 生成 8 字符唯一分享码
func generateShareCode() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
