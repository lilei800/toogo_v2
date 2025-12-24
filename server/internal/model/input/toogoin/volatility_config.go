// Package toogoin
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package toogoin

import (
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/form"
)

// VolatilityConfigListInp 波动率配置列表输入
type VolatilityConfigListInp struct {
	form.PageReq
	Symbol   string `json:"symbol" description:"交易对筛选（支持模糊搜索）"`
	IsActive int    `json:"isActive" description:"是否启用: -1=全部, 0=否, 1=是"`
}

// VolatilityConfigListModel 波动率配置列表返回
type VolatilityConfigListModel struct {
	*entity.ToogoVolatilityConfig
}

// VolatilityConfigEditInp 编辑波动率配置输入（适配新算法：市场状态阈值 + delta值 + DThreshold + 5个时间周期权重）
type VolatilityConfigEditInp struct {
	Id                      int64   `json:"id" description:"ID"`
	Symbol                   *string `json:"symbol" description:"交易对（留空表示全局配置，如：BTCUSDT表示BTCUSDT特定配置）"`
	HighVolatilityThreshold  float64 `json:"highVolatilityThreshold" v:"required|min:0.1" description:"高波动阈值HighV（V >= HighV && D < 0.4 -> 高波动）"`
	LowVolatilityThreshold   float64 `json:"lowVolatilityThreshold" v:"required|min:0.01" description:"低波动阈值LowV（V < LowV -> 低波动）"`
	TrendStrengthThreshold   float64 `json:"trendStrengthThreshold" v:"required|min:0.1|max:1" description:"趋势阈值TrendV（V >= TrendV && D >= DThreshold -> 趋势）"`
	DThreshold              float64 `json:"dThreshold" v:"required|min:0|max:1" description:"方向一致性阈值DThreshold（用于判断趋势市场，建议0.7）"`
	RangeVolatilityThreshold float64 `json:"rangeVolatilityThreshold" description:"震荡市场波动率阈值（暂不使用，默认0）"`
	Delta1m                 float64 `json:"delta1m" v:"required|min:0.1" description:"1分钟周期波动点数阈值delta（用于计算V = (H-L)/delta）"`
	Delta5m                 float64 `json:"delta5m" v:"required|min:0.1" description:"5分钟周期波动点数阈值delta"`
	Delta15m                float64 `json:"delta15m" v:"required|min:0.1" description:"15分钟周期波动点数阈值delta"`
	Delta30m                float64 `json:"delta30m" v:"required|min:0.1" description:"30分钟周期波动点数阈值delta"`
	Delta1h                 float64 `json:"delta1h" v:"required|min:0.1" description:"1小时周期波动点数阈值delta"`
	Weight1m                 float64 `json:"weight1m" v:"required|min:0|max:1" description:"1分钟周期权重"`
	Weight5m                float64 `json:"weight5m" v:"required|min:0|max:1" description:"5分钟周期权重"`
	Weight15m               float64 `json:"weight15m" v:"required|min:0|max:1" description:"15分钟周期权重"`
	Weight30m               float64 `json:"weight30m" v:"required|min:0|max:1" description:"30分钟周期权重"`
	Weight1h                float64 `json:"weight1h" v:"required|min:0|max:1" description:"1小时周期权重"`
	IsActive                int     `json:"isActive" v:"required|in:0,1" description:"是否启用"`
}

// VolatilityConfigDeleteInp 删除波动率配置输入
type VolatilityConfigDeleteInp struct {
	Id int64 `json:"id" v:"required" description:"ID"`
}

// VolatilityConfigBatchEditInp 批量编辑波动率配置输入（为多个货币对批量设置，适配新算法）
type VolatilityConfigBatchEditInp struct {
	Symbols                  []string `json:"symbols" v:"required#请选择至少一个交易对" description:"交易对列表"`
	HighVolatilityThreshold  float64  `json:"highVolatilityThreshold" v:"required|min:0.1" description:"高波动阈值HighV"`
	LowVolatilityThreshold   float64  `json:"lowVolatilityThreshold" v:"required|min:0.01" description:"低波动阈值LowV"`
	TrendStrengthThreshold   float64  `json:"trendStrengthThreshold" v:"required|min:0.1|max:1" description:"趋势阈值TrendV"`
	DThreshold              float64  `json:"dThreshold" v:"required|min:0|max:1" description:"方向一致性阈值DThreshold"`
	RangeVolatilityThreshold float64  `json:"rangeVolatilityThreshold" description:"震荡市场波动率阈值（暂不使用，默认0）"`
	Delta1m                 float64  `json:"delta1m" v:"required|min:0.1" description:"1分钟周期delta"`
	Delta5m                 float64  `json:"delta5m" v:"required|min:0.1" description:"5分钟周期delta"`
	Delta15m                float64  `json:"delta15m" v:"required|min:0.1" description:"15分钟周期delta"`
	Delta30m                float64  `json:"delta30m" v:"required|min:0.1" description:"30分钟周期delta"`
	Delta1h                 float64  `json:"delta1h" v:"required|min:0.1" description:"1小时周期delta"`
	Weight1m                 float64  `json:"weight1m" v:"required|min:0|max:1" description:"1分钟周期权重"`
	Weight5m                float64  `json:"weight5m" v:"required|min:0|max:1" description:"5分钟周期权重"`
	Weight15m               float64  `json:"weight15m" v:"required|min:0|max:1" description:"15分钟周期权重"`
	Weight30m               float64  `json:"weight30m" v:"required|min:0|max:1" description:"30分钟周期权重"`
	Weight1h                float64  `json:"weight1h" v:"required|min:0|max:1" description:"1小时周期权重"`
	IsActive                int      `json:"isActive" v:"required|in:0,1" description:"是否启用"`
}

