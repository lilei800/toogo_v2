// Package engine 机器人引擎模块 - 核心引擎
package engine

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"hotgo/internal/library/config"
	"hotgo/internal/library/exchange"
	"hotgo/internal/library/market"
	"hotgo/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// NewRobotEngine 创建机器人引擎
func NewRobotEngine(ctx context.Context, robot *entity.TradingRobot, apiConfig *entity.TradingApiConfig, ex exchange.Exchange) *RobotEngine {
	engine := &RobotEngine{
		Robot:            robot,
		APIConfig:        apiConfig,
		Platform:         strings.ToLower(strings.TrimSpace(apiConfig.Platform)),
		Exchange:         ex,
		PositionTrackers: make(map[string]*PositionTracker),
		StopCh:           make(chan struct{}),
		MarketRiskMapping: map[string]string{
			"trend":    "aggressive",
			"range": "balanced",
			"high_vol": "aggressive",
			"low_vol":  "conservative",
		},
	}

	// 初始化价格窗口
	engine.PriceWin = NewPriceWindow(strings.ToUpper(strings.TrimSpace(robot.Symbol)), 60, 10)

	// 从机器人配置加载风险映射
	engine.loadRiskConfigFromRobot(ctx)

	// 初始化各模块 (已移除 RiskManager，直接使用策略模板参数)
	engine.Analyzer = NewRobotAnalyzer(engine)
	engine.SignalGen = NewRobotSignalGen(engine)
	engine.Trader = NewRobotTrader(engine)

	return engine
}

// loadRiskConfigFromRobot 从机器人配置加载风险映射
func (e *RobotEngine) loadRiskConfigFromRobot(ctx context.Context) {
	if e.Robot.CurrentStrategy == "" {
		return
	}

	var configData map[string]interface{}
	if err := json.Unmarshal([]byte(e.Robot.CurrentStrategy), &configData); err != nil {
		return
	}

	if riskConfig, ok := configData["riskConfig"].(map[string]interface{}); ok {
		if mapping, ok := riskConfig["marketRiskMapping"].(map[string]interface{}); ok {
			for k, v := range mapping {
				if vs, ok := v.(string); ok {
					// 兼容旧版本：将 'volatile' 自动转换为 'range'
					if k == "volatile" {
						e.MarketRiskMapping["range"] = vs
						g.Log().Debugf(ctx, "[Engine] robotId=%d 配置迁移: volatile → range", e.Robot.Id)
					} else {
						e.MarketRiskMapping[k] = vs
					}
				}
			}
			g.Log().Infof(ctx, "[Engine] robotId=%d 加载风险配置映射: %v", e.Robot.Id, e.MarketRiskMapping)
		}
	}
}

// Start 启动引擎
func (e *RobotEngine) Start(ctx context.Context) error {
	e.Mu.Lock()
	if e.Running {
		e.Mu.Unlock()
		return nil
	}
	e.Running = true
	e.Mu.Unlock()

	g.Log().Infof(ctx, "[Engine] 机器人引擎启动: robotId=%d, symbol=%s", e.Robot.Id, e.Robot.Symbol)

	// 订阅行情
	market.GetMarketServiceManager().Subscribe(ctx, e.Platform, e.Robot.Symbol, e.Exchange)

	// 启动主循环（合并多个定时任务）
	go e.runMainLoop(ctx)

	return nil
}

// Stop 停止引擎
func (e *RobotEngine) Stop() {
	e.Mu.Lock()
	defer e.Mu.Unlock()

	if !e.Running {
		return
	}

	e.Running = false
	close(e.StopCh)

	// 取消订阅行情
	market.GetMarketServiceManager().Unsubscribe(e.Platform, e.Robot.Symbol)

	g.Log().Infof(context.Background(), "[Engine] 机器人引擎停止: robotId=%d", e.Robot.Id)
}

// IsRunning 检查是否运行中
func (e *RobotEngine) IsRunning() bool {
	e.Mu.RLock()
	defer e.Mu.RUnlock()
	return e.Running
}

