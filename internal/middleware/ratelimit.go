package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// tokenBucket 单个 IP 的令牌桶
type tokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens/sec
	lastRefill time.Time
}

// allow 尝试消费一个令牌，返回是否允许
func (b *tokenBucket) allow(now time.Time) bool {
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.lastRefill = now

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true
	}
	return false
}

// IPRateLimiter IP 级令牌桶限流器
type IPRateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*tokenBucket
	rate     float64       // 每秒填充的令牌数
	burst    int           // 桶容量（突发上限）
	cleanup  time.Duration // 过期桶清理间隔
	lastClean time.Time
}

// NewIPRateLimiter 创建 IP 限流器
//   - rate: 每秒允许的请求数（令牌填充速率）
//   - burst: 突发请求上限（桶容量）
func NewIPRateLimiter(rate float64, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		buckets:   make(map[string]*tokenBucket),
		rate:      rate,
		burst:     burst,
		cleanup:   5 * time.Minute,
		lastClean: time.Now(),
	}
}

// getBucket 获取或创建 IP 对应的令牌桶
func (rl *IPRateLimiter) getBucket(ip string) *tokenBucket {
	b, ok := rl.buckets[ip]
	if !ok {
		b = &tokenBucket{
			tokens:     float64(rl.burst),
			maxTokens:  float64(rl.burst),
			refillRate: rl.rate,
			lastRefill: time.Now(),
		}
		rl.buckets[ip] = b
	}
	return b
}

// cleanStale 清理长时间不活跃的桶，防止内存泄漏
func (rl *IPRateLimiter) cleanStale(now time.Time) {
	if now.Sub(rl.lastClean) < rl.cleanup {
		return
	}
	rl.lastClean = now
	threshold := now.Add(-10 * time.Minute)
	for ip, b := range rl.buckets {
		if b.lastRefill.Before(threshold) {
			delete(rl.buckets, ip)
		}
	}
}

// Allow 检查指定 IP 是否允许请求
func (rl *IPRateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	rl.cleanStale(now)
	return rl.getBucket(ip).allow(now)
}

// Middleware 返回 Gin 限流中间件
// 返回 429 Too Many Requests 当请求超过限速
func (rl *IPRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
