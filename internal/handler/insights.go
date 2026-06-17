package handler

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// Connection 跨材料知识关联
type Connection struct {
	MaterialA       MaterialBrief   `json:"material_a"`
	MaterialB       MaterialBrief   `json:"material_b"`
	SharedConcepts  []SharedConcept `json:"shared_concepts"`
	SimilarityScore float64         `json:"similarity_score"`
}

// MaterialBrief 材料简要信息
type MaterialBrief struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Tags  string `json:"tags"`
}

// SharedConcept 共同知识点
type SharedConcept struct {
	ConceptA string `json:"concept_a"` // 材料 A 中的概念
	ConceptB string `json:"concept_b"` // 材料 B 中的概念
	MatchType string `json:"match_type"` // exact / contains / related
}

// InsightsResponse 知识洞察响应
type InsightsResponse struct {
	Connections  []Connection     `json:"connections"`
	Materials    []MaterialBrief  `json:"materials"`
	TotalPairs   int              `json:"total_pairs"`
	StrongCount  int              `json:"strong_count"`  // 强关联数 (score >= 0.5)
	MediumCount  int              `json:"medium_count"`  // 中关联数 (0.2 <= score < 0.5)
	TopConcepts  []ConceptFreq    `json:"top_concepts"`  // 跨材料出现最多的概念
}

// ConceptFreq 概念出现频率
type ConceptFreq struct {
	Concept      string   `json:"concept"`
	MaterialCount int     `json:"material_count"`
	MaterialTitles []string `json:"material_titles"`
}

// analysisDataSimple 用于解析 analysis_data 的简化结构
type analysisDataSimple struct {
	Summary   string `json:"summary"`
	KeyPoints []struct {
		Concept    string `json:"concept"`
		Detail     string `json:"detail"`
		Difficulty string `json:"difficulty"`
	} `json:"key_points"`
	Relationships []struct {
		From string `json:"from"`
		To   string `json:"to"`
		Type string `json:"type"`
	} `json:"relationships"`
}

// materialConcepts 材料概念集合（用于跨材料关联计算）
type materialConcepts struct {
	brief    MaterialBrief
	concepts []string          // 所有概念名（小写归一化）
	original map[string]string // 小写 -> 原始概念名
}

