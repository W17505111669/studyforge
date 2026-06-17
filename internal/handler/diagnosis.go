package handler

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"studyforge/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== AI 知识弱点诊断 ====================

// DiagnosisWeakPoint 单个弱点
type DiagnosisWeakPoint struct {
	ID          string `json:"id"`
	Dimension   string `json:"dimension"` // card/tag/material/type/difficulty
	Label       string `json:"label"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion"`
	ActionURL   string `json:"action_url"`
	Severity    string `json:"severity"` // high/medium/low
	Metric      string `json:"metric"`
}

// RadarDimension 雷达图维度数据
type RadarDimension struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"` // 0-100
}

// DiagnosisResponse 诊断报告响应
type DiagnosisResponse struct {
	RadarDimensions []RadarDimension     `json:"radar_dimensions"`
	WeakPoints      []DiagnosisWeakPoint `json:"weak_points"`
	OverallScore    float64              `json:"overall_score"`
	CardStats       struct {
		Total       int64 `json:"total"`
		Weak        int   `json:"weak"`
		NotReviewed int64 `json:"not_reviewed"`
	} `json:"card_stats"`
	QuizStats struct {
		TotalAttempts int64   `json:"total_attempts"`
		Accuracy      float64 `json:"accuracy"`
	} `json:"quiz_stats"`
	TypeAccuracy  map[string]float64 `json:"type_accuracy"`
	DifficultyAcc map[string]float64 `json:"difficulty_accuracy"`
}

// diagStat 诊断统计辅助结构
type diagStat struct {
	total   int
	correct int
}

