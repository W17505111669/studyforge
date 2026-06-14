package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// ChatFunc LLM 对话函数类型（避免循环依赖 agent 包）
type ChatFunc func(ctx context.Context, systemPrompt, userMessage string) (string, error)

// Judge LLM-as-Judge 自动评估器
// 用一个独立的 LLM 调用来评估 Agent 输出的质量
type Judge struct {
	chatFn ChatFunc
}

// NewJudge 创建评估器，接受一个 LLM 对话函数
func NewJudge(chatFn ChatFunc) *Judge {
	return &Judge{chatFn: chatFn}
}

// EvalResult 评估结果
type EvalResult struct {
	Relevance    float64 `json:"relevance"`    // 相关性 0-10
	Accuracy     float64 `json:"accuracy"`     // 准确性 0-10
	Completeness float64 `json:"completeness"` // 完整性 0-10
	Overall      float64 `json:"overall"`      // 综合分
	Feedback     string  `json:"feedback"`     // 评估反馈
}

const judgePrompt = `你是一个 AI 输出质量评估专家。请从以下三个维度评估 AI 的输出质量，每个维度打分 0-10：

1. 相关性 (relevance)：AI 的输出是否回答了用户的问题？是否紧扣主题？
2. 准确性 (accuracy)：AI 输出中的事实陈述是否正确？有没有明显的错误？
3. 完整性 (completeness)：AI 的输出是否涵盖了问题的关键方面？有没有遗漏重要信息？

请严格按照以下 JSON 格式输出，不要输出任何其他内容：
{
  "relevance": 8.5,
  "accuracy": 9.0,
  "completeness": 7.5,
  "overall": 8.3,
  "feedback": "简要评价（1-2 句话）"
}`

// Evaluate 评估一段 Agent 输出
func (j *Judge) Evaluate(ctx context.Context, userQuery, agentOutput string) (*EvalResult, error) {
	prompt := fmt.Sprintf("用户问题：%s\n\nAI 输出：%s", userQuery, agentOutput)

	response, err := j.chatFn(ctx, judgePrompt, prompt)
	if err != nil {
		log.Printf("Judge 评估失败: %v", err)
		return nil, err
	}

	var result EvalResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("Judge 输出解析失败: %w", err)
	}

	return &result, nil
}
