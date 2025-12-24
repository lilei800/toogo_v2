// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package trading

import (
	"hotgo/internal/model/input"
	"hotgo/internal/model/input/toogoin"

	"github.com/gogf/gf/v2/frame/g"
)

// RobotListReq 机器人列表请求
type RobotListReq struct {
	g.Meta `path:"/trading/robot/list" method:"get" tags:"交易管理" summary:"机器人列表" dc:"获取机器人列表"`
	input.TradingRobotListInp
}

// RobotListRes 机器人列表响应
type RobotListRes struct {
	List       []*input.TradingRobotListModel `json:"list" dc:"列表数据"`
	TotalCount int                            `json:"totalCount" dc:"总数"`
	Page       int                            `json:"page" dc:"当前页"`
	PageSize   int                            `json:"pageSize" dc:"每页数量"`
}

// RobotCreateReq 创建机器人请求
type RobotCreateReq struct {
	g.Meta `path:"/trading/robot/create" method:"post" tags:"交易管理" summary:"创建机器人" dc:"创建新的交易机器人"`
	input.TradingRobotCreateInp
}

// RobotCreateRes 创建机器人响应
type RobotCreateRes struct {
	Id int64 `json:"id" dc:"机器人ID"`
}

// RobotUpdateReq 更新机器人请求
type RobotUpdateReq struct {
	g.Meta `path:"/trading/robot/update" method:"post" tags:"交易管理" summary:"更新机器人" dc:"更新机器人配置"`
	input.TradingRobotUpdateInp
}

// RobotUpdateRes 更新机器人响应
type RobotUpdateRes struct{}

// RobotDeleteReq 删除机器人请求
type RobotDeleteReq struct {
	g.Meta `path:"/trading/robot/delete" method:"post" tags:"交易管理" summary:"删除机器人" dc:"删除指定的机器人"`
	input.TradingRobotDeleteInp
}

// RobotDeleteRes 删除机器人响应
type RobotDeleteRes struct{}

// RobotViewReq 查看机器人请求
type RobotViewReq struct {
	g.Meta `path:"/trading/robot/view" method:"get" tags:"交易管理" summary:"查看机器人详情" dc:"查看机器人详细信息"`
	input.TradingRobotViewInp
}

// RobotViewRes 查看机器人响应
type RobotViewRes struct {
	*input.TradingRobotViewModel
}

// RobotStartReq 启动机器人请求
type RobotStartReq struct {
	g.Meta `path:"/trading/robot/start" method:"post" tags:"交易管理" summary:"启动机器人" dc:"启动指定的机器人"`
	input.TradingRobotStartInp
}

// RobotStartRes 启动机器人响应
type RobotStartRes struct{}

// RobotPauseReq 暂停机器人请求
type RobotPauseReq struct {
	g.Meta `path:"/trading/robot/pause" method:"post" tags:"交易管理" summary:"暂停机器人" dc:"暂停指定的机器人"`
	input.TradingRobotPauseInp
}

// RobotPauseRes 暂停机器人响应
type RobotPauseRes struct{}

// RobotStopReq 停止机器人请求
type RobotStopReq struct {
	g.Meta `path:"/trading/robot/stop" method:"post" tags:"交易管理" summary:"停止机器人" dc:"停止指定的机器人"`
	input.TradingRobotStopInp
}

// RobotStopRes 停止机器人响应
type RobotStopRes struct{}

// RobotStatsReq 获取运行统计请求
type RobotStatsReq struct {
	g.Meta `path:"/trading/robot/stats" method:"get" tags:"交易管理" summary:"获取运行统计" dc:"获取机器人运行统计数据"`
	input.TradingRobotStatsInp
}

// RobotStatsRes 获取运行统计响应
type RobotStatsRes struct {
	*input.TradingRobotStatsModel
}

// RobotRecommendStrategyReq 推荐策略请求
type RobotRecommendStrategyReq struct {
	g.Meta `path:"/trading/robot/recommendStrategy" method:"get" tags:"交易管理" summary:"推荐策略" dc:"根据风险偏好和市场状态推荐策略"`
	input.TradingRobotRecommendStrategyInp
}

// RobotRecommendStrategyRes 推荐策略响应
type RobotRecommendStrategyRes struct {
	*input.TradingRobotRecommendStrategyModel
}

// RobotPositionsReq 获取机器人持仓请求
type RobotPositionsReq struct {
	g.Meta `path:"/trading/robot/positions" method:"get" tags:"交易管理" summary:"获取机器人持仓" dc:"获取机器人的实时持仓"`
	toogoin.GetPositionsInp
}