// GetDiagnosis 获取多维度知识弱点诊断
// GET /api/diagnosis
func (h *Handler) GetDiagnosis(c *gin.Context) {
	userID := c.GetString("userID")

	resp := &DiagnosisResponse{
		WeakPoints:    []DiagnosisWeakPoint{},
		TypeAccuracy:  make(map[string]float64),
		DifficultyAcc: make(map[string]float64),
	}

	// ===== 维度1: 卡片薄弱知识点 (review_count>=2 且 ease_factor<2.0) =====
	var weakCards []model.Card
	h.DB.Where("user_id = ? AND review_count >= 2 AND ease_factor < 2.0", userID).
		Order("ease_factor ASC").Limit(5).Find(&weakCards)

	for _, card := range weakCards {
		severity := "medium"
		if card.EaseFactor < 1.5 {
			severity = "high"
		} else if card.EaseFactor >= 1.8 {
			severity = "low"
		}
		resp.WeakPoints = append(resp.WeakPoints, DiagnosisWeakPoint{
			ID:          card.ID,
			Dimension:   "card",
			Label:       card.Concept,
			Description: fmt.Sprintf("复习 %d 次仍困难，难度系数 %.2f（越低越难）", card.ReviewCount, card.EaseFactor),
			Suggestion:  "建议使用「不熟」按钮重置间隔，配合记忆技巧反复练习",
			ActionURL:   "/cards",
			Severity:    severity,
			Metric:      fmt.Sprintf("复习%d次 · 系数%.2f", card.ReviewCount, card.EaseFactor),
		})
	}

	// ===== 维度2: 按材料标签的正确率 (quiz_attempts -> quizzes -> materials.tags) =====
	type TagAttempt struct {
		MaterialTags string
		IsCorrect    bool
	}
	var tagAttempts []TagAttempt
	h.DB.Table("quiz_attempts").
		Select("COALESCE(materials.tags, '') as material_tags, quiz_attempts.is_correct").
		Joins("JOIN quizzes ON quizzes.id = quiz_attempts.quiz_id").
		Joins("LEFT JOIN materials ON materials.id = quizzes.material_id").
		Where("quiz_attempts.user_id = ?", userID).
		Find(&tagAttempts)

	tagStatMap := make(map[string]*diagStat)
	for _, ta := range tagAttempts {
		if ta.MaterialTags == "" {
			continue
		}
		for _, tag := range strings.Split(ta.MaterialTags, ",") {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			if _, ok := tagStatMap[tag]; !ok {
				tagStatMap[tag] = &diagStat{}
			}
			tagStatMap[tag].total++
			if ta.IsCorrect {
				tagStatMap[tag].correct++
			}
		}
	}
	type weakTag struct {
		tag   string
		rate  float64
		total int
	}
	var weakTags []weakTag
	for tag, s := range tagStatMap {
		if s.total < 2 {
			continue
		}
		rate := float64(s.correct) / float64(s.total)
		if rate < 0.6 {
			weakTags = append(weakTags, weakTag{tag: tag, rate: rate, total: s.total})
		}
	}
	sort.Slice(weakTags, func(i, j int) bool { return weakTags[i].rate < weakTags[j].rate })
	for i, wt := range weakTags {
		if i >= 5 {
			break
		}
		pct := int(wt.rate * 100)
		resp.WeakPoints = append(resp.WeakPoints, DiagnosisWeakPoint{
			Dimension:   "tag",
			Label:       wt.tag,
			Description: fmt.Sprintf("该标签下答题正确率仅 %d%%（共 %d 题）", pct, wt.total),
			Suggestion:  fmt.Sprintf("建议复习「%s」相关材料的知识卡片，加深理解", wt.tag),
			ActionURL:   "/cards",
			Severity:    diagSeverityFromRate(wt.rate),
			Metric:      fmt.Sprintf("正确率%d%% · %d题", pct, wt.total),
		})
	}

	// ===== 维度3: 未充分学习的材料（复习率低） =====
	type MaterialReviewInfo struct {
		ID        string
		Title     string
		CardCount int64
	}
	var matInfos []MaterialReviewInfo
	h.DB.Table("materials").
		Select("materials.id, materials.title, COUNT(cards.id) as card_count").
		Joins("LEFT JOIN cards ON cards.material_id = materials.id AND cards.user_id = ?", userID).
		Where("materials.user_id = ? AND materials.status IN ?", userID, []string{"completed", "partial"}).
		Group("materials.id, materials.title").
		Having("COUNT(cards.id) > 0").
		Find(&matInfos)

	type weakMat struct {
		id, title  string
		reviewRate float64
		cardCount  int64
	}
	var weakMats []weakMat
	for _, m := range matInfos {
		if m.CardCount == 0 {
			continue
		}
		var reviewed int64
		h.DB.Model(&model.Card{}).
			Where("material_id = ? AND user_id = ? AND review_count > 0", m.ID, userID).
			Count(&reviewed)
		rate := float64(reviewed) / float64(m.CardCount)
		if rate < 0.5 {
			weakMats = append(weakMats, weakMat{m.ID, m.Title, rate, m.CardCount})
		}
	}
	sort.Slice(weakMats, func(i, j int) bool { return weakMats[i].reviewRate < weakMats[j].reviewRate })
	for i, wm := range weakMats {
		if i >= 3 {
			break
		}
		pct := int(wm.reviewRate * 100)
		resp.WeakPoints = append(resp.WeakPoints, DiagnosisWeakPoint{
			ID:          wm.id,
			Dimension:   "material",
			Label:       wm.title,
			Description: fmt.Sprintf("%d 张卡片中仅 %d%% 已复习", wm.cardCount, pct),
			Suggestion:  "建议开始学习该材料的知识卡片，建立基础记忆",
			ActionURL:   fmt.Sprintf("/materials/%s", wm.id),
			Severity:    diagSeverityFromRate(wm.reviewRate),
			Metric:      fmt.Sprintf("复习率%d%% · %d张卡片", pct, wm.cardCount),
		})
	}

	// ===== 维度4: 题型维度正确率 =====
	type QuizTypeAttempt struct {
		QuizType  string
		IsCorrect bool
	}
	var typeAttempts []QuizTypeAttempt
	h.DB.Table("quiz_attempts").
		Select("quizzes.type as quiz_type, quiz_attempts.is_correct").
		Joins("JOIN quizzes ON quizzes.id = quiz_attempts.quiz_id").
		Where("quiz_attempts.user_id = ?", userID).
		Find(&typeAttempts)

	typeStatMap := make(map[string]*diagStat)
	for _, a := range typeAttempts {
		if _, ok := typeStatMap[a.QuizType]; !ok {
			typeStatMap[a.QuizType] = &diagStat{}
		}
		typeStatMap[a.QuizType].total++
		if a.IsCorrect {
			typeStatMap[a.QuizType].correct++
		}
	}
	typeLabels := map[string]string{"choice": "选择题", "fill": "填空题", "short_answer": "简答题", "judge": "判断题"}
	for t, s := range typeStatMap {
		if s.total >= 2 {
			rate := float64(s.correct) / float64(s.total)
			resp.TypeAccuracy[t] = math.Round(rate*10000) / 100
			if rate < 0.6 {
				pct := int(rate * 100)
				label := typeLabels[t]
				if label == "" {
					label = t
				}
				resp.WeakPoints = append(resp.WeakPoints, DiagnosisWeakPoint{
					Dimension:   "type",
					Label:       label,
					Description: fmt.Sprintf("正确率仅 %d%%（共 %d 题）", pct, s.total),
					Suggestion:  fmt.Sprintf("建议进行%s专项练习，熟悉出题规律", label),
					ActionURL:   "/quiz",
					Severity:    diagSeverityFromRate(rate),
					Metric:      fmt.Sprintf("正确率%d%% · %d题", pct, s.total),
				})
			}
		}
	}

	// ===== 维度5: 难度维度正确率 =====
	type DiffAttempt struct {
		Difficulty string
		IsCorrect  bool
	}
	var diffAttempts []DiffAttempt
	h.DB.Table("quiz_attempts").
		Select("quizzes.difficulty as difficulty, quiz_attempts.is_correct").
		Joins("JOIN quizzes ON quizzes.id = quiz_attempts.quiz_id").
		Where("quiz_attempts.user_id = ?", userID).
		Find(&diffAttempts)

	diffStatMap := make(map[string]*diagStat)
	for _, a := range diffAttempts {
		if _, ok := diffStatMap[a.Difficulty]; !ok {
			diffStatMap[a.Difficulty] = &diagStat{}
		}
		diffStatMap[a.Difficulty].total++
		if a.IsCorrect {
			diffStatMap[a.Difficulty].correct++
		}
	}
	diffLabels := map[string]string{"easy": "简单", "medium": "中等", "hard": "困难"}
	for d, s := range diffStatMap {
		if s.total >= 2 {
			rate := float64(s.correct) / float64(s.total)
			resp.DifficultyAcc[d] = math.Round(rate*10000) / 100
			if rate < 0.6 {
				pct := int(rate * 100)
				label := diffLabels[d]
				if label == "" {
					label = d
				}
				resp.WeakPoints = append(resp.WeakPoints, DiagnosisWeakPoint{
					Dimension:   "difficulty",
					Label:       fmt.Sprintf("%s难度", label),
					Description: fmt.Sprintf("正确率仅 %d%%（共 %d 题）", pct, s.total),
					Suggestion:  fmt.Sprintf("建议回顾%s难度的相关知识点，加强基础理解", label),
					ActionURL:   "/quiz",
					Severity:    diagSeverityFromRate(rate),
					Metric:      fmt.Sprintf("正确率%d%% · %d题", pct, s.total),
				})
			}
		}
	}

	// ===== 统计摘要 =====
	h.DB.Model(&model.Card{}).Where("user_id = ?", userID).Count(&resp.CardStats.Total)
	var notReviewed int64
	h.DB.Model(&model.Card{}).
		Where("user_id = ? AND (next_review_at IS NULL OR next_review_at <= ?)", userID, time.Now()).
		Count(&notReviewed)
	resp.CardStats.NotReviewed = notReviewed
	resp.CardStats.Weak = len(weakCards)

	var totalAttempts int64
	h.DB.Model(&model.QuizAttempt{}).Where("user_id = ?", userID).Count(&totalAttempts)
	resp.QuizStats.TotalAttempts = totalAttempts
	if totalAttempts > 0 {
		var correctAttempts int64
		h.DB.Model(&model.QuizAttempt{}).
			Where("user_id = ? AND is_correct = ?", userID, true).
			Count(&correctAttempts)
		resp.QuizStats.Accuracy = math.Round(float64(correctAttempts)/float64(totalAttempts)*10000) / 100
	}

	// ===== 雷达图五维度得分 =====
	memoryScore := h.calcMemoryScore(userID)
	understandingScore := resp.QuizStats.Accuracy
	if understandingScore == 0 && totalAttempts == 0 {
		understandingScore = 50
	}
	applicationScore := diagCalcGroupScore(diffStatMap, []string{"medium", "hard"})
	analysisScore := diagCalcGroupScore(typeStatMap, []string{"fill", "short_answer"})
	synthesisScore := 0.3*memoryScore + 0.25*understandingScore + 0.2*applicationScore + 0.25*analysisScore

	if totalAttempts == 0 {
		applicationScore = 50
		analysisScore = 50
		synthesisScore = 0.3*memoryScore + 0.7*50
	}

	resp.RadarDimensions = []RadarDimension{
		{Name: "记忆", Score: math.Round(memoryScore*100) / 100},
		{Name: "理解", Score: math.Round(understandingScore*100) / 100},
		{Name: "应用", Score: math.Round(applicationScore*100) / 100},
		{Name: "分析", Score: math.Round(analysisScore*100) / 100},
		{Name: "综合", Score: math.Round(synthesisScore*100) / 100},
	}
	resp.OverallScore = math.Round(synthesisScore*100) / 100

	c.JSON(http.StatusOK, gin.H{"diagnosis": resp})
}

// calcMemoryScore 记忆力得分：ease_factor>=2.3 的卡片比例
func (h *Handler) calcMemoryScore(userID string) float64 {
	var cards []model.Card
	h.DB.Select("ease_factor, review_count").
		Where("user_id = ? AND review_count > 0", userID).
		Find(&cards)
	if len(cards) == 0 {
		return 50
	}
	var strong int
	for _, c := range cards {
		if c.EaseFactor >= 2.3 {
			strong++
		}
	}
	return float64(strong) / float64(len(cards)) * 100
}

// diagCalcGroupScore 按指定 key 子集计算得分
func diagCalcGroupScore(statMap map[string]*diagStat, keys []string) float64 {
	var total, correct int
	for _, k := range keys {
		if s, ok := statMap[k]; ok {
			total += s.total
			correct += s.correct
		}
	}
	if total < 2 {
		return 50
	}
	return float64(correct) / float64(total) * 100
}

// diagSeverityFromRate 根据正确率判断严重程度
func diagSeverityFromRate(rate float64) string {
	if rate < 0.3 {
		return "high"
	} else if rate < 0.5 {
		return "medium"
	}
	return "low"
}
