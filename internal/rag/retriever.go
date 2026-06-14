package rag

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
)

// EmbeddingProvider 向量生成接口（解耦 agent 包，避免循环依赖）
type EmbeddingProvider interface {
	CreateEmbedding(ctx context.Context, text string) ([]float32, error)
}

// VectorStore 内存向量存储（开发阶段替代 Qdrant，接口完全兼容）
// TODO: 后续替换为 Qdrant gRPC 实现，接口保持不变
type VectorStore struct {
	mu         sync.RWMutex
	collection string
	vectorSize int
	points     []vectorPoint
	embedder   EmbeddingProvider // 用于 query embedding
}

type vectorPoint struct {
	ID       string
	Vector   []float32
	Payload  map[string]string
	UserID   string // 所属用户
	MaterialID string
}

// NewVectorStore 创建内存向量存储
func NewVectorStore(host string, port int, collection string, vectorSize int, embedder EmbeddingProvider) (*VectorStore, error) {
	store := &VectorStore{
		collection: collection,
		vectorSize: vectorSize,
		points:     make([]vectorPoint, 0),
		embedder:   embedder,
	}
	return store, nil
}

// EnsureCollection 内存模式无需创建集合，直接返回
func (vs *VectorStore) EnsureCollection(ctx context.Context) error {
	return nil
}

// UpsertPoints 将向量数据写入内存存储
func (vs *VectorStore) UpsertPoints(ctx context.Context, materialID string, chunks []Chunk, vectors [][]float32) error {
	if len(chunks) != len(vectors) {
		return fmt.Errorf("chunks 和 vectors 数量不匹配: %d vs %d", len(chunks), len(vectors))
	}

	vs.mu.Lock()
	defer vs.mu.Unlock()

	for i := range chunks {
		point := vectorPoint{
			ID:     fmt.Sprintf("%s-chunk-%d", materialID, i),
			Vector: vectors[i],
			Payload: map[string]string{
				"material_id": materialID,
				"content":     chunks[i].Content,
				"chunk_index": fmt.Sprintf("%d", chunks[i].Index),
			},
			MaterialID: materialID,
			UserID:     chunks[i].Metadata["user_id"],
		}
		vs.points = append(vs.points, point)
	}

	return nil
}

// IndexMaterial 一体化索引：将材料内容切分 → 向量化 → 存入向量库
// chunkSize: 每个 chunk 的目标字符数, overlap: 重叠字符数
func (vs *VectorStore) IndexMaterial(ctx context.Context, materialID, userID, content string, chunkSize, overlap int) error {
	// 1. 切分文档
	chunks := ChunkDocument(content, chunkSize, overlap)
	if len(chunks) == 0 {
		log.Printf("材料 %s 内容为空，跳过索引", materialID)
		return nil
	}

	// 给每个 chunk 添加 user_id 元数据
	for i := range chunks {
		if chunks[i].Metadata == nil {
			chunks[i].Metadata = make(map[string]string)
		}
		chunks[i].Metadata["user_id"] = userID
		chunks[i].Metadata["material_id"] = materialID
	}

	// 2. 向量化（使用 EmbeddingProvider 接口）
	vectors := make([][]float32, len(chunks))
	for i, chunk := range chunks {
		vec, err := vs.embedder.CreateEmbedding(ctx, chunk.Content)
		if err != nil {
			return fmt.Errorf("材料 %s chunk %d embedding 失败: %w", materialID, i, err)
		}
		vectors[i] = vec
	}

	// 3. 存入向量库
	if err := vs.UpsertPoints(ctx, materialID, chunks, vectors); err != nil {
		return fmt.Errorf("材料 %s 写入向量库失败: %w", materialID, err)
	}

	log.Printf("材料 %s 索引完成: %d 个 chunks", materialID, len(chunks))
	return nil
}

// Search 语义检索：余弦相似度匹配最相关的文档片段
func (vs *VectorStore) Search(ctx context.Context, queryVector []float32, materialID string, topK int) ([]SearchResult, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	type scored struct {
		content    string
		score      float32
		materialID string
	}

	var candidates []scored
	for _, p := range vs.points {
		// 按 material_id 过滤（如果指定）
		if materialID != "" && p.MaterialID != materialID {
			continue
		}
		score := cosineSimilarity(queryVector, p.Vector)
		candidates = append(candidates, scored{
			content:    p.Payload["content"],
			score:      score,
			materialID: p.MaterialID,
		})
	}

	// 按相似度降序排序
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	// 取 topK
	limit := topK
	if limit > len(candidates) {
		limit = len(candidates)
	}

	var results []SearchResult
	for i := 0; i < limit; i++ {
		results = append(results, SearchResult{
			Content:    candidates[i].content,
			Score:      candidates[i].score,
			MaterialID: candidates[i].materialID,
		})
	}

	return results, nil
}

// SemanticSearch 一体化语义搜索：输入查询文本 → embedding → 向量检索
// 支持按 userID 过滤（只搜索该用户的材料）
func (vs *VectorStore) SemanticSearch(ctx context.Context, query string, userID string, topK int) ([]SearchResult, error) {
	// 1. 将查询文本向量化
	queryVec, err := vs.embedder.CreateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询 embedding 失败: %w", err)
	}

	// 2. 在用户的所有材料中搜索
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	type scored struct {
		content    string
		score      float32
		materialID string
	}

	var candidates []scored
	for _, p := range vs.points {
		// 按用户过滤
		if userID != "" && p.UserID != userID {
			continue
		}
		score := cosineSimilarity(queryVec, p.Vector)
		candidates = append(candidates, scored{
			content:    p.Payload["content"],
			score:      score,
			materialID: p.MaterialID,
		})
	}

	// 按相似度降序排序
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	limit := topK
	if limit > len(candidates) {
		limit = len(candidates)
	}

	var results []SearchResult
	for i := 0; i < limit; i++ {
		results = append(results, SearchResult{
			Content:    candidates[i].content,
			Score:      candidates[i].score,
			MaterialID: candidates[i].materialID,
		})
	}

	return results, nil
}

// SearchResult 检索结果
type SearchResult struct {
	Content    string  `json:"content"`
	Score      float32 `json:"score"`
	MaterialID string  `json:"material_id"`
}

// Close 关闭连接（内存模式无需操作）
func (vs *VectorStore) Close() {
	// 内存模式无需关闭
}

// DeleteMaterialVectors 删除指定材料的所有向量数据
func (vs *VectorStore) DeleteMaterialVectors(materialID string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	var remaining []vectorPoint
	deleted := 0
	for _, p := range vs.points {
		if p.MaterialID == materialID {
			deleted++
		} else {
			remaining = append(remaining, p)
		}
	}
	vs.points = remaining

	if deleted > 0 {
		log.Printf("已删除材料 %s 的 %d 条向量数据", materialID, deleted)
	}
}

// cosineSimilarity 计算两个向量的余弦相似度
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		ai, bi := float64(a[i]), float64(b[i])
		dotProduct += ai * bi
		normA += ai * ai
		normB += bi * bi
	}

	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}

	return float32(dotProduct / denom)
}
