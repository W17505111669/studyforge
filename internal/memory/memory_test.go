package memory

import (
	"strings"
	"sync"
	"testing"
)

// ==================== NewConversationMemory ====================

func TestNewConversationMemory_DefaultSize(t *testing.T) {
	mem := NewConversationMemory(0)
	if mem.bufferSize != 10 {
		t.Errorf("bufferSize = %d, want 10 (default)", mem.bufferSize)
	}
}

func TestNewConversationMemory_NegativeSize(t *testing.T) {
	mem := NewConversationMemory(-5)
	if mem.bufferSize != 10 {
		t.Errorf("bufferSize = %d, want 10 (default for negative)", mem.bufferSize)
	}
}

func TestNewConversationMemory_CustomSize(t *testing.T) {
	mem := NewConversationMemory(20)
	if mem.bufferSize != 20 {
		t.Errorf("bufferSize = %d, want 20", mem.bufferSize)
	}
}

func TestNewConversationMemory_InitialState(t *testing.T) {
	mem := NewConversationMemory(10)
	if len(mem.buffer) != 0 {
		t.Errorf("initial buffer length = %d, want 0", len(mem.buffer))
	}
	if mem.summary != "" {
		t.Errorf("initial summary = %q, want empty", mem.summary)
	}
	if mem.roundCount != 0 {
		t.Errorf("initial roundCount = %d, want 0", mem.roundCount)
	}
}

// ==================== AddMessage ====================

func TestAddMessage_SingleMessage(t *testing.T) {
	mem := NewConversationMemory(10)
	mem.AddMessage("user", "你好")

	msgs := mem.GetShortTermMemory()
	if len(msgs) != 1 {
		t.Fatalf("GetShortTermMemory length = %d, want 1", len(msgs))
	}
	if msgs[0].Role != "user" || msgs[0].Content != "你好" {
		t.Errorf("message = {%q, %q}, want {user, 你好}", msgs[0].Role, msgs[0].Content)
	}
}

func TestAddMessage_MultipleMessages(t *testing.T) {
	mem := NewConversationMemory(10)
	mem.AddMessage("user", "什么是 Go 协程？")
	mem.AddMessage("assistant", "Go 协程是轻量级线程...")
	mem.AddMessage("user", "怎么创建协程？")
	mem.AddMessage("assistant", "使用 go 关键字...")

	msgs := mem.GetShortTermMemory()
	if len(msgs) != 4 {
		t.Fatalf("GetShortTermMemory length = %d, want 4", len(msgs))
	}

	// 验证顺序正确
	expected := []struct{ role, content string }{
		{"user", "什么是 Go 协程？"},
		{"assistant", "Go 协程是轻量级线程..."},
		{"user", "怎么创建协程？"},
		{"assistant", "使用 go 关键字..."},
	}
	for i, exp := range expected {
		if msgs[i].Role != exp.role || msgs[i].Content != exp.content {
			t.Errorf("msgs[%d] = {%q, %q}, want {%q, %q}", i, msgs[i].Role, msgs[i].Content, exp.role, exp.content)
		}
	}
}

func TestAddMessage_GetBufferTextContainsNewMessage(t *testing.T) {
	mem := NewConversationMemory(10)
	mem.AddMessage("user", "什么是 RAG？")
	mem.AddMessage("assistant", "RAG 是检索增强生成...")

	text := mem.GetBufferText()
	if !strings.Contains(text, "用户：什么是 RAG？") {
		t.Errorf("GetBufferText missing user message, got:\n%s", text)
	}
	if !strings.Contains(text, "助手：RAG 是检索增强生成...") {
		t.Errorf("GetBufferText missing assistant message, got:\n%s", text)
	}
}

// ==================== 滑动窗口 ====================

func TestAddMessage_SlidingWindow(t *testing.T) {
	// bufferSize=3 → 滑动窗口阈值 = 3*2 = 6 条消息
	mem := NewConversationMemory(3)

	// 添加 4 轮对话（8 条消息）
	for i := 0; i < 4; i++ {
		mem.AddMessage("user", "问题"+string(rune('A'+i)))
		mem.AddMessage("assistant", "回答"+string(rune('A'+i)))
	}

	msgs := mem.GetShortTermMemory()
	// 超过 6 条后，最早的一轮（2条）被移除
	// 第1轮添加后: 2条 ≤ 6, 不触发
	// 第2轮添加后: 4条 ≤ 6, 不触发
	// 第3轮添加后: 6条 ≤ 6, 不触发
	// 第4轮添加后: 8条 > 6, 触发移除最早2条 → 剩6条
	if len(msgs) != 6 {
		t.Fatalf("sliding window: length = %d, want 6", len(msgs))
	}

	// 最早的消息应该是问题B（问题A和回答A被移除）
	if msgs[0].Content != "问题B" {
		t.Errorf("first message after sliding = %q, want 问题B", msgs[0].Content)
	}
}

