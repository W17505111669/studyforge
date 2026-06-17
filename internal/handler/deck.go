package handler

import (
	"log"
	"net/http"
	"sort"
	"strings"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 卡片组（牌组）管理 ====================

// CreateDeck 从现有卡片创建牌组
// POST /api/decks
func (h *Handler) CreateDeck(c *gin.Context) {
	userID := c.GetString("userID")

	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Tags        string   `json:"tags"`
		CardIDs     []string `json:"card_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供牌组名称和卡片"})
		return
	}

	if len(req.CardIDs) == 0 || len(req.CardIDs) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "牌组卡片数量需在 1-200 之间"})
		return
	}

	if len(req.Name) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "牌组名称不能超过 200 字"})
		return
	}

	// 验证卡片属于当前用户
	var cards []model.Card
	h.DB.Where("id IN ? AND user_id = ?", req.CardIDs, userID).Find(&cards)
	if len(cards) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未找到有效的卡片"})
		return
	}

	// 创建牌组
	deck := model.Deck{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Tags:        req.Tags,
		CardCount:   len(cards),
	}
	if err := h.DB.Create(&deck).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建牌组失败"})
		return
	}

	// 创建牌组-卡片关联
	deckCards := make([]model.DeckCard, len(cards))
	for i, card := range cards {
		deckCards[i] = model.DeckCard{
			DeckID: deck.ID,
			CardID: card.ID,
		}
	}
	if err := h.DB.CreateInBatches(&deckCards, 50).Error; err != nil {
		log.Printf("创建牌组卡片关联失败: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "牌组创建成功",
		"deck":    deck,
	})
}

// ListDecks 列出用户的牌组
// GET /api/decks
func (h *Handler) ListDecks(c *gin.Context) {
	userID := c.GetString("userID")

	var decks []model.Deck
	h.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&decks)

	c.JSON(http.StatusOK, gin.H{"decks": decks})
}

// GetDeck 获取牌组详情（含卡片列表）
// GET /api/decks/:id
func (h *Handler) GetDeck(c *gin.Context) {
	userID := c.GetString("userID")
	deckID := c.Param("id")

	var deck model.Deck
	if err := h.DB.Where("id = ? AND user_id = ?", deckID, userID).First(&deck).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "牌组不存在"})
		return
	}

	// 查关联的卡片 ID
	var deckCards []model.DeckCard
	h.DB.Where("deck_id = ?", deckID).Find(&deckCards)
	cardIDs := make([]string, len(deckCards))
	for i, dc := range deckCards {
		cardIDs[i] = dc.CardID
	}

	// 查卡片详情
	var cards []model.Card
	if len(cardIDs) > 0 {
		h.DB.Where("id IN ?", cardIDs).Find(&cards)
	}

	c.JSON(http.StatusOK, gin.H{
		"deck":  deck,
		"cards": cards,
	})
}

// DeleteDeck 删除牌组
// DELETE /api/decks/:id
func (h *Handler) DeleteDeck(c *gin.Context) {
	userID := c.GetString("userID")
	deckID := c.Param("id")

	var deck model.Deck
	if err := h.DB.Where("id = ? AND user_id = ?", deckID, userID).First(&deck).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "牌组不存在"})
		return
	}

	// 删除关联和牌组
	h.DB.Where("deck_id = ?", deckID).Delete(&model.DeckCard{})
	h.DB.Delete(&deck)

	c.JSON(http.StatusOK, gin.H{"message": "牌组已删除"})
}

// ToggleDeckShare 切换牌组公开状态
// PUT /api/decks/:id/share
func (h *Handler) ToggleDeckShare(c *gin.Context) {
	userID := c.GetString("userID")
	deckID := c.Param("id")

	var deck model.Deck
	if err := h.DB.Where("id = ? AND user_id = ?", deckID, userID).First(&deck).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "牌组不存在"})
		return
	}

	if deck.IsPublic {
		h.DB.Model(&deck).Updates(map[string]interface{}{
			"is_public":  false,
			"share_code": "",
		})
		c.JSON(http.StatusOK, gin.H{
			"is_public":  false,
			"share_code": "",
			"message":    "已取消分享",
		})
	} else {
		code := generateShareCode()
		h.DB.Model(&deck).Updates(map[string]interface{}{
			"is_public":  true,
			"share_code": code,
		})
		c.JSON(http.StatusOK, gin.H{
			"is_public":  true,
			"share_code": code,
			"message":    "已公开分享",
		})
	}
}

// ==================== 牌组市场 ====================

// ListMarketDecks 浏览公开牌组市场
// GET /api/market/decks?q=keyword&tag=Go&sort=latest|popular&limit=20&offset=0
func (h *Handler) ListMarketDecks(c *gin.Context) {
	limit, offset := parsePagination(c)
	q := strings.TrimSpace(c.Query("q"))
	tag := strings.TrimSpace(c.Query("tag"))
	sortBy := c.DefaultQuery("sort", "latest")

	query := h.DB.Model(&model.Deck{}).Where("is_public = ?", true)

	if q != "" {
		query = query.Where("name LIKE ?", "%"+q+"%")
	}
	if tag != "" {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
	}

	switch sortBy {
	case "popular":
		query = query.Order("collect_count DESC, created_at DESC")
	default:
		query = query.Order("created_at DESC")
	}

	var total int64
	query.Count(&total)

	var decks []model.Deck
	if err := query.Limit(limit).Offset(offset).Find(&decks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载市场牌组失败"})
		return
	}

	// 构建响应（附加作者昵称）
	type DeckMarketItem struct {
		model.Deck
		AuthorName string `json:"author_name"`
	}

	userIDs := make([]string, 0, len(decks))
	for _, d := range decks {
		userIDs = append(userIDs, d.UserID)
	}

	authorMap := make(map[string]string)
	if len(userIDs) > 0 {
		var users []model.User
		h.DB.Select("id, nickname, username").Where("id IN ?", userIDs).Find(&users)
		for _, u := range users {
			name := u.Nickname
			if name == "" {
				name = u.Username
			}
			authorMap[u.ID] = name
		}
	}

	items := make([]DeckMarketItem, len(decks))
	for i, d := range decks {
		author := authorMap[d.UserID]
		if author == "" {
			author = "匿名用户"
		}
		items[i] = DeckMarketItem{
			Deck:       d,
			AuthorName: author,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// PreviewMarketDeck 预览公开牌组（不暴露完整卡片内容）
// GET /api/market/decks/:share_code
func (h *Handler) PreviewMarketDeck(c *gin.Context) {
	shareCode := c.Param("share_code")
	if shareCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少分享码"})
		return
	}

	var deck model.Deck
	if err := h.DB.Where("share_code = ? AND is_public = ?", shareCode, true).First(&deck).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "牌组不存在或已取消分享"})
		return
	}

	// 查作者昵称
	authorName := "匿名用户"
	var author model.User
	if err := h.DB.Select("nickname, username").Where("id = ?", deck.UserID).First(&author).Error; err == nil {
		authorName = author.Nickname
		if authorName == "" {
			authorName = author.Username
		}
	}

	// 查关联卡片（仅返回概要）
	var deckCards []model.DeckCard
	h.DB.Where("deck_id = ?", deck.ID).Find(&deckCards)
	cardIDs := make([]string, len(deckCards))
	for i, dc := range deckCards {
		cardIDs[i] = dc.CardID
	}

	var cards []model.Card
	if len(cardIDs) > 0 {
		h.DB.Select("id, concept, difficulty, tags").Where("id IN ?", cardIDs).Find(&cards)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          deck.ID,
		"name":        deck.Name,
		"description": deck.Description,
		"tags":        deck.Tags,
		"author_name": authorName,
		"card_count":  deck.CardCount,
		"collect_count": deck.CollectCount,
		"cards":       cards,
		"created_at":  deck.CreatedAt,
		"share_code":  deck.ShareCode,
	})
}

// CollectMarketDeck 收藏公开牌组到自己的库
// POST /api/market/decks/:share_code/collect
func (h *Handler) CollectMarketDeck(c *gin.Context) {
	userID := c.GetString("userID")
	shareCode := c.Param("share_code")
	if shareCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少分享码"})
		return
	}

	var original model.Deck
	if err := h.DB.Where("share_code = ? AND is_public = ?", shareCode, true).First(&original).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "牌组不存在或已取消分享"})
		return
	}

	if original.UserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能收藏自己的牌组"})
		return
	}

	// 查找原始牌组的卡片
	var deckCards []model.DeckCard
	h.DB.Where("deck_id = ?", original.ID).Find(&deckCards)
	cardIDs := make([]string, len(deckCards))
	for i, dc := range deckCards {
		cardIDs[i] = dc.CardID
	}

	var cards []model.Card
	if len(cardIDs) > 0 {
		h.DB.Where("id IN ?", cardIDs).Find(&cards)
	}

	// 深拷贝：创建一个新材料承载牌组的卡片
	newMaterial := model.Material{
		UserID:  userID,
		Title:   original.Name + "（牌组收藏）",
		Status:  "completed",
		Tags:    original.Tags,
	}
	if err := h.DB.Create(&newMaterial).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "收藏失败"})
		return
	}

	// 深拷贝卡片
	newCards := make([]model.Card, len(cards))
	for i, card := range cards {
		difficulty := card.Difficulty
		if difficulty == "" {
			difficulty = "medium"
		}
		newCards[i] = model.Card{
			MaterialID: newMaterial.ID,
			UserID:     userID,
			Concept:    card.Concept,
			Detail:     card.Detail,
			Formula:    card.Formula,
			MemoryTip:  card.MemoryTip,
			Difficulty: difficulty,
			Tags:       card.Tags,
			EaseFactor: 2.5,
		}
	}
	if len(newCards) > 0 {
		if err := h.DB.CreateInBatches(&newCards, 50).Error; err != nil {
			log.Printf("收藏牌组卡片失败: %v", err)
		}
	}

	// 创建新牌组
	newDeck := model.Deck{
		UserID:      userID,
		Name:        original.Name + "（收藏）",
		Description: original.Description,
		Tags:        original.Tags,
		CardCount:   len(newCards),
	}
	if err := h.DB.Create(&newDeck).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建收藏牌组失败"})
		return
	}

	// 创建新牌组-卡片关联
	newDeckCards := make([]model.DeckCard, len(newCards))
	for i, nc := range newCards {
		newDeckCards[i] = model.DeckCard{
			DeckID: newDeck.ID,
			CardID: nc.ID,
		}
	}
	if len(newDeckCards) > 0 {
		h.DB.CreateInBatches(&newDeckCards, 50)
	}

	// 递增原始牌组收藏数
	h.DB.Model(&original).Update("collect_count", original.CollectCount+1)

	c.JSON(http.StatusOK, gin.H{
		"message":    "牌组收藏成功",
		"deck_id":    newDeck.ID,
		"card_count": len(newCards),
	})
}

// GetMarketDeckTags 获取市场所有公开牌组的标签
// GET /api/market/decks/tags
func (h *Handler) GetMarketDeckTags(c *gin.Context) {
	var decks []model.Deck
	h.DB.Select("tags").Where("is_public = ? AND tags != ''", true).Find(&decks)

	tagCount := make(map[string]int)
	for _, d := range decks {
		for _, t := range strings.Split(d.Tags, ",") {
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
