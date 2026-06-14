package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-for-middleware-tests"

// ────────────────────────── JWT 中间件测试 ──────────────────────────

// setupJWTRouter 搭建含 JWT 中间件的测试路由
func setupJWTRouter(secret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuth(secret))
	r.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get("userID")
		username, _ := c.Get("username")
		c.JSON(http.StatusOK, gin.H{
			"user_id":  userID,
			"username": username,
		})
	})
	return r
}

// makeToken 生成自定义 claims 的 JWT token（自动补全 iss/aud 以匹配中间件校验）
func makeToken(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()
	// 自动补全 iss/aud，与 JWTAuth 中间件的 jwt.WithIssuer/WithAudience 保持一致
	if _, ok := claims["iss"]; !ok {
		claims["iss"] = "studyforge-pro"
	}
	if _, ok := claims["aud"]; !ok {
		claims["aud"] = "studyforge-client"
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("生成 token 失败: %v", err)
	}
	return s
}

func TestJWT_NoToken_Returns401(t *testing.T) {
	router := setupJWTRouter(testSecret)
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望 401，实际 %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["error"]; !ok {
		t.Error("响应中缺少 error 字段")
	}
}

func TestJWT_InvalidFormat_Returns401(t *testing.T) {
	router := setupJWTRouter(testSecret)

	tests := []struct {
		name   string
		header string
	}{
		{"无 Bearer 前缀", "some-token-value"},
		{"仅 Bearer", "Bearer"},
		{"多余空格", "Bearer  token extra"},
		{"空 Bearer", "Bearer "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tt.header)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("[%s] 期望 401，实际 %d", tt.name, w.Code)
			}
		})
	}
}

func TestJWT_ExpiredToken_Returns401(t *testing.T) {
	router := setupJWTRouter(testSecret)

	tokenStr := makeToken(t, testSecret, jwt.MapClaims{
		"user_id":  "user-123",
		"username": "testuser",
		"exp":      time.Now().Add(-1 * time.Hour).Unix(),
		"iat":      time.Now().Add(-2 * time.Hour).Unix(),
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望 401（过期 token），实际 %d", w.Code)
	}
}

func TestJWT_WrongSecret_Returns401(t *testing.T) {
	router := setupJWTRouter(testSecret)

	// 用不同的 secret 签名
	tokenStr := makeToken(t, "wrong-secret-key", jwt.MapClaims{
		"user_id":  "user-123",
		"username": "testuser",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望 401（错误签名），实际 %d", w.Code)
	}
}

func TestJWT_MissingUserID_Returns401(t *testing.T) {
	router := setupJWTRouter(testSecret)

	tokenStr := makeToken(t, testSecret, jwt.MapClaims{
		"username": "testuser",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望 401（缺少 user_id），实际 %d", w.Code)
	}
}

func TestJWT_EmptyUserID_Returns401(t *testing.T) {
	router := setupJWTRouter(testSecret)

	tokenStr := makeToken(t, testSecret, jwt.MapClaims{
		"user_id":  "",
		"username": "testuser",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望 401（空 user_id），实际 %d", w.Code)
	}
}

func TestJWT_ValidToken_SetsContext(t *testing.T) {
	router := setupJWTRouter(testSecret)

	tokenStr := makeToken(t, testSecret, jwt.MapClaims{
		"user_id":  "user-abc-123",
		"username": "alice",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望 200，实际 %d, body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["user_id"] != "user-abc-123" {
		t.Errorf("期望 user_id=user-abc-123，实际 %v", resp["user_id"])
	}
	if resp["username"] != "alice" {
		t.Errorf("期望 username=alice，实际 %v", resp["username"])
	}
}

func TestJWT_WrongSigningMethod_Returns401(t *testing.T) {
	router := setupJWTRouter(testSecret)

	// 使用 none 签名方法，应被拒绝
	claims := jwt.MapClaims{
		"user_id":  "user-123",
		"username": "hacker",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("生成 none token 失败: %v", err)
	}

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望 401（none 签名方法），实际 %d", w.Code)
	}
}

// ────────────────────────── GenerateToken + ValidateToken 测试 ──────────────────────────

func TestGenerateToken_Valid(t *testing.T) {
	tokenStr, err := GenerateToken("user-1", "bob", testSecret, 24)
	if err != nil {
		t.Fatalf("GenerateToken 失败: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("生成的 token 为空")
	}

	// 验证生成的 token 可以被正确解析
	userID, err := ValidateToken(tokenStr, testSecret)
	if err != nil {
		t.Fatalf("ValidateToken 失败: %v", err)
	}
	if userID != "user-1" {
		t.Errorf("期望 userID=user-1，实际 %s", userID)
	}
}

func TestValidateToken_Expired(t *testing.T) {
	tokenStr, err := GenerateToken("user-1", "bob", testSecret, 0)
	if err != nil {
		t.Fatalf("GenerateToken 失败: %v", err)
	}

	// expire 0 hours → token 已过期（exp = now + 0 = now, jwt 认为已过期）
	_, err = ValidateToken(tokenStr, testSecret)
	if err == nil {
		t.Error("期望过期 token 返回错误，但 err 为 nil")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	tokenStr, err := GenerateToken("user-1", "bob", testSecret, 24)
	if err != nil {
		t.Fatalf("GenerateToken 失败: %v", err)
	}

	_, err = ValidateToken(tokenStr, "different-secret")
	if err == nil {
		t.Error("期望错误 secret 返回错误，但 err 为 nil")
	}
}

func TestValidateToken_InvalidTokenString(t *testing.T) {
	_, err := ValidateToken("not-a-valid-jwt", testSecret)
	if err == nil {
		t.Error("期望无效 token 字符串返回错误，但 err 为 nil")
	}
}

// ────────────────────────── Rate Limit 中间件测试 ──────────────────────────

// setupRateLimitRouter 搭建含限流中间件的测试路由
func setupRateLimitRouter(rl *IPRateLimiter) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(rl.Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return r
}

func TestRateLimit_WithinBurst_AllPass(t *testing.T) {
	rl := NewIPRateLimiter(10, 5) // 10 req/s, burst 5
	router := setupRateLimitRouter(rl)

	// burst=5，5 个请求应该全部通过
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("第 %d 个请求期望 200，实际 %d", i+1, w.Code)
		}
	}
}

func TestRateLimit_ExceedBurst_Returns429(t *testing.T) {
	rl := NewIPRateLimiter(1, 3) // 1 req/s, burst 3
	router := setupRateLimitRouter(rl)

	// 前 3 个请求（burst=3）应通过
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("第 %d 个请求（burst 内）期望 200，实际 %d", i+1, w.Code)
		}
	}

	// 第 4 个请求应被限流
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("超出 burst 的请求期望 429，实际 %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if _, ok := resp["error"]; !ok {
		t.Error("429 响应中缺少 error 字段")
	}
}

func TestRateLimit_TokenRecovery_PassAgain(t *testing.T) {
	rl := NewIPRateLimiter(100, 2) // 100 req/s, burst 2 — 快速填充
	router := setupRateLimitRouter(rl)

	// 消耗完所有令牌（burst=2）
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "10.0.0.2:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("初始请求 %d 期望 200，实际 %d", i+1, w.Code)
		}
	}

	// 此时桶已空，下一个应被限流
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.2:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("桶空后期望 429，实际 %d", w.Code)
	}

	// 等待令牌恢复（100 req/s → 10ms 即可恢复 1 个令牌）
	time.Sleep(20 * time.Millisecond)

	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.2:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("令牌恢复后期望 200，实际 %d", w.Code)
	}
}

