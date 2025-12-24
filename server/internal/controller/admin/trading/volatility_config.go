package trading

import (
	"context"

	"hotgo/api/admin/trading"
	"hotgo/internal/model/input/form"
	"hotgo/internal/model/input/toogoin"
	"hotgo/internal/service"
)

// VolatilityConfig 波动率配置控制器（支持每个货币对独立配置）
var VolatilityConfig = cVolatilityConfig{}

type cVolatilityConfig struct{}

// List 波动率配置列表（支持按货币对筛选）
func (c *cVolatilityConfig) List(ctx context.Context, req *trading.VolatilityConfigListReq) (res *trading.VolatilityConfigListRes, err error) {
	isActive := -1
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	list, total, err := service.ToogoVolatilityConfig().List(ctx, &toogoin.VolatilityConfigListInp{
		PageReq:  form.PageReq{Page: req.Page, PerPage: req.PageSize},
		Symbol:   req.Symbol,
		IsActive: isActive,
	})
	if err != nil {
		return nil, err
	}

	res = &trading.VolatilityConfigListRes{
		List:  list,
		Total: total,
		Page:  req.Page,
	}
	return
}

// Create 创建波动率配置（简化版：市场状态阈值 + 5个时间周期权重）
func (c *cVolatilityConfig) Create(ctx context.Context, req *trading.VolatilityConfigCreateReq) (res *trading.VolatilityConfigCreateRes, err error) {
	err = service.ToogoVolatilityConfig().Edit(ctx, &toogoin.VolatilityConfigEditInp{
		Id:                       0, // 创建时ID为0
		Symbol:                   req.Symbol,
		HighVolatilityThreshold:  req.HighVolatilityThreshold,
		LowVolatilityThreshold:   req.LowVolatilityThreshold,
		TrendStrengthThreshold:   req.TrendStrengthThreshold,
		DThreshold:               req.DThreshold,
		RangeVolatilityThreshold: 0.0, // 暂不使用，设为0
		Delta1m:                  req.Delta1m,
		Delta5m:                  req.Delta5m,
		Delta15m:                 req.Delta15m,
		Delta30m:                 req.Delta30m,
		Delta1h:                  req.Delta1h,
		Weight1m:                 req.Weight1m,
		Weight5m:                 req.Weight5m,
		Weight15m:                req.Weight15m,
		Weight30m:                req.Weight30m,
		Weight1h:                 req.Weight1h,
		IsActive:                 req.IsActive,
	})
	if err != nil {
		return nil, err
	}

	res = &trading.VolatilityConfigCreateRes{}
	return
}

// Update 更新波动率配置（简化版：5个时间周期权重）
func (c *cVolatilityConfig) Update(ctx context.Context, req *trading.VolatilityConfigUpdateReq) (res *trading.VolatilityConfigUpdateRes, err error) {
	err = service.ToogoVolatilityConfig().Edit(ctx, &toogoin.VolatilityConfigEditInp{
		Id:                       req.Id,
		Symbol:                   req.Symbol,
		HighVolatilityThreshold:  req.HighVolatilityThreshold,
		LowVolatilityThreshold:   req.LowVolatilityThreshold,
		TrendStrengthThreshold:   req.TrendStrengthThreshold,
		DThreshold:               req.DThreshold,
		RangeVolatilityThreshold: 0.0, // 暂不使用，设为0
		Delta1m:                  req.Delta1m,
		Delta5m:                  req.Delta5m,
		Delta15m:                 req.Delta15m,
		Delta30m:                 req.Delta30m,
		Delta1h:                  req.Delta1h,
		Weight1m:                 req.Weight1m,
		Weight5m:                 req.Weight5m,
		Weight15m:                req.Weight15m,
		Weight30m:                req.Weight30m,
		Weight1h:                 req.Weight1h,
		IsActive:                 req.IsActive,
	})
	if err != nil {
		return nil, err
	}

	res = &trading.VolatilityConfigUpdateRes{}
	return
}

// Delete 删除波动率配置
func (c *cVolatilityConfig) Delete(ctx context.Context, req *trading.VolatilityConfigDeleteReq) (res *trading.VolatilityConfigDeleteRes, err error) {
	err = service.ToogoVolatilityConfig().Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	res = &trading.VolatilityConfigDeleteRes{}
	return
}

// BatchEdit 批量编辑波动率配置（简化版：5个时间周期权重，为多个货币对批量设置相同配置）
func (c *cVolatilityConfig) BatchEdit(ctx context.Context, req *trading.VolatilityConfigBatchEditReq) (res *trading.VolatilityConfigBatchEditRes, err error) {
	err = service.ToogoVolatilityConfig().BatchEdit(ctx, &toogoin.VolatilityConfigBatchEditInp{
		Symbols:                 req.Symbols,
		HighVolatilityThreshold: req.HighVolatilityThreshold,
		LowVolatilityThreshold:  req.LowVolatilityThreshold,
		TrendStrengthThreshold:  req.TrendStrengthThreshold,
		DThreshold:              req.DThreshold,
		Weight1m:                req.Weight1m,
		Weight5m:                req.Weight5m,
		Weight15m:               req.Weight15m,
		Weight30m:               req.Weight30m,
		Weight1h:                req.Weight1h,
		Delta1m:                 req.Delta1m,
		Delta5m:                 req.Delta5m,
		Delta15m:                req.Delta15m,
		Delta30m:                req.Delta30m,
		Delta1h:                 req.Delta1h,
		IsActive:                req.IsActive,
	})
	if err != nil {
		return nil, err
	}

	res = &trading.VolatilityConfigBatchEditRes{}
	return
}

// GetSymbols 获取所有已配置的交易对列表
func (c *cVolatilityConfig) GetSymbols(ctx context.Context, req *trading.VolatilityConfigGetSymbolsReq) (res *trading.VolatilityConfigGetSymbolsRes, err error) {
	symbols, err := service.ToogoVolatilityConfig().GetAllSymbols(ctx)
	if err != nil {
		return nil, err
	}

	res = &trading.VolatilityConfigGetSymbolsRes{
		Symbols: symbols,
	}
	return
}
