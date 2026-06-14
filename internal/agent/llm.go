package agent

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// LLMClient 封装 OpenAI 兼容 API 调用
type LLMClient struct {
	client         *openai.Client
	modelName      string
	embeddingModel openai.EmbeddingModel
	ModelName      string // 外部可读的模型名（用于 Trace 记录）
}

// NewLLMClient 创建 LLM 客户端
func NewLLMClient(apiKey, baseURL, modelName, embeddingModelName string) *LLMClient {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	// 设置 HTTP 超时（防止 LLM API 挂起导致 goroutine 永久阻塞）
	config.HTTPClient = &http.Client{
		Timeout: 120 * time.Second,
	}

	// 将字符串模型名映射到 go-openai 的 EmbeddingModel 类型
	var embModel openai.EmbeddingModel
	switch embeddingModelName {
	case "text-embedding-v3":
		embModel = openai.EmbeddingModel(embeddingModelName)
	case "text-embedding-ada-002":
		embModel = openai.AdaEmbeddingV2
	case "text-embedding-3-small":
		embModel = openai.EmbeddingModel("text-embedding-3-small")
	default:
		if embeddingModelName != "" {
			embModel = openai.EmbeddingModel(embeddingModelName)
		} else {
			embModel = openai.AdaEmbeddingV2
		}
	}

	return &LLMClient{
		client:         openai.NewClientWithConfig(config),
		modelName:      modelName,
		embeddingModel: embModel,
		ModelName:      modelName,
	}
}

// isRetryableError 判断错误是否为瞬时故障，可重试
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// HTTP 429 (限流), 500/502/503 (服务端错误), 网络超时
	return strings.Contains(msg, "429") ||
		strings.Contains(msg, "500") ||
		strings.Contains(msg, "502") ||
		strings.Contains(msg, "503") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "connection refused")
}

// retryChat 带重试的对话调用（最多重试 maxRetries 次，指数退避）
func (l *LLMClient) retryChat(ctx context.Context, req openai.ChatCompletionRequest, maxRetries int) (openai.ChatCompletionResponse, error) {
	var resp openai.ChatCompletionResponse
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err = l.client.CreateChatCompletion(ctx, req)
		if err == nil {
			return resp, nil
		}

		if !isRetryableError(err) || attempt == maxRetries {
			return resp, err
		}

		// 指数退避：1s, 2s, 4s...
		delay := time.Duration(1<<uint(attempt)) * time.Second
		log.Printf("LLM 调用失败 (第 %d 次)，%v 后重试: %v", attempt+1, delay, err)

		select {
		case <-time.After(delay):
			// 继续重试
		case <-ctx.Done():
			return resp, ctx.Err()
		}
	}
	return resp, err
}

// Chat 发送对话请求，返回 AI 回复文本（含重试）
func (l *LLMClient) Chat(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	resp, err := l.retryChat(ctx, openai.ChatCompletionRequest{
		Model: l.modelName,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userMessage},
		},
		Temperature: 0.7,
	}, 2) // 最多重试 2 次
	if err != nil {
		return "", fmt.Errorf("LLM 调用失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("LLM 未返回结果")
	}

	return resp.Choices[0].Message.Content, nil
}

// ChatWithUsage 发送对话请求，同时返回 AI 回复文本和 Token 使用量（含重试）
func (l *LLMClient) ChatWithUsage(ctx context.Context, systemPrompt, userMessage string) (content string, inputTokens, outputTokens int, err error) {
	resp, err := l.retryChat(ctx, openai.ChatCompletionRequest{
		Model: l.modelName,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userMessage},
		},
		Temperature: 0.7,
	}, 2)
	if err != nil {
		return "", 0, 0, fmt.Errorf("LLM 调用失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", 0, 0, fmt.Errorf("LLM 未返回结果")
	}

	return resp.Choices[0].Message.Content, resp.Usage.PromptTokens, resp.Usage.CompletionTokens, nil
}

// ChatWithTools 发送带 Function Calling 的对话请求（含重试）
func (l *LLMClient) ChatWithTools(ctx context.Context, messages []openai.ChatCompletionMessage, tools []openai.Tool) (*openai.ChatCompletionResponse, error) {
	resp, err := l.retryChat(ctx, openai.ChatCompletionRequest{
		Model:    l.modelName,
		Messages: messages,
		Tools:    tools,
	}, 2)
	if err != nil {
		return nil, fmt.Errorf("LLM 工具调用失败: %w", err)
	}
	return &resp, nil
}

// CreateEmbedding 生成文本向量（含重试）
func (l *LLMClient) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	var resp openai.EmbeddingResponse
	var err error

	for attempt := 0; attempt <= 2; attempt++ {
		resp, err = l.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Input: []string{text},
			Model: l.embeddingModel,
		})
		if err == nil {
			break
		}
		if !isRetryableError(err) || attempt == 2 {
			break
		}
		delay := time.Duration(1<<uint(attempt)) * time.Second
		log.Printf("Embedding 调用失败 (第 %d 次)，%v 后重试: %v", attempt+1, delay, err)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if err != nil {
		return nil, fmt.Errorf("Embedding 生成失败: %w", err)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("Embedding 未返回结果")
	}
	return resp.Data[0].Embedding, nil
}

// ChatStream 流式对话：返回一个流式读取器，用于 SSE 逐字输出
func (l *LLMClient) ChatStream(ctx context.Context, systemPrompt, userMessage string) (*openai.ChatCompletionStream, error) {
	stream, err := l.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: l.modelName,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userMessage},
		},
		Temperature: 0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM 流式调用失败: %w", err)
	}
	return stream, nil
}

// ChatStreamFromMessages 基于完整消息历史（含工具调用结果）进行流式对话
func (l *LLMClient) ChatStreamFromMessages(ctx context.Context, messages []openai.ChatCompletionMessage) (*openai.ChatCompletionStream, error) {
	stream, err := l.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       l.modelName,
		Messages:    messages,
		Temperature: 0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM 流式调用失败: %w", err)
	}
	return stream, nil
}
