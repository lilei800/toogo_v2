// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package trading

import (
	"context"
	"encoding/json"
	"fmt"
	"hotgo/addons/exchange"
	"hotgo/api/admin/trading"
	"hotgo/internal/dao"
	"hotgo/internal/library/contexts"
	"hotgo/internal/library/market"
	"hotgo/internal/logic/toogo"
	"hotgo/internal/model/do"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type monitorImpl struct{}

var Monitor = &monitorImpl{}

// GetTicker 获取实时行情
func (s *monitorImpl) GetTicker(ctx context.Context, in *input.TradingMonitorTickerInp) (*input.TradingMonitorTickerModel, error) {
	// 获取交易所实例
	exchange, err := ExchangeManager.GetExchange(ctx, in.ApiConfigId)
	if err != nil {
		return nil, err
	}

	// 获取行情
	ticker, err := exchange.GetTicker(ctx, in.Symbol)
	if err != nil {
		return nil, err
	}

	out := &input.TradingMonitorTickerModel{
		Symbol:    ticker.Symbol,
		LastPrice: ticker.LastPrice,
		High24h:   ticker.High24h,
		Low24h:    ticker.Low24h,
		Volume24h: ticker.Volume24h,
		Change24h: ticker.Change24h,
		Timestamp: ticker.Timestamp.Format("2006-01-02 15:04:05"),
	}

	return out, nil
}

// GetMarketState 分析市场状态
func (s *monitorImpl) GetMarketState(ctx context.Context, in *input.TradingMonitorMarketStateInp) (*input.TradingMonitorMarketStateModel, error) {
	// 获取交易所实例
	exchange, err := ExchangeManager.GetExchange(ctx, in.ApiConfigId)
	if err != nil {
		return nil, err
	}

	// 获取实时行情
	ticker, err := exchange.GetTicker(ctx, in.Symbol)
	if err != nil {
		return nil, err
	}

	// 分析市场状态
	analysis := s.analyzeMarket(ticker)

	out := &input.TradingMonitorMarketStateModel{
		Symbol:          ticker.Symbol,
		CurrentPrice:    ticker.LastPrice,
		MarketState:     analysis.MarketState,
		VolatilityIndex: analysis.VolatilityIndex,
		TrendIndex:      analysis.TrendIndex,
		SignalType:      analysis.SignalType,
		SignalStrength:  analysis.SignalStrength,
		RecommendedRisk: analysis.RecommendedRisk,
		Analysis:        analysis.Analysis,
		Timestamp:       time.Now().Format("2006-01-02 15:04:05"),
	}

	return out, nil
}

// MarketAnalysis 市场分析结果
type MarketAnalysis struct {
	MarketState     string  // 市场状态
	VolatilityIndex float64 // 波动率指数
	TrendIndex      float64 // 趋势指数
	SignalType      string  // 信号类型
	SignalStrength  float64 // 信号强度
	RecommendedRisk string  // 推荐风险偏好
	Analysis        string  // 市场分析
	CurrentPrice    float64 // 当前价格
}

// analyzeMarket 分析市场
func (s *monitorImpl) analyzeMarket(ticker *exchange.Ticker) *MarketAnalysis {
	analysis := &MarketAnalysis{}

	// 计算波动率指数 (基于24小时高低价差)
	priceRange := ticker.High24h - ticker.Low24h
	volatilityPercent := (priceRange / ticker.LastPrice) * 100
	analysis.VolatilityIndex = math.Round(volatilityPercent*100) / 100

	// 计算趋势指数 (基于24小时涨跌幅)
	analysis.TrendIndex = math.Round(ticker.Change24h*100) / 100

	// 判断市场状态
	if volatilityPercent >= 5 {
		analysis.MarketState = "high-volatility"
	} else if volatilityPercent <= 1 {
		analysis.MarketState = "low-volatility"
	} else if math.Abs(ticker.Change24h) >= 3 {
		analysis.MarketState = "trend"
	} else {
		analysis.MarketState = "range"
	}

	// 生成信号
	if ticker.Change24h > 2 {
		analysis.SignalType = "long"
		analysis.SignalStrength = math.Min(ticker.Change24h*10, 100)
	} else if ticker.Change24h < -2 {
		analysis.SignalType = "short"
		analysis.SignalStrength = math.Min(math.Abs(ticker.Change24h)*10, 100)
	} else {
		analysis.SignalType = "neutral"
		analysis.SignalStrength = 30
	}

	// 推荐风险偏好
	if volatilityPercent >= 5 {
		analysis.RecommendedRisk = "conservative"
		analysis.Analysis = "市场波动剧烈，建议采用保守策略，降低杠杆，控制仓位"
	} else if volatilityPercent >= 3 {
		analysis.RecommendedRisk = "balanced"
		analysis.Analysis = "市场波动适中，可采用平衡策略，适度杠杆"
	} else {
		analysis.RecommendedRisk = "aggressive"
		analysis.Analysis = "市场波动较小，可采用激进策略，提高收益率"
	}

	return analysis
}

// SaveMonitorLog 保存监控日志
func (s *monitorImpl) SaveMonitorLog(ctx context.Context, robotId int64, symbol string, analysis *MarketAnalysis, actionTaken, actionResult string) error {
	userId := contexts.GetUserId(ctx)
	tenantId := contexts.GetTenantId(ctx)

	if userId <= 0 {
		return gerror.New("用户未登录")
	}

	signalDetailMap := map[string]interface{}{
		"marketState":     analysis.MarketState,
		"volatilityIndex": analysis.VolatilityIndex,
		"trendIndex":      analysis.TrendIndex,
		"signalStrength":  analysis.SignalStrength,
		"recommendedRisk": analysis.RecommendedRisk,
		"analysis":        analysis.Analysis,
		"actionTaken":     actionTaken,
		"actionResult":    actionResult,
	}

	signalDetail, _ := json.Marshal(signalDetailMap)

	logData := &do.TradingMonitorLog{
		TenantId:       tenantId,
		UserId:         userId,
		RobotId:        robotId,
		Symbol:         symbol,
		CurrentPrice:   analysis.CurrentPrice, // 从分析结果中获取当前价格
		SignalType:     analysis.SignalType,
		SignalStrength: analysis.SignalStrength,
		SignalDetail:   string(signalDetail),
		MarketState:    analysis.MarketState,
	}

	_, err := dao.TradingMonitorLog.Ctx(ctx).Data(logData).Insert()
	if err != nil {
		g.Log().Errorf(ctx, "保存监控日志失败: %v", err)
		return err
	}

	return nil
}

// GetMonitorLogs 获取监控日志列表
func (s *monitorImpl) GetMonitorLogs(ctx context.Context, in *input.TradingMonitorLogListInp) (list []*input.TradingMonitorLogListModel, totalCount int, err error) {
	mod := dao.TradingMonitorLog.Ctx(ctx)

	// 租户隔离
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		err = gerror.New("用户未登录")
		return
	}
	mod = mod.Where(dao.TradingMonitorLog.Columns().UserId, memberId)

	// 条件筛选
	if in.RobotId > 0 {
		mod = mod.Where(dao.TradingMonitorLog.Columns().RobotId, in.RobotId)
	}
	if in.SignalType != "" {
		mod = mod.Where(dao.TradingMonitorLog.Columns().SignalType, in.SignalType)
	}
	if in.Symbol != "" {
		mod = mod.Where(dao.TradingMonitorLog.Columns().Symbol, in.Symbol)
	}
	if in.StartDate != "" {
		mod = mod.WhereGTE(dao.TradingMonitorLog.Columns().CreatedAt, in.StartDate)
	}
	if in.EndDate != "" {
		mod = mod.WhereLTE(dao.TradingMonitorLog.Columns().CreatedAt, in.EndDate+" 23:59:59")
	}

	totalCount, err = mod.Count()
	if err != nil {
		return nil, 0, err
	}

	if totalCount == 0 {
		return
	}

	// 查询日志列表
	var logs []*entity.TradingMonitorLog
	err = mod.Page(in.Page, in.PageSize).
		Order(dao.TradingMonitorLog.Columns().CreatedAt + " DESC").
		Scan(&logs)

	if err != nil {
		return nil, 0, err
	}

	// 获取机器人名称
	robotIds := make([]int64, 0)
	for _, log := range logs {
		robotIds = append(robotIds, log.RobotId)
	}

	var robots []*entity.TradingRobot
	if len(robotIds) > 0 {
		err = dao.TradingRobot.Ctx(ctx).
			WhereIn(dao.TradingRobot.Columns().Id, robotIds).
			Scan(&robots)

		if err != nil {
			return nil, 0, err
		}
	}

	robotMap := make(map[int64]string)
	for _, robot := range robots {
		robotMap[robot.Id] = robot.RobotName
	}

	// 转换为输出模型
	list = make([]*input.TradingMonitorLogListModel, 0, len(logs))
	for _, log := range logs {
		// 从SignalDetail中解析额外信息
		var signalDetailMap map[string]interface{}
		_ = json.Unmarshal([]byte(log.SignalDetail), &signalDetailMap)

		actionTaken := ""
		actionResult := ""
		var volatilityIndex, trendIndex float64

		if signalDetailMap != nil {
			if v, ok := signalDetailMap["actionTaken"].(string); ok {
				actionTaken = v
			}
			if v, ok := signalDetailMap["actionResult"].(string); ok {
				actionResult = v
			}
			if v, ok := signalDetailMap["volatilityIndex"].(float64); ok {
				volatilityIndex = v
			}
			if v, ok := signalDetailMap["trendIndex"].(float64); ok {
				trendIndex = v
			}
		}

		item := &input.TradingMonitorLogListModel{
			Id:              log.Id,
			RobotId:         log.RobotId,
			RobotName:       robotMap[log.RobotId],
			Symbol:          log.Symbol,
			CurrentPrice:    log.CurrentPrice,
			SignalType:      log.SignalType,
			SignalStrength:  log.SignalStrength,
			SignalDetail:    log.SignalDetail,
			ActionTaken:     actionTaken,
			ActionResult:    actionResult,
			VolatilityIndex: volatilityIndex,
			TrendIndex:      trendIndex,
			CreateTime:      gtime.New(log.CreatedAt),
		}
		list = append(list, item)
	}

	return
}

