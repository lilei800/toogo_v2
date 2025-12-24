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

var ApiConfig = cApiConfig{}

type cApiConfig struct{}

// List 获取API配置列表
func (c *cApiConfig) List(ctx context.Context, req *trading.ApiConfigListReq) (res *trading.ApiConfigListRes, err error) {
	list, totalCount, err := tradingLogic.ApiConfig.List(ctx, &req.TradingApiConfigListInp)
	if err != nil {
		return nil, err
	}

	res = &trading.ApiConfigListRes{
		List:       list,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}
	return
}

// Create 创建API配置
func (c *cApiConfig) Create(ctx context.Context, req *trading.ApiConfigCreateReq) (res *trading.ApiConfigCreateRes, err error) {
	id, err := tradingLogic.ApiConfig.Create(ctx, &req.TradingApiConfigCreateInp)
	if err != nil {
		return nil, err
	}

	res = &trading.ApiConfigCreateRes{Id: id}
	return
}

// Update 更新API配置
func (c *cApiConfig) Update(ctx context.Context, req *trading.ApiConfigUpdateReq) (res *trading.ApiConfigUpdateRes, err error) {
	err = tradingLogic.ApiConfig.Update(ctx, &req.TradingApiConfigUpdateInp)
	if err != nil {
		return nil, err
	}

	res = &trading.ApiConfigUpdateRes{}
	return
}

// Delete 删除API配置
func (c *cApiConfig) Delete(ctx context.Context, req *trading.ApiConfigDeleteReq) (res *trading.ApiConfigDeleteRes, err error) {
	err = tradingLogic.ApiConfig.Delete(ctx, &req.TradingApiConfigDeleteInp)
	if err != nil {
		return nil, err
	}

	res = &trading.ApiConfigDeleteRes{}
	return
}

// View 查看API配置详情
func (c *cApiConfig) View(ctx context.Context, req *trading.ApiConfigViewReq) (res *trading.ApiConfigViewRes, err error) {
	out, err := tradingLogic.ApiConfig.View(ctx, &req.TradingApiConfigViewInp)
	if err != nil {
		return nil, err
	}

	res = &trading.ApiConfigViewRes{
		TradingApiConfigViewModel: out,
	}
	return
}

// Test 测试API连接
func (c *cApiConfig) Test(ctx context.Context, req *trading.ApiConfigTestReq) (res *trading.ApiConfigTestRes, err error) {
	out, err := tradingLogic.ApiConfig.Test(ctx, &req.TradingApiConfigTestInp)
	if err != nil {
		return nil, err
	}

	res = &trading.ApiConfigTestRes{
		Success: out.Success,
		Message: out.Message,
		Balance: out.Balance,
		Latency: out.Latency,
	}
	return
}

// SetDefault 设为默认配置
func (c *cApiConfig) SetDefault(ctx context.Context, req *trading.ApiConfigSetDefaultReq) (res *trading.ApiConfigSetDefaultRes, err error) {
	err = tradingLogic.ApiConfig.SetDefault(ctx, &req.TradingApiConfigSetDefaultInp)
	if err != nil {
		return nil, err
	}

	res = &trading.ApiConfigSetDefaultRes{}
	return
}

// Platforms 获取支持的平台列表
func (c *cApiConfig) Platforms(ctx context.Context, req *trading.ApiConfigPlatformsReq) (res *trading.ApiConfigPlatformsRes, err error) {
	list, err := tradingLogic.ApiConfig.GetPlatforms(ctx)
	if err != nil {
		return nil, err
	}

	res = &trading.ApiConfigPlatformsRes{
		List: list,
	}
	return
}
