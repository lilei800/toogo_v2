// Package entity 预警日志实体
package entity

import "github.com/gogf/gf/v2/os/gtime"

// TradingMarketStateLog 市场状态预警日志
type TradingMarketStateLog struct {
	Id            int64       `json:"id"            orm:"id"              description:"日志ID"`
	Platform      string      `json:"platform"      orm:"platform"        description:"交易所平台"`
	Symbol        string      `json:"symbol"        orm:"symbol"          description:"交易对"`
	PrevState     string      `json:"prevState"     orm:"prev_state"      description:"之前的市场状态"`
	NewState      string      `json:"newState"      orm:"new_state"       description:"新的市场状态"`
	Confidence    float64     `json:"confidence"    orm:"confidence"      description:"置信度"`
	TrendStrength float64     `json:"trendStrength" orm:"trend_strength"  description:"趋势强度"`
	Volatility    float64     `json:"volatility"    orm:"volatility"      description:"波动率"`
	TrendScore    float64     `json:"trendScore"    orm:"trend_score"     description:"趋势评分"`
	MomentumScore float64     `json:"momentumScore" orm:"momentum_score"  description:"动量评分"`
	Reason        string      `json:"reason"        orm:"reason"          description:"预警原因"`
	Indicators    string      `json:"indicators"    orm:"indicators"      description:"技术指标JSON"`
	CreatedAt     *gtime.Time `json:"createdAt"     orm:"created_at"      description:"创建时间"`
}

// TradingRiskPreferenceLog 风险偏好预警日志
type TradingRiskPreferenceLog struct {
	Id                     int64       `json:"id"                     orm:"id"                        description:"日志ID"`
	RobotId                int64       `json:"robotId"                orm:"robot_id"                  description:"机器人ID"`
	UserId                 int64       `json:"userId"                 orm:"user_id"                   description:"用户ID"`
	Platform               string      `json:"platform"               orm:"platform"                  description:"交易所平台"`
	Symbol                 string      `json:"symbol"                 orm:"symbol"                    description:"交易对"`
	PrevPreference         string      `json:"prevPreference"         orm:"prev_preference"           description:"之前的风险偏好"`
	NewPreference          string      `json:"newPreference"          orm:"new_preference"            description:"新的风险偏好"`
	WinProbability         float64     `json:"winProbability"         orm:"win_probability"           description:"胜算概率"`
	MarketScore            float64     `json:"marketScore"            orm:"market_score"              description:"市场状态评分"`
	TechnicalScore         float64     `json:"technicalScore"         orm:"technical_score"           description:"技术指标评分"`
	AccountScore           float64     `json:"accountScore"           orm:"account_score"             description:"账户状况评分"`
	HistoryScore           float64     `json:"historyScore"           orm:"history_score"             description:"历史表现评分"`
	VolatilityRisk         float64     `json:"volatilityRisk"         orm:"volatility_risk"           description:"波动风险"`
	SuggestedLeverage      int         `json:"suggestedLeverage"      orm:"suggested_leverage"        description:"建议杠杆"`
	SuggestedMarginPercent float64     `json:"suggestedMarginPercent" orm:"suggested_margin_percent"  description:"建议保证金比例"`
	SuggestedStopLoss      float64     `json:"suggestedStopLoss"      orm:"suggested_stop_loss"       description:"建议止损比例"`
	SuggestedTakeProfit    float64     `json:"suggestedTakeProfit"    orm:"suggested_take_profit"     description:"建议止盈比例"`
	Reason                 string      `json:"reason"                 orm:"reason"                    description:"预警原因"`
	Factors                string      `json:"factors"                orm:"factors"                   description:"评估因素JSON"`
	CreatedAt              *gtime.Time `json:"createdAt"              orm:"created_at"                description:"创建时间"`
}

