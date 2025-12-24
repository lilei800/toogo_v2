// Package trading 预警日志API
package trading

import (
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/form"

	"github.com/gogf/gf/v2/frame/g"
)

// AlertLogListReq 预警日志列表请求
type AlertLogListReq struct {
	g.Meta   `path:"/alertLog/list" method:"get" tags:"交易-预警日志" summary:"预警日志列表"`
	Type     string `json:"type" dc:"日志类型: market_state/risk_preference/direction"`
	Platform string `json:"platform" dc:"交易所平台"`
	Symbol   string `json:"symbol" dc:"交易对"`
	RobotId  int64  `json:"robotId" dc:"机器人ID（仅风险偏好日志）"`
	form.PageReq
}

type AlertLogListRes struct {
	List  interface{} `json:"list" dc:"列表数据"`
	Page  int         `json:"page" dc:"当前页码"`
	Total int         `json:"total" dc:"总数"`
}

// MarketAnalysisReq 市场分析数据请求
type MarketAnalysisReq struct {
	g.Meta   `path:"/alertLog/marketAnalysis" method:"get" tags:"交易-预警日志" summary:"市场分析数据"`
	Platform string `json:"platform" v:"required" dc:"交易所平台"`
	Symbol   string `json:"symbol" v:"required" dc:"交易对"`
}

type MarketAnalysisRes struct {
	Platform        string                    `json:"platform" dc:"交易所平台"`
	Symbol          string                    `json:"symbol" dc:"交易对"`
	CurrentPrice    float64                   `json:"currentPrice" dc:"当前价格"`
	MarketState     string                    `json:"marketState" dc:"市场状态"`
	MarketStateConf float64                   `json:"marketStateConf" dc:"置信度"`
	TrendStrength   float64                   `json:"trendStrength" dc:"趋势强度"`
	Volatility      float64                   `json:"volatility" dc:"波动率"`
	SupportLevel    float64                   `json:"supportLevel" dc:"支撑位"`
	ResistanceLevel float64                   `json:"resistanceLevel" dc:"阻力位"`
	Indicators      *TechnicalIndicators      `json:"indicators" dc:"技术指标"`
	TimeframeData   map[string]*TimeframeInfo `json:"timeframeData" dc:"多周期数据"`
}

type TechnicalIndicators struct {
	TrendScore      float64 `json:"trendScore" dc:"趋势评分"`
	VolatilityScore float64 `json:"volatilityScore" dc:"波动评分"`
}

type TimeframeInfo struct {
	Interval      string  `json:"interval" dc:"周期"`
	Trend         string  `json:"trend" dc:"趋势"`
	TrendStrength float64 `json:"trendStrength" dc:"趋势强度"`
	MACD          float64 `json:"macd" dc:"MACD"`
	EMA12         float64 `json:"ema12" dc:"EMA12"`
	EMA26         float64 `json:"ema26" dc:"EMA26"`
}

// DirectionSignalReq 方向信号请求
type DirectionSignalReq struct {
	g.Meta   `path:"/alertLog/directionSignal" method:"get" tags:"交易-预警日志" summary:"方向信号"`
	Platform string `json:"platform" v:"required" dc:"交易所平台"`
	Symbol   string `json:"symbol" v:"required" dc:"交易对"`
}

type DirectionSignalRes struct {
	Platform   string  `json:"platform" dc:"交易所平台"`
	Symbol     string  `json:"symbol" dc:"交易对"`
	Direction  string  `json:"direction" dc:"方向"`
	Strength   float64 `json:"strength" dc:"信号强度"`
	Confidence float64 `json:"confidence" dc:"置信度"`
	Action     string  `json:"action" dc:"建议操作"`
	EntryPrice float64 `json:"entryPrice" dc:"入场价"`
	StopLoss   float64 `json:"stopLoss" dc:"止损价"`
	Reason     string  `json:"reason" dc:"原因"`
	// ============ 窗口信号数据（toogo实时信号逻辑） ============
	SignalType      string  `json:"signalType" dc:"信号类型: window/analysis"`
	WindowMaxPrice  float64 `json:"windowMaxPrice" dc:"窗口最高价"`
	WindowMinPrice  float64 `json:"windowMinPrice" dc:"窗口最低价"`
	CurrentPrice    float64 `json:"currentPrice" dc:"当前价格"`
	DistanceFromMin float64 `json:"distanceFromMin" dc:"距最低价距离"`
	DistanceFromMax float64 `json:"distanceFromMax" dc:"距最高价距离"`
	SignalThreshold float64 `json:"signalThreshold" dc:"信号阈值"`
	SignalProgress  float64 `json:"signalProgress" dc:"信号进度百分比"`
}

