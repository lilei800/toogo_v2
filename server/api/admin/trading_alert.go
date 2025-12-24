// Package admin 预警日志API
package admin

import (
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/form"

	"github.com/gogf/gf/v2/frame/g"
)

// ========== 市场状态预警日志 ==========

// MarketStateLogListReq 市场状态预警日志列表请求
type MarketStateLogListReq struct {
	g.Meta   `path:"/trading/alert/marketState/list" method:"get" tags:"预警日志" summary:"市场状态预警列表"`
	Platform string `json:"platform" dc:"交易所平台"`
	Symbol   string `json:"symbol" dc:"交易对"`
	NewState string `json:"newState" dc:"市场状态"`
	form.PageReq
}

type MarketStateLogListRes struct {
	List  []*entity.TradingMarketStateLog `json:"list" dc:"列表"`
	Total int                             `json:"total" dc:"总数"`
	form.PageRes
}

// ========== 风险偏好预警日志 ==========

// RiskPreferenceLogListReq 风险偏好预警日志列表请求
type RiskPreferenceLogListReq struct {
	g.Meta        `path:"/trading/alert/riskPreference/list" method:"get" tags:"预警日志" summary:"风险偏好预警列表"`
	RobotId       int64  `json:"robotId" dc:"机器人ID"`
	UserId        int64  `json:"userId" dc:"用户ID"`
	NewPreference string `json:"newPreference" dc:"风险偏好"`
	form.PageReq
}

type RiskPreferenceLogListRes struct {
	List  []*entity.TradingRiskPreferenceLog `json:"list" dc:"列表"`
	Total int                                `json:"total" dc:"总数"`
	form.PageRes
}

// ========== 方向预警日志 ==========

// DirectionLogListReq 方向预警日志列表请求
type DirectionLogListReq struct {
	g.Meta     `path:"/trading/alert/direction/list" method:"get" tags:"预警日志" summary:"方向预警列表"`
	RobotId    int64  `json:"robotId" dc:"机器人ID"`
	Symbol     string `json:"symbol" dc:"交易对"`
	SignalType string `json:"signalType" dc:"信号类型(long/short)"`
	form.PageReq
}

// DirectionLogItem 方向预警日志项
type DirectionLogItem struct {
	Id             int64   `json:"id" dc:"ID"`
	RobotId        int64   `json:"robotId" dc:"机器人ID"`
	StrategyId     int64   `json:"strategyId" dc:"策略ID"`
	Symbol         string  `json:"symbol" dc:"交易对"`
	SignalType     string  `json:"signalType" dc:"信号类型(long/short)"`
	SignalSource   string  `json:"signalSource" dc:"信号来源"`
	SignalStrength float64 `json:"signalStrength" dc:"信号强度"`
	CurrentPrice   float64 `json:"currentPrice" dc:"当前价格"`
	WindowMinPrice float64 `json:"windowMinPrice" dc:"窗口最低价"`
	WindowMaxPrice float64 `json:"windowMaxPrice" dc:"窗口最高价"`
	Threshold      float64 `json:"threshold" dc:"波动阈值"`
	TargetPrice    float64 `json:"targetPrice" dc:"目标价格"`
	StopLoss       float64 `json:"stopLoss" dc:"止损价"`
	TakeProfit     float64 `json:"takeProfit" dc:"止盈价"`
	Executed       int     `json:"executed" dc:"是否执行"`
	ExecuteResult  string  `json:"executeResult" dc:"执行结果"`
	Reason         string  `json:"reason" dc:"原因"`
	MarketState    string  `json:"marketState" dc:"市场状态"`
	RiskPreference string  `json:"riskPreference" dc:"风险偏好"`
	CreatedAt      string  `json:"createdAt" dc:"创建时间"`
}

type DirectionLogListRes struct {
	List  []*DirectionLogItem `json:"list" dc:"列表"`
	Total int                 `json:"total" dc:"总数"`
	form.PageRes
}

// ========== 机器人实时状态 ==========

// RobotRealtimeReq 获取机器人实时状态请求
type RobotRealtimeReq struct {
	g.Meta  `path:"/trading/alert/robot/realtime" method:"get" tags:"预警日志" summary:"获取机器人实时状态"`
	RobotId int64 `json:"robotId" v:"required#机器人ID不能为空" dc:"机器人ID"`
}

