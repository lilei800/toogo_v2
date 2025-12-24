// Package trading
package trading

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ==================== 波动率配置 API（支持每个货币对独立配置） ====================

type VolatilityConfigListReq struct {
	g.Meta   `path:"/volatility/config/list" method:"get" tags:"量化管理" summary:"波动率配置列表"`
	Page     int    `json:"page" d:"1"`
	PageSize int    `json:"pageSize" d:"10"`
	Symbol   string `json:"symbol"`   // 交易对筛选（支持模糊搜索）
	IsActive *int   `json:"isActive"` // 是否启用: -1=全部, 0=否, 1=是
}

type VolatilityConfigListRes struct {
	List  interface{} `json:"list"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
}

// VolatilityConfigCreateReq 创建波动率配置（简化版：市场状态阈值 + 5个时间周期权重）
type VolatilityConfigCreateReq struct {
	g.Meta                  `path:"/volatility/config/create" method:"post" tags:"量化管理" summary:"创建波动率配置"`
	Symbol                  *string `json:"symbol" description:"交易对（留空表示全局配置，如：BTCUSDT表示BTCUSDT特定配置）"`
	HighVolatilityThreshold float64 `json:"highVolatilityThreshold" v:"required|min:0.1#请输入高波动阈值|高波动阈值不能小于0.1"`
	LowVolatilityThreshold  float64 `json:"lowVolatilityThreshold" v:"required|min:0.01#请输入低波动阈值|低波动阈值不能小于0.01"`
	TrendStrengthThreshold  float64 `json:"trendStrengthThreshold" v:"required|min:0.1#请输入趋势阈值|趋势阈值不能小于0.1"`
	DThreshold              float64 `json:"dThreshold" v:"required|min:0|max:1#请输入方向一致性阈值|方向一致性阈值需在0-1之间"`
	Delta1m                 float64 `json:"delta1m" v:"required|min:0.1#请输入1分钟delta|delta不能小于0.1"`
	Delta5m                 float64 `json:"delta5m" v:"required|min:0.1#请输入5分钟delta|delta不能小于0.1"`
	Delta15m                float64 `json:"delta15m" v:"required|min:0.1#请输入15分钟delta|delta不能小于0.1"`
	Delta30m                float64 `json:"delta30m" v:"required|min:0.1#请输入30分钟delta|delta不能小于0.1"`
	Delta1h                 float64 `json:"delta1h" v:"required|min:0.1#请输入1小时delta|delta不能小于0.1"`
	Weight1m                float64 `json:"weight1m" v:"required|min:0|max:1#请输入1分钟权重|权重需在0-1之间"`
	Weight5m                float64 `json:"weight5m" v:"required|min:0|max:1#请输入5分钟权重|权重需在0-1之间"`
	Weight15m               float64 `json:"weight15m" v:"required|min:0|max:1#请输入15分钟权重|权重需在0-1之间"`
	Weight30m               float64 `json:"weight30m" v:"required|min:0|max:1#请输入30分钟权重|权重需在0-1之间"`
	Weight1h                float64 `json:"weight1h" v:"required|min:0|max:1#请输入1小时权重|权重需在0-1之间"`
	IsActive                int     `json:"isActive" v:"required|in:0,1#请选择是否启用|启用状态无效"`
}

type VolatilityConfigCreateRes struct{}

// VolatilityConfigUpdateReq 更新波动率配置（简化版：5个时间周期权重）
type VolatilityConfigUpdateReq struct {
	g.Meta                  `path:"/volatility/config/update" method:"post" tags:"量化管理" summary:"更新波动率配置"`
	Id                      int64   `json:"id" v:"required#请指定配置ID"`
	Symbol                  *string `json:"symbol" description:"交易对（留空表示全局配置，如：BTCUSDT表示BTCUSDT特定配置）"`
	HighVolatilityThreshold float64 `json:"highVolatilityThreshold" v:"required|min:0.1"`
	LowVolatilityThreshold  float64 `json:"lowVolatilityThreshold" v:"required|min:0.01"`
	TrendStrengthThreshold  float64 `json:"trendStrengthThreshold" v:"required|min:0.1"`
	DThreshold              float64 `json:"dThreshold" v:"required|min:0|max:1"`
	Delta1m                 float64 `json:"delta1m" v:"required|min:0.1"`
	Delta5m                 float64 `json:"delta5m" v:"required|min:0.1"`
	Delta15m                float64 `json:"delta15m" v:"required|min:0.1"`
	Delta30m                float64 `json:"delta30m" v:"required|min:0.1"`
	Delta1h                 float64 `json:"delta1h" v:"required|min:0.1"`
	Weight1m                float64 `json:"weight1m" v:"required|min:0|max:1"`
	Weight5m                float64 `json:"weight5m" v:"required|min:0|max:1"`
	Weight15m               float64 `json:"weight15m" v:"required|min:0|max:1"`
	Weight30m               float64 `json:"weight30m" v:"required|min:0|max:1"`
	Weight1h                float64 `json:"weight1h" v:"required|min:0|max:1"`
	IsActive                int     `json:"isActive" v:"required|in:0,1"`
}

type VolatilityConfigUpdateRes struct{}

type VolatilityConfigDeleteReq struct {
	g.Meta `path:"/volatility/config/delete" method:"post" tags:"量化管理" summary:"删除波动率配置"`
	Id     int64 `json:"id" v:"required#请指定配置ID"`
}

type VolatilityConfigDeleteRes struct{}

// VolatilityConfigBatchEditReq 批量编辑波动率配置（简化版：5个时间周期权重）
type VolatilityConfigBatchEditReq struct {
	g.Meta                  `path:"/volatility/config/batch-edit" method:"post" tags:"量化管理" summary:"批量编辑波动率配置（为多个货币对批量设置）"`
	Symbols                 []string `json:"symbols" v:"required#请选择至少一个交易对"`
	HighVolatilityThreshold float64  `json:"highVolatilityThreshold" v:"required|min:0.1"`
	LowVolatilityThreshold  float64  `json:"lowVolatilityThreshold" v:"required|min:0.01"`
	TrendStrengthThreshold  float64  `json:"trendStrengthThreshold" v:"required|min:0.1"`
	DThreshold              float64  `json:"dThreshold" v:"required|min:0|max:1"`
	Delta1m                 float64  `json:"delta1m" v:"required|min:0.1"`
	Delta5m                 float64  `json:"delta5m" v:"required|min:0.1"`
	Delta15m                float64  `json:"delta15m" v:"required|min:0.1"`
	Delta30m                float64  `json:"delta30m" v:"required|min:0.1"`
	Delta1h                 float64  `json:"delta1h" v:"required|min:0.1"`
	Weight1m                float64  `json:"weight1m" v:"required|min:0|max:1"`
	Weight5m                float64  `json:"weight5m" v:"required|min:0|max:1"`
	Weight15m               float64  `json:"weight15m" v:"required|min:0|max:1"`
	Weight30m               float64  `json:"weight30m" v:"required|min:0|max:1"`
	Weight1h                float64  `json:"weight1h" v:"required|min:0|max:1"`
	IsActive                int      `json:"isActive" v:"required|in:0,1"`
}

type VolatilityConfigBatchEditRes struct{}

type VolatilityConfigGetSymbolsReq struct {
	g.Meta `path:"/volatility/config/symbols" method:"get" tags:"量化管理" summary:"获取所有已配置的交易对列表"`
}

type VolatilityConfigGetSymbolsRes struct {
	Symbols []string `json:"symbols"`
}