// GetRobotAnalysis 获取机器人实时分析数据
func (s *monitorImpl) GetRobotAnalysis(ctx context.Context, robotId int64) (*trading.MonitorRobotAnalysisRes, error) {
	// 获取机器人信息
	var robot *entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, robotId).Scan(&robot)
	if err != nil {
		return nil, gerror.Newf("获取机器人信息失败: %v", err)
	}
	if robot == nil {
		return nil, gerror.New("机器人不存在")
	}

	res := &trading.MonitorRobotAnalysisRes{
		RobotId:   robotId,
		Connected: false,
	}

	// 获取API配置以确定平台
	var apiConfig *entity.TradingApiConfig
	err = dao.TradingApiConfig.Ctx(ctx).Where(dao.TradingApiConfig.Columns().Id, robot.ApiConfigId).Scan(&apiConfig)
	if err != nil || apiConfig == nil {
		res.ConnectionError = "API配置不存在"
		return res, nil
	}
	platform := apiConfig.Platform
	platform = strings.ToLower(strings.TrimSpace(platform))
	robot.Symbol = strings.ToUpper(strings.TrimSpace(robot.Symbol))

	// ====== 优先从RobotEngine获取窗口信号数据 ======
	engine := toogo.GetRobotTaskManager().GetEngine(robotId)
	if engine != nil {
		// 获取窗口价格数据
		res.PriceWindow = s.buildPriceWindow(engine)
		res.SignalHistory = s.buildSignalHistory(engine)

		// 如果引擎有最新信号，使用窗口信号
		if engine.LastSignal != nil {
			res.Signal = s.buildWindowSignal(engine)
		}
	}

	// ====== 优先从全局MarketServiceManager获取缓存数据 ======
	ticker := market.GetMarketServiceManager().GetTicker(platform, robot.Symbol)

	// 如果缓存中有数据且数据新鲜，直接使用
	if ticker != nil {
		// 统一口径：列表页“连接状态/当前价”优先使用 LastPrice；缺失则用 MarkPrice 兜底
		lastPrice := ticker.LastPrice
		if lastPrice <= 0 {
			lastPrice = ticker.EffectiveMarkPrice()
		}

		// 设置连接成功
		res.Connected = true

		// 填充行情数据
		res.Ticker = &trading.RobotTickerInfo{
			Symbol:        ticker.Symbol,
			LastPrice:     lastPrice,
			High24h:       ticker.High24h,
			Low24h:        ticker.Low24h,
			Volume24h:     ticker.Volume24h,
			Change24h:     ticker.Change24h,
			ChangePercent: ticker.PriceChangePercent,
			Timestamp:     time.Unix(ticker.Timestamp/1000, 0).Format("2006-01-02 15:04:05"),
		}

		// 转换为内部Ticker格式用于分析
		exchangeTicker := &exchange.Ticker{
			Symbol:    ticker.Symbol,
			LastPrice: lastPrice,
			High24h:   ticker.High24h,
			Low24h:    ticker.Low24h,
			Volume24h: ticker.Volume24h,
			Change24h: ticker.Change24h,
		}

		// 进行市场分析
		analysis := s.analyzeMarket(exchangeTicker)

		// 填充市场分析数据
		res.Market = s.buildMarketAnalysis(exchangeTicker, analysis)

		// 填充风险评估数据
		res.Risk = s.buildRiskEvaluation(exchangeTicker, analysis, robot)

		// 如果还没有窗口信号，使用传统信号
		if res.Signal == nil {
			res.Signal = s.buildDirectionSignal(exchangeTicker, analysis)
		} else {
			// 补充传统信号字段到窗口信号
			traditionalSignal := s.buildDirectionSignal(exchangeTicker, analysis)
			if res.Signal.RiskRewardRatio == 0 {
				res.Signal.RiskRewardRatio = traditionalSignal.RiskRewardRatio
			}
			if res.Signal.TimeWindow == "" {
				res.Signal.TimeWindow = traditionalSignal.TimeWindow
			}
		}

		// 尝试获取账户信息（可能失败，但不影响连接状态）
		exchangeInst, err := ExchangeManager.GetExchange(ctx, robot.ApiConfigId)
		if err == nil {
			res.Account = s.buildAccountInfo(ctx, exchangeInst, robot)
		} else {
			// 账户信息获取失败，使用空数据
			res.Account = &trading.RobotAccountInfo{}
		}

		// 填充机器人配置信息
		res.Config = s.buildConfigInfo(ctx, robot)

		// 多周期市场状态实时明细（新算法 + 平滑机制）
		res.MarketStateRealtime = s.buildMarketStateRealtime(platform, robot.Symbol)

		return res, nil
	}

	// ====== 缓存无数据时，返回连接失败状态 ======
	// 不直接请求API，避免超时
	res.ConnectionError = "市场数据未就绪，请稍后"
	res.Connected = false

	// 填充机器人配置信息（即使无连接也显示配置）
	res.Config = s.buildConfigInfo(ctx, robot)

	return res, nil
}