// RobotPositionsRes 获取机器人持仓响应
type RobotPositionsRes struct {
	List []*toogoin.PositionModel `json:"list" dc:"持仓列表"`
}

// RobotOrdersReq 获取机器人挂单请求
type RobotOrdersReq struct {
	g.Meta `path:"/trading/robot/orders" method:"get" tags:"交易管理" summary:"获取机器人挂单" dc:"获取机器人的当前挂单"`
	toogoin.GetOrdersInp
}

// RobotOrdersRes 获取机器人挂单响应
type RobotOrdersRes struct {
	List []*toogoin.OrderModel `json:"list" dc:"挂单列表"`
}

// RobotOrderHistoryReq 获取机器人历史订单请求
type RobotOrderHistoryReq struct {
	g.Meta `path:"/trading/robot/orderHistory" method:"get" tags:"交易管理" summary:"获取机器人历史订单" dc:"获取机器人的历史订单"`
	toogoin.GetOrdersInp
}

// RobotOrderHistoryRes 获取机器人历史订单响应
type RobotOrderHistoryRes struct {
	List []*toogoin.OrderModel `json:"list" dc:"订单列表"`
}

// RobotClosePositionReq 手动平仓请求
type RobotClosePositionReq struct {
	g.Meta `path:"/trading/robot/closePosition" method:"post" tags:"交易管理" summary:"手动平仓" dc:"手动平仓指定持仓"`
	toogoin.ClosePositionInp
}

// RobotClosePositionRes 手动平仓响应
type RobotClosePositionRes struct{}

// RobotCancelOrderReq 撤销挂单请求
type RobotCancelOrderReq struct {
	g.Meta  `path:"/trading/robot/cancelOrder" method:"post" tags:"交易管理" summary:"撤销挂单" dc:"撤销指定的挂单"`
	RobotId int64  `json:"robotId" v:"required#机器人ID不能为空" dc:"机器人ID"`
	OrderId string `json:"orderId" v:"required#订单ID不能为空" dc:"订单ID"`
}

// RobotCancelOrderRes 撤销挂单响应
type RobotCancelOrderRes struct{}

// RobotSetTakeProfitSwitchReq 设置止盈回撤开关请求
type RobotSetTakeProfitSwitchReq struct {
	g.Meta       `path:"/trading/robot/setTakeProfitSwitch" method:"post" tags:"交易管理" summary:"设置止盈回撤开关" dc:"设置指定持仓的止盈回撤开关状态"`
	RobotId      int64  `json:"robotId" v:"required#机器人ID不能为空" dc:"机器人ID"`
	PositionSide string `json:"positionSide" v:"required#持仓方向不能为空" dc:"持仓方向: LONG/SHORT"`
	Enabled      bool   `json:"enabled" dc:"是否启用止盈回撤"`
}

// RobotSetTakeProfitSwitchRes 设置止盈回撤开关响应
type RobotSetTakeProfitSwitchRes struct{}

// RobotRiskConfigReq 获取风险配置请求
type RobotRiskConfigReq struct {
	g.Meta  `path:"/trading/robot/riskConfig" method:"get" tags:"交易管理" summary:"获取风险配置" dc:"获取机器人的风险偏好配置"`
	RobotId int64 `json:"robotId" v:"required#机器人ID不能为空" dc:"机器人ID"`
}

// RobotRiskConfigRes 获取风险配置响应
type RobotRiskConfigRes struct {
	Config *RiskConfig `json:"config" dc:"风险配置"`
}

// RobotSaveRiskConfigReq 保存风险配置请求
type RobotSaveRiskConfigReq struct {
	g.Meta  `path:"/trading/robot/riskConfig/save" method:"post" tags:"交易管理" summary:"保存风险配置" dc:"保存机器人的风险偏好配置"`
	RobotId int64       `json:"robotId" v:"required#机器人ID不能为空" dc:"机器人ID"`
	Config  *RiskConfig `json:"config" v:"required#配置不能为空" dc:"风险配置"`
}

// RobotSaveRiskConfigRes 保存风险配置响应
type RobotSaveRiskConfigRes struct{}

// RiskConfig 风险配置
type RiskConfig struct {
	MarketRiskMapping map[string]string     `json:"marketRiskMapping" dc:"市场状态→风险偏好映射"`
	RiskParams        map[string]RiskParams `json:"riskParams" dc:"风险偏好参数配置"`
}

