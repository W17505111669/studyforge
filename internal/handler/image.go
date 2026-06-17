package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 允许上传的图片 MIME 类型
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/svg+xml": true,
}

// 允许的图片扩展名
var allowedImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".svg":  true,
}

const maxImageSize = 5 * 1024 * 1024 // 5MB

// UploadImage 上传图片到 ./uploads/images/ 目录
// POST /api/images/upload  (multipart/form-data, field: "file")
func (h *Handler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的图片"})
		return
	}
	defer file.Close()

	// 验证扩展名（防绕过 Content-Type）
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedImageExtensions[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 JPG、PNG、WebP、SVG 格式图片"})
		return
	}

	// 验证文件大小
	if header.Size > maxImageSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("图片大小不能超过 5MB，当前 %.1fMB", float64(header.Size)/1024/1024)})
		return
	}

	// 验证 MIME 类型
	contentType := header.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		// SVG 有时 Content-Type 为 text/xml 或 application/xml
		if ext == ".svg" && (contentType == "text/xml" || contentType == "application/xml" || contentType == "") {
			// 允许 SVG 通过
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的图片格式"})
			return
		}
	}

	// 确保上传目录存在
	uploadDir := filepath.Join(".", "uploads", "images")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
		return
	}

	// 生成 UUID 文件名 + 原始扩展名
	filename := uuid.New().String() + ext
	safePath := filepath.Join(uploadDir, filename)

	// 双重检查：确保路径未逃逸出 uploadDir（防路径穿越）
	absUpload, _ := filepath.Abs(uploadDir)
	absFile, _ := filepath.Abs(safePath)
	if !strings.HasPrefix(absFile, absUpload) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件路径"})
		return
	}

	// 写入文件
	dst, err := os.Create(safePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(safePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "写入文件失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":      "/api/images/" + filename,
		"filename": filename,
		"size":     written,
	})
}

// ServeImage 提供图片文件的静态访问
// GET /api/images/:filename
func (h *Handler) ServeImage(c *gin.Context) {
	filename := c.Param("filename")

	// 防路径穿越：只允许简单文件名（无路径分隔符、无 ..）
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") || strings.Contains(filename, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件名"})
		return
	}

	filePath := filepath.Join(".", "uploads", "images", filename)

	// 验证路径在 uploads/images 目录内
	absUpload, _ := filepath.Abs(filepath.Join(".", "uploads", "images"))
	absFile, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absFile, absUpload) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件路径"})
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "图片不存在"})
		return
	}

	// 缓存 30 天（UUID 文件名不可猜测，缓存安全）
	c.Header("Cache-Control", "public, max-age=2592000")
	c.File(filePath)
}
