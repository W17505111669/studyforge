package agent

import (
	"context"
	"encoding/json"
	"fmt"
)

// CardMakerAgent 卡片师 Agent：生成学习卡片
type CardMakerAgent struct {
	llm *LLMClient
}

func NewCardMakerAgent(llm *LLMClient) *CardMakerAgent {
	return &CardMakerAgent{llm: llm}
}

// CardMakerOutput 卡片师输出
type CardMakerOutput struct {
	Cards []CardItem `json:"cards"`
}

type CardItem struct {
	Concept    string `json:"concept"`     // 核心概念（一句话）
	Detail     string `json:"detail"`      // 详细解释
	Formula    string `json:"formula"`     // 关键公式/代码（如有）
	MemoryTip  string `json:"memory_tip"`  // 记忆技巧
	Difficulty string `json:"difficulty"`  // easy, medium, hard
	Tags       string `json:"tags"`        // 逗号分隔标签
}

const cardMakerPrompt = `你是 StudyForge 的「卡片师 Agent」。你的任务是把学习材料中的知识点转化为精美的学习卡片数据。

请严格按照以下 JSON 格式输出，不要输出任何其他内容：
{
  "cards": [
    {
      "concept": "核心概念名称（一句话）",
      "detail": "详细解释（3-5 句话，通俗易懂）",
      "formula": "关键公式或代码（如有，没有则留空字符串）",
      "memory_tip": "记忆技巧或口诀（帮助记住这个知识点）",
      "difficulty": "easy/medium/hard",
      "tags": "标签1,标签2,标签3"
    }
  ]
}

要求：
1. 每张卡片聚焦一个核心概念
2. detail 要用通俗的语言解释，就像给同学讲题
3. memory_tip 要有趣、好记，可以用类比或口诀
4. 生成 5-12 张卡片
5. 使用中文`

// Generate 生成学习卡片
func (c *CardMakerAgent) Generate(ctx context.Context, content string) (*CardMakerOutput, int, int, error) {
	response, inputTokens, outputTokens, err := c.llm.ChatWithUsage(ctx, cardMakerPrompt, content)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("CardMaker Agent 失败: %w", err)
	}

	var output CardMakerOutput
	if err := json.Unmarshal([]byte(extractJSON(response)), &output); err != nil {
		return nil, inputTokens, outputTokens, fmt.Errorf("CardMaker 输出解析失败: %w (原始输出前200字: %.200s)", err, response)
	}

	return &output, inputTokens, outputTokens, nil
}
