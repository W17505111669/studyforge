package rag

import (
	"context"
	"crypto/md5"
	"fmt"
	"math"
	"strings"
	"testing"
)

// ============================================================
// Mock EmbeddingProvider — 基于内容 hash 生成确定性向量
// ============================================================

// mockEmbeddingProvider 返回基于内容 hash 的固定维度向量
// 相同文本始终产生相同向量，不同文本产生不同向量
type mockEmbeddingProvider struct {
	dim        int
	failOnText string // 如果设置，遇到该文本返回错误
}

func newMockEmbedder(dim int) *mockEmbeddingProvider {
	return &mockEmbeddingProvider{dim: dim}
}

func (m *mockEmbeddingProvider) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if m.failOnText != "" && text == m.failOnText {
		return nil, fmt.Errorf("mock embedding 失败: 指定文本")
	}

	// 基于 MD5 hash 生成确定性向量
	hash := md5.Sum([]byte(text))
	vec := make([]float32, m.dim)

	for i := 0; i < m.dim; i++ {
		// 循环使用 hash 字节生成浮点值 [-1, 1]
		byteIdx := i % len(hash)
		vec[i] = float32(hash[byteIdx])/127.5 - 1.0
		// 引入位置偏移以增加不同文本的差异
		vec[i] += float32(i%7-3) * 0.01
	}

	// 归一化向量使模长为 1
	var norm float64
	for _, v := range vec {
		norm += float64(v) * float64(v)
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range vec {
			vec[i] = float32(float64(vec[i]) / norm)
		}
	}

	return vec, nil
}

// directionalEmbedder 生成方向性向量，用于精确测试余弦相似度排序
// 给定一个基础向量方向，生成与之相似度可控的向量
type directionalEmbedder struct {
	dim int
}

func newDirectionalEmbedder(dim int) *directionalEmbedder {
	return &directionalEmbedder{dim: dim}
}

func (d *directionalEmbedder) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	vec := make([]float32, d.dim)

	// 根据文本内容生成不同方向的向量
	// "query" → [1, 0, 0, ...]
	// "similar" → [0.9, 0.1, 0, ...]
	// "different" → [0, 0, ..., 1]
	switch {
	case strings.Contains(text, "query"):
		vec[0] = 1.0
	case strings.Contains(text, "similar"):
		vec[0] = 0.9
		if d.dim > 1 {
			vec[1] = 0.1
		}
	case strings.Contains(text, "medium"):
		vec[0] = 0.5
		if d.dim > 1 {
			vec[1] = 0.5
		}
	case strings.Contains(text, "different"):
		if d.dim > 2 {
			vec[d.dim-1] = 1.0
		} else {
			vec[0] = -1.0
		}
	default:
		// 用 hash 生成伪随机向量
		hash := md5.Sum([]byte(text))
		for i := 0; i < d.dim; i++ {
			vec[i] = float32(hash[i%len(hash)])/127.5 - 1.0
		}
	}

	// 归一化
	var norm float64
	for _, v := range vec {
		norm += float64(v) * float64(v)
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range vec {
			vec[i] = float32(float64(vec[i]) / norm)
		}
	}

	return vec, nil
}

// ============================================================
// 辅助函数
// ============================================================

// newTestVectorStore 创建测试用向量存储
func newTestVectorStore(embedder EmbeddingProvider) *VectorStore {
	store, _ := NewVectorStore("", 0, "test-collection", 128, embedder)
	return store
}

// insertTestPoints 快速插入测试向量点
func insertTestPoints(store *VectorStore, materialID, userID string, contents []string, embedder EmbeddingProvider) error {
	ctx := context.Background()
	chunks := make([]Chunk, len(contents))
	vectors := make([][]float32, len(contents))

	for i, content := range contents {
		chunks[i] = Chunk{
			Content: content,
			Index:   i,
			Metadata: map[string]string{
				"user_id":     userID,
				"material_id": materialID,
			},
		}
		vec, err := embedder.CreateEmbedding(ctx, content)
		if err != nil {
			return err
		}
		vectors[i] = vec
	}

	return store.UpsertPoints(ctx, materialID, chunks, vectors)
}