func TestAddMessage_SlidingWindowPreservesOrder(t *testing.T) {
	mem := NewConversationMemory(2) // 阈值 = 4 条

	mem.AddMessage("user", "Q1")
	mem.AddMessage("assistant", "A1")
	mem.AddMessage("user", "Q2")
	mem.AddMessage("assistant", "A2")
	// 此时 4 条，不触发（4 > 4 为 false）

	msgs := mem.GetShortTermMemory()
	if len(msgs) != 4 {
		t.Fatalf("before overflow: length = %d, want 4", len(msgs))
	}

	// 第 5 条触发滑动窗口
	mem.AddMessage("user", "Q3")
	// buffer: [Q1, A1, Q2, A2, Q3] → 5 > 4 → buffer[2:] = [Q2, A2, Q3]
	msgs = mem.GetShortTermMemory()
	if len(msgs) != 3 {
		t.Fatalf("after overflow: length = %d, want 3", len(msgs))
	}
	if msgs[0].Content != "Q2" {
		t.Errorf("first message = %q, want Q2", msgs[0].Content)
	}
	if msgs[1].Content != "A2" {
		t.Errorf("second message = %q, want A2", msgs[1].Content)
	}
	if msgs[2].Content != "Q3" {
		t.Errorf("third message = %q, want Q3", msgs[2].Content)
	}
}

// ==================== GetShortTermMemory ====================

func TestGetShortTermMemory_DeepCopy(t *testing.T) {
	mem := NewConversationMemory(10)
	mem.AddMessage("user", "原始消息")

	msgs := mem.GetShortTermMemory()
	// 修改返回的切片不应影响原始数据
	msgs[0].Content = "被修改"

	original := mem.GetShortTermMemory()
	if original[0].Content != "原始消息" {
		t.Errorf("deep copy failed: original content = %q, want 原始消息", original[0].Content)
	}
}

func TestGetShortTermMemory_Empty(t *testing.T) {
	mem := NewConversationMemory(10)
	msgs := mem.GetShortTermMemory()
	if len(msgs) != 0 {
		t.Errorf("empty memory: length = %d, want 0", len(msgs))
	}
}

// ==================== GetLongTermSummary & UpdateSummary ====================

func TestGetLongTermSummary_InitiallyEmpty(t *testing.T) {
	mem := NewConversationMemory(10)
	if s := mem.GetLongTermSummary(); s != "" {
		t.Errorf("initial summary = %q, want empty", s)
	}
}

func TestUpdateSummary_SetsSummary(t *testing.T) {
	mem := NewConversationMemory(10)
	mem.UpdateSummary("用户讨论了 Go 并发编程和 channel 的使用。")

	s := mem.GetLongTermSummary()
	if s != "用户讨论了 Go 并发编程和 channel 的使用。" {
		t.Errorf("summary = %q, want expected text", s)
	}
}

func TestUpdateSummary_OverwritesPrevious(t *testing.T) {
	mem := NewConversationMemory(10)
	mem.UpdateSummary("第一版摘要")
	mem.UpdateSummary("第二版摘要（合并了新内容）")

	s := mem.GetLongTermSummary()
	if s != "第二版摘要（合并了新内容）" {
		t.Errorf("summary = %q, want 第二版摘要（合并了新内容）", s)
	}
}

// ==================== IncrementRound ====================

func TestIncrementRound_StartsAtZero(t *testing.T) {
	mem := NewConversationMemory(10)
	// roundCount 初始为 0，NeedsSummary 应为 false
	if mem.NeedsSummary() {
		t.Error("NeedsSummary() = true at roundCount 0, want false")
	}
}

func TestIncrementRound_IncrementsEachCall(t *testing.T) {
	mem := NewConversationMemory(10)
	for i := 0; i < 5; i++ {
		mem.IncrementRound()
	}
	// roundCount 应为 5
	if mem.NeedsSummary() {
		t.Error("NeedsSummary() = true at roundCount 5, want false")
	}
}

