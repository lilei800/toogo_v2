// Package toogo
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package toogo

import (
	"context"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"hotgo/internal/dao"
	configlib "hotgo/internal/library/config"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/toogoin"
	"hotgo/internal/service"
)

type sToogoVolatilityConfig struct{}

func NewToogoVolatilityConfig() *sToogoVolatilityConfig {
	return &sToogoVolatilityConfig{}
}

func init() {
	service.RegisterToogoVolatilityConfig(NewToogoVolatilityConfig())
}

// List 波动率配置列表（支持按货币对筛选）
func (s *sToogoVolatilityConfig) List(ctx context.Context, in *toogoin.VolatilityConfigListInp) (list []*entity.ToogoVolatilityConfig, total int, err error) {
	mod := dao.ToogoVolatilityConfig.Ctx(ctx)
	cols := dao.ToogoVolatilityConfig.Columns()

	// 支持模糊搜索交易对
	if in.Symbol != "" {
		mod = mod.WhereLike(cols.Symbol, "%"+in.Symbol+"%")
	}
	// IsActive: -1=全部, 0=否, 1=是
	// 只有当 IsActive >= 0 时才添加筛选条件（-1表示全部，不筛选）
	if in.IsActive >= 0 {
		mod = mod.Where(cols.IsActive, in.IsActive)
	}

	// 排序：全局配置在前，然后按交易对字母顺序
	// 注意：使用COALESCE处理NULL值，避免SQL错误
	err = mod.OrderAsc("COALESCE(symbol, '')").
		Page(in.Page, in.PerPage).
		ScanAndCount(&list, &total, true)
	if err != nil {
		err = gerror.Wrap(err, "获取波动率配置列表失败")
		return
	}
	return
}

// Edit 编辑波动率配置（简化版：市场状态阈值 + 5个时间周期权重）
func (s *sToogoVolatilityConfig) Edit(ctx context.Context, in *toogoin.VolatilityConfigEditInp) error {
	cols := dao.ToogoVolatilityConfig.Columns()
	
	// 检查交易对是否已存在（排除当前记录）
	// 只有当创建新配置（id=0）且指定了交易对时才检查
	if in.Id == 0 && in.Symbol != nil && *in.Symbol != "" {
		var existing *entity.ToogoVolatilityConfig
		err := dao.ToogoVolatilityConfig.Ctx(ctx).
			Where(cols.Symbol, *in.Symbol).
			Scan(&existing)
		if err == nil && existing != nil {
			return gerror.Newf("交易对 %s 的配置已存在（ID: %d），请使用编辑功能或先删除现有配置", *in.Symbol, existing.Id)
		}
	}
	
	data := g.Map{
		cols.Symbol:                   in.Symbol,
		cols.HighVolatilityThreshold:  in.HighVolatilityThreshold,
		cols.LowVolatilityThreshold:   in.LowVolatilityThreshold,
		cols.TrendStrengthThreshold:   in.TrendStrengthThreshold,
		cols.DThreshold:              in.DThreshold,
		cols.RangeVolatilityThreshold: in.RangeVolatilityThreshold,
		cols.Delta1m:                 in.Delta1m,
		cols.Delta5m:                 in.Delta5m,
		cols.Delta15m:                in.Delta15m,
		cols.Delta30m:                in.Delta30m,
		cols.Delta1h:                 in.Delta1h,
		cols.Weight1m:                 in.Weight1m,
		cols.Weight5m:                 in.Weight5m,
		cols.Weight15m:                in.Weight15m,
		cols.Weight30m:                in.Weight30m,
		cols.Weight1h:                 in.Weight1h,
		cols.IsActive:                 in.IsActive,
	}

	if in.Id > 0 {
		data[cols.UpdatedAt] = gtime.Now()
		_, err := dao.ToogoVolatilityConfig.Ctx(ctx).Where(dao.ToogoVolatilityConfig.Columns().Id, in.Id).Data(data).Update()
		if err != nil {
			return gerror.Wrap(err, "更新波动率配置失败")
		}
		g.Log().Infof(ctx, "[VolatilityConfig] 更新配置: id=%d, symbol=%v, high=%.2f, low=%.2f, trend=%.2f, dThreshold=%.2f, deltas=[1m:%.2f,5m:%.2f,15m:%.2f,30m:%.2f,1h:%.2f]",
			in.Id, in.Symbol, in.HighVolatilityThreshold, in.LowVolatilityThreshold, in.TrendStrengthThreshold, in.DThreshold,
			in.Delta1m, in.Delta5m, in.Delta15m, in.Delta30m, in.Delta1h)
	} else {
		data[cols.CreatedAt] = gtime.Now()
		data[cols.UpdatedAt] = gtime.Now()
		_, err := dao.ToogoVolatilityConfig.Ctx(ctx).Data(data).Insert()
		if err != nil {
			return gerror.Wrap(err, "创建波动率配置失败")
		}
		g.Log().Infof(ctx, "[VolatilityConfig] 创建配置: symbol=%v, high=%.2f, low=%.2f, trend=%.2f, dThreshold=%.2f, deltas=[1m:%.2f,5m:%.2f,15m:%.2f,30m:%.2f,1h:%.2f]",
			in.Symbol, in.HighVolatilityThreshold, in.LowVolatilityThreshold, in.TrendStrengthThreshold, in.DThreshold,
			in.Delta1m, in.Delta5m, in.Delta15m, in.Delta30m, in.Delta1h)
	}
	
	// 触发全局配置管理器重载（支持热更新）
	symbolStr := ""
	if in.Symbol != nil {
		symbolStr = *in.Symbol
	}
	go configlib.GetVolatilityConfigManager().ReloadConfig(ctx, symbolStr)
	
	return nil
}