// buildMarketStateRealtime 构建多周期市场状态实时明细（新算法 + 平滑机制）
func (s *monitorImpl) buildMarketStateRealtime(platform, symbol string) *trading.RobotMarketStateRealtime {
	platform = strings.ToLower(strings.TrimSpace(platform))
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if platform == "" || symbol == "" {
		return nil
	}
	analysis := market.GetMarketAnalyzer().GetAnalysis(platform, symbol)
	if analysis == nil {
		return nil
	}

	normalize := func(state string) string {
		if state == "range" {
			return "volatile"
		}
		return state
	}

	// 固定顺序输出，便于前端展示
	order := map[string]int{"1m": 1, "5m": 2, "15m": 3, "30m": 4, "1h": 5}
	timeframes := make([]*trading.RobotMarketStateTimeframe, 0, len(analysis.TimeframeAnalysis))
	for interval, tf := range analysis.TimeframeAnalysis {
		if tf == nil {
			continue
		}
		timeframes = append(timeframes, &trading.RobotMarketStateTimeframe{
			Interval:      interval,
			Weight:        tf.Weight,
			Open:          tf.Open,
			High:          tf.High,
			Low:           tf.Low,
			Close:         tf.Close,
			Delta:         tf.Delta,
			V:             tf.V,
			D:             tf.D,
			RawState:      normalize(tf.RawState),
			SmoothedState: normalize(tf.SmoothedState),
			SmoothedConf:  tf.SmoothedConf,
		})
	}
	sort.Slice(timeframes, func(i, j int) bool {
		return order[timeframes[i].Interval] < order[timeframes[j].Interval]
	})

	finalState := normalize(string(analysis.MarketState))

	// 生成“播报”摘要（可直接展示/复制）
	broadcast := fmt.Sprintf("最终=%s(%.2f)", finalState, analysis.MarketStateConf)
	for _, tf := range timeframes {
		broadcast += fmt.Sprintf(" | %s:%s V=%.2f D=%.2f w=%.2f", tf.Interval, tf.SmoothedState, tf.V, tf.D, tf.Weight)
	}

	return &trading.RobotMarketStateRealtime{
		Platform:   platform,
		Symbol:     symbol,
		State:      finalState,
		Confidence: analysis.MarketStateConf,
		VoteRatio:  analysis.VoteRatio,
		UpdatedAt:  analysis.UpdatedAt.Format("2006-01-02 15:04:05"),
		Timeframes: timeframes,
		Broadcast:  broadcast,
	}
}