// UpdateRobot 更新机器人配置
func (e *RobotEngine) UpdateRobot(robot *entity.TradingRobot) {
	e.Mu.Lock()
	defer e.Mu.Unlock()
	e.Robot = robot
}

// runMainLoop 主循环（优化：合并多个定时器）
func (e *RobotEngine) runMainLoop(ctx context.Context) {
	fastTicker := time.NewTicker(TradingInterval) // 500ms
	slowTicker := time.NewTicker(SyncInterval)    // 5s - 用于同步账户和持仓数据
	defer fastTicker.Stop()
	defer slowTicker.Stop()

	var cycleCount int

	// 立即执行一次账户数据同步
	e.syncAccountData(ctx)

	for {
		select {
		case <-e.StopCh:
			return
		case <-fastTicker.C:
			cycleCount++

			// 每500ms: 交易检查
			e.doTradingCheck(ctx)

			// 每1s (每2次): 市场分析和信号生成
			if cycleCount%2 == 0 {
				e.doAnalysis(ctx)
				e.doSignalGeneration(ctx)
			}

		case <-slowTicker.C:
			// 每5s: 同步账户余额和持仓数据
			e.syncAccountData(ctx)
		}
	}
}

// doAnalysis 执行市场分析
func (e *RobotEngine) doAnalysis(ctx context.Context) {
	// 获取行情数据
	ticker := market.GetMarketServiceManager().GetTicker(e.Platform, e.Robot.Symbol)
	if ticker != nil {
		e.Mu.Lock()
		e.LastTicker = ticker
		e.LastTickerUpdate = time.Now()
		e.Mu.Unlock()

		// 添加价格点到窗口
		if ticker.LastPrice > 0 {
			e.PriceWin.AddPoint(ticker.LastPrice)
		}
	}

	// 获取K线数据
	klines := market.GetMarketServiceManager().GetMultiTimeframeKlines(e.Platform, e.Robot.Symbol)
	if klines != nil {
		e.Mu.Lock()
		e.LastKlines = klines
		e.Mu.Unlock()
	}

	// 加载波动率配置
	e.loadVolatilityConfig(ctx)

	// 执行分析
	if e.LastKlines != nil {
		analysis := e.Analyzer.Analyze(ctx)
		if analysis != nil {
			e.Mu.Lock()
			e.LastAnalysis = analysis
			e.LastAnalysisUpdate = time.Now()
			e.Mu.Unlock()

			// 检测市场状态变化
			e.checkAndUpdateStrategyConfig(ctx, analysis.MarketState)
		}
	}
}

// syncAccountData 同步账户数据（余额和持仓）
func (e *RobotEngine) syncAccountData(ctx context.Context) {
	// 获取账户余额
	balance, err := e.Exchange.GetBalance(ctx)
	if err == nil {
		e.Mu.Lock()
		e.AccountBalance = balance
		e.LastBalanceUpdate = time.Now()
		e.Mu.Unlock()
	} else {
		g.Log().Debugf(ctx, "[Engine] robotId=%d 获取余额失败: %v", e.Robot.Id, err)
	}

	// 获取持仓
	positions, err := e.Exchange.GetPositions(ctx, e.Robot.Symbol)
	if err == nil {
		e.Mu.Lock()
		e.CurrentPositions = positions
		e.LastPositionUpdate = time.Now()
		e.Mu.Unlock()
	}
}

// doSignalGeneration 生成方向信号
func (e *RobotEngine) doSignalGeneration(ctx context.Context) {
	signal := e.SignalGen.Generate(ctx)
	if signal != nil {
		e.Mu.Lock()
		e.LastSignal = signal
		e.LastSignalUpdate = time.Now()
		e.Mu.Unlock()
	}
}

// doTradingCheck 交易检查
func (e *RobotEngine) doTradingCheck(ctx context.Context) {
	// 【已删除】自动平仓检查已删除
	// go e.checkClosePosition(ctx)

	// 检查开仓
	go e.checkOpenPosition(ctx)
}