// GetConnections 获取跨材料知识关联
func (h *Handler) GetConnections(c *gin.Context) {
	userID := c.GetString("user_id")

	// 查询用户所有已完成且有分析数据的材料
	var materials []struct {
		ID           string
		Title        string
		Tags         string
		AnalysisData string
	}

	err := h.DB.Table("materials").
		Select("id, title, tags, analysis_data").
		Where("user_id = ? AND status = ? AND analysis_data IS NOT NULL AND analysis_data != ''", userID, "completed").
		Find(&materials).Error

	if err != nil {
		log.Printf("GetConnections: 查询材料失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询材料失败"})
		return
	}

	if len(materials) < 2 {
		c.JSON(http.StatusOK, InsightsResponse{
			Connections: []Connection{},
			Materials:   []MaterialBrief{},
			TotalPairs:  0,
			StrongCount: 0,
			MediumCount: 0,
			TopConcepts: []ConceptFreq{},
		})
		return
	}

	// 解析每个材料的知识点
	var parsed []materialConcepts
	var materialBriefs []MaterialBrief

	for _, m := range materials {
		var data analysisDataSimple
		if err := json.Unmarshal([]byte(m.AnalysisData), &data); err != nil {
			continue // 跳过解析失败的材料
		}

		brief := MaterialBrief{ID: m.ID, Title: m.Title, Tags: m.Tags}
		materialBriefs = append(materialBriefs, brief)

		concepts := make([]string, 0, len(data.KeyPoints))
		original := make(map[string]string)
		for _, kp := range data.KeyPoints {
			if kp.Concept == "" {
				continue
			}
			lower := strings.ToLower(strings.TrimSpace(kp.Concept))
			concepts = append(concepts, lower)
			original[lower] = kp.Concept
		}

		// 也把 relationships 中的概念加入
		for _, rel := range data.Relationships {
			for _, name := range []string{rel.From, rel.To} {
				if name == "" {
					continue
				}
				lower := strings.ToLower(strings.TrimSpace(name))
				if _, exists := original[lower]; !exists {
					concepts = append(concepts, lower)
					original[lower] = name
				}
			}
		}

		parsed = append(parsed, materialConcepts{
			brief:    brief,
			concepts: concepts,
			original: original,
		})
	}

	// 计算两两材料间的关联
	connections := make([]Connection, 0)
	// 概念 -> 出现在哪些材料中
	conceptMaterials := make(map[string][]string) // concept_lower -> []materialTitle

	for i := 0; i < len(parsed); i++ {
		for _, c := range parsed[i].concepts {
			conceptMaterials[c] = appendUnique(conceptMaterials[c], parsed[i].brief.Title)
		}
	}

	for i := 0; i < len(parsed); i++ {
		for j := i + 1; j < len(parsed); j++ {
			shared := findSharedConcepts(parsed[i], parsed[j])
			if len(shared) == 0 {
				continue
			}

			// Jaccard 相似度
			unionSize := len(unionSets(parsed[i].concepts, parsed[j].concepts))
			score := 0.0
			if unionSize > 0 {
				score = float64(len(shared)) / float64(unionSize)
			}
			// 增加 shared 数量权重（避免只有 1 个共享概念但集合很小导致分数过高）
			quantityBonus := math.Min(float64(len(shared))*0.05, 0.3)
			finalScore := math.Min(score+quantityBonus, 1.0)
			finalScore = math.Round(finalScore*100) / 100

			connections = append(connections, Connection{
				MaterialA:       parsed[i].brief,
				MaterialB:       parsed[j].brief,
				SharedConcepts:  shared,
				SimilarityScore: finalScore,
			})
		}
	}

	// 按相似度降序排序
	sort.SliceStable(connections, func(a, b int) bool {
		return connections[a].SimilarityScore > connections[b].SimilarityScore
	})

	// 统计强/中关联
	strongCount := 0
	mediumCount := 0
	for _, c := range connections {
		if c.SimilarityScore >= 0.5 {
			strongCount++
		} else if c.SimilarityScore >= 0.2 {
			mediumCount++
		}
	}

	// 计算跨材料出现最多的概念 (top 10)
	type conceptInfo struct {
		name    string
		titles  []string
	}
	conceptInfoMap := make(map[string]*conceptInfo)
	for concept, titles := range conceptMaterials {
		if len(titles) >= 2 {
			origName := ""
			// 找到原始名称
			for _, p := range parsed {
				if on, ok := p.original[concept]; ok {
					origName = on
					break
				}
			}
			if origName == "" {
				origName = concept
			}
			conceptInfoMap[concept] = &conceptInfo{name: origName, titles: titles}
		}
	}

	topConcepts := make([]ConceptFreq, 0, len(conceptInfoMap))
	for _, info := range conceptInfoMap {
		topConcepts = append(topConcepts, ConceptFreq{
			Concept:        info.name,
			MaterialCount:  len(info.titles),
			MaterialTitles: info.titles,
		})
	}
	sort.SliceStable(topConcepts, func(a, b int) bool {
		return topConcepts[a].MaterialCount > topConcepts[b].MaterialCount
	})
	if len(topConcepts) > 10 {
		topConcepts = topConcepts[:10]
	}

	c.JSON(http.StatusOK, InsightsResponse{
		Connections: connections,
		Materials:   materialBriefs,
		TotalPairs:  len(connections),
		StrongCount: strongCount,
		MediumCount: mediumCount,
		TopConcepts: topConcepts,
	})
}

// findSharedConcepts 找出两个材料间的共同概念
func findSharedConcepts(a, b materialConcepts) []SharedConcept {
	bSet := make(map[string]bool)
	for _, c := range b.concepts {
		bSet[c] = true
	}

	seen := make(map[string]bool)
	var shared []SharedConcept

	for _, ca := range a.concepts {
		if seen[ca] {
			continue
		}
		// 精确匹配
		if bSet[ca] {
			origA := a.original[ca]
			origB := b.original[ca]
			shared = append(shared, SharedConcept{
				ConceptA:  origA,
				ConceptB:  origB,
				MatchType: "exact",
			})
			seen[ca] = true
			continue
		}
		// 包含匹配（一个概念名包含另一个）
		for _, cb := range b.concepts {
			if seen[ca] {
				break
			}
			if len(ca) < 2 || len(cb) < 2 {
				continue
			}
			if strings.Contains(ca, cb) || strings.Contains(cb, ca) {
				origA := a.original[ca]
				origB := b.original[cb]
				// 避免与自身匹配
				if strings.ToLower(origA) == strings.ToLower(origB) {
					continue
				}
				shared = append(shared, SharedConcept{
					ConceptA:  origA,
					ConceptB:  origB,
					MatchType: "contains",
				})
				seen[ca] = true
				break
			}
		}
	}

	return shared
}

// unionSets 计算两个字符串切片的并集大小
func unionSets(a, b []string) map[string]bool {
	union := make(map[string]bool)
	for _, s := range a {
		union[s] = true
	}
	for _, s := range b {
		union[s] = true
	}
	return union
}

// appendUnique 向切片追加不重复的元素
func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}