// RobotRiskEvalReq 机器人风险评估请求
type RobotRiskEvalReq struct {
	g.Meta  `path:"/alertLog/robotRiskEval" method:"get" tags:"交易-预警日志" summary:"机器人风险评估"`
	RobotId int64 `json:"robotId" v:"required" dc:"机器人ID"`
}

type RobotRiskEvalRes struct {
	RobotId                int64   `json:"robotId" dc:"机器人ID"`
	WinProbability         float64 `json:"winProbability" dc:"胜算概率"`
	RiskPreference         string  `json:"riskPreference" dc:"风险偏好"`
	MarketScore            float64 `json:"marketScore" dc:"市场评分"`
	TechnicalScore         float64 `json:"technicalScore" dc:"技术评分"`
	AccountScore           float64 `json:"accountScore" dc:"账户评分"`
	HistoryScore           float64 `json:"historyScore" dc:"历史评分"`
	VolatilityRisk         float64 `json:"volatilityRisk" dc:"波动风险"`
	SuggestedLeverage      int     `json:"suggestedLeverage" dc:"建议杠杆"`
	SuggestedMarginPercent float64 `json:"suggestedMarginPercent" dc:"建议保证金比例"`
	Reason                 string  `json:"reason" dc:"评估说明"`
}

// RobotStatusReq 机器人实时状态请求
type RobotStatusReq struct {
	g.Meta  `path:"/alertLog/robotStatus" method:"get" tags:"交易-预警日志" summary:"机器人实时状态"`
	RobotId int64 `json:"robotId" v:"required" dc:"机器人ID"`
}

type RobotStatusRes struct {
	RobotId          int64               `json:"robotId" dc:"机器人ID"`
	Symbol           string              `json:"symbol" dc:"交易对"`
	Status           int                 `json:"status" dc:"机器人状态"`
	CurrentPrice     float64             `json:"currentPrice" dc:"当前价格"`
	AccountBalance   float64             `json:"accountBalance" dc:"账户余额"`
	AvailableBalance float64             `json:"availableBalance" dc:"可用余额"`
	Positions        []*PositionInfo     `json:"positions" dc:"持仓列表"`
	RiskEvaluation   *RobotRiskEvalRes   `json:"riskEvaluation" dc:"风险评估"`
	DirectionSignal  *DirectionSignalRes `json:"directionSignal" dc:"方向信号"`
	MarketState      string              `json:"marketState" dc:"市场状态"`
}

type PositionInfo struct {
	PositionSide   string  `json:"positionSide" dc:"持仓方向"`
	PositionAmt    float64 `json:"positionAmt" dc:"持仓数量"`
	EntryPrice     float64 `json:"entryPrice" dc:"开仓价格"`
	MarkPrice      float64 `json:"markPrice" dc:"标记价格"`
	UnrealizedPnl  float64 `json:"unrealizedPnl" dc:"未实现盈亏"`
	Leverage       int     `json:"leverage" dc:"杠杆"`
	IsolatedMargin float64 `json:"isolatedMargin" dc:"保证金"`
}

// EngineStatusReq 引擎状态请求
type EngineStatusReq struct {
	g.Meta `path:"/alertLog/engineStatus" method:"get" tags:"交易-预警日志" summary:"引擎状态"`
}

type EngineStatusRes struct {
	Running             bool   `json:"running" dc:"是否运行中"`
	ActiveRobots        int    `json:"activeRobots" dc:"活跃机器人数"`
	ActiveSubscriptions int    `json:"activeSubscriptions" dc:"活跃订阅数"`
	EngineVersion       string `json:"engineVersion" dc:"引擎版本"`
}

// MarketStateLogListModel 市场状态日志列表项
type MarketStateLogListModel struct {
	*entity.TradingMarketStateLog
}

// RiskPreferenceLogListModel 风险偏好日志列表项
type RiskPreferenceLogListModel struct {
	*entity.TradingRiskPreferenceLog
}

// DirectionLogListModel 方向日志列表项
type DirectionLogListModel struct {
	*entity.TradingDirectionLog
}