type RobotRealtimeRes struct {
	*entity.TradingRobotRealtime
}

// RobotRealtimeListReq 机器人实时状态列表请求
type RobotRealtimeListReq struct {
	g.Meta `path:"/trading/alert/robot/realtimeList" method:"get" tags:"预警日志" summary:"机器人实时状态列表"`
	UserId int64 `json:"userId" dc:"用户ID"`
	form.PageReq
}

type RobotRealtimeListRes struct {
	List  []*entity.TradingRobotRealtime `json:"list" dc:"列表"`
	Total int                            `json:"total" dc:"总数"`
	form.PageRes
}

// ========== 市场分析 ==========

// MarketAnalysisReq 获取市场分析请求
type MarketAnalysisReq struct {
	g.Meta   `path:"/trading/alert/market/analysis" method:"get" tags:"预警日志" summary:"获取市场分析"`
	Platform string `json:"platform" v:"required#交易所不能为空" dc:"交易所平台"`
	Symbol   string `json:"symbol" v:"required#交易对不能为空" dc:"交易对"`
}

type MarketAnalysisRes struct {
	Platform        string                 `json:"platform" dc:"交易所"`
	Symbol          string                 `json:"symbol" dc:"交易对"`
	CurrentPrice    float64                `json:"currentPrice" dc:"当前价格"`
	MarketState     string                 `json:"marketState" dc:"市场状态"`
	MarketStateConf float64                `json:"marketStateConf" dc:"置信度"`
	TrendStrength   float64                `json:"trendStrength" dc:"趋势强度"`
	Volatility      float64                `json:"volatility" dc:"波动率"`
	SupportLevel    float64                `json:"supportLevel" dc:"支撑位"`
	ResistanceLevel float64                `json:"resistanceLevel" dc:"阻力位"`
	Indicators      map[string]interface{} `json:"indicators" dc:"技术指标"`
	TimeframeData   map[string]interface{} `json:"timeframeData" dc:"多周期数据"`
}

// DirectionSignalReq 获取方向信号请求
type DirectionSignalReq struct {
	g.Meta   `path:"/trading/alert/direction/signal" method:"get" tags:"预警日志" summary:"获取方向信号"`
	Platform string `json:"platform" v:"required#交易所不能为空" dc:"交易所平台"`
	Symbol   string `json:"symbol" v:"required#交易对不能为空" dc:"交易对"`
}

type DirectionSignalRes struct {
	Platform         string                 `json:"platform" dc:"交易所"`
	Symbol           string                 `json:"symbol" dc:"交易对"`
	Direction        string                 `json:"direction" dc:"方向"`
	Strength         float64                `json:"strength" dc:"强度"`
	Confidence       float64                `json:"confidence" dc:"置信度"`
	Action           string                 `json:"action" dc:"建议操作"`
	EntryPrice       float64                `json:"entryPrice" dc:"入场价"`
	StopLoss         float64                `json:"stopLoss" dc:"止损价"`
	TakeProfit1      float64                `json:"takeProfit1" dc:"止盈目标1"`
	TakeProfit2      float64                `json:"takeProfit2" dc:"止盈目标2"`
	Reason           string                 `json:"reason" dc:"原因"`
	TimeframeSignals map[string]interface{} `json:"timeframeSignals" dc:"各周期信号"`
}

// RiskEvaluationReq 获取风险评估请求
type RiskEvaluationReq struct {
	g.Meta  `path:"/trading/alert/risk/evaluation" method:"get" tags:"预警日志" summary:"获取风险评估"`
	RobotId int64 `json:"robotId" v:"required#机器人ID不能为空" dc:"机器人ID"`
}

type RiskEvaluationRes struct {
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
	SuggestedStopLoss      float64 `json:"suggestedStopLoss" dc:"建议止损"`
	SuggestedTakeProfit    float64 `json:"suggestedTakeProfit" dc:"建议止盈"`
	RiskLevel              int     `json:"riskLevel" dc:"风险等级"`
	Reason                 string  `json:"reason" dc:"评估原因"`
}

// EngineStatusReq 获取引擎状态请求
type EngineStatusReq struct {
	g.Meta `path:"/trading/alert/engine/status" method:"get" tags:"全局引擎" summary:"获取引擎状态"`
}

