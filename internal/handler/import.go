package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

const maxApkgSize = 50 << 20 // 50MB

// AnkiPreviewCard Anki 导入预览卡片
type AnkiPreviewCard struct {
	Index   int      `json:"index"`
	Concept string   `json:"concept"`
	Detail  string   `json:"detail"`
	Tags    string   `json:"tags"`
	Fields  []string `json:"fields"`
}

// AnkiImportRequest Anki 导入确认请求
type AnkiImportRequest struct {
	Cards      []AnkiPreviewCard `json:"cards" binding:"required"`
	MaterialID string            `json:"material_id"` // 可选：归入已有材料
}

// ImportAnkiPreview 解析 .apkg 文件并返回卡片预览
// POST /api/import/anki
func (h *Handler) ImportAnkiPreview(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择 .apkg 文件"})
		return
	}
	defer file.Close()

	// 文件大小检查
	if header.Size > maxApkgSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件过大，最大支持 50MB"})
		return
	}

	// 扩展名检查
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".apkg") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 .apkg 格式的 Anki 牌组文件"})
		return
	}

	// 保存到临时文件
	tmpDir := os.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "anki-*.apkg")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建临时文件失败"})
		return
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // 函数退出时清理

	if _, err := io.Copy(tmpFile, file); err != nil {
		tmpFile.Close()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}
	tmpFile.Close()

	// 解析 .apkg
	cards, modelName, fieldNames, parseErr := parseApkg(tmpPath)
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("解析 .apkg 文件失败: %v", parseErr)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cards":       cards,
		"total":       len(cards),
		"model_name":  modelName,
		"field_names": fieldNames,
		"filename":    header.Filename,
	})
}

// ImportAnkiConfirm 确认导入选中的 Anki 卡片
// POST /api/import/anki/confirm
func (h *Handler) ImportAnkiConfirm(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req AnkiImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if len(req.Cards) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有选择任何卡片"})
		return
	}

	// 限制最大导入数量
	if len(req.Cards) > 5000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "单次导入最多 5000 张卡片"})
		return
	}

	// 确定目标材料
	materialID := req.MaterialID
	if materialID == "" {
		// 创建默认导入材料
		material := model.Material{
			UserID:      userID,
			Title:       "Anki 导入牌组",
			Content:     "从 Anki .apkg 文件导入的卡片",
			Status:      "completed",
			ContentType: "anki",
		}
		if err := h.DB.Create(&material).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建导入材料失败"})
			return
		}
		materialID = material.ID
	} else {
		// 验证材料归属
		var count int64
		h.DB.Model(&model.Material{}).Where("id = ? AND user_id = ?", materialID, userID).Count(&count)
		if count == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "目标材料不存在"})
			return
		}
	}

	// 批量创建卡片
	var created int
	for _, ac := range req.Cards {
		concept := strings.TrimSpace(ac.Concept)
		if concept == "" {
			continue
		}

		tags := strings.TrimSpace(ac.Tags)
		if tags != "" {
			tags = "anki," + tags
		} else {
			tags = "anki"
		}

		card := model.Card{
			UserID:     userID,
			MaterialID: materialID,
			Concept:    concept,
			Detail:     strings.TrimSpace(ac.Detail),
			Tags:       tags,
			Difficulty: "medium",
		}
		if err := h.DB.Create(&card).Error; err != nil {
			log.Printf("导入 Anki 卡片失败: %v", err)
			continue
		}
		created++
	}

	c.JSON(http.StatusOK, gin.H{
		"imported":   created,
		"total":      len(req.Cards),
		"material_id": materialID,
	})
}

// ==================== .apkg 解析逻辑 ====================

// parseApkg 解析 Anki .apkg 文件（SQLite 数据库）
// 返回：预览卡片列表、模型名称、字段名称列表、错误
func parseApkg(path string) ([]AnkiPreviewCard, string, []string, error) {
	dsn := fmt.Sprintf("file:%s?mode=ro&immutable=1", path)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, "", nil, fmt.Errorf("打开 .apkg 失败: %w", err)
	}
	defer db.Close()

	// 连通性测试
	if err := db.Ping(); err != nil {
		return nil, "", nil, fmt.Errorf("连接 .apkg 数据库失败: %w", err)
	}

	// 获取模型定义
	modelName, fieldNames, err := getAnkiModels(db)
	if err != nil {
		return nil, "", nil, err
	}

	// 解析笔记（卡片）
	cards, err := getAnkiNotes(db, fieldNames)
	if err != nil {
		return nil, "", nil, err
	}

	if len(cards) == 0 {
		return nil, "", nil, fmt.Errorf("牌组中没有找到卡片")
	}

	return cards, modelName, fieldNames, nil
}

