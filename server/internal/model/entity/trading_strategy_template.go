// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingStrategyTemplate is the golang structure for table trading_strategy_template.
type TradingStrategyTemplate struct {
	Id                       int64       `json:"id"                       orm:"id"                          description:"主键ID"`
	GroupId                  int64       `json:"groupId"                  orm:"group_id"                    description:"策略组ID"`
	StrategyKey              string      `json:"strategyKey"              orm:"strategy_key"                description:"策略KEY：conservative_trend"`
	StrategyName             string      `json:"strategyName"             orm:"strategy_name"               description:"策略名称"`
	RiskPreference           string      `json:"riskPreference"           orm:"risk_preference"             description:"风险偏好：conservative/balanced/aggressive"`
	MarketState              string      `json:"marketState"              orm:"market_state"                description:"市场状态：trend/volatile/high-volatility/low-volatility"`
	MonitorWindow            int         `json:"monitorWindow"            orm:"monitor_window"              description:"监控时间窗口(秒)"`
	VolatilityThreshold      float64     `json:"volatilityThreshold"      orm:"volatility_threshold"        description:"波动阈值(USDT)"`
	Leverage                 int         `json:"leverage"                 orm:"leverage"                    description:"杠杆倍数"`
	MarginPercent            float64     `json:"marginPercent"            orm:"margin_percent"              description:"保证金比例(%)"`
	StopLossPercent          float64     `json:"stopLossPercent"          orm:"stop_loss_percent"           description:"止损百分比(%)"`
	ProfitRetreatPercent     float64     `json:"profitRetreatPercent"     orm:"profit_retreat_percent"      description:"止盈回撤百分比(%)"`
	AutoStartRetreatPercent  float64     `json:"autoStartRetreatPercent"  orm:"auto_start_retreat_percent"  description:"启动回撤百分比(%)"`
	ConfigJson               string      `json:"configJson"               orm:"config_json"                 description:"其他配置(JSON)"`
	Description              string      `json:"description"              orm:"description"                 description:"策略描述"`
	IsActive                 int         `json:"isActive"                 orm:"is_active"                   description:"是否激活"`
	Sort                     int         `json:"sort"                     orm:"sort"                        description:"排序"`
	CreatedAt                *gtime.Time `json:"createdAt"                orm:"created_at"                  description:"创建时间"`
	UpdatedAt                *gtime.Time `json:"updatedAt"                orm:"updated_at"                  description:"更新时间"`
}

