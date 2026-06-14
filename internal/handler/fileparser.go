package handler

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/nguyenthenguyen/docx"
)

// maxExtractSize 提取文本的最大字节数（防止 zip bomb 式攻击）
const maxExtractSize = 50 * 1024 * 1024 // 50MB

// extractTextFromFile 根据文件扩展名提取文本内容
func extractTextFromFile(reader io.Reader, filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".pdf":
		return extractPDF(filename)
	case ".docx":
		return extractDOCX(filename)
	case ".md", ".txt", ".markdown":
		return extractPlainText(reader)
	default:
		return "", fmt.Errorf("不支持的文件格式: %s（支持 .pdf .docx .md .txt）", ext)
	}
}

// extractPDF 从 PDF 文件提取文本
func extractPDF(fp string) (string, error) {
	f, r, err := pdf.Open(fp)
	if err != nil {
		return "", fmt.Errorf("PDF 打开失败: %w", err)
	}
	defer f.Close()

	reader, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("PDF 文本提取失败: %w", err)
	}

	data, err := io.ReadAll(io.LimitReader(reader, maxExtractSize))
	if err != nil {
		return "", fmt.Errorf("PDF 读取失败: %w", err)
	}
	return string(data), nil
}

// extractDOCX 从 DOCX 文件提取文本
func extractDOCX(fp string) (string, error) {
	r, err := docx.ReadDocxFile(fp)
	if err != nil {
		return "", fmt.Errorf("DOCX 读取失败: %w", err)
	}
	defer r.Close()

	doc := r.Editable()
	content := doc.GetContent()
	// 截断过大的内容
	if len(content) > maxExtractSize {
		content = content[:maxExtractSize]
	}
	return content, nil
}

// extractPlainText 从纯文本/Markdown 文件提取内容
func extractPlainText(reader io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maxExtractSize))
	if err != nil {
		return "", fmt.Errorf("文件读取失败: %w", err)
	}
	return string(data), nil
}