// TradingDirectionLog 方向预警日志
type TradingDirectionLog struct {
	Id               int64       `json:"id"               orm:"id"                 description:"日志ID"`
	Platform         string      `json:"platform"         orm:"platform"           description:"交易所平台"`
	Symbol           string      `json:"symbol"           orm:"symbol"             description:"交易对"`
	PrevDirection    string      `json:"prevDirection"    orm:"prev_direction"     description:"之前的方向"`
	NewDirection     string      `json:"newDirection"     orm:"new_direction"      description:"新的方向"`
	Strength         float64     `json:"strength"         orm:"strength"           description:"信号强度"`
	Confidence       float64     `json:"confidence"       orm:"confidence"         description:"置信度"`
	Action           string      `json:"action"           orm:"action"             description:"建议操作"`
	TrendSignal      string      `json:"trendSignal"      orm:"trend_signal"       description:"趋势信号"`
	MomentumSignal   string      `json:"momentumSignal"   orm:"momentum_signal"    description:"动量信号"`
	PatternSignal    string      `json:"patternSignal"    orm:"pattern_signal"     description:"形态信号"`
	NearSupport      int         `json:"nearSupport"      orm:"near_support"       description:"是否接近支撑位"`
	NearResistance   int         `json:"nearResistance"   orm:"near_resistance"    description:"是否接近阻力位"`
	EntryPrice       float64     `json:"entryPrice"       orm:"entry_price"        description:"建议入场价"`
	StopLoss         float64     `json:"stopLoss"         orm:"stop_loss"          description:"建议止损价"`
	TakeProfit1      float64     `json:"takeProfit1"      orm:"take_profit_1"      description:"止盈目标1"`
	TakeProfit2      float64     `json:"takeProfit2"      orm:"take_profit_2"      description:"止盈目标2"`
	Reason           string      `json:"reason"           orm:"reason"             description:"预警原因"`
	TimeframeSignals string      `json:"timeframeSignals" orm:"timeframe_signals"  description:"各周期信号JSON"`
	Indicators       string      `json:"indicators"       orm:"indicators"         description:"技术指标JSON"`
	CreatedAt        *gtime.Time `json:"createdAt"        orm:"created_at"         description:"创建时间"`
}

// TradingRobotRealtime 机器人实时状态
type TradingRobotRealtime struct {
	Id                  int64       `json:"id"                  orm:"id"                    description:"ID"`
	RobotId             int64       `json:"robotId"             orm:"robot_id"              description:"机器人ID"`
	UserId              int64       `json:"userId"              orm:"user_id"               description:"用户ID"`
	Platform            string      `json:"platform"            orm:"platform"              description:"交易所平台"`
	Symbol              string      `json:"symbol"              orm:"symbol"                description:"交易对"`
	CurrentPrice        float64     `json:"currentPrice"        orm:"current_price"         description:"当前价格"`
	PriceChange24h      float64     `json:"priceChange24h"      orm:"price_change_24h"      description:"24h涨跌幅"`
	MarketState         string      `json:"marketState"         orm:"market_state"          description:"市场状态"`
	MarketStateConf     float64     `json:"marketStateConf"     orm:"market_state_conf"     description:"市场状态置信度"`
	TrendStrength       float64     `json:"trendStrength"       orm:"trend_strength"        description:"趋势强度"`
	Volatility          float64     `json:"volatility"          orm:"volatility"            description:"波动率"`
	RiskPreference      string      `json:"riskPreference"      orm:"risk_preference"       description:"风险偏好"`
	WinProbability      float64     `json:"winProbability"      orm:"win_probability"       description:"胜算概率"`
	RiskLevel           int         `json:"riskLevel"           orm:"risk_level"            description:"风险等级"`
	Direction           string      `json:"direction"           orm:"direction"             description:"方向信号"`
	DirectionStrength   float64     `json:"directionStrength"   orm:"direction_strength"    description:"方向强度"`
	DirectionConfidence float64     `json:"directionConfidence" orm:"direction_confidence"  description:"方向置信度"`
	SuggestedAction     string      `json:"suggestedAction"     orm:"suggested_action"      description:"建议操作"`
	HasPosition         int         `json:"hasPosition"         orm:"has_position"          description:"是否有持仓"`
	PositionSide        string      `json:"positionSide"        orm:"position_side"         description:"持仓方向"`
	PositionAmt         float64     `json:"positionAmt"         orm:"position_amt"          description:"持仓数量"`
	PositionPnl         float64     `json:"positionPnl"         orm:"position_pnl"          description:"持仓盈亏"`
	PositionPnlPercent  float64     `json:"positionPnlPercent"  orm:"position_pnl_percent"  description:"持仓盈亏比例"`
	AccountBalance      float64     `json:"accountBalance"      orm:"account_balance"       description:"账户余额"`
	AvailableBalance    float64     `json:"availableBalance"    orm:"available_balance"     description:"可用余额"`
	UpdatedAt           *gtime.Time `json:"updatedAt"           orm:"updated_at"            description:"更新时间"`
}