// buildAccountInfo 构建账户信息
// 【优化】从RobotEngine缓存获取数据，避免频繁调用API
func (s *monitorImpl) buildAccountInfo(ctx context.Context, exchangeInst exchange.IExchange, robot *entity.TradingRobot) *trading.RobotAccountInfo {
	// 【优化】从RobotEngine获取缓存数据，避免每秒调用API
	engine := toogo.GetRobotTaskManager().GetEngine(robot.Id)
	if engine == nil {
		// 引擎不存在，返回nil避免前端闪烁（不返回全0的空对象）
		return nil
	}

	// 余额策略：
	// - 引擎已取消“定时同步余额”（只在下单/交易后刷新），但列表页仍需要展示账户权益/可用余额
	// - 因此这里做一个轻量的“按需刷新”：余额为空或过期时，使用 GetBalanceSmart 节流拉一次（内部有去重）
	// - 若依然拿不到余额则返回 nil（避免前端显示 0.00 的假数据/闪烁）
	cachedBal, cachedPositions, lastBalAt, _ := engine.GetAccountSnapshot()

	// 5秒内视为新鲜；为空或过期则触发一次刷新
	if cachedBal == nil || time.Since(lastBalAt) >= 5*time.Second {
		if bal, _ := engine.GetBalanceSmart(ctx, 5*time.Second); bal != nil {
			cachedBal = bal
		}
	}
	if cachedBal == nil {
		return nil
	}

	account := &trading.RobotAccountInfo{}

	// 从引擎缓存获取余额
	// 账户权益 = 总余额（包含未实现盈亏）
	account.AccountEquity = cachedBal.TotalBalance
	account.TotalBalance = cachedBal.TotalBalance // 兼容旧字段
	account.AvailableBalance = cachedBal.AvailableBalance

	// 从引擎缓存获取持仓
	if cachedPositions != nil {
		for _, pos := range cachedPositions {
			if pos == nil || pos.PositionAmt == 0 {
				continue
			}
			// 计算保证金 = 持仓价值 / 杠杆
			posValue := math.Abs(pos.PositionAmt) * pos.EntryPrice
			if pos.Leverage > 0 {
				account.UsedMargin += posValue / float64(pos.Leverage)
			}
			account.UnrealizedPnl += pos.UnrealizedPnl
		}
	}

	// 计算钱包余额：账户权益 - 未实现盈亏
	// AccountEquity = WalletBalance + UnrealizedPnl
	account.WalletBalance = account.AccountEquity - account.UnrealizedPnl

	// 计算保证金率
	if account.AccountEquity > 0 {
		account.MarginRatio = (account.UsedMargin / account.AccountEquity) * 100
	}

	account.TodayPnl = account.UnrealizedPnl
	return account
}

