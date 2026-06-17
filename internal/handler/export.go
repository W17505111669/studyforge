package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 学习数据导出 ====================

// ExportSection 导出数据分区
type ExportSection struct {
	Name string      `json:"name"`
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// ExportData 导出学习数据为 CSV 或 JSON
// GET /api/export/data?type=cards|quizzes|chats|mistakes|all&format=csv|json&date_from=2026-01-01&date_to=2026-12-31
func (h *Handler) ExportData(c *gin.Context) {
	userID := c.GetString("userID")
	exportType := c.DefaultQuery("type", "all")
	format := c.DefaultQuery("format", "json")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	// 解析日期范围
	var fromTime, toTime time.Time
	var hasDateFrom, hasDateTo bool
	if dateFrom != "" {
		t, err := time.Parse("2006-01-02", dateFrom)
		if err == nil {
			fromTime = t
			hasDateFrom = true
		}
	}
	if dateTo != "" {
		t, err := time.Parse("2006-01-02", dateTo)
		if err == nil {
			toTime = t.Add(24*time.Hour - time.Second) // 包含当天全天
			hasDateTo = true
		}
	}

	// 收集导出数据
	var sections []ExportSection
	typeSet := parseTypeSet(exportType)

	if typeSet["cards"] {
		cards, err := h.exportCards(userID, hasDateFrom, fromTime, hasDateTo, toTime)
		if err == nil {
			sections = append(sections, ExportSection{Name: "知识卡片", Type: "cards", Data: cards})
		}
	}
	if typeSet["quizzes"] {
		quizzes, err := h.exportQuizzes(userID, hasDateFrom, fromTime, hasDateTo, toTime)
		if err == nil {
			sections = append(sections, ExportSection{Name: "练习题", Type: "quizzes", Data: quizzes})
		}
	}
	if typeSet["chats"] {
		chats, err := h.exportChats(userID, hasDateFrom, fromTime, hasDateTo, toTime)
		if err == nil {
			sections = append(sections, ExportSection{Name: "对话记录", Type: "chats", Data: chats})
		}
	}
	if typeSet["mistakes"] {
		mistakes, err := h.exportMistakes(userID, hasDateFrom, fromTime, hasDateTo, toTime)
		if err == nil {
			sections = append(sections, ExportSection{Name: "错题记录", Type: "mistakes", Data: mistakes})
		}
	}

	if len(sections) == 0 {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "message": "没有找到符合条件的数据"})
		return
	}

	if format == "csv" {
		h.exportCSV(c, sections, exportType)
	} else {
		h.exportJSON(c, sections, exportType)
	}
}

// ExportDataPreview 预览导出数据（前 10 条）
// GET /api/export/preview?type=cards|quizzes|chats|mistakes&date_from=&date_to=
func (h *Handler) ExportDataPreview(c *gin.Context) {
	userID := c.GetString("userID")
	exportType := c.DefaultQuery("type", "cards")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	var fromTime, toTime time.Time
	var hasDateFrom, hasDateTo bool
	if dateFrom != "" {
		t, err := time.Parse("2006-01-02", dateFrom)
		if err == nil {
			fromTime = t
			hasDateFrom = true
		}
	}
	if dateTo != "" {
		t, err := time.Parse("2006-01-02", dateTo)
		if err == nil {
			toTime = t.Add(24*time.Hour - time.Second)
			hasDateTo = true
		}
	}

	// 只预览单个类型
	result := gin.H{}
	switch exportType {
	case "cards":
		cards, err := h.exportCards(userID, hasDateFrom, fromTime, hasDateTo, toTime)
		if err == nil {
			preview := cards
			if len(preview) > 10 {
				preview = preview[:10]
			}
			result["data"] = preview
			result["total"] = len(cards)
		}
	case "quizzes":
		quizzes, err := h.exportQuizzes(userID, hasDateFrom, fromTime, hasDateTo, toTime)
		if err == nil {
			preview := quizzes
			if len(preview) > 10 {
				preview = preview[:10]
			}
			result["data"] = preview
			result["total"] = len(quizzes)
		}
	case "chats":
		chats, err := h.exportChats(userID, hasDateFrom, fromTime, hasDateTo, toTime)
		if err == nil {
			preview := chats
			if len(preview) > 10 {
				preview = preview[:10]
			}
			result["data"] = preview
			result["total"] = len(chats)
		}
	case "mistakes":
		mistakes, err := h.exportMistakes(userID, hasDateFrom, fromTime, hasDateTo, toTime)
		if err == nil {
			preview := mistakes
			if len(preview) > 10 {
				preview = preview[:10]
			}
			result["data"] = preview
			result["total"] = len(mistakes)
		}
	default:
		result["data"] = []interface{}{}
		result["total"] = 0
	}

	c.JSON(http.StatusOK, result)
}

