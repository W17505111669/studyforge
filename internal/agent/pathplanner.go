package agent

import (
	"context"
	"encoding/json"
	"fmt"
)

// PathPlannerAgent 学习路径规划 Agent：根据用户的材料列表生成学习路线图
type PathPlannerAgent struct {
	llm *LLMClient
}

func NewPathPlannerAgent(llm *LLMClient) *PathPlannerAgent {
	return &PathPlannerAgent{llm: llm}
}

// PathStep 学习路径中的一个步骤
type PathStep struct {
	Title            string   `json:"title"`              // 步骤标题
	Description      string   `json:"description"`        // 步骤说明
	EstimatedMinutes int      `json:"estimated_minutes"`  // 预估时长（分钟）
	MaterialIDs      []string `json:"material_ids"`       // 关联材料 ID
	Prerequisites    []int    `json:"prerequisites"`      // 前置依赖步骤索引（0-based）
	Difficulty       string   `json:"difficulty"`         // easy / medium / hard
}

// PathPlannerOutput 路径规划 Agent 输出结构
type PathPlannerOutput struct {
	Overview    string     `json:"overview"`     // 整体学习路线概述
	TotalHours  float64    `json:"total_hours"`  // 预估总学时
	Steps       []PathStep `json:"ordered_steps"` // 有序学习步骤
}

const pathPlannerPrompt = `你是 StudyForge 的「学习路径规划师 Agent」。你的任务是根据用户提供的所有学习材料及其知识点摘要，生成一份科学合理的学习路径规划。

请严格按照以下 JSON 格式输出，不要输出任何其他内容：
{
  "overview": "整体学习路线的简要概述（2-3 句话）",
  "total_hours": 5.5,
  "ordered_steps": [
    {
      "title": "步骤标题",
      "description": "详细描述这一步要学什么、为什么这样安排（2-3 句话）",
      "estimated_minutes": 45,
      "material_ids": ["材料ID1", "材料ID2"],
      "prerequisites": [],
      "difficulty": "easy/medium/hard"
    }
  ]
}

规划原则：
1. 遵循由浅入深、循序渐进的学习规律
2. 前置依赖关系必须合理（prerequisites 引用步骤索引，0-based）
3. 每个步骤关联 1-3 个最相关的材料 ID
4. 预估时长根据内容难度和数量合理分配（每步 15-120 分钟）
5. 步骤数量控制在 4-12 步
6. 如果材料之间存在明确的 prerequisite 关系，优先安排基础材料
7. 使用中文输出`

// Generate 根据材料摘要生成学习路径
func (p *PathPlannerAgent) Generate(ctx context.Context, materialSummary string) (*PathPlannerOutput, int, int, error) {
	response, inputTokens, outputTokens, err := p.llm.ChatWithUsage(ctx, pathPlannerPrompt, materialSummary)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("PathPlanner Agent 失败: %w", err)
	}

	var output PathPlannerOutput
	cleaned := extractJSON(response)
	if err := json.Unmarshal([]byte(cleaned), &output); err != nil {
		return nil, inputTokens, outputTokens, fmt.Errorf("PathPlanner 输出解析失败: %w (原始输出前200字: %.200s)", err, response)
	}

	return &output, inputTokens, outputTokens, nil
}
