package agent

import (
	"context"
	"encoding/json"
	"fmt"
)

// AnalystAgent 分析师 Agent：提取核心知识点和概念关系
type AnalystAgent struct {
	llm *LLMClient
}

func NewAnalystAgent(llm *LLMClient) *AnalystAgent {
	return &AnalystAgent{llm: llm}
}

// AnalystOutput 分析师输出结构
type AnalystOutput struct {
	Summary      string        `json:"summary"`       // 材料总体概述
	KeyPoints    []KeyPoint    `json:"key_points"`    // 核心知识点列表
	Relationships []Relation   `json:"relationships"` // 概念间关系
	Importance   string        `json:"importance"`    // 重要程度评估
}

type KeyPoint struct {
	Concept    string `json:"concept"`    // 概念名称
	Detail     string `json:"detail"`     // 详细说明
	Difficulty string `json:"difficulty"` // easy / medium / hard
}

type Relation struct {
	From string `json:"from"` // 概念 A
	To   string `json:"to"`   // 概念 B
	Type string `json:"type"` // "prerequisite", "related", "analogy"
}

const analystPrompt = `你是 StudyForge 的「分析师 Agent」。你的任务是分析学习材料，提取核心知识点。

请严格按照以下 JSON 格式输出，不要输出任何其他内容：
{
  "summary": "材料的总体概述（2-3 句话）",
  "key_points": [
    {
      "concept": "概念名称",
      "detail": "详细说明（2-3 句话）",
      "difficulty": "easy/medium/hard"
    }
  ],
  "relationships": [
    {
      "from": "概念A",
      "to": "概念B",
      "type": "依赖"
    }
  ],
  "importance": "对这段材料重要程度的评估（1-2 句话）"
}

要求：
1. 提取 5-15 个核心知识点
2. 标注每个知识点的难度等级
3. 识别知识点之间的关系，type 使用中文标签，可选类型包括：
   - 依赖：B 需要 A 作为基础才能理解
   - 前置条件：学习 B 之前必须掌握 A
   - 包含：A 包含 B 作为子概念
   - 关联：两个概念有密切关联但不存在依赖
   - 基于：B 是在 A 的基础上发展而来
   - 对比：两个概念可以相互比较
   - 推导：B 可以从 A 推导出来
   - 实例化：B 是 A 的具体实现或应用
4. 尽量多提取关系（至少 8 条），让知识图谱更丰富
5. 使用中文输出`

// Analyze 分析学习材料
func (a *AnalystAgent) Analyze(ctx context.Context, content string) (*AnalystOutput, int, int, error) {
	response, inputTokens, outputTokens, err := a.llm.ChatWithUsage(ctx, analystPrompt, content)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("Analyst Agent 失败: %w", err)
	}

	var output AnalystOutput
	if err := json.Unmarshal([]byte(extractJSON(response)), &output); err != nil {
		return nil, inputTokens, outputTokens, fmt.Errorf("Analyst 输出解析失败: %w (原始输出前200字: %.200s)", err, response)
	}

	return &output, inputTokens, outputTokens, nil
}
