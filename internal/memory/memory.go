package memory

import (
	"strings"
	"sync"
)

// ConversationMemory 对话记忆管理器
// 三层记忆架构：短期对话 + 长期摘要 + RAG 语义检索
type ConversationMemory struct {
	mu sync.RWMutex

	// 短期记忆：滑动窗口，保留最近 N 轮对话
	buffer     []Message
	bufferSize int // 默认 10

	// 长期记忆：对话压缩摘要
	summary string

	// 对话轮次计数（每次 assistant 回复算一轮），用于触发自动摘要
	roundCount int
}

// Message 对话消息
type Message struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"`
}

// NewConversationMemory 创建对话记忆
func NewConversationMemory(bufferSize int) *ConversationMemory {
	if bufferSize <= 0 {
		bufferSize = 10
	}
	return &ConversationMemory{
		buffer:     make([]Message, 0),
		bufferSize: bufferSize,
	}
}

// AddMessage 添加一条对话消息到短期记忆
func (m *ConversationMemory) AddMessage(role, content string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.buffer = append(m.buffer, Message{
		Role:    role,
		Content: content,
	})

	// 滑动窗口：超过 bufferSize 时，移除最早的消息
	if len(m.buffer) > m.bufferSize*2 { // *2 因为 user + assistant 是一对
		m.buffer = m.buffer[2:] // 移除最早的一轮对话
	}
}

// GetShortTermMemory 获取短期记忆（最近 N 轮对话）
func (m *ConversationMemory) GetShortTermMemory() []Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Message, len(m.buffer))
	copy(result, m.buffer)
	return result
}

// GetLongTermSummary 获取长期记忆（对话摘要）
func (m *ConversationMemory) GetLongTermSummary() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.summary
}

// UpdateSummary 更新长期记忆摘要
// 通常在每轮对话结束后，由 LLM 生成压缩摘要
func (m *ConversationMemory) UpdateSummary(newSummary string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.summary = newSummary
}

// IncrementRound 增加一轮对话计数
// 应在每次 assistant 回复添加到短期记忆后调用
func (m *ConversationMemory) IncrementRound() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roundCount++
}

// NeedsSummary 检查是否已达到触发摘要的轮次阈值（每 10 轮）
func (m *ConversationMemory) NeedsSummary() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.roundCount > 0 && m.roundCount%10 == 0
}

// GetBufferText 获取短期记忆的格式化对话文本（用于 LLM 生成摘要）
func (m *ConversationMemory) GetBufferText() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sb strings.Builder
	for _, msg := range m.buffer {
		roleLabel := "用户"
		if msg.Role == "assistant" {
			roleLabel = "助手"
		}
		sb.WriteString(roleLabel + "：" + msg.Content + "\n")
	}
	return sb.String()
}

// BuildContext 构建完整的上下文（合并三层记忆）
// ragContext: 从 RAG 检索到的相关文档片段
func (m *ConversationMemory) BuildContext(ragContext string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var ctx string

	// 第一层：长期摘要
	if m.summary != "" {
		ctx += "【历史对话摘要】\n" + m.summary + "\n\n"
	}

	// 第二层：RAG 检索结果
	if ragContext != "" {
		ctx += "【相关学习材料】\n" + ragContext + "\n\n"
	}

	// 第三层：短期对话（最近几轮）
	if len(m.buffer) > 0 {
		ctx += "【最近对话记录】\n"
		for _, msg := range m.buffer {
			roleLabel := "用户"
			if msg.Role == "assistant" {
				roleLabel = "助手"
			}
			ctx += roleLabel + "：" + msg.Content + "\n"
		}
	}

	return ctx
}
