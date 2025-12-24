// Package exchange API请求重试机制
package exchange

import (
	"context"
	"math"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries int           // 最大重试次数
	BaseDelay  time.Duration // 基础延迟
	MaxDelay   time.Duration // 最大延迟
	Multiplier float64       // 延迟倍数
}

// DefaultRetryConfig 默认重试配置
var DefaultRetryConfig = RetryConfig{
	MaxRetries: 3,
	BaseDelay:  100 * time.Millisecond,
	MaxDelay:   5 * time.Second,
	Multiplier: 2.0,
}

// RetryableError 可重试的错误类型
type RetryableError struct {
	Err       error
	Retryable bool
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// IsRetryable 判断错误是否可重试
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否为RetryableError
	var retryableErr *RetryableError
	if gerror.As(err, &retryableErr) {
		return retryableErr.Retryable
	}

	// 默认对网络错误等进行重试
	errStr := err.Error()
	retryableKeywords := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"no such host",
		"temporary failure",
		"network is unreachable",
		"i/o timeout",
		"EOF",
	}

	for _, keyword := range retryableKeywords {
		if containsIgnoreCase(errStr, keyword) {
			return true
		}
	}

	return false
}

// containsIgnoreCase 忽略大小写检查字符串包含
func containsIgnoreCase(s, substr string) bool {
	sLower := make([]byte, len(s))
	substrLower := make([]byte, len(substr))
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			sLower[i] = s[i] + 32
		} else {
			sLower[i] = s[i]
		}
	}
	for i := 0; i < len(substr); i++ {
		if substr[i] >= 'A' && substr[i] <= 'Z' {
			substrLower[i] = substr[i] + 32
		} else {
			substrLower[i] = substr[i]
		}
	}
	return bytesContains(sLower, substrLower)
}

func bytesContains(s, substr []byte) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// WithRetry 带重试的函数执行
func WithRetry(ctx context.Context, fn func() error, config *RetryConfig) error {
	if config == nil {
		config = &DefaultRetryConfig
	}

	var lastErr error
	for i := 0; i <= config.MaxRetries; i++ {
		// 执行函数
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// 检查是否为最后一次尝试
		if i == config.MaxRetries {
			break
		}

		// 检查是否可重试
		if !IsRetryable(err) {
			return err
		}

		// 检查context是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 计算延迟时间（指数退避）
		delay := time.Duration(float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(i)))
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		g.Log().Debugf(ctx, "[Retry] 第%d次重试失败: %v, 等待%v后重试", i+1, err, delay)

		// 等待
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	return gerror.Wrapf(lastErr, "failed after %d retries", config.MaxRetries+1)
}

// WithRetryResult 带重试的函数执行（返回结果）
func WithRetryResult[T any](ctx context.Context, fn func() (T, error), config *RetryConfig) (T, error) {
	if config == nil {
		config = &DefaultRetryConfig
	}

	var result T
	var lastErr error

	for i := 0; i <= config.MaxRetries; i++ {
		// 执行函数
		var err error
		result, err = fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// 检查是否为最后一次尝试
		if i == config.MaxRetries {
			break
		}

		// 检查是否可重试
		if !IsRetryable(err) {
			return result, err
		}

		// 检查context是否已取消
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		// 计算延迟时间
		delay := time.Duration(float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(i)))
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		g.Log().Debugf(ctx, "[Retry] 第%d次重试失败: %v, 等待%v后重试", i+1, err, delay)

		// 等待
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return result, ctx.Err()
		case <-timer.C:
		}
	}

	return result, gerror.Wrapf(lastErr, "failed after %d retries", config.MaxRetries+1)
}

// ExchangeRequest 交易所请求封装（带限流和重试）
func ExchangeRequest[T any](ctx context.Context, platform string, fn func() (T, error)) (T, error) {
	// 限流
	if err := WaitForRateLimit(ctx, platform); err != nil {
		var zero T
		return zero, gerror.Wrap(err, "rate limit error")
	}

	// 带重试执行
	return WithRetryResult(ctx, fn, &DefaultRetryConfig)
}

// ExchangeRequestNoResult 交易所请求封装（无返回值）
func ExchangeRequestNoResult(ctx context.Context, platform string, fn func() error) error {
	// 限流
	if err := WaitForRateLimit(ctx, platform); err != nil {
		return gerror.Wrap(err, "rate limit error")
	}

	// 带重试执行
	return WithRetry(ctx, fn, &DefaultRetryConfig)
}