// buildConfigInfo 构建机器人配置信息
// 【重新设计】统一逻辑：无论机器人是否启动，都使用相同的策略参数获取逻辑
// 1. 获取全局实时市场状态
// 2. 根据创建机器人时提交的映射关系选择风险偏好
// 3. 根据实时市场状态+风险偏好获取策略组中对应的策略
func (s *monitorImpl) buildConfigInfo(ctx context.Context, robot *entity.TradingRobot) *trading.RobotConfigInfo {
	config := &trading.RobotConfigInfo{
		AutoTradeEnabled: robot.AutoTradeEnabled == 1,
		AutoCloseEnabled: robot.AutoCloseEnabled == 1,
		DualSidePosition: robot.DualSidePosition == 1,
		UseMonitorSignal: robot.UseMonitorSignal == 1,
		MaxProfit:        robot.MaxProfitTarget,
		MaxLoss:          robot.MaxLossAmount,
		TotalProfit:      robot.TotalProfit,
		LongCount:        robot.LongCount,
		ShortCount:       robot.ShortCount,
	}

	// 计算运行时长
	if robot.StartTime != nil {
		config.StartTime = robot.StartTime.Format("2006-01-02 15:04:05")
		config.RuntimeSeconds = int64(time.Since(robot.StartTime.Time).Seconds())
	}

	// 【步骤1】获取全局实时市场状态
	// 获取平台信息（优先从引擎获取，否则从API配置获取）
	var platform string
	engine := toogo.GetRobotTaskManager().GetEngine(robot.Id)
	if engine != nil {
		platform = engine.Platform
	} else {
		// 引擎不存在时，从API配置获取平台信息
		var apiConfig *entity.TradingApiConfig
		if err := dao.TradingApiConfig.Ctx(ctx).Where(dao.TradingApiConfig.Columns().Id, robot.ApiConfigId).Scan(&apiConfig); err == nil && apiConfig != nil {
			platform = apiConfig.Platform
		}
	}

	platform = strings.ToLower(strings.TrimSpace(platform))
	robot.Symbol = strings.ToUpper(strings.TrimSpace(robot.Symbol))

	if platform == "" {
		errMsg := "无法获取交易平台信息"
		g.Log().Errorf(ctx, "[Monitor] robotId=%d %s", robot.Id, errMsg)
		config.ErrorMessage = errMsg
		return config
	}

	// 从全局市场分析器获取实时市场状态
	globalAnalysis := market.GetMarketAnalyzer().GetAnalysis(platform, robot.Symbol)
	marketState := ""
	if globalAnalysis != nil {
		marketState = string(globalAnalysis.MarketState)
		if marketState != "" {
			marketState = normalizeMarketState(marketState)
			config.MarketState = marketState
		}
	}

	// 如果全局市场分析器没有数据，返回错误
	if marketState == "" {
		errMsg := "全局市场分析器未返回市场状态数据，请检查市场分析服务是否正常运行"
		g.Log().Errorf(ctx, "[Monitor] robotId=%d %s", robot.Id, errMsg)
		config.ErrorMessage = errMsg
		return config
	}

	// 【步骤2】根据创建机器人时提交的映射关系选择风险偏好
	// 【重要】从 remark 字段解析映射关系（创建时保存的独立映射关系）
	riskPref := ""
	if robot.Remark != "" {
		var mapping map[string]string
		if err := json.Unmarshal([]byte(robot.Remark), &mapping); err == nil {
			// 规范化市场状态键，确保与映射关系中的key一致
			normalizedMarketState := normalizeMarketState(marketState)
			if pref, ok := mapping[normalizedMarketState]; ok && pref != "" {
				riskPref = pref
				g.Log().Debugf(ctx, "[Monitor] robotId=%d 从 remark 字段映射关系获取风险偏好: 市场状态=%s → 风险偏好=%s",
					robot.Id, normalizedMarketState, riskPref)
			}
		} else {
			g.Log().Warningf(ctx, "[Monitor] robotId=%d remark 字段不是有效的JSON格式: %s，错误: %v", robot.Id, robot.Remark, err)
		}
	}

	// 如果映射关系中没有找到，返回错误
	if riskPref == "" {
		errMsg := fmt.Sprintf("市场状态=%s 在映射关系中未找到对应的风险偏好，请检查机器人的风险配置映射关系（remark=%s）", marketState, robot.Remark)
		g.Log().Errorf(ctx, "[Monitor] robotId=%d %s", robot.Id, errMsg)
		config.ErrorMessage = errMsg
		config.MarketState = marketState
		return config
	}
	config.RiskPreference = riskPref

	// 【步骤3】根据实时市场状态+风险偏好获取策略组中对应的策略
	// 获取策略组ID（优先级：机器人.StrategyGroupId > CurrentStrategy.groupId）
	groupId := robot.StrategyGroupId
	if groupId == 0 && robot.CurrentStrategy != "" {
		var strategyData map[string]interface{}
		if err := json.Unmarshal([]byte(robot.CurrentStrategy), &strategyData); err == nil {
			if gid, ok := strategyData["groupId"].(float64); ok {
				groupId = int64(gid)
			}
		}
	}

	if groupId == 0 {
		errMsg := "机器人未绑定策略组ID，无法加载策略参数"
		g.Log().Errorf(ctx, "[Monitor] robotId=%d %s", robot.Id, errMsg)
		config.ErrorMessage = errMsg
		return config
	}

	// 从策略模板表查询对应的策略（尝试多种市场状态名称，兼容旧数据）
	marketStatesToTry := []string{marketState}
	// 添加兼容格式
	if marketState == "volatile" {
		marketStatesToTry = append(marketStatesToTry, "range")
	} else if marketState == "range" {
		marketStatesToTry = append(marketStatesToTry, "volatile")
	} else if marketState == "high_vol" {
		marketStatesToTry = append(marketStatesToTry, "high-volatility")
	} else if marketState == "low_vol" {
		marketStatesToTry = append(marketStatesToTry, "low-volatility")
	}

	var strategy *entity.TradingStrategyTemplate
	var queryErr error
	for _, ms := range marketStatesToTry {
		queryErr = dao.TradingStrategyTemplate.Ctx(ctx).
			Where("group_id", groupId).
			Where(dao.TradingStrategyTemplate.Columns().MarketState, ms).
			Where(dao.TradingStrategyTemplate.Columns().RiskPreference, riskPref).
			Scan(&strategy)
		if queryErr == nil && strategy != nil {
			break
		}
	}

	if queryErr != nil || strategy == nil {
		errMsg := fmt.Sprintf("找不到策略模板: groupId=%d, marketState=%s, riskPreference=%s，请检查策略模板配置",
			groupId, marketState, riskPref)
		g.Log().Errorf(ctx, "[Monitor] robotId=%d %s", robot.Id, errMsg)
		config.ErrorMessage = errMsg
		return config
	}

	// 填充策略参数（直接使用策略模板的值）
	config.Leverage = strategy.Leverage
	config.MarginPercent = strategy.MarginPercent
	config.StopLossPercent = strategy.StopLossPercent
	config.AutoStartRetreat = strategy.AutoStartRetreatPercent
	config.TakeProfitPercent = strategy.ProfitRetreatPercent
	config.TimeWindow = strategy.MonitorWindow
	config.Threshold = strategy.VolatilityThreshold

	// 填充策略组信息
	config.StrategyGroupId = groupId
	config.StrategyName = strategy.StrategyName

	// 查询策略组名称
	var strategyGroup *entity.TradingStrategyGroup
	if err := g.DB().Model("hg_trading_strategy_group").Ctx(ctx).Where("id", groupId).Scan(&strategyGroup); err == nil && strategyGroup != nil {
		config.StrategyGroupName = strategyGroup.GroupName
	}

	g.Log().Infof(ctx, "[Monitor] robotId=%d 策略参数已加载: market=%s, risk=%s, 策略组=%s, 策略=%s, 杠杆=%d, 保证金=%.1f%%, 止损=%.1f%%, 启动止盈=%.1f%%, 止盈回撤=%.1f%%, 窗口=%d, 波动=%.1f",
		robot.Id, marketState, riskPref, config.StrategyGroupName, config.StrategyName, config.Leverage, config.MarginPercent,
		config.StopLossPercent, config.AutoStartRetreat, config.TakeProfitPercent,
		config.TimeWindow, config.Threshold)

	return config
}

