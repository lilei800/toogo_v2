// Package middleware
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
// @Description 请求限流中间件
package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"hotgo/internal/consts"
	"hotgo/internal/library/contexts"
	"hotgo/internal/library/response"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Window      int  // 时间窗口（秒）
	MaxRequests int  // 最大请求数
	ByIP        bool // 是否按IP限流
	ByUser      bool // 是否按用户限流
}

// DefaultRateLimitConfig 默认限流配置
var DefaultRateLimitConfig = &RateLimitConfig{
	Window:      consts.RateLimitWindow,
	MaxRequests: consts.RateLimitMaxRequests,
	ByIP:        true,
	ByUser:      false,
}

// TradingRateLimitConfig 交易接口限流配置
var TradingRateLimitConfig = &RateLimitConfig{
	Window:      consts.RateLimitWindow,
	MaxRequests: consts.RateLimitMaxRequestsForTrading,
	ByIP:        true,
	ByUser:      true,
}

// LoginRateLimitConfig 登录接口限流配置
var LoginRateLimitConfig = &RateLimitConfig{
	Window:      consts.RateLimitWindow,
	MaxRequests: consts.RateLimitMaxRequestsForLogin,
	ByIP:        true,
	ByUser:      false,
}

// RateLimit 通用限流中间件
func (s *sMiddleware) RateLimit(r *ghttp.Request) {
	if err := checkRateLimit(r.Context(), r, DefaultRateLimitConfig); err != nil {
		response.JsonExit(r, gcode.CodeOperationFailed.Code(), err.Error())
	}
	r.Middleware.Next()
}

// RateLimitForTrading 交易接口限流中间件
func (s *sMiddleware) RateLimitForTrading(r *ghttp.Request) {
	if err := checkRateLimit(r.Context(), r, TradingRateLimitConfig); err != nil {
		response.JsonExit(r, gcode.CodeOperationFailed.Code(), err.Error())
	}
	r.Middleware.Next()
}

// RateLimitForLogin 登录接口限流中间件
func (s *sMiddleware) RateLimitForLogin(r *ghttp.Request) {
	if err := checkRateLimit(r.Context(), r, LoginRateLimitConfig); err != nil {
		response.JsonExit(r, gcode.CodeOperationFailed.Code(), err.Error())
	}
	r.Middleware.Next()
}

// checkRateLimit 检查限流
func checkRateLimit(ctx context.Context, r *ghttp.Request, config *RateLimitConfig) error {
	// 构建限流键
	key := buildRateLimitKey(r, config)
	
	// 获取Redis实例
	redis := g.Redis()
	if redis == nil {
		// Redis不可用时跳过限流
		g.Log().Warning(ctx, "Redis unavailable, rate limit skipped")
		return nil
	}

	// 获取当前计数
	countVal, err := redis.Get(ctx, key)
	if err != nil {
		g.Log().Warningf(ctx, "Rate limit get error: %v", err)
		return nil
	}

	count := countVal.Int()
	
	// 检查是否超过限制
	if count >= config.MaxRequests {
		g.Log().Warningf(ctx, "Rate limit exceeded: key=%s, count=%d, limit=%d", key, count, config.MaxRequests)
		return gerror.New("请求过于频繁，请稍后再试")
	}

	// 增加计数
	_, err = redis.Incr(ctx, key)
	if err != nil {
		g.Log().Warningf(ctx, "Rate limit incr error: %v", err)
		return nil
	}

	// 如果是新建的键，设置过期时间
	if count == 0 {
		_, err = redis.Expire(ctx, key, int64(config.Window))
		if err != nil {
			g.Log().Warningf(ctx, "Rate limit expire error: %v", err)
		}
	}

	// 设置响应头
	r.Response.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.MaxRequests))
	r.Response.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", config.MaxRequests-count-1))
	r.Response.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Duration(config.Window)*time.Second).Unix()))

	return nil
}

// buildRateLimitKey 构建限流键
func buildRateLimitKey(r *ghttp.Request, config *RateLimitConfig) string {
	var key string

	// 基础前缀 + 路径
	path := r.URL.Path
	key = consts.RateLimitCachePrefix + path

	// 按IP限流
	if config.ByIP {
		ip := r.GetClientIp()
		key = key + ":ip:" + ip
	}

	// 按用户限流
	if config.ByUser {
		ctx := r.Context()
		model := contexts.Get(ctx)
		if model != nil && model.User != nil {
			key = key + ":user:" + fmt.Sprintf("%d", model.User.Id)
		}
	}

	return key
}

// ClearRateLimit 清除限流计数（用于特殊情况）
func ClearRateLimit(ctx context.Context, key string) error {
	redis := g.Redis()
	if redis == nil {
		return nil
	}
	_, err := redis.Del(ctx, consts.RateLimitCachePrefix+key)
	return err
}

// GetRateLimitStatus 获取限流状态
func GetRateLimitStatus(ctx context.Context, key string, config *RateLimitConfig) (count int, remaining int, err error) {
	redis := g.Redis()
	if redis == nil {
		return 0, config.MaxRequests, nil
	}

	fullKey := consts.RateLimitCachePrefix + key
	countVal, err := redis.Get(ctx, fullKey)
	if err != nil {
		return 0, config.MaxRequests, err
	}

	count = countVal.Int()
	remaining = config.MaxRequests - count
	if remaining < 0 {
		remaining = 0
	}

	return count, remaining, nil
}

