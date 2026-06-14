package handler

import (
	"math/rand"
	"net/http"
	"sort"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 错题本 ====================

// MistakeItem 错题列表响应项（含题目详情+材料标题）
type MistakeItem struct {
	ID            string `json:"id"`
	QuizID        string `json:"quiz_id"`
	UserAnswer    string `json:"user_answer"`
	CorrectAnswer string `json:"correct_answer"`
	MistakeAt     string `json:"mistake_at"`
	Reviewed      bool   `json:"reviewed"`
	// 题目详情
	Question      string `json:"question"`
	QuizType      string `json:"quiz_type"`
	Difficulty    string `json:"difficulty"`
	Options       string `json:"options"`
	Explanation   string `json:"explanation"`
	MaterialID    string `json:"material_id"`
	MaterialTitle string `json:"material_title"`
}

// ListMistakes 获取用户错题列表（含题目详情+材料标题）
// GET /api/mistakes
func (h *Handler) ListMistakes(c *gin.Context) {
	userID := c.GetString("userID")

	// 分页参数
	limit, offset := parsePagination(c)

	// 查询总数
	var total int64
	h.DB.Model(&model.QuizMistake{}).Where("user_id = ?", userID).Count(&total)

	// 查询错题（按错题时间倒序）
	var mistakes []model.QuizMistake
	h.DB.Where("user_id = ?", userID).
		Order("mistake_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&mistakes)

	// 批量加载关联的 Quiz 详情（避免 N+1）
	quizIDs := make([]string, 0, len(mistakes))
	for _, m := range mistakes {
		quizIDs = append(quizIDs, m.QuizID)
	}
	quizMap := make(map[string]*model.Quiz)
	if len(quizIDs) > 0 {
		var quizzes []model.Quiz
		h.DB.Where("id IN ? AND user_id = ?", quizIDs, userID).Find(&quizzes)
		for i := range quizzes {
			quizMap[quizzes[i].ID] = &quizzes[i]
		}
	}

	// 批量查询关联的材料标题
	materialIDs := make([]string, 0)
	materialSet := make(map[string]bool)
	for _, q := range quizMap {
		if !materialSet[q.MaterialID] {
			materialIDs = append(materialIDs, q.MaterialID)
			materialSet[q.MaterialID] = true
		}
	}
	materialTitles := make(map[string]string)
	if len(materialIDs) > 0 {
		var materials []model.Material
		h.DB.Select("id, title").Where("id IN ? AND user_id = ?", materialIDs, userID).Find(&materials)
		for _, mat := range materials {
			materialTitles[mat.ID] = mat.Title
		}
	}

	// 组装响应
	items := make([]MistakeItem, 0, len(mistakes))
	for _, m := range mistakes {
		item := MistakeItem{
			ID:            m.ID,
			QuizID:        m.QuizID,
			UserAnswer:    m.UserAnswer,
			CorrectAnswer: m.CorrectAnswer,
			MistakeAt:     m.MistakeAt.Format(time.RFC3339),
			Reviewed:      m.Reviewed,
		}
		if q, ok := quizMap[m.QuizID]; ok {
			item.Question = q.Question
			item.QuizType = q.Type
			item.Difficulty = q.Difficulty
			item.Options = q.Options
			item.Explanation = q.Explanation
			item.MaterialID = q.MaterialID
			item.MaterialTitle = materialTitles[q.MaterialID]
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// ReviewMistake 标记错题为已复习
// POST /api/mistakes/:id/review
func (h *Handler) ReviewMistake(c *gin.Context) {
	userID := c.GetString("userID")
	mistakeID := c.Param("id")

	var mistake model.QuizMistake
	if err := h.DB.Where("id = ? AND user_id = ?", mistakeID, userID).First(&mistake).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "错题不存在"})
		return
	}

	h.DB.Model(&mistake).Update("reviewed", true)
	c.JSON(http.StatusOK, gin.H{"message": "已标记为已复习"})
}

// BatchReviewMistakes 批量标记错题为已复习
// POST /api/mistakes/batch-review
func (h *Handler) BatchReviewMistakes(c *gin.Context) {
	userID := c.GetString("userID")

	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供错题 ID 列表"})
		return
	}

	h.DB.Model(&model.QuizMistake{}).
		Where("id IN ? AND user_id = ?", req.IDs, userID).
		Update("reviewed", true)

	c.JSON(http.StatusOK, gin.H{"message": "批量标记完成"})
}

// DeleteMistake 删除单条错题记录
// DELETE /api/mistakes/:id
func (h *Handler) DeleteMistake(c *gin.Context) {
	userID := c.GetString("userID")
	mistakeID := c.Param("id")

	result := h.DB.Where("id = ? AND user_id = ?", mistakeID, userID).Delete(&model.QuizMistake{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "错题不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// GetMistakeStats 获取错题统计
// GET /api/mistakes/stats
func (h *Handler) GetMistakeStats(c *gin.Context) {
	userID := c.GetString("userID")

	var total, reviewed, unreviewed int64
	h.DB.Model(&model.QuizMistake{}).Where("user_id = ?", userID).Count(&total)
	h.DB.Model(&model.QuizMistake{}).Where("user_id = ? AND reviewed = ?", userID, true).Count(&reviewed)
	unreviewed = total - reviewed

	// 按难度统计
	type DiffStat struct {
		Difficulty string `json:"difficulty"`
		Count      int64  `json:"count"`
	}
	var diffStats []DiffStat
	h.DB.Model(&model.QuizMistake{}).
		Select("quizzes.difficulty, COUNT(*) as count").
		Joins("JOIN quizzes ON quizzes.id = quiz_mistakes.quiz_id").
		Where("quiz_mistakes.user_id = ?", userID).
		Group("quizzes.difficulty").
		Scan(&diffStats)

	c.JSON(http.StatusOK, gin.H{
		"total":           total,
		"reviewed":        reviewed,
		"unreviewed":      unreviewed,
		"by_difficulty":   diffStats,
	})
}

// ConsolidatePractice 巩固强化 — 根据错题类型找同类题练习
// POST /api/mistakes/consolidate
func (h *Handler) ConsolidatePractice(c *gin.Context) {
	userID := c.GetString("userID")

	// 1. 查询用户所有错题
	var mistakes []model.QuizMistake
	h.DB.Where("user_id = ?", userID).Find(&mistakes)

	if len(mistakes) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"quizzes":    []interface{}{},
			"total":      0,
			"weak_areas": []interface{}{},
			"message":    "暂无错题，无法生成巩固练习",
		})
		return
	}

	// 2. 收集 quiz ID + 加载错题详情（合并分析）
	mistakeQuizIDs := make([]string, 0, len(mistakes))
	for _, m := range mistakes {
		mistakeQuizIDs = append(mistakeQuizIDs, m.QuizID)
	}

	var mistakeQuizzes []model.Quiz
	h.DB.Where("id IN ? AND user_id = ?", mistakeQuizIDs, userID).Find(&mistakeQuizzes)

	if len(mistakeQuizzes) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"quizzes":    []interface{}{},
			"total":      0,
			"weak_areas": []interface{}{},
			"message":    "错题关联的题目数据缺失",
		})
		return
	}

	// 单次遍历：同时收集统计信息（替代多次循环）
	type matStats struct {
		count    int
		types    map[string]int
		diffs    map[string]int
	}
	matStatsMap := make(map[string]*matStats) // materialID → stats
	globalTypeSet := make(map[string]int)
	globalDiffSet := make(map[string]int)

	for _, q := range mistakeQuizzes {
		globalTypeSet[q.Type]++
		globalDiffSet[q.Difficulty]++
		if _, ok := matStatsMap[q.MaterialID]; !ok {
			matStatsMap[q.MaterialID] = &matStats{types: make(map[string]int), diffs: make(map[string]int)}
		}
		ms := matStatsMap[q.MaterialID]
		ms.count++
		ms.types[q.Type]++
		ms.diffs[q.Difficulty]++
	}

	// 3. 收集已作答过的 quiz ID
	attemptedQuizIDs := make(map[string]bool, len(mistakeQuizIDs))
	for _, id := range mistakeQuizIDs {
		attemptedQuizIDs[id] = true
	}
	var recentAttempts []model.QuizAttempt
	h.DB.Select("quiz_id").Where("user_id = ?", userID).
		Order("created_at DESC").Limit(100).Find(&recentAttempts)
	for _, a := range recentAttempts {
		attemptedQuizIDs[a.QuizID] = true
	}

	excludeIDs := make([]string, 0, len(attemptedQuizIDs))
	for id := range attemptedQuizIDs {
		excludeIDs = append(excludeIDs, id)
	}

	// 4. 材料按错题数降序排序（sort.Slice 替代冒泡）
	type materialScore struct {
		materialID string
		count      int
	}
	materialScores := make([]materialScore, 0, len(matStatsMap))
	for mid, s := range matStatsMap {
		materialScores = append(materialScores, materialScore{mid, s.count})
	}
	sort.Slice(materialScores, func(i, j int) bool {
		return materialScores[i].count > materialScores[j].count
	})

	// 5. 批量从薄弱材料中找未做过的题（单次查询 + Go 层分组，替代逐材料 RANDOM()）
	const maxQuizzes = 15
	var consolidateQuizzes []model.Quiz
	usedIDs := make(map[string]bool)

	if len(excludeIDs) > 0 {
		weakMatIDs := make([]string, 0, len(materialScores))
		for _, ms := range materialScores {
			weakMatIDs = append(weakMatIDs, ms.materialID)
		}

		var candidateQuizzes []model.Quiz
		h.DB.Where("material_id IN ? AND user_id = ? AND id NOT IN ?", weakMatIDs, userID, excludeIDs).
			Find(&candidateQuizzes)

		// Go 层按材料分组
		byMaterial := make(map[string][]model.Quiz)
		for _, q := range candidateQuizzes {
			byMaterial[q.MaterialID] = append(byMaterial[q.MaterialID], q)
		}
		// Go 层随机打乱（替代 SQL ORDER BY RANDOM()）
		for _, qs := range byMaterial {
			rand.Shuffle(len(qs), func(i, j int) { qs[i], qs[j] = qs[j], qs[i] })
		}

		// 按薄弱度顺序取题
		for _, ms := range materialScores {
			if len(consolidateQuizzes) >= maxQuizzes {
				break
			}
			for _, q := range byMaterial[ms.materialID] {
				if len(consolidateQuizzes) >= maxQuizzes {
					break
				}
				if !usedIDs[q.ID] {
					consolidateQuizzes = append(consolidateQuizzes, q)
					usedIDs[q.ID] = true
				}
			}
		}
	}

	// 6. 不够则扩展到同题型+同难度的跨材料题目
	if len(consolidateQuizzes) < maxQuizzes {
		remaining := maxQuizzes - len(consolidateQuizzes)
		topType, topTypeCount := "", 0
		for t, cnt := range globalTypeSet {
			if cnt > topTypeCount {
				topType, topTypeCount = t, cnt
			}
		}
		topDiff, topDiffCount := "", 0
		for d, cnt := range globalDiffSet {
			if cnt > topDiffCount {
				topDiff, topDiffCount = d, cnt
			}
		}

		allExcludeIDs := make([]string, 0, len(excludeIDs)+len(usedIDs))
		allExcludeIDs = append(allExcludeIDs, excludeIDs...)
		for id := range usedIDs {
			allExcludeIDs = append(allExcludeIDs, id)
		}

		query := h.DB.Where("user_id = ?", userID)
		if len(allExcludeIDs) > 0 {
			query = query.Where("id NOT IN ?", allExcludeIDs)
		}
		if topType != "" {
			query = query.Where("type = ?", topType)
		}
		if topDiff != "" {
			query = query.Where("difficulty = ?", topDiff)
		}

		var extraQuizzes []model.Quiz
		candidateLimit := remaining * 5
		if candidateLimit > 200 {
			candidateLimit = 200
		}
		query.Limit(candidateLimit).Find(&extraQuizzes)
		rand.Shuffle(len(extraQuizzes), func(i, j int) { extraQuizzes[i], extraQuizzes[j] = extraQuizzes[j], extraQuizzes[i] })
		for _, q := range extraQuizzes {
			if len(consolidateQuizzes) >= maxQuizzes {
				break
			}
			if !usedIDs[q.ID] {
				consolidateQuizzes = append(consolidateQuizzes, q)
				usedIDs[q.ID] = true
			}
		}
	}

	// 6b. 仍未找到 → 放宽限制，允许复用已答过的题（仅排除错题本身）
	if len(consolidateQuizzes) == 0 {
		mistakeOnlyIDs := make([]string, len(mistakeQuizIDs))
		copy(mistakeOnlyIDs, mistakeQuizIDs)

		var fallbackQuizzes []model.Quiz
		query := h.DB.Where("user_id = ?", userID)
		if len(mistakeOnlyIDs) > 0 {
			query = query.Where("id NOT IN ?", mistakeOnlyIDs)
		}
		query.Find(&fallbackQuizzes)
		rand.Shuffle(len(fallbackQuizzes), func(i, j int) { fallbackQuizzes[i], fallbackQuizzes[j] = fallbackQuizzes[j], fallbackQuizzes[i] })

		for _, q := range fallbackQuizzes {
			if len(consolidateQuizzes) >= maxQuizzes {
				break
			}
			consolidateQuizzes = append(consolidateQuizzes, q)
		}
	}

	if len(consolidateQuizzes) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"quizzes":    []interface{}{},
			"total":      0,
			"weak_areas": []interface{}{},
			"message":    "题库中没有其他题目可供巩固练习",
		})
		return
	}

	// 7. 一次性批量查询所有相关材料标题（错题材料 + 巩固题材料，替代 N+1）
	allMatIDs := make([]string, 0)
	allMatSet := make(map[string]bool)
	for mid := range matStatsMap {
		if !allMatSet[mid] {
			allMatIDs = append(allMatIDs, mid)
			allMatSet[mid] = true
		}
	}
	for _, q := range consolidateQuizzes {
		if !allMatSet[q.MaterialID] {
			allMatIDs = append(allMatIDs, q.MaterialID)
			allMatSet[q.MaterialID] = true
		}
	}
	matTitles := make(map[string]string)
	if len(allMatIDs) > 0 {
		var mats []model.Material
		h.DB.Select("id, title").Where("id IN ? AND user_id = ?", allMatIDs, userID).Find(&mats)
		for _, m := range mats {
			matTitles[m.ID] = m.Title
		}
	}

	// 8. 构建响应
	type ConsolidateItem struct {
		ID            string `json:"id"`
		Question      string `json:"question"`
		Type          string `json:"type"`
		Difficulty    string `json:"difficulty"`
		Options       string `json:"options"`
		Answer        string `json:"answer"`
		Explanation   string `json:"explanation"`
		MaterialID    string `json:"material_id"`
		MaterialTitle string `json:"material_title"`
	}

	items := make([]ConsolidateItem, 0, len(consolidateQuizzes))
	for _, q := range consolidateQuizzes {
		items = append(items, ConsolidateItem{
			ID:            q.ID,
			Question:      q.Question,
			Type:          q.Type,
			Difficulty:    q.Difficulty,
			Options:       q.Options,
			Answer:        q.Answer,
			Explanation:   q.Explanation,
			MaterialID:    q.MaterialID,
			MaterialTitle: matTitles[q.MaterialID],
		})
	}

	// 9. 构建薄弱点分析（使用预计算的 matStatsMap，零额外 DB 查询）
	type WeakArea struct {
		MaterialID     string `json:"material_id"`
		MaterialTitle  string `json:"material_title"`
		MistakeCount   int    `json:"mistake_count"`
		MainType       string `json:"main_type"`
		MainDifficulty string `json:"main_difficulty"`
	}

	weakAreas := make([]WeakArea, 0, len(materialScores))
	for _, ms := range materialScores {
		s := matStatsMap[ms.materialID]
		mainType, mainTypeCount := "", 0
		for t, cnt := range s.types {
			if cnt > mainTypeCount {
				mainType, mainTypeCount = t, cnt
			}
		}
		mainDiff, mainDiffCount := "", 0
		for d, cnt := range s.diffs {
			if cnt > mainDiffCount {
				mainDiff, mainDiffCount = d, cnt
			}
		}
		weakAreas = append(weakAreas, WeakArea{
			MaterialID:     ms.materialID,
			MaterialTitle:  matTitles[ms.materialID],
			MistakeCount:   ms.count,
			MainType:       mainType,
			MainDifficulty: mainDiff,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"quizzes":    items,
		"total":      len(items),
		"weak_areas": weakAreas,
		"mistake_stats": gin.H{
			"total_mistakes": len(mistakes),
			"by_type":        globalTypeSet,
			"by_difficulty":  globalDiffSet,
		},
	})
}

