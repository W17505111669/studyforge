package rag

import (
	"strings"
	"unicode/utf8"
)

// Chunk 文档切片
type Chunk struct {
	ID       string `json:"id"`
	Content  string `json:"content"`
	Index    int    `json:"index"`      // 在原文中的位置序号
	Metadata map[string]string `json:"metadata"` // 元数据（材料 ID、标题等）
}

// ChunkDocument 将文档按指定大小切分
// chunkSize: 每个 chunk 的目标字符数
// overlap: 相邻 chunk 的重叠字符数
func ChunkDocument(content string, chunkSize, overlap int) []Chunk {
	if content == "" {
		return nil
	}

	// 先按段落切分
	paragraphs := splitParagraphs(content)

	var chunks []Chunk
	var currentBuilder strings.Builder
	chunkIndex := 0

	for _, para := range paragraphs {
		// 如果单个段落就超过 chunkSize，直接切分
		if utf8.RuneCountInString(para) > chunkSize {
			// 先把当前累积的内容存为一个 chunk
			if currentBuilder.Len() > 0 {
				chunks = append(chunks, Chunk{
					Content: strings.TrimSpace(currentBuilder.String()),
					Index:   chunkIndex,
				})
				chunkIndex++
				currentBuilder.Reset()
			}
			// 把长段落按 chunkSize 切分
			subChunks := splitBySize(para, chunkSize, overlap)
			for _, sc := range subChunks {
				chunks = append(chunks, Chunk{
					Content: sc,
					Index:   chunkIndex,
				})
				chunkIndex++
			}
			continue
		}

		// 检查加入当前段落后是否超过 chunkSize
		newLen := utf8.RuneCountInString(currentBuilder.String()) + utf8.RuneCountInString(para) + 2 // +2 for "\n\n"
		if newLen > chunkSize && currentBuilder.Len() > 0 {
			// 当前 chunk 已满，保存并开始新的
			chunks = append(chunks, Chunk{
				Content: strings.TrimSpace(currentBuilder.String()),
				Index:   chunkIndex,
			})
			chunkIndex++

			// 保留 overlap：取当前 chunk 末尾的 overlap 字符
			currentStr := currentBuilder.String()
			if utf8.RuneCountInString(currentStr) > overlap {
				runes := []rune(currentStr)
				currentBuilder.Reset()
				currentBuilder.WriteString(string(runes[len(runes)-overlap:]))
			} else {
				currentBuilder.Reset()
			}
		}

		if currentBuilder.Len() > 0 {
			currentBuilder.WriteString("\n\n")
		}
		currentBuilder.WriteString(para)
	}

	// 保存最后一个 chunk
	if currentBuilder.Len() > 0 {
		chunks = append(chunks, Chunk{
			Content: strings.TrimSpace(currentBuilder.String()),
			Index:   chunkIndex,
		})
	}

	return chunks
}

// splitParagraphs 按双换行或单换行切分段落
func splitParagraphs(content string) []string {
	// 先尝试按双换行切分
	paragraphs := strings.Split(content, "\n\n")

	var result []string
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}

	// 如果只有一个段落，按单换行再切
	if len(result) <= 1 {
		lines := strings.Split(content, "\n")
		result = nil
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l != "" {
				result = append(result, l)
			}
		}
	}

	return result
}

// splitBySize 将长文本按指定大小切分（带 overlap）
func splitBySize(text string, size, overlap int) []string {
	runes := []rune(text)
	var result []string

	// 防止 overlap >= size 导致无限循环
	step := size - overlap
	if step <= 0 {
		step = size
	}

	for i := 0; i < len(runes); i += step {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		result = append(result, string(runes[i:end]))
		if end == len(runes) {
			break
		}
	}

	return result
}
