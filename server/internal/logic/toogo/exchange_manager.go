// Package toogo 交易所管理器
package toogo

import (
	"context"
	"strings"
	"sync"

	"hotgo/internal/dao"
	"hotgo/internal/library/exchange"
	"hotgo/internal/model/entity"
	"hotgo/utility/encrypt"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// ExchangeManager 交易所管理器
type ExchangeManager struct {
	mu        sync.RWMutex
	exchanges map[int64]exchange.Exchange // key: apiConfigId
}

var (
	exchangeManager     *ExchangeManager
	exchangeManagerOnce sync.Once
)

// GetExchangeManager 获取交易所管理器单例
func GetExchangeManager() *ExchangeManager {
	exchangeManagerOnce.Do(func() {
		exchangeManager = &ExchangeManager{
			exchanges: make(map[int64]exchange.Exchange),
		}
	})
	return exchangeManager
}

// GetExchangeFromConfig 从API配置获取或创建交易所实例
func (m *ExchangeManager) GetExchangeFromConfig(ctx context.Context, apiConfig *entity.TradingApiConfig) (exchange.Exchange, error) {
	if apiConfig == nil {
		return nil, gerror.New("API配置不能为空")
	}

	m.mu.RLock()
	ex, exists := m.exchanges[apiConfig.Id]
	m.mu.RUnlock()

	if exists {
		return ex, nil
	}

	// 创建新实例
	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查
	if ex, exists = m.exchanges[apiConfig.Id]; exists {
		return ex, nil
	}

	// 解密 API Key（数据库中是加密存储的）
	apiKey, err := encrypt.AesDecrypt(apiConfig.ApiKey)
	if err != nil {
		g.Log().Warningf(ctx, "[ExchangeManager] 解密ApiKey失败: %v, 尝试使用原始值", err)
		apiKey = apiConfig.ApiKey // 兼容未加密的旧数据
	}

	// 解密 Secret Key
	secretKey, err := encrypt.AesDecrypt(apiConfig.SecretKey)
	if err != nil {
		g.Log().Warningf(ctx, "[ExchangeManager] 解密SecretKey失败: %v, 尝试使用原始值", err)
		secretKey = apiConfig.SecretKey
	}

	// 解密 Passphrase
	passphrase := ""
	if apiConfig.Passphrase != "" {
		passphrase, err = encrypt.AesDecrypt(apiConfig.Passphrase)
		if err != nil {
			g.Log().Warningf(ctx, "[ExchangeManager] 解密Passphrase失败: %v, 尝试使用原始值", err)
			passphrase = apiConfig.Passphrase
		}
	}

	// 获取代理配置
	proxyConfig := m.getProxyConfig(ctx)

	config := &exchange.Config{
		Platform:   strings.ToLower(strings.TrimSpace(apiConfig.Platform)),
		ApiKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
		IsTestnet:  false, // 可以从配置读取
		Proxy:      proxyConfig,
	}

	ex, err = exchange.NewExchange(config)
	if err != nil {
		return nil, err
	}

	m.exchanges[apiConfig.Id] = ex
	g.Log().Infof(ctx, "[ExchangeManager] 创建交易所实例: platform=%s, apiConfigId=%d", apiConfig.Platform, apiConfig.Id)

	return ex, nil
}

// GetExchangeById 通过API Config ID获取交易所实例
func (m *ExchangeManager) GetExchangeById(ctx context.Context, apiConfigId int64) (exchange.Exchange, error) {
	m.mu.RLock()
	ex, exists := m.exchanges[apiConfigId]
	m.mu.RUnlock()

	if exists {
		return ex, nil
	}

	return nil, gerror.Newf("交易所实例不存在: apiConfigId=%d", apiConfigId)
}

// RemoveExchange 移除交易所实例
func (m *ExchangeManager) RemoveExchange(apiConfigId int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.exchanges, apiConfigId)
}

// getProxyConfig 获取代理配置（从数据库读取全局配置）
func (m *ExchangeManager) getProxyConfig(ctx context.Context) *exchange.ProxyConfig {
	// 从数据库读取全局代理配置（user_id=0, tenant_id=0）
	var config *entity.TradingProxyConfig
	err := dao.TradingProxyConfig.Ctx(ctx).
		Where(dao.TradingProxyConfig.Columns().UserId, 0).   // user_id=0 表示全局配置
		Where(dao.TradingProxyConfig.Columns().TenantId, 0). // tenant_id=0 表示全局配置
		Where(dao.TradingProxyConfig.Columns().Enabled, 1).  // 只获取启用的配置
		Scan(&config)

	if err != nil || config == nil {
		// 如果没有配置或未启用，返回nil（不使用代理）
		return nil
	}

	// 解析代理地址（格式：host:port）
	host := config.ProxyAddress
	port := 0
	if idx := strings.Index(config.ProxyAddress, ":"); idx > 0 {
		host = config.ProxyAddress[:idx]
		portStr := config.ProxyAddress[idx+1:]
		port = g.NewVar(portStr).Int()
	}

	proxyConfig := &exchange.ProxyConfig{
		Enabled:  true,
		Type:     config.ProxyType,
		Host:     host,
		Port:     port,
		Username: "",
		Password: "",
	}

	// 如果启用了认证，解密密码
	if config.AuthEnabled == 1 && config.Username != "" {
		proxyConfig.Username = config.Username
		if config.Password != "" {
			password, err := encrypt.AesDecrypt(config.Password)
			if err == nil {
				proxyConfig.Password = password
			}
		}
	}

	return proxyConfig
}

// TestConnection 测试API连接
func (m *ExchangeManager) TestConnection(ctx context.Context, apiConfig *entity.TradingApiConfig) error {
	ex, err := m.GetExchangeFromConfig(ctx, apiConfig)
	if err != nil {
		return err
	}

	// 尝试获取余额来测试连接
	_, err = ex.GetBalance(ctx)
	if err != nil {
		return gerror.Wrap(err, "API连接测试失败")
	}

	return nil
}
