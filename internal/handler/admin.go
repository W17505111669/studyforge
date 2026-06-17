package handler

import (
	"net/http"
	"strconv"
	"strings"

	"studyforge/internal/middleware"

	"github.com/gin-gonic/gin"
)

// ==================== Admin 数据库诊断 ====================

// GetDBStats 获取数据库统计信息（各表行数、索引列表）
// GET /api/admin/db-stats
func (h *Handler) GetDBStats(c *gin.Context) {
	// 各表行数统计
	type TableStat struct {
		Table string `json:"table"`
		Count int64  `json:"count"`
	}

	// 需要统计的所有表
	tableNames := []string{
		"users",
		"materials",
		"cards",
		"quizzes",
		"quiz_attempts",
		"llm_traces",
		"conversations",
		"chat_messages",
		"user_achievements",
		"quiz_mistakes",
		"notifications",
		"pomodoro_sessions",
		"learning_goals",
		"study_streaks",
		"daily_tasks",
		"notes",
		"note_folders",
		"decks",
		"deck_cards",
		"friendships",
		"study_groups",
		"group_members",
		"group_goals",
		"exam_sessions",
		"explain_caches",
	}

	tableStats := make([]TableStat, 0, len(tableNames))
	for _, table := range tableNames {
		var count int64
		h.DB.Table(table).Count(&count)
		tableStats = append(tableStats, TableStat{
			Table: table,
			Count: count,
		})
	}

	// 索引列表（从 SQLite 元数据获取）
	type IndexInfo struct {
		Table   string `json:"table"`
		Name    string `json:"name"`
		Unique  bool   `json:"unique"`
		Columns string `json:"columns"`
	}

	var indexes []IndexInfo

	// Step 1: 从 sqlite_master 获取所有用户创建的索引
	type sqliteIndexRow struct {
		TblName string
		Name    string
		SQL     string
	}
	var rawIndexes []sqliteIndexRow
	h.DB.Raw(`
		SELECT tbl_name, name, sql
		FROM sqlite_master
		WHERE type = 'index'
		AND name NOT LIKE 'sqlite_autoindex_%'
		ORDER BY tbl_name, name
	`).Scan(&rawIndexes)

	// Step 2: 对每个索引获取列信息
	for _, ri := range rawIndexes {
		type indexColumn struct {
			Cid  int
			Name string
		}
		var cols []indexColumn
		h.DB.Raw("PRAGMA index_info(?)", ri.Name).Scan(&cols)

		colNames := ""
		for i, col := range cols {
			if i > 0 {
				colNames += ", "
			}
			colNames += col.Name
		}

		// 通过 SQL 定义判断是否唯一索引
		isUnique := strings.Contains(strings.ToUpper(ri.SQL), "UNIQUE")

		indexes = append(indexes, IndexInfo{
			Table:   ri.TblName,
			Name:    ri.Name,
			Unique:  isUnique,
			Columns: colNames,
		})
	}

	// 数据库文件大小（通过 page_count * page_size 计算）
	var pageCount, pageSize int64
	h.DB.Raw("PRAGMA page_count").Scan(&pageCount)
	h.DB.Raw("PRAGMA page_size").Scan(&pageSize)
	dbSizeBytes := pageCount * pageSize

	// 数据库总行数
	var totalRows int64
	for _, ts := range tableStats {
		totalRows += ts.Count
	}

	c.JSON(http.StatusOK, gin.H{
		"tables":      tableStats,
		"indexes":     indexes,
		"total_tables": len(tableStats),
		"total_indexes": len(indexes),
		"total_rows":  totalRows,
		"db_size_bytes": dbSizeBytes,
		"db_size_mb":   float64(dbSizeBytes) / (1024 * 1024),
	})
}

// ==================== API 性能监控 ====================

// GetAPIMetrics 获取 API 性能监控数据
// GET /api/admin/metrics?range=1|24|168
func (h *Handler) GetAPIMetrics(c *gin.Context) {
	rangeHours := 24 // 默认 24 小时
	if r := c.Query("range"); r != "" {
		if v, err := strconv.Atoi(r); err == nil && v > 0 {
			rangeHours = v
		}
	}

	// 限制范围上限为 7 天
	if rangeHours > 168 {
		rangeHours = 168
	}

	summary := middleware.ComputeSummary(rangeHours)
	c.JSON(http.StatusOK, summary)
}