// ==================== 数据查询 ====================

// CardExportItem 卡片导出项
type CardExportItem struct {
	Concept      string  `json:"concept"`
	Detail       string  `json:"detail"`
	Tags         string  `json:"tags"`
	Difficulty   string  `json:"difficulty"`
	ReviewCount  int     `json:"review_count"`
	EaseFactor   float64 `json:"ease_factor"`
	NextReviewAt string  `json:"next_review_at"`
	IsBookmarked bool    `json:"is_bookmarked"`
	UserNote     string  `json:"user_note"`
	MaterialTitle string `json:"material_title"`
	CreatedAt    string  `json:"created_at"`
}

func (h *Handler) exportCards(userID string, hasFrom bool, fromTime time.Time, hasTo bool, toTime time.Time) ([]CardExportItem, error) {
	var cards []model.Card
	query := h.DB.Where("user_id = ?", userID)
	if hasFrom {
		query = query.Where("created_at >= ?", fromTime)
	}
	if hasTo {
		query = query.Where("created_at <= ?", toTime)
	}
	if err := query.Order("created_at DESC").Find(&cards).Error; err != nil {
		return nil, err
	}

	// 批量查材料标题
	materialIDs := make([]string, 0)
	for _, c := range cards {
		materialIDs = append(materialIDs, c.MaterialID)
	}
	materialTitleMap := make(map[string]string)
	if len(materialIDs) > 0 {
		var materials []model.Material
		h.DB.Select("id, title").Where("id IN ?", materialIDs).Find(&materials)
		for _, m := range materials {
			materialTitleMap[m.ID] = m.Title
		}
	}

	items := make([]CardExportItem, len(cards))
	for i, c := range cards {
		nextReview := ""
		if c.NextReviewAt != nil {
			nextReview = c.NextReviewAt.Format("2006-01-02 15:04")
		}
		items[i] = CardExportItem{
			Concept:       c.Concept,
			Detail:        c.Detail,
			Tags:          c.Tags,
			Difficulty:    c.Difficulty,
			ReviewCount:   c.ReviewCount,
			EaseFactor:    c.EaseFactor,
			NextReviewAt:  nextReview,
			IsBookmarked:  c.IsBookmarked,
			UserNote:      c.UserNote,
			MaterialTitle: materialTitleMap[c.MaterialID],
			CreatedAt:     c.CreatedAt.Format("2006-01-02 15:04"),
		}
	}
	return items, nil
}

// QuizExportItem 练习题导出项
type QuizExportItem struct {
	Question      string `json:"question"`
	Type          string `json:"type"`
	Difficulty    string `json:"difficulty"`
	Answer        string `json:"answer"`
	Explanation   string `json:"explanation"`
	Hint1         string `json:"hint_1"`
	Hint2         string `json:"hint_2"`
	Hint3         string `json:"hint_3"`
	MaterialTitle string `json:"material_title"`
	CreatedAt     string `json:"created_at"`
}

func (h *Handler) exportQuizzes(userID string, hasFrom bool, fromTime time.Time, hasTo bool, toTime time.Time) ([]QuizExportItem, error) {
	var quizzes []model.Quiz
	query := h.DB.Where("user_id = ?", userID)
	if hasFrom {
		query = query.Where("created_at >= ?", fromTime)
	}
	if hasTo {
		query = query.Where("created_at <= ?", toTime)
	}
	if err := query.Order("created_at DESC").Find(&quizzes).Error; err != nil {
		return nil, err
	}

	// 批量查材料标题
	materialIDs := make([]string, 0)
	for _, q := range quizzes {
		materialIDs = append(materialIDs, q.MaterialID)
	}
	materialTitleMap := make(map[string]string)
	if len(materialIDs) > 0 {
		var materials []model.Material
		h.DB.Select("id, title").Where("id IN ?", materialIDs).Find(&materials)
		for _, m := range materials {
			materialTitleMap[m.ID] = m.Title
		}
	}

	items := make([]QuizExportItem, len(quizzes))
	for i, q := range quizzes {
		items[i] = QuizExportItem{
			Question:      q.Question,
			Type:          q.Type,
			Difficulty:    q.Difficulty,
			Answer:        q.Answer,
			Explanation:   q.Explanation,
			Hint1:         q.Hint1,
			Hint2:         q.Hint2,
			Hint3:         q.Hint3,
			MaterialTitle: materialTitleMap[q.MaterialID],
			CreatedAt:     q.CreatedAt.Format("2006-01-02 15:04"),
		}
	}
	return items, nil
}

