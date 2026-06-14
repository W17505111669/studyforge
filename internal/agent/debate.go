package agent

import (
	"context"
	"fmt"
	"time"
)

// DebateMessage 辩论中的一条发言
type DebateMessage struct {
	AgentName string `json:"agent_name"` // Analyst / QuizMaster / CardMaker / Summary
	Role      string `json:"role"`       // 角色描述
	Avatar    string `json:"avatar"`     // 前端头像标识
	Color     string `json:"color"`      // 前端颜色标识
	Round     int    `json:"round"`      // 第几轮（1-based）
	Content   string `json:"content"`    // 发言内容（Markdown 格式）
	Duration  int64  `json:"duration_ms"` // 耗时毫秒
}

// DebateOutput 辩论完整输出
type DebateOutput struct {
	Concept  string          `json:"concept"`  // 辩论主题
	Summary  string          `json:"summary"`  // 最终总结
	Messages []DebateMessage `json:"messages"` // 按顺序的发言列表
}

// DebateOrchestrator 多 Agent 辩论编排器
// 让 Analyst / QuizMaster / CardMaker 三个 Agent 从不同视角讨论同一个知识点
type DebateOrchestrator struct {
	llm *LLMClient
}

// NewDebateOrchestrator 创建辩论编排器
func NewDebateOrchestrator(llm *LLMClient) *DebateOrchestrator {
	return &DebateOrchestrator{llm: llm}
}

// 各 Agent 辩论角色系统提示
const analystDebateSystem = `你是 StudyForge 多 Agent 辩论中的「分析师」。你的专长是深入分析知识点的本质结构和核心原理。
请从以下角度发表观点（200-400 字）：核心本质、关键组成部分、学科重要性、常见理解误区。
用 Markdown 格式输出，通俗易懂，像教授在课堂上讲解。使用中文。不要输出 JSON。`

const quizMasterDebateSystem = `你是 StudyForge 多 Agent 辩论中的「出题官」。你的专长是从考试和评估的角度分析知识点，发现易错点和考察重点。
请从以下角度发表观点（200-400 字）：经典考法与题型、学生常见错误及原因、如何判断掌握深度、检验理解的思考题。
结合前面发言者的观点评论或补充，指出遗漏或偏差。用 Markdown 格式输出。使用中文。不要输出 JSON。`

const cardMakerDebateSystem = `你是 StudyForge 多 Agent 辩论中的「记忆大师」。你的专长是设计记忆策略和学习方法，帮助学生高效记住知识点。
请从以下角度发表观点（200-400 字）：记忆技巧（口诀/类比/联想）、一句话核心概括、生活经验关联、长期复习策略。
结合前面发言者的观点提炼实用记忆方法，给出具体可操作的建议。用 Markdown 格式输出。使用中文。不要输出 JSON。`

const debateSummarySystem = `你是 StudyForge 的辩论总结员。请根据三位专家的讨论写一份精炼总结（150-300 字）。
总结应包含：核心共识、各自独特视角、对学习者实用建议。
用 Markdown 格式输出，结构清晰，重点突出。使用中文。不要输出 JSON。`

// RunDebate 执行多 Agent 辩论
// 三轮发言顺序：Analyst → QuizMaster → CardMaker，每轮包含前面轮次的上下文
// 最后生成一份综合总结
func (d *DebateOrchestrator) RunDebate(ctx context.Context, concept string) (*DebateOutput, int, int, error) {
	output := &DebateOutput{
		Concept:  concept,
		Messages: make([]DebateMessage, 0, 4), // 3 rounds + summary
	}

	var totalInput, totalOutput int
	var debateContext string

	// ========== Round 1: Analyst 分析核心概念 ==========
	start := time.Now()
	userMsg1 := fmt.Sprintf("辩论主题：%s\n\n这是第一轮发言，请从分析师的角度深入分析这个概念。", concept)
	analystContent, in, out, err := d.llm.ChatWithUsage(ctx, analystDebateSystem, userMsg1)
	if err != nil {
		return nil, totalInput, totalOutput, fmt.Errorf("Analyst 辩论失败: %w", err)
	}
	totalInput += in
	totalOutput += out

	output.Messages = append(output.Messages, DebateMessage{
		AgentName: "Analyst",
		Role:      "分析师",
		Avatar:    "analyst",
		Color:     "blue",
		Round:     1,
		Content:   analystContent,
		Duration:  time.Since(start).Milliseconds(),
	})
	debateContext = fmt.Sprintf("【第一轮 · 分析师观点】\n%s", analystContent)

	// ========== Round 2: QuizMaster 从出题者角度分析 ==========
	start = time.Now()
	userMsg2 := fmt.Sprintf("辩论主题：%s\n\n以下是前面分析师的发言：\n%s\n\n请从出题官的角度发表你的观点。", concept, debateContext)
	quizContent, in, out, err := d.llm.ChatWithUsage(ctx, quizMasterDebateSystem, userMsg2)
	if err != nil {
		return nil, totalInput, totalOutput, fmt.Errorf("QuizMaster 辩论失败: %w", err)
	}
	totalInput += in
	totalOutput += out

	output.Messages = append(output.Messages, DebateMessage{
		AgentName: "QuizMaster",
		Role:      "出题官",
		Avatar:    "quizmaster",
		Color:     "amber",
		Round:     2,
		Content:   quizContent,
		Duration:  time.Since(start).Milliseconds(),
	})
	debateContext = fmt.Sprintf("%s\n\n【第二轮 · 出题官观点】\n%s", debateContext, quizContent)

	// ========== Round 3: CardMaker 总结记忆策略 ==========
	start = time.Now()
	userMsg3 := fmt.Sprintf("辩论主题：%s\n\n以下是前面两位专家的发言：\n%s\n\n请从记忆大师的角度发表你的观点。", concept, debateContext)
	cardContent, in, out, err := d.llm.ChatWithUsage(ctx, cardMakerDebateSystem, userMsg3)
	if err != nil {
		return nil, totalInput, totalOutput, fmt.Errorf("CardMaker 辩论失败: %w", err)
	}
	totalInput += in
	totalOutput += out

	output.Messages = append(output.Messages, DebateMessage{
		AgentName: "CardMaker",
		Role:      "记忆大师",
		Avatar:    "cardmaker",
		Color:     "emerald",
		Round:     3,
		Content:   cardContent,
		Duration:  time.Since(start).Milliseconds(),
	})

	// ========== 综合总结 ==========
	allDebate := fmt.Sprintf("【分析师观点】\n%s\n\n【出题官观点】\n%s\n\n【记忆大师观点】\n%s",
		analystContent, quizContent, cardContent)

	start = time.Now()
	summaryUserMsg := fmt.Sprintf("辩论主题：%s\n\n以下是三位专家的完整讨论：\n%s\n\n请撰写总结。", concept, allDebate)
	summary, in, out, err := d.llm.ChatWithUsage(ctx, debateSummarySystem, summaryUserMsg)
	if err != nil {
		// 总结失败不阻塞，使用简易摘要
		summary = fmt.Sprintf("本次辩论围绕「%s」展开，三位专家分别从概念本质、考试应用和记忆策略三个角度进行了深入讨论。", concept)
	} else {
		totalInput += in
		totalOutput += out
	}

	output.Messages = append(output.Messages, DebateMessage{
		AgentName: "Summary",
		Role:      "总结员",
		Avatar:    "summary",
		Color:     "purple",
		Round:     4,
		Content:   summary,
		Duration:  time.Since(start).Milliseconds(),
	})
	output.Summary = summary

	return output, totalInput, totalOutput, nil
}
