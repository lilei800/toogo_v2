// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package trading

import (
	"context"
	"hotgo/api/admin/trading"
	tradingLogic "hotgo/internal/logic/trading"
)

var Monitor = cMonitor{}

type cMonitor struct{}

// Ticker 获取实时行情
func (c *cMonitor) Ticker(ctx context.Context, req *trading.MonitorTickerReq) (res *trading.MonitorTickerRes, err error) {
	out, err := tradingLogic.Monitor.GetTicker(ctx, &req.TradingMonitorTickerInp)
	if err != nil {
		return nil, err
	}

	res = &trading.MonitorTickerRes{
		TradingMonitorTickerModel: out,
	}
	return
}

// MarketState 获取市场状态
func (c *cMonitor) MarketState(ctx context.Context, req *trading.MonitorMarketStateReq) (res *trading.MonitorMarketStateRes, err error) {
	out, err := tradingLogic.Monitor.GetMarketState(ctx, &req.TradingMonitorMarketStateInp)
	if err != nil {
		return nil, err
	}

	res = &trading.MonitorMarketStateRes{
		TradingMonitorMarketStateModel: out,
	}
	return
}

// Logs 获取监控日志列表
func (c *cMonitor) Logs(ctx context.Context, req *trading.MonitorLogsReq) (res *trading.MonitorLogsRes, err error) {
	list, totalCount, err := tradingLogic.Monitor.GetMonitorLogs(ctx, &req.TradingMonitorLogListInp)
	if err != nil {
		return nil, err
	}

	res = &trading.MonitorLogsRes{
		List:       list,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}
	return
}

// RobotAnalysis 获取机器人实时分析数据
func (c *cMonitor) RobotAnalysis(ctx context.Context, req *trading.MonitorRobotAnalysisReq) (res *trading.MonitorRobotAnalysisRes, err error) {
	out, err := tradingLogic.Monitor.GetRobotAnalysis(ctx, req.RobotId)
	if err != nil {
		// 返回连接失败的状态而不是直接报错
		res = &trading.MonitorRobotAnalysisRes{
			RobotId:         req.RobotId,
			Connected:       false,
			ConnectionError: err.Error(),
		}
		return res, nil
	}
	return out, nil
}

// BatchRobotAnalysis 批量获取机器人实时分析数据
func (c *cMonitor) BatchRobotAnalysis(ctx context.Context, req *trading.MonitorBatchRobotAnalysisReq) (res *trading.MonitorBatchRobotAnalysisRes, err error) {
	out, err := tradingLogic.Monitor.GetBatchRobotAnalysis(ctx, req.RobotIds)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Kline 获取K线数据
func (c *cMonitor) Kline(ctx context.Context, req *trading.MonitorKlineReq) (res *trading.MonitorKlineRes, err error) {
	out, err := tradingLogic.Monitor.GetKline(ctx, req.ApiConfigId, req.Symbol, req.Interval, req.Limit)
	if err != nil {
		return nil, err
	}
	return out, nil
}