// ChatExportItem 对话导出项
type ChatExportItem struct {
	ConversationTitle string `json:"conversation_title"`
	Content           string `json:"content"`
	Role              string `json:"role"`
	CreatedAt         string `json:"created_at"`
}

func (h *Handler) exportChats(userID string, hasFrom bool, fromTime time.Time, hasTo bool, toTime time.Time) ([]ChatExportItem, error) {
	// 先查对话会话
	var conversations []model.Conversation
	convQuery := h.DB.Where("user_id = ?", userID)
	if hasFrom {
		convQuery = convQuery.Where("created_at >= ?", fromTime)
	}
	if hasTo {
		convQuery = convQuery.Where("created_at <= ?", toTime)
	}
	if err := convQuery.Order("created_at DESC").Find(&conversations).Error; err != nil {
		return nil, err
	}

	if len(conversations) == 0 {
		return []ChatExportItem{}, nil
	}

	// 批量查所有消息
	convIDs := make([]string, len(conversations))
	convTitleMap := make(map[string]string)
	for i, conv := range conversations {
		convIDs[i] = conv.ID
		convTitleMap[conv.ID] = conv.Title
	}

	var messages []model.ChatMessage
	if err := h.DB.Where("conversation_id IN ?", convIDs).Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, err
	}

	items := make([]ChatExportItem, len(messages))
	for i, m := range messages {
		items[i] = ChatExportItem{
			ConversationTitle: convTitleMap[m.ConversationID],
			Content:           m.Content,
			Role:              m.Role,
			CreatedAt:         m.CreatedAt.Format("2006-01-02 15:04"),
		}
	}
	return items, nil
}

// MistakeExportItem 错题导出项
type MistakeExportItem struct {
	Question       string `json:"question"`
	Type           string `json:"type"`
	Difficulty     string `json:"difficulty"`
	UserAnswer     string `json:"user_answer"`
	CorrectAnswer  string `json:"correct_answer"`
	Explanation    string `json:"explanation"`
	MaterialTitle  string `json:"material_title"`
	MistakeAt      string `json:"mistake_at"`
}

func (h *Handler) exportMistakes(userID string, hasFrom bool, fromTime time.Time, hasTo bool, toTime time.Time) ([]MistakeExportItem, error) {
	var mistakes []model.QuizMistake
	query := h.DB.Where("user_id = ?", userID)
	if hasFrom {
		query = query.Where("mistake_at >= ?", fromTime)
	}
	if hasTo {
		query = query.Where("mistake_at <= ?", toTime)
	}
	if err := query.Order("mistake_at DESC").Find(&mistakes).Error; err != nil {
		return nil, err
	}

	if len(mistakes) == 0 {
		return []MistakeExportItem{}, nil
	}

	// 批量查 Quiz 详情
	quizIDs := make([]string, len(mistakes))
	for i, m := range mistakes {
		quizIDs[i] = m.QuizID
	}
	quizMap := make(map[string]*model.Quiz)
	var quizzes []model.Quiz
	h.DB.Where("id IN ?", quizIDs).Find(&quizzes)
	for i := range quizzes {
		quizMap[quizzes[i].ID] = &quizzes[i]
	}

	// 批量查材料标题
	materialIDs := make([]string, 0)
	for _, q := range quizMap {
		if q.MaterialID != "" {
			materialIDs = append(materialIDs, q.MaterialID)
		}
	}
	materialTitleMap := make(map[string]string)
	if len(materialIDs) > 0 {
		var materials []model.Material
		h.DB.Select("id, title").Where("id IN ?", materialIDs).Find(&materials)
		for _, m := range materials {
			materialTitleMap[m.ID] = m.Title
		}
	}

	items := make([]MistakeExportItem, len(mistakes))
	for i, m := range mistakes {
		q := quizMap[m.QuizID]
		item := MistakeExportItem{
			UserAnswer:    m.UserAnswer,
			CorrectAnswer: m.CorrectAnswer,
			MistakeAt:     m.MistakeAt.Format("2006-01-02 15:04"),
		}
		if q != nil {
			item.Question = q.Question
			item.Type = q.Type
			item.Difficulty = q.Difficulty
			item.Explanation = q.Explanation
			item.MaterialTitle = materialTitleMap[q.MaterialID]
		}
		items[i] = item
	}
	return items, nil
}

// ==================== 格式化输出 ====================

