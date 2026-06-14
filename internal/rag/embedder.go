package rag

import (
	"context"
	"fmt"
	"log"
)

// Embedder 文档向量化器（使用 EmbeddingProvider 接口）
type Embedder struct {
	provider EmbeddingProvider
}

// NewEmbedder 创建 Embedder
func NewEmbedder(provider EmbeddingProvider) *Embedder {
	return &Embedder{provider: provider}
}

// EmbedChunks 将一组文档切片转为向量
func (e *Embedder) EmbedChunks(ctx context.Context, chunks []Chunk) ([][]float32, error) {
	vectors := make([][]float32, len(chunks))

	for i, chunk := range chunks {
		vec, err := e.provider.CreateEmbedding(ctx, chunk.Content)
		if err != nil {
			log.Printf("Chunk %d embedding 失败: %v", i, err)
			return nil, fmt.Errorf("第 %d 个 chunk embedding 失败: %w", i, err)
		}
		vectors[i] = vec
	}

	return vectors, nil
}