// BatchEdit 批量编辑波动率配置（为多个货币对批量设置相同配置）
func (s *sToogoVolatilityConfig) BatchEdit(ctx context.Context, in *toogoin.VolatilityConfigBatchEditInp) error {
	if len(in.Symbols) == 0 {
		return gerror.New("请选择至少一个交易对")
	}

	cols := dao.ToogoVolatilityConfig.Columns()
	now := gtime.Now()
	
	// 批量插入或更新
	for _, symbol := range in.Symbols {
		if symbol == "" {
			continue
		}
		
		// 检查是否已存在
		var existing *entity.ToogoVolatilityConfig
		_ = dao.ToogoVolatilityConfig.Ctx(ctx).
			Where(cols.Symbol, symbol).
			Scan(&existing)
		
		data := g.Map{
			cols.Symbol:                   symbol,
			cols.HighVolatilityThreshold:  in.HighVolatilityThreshold,
			cols.LowVolatilityThreshold:   in.LowVolatilityThreshold,
			cols.TrendStrengthThreshold:   in.TrendStrengthThreshold,
			cols.DThreshold:              in.DThreshold,
			cols.RangeVolatilityThreshold: in.RangeVolatilityThreshold,
			cols.Delta1m:                 in.Delta1m,
			cols.Delta5m:                 in.Delta5m,
			cols.Delta15m:                in.Delta15m,
			cols.Delta30m:                in.Delta30m,
			cols.Delta1h:                 in.Delta1h,
			cols.Weight1m:                 in.Weight1m,
			cols.Weight5m:                 in.Weight5m,
			cols.Weight15m:                in.Weight15m,
			cols.Weight30m:                in.Weight30m,
			cols.Weight1h:                 in.Weight1h,
			cols.IsActive:                 in.IsActive,
		}
		
		if existing != nil {
			// 更新
			data[cols.UpdatedAt] = now
			_, updateErr := dao.ToogoVolatilityConfig.Ctx(ctx).
				Where(cols.Symbol, symbol).
				Data(data).
				Update()
			if updateErr != nil {
				g.Log().Errorf(ctx, "[VolatilityConfig] 批量更新失败: symbol=%s, err=%v", symbol, updateErr)
			}
		} else {
			// 插入
			data[cols.CreatedAt] = now
			data[cols.UpdatedAt] = now
			_, insertErr := dao.ToogoVolatilityConfig.Ctx(ctx).Data(data).Insert()
			if insertErr != nil {
				g.Log().Errorf(ctx, "[VolatilityConfig] 批量创建失败: symbol=%s, err=%v", symbol, insertErr)
			}
		}
	}
	
	g.Log().Infof(ctx, "[VolatilityConfig] 批量更新配置: symbols=%v, high=%.2f, low=%.2f, trend=%.2f, dThreshold=%.2f, deltas=[1m:%.2f,5m:%.2f,15m:%.2f,30m:%.2f,1h:%.2f], weights=[1m:%.2f,5m:%.2f,15m:%.2f,30m:%.2f,1h:%.2f]",
		in.Symbols, in.HighVolatilityThreshold, in.LowVolatilityThreshold, in.TrendStrengthThreshold, in.DThreshold,
		in.Delta1m, in.Delta5m, in.Delta15m, in.Delta30m, in.Delta1h,
		in.Weight1m, in.Weight5m, in.Weight15m, in.Weight30m, in.Weight1h)
	
	// 触发全局配置管理器重载所有相关配置
	go func() {
		for _, symbol := range in.Symbols {
			configlib.GetVolatilityConfigManager().ReloadConfig(ctx, symbol)
		}
	}()
	
	return nil
}

