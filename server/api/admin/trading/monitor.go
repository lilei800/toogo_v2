// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package trading

import (
	"hotgo/internal/model/input"

	"github.com/gogf/gf/v2/frame/g"
)

// MonitorTickerReq 获取实时行情请求
type MonitorTickerReq struct {
	g.Meta `path:"/trading/monitor/ticker" method:"get" tags:"交易监控" summary:"获取实时行情" dc:"获取指定交易对的实时行情"`
	input.TradingMonitorTickerInp
}

// MonitorTickerRes 获取实时行情响应
type MonitorTickerRes struct {
	*input.TradingMonitorTickerModel
}

// MonitorMarketStateReq 获取市场状态请求
type MonitorMarketStateReq struct {
	g.Meta `path:"/trading/monitor/marketState" method:"get" tags:"交易监控" summary:"获取市场状态" dc:"分析市场状态并给出交易信号"`
	input.TradingMonitorMarketStateInp
}

// MonitorMarketStateRes 获取市场状态响应
type MonitorMarketStateRes struct {
	*input.TradingMonitorMarketStateModel
}

// MonitorLogsReq 监控日志列表请求
type MonitorLogsReq struct {
	g.Meta `path:"/trading/monitor/logs" method:"get" tags:"交易监控" summary:"监控日志列表" dc:"获取交易监控日志列表"`
	input.TradingMonitorLogListInp
}

// MonitorLogsRes 监控日志列表响应
type MonitorLogsRes struct {
	List       []*input.TradingMonitorLogListModel `json:"list" dc:"列表数据"`
	TotalCount int                                 `json:"totalCount" dc:"总数"`
	Page       int                                 `json:"page" dc:"当前页"`
	PageSize   int                                 `json:"pageSize" dc:"每页数量"`
}

// MonitorRobotAnalysisReq 获取机器人实时分析数据请求
type MonitorRobotAnalysisReq struct {
	g.Meta  `path:"/trading/monitor/robotAnalysis" method:"get" tags:"交易监控" summary:"获取机器人实时分析" dc:"获取机器人的实时行情和分析数据"`
	RobotId int64 `json:"robotId" v:"required" dc:"机器人ID"`
}

// MonitorRobotAnalysisRes 获取机器人实时分析数据响应
type MonitorRobotAnalysisRes struct {
	RobotId int64 `json:"robotId" dc:"机器人ID"`
	// 实时行情
	Ticker *RobotTickerInfo `json:"ticker" dc:"实时行情"`
	// 市场分析
	Market *RobotMarketAnalysis `json:"market" dc:"市场分析"`
	// 风险评估
	Risk *RobotRiskEvaluation `json:"risk" dc:"风险评估"`
	// 方向信号
	Signal *RobotDirectionSignal `json:"signal" dc:"方向信号"`
	// 账户信息
	Account *RobotAccountInfo `json:"account" dc:"账户信息"`
	// 机器人配置
	Config *RobotConfigInfo `json:"config" dc:"机器人配置"`
	// 连接状态
	Connected       bool   `json:"connected" dc:"是否已连接"`
	ConnectionError string `json:"connectionError" dc:"连接错误信息"`

	// ============ 窗口价格曲线数据（toogo实时信号逻辑）============
	PriceWindow   []*PricePoint        `json:"priceWindow" dc:"窗口价格数据"`
	SignalHistory []*SignalHistoryItem `json:"signalHistory" dc:"信号历史"`

	// ============ 多周期市场状态实时明细（新算法 + 平滑机制）============
	MarketStateRealtime *RobotMarketStateRealtime `json:"marketStateRealtime" dc:"多周期市场状态实时明细（含平滑后状态与播报）"`
}

