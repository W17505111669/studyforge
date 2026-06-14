package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 应用全局配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	LLM      LLMConfig      `yaml:"llm"`
	Qdrant   QdrantConfig   `yaml:"qdrant"`
	RAG      RAGConfig      `yaml:"rag"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

type LLMConfig struct {
	APIKey         string `yaml:"api_key"`
	BaseURL        string `yaml:"base_url"`
	Model          string `yaml:"model"`
	EmbeddingModel string `yaml:"embedding_model"`
}

type QdrantConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Collection string `yaml:"collection"`
	VectorSize int    `yaml:"vector_size"`
}

type RAGConfig struct {
	ChunkSize    int `yaml:"chunk_size"`
	ChunkOverlap int `yaml:"chunk_overlap"`
	TopK         int `yaml:"top_k"`
}

// LoadConfig 从 YAML 文件加载配置，环境变量优先级高于 YAML 文件
// 如果 YAML 文件不存在，则完全使用环境变量+默认值
func LoadConfig(path string) (*Config, error) {
	cfg := defaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("读取配置文件 %s 失败: %w", path, err)
		}
		// 文件不存在，使用默认值+环境变量
		log.Printf("配置文件 %s 不存在，使用环境变量和默认值", path)
	} else {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("解析配置文件 %s 失败: %w", path, err)
		}
	}

	// 环境变量覆盖（优先级最高）
	cfg.applyEnvOverrides()

	return cfg, nil
}

// defaultConfig 返回带默认值的配置
func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
		Database: DatabaseConfig{
			Path: "./data/studyforge.db",
		},
		JWT: JWTConfig{
			ExpireHours: 72,
		},
		LLM: LLMConfig{
			BaseURL:        "https://dashscope.aliyuncs.com/compatible-mode/v1",
			Model:          "qwen-plus",
			EmbeddingModel: "text-embedding-v3",
		},
		Qdrant: QdrantConfig{
			Host:       "localhost",
			Port:       6333,
			Collection: "studyforge_docs",
			VectorSize: 1024,
		},
		RAG: RAGConfig{
			ChunkSize:    400,
			ChunkOverlap: 50,
			TopK:         5,
		},
	}
}

// applyEnvOverrides 用环境变量覆盖配置值
func (c *Config) applyEnvOverrides() {
	// ===== 服务器 =====
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.Server.Port = n
		}
	}
	// Railway / Render 等平台通过 PORT 环境变量分配端口
	if v := os.Getenv("PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.Server.Port = n
		}
	}
	if v := os.Getenv("SERVER_HOST"); v != "" {
		c.Server.Host = v
	}

	// ===== 数据库 =====
	if v := os.Getenv("DB_PATH"); v != "" {
		c.Database.Path = v
	}

	// ===== JWT =====
	if v := os.Getenv("JWT_SECRET"); v != "" {
		c.JWT.Secret = v
	}
	if v := os.Getenv("JWT_EXPIRE_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.JWT.ExpireHours = n
		}
	}

	// ===== LLM =====
	if v := os.Getenv("LLM_API_KEY"); v != "" {
		c.LLM.APIKey = v
	}
	if v := os.Getenv("LLM_BASE_URL"); v != "" {
		c.LLM.BaseURL = v
	}
	if v := os.Getenv("LLM_MODEL"); v != "" {
		c.LLM.Model = v
	}
	if v := os.Getenv("LLM_EMBEDDING_MODEL"); v != "" {
		c.LLM.EmbeddingModel = v
	}

	// ===== Qdrant =====
	if v := os.Getenv("QDRANT_HOST"); v != "" {
		c.Qdrant.Host = v
	}
	if v := os.Getenv("QDRANT_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.Qdrant.Port = n
		}
	}
	if v := os.Getenv("QDRANT_COLLECTION"); v != "" {
		c.Qdrant.Collection = v
	}
	if v := os.Getenv("QDRANT_VECTOR_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.Qdrant.VectorSize = n
		}
	}

	// ===== RAG =====
	if v := os.Getenv("RAG_CHUNK_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.RAG.ChunkSize = n
		}
	}
	if v := os.Getenv("RAG_CHUNK_OVERLAP"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			c.RAG.ChunkOverlap = n
		}
	}
	if v := os.Getenv("RAG_TOP_K"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.RAG.TopK = n
		}
	}
}

// Validate 校验必需的配置项，返回友好错误信息
func (c *Config) Validate() error {
	var errors []string

	// ===== 必需字段 =====
	if c.LLM.APIKey == "" {
		errors = append(errors, "  • LLM API Key 未配置 → 设置环境变量 LLM_API_KEY 或在 config.yaml 中填入 llm.api_key")
	}
	if c.JWT.Secret == "" {
		errors = append(errors, "  • JWT Secret 未配置 → 设置环境变量 JWT_SECRET 或在 config.yaml 中填入 jwt.secret")
	}

	// ===== 基本合理性检查 =====
	if c.Database.Path == "" {
		errors = append(errors, "  • 数据库路径为空 → 设置环境变量 DB_PATH 或在 config.yaml 中填入 database.path")
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		errors = append(errors, fmt.Sprintf("  • 服务端口 %d 无效 → 设置 SERVER_PORT 为 1-65535 之间的值", c.Server.Port))
	}
	if c.JWT.ExpireHours <= 0 {
		c.JWT.ExpireHours = 72
	}
	if c.LLM.BaseURL == "" {
		errors = append(errors, "  • LLM Base URL 为空 → 设置环境变量 LLM_BASE_URL 或在 config.yaml 中填入 llm.base_url")
	}
	if c.Qdrant.Host == "" {
		errors = append(errors, "  • Qdrant 主机为空 → 设置环境变量 QDRANT_HOST 或在 config.yaml 中填入 qdrant.host")
	}
	if c.Qdrant.VectorSize <= 0 {
		c.Qdrant.VectorSize = 1024
	}
	if c.RAG.ChunkSize <= 0 {
		c.RAG.ChunkSize = 400
	}
	if c.RAG.ChunkOverlap < 0 {
		c.RAG.ChunkOverlap = 50
	}
	if c.RAG.TopK <= 0 {
		c.RAG.TopK = 5
	}
	if c.RAG.ChunkOverlap >= c.RAG.ChunkSize {
		errors = append(errors, fmt.Sprintf("  • RAG chunk_overlap(%d) 必须小于 chunk_size(%d)", c.RAG.ChunkOverlap, c.RAG.ChunkSize))
	}

	if len(errors) > 0 {
		return fmt.Errorf("配置校验失败（%d 项）:\n%s\n\n提示: 可通过 config.yaml 文件或环境变量设置，详见 .env.example",
			len(errors), strings.Join(errors, "\n"))
	}
	return nil
}

// PrintSummary 打印配置摘要，敏感值已遮蔽
func (c *Config) PrintSummary() {
	log.Println("═══════════════════════════════════════════")
	log.Println("  StudyForge Pro — 配置摘要")
	log.Println("═══════════════════════════════════════════")

	// 服务器
	log.Printf("  [服务器]  %s:%d", c.Server.Host, c.Server.Port)

	// 数据库
	log.Printf("  [数据库]  %s", c.Database.Path)

	// JWT
	log.Printf("  [JWT]     secret=%s, expire=%dh", maskSecret(c.JWT.Secret), c.JWT.ExpireHours)

	// LLM
	log.Printf("  [LLM]     api_key=%s, model=%s, embedding=%s",
		maskSecret(c.LLM.APIKey), c.LLM.Model, c.LLM.EmbeddingModel)
	log.Printf("  [LLM]     base_url=%s", c.LLM.BaseURL)

	// Qdrant
	log.Printf("  [Qdrant]  %s:%d, collection=%s, vector_size=%d",
		c.Qdrant.Host, c.Qdrant.Port, c.Qdrant.Collection, c.Qdrant.VectorSize)

	// RAG
	log.Printf("  [RAG]     chunk_size=%d, overlap=%d, top_k=%d",
		c.RAG.ChunkSize, c.RAG.ChunkOverlap, c.RAG.TopK)

	// 环境变量提示
	envVars := []string{}
	if os.Getenv("LLM_API_KEY") != "" {
		envVars = append(envVars, "LLM_API_KEY")
	}
	if os.Getenv("JWT_SECRET") != "" {
		envVars = append(envVars, "JWT_SECRET")
	}
	if os.Getenv("CORS_ORIGINS") != "" {
		envVars = append(envVars, "CORS_ORIGINS")
	}
	if len(envVars) > 0 {
		log.Printf("  [环境]    已设置环境变量: %s", strings.Join(envVars, ", "))
	}

	log.Println("═══════════════════════════════════════════")
}

// maskSecret 遮蔽敏感值，仅显示前4位
func maskSecret(s string) string {
	if len(s) <= 4 {
		return "***"
	}
	return s[:4] + "***"
}

// GetConfigPath 获取配置文件路径，支持 CONFIG_PATH 环境变量
func GetConfigPath() string {
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		return p
	}
	return "config.yaml"
}