// GetBySymbol 获取波动率配置（优先货币对特定配置，其次全局配置）
func (s *sToogoVolatilityConfig) GetBySymbol(ctx context.Context, symbol string) (*entity.ToogoVolatilityConfig, error) {
	cols := dao.ToogoVolatilityConfig.Columns()

	// 1. 优先获取货币对特定配置
	var config *entity.ToogoVolatilityConfig
	err := dao.ToogoVolatilityConfig.Ctx(ctx).
		Where(cols.Symbol, symbol).
		Where(cols.IsActive, 1).
		Scan(&config)

	if err == nil && config != nil {
		g.Log().Debugf(ctx, "[VolatilityConfig] 使用货币对特定配置: symbol=%s, high=%.2f, low=%.2f",
			symbol, config.HighVolatilityThreshold, config.LowVolatilityThreshold)
		return config, nil
	}

	// 2. 获取全局配置（symbol为NULL）
	err = dao.ToogoVolatilityConfig.Ctx(ctx).
		WhereNull(cols.Symbol).
		Where(cols.IsActive, 1).
		Scan(&config)

	if err == nil && config != nil {
		g.Log().Debugf(ctx, "[VolatilityConfig] 使用全局配置: symbol=%s, high=%.2f, low=%.2f",
			symbol, config.HighVolatilityThreshold, config.LowVolatilityThreshold)
		return config, nil
	}

	// 3. 返回默认配置（适配新算法）
	defaultConfig := &entity.ToogoVolatilityConfig{
		HighVolatilityThreshold: 2.0,
		LowVolatilityThreshold:  1.0,
		TrendStrengthThreshold:  1.2,
		DThreshold:              0.7,
		Delta1m:                2.0, // 1分钟delta
		Delta5m:                2.0, // 5分钟delta
		Delta15m:               3.0, // 15分钟delta
		Delta30m:               3.0, // 30分钟delta
		Delta1h:                5.0, // 1小时delta
		Weight1m:                0.20, // 1分钟权重20%
		Weight5m:                0.25, // 5分钟权重25%
		Weight15m:               0.25, // 15分钟权重25%
		Weight30m:               0.20, // 30分钟权重20%
		Weight1h:                0.10, // 1小时权重10%
	}
	g.Log().Debugf(ctx, "[VolatilityConfig] 使用默认配置: symbol=%s", symbol)
	return defaultConfig, nil
}

// Delete 删除波动率配置
func (s *sToogoVolatilityConfig) Delete(ctx context.Context, id int64) error {
	// 检查是否是全局配置
	var config *entity.ToogoVolatilityConfig
	err := dao.ToogoVolatilityConfig.Ctx(ctx).Where(dao.ToogoVolatilityConfig.Columns().Id, id).Scan(&config)
	if err == nil && config != nil && config.Symbol == nil {
		return gerror.New("不能删除全局配置，请先创建新的全局配置后再删除")
	}
	
	_, err = dao.ToogoVolatilityConfig.Ctx(ctx).Where(dao.ToogoVolatilityConfig.Columns().Id, id).Delete()
	if err != nil {
		return gerror.Wrap(err, "删除波动率配置失败")
	}
	return nil
}

// GetAllSymbols 获取所有已配置的交易对列表
func (s *sToogoVolatilityConfig) GetAllSymbols(ctx context.Context) ([]string, error) {
	var configs []*entity.ToogoVolatilityConfig
	err := dao.ToogoVolatilityConfig.Ctx(ctx).
		Where("symbol IS NOT NULL").
		Where("is_active", 1).
		OrderAsc("symbol").
		Scan(&configs)
	
	if err != nil {
		return nil, gerror.Wrap(err, "获取交易对列表失败")
	}
	
	symbols := make([]string, 0, len(configs))
	for _, config := range configs {
		if config.Symbol != nil && *config.Symbol != "" {
			symbols = append(symbols, *config.Symbol)
		}
	}
	return symbols, nil
}