// RobotMarketStateRealtime 多周期市场状态实时数据（用于前端机器人列表展示/播报）
type RobotMarketStateRealtime struct {
	Platform string `json:"platform" dc:"平台"`
	Symbol   string `json:"symbol" dc:"交易对"`

	// 最终市场状态（平滑后）
	State      string  `json:"state" dc:"最终市场状态(low_vol/volatile/high_vol/trend)"`
	Confidence float64 `json:"confidence" dc:"最终市场状态置信度(0-1)"`
	VoteRatio  float64 `json:"voteRatio" dc:"投票占比(0-1)，最大投票权重/总权重"`
	UpdatedAt  string  `json:"updatedAt" dc:"更新时间(本地格式)"`

	// 各周期明细（1m/5m/15m/30m/1h）
	Timeframes []*RobotMarketStateTimeframe `json:"timeframes" dc:"各周期实时明细"`

	// 播报摘要（可直接展示/复制）
	Broadcast string `json:"broadcast" dc:"播报摘要"`
}

// RobotMarketStateTimeframe 单周期实时明细
type RobotMarketStateTimeframe struct {
	Interval string  `json:"interval" dc:"周期"`
	Weight   float64 `json:"weight" dc:"权重"`

	Open  float64 `json:"open" dc:"开盘价"`
	High  float64 `json:"high" dc:"最高价"`
	Low   float64 `json:"low" dc:"最低价"`
	Close float64 `json:"close" dc:"收盘价"`

	Delta float64 `json:"delta" dc:"delta"`
	V     float64 `json:"v" dc:"V=(H-L)/delta"`
	D     float64 `json:"d" dc:"方向一致性D(0-1)"`

	RawState      string  `json:"rawState" dc:"原始状态(未平滑)"`
	SmoothedState string  `json:"smoothedState" dc:"平滑后状态(参与投票)"`
	SmoothedConf  float64 `json:"smoothedConf" dc:"平滑置信度(0-1)"`
}

// RobotAccountInfo 机器人账户信息
type RobotAccountInfo struct {
	AccountEquity    float64 `json:"accountEquity" dc:"账户权益(USDT)，包含未实现盈亏"`
	WalletBalance    float64 `json:"walletBalance" dc:"钱包余额(USDT)，不包含未实现盈亏"`
	AvailableBalance float64 `json:"availableBalance" dc:"可用余额(USDT)"`
	UsedMargin       float64 `json:"usedMargin" dc:"已用保证金(USDT)"`
	UnrealizedPnl    float64 `json:"unrealizedPnl" dc:"未实现盈亏(USDT)"`
	TodayPnl         float64 `json:"todayPnl" dc:"今日盈亏(USDT)"`
	MarginRatio      float64 `json:"marginRatio" dc:"保证金率(%)"`
	// 兼容旧字段名
	TotalBalance float64 `json:"totalBalance" dc:"账户权益(USDT)，与accountEquity相同"`
}

// RobotConfigInfo 机器人配置信息
type RobotConfigInfo struct {
	AutoTradeEnabled  bool    `json:"autoTradeEnabled" dc:"自动下单"`
	AutoCloseEnabled  bool    `json:"autoCloseEnabled" dc:"自动平仓"`
	DualSidePosition  bool    `json:"dualSidePosition" dc:"双向开单"`
	UseMonitorSignal  bool    `json:"useMonitorSignal" dc:"信号监控"`
	RiskPreference    string  `json:"riskPreference" dc:"风险偏好"`
	MarketState       string  `json:"marketState" dc:"市场状态"`
	Leverage          int     `json:"leverage" dc:"杠杆倍数"`
	MarginPercent     float64 `json:"marginPercent" dc:"保证金比例"`
	StopLossPercent   float64 `json:"stopLossPercent" dc:"止损比例"`
	AutoStartRetreat  float64 `json:"autoStartRetreat" dc:"启动止盈百分比"`
	TakeProfitPercent float64 `json:"takeProfitPercent" dc:"止盈回撤"`
	MaxProfit         float64 `json:"maxProfit" dc:"最大盈利目标"`
	MaxLoss           float64 `json:"maxLoss" dc:"最大亏损限制"`
	RuntimeSeconds    int64   `json:"runtimeSeconds" dc:"运行时长(秒)"`
	StartTime         string  `json:"startTime" dc:"启动时间"`
	TotalProfit       float64 `json:"totalProfit" dc:"累计盈亏"`
	LongCount         int     `json:"longCount" dc:"做多次数"`
	ShortCount        int     `json:"shortCount" dc:"做空次数"`
	TimeWindow        int     `json:"timeWindow" dc:"时间窗口(秒)"`
	Threshold         float64 `json:"threshold" dc:"波动阈值(USDT)"`
	ErrorMessage      string  `json:"errorMessage" dc:"错误信息（策略模板加载失败时显示）"`
	StrategyGroupId   int64   `json:"strategyGroupId" dc:"策略组ID"`
	StrategyGroupName string  `json:"strategyGroupName" dc:"策略组名称"`
	StrategyName      string  `json:"strategyName" dc:"策略模板名称"`
}