// ==================== NeedsSummary ====================

func TestNeedsSummary_VariousRoundCounts(t *testing.T) {
	tests := []struct {
		rounds int
		want   bool
	}{
		{0, false},   // 初始状态
		{1, false},   // 1 轮
		{5, false},   // 5 轮
		{9, false},   // 9 轮（差 1 轮）
		{10, true},   // 正好 10 轮 → 触发
		{11, false},  // 11 轮（已过触发点）
		{15, false},  // 15 轮
		{19, false},  // 19 轮
		{20, true},   // 正好 20 轮 → 再次触发
		{21, false},  // 21 轮
		{100, true},  // 100 轮 → 触发
		{101, false}, // 101 轮
	}

	for _, tt := range tests {
		mem := NewConversationMemory(10)
		for i := 0; i < tt.rounds; i++ {
			mem.IncrementRound()
		}
		got := mem.NeedsSummary()
		if got != tt.want {
			t.Errorf("NeedsSummary() at roundCount=%d: got %v, want %v", tt.rounds, got, tt.want)
		}
	}
}

func TestNeedsSummary_AfterResetBehavior(t *testing.T) {
	// 模拟实际场景：达到 10 轮后触发摘要，然后继续对话
	mem := NewConversationMemory(10)

	// 达到 10 轮
	for i := 0; i < 10; i++ {
		mem.IncrementRound()
	}
	if !mem.NeedsSummary() {
		t.Error("NeedsSummary() should be true at round 10")
	}

	// 摘要完成后继续对话（roundCount 不重置，只继续递增）
	mem.IncrementRound() // 11
	if mem.NeedsSummary() {
		t.Error("NeedsSummary() should be false at round 11")
	}

	// 继续到 20 轮
	for i := 12; i <= 20; i++ {
		mem.IncrementRound()
	}
	if !mem.NeedsSummary() {
		t.Error("NeedsSummary() should be true at round 20")
	}
}

// ==================== GetBufferText ====================

func TestGetBufferText_Empty(t *testing.T) {
	mem := NewConversationMemory(10)
	text := mem.GetBufferText()
	if text != "" {
		t.Errorf("GetBufferText empty: got %q, want empty", text)
	}
}

func TestGetBufferText_RoleLabels(t *testing.T) {
	mem := NewConversationMemory(10)
	mem.AddMessage("user", "你好")
	mem.AddMessage("assistant", "你好！有什么可以帮你的？")

	text := mem.GetBufferText()
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) != 2 {
		t.Fatalf("GetBufferText lines = %d, want 2", len(lines))
	}
	if !strings.HasPrefix(lines[0], "用户：") {
		t.Errorf("user line prefix: got %q, want 用户：", lines[0])
	}
	if !strings.HasPrefix(lines[1], "助手：") {
		t.Errorf("assistant line prefix: got %q, want 助手：", lines[1])
	}
}

func TestGetBufferText_Format(t *testing.T) {
	mem := NewConversationMemory(10)
	mem.AddMessage("user", "什么是闭包？")
	mem.AddMessage("assistant", "闭包是一个函数值，它引用了其外部作用域的变量。")

	text := mem.GetBufferText()
	expected := "用户：什么是闭包？\n助手：闭包是一个函数值，它引用了其外部作用域的变量。\n"
	if text != expected {
		t.Errorf("GetBufferText:\ngot:  %q\nwant: %q", text, expected)
	}
}

// ==================== BuildContext ====================

func TestBuildContext_AllEmpty(t *testing.T) {
	mem := NewConversationMemory(10)
	ctx := mem.BuildContext("")
	if ctx != "" {
		t.Errorf("BuildContext all empty: got %q, want empty", ctx)
	}
}

func TestBuildContext_SummaryOnly(t *testing.T) {
	mem := NewConversationMemory(10)
	mem.UpdateSummary("用户正在学习 Go 语言基础。")

	ctx := mem.BuildContext("")
	if !strings.Contains(ctx, "【历史对话摘要】") {
		t.Error("BuildContext missing 历史对话摘要 header")
	}
	if !strings.Contains(ctx, "用户正在学习 Go 语言基础。") {
		t.Error("BuildContext missing summary content")
	}
	if strings.Contains(ctx, "【相关学习材料】") {
		t.Error("BuildContext should not contain RAG section when empty")
	}
	if strings.Contains(ctx, "【最近对话记录】") {
		t.Error("BuildContext should not contain buffer section when empty")
	}
}