func TestRateLimit_DifferentIPs_Independent(t *testing.T) {
	rl := NewIPRateLimiter(1, 2) // 1 req/s, burst 2
	router := setupRateLimitRouter(rl)

	// IP-A 消耗完 burst
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "1.1.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("IP-A 第 %d 个请求期望 200，实际 %d", i+1, w.Code)
		}
	}

	// IP-A 被限流
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "1.1.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("IP-A 超限期望 429，实际 %d", w.Code)
	}

	// IP-B 不受影响，应该通过
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "2.2.2.2:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("IP-B 未超限期望 200，实际 %d", w.Code)
	}
}

// ────────────────────────── tokenBucket 单元测试 ──────────────────────────

func TestTokenBucket_Allow_ConsumeAndRefill(t *testing.T) {
	now := time.Now()
	b := &tokenBucket{
		tokens:     3.0,
		maxTokens:  3.0,
		refillRate: 10.0, // 10 tokens/sec
		lastRefill: now,
	}

	// 连续消费 3 个令牌
	for i := 0; i < 3; i++ {
		if !b.allow(now) {
			t.Errorf("第 %d 次 allow 应返回 true", i+1)
		}
	}

	// 第 4 次应返回 false（桶已空）
	if b.allow(now) {
		t.Error("桶空后 allow 应返回 false")
	}

	// 模拟 100ms 后（应补充 1 个令牌）
	future := now.Add(100 * time.Millisecond)
	if !b.allow(future) {
		t.Error("100ms 后补充 1 个令牌，allow 应返回 true")
	}

	// 再次消耗完
	if b.allow(future) {
		t.Error("再次消耗完后 allow 应返回 false")
	}
}

func TestTokenBucket_Refill_CappedAtMax(t *testing.T) {
	now := time.Now()
	b := &tokenBucket{
		tokens:     2.0,
		maxTokens:  5.0,
		refillRate: 100.0, // 100 tokens/sec
		lastRefill: now,
	}

	// 1 秒后理论上填充 100 个，但应被 cap 到 maxTokens=5
	future := now.Add(1 * time.Second)
	b.allow(future)

	if b.tokens > b.maxTokens {
		t.Errorf("令牌数 %.2f 超过上限 %.2f", b.tokens, b.maxTokens)
	}
	// allow 消费了 1 个，所以 tokens 应该 = maxTokens - 1 = 4.0
	expected := b.maxTokens - 1.0
	if b.tokens != expected {
		t.Errorf("期望 tokens=%.2f，实际 %.2f", expected, b.tokens)
	}
}