type EngineStatusRes struct {
	Running             bool `json:"running" dc:"是否运行中"`
	ActiveRobots        int  `json:"activeRobots" dc:"活跃机器人数"`
	ActiveSubscriptions int  `json:"activeSubscriptions" dc:"活跃订阅数"`
}

// ========== 全局引擎管理 ==========

// GlobalEngineDetailReq 获取全局引擎详情请求
type GlobalEngineDetailReq struct {
	g.Meta `path:"/trading/alert/engine/detail" method:"get" tags:"全局引擎" summary:"获取全局引擎详情"`
}

type GlobalEngineDetailRes struct {
	// 基础状态
	Running   bool   `json:"running" dc:"是否运行中"`
	StartTime string `json:"startTime" dc:"启动时间"`
	Uptime    int64  `json:"uptime" dc:"运行时长(秒)"`

	// MarketDataService 行情数据服务
	MarketDataService *MarketDataServiceStatus `json:"marketDataService" dc:"行情数据服务"`

	// MarketAnalyzer 市场分析引擎
	MarketAnalyzer *MarketAnalyzerStatus `json:"marketAnalyzer" dc:"市场分析引擎"`

	// DirectionSignalService 方向信号服务
	DirectionSignalService *DirectionSignalServiceStatus `json:"directionSignalService" dc:"方向信号服务"`

	// RobotTaskManager 机器人任务管理器
	RobotTaskManager *RobotTaskManagerStatus `json:"robotTaskManager" dc:"机器人任务管理器"`

	// AlertLogger 预警日志服务
	AlertLogger *AlertLoggerStatus `json:"alertLogger" dc:"预警日志服务"`

	// TradeStatistics 交易统计服务
	TradeStatistics *TradeStatisticsStatus `json:"tradeStatistics" dc:"交易统计服务"`
}

// MarketDataServiceStatus 行情数据服务状态
type MarketDataServiceStatus struct {
	Running       bool                  `json:"running" dc:"是否运行中"`
	Subscriptions int                   `json:"subscriptions" dc:"订阅数量"`
	TickerCount   int                   `json:"tickerCount" dc:"行情缓存数"`
	KlineCount    int                   `json:"klineCount" dc:"K线缓存数"`
	TickerList    []*SubscriptionDetail `json:"tickerList" dc:"订阅列表"`
}

// SubscriptionDetail 订阅详情
type SubscriptionDetail struct {
	Platform   string  `json:"platform" dc:"交易所"`
	Symbol     string  `json:"symbol" dc:"交易对"`
	LastPrice  float64 `json:"lastPrice" dc:"最新价"`
	Change24h  float64 `json:"change24h" dc:"24H涨跌"`
	RefCount   int     `json:"refCount" dc:"引用计数"`
	LastUpdate string  `json:"lastUpdate" dc:"最后更新"`
	DataFresh  bool    `json:"dataFresh" dc:"数据新鲜"`
}

// MarketAnalyzerStatus 市场分析引擎状态
type MarketAnalyzerStatus struct {
	Running       bool              `json:"running" dc:"是否运行中"`
	AnalysisCount int               `json:"analysisCount" dc:"分析数据数"`
	AnalysisList  []*AnalysisDetail `json:"analysisList" dc:"分析列表"`
}

// AnalysisDetail 分析详情
type AnalysisDetail struct {
	Platform      string  `json:"platform" dc:"交易所"`
	Symbol        string  `json:"symbol" dc:"交易对"`
	MarketState   string  `json:"marketState" dc:"市场状态"`
	TrendStrength float64 `json:"trendStrength" dc:"趋势强度"`
	Volatility    float64 `json:"volatility" dc:"波动率"`
	LastUpdate    string  `json:"lastUpdate" dc:"最后更新"`
}

// DirectionSignalServiceStatus 方向信号服务状态
type DirectionSignalServiceStatus struct {
	Running     bool            `json:"running" dc:"是否运行中"`
	SignalCount int             `json:"signalCount" dc:"信号数量"`
	SignalList  []*SignalDetail `json:"signalList" dc:"信号列表"`
}

// SignalDetail 信号详情
type SignalDetail struct {
	Platform   string  `json:"platform" dc:"交易所"`
	Symbol     string  `json:"symbol" dc:"交易对"`
	Direction  string  `json:"direction" dc:"方向"`
	Strength   float64 `json:"strength" dc:"强度"`
	Confidence float64 `json:"confidence" dc:"置信度"`
	Action     string  `json:"action" dc:"建议操作"`
	LastUpdate string  `json:"lastUpdate" dc:"最后更新"`
}

