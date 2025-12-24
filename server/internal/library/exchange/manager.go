// Package exchange
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
// @Description 交易所管理器
package exchange

import (
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"hotgo/internal/dao"
	"hotgo/internal/model/entity"
	"sync"
)

// Manager 交易所管理器
type Manager struct {
	exchanges sync.Map // map[int64]Exchange, key为apiConfigId
}

var manager = &Manager{}

// GetManager 获取管理器实例
func GetManager() *Manager {
	return manager
}

// GetExchange 根据API配置ID获取交易所实例
func (m *Manager) GetExchange(ctx context.Context, apiConfigId int64) (Exchange, error) {
	// 先从缓存获取
	if ex, ok := m.exchanges.Load(apiConfigId); ok {
		return ex.(Exchange), nil
	}

	// 查询API配置
	var apiConfig *entity.TradingApiConfig
	err := dao.TradingApiConfig.Ctx(ctx).Where(dao.TradingApiConfig.Columns().Id, apiConfigId).Scan(&apiConfig)
	if err != nil {
		return nil, gerror.Wrap(err, "获取API配置失败")
	}
	if apiConfig == nil {
		return nil, gerror.Newf("API配置不存在: %d", apiConfigId)
	}

	// 创建交易所实例
	ex, err := NewExchange(&Config{
		Platform:   apiConfig.Platform,
		ApiKey:     apiConfig.ApiKey,
		SecretKey:  apiConfig.SecretKey,
		Passphrase: apiConfig.Passphrase,
		IsTestnet:  false, // 默认为主网
	})
	if err != nil {
		return nil, err
	}

	// 缓存
	m.exchanges.Store(apiConfigId, ex)
	g.Log().Infof(ctx, "创建交易所实例: platform=%s, apiConfigId=%d", apiConfig.Platform, apiConfigId)

	return ex, nil
}

// RemoveExchange 移除交易所实例 (API配置更新时调用)
func (m *Manager) RemoveExchange(apiConfigId int64) {
	m.exchanges.Delete(apiConfigId)
}

// TestConnection 测试API连接
func (m *Manager) TestConnection(ctx context.Context, config *Config) error {
	ex, err := NewExchange(config)
	if err != nil {
		return err
	}

	// 测试获取余额
	balance, err := ex.GetBalance(ctx)
	if err != nil {
		return gerror.Wrap(err, "API连接测试失败")
	}

	g.Log().Infof(ctx, "API连接测试成功: platform=%s, balance=%.4f USDT",
		config.Platform, balance.AvailableBalance)
	return nil
}