// buildMarketAnalysis 构建市场分析数据
func (s *monitorImpl) buildMarketAnalysis(ticker *exchange.Ticker, analysis *MarketAnalysis) *trading.RobotMarketAnalysis {
	// 转换市场状态
	state := "RANGING"
	switch analysis.MarketState {
	case "trend":
		if analysis.TrendIndex > 0 {
			if analysis.TrendIndex > 3 {
				state = "STRONG_UPTREND"
			} else {
				state = "MILD_UPTREND"
			}
		} else {
			if analysis.TrendIndex < -3 {
				state = "STRONG_DOWNTREND"
			} else {
				state = "MILD_DOWNTREND"
			}
		}
	case "high-volatility":
		state = "HIGH_VOLATILITY"
	case "low-volatility":
		state = "LOW_VOLATILITY"
	}

	// 转换波动等级
	volatilityLevel := "NORMAL"
	if analysis.VolatilityIndex < 1 {
		volatilityLevel = "LOW"
	} else if analysis.VolatilityIndex > 5 {
		volatilityLevel = "EXTREME"
	} else if analysis.VolatilityIndex > 3 {
		volatilityLevel = "HIGH"
	}

	// 生成多周期信号（基于当前行情模拟）
	// 注意：这个逻辑应该从 RobotEngine 获取真实的多周期分析结果
	// 这里只是临时方案，如果引擎有数据应该优先使用引擎数据
	change := ticker.Change24h
	timeFrameSignals := map[string]string{
		"5m":  s.getSignalText(change * 0.4),
		"15m": s.getSignalText(change * 0.6),
		"1h":  s.getSignalText(change),
	}

	// 转换建议操作
	suggestAction := "WAIT"
	switch analysis.SignalType {
	case "long":
		if analysis.SignalStrength > 70 {
			suggestAction = "STRONG_BUY"
		} else {
			suggestAction = "BUY"
		}
	case "short":
		if analysis.SignalStrength > 70 {
			suggestAction = "STRONG_SELL"
		} else {
			suggestAction = "SELL"
		}
	}
	if analysis.VolatilityIndex > 5 {
		suggestAction = "CAUTION"
	}

	return &trading.RobotMarketAnalysis{
		State:            state,
		TrendScore:       analysis.TrendIndex * 10,
		VolatilityLevel:  volatilityLevel,
		Confidence:       analysis.SignalStrength / 100,
		SuggestAction:    suggestAction,
		TimeFrameSignals: timeFrameSignals,
	}
}

// buildMarketAnalysisFromEngine 从 RobotEngine 构建市场分析数据（确保与实时信号一致）
func (s *monitorImpl) buildMarketAnalysisFromEngine(engine *toogo.RobotEngine) *trading.RobotMarketAnalysis {
	analysis := engine.LastAnalysis
	if analysis == nil {
		return &trading.RobotMarketAnalysis{
			State: "range",
		}
	}

	// 转换市场状态（使用与 TradingEngine 相同的格式）
	state := analysis.MarketState
	if state == "" {
		state = "range"
	}

	// 转换波动等级
	volatilityLevel := "NORMAL"
	if analysis.Volatility < 0.2 {
		volatilityLevel = "LOW"
	} else if analysis.Volatility > 1.2 {
		volatilityLevel = "EXTREME"
	} else if analysis.Volatility > 0.8 {
		volatilityLevel = "HIGH"
	}

	// 生成多周期信号（如果有的话）
	// 使用 Direction 和 Strength 来判断，而不是 TrendStrength
	timeFrameSignals := make(map[string]string)
	if analysis.TimeframeScores != nil && len(analysis.TimeframeScores) > 0 {
		for tf, score := range analysis.TimeframeScores {
			if score == nil {
				continue
			}
			// Direction: "up"/"down"/"neutral"
			// Strength: 0-100 (方向强度，50+表示有明确方向)
			// TrendStrength: 0-1 (趋势强度，用于辅助判断)
			// 判断逻辑：方向明确且强度足够才显示看多/看空，否则显示震荡
			if score.Direction == "up" && score.Strength >= 50 {
				// 根据强度显示不同的信号强度
				if score.Strength >= 70 {
					timeFrameSignals[tf] = "强看多"
				} else {
					timeFrameSignals[tf] = "看多"
				}
			} else if score.Direction == "down" && score.Strength >= 50 {
				if score.Strength >= 70 {
					timeFrameSignals[tf] = "强看空"
				} else {
					timeFrameSignals[tf] = "看空"
				}
			} else {
				// Direction 为 "neutral" 或 Strength < 50，显示震荡
				timeFrameSignals[tf] = "震荡"
			}
		}
	}

	// 转换建议操作
	suggestAction := "WAIT"
	if analysis.TrendDirection == "up" && analysis.MarketStateConf > 0.6 {
		suggestAction = "BUY"
	} else if analysis.TrendDirection == "down" && analysis.MarketStateConf > 0.6 {
		suggestAction = "SELL"
	}

	// 计算趋势评分（基于趋势强度）
	trendScore := analysis.TrendStrength
	if analysis.TrendDirection == "down" {
		trendScore = -trendScore
	}

	return &trading.RobotMarketAnalysis{
		State:            state,
		TrendScore:       trendScore,
		VolatilityLevel:  volatilityLevel,
		Confidence:       analysis.MarketStateConf,
		SuggestAction:    suggestAction,
		TimeFrameSignals: timeFrameSignals,
	}
}

