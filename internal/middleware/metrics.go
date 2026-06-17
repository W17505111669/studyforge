package middleware

import (
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ==================== API 性能监控中间件 ====================

// MetricEntry 单条请求性能记录
type MetricEntry struct {
	Path       string    `json:"path"`
	Method     string    `json:"method"`
	StatusCode int       `json:"status_code"`
	LatencyMs  float64   `json:"latency_ms"`
	UserID     string    `json:"user_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// MetricsStore 内存环形缓冲区（保留最近 N 条请求记录）
type MetricsStore struct {
	mu      sync.RWMutex
	entries []MetricEntry
	head    int // 下一个写入位置
	count   int // 当前有效记录数
	size    int // 缓冲区容量
}

// GlobalMetrics 全局性能指标存储（包级别单例）
var GlobalMetrics = NewMetricsStore(2000)

// NewMetricsStore 创建环形缓冲区
func NewMetricsStore(size int) *MetricsStore {
	return &MetricsStore{
		entries: make([]MetricEntry, size),
		size:    size,
	}
}

// Record 写入一条指标记录
func (s *MetricsStore) Record(entry MetricEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[s.head] = entry
	s.head = (s.head + 1) % s.size
	if s.count < s.size {
		s.count++
	}
}

// Snapshot 获取当前所有有效记录的快照（线程安全）
func (s *MetricsStore) Snapshot() []MetricEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.count == 0 {
		return nil
	}

	result := make([]MetricEntry, s.count)
	if s.count < s.size {
		// 还没写满一圈，直接从头开始拷贝
		copy(result, s.entries[:s.count])
	} else {
		// 已经写满一圈，从 head 开始读（head 是最老的记录）
		n := copy(result, s.entries[s.head:])
		copy(result[n:], s.entries[:s.head])
	}
	return result
}

// MetricsMiddleware API 性能监控中间件
// 记录每个 API 请求的处理时间，存入环形缓冲区
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		userID, _ := c.Get("userID")
		uid, _ := userID.(string)

		GlobalMetrics.Record(MetricEntry{
			Path:       path,
			Method:     c.Request.Method,
			StatusCode: c.Writer.Status(),
			LatencyMs:  float64(latency.Nanoseconds()) / 1e6,
			UserID:     uid,
			Timestamp:  start,
		})
	}
}

// ==================== 统计分析 ====================

// EndpointStat 端点维度的聚合统计
type EndpointStat struct {
	Path      string  `json:"path"`
	Method    string  `json:"method"`
	Count     int     `json:"count"`
	AvgMs     float64 `json:"avg_ms"`
	P50Ms     float64 `json:"p50_ms"`
	P95Ms     float64 `json:"p95_ms"`
	P99Ms     float64 `json:"p99_ms"`
	MaxMs     float64 `json:"max_ms"`
	ErrorRate float64 `json:"error_rate"` // 4xx+5xx 占比
}

// TimeBucket 时间桶（用于时间序列折线图）
type TimeBucket struct {
	Time    string  `json:"time"`
	AvgMs   float64 `json:"avg_ms"`
	P95Ms   float64 `json:"p95_ms"`
	Count   int     `json:"count"`
	ErrRate float64 `json:"err_rate"`
}

// HeatCell 热力图单元格（小时×天）
type HeatCell struct {
	Day   int     `json:"day"`   // 0=周一 ... 6=周日
	Hour  int     `json:"hour"`  // 0-23
	Value int     `json:"value"` // 请求数量
}

// StatusCount HTTP 状态码分布
type StatusCount struct {
	Status int `json:"status"`
	Count  int `json:"count"`
}

// MetricsSummary 汇总统计响应
type MetricsSummary struct {
	// 总体指标
	TotalRequests int     `json:"total_requests"`
	AvgMs         float64 `json:"avg_ms"`
	P50Ms         float64 `json:"p50_ms"`
	P95Ms         float64 `json:"p95_ms"`
	P99Ms         float64 `json:"p99_ms"`
	ErrorRate     float64 `json:"error_rate"`
	QPS           float64 `json:"qps"`

	// 分维度
	Endpoints    []EndpointStat `json:"endpoints"`
	TimeSeries   []TimeBucket   `json:"time_series"`
	Heatmap      []HeatCell     `json:"heatmap"`
	StatusDist   []StatusCount  `json:"status_dist"`

	// 元数据
	RangeLabel string `json:"range_label"`
	RecordedAt int64  `json:"recorded_at"` // Unix 秒
}

// ComputeSummary 根据时间范围过滤并计算汇总统计
func ComputeSummary(rangeHours int) MetricsSummary {
	all := GlobalMetrics.Snapshot()
	if len(all) == 0 {
		return MetricsSummary{
			RangeLabel: rangeLabel(rangeHours),
			RecordedAt: time.Now().Unix(),
		}
	}

	// 按时间范围过滤
	cutoff := time.Now().Add(-time.Duration(rangeHours) * time.Hour)
	var filtered []MetricEntry
	for _, e := range all {
		if e.Timestamp.After(cutoff) {
			filtered = append(filtered, e)
		}
	}

	if len(filtered) == 0 {
		return MetricsSummary{
			RangeLabel: rangeLabel(rangeHours),
			RecordedAt: time.Now().Unix(),
		}
	}

	// === 总体延迟 ===
	latencies := make([]float64, len(filtered))
	var totalLat float64
	var errorCount int
	for i, e := range filtered {
		latencies[i] = e.LatencyMs
		totalLat += e.LatencyMs
		if e.StatusCode >= 400 {
			errorCount++
		}
	}
	sort.Float64s(latencies)

	// 计算时间跨度（秒），用于 QPS
	earliest := filtered[0].Timestamp
	latest := filtered[len(filtered)-1].Timestamp
	spanSec := latest.Sub(earliest).Seconds()
	if spanSec < 1 {
		spanSec = 1
	}

	summary := MetricsSummary{
		TotalRequests: len(filtered),
		AvgMs:         round2(totalLat / float64(len(filtered))),
		P50Ms:         round2(percentile(latencies, 50)),
		P95Ms:         round2(percentile(latencies, 95)),
		P99Ms:         round2(percentile(latencies, 99)),
		ErrorRate:     round2(float64(errorCount) / float64(len(filtered)) * 100),
		QPS:           round2(float64(len(filtered)) / spanSec),
		RangeLabel:    rangeLabel(rangeHours),
		RecordedAt:    time.Now().Unix(),
	}

	// === 端点维度 ===
	epMap := make(map[string][]MetricEntry)
	for _, e := range filtered {
		key := e.Method + " " + e.Path
		epMap[key] = append(epMap[key], e)
	}

	endpoints := make([]EndpointStat, 0, len(epMap))
	for key, entries := range epMap {
		els := make([]float64, len(entries))
		var sum float64
		var errs int
		for i, e := range entries {
			els[i] = e.LatencyMs
			sum += e.LatencyMs
			if e.StatusCode >= 400 {
				errs++
			}
		}
		sort.Float64s(els)
		endpoints = append(endpoints, EndpointStat{
			Path:      entries[0].Path,
			Method:    entries[0].Method,
			Count:     len(entries),
			AvgMs:     round2(sum / float64(len(entries))),
			P50Ms:     round2(percentile(els, 50)),
			P95Ms:     round2(percentile(els, 95)),
			P99Ms:     round2(percentile(els, 99)),
			MaxMs:     round2(els[len(els)-1]),
			ErrorRate: round2(float64(errs) / float64(len(entries)) * 100),
		})
		_ = key
	}
	// 按平均延迟降序排列
	sort.Slice(endpoints, func(i, j int) bool {
		return endpoints[i].AvgMs > endpoints[j].AvgMs
	})
	summary.Endpoints = endpoints

	// === 时间序列（按时间分桶）===
	bucketMinutes := 5
	if rangeHours > 24 {
		bucketMinutes = 60
	} else if rangeHours > 4 {
		bucketMinutes = 15
	}

	type tBucket struct {
		entries []MetricEntry
	}
	bucketMap := make(map[string]*tBucket)
	for _, e := range filtered {
		t := e.Timestamp.Truncate(time.Duration(bucketMinutes) * time.Minute)
		key := t.Format("15:04")
		if rangeHours > 24 {
			key = t.Format("01/02 15:04")
		}
		b, ok := bucketMap[key]
		if !ok {
			b = &tBucket{}
			bucketMap[key] = b
		}
		b.entries = append(b.entries, e)
	}

	// 排序时间桶
	timeKeys := make([]string, 0, len(bucketMap))
	for k := range bucketMap {
		timeKeys = append(timeKeys, k)
	}
	sort.Strings(timeKeys)

	timeSeries := make([]TimeBucket, 0, len(timeKeys))
	for _, key := range timeKeys {
		b := bucketMap[key]
		els := make([]float64, len(b.entries))
		var sum float64
		var errs int
		for i, e := range b.entries {
			els[i] = e.LatencyMs
			sum += e.LatencyMs
			if e.StatusCode >= 400 {
				errs++
			}
		}
		sort.Float64s(els)
		timeSeries = append(timeSeries, TimeBucket{
			Time:    key,
			AvgMs:   round2(sum / float64(len(b.entries))),
			P95Ms:   round2(percentile(els, 95)),
			Count:   len(b.entries),
			ErrRate: round2(float64(errs) / float64(len(b.entries)) * 100),
		})
	}
	summary.TimeSeries = timeSeries

	// === 热力图（小时×天）===
	heatMap := make(map[[2]int]int) // [day, hour] -> count
	for _, e := range filtered {
		weekday := int(e.Timestamp.Weekday())
		// 转换为 0=周一 ... 6=周日
		day := (weekday + 6) % 7
		hour := e.Timestamp.Hour()
		heatMap[[2]int{day, hour}]++
	}

	heatmap := make([]HeatCell, 0, len(heatMap))
	for k, v := range heatMap {
		heatmap = append(heatmap, HeatCell{
			Day:   k[0],
			Hour:  k[1],
			Value: v,
		})
	}
	summary.Heatmap = heatmap

	// === 状态码分布 ===
	statusMap := make(map[int]int)
	for _, e := range filtered {
		statusMap[e.StatusCode]++
	}
	statusDist := make([]StatusCount, 0, len(statusMap))
	for s, c := range statusMap {
		statusDist = append(statusDist, StatusCount{Status: s, Count: c})
	}
	sort.Slice(statusDist, func(i, j int) bool {
		return statusDist[i].Count > statusDist[j].Count
	})
	summary.StatusDist = statusDist

	return summary
}

// ==================== 辅助函数 ====================

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p / 100)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func round2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}

func rangeLabel(hours int) string {
	switch hours {
	case 1:
		return "最近 1 小时"
	case 24:
		return "最近 24 小时"
	case 168:
		return "最近 7 天"
	default:
		return "自定义范围"
	}
}