// ────────────────────────── cleanStale 测试 ──────────────────────────

func TestCleanStale_RemovesOldBuckets(t *testing.T) {
	rl := NewIPRateLimiter(10, 5)

	now := time.Now()

	// 手动添加一些桶
	rl.buckets["active-ip"] = &tokenBucket{
		tokens:     5.0,
		maxTokens:  5.0,
		refillRate: 10.0,
		lastRefill: now.Add(-2 * time.Minute), // 2 分钟前，不算旧
	}
	rl.buckets["stale-ip-1"] = &tokenBucket{
		tokens:     5.0,
		maxTokens:  5.0,
		refillRate: 10.0,
		lastRefill: now.Add(-15 * time.Minute), // 15 分钟前，应被清理
	}
	rl.buckets["stale-ip-2"] = &tokenBucket{
		tokens:     5.0,
		maxTokens:  5.0,
		refillRate: 10.0,
		lastRefill: now.Add(-20 * time.Minute), // 20 分钟前，应被清理
	}

	if len(rl.buckets) != 3 {
		t.Fatalf("初始应有 3 个桶，实际 %d", len(rl.buckets))
	}

	// 强制 lastClean 为很久之前，使 cleanStale 实际执行清理
	rl.lastClean = now.Add(-10 * time.Minute)

	rl.cleanStale(now)

	if _, ok := rl.buckets["active-ip"]; !ok {
		t.Error("active-ip 不应被清理")
	}
	if _, ok := rl.buckets["stale-ip-1"]; ok {
		t.Error("stale-ip-1 应该已被清理")
	}
	if _, ok := rl.buckets["stale-ip-2"]; ok {
		t.Error("stale-ip-2 应该已被清理")
	}

	if len(rl.buckets) != 1 {
		t.Errorf("清理后应剩余 1 个桶，实际 %d", len(rl.buckets))
	}
}

func TestCleanStale_SkipsWhenTooSoon(t *testing.T) {
	rl := NewIPRateLimiter(10, 5)
	now := time.Now()

	rl.buckets["some-ip"] = &tokenBucket{
		tokens:     5.0,
		maxTokens:  5.0,
		refillRate: 10.0,
		lastRefill: now.Add(-15 * time.Minute), // 很旧，但清理不会执行
	}

	// lastClean 刚更新过，cleanStale 应跳过
	rl.lastClean = now.Add(-1 * time.Minute)

	rl.cleanStale(now)

	if _, ok := rl.buckets["some-ip"]; !ok {
		t.Error("清理间隔未到，桶不应被清理")
	}
	if len(rl.buckets) != 1 {
		t.Errorf("应有 1 个桶，实际 %d", len(rl.buckets))
	}
}

func TestCleanStale_UpdatesLastClean(t *testing.T) {
	rl := NewIPRateLimiter(10, 5)
	now := time.Now()
	rl.lastClean = now.Add(-10 * time.Minute)

	rl.cleanStale(now)

	if !rl.lastClean.Equal(now) {
		t.Errorf("lastClean 应更新为 now，实际 %v", rl.lastClean)
	}
}

// ────────────────────────── NewIPRateLimiter 测试 ──────────────────────────

func TestNewIPRateLimiter_Defaults(t *testing.T) {
	rl := NewIPRateLimiter(5.0, 10)

	if rl.rate != 5.0 {
		t.Errorf("期望 rate=5.0，实际 %.2f", rl.rate)
	}
	if rl.burst != 10 {
		t.Errorf("期望 burst=10，实际 %d", rl.burst)
	}
	if rl.cleanup != 5*time.Minute {
		t.Errorf("期望 cleanup=5m，实际 %v", rl.cleanup)
	}
	if len(rl.buckets) != 0 {
		t.Errorf("初始 buckets 应为空，实际 %d", len(rl.buckets))
	}
}

func TestIPRateLimiter_Allow_CreatesBucketOnFirstCall(t *testing.T) {
	rl := NewIPRateLimiter(10, 5)

	result := rl.Allow("10.0.0.100")
	if !result {
		t.Error("首次 Allow 应返回 true")
	}

	if _, ok := rl.buckets["10.0.0.100"]; !ok {
		t.Error("首次 Allow 后应创建对应 IP 的桶")
	}
}

// ────────────────────────── 并发安全测试 ──────────────────────────

func TestRateLimit_ConcurrentAccess(t *testing.T) {
	rl := NewIPRateLimiter(1000, 100) // 高速率避免测试过慢
	router := setupRateLimitRouter(rl)

	done := make(chan bool, 20)

	for i := 0; i < 20; i++ {
		go func(idx int) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "concurrent-ip:12345"
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			// 不验证具体状态码，只确保不会 panic
			done <- true
		}(i)
	}

	for i := 0; i < 20; i++ {
		<-done
	}
}

// ────────────────────────── TestMain ──────────────────────────

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