// RobotTaskManagerStatus 机器人任务管理器状态
type RobotTaskManagerStatus struct {
	Running      bool                  `json:"running" dc:"是否运行中"`
	ActiveRobots int                   `json:"activeRobots" dc:"活跃机器人数"`
	RobotList    []*ManagedRobotDetail `json:"robotList" dc:"机器人列表"`
}

// ManagedRobotDetail 被管理机器人详情
type ManagedRobotDetail struct {
	RobotId         int64   `json:"robotId" dc:"机器人ID"`
	RobotName       string  `json:"robotName" dc:"机器人名称"`
	Platform        string  `json:"platform" dc:"交易所"`
	Symbol          string  `json:"symbol" dc:"交易对"`
	Connected       bool    `json:"connected" dc:"连接状态(行情)"`
	ApiConnected    bool    `json:"apiConnected" dc:"API连接状态"`
	ApiError        string  `json:"apiError" dc:"API错误信息"`
	HasPosition     bool    `json:"hasPosition" dc:"是否持仓"`
	PositionSide    string  `json:"positionSide" dc:"持仓方向"`
	PositionAmt     float64 `json:"positionAmt" dc:"持仓数量"`
	EntryPrice      float64 `json:"entryPrice" dc:"开仓均价"`
	UnrealizedPnl   float64 `json:"unrealizedPnl" dc:"未实现盈亏"`
	TotalBalance    float64 `json:"totalBalance" dc:"账户总余额"`
	AvailBalance    float64 `json:"availBalance" dc:"可用余额"`
	RiskPreference  string  `json:"riskPreference" dc:"风险偏好"`
	WinProbability  float64 `json:"winProbability" dc:"胜算概率"`
	DirectionSignal string  `json:"directionSignal" dc:"方向信号"`
	LastRiskEval    string  `json:"lastRiskEval" dc:"上次风险评估时间"`
	LastUpdate      string  `json:"lastUpdate" dc:"最后更新"`
}

// AlertLoggerStatus 预警日志服务状态
type AlertLoggerStatus struct {
	Running            bool  `json:"running" dc:"是否运行中"`
	MarketStateLogs    int64 `json:"marketStateLogs" dc:"市场状态日志数"`
	RiskPreferenceLogs int64 `json:"riskPreferenceLogs" dc:"风险偏好日志数"`
	DirectionLogs      int64 `json:"directionLogs" dc:"方向日志数"`
}

// TradeStatisticsStatus 交易统计服务状态
type TradeStatisticsStatus struct {
	Running     bool    `json:"running" dc:"是否运行中"`
	TotalTrades int     `json:"totalTrades" dc:"总交易数"`
	TodayTrades int     `json:"todayTrades" dc:"今日交易数"`
	WinRate     float64 `json:"winRate" dc:"胜率"`
	TotalProfit float64 `json:"totalProfit" dc:"总盈亏"`
	TodayProfit float64 `json:"todayProfit" dc:"今日盈亏"`
}

// GlobalEngineStartReq 启动全局引擎请求
type GlobalEngineStartReq struct {
	g.Meta `path:"/trading/alert/engine/start" method:"post" tags:"全局引擎" summary:"启动全局引擎"`
}

type GlobalEngineStartRes struct {
	Success bool   `json:"success" dc:"是否成功"`
	Message string `json:"message" dc:"消息"`
}

// GlobalEngineStopReq 停止全局引擎请求
type GlobalEngineStopReq struct {
	g.Meta `path:"/trading/alert/engine/stop" method:"post" tags:"全局引擎" summary:"停止全局引擎"`
}

type GlobalEngineStopRes struct {
	Success bool   `json:"success" dc:"是否成功"`
	Message string `json:"message" dc:"消息"`
}

// GlobalEngineRestartReq 重启全局引擎请求
type GlobalEngineRestartReq struct {
	g.Meta `path:"/trading/alert/engine/restart" method:"post" tags:"全局引擎" summary:"重启全局引擎"`
}

type GlobalEngineRestartRes struct {
	Success bool   `json:"success" dc:"是否成功"`
	Message string `json:"message" dc:"消息"`
}
