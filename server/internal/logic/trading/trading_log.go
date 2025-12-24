// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
// @Description 交易日志服务
package trading

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TradingLogService 交易日志服务
type TradingLogService struct{}

var tradingLogService = &TradingLogService{}

// GetTradingLogService 获取交易日志服务
func GetTradingLogService() *TradingLogService {
	return tradingLogService
}

// OperationLog 操作日志
type OperationLog struct {
	RobotId      int64       `json:"robotId"`
	UserId       int64       `json:"userId"`
	Operation    string      `json:"operation"`
	Symbol       string      `json:"symbol"`
	Side         string      `json:"side"`
	PositionSide string      `json:"positionSide"`
	OrderType    string      `json:"orderType"`
	Quantity     float64     `json:"quantity"`
	Price        float64     `json:"price"`
	OrderId      string      `json:"orderId"`
	Status       string      `json:"status"`
	ErrorCode    string      `json:"errorCode"`
	ErrorMsg     string      `json:"errorMsg"`
	RequestData  interface{} `json:"requestData"`
	ResponseData interface{} `json:"responseData"`
	ExecuteTime  int64       `json:"executeTime"`
}

// LogOperation 记录交易操作
func (s *TradingLogService) LogOperation(ctx context.Context, log *OperationLog) error {
	requestData, _ := json.Marshal(log.RequestData)
	responseData, _ := json.Marshal(log.ResponseData)

	_, err := g.DB().Model("hg_trading_operation_log").Ctx(ctx).Insert(g.Map{
		"robot_id":      log.RobotId,
		"user_id":       log.UserId,
		"operation":     log.Operation,
		"symbol":        log.Symbol,
		"side":          log.Side,
		"position_side": log.PositionSide,
		"order_type":    log.OrderType,
		"quantity":      log.Quantity,
		"price":         log.Price,
		"order_id":      log.OrderId,
		"status":        log.Status,
		"error_code":    log.ErrorCode,
		"error_msg":     log.ErrorMsg,
		"request_data":  string(requestData),
		"response_data": string(responseData),
		"execute_time":  log.ExecuteTime,
	})
	return err
}

// DailyStats 日统计
type DailyStats struct {
	RobotId        int64   `json:"robotId"`
	UserId         int64   `json:"userId"`
	Date           string  `json:"date"`
	Symbol         string  `json:"symbol"`
	TotalTrades    int     `json:"totalTrades"`
	WinTrades      int     `json:"winTrades"`
	LossTrades     int     `json:"lossTrades"`
	TotalVolume    float64 `json:"totalVolume"`
	TotalPnl       float64 `json:"totalPnl"`
	RealizedPnl    float64 `json:"realizedPnl"`
	Commission     float64 `json:"commission"`
	MaxProfit      float64 `json:"maxProfit"`
	MaxLoss        float64 `json:"maxLoss"`
	MaxDrawdown    float64 `json:"maxDrawdown"`
	WinRate        float64 `json:"winRate"`
	ProfitFactor   float64 `json:"profitFactor"`
	AvgHoldingTime int     `json:"avgHoldingTime"`
}

// UpdateDailyStats 更新日统计
func (s *TradingLogService) UpdateDailyStats(ctx context.Context, stats *DailyStats) error {
	// 计算胜率
	if stats.TotalTrades > 0 {
		stats.WinRate = float64(stats.WinTrades) / float64(stats.TotalTrades)
	}

	// 计算盈亏比
	if stats.MaxLoss != 0 {
		stats.ProfitFactor = stats.MaxProfit / (-stats.MaxLoss)
	}

	_, err := g.DB().Model("hg_trading_daily_stats").Ctx(ctx).Save(g.Map{
		"robot_id":         stats.RobotId,
		"user_id":          stats.UserId,
		"date":             stats.Date,
		"symbol":           stats.Symbol,
		"total_trades":     stats.TotalTrades,
		"win_trades":       stats.WinTrades,
		"loss_trades":      stats.LossTrades,
		"total_volume":     stats.TotalVolume,
		"total_pnl":        stats.TotalPnl,
		"realized_pnl":     stats.RealizedPnl,
		"commission":       stats.Commission,
		"max_profit":       stats.MaxProfit,
		"max_loss":         stats.MaxLoss,
		"max_drawdown":     stats.MaxDrawdown,
		"win_rate":         stats.WinRate,
		"profit_factor":    stats.ProfitFactor,
		"avg_holding_time": stats.AvgHoldingTime,
	})
	return err
}

