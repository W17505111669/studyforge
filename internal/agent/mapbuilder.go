package agent

import (
	"context"
	"encoding/json"
	"fmt"
)

// MapBuilderAgent 图谱师 Agent：生成知识图谱数据
type MapBuilderAgent struct {
	llm *LLMClient
}

func NewMapBuilderAgent(llm *LLMClient) *MapBuilderAgent {
	return &MapBuilderAgent{llm: llm}
}

// MapBuilderOutput 图谱师输出（ECharts Graph 格式）
type MapBuilderOutput struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type GraphNode struct {
	Name       string `json:"name"`       // 节点名称
	Category   string `json:"category"`   // 分类标签（中文，如"核心概念"、"算法"等）
	SymbolSize int    `json:"symbolSize"` // 节点大小（根据重要程度）
}

type GraphEdge struct {
	Source string `json:"source"` // 起始节点
	Target string `json:"target"` // 目标节点
	Label  string `json:"label,omitempty"` // 关系标签
}

const mapBuilderPrompt = `你是 StudyForge 的「图谱师 Agent」。你的任务是分析学习材料，生成知识图谱的节点和边数据。

【最重要：边的生成规则】
输入材料中包含「概念关系」部分（格式：- A → B: 关系类型），这些是已经由分析师提取好的知识点关系。
你必须：
1. 将「概念关系」中的每一条关系直接作为边输出，保留其关系类型作为 label
2. 可以在此基础上补充额外的关系边，但不能遗漏已有的关系
3. 边的 source 和 target 必须与节点的 name 完全一致

请严格按照以下 JSON 格式输出，不要输出任何其他内容：
{
  "nodes": [
    {
      "name": "知识点名称",
      "category": "核心概念",
      "symbolSize": 40
    }
  ],
  "edges": [
    {
      "source": "知识点A",
      "target": "知识点B",
      "label": "依赖"
    }
  ]
}

要求：
1. 提取 8-20 个知识点作为节点，节点名称必须与输入中的「核心知识点」概念名称一致
2. category 必须是以下中文分类之一（不要用数字）：
   - "核心概念"：最基础最重要的概念
   - "原理"：底层原理、理论基础
   - "算法"：算法、策略、方法论
   - "应用"：实际应用场景、工程实践
   - "模型"：模型架构、系统框架
   - "方法"：技术方法、工具、技巧
   - "公式"：数学公式、定理、推导
   - "功能"：功能特性、能力描述
3. symbolSize 范围 20-60，核心概念大、次要概念小
4. edges 表示知识点间的关系
5. label 必须标注具体的关系类型，例如：依赖、前置条件、包含、关联、基于、对比、推导、实例化
6. 确保图是连通的（没有孤立节点），每个节点至少有一条边
7. 使用中文`

// Generate 生成知识图谱数据
func (m *MapBuilderAgent) Generate(ctx context.Context, content string) (*MapBuilderOutput, int, int, error) {
	response, inputTokens, outputTokens, err := m.llm.ChatWithUsage(ctx, mapBuilderPrompt, content)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("MapBuilder Agent 失败: %w", err)
	}

	var output MapBuilderOutput
	if err := json.Unmarshal([]byte(extractJSON(response)), &output); err != nil {
		return nil, inputTokens, outputTokens, fmt.Errorf("MapBuilder 输出解析失败: %w (原始输出前200字: %.200s)", err, response)
	}

	return &output, inputTokens, outputTokens, nil
}