// getAnkiModels 从 col 表获取 Anki 模型定义
func getAnkiModels(db *sql.DB) (string, []string, error) {
	var modelsJSON string
	err := db.QueryRow("SELECT models FROM col").Scan(&modelsJSON)
	if err != nil {
		// 尝试旧版 Anki 格式（models 在 JSON 字段中）
		return "Basic", []string{"Front", "Back"}, nil
	}

	var models map[string]json.RawMessage
	if err := json.Unmarshal([]byte(modelsJSON), &models); err != nil {
		// JSON 解析失败时使用默认字段名
		return "Basic", []string{"Front", "Back"}, nil
	}

	// 获取第一个模型
	for _, modelRaw := range models {
		var m struct {
			Name   string `json:"name"`
			Flds   []struct {
				Name string `json:"name"`
				Ord  int    `json:"ord"`
			} `json:"flds"`
		}
		if err := json.Unmarshal(modelRaw, &m); err != nil {
			continue
		}
		if len(m.Flds) == 0 {
			continue
		}

		// 按 ord 排序
		for i := 0; i < len(m.Flds); i++ {
			for j := i + 1; j < len(m.Flds); j++ {
				if m.Flds[j].Ord < m.Flds[i].Ord {
					m.Flds[i], m.Flds[j] = m.Flds[j], m.Flds[i]
				}
			}
		}

		var names []string
		for _, f := range m.Flds {
			names = append(names, f.Name)
		}
		return m.Name, names, nil
	}

	return "Basic", []string{"Front", "Back"}, nil
}

// getAnkiNotes 从 notes 表解析卡片数据
func getAnkiNotes(db *sql.DB, fieldNames []string) ([]AnkiPreviewCard, error) {
	rows, err := db.Query("SELECT flds, tags FROM notes")
	if err != nil {
		return nil, fmt.Errorf("查询笔记失败: %w", err)
	}
	defer rows.Close()

	var cards []AnkiPreviewCard
	idx := 0

	for rows.Next() {
		var flds, tags string
		if err := rows.Scan(&flds, &tags); err != nil {
			continue
		}

		// Anki 字段分隔符：\x1f (Unit Separator)
		fields := strings.Split(flds, "\x1f")

		// 映射字段名
		namedFields := make(map[string]string)
		for i, f := range fields {
			if i < len(fieldNames) {
				namedFields[fieldNames[i]] = f
			}
		}

		// concept = 第一个字段（正面/Front）
		concept := ""
		if len(fields) > 0 {
			concept = stripHTML(fields[0])
		}

		// detail = 第二个字段及之后（背面/Back + 额外字段）
		detail := buildDetail(fields, fieldNames)

		// 标签：Anki 存储为 " tag1 tag2 " 格式（空格分隔，前后有空格）
		tagStr := strings.TrimSpace(strings.ReplaceAll(tags, " ", ", "))

		if idx >= 5000 {
			break // 安全上限
		}

		cards = append(cards, AnkiPreviewCard{
			Index:   idx,
			Concept: concept,
			Detail:  detail,
			Tags:    tagStr,
			Fields:  fields,
		})
		idx++
	}

	return cards, rows.Err()
}

// buildDetail 构建卡片背面内容
func buildDetail(fields []string, fieldNames []string) string {
	if len(fields) <= 1 {
		return ""
	}

	var parts []string
	for i := 1; i < len(fields); i++ {
		content := stripHTML(fields[i])
		if content == "" {
			continue
		}
		if i < len(fieldNames) {
			parts = append(parts, fmt.Sprintf("**%s:** %s", fieldNames[i], content))
		} else {
			parts = append(parts, content)
		}
	}
	return strings.Join(parts, "\n\n")
}

// stripHTML 去除 HTML 标签并解码基本实体
func stripHTML(s string) string {
	// 将 <br> / <br/> / <div> 转为换行
	re := regexp.MustCompile(`(?i)<br\s*/?>|</div>|</p>|</li>`)
	s = re.ReplaceAllString(s, "\n")
	// 去除所有 HTML 标签
	re = regexp.MustCompile(`<[^>]+>`)
	s = re.ReplaceAllString(s, "")
	// 解码基本 HTML 实体
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	// 清理多余空行
	re = regexp.MustCompile(`\n{3,}`)
	s = re.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}