// checkOpenPosition 检查是否应该开仓
func (e *RobotEngine) checkOpenPosition(ctx context.Context) {
	if !e.OrderLock.TryLock() {
		return
	}
	defer e.OrderLock.Unlock()

	e.Trader.CheckAndOpenPosition(ctx)
}

// checkClosePosition 检查是否应该平仓
// 【已删除】自动平仓功能已删除，此函数不再执行任何操作
func (e *RobotEngine) checkClosePosition(ctx context.Context) {
	// 自动平仓功能已删除
	return
}

// checkAndUpdateStrategyConfig 检测市场状态变化并更新策略配置
func (e *RobotEngine) checkAndUpdateStrategyConfig(ctx context.Context, currentMarketState string) {
	if currentMarketState == "" || currentMarketState == e.LastMarketState {
		return
	}

	oldState := e.LastMarketState
	e.LastMarketState = currentMarketState

	// 根据市场状态获取对应的风险偏好
	riskPreference := e.MarketRiskMapping[currentMarketState]
	if riskPreference == "" {
		riskPreference = "balanced"
	}

	// 加载策略参数
	strategyParams := LoadStrategyParams(ctx, e.Robot, currentMarketState, riskPreference)

	e.Mu.Lock()
	e.CurrentStrategyParams = strategyParams
	e.Mu.Unlock()

	// 更新价格窗口配置
	if strategyParams.Window > 0 && strategyParams.Threshold > 0 {
		e.PriceWin.UpdateConfig(strategyParams.Window, strategyParams.Threshold)
		g.Log().Infof(ctx, "[Engine] robotId=%d 市场状态变化: %s → %s, 风险偏好: %s",
			e.Robot.Id, oldState, currentMarketState, riskPreference)
	}
}

// loadVolatilityConfig 从全局配置管理器加载波动率配置
func (e *RobotEngine) loadVolatilityConfig(ctx context.Context) {
	// 从全局配置管理器获取配置（只读，无数据库查询，无内存复制）
	globalConfig := config.GetVolatilityConfigManager().GetConfig(e.Robot.Symbol)
	
	if globalConfig == nil {
		// 极端情况：配置管理器未初始化，使用默认配置
		e.Mu.Lock()
		if e.VolatilityConfig == nil {
			e.VolatilityConfig = &VolatilityConfig{
				HighVolatilityThreshold: HighVolatilityThreshold,
				LowVolatilityThreshold:  LowVolatilityThreshold,
				TrendStrengthThreshold:  TrendStrengthThreshold,
				Symbol:                  "默认",
			}
		}
		e.Mu.Unlock()
		return
	}
	
	// 转换为引擎使用的配置格式（只在配置变化时才更新）
	e.Mu.Lock()
	needUpdate := e.VolatilityConfig == nil || 
		e.VolatilityConfig.HighVolatilityThreshold != globalConfig.HighVolatilityThreshold ||
		e.VolatilityConfig.LowVolatilityThreshold != globalConfig.LowVolatilityThreshold ||
		e.VolatilityConfig.TrendStrengthThreshold != globalConfig.TrendStrengthThreshold
	
	if needUpdate {
		e.VolatilityConfig = &VolatilityConfig{
			HighVolatilityThreshold: globalConfig.HighVolatilityThreshold,
			LowVolatilityThreshold:  globalConfig.LowVolatilityThreshold,
			TrendStrengthThreshold:  globalConfig.TrendStrengthThreshold,
			Weight1m:                globalConfig.Weight1m,
			Weight5m:                globalConfig.Weight5m,
			Weight15m:               globalConfig.Weight15m,
			Weight30m:               globalConfig.Weight30m,
			Weight1h:                globalConfig.Weight1h,
			Symbol:                  globalConfig.Symbol,
		}
		g.Log().Debugf(ctx, "[Engine] robotId=%d 配置已更新: symbol=%s, high=%.2f, low=%.2f, trend=%.2f",
			e.Robot.Id, globalConfig.Symbol, globalConfig.HighVolatilityThreshold, 
			globalConfig.LowVolatilityThreshold, globalConfig.TrendStrengthThreshold)
	}
	e.Mu.Unlock()
}

