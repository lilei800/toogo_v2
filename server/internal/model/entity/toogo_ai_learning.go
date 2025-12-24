// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ToogoAiLearning is the golang structure for table hg_toogo_ai_learning.
type ToogoAiLearning struct {
	Id               int64       `json:"id"               orm:"id"                description:"主键ID"`
	Symbol           string      `json:"symbol"           orm:"symbol"            description:"交易对"`
	TimeFrame        string      `json:"timeFrame"        orm:"time_frame"        description:"时间周期: 1m/5m/15m/30m/1h"`
	MarketState      string      `json:"marketState"      orm:"market_state"      description:"市场状态"`
	RiskPreference   string      `json:"riskPreference"   orm:"risk_preference"   description:"风险偏好"`
	PriceWeight      float64     `json:"priceWeight"      orm:"price_weight"      description:"价格权重"`
	VolumeWeight     float64     `json:"volumeWeight"     orm:"volume_weight"     description:"成交量权重"`
	TrendWeight      float64     `json:"trendWeight"      orm:"trend_weight"      description:"趋势权重"`
	VolatilityWeight float64     `json:"volatilityWeight" orm:"volatility_weight" description:"波动率权重"`
	TotalSignals     int         `json:"totalSignals"     orm:"total_signals"     description:"总信号数"`
	CorrectSignals   int         `json:"correctSignals"   orm:"correct_signals"   description:"正确信号数"`
	AccuracyRate     float64     `json:"accuracyRate"     orm:"accuracy_rate"     description:"准确率"`
	TotalProfit      float64     `json:"totalProfit"      orm:"total_profit"      description:"总收益"`
	AvgProfit        float64     `json:"avgProfit"        orm:"avg_profit"        description:"平均收益"`
	LastUpdate       *gtime.Time `json:"lastUpdate"       orm:"last_update"       description:"最后更新时间"`
	CreatedAt        *gtime.Time `json:"createdAt"        orm:"created_at"        description:"创建时间"`
	UpdatedAt        *gtime.Time `json:"updatedAt"        orm:"updated_at"        description:"更新时间"`
}

