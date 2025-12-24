// Package sys
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
// @Description 安全服务
package sys

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
	"hotgo/internal/consts"
	"hotgo/internal/service"
	"hotgo/utility/encrypt"
)

func init() {
	service.RegisterSysSecurity(NewSysSecurity())
}

// sSysSecurity 安全服务
type sSysSecurity struct {
	cache *gcache.Cache
}

// NewSysSecurity 创建安全服务
func NewSysSecurity() *sSysSecurity {
	return &sSysSecurity{
		cache: gcache.New(),
	}
}

// CheckLoginAttempts 检查登录尝试次数
func (s *sSysSecurity) CheckLoginAttempts(ctx context.Context, identifier string) error {
	key := fmt.Sprintf("login_attempts:%s", identifier)
	
	val, err := s.cache.Get(ctx, key)
	if err != nil {
		return nil // 缓存错误不影响登录
	}
	
	attempts := val.Int()
	if attempts >= consts.MaxLoginAttempts {
		return gerror.Newf("登录失败次数过多，请%d分钟后再试", consts.LoginLockDuration/60)
	}
	
	return nil
}

// RecordLoginAttempt 记录登录尝试
func (s *sSysSecurity) RecordLoginAttempt(ctx context.Context, identifier string, success bool) error {
	key := fmt.Sprintf("login_attempts:%s", identifier)
	
	if success {
		// 登录成功，清除计数
		_, err := s.cache.Remove(ctx, key)
		return err
	}
	
	// 登录失败，增加计数
	val, _ := s.cache.Get(ctx, key)
	attempts := val.Int() + 1
	
	return s.cache.Set(ctx, key, attempts, time.Duration(consts.LoginLockDuration)*time.Second)
}

// EncryptSensitiveData 加密敏感数据
func (s *sSysSecurity) EncryptSensitiveData(ctx context.Context, data string) (string, error) {
	return encrypt.EncryptApiKey(data)
}

// DecryptSensitiveData 解密敏感数据
func (s *sSysSecurity) DecryptSensitiveData(ctx context.Context, encryptedData string) (string, error) {
	return encrypt.DecryptApiKey(encryptedData)
}

// ValidateIPWhitelist 验证IP白名单
func (s *sSysSecurity) ValidateIPWhitelist(ctx context.Context, ip string, whitelist []string) bool {
	if len(whitelist) == 0 {
		return true
	}
	
	for _, allowed := range whitelist {
		if allowed == "*" || allowed == ip {
			return true
		}
		// TODO: 支持CIDR格式的IP范围
	}
	
	return false
}

// LogSensitiveOperation 记录敏感操作
func (s *sSysSecurity) LogSensitiveOperation(ctx context.Context, userId int64, operation string, details string) error {
	g.Log().Infof(ctx, "Sensitive operation: user=%d, op=%s, details=%s", userId, operation, details)
	// TODO: 写入审计日志表
	return nil
}

// GenerateSecureToken 生成安全令牌
func (s *sSysSecurity) GenerateSecureToken(ctx context.Context, length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	token := make([]byte, length)
	
	for i := range token {
		token[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(time.Nanosecond)
	}
	
	return string(token), nil
}

// VerifyCSRFToken 验证CSRF令牌
func (s *sSysSecurity) VerifyCSRFToken(ctx context.Context, sessionId, token string) bool {
	key := fmt.Sprintf("csrf:%s", sessionId)
	
	val, err := s.cache.Get(ctx, key)
	if err != nil {
		return false
	}
	
	return val.String() == token
}

// GenerateCSRFToken 生成CSRF令牌
func (s *sSysSecurity) GenerateCSRFToken(ctx context.Context, sessionId string) (string, error) {
	token, err := s.GenerateSecureToken(ctx, consts.CSRFTokenLength)
	if err != nil {
		return "", err
	}
	
	key := fmt.Sprintf("csrf:%s", sessionId)
	err = s.cache.Set(ctx, key, token, time.Duration(consts.SessionTimeout)*time.Second)
	if err != nil {
		return "", err
	}
	
	return token, nil
}

// IsApiKeyExpired 检查API密钥是否过期
func (s *sSysSecurity) IsApiKeyExpired(ctx context.Context, createdAt time.Time) bool {
	maxAge := time.Duration(consts.ApiKeyMaxAge) * 24 * time.Hour
	return time.Since(createdAt) > maxAge
}

// MaskSensitiveString 遮蔽敏感字符串
func (s *sSysSecurity) MaskSensitiveString(data string, showFirst, showLast int) string {
	if len(data) <= showFirst+showLast {
		return "****"
	}
	
	masked := data[:showFirst]
	for i := 0; i < len(data)-showFirst-showLast; i++ {
		masked += "*"
	}
	masked += data[len(data)-showLast:]
	
	return masked
}

