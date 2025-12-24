// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package trading

import (
	"context"
	"hotgo/api/admin/trading"
	tradingLogic "hotgo/internal/logic/trading"
)

var ProxyConfig = cProxyConfig{}

type cProxyConfig struct{}

// Get 获取代理配置
func (c *cProxyConfig) Get(ctx context.Context, req *trading.ProxyConfigGetReq) (res *trading.ProxyConfigGetRes, err error) {
	out, err := tradingLogic.ProxyConfig.Get(ctx, &req.TradingProxyConfigGetInp)
	if err != nil {
		return nil, err
	}

	res = &trading.ProxyConfigGetRes{
		TradingProxyConfigModel: out,
	}
	return
}

// Save 保存代理配置
func (c *cProxyConfig) Save(ctx context.Context, req *trading.ProxyConfigSaveReq) (res *trading.ProxyConfigSaveRes, err error) {
	err = tradingLogic.ProxyConfig.Save(ctx, &req.TradingProxyConfigSaveInp)
	if err != nil {
		return nil, err
	}

	res = &trading.ProxyConfigSaveRes{}
	return
}

// Test 测试代理连接
func (c *cProxyConfig) Test(ctx context.Context, req *trading.ProxyConfigTestReq) (res *trading.ProxyConfigTestRes, err error) {
	out, err := tradingLogic.ProxyConfig.Test(ctx, &req.TradingProxyConfigTestInp)
	if err != nil {
		return nil, err
	}

	res = &trading.ProxyConfigTestRes{
		Success:    out.Success,
		Message:    out.Message,
		ExternalIP: out.ExternalIP,
		Latency:    out.Latency,
	}
	return
}

// Toggle 切换启用状态
func (c *cProxyConfig) Toggle(ctx context.Context, req *trading.ProxyConfigToggleReq) (res *trading.ProxyConfigToggleRes, err error) {
	err = tradingLogic.ProxyConfig.Toggle(ctx, &req.TradingProxyConfigToggleInp)
	if err != nil {
		return nil, err
	}

	res = &trading.ProxyConfigToggleRes{}
	return
}