// convertEngineAnalysisToMarketAnalysis 将引擎分析结果转换为市场分析格式（用于风险评估）
func (s *monitorImpl) convertEngineAnalysisToMarketAnalysis(analysis *toogo.RobotMarketAnalysis) *MarketAnalysis {
	// 确定信号类型
	signalType := "neutral"
	signalStrength := 50.0
	if analysis.TrendDirection == "up" {
		signalType = "long"
		signalStrength = analysis.TrendStrength
	} else if analysis.TrendDirection == "down" {
		signalType = "short"
		signalStrength = analysis.TrendStrength
	}

	// 确定波动指数
	volatilityIndex := 2.0
	if analysis.Volatility < 0.2 {
		volatilityIndex = 1.0
	} else if analysis.Volatility > 1.2 {
		volatilityIndex = 5.0
	} else if analysis.Volatility > 0.8 {
		volatilityIndex = 4.0
	}

	// 确定趋势指数
	trendIndex := 0.0
	if analysis.TrendDirection == "up" {
		trendIndex = analysis.TrendStrength / 10
	} else if analysis.TrendDirection == "down" {
		trendIndex = -analysis.TrendStrength / 10
	}

	return &MarketAnalysis{
		MarketState:     analysis.MarketState,
		SignalType:      signalType,
		SignalStrength:  signalStrength,
		TrendIndex:      trendIndex,
		VolatilityIndex: volatilityIndex,
		RecommendedRisk: "balanced", // 默认平衡型
	}
}

// getSignalText 根据变化量获取信号文本
func (s *monitorImpl) getSignalText(change float64) string {
	if change > 0.5 {
		return "看多"
	} else if change < -0.5 {
		return "看空"
	}
	return "震荡"
}

// buildRiskEvaluation 构建风险评估数据
func (s *monitorImpl) buildRiskEvaluation(ticker *exchange.Ticker, analysis *MarketAnalysis, robot *entity.TradingRobot) *trading.RobotRiskEvaluation {
	// 计算胜算概率
	winProbability := 50.0 + analysis.SignalStrength*0.3
	if analysis.VolatilityIndex > 4 {
		winProbability = math.Max(30, winProbability-20)
	} else if analysis.SignalStrength > 70 && analysis.VolatilityIndex < 2 {
		winProbability = math.Min(85, winProbability+10)
	}

	// 根据风险偏好确定参数
	preferenceType := analysis.RecommendedRisk
	var suggestLeverage int
	var suggestPosition float64
	var suggestStopLoss float64

	switch preferenceType {
	case "conservative":
		suggestLeverage = 5
		suggestPosition = 0.1
		suggestStopLoss = 3.0
	case "aggressive":
		suggestLeverage = 20
		suggestPosition = 0.3
		suggestStopLoss = 8.0
	default: // balanced
		suggestLeverage = 10
		suggestPosition = 0.2
		suggestStopLoss = 5.0
	}

	// 生成风险提示
	reasons := make([]string, 0)
	if analysis.VolatilityIndex > 3 {
		reasons = append(reasons, "市场波动较大，建议降低风险")
	} else {
		reasons = append(reasons, "市场波动正常")
	}
	if analysis.SignalStrength > 60 {
		reasons = append(reasons, "信号强度较高，可适当加仓")
	} else {
		reasons = append(reasons, "信号强度一般，建议观望")
	}
	if winProbability > 60 {
		reasons = append(reasons, "胜算概率较高")
	} else {
		reasons = append(reasons, "胜算概率一般")
	}

	return &trading.RobotRiskEvaluation{
		PreferenceType:  preferenceType,
		WinProbability:  math.Round(winProbability*10) / 10,
		SuggestLeverage: suggestLeverage,
		SuggestPosition: suggestPosition,
		SuggestStopLoss: suggestStopLoss,
		Reasons:         reasons,
	}
}

// buildDirectionSignal 构建方向信号数据
func (s *monitorImpl) buildDirectionSignal(ticker *exchange.Ticker, analysis *MarketAnalysis) *trading.RobotDirectionSignal {
	direction := "WAIT"
	switch analysis.SignalType {
	case "long":
		direction = "LONG"
	case "short":
		direction = "SHORT"
	}

	// 计算风险收益比
	riskRewardRatio := 0.0
	if direction != "WAIT" {
		riskRewardRatio = 2 + math.Abs(analysis.TrendIndex)*0.2
	}

	// 建议持仓时间
	timeWindow := "15-60分钟"
	if analysis.VolatilityIndex > 3 {
		timeWindow = "5-15分钟"
	} else if analysis.VolatilityIndex < 1 {
		timeWindow = "1-4小时"
	}

	// 生成操作建议
	recommendation := ""
	if direction == "LONG" {
		leverage := 10
		stopLoss := 3.0
		if analysis.VolatilityIndex > 2 {
			leverage = 5
			stopLoss = 5.0
		}
		recommendation = fmt.Sprintf("建议做多，使用%dx杠杆，止损%.0f%%", leverage, stopLoss)
	} else if direction == "SHORT" {
		leverage := 10
		stopLoss := 3.0
		if analysis.VolatilityIndex > 2 {
			leverage = 5
			stopLoss = 5.0
		}
		recommendation = fmt.Sprintf("建议做空，使用%dx杠杆，止损%.0f%%", leverage, stopLoss)
	} else {
		recommendation = "当前信号不明确，建议等待更好的入场时机"
	}

	return &trading.RobotDirectionSignal{
		Direction:       direction,
		SignalStrength:  analysis.SignalStrength,
		Confidence:      analysis.SignalStrength / 100,
		RiskRewardRatio: math.Round(riskRewardRatio*100) / 100,
		TimeWindow:      timeWindow,
		Recommendation:  recommendation,
		SignalType:      "analysis",
	}
}

