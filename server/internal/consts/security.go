// Package consts
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
// @Description 安全相关常量
package consts

// 加密相关常量
const (
	// EncryptionKeyLength AES-256密钥长度
	EncryptionKeyLength = 32

	// DefaultEncryptionKey 默认加密密钥（生产环境必须修改）
	DefaultEncryptionKey = "toogo_secure_key_2024_change_me!"

	// ApiKeyPrefix API密钥前缀（用于识别加密后的密钥）
	ApiKeyPrefix = "ENC:"
)

// 请求限流相关常量
const (
	// RateLimitWindow 限流窗口时间（秒）
	RateLimitWindow = 60

	// RateLimitMaxRequests 限流窗口内最大请求数
	RateLimitMaxRequests = 100

	// RateLimitMaxRequestsForTrading 交易接口限流（更严格）
	RateLimitMaxRequestsForTrading = 30

	// RateLimitMaxRequestsForLogin 登录接口限流
	RateLimitMaxRequestsForLogin = 10

	// RateLimitCachePrefix 限流缓存前缀
	RateLimitCachePrefix = "rate_limit:"
)

// 安全相关常量
const (
	// MaxLoginAttempts 最大登录尝试次数
	MaxLoginAttempts = 5

	// LoginLockDuration 登录锁定时间（秒）
	LoginLockDuration = 300

	// ApiKeyMaxAge API密钥最大有效期（天）
	ApiKeyMaxAge = 90

	// SessionTimeout 会话超时时间（秒）
	SessionTimeout = 7200

	// CSRFTokenLength CSRF令牌长度
	CSRFTokenLength = 32
)

// 敏感操作类型
const (
	SensitiveOpWithdraw   = "withdraw"   // 提现
	SensitiveOpTransfer   = "transfer"   // 转账
	SensitiveOpApiKey     = "api_key"    // API密钥操作
	SensitiveOpPassword   = "password"   // 密码修改
	SensitiveOpRobotStart = "robot_start" // 机器人启动
)

// IP白名单相关
const (
	// IPWhitelistCacheKey IP白名单缓存键
	IPWhitelistCacheKey = "security:ip_whitelist"

	// IPBlacklistCacheKey IP黑名单缓存键
	IPBlacklistCacheKey = "security:ip_blacklist"
)