// GetStatus 获取引擎状态
func (e *RobotEngine) GetStatus() *EngineStatus {
	e.Mu.RLock()
	defer e.Mu.RUnlock()

	status := &EngineStatus{
		RobotId:   e.Robot.Id,
		Symbol:    e.Robot.Symbol,
		Platform:  e.Platform,
		Running:   e.Running,
		Connected: e.LastTicker != nil && time.Since(e.LastTickerUpdate) < 10*time.Second,
	}

	if e.LastTicker != nil {
		status.LastPrice = e.LastTicker.LastPrice
	}

	if e.AccountBalance != nil {
		status.TotalBalance = e.AccountBalance.TotalBalance
		status.AvailBalance = e.AccountBalance.AvailableBalance
	}

	if e.LastAnalysis != nil {
		status.MarketState = e.LastAnalysis.MarketState
		status.TrendDirection = e.LastAnalysis.TrendDirection
		status.Volatility = e.LastAnalysis.Volatility
	}

	if e.LastSignal != nil {
		status.SignalDirection = e.LastSignal.Direction
		status.SignalStrength = e.LastSignal.Strength
		status.SignalConfidence = e.LastSignal.Confidence
		status.SignalProgress = e.LastSignal.SignalProgress
		status.SignalReason = e.LastSignal.Reason
	}

	// 持仓信息
	for _, pos := range e.CurrentPositions {
		if pos.PositionAmt != 0 {
			status.HasPosition = true
			status.PositionSide = pos.PositionSide
			status.PositionAmt = pos.PositionAmt
			status.EntryPrice = pos.EntryPrice
			status.UnrealizedPnl = pos.UnrealizedPnl
			break
		}
	}

	// 策略配置
	if e.PriceWin != nil && e.PriceWin.Config != nil {
		status.StrategyWindow = e.PriceWin.Config.Window
		status.StrategyThreshold = e.PriceWin.Config.Threshold
	}

	status.CurrentMarketState = e.LastMarketState
	if e.LastMarketState != "" && e.MarketRiskMapping != nil {
		status.CurrentRiskPref = e.MarketRiskMapping[e.LastMarketState]
	}

	// 价格窗口数据
	stats := e.PriceWin.GetStats()
	status.WindowMinPrice = stats.MinPrice
	status.WindowMaxPrice = stats.MaxPrice
	status.WindowCurrentPrice = stats.CurrentPrice

	if e.PriceWin.Config != nil {
		status.LongTriggerPrice = stats.MinPrice + e.PriceWin.Config.Threshold
		status.ShortTriggerPrice = stats.MaxPrice - e.PriceWin.Config.Threshold
	}

	// 转换价格窗口数据
	points := e.PriceWin.GetPoints()
	status.PriceWindowData = make([]PriceWindowPoint, len(points))
	for i, p := range points {
		status.PriceWindowData[i] = PriceWindowPoint{
			Timestamp: p.Timestamp,
			Price:     p.Price,
		}
	}

	return status
}

// HasActivePosition 检查是否有活跃持仓
func (e *RobotEngine) HasActivePosition(side string) bool {
	e.Mu.RLock()
	defer e.Mu.RUnlock()

	for _, pos := range e.CurrentPositions {
		if pos.PositionAmt != 0 && pos.PositionSide == side {
			return true
		}
	}
	return false
}

// GetPosition 获取指定方向持仓
func (e *RobotEngine) GetPosition(side string) *exchange.Position {
	e.Mu.RLock()
	defer e.Mu.RUnlock()

	for _, pos := range e.CurrentPositions {
		if pos.PositionSide == side && pos.PositionAmt != 0 {
			return pos
		}
	}
	return nil
}

