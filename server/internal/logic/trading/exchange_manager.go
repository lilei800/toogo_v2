// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package trading

import (
	"context"
	"fmt"
	"hotgo/addons/exchange"
	"hotgo/addons/exchange_bitget/service"
	"hotgo/internal/library/contexts"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
)

type exchangeManagerImpl struct {
	factory *service.ExchangeFactory
	cache   sync.Map // 缓存交易所实例
}

var ExchangeManager = &exchangeManagerImpl{
	factory: service.NewExchangeFactory(),
}

// GetExchange 获取交易所实例
func (s *exchangeManagerImpl) GetExchange(ctx context.Context, apiConfigId int64) (exchange.IExchange, error) {
	userId := contexts.GetUserId(ctx)
	tenantId := contexts.GetTenantId(ctx)

	if userId <= 0 {
		return nil, gerror.New("用户未登录")
	}

	// 从缓存中获取
	cacheKey := getCacheKey(tenantId, userId, apiConfigId)
	if cached, ok := s.cache.Load(cacheKey); ok {
		return cached.(exchange.IExchange), nil
	}

	// 创建新实例
	exchangeInst, err := s.factory.CreateExchange(ctx, apiConfigId, userId, tenantId)
	if err != nil {
		return nil, err
	}

	// 缓存实例
	s.cache.Store(cacheKey, exchangeInst)

	return exchangeInst, nil
}

// ClearCache 清除缓存
func (s *exchangeManagerImpl) ClearCache(tenantId int64, userId int64, apiConfigId int64) {
	cacheKey := getCacheKey(tenantId, userId, apiConfigId)
	s.cache.Delete(cacheKey)
}

// getCacheKey 生成缓存Key
func getCacheKey(tenantId, userId, apiConfigId int64) string {
	return fmt.Sprintf("%d:%d:%d", tenantId, userId, apiConfigId)
}
