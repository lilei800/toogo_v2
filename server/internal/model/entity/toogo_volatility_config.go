// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ToogoVolatilityConfig 量化管理波动率配置（适配新算法）
// 支持为每个货币对（交易对）配置独立的市场状态阈值、delta值和5个时间周期权重
type ToogoVolatilityConfig struct {
	Id                      int64       `json:"id"                       orm:"id"                          description:"主键ID"`
	Symbol                  *string     `json:"symbol"                   orm:"symbol"                      description:"交易对（NULL表示全局配置，如：BTCUSDT表示BTCUSDT特定配置）"`
	HighVolatilityThreshold float64     `json:"highVolatilityThreshold"  orm:"high_volatility_threshold"   description:"高波动阈值HighV（判断高波动市场：V >= HighV && D < 0.4）"`
	LowVolatilityThreshold  float64     `json:"lowVolatilityThreshold"   orm:"low_volatility_threshold"    description:"低波动阈值LowV（判断低波动市场：V < LowV）"`
	TrendStrengthThreshold   float64     `json:"trendStrengthThreshold"   orm:"trend_strength_threshold"    description:"趋势阈值TrendV（判断趋势市场：V >= TrendV && D >= DThreshold）"`
	DThreshold              float64     `json:"dThreshold"               orm:"d_threshold"                 description:"方向一致性阈值DThreshold（用于判断趋势市场，0-1之间，建议0.7）"`
	RangeVolatilityThreshold float64     `json:"rangeVolatilityThreshold" orm:"range_volatility_threshold"  description:"震荡市场波动率阈值（暂不使用，默认0）"`
	Delta1m                 float64     `json:"delta1m"                   orm:"delta_1m"                   description:"1分钟周期波动点数阈值delta（用于计算V = (H-L)/delta）"`
	Delta5m                 float64     `json:"delta5m"                   orm:"delta_5m"                   description:"5分钟周期波动点数阈值delta"`
	Delta15m                float64     `json:"delta15m"                  orm:"delta_15m"                  description:"15分钟周期波动点数阈值delta"`
	Delta30m                float64     `json:"delta30m"                  orm:"delta_30m"                  description:"30分钟周期波动点数阈值delta"`
	Delta1h                 float64     `json:"delta1h"                   orm:"delta_1h"                   description:"1小时周期波动点数阈值delta"`
	Weight1m                 float64     `json:"weight1m"                 orm:"weight_1m"                   description:"1分钟周期权重"`
	Weight5m                float64     `json:"weight5m"                 orm:"weight_5m"                   description:"5分钟周期权重"`
	Weight15m               float64     `json:"weight15m"                orm:"weight_15m"                  description:"15分钟周期权重"`
	Weight30m               float64     `json:"weight30m"                orm:"weight_30m"                  description:"30分钟周期权重"`
	Weight1h                float64     `json:"weight1h"                 orm:"weight_1h"                   description:"1小时周期权重"`
	IsActive                int         `json:"isActive"                 orm:"is_active"                   description:"是否启用: 0=否, 1=是"`
	CreatedAt               *gtime.Time `json:"createdAt"                orm:"created_at"                  description:"创建时间"`
	UpdatedAt               *gtime.Time `json:"updatedAt"                orm:"updated_at"                  description:"更新时间"`
}