func TestBuildContext_RAGOnly(t *testing.T) {
	mem := NewConversationMemory(10)
	ragCtx := "Go 语言的并发模型使用 goroutine 和 channel。"

	ctx := mem.BuildContext(ragCtx)
	if !strings.Contains(ctx, "【相关学习材料】") {
		t.Error("BuildContext missing RAG header")
	}
	if !strings.Contains(ctx, ragCtx) {
		t.Error("BuildContext missing RAG content")
	}
	if strings.Contains(ctx, "【历史对话摘要】") {
		t.Error("BuildContext should not contain summary section when empty")
	}
}

func TestBuildContext_BufferOnly(t *testing.T) {
	mem := NewConversationMemory(10)
	mem.AddMessage("user", "继续上次的话题")
	mem.AddMessage("assistant", "好的，上次我们讨论了...")

	ctx := mem.BuildContext("")
	if !strings.Contains(ctx, "【最近对话记录】") {
		t.Error("BuildContext missing buffer header")
	}
	if !strings.Contains(ctx, "用户：继续上次的话题") {
		t.Error("BuildContext missing user message")
	}
	if !strings.Contains(ctx, "助手：好的，上次我们讨论了...") {
		t.Error("BuildContext missing assistant message")
	}
}

func TestBuildContext_AllThreeLayers(t *testing.T) {
	mem := NewConversationMemory(10)
	mem.UpdateSummary("用户之前讨论了 Go 并发编程。")
	mem.AddMessage("user", "channel 怎么使用？")
	mem.AddMessage("assistant", "使用 make(chan Type) 创建...")

	ragCtx := "channel 是 Go 语言中用于 goroutine 之间通信的机制。"
	ctx := mem.BuildContext(ragCtx)

	// 验证三层都存在
	if !strings.Contains(ctx, "【历史对话摘要】") {
		t.Error("missing summary section")
	}
	if !strings.Contains(ctx, "【相关学习材料】") {
		t.Error("missing RAG section")
	}
	if !strings.Contains(ctx, "【最近对话记录】") {
		t.Error("missing buffer section")
	}

	// 验证顺序：摘要 < RAG < 短期对话
	summaryIdx := strings.Index(ctx, "【历史对话摘要】")
	ragIdx := strings.Index(ctx, "【相关学习材料】")
	bufferIdx := strings.Index(ctx, "【最近对话记录】")

	if summaryIdx >= ragIdx {
		t.Errorf("summary (idx %d) should appear before RAG (idx %d)", summaryIdx, ragIdx)
	}
	if ragIdx >= bufferIdx {
		t.Errorf("RAG (idx %d) should appear before buffer (idx %d)", ragIdx, bufferIdx)
	}
}

func TestBuildContext_IncludesSummaryAfterUpdate(t *testing.T) {
	mem := NewConversationMemory(10)

	// 模拟完整对话流程
	for i := 0; i < 5; i++ {
		mem.AddMessage("user", "问题"+string(rune('1'+i)))
		mem.AddMessage("assistant", "回答"+string(rune('1'+i)))
	}

	// 模拟摘要生成
	mem.UpdateSummary("用户提出了5个关于Go语言的问题，涵盖了并发、channel和goroutine等主题。")

	ctx := mem.BuildContext("")
	if !strings.Contains(ctx, "用户提出了5个关于Go语言的问题") {
		t.Error("BuildContext should include updated summary")
	}
	if !strings.Contains(ctx, "【最近对话记录】") {
		t.Error("BuildContext should still include recent messages")
	}
}

// ==================== 综合场景测试 ====================

func TestConversationMemory_FullConversationFlow(t *testing.T) {
	mem := NewConversationMemory(5) // 小缓冲区方便测试滑动窗口

	// 模拟 12 轮对话
	for i := 0; i < 12; i++ {
		mem.AddMessage("user", "第"+string(rune('1'+i))+"个问题")
		mem.AddMessage("assistant", "第"+string(rune('1'+i))+"个回答")
		mem.IncrementRound()
	}

	// 12 轮，NeedsSummary 在 10 时为 true，12 时为 false
	if mem.NeedsSummary() {
		t.Error("NeedsSummary should be false at round 12")
	}

	// 验证短期记忆（滑动窗口生效）
	msgs := mem.GetShortTermMemory()
	// bufferSize=5, 阈值=10条, 24条消息会多次触发滑动窗口
	if len(msgs) > 10 {
		t.Errorf("short term memory length = %d, should be ≤ 10", len(msgs))
	}

	// 设置摘要后 BuildContext 应包含摘要
	mem.UpdateSummary("这是一段很长的对话摘要...")
	ctx := mem.BuildContext("一些 RAG 上下文")
	if !strings.Contains(ctx, "【历史对话摘要】") {
		t.Error("BuildContext missing summary after UpdateSummary")
	}
	if !strings.Contains(ctx, "【相关学习材料】") {
		t.Error("BuildContext missing RAG")
	}
}