// RobotTickerInfo 机器人行情信息
type RobotTickerInfo struct {
	Symbol        string  `json:"symbol" dc:"交易对"`
	LastPrice     float64 `json:"lastPrice" dc:"最新价格"`
	High24h       float64 `json:"high24h" dc:"24小时最高价"`
	Low24h        float64 `json:"low24h" dc:"24小时最低价"`
	Volume24h     float64 `json:"volume24h" dc:"24小时成交量"`
	Change24h     float64 `json:"change24h" dc:"24小时涨跌幅"`
	ChangePercent float64 `json:"changePercent" dc:"24小时涨跌百分比"`
	Timestamp     string  `json:"timestamp" dc:"更新时间"`
}

// RobotMarketAnalysis 机器人市场分析
type RobotMarketAnalysis struct {
	State            string            `json:"state" dc:"市场状态(UPTREND/DOWNTREND/RANGING/HIGH_VOLATILITY/LOW_VOLATILITY)"`
	TrendScore       float64           `json:"trendScore" dc:"趋势评分(-100到100)"`
	VolatilityLevel  string            `json:"volatilityLevel" dc:"波动等级(LOW/NORMAL/HIGH/EXTREME)"`
	Confidence       float64           `json:"confidence" dc:"分析置信度(0-1)"`
	SuggestAction    string            `json:"suggestAction" dc:"建议操作(BUY/SELL/WAIT/CAUTION)"`
	TimeFrameSignals map[string]string `json:"timeFrameSignals" dc:"多周期信号"`
}

// RobotRiskEvaluation 机器人风险评估
type RobotRiskEvaluation struct {
	PreferenceType  string   `json:"preferenceType" dc:"风险偏好(conservative/balanced/aggressive)"`
	WinProbability  float64  `json:"winProbability" dc:"胜算概率(0-100)"`
	SuggestLeverage int      `json:"suggestLeverage" dc:"建议杠杆"`
	SuggestPosition float64  `json:"suggestPosition" dc:"建议仓位比例(0-1)"`
	SuggestStopLoss float64  `json:"suggestStopLoss" dc:"建议止损百分比"`
	Reasons         []string `json:"reasons" dc:"风险提示原因"`
}

