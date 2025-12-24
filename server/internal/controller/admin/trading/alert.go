// Package trading 预警日志控制器
package trading

import (
	"context"
	"time"

	"hotgo/api/admin"
	"hotgo/internal/dao"
	"hotgo/internal/library/market"
	"hotgo/internal/logic/toogo"
	"hotgo/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// AlertController 预警日志控制器
var AlertController = new(cAlertController)

type cAlertController struct{}

// MarketStateLogList 市场状态预警日志列表
func (c *cAlertController) MarketStateLogList(ctx context.Context, req *admin.MarketStateLogListReq) (res *admin.MarketStateLogListRes, err error) {
	res = &admin.MarketStateLogListRes{}

	model := dao.TradingMarketStateLog.Ctx(ctx)

	if req.Platform != "" {
		model = model.Where("platform", req.Platform)
	}
	if req.Symbol != "" {
		model = model.Where("symbol", req.Symbol)
	}
	if req.NewState != "" {
		model = model.Where("new_state", req.NewState)
	}

	res.Total, err = model.Count()
	if err != nil {
		return nil, err
	}

	var list []*entity.TradingMarketStateLog
	err = model.Page(req.Page, req.PerPage).OrderDesc("id").Scan(&list)
	if err != nil {
		return nil, err
	}

	res.List = list
	return
}

// RiskPreferenceLogList 风险偏好预警日志列表
func (c *cAlertController) RiskPreferenceLogList(ctx context.Context, req *admin.RiskPreferenceLogListReq) (res *admin.RiskPreferenceLogListRes, err error) {
	res = &admin.RiskPreferenceLogListRes{}

	model := dao.TradingRiskPreferenceLog.Ctx(ctx)

	if req.RobotId > 0 {
		model = model.Where("robot_id", req.RobotId)
	}
	if req.UserId > 0 {
		model = model.Where("user_id", req.UserId)
	}
	if req.NewPreference != "" {
		model = model.Where("new_preference", req.NewPreference)
	}

	res.Total, err = model.Count()
	if err != nil {
		return nil, err
	}

	var list []*entity.TradingRiskPreferenceLog
	err = model.Page(req.Page, req.PerPage).OrderDesc("id").Scan(&list)
	if err != nil {
		return nil, err
	}

	res.List = list
	return
}

// DirectionLogList 方向预警日志列表（从 hg_trading_signal_log 读取）
func (c *cAlertController) DirectionLogList(ctx context.Context, req *admin.DirectionLogListReq) (res *admin.DirectionLogListRes, err error) {
	res = &admin.DirectionLogListRes{}

	model := g.DB().Model("hg_trading_signal_log").Ctx(ctx)

	// 只查询有价值的信号（long/short）
	model = model.WhereIn("signal_type", []string{"long", "short", "LONG", "SHORT"})

	if req.RobotId > 0 {
		model = model.Where("robot_id", req.RobotId)
	}
	if req.Symbol != "" {
		model = model.WhereLike("symbol", "%"+req.Symbol+"%")
	}
	if req.SignalType != "" {
		model = model.Where("signal_type", req.SignalType)
	}

	res.Total, err = model.Clone().Count()
	if err != nil {
		return nil, err
	}

	records, err := model.Page(req.Page, req.PerPage).OrderDesc("id").All()
	if err != nil {
		return nil, err
	}

	res.List = make([]*admin.DirectionLogItem, 0, len(records))
	for _, r := range records {
		item := &admin.DirectionLogItem{
			Id:             r["id"].Int64(),
			RobotId:        r["robot_id"].Int64(),
			StrategyId:     r["strategy_id"].Int64(),
			Symbol:         r["symbol"].String(),
			SignalType:     r["signal_type"].String(),
			SignalSource:   r["signal_source"].String(),
			SignalStrength: r["signal_strength"].Float64(),
			CurrentPrice:   r["current_price"].Float64(),
			WindowMinPrice: r["window_min_price"].Float64(),
			WindowMaxPrice: r["window_max_price"].Float64(),
			Threshold:      r["threshold"].Float64(),
			TargetPrice:    r["target_price"].Float64(),
			StopLoss:       r["stop_loss"].Float64(),
			TakeProfit:     r["take_profit"].Float64(),
			Executed:       r["executed"].Int(),
			ExecuteResult:  r["execute_result"].String(),
			Reason:         r["reason"].String(),
			MarketState:    r["market_state"].String(),
			RiskPreference: r["risk_preference"].String(),
			CreatedAt:      r["created_at"].String(),
		}
		res.List = append(res.List, item)
	}

	return
}

// RobotRealtime 获取机器人实时状态
func (c *cAlertController) RobotRealtime(ctx context.Context, req *admin.RobotRealtimeReq) (res *admin.RobotRealtimeRes, err error) {
	res = &admin.RobotRealtimeRes{}

	var data *entity.TradingRobotRealtime
	err = dao.TradingRobotRealtime.Ctx(ctx).Where("robot_id", req.RobotId).Scan(&data)
	if err != nil {
		return nil, err
	}

	res.TradingRobotRealtime = data
	return
}

// RobotRealtimeList 机器人实时状态列表
func (c *cAlertController) RobotRealtimeList(ctx context.Context, req *admin.RobotRealtimeListReq) (res *admin.RobotRealtimeListRes, err error) {
	res = &admin.RobotRealtimeListRes{}

	model := dao.TradingRobotRealtime.Ctx(ctx)

	if req.UserId > 0 {
		model = model.Where("user_id", req.UserId)
	}

	res.Total, err = model.Count()
	if err != nil {
		return nil, err
	}

	var list []*entity.TradingRobotRealtime
	err = model.Page(req.Page, req.PerPage).OrderDesc("updated_at").Scan(&list)
	if err != nil {
		return nil, err
	}

	res.List = list
	return
}

// MarketAnalysis 获取市场分析
func (c *cAlertController) MarketAnalysis(ctx context.Context, req *admin.MarketAnalysisReq) (res *admin.MarketAnalysisRes, err error) {
	analysis := market.GetMarketAnalyzer().GetAnalysis(req.Platform, req.Symbol)
	if analysis == nil {
		return nil, gerror.New("暂无市场分析数据，请确保有机器人在运行")
	}

	res = &admin.MarketAnalysisRes{
		Platform:        analysis.Platform,
		Symbol:          analysis.Symbol,
		CurrentPrice:    analysis.CurrentPrice,
		MarketState:     string(analysis.MarketState),
		MarketStateConf: analysis.MarketStateConf,
		TrendStrength:   analysis.TrendStrength,
		Volatility:      analysis.Volatility,
		SupportLevel:    analysis.SupportLevel,
		ResistanceLevel: analysis.ResistanceLevel,
	}

	// 添加技术指标
	if analysis.Indicators != nil {
		res.Indicators = map[string]interface{}{
			"trendScore":      analysis.Indicators.TrendScore,
			"volatilityScore": analysis.Indicators.VolatilityScore,
		}
	}

	// 添加多周期数据
	res.TimeframeData = make(map[string]interface{})
	for interval, tf := range analysis.TimeframeAnalysis {
		res.TimeframeData[interval] = map[string]interface{}{
			"trend":    tf.Trend,
			"strength":        tf.TrendStrength,
			"priceVolatility": tf.PriceVolatility,
			"priceChangeRate": tf.PriceChangeRate,
		}
	}

	return
}

// DirectionSignal 获取方向信号
func (c *cAlertController) DirectionSignal(ctx context.Context, req *admin.DirectionSignalReq) (res *admin.DirectionSignalRes, err error) {
	signal := market.GetDirectionSignalService().GetSignal(req.Platform, req.Symbol)
	if signal == nil {
		return nil, gerror.New("暂无方向信号数据，请确保有机器人在运行")
	}

	res = &admin.DirectionSignalRes{
		Platform:   signal.Platform,
		Symbol:     signal.Symbol,
		Direction:  string(signal.Direction),
		Strength:   signal.Strength,
		Confidence: signal.Confidence,
		Action:     string(signal.Action),
		EntryPrice: signal.EntryPrice,
		StopLoss:   signal.StopLoss,
		Reason:     signal.Reason,
	}

	// 添加各周期信号
	res.TimeframeSignals = make(map[string]interface{})
	for interval, dir := range signal.TimeframeSignals {
		res.TimeframeSignals[interval] = string(dir)
	}

	return
}

// RiskEvaluation 获取风险评估
func (c *cAlertController) RiskEvaluation(ctx context.Context, req *admin.RiskEvaluationReq) (res *admin.RiskEvaluationRes, err error) {
	eval := market.GetRiskEvaluator().GetEvaluation(req.RobotId)
	if eval == nil {
		return nil, gerror.New("暂无风险评估数据，请确保机器人在运行")
	}

	res = &admin.RiskEvaluationRes{
		RobotId:                eval.RobotId,
		WinProbability:         eval.WinProbability,
		RiskPreference:         string(eval.RiskPreference),
		MarketScore:            eval.MarketScore,
		TechnicalScore:         eval.TechnicalScore,
		AccountScore:           eval.AccountScore,
		HistoryScore:           eval.HistoryScore,
		VolatilityRisk:         eval.VolatilityRisk,
		SuggestedLeverage:      eval.SuggestedLeverage,
		SuggestedMarginPercent: eval.SuggestedMarginPercent,
		SuggestedStopLoss:      eval.SuggestedStopLoss,
		SuggestedTakeProfit:    eval.SuggestedTakeProfit,
		RiskLevel:              eval.RiskLevel,
		Reason:                 eval.Reason,
	}

	return
}

// EngineStatus 获取引擎状态
func (c *cAlertController) EngineStatus(ctx context.Context, req *admin.EngineStatusReq) (res *admin.EngineStatusRes, err error) {
	res = &admin.EngineStatusRes{
		Running:             true,
		ActiveRobots:        toogo.GetRobotTaskManager().GetActiveRobotCount(),
		ActiveSubscriptions: market.GetMarketServiceManager().GetActiveServiceCount(),
	}
	return
}

// GlobalEngineDetail 获取全局引擎详情
func (c *cAlertController) GlobalEngineDetail(ctx context.Context, req *admin.GlobalEngineDetailReq) (res *admin.GlobalEngineDetailRes, err error) {
	res = &admin.GlobalEngineDetailRes{
		Running: true,
	}

	// MarketDataService 状态
	res.MarketDataService = c.getMarketDataServiceStatus(ctx)

	// MarketAnalyzer 状态
	res.MarketAnalyzer = c.getMarketAnalyzerStatus(ctx)

	// DirectionSignalService 状态
	res.DirectionSignalService = c.getDirectionSignalServiceStatus(ctx)

	// RobotTaskManager 状态
	res.RobotTaskManager = c.getRobotTaskManagerStatus(ctx)

	// AlertLogger 状态
	res.AlertLogger = c.getAlertLoggerStatus(ctx)

	// TradeStatistics 状态
	res.TradeStatistics = c.getTradeStatisticsStatus(ctx)

	return
}

// getMarketDataServiceStatus 获取行情数据服务状态
func (c *cAlertController) getMarketDataServiceStatus(ctx context.Context) *admin.MarketDataServiceStatus {
	msm := market.GetMarketServiceManager()
	status := &admin.MarketDataServiceStatus{
		Running:       msm.IsRunning(),
		Subscriptions: 0,
		TickerList:    make([]*admin.SubscriptionDetail, 0),
	}

	// 获取所有交易所服务的订阅详情
	allServices := msm.GetAllServices()
	for platform, svc := range allServices {
		subscriptions := svc.GetAllSubscriptions()
		status.Subscriptions += len(subscriptions)

		for symbol, refCount := range subscriptions {
			ticker := svc.GetTicker(symbol)
			detail := &admin.SubscriptionDetail{
				Platform:  platform,
				Symbol:    symbol,
				RefCount:  refCount,
				DataFresh: svc.IsDataFresh(symbol, 10*time.Second),
			}
			if ticker != nil {
				detail.LastPrice = ticker.LastPrice
				detail.Change24h = ticker.Change24h
				detail.LastUpdate = time.Unix(ticker.Timestamp/1000, 0).Format("2006-01-02 15:04:05")
			}
			status.TickerList = append(status.TickerList, detail)
		}
	}

	status.TickerCount = len(status.TickerList)
	return status
}

// getMarketAnalyzerStatus 获取市场分析引擎状态
func (c *cAlertController) getMarketAnalyzerStatus(ctx context.Context) *admin.MarketAnalyzerStatus {
	ma := market.GetMarketAnalyzer()
	status := &admin.MarketAnalyzerStatus{
		Running:      true,
		AnalysisList: make([]*admin.AnalysisDetail, 0),
	}

	// 获取所有分析数据
	analyses := ma.GetAllAnalyses()
	for _, analysis := range analyses {
		detail := &admin.AnalysisDetail{
			Platform:      analysis.Platform,
			Symbol:        analysis.Symbol,
			MarketState:   string(analysis.MarketState),
			TrendStrength: analysis.TrendStrength,
			Volatility:    analysis.Volatility,
			LastUpdate:    analysis.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		status.AnalysisList = append(status.AnalysisList, detail)
	}

	status.AnalysisCount = len(status.AnalysisList)
	return status
}

// getDirectionSignalServiceStatus 获取方向信号服务状态
func (c *cAlertController) getDirectionSignalServiceStatus(ctx context.Context) *admin.DirectionSignalServiceStatus {
	dss := market.GetDirectionSignalService()
	status := &admin.DirectionSignalServiceStatus{
		Running:    true,
		SignalList: make([]*admin.SignalDetail, 0),
	}

	// 获取所有信号
	signals := dss.GetAllSignals()
	for _, signal := range signals {
		detail := &admin.SignalDetail{
			Platform:   signal.Platform,
			Symbol:     signal.Symbol,
			Direction:  string(signal.Direction),
			Strength:   signal.Strength,
			Confidence: signal.Confidence,
			Action:     string(signal.Action),
			LastUpdate: signal.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		status.SignalList = append(status.SignalList, detail)
	}

	status.SignalCount = len(status.SignalList)
	return status
}

// getRobotTaskManagerStatus 获取机器人任务管理器状态
func (c *cAlertController) getRobotTaskManagerStatus(ctx context.Context) *admin.RobotTaskManagerStatus {
	rtm := toogo.GetRobotTaskManager()
	status := &admin.RobotTaskManagerStatus{
		Running:      rtm.IsRunning(),
		ActiveRobots: rtm.GetActiveRobotCount(),
		RobotList:    make([]*admin.ManagedRobotDetail, 0),
	}

	// 获取所有被管理的机器人
	robots := rtm.GetAllManagedRobots()
	for _, managed := range robots {
		detail := &admin.ManagedRobotDetail{
			RobotId:   managed.Robot.Id,
			RobotName: managed.Robot.RobotName,
			Platform:  managed.Platform,
			Symbol:    managed.Robot.Symbol,
		}

		// 检查行情连接状态
		ticker := market.GetMarketServiceManager().GetTicker(managed.Platform, managed.Robot.Symbol)
		detail.Connected = ticker != nil

		// 检查API连接状态 - 通过尝试获取账户余额来判断
		if managed.Exchange != nil {
			// 使用缓存的账户余额
			if managed.AccountBalance != nil {
				detail.ApiConnected = true
				detail.TotalBalance = managed.AccountBalance.TotalBalance
				detail.AvailBalance = managed.AccountBalance.AvailableBalance
			} else {
				// 如果缓存为空，实时获取一次
				balance, err := managed.Exchange.GetBalance(ctx)
				if err != nil {
					detail.ApiConnected = false
					detail.ApiError = err.Error()
				} else {
					detail.ApiConnected = true
					detail.TotalBalance = balance.TotalBalance
					detail.AvailBalance = balance.AvailableBalance
				}
			}

			// 使用缓存的持仓数据
			if managed.CurrentPositions != nil && len(managed.CurrentPositions) > 0 {
				for _, pos := range managed.CurrentPositions {
					if pos.PositionAmt != 0 {
						detail.HasPosition = true
						detail.PositionSide = pos.PositionSide
						detail.PositionAmt = pos.PositionAmt
						detail.EntryPrice = pos.EntryPrice
						detail.UnrealizedPnl = pos.UnrealizedPnl
						break
					}
				}
			} else if detail.ApiConnected && detail.ApiError == "" {
				// 如果缓存为空且API连接正常，实时获取一次
				positions, err := managed.Exchange.GetPositions(ctx, managed.Robot.Symbol)
				if err != nil {
					if detail.ApiError == "" {
						detail.ApiError = "获取持仓失败: " + err.Error()
					}
				} else {
					for _, pos := range positions {
						if pos.PositionAmt != 0 {
							detail.HasPosition = true
							detail.PositionSide = pos.PositionSide
							detail.PositionAmt = pos.PositionAmt
							detail.EntryPrice = pos.EntryPrice
							detail.UnrealizedPnl = pos.UnrealizedPnl
							break
						}
					}
				}
			}
		} else {
			detail.ApiConnected = false
			detail.ApiError = "交易所实例未初始化"
		}

		// 风险评估已移除（新架构直接使用策略模板参数）
		// RiskEvaluation 和 LastRiskEval 不再由 RiskManager 动态计算
		detail.RiskPreference = managed.Robot.RiskPreference // 直接使用机器人配置

		// 方向信号
		if managed.DirectionSignal != nil {
			detail.DirectionSignal = string(managed.DirectionSignal.Direction)
		}

		detail.LastUpdate = time.Now().Format("2006-01-02 15:04:05")
		status.RobotList = append(status.RobotList, detail)
	}

	return status
}

// getAlertLoggerStatus 获取预警日志服务状态
func (c *cAlertController) getAlertLoggerStatus(ctx context.Context) *admin.AlertLoggerStatus {
	status := &admin.AlertLoggerStatus{
		Running: true,
	}

	// 统计日志数量
	marketStateCount, _ := dao.TradingMarketStateLog.Ctx(ctx).Count()
	riskPreferenceCount, _ := dao.TradingRiskPreferenceLog.Ctx(ctx).Count()
	directionCount, _ := dao.TradingDirectionLog.Ctx(ctx).Count()

	status.MarketStateLogs = int64(marketStateCount)
	status.RiskPreferenceLogs = int64(riskPreferenceCount)
	status.DirectionLogs = int64(directionCount)

	return status
}

// getTradeStatisticsStatus 获取交易统计服务状态
func (c *cAlertController) getTradeStatisticsStatus(ctx context.Context) *admin.TradeStatisticsStatus {
	status := &admin.TradeStatisticsStatus{
		Running: true,
	}

	// 统计今日交易
	today := time.Now().Format("2006-01-02")
	status.TodayTrades, _ = dao.TradingOrder.Ctx(ctx).WhereGTE("created_at", today).Count()
	status.TotalTrades, _ = dao.TradingOrder.Ctx(ctx).Count()

	// 计算胜率和盈亏
	winCount, _ := dao.TradingOrder.Ctx(ctx).WhereGT("profit", 0).Count()
	totalWithProfit, _ := dao.TradingOrder.Ctx(ctx).WhereNotNull("profit").Count()
	if totalWithProfit > 0 {
		status.WinRate = float64(winCount) / float64(totalWithProfit) * 100
	}

	totalProfit, _ := dao.TradingOrder.Ctx(ctx).Sum("profit")
	status.TotalProfit = totalProfit

	todayProfit, _ := dao.TradingOrder.Ctx(ctx).WhereGTE("created_at", today).Sum("profit")
	status.TodayProfit = todayProfit

	return status
}

// GlobalEngineStart 启动全局引擎
func (c *cAlertController) GlobalEngineStart(ctx context.Context, req *admin.GlobalEngineStartReq) (res *admin.GlobalEngineStartRes, err error) {
	res = &admin.GlobalEngineStartRes{}

	err = toogo.GetRobotTaskManager().Start(ctx)
	if err != nil {
		res.Success = false
		res.Message = err.Error()
		return res, nil
	}

	res.Success = true
	res.Message = "全局引擎启动成功"
	return
}

// GlobalEngineStop 停止全局引擎
func (c *cAlertController) GlobalEngineStop(ctx context.Context, req *admin.GlobalEngineStopReq) (res *admin.GlobalEngineStopRes, err error) {
	res = &admin.GlobalEngineStopRes{}

	toogo.GetRobotTaskManager().Stop()

	res.Success = true
	res.Message = "全局引擎已停止"
	return
}

// GlobalEngineRestart 重启全局引擎
func (c *cAlertController) GlobalEngineRestart(ctx context.Context, req *admin.GlobalEngineRestartReq) (res *admin.GlobalEngineRestartRes, err error) {
	res = &admin.GlobalEngineRestartRes{}

	// 先停止
	toogo.GetRobotTaskManager().Stop()

	// 等待1秒
	time.Sleep(time.Second)

	// 再启动
	err = toogo.GetRobotTaskManager().Start(ctx)
	if err != nil {
		res.Success = false
		res.Message = err.Error()
		return res, nil
	}

	res.Success = true
	res.Message = "全局引擎重启成功"
	return
}

// splitKey 分割key
func splitKey(key string) []string {
	idx := 0
	for i, c := range key {
		if c == ':' {
			idx = i
			break
		}
	}
	if idx == 0 {
		return []string{key}
	}
	return []string{key[:idx], key[idx+1:]}
}