// ==================== 并发安全测试 ====================

func TestConversationMemory_ConcurrentSafety(t *testing.T) {
	mem := NewConversationMemory(100)
	var wg sync.WaitGroup

	// 10 个 goroutine 并发写消息
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				mem.AddMessage("user", "goroutine消息")
			}
		}(i)
	}

	// 5 个 goroutine 并发读
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_ = mem.GetShortTermMemory()
				_ = mem.GetLongTermSummary()
				_ = mem.GetBufferText()
				_ = mem.BuildContext("rag")
				_ = mem.NeedsSummary()
			}
		}()
	}

	// 3 个 goroutine 并发 IncrementRound + UpdateSummary
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 30; j++ {
				mem.IncrementRound()
				mem.UpdateSummary("并发摘要更新")
			}
		}()
	}

	wg.Wait()
	// 如果执行到这里没有 race condition 或 panic，测试通过
}

func TestConversationMemory_ConcurrentReadWrite(t *testing.T) {
	mem := NewConversationMemory(10)
	var wg sync.WaitGroup

	// writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			mem.AddMessage("user", "写入")
			mem.AddMessage("assistant", "回复")
			mem.IncrementRound()
		}
	}()

	// reader
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			ctx := mem.BuildContext("rag context")
			_ = len(ctx)
		}
	}()

	wg.Wait()
}

// ==================== 边界条件 ====================

func TestNewConversationMemory_SizeOne(t *testing.T) {
	mem := NewConversationMemory(1) // 最小有效缓冲区

	mem.AddMessage("user", "Q1")
	mem.AddMessage("assistant", "A1")
	mem.AddMessage("user", "Q2")
	// 3 条 > 1*2=2, 触发滑动窗口 → 移除 Q1+A1 → 剩 Q2
	// 但 Q2 只有 1 条，移除前 2 条后只剩 Q2
	// 实际上：添加 A1 后 2 条，不大于 2，不触发；添加 Q2 后 3 条 > 2，触发移除前 2 条

	msgs := mem.GetShortTermMemory()
	if len(msgs) != 1 {
		t.Fatalf("bufferSize=1: length = %d, want 1", len(msgs))
	}
	if msgs[0].Content != "Q2" {
		t.Errorf("bufferSize=1: remaining message = %q, want Q2", msgs[0].Content)
	}
}

func TestAddMessage_EmptyContent(t *testing.T) {
	mem := NewConversationMemory(10)
	mem.AddMessage("user", "")
	mem.AddMessage("assistant", "")

	msgs := mem.GetShortTermMemory()
	if len(msgs) != 2 {
		t.Fatalf("empty content: length = %d, want 2", len(msgs))
	}
	if msgs[0].Content != "" || msgs[1].Content != "" {
		t.Error("empty content messages should preserve empty string")
	}
}

func TestBuildContext_SummaryAndBufferNoRAG(t *testing.T) {
	mem := NewConversationMemory(10)
	mem.UpdateSummary("摘要内容")
	mem.AddMessage("user", "新消息")

	ctx := mem.BuildContext("")
	if !strings.Contains(ctx, "【历史对话摘要】") {
		t.Error("missing summary")
	}
	if strings.Contains(ctx, "【相关学习材料】") {
		t.Error("should not have RAG section")
	}
	if !strings.Contains(ctx, "【最近对话记录】") {
		t.Error("missing buffer section")
	}
}

func TestBuildContext_RAGAndBufferNoSummary(t *testing.T) {
	mem := NewConversationMemory(10)
	mem.AddMessage("user", "消息")

	ctx := mem.BuildContext("RAG 检索结果")
	if strings.Contains(ctx, "【历史对话摘要】") {
		t.Error("should not have summary section")
	}
	if !strings.Contains(ctx, "【相关学习材料】") {
		t.Error("missing RAG section")
	}
	if !strings.Contains(ctx, "【最近对话记录】") {
		t.Error("missing buffer section")
	}
}