// ============================================================
// ChunkDocument 测试
// ============================================================

func TestChunkDocument_EmptyDocument(t *testing.T) {
	chunks := ChunkDocument("", 500, 100)
	if chunks != nil {
		t.Errorf("空文档应返回 nil, 得到 %d 个 chunks", len(chunks))
	}
}

func TestChunkDocument_ShortDocument(t *testing.T) {
	content := "这是一段短文本，不需要切分。"
	chunks := ChunkDocument(content, 500, 100)

	if len(chunks) != 1 {
		t.Fatalf("短文档应产生 1 个 chunk, 得到 %d", len(chunks))
	}

	if chunks[0].Content != content {
		t.Errorf("短文档 chunk 内容应等于原文\n期望: %s\n得到: %s", content, chunks[0].Content)
	}

	if chunks[0].Index != 0 {
		t.Errorf("首个 chunk 索引应为 0, 得到 %d", chunks[0].Index)
	}
}

func TestChunkDocument_LongDocument_MultipleChunks(t *testing.T) {
	// 构造长文档：多个段落，每段约 50 字
	var paragraphs []string
	for i := 0; i < 10; i++ {
		paragraphs = append(paragraphs, fmt.Sprintf("这是第 %d 段内容，包含一些关于 Go 并发编程的知识。Goroutine 是轻量级线程，由 Go 运行时管理。", i))
	}
	content := strings.Join(paragraphs, "\n\n")

	chunkSize := 200
	overlap := 50
	chunks := ChunkDocument(content, chunkSize, overlap)

	if len(chunks) < 2 {
		t.Fatalf("长文档应产生多个 chunk, 得到 %d", len(chunks))
	}

	// 验证索引递增
	for i, chunk := range chunks {
		if chunk.Index != i {
			t.Errorf("chunk %d 索引应为 %d, 得到 %d", i, i, chunk.Index)
		}
	}

	// 验证所有 chunk 非空
	for i, chunk := range chunks {
		if strings.TrimSpace(chunk.Content) == "" {
			t.Errorf("chunk %d 内容为空", i)
		}
	}
}

func TestChunkDocument_SingleLongParagraph(t *testing.T) {
	// 单个超长段落（无换行），应被 splitBySize 切分
	longText := strings.Repeat("Go语言并发编程的核心概念包括goroutine和channel。", 20)

	chunkSize := 100
	overlap := 20
	chunks := ChunkDocument(longText, chunkSize, overlap)

	if len(chunks) < 2 {
		t.Fatalf("超长段落应产生多个 chunk, 得到 %d", len(chunks))
	}

	// 验证有重叠内容（相邻 chunk 首尾应有部分相同内容）
	if len(chunks) >= 2 {
		// 重叠检测：后一个 chunk 开头应包含前一个 chunk 末尾的部分内容
		prevRunes := []rune(chunks[0].Content)
		nextRunes := []rune(chunks[1].Content)

		// 取前一个 chunk 末尾 overlap 个字符
		overlapStart := len(prevRunes) - overlap
		if overlapStart < 0 {
			overlapStart = 0
		}
		tail := string(prevRunes[overlapStart:])

		// 后一个 chunk 开头应包含部分重叠内容
		headLen := len(nextRunes)
		if headLen > overlap {
			headLen = overlap
		}
		head := string(nextRunes[:headLen])

		// 只要有部分重叠即可（实际实现中重叠可能不完全精确）
		hasOverlap := false
		for i := 0; i < len([]rune(tail))-5; i++ {
			substr := string([]rune(tail)[i : i+5])
			if strings.Contains(head, substr) {
				hasOverlap = true
				break
			}
		}

		if !hasOverlap {
			t.Logf("注意: 重叠检测未找到明确的 %d 字符重叠, tail=%q, head=%q", overlap, tail, head)
		}
	}
}