// RetryMistakes 错题重练 — 从未复习的错题中选取题目组成练习
// POST /api/mistakes/retry
func (h *Handler) RetryMistakes(c *gin.Context) {
	userID := c.GetString("userID")

	// 查询未复习的错题（最多 20 道，按时间倒序）
	var mistakes []model.QuizMistake
	h.DB.Where("user_id = ? AND reviewed = ?", userID, false).
		Order("mistake_at DESC").
		Limit(20).
		Find(&mistakes)

	if len(mistakes) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "没有待复习的错题",
			"quizzes": []interface{}{},
			"total":   0,
		})
		return
	}

	// 收集 quiz ID 并查询完整题目
	quizIDs := make([]string, 0, len(mistakes))
	mistakeMap := make(map[string]string) // quizID -> userAnswer (上次答错的答案)
	for _, m := range mistakes {
		quizIDs = append(quizIDs, m.QuizID)
		mistakeMap[m.QuizID] = m.UserAnswer
	}

	var quizzes []model.Quiz
	h.DB.Where("id IN ? AND user_id = ?", quizIDs, userID).Find(&quizzes)

	// 构建响应（附带上次错误答案）
	type RetryQuizItem struct {
		model.Quiz
		LastWrongAnswer string `json:"last_wrong_answer"`
		MistakeID       string `json:"mistake_id"`
	}

	items := make([]RetryQuizItem, 0, len(quizzes))
	for _, q := range quizzes {
		item := RetryQuizItem{
			Quiz:            q,
			LastWrongAnswer: mistakeMap[q.ID],
		}
		// 找到对应的 mistake ID
		for _, m := range mistakes {
			if m.QuizID == q.ID {
				item.MistakeID = m.ID
				break
			}
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"quizzes": items,
		"total":   len(items),
	})
}