// buildWindowSignal 构建窗口信号数据
func (s *monitorImpl) buildWindowSignal(engine *toogo.RobotEngine) *trading.RobotDirectionSignal {
	signal := engine.LastSignal
	if signal == nil {
		return nil
	}

	// 生成操作建议
	recommendation := signal.Reason
	if signal.Direction == "LONG" && signal.Action == "OPEN_LONG" {
		recommendation = "窗口信号触发做多，建议顺势开仓"
	} else if signal.Direction == "SHORT" && signal.Action == "OPEN_SHORT" {
		recommendation = "窗口信号触发做空，建议顺势开仓"
	} else if signal.Direction == "NEUTRAL" {
		recommendation = "监控中，等待信号触发"
	}

	// 监控窗口配置
	monitorWindow := 60
	strategyThreshold := 50.0
	if engine.MonitorConfig != nil {
		monitorWindow = engine.MonitorConfig.Window
		strategyThreshold = engine.MonitorConfig.Threshold
	}

	// 获取当前市场状态和风险偏好（从RobotEngine的分析结果）
	currentMarketState := ""
	currentRiskPref := ""

	// 【优化】从全局市场分析器获取市场状态
	if currentMarketState == "" {
		globalAnalysis := market.GetMarketAnalyzer().GetAnalysis(engine.Platform, engine.Robot.Symbol)
		if globalAnalysis != nil {
			currentMarketState = string(globalAnalysis.MarketState)
		}
	}

	// 【统一数据源】从 remark 字段获取映射关系（与 buildConfigInfo 一致）
	// 确保 signal.currentRiskPref 和 config.riskPreference 使用相同的数据源
	if currentRiskPref == "" && currentMarketState != "" {
		// 规范化市场状态
		normalizedMarketState := normalizeMarketState(currentMarketState)

		// 从 remark 字段解析映射关系（创建时保存的独立映射关系）
		if engine.Robot.Remark != "" {
			var mapping map[string]string
			if err := json.Unmarshal([]byte(engine.Robot.Remark), &mapping); err == nil {
				if pref, ok := mapping[normalizedMarketState]; ok && pref != "" {
					currentRiskPref = pref
				}
			}
		}
	}

	// 【严格模式】不使用降级方案，保持与 buildConfigInfo 一致
	// 如果映射关系中找不到，currentRiskPref 保持为空

	return &trading.RobotDirectionSignal{
		Direction:          signal.Direction,
		SignalStrength:     signal.Strength,
		Confidence:         signal.Confidence / 100,
		RiskRewardRatio:    2.0,
		TimeWindow:         fmt.Sprintf("%d秒窗口", monitorWindow),
		Recommendation:     recommendation,
		SignalType:         signal.SignalType,
		WindowMaxPrice:     signal.WindowMaxPrice,
		WindowMinPrice:     signal.WindowMinPrice,
		CurrentPrice:       signal.CurrentPrice,
		DistanceFromMin:    signal.DistanceFromMin,
		DistanceFromMax:    signal.DistanceFromMax,
		SignalThreshold:    signal.SignalThreshold,
		SignalProgress:     signal.SignalProgress,
		Action:             signal.Action,
		Reason:             signal.Reason,
		MonitorWindow:      monitorWindow,
		PricePointsCount:   len(engine.PriceWindow),
		StrategyWindow:     monitorWindow,
		StrategyThreshold:  strategyThreshold,
		CurrentMarketState: currentMarketState,
		CurrentRiskPref:    currentRiskPref,
	}
}

// buildPriceWindow 构建窗口价格数据
func (s *monitorImpl) buildPriceWindow(engine *toogo.RobotEngine) []*trading.PricePoint {
	engine.GetPriceLock().RLock()
	defer engine.GetPriceLock().RUnlock()

	priceWindow := make([]*trading.PricePoint, 0, len(engine.PriceWindow))
	for _, p := range engine.PriceWindow {
		priceWindow = append(priceWindow, &trading.PricePoint{
			Timestamp: p.Timestamp,
			Price:     p.Price,
		})
	}
	return priceWindow
}

// buildSignalHistory 构建信号历史数据
func (s *monitorImpl) buildSignalHistory(engine *toogo.RobotEngine) []*trading.SignalHistoryItem {
	engine.GetPriceLock().RLock()
	defer engine.GetPriceLock().RUnlock()

	signalHistory := make([]*trading.SignalHistoryItem, 0, len(engine.SignalHistory))
	for _, h := range engine.SignalHistory {
		signalHistory = append(signalHistory, &trading.SignalHistoryItem{
			Timestamp: h.Timestamp,
			Signal:    h.Signal,
		})
	}
	return signalHistory
}

// GetBatchRobotAnalysis 批量获取机器人分析数据
func (s *monitorImpl) GetBatchRobotAnalysis(ctx context.Context, robotIdsStr string) (*trading.MonitorBatchRobotAnalysisRes, error) {
	robotIdStrs := strings.Split(robotIdsStr, ",")
	result := &trading.MonitorBatchRobotAnalysisRes{
		List: make([]*trading.MonitorRobotAnalysisRes, 0),
	}

	for _, idStr := range robotIdStrs {
		robotId, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			continue
		}

		analysis, err := s.GetRobotAnalysis(ctx, robotId)
		if err != nil {
			// 添加失败的机器人状态
			result.List = append(result.List, &trading.MonitorRobotAnalysisRes{
				RobotId:         robotId,
				Connected:       false,
				ConnectionError: err.Error(),
			})
			continue
		}
		result.List = append(result.List, analysis)
	}

	return result, nil
}

// GetKline 获取K线数据
func (s *monitorImpl) GetKline(ctx context.Context, apiConfigId int64, symbol, interval string, limit int) (*trading.MonitorKlineRes, error) {
	// 获取交易所实例
	exchangeInst, err := ExchangeManager.GetExchange(ctx, apiConfigId)
	if err != nil {
		return nil, err
	}

	// 默认值
	if interval == "" {
		interval = "1h"
	}
	if limit <= 0 {
		limit = 100
	}

	// 获取K线数据
	klines, err := exchangeInst.GetKline(ctx, symbol, interval, limit)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	list := make([]*trading.KlineDataItem, 0, len(klines))
	for _, k := range klines {
		list = append(list, &trading.KlineDataItem{
			OpenTime:  k.OpenTime.UnixMilli(),
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
			CloseTime: k.CloseTime.UnixMilli(),
		})
	}

	return &trading.MonitorKlineRes{
		Symbol:   symbol,
		Interval: interval,
		List:     list,
	}, nil
}