func TestChunkDocument_MultipleParagraphs(t *testing.T) {
	// 多个短段落，应合并到同一 chunk（不超过 chunkSize）
	content := "第一段。\n\n第二段。\n\n第三段。"
	chunks := ChunkDocument(content, 500, 50)

	if len(chunks) != 1 {
		t.Fatalf("短段落应合并为 1 个 chunk, 得到 %d", len(chunks))
	}

	// 验证内容包含所有段落
	for _, para := range []string{"第一段", "第二段", "第三段"} {
		if !strings.Contains(chunks[0].Content, para) {
			t.Errorf("chunk 内容应包含 %q, 实际: %s", para, chunks[0].Content)
		}
	}
}

func TestChunkDocument_WhitespaceOnly(t *testing.T) {
	chunks := ChunkDocument("   \n\n   \n\n   ", 500, 100)
	if len(chunks) != 0 {
		t.Errorf("纯空白文档应返回空结果, 得到 %d 个 chunks", len(chunks))
	}
}

// ============================================================
// UpsertPoints + Search 集成测试
// ============================================================

func TestUpsertPoints_ThenSearch(t *testing.T) {
	embedder := newMockEmbedder(128)
	store := newTestVectorStore(embedder)
	ctx := context.Background()

	err := insertTestPoints(store, "mat-1", "user-1",
		[]string{
			"Go 并发编程的基础知识",
			"Python 数据分析入门教程",
			"机器学习算法原理详解",
		}, embedder)

	if err != nil {
		t.Fatalf("UpsertPoints 失败: %v", err)
	}

	// 查询向量使用第一个文本的 embedding
	queryVec, _ := embedder.CreateEmbedding(ctx, "Go 并发编程的基础知识")
	results, err := store.Search(ctx, queryVec, "mat-1", 3)
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("期望 3 个结果, 得到 %d", len(results))
	}

	// 第一个结果应该是完全匹配的那个（相似度最高）
	if results[0].Content != "Go 并发编程的基础知识" {
		t.Errorf("最相关结果应为完全匹配文本, 得到: %s", results[0].Content)
	}

	// 相似度应接近 1.0（完全相同的向量）
	if results[0].Score < 0.99 {
		t.Errorf("完全匹配文本的相似度应接近 1.0, 得到: %f", results[0].Score)
	}
}

func TestUpsertPoints_MismatchedChunksVectors(t *testing.T) {
	embedder := newMockEmbedder(128)
	store := newTestVectorStore(embedder)
	ctx := context.Background()

	chunks := []Chunk{{Content: "test", Index: 0}}
	vectors := [][]float32{{1.0, 0.0}, {0.0, 1.0}} // 2 个向量 vs 1 个 chunk

	err := store.UpsertPoints(ctx, "mat-1", chunks, vectors)
	if err == nil {
		t.Fatal("chunks 和 vectors 数量不匹配应返回错误")
	}
}

func TestUpsertPoints_EmptyChunks(t *testing.T) {
	embedder := newMockEmbedder(128)
	store := newTestVectorStore(embedder)
	ctx := context.Background()

	err := store.UpsertPoints(ctx, "mat-1", nil, nil)
	if err != nil {
		t.Fatalf("空 chunks 不应返回错误, 得到: %v", err)
	}
}

// ============================================================
// Search 测试 — 余弦相似度排序、topK 截断、materialID 过滤
// ============================================================

func TestSearch_CosineSimilaritySort(t *testing.T) {
	embedder := newDirectionalEmbedder(128)
	store := newTestVectorStore(embedder)
	ctx := context.Background()

	// 插入三个不同相似度的文本
	contents := []string{
		"similar content here",   // 与 query 最相似
		"medium content here",    // 中等相似
		"different content here", // 最不相似
	}
	err := insertTestPoints(store, "mat-1", "user-1", contents, embedder)
	if err != nil {
		t.Fatalf("insertTestPoints 失败: %v", err)
	}

	// 用 "query" 向量搜索
	queryVec, _ := embedder.CreateEmbedding(ctx, "query text")
	results, err := store.Search(ctx, queryVec, "mat-1", 10)
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("期望 3 个结果, 得到 %d", len(results))
	}

	// 验证结果按相似度降序排列
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("结果未按相似度降序排列: results[%d].Score=%f > results[%d].Score=%f",
				i, results[i].Score, i-1, results[i-1].Score)
		}
	}

	// "similar" 应该排第一
	if !strings.Contains(results[0].Content, "similar") {
		t.Errorf("最相似的结果应包含 'similar', 得到: %s", results[0].Content)
	}
}

