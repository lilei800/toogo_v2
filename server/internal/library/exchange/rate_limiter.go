// Package exchange API请求限流器
package exchange

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

// RateLimiter 请求限流器
type RateLimiter struct {
	mu sync.Mutex

	// 配置
	requestsPerSecond float64 // 每秒请求数
	burst             int     // 突发请求数

	// 状态
	tokens     float64   // 当前令牌数
	lastUpdate time.Time // 上次更新时间
}

// NewRateLimiter 创建限流器
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	if rps <= 0 {
		rps = 10 // 默认每秒10次
	}
	if burst <= 0 {
		burst = 20 // 默认突发20次
	}

	return &RateLimiter{
		requestsPerSecond: rps,
		burst:             burst,
		tokens:            float64(burst),
		lastUpdate:        time.Now(),
	}
}

// Wait 等待获取令牌
// 如果超过限制，会阻塞等待
func (r *RateLimiter) Wait(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 更新令牌数
	now := time.Now()
	elapsed := now.Sub(r.lastUpdate).Seconds()
	r.tokens += elapsed * r.requestsPerSecond
	if r.tokens > float64(r.burst) {
		r.tokens = float64(r.burst)
	}
	r.lastUpdate = now

	// 检查是否有可用令牌
	if r.tokens >= 1 {
		r.tokens--
		return nil
	}

	// 计算需要等待的时间
	waitTime := time.Duration((1-r.tokens)/r.requestsPerSecond*1000) * time.Millisecond

	// 检查context是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 限制最大等待时间
	maxWait := 5 * time.Second
	if waitTime > maxWait {
		return gerror.Newf("rate limit exceeded, would need to wait %v", waitTime)
	}

	// 等待
	timer := time.NewTimer(waitTime)
	select {
	case <-ctx.Done():
		timer.Stop()
		return ctx.Err()
	case <-timer.C:
		r.tokens = 0 // 消耗令牌
		return nil
	}
}

// TryAcquire 尝试获取令牌（非阻塞）
// 返回是否成功获取
func (r *RateLimiter) TryAcquire() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 更新令牌数
	now := time.Now()
	elapsed := now.Sub(r.lastUpdate).Seconds()
	r.tokens += elapsed * r.requestsPerSecond
	if r.tokens > float64(r.burst) {
		r.tokens = float64(r.burst)
	}
	r.lastUpdate = now

	if r.tokens >= 1 {
		r.tokens--
		return true
	}
	return false
}

// Available 获取当前可用令牌数
func (r *RateLimiter) Available() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 更新令牌数
	now := time.Now()
	elapsed := now.Sub(r.lastUpdate).Seconds()
	r.tokens += elapsed * r.requestsPerSecond
	if r.tokens > float64(r.burst) {
		r.tokens = float64(r.burst)
	}
	r.lastUpdate = now

	return int(r.tokens)
}

// ExchangeRateLimiters 交易所限流器管理
var (
	exchangeLimiters     = make(map[string]*RateLimiter)
	exchangeLimitersMu   sync.RWMutex
	defaultLimiterConfig = struct {
		RPS   float64
		Burst int
	}{
		RPS:   10, // 默认每秒10次
		Burst: 20, // 默认突发20次
	}
)

// 各交易所限流配置
var exchangeLimiterConfigs = map[string]struct {
	RPS   float64
	Burst int
}{
	"binance": {RPS: 10, Burst: 50},
	"okx":     {RPS: 10, Burst: 30},
	"gate":    {RPS: 10, Burst: 30},
}

// GetExchangeLimiter 获取指定交易所的限流器
func GetExchangeLimiter(platform string) *RateLimiter {
	exchangeLimitersMu.RLock()
	limiter, exists := exchangeLimiters[platform]
	exchangeLimitersMu.RUnlock()

	if exists {
		return limiter
	}

	// 创建新的限流器
	exchangeLimitersMu.Lock()
	defer exchangeLimitersMu.Unlock()

	// 双重检查
	if limiter, exists = exchangeLimiters[platform]; exists {
		return limiter
	}

	// 获取配置
	config, ok := exchangeLimiterConfigs[platform]
	if !ok {
		config = struct {
			RPS   float64
			Burst int
		}{
			RPS:   defaultLimiterConfig.RPS,
			Burst: defaultLimiterConfig.Burst,
		}
	}

	limiter = NewRateLimiter(config.RPS, config.Burst)
	exchangeLimiters[platform] = limiter
	return limiter
}

// WaitForRateLimit 等待限流
func WaitForRateLimit(ctx context.Context, platform string) error {
	return GetExchangeLimiter(platform).Wait(ctx)
}

