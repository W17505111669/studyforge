package model

import (
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ==================== 辅助函数 ====================

// setupModelTestDB 创建 SQLite 内存数据库并自动迁移所有模型
func setupModelTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("测试数据库初始化失败: %v", err)
	}
	err = db.AutoMigrate(
		&User{},
		&Material{},
		&Card{},
		&Quiz{},
		&QuizAttempt{},
		&LLMTrace{},
		&Conversation{},
		&ChatMessage{},
		&UserAchievement{},
	)
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

// isValidUUID 检查字符串是否为合法 UUID v4 格式
func isValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// ==================== BeforeCreate UUID 自动生成测试 ====================

func TestUser_BeforeCreate_GeneratesUUID(t *testing.T) {
	db := setupModelTestDB(t)

	user := User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpw",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	if user.ID == "" {
		t.Error("用户 ID 不应为空")
	}
	if !isValidUUID(user.ID) {
		t.Errorf("用户 ID 不是合法 UUID: %s", user.ID)
	}
}

func TestUser_BeforeCreate_PreservesExistingID(t *testing.T) {
	db := setupModelTestDB(t)

	customID := "custom-user-id-12345"
	user := User{
		ID:       customID,
		Username: "testuser2",
		Email:    "test2@example.com",
		Password: "hashedpw",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	if user.ID != customID {
		t.Errorf("已有 ID 不应被覆盖，期望 %s，实际 %s", customID, user.ID)
	}
}

func TestMaterial_BeforeCreate_GeneratesUUID(t *testing.T) {
	db := setupModelTestDB(t)

	// 先创建关联用户
	user := User{Username: "matuser", Email: "mat@test.com", Password: "pw"}
	db.Create(&user)

	mat := Material{
		UserID:      user.ID,
		Title:       "测试材料",
		ContentType: "text",
		Content:     "这是测试内容",
	}
	if err := db.Create(&mat).Error; err != nil {
		t.Fatalf("创建材料失败: %v", err)
	}
	if mat.ID == "" {
		t.Error("材料 ID 不应为空")
	}
	if !isValidUUID(mat.ID) {
		t.Errorf("材料 ID 不是合法 UUID: %s", mat.ID)
	}
}

func TestCard_BeforeCreate_GeneratesUUID(t *testing.T) {
	db := setupModelTestDB(t)

	user := User{Username: "carduser", Email: "card@test.com", Password: "pw"}
	db.Create(&user)
	mat := Material{UserID: user.ID, Title: "材料", ContentType: "text"}
	db.Create(&mat)

	card := Card{
		UserID:     user.ID,
		MaterialID: mat.ID,
		Concept:    "测试概念",
		Detail:     "测试详情",
		EaseFactor: 2.5,
	}
	if err := db.Create(&card).Error; err != nil {
		t.Fatalf("创建卡片失败: %v", err)
	}
	if card.ID == "" {
		t.Error("卡片 ID 不应为空")
	}
	if !isValidUUID(card.ID) {
		t.Errorf("卡片 ID 不是合法 UUID: %s", card.ID)
	}
}

func TestQuiz_BeforeCreate_GeneratesUUID(t *testing.T) {
	db := setupModelTestDB(t)

	user := User{Username: "quizuser", Email: "quiz@test.com", Password: "pw"}
	db.Create(&user)
	mat := Material{UserID: user.ID, Title: "材料", ContentType: "text"}
	db.Create(&mat)

	quiz := Quiz{
		UserID:     user.ID,
		MaterialID: mat.ID,
		Question:   "1+1=?",
		Type:       "choice",
		Difficulty: "easy",
		Options:    `["1","2","3","4"]`,
		Answer:     "2",
	}
	if err := db.Create(&quiz).Error; err != nil {
		t.Fatalf("创建练习题失败: %v", err)
	}
	if quiz.ID == "" {
		t.Error("练习题 ID 不应为空")
	}
	if !isValidUUID(quiz.ID) {
		t.Errorf("练习题 ID 不是合法 UUID: %s", quiz.ID)
	}
}

func TestQuizAttempt_BeforeCreate_GeneratesUUID(t *testing.T) {
	db := setupModelTestDB(t)

	user := User{Username: "attuser", Email: "att@test.com", Password: "pw"}
	db.Create(&user)
	mat := Material{UserID: user.ID, Title: "材料", ContentType: "text"}
	db.Create(&mat)
	quiz := Quiz{UserID: user.ID, MaterialID: mat.ID, Question: "Q?", Type: "choice", Answer: "A"}
	db.Create(&quiz)

	attempt := QuizAttempt{
		UserID:    user.ID,
		QuizID:    quiz.ID,
		Answer:    "A",
		IsCorrect: true,
	}
	if err := db.Create(&attempt).Error; err != nil {
		t.Fatalf("创建答题记录失败: %v", err)
	}
	if attempt.ID == "" {
		t.Error("答题记录 ID 不应为空")
	}
	if !isValidUUID(attempt.ID) {
		t.Errorf("答题记录 ID 不是合法 UUID: %s", attempt.ID)
	}
}

func TestConversation_BeforeCreate_GeneratesUUID(t *testing.T) {
	db := setupModelTestDB(t)

	user := User{Username: "convuser", Email: "conv@test.com", Password: "pw"}
	db.Create(&user)

	conv := Conversation{
		UserID: user.ID,
		Title:  "测试对话",
	}
	if err := db.Create(&conv).Error; err != nil {
		t.Fatalf("创建对话失败: %v", err)
	}
	if conv.ID == "" {
		t.Error("对话 ID 不应为空")
	}
	if !isValidUUID(conv.ID) {
		t.Errorf("对话 ID 不是合法 UUID: %s", conv.ID)
	}
}

func TestChatMessage_BeforeCreate_GeneratesUUID(t *testing.T) {
	db := setupModelTestDB(t)

	user := User{Username: "msguser", Email: "msg@test.com", Password: "pw"}
	db.Create(&user)
	conv := Conversation{UserID: user.ID, Title: "对话"}
	db.Create(&conv)

	msg := ChatMessage{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        "你好",
	}
	if err := db.Create(&msg).Error; err != nil {
		t.Fatalf("创建消息失败: %v", err)
	}
	if msg.ID == "" {
		t.Error("消息 ID 不应为空")
	}
	if !isValidUUID(msg.ID) {
		t.Errorf("消息 ID 不是合法 UUID: %s", msg.ID)
	}
}

func TestUserAchievement_BeforeCreate_GeneratesUUID(t *testing.T) {
	db := setupModelTestDB(t)

	user := User{Username: "achuser", Email: "ach@test.com", Password: "pw"}
	db.Create(&user)

	ua := UserAchievement{
		UserID:        user.ID,
		AchievementID: "first_upload",
		UnlockedAt:    time.Now(),
	}
	if err := db.Create(&ua).Error; err != nil {
		t.Fatalf("创建成就记录失败: %v", err)
	}
	if ua.ID == "" {
		t.Error("成就记录 ID 不应为空")
	}
	if !isValidUUID(ua.ID) {
		t.Errorf("成就记录 ID 不是合法 UUID: %s", ua.ID)
	}
}

func TestLLMTrace_BeforeCreate_GeneratesUUID_AndSetsLatencyMs(t *testing.T) {
	db := setupModelTestDB(t)

	user := User{Username: "traceuser", Email: "trace@test.com", Password: "pw"}
	db.Create(&user)

	trace := LLMTrace{
		UserID:       user.ID,
		AgentName:    "Analyst",
		Model:        "qwen-plus",
		InputTokens:  100,
		OutputTokens: 200,
		TotalTokens:  300,
		Latency:      2500 * time.Millisecond,
		Status:       "success",
	}
	if err := db.Create(&trace).Error; err != nil {
		t.Fatalf("创建追踪记录失败: %v", err)
	}
	if trace.ID == "" {
		t.Error("追踪记录 ID 不应为空")
	}
	if !isValidUUID(trace.ID) {
		t.Errorf("追踪记录 ID 不是合法 UUID: %s", trace.ID)
	}
	// LatencyMs 应在 BeforeCreate 中自动设置
	if trace.LatencyMs != 2500 {
		t.Errorf("LatencyMs 期望 2500，实际 %d", trace.LatencyMs)
	}
}

// TestAllModels_UUIDUniqueness 批量创建多条记录，验证 UUID 唯一性
func TestAllModels_UUIDUniqueness(t *testing.T) {
	db := setupModelTestDB(t)

	user1 := User{Username: "u1", Email: "u1@test.com", Password: "pw"}
	user2 := User{Username: "u2", Email: "u2@test.com", Password: "pw"}
	db.Create(&user1)
	db.Create(&user2)

	if user1.ID == user2.ID {
		t.Error("不同用户不应生成相同 UUID")
	}

	// 验证多个材料的 UUID 唯一性
	mat1 := Material{UserID: user1.ID, Title: "M1", ContentType: "text"}
	mat2 := Material{UserID: user1.ID, Title: "M2", ContentType: "text"}
	db.Create(&mat1)
	db.Create(&mat2)
	if mat1.ID == mat2.ID {
		t.Error("不同材料不应生成相同 UUID")
	}

	// 验证多个卡片的 UUID 唯一性
	card1 := Card{UserID: user1.ID, MaterialID: mat1.ID, Concept: "C1", EaseFactor: 2.5}
	card2 := Card{UserID: user1.ID, MaterialID: mat1.ID, Concept: "C2", EaseFactor: 2.5}
	db.Create(&card1)
	db.Create(&card2)
	if card1.ID == card2.ID {
		t.Error("不同卡片不应生成相同 UUID")
	}
}

// ==================== UploadRequest 验证测试 ====================

// validateStruct 使用 go-playground/validator 验证结构体（模拟 Gin 的 binding 行为）
// Gin 使用 "binding" 标签名而非默认的 "validate"，需要设置标签名
func validateStruct(v interface{}) error {
	validate := validator.New()
	validate.SetTagName("binding")
	return validate.Struct(v)
}

func TestUploadRequest_Validation_RequiredTitle(t *testing.T) {
	req := UploadRequest{
		Title:       "",
		ContentType: "text",
	}
	err := validateStruct(req)
	if err == nil {
		t.Error("Title 为空时应验证失败")
	}
	if !strings.Contains(err.Error(), "Title") {
		t.Errorf("错误应包含 Title 字段，实际: %v", err)
	}
}

func TestUploadRequest_Validation_RequiredContentType(t *testing.T) {
	req := UploadRequest{
		Title:       "测试材料",
		ContentType: "",
	}
	err := validateStruct(req)
	if err == nil {
		t.Error("ContentType 为空时应验证失败")
	}
	if !strings.Contains(err.Error(), "ContentType") {
		t.Errorf("错误应包含 ContentType 字段，实际: %v", err)
	}
}

func TestUploadRequest_Validation_ContentTypeOneof(t *testing.T) {
	// 测试非法值
	req := UploadRequest{
		Title:       "测试",
		ContentType: "pdf", // 不在 oneof=text url 中
	}
	err := validateStruct(req)
	if err == nil {
		t.Error("ContentType 为 pdf 时应验证失败（oneof=text url 约束）")
	}

	// 测试合法值 "text"
	reqText := UploadRequest{
		Title:       "测试",
		ContentType: "text",
	}
	if err := validateStruct(reqText); err != nil {
		t.Errorf("ContentType 为 text 时应验证通过，实际错误: %v", err)
	}

	// 测试合法值 "url"
	reqURL := UploadRequest{
		Title:       "测试",
		ContentType: "url",
		SourceURL:   "https://example.com",
	}
	if err := validateStruct(reqURL); err != nil {
		t.Errorf("ContentType 为 url 时应验证通过，实际错误: %v", err)
	}
}

func TestUploadRequest_Validation_ValidRequest(t *testing.T) {
	req := UploadRequest{
		Title:       "Go 并发编程",
		ContentType: "text",
		Content:     "这是一段关于 Go 并发的学习材料...",
	}
	if err := validateStruct(req); err != nil {
		t.Errorf("完整有效请求应验证通过，实际错误: %v", err)
	}
}

// ==================== Card SM-2 间隔重复计算测试 ====================

func TestCard_ApplyReview_Mastered_FirstReview(t *testing.T) {
	card := Card{
		Concept:      "goroutine",
		EaseFactor:   2.5,
		IntervalDays: 0, // 新卡片
		ReviewCount:  0,
	}

	nextReview := card.ApplyReview("mastered")

	if card.ReviewCount != 1 {
		t.Errorf("ReviewCount 期望 1，实际 %d", card.ReviewCount)
	}
	if card.IntervalDays != 1 {
		t.Errorf("首次 mastered 后 IntervalDays 期望 1，实际 %d", card.IntervalDays)
	}
	if card.EaseFactor != 2.6 {
		t.Errorf("EaseFactor 期望 2.6，实际 %.2f", card.EaseFactor)
	}
	if card.LastReviewedAt == nil {
		t.Error("LastReviewedAt 不应为 nil")
	}
	if card.NextReviewAt == nil {
		t.Error("NextReviewAt 不应为 nil")
	}
	// nextReview 应为大约 1 天后
	expectedDate := time.Now().AddDate(0, 0, 1)
	diff := expectedDate.Sub(nextReview)
	if diff < -time.Minute || diff > time.Minute {
		t.Errorf("nextReview 期望约 1 天后，实际 %v", nextReview)
	}
}

func TestCard_ApplyReview_Mastered_SecondReview(t *testing.T) {
	card := Card{
		Concept:      "channel",
		EaseFactor:   2.6, // 已复习过一次
		IntervalDays: 1,   // 当前间隔 1 天
		ReviewCount:  1,
	}

	card.ApplyReview("mastered")

	if card.ReviewCount != 2 {
		t.Errorf("ReviewCount 期望 2，实际 %d", card.ReviewCount)
	}
	if card.IntervalDays != 3 {
		t.Errorf("第二次 mastered 后 IntervalDays 期望 3，实际 %d", card.IntervalDays)
	}
	if card.EaseFactor != 2.7 {
		t.Errorf("EaseFactor 期望 2.7，实际 %.2f", card.EaseFactor)
	}
}

func TestCard_ApplyReview_Mastered_ThirdReview(t *testing.T) {
	card := Card{
		Concept:      "mutex",
		EaseFactor:   2.7,
		IntervalDays: 3, // 当前间隔 3 天
		ReviewCount:  2,
	}

	card.ApplyReview("mastered")

	// 第三次起：interval = int(3 * 2.7) = int(8.1) = 8
	if card.IntervalDays != 8 {
		t.Errorf("第三次 mastered 后 IntervalDays 期望 8，实际 %d", card.IntervalDays)
	}
	if card.ReviewCount != 3 {
		t.Errorf("ReviewCount 期望 3，实际 %d", card.ReviewCount)
	}
}

func TestCard_ApplyReview_Mastered_EaseFactorCap(t *testing.T) {
	card := Card{
		Concept:      "select",
		EaseFactor:   2.95, // 接近上限
		IntervalDays: 10,
		ReviewCount:  5,
	}

	card.ApplyReview("mastered")

	// 2.95 + 0.1 = 3.05 → 被限制到 3.0
	if card.EaseFactor != 3.0 {
		t.Errorf("EaseFactor 不应超过 3.0，实际 %.2f", card.EaseFactor)
	}
}

func TestCard_ApplyReview_Mastered_AlreadyAtCap(t *testing.T) {
	card := Card{
		Concept:      "defer",
		EaseFactor:   3.0, // 已在上限
		IntervalDays: 20,
		ReviewCount:  10,
	}

	card.ApplyReview("mastered")

	if card.EaseFactor != 3.0 {
		t.Errorf("EaseFactor 应保持 3.0，实际 %.2f", card.EaseFactor)
	}
	// interval = int(20 * 3.0) = 60
	if card.IntervalDays != 60 {
		t.Errorf("IntervalDays 期望 60，实际 %d", card.IntervalDays)
	}
}

func TestCard_ApplyReview_Review_ResetInterval(t *testing.T) {
	card := Card{
		Concept:      "interface",
		EaseFactor:   2.5,
		IntervalDays: 10, // 之前间隔 10 天
		ReviewCount:  5,
	}

	card.ApplyReview("review")

	if card.IntervalDays != 1 {
		t.Errorf("review 后 IntervalDays 期望重置为 1，实际 %d", card.IntervalDays)
	}
	if card.ReviewCount != 6 {
		t.Errorf("ReviewCount 期望 6，实际 %d", card.ReviewCount)
	}
	if card.EaseFactor != 2.3 {
		t.Errorf("EaseFactor 期望 2.3（2.5-0.2），实际 %.2f", card.EaseFactor)
	}
}

func TestCard_ApplyReview_Review_EaseFactorFloor(t *testing.T) {
	card := Card{
		Concept:      "pointer",
		EaseFactor:   1.35, // 接近下限
		IntervalDays: 1,
		ReviewCount:  3,
	}

	card.ApplyReview("review")

	// 1.35 - 0.2 = 1.15 → 被限制到 1.3
	if card.EaseFactor != 1.3 {
		t.Errorf("EaseFactor 不应低于 1.3，实际 %.2f", card.EaseFactor)
	}
}

func TestCard_ApplyReview_Review_AlreadyAtFloor(t *testing.T) {
	card := Card{
		Concept:      "slice",
		EaseFactor:   1.3, // 已在下限
		IntervalDays: 1,
		ReviewCount:  8,
	}

	card.ApplyReview("review")

	if card.EaseFactor != 1.3 {
		t.Errorf("EaseFactor 应保持 1.3，实际 %.2f", card.EaseFactor)
	}
}

func TestCard_ApplyReview_MultipleMastered_Sequence(t *testing.T) {
	// 模拟连续多次 mastered 的完整间隔递增序列
	card := Card{
		Concept:    "Go 并发全链路",
		EaseFactor: 2.5,
	}

	// 第 1 次 mastered：0 → 1
	card.ApplyReview("mastered")
	if card.IntervalDays != 1 {
		t.Errorf("第1次 mastered: IntervalDays 期望 1，实际 %d", card.IntervalDays)
	}

	// 第 2 次 mastered：1 → 3
	card.ApplyReview("mastered")
	if card.IntervalDays != 3 {
		t.Errorf("第2次 mastered: IntervalDays 期望 3，实际 %d", card.IntervalDays)
	}

	// 第 3 次 mastered：3 → int(3 * 2.7) = 8
	card.ApplyReview("mastered")
	if card.IntervalDays != 8 {
		t.Errorf("第3次 mastered: IntervalDays 期望 8，实际 %d", card.IntervalDays)
	}

	// 第 4 次 mastered：8 → int(8 * 2.8) = 22
	card.ApplyReview("mastered")
	if card.IntervalDays != 22 {
		t.Errorf("第4次 mastered: IntervalDays 期望 22，实际 %d", card.IntervalDays)
	}

	if card.ReviewCount != 4 {
		t.Errorf("ReviewCount 期望 4，实际 %d", card.ReviewCount)
	}
}

func TestCard_ApplyReview_ReviewThenMastered(t *testing.T) {
	// 模拟：先 review（降低间隔），再 mastered（重新开始递增）
	card := Card{
		Concept:      "map",
		EaseFactor:   2.5,
		IntervalDays: 10,
		ReviewCount:  5,
	}

	// review 重置
	card.ApplyReview("review")
	if card.IntervalDays != 1 {
		t.Errorf("review 后 IntervalDays 期望 1，实际 %d", card.IntervalDays)
	}
	if card.EaseFactor != 2.3 {
		t.Errorf("review 后 EaseFactor 期望 2.3，实际 %.2f", card.EaseFactor)
	}

	// mastered 从 1 开始恢复
	card.ApplyReview("mastered")
	if card.IntervalDays != 3 {
		t.Errorf("mastered 后 IntervalDays 期望 3，实际 %d", card.IntervalDays)
	}
}

// ==================== Achievement 定义完整性测试 ====================

func TestAllAchievements_NoDuplicateIDs(t *testing.T) {
	seen := make(map[string]bool)
	for _, a := range AllAchievements {
		if seen[a.ID] {
			t.Errorf("成就 ID 重复: %s", a.ID)
		}
		seen[a.ID] = true
	}
}

func TestAllAchievements_RequiredFields(t *testing.T) {
	for _, a := range AllAchievements {
		if a.ID == "" {
			t.Errorf("成就缺少 ID")
		}
		if a.Name == "" {
			t.Errorf("成就 %s 缺少 Name", a.ID)
		}
		if a.Description == "" {
			t.Errorf("成就 %s 缺少 Description", a.ID)
		}
		if a.Icon == "" {
			t.Errorf("成就 %s 缺少 Icon", a.ID)
		}
		if a.Category == "" {
			t.Errorf("成就 %s 缺少 Category", a.ID)
		}
		if a.Target <= 0 {
			t.Errorf("成就 %s 的 Target 应大于 0，实际 %d", a.ID, a.Target)
		}
	}
}

func TestAllAchievements_ValidCategories(t *testing.T) {
	validCategories := map[string]bool{
		"learning":    true,
		"practice":    true,
		"review":      true,
		"exploration": true,
		"special":     true,
	}

	for _, a := range AllAchievements {
		if !validCategories[a.Category] {
			t.Errorf("成就 %s 的 Category '%s' 不在合法类别中", a.ID, a.Category)
		}
	}
}

func TestAllAchievements_Count(t *testing.T) {
	// 确保至少有 18 种成就（当前已定义的数量）
	if len(AllAchievements) < 18 {
		t.Errorf("成就数量不应少于 18 种，实际 %d", len(AllAchievements))
	}
}

func TestAchievementCategories_MatchesAllCategories(t *testing.T) {
	// AchievementCategories map 应覆盖所有在 AllAchievements 中出现的类别
	usedCategories := make(map[string]bool)
	for _, a := range AllAchievements {
		usedCategories[a.Category] = true
	}

	for cat := range usedCategories {
		if _, ok := AchievementCategories[cat]; !ok {
			t.Errorf("AchievementCategories 缺少类别 '%s'，但 AllAchievements 中存在该类别", cat)
		}
	}
}

// ==================== LLMTrace LatencyMs 计算测试 ====================

func TestLLMTrace_LatencyMsConversion(t *testing.T) {
	db := setupModelTestDB(t)

	user := User{Username: "latuser", Email: "lat@test.com", Password: "pw"}
	db.Create(&user)

	// 测试不同延迟值的转换
	testCases := []struct {
		name     string
		latency  time.Duration
		expected int64
	}{
		{"1秒", 1 * time.Second, 1000},
		{"500毫秒", 500 * time.Millisecond, 500},
		{"2.5秒", 2500 * time.Millisecond, 2500},
		{"0.1秒", 100 * time.Millisecond, 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trace := LLMTrace{
				UserID:    user.ID,
				AgentName: "Analyst",
				Latency:   tc.latency,
				Status:    "success",
			}
			if err := db.Create(&trace).Error; err != nil {
				t.Fatalf("创建追踪记录失败: %v", err)
			}
			if trace.LatencyMs != tc.expected {
				t.Errorf("LatencyMs 期望 %d，实际 %d", tc.expected, trace.LatencyMs)
			}
		})
	}
}