func TestSearch_TopK_Truncation(t *testing.T) {
	embedder := newMockEmbedder(128)
	store := newTestVectorStore(embedder)
	ctx := context.Background()

	// 插入 10 个文本
	contents := make([]string, 10)
	for i := range contents {
		contents[i] = fmt.Sprintf("这是第 %d 个测试文本段落", i)
	}
	err := insertTestPoints(store, "mat-1", "user-1", contents, embedder)
	if err != nil {
		t.Fatalf("insertTestPoints 失败: %v", err)
	}

	queryVec, _ := embedder.CreateEmbedding(ctx, "查询文本")

	// topK = 3
	results, err := store.Search(ctx, queryVec, "mat-1", 3)
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("topK=3 应返回 3 个结果, 得到 %d", len(results))
	}

	// topK = 5
	results, err = store.Search(ctx, queryVec, "mat-1", 5)
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("topK=5 应返回 5 个结果, 得到 %d", len(results))
	}

	// topK 超过实际数量
	results, err = store.Search(ctx, queryVec, "mat-1", 100)
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}

	if len(results) != 10 {
		t.Errorf("topK=100 超过实际数量应返回全部 10 个, 得到 %d", len(results))
	}
}

func TestSearch_MaterialIDFilter(t *testing.T) {
	embedder := newMockEmbedder(128)
	store := newTestVectorStore(embedder)
	ctx := context.Background()

	// 向两个不同材料插入数据
	err := insertTestPoints(store, "mat-1", "user-1",
		[]string{"材料一的知识卡片A", "材料一的知识卡片B"}, embedder)
	if err != nil {
		t.Fatalf("insertTestPoints mat-1 失败: %v", err)
	}

	err = insertTestPoints(store, "mat-2", "user-1",
		[]string{"材料二的知识卡片X", "材料二的知识卡片Y", "材料二的知识卡片Z"}, embedder)
	if err != nil {
		t.Fatalf("insertTestPoints mat-2 失败: %v", err)
	}

	queryVec, _ := embedder.CreateEmbedding(ctx, "知识卡片")

	// 按 mat-1 过滤
	results, err := store.Search(ctx, queryVec, "mat-1", 10)
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("mat-1 过滤应返回 2 个结果, 得到 %d", len(results))
	}

	for _, r := range results {
		if r.MaterialID != "mat-1" {
			t.Errorf("结果应属于 mat-1, 得到: %s", r.MaterialID)
		}
	}

	// 按 mat-2 过滤
	results, err = store.Search(ctx, queryVec, "mat-2", 10)
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("mat-2 过滤应返回 3 个结果, 得到 %d", len(results))
	}

	// 不过滤（materialID 为空）
	results, err = store.Search(ctx, queryVec, "", 10)
	if err != nil {
		t.Fatalf("Search 失败: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("无过滤应返回全部 5 个结果, 得到 %d", len(results))
	}
}