func (h *Handler) exportJSON(c *gin.Context, sections []ExportSection, exportType string) {
	if exportType == "all" {
		// 多类型：返回分区结构
		result := make(map[string]interface{})
		for _, s := range sections {
			result[s.Type] = s.Data
		}
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=studyforge-export-%s.json", time.Now().Format("20060102-150405")))
		c.Header("Content-Type", "application/json; charset=utf-8")
		data, _ := json.MarshalIndent(result, "", "  ")
		c.Data(http.StatusOK, "application/json; charset=utf-8", data)
	} else {
		// 单类型：直接返回数组
		if len(sections) > 0 {
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=studyforge-%s-%s.json", exportType, time.Now().Format("20060102-150405")))
			c.Header("Content-Type", "application/json; charset=utf-8")
			data, _ := json.MarshalIndent(sections[0].Data, "", "  ")
			c.Data(http.StatusOK, "application/json; charset=utf-8", data)
		}
	}
}

func (h *Handler) exportCSV(c *gin.Context, sections []ExportSection, exportType string) {
	var filename string
	if exportType == "all" {
		filename = fmt.Sprintf("studyforge-export-%s.csv", time.Now().Format("20060102-150405"))
	} else {
		filename = fmt.Sprintf("studyforge-%s-%s.csv", exportType, time.Now().Format("20060102-150405"))
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "text/csv; charset=utf-8")

	// UTF-8 BOM 头（兼容 Excel 中文显示）
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	for i, section := range sections {
		if i > 0 {
			// 多个 section 之间空一行
			writer.Write([]string{""})
		}

		switch section.Type {
		case "cards":
			if items, ok := section.Data.([]CardExportItem); ok {
				writer.Write([]string{"# 知识卡片"})
				writer.Write([]string{"概念", "详情", "标签", "难度", "复习次数", "难度系数", "下次复习", "书签", "笔记", "所属材料", "创建时间"})
				for _, item := range items {
					bookmarked := "否"
					if item.IsBookmarked {
						bookmarked = "是"
					}
					writer.Write([]string{
						item.Concept, item.Detail, item.Tags, item.Difficulty,
						fmt.Sprintf("%d", item.ReviewCount),
						fmt.Sprintf("%.2f", item.EaseFactor),
						item.NextReviewAt, bookmarked, item.UserNote,
						item.MaterialTitle, item.CreatedAt,
					})
				}
			}
		case "quizzes":
			if items, ok := section.Data.([]QuizExportItem); ok {
				writer.Write([]string{"# 练习题"})
				writer.Write([]string{"题目", "题型", "难度", "答案", "解析", "提示1", "提示2", "提示3", "所属材料", "创建时间"})
				for _, item := range items {
					writer.Write([]string{
						item.Question, exportTypeLabel(item.Type), item.Difficulty,
						item.Answer, item.Explanation,
						item.Hint1, item.Hint2, item.Hint3,
						item.MaterialTitle, item.CreatedAt,
					})
				}
			}
		case "chats":
			if items, ok := section.Data.([]ChatExportItem); ok {
				writer.Write([]string{"# 对话记录"})
				writer.Write([]string{"会话标题", "消息内容", "角色", "时间"})
				for _, item := range items {
					role := item.Role
					if role == "user" {
						role = "用户"
					} else if role == "assistant" {
						role = "AI"
					}
					writer.Write([]string{
						item.ConversationTitle, item.Content, role, item.CreatedAt,
					})
				}
			}
		case "mistakes":
			if items, ok := section.Data.([]MistakeExportItem); ok {
				writer.Write([]string{"# 错题记录"})
				writer.Write([]string{"题目", "题型", "难度", "我的答案", "正确答案", "解析", "所属材料", "错误时间"})
				for _, item := range items {
					writer.Write([]string{
						item.Question, exportTypeLabel(item.Type), item.Difficulty,
						item.UserAnswer, item.CorrectAnswer, item.Explanation,
						item.MaterialTitle, item.MistakeAt,
					})
				}
			}
		}
	}
}

// ==================== 辅助函数 ====================

func parseTypeSet(t string) map[string]bool {
	set := make(map[string]bool)
	if t == "all" || t == "" {
		set["cards"] = true
		set["quizzes"] = true
		set["chats"] = true
		set["mistakes"] = true
		return set
	}
	for _, item := range strings.Split(t, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			set[item] = true
		}
	}
	return set
}

func exportTypeLabel(t string) string {
	switch t {
	case "choice":
		return "选择题"
	case "fill":
		return "填空题"
	case "judge":
		return "判断题"
	case "short_answer":
		return "简答题"
	default:
		return t
	}
}