// GetDailyStats 获取日统计
func (s *TradingLogService) GetDailyStats(ctx context.Context, robotId int64, date string) (*DailyStats, error) {
	var stats DailyStats
	err := g.DB().Model("hg_trading_daily_stats").Ctx(ctx).
		Where("robot_id", robotId).
		Where("date", date).
		Scan(&stats)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetRobotStatsRange 获取机器人统计范围
func (s *TradingLogService) GetRobotStatsRange(ctx context.Context, robotId int64, startDate, endDate string) ([]*DailyStats, error) {
	var stats []*DailyStats
	err := g.DB().Model("hg_trading_daily_stats").Ctx(ctx).
		Where("robot_id", robotId).
		WhereBetween("date", startDate, endDate).
		Order("date ASC").
		Scan(&stats)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// UserSummary 用户汇总
type UserSummary struct {
	UserId         int64     `json:"userId"`
	TotalRobots    int       `json:"totalRobots"`
	ActiveRobots   int       `json:"activeRobots"`
	TotalTrades    int       `json:"totalTrades"`
	TotalVolume    float64   `json:"totalVolume"`
	TotalPnl       float64   `json:"totalPnl"`
	TotalCommission float64  `json:"totalCommission"`
	OverallWinRate float64   `json:"overallWinRate"`
	BestRobotId    int64     `json:"bestRobotId"`
	BestRobotPnl   float64   `json:"bestRobotPnl"`
	FirstTradeTime time.Time `json:"firstTradeTime"`
	LastTradeTime  time.Time `json:"lastTradeTime"`
}

// UpdateUserSummary 更新用户汇总
func (s *TradingLogService) UpdateUserSummary(ctx context.Context, userId int64) error {
	// 聚合用户所有机器人的统计数据
	result, err := g.DB().Model("hg_trading_daily_stats").Ctx(ctx).
		Fields("SUM(total_trades) as total_trades, SUM(total_volume) as total_volume, SUM(total_pnl) as total_pnl, SUM(commission) as total_commission, SUM(win_trades) as win_trades").
		Where("user_id", userId).
		One()
	if err != nil {
		return err
	}

	totalTrades := result["total_trades"].Int()
	winTrades := result["win_trades"].Int()
	winRate := float64(0)
	if totalTrades > 0 {
		winRate = float64(winTrades) / float64(totalTrades)
	}

	// 获取机器人数量
	robotCount, _ := g.DB().Model("hg_trading_robot").Ctx(ctx).Where("user_id", userId).Count()
	activeRobots, _ := g.DB().Model("hg_trading_robot").Ctx(ctx).Where("user_id", userId).Where("status", 2).Count()

	// 获取最佳机器人
	bestRobot, _ := g.DB().Model("hg_trading_robot").Ctx(ctx).
		Fields("id, total_profit").
		Where("user_id", userId).
		Order("total_profit DESC").
		One()

	_, err = g.DB().Model("hg_trading_user_summary").Ctx(ctx).Save(g.Map{
		"user_id":          userId,
		"total_robots":     robotCount,
		"active_robots":    activeRobots,
		"total_trades":     result["total_trades"].Int(),
		"total_volume":     result["total_volume"].Float64(),
		"total_pnl":        result["total_pnl"].Float64(),
		"total_commission": result["total_commission"].Float64(),
		"overall_win_rate": winRate,
		"best_robot_id":    bestRobot["id"].Int64(),
		"best_robot_pnl":   bestRobot["total_profit"].Float64(),
		"last_trade_time":  gtime.Now(),
	})
	return err
}

// SignalLog 信号日志
type SignalLog struct {
	RobotId        int64       `json:"robotId"`
	StrategyId     int64       `json:"strategyId"`
	Symbol         string      `json:"symbol"`
	SignalType     string      `json:"signalType"`
	SignalSource   string      `json:"signalSource"`
	SignalStrength float64     `json:"signalStrength"`
	CurrentPrice   float64     `json:"currentPrice"`
	TargetPrice    float64     `json:"targetPrice"`
	StopLoss       float64     `json:"stopLoss"`
	TakeProfit     float64     `json:"takeProfit"`
	Executed       bool        `json:"executed"`
	ExecuteResult  string      `json:"executeResult"`
	Reason         string      `json:"reason"`
	Indicators     interface{} `json:"indicators"`
}

// LogSignal 记录交易信号
func (s *TradingLogService) LogSignal(ctx context.Context, log *SignalLog) error {
	indicators, _ := json.Marshal(log.Indicators)

	executed := 0
	if log.Executed {
		executed = 1
	}

	_, err := g.DB().Model("hg_trading_signal_log").Ctx(ctx).Insert(g.Map{
		"robot_id":        log.RobotId,
		"strategy_id":     log.StrategyId,
		"symbol":          log.Symbol,
		"signal_type":     log.SignalType,
		"signal_source":   log.SignalSource,
		"signal_strength": log.SignalStrength,
		"current_price":   log.CurrentPrice,
		"target_price":    log.TargetPrice,
		"stop_loss":       log.StopLoss,
		"take_profit":     log.TakeProfit,
		"executed":        executed,
		"execute_result":  log.ExecuteResult,
		"reason":          log.Reason,
		"indicators":      string(indicators),
	})
	return err
}

// SystemMonitor 系统监控数据
type SystemMonitor struct {
	Date          string  `json:"date"`
	Hour          int     `json:"hour"`
	TotalUsers    int     `json:"totalUsers"`
	TotalRobots   int     `json:"totalRobots"`
	TotalOrders   int     `json:"totalOrders"`
	TotalVolume   float64 `json:"totalVolume"`
	TotalPnl      float64 `json:"totalPnl"`
	ApiCalls      int     `json:"apiCalls"`
	ApiErrors     int     `json:"apiErrors"`
	AvgLatency    int     `json:"avgLatency"`
	MaxLatency    int     `json:"maxLatency"`
	WsConnections int     `json:"wsConnections"`
}

// UpdateSystemMonitor 更新系统监控数据
func (s *TradingLogService) UpdateSystemMonitor(ctx context.Context, monitor *SystemMonitor) error {
	_, err := g.DB().Model("hg_trading_system_monitor").Ctx(ctx).Save(g.Map{
		"date":           monitor.Date,
		"hour":           monitor.Hour,
		"total_users":    monitor.TotalUsers,
		"total_robots":   monitor.TotalRobots,
		"total_orders":   monitor.TotalOrders,
		"total_volume":   monitor.TotalVolume,
		"total_pnl":      monitor.TotalPnl,
		"api_calls":      monitor.ApiCalls,
		"api_errors":     monitor.ApiErrors,
		"avg_latency":    monitor.AvgLatency,
		"max_latency":    monitor.MaxLatency,
		"ws_connections": monitor.WsConnections,
	})
	return err
}

// GetOperationLogs 获取操作日志
func (s *TradingLogService) GetOperationLogs(ctx context.Context, robotId int64, page, pageSize int) ([]gdb.Record, int, error) {
	model := g.DB().Model("hg_trading_operation_log").Ctx(ctx)
	if robotId > 0 {
		model = model.Where("robot_id", robotId)
	}

	total, err := model.Clone().Count()
	if err != nil {
		return nil, 0, err
	}

	list, err := model.Page(page, pageSize).Order("created_at DESC").All()
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// GetSignalLogs 获取信号日志
func (s *TradingLogService) GetSignalLogs(ctx context.Context, robotId int64, page, pageSize int) ([]gdb.Record, int, error) {
	model := g.DB().Model("hg_trading_signal_log").Ctx(ctx)
	if robotId > 0 {
		model = model.Where("robot_id", robotId)
	}

	total, err := model.Clone().Count()
	if err != nil {
		return nil, 0, err
	}

	list, err := model.Page(page, pageSize).Order("created_at DESC").All()
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// CleanOldLogs 清理过期日志
func (s *TradingLogService) CleanOldLogs(ctx context.Context, days int) error {
	expireTime := gtime.Now().AddDate(0, 0, -days)

	// 清理操作日志
	_, err := g.DB().Model("hg_trading_operation_log").Ctx(ctx).
		Where("created_at < ?", expireTime).Delete()
	if err != nil {
		return err
	}

	// 清理信号日志
	_, err = g.DB().Model("hg_trading_signal_log").Ctx(ctx).
		Where("created_at < ?", expireTime).Delete()
	if err != nil {
		return err
	}

	g.Log().Infof(ctx, "Cleaned old trading logs older than %d days", days)
	return nil
}