// RobotDirectionSignal 机器人方向信号
type RobotDirectionSignal struct {
	Direction       string  `json:"direction" dc:"方向(LONG/SHORT/WAIT)"`
	SignalStrength  float64 `json:"signalStrength" dc:"信号强度(0-100)"`
	Confidence      float64 `json:"confidence" dc:"置信度(0-1)"`
	RiskRewardRatio float64 `json:"riskRewardRatio" dc:"风险收益比"`
	TimeWindow      string  `json:"timeWindow" dc:"建议持仓时间"`
	Recommendation  string  `json:"recommendation" dc:"操作建议"`

	// ============ 窗口信号相关（toogo实时信号逻辑）============
	SignalType      string  `json:"signalType" dc:"信号类型(window/analysis)"`
	WindowMaxPrice  float64 `json:"windowMaxPrice" dc:"窗口最高价"`
	WindowMinPrice  float64 `json:"windowMinPrice" dc:"窗口最低价"`
	CurrentPrice    float64 `json:"currentPrice" dc:"当前价格"`
	DistanceFromMin float64 `json:"distanceFromMin" dc:"距最低价距离"`
	DistanceFromMax float64 `json:"distanceFromMax" dc:"距最高价距离"`
	SignalThreshold float64 `json:"signalThreshold" dc:"信号阈值"`
	SignalProgress  float64 `json:"signalProgress" dc:"信号进度(0-100)"`
	Action          string  `json:"action" dc:"建议操作(OPEN_LONG/OPEN_SHORT/CLOSE_LONG/CLOSE_SHORT/HOLD)"`
	Reason          string  `json:"reason" dc:"信号原因"`

	// 窗口监控配置
	MonitorWindow    int `json:"monitorWindow" dc:"监控窗口(秒)"`
	PricePointsCount int `json:"pricePointsCount" dc:"价格数据点数"`

	// 策略配置（从策略模板加载）
	StrategyWindow     int     `json:"strategyWindow" dc:"策略时间窗口(秒)"`
	StrategyThreshold  float64 `json:"strategyThreshold" dc:"策略波动阈值(USDT)"`
	CurrentMarketState string  `json:"currentMarketState" dc:"当前市场状态"`
	CurrentRiskPref    string  `json:"currentRiskPref" dc:"当前风险偏好"`
}

// PricePoint 价格数据点
type PricePoint struct {
	Timestamp int64   `json:"timestamp" dc:"时间戳(毫秒)"`
	Price     float64 `json:"price" dc:"价格"`
}

// SignalHistoryItem 信号历史项
type SignalHistoryItem struct {
	Timestamp int64  `json:"timestamp" dc:"时间戳(毫秒)"`
	Signal    string `json:"signal" dc:"信号方向(long/short/neutral)"`
}

// MonitorBatchRobotAnalysisReq 批量获取机器人实时分析数据请求
type MonitorBatchRobotAnalysisReq struct {
	g.Meta   `path:"/trading/monitor/batchRobotAnalysis" method:"get" tags:"交易监控" summary:"批量获取机器人实时分析" dc:"批量获取多个机器人的实时行情和分析数据"`
	RobotIds string `json:"robotIds" v:"required" dc:"机器人ID列表(逗号分隔)"`
}

// MonitorBatchRobotAnalysisRes 批量获取机器人实时分析数据响应
type MonitorBatchRobotAnalysisRes struct {
	List []*MonitorRobotAnalysisRes `json:"list" dc:"机器人分析数据列表"`
}

// MonitorKlineReq 获取K线数据请求
type MonitorKlineReq struct {
	g.Meta      `path:"/trading/monitor/kline" method:"get" tags:"交易监控" summary:"获取K线数据" dc:"获取指定交易对的K线数据"`
	ApiConfigId int64  `json:"apiConfigId" v:"required" dc:"API配置ID"`
	Symbol      string `json:"symbol" v:"required" dc:"交易对"`
	Interval    string `json:"interval" dc:"K线周期(1m/5m/15m/30m/1h/4h/1d)"`
	Limit       int    `json:"limit" dc:"数据条数"`
}

// MonitorKlineRes 获取K线数据响应
type MonitorKlineRes struct {
	Symbol   string           `json:"symbol" dc:"交易对"`
	Interval string           `json:"interval" dc:"K线周期"`
	List     []*KlineDataItem `json:"list" dc:"K线数据列表"`
}

// KlineDataItem K线数据项
type KlineDataItem struct {
	OpenTime  int64   `json:"openTime" dc:"开盘时间戳"`
	Open      float64 `json:"open" dc:"开盘价"`
	High      float64 `json:"high" dc:"最高价"`
	Low       float64 `json:"low" dc:"最低价"`
	Close     float64 `json:"close" dc:"收盘价"`
	Volume    float64 `json:"volume" dc:"成交量"`
	CloseTime int64   `json:"closeTime" dc:"收盘时间戳"`
}