// RiskParams 风险参数
type RiskParams struct {
	LeverageMin          int     `json:"leverageMin" dc:"最小杠杆"`
	LeverageMax          int     `json:"leverageMax" dc:"最大杠杆"`
	MarginPercentMin     float64 `json:"marginPercentMin" dc:"最小保证金比例"`
	MarginPercentMax     float64 `json:"marginPercentMax" dc:"最大保证金比例"`
	StopLossPercent      float64 `json:"stopLossPercent" dc:"止损百分比"`
	ProfitRetreatPercent float64 `json:"profitRetreatPercent" dc:"盈利回撤百分比"`
}

// RobotSignalLogsReq 获取方向预警日志请求
type RobotSignalLogsReq struct {
	g.Meta  `path:"/trading/robot/signalLogs" method:"get" tags:"交易管理" summary:"获取方向预警日志" dc:"获取机器人的方向预警日志"`
	RobotId int64 `json:"robotId" v:"required#机器人ID不能为空" dc:"机器人ID"`
	Limit   int   `json:"limit" dc:"限制数量"`
}

// RobotSignalLogsRes 获取方向预警日志响应
type RobotSignalLogsRes struct {
	List []*SignalLogItem `json:"list" dc:"日志列表"`
}

// SignalLogItem 信号日志项
type SignalLogItem struct {
	Id             int64   `json:"id" dc:"ID"`
	RobotId        int64   `json:"robotId" dc:"机器人ID"`
	Symbol         string  `json:"symbol" dc:"交易对"`
	SignalType     string  `json:"signalType" dc:"信号类型"`
	SignalStrength float64 `json:"signalStrength" dc:"信号强度"`
	CurrentPrice   float64 `json:"currentPrice" dc:"当前价格"`
	WindowMinPrice float64 `json:"windowMinPrice" dc:"窗口最低价"`
	WindowMaxPrice float64 `json:"windowMaxPrice" dc:"窗口最高价"`
	Threshold      float64 `json:"threshold" dc:"波动阈值"`
	Reason         string  `json:"reason" dc:"原因"`
	MarketState    string  `json:"marketState" dc:"市场状态"`
	RiskPreference string  `json:"riskPreference" dc:"风险偏好"`
	Executed       bool    `json:"executed" dc:"是否执行"`
	ExecuteResult  string  `json:"executeResult" dc:"执行结果/未执行原因"`
	IsProcessed    bool    `json:"isProcessed" dc:"是否已读：true=已读，false=未读"`
	CreatedAt      string  `json:"createdAt" dc:"创建时间"`
}

// RobotExecutionLogsReq 获取交易执行日志请求
type RobotExecutionLogsReq struct {
	g.Meta  `path:"/trading/robot/executionLogs" method:"get" tags:"交易管理" summary:"获取交易执行日志" dc:"获取机器人的交易执行日志（下单、平仓等）"`
	RobotId int64 `json:"robotId" v:"required#机器人ID不能为空" dc:"机器人ID"`
	Limit   int   `json:"limit" dc:"限制数量"`
}

// RobotExecutionLogsRes 获取交易执行日志响应
type RobotExecutionLogsRes struct {
	List []*ExecutionLogItem `json:"list" dc:"日志列表"`
}

// ExecutionLogItem 交易执行日志项
type ExecutionLogItem struct {
	Id          int64  `json:"id" dc:"ID"`
	SignalLogId int64  `json:"signalLogId" dc:"关联的预警日志ID"`
	RobotId     int64  `json:"robotId" dc:"机器人ID"`
	OrderId     int64  `json:"orderId" dc:"关联的订单ID"`
	EventType   string `json:"eventType" dc:"事件类型：signal_detected/order_attempt/order_success/order_failed/position_monitor/position_close/stop_loss/take_profit"`
	EventData   string `json:"eventData" dc:"事件数据（JSON格式）"`
	Status      string `json:"status" dc:"状态：pending/success/failed"`
	Message     string `json:"message" dc:"消息"`
	CreatedAt   string `json:"createdAt" dc:"创建时间"`
}

// RobotReloadStrategyReq 重新加载策略配置请求
type RobotReloadStrategyReq struct {
	g.Meta `path:"/trading/robot/reloadStrategy" method:"post" tags:"交易管理" summary:"重新加载策略" dc:"重新加载运行中机器人的策略配置"`
	Id     int64 `json:"id" v:"required#机器人ID不能为空" dc:"机器人ID"`
}

// RobotReloadStrategyRes 重新加载策略配置响应
type RobotReloadStrategyRes struct{}
