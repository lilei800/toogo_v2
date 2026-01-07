// Package exchange API错误类型定义
package exchange

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// ErrorCode 交易所错误码
type ErrorCode int

const (
	// 通用错误码
	ErrCodeUnknown     ErrorCode = 0
	ErrCodeRateLimit   ErrorCode = -1015 // 请求过于频繁
	ErrCodeIPBanned    ErrorCode = -1003 // IP 被封禁
	ErrCodeServerBusy  ErrorCode = -1000 // 服务器繁忙
	ErrCodeInvalidAPI  ErrorCode = -2015 // API Key 无效
	ErrCodeNoPermission ErrorCode = -2010 // 没有权限

	// Binance 特定错误码
	ErrCodeBinanceTooManyRequests ErrorCode = -1015
	ErrCodeBinanceIPBanned        ErrorCode = -1003
	ErrCodeBinanceWAFLimit        ErrorCode = -1010 // WAF 限制
)

// APIError API错误
type APIError struct {
	Code       ErrorCode // 错误码
	Message    string    // 错误消息
	StatusCode int       // HTTP状态码
	Platform   string    // 交易所平台
	RawBody    string    // 原始响应
}

func (e *APIError) Error() string {
	return e.Message
}

// IsRateLimitError 判断是否为限流错误
func (e *APIError) IsRateLimitError() bool {
	// HTTP 429 = Too Many Requests
	if e.StatusCode == 429 {
		return true
	}

	// 检查常见限流错误码
	// 注意：ErrCodeRateLimit 和 ErrCodeBinanceTooManyRequests 都是 -1015
	switch e.Code {
	case ErrCodeRateLimit:
		return true
	}

	// 检查错误消息中的关键字
	lowerMsg := strings.ToLower(e.Message)
	if strings.Contains(lowerMsg, "rate limit") ||
		strings.Contains(lowerMsg, "too many") ||
		strings.Contains(lowerMsg, "request limit") ||
		strings.Contains(lowerMsg, "请求过于频繁") {
		return true
	}

	return false
}

// IsIPBannedError 判断是否为IP封禁错误
func (e *APIError) IsIPBannedError() bool {
	// HTTP 403 + 特定错误码
	if e.StatusCode == 403 {
		return true
	}

	// 注意：ErrCodeIPBanned 和 ErrCodeBinanceIPBanned 都是 -1003
	switch e.Code {
	case ErrCodeIPBanned, ErrCodeBinanceWAFLimit:
		return true
	}

	lowerMsg := strings.ToLower(e.Message)
	if strings.Contains(lowerMsg, "ip ban") ||
		strings.Contains(lowerMsg, "ip blocked") ||
		strings.Contains(lowerMsg, "waf") ||
		strings.Contains(lowerMsg, "forbidden") {
		return true
	}

	return false
}

// IsAuthError 判断是否为认证错误
func (e *APIError) IsAuthError() bool {
	switch e.Code {
	case ErrCodeInvalidAPI, ErrCodeNoPermission:
		return true
	}

	if e.StatusCode == 401 {
		return true
	}

	lowerMsg := strings.ToLower(e.Message)
	return strings.Contains(lowerMsg, "api key") ||
		strings.Contains(lowerMsg, "signature") ||
		strings.Contains(lowerMsg, "unauthorized")
}

// IsCriticalError 判断是否为严重错误（需要停止机器人）
func (e *APIError) IsCriticalError() bool {
	return e.IsIPBannedError() || e.IsAuthError()
}

// ParseAPIError 解析API错误响应
func ParseAPIError(platform string, statusCode int, body string) *APIError {
	apiErr := &APIError{
		StatusCode: statusCode,
		Platform:   platform,
		RawBody:    body,
		Message:    body,
	}

	// 尝试解析错误码
	// Binance 格式: {"code":-1015,"msg":"Too many requests..."}
	codePattern := regexp.MustCompile(`"code"\s*:\s*(-?\d+|"(\d+)")`)
	msgPattern := regexp.MustCompile(`"msg"\s*:\s*"([^"]*)"`)
	// Gate 常见格式: {"label":"USER_NOT_FOUND","message":"..."}
	labelPattern := regexp.MustCompile(`"label"\s*:\s*"([^"]*)"`)
	messagePattern := regexp.MustCompile(`"message"\s*:\s*"([^"]*)"`)

	if matches := codePattern.FindStringSubmatch(body); len(matches) > 1 {
		codeStr := matches[1]
		// 去除引号
		codeStr = strings.Trim(codeStr, "\"")
		if code, err := strconv.Atoi(codeStr); err == nil {
			apiErr.Code = ErrorCode(code)
		}
	}

	if matches := msgPattern.FindStringSubmatch(body); len(matches) > 1 {
		apiErr.Message = matches[1]
	}
	// 兼容 Gate 的 message 字段（只有当 msg 未命中时再尝试，避免覆盖 Binance/Bitget）
	if apiErr.Message == body {
		if matches := messagePattern.FindStringSubmatch(body); len(matches) > 1 {
			apiErr.Message = matches[1]
		}
	}
	// label 作为前缀，帮助定位错误类型
	if matches := labelPattern.FindStringSubmatch(body); len(matches) > 1 {
		lbl := strings.TrimSpace(matches[1])
		if lbl != "" {
			// 避免重复拼接
			if !strings.Contains(strings.ToUpper(apiErr.Message), strings.ToUpper(lbl)) {
				apiErr.Message = lbl + ": " + apiErr.Message
			}
		}
	}

	return apiErr
}

// WrapAsAPIError 将普通错误包装为API错误
func WrapAsAPIError(platform string, statusCode int, body string, originalErr error) error {
	apiErr := ParseAPIError(platform, statusCode, body)

	// 判断是否为限流或IP封禁错误
	if apiErr.IsRateLimitError() {
		return gerror.Wrapf(apiErr, "[%s] API限流: %s", platform, apiErr.Message)
	}

	if apiErr.IsIPBannedError() {
		return gerror.Wrapf(apiErr, "[%s] IP被封禁: %s", platform, apiErr.Message)
	}

	if apiErr.IsAuthError() {
		return gerror.Wrapf(apiErr, "[%s] API认证失败: %s", platform, apiErr.Message)
	}

	return gerror.Newf("[%s] API error (code=%d): %s", platform, apiErr.Code, apiErr.Message)
}

// IsRateLimitErr 判断错误是否为限流错误（通用方法）
func IsRateLimitErr(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否为 APIError 类型
	var apiErr *APIError
	if gerror.As(err, &apiErr) {
		return apiErr.IsRateLimitError()
	}

	// 检查错误消息
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "too many") ||
		strings.Contains(errStr, "限流") ||
		strings.Contains(errStr, "-1015")
}

// IsIPBannedErr 判断错误是否为IP封禁错误（通用方法）
func IsIPBannedErr(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *APIError
	if gerror.As(err, &apiErr) {
		return apiErr.IsIPBannedError()
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "ip ban") ||
		strings.Contains(errStr, "forbidden") ||
		strings.Contains(errStr, "-1003")
}

// IsCriticalErr 判断错误是否为严重错误（需要停止机器人）
func IsCriticalErr(err error) bool {
	if err == nil {
		return false
	}

	var apiErr *APIError
	if gerror.As(err, &apiErr) {
		return apiErr.IsCriticalError()
	}

	return IsIPBannedErr(err)
}