// ==================== RegisterRequest / LoginRequest 验证测试 ====================

func TestRegisterRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     RegisterRequest
		wantErr bool
		errField string
	}{
		{
			name:    "有效请求",
			req:     RegisterRequest{Username: "testuser", Email: "test@example.com", Password: "password123"},
			wantErr: false,
		},
		{
			name:     "用户名过短",
			req:      RegisterRequest{Username: "ab", Email: "test@example.com", Password: "password123"},
			wantErr:  true,
			errField: "Username",
		},
		{
			name:     "用户名为空",
			req:      RegisterRequest{Username: "", Email: "test@example.com", Password: "password123"},
			wantErr:  true,
			errField: "Username",
		},
		{
			name:     "邮箱为空",
			req:      RegisterRequest{Username: "testuser", Email: "", Password: "password123"},
			wantErr:  true,
			errField: "Email",
		},
		{
			name:     "无效邮箱格式",
			req:      RegisterRequest{Username: "testuser", Email: "not-an-email", Password: "password123"},
			wantErr:  true,
			errField: "Email",
		},
		{
			name:     "密码过短",
			req:      RegisterRequest{Username: "testuser", Email: "test@example.com", Password: "12345"},
			wantErr:  true,
			errField: "Password",
		},
		{
			name:     "密码为空",
			req:      RegisterRequest{Username: "testuser", Email: "test@example.com", Password: ""},
			wantErr:  true,
			errField: "Password",
		},
		{
			name:    "恰好满足最低要求",
			req:     RegisterRequest{Username: "abc", Email: "a@b.com", Password: "123456"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStruct(tt.req)
			if tt.wantErr && err == nil {
				t.Errorf("期望验证失败，但通过了")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("期望验证通过，但失败: %v", err)
			}
			if tt.wantErr && tt.errField != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errField) {
					t.Errorf("错误应包含字段 '%s'，实际: %v", tt.errField, err)
				}
			}
		})
	}
}

func TestLoginRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     LoginRequest
		wantErr bool
	}{
		{
			name:    "有效请求",
			req:     LoginRequest{Username: "testuser", Password: "password"},
			wantErr: false,
		},
		{
			name:    "用户名为空",
			req:     LoginRequest{Username: "", Password: "password"},
			wantErr: true,
		},
		{
			name:    "密码为空",
			req:     LoginRequest{Username: "testuser", Password: ""},
			wantErr: true,
		},
		{
			name:    "全部为空",
			req:     LoginRequest{Username: "", Password: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStruct(tt.req)
			if tt.wantErr && err == nil {
				t.Errorf("期望验证失败，但通过了")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("期望验证通过，但失败: %v", err)
			}
		})
	}
}

// ==================== AnswerRequest 验证测试 ====================

func TestAnswerRequest_Validation(t *testing.T) {
	// 有效
	req := AnswerRequest{Answer: "B"}
	if err := validateStruct(req); err != nil {
		t.Errorf("有效答案应通过验证: %v", err)
	}

	// 空答案
	req2 := AnswerRequest{Answer: ""}
	if err := validateStruct(req2); err == nil {
		t.Error("空答案应验证失败")
	}
}
