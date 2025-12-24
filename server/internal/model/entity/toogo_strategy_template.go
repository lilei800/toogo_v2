// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/os/gtime"
)

// ToogoStrategyTemplate is the golang structure for table hg_toogo_strategy_template.
type ToogoStrategyTemplate struct {
	Id                   int64       `json:"id"                   orm:"id"                     description:"主键ID"`
	StrategyKey          string      `json:"strategyKey"          orm:"strategy_key"           description:"策略KEY"`
	StrategyName         string      `json:"strategyName"         orm:"strategy_name"          description:"策略名称"`
	RiskPreference       string      `json:"riskPreference"       orm:"risk_preference"        description:"风险偏好: conservative/balanced/aggressive"`
	MarketState          string      `json:"marketState"          orm:"market_state"           description:"市场状态: trend/volatile/high_volatility/low_volatility"`
	TimeWindow           int         `json:"timeWindow"           orm:"time_window"            description:"时间窗口(秒)"`
	VolatilityPoints     float64     `json:"volatilityPoints"     orm:"volatility_points"      description:"波动点数(USDT)"`
	LeverageMin          int         `json:"leverageMin"          orm:"leverage_min"           description:"杠杆倍数最小值"`
	LeverageMax          int         `json:"leverageMax"          orm:"leverage_max"           description:"杠杆倍数最大值"`
	MarginPercentMin     float64     `json:"marginPercentMin"     orm:"margin_percent_min"     description:"保证金比例最小值(%)"`
	MarginPercentMax     float64     `json:"marginPercentMax"     orm:"margin_percent_max"     description:"保证金比例最大值(%)"`
	StopLossPercent      float64     `json:"stopLossPercent"      orm:"stop_loss_percent"      description:"止损百分比(%)"`
	ProfitRetreatPercent float64     `json:"profitRetreatPercent" orm:"profit_retreat_percent" description:"止盈回撤百分比(%)"`
	StartRetreatPercent  float64     `json:"startRetreatPercent"  orm:"start_retreat_percent"  description:"启动回撤百分比(%)"`
	ReverseLossRetreat   float64     `json:"reverseLossRetreat"   orm:"reverse_loss_retreat"   description:"反向-亏损订单回撤百分比"`
	ReverseProfitRetreat float64     `json:"reverseProfitRetreat" orm:"reverse_profit_retreat" description:"反向-盈利订单回撤百分比"`
	VolatilityConfig     *gjson.Json `json:"volatilityConfig"     orm:"volatility_config"      description:"多周期波动率配置(JSON)"`
	Description          string      `json:"description"          orm:"description"            description:"策略描述"`
	IsOfficial           int         `json:"isOfficial"           orm:"is_official"            description:"是否官方推荐: 0=否, 1=是"`
	IsActive             int         `json:"isActive"             orm:"is_active"              description:"是否激活: 0=否, 1=是"`
	Sort                 int         `json:"sort"                 orm:"sort"                   description:"排序"`
	CreatedAt            *gtime.Time `json:"createdAt"            orm:"created_at"             description:"创建时间"`
	UpdatedAt            *gtime.Time `json:"updatedAt"            orm:"updated_at"             description:"更新时间"`
}