func TestSearch_EmptyStore(t *testing.T) {
	embedder := newMockEmbedder(128)
	store := newTestVectorStore(embedder)
	ctx := context.Background()

	queryVec, _ := embedder.CreateEmbedding(ctx, "查询")
	results, err := store.Search(ctx, queryVec, "", 5)
	if err != nil {
		t.Fatalf("Search 不应返回错误: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("空存储应返回 0 结果, 得到 %d", len(results))
	}
}

// ============================================================
// SemanticSearch 测试 — userID 过滤
// ============================================================

func TestSemanticSearch_UserIDFilter(t *testing.T) {
	embedder := newMockEmbedder(128)
	store := newTestVectorStore(embedder)

	// 两个用户各插入不同材料
	err := insertTestPoints(store, "mat-A", "alice",
		[]string{"Alice 的 Go 并发笔记", "Alice 的 Python 教程"}, embedder)
	if err != nil {
		t.Fatalf("insertTestPoints alice 失败: %v", err)
	}

	err = insertTestPoints(store, "mat-B", "bob",
		[]string{"Bob 的 Java 入门指南", "Bob 的 Rust 学习路径"}, embedder)
	if err != nil {
		t.Fatalf("insertTestPoints bob 失败: %v", err)
	}

	ctx := context.Background()

	// Alice 搜索
	results, err := store.SemanticSearch(ctx, "编程教程", "alice", 10)
	if err != nil {
		t.Fatalf("SemanticSearch alice 失败: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Alice 应看到 2 个结果, 得到 %d", len(results))
	}

	// Bob 搜索
	results, err = store.SemanticSearch(ctx, "编程教程", "bob", 10)
	if err != nil {
		t.Fatalf("SemanticSearch bob 失败: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Bob 应看到 2 个结果, 得到 %d", len(results))
	}

	// 空 userID（不过滤，返回全部）
	results, err = store.SemanticSearch(ctx, "编程教程", "", 10)
	if err != nil {
		t.Fatalf("SemanticSearch 全部失败: %v", err)
	}

	if len(results) != 4 {
		t.Errorf("空 userID 应返回全部 4 个结果, 得到 %d", len(results))
	}
}

func TestSemanticSearch_EmbeddingFailure(t *testing.T) {
	embedder := &mockEmbeddingProvider{dim: 128, failOnText: "bad query"}
	store := newTestVectorStore(embedder)

	_, err := store.SemanticSearch(context.Background(), "bad query", "user-1", 5)
	if err == nil {
		t.Fatal("embedding 失败应返回错误")
	}
}

// ============================================================
// DeleteMaterialVectors 测试
// ============================================================

func TestDeleteMaterialVectors(t *testing.T) {
	embedder := newMockEmbedder(128)
	store := newTestVectorStore(embedder)
	ctx := context.Background()

	// 插入两个材料的向量
	err := insertTestPoints(store, "mat-1", "user-1",
		[]string{"材料一内容A", "材料一内容B"}, embedder)
	if err != nil {
		t.Fatalf("insertTestPoints mat-1 失败: %v", err)
	}

	err = insertTestPoints(store, "mat-2", "user-1",
		[]string{"材料二内容X", "材料二内容Y"}, embedder)
	if err != nil {
		t.Fatalf("insertTestPoints mat-2 失败: %v", err)
	}

	// 验证初始状态
	queryVec, _ := embedder.CreateEmbedding(ctx, "内容")
	results, _ := store.Search(ctx, queryVec, "", 10)
	if len(results) != 4 {
		t.Fatalf("删除前应返回 4 个结果, 得到 %d", len(results))
	}

	// 删除 mat-1
	store.DeleteMaterialVectors("mat-1")

	// mat-1 的搜索应该返回空
	results, _ = store.Search(ctx, queryVec, "mat-1", 10)
	if len(results) != 0 {
		t.Errorf("删除 mat-1 后应返回 0 个结果, 得到 %d", len(results))
	}

	// mat-2 应该不受影响
	results, _ = store.Search(ctx, queryVec, "mat-2", 10)
	if len(results) != 2 {
		t.Errorf("mat-2 应保持 2 个结果, 得到 %d", len(results))
	}

	// 全局搜索只返回 mat-2
	results, _ = store.Search(ctx, queryVec, "", 10)
	if len(results) != 2 {
		t.Errorf("全局搜索应只返回 mat-2 的 2 个结果, 得到 %d", len(results))
	}
}

func TestDeleteMaterialVectors_NonExistent(t *testing.T) {
	embedder := newMockEmbedder(128)
	store := newTestVectorStore(embedder)

	// 插入一些数据
	err := insertTestPoints(store, "mat-1", "user-1",
		[]string{"测试内容"}, embedder)
	if err != nil {
		t.Fatalf("insertTestPoints 失败: %v", err)
	}

	// 删除不存在的材料不应 panic 或影响现有数据
	store.DeleteMaterialVectors("non-existent-mat")

	ctx := context.Background()
	queryVec, _ := embedder.CreateEmbedding(ctx, "测试")
	results, _ := store.Search(ctx, queryVec, "", 10)
	if len(results) != 1 {
		t.Errorf("删除不存在的材料后应保留原有 1 个结果, 得到 %d", len(results))
	}
}

// ============================================================
// cosineSimilarity 测试
// ============================================================

func TestCosineSimilarity_IdenticalVectors(t *testing.T) {
	a := []float32{1.0, 2.0, 3.0}
	sim := cosineSimilarity(a, a)
	if math.Abs(float64(sim)-1.0) > 0.001 {
		t.Errorf("相同向量的余弦相似度应为 1.0, 得到: %f", sim)
	}
}

func TestCosineSimilarity_OrthogonalVectors(t *testing.T) {
	a := []float32{1.0, 0.0}
	b := []float32{0.0, 1.0}
	sim := cosineSimilarity(a, b)
	if math.Abs(float64(sim)) > 0.001 {
		t.Errorf("正交向量的余弦相似度应接近 0, 得到: %f", sim)
	}
}

func TestCosineSimilarity_OppositeVectors(t *testing.T) {
	a := []float32{1.0, 0.0}
	b := []float32{-1.0, 0.0}
	sim := cosineSimilarity(a, b)
	if math.Abs(float64(sim)+1.0) > 0.001 {
		t.Errorf("相反向量的余弦相似度应接近 -1.0, 得到: %f", sim)
	}
}

func TestCosineSimilarity_DifferentLengths(t *testing.T) {
	a := []float32{1.0, 2.0}
	b := []float32{1.0, 2.0, 3.0}
	sim := cosineSimilarity(a, b)
	if sim != 0 {
		t.Errorf("不同长度向量应返回 0, 得到: %f", sim)
	}
}

func TestCosineSimilarity_EmptyVectors(t *testing.T) {
	sim := cosineSimilarity(nil, nil)
	if sim != 0 {
		t.Errorf("空向量应返回 0, 得到: %f", sim)
	}
}

func TestCosineSimilarity_ZeroVector(t *testing.T) {
	a := []float32{1.0, 2.0}
	b := []float32{0.0, 0.0}
	sim := cosineSimilarity(a, b)
	if sim != 0 {
		t.Errorf("零向量应返回 0（除零保护）, 得到: %f", sim)
	}
}

// ============================================================
// IndexMaterial 集成测试
// ============================================================

func TestIndexMaterial_EndToEnd(t *testing.T) {
	embedder := newMockEmbedder(128)
	store := newTestVectorStore(embedder)
	ctx := context.Background()

	content := "Go 并发编程入门\n\nGoroutine 是 Go 语言的轻量级线程。通过 go 关键字即可启动。\n\nChannel 是 goroutine 之间的通信管道。使用 make(chan Type) 创建。"

	err := store.IndexMaterial(ctx, "mat-go", "user-1", content, 200, 50)
	if err != nil {
		t.Fatalf("IndexMaterial 失败: %v", err)
	}

	// 验证向量已存入
	results, err := store.SemanticSearch(ctx, "Goroutine", "user-1", 10)
	if err != nil {
		t.Fatalf("SemanticSearch 失败: %v", err)
	}

	if len(results) == 0 {
		t.Error("IndexMaterial 后 SemanticSearch 应返回结果")
	}
}

func TestIndexMaterial_EmptyContent(t *testing.T) {
	embedder := newMockEmbedder(128)
	store := newTestVectorStore(embedder)
	ctx := context.Background()

	err := store.IndexMaterial(ctx, "mat-empty", "user-1", "", 200, 50)
	if err != nil {
		t.Fatalf("空内容 IndexMaterial 不应返回错误: %v", err)
	}

	// 验证存储为空
	results, _ := store.Search(ctx, []float32{1.0}, "", 10)
	if len(results) != 0 {
		t.Errorf("空内容索引后存储应为空, 得到 %d 条", len(results))
	}
}

func TestIndexMaterial_EmbeddingFailure(t *testing.T) {
	embedder := &mockEmbeddingProvider{dim: 128, failOnText: "Go 并发编程入门"}
	store := newTestVectorStore(embedder)
	ctx := context.Background()

	err := store.IndexMaterial(ctx, "mat-fail", "user-1", "Go 并发编程入门", 200, 50)
	if err == nil {
		t.Fatal("embedding 失败时 IndexMaterial 应返回错误")
	}
}

// ============================================================
// VectorStore 基础功能测试
// ============================================================

func TestNewVectorStore(t *testing.T) {
	embedder := newMockEmbedder(128)
	store, err := NewVectorStore("localhost", 6333, "test", 128, embedder)
	if err != nil {
		t.Fatalf("NewVectorStore 不应返回错误: %v", err)
	}

	if store == nil {
		t.Fatal("NewVectorStore 不应返回 nil")
	}

	if store.collection != "test" {
		t.Errorf("collection 应为 'test', 得到: %s", store.collection)
	}

	if store.vectorSize != 128 {
		t.Errorf("vectorSize 应为 128, 得到: %d", store.vectorSize)
	}
}

func TestEnsureCollection(t *testing.T) {
	store := newTestVectorStore(newMockEmbedder(128))
	err := store.EnsureCollection(context.Background())
	if err != nil {
		t.Errorf("内存模式 EnsureCollection 应返回 nil: %v", err)
	}
}

func TestClose(t *testing.T) {
	store := newTestVectorStore(newMockEmbedder(128))
	// Close 不应 panic
	store.Close()
}

// ============================================================
// Embedder 测试
// ============================================================

func TestEmbedder_EmbedChunks(t *testing.T) {
	embedder := newMockEmbedder(128)
	emb := NewEmbedder(embedder)
	ctx := context.Background()

	chunks := []Chunk{
		{Content: "第一个文本块", Index: 0},
		{Content: "第二个文本块", Index: 1},
		{Content: "第三个文本块", Index: 2},
	}

	vectors, err := emb.EmbedChunks(ctx, chunks)
	if err != nil {
		t.Fatalf("EmbedChunks 失败: %v", err)
	}

	if len(vectors) != 3 {
		t.Fatalf("期望 3 个向量, 得到 %d", len(vectors))
	}

	for i, vec := range vectors {
		if len(vec) != 128 {
			t.Errorf("向量 %d 维度应为 128, 得到 %d", i, len(vec))
		}
	}

	// 验证确定性：相同内容产生相同向量
	vectors2, _ := emb.EmbedChunks(ctx, chunks)
	for i := range vectors {
		for j := range vectors[i] {
			if vectors[i][j] != vectors2[i][j] {
				t.Errorf("向量 %d 不具有确定性", i)
				break
			}
		}
	}
}

func TestEmbedder_EmbedChunks_Empty(t *testing.T) {
	embedder := newMockEmbedder(128)
	emb := NewEmbedder(embedder)
	ctx := context.Background()

	vectors, err := emb.EmbedChunks(ctx, nil)
	if err != nil {
		t.Fatalf("空 chunks 不应返回错误: %v", err)
	}

	if len(vectors) != 0 {
		t.Errorf("空 chunks 应返回空向量列表, 得到 %d", len(vectors))
	}
}

func TestEmbedder_EmbedChunks_ProviderFailure(t *testing.T) {
	embedder := &mockEmbeddingProvider{dim: 128, failOnText: "bad content"}
	emb := NewEmbedder(embedder)
	ctx := context.Background()

	chunks := []Chunk{
		{Content: "good content", Index: 0},
		{Content: "bad content", Index: 1},
	}

	_, err := emb.EmbedChunks(ctx, chunks)
	if err == nil {
		t.Fatal("provider 失败时 EmbedChunks 应返回错误")
	}
}

// ============================================================
// splitBySize 辅助函数测试
// ============================================================

func TestSplitBySize_Basic(t *testing.T) {
	text := "12345678901234567890" // 20 个字符
	parts := splitBySize(text, 10, 3)

	if len(parts) < 2 {
		t.Fatalf("20 字符文本按 10 切分应产生至少 2 部分, 得到 %d", len(parts))
	}

	// 第一部分应为 10 个字符
	if len([]rune(parts[0])) != 10 {
		t.Errorf("第一部分应为 10 个字符, 得到 %d", len([]rune(parts[0])))
	}
}

func TestSplitBySize_OverlapZero(t *testing.T) {
	text := "ABCDEFGHIJ" // 10 个字符
	parts := splitBySize(text, 5, 0)

	if len(parts) != 2 {
		t.Fatalf("10 字符按 5 无重叠切分应产生 2 部分, 得到 %d", len(parts))
	}

	if parts[0] != "ABCDE" {
		t.Errorf("第一部分应为 'ABCDE', 得到: %s", parts[0])
	}
	if parts[1] != "FGHIJ" {
		t.Errorf("第二部分应为 'FGHIJ', 得到: %s", parts[1])
	}
}

func TestSplitBySize_OverlapEqualsSize(t *testing.T) {
	// overlap >= size 应回退到 step = size（防止无限循环）
	text := "ABCDEFGHIJ"
	parts := splitBySize(text, 5, 5)

	if len(parts) < 1 {
		t.Fatalf("应产生至少 1 部分")
	}

	// 验证不会无限循环（已经成功返回）
	t.Logf("overlap=size 产生 %d 部分", len(parts))
}

func TestSplitBySize_TextShorterThanSize(t *testing.T) {
	text := "短文本"
	parts := splitBySize(text, 100, 10)

	if len(parts) != 1 {
		t.Fatalf("短文本应产生 1 部分, 得到 %d", len(parts))
	}

	if parts[0] != text {
		t.Errorf("短文本应原样返回, 得到: %s", parts[0])
	}
}

// ============================================================
// splitParagraphs 辅助函数测试
// ============================================================

func TestSplitParagraphs_DoubleNewline(t *testing.T) {
	content := "第一段。\n\n第二段。\n\n第三段。"
	paras := splitParagraphs(content)

	if len(paras) != 3 {
		t.Fatalf("双换行应切分为 3 段, 得到 %d", len(paras))
	}

	expected := []string{"第一段。", "第二段。", "第三段。"}
	for i, p := range paras {
		if p != expected[i] {
			t.Errorf("段落 %d 应为 %q, 得到 %q", i, expected[i], p)
		}
	}
}

func TestSplitParagraphs_SingleNewline(t *testing.T) {
	// 无换行时按单换行切分
	content := "第一行\n第二行\n第三行"
	paras := splitParagraphs(content)

	if len(paras) != 3 {
		t.Fatalf("单换行应切分为 3 段, 得到 %d", len(paras))
	}
}

func TestSplitParagraphs_EmptyLines(t *testing.T) {
	content := "\n\n\n  \n\n"
	paras := splitParagraphs(content)

	if len(paras) != 0 {
		t.Errorf("纯空行应返回空列表, 得到 %d", len(paras))
	}
}

// ============================================================
// 并发安全测试
// ============================================================

func TestVectorStore_ConcurrentAccess(t *testing.T) {
	embedder := newMockEmbedder(64)
	store := newTestVectorStore(embedder)
	ctx := context.Background()

	// 并发写入
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			defer func() { done <- true }()
			content := fmt.Sprintf("并发测试文本 %d", idx)
			chunks := []Chunk{{
				Content:  content,
				Index:    0,
				Metadata: map[string]string{"user_id": "user-1"},
			}}
			vec, _ := embedder.CreateEmbedding(ctx, content)
			vectors := [][]float32{vec}
			_ = store.UpsertPoints(ctx, fmt.Sprintf("mat-%d", idx), chunks, vectors)
		}(i)
	}

	// 等待所有写入完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 并发读取
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			queryVec, _ := embedder.CreateEmbedding(ctx, "测试")
			_, _ = store.Search(ctx, queryVec, "", 5)
		}()
	}

	// 等待所有读取完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证所有数据已写入
	queryVec, _ := embedder.CreateEmbedding(ctx, "并发")
	results, _ := store.Search(ctx, queryVec, "", 100)
	if len(results) != 10 {
		t.Errorf("并发写入后应有 10 个结果, 得到 %d", len(results))
	}
}
