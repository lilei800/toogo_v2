// Package toogo 机器人核心引擎
// 每个机器人独立的引擎，包含市场分析、风险评估、方向信号、自动交易等功能
package toogo

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"hotgo/internal/dao"
	"hotgo/internal/library/exchange"
	"hotgo/internal/library/market"
	"hotgo/internal/model/entity"
	"hotgo/internal/service"
	"hotgo/internal/websocket"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/grand"
)

// RobotEngine 机器人核心引擎
// 每个机器人一个独立的引擎实例
type RobotEngine struct {
	mu sync.RWMutex

	// 基础配置
	Robot     *entity.TradingRobot
	APIConfig *entity.TradingApiConfig
	Platform  string
	Exchange  exchange.Exchange

	// ============ 模块 (已移除 RiskManager，直接使用策略模板参数) ============
	Analyzer  *RobotAnalyzer  // 市场分析模块
	SignalGen *RobotSignalGen // 信号生成模块
	Trader    *RobotTrader    // 交易执行模块

	// ============ 状态缓存（统一数据中心，所有请求都使用这里的缓存） ============
	LastTicker       *exchange.Ticker     // 最新行情
	LastKlines       *market.KlineCache   // 最新K线
	LastAnalysis     *RobotMarketAnalysis // 最新分析结果
	LastSignal       *RobotSignal         // 最新方向信号
	CurrentPositions []*exchange.Position // 当前持仓
	AccountBalance   *exchange.Balance    // 账户余额
	OrderHistory     []*exchange.Order    // 订单历史缓存

	// ============ 时间记录 ============
	LastTickerUpdate           time.Time
	LastAnalysisUpdate         time.Time
	LastSignalUpdate           time.Time
	LastProcessedSignalTime    time.Time // 【新增】上次已处理的信号时间戳（用于防止重复下单）
	LastPositionUpdate         time.Time
	LastBalanceUpdate          time.Time
	LastOrderHistoryUpdate     time.Time // 订单历史更新时间
	LastOrderSync              time.Time // 上次订单状态同步时间（用于控制同步频率）
	LastSyncError              error     // 上次同步错误
	LastVolatilityConfigUpdate time.Time // 波动率配置更新时间（减少数据库查询）
	LastStrategyParamsUpdate   time.Time // 策略参数更新时间（减少数据库查询）

	SyncErrorCount int // 连续同步错误次数

	// ============ 持仓跟踪 ============
	PositionTrackers map[string]*PositionTracker

	// ============ 窗口价格监控（toogo实时信号逻辑） ============
	PriceWindow      []PricePoint        // 窗口内价格序列
	SignalHistory    []SignalHistoryItem // 信号历史
	MonitorConfig    *MonitorConfig      // 监控配置
	LastAlertedLong  *float64            // 上次做多预警的基准价
	LastAlertedShort *float64            // 上次做空预警的基准价
	LastWindowMin    *float64            // 上次窗口最低价
	LastWindowMax    *float64            // 上次窗口最高价
	LastWindowSignal string              // 上次窗口信号方向

	// ============ 市场状态与策略配置 ============
	LastMarketState       string            // 上次市场状态（用于检测市场状态变化，避免重复加载策略）
	MarketRiskMapping     map[string]string // 市场状态 → 风险偏好映射（从机器人 remark 字段加载，不再从 CurrentStrategy JSON 加载）
	CurrentStrategyParams *StrategyParams   // 当前使用的策略参数（根据市场状态从模板加载）
	VolatilityConfig      *VolatilityConfig // 波动率配置（简化版：市场状态阈值 + 时间周期权重）
	LastSetLeverage       int               // 上次设置的杠杆（避免重复调用API）

	// ============ 运行状态 ============
	running bool
	stopCh  chan struct{}

	// ============ 锁 ============
	orderLock sync.Mutex
	priceLock sync.RWMutex // 价格窗口数据锁

	// ============ 并发控制 ============
	processingPriceUpdate int32 // 是否正在处理价格更新（原子操作，防止goroutine堆积）

	// ============ API 请求去重（singleflight模式） ============
	positionFetching int32 // 是否正在获取持仓（原子操作，防止重复请求）
	balanceFetching  int32 // 是否正在获取余额（原子操作，防止重复请求）
}

// GetAccountSnapshot 获取账户缓存快照（线程安全）
// 用于跨包读取引擎内的余额/持仓缓存（避免直接访问未导出的 mu）。
func (e *RobotEngine) GetAccountSnapshot() (bal *exchange.Balance, positions []*exchange.Position, lastBalAt, lastPosAt time.Time) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	bal = e.AccountBalance
	lastBalAt = e.LastBalanceUpdate
	lastPosAt = e.LastPositionUpdate
	// 返回 slice 的浅拷贝，避免调用方误改内部 slice 头
	if e.CurrentPositions != nil {
		positions = append([]*exchange.Position(nil), e.CurrentPositions...)
	}
	return
}

// RobotMarketAnalysis 机器人市场分析结果
type RobotMarketAnalysis struct {
	Timestamp time.Time

	// 市场状态
	MarketState     string  // trend/volatile/high_vol/low_vol
	MarketStateConf float64 // 市场状态置信度
	TrendDirection  string  // up/down/neutral
	TrendStrength   float64 // 趋势强度 0-100
	Volatility      float64 // 波动率
	VolatilityLevel string  // low/normal/high

	// 多周期分析
	TimeframeScores map[string]*TimeframeScore // 各周期评分

	// 技术指标
	Indicators *TechnicalIndicators
}

// TimeframeScore 单周期评分（精简版）
type TimeframeScore struct {
	Timeframe     string
	Direction     string  // up/down/neutral
	Strength      float64 // 方向强度 0-100
	TrendStrength float64 // 趋势强度 0-1
	Volatility    float64 // 波动率
	MarketState   string  // trend/volatile/high_vol/low_vol
	MACD          float64
	EMA12         float64
	EMA26         float64
	KlinesCount   int
}

// 周期权重配置（精简为3个核心周期）
var timeframeWeights = map[string]float64{
	"5m":  0.20, // 短期
	"15m": 0.35, // 中期（主要）
	"1h":  0.45, // 长期（最重要）
}

// 波动率阈值配置
const (
	highVolatilityThreshold = 2.0  // 高波动阈值
	lowVolatilityThreshold  = 0.5  // 低波动阈值
	trendStrengthThreshold  = 0.35 // 趋势强度阈值
)

// TechnicalIndicators 技术指标汇总（精简版）
type TechnicalIndicators struct {
	TrendScore      float64 // 趋势综合评分 -100 ~ 100
	VolatilityScore float64 // 波动评分 0-100
}

// PositionTracker 持仓跟踪器（纯内存，每个新订单自动重置）
type PositionTracker struct {
	PositionSide      string    // 持仓方向 LONG/SHORT
	EntryMargin       float64   // 开仓保证金
	EntryTime         time.Time // 开仓时间
	HighestProfit     float64   // 最高盈利金额（只增不减）
	LowestProfit      float64   // 最低盈利金额（负数表示亏损）
	TakeProfitEnabled bool      // 止盈回撤是否已启用（由用户手动开启或自动触发）
	OrderId           int64     // 关联的订单ID（用于检测订单变化）

	// ===== 冻结参数（开仓时确定，用于前端血条/展示；优先内存，无则DB兜底） =====
	ParamsLoaded            bool    // 是否已加载冻结参数
	StopLossPercent         float64 // 止损百分比(%)
	AutoStartRetreatPercent float64 // 启动止盈百分比(%)
	ProfitRetreatPercent    float64 // 止盈回撤百分比(%)
	MarginPercent           float64 // 保证金比例(%)
	MarketState             string  // 开仓时市场状态
	RiskPreference          string  // 开仓时风险偏好
}

// RobotSignal 机器人方向信号
type RobotSignal struct {
	Timestamp time.Time

	// 方向 LONG/SHORT/NEUTRAL
	Direction string

	// 信号强度 0-100
	Strength float64

	// 置信度 0-100
	Confidence float64

	// 建议操作 OPEN_LONG/OPEN_SHORT/CLOSE_LONG/CLOSE_SHORT/HOLD
	Action string

	// 多周期对齐数
	AlignedTimeframes int

	// 信号原因
	Reason string

	// ============ 窗口信号相关（toogo实时信号逻辑） ============
	// 窗口内最高价
	WindowMaxPrice float64
	// 窗口内最低价
	WindowMinPrice float64
	// 当前价格
	CurrentPrice float64
	// 距最低价距离
	DistanceFromMin float64
	// 距最高价距离
	DistanceFromMax float64
	// 信号阈值
	SignalThreshold float64
	// 信号进度百分比 0-100
	SignalProgress float64
	// 信号类型: window(窗口信号)/analysis(分析信号)
	SignalType string
}

// PricePoint 价格数据点
type PricePoint struct {
	Timestamp int64   `json:"timestamp"` // 毫秒时间戳
	Price     float64 `json:"price"`     // 价格
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	Symbol    string  // 监控交易对
	Window    int     // 窗口期（秒）
	Threshold float64 // 信号阈值（价格波动幅度）
}

// SignalHistoryItem 信号历史记录
type SignalHistoryItem struct {
	Timestamp int64  `json:"timestamp"` // 毫秒时间戳
	Signal    string `json:"signal"`    // long/short/neutral
}

// NewRobotEngine 创建机器人引擎
func NewRobotEngine(ctx context.Context, robot *entity.TradingRobot, apiConfig *entity.TradingApiConfig, ex exchange.Exchange) *RobotEngine {
	engine := &RobotEngine{
		Robot:            robot,
		APIConfig:        apiConfig,
		Platform:         strings.ToLower(strings.TrimSpace(apiConfig.Platform)),
		Exchange:         ex,
		PositionTrackers: make(map[string]*PositionTracker),
		stopCh:           make(chan struct{}),
		// 初始化窗口监控相关
		PriceWindow:      make([]PricePoint, 0, 1000),
		SignalHistory:    make([]SignalHistoryItem, 0, 100),
		LastWindowSignal: "neutral",
	}

	// 初始化监控配置（使用默认值，可通过API更新）
	engine.MonitorConfig = &MonitorConfig{
		Symbol:    strings.ToUpper(strings.TrimSpace(robot.Symbol)),
		Window:    60, // 默认60秒窗口
		Threshold: 10, // 默认10 USDT阈值
	}

	// 初始化映射关系（空映射，将从机器人配置加载）
	// 【重要】不使用默认值，必须从机器人创建时保存的 remark 字段中加载
	engine.MarketRiskMapping = make(map[string]string)

	// 从机器人配置加载风险映射（创建时保存的映射关系，必选）
	// 验证逻辑已在 loadRiskConfigFromRobot 中完成，如果失败会记录错误日志
	engine.loadRiskConfigFromRobot(ctx)

	// 初始化各模块 (已移除 RiskManager，直接使用策略模板参数)
	engine.Analyzer = NewRobotAnalyzer(engine)
	engine.SignalGen = NewRobotSignalGen(engine)
	engine.Trader = NewRobotTrader(engine)

	return engine
}

// loadRiskConfigFromRobot 从机器人配置加载风险映射
// 【重要】映射关系存储在 Robot.Remark 字段中（JSON格式），这是创建机器人时保存的独立映射关系
// 每个机器人都有自己独立的映射关系，必须使用创建时保存的映射关系
// 【严格模式】创建机器人时映射关系是必选的，如果无法加载，直接报错，不使用任何备用方案
func (e *RobotEngine) loadRiskConfigFromRobot(ctx context.Context) {
	// 【必须】从 remark 字段解析映射关系（创建时保存的独立映射关系）
	if e.Robot.Remark == "" {
		g.Log().Errorf(ctx, "[RobotEngine] robotId=%d ❌ remark 字段为空，无法加载映射关系。创建机器人时映射关系是必选的，请检查数据完整性", e.Robot.Id)
		e.mu.Lock()
		e.MarketRiskMapping = make(map[string]string) // 保持为空
		e.mu.Unlock()
		return
	}

	// Parse mapping from remark (compatible with 2 formats)
	// 1) legacy: remark is map[string]string
	// 2) new: remark is RiskConfig JSON containing field "marketRiskMapping"
	var mapping map[string]string
	if err := json.Unmarshal([]byte(e.Robot.Remark), &mapping); err != nil || mapping == nil || len(mapping) == 0 {
		var wrapper struct {
			MarketRiskMapping map[string]string `json:"marketRiskMapping"`
		}
		if err2 := json.Unmarshal([]byte(e.Robot.Remark), &wrapper); err2 != nil || wrapper.MarketRiskMapping == nil || len(wrapper.MarketRiskMapping) == 0 {
			g.Log().Errorf(ctx, "[RobotEngine] robotId=%d ??remark parse failed (mapping/RiskConfig): %s, err=%v/%v", e.Robot.Id, e.Robot.Remark, err, err2)
			e.mu.Lock()
			e.MarketRiskMapping = make(map[string]string) // keep empty
			e.mu.Unlock()
			return
		}
		mapping = wrapper.MarketRiskMapping
	}

	// 成功解析映射关系JSON
	e.mu.Lock()
	e.MarketRiskMapping = make(map[string]string)
	for k, v := range mapping {
		// 规范化市场状态键
		normalizedKey := normalizeMarketState(k)
		e.MarketRiskMapping[normalizedKey] = v
		if normalizedKey != k {
			g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 映射关系键规范化: %s → %s", e.Robot.Id, k, normalizedKey)
		}
	}
	e.mu.Unlock()

	// 验证映射关系完整性（必须包含所有4种市场状态）
	requiredStates := []string{"trend", "volatile", "high_vol", "low_vol"}
	missingStates := []string{}
	e.mu.RLock()
	for _, state := range requiredStates {
		if _, exists := e.MarketRiskMapping[state]; !exists {
			missingStates = append(missingStates, state)
		}
	}
	e.mu.RUnlock()

	if len(missingStates) > 0 {
		g.Log().Errorf(ctx, "[RobotEngine] robotId=%d ❌ 映射关系不完整，缺少以下市场状态: %v，当前映射关系: %v。创建机器人时映射关系是必选的，请检查数据完整性", e.Robot.Id, missingStates, e.MarketRiskMapping)
		// 清空映射关系，让后续逻辑报错
		e.mu.Lock()
		e.MarketRiskMapping = make(map[string]string)
		e.mu.Unlock()
		return
	}

	g.Log().Infof(ctx, "[RobotEngine] robotId=%d ✅ 成功从 remark 字段加载创建时保存的映射关系: %v", e.Robot.Id, e.MarketRiskMapping)
}

// Start 启动引擎
func (e *RobotEngine) Start(ctx context.Context) error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return nil
	}
	e.running = true
	e.mu.Unlock()

	g.Log().Infof(ctx, "[RobotEngine] 机器人引擎启动: robotId=%d, symbol=%s", e.Robot.Id, e.Robot.Symbol)

	// 【优化】订阅行情并注册价格更新回调（用于实时平仓检查）
	market.GetMarketServiceManager().SubscribeWithCallback(ctx, e.Platform, e.Robot.Symbol, e.Exchange, func(ticker *exchange.Ticker) {
		// WebSocket价格推送回调 - 实时触发平仓检查
		e.OnPriceUpdate(ctx, ticker)
	})

	// 私有WS（订单/持仓/账户变更）按 apiConfigId 复用：事件驱动触发同步，减少轮询
	// 失败不阻断引擎启动（仍有定期兜底对账）
	if e.APIConfig != nil {
		if err := GetPrivateStreamManager().Acquire(ctx, e.APIConfig, e.Robot.Symbol, e.Robot.Id); err != nil {
			g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 私有WS启动失败(忽略): %v", e.Robot.Id, err)
		}
	}

	// 【优化】等待行情服务获取初始数据（首次订阅会自动获取）
	// MarketServiceManager.Subscribe 内部会调用 fetchInitialData 获取历史K线
	// 这里等待一段时间让初始数据加载完成（最多等待3秒，每500ms检查一次）
	maxWait := 3 * time.Second
	checkInterval := 500 * time.Millisecond
	startTime := time.Now()

	for time.Since(startTime) < maxWait {
		klineCache := market.GetMarketServiceManager().GetMultiTimeframeKlines(e.Platform, e.Robot.Symbol)
		if klineCache != nil && len(klineCache.Klines1m) > 0 {
			g.Log().Infof(ctx, "[RobotEngine] 已获取历史K线数据: robotId=%d, platform=%s, symbol=%s, 1m=%d条, 耗时=%v",
				e.Robot.Id, e.Platform, e.Robot.Symbol, len(klineCache.Klines1m), time.Since(startTime))
			break
		}
		time.Sleep(checkInterval)
	}

	// 最终检查
	klineCache := market.GetMarketServiceManager().GetMultiTimeframeKlines(e.Platform, e.Robot.Symbol)
	if klineCache == nil || len(klineCache.Klines1m) == 0 {
		g.Log().Warningf(ctx, "[RobotEngine] 历史K线数据获取超时: robotId=%d, platform=%s, symbol=%s, 等待时间=%v",
			e.Robot.Id, e.Platform, e.Robot.Symbol, time.Since(startTime))
	}

	// 启动统一主循环（优化：4个循环合并为1个，减少goroutine开销）
	go e.runMainLoop(ctx)

	return nil
}

// Stop 停止引擎
func (e *RobotEngine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return
	}

	e.running = false
	close(e.stopCh)

	// 取消订阅行情
	market.GetMarketServiceManager().Unsubscribe(e.Platform, e.Robot.Symbol)

	// 释放私有WS引用
	if e.APIConfig != nil {
		GetPrivateStreamManager().Release(e.Platform, e.APIConfig.Id, e.Robot.Symbol, e.Robot.Id)
	}

	g.Log().Infof(context.Background(), "[RobotEngine] 机器人引擎停止: robotId=%d", e.Robot.Id)
}

// IsRunning 检查是否运行中
func (e *RobotEngine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

// UpdateRobot 更新机器人配置
func (e *RobotEngine) UpdateRobot(robot *entity.TradingRobot) {
	e.mu.Lock()
	// 检查CurrentStrategy是否发生变化（用于判断是否需要重新加载策略参数）
	oldCurrentStrategy := e.Robot.CurrentStrategy
	e.Robot = robot
	newCurrentStrategy := robot.CurrentStrategy
	e.mu.Unlock()

	// 重新加载风险配置映射（如果CurrentStrategy发生变化）
	ctx := context.Background()
	e.loadRiskConfigFromRobot(ctx)

	// 如果CurrentStrategy发生变化，触发策略参数重新加载（实时生效）
	if oldCurrentStrategy != newCurrentStrategy {
		g.Log().Infof(ctx, "[RobotEngine] robotId=%d CurrentStrategy已更新，触发策略参数重新加载", robot.Id)

		// 【优化】从全局市场分析器获取市场状态，触发策略参数重新加载
		globalAnalysis := market.GetMarketAnalyzer().GetAnalysis(e.Platform, e.Robot.Symbol)
		if globalAnalysis != nil {
			marketState := normalizeMarketState(string(globalAnalysis.MarketState))
			if marketState != "" {
				// 触发策略参数重新加载
				e.checkAndUpdateStrategyConfig(ctx, marketState)
			}
		}
	}

	g.Log().Infof(context.Background(), "[RobotEngine] 机器人配置已更新: robotId=%d, autoTradeEnabled=%d, autoCloseEnabled=%d",
		robot.Id, robot.AutoTradeEnabled, robot.AutoCloseEnabled)
}

// GetPositionTracker 获取持仓跟踪器（供外部查询使用）
func (e *RobotEngine) GetPositionTracker(positionSide string) *PositionTracker {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.PositionTrackers[positionSide]
}

// ClearPositionTracker 清除持仓跟踪器（手动平仓后调用）
func (e *RobotEngine) ClearPositionTracker(positionSide string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.PositionTrackers, positionSide)
}

// SetTakeProfitEnabled 设置止盈回撤启用状态（供前端开关调用）
// 【内存操作】不再依赖数据库
// 【重要修复】手动启动时需要正确初始化 HighestProfit，否则止盈检查条件永远不满足
func (e *RobotEngine) SetTakeProfitEnabled(positionSide string, enabled bool) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	tracker := e.PositionTrackers[positionSide]
	if tracker == nil {
		return false
	}

	wasEnabled := tracker.TakeProfitEnabled

	// 不允许关闭已开启的止盈回撤（只能开启，不能关闭）
	if tracker.TakeProfitEnabled && !enabled {
		return false
	}

	tracker.TakeProfitEnabled = enabled
	if enabled {
		// 【关键修复】从“未启用 -> 启用”时，必须把最高盈利基准重置为当前时刻的起点，
		// 否则如果历史 HighestProfit 很大、当前盈利已经回撤，会导致“刚启用就立刻触发止盈”。
		if !wasEnabled {
			tracker.HighestProfit = 0 // 让 checkTakeProfitAndClose 在下一次tick用当前盈利初始化为100%血条
		} else if tracker.HighestProfit <= 0 {
			// 兜底：极小正值，避免分母为0
			tracker.HighestProfit = 0.0001
		}
		g.Log().Infof(context.Background(), "[RobotEngine] robotId=%d 手动启动止盈回撤: positionSide=%s, highestProfit=%.4f",
			e.Robot.Id, positionSide, tracker.HighestProfit)
	}
	return true
}

// MarkProfitRetreatStarted 将“止盈回撤已启动”持久化到数据库（用于服务重启后恢复状态）
// - 只允许从 0 -> 1（符合“不可关闭原则”）
// - highestProfit 若可获得则一并写入，避免回撤计算分母为 0
func (e *RobotEngine) MarkProfitRetreatStarted(ctx context.Context, positionSide string) {
	// 尽量使用内存里的最高盈利同步写入
	highestProfit := 0.0
	if tracker := e.GetPositionTracker(positionSide); tracker != nil {
		highestProfit = tracker.HighestProfit
	}
	e.persistProfitRetreatStarted(ctx, positionSide, highestProfit)
}

func (e *RobotEngine) persistProfitRetreatStarted(ctx context.Context, positionSide string, highestProfit float64) {
	if ctx == nil {
		ctx = context.Background()
	}
	e.mu.RLock()
	robot := e.Robot
	e.mu.RUnlock()
	if robot == nil {
		return
	}

	direction := "long"
	if strings.ToUpper(strings.TrimSpace(positionSide)) == "SHORT" {
		direction = "short"
	}

	update := g.Map{
		"profit_retreat_started": 1,
		"updated_at":             gtime.Now(),
	}
	if highestProfit > 0 {
		update["highest_profit"] = highestProfit
	}

	_, err := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", robot.Id).
		Where("direction", direction).
		Where("status", OrderStatusOpen).
		Update(update)
	if err != nil {
		g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 持久化止盈回撤启动状态失败: positionSide=%s, err=%v",
			robot.Id, positionSide, err)
	}
}

// initTrackerFromDB 在首次创建 PositionTracker 时，从数据库恢复“止盈回撤启动/最高盈利”等关键状态
// 仅在 tracker 新建时调用，避免高频查询。
func (e *RobotEngine) initTrackerFromDB(ctx context.Context, positionSide string, tracker *PositionTracker) {
	if tracker == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	e.mu.RLock()
	robot := e.Robot
	e.mu.RUnlock()
	if robot == nil {
		return
	}

	direction := "long"
	if strings.ToUpper(strings.TrimSpace(positionSide)) == "SHORT" {
		direction = "short"
	}

	var order *entity.TradingOrder
	err := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", robot.Id).
		Where("direction", direction).
		Where("status", OrderStatusOpen).
		Fields("id", "profit_retreat_started", "highest_profit", "margin").
		OrderDesc("id").
		Scan(&order)
	if err != nil || order == nil {
		return
	}

	if order.ProfitRetreatStarted == 1 {
		tracker.TakeProfitEnabled = true
	}
	if order.HighestProfit > tracker.HighestProfit {
		tracker.HighestProfit = order.HighestProfit
	}
	if tracker.EntryMargin <= 0 && order.Margin > 0 {
		tracker.EntryMargin = order.Margin
	}
}

// GetOrCreatePositionTracker 获取或创建持仓跟踪器
// 【内存操作】确保跟踪器存在
func (e *RobotEngine) GetOrCreatePositionTracker(positionSide string, margin float64) *PositionTracker {
	e.mu.Lock()
	defer e.mu.Unlock()

	tracker := e.PositionTrackers[positionSide]
	if tracker == nil {
		tracker = &PositionTracker{
			PositionSide:      positionSide,
			EntryMargin:       margin,
			EntryTime:         time.Now(),
			HighestProfit:     0,
			TakeProfitEnabled: false,
		}
		e.PositionTrackers[positionSide] = tracker
	}
	return tracker
}

// ClearPosition 清除指定方向的持仓（从内存中删除）
// 【注意】此方法会清除内存中的持仓数据，但不会影响交易所实际持仓
// 如果交易所仍有持仓，系统会在下次同步时重新加载
func (e *RobotEngine) ClearPosition(ctx context.Context, positionSide string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 清除持仓跟踪器
	delete(e.PositionTrackers, positionSide)

	// 直接从 CurrentPositions 移除该方向持仓，避免后续接口仍返回“0数量残留对象”
	if e.CurrentPositions != nil && len(e.CurrentPositions) > 0 {
		newList := make([]*exchange.Position, 0, len(e.CurrentPositions))
		for _, p := range e.CurrentPositions {
			if p == nil {
				continue
			}
			if strings.ToUpper(strings.TrimSpace(p.PositionSide)) == strings.ToUpper(strings.TrimSpace(positionSide)) {
				continue
			}
			newList = append(newList, p)
		}
		e.CurrentPositions = newList
		e.LastPositionUpdate = time.Now()
		g.Log().Infof(ctx, "[RobotEngine] robotId=%d 已从内存移除持仓: positionSide=%s", e.Robot.Id, positionSide)
	}
}

// ClearAllPositions 清除所有持仓（从内存中删除）
// 【警告】此方法会清除所有内存中的持仓数据，但不会影响交易所实际持仓
// 如果交易所仍有持仓，系统会在下次同步时重新加载
// 【使用场景】用于重置机器人状态、修复数据不一致等问题
func (e *RobotEngine) ClearAllPositions(ctx context.Context) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 清除所有持仓跟踪器
	e.PositionTrackers = make(map[string]*PositionTracker)

	// 清空 CurrentPositions
	e.CurrentPositions = make([]*exchange.Position, 0)

	g.Log().Infof(ctx, "[RobotEngine] robotId=%d 已清除所有内存中的持仓数据", e.Robot.Id)
}

// ==================== 统一主循环（优化：4合1） ====================

// runMainLoop 统一主循环
// 优化：将4个独立循环合并为1个，减少goroutine数量75%
// 调度策略：
//   - 每500ms: 交易检查（高频，最重要）
//   - 每1秒(500ms*2): 市场分析 + 信号生成
//   - 每3秒(500ms*6): 风险评估
func (e *RobotEngine) runMainLoop(ctx context.Context) {
	// 【健壮性优化】添加 panic 恢复机制，确保单个引擎异常不影响其他引擎
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "[RobotEngine] panic recovered: robotId=%d, err=%v, stack=%s",
				e.Robot.Id, r, getStackTrace())
			// 尝试优雅停止引擎
			e.mu.Lock()
			e.running = false
			e.mu.Unlock()
		}
	}()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var tickCount int64 = 0

	// 立即执行一次初始化
	e.doAnalysis(ctx)
	e.syncAccountDataIfNeeded(ctx, "init") // 初始化时强制同步
	e.doSignalGeneration(ctx)

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			tickCount++

			// 【纯事件驱动架构】每1秒执行市场分析和信号生成
			// - doAnalysis: 价格更新时立即触发平仓检查（事件驱动）
			// - doSignalGeneration: 信号生成时立即触发开仓检查（事件驱动）
			// 不再使用定时兜底检查，完全依赖事件驱动，提高响应速度和效率
			if tickCount%2 == 0 {
				e.doAnalysis(ctx)
				e.doSignalGeneration(ctx)
			}

			// 【事件驱动优化】每2分钟(tickCount % 240 == 0): 兜底同步
			// 主要同步在事件驱动中完成（开仓前、平仓后）
			// 此处仅作为最终一致性保障，频率大幅降低
			if tickCount%240 == 0 {
				e.syncAccountDataIfNeeded(ctx, "periodic")
			}

			// 防止溢出，每10分钟重置计数器
			if tickCount >= 1200 {
				tickCount = 0
			}
		}
	}
}

// getStackTrace 获取堆栈跟踪（高效实现，避免使用debug.Stack）
func getStackTrace() string {
	// 简化实现，只返回基本信息
	return "see logs for details"
}

// ==================== 核心逻辑 ====================

// OnPriceUpdate 价格更新回调（WebSocket推送触发）
// 【核心优化】实时触发止损止盈检查，响应速度从2秒降低到毫秒级
func (e *RobotEngine) OnPriceUpdate(ctx context.Context, ticker *exchange.Ticker) {
	if ticker == nil {
		return
	}

	// pricePoint：用于价格窗口/信号（优先成交价 LastPrice；缺失则用 MarkPrice 兜底）
	pricePoint := ticker.LastPrice
	if pricePoint <= 0 {
		pricePoint = ticker.EffectiveMarkPrice()
	}
	if pricePoint <= 0 {
		return
	}

	// riskPrice：用于止损止盈/风控（MarkPrice 优先；缺失则用 pricePoint 兜底）
	riskPrice := ticker.EffectiveMarkPrice()
	if riskPrice <= 0 {
		riskPrice = pricePoint
	}

	// 【优化1】防止goroutine堆积（原子操作）
	if !atomic.CompareAndSwapInt32(&e.processingPriceUpdate, 0, 1) {
		return // 已有价格更新在处理中，跳过本次
	}
	defer atomic.StoreInt32(&e.processingPriceUpdate, 0)

	// 更新价格窗口（用于信号生成）
	e.priceLock.Lock()
	now := time.Now()
	e.PriceWindow = append(e.PriceWindow, PricePoint{
		Price:     pricePoint,
		Timestamp: now.UnixMilli(), // 使用毫秒时间戳
	})
	// 保持窗口大小
	window, _ := e.getRealTimeWindowAndThreshold()
	if window > 0 {
		cutoffTimestamp := now.Add(-time.Duration(window) * time.Second).UnixMilli()
		newWindow := make([]PricePoint, 0, len(e.PriceWindow))
		for _, p := range e.PriceWindow {
			if p.Timestamp > cutoffTimestamp {
				newWindow = append(newWindow, p)
			}
		}
		e.PriceWindow = newWindow
	}
	e.priceLock.Unlock()

	// 【核心优化】实时检查止损和止盈（WebSocket触发，毫秒级响应）
	// 只有在持仓存在时才检查（避免无意义的检查）
	e.mu.RLock()
	hasPosition := len(e.CurrentPositions) > 0
	autoCloseEnabled := e.Robot != nil && e.Robot.AutoCloseEnabled == 1
	e.mu.RUnlock()

	if hasPosition && autoCloseEnabled {
		// 异步执行平仓检查（避免阻塞价格更新）
		go func() {
			checkCtx := context.Background()
			e.checkStopLossAndClose(checkCtx, riskPrice)
			e.checkTakeProfitAndClose(checkCtx, riskPrice)
		}()
	}

	// 评估窗口信号（可能触发开仓）
	signal := e.EvaluateWindowSignal()
	if signal != nil {
		e.mu.Lock()
		e.LastSignal = signal
		e.LastSignalUpdate = time.Now()
		e.mu.Unlock()
	}
}

// doAnalysis 执行市场分析
// 【健壮性优化】添加 panic 恢复，确保单个分析失败不影响主循环
// 【纯事件驱动架构】当价格更新时，立即触发平仓检查
func (e *RobotEngine) doAnalysis(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "[RobotEngine] doAnalysis panic recovered: robotId=%d, err=%v",
				e.Robot.Id, r)
		}
	}()

	// 获取行情数据
	ticker := market.GetMarketServiceManager().GetTicker(e.Platform, e.Robot.Symbol)
	if ticker == nil || ticker.EffectiveMarkPrice() <= 0 {
		return
	}
	// pricePoint 用于信号窗口（优先成交价LastPrice；缺失则用MarkPrice兜底）
	pricePoint := ticker.LastPrice
	if pricePoint <= 0 {
		pricePoint = ticker.EffectiveMarkPrice()
	}
	// riskPrice 用于盈亏/止盈止损/风控（MarkPrice优先，LastPrice兜底）
	riskPrice := ticker.EffectiveMarkPrice()

	// 【效率优化】减少锁持有时间，快速更新
	e.mu.RLock()
	hasPosition := len(e.CurrentPositions) > 0
	e.mu.RUnlock()

	// 【优化】只有当有持仓时才检查价格变化，避免无持仓时的无效检查
	if !hasPosition {
		// 无持仓时，只更新价格数据，不触发平仓检查
		e.mu.Lock()
		e.LastTicker = ticker
		e.LastTickerUpdate = time.Now()
		e.mu.Unlock()
		// 添加价格点到窗口（用于信号生成）
		e.AddPricePoint(pricePoint)
		return
	}

	// 有持仓时，检查价格变化
	e.mu.Lock()
	oldPrice := 0.0
	if e.LastTicker != nil {
		oldPrice = e.LastTicker.EffectiveMarkPrice()
	}
	e.LastTicker = ticker
	e.LastTickerUpdate = time.Now()
	// 【优化】价格变化阈值：根据币种精度调整，避免微小波动触发检查
	priceChanged := (oldPrice == 0 || math.Abs(riskPrice-oldPrice) > 0.0001)
	e.mu.Unlock()

	// 【事件驱动】当价格更新时，立即触发以下操作：
	// 1. 更新订单未实现盈亏（基于实时价格计算，轻量级）
	// 2. 检查止损进度，达到100%时立即执行平仓
	// 3. 检查启动止盈和止盈回撤，达到条件时立即执行平仓
	if priceChanged {
		// 【并发控制】防止goroutine堆积：如果上一次价格处理还未完成，跳过本次
		// 使用原子操作 CAS 确保同一时刻只有一个goroutine在执行
		if !atomic.CompareAndSwapInt32(&e.processingPriceUpdate, 0, 1) {
			// 已有goroutine在处理，跳过本次
			g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 价格更新跳过（上次处理尚未完成）", e.Robot.Id)
		} else {
			g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 【事件驱动】价格更新，立即触发订单更新: price=%.4f, oldPrice=%.4f",
				e.Robot.Id, riskPrice, oldPrice)
			// 【修复竞态条件】先更新数据库，再检查止损/止盈
			// 使用串行执行避免竞态条件
			go func() {
				defer atomic.StoreInt32(&e.processingPriceUpdate, 0) // 处理完成后释放标志
				// 1. 先更新未实现盈亏到数据库
				e.updateOrdersUnrealizedPnl(ctx, riskPrice)
				// 2. 再检查止损（使用刚更新的数据）
				e.checkStopLossAndClose(ctx, riskPrice)
				// 3. 再检查止盈（使用刚更新的数据）
				e.checkTakeProfitAndClose(ctx, riskPrice)
			}()
		}
	}

	// 添加价格点到窗口（toogo实时信号逻辑）
	e.AddPricePoint(pricePoint)

	// 获取K线数据
	klines := market.GetMarketServiceManager().GetMultiTimeframeKlines(e.Platform, e.Robot.Symbol)
	if klines != nil {
		// 【效率优化】快速更新K线数据
		e.mu.Lock()
		e.LastKlines = klines
		e.mu.Unlock()
	}

	// 加载波动率配置（支持每个货币对独立配置）
	e.loadVolatilityConfig(ctx)

	// 【架构优化】从全局市场分析器获取市场状态（按 platform+symbol 共享，避免重复计算）
	// 每个币种（platform+symbol）有独立的市场状态信号，所有交易该币种的机器人共享同一套信号
	// 例如：bitget:BTCUSDT 的所有机器人共享同一套市场状态，binance:BTCUSDT 的机器人共享另一套
	// 统一使用全局服务，不降级到本地计算，确保所有机器人使用一致的市场状态
	globalAnalysis := market.GetMarketAnalyzer().GetAnalysis(e.Platform, e.Robot.Symbol)
	if globalAnalysis != nil {
		marketState := normalizeMarketState(string(globalAnalysis.MarketState))
		if marketState != "" {
			// 检测市场状态变化，并加载对应的策略配置
			e.checkAndUpdateStrategyConfig(ctx, marketState)
		}
	}
	// 如果全局服务不可用，直接跳过市场状态分析，不降级到本地计算
	// 这样可以确保所有机器人都使用统一的市场状态计算结果
}

// checkAndUpdateStrategyConfig 检测市场状态变化并更新策略配置
func (e *RobotEngine) checkAndUpdateStrategyConfig(ctx context.Context, currentMarketState string) {
	// 规范化市场状态
	currentMarketState = normalizeMarketState(currentMarketState)

	// 如果市场状态为空，不更新
	if currentMarketState == "" {
		g.Log().Debugf(ctx, "[RobotEngine] robotId=%d checkAndUpdateStrategyConfig跳过: 市场状态为空", e.Robot.Id)
		return
	}

	// 【优化】检查是否需要重新加载（策略未加载或市场状态变化）
	e.mu.RLock()
	needReload := e.CurrentStrategyParams == nil || e.LastMarketState != currentMarketState
	e.mu.RUnlock()

	if !needReload {
		return // 策略已加载且市场状态未变化，无需重新加载
	}

	// 【重要】始终使用创建机器人时保存的映射关系（从 remark 字段加载）
	// 映射关系是创建时保存的，如果用户修改了映射关系，会通过UpdateRobot实时生效
	// 根据当前市场状态，从映射关系中获取对应的风险偏好
	// 【严格模式】如果映射关系中没有找到，直接报错，不允许降级
	e.mu.RLock()
	riskPreference := e.MarketRiskMapping[currentMarketState]
	mappingCopy := make(map[string]string)
	for k, v := range e.MarketRiskMapping {
		mappingCopy[k] = v
	}
	e.mu.RUnlock()

	if riskPreference == "" {
		// 映射关系中没有找到，输出详细调试信息
		g.Log().Errorf(ctx, "[RobotEngine] robotId=%d 【策略加载失败】市场状态=%s 在映射关系中未找到对应的风险偏好。当前映射关系=%v，Remark字段=%s",
			e.Robot.Id, currentMarketState, mappingCopy, e.Robot.Remark)
		return // 不更新策略参数，保持为空，后续操作会检查并阻止
	}

	g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 从创建时的映射关系获取风险偏好: 市场状态=%s → 风险偏好=%s",
		e.Robot.Id, currentMarketState, riskPreference)

	g.Log().Infof(ctx, "[RobotEngine] robotId=%d 加载策略参数: 市场状态=%s, 风险偏好=%s(来自创建时的映射关系，用户修改后实时生效)",
		e.Robot.Id, currentMarketState, riskPreference)

	// 从策略模板加载完整策略参数（包括杠杆、保证金比例、止损等）
	// 查询条件：策略组ID + 当前市场状态 + 创建机器人时的风险偏好
	strategyParams, err := e.loadFullStrategyParams(ctx, currentMarketState, riskPreference)
	if err != nil {
		g.Log().Errorf(ctx, "[RobotEngine] robotId=%d 加载策略参数失败: %v", e.Robot.Id, err)
		// 不更新策略参数，保持为空，后续操作会检查并阻止
		return
	}

	// 更新当前策略参数和市场状态
	e.mu.Lock()
	e.CurrentStrategyParams = strategyParams
	e.LastMarketState = currentMarketState
	e.mu.Unlock()

	// 【重要】输出完整的策略参数，包括止盈止损参数
	g.Log().Infof(ctx, "[RobotEngine] robotId=%d 【策略加载成功】市场状态=%s, 风险偏好=%s, 止损=%.1f%%, 启动止盈=%.1f%%, 止盈回撤=%.1f%%, 杠杆=%d-%d, 保证金=%.1f-%.1f%%",
		e.Robot.Id, currentMarketState, riskPreference,
		strategyParams.StopLossPercent, strategyParams.AutoStartRetreatPercent, strategyParams.ProfitRetreatPercent,
		strategyParams.LeverageMin, strategyParams.LeverageMax,
		strategyParams.MarginPercentMin, strategyParams.MarginPercentMax)
}

// normalizeMarketState 规范化市场状态格式
// 统一格式: trend, volatile, high_vol, low_vol
// 兼容旧格式: range → volatile, high-volatility → high_vol, low-volatility → low_vol
func normalizeMarketState(marketState string) string {
	if marketState == "" {
		return "trend" // 默认值
	}

	switch marketState {
	case "range":
		return "volatile"
	case "high-volatility":
		return "high_vol"
	case "low-volatility":
		return "low_vol"
	case "trend", "volatile", "high_vol", "low_vol":
		return marketState
	default:
		// 未知格式，返回默认值
		g.Log().Warningf(context.Background(), "[RobotEngine] 未知市场状态格式: %s，使用默认值 trend", marketState)
		return "trend"
	}
}

// mapMarketStateToDb 将代码内部的市场状态映射到数据库存储的格式（兼容性函数，保留）
// 统一格式: trend, volatile, high_vol, low_vol
func mapMarketStateToDb(marketState string) string {
	// 使用规范化函数
	return normalizeMarketState(marketState)
}

// 默认策略参数配置（当策略模板找不到时使用）
// 按照：市场状态 -> 风险偏好 -> (时间窗口秒, 波动阈值USDT)
var defaultStrategyParams = map[string]map[string]struct {
	Window    int
	Threshold float64
}{
	"trend": {
		"conservative": {Window: 120, Threshold: 15},
		"balanced":     {Window: 90, Threshold: 12},
		"aggressive":   {Window: 60, Threshold: 8},
	},
	"volatile": {
		"conservative": {Window: 90, Threshold: 20},
		"balanced":     {Window: 60, Threshold: 15},
		"aggressive":   {Window: 45, Threshold: 10},
	},
	"high_vol": {
		"conservative": {Window: 60, Threshold: 25},
		"balanced":     {Window: 45, Threshold: 20},
		"aggressive":   {Window: 30, Threshold: 15},
	},
	"low_vol": {
		"conservative": {Window: 180, Threshold: 8},
		"balanced":     {Window: 120, Threshold: 6},
		"aggressive":   {Window: 90, Threshold: 5},
	},
}

// StrategyParams 策略参数（从策略模板加载的完整参数）
type StrategyParams struct {
	Window                  int     // 时间窗口(秒)
	Threshold               float64 // 波动阈值(USDT)
	LeverageMin             int     // 杠杆最小值
	LeverageMax             int     // 杠杆最大值
	MarginPercentMin        float64 // 保证金比例最小值
	MarginPercentMax        float64 // 保证金比例最大值
	StopLossPercent         float64 // 止损百分比
	ProfitRetreatPercent    float64 // 止盈回撤百分比
	AutoStartRetreatPercent float64 // 启动止盈百分比
}

// VolatilityConfig 波动率配置（市场状态阈值 + 5个时间周期权重）
type VolatilityConfig struct {
	HighVolatilityThreshold float64 // 高波动阈值（判断高波动市场）
	LowVolatilityThreshold  float64 // 低波动阈值（判断低波动市场）
	TrendStrengthThreshold  float64 // 趋势强度阈值（判断趋势市场）
	Weight1m                float64 // 1分钟周期权重
	Weight5m                float64 // 5分钟周期权重
	Weight15m               float64 // 15分钟周期权重
	Weight30m               float64 // 30分钟周期权重
	Weight1h                float64 // 1小时周期权重
	Symbol                  string  // 配置名称（交易对或"默认"/"全局"）
}

// LoadFullStrategyParams 从策略模板加载完整参数（公开方法，供外部调用）
func (e *RobotEngine) LoadFullStrategyParams(ctx context.Context, marketState, riskPreference string) (*StrategyParams, error) {
	return e.loadFullStrategyParams(ctx, marketState, riskPreference)
}

// RefreshStrategyParams 刷新策略参数缓存（强制重新加载）
// 当策略模板被修改时调用此方法，清除缓存并重新加载最新参数
func (e *RobotEngine) RefreshStrategyParams(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 清除当前策略参数缓存
	e.CurrentStrategyParams = nil

	// 【优化】从全局市场分析器获取当前市场状态
	marketState := ""
	globalAnalysis := market.GetMarketAnalyzer().GetAnalysis(e.Platform, e.Robot.Symbol)
	if globalAnalysis != nil {
		marketState = normalizeMarketState(string(globalAnalysis.MarketState))
	}
	if marketState == "" {
		marketState = "trend" // 默认值
	}
	marketState = normalizeMarketState(marketState) // 确保规范化

	// 【严格模式】从映射关系获取风险偏好，找不到直接报错
	riskPreference := e.MarketRiskMapping[marketState]
	if riskPreference == "" {
		errMsg := fmt.Sprintf("机器人ID=%d 市场状态=%s 在映射关系中未找到对应的风险偏好，无法刷新策略参数。请检查机器人的风险配置映射关系是否完整", e.Robot.Id, marketState)
		g.Log().Errorf(ctx, "[RobotEngine] %s", errMsg)
		return gerror.New(errMsg)
	}

	// 重新加载策略参数
	strategyParams, err := e.loadFullStrategyParams(ctx, marketState, riskPreference)
	if err != nil {
		g.Log().Errorf(ctx, "[RobotEngine] robotId=%d 刷新策略参数失败: %v", e.Robot.Id, err)
		return err
	}

	// 更新缓存
	e.CurrentStrategyParams = strategyParams

	// 【已废弃】不再更新监控配置，窗口和波动值现在实时获取
	// 策略参数已加载到 CurrentStrategyParams，窗口和波动值会在使用时实时获取

	g.Log().Infof(ctx, "[RobotEngine] robotId=%d 策略参数缓存已刷新: market=%s, risk=%s, 窗口=%ds, 波动=%.2f, 杠杆=%d-%d, 保证金=%.1f-%.1f%%",
		e.Robot.Id, marketState, riskPreference,
		strategyParams.Window, strategyParams.Threshold,
		strategyParams.LeverageMin, strategyParams.LeverageMax,
		strategyParams.MarginPercentMin, strategyParams.MarginPercentMax)

	return nil
}

// loadFullStrategyParams 从策略模板加载完整参数（包括杠杆、保证金等）
// 如果找不到策略模板，返回错误，不使用默认值
func (e *RobotEngine) loadFullStrategyParams(ctx context.Context, marketState, riskPreference string) (*StrategyParams, error) {
	params := &StrategyParams{}

	// 规范化市场状态（统一格式）
	normalizedMarketState := normalizeMarketState(marketState)

	// 1. 获取策略组ID（优先级：机器人.StrategyGroupId > CurrentStrategy.group_id）
	var groupId int64 = 0

	// 1.1 优先使用机器人绑定的策略组ID
	if e.Robot.StrategyGroupId > 0 {
		groupId = e.Robot.StrategyGroupId
		g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 使用机器人绑定的策略组: groupId=%d", e.Robot.Id, groupId)
	}

	// 1.2 其次从 CurrentStrategy JSON 中获取（兼容旧数据）
	if groupId == 0 && e.Robot.CurrentStrategy != "" {
		var strategyData map[string]interface{}
		if err := json.Unmarshal([]byte(e.Robot.CurrentStrategy), &strategyData); err == nil {
			// 支持 groupId 和 group_id 两种格式（兼容旧数据）
			if gid, ok := strategyData["groupId"].(float64); ok {
				groupId = int64(gid)
				g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 从CurrentStrategy获取策略组: groupId=%d", e.Robot.Id, groupId)
			} else if gid, ok := strategyData["group_id"].(float64); ok {
				groupId = int64(gid)
				g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 从CurrentStrategy获取策略组: group_id=%d", e.Robot.Id, groupId)
			}
		}
	}

	// 2. 检查是否有策略组ID
	if groupId == 0 {
		errMsg := fmt.Sprintf("机器人ID=%d 未绑定策略组ID，无法加载策略参数", e.Robot.Id)
		g.Log().Errorf(ctx, "[RobotEngine] %s", errMsg)
		return nil, gerror.New(errMsg)
	}

	// 3. 从策略模板表中查询对应的策略（尝试多种市场状态名称，兼容旧数据）
	marketStatesToTry := []string{
		normalizedMarketState, // 标准格式（优先级最高）
	}

	// 如果原始格式与规范化格式不同，添加原始格式
	if normalizedMarketState != marketState {
		marketStatesToTry = append(marketStatesToTry, marketState)
	}

	// 添加兼容格式（仅当与标准格式不同时）
	if normalizedMarketState == "volatile" && marketState != "volatile" {
		marketStatesToTry = append(marketStatesToTry, "range") // 兼容旧格式
	}
	if normalizedMarketState == "high_vol" && marketState != "high_vol" {
		marketStatesToTry = append(marketStatesToTry, "high-volatility") // 兼容数据库格式
	}
	if normalizedMarketState == "low_vol" && marketState != "low_vol" {
		marketStatesToTry = append(marketStatesToTry, "low-volatility") // 兼容数据库格式
	}

	for _, ms := range marketStatesToTry {
		var strategy *entity.TradingStrategyTemplate
		err := dao.TradingStrategyTemplate.Ctx(ctx).
			Where("group_id", groupId).
			Where(dao.TradingStrategyTemplate.Columns().MarketState, ms).
			Where(dao.TradingStrategyTemplate.Columns().RiskPreference, riskPreference).
			// 移除 IsActive 限制，只要策略模板存在就可以使用
			Scan(&strategy)

		if err == nil && strategy != nil {
			params.Window = strategy.MonitorWindow
			params.Threshold = strategy.VolatilityThreshold
			params.LeverageMin = strategy.Leverage
			params.LeverageMax = strategy.Leverage
			params.MarginPercentMin = strategy.MarginPercent
			params.MarginPercentMax = strategy.MarginPercent
			params.StopLossPercent = strategy.StopLossPercent
			params.ProfitRetreatPercent = strategy.ProfitRetreatPercent
			params.AutoStartRetreatPercent = strategy.AutoStartRetreatPercent

			g.Log().Infof(ctx, "[RobotEngine] robotId=%d 从策略模板加载参数: market=%s(规范化=%s,查询=%s), risk=%s, 窗口=%d, 波动=%.1f, 杠杆=%d, 保证金=%.1f%%, 止损=%.1f%%, 启动止盈=%.1f%%, 止盈回撤=%.1f%%",
				e.Robot.Id, marketState, normalizedMarketState, ms, riskPreference,
				params.Window, params.Threshold,
				params.LeverageMin, params.MarginPercentMin,
				params.StopLossPercent, params.AutoStartRetreatPercent, params.ProfitRetreatPercent)
			return params, nil
		}
	}

	// 4. 找不到策略模板，返回详细错误信息
	// 查询策略组中所有可用的市场状态和风险偏好组合，帮助用户了解可用的配置
	var availableCombinations []struct {
		MarketState    string `json:"marketState"`
		RiskPreference string `json:"riskPreference"`
	}
	_ = dao.TradingStrategyTemplate.Ctx(ctx).
		Fields("market_state", "risk_preference").
		Where("group_id", groupId).
		Group("market_state", "risk_preference").
		Scan(&availableCombinations)

	var availableInfo string
	if len(availableCombinations) > 0 {
		availableInfo = "可用配置: "
		for _, combo := range availableCombinations {
			availableInfo += fmt.Sprintf("%s/%s ", combo.MarketState, combo.RiskPreference)
		}
	} else {
		availableInfo = "策略组中没有任何策略模板"
	}

	errMsg := fmt.Sprintf("机器人ID=%d 找不到策略模板: groupId=%d, marketState=%s/%s, riskPreference=%s。%s",
		e.Robot.Id, groupId, marketState, normalizedMarketState, riskPreference, availableInfo)
	g.Log().Errorf(ctx, "[RobotEngine] %s", errMsg)
	return nil, gerror.New(errMsg)
}

// loadStrategyParams 从策略模板加载参数
// 返回值：window=时间窗口(秒), threshold=波动阈值(USDT)
// 如果找不到策略模板，使用默认参数
func (e *RobotEngine) loadStrategyParams(ctx context.Context, marketState, riskPreference string) (window int, threshold float64) {
	// 1. 优先从机器人的 CurrentStrategy JSON 中获取策略组ID和参数
	var groupId int64 = 0
	if e.Robot.CurrentStrategy != "" {
		var strategyData map[string]interface{}
		if err := json.Unmarshal([]byte(e.Robot.CurrentStrategy), &strategyData); err == nil {
			if gid, ok := strategyData["group_id"].(float64); ok {
				groupId = int64(gid)
			}
			// 尝试从顶层直接获取时间窗口和波动值
			if mw, ok := strategyData["monitor_window"].(float64); ok && mw > 0 {
				window = int(mw)
			}
			if vt, ok := strategyData["volatility_threshold"].(float64); ok && vt > 0 {
				threshold = vt
			}
			// 如果从JSON中获取到了参数，直接返回
			if window > 0 && threshold > 0 {
				g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 从CurrentStrategy加载: window=%d, threshold=%.2f",
					e.Robot.Id, window, threshold)
				return
			}
		}
	}

	// 2. 如果有策略组ID，从策略模板表中查询对应的策略
	if groupId > 0 {
		var strategy *entity.TradingStrategyTemplate
		err := dao.TradingStrategyTemplate.Ctx(ctx).
			Where("group_id", groupId).
			Where(dao.TradingStrategyTemplate.Columns().MarketState, marketState).
			Where(dao.TradingStrategyTemplate.Columns().RiskPreference, riskPreference).
			Where(dao.TradingStrategyTemplate.Columns().IsActive, 1).
			Scan(&strategy)

		if err == nil && strategy != nil {
			window = strategy.MonitorWindow
			threshold = strategy.VolatilityThreshold
			g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 从策略组%d加载: market=%s, risk=%s, window=%d, threshold=%.2f",
				e.Robot.Id, groupId, marketState, riskPreference, window, threshold)
			return
		} else {
			g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 策略组%d未找到模板: market=%s, risk=%s, 将使用默认参数",
				e.Robot.Id, groupId, marketState, riskPreference)
		}
	}

	// 3. 从默认策略参数获取（回退机制）
	if marketParams, ok := defaultStrategyParams[marketState]; ok {
		if riskParams, ok := marketParams[riskPreference]; ok {
			window = riskParams.Window
			threshold = riskParams.Threshold
			g.Log().Infof(ctx, "[RobotEngine] robotId=%d 使用默认策略参数: market=%s, risk=%s, window=%d, threshold=%.2f",
				e.Robot.Id, marketState, riskPreference, window, threshold)
			return
		}
	}

	// 4. 最终回退：使用平衡型默认值
	window = 60
	threshold = 15
	g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 使用最终默认参数: window=%d, threshold=%.2f",
		e.Robot.Id, window, threshold)
	return
}

// ==================== 智能 API 调用（防止重复请求和goroutine堆积） ====================

// GetPositionsSmart 智能获取持仓（使用缓存 + singleflight 模式）
// maxCacheAge: 缓存最大有效期，超过则重新获取
// 特性：
//  1. 缓存有效期内直接返回缓存
//  2. 使用 singleflight 模式，同一时刻只有一个 goroutine 真正调用 API
//  3. 其他 goroutine 等待结果或使用旧缓存
func (e *RobotEngine) GetPositionsSmart(ctx context.Context, maxCacheAge time.Duration) ([]*exchange.Position, error) {
	// 1. 检查缓存是否有效
	e.mu.RLock()
	cachedPositions := e.CurrentPositions
	lastUpdate := e.LastPositionUpdate
	e.mu.RUnlock()

	if time.Since(lastUpdate) < maxCacheAge {
		return cachedPositions, nil
	}

	// 2. 缓存过期，尝试获取新数据
	// 使用 CAS 确保只有一个 goroutine 执行实际的 API 调用
	if !atomic.CompareAndSwapInt32(&e.positionFetching, 0, 1) {
		// 已有其他 goroutine 在获取，等待一小段时间后返回缓存
		// 不阻塞等待，直接返回旧缓存，避免 goroutine 堆积
		g.Log().Debugf(ctx, "[RobotEngine] robotId=%d GetPositions请求合并，使用缓存", e.Robot.Id)
		return cachedPositions, nil
	}

	// 3. 执行实际的 API 调用
	defer atomic.StoreInt32(&e.positionFetching, 0)

	if e.Exchange == nil {
		return cachedPositions, gerror.New("交易所实例不存在")
	}

	positions, err := e.Exchange.GetPositions(ctx, e.Robot.Symbol)
	if err != nil {
		// API 失败，返回旧缓存
		g.Log().Debugf(ctx, "[RobotEngine] robotId=%d GetPositions失败，使用旧缓存: %v", e.Robot.Id, err)
		return cachedPositions, nil
	}

	// 4. 更新缓存
	e.mu.Lock()
	e.CurrentPositions = positions
	e.LastPositionUpdate = time.Now()
	e.mu.Unlock()

	return positions, nil
}

// GetBalanceSmart 智能获取余额（使用缓存 + singleflight 模式）
func (e *RobotEngine) GetBalanceSmart(ctx context.Context, maxCacheAge time.Duration) (*exchange.Balance, error) {
	// 1. 检查缓存是否有效
	e.mu.RLock()
	cachedBalance := e.AccountBalance
	lastUpdate := e.LastBalanceUpdate
	e.mu.RUnlock()

	if cachedBalance != nil && time.Since(lastUpdate) < maxCacheAge {
		return cachedBalance, nil
	}

	// 2. 缓存过期，尝试获取新数据
	if !atomic.CompareAndSwapInt32(&e.balanceFetching, 0, 1) {
		// 已有其他 goroutine 在获取，返回缓存
		g.Log().Debugf(ctx, "[RobotEngine] robotId=%d GetBalance请求合并，使用缓存", e.Robot.Id)
		return cachedBalance, nil
	}

	// 3. 执行实际的 API 调用
	defer atomic.StoreInt32(&e.balanceFetching, 0)

	if e.Exchange == nil {
		return cachedBalance, gerror.New("交易所实例不存在")
	}

	balance, err := e.Exchange.GetBalance(ctx)
	if err != nil {
		g.Log().Debugf(ctx, "[RobotEngine] robotId=%d GetBalance失败，使用旧缓存: %v", e.Robot.Id, err)
		return cachedBalance, nil
	}

	// 4. 更新缓存
	e.mu.Lock()
	e.AccountBalance = balance
	e.LastBalanceUpdate = time.Now()
	e.mu.Unlock()

	return balance, nil
}

// ForceRefreshPositions 强制刷新持仓（用于平仓等关键操作）
// 直接调用 API，忽略缓存，但仍使用 singleflight 防止并发
func (e *RobotEngine) ForceRefreshPositions(ctx context.Context) ([]*exchange.Position, error) {
	if e.Exchange == nil {
		return nil, gerror.New("交易所实例不存在")
	}

	// 使用 singleflight 模式
	if !atomic.CompareAndSwapInt32(&e.positionFetching, 0, 1) {
		// 等待其他 goroutine 完成，然后返回最新缓存
		time.Sleep(100 * time.Millisecond)
		e.mu.RLock()
		positions := e.CurrentPositions
		e.mu.RUnlock()
		return positions, nil
	}
	defer atomic.StoreInt32(&e.positionFetching, 0)

	positions, err := e.Exchange.GetPositions(ctx, e.Robot.Symbol)
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	e.CurrentPositions = positions
	e.LastPositionUpdate = time.Now()
	e.mu.Unlock()

	return positions, nil
}

// syncAccountData 同步账户数据（持仓、订单历史）
// 【事件驱动优化】只同步持仓，余额只在下单前获取
// 减少不必要的API调用
func (e *RobotEngine) syncAccountData(ctx context.Context) {
	// 【优化】余额不再定期同步，只在下单前通过 GetBalanceSmart(ctx, 0) 获取
	// 这样可以减少一半的API调用

	// 获取持仓（使用智能方法）
	positions, err := e.GetPositionsSmart(ctx, 0) // 强制刷新获取最新持仓
	if err == nil {
		// 【新增】检测手动平仓（在更新内存状态前）
		e.detectManualClose(ctx, positions)

		e.mu.Lock()
		e.CurrentPositions = positions
		e.LastPositionUpdate = time.Now()
		e.LastSyncError = nil
		e.SyncErrorCount = 0
		e.mu.Unlock()
	} else {
		e.mu.Lock()
		e.LastSyncError = err
		e.SyncErrorCount++
		e.mu.Unlock()
		g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 获取持仓失败: %v (连续错误%d次)", e.Robot.Id, err, e.SyncErrorCount)
	}

	// 【优化】订单历史每120秒同步一次（进一步降低频率）
	e.mu.RLock()
	lastOrderHistoryUpdate := e.LastOrderHistoryUpdate
	e.mu.RUnlock()
	if time.Since(lastOrderHistoryUpdate) >= 120*time.Second {
		orderHistory, err := e.Exchange.GetOrderHistory(ctx, e.Robot.Symbol, 50)
		if err == nil {
			e.mu.Lock()
			e.OrderHistory = orderHistory
			e.LastOrderHistoryUpdate = time.Now()
			e.mu.Unlock()
		}
	}
}

// refreshBalanceCacheAfterTrade 交易事件后刷新余额缓存（用于机器人列表显示的账户权益/钱包余额）
// 设计目标：
// - 不做高频定时拉取，避免占用API
// - 在“下单前/平仓后/检测到手动平仓”等关键事件点顺便刷新
// - 使用 GetBalanceSmart 的去重能力，避免并发堆积
func (e *RobotEngine) refreshBalanceCacheAfterTrade(ctx context.Context, reason string) {
	// 简单节流：2秒内不重复刷新（同一笔平仓可能触发多条路径）
	e.mu.RLock()
	last := e.LastBalanceUpdate
	e.mu.RUnlock()
	if !last.IsZero() && time.Since(last) < 2*time.Second {
		return
	}

	bal, err := e.GetBalanceSmart(ctx, 0) // 强制刷新（smart内部会做去重/降级）
	if err != nil || bal == nil {
		g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 刷新余额缓存失败(忽略): reason=%s, err=%v", e.Robot.Id, reason, err)
		return
	}

	// GetBalanceSmart 内部已更新 AccountBalance/LastBalanceUpdate，这里只做日志
	g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 余额缓存已刷新: reason=%s, available=%.4f, total=%.4f",
		e.Robot.Id, reason, bal.AvailableBalance, bal.TotalBalance)
}

// GetCachedPositions 获取缓存的持仓数据（供其他模块使用）
func (e *RobotEngine) GetCachedPositions() ([]*exchange.Position, time.Time) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.CurrentPositions, e.LastPositionUpdate
}

// GetCachedOrderHistory 获取缓存的订单历史（供其他模块使用）
func (e *RobotEngine) GetCachedOrderHistory() ([]*exchange.Order, time.Time) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.OrderHistory, e.LastOrderHistoryUpdate
}

// updateOrderStatusAfterClose 平仓成功后更新数据库订单状态
// 【优化】直接更新数据库，不需要调用API同步
// closeType: "stop_loss"/"take_profit"/"manual"/"unknown"
func (e *RobotEngine) updateOrderStatusAfterClose(ctx context.Context, pos *exchange.Position, closeOrder *exchange.Order, closeType string) {
	robot := e.Robot
	if robot == nil {
		return
	}

	// 【新增】落库成交流水（幂等去重），用于“成交流水”页面展示手续费/已实现盈亏
	// 说明：自动平仓/止损止盈/引擎内平仓都走这里；如果不落库，会导致页面看起来“没有自动更新”。
	{
		sym := strings.TrimSpace(robot.Symbol)
		if pos != nil && strings.TrimSpace(pos.Symbol) != "" {
			sym = strings.TrimSpace(pos.Symbol)
		}
		if sym != "" {
			if saved, matched, err := fetchAndStoreTradeHistory(ctx, e.Exchange, robot.ApiConfigId, robot.Exchange, sym, 800); err != nil {
				g.Log().Warningf(ctx, "[RobotEngine] 平仓后落库成交流水失败(不影响平仓): robotId=%d, closeType=%s, symbol=%s, err=%v",
					robot.Id, closeType, sym, err)
			} else {
				g.Log().Debugf(ctx, "[RobotEngine] 平仓后已落库成交流水: robotId=%d, closeType=%s, symbol=%s, saved=%d, matched=%d",
					robot.Id, closeType, sym, saved, matched)
			}
		}
	}

	// 确定方向
	direction := "long"
	if pos.PositionSide == "SHORT" {
		direction = "short"
	}

	// 查询本地持仓中订单
	var localOrder *entity.TradingOrder
	err := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", robot.Id).
		Where("direction", direction).
		Where("status", OrderStatusOpen).
		OrderDesc("id").
		Limit(1).
		Scan(&localOrder)
	if err != nil || localOrder == nil {
		if err != nil {
			g.Log().Errorf(ctx, "[RobotEngine] robotId=%d 平仓后同步失败(查询本地OPEN订单失败): direction=%s, err=%v", robot.Id, direction, err)
		}
		return
	}

	// 平仓价格/盈亏/手续费：优先按“平仓订单ID”从成交(fill)记录汇总（以平台为准）
	closePrice := float64(0)
	if closeOrder != nil && closeOrder.AvgPrice > 0 {
		closePrice = closeOrder.AvgPrice
	}
	var (
		realizedProfit float64
		closeFee       float64
		closeFeeCoin   string
		closeTs        int64
	)
	if closeOrder != nil && strings.TrimSpace(closeOrder.OrderId) != "" {
		// 有明确平仓 orderId：直接按 orderId 汇总，避免“方向+时间”猜测误匹配
		if agg, ok := tryAggFromTradeHistoryByOrderID(ctx, e.Exchange, robot.Symbol, closeOrder.OrderId, 800); ok {
			if agg.AvgPrice > 0 {
				closePrice = agg.AvgPrice
			}
			realizedProfit = agg.RealizedPnl
			closeFee = agg.Commission
			closeFeeCoin = agg.FeeCoin
			closeTs = agg.MaxTs
		}
	}

	// 兜底：如果汇总失败，再尝试用“方向+开仓时间”推断（用于部分交易所回传 orderId 不稳定的情况）
	if (closePrice <= 0 || closeTs <= 0) && localOrder.OpenTime != nil {
		positionSide := "LONG"
		if direction == "short" {
			positionSide = "SHORT"
		}
		if cp, rp, fee, feeCoin, oid, ts, ok := tryCloseInfoFromTradeHistory(ctx, e.Exchange, robot.Symbol, positionSide, localOrder.OpenTime); ok {
			if cp > 0 {
				closePrice = cp
			}
			realizedProfit = rp
			closeFee = fee
			closeFeeCoin = feeCoin
			closeTs = ts
			// 若 closeOrderId 为空，则用推断到的
			if closeOrder != nil && strings.TrimSpace(closeOrder.OrderId) == "" && oid != "" {
				closeOrder.OrderId = oid
			}
		}
	}

	// 最后兜底：避免业务无法结算（算力扣除/佣金）——使用估算盈亏
	if math.Abs(realizedProfit) < 0.0000001 {
		openPrice := localOrder.OpenPrice
		if openPrice <= 0 && pos != nil && pos.EntryPrice > 0 {
			openPrice = pos.EntryPrice
		}
		qty := localOrder.Quantity
		if qty <= 0 && pos != nil && math.Abs(pos.PositionAmt) > 0.0001 {
			qty = math.Abs(pos.PositionAmt)
		}
		if closePrice <= 0 && pos != nil && pos.MarkPrice > 0 {
			closePrice = pos.MarkPrice
		}
		if closePrice > 0 && openPrice > 0 && qty > 0 {
			if direction == "long" {
				realizedProfit = (closePrice - openPrice) * qty
			} else {
				realizedProfit = (openPrice - closePrice) * qty
			}
		}
	}

	// 统一走 CloseOrder：补全平仓字段 + 扣算力（盈利单）
	finalCloseOrder := closeOrder
	if finalCloseOrder == nil {
		finalCloseOrder = &exchange.Order{}
	}
	// CloseOrder 用 CreateTime/UpdateTime 作为 close_time；这里尽量用成交的 maxTs
	if closeTs > 0 {
		finalCloseOrder.CreateTime = closeTs
		finalCloseOrder.UpdateTime = closeTs
	}
	// CloseOrder 会把 Fee/FeeCoin 写到 close_fee/close_fee_coin
	if closeFee > 0 {
		finalCloseOrder.Fee = closeFee
	}
	if closeFeeCoin != "" {
		finalCloseOrder.FeeCoin = closeFeeCoin
	}

	GetOrderStatusSyncService().CloseOrder(ctx, localOrder, closePrice, realizedProfit, closeType, finalCloseOrder, pos)
}

// removePositionFromCache 从内存缓存中移除已平仓的持仓
func (e *RobotEngine) removePositionFromCache(positionSide string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.CurrentPositions == nil {
		return
	}

	// 直接从缓存中移除该方向的持仓，避免“0数量残留对象”导致前端仍渲染
	newPositions := make([]*exchange.Position, 0, len(e.CurrentPositions))
	for _, pos := range e.CurrentPositions {
		if pos == nil {
			continue
		}
		if strings.ToUpper(strings.TrimSpace(pos.PositionSide)) == strings.ToUpper(strings.TrimSpace(positionSide)) {
			continue
		}
		newPositions = append(newPositions, pos)
	}
	e.CurrentPositions = newPositions
	e.LastPositionUpdate = time.Now()
}

// detectManualClose 检测手动平仓（交易所手动平仓）
// 【修复】检测到手动平仓后，立即更新数据库订单状态
func (e *RobotEngine) detectManualClose(ctx context.Context, exchangePositions []*exchange.Position) {
	// 构建交易所持仓映射（只统计有持仓的方向）
	exchangePosMap := make(map[string]bool)
	for _, pos := range exchangePositions {
		if math.Abs(pos.PositionAmt) > 0.0001 {
			exchangePosMap[pos.PositionSide] = true
		}
	}

	// 检查本地内存中的持仓
	e.mu.RLock()
	localPositions := e.CurrentPositions
	e.mu.RUnlock()

	if len(localPositions) == 0 {
		return
	}

	// 检查是否有手动平仓
	for _, localPos := range localPositions {
		if math.Abs(localPos.PositionAmt) > 0.0001 {
			// 本地有持仓
			if !exchangePosMap[localPos.PositionSide] {
				// 交易所已无该方向持仓，但本地有持仓 → 检测到手动平仓
				g.Log().Infof(ctx, "[RobotEngine] 检测到手动平仓: robotId=%d, side=%s",
					e.Robot.Id, localPos.PositionSide)

				// 【修复】立即更新数据库订单状态为已平仓（补全平仓价/盈亏，并触发扣算力）
				e.updateOrderStatusOnManualClose(ctx, localPos)

				// 使用统一的清除方法清除内存中的持仓状态
				e.ClearPosition(ctx, localPos.PositionSide)

				// 清除持仓跟踪器
				e.ClearPositionTracker(localPos.PositionSide)

				// 【页面显示优化】检测到手动平仓后，顺便刷新余额缓存
				e.refreshBalanceCacheAfterTrade(ctx, "after_manual_close_detected")

				g.Log().Infof(ctx, "[RobotEngine] robotId=%d 手动平仓处理完成: side=%s，已更新数据库和内存", e.Robot.Id, localPos.PositionSide)
			}
		}
	}
}

// updateOrderStatusOnManualClose 手动平仓后更新数据库订单状态
func (e *RobotEngine) updateOrderStatusOnManualClose(ctx context.Context, localPos *exchange.Position) {
	robot := e.Robot
	if robot == nil {
		return
	}
	if localPos == nil {
		return
	}

	// 【新增】落库成交流水（幂等去重）
	// 场景：用户在交易所/外部手动平仓 → 引擎检测到持仓消失并结算订单。
	// 若这里不落库，订单会结算但“成交流水”页面没有新记录。
	{
		sym := strings.TrimSpace(localPos.Symbol)
		if sym == "" {
			sym = strings.TrimSpace(robot.Symbol)
		}
		if sym != "" {
			if saved, matched, err := fetchAndStoreTradeHistory(ctx, e.Exchange, robot.ApiConfigId, robot.Exchange, sym, 800); err != nil {
				g.Log().Warningf(ctx, "[RobotEngine] 外部手动平仓检测后落库成交流水失败(不影响结算): robotId=%d, symbol=%s, err=%v",
					robot.Id, sym, err)
			} else {
				g.Log().Debugf(ctx, "[RobotEngine] 外部手动平仓检测后已落库成交流水: robotId=%d, symbol=%s, saved=%d, matched=%d",
					robot.Id, sym, saved, matched)
			}
		}
	}

	// 确定方向
	direction := "long"
	if localPos.PositionSide == "SHORT" {
		direction = "short"
	}

	// 查询本地持仓中订单
	var order *entity.TradingOrder
	err := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", robot.Id).
		Where("direction", direction).
		Where("status", OrderStatusOpen).
		OrderDesc("id").
		Limit(1).
		Scan(&order)

	if err != nil {
		g.Log().Errorf(ctx, "[RobotEngine] robotId=%d 手动平仓同步失败(查询本地订单失败): direction=%s, err=%v", robot.Id, direction, err)
		return
	}
	if order == nil {
		// 数据库没有 OPEN 订单也不应该阻塞：只清理内存，避免机器人持续误判持仓
		g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 检测到平台手动平仓但本地无OPEN订单: direction=%s", robot.Id, direction)
		return
	}

	// 获取平仓价格：优先行情缓存，兜底调用交易所
	closePrice := order.OpenPrice
	ticker := market.GetMarketServiceManager().GetTicker(robot.Exchange, localPos.Symbol)
	if ticker != nil && ticker.EffectiveMarkPrice() > 0 {
		// 平仓估值：MarkPrice优先，LastPrice兜底（更贴近交易所风险口径）
		closePrice = ticker.EffectiveMarkPrice()
	} else {
		apiTicker, err2 := e.Exchange.GetTicker(ctx, localPos.Symbol)
		if err2 == nil && apiTicker != nil && apiTicker.EffectiveMarkPrice() > 0 {
			closePrice = apiTicker.EffectiveMarkPrice()
		}
	}

	// 优先从交易所成交记录补齐“平仓价/已实现盈亏/平仓手续费/平仓时间/平仓订单ID”，以交易所为准
	var (
		realizedProfit float64
		closeOrder     *exchange.Order
	)
	positionSide := "LONG"
	if direction == "short" {
		positionSide = "SHORT"
	}
	if cp, rp, fee, feeCoin, oid, ts, ok := tryCloseInfoFromTradeHistory(ctx, e.Exchange, localPos.Symbol, positionSide, order.OpenTime); ok {
		closePrice = cp
		realizedProfit = rp
		closeOrder = &exchange.Order{OrderId: oid, Fee: fee, FeeCoin: feeCoin, CreateTime: ts, UpdateTime: ts}
	} else {
		// 兜底：用行情估算（可能与交易所实际有偏差）
		openPrice := order.OpenPrice
		if openPrice <= 0 && localPos.EntryPrice > 0 {
			openPrice = localPos.EntryPrice
		}
		qty := order.Quantity
		if qty <= 0 && math.Abs(localPos.PositionAmt) > 0.0001 {
			qty = math.Abs(localPos.PositionAmt)
		}
		if direction == "long" {
			realizedProfit = (closePrice - openPrice) * qty
		} else {
			realizedProfit = (openPrice - closePrice) * qty
		}
	}

	// 统一走 CloseOrder：补全平仓字段 + 扣算力（盈利单）
	GetOrderStatusSyncService().CloseOrder(ctx, order, closePrice, realizedProfit, "手动平仓(平台)", closeOrder, nil)
}

// syncPositionsToDatabase 同步持仓状态到数据库（确保内存与数据库一致）
// 【重要】因为开仓检查使用内存数据，必须确保内存与数据库一致
func (e *RobotEngine) syncPositionsToDatabase(ctx context.Context, positions []*exchange.Position) {
	// 构建持仓映射
	positionMap := make(map[string]bool)
	for _, pos := range positions {
		if math.Abs(pos.PositionAmt) > 0.0001 {
			positionSide := pos.PositionSide
			positionMap[positionSide] = true
		}
	}

	// 查询本地"持仓中"的订单
	var localOrders []*entity.TradingOrder
	err := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", e.Robot.Id).
		Where("status", OrderStatusOpen). // 持仓中（使用统一的订单状态常量）
		Scan(&localOrders)
	if err != nil {
		return
	}

	// 更新订单状态：如果内存中没有持仓但数据库显示持仓中，更新数据库状态
	for _, order := range localOrders {
		positionSide := "LONG"
		if order.Direction == "short" {
			positionSide = "SHORT"
		}

		hasPosition := positionMap[positionSide]
		if !hasPosition {
			// 内存中没有持仓，但数据库显示持仓中，更新数据库状态为已平仓
			// 这种情况可能是手动平仓，由订单同步服务处理，这里只记录日志
			g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 内存中无持仓但数据库显示持仓中: orderId=%d, positionSide=%s，等待订单同步服务处理", e.Robot.Id, order.Id, positionSide)
		}
	}
}

// syncAccountDataIfNeeded 智能同步账户数据（时间窗口 + 条件判断）
// 策略：
//   - 开仓前：强制同步（确保准确性）
//   - 平仓检查：有持仓时3秒内不重复同步，无持仓时5秒内不重复同步
//   - 定期同步：10秒一次（兜底）
//   - 错误重试：连续错误时降低同步频率，避免API限流
func (e *RobotEngine) syncAccountDataIfNeeded(ctx context.Context, syncType string) bool {
	e.mu.RLock()
	lastPositionUpdate := e.LastPositionUpdate
	syncErrorCount := e.SyncErrorCount
	hasPosition := false
	for _, pos := range e.CurrentPositions {
		if pos.PositionAmt != 0 {
			hasPosition = true
			break
		}
	}
	e.mu.RUnlock()

	timeSinceLastPositionUpdate := time.Since(lastPositionUpdate)

	// ========== 【事件驱动同步架构】==========
	// 核心原则：减少API调用，只在必要时同步
	// - 无持仓时：完全不轮询同步
	// - 有持仓时：使用内存缓存的持仓数据，不主动调用API
	// - 同步时机：仅在 下单前、平仓后 两个关键节点调用API

	// 【场景1】开仓前：必须同步（获取最新持仓和余额）
	if syncType == "before_open" {
		e.syncAccountData(ctx)
		return true
	}

	// 【场景2】交易后：延迟同步确保状态更新
	// 【优化】检查距离上次同步是否超过2秒，避免频繁调用API
	if syncType == "after_trade" {
		if timeSinceLastPositionUpdate < 2*time.Second {
			// 2秒内已同步过，跳过本次（减少API调用）
			return false
		}
		// 延迟1秒后同步，等待交易所处理完成
		time.Sleep(1 * time.Second)
		e.syncAccountData(ctx)
		return true
	}

	// 【场景3】无持仓时：完全不同步，节省API调用
	if !hasPosition {
		return false
	}

	// 【场景4】有持仓时的定期同步：大幅降低频率（60秒一次）
	// 止损/止盈检查使用内存缓存的 CurrentPositions，不需要频繁刷新
	// 只有持仓数据过期超过60秒才同步，用于保持数据最终一致性
	if syncType == "periodic" {
		if timeSinceLastPositionUpdate >= 60*time.Second {
			e.syncAccountData(ctx)
			return true
		}
		return false
	}

	// 【场景5】平仓检查：使用内存缓存，不调用API
	if syncType == "before_close_check" {
		// 止损/止盈计算使用 e.CurrentPositions（内存缓存）
		// 不需要每次检查都调用API，只有执行平仓时才需要刷新
		return false
	}

	// 【错误重试场景】连续错误时降低频率
	if syncErrorCount > 0 {
		minInterval := time.Duration(syncErrorCount) * 10 * time.Second
		if minInterval > 60*time.Second {
			minInterval = 60 * time.Second
		}
		if timeSinceLastPositionUpdate < minInterval {
			return false
		}
	}

	// 默认：不同步（事件驱动架构下，默认不主动同步）
	return false
}

// doSignalGeneration 生成方向信号
// 【简化】信号生成和下单触发已在 EvaluateWindowSignal 中完成
// 此函数只负责更新引擎状态，不再重复触发下单检查
func (e *RobotEngine) doSignalGeneration(ctx context.Context) {
	signal := e.SignalGen.Generate(ctx)
	if signal == nil {
		return
	}

	// 更新引擎状态（用于页面展示等）
	e.mu.Lock()
	e.LastSignal = signal
	e.LastSignalUpdate = time.Now()
	e.mu.Unlock()

	if signal.Direction != "NEUTRAL" {
		g.Log().Debugf(ctx, "[RobotEngine] robotId=%d, 方向信号: %s, 强度=%.2f%%, 置信度=%.2f%%",
			e.Robot.Id, signal.Direction, signal.Strength, signal.Confidence)
	}

	// 【重要】不再触发下单检查！
	// 下单逻辑已在 EvaluateWindowSignal 中通过 saveSignalAlertSimple + TryAutoTradeAndUpdate 完成
	// 这里重复触发会导致：1. 预警记录重复 2. 下单逻辑重复执行
}

// checkOpenPositionWithSignal 使用指定信号检查并开仓（事件驱动入口）
func (e *RobotEngine) checkOpenPositionWithSignal(ctx context.Context, signal *RobotSignal) {
	if signal == nil {
		return
	}
	e.Trader.CheckAndOpenPositionWithSignal(ctx, signal)
}

// checkOpenPosition 检查是否应该开仓
func (e *RobotEngine) checkOpenPosition(ctx context.Context) {
	if !e.orderLock.TryLock() {
		return
	}
	defer e.orderLock.Unlock()

	e.Trader.CheckAndOpenPosition(ctx)
}

// checkClosePosition 已删除 - 自动平仓逻辑已删除

// checkStopLossAndClose 检查止损进度并执行平仓
// 【重要修复】基于实时价格计算未实现盈亏，而不是使用可能过时2分钟的交易所API数据
func (e *RobotEngine) checkStopLossAndClose(ctx context.Context, currentPrice float64) {
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "[RobotEngine] checkStopLossAndClose panic recovered: robotId=%d, err=%v",
				e.Robot.Id, r)
		}
	}()

	if currentPrice <= 0 {
		return
	}

	// 【重要】检查自动平仓开关
	e.mu.RLock()
	robot := e.Robot
	strategyParams := e.CurrentStrategyParams
	positions := e.CurrentPositions // 使用引擎缓存的持仓（交易所实时数据）
	e.mu.RUnlock()

	if robot == nil || robot.AutoCloseEnabled != 1 {
		return // 自动平仓未开启，不执行止损
	}

	// 如果没有持仓，无需检查
	if len(positions) == 0 {
		return
	}

	// 【重要】从策略模板读取止损参数
	var stopLossPercent float64
	if strategyParams != nil {
		stopLossPercent = strategyParams.StopLossPercent
	} else {
		g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 止损检查跳过: CurrentStrategyParams为nil", robot.Id)
		return
	}
	if stopLossPercent <= 0 {
		return // 未设置止损百分比
	}

	// 【优化】直接使用交易所返回的持仓数据，不再查询数据库
	for _, pos := range positions {
		// 跳过无效持仓
		if math.Abs(pos.PositionAmt) < 0.0001 {
			continue
		}

		// 【重要修复】基于实时价格计算未实现盈亏，而不是使用 pos.UnrealizedPnl（可能过时2分钟）
		// 做多：(当前价格 - 开仓价格) * 数量
		// 做空：(开仓价格 - 当前价格) * 数量
		var realTimeUnrealizedPnl float64
		if pos.EntryPrice > 0 && math.Abs(pos.PositionAmt) > 0 {
			if pos.PositionSide == "LONG" {
				realTimeUnrealizedPnl = (currentPrice - pos.EntryPrice) * math.Abs(pos.PositionAmt)
			} else {
				realTimeUnrealizedPnl = (pos.EntryPrice - currentPrice) * math.Abs(pos.PositionAmt)
			}
		} else {
			// 如果没有开仓价格，降级使用交易所返回的值（可能过时）
			realTimeUnrealizedPnl = pos.UnrealizedPnl
		}

		// 只有亏损时才检查止损（realTimeUnrealizedPnl < 0）
		if realTimeUnrealizedPnl >= 0 {
			continue
		}

		// 计算保证金（从持仓数据计算）
		// 保证金 = |持仓数量| × 开仓价格 / 杠杆
		// 或直接使用交易所返回的 marginSize（如果有）
		margin := pos.Margin
		if margin <= 0 && pos.EntryPrice > 0 && robot.Leverage > 0 {
			margin = math.Abs(pos.PositionAmt) * pos.EntryPrice / float64(robot.Leverage)
		}
		if margin <= 0 {
			continue
		}

		// 计算止损进度
		// 止损金额 = 保证金 × 止损百分比
		// 止损进度 = |未实现盈亏| / 止损金额 × 100%
		stopLossAmount := margin * (stopLossPercent / 100.0)
		absUnrealizedPnl := math.Abs(realTimeUnrealizedPnl)
		progress := (absUnrealizedPnl / stopLossAmount) * 100.0

		// 如果止损进度达到100%，立即执行平仓
		if progress >= 100.0 {
			g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 止损进度达到100%%，立即执行平仓: positionSide=%s, progress=%.2f%%, unrealizedPnl=%.4f, margin=%.4f, stopLossPercent=%.2f%%",
				e.Robot.Id, pos.PositionSide, progress, pos.UnrealizedPnl, margin, stopLossPercent)

			// 执行平仓（使用交易所持仓数据）
			e.executeStopLossCloseByPosition(ctx, pos)
		}
	}
}

// saveCloseLog 保存平仓日志（止损、止盈、手动平仓通用）
// closeType: "stop_loss" 止损, "take_profit" 止盈, "manual" 手动
func (e *RobotEngine) saveCloseLog(ctx context.Context, closeType string, pos *exchange.Position, closeOrder *exchange.Order, errMsg string) {
	robot := e.Robot
	if robot == nil {
		return
	}

	// 查询关联的本地订单（用于获取 orderId）
	var localOrderId int64 = 0
	direction := "long"
	if pos.PositionSide == "SHORT" {
		direction = "short"
	}
	var localOrder struct {
		Id int64
	}
	err := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", robot.Id).
		Where("direction", direction).
		Where("status", OrderStatusOpen).
		Fields("id").
		Scan(&localOrder)
	if err == nil && localOrder.Id > 0 {
		localOrderId = localOrder.Id
	}

	// 确定事件类型和状态
	eventType := "close_" + closeType // close_stop_loss, close_take_profit, close_manual
	status := "success"
	message := ""
	if errMsg != "" {
		status = "failed"
		message = errMsg
	} else {
		switch closeType {
		case "stop_loss":
			message = fmt.Sprintf("止损平仓成功: %s方向, 数量%.6f, 盈亏%.4f USDT", pos.PositionSide, math.Abs(pos.PositionAmt), pos.UnrealizedPnl)
		case "take_profit":
			message = fmt.Sprintf("止盈平仓成功: %s方向, 数量%.6f, 盈亏%.4f USDT", pos.PositionSide, math.Abs(pos.PositionAmt), pos.UnrealizedPnl)
		case "manual":
			message = fmt.Sprintf("手动平仓成功: %s方向, 数量%.6f, 盈亏%.4f USDT", pos.PositionSide, math.Abs(pos.PositionAmt), pos.UnrealizedPnl)
		}
	}

	// 构建事件数据
	eventData := map[string]interface{}{
		"close_type":     closeType,
		"symbol":         pos.Symbol,
		"position_side":  pos.PositionSide,
		"quantity":       math.Abs(pos.PositionAmt),
		"entry_price":    pos.EntryPrice,
		"unrealized_pnl": pos.UnrealizedPnl,
		"margin":         pos.Margin,
	}
	if closeOrder != nil {
		eventData["exchange_order_id"] = closeOrder.OrderId
		eventData["avg_price"] = closeOrder.AvgPrice
		eventData["filled_qty"] = closeOrder.FilledQty
	}

	// 序列化事件数据为JSON
	eventDataJSON := "{}"
	if len(eventData) > 0 {
		data, jsonErr := json.Marshal(eventData)
		if jsonErr == nil {
			eventDataJSON = string(data)
		}
	}

	// 写入交易日志
	_, insertErr := g.DB().Model("hg_trading_execution_log").Ctx(ctx).Insert(g.Map{
		"signal_log_id": 0, // 平仓不关联预警记录
		"robot_id":      robot.Id,
		"order_id":      localOrderId,
		"event_type":    eventType,
		"event_data":    eventDataJSON,
		"status":        status,
		"message":       message,
		"created_at":    time.Now(),
	})
	if insertErr != nil {
		g.Log().Warningf(ctx, "[RobotEngine] 保存平仓日志失败: robotId=%d, closeType=%s, err=%v", robot.Id, closeType, insertErr)
	} else {
		g.Log().Debugf(ctx, "[RobotEngine] 平仓日志已保存: robotId=%d, closeType=%s, status=%s", robot.Id, closeType, status)
	}
}

// executeStopLossCloseByPosition 使用交易所持仓数据执行止损平仓
// executeStopLossCloseByPosition 执行止损平仓
// 【优化】直接使用传入的持仓数据，不再重复调用API
// 因为止损检查时已经使用了最新的 CurrentPositions
func (e *RobotEngine) executeStopLossCloseByPosition(ctx context.Context, pos *exchange.Position) {
	robot := e.Robot

	// 【优化】直接使用传入的持仓数据，不再调用 GetPositionsSmart
	// 理由：checkStopLossAndClose 已经使用了 CurrentPositions，数据是最新的
	quantity := math.Abs(pos.PositionAmt)
	if quantity <= 0 {
		g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 止损平仓跳过: 持仓数量为0或无效", robot.Id)
		return
	}

	g.Log().Infof(ctx, "[RobotEngine] robotId=%d 执行止损平仓: symbol=%s, positionSide=%s, quantity=%.6f, unrealizedPnl=%.4f",
		robot.Id, robot.Symbol, pos.PositionSide, quantity, pos.UnrealizedPnl)

	// 调用交易所API执行平仓
	closeOrder, err := e.Exchange.ClosePosition(ctx, robot.Symbol, pos.PositionSide, quantity)
	if err != nil {
		g.Log().Errorf(ctx, "[RobotEngine] robotId=%d 止损平仓失败: positionSide=%s, err=%v",
			robot.Id, pos.PositionSide, err)
		// 【新增】保存失败日志
		e.saveCloseLog(ctx, "stop_loss", pos, nil, err.Error())
		return
	}

	g.Log().Infof(ctx, "[RobotEngine] robotId=%d 止损平仓成功: positionSide=%s, exchangeOrderId=%s, unrealizedPnl=%.4f",
		robot.Id, pos.PositionSide, closeOrder.OrderId, pos.UnrealizedPnl)

	// 【新增】保存成功日志
	e.saveCloseLog(ctx, "stop_loss", pos, closeOrder, "")

	// 【重要】平仓成功后更新数据库订单状态（不需要调用API同步）
	e.updateOrderStatusAfterClose(ctx, pos, closeOrder, "stop_loss")

	// 【重要】平仓成功后清除 PositionTracker，为下一个新订单做准备
	e.ClearPositionTracker(pos.PositionSide)

	// 【优化】更新内存中的持仓数据（移除已平仓的持仓）
	e.removePositionFromCache(pos.PositionSide)

	// 【页面显示优化】平仓后顺便刷新余额缓存（账户权益/钱包余额），避免长期不更新
	e.refreshBalanceCacheAfterTrade(ctx, "after_stop_loss_close")
	g.Log().Infof(ctx, "[RobotEngine] robotId=%d 止损平仓完成: 已更新订单状态和内存缓存", robot.Id)
}

// calcStopLossProgress 计算止损进度（与前端显示的血条进度计算逻辑一致）
// 公式：
//
//	①、止损金额计算：止损金额 = 保证金 × (止损百分比 / 100%)
//	②、止损进度计算（血条进度）：止损进度 = |未实现盈亏| / 止损金额 × 100%
//	③、触发平仓条件：止损进度 ≥ 100%
func (e *RobotEngine) calcStopLossProgress(order *entity.TradingOrder, stopLossPercent float64) float64 {
	// 如果未实现盈亏 >= 0（盈利或持平），返回0（不显示进度）
	if order.UnrealizedProfit >= 0 {
		return 0
	}

	// 如果保证金 <= 0 或止损百分比 <= 0，返回0
	if order.Margin <= 0 || stopLossPercent <= 0 {
		return 0
	}

	// ①、止损金额计算：止损金额 = 保证金 × (止损百分比 / 100%)
	stopLossAmount := order.Margin * (stopLossPercent / 100.0)

	// ②、止损进度计算（血条进度）：止损进度 = |未实现盈亏| / 止损金额 × 100%
	absUnrealizedPnl := math.Abs(order.UnrealizedProfit)
	progress := (absUnrealizedPnl / stopLossAmount) * 100.0

	// 限制最大值为100%
	if progress > 100.0 {
		progress = 100.0
	}

	return progress
}

// executeStopLossClose 执行止损平仓
func (e *RobotEngine) executeStopLossClose(ctx context.Context, order *entity.TradingOrder) {
	robot := e.Robot

	// 确定持仓方向（LONG/SHORT）
	positionSide := "LONG"
	if order.Direction == "short" {
		positionSide = "SHORT"
	}

	// 获取持仓数量
	quantity := math.Abs(order.Quantity)
	if quantity <= 0 {
		g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 止损平仓失败: orderId=%d, quantity=%.6f (无效)",
			robot.Id, order.Id, quantity)
		return
	}

	g.Log().Infof(ctx, "[RobotEngine] robotId=%d 执行止损平仓: orderId=%d, symbol=%s, positionSide=%s, quantity=%.6f, unrealizedPnl=%.4f",
		robot.Id, order.Id, robot.Symbol, positionSide, quantity, order.UnrealizedProfit)

	// 调用交易所API执行平仓
	closeOrder, err := e.Exchange.ClosePosition(ctx, robot.Symbol, positionSide, quantity)
	if err != nil {
		g.Log().Errorf(ctx, "[RobotEngine] robotId=%d 止损平仓失败: orderId=%d, err=%v",
			robot.Id, order.Id, err)
		return
	}

	g.Log().Infof(ctx, "[RobotEngine] robotId=%d 止损平仓成功: orderId=%d, exchangeOrderId=%s, unrealizedPnl=%.4f",
		robot.Id, order.Id, closeOrder.OrderId, order.UnrealizedProfit)

	// 【优化】立即同步持仓数据和订单状态，确保状态一致性
	go func() {
		e.syncAccountDataIfNeeded(ctx, "after_trade")
		g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 止损平仓成功，等待下次自动同步: orderId=%d",
			robot.Id, order.Id)
	}()
}

// checkTakeProfitAndClose 检查启动止盈和止盈回撤并执行平仓
// 【重要修复】基于实时价格计算未实现盈亏，而不是使用可能过时2分钟的交易所API数据
// 【优化】完全使用内存数据（PositionTracker），不再依赖数据库
func (e *RobotEngine) checkTakeProfitAndClose(ctx context.Context, currentPrice float64) {
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "[RobotEngine] checkTakeProfitAndClose panic recovered: robotId=%d, err=%v",
				e.Robot.Id, r)
		}
	}()

	if currentPrice <= 0 {
		return
	}

	// 【重要】检查自动平仓开关
	e.mu.RLock()
	robot := e.Robot
	strategyParams := e.CurrentStrategyParams
	positions := e.CurrentPositions // 使用引擎缓存的持仓（交易所实时数据）
	e.mu.RUnlock()

	if robot == nil || robot.AutoCloseEnabled != 1 {
		return // 自动平仓未开启，不执行止盈回撤
	}

	// 如果没有持仓，无需检查
	if len(positions) == 0 {
		return
	}

	// 【重要】从策略模板读取止盈参数
	var autoStartPercent, profitRetreatPercent float64
	if strategyParams != nil {
		autoStartPercent = strategyParams.AutoStartRetreatPercent
		profitRetreatPercent = strategyParams.ProfitRetreatPercent
		g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 止盈参数(来自策略模板): autoStartPercent=%.2f%%, profitRetreatPercent=%.2f%%",
			robot.Id, autoStartPercent, profitRetreatPercent)
	} else {
		// 【调试】策略模板为空，输出详细信息帮助排查
		g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 【问题】CurrentStrategyParams为nil! 需要检查策略模板加载逻辑。MarketRiskMapping=%v",
			robot.Id, e.MarketRiskMapping)
		return
	}

	// 【优化】遍历交易所实时持仓
	for _, pos := range positions {
		// 跳过无效持仓
		if math.Abs(pos.PositionAmt) < 0.0001 {
			continue
		}

		// 【重要修复】基于实时价格计算未实现盈亏，而不是使用 pos.UnrealizedPnl（可能过时2分钟）
		// 做多：(当前价格 - 开仓价格) * 数量
		// 做空：(开仓价格 - 当前价格) * 数量
		var realTimeUnrealizedPnl float64
		if pos.EntryPrice > 0 && math.Abs(pos.PositionAmt) > 0 {
			if pos.PositionSide == "LONG" {
				realTimeUnrealizedPnl = (currentPrice - pos.EntryPrice) * math.Abs(pos.PositionAmt)
			} else {
				realTimeUnrealizedPnl = (pos.EntryPrice - currentPrice) * math.Abs(pos.PositionAmt)
			}
		} else {
			// 如果没有开仓价格，降级使用交易所返回的值（可能过时）
			realTimeUnrealizedPnl = pos.UnrealizedPnl
		}

		// 计算保证金
		margin := pos.Margin
		if margin <= 0 && pos.EntryPrice > 0 && robot.Leverage > 0 {
			margin = math.Abs(pos.PositionAmt) * pos.EntryPrice / float64(robot.Leverage)
		}
		if margin <= 0 {
			continue
		}

		// 【内存优化】获取或创建持仓跟踪器
		isNewTracker := false
		e.mu.Lock()
		tracker := e.PositionTrackers[pos.PositionSide]
		if tracker == nil {
			// 首次创建跟踪器
			tracker = &PositionTracker{
				PositionSide:      pos.PositionSide,
				EntryMargin:       margin,
				EntryTime:         time.Now(),
				HighestProfit:     0,
				TakeProfitEnabled: false,
			}
			e.PositionTrackers[pos.PositionSide] = tracker
			isNewTracker = true
			g.Log().Infof(ctx, "[RobotEngine] robotId=%d 创建持仓跟踪器: positionSide=%s", e.Robot.Id, pos.PositionSide)
		}
		e.mu.Unlock()

		// 【恢复机制】服务重启/状态丢失时：从数据库恢复止盈回撤启动状态与最高盈利
		if isNewTracker {
			e.initTrackerFromDB(ctx, pos.PositionSide, tracker)
		}

		// 【内存优化】更新最高盈利（只增不减）
		if realTimeUnrealizedPnl > tracker.HighestProfit {
			tracker.HighestProfit = realTimeUnrealizedPnl
		}

		// 【自动启动止盈回撤】检查是否满足启动条件
		// 条件：当前盈利百分比 = 未实现盈亏/保证金×100% >= 设定的启动止盈百分比
		if !tracker.TakeProfitEnabled && autoStartPercent > 0 && realTimeUnrealizedPnl > 0 {
			currentProfitPercent := (realTimeUnrealizedPnl / margin) * 100.0
			if currentProfitPercent >= autoStartPercent {
				tracker.TakeProfitEnabled = true
				// 【重要】启动时，当前盈利就是最高盈利的起点
				if realTimeUnrealizedPnl > tracker.HighestProfit {
					tracker.HighestProfit = realTimeUnrealizedPnl
				}
				// 【持久化】写入数据库，支持服务重启后继续止盈回撤（不可关闭原则）
				go e.persistProfitRetreatStarted(ctx, pos.PositionSide, tracker.HighestProfit)
				g.Log().Infof(ctx, "[RobotEngine] robotId=%d 【自动启动】止盈回撤已自动启动: positionSide=%s, currentProfitPercent=%.2f%% >= autoStartPercent=%.2f%%, highestProfit=%.4f",
					e.Robot.Id, pos.PositionSide, currentProfitPercent, autoStartPercent, tracker.HighestProfit)
			}
		}

		// 【内存优化】止盈回撤状态从内存获取
		isTakeProfitEnabled := tracker.TakeProfitEnabled

		// 如果止盈回撤未启动，跳过止盈回撤检查（继续下一个持仓）
		if !isTakeProfitEnabled {
			continue
		}

		// 【重要修复】如果刚手动启动止盈（HighestProfit很小），需要用当前盈亏初始化
		// 这是因为 SetTakeProfitEnabled 只能设置一个极小值，实际的最高盈利需要在这里设置
		if tracker.HighestProfit <= 0.001 && realTimeUnrealizedPnl > 0 {
			tracker.HighestProfit = realTimeUnrealizedPnl
			g.Log().Infof(ctx, "[RobotEngine] robotId=%d 初始化止盈最高盈利: positionSide=%s, highestProfit=%.4f",
				e.Robot.Id, pos.PositionSide, tracker.HighestProfit)
		}

		// 如果已启动止盈回撤，检查止盈回撤条件
		// 【修复】移除 tracker.HighestProfit > 0 条件，改为在内部处理
		if isTakeProfitEnabled && profitRetreatPercent > 0 {
			// 【调试日志】输出每次检查的详细数据
			g.Log().Infof(ctx, "[RobotEngine] robotId=%d 【止盈检查】positionSide=%s, takeProfitEnabled=%v, profitRetreatPercent=%.2f%%, highestProfit=%.4f, currentPnl=%.4f, entryPrice=%.4f, currentPrice=%.4f",
				e.Robot.Id, pos.PositionSide, isTakeProfitEnabled, profitRetreatPercent, tracker.HighestProfit, realTimeUnrealizedPnl, pos.EntryPrice, currentPrice)

			// 【修复】如果最高盈利还没有被正确初始化（小于等于0.001），跳过本次检查
			// 等待下次有正向盈利时初始化
			if tracker.HighestProfit <= 0.001 {
				g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 止盈回撤已启动但最高盈利未初始化（<=0.001），等待盈利出现: positionSide=%s, currentPnl=%.4f, highestProfit=%.4f",
					e.Robot.Id, pos.PositionSide, realTimeUnrealizedPnl, tracker.HighestProfit)
				continue
			}

			// 计算当前回撤百分比（使用实时盈亏）
			// 公式：(最高盈利 - 当前盈利) / 最高盈利 × 100%
			currentRetreatPercent := ((tracker.HighestProfit - realTimeUnrealizedPnl) / tracker.HighestProfit) * 100.0

			// 计算血条百分比（供调试用）
			bloodBarPercent := 100.0 - (currentRetreatPercent/profitRetreatPercent)*100.0
			if bloodBarPercent < 0 {
				bloodBarPercent = 0
			}
			if bloodBarPercent > 100 {
				bloodBarPercent = 100
			}

			g.Log().Infof(ctx, "[RobotEngine] robotId=%d 【止盈计算】回撤百分比=%.2f%%, 设定阈值=%.2f%%, 血条=%.2f%%, 是否触发=%v",
				e.Robot.Id, currentRetreatPercent, profitRetreatPercent, bloodBarPercent, currentRetreatPercent >= profitRetreatPercent)

			// 【安全检查】如果回撤百分比为负数（当前盈利超过最高盈利，应该更新最高盈利）
			if currentRetreatPercent < 0 {
				// 当前盈利超过了最高盈利，更新最高盈利
				tracker.HighestProfit = realTimeUnrealizedPnl
				g.Log().Infof(ctx, "[RobotEngine] robotId=%d 当前盈利超过最高盈利，更新: highestProfit=%.4f",
					e.Robot.Id, tracker.HighestProfit)
				continue
			}

			// 【关键】如果达到止盈回撤百分比，执行平仓
			// 血条 = 100% - (currentRetreatPercent / profitRetreatPercent) × 100%
			// 当血条为0%时，currentRetreatPercent >= profitRetreatPercent
			// 【修复】回撤百分比异常大（>200%）时，也应该触发止盈，而不是跳过
			// 这通常发生在从盈利大幅回撤到亏损的情况，更应该立即止盈
			if currentRetreatPercent >= profitRetreatPercent || currentRetreatPercent > 200 {
				if currentRetreatPercent > 200 {
					g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 【触发止盈-异常回撤】回撤百分比异常大: %.2f%%（当前盈亏=%.4f, 最高盈利=%.4f），立即执行平仓",
						e.Robot.Id, currentRetreatPercent, realTimeUnrealizedPnl, tracker.HighestProfit)
				} else {
					g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 【触发止盈】止盈回撤达到阈值，立即执行平仓: positionSide=%s, currentRetreatPercent=%.2f%%, profitRetreatPercent=%.2f%%, highestProfit=%.4f, realTimeUnrealizedPnl=%.4f",
						e.Robot.Id, pos.PositionSide, currentRetreatPercent, profitRetreatPercent, tracker.HighestProfit, realTimeUnrealizedPnl)
				}
				// 执行平仓（使用交易所持仓数据）
				e.executeTakeProfitCloseByPosition(ctx, pos, "take_profit")
			}
		} else {
			// 调试：输出为什么没有进入止盈检查
			if isTakeProfitEnabled && profitRetreatPercent <= 0 {
				g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 止盈检查跳过: profitRetreatPercent=%.2f%% (<=0)", e.Robot.Id, profitRetreatPercent)
			}
		}
	}
}

// executeTakeProfitCloseByPosition 执行止盈平仓
// 【优化】直接使用传入的持仓数据，不再重复调用API
func (e *RobotEngine) executeTakeProfitCloseByPosition(ctx context.Context, pos *exchange.Position, reason string) {
	robot := e.Robot

	// 【优化】直接使用传入的持仓数据，不再调用 GetPositionsSmart
	// 理由：checkTakeProfitAndClose 已经使用了 CurrentPositions，数据是最新的
	quantity := math.Abs(pos.PositionAmt)
	if quantity <= 0 {
		g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 止盈平仓跳过: 持仓数量为0或无效", robot.Id)
		return
	}

	// 【防重复平仓】先检查数据库中订单状态，如果已经是平仓中或已平仓，则跳过
	direction := "long"
	if pos.PositionSide == "SHORT" {
		direction = "short"
	}
	orderCount, err := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", robot.Id).
		Where("direction", direction).
		Where("status", OrderStatusOpen). // 只查询持仓中的订单
		Count()
	if err != nil || orderCount == 0 {
		// 重要：这里不再直接 return。
		// 真实场景里经常出现“交易所确实有持仓，但本地订单表没有 status=持仓中”的情况（比如历史数据/同步异常/重启丢状态）。
		// 如果此时直接跳过，会导致“血条到0%但不止盈”。
		// 处理策略：继续执行交易所平仓，同时输出 Warning 提示修复订单同步。
		g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 止盈触发但数据库无持仓中订单（direction=%s, orderCount=%d, err=%v）。将继续调用交易所平仓；同时建议检查订单状态同步逻辑。",
			robot.Id, direction, orderCount, err)
	}

	// 【防重复平仓】使用订单锁
	e.orderLock.Lock()
	defer e.orderLock.Unlock()

	g.Log().Infof(ctx, "[RobotEngine] robotId=%d 执行止盈平仓: symbol=%s, positionSide=%s, quantity=%.6f, unrealizedPnl=%.4f, reason=%s",
		robot.Id, robot.Symbol, pos.PositionSide, quantity, pos.UnrealizedPnl, reason)

	// 调用交易所API执行平仓
	closeOrder, closeErr := e.Exchange.ClosePosition(ctx, robot.Symbol, pos.PositionSide, quantity)
	if closeErr != nil {
		g.Log().Errorf(ctx, "[RobotEngine] robotId=%d 止盈平仓失败: positionSide=%s, err=%v",
			robot.Id, pos.PositionSide, closeErr)
		// 【新增】保存失败日志
		e.saveCloseLog(ctx, "take_profit", pos, nil, closeErr.Error())
		return
	}

	g.Log().Infof(ctx, "[RobotEngine] robotId=%d 止盈平仓成功: positionSide=%s, exchangeOrderId=%s, unrealizedPnl=%.4f, reason=%s",
		robot.Id, pos.PositionSide, closeOrder.OrderId, pos.UnrealizedPnl, reason)

	// 【新增】保存成功日志
	e.saveCloseLog(ctx, "take_profit", pos, closeOrder, "")

	// 【重要】平仓成功后更新数据库订单状态（不需要调用API同步）
	e.updateOrderStatusAfterClose(ctx, pos, closeOrder, reason)

	// 【重要】平仓成功后清除 PositionTracker，为下一个新订单做准备
	e.ClearPositionTracker(pos.PositionSide)

	// 【优化】更新内存中的持仓数据（移除已平仓的持仓）
	e.removePositionFromCache(pos.PositionSide)

	// 【页面显示优化】平仓后顺便刷新余额缓存（账户权益/钱包余额），避免长期不更新
	e.refreshBalanceCacheAfterTrade(ctx, "after_take_profit_close")
	g.Log().Infof(ctx, "[RobotEngine] robotId=%d 止盈平仓完成: 已更新订单状态和内存缓存", robot.Id)
}

// executeTakeProfitClose 执行止盈平仓
func (e *RobotEngine) executeTakeProfitClose(ctx context.Context, order *entity.TradingOrder, reason string) {
	robot := e.Robot

	// 确定持仓方向（LONG/SHORT）
	positionSide := "LONG"
	if order.Direction == "short" {
		positionSide = "SHORT"
	}

	// 获取持仓数量
	quantity := math.Abs(order.Quantity)
	if quantity <= 0 {
		g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 止盈平仓失败: orderId=%d, quantity=%.6f (无效)",
			robot.Id, order.Id, quantity)
		return
	}

	g.Log().Infof(ctx, "[RobotEngine] robotId=%d 执行止盈平仓: orderId=%d, symbol=%s, positionSide=%s, quantity=%.6f, unrealizedPnl=%.4f, reason=%s",
		robot.Id, order.Id, robot.Symbol, positionSide, quantity, order.UnrealizedProfit, reason)

	// 调用交易所API执行平仓
	closeOrder, err := e.Exchange.ClosePosition(ctx, robot.Symbol, positionSide, quantity)
	if err != nil {
		g.Log().Errorf(ctx, "[RobotEngine] robotId=%d 止盈平仓失败: orderId=%d, err=%v",
			robot.Id, order.Id, err)
		return
	}

	g.Log().Infof(ctx, "[RobotEngine] robotId=%d 止盈平仓成功: orderId=%d, exchangeOrderId=%s, unrealizedPnl=%.4f, reason=%s",
		robot.Id, order.Id, closeOrder.OrderId, order.UnrealizedProfit, reason)

	// 【优化】立即同步持仓数据和订单状态，确保状态一致性
	go func() {
		e.syncAccountDataIfNeeded(ctx, "after_trade")
		g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 止盈平仓成功，等待下次自动同步: orderId=%d",
			robot.Id, order.Id)
	}()
}

// updateOrdersUnrealizedPnl 基于实时价格更新订单未实现盈亏（事件驱动，轻量级）
// 【优化】价格更新时立即计算并更新未实现盈亏，不需要调用交易所API
func (e *RobotEngine) updateOrdersUnrealizedPnl(ctx context.Context, currentPrice float64) {
	if currentPrice <= 0 {
		return
	}

	// 查询所有持仓中的订单
	var orders []*entity.TradingOrder
	err := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", e.Robot.Id).
		Where("status", OrderStatusOpen). // 持仓中（使用统一的订单状态常量）
		Fields("id", "direction", "open_price", "quantity", "leverage", "margin", "unrealized_profit", "highest_profit").
		Scan(&orders)
	if err != nil || len(orders) == 0 {
		return
	}

	// 批量更新未实现盈亏
	updateBatch := make([]g.Map, 0)
	for _, order := range orders {
		if order.OpenPrice <= 0 || order.Quantity <= 0 {
			continue
		}

		// 计算未实现盈亏
		// 做多：(当前价格 - 开仓价格) * 数量
		// 做空：(开仓价格 - 当前价格) * 数量
		var unrealizedPnl float64
		if order.Direction == "long" {
			unrealizedPnl = (currentPrice - order.OpenPrice) * order.Quantity
		} else {
			unrealizedPnl = (order.OpenPrice - currentPrice) * order.Quantity
		}

		// 检查是否需要更新：未实现盈亏变化超过0.01 USDT
		pnlChanged := math.Abs(order.UnrealizedProfit-unrealizedPnl) >= 0.01
		if !pnlChanged {
			continue
		}

		// 更新最高盈利（只增不减）
		highestProfit := order.HighestProfit
		if unrealizedPnl > highestProfit {
			highestProfit = unrealizedPnl
		}

		updateBatch = append(updateBatch, g.Map{
			"id":                order.Id,
			"unrealized_profit": unrealizedPnl,
			"highest_profit":    highestProfit,
			"updated_at":        gtime.Now(),
		})
	}

	// 批量更新数据库
	if len(updateBatch) > 0 {
		for _, updateData := range updateBatch {
			orderId := updateData["id"].(int64)
			delete(updateData, "id")
			_, _ = dao.TradingOrder.Ctx(ctx).
				Where("id", orderId).
				Data(updateData).
				Update()
		}
		g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 【事件驱动】已更新 %d 个订单的未实现盈亏: currentPrice=%.4f", e.Robot.Id, len(updateBatch), currentPrice)
	}
}

// ==================== 窗口价格管理（toogo实时信号逻辑） ====================

// AddPricePoint 添加价格数据点
// 【效率优化】限制窗口大小，防止内存泄漏
func (e *RobotEngine) AddPricePoint(price float64) {
	e.priceLock.Lock()
	defer e.priceLock.Unlock()

	now := time.Now().UnixMilli()
	e.PriceWindow = append(e.PriceWindow, PricePoint{
		Timestamp: now,
		Price:     price,
	})

	// 【健壮性优化】限制窗口最大大小，防止内存无限增长
	const maxWindowSize = 10000 // 最大窗口大小（约10000个价格点）
	if len(e.PriceWindow) > maxWindowSize {
		// 保留最新的数据，删除最旧的数据（高效：直接切片）
		e.PriceWindow = e.PriceWindow[len(e.PriceWindow)-maxWindowSize:]
	}

	// 修剪窗口期外的数据
	e.pruneWindowData(now)
}

// pruneWindowData 修剪窗口期外的价格数据
// 【效率优化】使用二分查找优化修剪性能（O(log n) vs O(n)）
func (e *RobotEngine) pruneWindowData(now int64) {
	// 【优化】实时获取窗口值
	window := e.getRealTimeWindow()
	if window <= 0 {
		return
	}

	if len(e.PriceWindow) == 0 {
		return
	}

	cutoff := now - int64(window)*1000

	// 【效率优化】如果第一个元素还在窗口内，说明所有元素都在窗口内，无需修剪
	if e.PriceWindow[0].Timestamp >= cutoff {
		return
	}

	// 【效率优化】使用二分查找找到第一个需要保留的元素位置
	// 找到第一个 timestamp >= cutoff 的位置
	left, right := 0, len(e.PriceWindow)
	for left < right {
		mid := (left + right) / 2
		if e.PriceWindow[mid].Timestamp < cutoff {
			left = mid + 1
		} else {
			right = mid
		}
	}

	// 高效：直接切片，避免逐个复制
	if left < len(e.PriceWindow) {
		e.PriceWindow = e.PriceWindow[left:]
	} else {
		// 所有数据都已过期，清空窗口
		e.PriceWindow = e.PriceWindow[:0]
	}
}

// GetPriceWindow 获取价格窗口数据
func (e *RobotEngine) GetPriceWindow() []PricePoint {
	e.priceLock.RLock()
	defer e.priceLock.RUnlock()

	result := make([]PricePoint, len(e.PriceWindow))
	copy(result, e.PriceWindow)
	return result
}

// GetSignalHistory 获取信号历史
func (e *RobotEngine) GetSignalHistory() []SignalHistoryItem {
	e.priceLock.RLock()
	defer e.priceLock.RUnlock()

	result := make([]SignalHistoryItem, len(e.SignalHistory))
	copy(result, e.SignalHistory)
	return result
}

// ClearPriceWindow 清空价格窗口（平仓后调用）
func (e *RobotEngine) ClearPriceWindow() {
	e.priceLock.Lock()
	defer e.priceLock.Unlock()

	e.PriceWindow = make([]PricePoint, 0, 1000)
	e.LastAlertedLong = nil
	e.LastAlertedShort = nil
	e.LastWindowMin = nil
	e.LastWindowMax = nil
	g.Log().Debugf(context.Background(), "[RobotEngine] robotId=%d 清空价格窗口数据", e.Robot.Id)
}

// UpdateMonitorConfig 已废弃 - 现在使用实时获取窗口和波动值
// 【已废弃】此函数已废弃，窗口和波动值现在实时从策略模板获取
func (e *RobotEngine) UpdateMonitorConfig(window int, threshold float64) {
	// 已废弃：不再更新 MonitorConfig，改为实时获取
	g.Log().Debugf(context.Background(), "[RobotEngine] robotId=%d UpdateMonitorConfig已废弃，窗口和波动值现在实时获取", e.Robot.Id)
}

// getRealTimeWindowAndThreshold 实时获取窗口和波动值（复用 monitor.go 中的逻辑）
// 步骤1：获取全局实时市场状态
// 步骤2：根据创建机器人时提交的映射关系选择风险偏好
// 步骤3：根据实时市场状态+风险偏好获取策略组中对应的策略内的时间窗口和波动值
func (e *RobotEngine) getRealTimeWindowAndThreshold() (window int, threshold float64) {
	ctx := context.Background()

	// 【步骤1】获取全局实时市场状态
	globalAnalysis := market.GetMarketAnalyzer().GetAnalysis(e.Platform, e.Robot.Symbol)
	if globalAnalysis == nil || globalAnalysis.MarketState == "" {
		g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 全局市场分析器未返回市场状态数据", e.Robot.Id)
		return 0, 0
	}
	marketState := normalizeMarketState(string(globalAnalysis.MarketState))
	if marketState == "" {
		g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 市场状态为空", e.Robot.Id)
		return 0, 0
	}

	// 【步骤2】根据创建机器人时提交的映射关系选择风险偏好
	// 【重要】使用引擎已加载的映射关系（从 remark 字段加载）
	e.mu.RLock()
	riskPref := e.MarketRiskMapping[marketState]
	e.mu.RUnlock()

	if riskPref == "" {
		g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 市场状态=%s 在映射关系中未找到对应的风险偏好（映射关系: %v）", e.Robot.Id, marketState, e.MarketRiskMapping)
		return 0, 0
	}

	// 【步骤3】根据实时市场状态+风险偏好获取策略组中对应的策略
	groupId := e.Robot.StrategyGroupId
	if groupId == 0 && e.Robot.CurrentStrategy != "" {
		var strategyData map[string]interface{}
		if err := json.Unmarshal([]byte(e.Robot.CurrentStrategy), &strategyData); err == nil {
			if gid, ok := strategyData["groupId"].(float64); ok {
				groupId = int64(gid)
			}
		}
	}
	if groupId == 0 {
		g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 未绑定策略组ID", e.Robot.Id)
		return 0, 0
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
		g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 找不到策略模板: groupId=%d, marketState=%s, riskPreference=%s",
			e.Robot.Id, groupId, marketState, riskPref)
		return 0, 0
	}

	// 返回窗口和波动值
	window = strategy.MonitorWindow
	threshold = strategy.VolatilityThreshold

	g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 实时获取窗口和波动值: market=%s, risk=%s, window=%ds, threshold=%.2f",
		e.Robot.Id, marketState, riskPref, window, threshold)

	return window, threshold
}

// getRealTimeWindow 实时获取窗口值（用于修剪窗口数据）
func (e *RobotEngine) getRealTimeWindow() int {
	window, _ := e.getRealTimeWindowAndThreshold()
	return window
}

// GetPriceLock 获取价格数据锁（供外部只读访问）
func (e *RobotEngine) GetPriceLock() *sync.RWMutex {
	return &e.priceLock
}

// GetWindowStats 获取窗口统计数据
func (e *RobotEngine) GetWindowStats() (minPrice, maxPrice, currentPrice float64, dataCount int) {
	e.priceLock.RLock()
	defer e.priceLock.RUnlock()

	dataCount = len(e.PriceWindow)
	if dataCount == 0 {
		return 0, 0, 0, 0
	}

	prices := make([]float64, dataCount)
	for i, p := range e.PriceWindow {
		prices[i] = p.Price
	}

	minPrice = prices[0]
	maxPrice = prices[0]
	for _, p := range prices {
		if p < minPrice {
			minPrice = p
		}
		if p > maxPrice {
			maxPrice = p
		}
	}
	currentPrice = prices[dataCount-1]
	return
}

// EvaluateWindowSignal 评估窗口信号（简化版 - 纯窗口逻辑）
// 【架构优化】预警只负责检测信号和保存预警记录，不负责下单逻辑
//
// 核心方向判断算法（简化）：
// 在时间窗口内维护最高价(maxPrice)和最低价(minPrice)
// - 做空信号：最高价 - 实时价格 >= 波动值 (价格从高点回落)
// - 做多信号：实时价格 - 最低价 >= 波动值 (价格从低点反弹)
//
// 优化：移除MACD/EMA权重，只使用纯窗口信号，逻辑更清晰
// 下单逻辑由其他模块负责（如定时任务扫描预警记录并下单）
func (e *RobotEngine) EvaluateWindowSignal() *RobotSignal {
	e.priceLock.Lock()
	defer e.priceLock.Unlock()

	// 【优化】实时获取窗口和波动值
	window, threshold := e.getRealTimeWindowAndThreshold()
	if window <= 0 || threshold <= 0 {
		return &RobotSignal{
			Timestamp:  time.Now(),
			Direction:  "NEUTRAL",
			SignalType: "window",
			Reason:     "无法获取窗口或波动值配置",
		}
	}

	signal := &RobotSignal{
		Timestamp:       time.Now(),
		SignalType:      "window",
		SignalThreshold: threshold,
	}

	dataCount := len(e.PriceWindow)
	if dataCount == 0 {
		signal.Direction = "NEUTRAL"
		signal.Reason = "等待价格数据..."
		return signal
	}

	if dataCount == 1 {
		signal.Direction = "NEUTRAL"
		signal.CurrentPrice = e.PriceWindow[0].Price
		signal.Reason = fmt.Sprintf("已获取初始价格 %.2f，等待更多数据...", signal.CurrentPrice)
		return signal
	}

	// ============ 计算窗口内最高价和最低价 ============
	minPrice := e.PriceWindow[0].Price
	maxPrice := e.PriceWindow[0].Price
	for _, p := range e.PriceWindow {
		if p.Price < minPrice {
			minPrice = p.Price
		}
		if p.Price > maxPrice {
			maxPrice = p.Price
		}
	}

	currentPrice := e.PriceWindow[dataCount-1].Price

	// ============ 核心方向判断（简化：纯窗口逻辑） ============
	// 计算距离
	distanceFromMax := maxPrice - currentPrice // 最高价 - 实时价格
	distanceFromMin := currentPrice - minPrice // 实时价格 - 最低价

	// 触发条件（简化版：只看窗口价格）
	// 做空：最高价 - 实时价格 >= 波动值
	shortTriggered := distanceFromMax >= threshold
	// 做多：实时价格 - 最低价 >= 波动值
	longTriggered := distanceFromMin >= threshold

	// 填充信号基础数据
	signal.WindowMinPrice = minPrice
	signal.WindowMaxPrice = maxPrice
	signal.CurrentPrice = currentPrice
	signal.DistanceFromMin = distanceFromMin
	signal.DistanceFromMax = distanceFromMax

	// 检测窗口基准价变化，重置预警标记
	if e.LastWindowMin != nil && minPrice != *e.LastWindowMin {
		e.LastAlertedLong = nil
	}
	if e.LastWindowMax != nil && maxPrice != *e.LastWindowMax {
		e.LastAlertedShort = nil
	}
	e.LastWindowMin = &minPrice
	e.LastWindowMax = &maxPrice

	// ============ 判断最终方向（简化版） ============
	var newSignal string = "neutral"
	// shouldAlert 用于控制是否更新预警基准价（避免重复预警）
	// 注意：现在不再在此处保存预警记录，只在条件检查时保存
	_ = false // 占位，避免未使用变量警告

	// 双向同时触发时（价格剧烈波动，窗口范围 >= 2*阈值），重置预警继续监控
	if longTriggered && shortTriggered {
		e.LastAlertedLong = nil
		e.LastAlertedShort = nil
		signal.Direction = "NEUTRAL"
		signal.Reason = fmt.Sprintf("窗口双向触发 | 高%.2f 实时%.2f 低%.2f | 振幅%.2f≥2×阈值%.0f，继续监控",
			maxPrice, currentPrice, minPrice, maxPrice-minPrice, threshold)
		signal.SignalProgress = 0
		e.LastWindowSignal = "neutral"
		return signal
	}

	// 单向触发判断（更新预警基准价）
	if longTriggered {
		newSignal = "long"
		if e.LastAlertedLong == nil || math.Abs(minPrice-*e.LastAlertedLong) > 0.0001 {
			e.LastAlertedLong = &minPrice
		}
	} else if shortTriggered {
		newSignal = "short"
		if e.LastAlertedShort == nil || math.Abs(maxPrice-*e.LastAlertedShort) > 0.0001 {
			e.LastAlertedShort = &maxPrice
		}
	}

	// 设置信号结果
	switch newSignal {
	case "long":
		signal.Direction = "LONG"
		signal.Strength = 100 // 简化版：触发即100%
		signal.Confidence = 100
		signal.Action = "OPEN_LONG"
		signal.Reason = fmt.Sprintf("📈 做多信号 | 实时%.2f - 低%.2f = %.2f ≥ 阈值%.0f",
			currentPrice, minPrice, distanceFromMin, threshold)
	case "short":
		signal.Direction = "SHORT"
		signal.Strength = 100 // 简化版：触发即100%
		signal.Confidence = 100
		signal.Action = "OPEN_SHORT"
		signal.Reason = fmt.Sprintf("📉 做空信号 | 高%.2f - 实时%.2f = %.2f ≥ 阈值%.0f",
			maxPrice, currentPrice, distanceFromMax, threshold)
	default:
		signal.Direction = "NEUTRAL"
		signal.Strength = 0
		signal.Confidence = 0
		signal.Action = "HOLD"
		// 计算进度（距离触发条件的百分比）
		longProgress := (distanceFromMin / threshold) * 100
		shortProgress := (distanceFromMax / threshold) * 100
		if longProgress > shortProgress {
			signal.SignalProgress = math.Min(100, longProgress)
		} else {
			signal.SignalProgress = math.Min(100, shortProgress)
		}
		signal.Reason = fmt.Sprintf("监控中 | 高%.2f 实时%.2f 低%.2f | 做多%.0f%% 做空%.0f%%",
			maxPrice, currentPrice, minPrice, longProgress, shortProgress)
	}

	// 记录信号历史
	e.SignalHistory = append(e.SignalHistory, SignalHistoryItem{
		Timestamp: time.Now().UnixMilli(),
		Signal:    newSignal,
	})
	if len(e.SignalHistory) > 100 {
		e.SignalHistory = e.SignalHistory[len(e.SignalHistory)-100:]
	}

	// 【优化】每个信号只保存一次预警记录
	// 只在信号方向变化时保存，信号持续存在时不重复保存
	if newSignal != "neutral" {
		signal.AlignedTimeframes = 1

		// 复制信号副本
		signalCopy := *signal

		// 信号方向变化时记录日志
		isNewDirection := newSignal != e.LastWindowSignal

		logCtx := context.Background()
		g.Log().Infof(logCtx, "[RobotEngine] robotId=%d 检测到信号: newSignal=%s, lastSignal=%s, isNew=%v",
			e.Robot.Id, newSignal, e.LastWindowSignal, isNewDirection)

		// 【优化】每个信号只保存一次：只在信号方向变化时保存
		if isNewDirection {
			g.Log().Infof(logCtx, "[RobotEngine] robotId=%d 新方向信号: %s, price=%.2f",
				e.Robot.Id, newSignal, currentPrice)
			e.LastWindowSignal = newSignal

			// 【关键修复】在触发下单前，先检查内存缓存是否已有该方向的持仓
			// 解决 LONG → NEUTRAL → LONG 导致的重复下单问题
			positionSide := "LONG"
			if newSignal == "short" {
				positionSide = "SHORT"
			}
			if e.HasActivePosition(positionSide) {
				g.Log().Infof(logCtx, "[RobotEngine] robotId=%d 内存缓存已有%s方向持仓，跳过下单",
					e.Robot.Id, positionSide)
				// 已有持仓，不触发下单（不需要保存预警记录，因为不会下单）
			} else {
				// 无持仓，保存预警记录并触发下单
				logId := e.saveSignalAlertSimple(&signalCopy)
				if logId > 0 {
					g.Log().Infof(logCtx, "[RobotEngine] 预警记录已保存: robotId=%d, logId=%d, direction=%s, action=%s",
						e.Robot.Id, logId, signalCopy.Direction, signalCopy.Action)

					// 触发自动下单（异步执行）
					go func() {
						ctx := context.Background()
						e.Trader.TryAutoTradeAndUpdate(ctx, &signalCopy, logId)
					}()
				} else {
					g.Log().Warningf(logCtx, "[RobotEngine] 保存预警记录失败: robotId=%d, direction=%s, action=%s",
						e.Robot.Id, signalCopy.Direction, signalCopy.Action)
				}
			}
		}
		// 信号方向不变时，不保存预警记录（每个信号只保存一次）
	} else {
		// 信号变为 neutral 时，重置状态
		e.LastWindowSignal = "neutral"
	}

	return signal
}

// saveSignalAlertSimple 简化版：保存信号预警记录（只记录信号，不记录执行结果）
// 【优化】预警日志只记录方向信号，执行结果记录在交易日志中
func (e *RobotEngine) saveSignalAlertSimple(signal *RobotSignal) int64 {
	ctx := context.Background()

	// 【重要】始终保存预警记录，不管自动下单是否开启
	// 预警日志只记录信号本身，不记录执行结果
	e.mu.RLock()
	robot := e.Robot
	e.mu.RUnlock()

	if robot == nil {
		g.Log().Errorf(ctx, "[RobotEngine] robot为nil，无法保存预警记录")
		return 0
	}

	// 获取市场状态（规范化）
	marketState := ""
	if e.LastAnalysis != nil {
		marketState = normalizeMarketState(e.LastAnalysis.MarketState)
	}

	// 优化 reason：添加更多详细信息
	reason := signal.Reason
	if signal.Direction == "LONG" {
		// 做多信号：添加价格区间和触发距离
		distance := signal.CurrentPrice - signal.WindowMinPrice
		reason = fmt.Sprintf("📈 做多信号 | 当前价: %.2f | 窗口最低: %.2f | 距离: %.2f | 阈值: %.0f | 触发条件: 当前价-最低价≥阈值",
			signal.CurrentPrice, signal.WindowMinPrice, distance, signal.SignalThreshold)
	} else if signal.Direction == "SHORT" {
		// 做空信号：添加价格区间和触发距离
		distance := signal.WindowMaxPrice - signal.CurrentPrice
		reason = fmt.Sprintf("📉 做空信号 | 当前价: %.2f | 窗口最高: %.2f | 距离: %.2f | 阈值: %.0f | 触发条件: 最高价-当前价≥阈值",
			signal.CurrentPrice, signal.WindowMaxPrice, distance, signal.SignalThreshold)
	}

	// 写入数据库（只保存信号信息，不保存执行结果）
	// 【PostgreSQL 兼容】使用 InsertAndGetId() 而不是 Insert() + LastInsertId()
	data := g.Map{
		"robot_id":         robot.Id,
		"strategy_id":      0,
		"symbol":           robot.Symbol,
		"signal_type":      signal.Direction,
		"signal_source":    "window_weighted",
		"signal_strength":  signal.Strength,
		"current_price":    signal.CurrentPrice,
		"window_min_price": signal.WindowMinPrice,
		"window_max_price": signal.WindowMaxPrice,
		"threshold":        signal.SignalThreshold,
		"market_state":     marketState,
		"risk_preference":  "", // 【已移除】不再使用 Robot.RiskPreference，统一从映射关系获取
		"target_price":     0,
		"stop_loss":        0,
		"take_profit":      0,
		"executed":         0,
		"execute_result":   "", // 【优化】不再保存执行结果，执行结果记录在交易日志中
		"is_processed":     0,  // 【新增】已读标识：0=未处理，1=已处理（用于防止重复下单）
		"reason":           reason,
		"indicators":       "{}",
	}

	// 【PostgreSQL 兼容】直接使用事务 + LASTVAL()，避免尝试失败
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		g.Log().Errorf(ctx, "[RobotEngine] 开启事务失败: %v", err)
		return 0
	}
	defer tx.Rollback()

	_, err = tx.Model("hg_trading_signal_log").Data(data).Insert()
	if err != nil {
		g.Log().Errorf(ctx, "[RobotEngine] 插入信号日志失败: %v", err)
		return 0
	}

	v, err := tx.GetValue("SELECT LASTVAL()")
	if err != nil {
		g.Log().Errorf(ctx, "[RobotEngine] 获取信号日志ID失败: %v", err)
		return 0
	}
	logId := v.Int64()

	err = tx.Commit()
	if err != nil {
		g.Log().Errorf(ctx, "[RobotEngine] 提交事务失败: %v", err)
		return 0
	}

	g.Log().Infof(ctx, "[RobotEngine] ✅ 预警记录已保存: robotId=%d, logId=%d, direction=%s",
		robot.Id, logId, signal.Direction)
	return logId
}

// saveSignalAlert 保存信号预警记录（每个新信号必须保存）
// 先检查条件，然后保存带结果的预警记录，返回记录ID
// 【保留兼容性，但简化流程中不再使用】
func (e *RobotEngine) saveSignalAlert(signal *RobotSignal, isNewDirection bool) int64 {
	ctx := context.Background()

	g.Log().Infof(ctx, "[RobotEngine] saveSignalAlert调用: robotId=%d, direction=%s, isNewDirection=%v",
		e.Robot.Id, signal.Direction, isNewDirection)

	// 如果不是新方向，检查是否有未处理的记录（避免重复处理）
	// 改进：只检查未处理的记录（executed=0），如果上一个信号已经处理完成，允许新信号
	if !isNewDirection {
		count, _ := g.DB().Model("hg_trading_signal_log").Ctx(ctx).
			Where("robot_id", e.Robot.Id).
			Where("signal_type", signal.Direction).
			Where("executed", 0). // 只检查未处理的记录
			Count()
		if count > 0 {
			g.Log().Debugf(ctx, "[RobotEngine] 有未处理的同方向记录，等待处理完成: robotId=%d", e.Robot.Id)
			return 0 // 有未处理的记录，跳过
		}
	}

	// 获取市场状态（规范化）
	marketState := ""
	if e.LastAnalysis != nil {
		marketState = normalizeMarketState(e.LastAnalysis.MarketState)
	}

	// 先检查条件，确定执行结果
	executeResult := e.checkSignalConditions(ctx, signal)

	// 写入数据库
	// 【PostgreSQL 兼容】使用 InsertAndGetId() 而不是 Insert() + LastInsertId()
	data := g.Map{
		"robot_id":         e.Robot.Id,
		"strategy_id":      0,
		"symbol":           e.Robot.Symbol,
		"signal_type":      signal.Direction,
		"signal_source":    "window_weighted",
		"signal_strength":  signal.Strength,
		"current_price":    signal.CurrentPrice,
		"window_min_price": signal.WindowMinPrice,
		"window_max_price": signal.WindowMaxPrice,
		"threshold":        signal.SignalThreshold,
		"market_state":     marketState,
		"risk_preference":  "", // 【已移除】不再使用 Robot.RiskPreference，统一从映射关系获取
		"target_price":     0,
		"stop_loss":        0,
		"take_profit":      0,
		"executed":         0,
		"execute_result":   executeResult,
		"reason":           signal.Reason,
		"indicators":       "{}",
	}

	// 【PostgreSQL 兼容】直接使用事务 + LASTVAL()，避免尝试失败
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		g.Log().Errorf(ctx, "[RobotEngine] 开启事务失败: %v", err)
		return 0
	}
	defer tx.Rollback()

	_, err = tx.Model("hg_trading_signal_log").Data(data).Insert()
	if err != nil {
		g.Log().Errorf(ctx, "[RobotEngine] 插入信号日志失败: %v", err)
		return 0
	}

	v, err := tx.GetValue("SELECT LASTVAL()")
	if err != nil {
		g.Log().Errorf(ctx, "[RobotEngine] 获取信号日志ID失败: %v", err)
		return 0
	}
	logId := v.Int64()

	err = tx.Commit()
	if err != nil {
		g.Log().Errorf(ctx, "[RobotEngine] 提交事务失败: %v", err)
		return 0
	}

	g.Log().Infof(ctx, "[RobotEngine] ✅ 预警记录已保存: robotId=%d, logId=%d, direction=%s, result=%s",
		e.Robot.Id, logId, signal.Direction, executeResult)
	return logId
}

// checkSignalConditions 检查信号条件，返回执行结果描述
// 【重新设计】简化检查逻辑：只检查自动交易开关和一个方向只能有一单
func (e *RobotEngine) checkSignalConditions(ctx context.Context, signal *RobotSignal) string {
	// 【步骤1】信号生成 → 立即触发检查（自动交易开关、一个方向只能有一单）
	e.mu.RLock()
	robot := e.Robot
	e.mu.RUnlock()

	// 检查自动交易开关
	if robot == nil {
		return "机器人不存在"
	}
	autoTradeEnabled := robot.AutoTradeEnabled
	if autoTradeEnabled != 1 {
		return "自动下单未开启"
	}

	// 检查信号操作
	if signal.Action != "OPEN_LONG" && signal.Action != "OPEN_SHORT" {
		return fmt.Sprintf("信号类型为%s，不是开仓信号", signal.Action)
	}

	// 检查一个方向只能有一单
	positionSide := "LONG"
	if signal.Direction == "SHORT" {
		positionSide = "SHORT"
	}
	dbDirection := "long"
	if positionSide == "SHORT" {
		dbDirection = "short"
	}
	count, err := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", e.Robot.Id).
		Where("direction", dbDirection).
		Where("status", OrderStatusOpen).
		Count()
	if err == nil && count > 0 {
		directionText := "多头"
		if positionSide == "SHORT" {
			directionText = "空头"
		}
		return fmt.Sprintf("%s方向已有持仓", directionText)
	}

	// 所有条件满足，准备下单
	return "准备下单"
}

// saveSignalAlertDeprecated 保存信号预警记录（已废弃）
func (e *RobotEngine) saveSignalAlertDeprecated(signal *RobotSignal, direction string) {
	ctx := context.Background()

	// 获取市场状态（规范化）
	marketState := ""
	if e.LastAnalysis != nil {
		marketState = normalizeMarketState(e.LastAnalysis.MarketState)
	}

	// 检查自动交易是否开启
	executeResult := "自动下单未开启"
	if e.Robot.AutoTradeEnabled == 1 {
		executeResult = "等待条件检查"
	}

	// 写入数据库
	_, err := g.DB().Model("hg_trading_signal_log").Ctx(ctx).Insert(g.Map{
		"robot_id":         e.Robot.Id,
		"strategy_id":      0,
		"symbol":           e.Robot.Symbol,
		"signal_type":      direction,
		"signal_source":    "window_weighted",
		"signal_strength":  signal.Strength,
		"current_price":    signal.CurrentPrice,
		"window_min_price": signal.WindowMinPrice,
		"window_max_price": signal.WindowMaxPrice,
		"threshold":        signal.SignalThreshold,
		"market_state":     marketState,
		"risk_preference":  "", // 【已移除】不再使用 Robot.RiskPreference，统一从映射关系获取
		"target_price":     0,
		"stop_loss":        0,
		"take_profit":      0,
		"executed":         0,
		"execute_result":   executeResult,
		"reason":           signal.Reason,
		"indicators":       "{}",
	})
	if err != nil {
		g.Log().Warningf(ctx, "[RobotEngine] 保存信号预警失败: robotId=%d, err=%v", e.Robot.Id, err)
	} else {
		g.Log().Infof(ctx, "[RobotEngine] 信号预警: robotId=%d, direction=%s, price=%.2f",
			e.Robot.Id, direction, signal.CurrentPrice)
	}
}

// saveSignalLog 保存有价值的方向信号日志到数据库
func (e *RobotEngine) saveSignalLog(signal *RobotSignal, direction string, longScore, shortScore float64, reasons []string) {
	ctx := context.Background()

	// 获取市场状态和风险偏好（规范化）
	marketState := ""
	if e.LastAnalysis != nil {
		marketState = normalizeMarketState(e.LastAnalysis.MarketState)
	}

	// 从映射关系获取风险偏好
	e.mu.RLock()
	riskPreference := e.MarketRiskMapping[marketState]
	e.mu.RUnlock()

	// 【严格模式】从映射关系获取风险偏好，找不到时记录警告（用于日志记录，允许为空）
	if riskPreference == "" {
		g.Log().Warningf(ctx, "[RobotEngine] robotId=%d 市场状态=%s 在映射关系中未找到对应的风险偏好，无法记录风险偏好信息", e.Robot.Id, marketState)
		riskPreference = "" // 保持为空，不降级
	}

	// 构建指标详情
	indicators := map[string]interface{}{
		"longScore":      longScore,
		"shortScore":     shortScore,
		"reasons":        reasons,
		"windowMin":      signal.WindowMinPrice,
		"windowMax":      signal.WindowMaxPrice,
		"currentPrice":   signal.CurrentPrice,
		"threshold":      signal.SignalThreshold,
		"marketState":    marketState,
		"riskPreference": riskPreference,
	}
	indicatorsJson, _ := json.Marshal(indicators)

	// 检查自动交易是否开启，提前确定执行结果
	executeResult := "自动下单未开启"
	if e.Robot.AutoTradeEnabled == 1 {
		executeResult = "等待下单条件检查"
	}

	// 写入数据库
	_, err := g.DB().Model("hg_trading_signal_log").Ctx(ctx).Insert(g.Map{
		"robot_id":         e.Robot.Id,
		"strategy_id":      0, // 策略ID从 CurrentStrategy JSON 中获取，此处简化为0
		"symbol":           e.Robot.Symbol,
		"signal_type":      direction, // long/short
		"signal_source":    "window_weighted",
		"signal_strength":  signal.Strength,
		"current_price":    signal.CurrentPrice,
		"window_min_price": signal.WindowMinPrice,
		"window_max_price": signal.WindowMaxPrice,
		"threshold":        signal.SignalThreshold,
		"market_state":     marketState,
		"risk_preference":  riskPreference,
		"target_price":     0, // 不使用目标价
		"stop_loss":        0,
		"take_profit":      0,
		"executed":         0,
		"execute_result":   executeResult,
		"reason":           signal.Reason,
		"indicators":       string(indicatorsJson),
	})
	if err != nil {
		g.Log().Warningf(ctx, "[RobotEngine] 保存信号日志失败: robotId=%d, err=%v", e.Robot.Id, err)
	} else {
		g.Log().Infof(ctx, "[RobotEngine] 保存方向信号: robotId=%d, direction=%s, strength=%.2f",
			e.Robot.Id, direction, signal.Strength)
	}
}

// getIndicatorSignals 获取MACD和EMA指标信号
// 返回: macdLong, macdShort, emaLong, emaShort
func (e *RobotEngine) getIndicatorSignals() (bool, bool, bool, bool) {
	analysis := e.LastAnalysis
	if analysis == nil || analysis.TimeframeScores == nil {
		return false, false, false, false
	}

	var macdLongCount, macdShortCount int
	var emaLongCount, emaShortCount int
	var totalCount int

	// 统计各周期的MACD和EMA信号
	for _, score := range analysis.TimeframeScores {
		totalCount++

		// MACD信号：MACD > 0 看多，MACD < 0 看空
		if score.MACD > 0 {
			macdLongCount++
		} else if score.MACD < 0 {
			macdShortCount++
		}

		// EMA信号：EMA12 > EMA26 看多，EMA12 < EMA26 看空
		if score.EMA12 > score.EMA26 {
			emaLongCount++
		} else if score.EMA12 < score.EMA26 {
			emaShortCount++
		}
	}

	if totalCount == 0 {
		return false, false, false, false
	}

	// 需要超过半数才确认方向
	halfCount := totalCount / 2

	macdLong := macdLongCount > halfCount
	macdShort := macdShortCount > halfCount
	emaLong := emaLongCount > halfCount
	emaShort := emaShortCount > halfCount

	return macdLong, macdShort, emaLong, emaShort
}

// ==================== 状态获取 ====================

// GetStatus 获取引擎状态
func (e *RobotEngine) GetStatus() *RobotEngineStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()

	status := &RobotEngineStatus{
		RobotId:   e.Robot.Id,
		Symbol:    e.Robot.Symbol,
		Platform:  e.Platform,
		Running:   e.running,
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
		status.MarketState = normalizeMarketState(e.LastAnalysis.MarketState)
		status.TrendDirection = e.LastAnalysis.TrendDirection
		status.Volatility = e.LastAnalysis.Volatility
	}

	// 【已移除】风险偏好不再从 Robot.RiskPreference 获取，统一从映射关系获取
	// status.RiskPreference 字段保留为空或从映射关系获取（如果需要显示）
	status.RiskPreference = ""

	if e.LastSignal != nil {
		status.SignalDirection = e.LastSignal.Direction
		status.SignalStrength = e.LastSignal.Strength
		status.SignalConfidence = e.LastSignal.Confidence
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

	// 策略配置（时间窗口和波动值）
	// 【优化】实时获取窗口和波动值
	window, threshold := e.getRealTimeWindowAndThreshold()
	status.StrategyWindow = window
	status.StrategyThreshold = threshold

	// 【优化】从全局市场分析器获取当前市场状态
	currentState := ""
	globalAnalysis := market.GetMarketAnalyzer().GetAnalysis(e.Platform, e.Robot.Symbol)
	if globalAnalysis != nil {
		currentState = normalizeMarketState(string(globalAnalysis.MarketState))
	}
	currentState = normalizeMarketState(currentState) // 确保规范化
	status.CurrentMarketState = currentState

	// 【严格模式】从映射关系获取当前风险偏好，找不到时保持为空（用于状态显示）
	currentRiskPref := ""
	if currentState != "" {
		e.mu.RLock()
		currentRiskPref = e.MarketRiskMapping[currentState]
		e.mu.RUnlock()
		if currentRiskPref == "" {
			g.Log().Warningf(context.Background(), "[RobotEngine] robotId=%d 市场状态=%s 在映射关系中未找到对应的风险偏好", e.Robot.Id, currentState)
		}
	}
	status.CurrentRiskPref = currentRiskPref

	// 价格窗口数据（用于实时图表）
	e.priceLock.Lock()
	if len(e.PriceWindow) > 0 {
		status.PriceWindowData = make([]PriceWindowPoint, len(e.PriceWindow))
		minPrice := e.PriceWindow[0].Price
		maxPrice := e.PriceWindow[0].Price
		for i, p := range e.PriceWindow {
			status.PriceWindowData[i] = PriceWindowPoint{
				Timestamp: p.Timestamp,
				Price:     p.Price,
			}
			if p.Price < minPrice {
				minPrice = p.Price
			}
			if p.Price > maxPrice {
				maxPrice = p.Price
			}
		}
		status.WindowMinPrice = minPrice
		status.WindowMaxPrice = maxPrice
		status.WindowCurrentPrice = e.PriceWindow[len(e.PriceWindow)-1].Price

		// 计算触发价格
		// 【优化】实时获取波动值
		_, threshold := e.getRealTimeWindowAndThreshold()
		if threshold > 0 {
			status.LongTriggerPrice = minPrice + threshold
			status.ShortTriggerPrice = maxPrice - threshold
		}
	}
	e.priceLock.Unlock()

	// 信号详情
	if e.LastSignal != nil {
		status.SignalProgress = e.LastSignal.SignalProgress
		status.SignalReason = e.LastSignal.Reason
	}

	return status
}

// HealthCheck 健康检查（轻量级，高效）
// 【健壮性优化】检查引擎是否正常运行
func (e *RobotEngine) HealthCheck() error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.running {
		return gerror.New("引擎未运行")
	}

	// 检查最后更新时间（如果超过30秒未更新，可能异常）
	now := time.Now()
	if e.LastAnalysisUpdate.IsZero() {
		// 刚启动，还未执行分析，正常
		return nil
	}

	// 如果超过30秒未更新分析结果，可能异常
	if now.Sub(e.LastAnalysisUpdate) > 30*time.Second {
		return gerror.Newf("市场分析超时: 最后更新=%v", e.LastAnalysisUpdate)
	}

	return nil
}

// RobotEngineStatus 机器人引擎状态
type RobotEngineStatus struct {
	RobotId  int64  `json:"robotId"`
	Symbol   string `json:"symbol"`
	Platform string `json:"platform"`
	Running  bool   `json:"running"`

	// 连接状态
	Connected bool    `json:"connected"`
	LastPrice float64 `json:"lastPrice"`

	// 账户
	TotalBalance float64 `json:"totalBalance"`
	AvailBalance float64 `json:"availBalance"`

	// 市场分析
	MarketState    string  `json:"marketState"`
	TrendDirection string  `json:"trendDirection"`
	Volatility     float64 `json:"volatility"`

	// 风险评估
	WinProbability float64 `json:"winProbability"`
	RiskPreference string  `json:"riskPreference"`

	// 方向信号
	SignalDirection  string  `json:"signalDirection"`
	SignalStrength   float64 `json:"signalStrength"`
	SignalConfidence float64 `json:"signalConfidence"`

	// 持仓
	HasPosition   bool    `json:"hasPosition"`
	PositionSide  string  `json:"positionSide"`
	PositionAmt   float64 `json:"positionAmt"`
	EntryPrice    float64 `json:"entryPrice"`
	UnrealizedPnl float64 `json:"unrealizedPnl"`

	// 价格窗口数据（用于实时图表）
	PriceWindowData    []PriceWindowPoint `json:"priceWindowData"`
	WindowMinPrice     float64            `json:"windowMinPrice"`
	WindowMaxPrice     float64            `json:"windowMaxPrice"`
	WindowCurrentPrice float64            `json:"windowCurrentPrice"`
	LongTriggerPrice   float64            `json:"longTriggerPrice"`
	ShortTriggerPrice  float64            `json:"shortTriggerPrice"`
	SignalProgress     float64            `json:"signalProgress"`
	SignalReason       string             `json:"signalReason"`

	// 策略配置（时间窗口和波动值）
	StrategyWindow     int     `json:"strategyWindow"`     // 时间窗口（秒）
	StrategyThreshold  float64 `json:"strategyThreshold"`  // 波动值（USDT）
	CurrentMarketState string  `json:"currentMarketState"` // 当前市场状态
	CurrentRiskPref    string  `json:"currentRiskPref"`    // 当前风险偏好
}

// PriceWindowPoint 价格窗口数据点（用于图表）
type PriceWindowPoint struct {
	Timestamp int64   `json:"timestamp"`
	Price     float64 `json:"price"`
}

// ==================== 辅助方法 ====================

// HasActivePosition 检查是否有活跃持仓
func (e *RobotEngine) HasActivePosition(side string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, pos := range e.CurrentPositions {
		if pos.PositionAmt != 0 && pos.PositionSide == side {
			return true
		}
	}
	return false
}

// GetPosition 获取指定方向持仓
func (e *RobotEngine) GetPosition(side string) *exchange.Position {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, pos := range e.CurrentPositions {
		if pos.PositionSide == side && pos.PositionAmt != 0 {
			return pos
		}
	}
	return nil
}

// ==================== 市场分析模块 ====================

// RobotAnalyzer 机器人市场分析器
type RobotAnalyzer struct {
	engine *RobotEngine
}

// NewRobotAnalyzer 创建分析器
func NewRobotAnalyzer(engine *RobotEngine) *RobotAnalyzer {
	return &RobotAnalyzer{engine: engine}
}

// Analyze 执行市场分析（精简版：只分析3个核心周期）
func (a *RobotAnalyzer) Analyze(ctx context.Context) *RobotMarketAnalysis {
	klines := a.engine.LastKlines
	if klines == nil {
		return nil
	}

	analysis := &RobotMarketAnalysis{
		Timestamp:       time.Now(),
		TimeframeScores: make(map[string]*TimeframeScore),
		Indicators:      &TechnicalIndicators{},
	}

	// 【短线优化】分析5个周期：1m/5m/15m/1h/1d（短线需要多周期综合分析）
	timeframes := map[string][]*exchange.Kline{
		"1m":  klines.Klines1m,  // 1分钟周期（短期波动）
		"5m":  klines.Klines5m,  // 5分钟周期（短期趋势）
		"15m": klines.Klines15m, // 15分钟周期（中期趋势）
		"1h":  klines.Klines1h,  // 1小时周期（中期趋势）
		"1d":  klines.Klines1d,  // 1天周期（长期趋势，参考）
	}

	for tf, data := range timeframes {
		// 【短线优化】根据权重优化最小K线要求：高权重周期增加，低权重周期减少
		minKlines := 26
		switch tf {
		case "1m":
			minKlines = 8 // 1m周期：8根（权重30%，高权重，约8分钟数据）
		case "5m":
			minKlines = 20 // 5m周期：20根（权重40%，最高权重，约100分钟数据）
		case "15m":
			minKlines = 12 // 15m周期：12根（权重20%，中权重，约3小时数据）
		case "1h":
			minKlines = 5 // 1h周期：5根（权重3%，最低权重，约5小时数据）
		case "1d":
			minKlines = 5 // 1d周期：5根（长期参考，约5天数据）
		}
		if len(data) < minKlines {
			continue
		}
		score := a.analyzeTimeframe(data)
		score.Timeframe = tf
		analysis.TimeframeScores[tf] = score
	}

	// 计算综合指标
	a.calculateOverallIndicators(analysis)

	// 判断市场状态
	a.determineMarketState(analysis)

	return analysis
}

// analyzeTimeframe 分析单周期（精简版：只用MACD判断趋势）
func (a *RobotAnalyzer) analyzeTimeframe(klines []*exchange.Kline) *TimeframeScore {
	score := &TimeframeScore{
		KlinesCount: len(klines),
	}

	if len(klines) < 26 {
		return score
	}

	// 计算收盘价序列
	closes := make([]float64, len(klines))
	highs := make([]float64, len(klines))
	lows := make([]float64, len(klines))
	for i, k := range klines {
		closes[i] = k.Close
		highs[i] = k.High
		lows[i] = k.Low
	}

	// 只计算EMA和MACD
	score.EMA12 = a.calculateEMA(closes, 12)
	score.EMA26 = a.calculateEMA(closes, 26)
	score.MACD = score.EMA12 - score.EMA26

	// 计算趋势强度
	score.TrendStrength = a.calculateTrendStrength(klines)

	// 计算波动率
	score.Volatility = a.calculateTimeframeVolatility(klines)

	// 【短线优化】判断方向和强度（适度敏感，平衡实时性和稳定性）
	// 短线操作需要及时响应，但不过度敏感
	if score.EMA12 > score.EMA26 && score.MACD > 0 {
		score.Direction = "up"
		// 【短线优化】适度敏感：从50+50改为45+55（比超短线保守，比长线敏感）
		score.Strength = math.Min(100, 45+score.TrendStrength*55)
	} else if score.EMA12 < score.EMA26 && score.MACD < 0 {
		score.Direction = "down"
		score.Strength = math.Min(100, 45+score.TrendStrength*55)
	} else {
		score.Direction = "neutral"
		score.Strength = 30
	}

	// 判断该周期的市场状态
	score.MarketState = a.determineTimeframeMarketState(score.TrendStrength, score.Volatility)

	return score
}

// calculateTrendStrength 计算趋势强度（参考toogo算法）
func (a *RobotAnalyzer) calculateTrendStrength(klines []*exchange.Kline) float64 {
	if len(klines) < 10 {
		return 0
	}

	// 使用线性回归计算趋势斜率
	n := len(klines)
	var sumX, sumY, sumXY, sumX2 float64
	for i, k := range klines {
		x := float64(i)
		y := k.Close
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// 计算斜率
	denominator := float64(n)*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0
	}
	slope := (float64(n)*sumXY - sumX*sumY) / denominator

	// 归一化斜率到0-1范围
	avgPrice := sumY / float64(n)
	if avgPrice == 0 {
		return 0
	}
	normalizedSlope := math.Abs(slope) / avgPrice * 100

	// 限制在0-1范围内
	return math.Min(1, normalizedSlope)
}

// calculateTimeframeVolatility 计算单周期波动率
func (a *RobotAnalyzer) calculateTimeframeVolatility(klines []*exchange.Kline) float64 {
	if len(klines) < 10 {
		return 1.0
	}

	// 计算ATR
	var atr float64
	for i := 1; i < len(klines); i++ {
		high := klines[i].High
		low := klines[i].Low
		prevClose := klines[i-1].Close

		tr := math.Max(high-low, math.Max(math.Abs(high-prevClose), math.Abs(low-prevClose)))
		atr += tr
	}
	atr /= float64(len(klines) - 1)

	// 相对波动率（ATR/当前价格 * 100）
	lastPrice := klines[len(klines)-1].Close
	if lastPrice > 0 {
		return (atr / lastPrice) * 100
	}
	return 1.0
}

// calculateBaselineVolatility 计算基准波动率（过去N天的平均波动率）
// 【方案1优化】支持多周期K线数据，实时计算，无需等待100根K线
func (a *RobotAnalyzer) calculateBaselineVolatility(klines []*exchange.Kline, days int) float64 {
	if len(klines) < 10 {
		return 1.0 // 默认值
	}

	// 【优化】降低最小K线要求，支持实时计算
	// 至少需要20根K线即可计算基准（之前是100根）
	minKlines := 20
	if len(klines) < minKlines {
		// 数据不足，使用当前波动率作为基准
		return a.calculateTimeframeVolatility(klines)
	}

	// 计算过去N天的每日波动率
	// 简化：使用滑动窗口计算平均波动率
	windowSize := len(klines)
	if windowSize > 200 {
		windowSize = 200 // 最多使用200根K线
	}

	startIdx := len(klines) - windowSize
	dailyVolatilities := []float64{}

	// 将K线数据分成多个窗口，每个窗口计算一个波动率
	// 【优化】根据K线数量动态调整窗口数量
	windowCount := 5 // 最少5个窗口
	if len(klines) >= 50 {
		windowCount = 10 // 数据充足时使用10个窗口
	}

	windowLength := windowSize / windowCount
	if windowLength < 5 {
		windowLength = 5 // 每个窗口至少5根K线
	}

	// 确保 windowLength 不超过剩余K线数量
	if windowLength > len(klines)-startIdx {
		windowLength = len(klines) - startIdx
	}

	for i := startIdx; i < len(klines); i += windowLength {
		endIdx := i + windowLength
		if endIdx > len(klines) {
			endIdx = len(klines)
		}
		if endIdx-i < 5 {
			break
		}

		windowKlines := klines[i:endIdx]
		vol := a.calculateTimeframeVolatility(windowKlines)
		if vol > 0 {
			dailyVolatilities = append(dailyVolatilities, vol)
		}
	}

	if len(dailyVolatilities) == 0 {
		// 如果无法计算，使用当前波动率
		return a.calculateTimeframeVolatility(klines)
	}

	// 计算平均值（基准波动率）
	sum := 0.0
	for _, v := range dailyVolatilities {
		sum += v
	}
	baseline := sum / float64(len(dailyVolatilities))

	// 确保基准值在合理范围内（0.1% - 10%）
	if baseline < 0.1 {
		baseline = 0.1
	} else if baseline > 10.0 {
		baseline = 10.0
	}

	return baseline
}

// calculateBaselineVolatilityMultiTimeframe 使用多周期K线计算基准波动率（【方案1优化】）
// 优先级：1h > 15m > 5m > 1m > 1d
// 选择数据量最充足的周期，确保实时计算
func (a *RobotAnalyzer) calculateBaselineVolatilityMultiTimeframe() float64 {
	klines := a.engine.LastKlines
	if klines == nil {
		return 1.0 // 默认值
	}

	// 【短线优化】按权重优先级尝试多个周期，选择数据量最充足的
	// 优先级：5m > 1m > 15m > 1h（根据权重从高到低）
	timeframes := []struct {
		name   string
		klines []*exchange.Kline
		minLen int
	}{
		{"5m", klines.Klines5m, 25},   // 5分钟：至少25根（权重40%，最高权重，约125分钟数据）
		{"1m", klines.Klines1m, 12},   // 1分钟：至少12根（权重30%，高权重，约12分钟数据）
		{"15m", klines.Klines15m, 15}, // 15分钟：至少15根（权重20%，中权重，约3.75小时数据）
		{"1h", klines.Klines1h, 5},    // 1小时：至少5根（权重3%，最低权重，约5小时数据）
	}

	// 优先选择数据量最充足的周期
	bestKlines := []*exchange.Kline{}
	bestTimeframe := ""
	maxLen := 0

	for _, tf := range timeframes {
		if len(tf.klines) >= tf.minLen && len(tf.klines) > maxLen {
			maxLen = len(tf.klines)
			bestKlines = tf.klines
			bestTimeframe = tf.name
		}
	}

	// 如果找到合适的周期，使用它计算基准波动率
	if len(bestKlines) > 0 {
		baseline := a.calculateBaselineVolatility(bestKlines, 30)
		g.Log().Debugf(context.Background(), "[RobotAnalyzer] 使用%s周期计算基准波动率: %.2f%%, K线数量=%d",
			bestTimeframe, baseline, len(bestKlines))
		return baseline
	}

	// 如果所有周期数据都不足，尝试使用任意有数据的周期
	for _, tf := range timeframes {
		if len(tf.klines) >= 10 {
			baseline := a.calculateBaselineVolatility(tf.klines, 30)
			g.Log().Debugf(context.Background(), "[RobotAnalyzer] 降级使用%s周期计算基准波动率: %.2f%%, K线数量=%d",
				tf.name, baseline, len(tf.klines))
			return baseline
		}
	}

	// 如果完全没有数据，返回默认值
	return 1.0
}

// determineTimeframeMarketState 判断单周期市场状态（【方案1】使用自适应波动率阈值）
func (a *RobotAnalyzer) determineTimeframeMarketState(trendStrength, volatility float64) string {
	// 【方案1优化】使用多周期K线数据计算基准波动率，实时计算，无需等待100根K线
	// 优先级：1h > 15m > 5m > 1m > 1d
	var baselineVol float64

	baselineVol = a.calculateBaselineVolatilityMultiTimeframe()

	// 如果多周期计算失败，降级到配置的阈值
	if baselineVol <= 0 || baselineVol > 10.0 {
		a.engine.mu.RLock()
		config := a.engine.VolatilityConfig
		a.engine.mu.RUnlock()

		if config != nil {
			// 使用配置的阈值作为基准（如果配置了货币对特定值）
			baselineVol = (config.HighVolatilityThreshold + config.LowVolatilityThreshold) / 2
		} else {
			baselineVol = (highVolatilityThreshold + lowVolatilityThreshold) / 2
		}
	}

	// 【方案1】动态阈值计算：基于基准波动率
	// 高波动阈值 = 基准的1.5倍，低波动阈值 = 基准的0.5倍
	highThreshold := baselineVol * 1.5
	lowThreshold := baselineVol * 0.5

	// 确保阈值在合理范围内
	if highThreshold < 0.5 {
		highThreshold = 0.5
	}
	if lowThreshold < 0.1 {
		lowThreshold = 0.1
	}
	if highThreshold > 10.0 {
		highThreshold = 10.0
	}

	// 【短线优化】优先判断趋势市场：使用适度的阈值（平衡敏感性和稳定性）
	// 注意：这里使用固定阈值，周期特定的阈值调整在analyzeTimeframe中通过强度计算实现
	effectiveThreshold := trendStrengthThreshold

	if trendStrength > effectiveThreshold && volatility >= lowThreshold && volatility <= highThreshold*1.5 {
		return "trend"
	}

	// 非趋势市场，根据波动率判断
	if volatility >= highThreshold {
		return "high_vol"
	} else if volatility <= lowThreshold {
		return "low_vol"
	}
	return "volatile" // 注意：统一使用 "volatile" 而不是 "range"
}

// calculateEMA 计算EMA
func (a *RobotAnalyzer) calculateEMA(data []float64, period int) float64 {
	if len(data) < period {
		return 0
	}

	multiplier := 2.0 / float64(period+1)
	ema := data[0]

	for i := 1; i < len(data); i++ {
		ema = (data[i]-ema)*multiplier + ema
	}

	return ema
}

// calculateOverallIndicators 计算综合指标（精简版）
func (a *RobotAnalyzer) calculateOverallIndicators(analysis *RobotMarketAnalysis) {
	var weightedTrendSum, totalWeight float64
	var avgVolatility float64
	var volatilityCount int

	for tf, score := range analysis.TimeframeScores {
		weight := timeframeWeights[tf]
		if weight == 0 {
			weight = 0.2
		}
		totalWeight += weight

		// 加权趋势评分（简化：只看方向和强度）
		if score.Direction == "up" {
			weightedTrendSum += score.Strength * weight
		} else if score.Direction == "down" {
			weightedTrendSum -= score.Strength * weight
		}

		// 累计波动率
		if score.Volatility > 0 {
			avgVolatility += score.Volatility
			volatilityCount++
		}
	}

	if totalWeight > 0 {
		analysis.Indicators.TrendScore = weightedTrendSum / totalWeight
	}

	// 计算平均波动率
	if volatilityCount > 0 {
		analysis.Volatility = avgVolatility / float64(volatilityCount)
	} else if klines := a.engine.LastKlines; klines != nil && len(klines.Klines5m) > 0 {
		analysis.Volatility = a.calculateTimeframeVolatility(klines.Klines5m)
	}

	analysis.Indicators.VolatilityScore = math.Min(100, analysis.Volatility*20)
}

// determineMarketState 综合判断市场状态（【方案2】使用加权投票机制）
func (a *RobotAnalyzer) determineMarketState(analysis *RobotMarketAnalysis) {
	total := len(analysis.TimeframeScores)
	if total == 0 {
		analysis.MarketState = "volatile"
		analysis.VolatilityLevel = "normal"
		return
	}

	// 【短线优化】加权投票机制：大幅提高短期周期权重，降低长期周期权重，更灵敏
	// 短线需要及时响应，优先关注短期波动和趋势
	weights := map[string]float64{
		"1m":  0.30, // 超短期：30%（捕捉短期波动，提高10%）
		"5m":  0.40, // 短期：40%（短期趋势，最重要，提高15%）
		"15m": 0.20, // 中期：20%（中期趋势，重要，降低5%）
		"1h":  0.08, // 中期：8%（中期趋势，稳定，降低12%）
		"1d":  0.02, // 长期：2%（长期趋势，参考，降低8%）
	}

	// 【方案2】使用加权投票统计各状态得分
	stateScores := map[string]float64{
		"trend":    0,
		"high_vol": 0,
		"low_vol":  0,
		"volatile": 0,
	}

	var upScore, downScore float64

	for tf, score := range analysis.TimeframeScores {
		weight := weights[tf]
		if weight == 0 {
			weight = 1.0 / float64(total) // 如果周期不在权重表中，平均分配
		}

		// 统计方向（加权）
		if score.Direction == "up" {
			upScore += weight
		} else if score.Direction == "down" {
			downScore += weight
		}

		// 【方案2】加权投票：各状态得分累加
		stateScores[score.MarketState] += weight
	}

	// 判断趋势方向（基于加权得分）
	if upScore >= 0.4 {
		analysis.TrendDirection = "up"
		analysis.TrendStrength = upScore * 100
	} else if downScore >= 0.4 {
		analysis.TrendDirection = "down"
		analysis.TrendStrength = downScore * 100
	} else {
		analysis.TrendDirection = "neutral"
		analysis.TrendStrength = math.Max(upScore, downScore) * 100
	}

	// 【方案2】找到得分最高的市场状态（需要达到最小阈值0.4）
	maxScore := 0.0
	finalState := "volatile"
	minThreshold := 0.4 // 最小阈值：需要至少40%的权重支持

	for state, score := range stateScores {
		if score > maxScore {
			maxScore = score
			finalState = state
		}
	}

	// 如果最高得分未达到阈值，使用默认状态
	if maxScore < minThreshold {
		finalState = "volatile"
		maxScore = 0.6 // 默认置信度
	}

	analysis.MarketState = finalState
	analysis.MarketStateConf = maxScore

	// 设置波动等级
	switch finalState {
	case "high_vol":
		analysis.VolatilityLevel = "high"
	case "low_vol":
		analysis.VolatilityLevel = "low"
	case "trend":
		analysis.VolatilityLevel = "normal"
	default:
		analysis.VolatilityLevel = "normal"
	}
}

// ==================== 信号生成模块 ====================

// RobotSignalGen 机器人信号生成器
type RobotSignalGen struct {
	engine *RobotEngine
}

// NewRobotSignalGen 创建信号生成器
func NewRobotSignalGen(engine *RobotEngine) *RobotSignalGen {
	return &RobotSignalGen{engine: engine}
}

// Generate 生成方向信号（简化版 - 纯窗口逻辑）
// 核心逻辑：只使用窗口信号，移除技术分析干扰
func (s *RobotSignalGen) Generate(ctx context.Context) *RobotSignal {
	// 直接评估窗口价格信号
	windowSignal := s.engine.EvaluateWindowSignal()
	if windowSignal == nil {
		return &RobotSignal{
			Timestamp:  time.Now(),
			Direction:  "NEUTRAL",
			SignalType: "none",
			Reason:     "等待数据...",
		}
	}

	// 直接返回窗口信号，不再做技术分析确认
	return windowSignal
}

// ==================== 交易执行模块 ====================

// RobotTrader 机器人交易执行器
type RobotTrader struct {
	engine *RobotEngine
}

// NewRobotTrader 创建交易执行器
func NewRobotTrader(engine *RobotEngine) *RobotTrader {
	return &RobotTrader{engine: engine}
}

// TryAutoTradeAndUpdate 尝试自动下单并更新预警记录
// 【重新设计】简化流程：
// 1. 信号生成 → 立即触发检查（自动交易开关、一个方向只能有一单）
// 2. 参数计算 → 获取市场状态和策略参数（与机器人详情页面相同的方法）
// 3. 创建订单到平台（完成）
func (t *RobotTrader) TryAutoTradeAndUpdate(ctx context.Context, signal *RobotSignal, logId int64) {
	robot := t.engine.Robot
	if robot == nil {
		if logId > 0 {
			t.saveExecutionLog(ctx, logId, 0, "order_failed", "failed", "机器人不存在", map[string]interface{}{
				"step": "robot_check",
			})
		}
		return
	}

	// 检查信号有效性
	if signal == nil {
		if logId > 0 {
			t.saveExecutionLog(ctx, logId, 0, "order_failed", "failed", "信号为空", map[string]interface{}{
				"step": "signal_check",
			})
		}
		return
	}

	if signal.Direction == "NEUTRAL" {
		if logId > 0 {
			t.saveExecutionLog(ctx, logId, 0, "order_failed", "failed", "信号为中性，不执行下单", map[string]interface{}{
				"step":      "signal_check",
				"direction": signal.Direction,
			})
		}
		return
	}

	// 只处理开仓信号
	if signal.Action != "OPEN_LONG" && signal.Action != "OPEN_SHORT" {
		if logId > 0 {
			t.saveExecutionLog(ctx, logId, 0, "order_failed", "failed", fmt.Sprintf("信号类型为%s，不是开仓信号", signal.Action), map[string]interface{}{
				"step":   "signal_check",
				"action": signal.Action,
			})
		}
		return
	}

	// 【重要】使用原子操作标记预警记录为已处理（防止并发重复下单）
	// 使用数据库的原子更新操作：UPDATE ... SET is_processed=1 WHERE id=? AND is_processed=0
	// 如果更新影响的行数为0，说明已经被其他goroutine处理了，直接返回
	if logId > 0 {
		result, err := g.DB().Model("hg_trading_signal_log").Ctx(ctx).
			Where("id", logId).
			Where("(is_processed = 0 OR is_processed IS NULL)"). // 只更新未处理的记录
			Update(g.Map{
				"is_processed": 1, // 标记为已处理
			})

		if err != nil {
			g.Log().Errorf(ctx, "[RobotTrader] robotId=%d 更新预警记录logId=%d的已读标识失败: %v", robot.Id, logId, err)
			// 即使更新失败，也继续执行（避免因数据库问题导致无法下单）
		} else {
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected == 0 {
				// 影响行数为0，说明已经被其他goroutine处理了
				g.Log().Infof(ctx, "[RobotTrader] robotId=%d 预警记录logId=%d已被其他goroutine处理（is_processed=1），跳过重复下单", robot.Id, logId)
				return
			}
			g.Log().Infof(ctx, "[RobotTrader] robotId=%d 预警记录logId=%d已标记为已处理（is_processed=1），开始执行下单", robot.Id, logId)
		}
	}

	// 检查自动交易开关
	t.engine.mu.RLock()
	autoTradeEnabled := robot.AutoTradeEnabled
	t.engine.mu.RUnlock()

	if autoTradeEnabled != 1 {
		if logId > 0 {
			t.saveExecutionLog(ctx, logId, 0, "order_failed", "failed", "自动下单未开启", map[string]interface{}{
				"step":             "auto_trade_check",
				"autoTradeEnabled": autoTradeEnabled,
			})
		}
		return
	}

	// 【优化】检查一个方向只能有一单
	positionSide := "LONG"
	if signal.Direction == "SHORT" {
		positionSide = "SHORT"
	}

	// 检查交易所实时持仓
	if t.engine.Exchange == nil {
		if logId > 0 {
			t.saveExecutionLog(ctx, logId, 0, "order_failed", "failed", "交易所实例不存在，无法检查持仓", map[string]interface{}{
				"step":      "position_check",
				"direction": positionSide,
			})
		}
		return
	}

	// 【优化】使用智能缓存获取持仓，1秒内的缓存视为有效（减少API调用）
	positions, err := t.engine.GetPositionsSmart(ctx, 1*time.Second)
	if err != nil {
		if logId > 0 {
			t.saveExecutionLog(ctx, logId, 0, "order_failed", "failed", fmt.Sprintf("获取持仓失败: %v", err), map[string]interface{}{
				"step":      "position_check",
				"direction": positionSide,
				"error":     err.Error(),
			})
		}
		return
	}

	// 更新缓存（供其他模块使用）
	t.engine.mu.Lock()
	t.engine.CurrentPositions = positions
	t.engine.LastPositionUpdate = time.Now()
	t.engine.mu.Unlock()

	// 开仓限制规则（按你的最新定义）：
	// - DualSidePosition=1（双向开单开启）：允许同时持有多+空，但【同方向只能一单】（不允许加仓）
	// - DualSidePosition=0（关闭）：【持仓内只能有一单】（不区分多空）
	//
	// 这里基于“交易所实时持仓”做判断，避免依赖DB/内存不同步。
	hasAnyPosition := false
	hasSameSidePosition := false
	existingPositionSide := ""
	for _, pos := range positions {
		if pos == nil || math.Abs(pos.PositionAmt) <= positionAmtEpsilon {
			continue
		}
		hasAnyPosition = true
		if existingPositionSide == "" {
			existingPositionSide = pos.PositionSide
		}
		if pos.PositionSide == positionSide {
			hasSameSidePosition = true
		}
	}

	// 单向：任意方向已有持仓都拒绝新开仓
	if robot.DualSidePosition == 0 && hasAnyPosition {
		existingDirectionText := "多头"
		if existingPositionSide == "SHORT" {
			existingDirectionText = "空头"
		}
		targetDirectionText := "多头"
		if positionSide == "SHORT" {
			targetDirectionText = "空头"
		}
		reason := fmt.Sprintf("单向持仓模式：已有持仓（%s），持仓内只能有一单，拒绝新开仓（目标=%s）", existingDirectionText, targetDirectionText)
		if logId > 0 {
			t.saveExecutionLog(ctx, logId, 0, "order_failed", "failed", reason, map[string]interface{}{
				"step":                 "single_position_check",
				"dualSidePosition":     robot.DualSidePosition,
				"existingPositionSide": existingPositionSide,
				"targetPositionSide":   positionSide,
				"source":               "exchange_realtime",
			})
		}
		g.Log().Infof(ctx, "[RobotTrader] robotId=%d %s，跳过下单", robot.Id, reason)
		return
	}

	// 双向：同方向已有持仓则拒绝（不允许加仓）；反方向可开（允许多空同持）
	if robot.DualSidePosition == 1 && hasSameSidePosition {
		directionText := "多头"
		if positionSide == "SHORT" {
			directionText = "空头"
		}
		reason := fmt.Sprintf("双向持仓模式：%s方向已有持仓，同方向只能一单（禁止加仓），拒绝新开仓", directionText)
		if logId > 0 {
			t.saveExecutionLog(ctx, logId, 0, "order_failed", "failed", reason, map[string]interface{}{
				"step":             "dual_side_same_direction_check",
				"dualSidePosition": robot.DualSidePosition,
				"positionSide":     positionSide,
				"source":           "exchange_realtime",
			})
		}
		g.Log().Infof(ctx, "[RobotTrader] robotId=%d %s，跳过下单", robot.Id, reason)
		return
	}

	// 获取锁并执行下单
	locked := false
	for i := 0; i < 5; i++ {
		if t.engine.orderLock.TryLock() {
			locked = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !locked {
		if logId > 0 {
			t.saveExecutionLog(ctx, logId, 0, "order_failed", "failed", "系统繁忙，获取锁超时", map[string]interface{}{
				"step": "lock_acquire",
			})
		}
		return
	}
	defer t.engine.orderLock.Unlock()

	// 【重要】获取锁后再次检查持仓（防止并发下单）
	// 【优化】使用刚刚更新的内存缓存，不再调用API
	// 理由：1. 锁本身防止并发 2. 第一次检查刚更新缓存 3. 减少API调用
	t.engine.mu.RLock()
	positionsAgain := t.engine.CurrentPositions
	t.engine.mu.RUnlock()

	// 获取锁后再次检查（防止并发下单穿透）
	hasAnyPositionAgain := false
	hasSameSidePositionAgain := false
	existingPositionSideAgain := ""
	for _, pos := range positionsAgain {
		if pos == nil || math.Abs(pos.PositionAmt) <= positionAmtEpsilon {
			continue
		}
		hasAnyPositionAgain = true
		if existingPositionSideAgain == "" {
			existingPositionSideAgain = pos.PositionSide
		}
		if pos.PositionSide == positionSide {
			hasSameSidePositionAgain = true
		}
	}

	if robot.DualSidePosition == 0 && hasAnyPositionAgain {
		existingDirectionText := "多头"
		if existingPositionSideAgain == "SHORT" {
			existingDirectionText = "空头"
		}
		targetDirectionText := "多头"
		if positionSide == "SHORT" {
			targetDirectionText = "空头"
		}
		reason := fmt.Sprintf("单向持仓模式：已有持仓（%s），持仓内只能有一单，拒绝新开仓（目标=%s）", existingDirectionText, targetDirectionText)
		if logId > 0 {
			t.saveExecutionLog(ctx, logId, 0, "order_failed", "failed", reason, map[string]interface{}{
				"step":                 "single_position_check_after_lock",
				"dualSidePosition":     robot.DualSidePosition,
				"existingPositionSide": existingPositionSideAgain,
				"targetPositionSide":   positionSide,
				"source":               "exchange_realtime_after_lock",
			})
		}
		g.Log().Infof(ctx, "[RobotTrader] robotId=%d 获取锁后检查：%s，跳过下单", robot.Id, reason)
		return
	}

	if robot.DualSidePosition == 1 && hasSameSidePositionAgain {
		directionText := "多头"
		if positionSide == "SHORT" {
			directionText = "空头"
		}
		reason := fmt.Sprintf("双向持仓模式：%s方向已有持仓，同方向只能一单（禁止加仓），拒绝新开仓", directionText)
		if logId > 0 {
			t.saveExecutionLog(ctx, logId, 0, "order_failed", "failed", reason, map[string]interface{}{
				"step":             "dual_side_same_direction_check_after_lock",
				"dualSidePosition": robot.DualSidePosition,
				"positionSide":     positionSide,
				"source":           "exchange_realtime_after_lock",
			})
		}
		g.Log().Infof(ctx, "[RobotTrader] robotId=%d 获取锁后检查：%s，跳过下单", robot.Id, reason)
		return
	}

	// 【步骤2】参数计算 → 获取市场状态和策略参数（与机器人详情页面相同的方法）
	// 【步骤3】创建订单到平台
	g.Log().Infof(ctx, "[RobotTrader] robotId=%d 【步骤2-3】开始执行下单: logId=%d, direction=%s", robot.Id, logId, signal.Direction)
	execErr := t.executeOpen(ctx, signal, logId)
	if execErr != nil {
		g.Log().Errorf(ctx, "[RobotTrader] robotId=%d 下单失败: logId=%d, err=%v", robot.Id, logId, execErr)
		// 失败日志在 executeOpen 中记录
	} else {
		g.Log().Infof(ctx, "[RobotTrader] robotId=%d 下单成功: logId=%d", robot.Id, logId)
		// 成功日志在 executeOpen 中记录，包含完整的策略参数信息
	}
}

// saveExecutionLog 保存交易执行日志（记录完整的交易执行过程）
// 【优化】交易日志记录完整的交易执行流程，使用TEXT类型可以记录详细信息
// 【优化】增加失败分类和结构化失败原因，便于前端展示
func (t *RobotTrader) saveExecutionLog(ctx context.Context, signalLogId int64, orderId int64, eventType string, status string, message string, eventData map[string]interface{}) {
	robot := t.engine.Robot
	if robot == nil {
		return
	}

	// 序列化事件数据为JSON
	eventDataJSON := "{}"
	if len(eventData) > 0 {
		data, err := json.Marshal(eventData)
		if err == nil {
			eventDataJSON = string(data)
		}
	}

	// 【新增】分析失败原因，提取分类和详情
	failureCategory, failureReason := t.analyzeFailureReason(eventType, message, eventData)

	// 写入交易日志
	insertData := g.Map{
		"signal_log_id":    signalLogId,
		"robot_id":         robot.Id,
		"order_id":         orderId,
		"event_type":       eventType,
		"event_data":       eventDataJSON,
		"status":           status,
		"message":          message,
		"failure_category": failureCategory, // 【新增】
		"failure_reason":   failureReason,   // 【新增】
		"created_at":       time.Now(),
	}
	_, err := g.DB().Model("hg_trading_execution_log").Ctx(ctx).Insert(insertData)
	if err != nil {
		// 兼容旧库：若 hg_trading_execution_log 未加 failure_category/failure_reason 字段，则退回不带新列的插入，避免核心日志写入完全失败
		// 常见错误：
		// - MySQL: Unknown column 'failure_category' in 'field list'
		// - PostgreSQL: column "failure_category" of relation "hg_trading_execution_log" does not exist
		errMsg := err.Error()
		if strings.Contains(errMsg, "Unknown column") || strings.Contains(errMsg, "does not exist") {
			delete(insertData, "failure_category")
			delete(insertData, "failure_reason")
			if _, rerr := g.DB().Model("hg_trading_execution_log").Ctx(ctx).Insert(insertData); rerr == nil {
				g.Log().Warningf(ctx, "[RobotTrader] 交易日志表缺少 failure_* 字段，已降级写入成功: robotId=%d, eventType=%s", robot.Id, eventType)
				// 视为写入成功，继续后续逻辑（例如回写 signal_log）
				err = nil
			} else {
				g.Log().Warningf(ctx, "[RobotTrader] 保存交易日志失败(降级仍失败): robotId=%d, eventType=%s, err=%v", robot.Id, eventType, rerr)
			}
		} else {
			g.Log().Warningf(ctx, "[RobotTrader] 保存交易日志失败: robotId=%d, eventType=%s, err=%v", robot.Id, eventType, err)
		}
	} else {
		if failureCategory != "" {
			g.Log().Debugf(ctx, "[RobotTrader] 交易日志已保存: robotId=%d, eventType=%s, status=%s, category=%s", robot.Id, eventType, status, failureCategory)
		} else {
			g.Log().Debugf(ctx, "[RobotTrader] 交易日志已保存: robotId=%d, eventType=%s, status=%s", robot.Id, eventType, status)
		}
	}

	// 兼容前端"信号日志/执行结果"展示：在最终态时回写 signal_log（避免一直显示"进行中/准备下单"）
	// 说明：很多页面仍在读 hg_trading_signal_log.execute_result/executed。
	if err == nil && signalLogId > 0 && (eventType == "order_failed" || eventType == "order_success") {
		result := message
		// 优先使用结构化失败原因
		if failureReason != "" {
			result = failureReason
		}
		// 避免字段过长（不同环境字段长度可能不同）
		if len(result) > 200 {
			result = result[:200]
		}
		_, _ = g.DB().Model("hg_trading_signal_log").Ctx(ctx).
			Where("id", signalLogId).
			Data(g.Map{
				"executed":       1,
				"execute_result": result,
			}).
			Update()
	}
}

// analyzeFailureReason 分析失败原因，提取分类和详情
// 【新增】自动分析失败原因，生成结构化说明和解决建议
func (t *RobotTrader) analyzeFailureReason(eventType string, message string, eventData map[string]interface{}) (category string, reason string) {
	// 只处理失败事件
	if eventType != "order_failed" {
		return "", ""
	}

	step, _ := eventData["step"].(string)

	switch step {
	case "robot_check", "signal_check":
		category = "system"
		reason = fmt.Sprintf("系统检查失败：%s", message)

	case "auto_trade_check":
		category = "config"
		autoTradeEnabled := 0
		if val, ok := eventData["autoTradeEnabled"]; ok {
			switch v := val.(type) {
			case int:
				autoTradeEnabled = v
			case float64:
				autoTradeEnabled = int(v)
			}
		}
		if autoTradeEnabled == 0 {
			reason = "自动交易开关未开启。解决方案：在机器人设置中开启自动交易开关"
		} else {
			reason = fmt.Sprintf("自动交易检查失败：%s", message)
		}

	case "position_check", "single_position_check", "dual_side_same_direction_check",
		"single_position_check_after_lock", "dual_side_same_direction_check_after_lock":
		category = "position"
		dualSidePosition := 0
		if val, ok := eventData["dualSidePosition"]; ok {
			switch v := val.(type) {
			case int:
				dualSidePosition = v
			case float64:
				dualSidePosition = int(v)
			}
		}

		positionSide, _ := eventData["positionSide"].(string)
		targetPositionSide, _ := eventData["targetPositionSide"].(string)
		existingPositionSide, _ := eventData["existingPositionSide"].(string)

		if dualSidePosition == 0 {
			// 单向持仓模式
			existing := translatePositionSide(existingPositionSide)
			target := translatePositionSide(targetPositionSide)
			reason = fmt.Sprintf("单向持仓模式限制：当前已有%s持仓，持仓内只能有一单。解决方案：1) 等待当前持仓平仓后再下单，2) 切换到双向持仓模式",
				existing)
			if target != "" && existing != target {
				reason = fmt.Sprintf("单向持仓模式限制：当前已有%s持仓，持仓内只能有一单，拒绝新开%s仓。解决方案：1) 等待当前持仓平仓，2) 切换到双向持仓模式",
					existing, target)
			}
		} else {
			// 双向持仓模式
			pos := translatePositionSide(positionSide)
			opposite := translateOppositePositionSide(positionSide)
			reason = fmt.Sprintf("双向持仓模式限制：%s方向已有持仓，同方向不允许加仓。解决方案：1) 等待当前%s持仓平仓后再下单，2) 开反方向的%s仓位",
				pos, pos, opposite)
		}

	case "balance_check":
		category = "balance"
		availableBalance := 0.0
		if val, ok := eventData["available_balance"]; ok {
			switch v := val.(type) {
			case float64:
				availableBalance = v
			case int:
				availableBalance = float64(v)
			}
		}
		if availableBalance <= 0 {
			reason = "账户余额不足或为0。解决方案：1) 充值到交易所账户，2) 降低保证金比例"
		} else {
			errorMsg, _ := eventData["error"].(string)
			if errorMsg != "" {
				reason = fmt.Sprintf("余额检查失败：%s", errorMsg)
			} else {
				reason = fmt.Sprintf("余额检查失败：%s", message)
			}
		}

	case "ticker_check":
		category = "system"
		reason = "无法获取实时行情数据。解决方案：1) 检查网络连接，2) 检查WebSocket服务是否运行"

	case "strategy_params":
		category = "strategy"
		errorMsg, _ := eventData["error"].(string)
		if strings.Contains(errorMsg, "未找到对应的风险偏好") {
			reason = "策略配置缺失：市场状态与风险偏好映射关系未配置。解决方案：1) 检查机器人的风险配置映射（remark字段），2) 重新创建机器人并设置完整的映射关系"
		} else if strings.Contains(errorMsg, "未返回市场状态") {
			reason = "市场分析服务未就绪：全局市场分析器未返回数据。解决方案：1) 等待市场分析服务启动，2) 检查市场分析服务是否正常运行"
		} else {
			reason = fmt.Sprintf("策略参数获取失败：%s", errorMsg)
		}

	case "pre_create_order":
		category = "system"
		reason = fmt.Sprintf("预创建订单失败：%s", message)

	case "exchange_api":
		category = "exchange"
		errorMsg, _ := eventData["error"].(string)
		reason = formatExchangeAPIError(errorMsg)

	case "order_status_update":
		category = "system"
		reason = fmt.Sprintf("订单状态更新失败：%s", message)

	case "lock_acquire":
		category = "system"
		reason = "系统繁忙，无法获取下单锁。解决方案：稍后再试或联系技术支持"

	default:
		category = "system"
		reason = message
	}

	return category, reason
}

// translatePositionSide 翻译持仓方向（英文→中文）
func translatePositionSide(positionSide string) string {
	switch strings.ToUpper(positionSide) {
	case "LONG":
		return "多头"
	case "SHORT":
		return "空头"
	default:
		return positionSide
	}
}

// translateOppositePositionSide 获取反向持仓方向
func translateOppositePositionSide(positionSide string) string {
	switch strings.ToUpper(positionSide) {
	case "LONG":
		return "空头"
	case "SHORT":
		return "多头"
	default:
		return positionSide
	}
}

// formatExchangeAPIError 格式化交易所API错误，提供友好的错误说明
func formatExchangeAPIError(errorMsg string) string {
	// 常见错误码映射
	errorMappings := map[string]string{
		"-1021":                   "时间戳错误。解决方案：检查服务器时间同步（可能需要重启服务或配置NTP）",
		"-2010":                   "订单被交易所拒绝。解决方案：1) 检查账户余额是否充足，2) 检查杠杆设置是否正确，3) 检查订单数量是否符合最小值要求",
		"-2015":                   "无效订单参数。解决方案：检查订单配置（数量、价格、杠杆）是否符合交易所要求",
		"-2019":                   "保证金不足。解决方案：1) 充值到交易所账户，2) 降低杠杆倍数，3) 降低保证金比例",
		"insufficient balance":    "余额不足。解决方案：充值到交易所账户",
		"insufficient margin":     "保证金不足。解决方案：1) 充值，2) 降低杠杆或保证金比例",
		"position not found":      "持仓不存在，可能已被平仓",
		"leverage not set":        "杠杆未设置。解决方案：检查杠杆配置",
		"symbol not found":        "交易对不存在或未开放",
		"market closed":           "市场已关闭，无法交易",
		"order would immediately": "订单会立即触发强平，被拒绝",
		"reduce only":             "只能平仓，不能开仓（可能是风控限制）",
	}

	// 查找匹配的错误码或关键字
	lowerMsg := strings.ToLower(errorMsg)
	for keyword, description := range errorMappings {
		if strings.Contains(lowerMsg, strings.ToLower(keyword)) {
			return fmt.Sprintf("交易所API错误 [%s]：%s", keyword, description)
		}
	}

	// 未匹配到具体错误，返回原始错误信息
	return fmt.Sprintf("交易所API错误：%s", errorMsg)
}

// updateSignalLog 更新预警记录的执行状态（已废弃，改为使用 saveExecutionLog）
// 【优化】不再更新预警记录，改为写入交易日志
func (t *RobotTrader) updateSignalLog(ctx context.Context, logId int64, executed int, result string) {
	if logId == 0 {
		return
	}

	// 【优化】改为写入交易日志，而不是更新预警记录
	eventType := "order_failed"
	status := "failed"
	if executed == 1 {
		eventType = "order_success"
		status = "success"
	} else if result == "" {
		eventType = "order_attempt"
		status = "pending"
	}

	// 构建事件数据
	eventData := map[string]interface{}{
		"executed": executed,
		"result":   result,
	}

	// 保存交易日志
	t.saveExecutionLog(ctx, logId, 0, eventType, status, result, eventData)
}

// TryProcessPendingSignal 尝试处理未处理的预警记录（状态为"准备下单"）
func (t *RobotTrader) TryProcessPendingSignal(ctx context.Context, signal *RobotSignal) {
	// 【重要】使用锁保护，确保读取最新的开关状态
	t.engine.mu.RLock()
	robot := t.engine.Robot
	autoTradeEnabled := 0
	if robot != nil {
		autoTradeEnabled = robot.AutoTradeEnabled
	}
	t.engine.mu.RUnlock()

	// 如果自动下单未开启，不需要处理
	if robot == nil || autoTradeEnabled != 1 {
		return
	}

	// 查找未处理的同方向记录（状态为"准备下单"且未标记为已处理）
	var logRecord struct {
		Id            int64  `json:"id"`
		ExecuteResult string `json:"execute_result"`
	}
	err := g.DB().Model("hg_trading_signal_log").Ctx(ctx).
		Where("robot_id", robot.Id).
		Where("signal_type", signal.Direction).
		Where("executed", 0).
		Where("execute_result", "准备下单").
		Where("(is_processed = 0 OR is_processed IS NULL)"). // 【重要】只查询未处理的预警记录，防止重复使用
		OrderDesc("id").
		Limit(1).
		Scan(&logRecord)

	if err != nil || logRecord.Id == 0 {
		return
	}

	// 尝试处理这个未完成的记录
	g.Log().Infof(ctx, "[RobotTrader] 发现未处理的预警记录: logId=%d, 尝试处理", logRecord.Id)
	t.TryAutoTradeAndUpdate(ctx, signal, logRecord.Id)
}

// TryAutoTrade 尝试自动下单（旧方法，保留兼容）
func (t *RobotTrader) TryAutoTrade(ctx context.Context, signal *RobotSignal) {
	// 【重要】使用锁保护，确保读取最新的开关状态
	t.engine.mu.RLock()
	robot := t.engine.Robot
	autoTradeEnabled := 0
	if robot != nil {
		autoTradeEnabled = robot.AutoTradeEnabled
	}
	t.engine.mu.RUnlock()

	if signal == nil || signal.Direction == "NEUTRAL" {
		return
	}

	// 提前过滤：只处理开仓信号
	if signal.Action != "OPEN_LONG" && signal.Action != "OPEN_SHORT" {
		return
	}

	// 检查是否开启自动交易（使用最新的开关状态）
	if robot == nil || autoTradeEnabled != 1 {
		return // 预警记录已保存，这里只是不下单
	}

	// 检查下单条件（简化：只检查自动交易开关和一个方向只能有一单）
	positionSide := "LONG"
	if signal.Direction == "SHORT" {
		positionSide = "SHORT"
	}
	dbDirection := "long"
	if positionSide == "SHORT" {
		dbDirection = "short"
	}
	count, err := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", robot.Id).
		Where("direction", dbDirection).
		Where("status", OrderStatusOpen).
		Count()
	if err == nil && count > 0 {
		g.Log().Debugf(ctx, "[RobotTrader] robotId=%d %s方向已有持仓", robot.Id, positionSide)
		return
	}

	// 尝试获取锁并下单
	if t.engine.orderLock.TryLock() {
		defer t.engine.orderLock.Unlock()
		err := t.executeOpen(ctx, signal, 0)
		if err != nil {
			g.Log().Warningf(ctx, "[RobotTrader] robotId=%d 下单失败: %v", robot.Id, err)
		}
	}
}

// CheckAndOpenPosition 定时检查并开仓（已废弃）
// 注意：开仓已由预警信号触发（TryAutoTradeAndUpdate），不再使用定时检查机制
// 保留此方法以避免编译错误，但不再被调用
func (t *RobotTrader) CheckAndOpenPosition(ctx context.Context) {
	// 已废弃：开仓已由预警信号触发，不再使用定时检查
	// 此方法保留以避免编译错误，但不会被调用
}

// CheckAndOpenPositionWithSignal 使用指定信号检查并开仓（事件驱动入口）
// 【优化】简化条件检查流程，提高性能和可读性
func (t *RobotTrader) CheckAndOpenPositionWithSignal(ctx context.Context, signal *RobotSignal) {
	// 【前置检查】快速过滤无效信号
	if signal == nil || signal.Direction == "NEUTRAL" {
		return
	}
	if signal.Action != "OPEN_LONG" && signal.Action != "OPEN_SHORT" {
		return
	}

	// 【重要】使用锁保护，确保读取最新的开关状态
	t.engine.mu.RLock()
	robot := t.engine.Robot
	autoTradeEnabled := 0
	if robot != nil {
		autoTradeEnabled = robot.AutoTradeEnabled
	}
	t.engine.mu.RUnlock()

	if robot == nil {
		return
	}

	g.Log().Infof(ctx, "[RobotTrader] robotId=%d 【事件驱动】开始检查开仓条件: direction=%s, action=%s, autoTradeEnabled=%d",
		robot.Id, signal.Direction, signal.Action, autoTradeEnabled)

	// 【订单事件】记录信号生成事件
	RecordSignalGenerated(ctx, robot.Id, signal)

	// 【条件1】检查是否开启自动交易
	if autoTradeEnabled != 1 {
		g.Log().Debugf(ctx, "[RobotTrader] robotId=%d 自动交易未开启(AutoTradeEnabled=%d)", robot.Id, autoTradeEnabled)
		t.saveUnexecutedSignal(ctx, signal, "自动下单未开启")
		t.updateProcessedSignalTime(signal) // 更新时间戳，避免重复检查
		return
	}

	// 检查方向
	positionSide := "LONG"
	if signal.Direction == "SHORT" {
		positionSide = "SHORT"
	}
	direction := "long"
	if signal.Direction == "SHORT" {
		direction = "short"
	}

	// 【优化】使用智能缓存获取持仓，1秒内的缓存视为有效
	positions, err := t.engine.GetPositionsSmart(ctx, 1*time.Second)
	if err != nil {
		g.Log().Warningf(ctx, "[RobotTrader] robotId=%d 获取持仓失败: %v", robot.Id, err)
		t.saveUnexecutedSignal(ctx, signal, fmt.Sprintf("获取持仓失败: %v", err))
		t.updateProcessedSignalTime(signal)
		return
	}

	// 开仓限制规则（按你的最新定义）：
	// - DualSidePosition=1：允许同时持有多+空，但【同方向只能一单】（不允许加仓）
	// - DualSidePosition=0：持仓内只能有一单（不区分多空）
	hasAnyPosition := false
	hasSameSidePosition := false
	existingPositionSide := ""
	for _, pos := range positions {
		if pos == nil || math.Abs(pos.PositionAmt) <= positionAmtEpsilon {
			continue
		}
		hasAnyPosition = true
		if existingPositionSide == "" {
			existingPositionSide = pos.PositionSide
		}
		if pos.PositionSide == positionSide {
			hasSameSidePosition = true
		}
	}

	// 【订单事件】记录开仓检查事件
	checkResult := map[string]interface{}{
		"has_exchange_position_any":       hasAnyPosition,
		"has_exchange_position_same_side": hasSameSidePosition,
		"dualSidePosition":                robot.DualSidePosition,
		"source":                          "exchange_realtime",
	}
	RecordCheckStarted(ctx, robot.Id, direction, checkResult)

	// 单向：任意方向已有持仓都拒绝
	if robot.DualSidePosition == 0 && hasAnyPosition {
		existingDirectionText := "多头"
		if existingPositionSide == "SHORT" {
			existingDirectionText = "空头"
		}
		targetDirectionText := "多头"
		if positionSide == "SHORT" {
			targetDirectionText = "空头"
		}
		reason := fmt.Sprintf("单向持仓模式：已有持仓（%s），持仓内只能有一单，拒绝新开仓（目标=%s）", existingDirectionText, targetDirectionText)
		g.Log().Infof(ctx, "[RobotTrader] robotId=%d %s", robot.Id, reason)
		t.saveUnexecutedSignal(ctx, signal, reason)
		t.updateProcessedSignalTime(signal)
		return
	}

	// 双向：同方向已有持仓则拒绝（不允许加仓）
	if robot.DualSidePosition == 1 && hasSameSidePosition {
		directionText := "多头"
		if positionSide == "SHORT" {
			directionText = "空头"
		}
		reason := fmt.Sprintf("双向持仓模式：%s方向已有持仓，同方向只能一单（禁止加仓），拒绝新开仓", directionText)
		g.Log().Infof(ctx, "[RobotTrader] robotId=%d %s", robot.Id, reason)
		t.saveUnexecutedSignal(ctx, signal, reason)
		t.updateProcessedSignalTime(signal)
		return
	}

	// 【条件3】检查算力
	if !t.checkPower(ctx) {
		t.saveUnexecutedSignal(ctx, signal, "算力不足，请充值")
		t.updateProcessedSignalTime(signal) // 更新时间戳，避免重复检查
		return
	}

	// 【执行开仓】所有条件满足，执行开仓
	g.Log().Infof(ctx, "[RobotTrader] robotId=%d 所有开仓条件满足，准备执行开仓: direction=%s, strength=%.2f, confidence=%.2f",
		robot.Id, signal.Direction, signal.Strength, signal.Confidence)

	openErr := t.executeOpen(ctx, signal, 0)
	if openErr != nil {
		g.Log().Warningf(ctx, "[RobotTrader] 开仓失败: robotId=%d, err=%v", robot.Id, openErr)
		t.saveUnexecutedSignal(ctx, signal, "开仓失败: "+openErr.Error())
		// 开仓失败时不更新已处理时间戳，允许重试（可能是临时错误）
	} else {
		// 开仓成功，更新已处理的信号时间戳，防止重复下单
		t.updateProcessedSignalTime(signal)
	}
}

// updateProcessedSignalTime 更新已处理的信号时间戳（防重复下单）
func (t *RobotTrader) updateProcessedSignalTime(signal *RobotSignal) {
	if signal == nil {
		return
	}
	t.engine.mu.Lock()
	if signal.Timestamp.After(t.engine.LastProcessedSignalTime) {
		t.engine.LastProcessedSignalTime = signal.Timestamp
		g.Log().Debugf(context.Background(), "[RobotTrader] robotId=%d 已更新已处理信号时间戳: signalTime=%v, direction=%s",
			t.engine.Robot.Id, signal.Timestamp, signal.Direction)
	}
	t.engine.mu.Unlock()
}

// checkOpenPositionInDB 检查数据库中是否有该方向的持仓订单（权威来源）
func (t *RobotTrader) checkOpenPositionInDB(ctx context.Context, direction string) (bool, error) {
	count, err := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", t.engine.Robot.Id).
		Where("symbol", t.engine.Robot.Symbol).
		Where("direction", direction).
		Where("status", OrderStatusOpen). // 持仓中
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// syncPositionFromDB 从数据库同步持仓到内存（修复不一致）
func (t *RobotTrader) syncPositionFromDB(ctx context.Context, direction string) {
	// 查询数据库中的持仓订单
	var orders []*entity.TradingOrder
	err := dao.TradingOrder.Ctx(ctx).
		Where("robot_id", t.engine.Robot.Id).
		Where("symbol", t.engine.Robot.Symbol).
		Where("direction", direction).
		Where("status", OrderStatusOpen).
		OrderDesc("created_at").
		Limit(1).
		Scan(&orders)
	if err != nil || len(orders) == 0 {
		return
	}

	order := orders[0]
	positionSide := "LONG"
	if direction == "short" {
		positionSide = "SHORT"
	}

	// 更新内存持仓
	t.engine.mu.Lock()
	if t.engine.CurrentPositions == nil {
		t.engine.CurrentPositions = make([]*exchange.Position, 0)
	}

	// 查找是否已有该方向的持仓
	found := false
	for i, p := range t.engine.CurrentPositions {
		if p.PositionSide == positionSide {
			// 更新现有持仓
			t.engine.CurrentPositions[i].PositionAmt = order.Quantity
			if direction == "short" {
				t.engine.CurrentPositions[i].PositionAmt = -order.Quantity
			}
			t.engine.CurrentPositions[i].EntryPrice = order.OpenPrice
			t.engine.CurrentPositions[i].IsolatedMargin = order.Margin
			found = true
			break
		}
	}

	if !found {
		// 添加新持仓
		positionAmt := order.Quantity
		if direction == "short" {
			positionAmt = -order.Quantity
		}
		t.engine.CurrentPositions = append(t.engine.CurrentPositions, &exchange.Position{
			Symbol:         order.Symbol,
			PositionSide:   positionSide,
			PositionAmt:    positionAmt,
			EntryPrice:     order.OpenPrice,
			IsolatedMargin: order.Margin,
			UnrealizedPnl:  0,
		})
	}
	t.engine.mu.Unlock()

	g.Log().Infof(ctx, "[RobotTrader] robotId=%d 已从数据库同步持仓到内存: direction=%s, orderId=%d", t.engine.Robot.Id, direction, order.Id)
}

// preCreateOrder 预创建订单记录（状态=PENDING，事务保护）
// 【优化】在下单前先创建订单记录，确保数据库保存成功后再调用交易所API
func (t *RobotTrader) preCreateOrder(ctx context.Context, signal *RobotSignal, strategyParams *StrategyParams, leverage int, marginPercent float64, marketState, riskPreference string, quantity, entryPrice, margin float64) (int64, error) {
	robot := t.engine.Robot

	// 【必须】验证必填字段
	if marketState == "" {
		return 0, gerror.New("市场状态为空，无法预创建订单")
	}
	if riskPreference == "" {
		return 0, gerror.New("风险偏好为空，无法预创建订单")
	}

	// 确定方向
	direction := "long"
	side := "BUY"
	if signal.Direction == "SHORT" {
		direction = "short"
		side = "SELL"
	}

	// 生成系统订单号
	// 【修复】使用 PHP 风格格式 "YmdHis" 替代 Go 标准格式，确保 gtime.Format 正确工作
	orderSn := fmt.Sprintf("TO%s%s", gtime.Now().Format("YmdHis"), grand.S(6))

	// 获取策略组ID
	strategyGroupId := robot.StrategyGroupId
	if strategyGroupId == 0 && robot.CurrentStrategy != "" {
		var configData map[string]interface{}
		if err := json.Unmarshal([]byte(robot.CurrentStrategy), &configData); err == nil {
			if groupIdVal, ok := configData["groupId"].(float64); ok {
				strategyGroupId = int64(groupIdVal)
			}
		}
	}

	// 确定订单类型详情
	orderTypeDetail := "market_open_long"
	if signal.Direction == "SHORT" {
		orderTypeDetail = "market_open_short"
	}

	// 计算开仓保证金
	openMargin := margin
	if openMargin <= 0 && quantity > 0 && entryPrice > 0 && leverage > 0 {
		openMargin = (quantity * entryPrice) / float64(leverage)
	}

	// 构建订单数据（状态=PENDING）
	orderData := g.Map{
		// 基础信息
		"user_id":           robot.UserId,
		"robot_id":          robot.Id,
		"strategy_group_id": strategyGroupId,
		"exchange":          t.engine.Platform,
		"symbol":            robot.Symbol,
		"order_sn":          orderSn,
		"exchange_order_id": "", // 预创建时还没有交易所订单ID
		"direction":         direction,
		"quantity":          quantity,

		// 价格信息
		"price":      entryPrice,
		"avg_price":  entryPrice,
		"open_price": entryPrice,
		"mark_price": 0.0,

		// 订单类型
		"order_type":        "MARKET",
		"order_type_detail": orderTypeDetail,
		"exchange_side":     side,

		// 开仓信息
		"open_time":     gtime.Now(),
		"open_margin":   openMargin,
		"margin":        margin, // 保证金（必填字段）
		"open_fee":      0.0,
		"open_fee_coin": "",

		// 市场状态和风险偏好
		"market_state": marketState,
		"risk_level":   riskPreference,

		// 策略参数
		"leverage":       leverage,
		"margin_percent": marginPercent,

		// 订单状态（预创建时为PENDING）
		"status":     OrderStatusPending,
		"created_at": gtime.Now(),
		"updated_at": gtime.Now(),
	}

	// 保存策略参数
	if strategyParams != nil {
		if strategyParams.StopLossPercent > 0 {
			orderData["stop_loss_percent"] = strategyParams.StopLossPercent
		}
		if strategyParams.AutoStartRetreatPercent > 0 {
			orderData["auto_start_retreat_percent"] = strategyParams.AutoStartRetreatPercent
		}
		if strategyParams.ProfitRetreatPercent > 0 {
			orderData["profit_retreat_percent"] = strategyParams.ProfitRetreatPercent
		}
		if strategyParams.LeverageMin > 0 {
			orderData["leverage_min"] = strategyParams.LeverageMin
		}
		if strategyParams.LeverageMax > 0 {
			orderData["leverage_max"] = strategyParams.LeverageMax
		}
		if strategyParams.MarginPercentMin > 0 {
			orderData["margin_percent_min"] = strategyParams.MarginPercentMin
		}
		if strategyParams.MarginPercentMax > 0 {
			orderData["margin_percent_max"] = strategyParams.MarginPercentMax
		}
	}

	// 【重要】在事务中插入订单记录
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return 0, gerror.Wrap(err, "开启事务失败")
	}
	defer tx.Rollback()

	// 【PostgreSQL 兼容】使用 InsertAndGetId() 而不是 Insert() + LastInsertId()
	orderId := int64(0)
	_, err = tx.Model("hg_trading_order").Data(orderData).Insert()
	if err == nil {
		v, e := tx.GetValue("SELECT LASTVAL()")
		if e != nil {
			err = e
		} else {
			orderId = v.Int64()
		}
	}
	if err != nil {
		errDetail := err.Error()
		g.Log().Errorf(ctx, "[RobotTrader] 预创建订单记录失败: robotId=%d, err=%v", robot.Id, err)

		// 检查常见错误
		if strings.Contains(errDetail, "Unknown column") || strings.Contains(errDetail, "doesn't exist") {
			return 0, gerror.Newf("数据库字段不存在: %s，请执行迁移脚本", errDetail)
		}
		if strings.Contains(errDetail, "Field") && strings.Contains(errDetail, "doesn't have a default value") {
			return 0, gerror.Newf("必填字段缺失: %s", errDetail)
		}

		return 0, gerror.Wrap(err, "预创建订单记录失败")
	}

	if orderId == 0 {
		return 0, gerror.New("订单ID为0，预创建订单记录失败")
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return 0, gerror.Wrap(err, "提交事务失败")
	}

	g.Log().Infof(ctx, "[RobotTrader] 预创建订单记录成功: robotId=%d, orderId=%d, direction=%s, status=PENDING", robot.Id, orderId, direction)

	// 【订单事件】记录预创建订单事件
	orderEventData := map[string]interface{}{
		"robot_id":          robot.Id,
		"symbol":            robot.Symbol,
		"direction":         direction,
		"quantity":          quantity,
		"entry_price":       entryPrice,
		"leverage":          leverage,
		"margin":            margin,
		"margin_percent":    marginPercent,
		"market_state":      marketState,
		"risk_preference":   riskPreference,
		"strategy_group_id": strategyGroupId,
	}
	RecordPreCreated(ctx, orderId, orderEventData)

	return orderId, nil
}

// updateOrderStatus 更新订单状态（事务保护）
func (t *RobotTrader) updateOrderStatus(ctx context.Context, orderId int64, status int, exchangeOrderId string, order *exchange.Order) error {
	updateData := g.Map{
		"status":     status,
		"updated_at": gtime.Now(),
	}

	// 【重要】当订单状态变为 OPEN（持仓中）时，清除最高盈利，让每个订单独立计算自己的最高盈利
	if status == OrderStatusOpen {
		updateData["highest_profit"] = 0
		g.Log().Infof(ctx, "[RobotTrader] 订单状态变为OPEN，清除最高盈利: orderId=%d", orderId)
	}

	// 如果提供了交易所订单ID，更新它
	if exchangeOrderId != "" {
		updateData["exchange_order_id"] = exchangeOrderId
	}

	// 如果提供了订单信息，更新相关字段
	if order != nil {
		if order.AvgPrice > 0 {
			updateData["avg_price"] = order.AvgPrice
			if status == OrderStatusOpen {
				updateData["open_price"] = order.AvgPrice
			}
		}
		if order.FilledQty > 0 {
			updateData["filled_qty"] = order.FilledQty
		}
		if order.ClientId != "" {
			updateData["client_order_id"] = order.ClientId
		}
		if order.CreateTime > 0 {
			orderCreateTime := gtime.NewFromTimeStamp(order.CreateTime / 1000)
			updateData["open_time"] = orderCreateTime
			if status == OrderStatusOpen {
				updateData["created_at"] = orderCreateTime
			}
		}
	}

	_, err := dao.TradingOrder.Ctx(ctx).
		Where(dao.TradingOrder.Columns().Id, orderId).
		Update(updateData)
	if err != nil {
		g.Log().Errorf(ctx, "[RobotTrader] 更新订单状态失败: orderId=%d, status=%d, err=%v", orderId, status, err)
		return gerror.Wrap(err, "更新订单状态失败")
	}

	g.Log().Infof(ctx, "[RobotTrader] 订单状态已更新: orderId=%d, status=%d, exchangeOrderId=%s", orderId, status, exchangeOrderId)
	return nil
}

// tryFillOpenFeeFromTradeHistory 尝试从成交(fill)记录中补齐“开仓手续费”
// 说明：
// - 多数交易所 CreateOrder 的响应不包含手续费/已实现盈亏，需要额外从 fills 获取。
// - 这里做轻量重试（成交明细可能有短暂延迟）。
func (t *RobotTrader) tryFillOpenFeeFromTradeHistory(ctx context.Context, localOrderId int64, symbol string, openOrderID string) {
	if localOrderId <= 0 || strings.TrimSpace(openOrderID) == "" {
		return
	}
	if t == nil || t.engine == nil || t.engine.Exchange == nil {
		return
	}

	// 已有 open_fee 则不重复写入
	one, _ := g.DB().Model("hg_trading_order").Ctx(ctx).
		Fields("open_fee", "open_fee_coin").
		Where("id", localOrderId).
		One()
	if one != nil && !one.IsEmpty() {
		if one["open_fee"].Float64() > 0 {
			return
		}
	}

	var (
		agg tradeAggByOrderId
		ok  bool
	)
	for i := 0; i < 3; i++ {
		agg, ok = tryAggFromTradeHistoryByOrderID(ctx, t.engine.Exchange, symbol, openOrderID, 800)
		if ok && agg.Commission > 0 {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
	if !ok || agg.Commission <= 0 {
		return
	}

	data := g.Map{
		"open_fee":      agg.Commission,
		"open_fee_coin": agg.FeeCoin,
		"updated_at":    gtime.Now(),
	}
	_, err := g.DB().Model("hg_trading_order").Ctx(ctx).
		Where("id", localOrderId).
		Data(data).
		Update()
	if err != nil {
		// 容错：部分环境可能尚未迁移 open_fee/open_fee_coin 字段
		if strings.Contains(strings.ToLower(err.Error()), "unknown column") {
			return
		}
		g.Log().Warningf(ctx, "[RobotTrader] 补齐开仓手续费失败: orderId=%d, exchangeOrderId=%s, err=%v", localOrderId, openOrderID, err)
		return
	}

	// 【事件驱动】开仓手续费补齐后，实时刷新运行区间汇总（避免等到平仓才看到手续费）
	var o struct {
		UserId   int64  `orm:"user_id"`
		RobotId  int64  `orm:"robot_id"`
		Exchange string `orm:"exchange"`
		Symbol   string `orm:"symbol"`
	}
	_ = g.DB().Model("hg_trading_order").Ctx(ctx).
		Fields("user_id", "robot_id", "exchange", "symbol").
		Where("id", localOrderId).
		Scan(&o)
	refreshCurrentRunSessionSummaryByRobot(ctx, o.UserId, o.RobotId, o.Exchange, o.Symbol)
}

// checkPower 检查算力是否充足
func (t *RobotTrader) checkPower(ctx context.Context) bool {
	robot := t.engine.Robot
	// 检查用户算力
	var wallet struct {
		Power     float64 `json:"power"`
		GiftPower float64 `json:"giftPower"`
	}
	err := g.DB().Model("hg_toogo_wallet").Ctx(ctx).
		Where("user_id", robot.UserId).
		Scan(&wallet)
	if err != nil {
		g.Log().Warningf(ctx, "[RobotTrader] 查询算力失败: robotId=%d, userId=%d, err=%v", robot.Id, robot.UserId, err)
		return false
	}
	totalPower := wallet.Power + wallet.GiftPower
	g.Log().Debugf(ctx, "[RobotTrader] 算力检查: robotId=%d, userId=%d, power=%.2f, giftPower=%.2f, total=%.2f",
		robot.Id, robot.UserId, wallet.Power, wallet.GiftPower, totalPower)
	// 至少需要1点算力
	return totalPower >= 1
}

// saveUnexecutedSignal 保存未执行的信号记录
func (t *RobotTrader) saveUnexecutedSignal(ctx context.Context, signal *RobotSignal, reason string) {
	robot := t.engine.Robot
	marketState := ""
	if t.engine.LastAnalysis != nil {
		marketState = t.engine.LastAnalysis.MarketState
	}

	// 检查30秒内是否已有相同方向+相同原因的记录，避免重复
	count, _ := g.DB().Model("hg_trading_signal_log").Ctx(ctx).
		Where("robot_id", robot.Id).
		Where("signal_type", signal.Direction).
		Where("execute_result", reason).
		Where("created_at > ?", time.Now().Add(-30*time.Second)).
		Count()
	if count > 0 {
		return // 30秒内已有相同记录，跳过
	}

	_, err := g.DB().Model("hg_trading_signal_log").Ctx(ctx).Insert(g.Map{
		"robot_id":         robot.Id,
		"strategy_id":      0,
		"symbol":           robot.Symbol,
		"signal_type":      signal.Direction,
		"signal_source":    "window_weighted",
		"signal_strength":  signal.Strength,
		"current_price":    signal.CurrentPrice,
		"window_min_price": signal.WindowMinPrice,
		"window_max_price": signal.WindowMaxPrice,
		"threshold":        signal.SignalThreshold,
		"market_state":     marketState,
		"risk_preference":  "", // 【已移除】不再使用 Robot.RiskPreference，统一从映射关系获取
		"target_price":     0,
		"stop_loss":        0,
		"take_profit":      0,
		"executed":         0, // 未执行
		"execute_result":   reason,
		"reason":           signal.Reason,
		"indicators":       "{}",
	})
	if err != nil {
		g.Log().Warningf(ctx, "[RobotTrader] 保存未执行信号失败: robotId=%d, err=%v", robot.Id, err)
	}
}

// CheckAndClosePosition 检查并平仓
// CheckAndClosePosition 已删除 - 自动平仓逻辑已删除

// executeOpen 执行开仓 (直接使用策略模板参数，不再依赖RiskManager)
// getStrategyParamsForTrade 获取策略参数（与机器人详情页面相同的方法）
// 【重新设计】统一逻辑：
// 1. 获取全局实时市场状态
// 2. 根据创建机器人时提交的映射关系选择风险偏好
// 3. 根据实时市场状态+风险偏好获取策略组中对应的策略
func (t *RobotTrader) getStrategyParamsForTrade(ctx context.Context) (marketState, riskPreference string, strategyParams *StrategyParams, err error) {
	robot := t.engine.Robot

	// 【步骤1】获取全局实时市场状态
	globalAnalysis := market.GetMarketAnalyzer().GetAnalysis(t.engine.Platform, robot.Symbol)
	if globalAnalysis == nil {
		return "", "", nil, gerror.New("全局市场分析器未返回市场状态数据，请检查市场分析服务是否正常运行")
	}
	marketState = normalizeMarketState(string(globalAnalysis.MarketState))
	if marketState == "" {
		return "", "", nil, gerror.New("全局市场分析器未返回市场状态数据")
	}

	// 【步骤2】根据创建机器人时提交的映射关系选择风险偏好
	// 【重要】使用引擎已加载的映射关系（从 remark 字段加载）
	t.engine.mu.RLock()
	riskPreference = t.engine.MarketRiskMapping[marketState]
	t.engine.mu.RUnlock()

	if riskPreference == "" {
		t.engine.mu.RLock()
		mappingStr := fmt.Sprintf("%v", t.engine.MarketRiskMapping)
		t.engine.mu.RUnlock()
		return marketState, "", nil, gerror.Newf("市场状态=%s 在映射关系中未找到对应的风险偏好，请检查机器人的风险配置映射关系（映射关系: %s）", marketState, mappingStr)
	}

	// 【步骤3】根据实时市场状态+风险偏好获取策略组中对应的策略
	strategyParams, err = t.engine.loadFullStrategyParams(ctx, marketState, riskPreference)
	if err != nil {
		return marketState, riskPreference, nil, gerror.Wrap(err, "策略参数加载失败")
	}

	return marketState, riskPreference, strategyParams, nil
}

func (t *RobotTrader) executeOpen(ctx context.Context, signal *RobotSignal, signalLogId int64) error {
	robot := t.engine.Robot

	// 【重要】下单时必须从交易所API获取最新余额，不允许使用本地缓存余额
	// 获取缓存余额仅用于对比和日志记录
	t.engine.mu.RLock()
	cachedBalance := t.engine.AccountBalance
	t.engine.mu.RUnlock()

	var balance *exchange.Balance
	balanceSource := "exchange" // 必须使用交易所API

	// 必须从交易所获取最新余额，不允许降级使用缓存
	if t.engine.Exchange == nil {
		errMsg := "余额不足（交易所实例不存在，无法获取交易所余额）"
		if signalLogId > 0 {
			t.saveExecutionLog(ctx, signalLogId, 0, "order_failed", "failed", errMsg, map[string]interface{}{
				"step": "balance_check",
			})
		}
		return gerror.New(errMsg)
	}

	// 【优化】下单前强制刷新余额，确保使用最新数据（缓存0秒 = 强制刷新）
	latestBalance, err := t.engine.GetBalanceSmart(ctx, 0)
	if err != nil || latestBalance == nil {
		errMsg := fmt.Sprintf("余额不足（无法获取余额: %v）", err)
		if signalLogId > 0 {
			errStr := ""
			if err != nil {
				errStr = err.Error()
			}
			t.saveExecutionLog(ctx, signalLogId, 0, "order_failed", "failed", errMsg, map[string]interface{}{
				"step":  "balance_check",
				"error": errStr,
			})
		}
		g.Log().Errorf(ctx, "[RobotTrader] robotId=%d 下单失败：无法获取余额，err=%v", robot.Id, err)
		return gerror.New(errMsg)
	}

	balance = latestBalance
	// 更新缓存，供前端页面和其他模块使用
	t.engine.mu.Lock()
	t.engine.AccountBalance = latestBalance
	t.engine.LastBalanceUpdate = time.Now()
	t.engine.mu.Unlock()

	// 对比缓存值和API值，记录差异
	if cachedBalance != nil {
		diff := latestBalance.AvailableBalance - cachedBalance.AvailableBalance
		if math.Abs(diff) > 0.01 {
			g.Log().Infof(ctx, "[RobotTrader] robotId=%d 【余额来源=交易所API】下单前已从交易所获取最新余额: %.2f USDT（缓存值=%.2f USDT，差异=%.2f USDT）",
				robot.Id, balance.AvailableBalance, cachedBalance.AvailableBalance, diff)
		} else {
			g.Log().Infof(ctx, "[RobotTrader] robotId=%d 【余额来源=交易所API】下单前已从交易所获取最新余额: %.2f USDT（与缓存一致）", robot.Id, balance.AvailableBalance)
		}
	} else {
		g.Log().Infof(ctx, "[RobotTrader] robotId=%d 【余额来源=交易所API】下单前已从交易所获取最新余额: %.2f USDT（缓存不存在）", robot.Id, balance.AvailableBalance)
	}

	// 检查交易所余额是否充足
	if balance.AvailableBalance <= 0 {
		errMsg := "余额不足（交易所余额为0或负数）"
		if signalLogId > 0 {
			t.saveExecutionLog(ctx, signalLogId, 0, "order_failed", "failed", errMsg, map[string]interface{}{
				"step":              "balance_check",
				"available_balance": balance.AvailableBalance,
				"total_balance":     balance.TotalBalance,
			})
		}
		return gerror.New(errMsg)
	}

	ticker := t.engine.LastTicker
	if ticker == nil {
		errMsg := "获取行情失败"
		if signalLogId > 0 {
			t.saveExecutionLog(ctx, signalLogId, 0, "order_failed", "failed", errMsg, map[string]interface{}{
				"step": "ticker_check",
			})
		}
		return gerror.New(errMsg)
	}

	// 【步骤2】参数计算 → 获取市场状态和策略参数（与机器人详情页面相同的方法）
	g.Log().Infof(ctx, "[RobotTrader] robotId=%d 【步骤2】开始获取策略参数", robot.Id)
	marketState, riskPreference, strategyParams, err := t.getStrategyParamsForTrade(ctx)
	if err != nil {
		g.Log().Errorf(ctx, "[RobotTrader] robotId=%d 【步骤2】获取策略参数失败: %v", robot.Id, err)
		if signalLogId > 0 {
			t.saveExecutionLog(ctx, signalLogId, 0, "order_failed", "failed", fmt.Sprintf("获取策略参数失败: %v", err), map[string]interface{}{
				"step":  "strategy_params",
				"error": err.Error(),
			})
		}
		return gerror.Wrap(err, "获取策略参数失败")
	}

	// 使用策略模板中的参数（直接使用，不再计算范围中值）
	leverage := strategyParams.LeverageMin
	if leverage <= 0 {
		leverage = 10
	}
	marginPercent := strategyParams.MarginPercentMin
	if marginPercent <= 0 {
		marginPercent = 10
	}

	g.Log().Infof(ctx, "[RobotTrader] robotId=%d 【步骤2】策略参数获取成功: 市场=%s, 风险偏好=%s, 杠杆=%dx, 保证金=%.1f%%, 止损=%.1f%%, 启动止盈=%.1f%%, 止盈回撤=%.1f%%, 时间窗口=%d秒, 波动阈值=%.1f USDT, 可用余额=%.2f USDT（来源=%s）",
		robot.Id, marketState, riskPreference, leverage, marginPercent,
		strategyParams.StopLossPercent, strategyParams.AutoStartRetreatPercent, strategyParams.ProfitRetreatPercent,
		strategyParams.Window, strategyParams.Threshold, balance.AvailableBalance, balanceSource)

	// 【订单金额计算】计算公式（基于可用余额，已扣除已用保证金）：
	// ① 保证金 = 可用余额（AvailableBalance）× 保证金比例 / 100
	// ② 订单金额 = 保证金 × 杠杆
	// ③ 下单数量 = 订单金额 / 当前价格
	// 【修复】直接使用 AvailableBalance，避免重复计算（AvailableBalance 已经是扣除已用保证金后的可用余额）
	margin := balance.AvailableBalance * marginPercent / 100
	orderValue := margin * float64(leverage)

	// 【详细日志】输出订单金额计算过程
	g.Log().Infof(ctx, "[RobotTrader] robotId=%d 【订单金额计算】交易所账户权益=%.2f USDT, 未实现盈亏=%.2f USDT, 可用余额=%.2f USDT, 保证金比例=%.1f%%, 杠杆=%dx, 计算过程: 保证金=%.2f × %.1f%% / 100 = %.2f USDT, 订单金额=%.2f × %d = %.2f USDT",
		robot.Id, balance.TotalBalance, balance.UnrealizedPnl, balance.AvailableBalance, marginPercent, leverage,
		balance.AvailableBalance, marginPercent, margin, margin, leverage, orderValue)

	quantity := margin * float64(leverage) / ticker.LastPrice

	// 【优化】最小下单数量为0.0001，如果计算出的数量小于最小值，使用最小值0.0001
	minQuantity := 0.0001
	if quantity < minQuantity {
		quantity = minQuantity
	}

	// 确定方向
	side := "BUY"
	positionSide := "LONG"
	if signal.Direction == "SHORT" {
		side = "SELL"
		positionSide = "SHORT"
	}

	entryPrice := ticker.LastPrice // 预估开仓价格

	// 【步骤3】创建订单到平台
	g.Log().Infof(ctx, "[RobotTrader] robotId=%d 【步骤3】开始创建订单: 方向=%s, 数量=%.4f, 价格=%.2f, 杠杆=%dx, 保证金=%.2f USDT",
		robot.Id, positionSide, quantity, entryPrice, leverage, margin)

	// 步骤3.1：预创建订单记录（状态=PENDING，事务保护）
	localOrderId, err := t.preCreateOrder(ctx, signal, strategyParams, leverage, marginPercent, marketState, riskPreference, quantity, entryPrice, margin)
	if err != nil {
		errMsg := fmt.Sprintf("预创建订单记录失败: robotId=%d, err=%v", robot.Id, err)
		g.Log().Errorf(ctx, "[RobotTrader] robotId=%d 【步骤3.1】%s", robot.Id, errMsg)
		if signalLogId > 0 {
			t.saveExecutionLog(ctx, signalLogId, 0, "order_failed", "failed", errMsg, map[string]interface{}{
				"step":  "pre_create_order",
				"error": err.Error(),
			})
		}
		return gerror.Wrap(err, "预创建订单记录失败，无法执行开仓")
	}

	g.Log().Infof(ctx, "[RobotTrader] robotId=%d 【步骤3.1】预创建订单记录成功: orderId=%d, 准备调用交易所API", robot.Id, localOrderId)

	// 步骤3.2：设置杠杆（只有杠杆变化时才调用API）
	// 【重要】交易所下单API不支持传入杠杆参数，必须先设置杠杆
	// 【优化】缓存已设置的杠杆，避免重复调用API
	t.engine.mu.RLock()
	lastSetLeverage := t.engine.LastSetLeverage
	t.engine.mu.RUnlock()

	if lastSetLeverage != leverage {
		g.Log().Infof(ctx, "[RobotTrader] robotId=%d 【步骤3.2】设置杠杆: symbol=%s, leverage=%dx (上次=%dx)",
			robot.Id, robot.Symbol, leverage, lastSetLeverage)
		if err := t.engine.Exchange.SetLeverage(ctx, robot.Symbol, leverage); err != nil {
			g.Log().Warningf(ctx, "[RobotTrader] robotId=%d 设置杠杆失败: %v（继续下单，使用平台当前杠杆）", robot.Id, err)
		} else {
			// 设置成功，更新缓存
			t.engine.mu.Lock()
			t.engine.LastSetLeverage = leverage
			t.engine.mu.Unlock()
		}
	} else {
		g.Log().Debugf(ctx, "[RobotTrader] robotId=%d 【步骤3.2】杠杆未变化，跳过设置: leverage=%dx", robot.Id, leverage)
	}

	// 步骤3.3：调用交易所API下单
	g.Log().Infof(ctx, "[RobotTrader] robotId=%d 【步骤3.3】调用交易所API下单: symbol=%s, side=%s, positionSide=%s, quantity=%.4f",
		robot.Id, robot.Symbol, side, positionSide, quantity)

	// 【订单日志1】提交API下单 - 记录提交的具体内容
	requestData := map[string]interface{}{
		"symbol":                 robot.Symbol,
		"side":                   side,
		"position_side":          positionSide,
		"type":                   "MARKET",
		"quantity":               quantity,
		"leverage":               leverage,
		"margin":                 margin,
		"margin_percent":         marginPercent,
		"entry_price":            entryPrice,
		"market_state":           marketState,
		"risk_preference":        riskPreference,
		"stop_loss_percent":      strategyParams.StopLossPercent,
		"auto_start_retreat":     strategyParams.AutoStartRetreatPercent,
		"profit_retreat_percent": strategyParams.ProfitRetreatPercent,
	}
	t.saveExecutionLog(ctx, signalLogId, localOrderId, "order_submit", "pending", fmt.Sprintf("提交API下单: %s方向, 数量%.4f, 价格%.2f, 杠杆%dx, 保证金%.2f USDT", positionSide, quantity, entryPrice, leverage, margin), requestData)

	order, err := t.engine.Exchange.CreateOrder(ctx, &exchange.OrderRequest{
		Symbol:       robot.Symbol,
		Side:         side,
		PositionSide: positionSide,
		Type:         "MARKET",
		Quantity:     quantity,
	})

	responseData := map[string]interface{}{}
	if err != nil {
		responseData["error"] = err.Error()
		g.Log().Errorf(ctx, "[RobotTrader] robotId=%d 【步骤3.3】交易所API下单失败: orderId=%d, err=%v", robot.Id, localOrderId, err)

		// 【订单事件】记录交易所下单失败事件
		RecordExchangeOrdered(ctx, localOrderId, "", requestData, responseData, false)
		RecordOrderFailed(ctx, localOrderId, "", err.Error(), map[string]interface{}{
			"step":         "exchange_api",
			"request_data": requestData,
			"error":        err.Error(),
		})

		// 交易所下单失败，更新订单状态为FAILED
		updateErr := t.updateOrderStatus(ctx, localOrderId, OrderStatusFailed, "", nil)
		if updateErr != nil {
			g.Log().Errorf(ctx, "[RobotTrader] robotId=%d 更新订单状态为FAILED失败: orderId=%d, err=%v", robot.Id, localOrderId, updateErr)
		}
		return gerror.Wrap(err, "交易所下单失败")
	}

	g.Log().Infof(ctx, "[RobotTrader] robotId=%d 【步骤3.3】交易所API下单成功: exchangeOrderId=%s, avgPrice=%.2f, filledQty=%.4f",
		robot.Id, order.OrderId, order.AvgPrice, order.FilledQty)

	// 【订单事件】记录交易所下单成功事件
	responseData = map[string]interface{}{
		"exchange_order_id": order.OrderId,
		"client_order_id":   order.ClientId,
		"avg_price":         order.AvgPrice,
		"filled_qty":        order.FilledQty,
		"status":            order.Status,
		"create_time":       order.CreateTime,
	}
	RecordExchangeOrdered(ctx, localOrderId, order.OrderId, requestData, responseData, true)

	// 步骤3.4：更新订单状态为OPEN（包含交易所订单ID）
	entryPrice = ticker.LastPrice
	if order.AvgPrice > 0 {
		entryPrice = order.AvgPrice
	}

	g.Log().Infof(ctx, "[RobotTrader] robotId=%d 【步骤3.4】更新订单状态为OPEN: orderId=%d, exchangeOrderId=%s, entryPrice=%.2f",
		robot.Id, localOrderId, order.OrderId, entryPrice)

	err = t.updateOrderStatus(ctx, localOrderId, OrderStatusOpen, order.OrderId, order)
	if err != nil {
		g.Log().Errorf(ctx, "[RobotTrader] robotId=%d 【步骤3.4】更新订单状态失败: orderId=%d, err=%v", robot.Id, localOrderId, err)
		RecordOrderFailed(ctx, localOrderId, order.OrderId, fmt.Sprintf("更新订单状态失败: %v", err), map[string]interface{}{
			"step":          "order_status_update",
			"target_status": OrderStatusOpen,
			"error":         err.Error(),
		})
	} else {
		g.Log().Infof(ctx, "[RobotTrader] robotId=%d 【步骤3.4】订单状态更新成功: orderId=%d, status=OPEN", robot.Id, localOrderId)
		RecordOrderFilled(ctx, localOrderId, order.OrderId, map[string]interface{}{
			"exchange_order_id": order.OrderId,
			"avg_price":         entryPrice,
			"quantity":          quantity,
			"filled_qty":        order.FilledQty,
			"leverage":          leverage,
			"margin":            margin,
		})

		// 【财务字段补齐】开仓手续费：从成交(fill)记录汇总回填（以平台为准）
		// 说明：大多数交易所“下单返回”不包含手续费，必须从 fills 获取。
		t.tryFillOpenFeeFromTradeHistory(ctx, localOrderId, robot.Symbol, order.OrderId)

		// 推送“订单变更”事件给前端：用于详情弹窗挂单/成交明细秒级刷新（不依赖10s轮询）
		if robot.UserId > 0 {
			dir := "long"
			if positionSide == "SHORT" {
				dir = "short"
			}
			websocket.SendToUser(robot.UserId, &websocket.WResponse{
				Event: "toogo/robot/trade/event",
				Data: g.Map{
					"type":            "order_delta",
					"action":          "open",
					"robotId":         robot.Id,
					"symbol":          robot.Symbol,
					"positionSide":    positionSide,
					"direction":       dir,
					"localOrderId":    localOrderId,
					"exchangeOrderId": order.OrderId,
					"status":          OrderStatusOpen,
					"quantity":        quantity,
					"avgPrice":        entryPrice,
					"margin":          margin,
					"marginPercent":   marginPercent,
					"marketState":     marketState,
					"riskPreference":  riskPreference,
					"ts":              gtime.Now().TimestampMilli(),
				},
			})
		}
	}

	// 标记信号日志最终成功（避免前端一直显示“进行中”）
	if signalLogId > 0 {
		t.saveExecutionLog(ctx, signalLogId, localOrderId, "order_success", "success",
			fmt.Sprintf("下单成功: %s %s qty=%.4f exchangeOrderId=%s", robot.Symbol, positionSide, quantity, order.OrderId),
			map[string]interface{}{
				"step":            "done",
				"exchangeOrderId": order.OrderId,
			})
	}

	g.Log().Infof(ctx, "[RobotTrader] robotId=%d 【完成】开仓成功: localOrderId=%d, exchangeOrderId=%s, side=%s, qty=%.4f, price=%.2f, leverage=%dx, margin=%.2f%%, market=%s, risk=%s, 止损=%.1f%%, 启动止盈=%.1f%%, 止盈回撤=%.1f%%",
		robot.Id, localOrderId, order.OrderId, side, quantity, entryPrice, leverage, marginPercent, marketState, riskPreference,
		strategyParams.StopLossPercent, strategyParams.AutoStartRetreatPercent, strategyParams.ProfitRetreatPercent)

	// 【优化】步骤5：更新内存缓存（成功后）
	t.engine.mu.Lock()
	// 初始化持仓跟踪器
	t.engine.PositionTrackers[positionSide] = &PositionTracker{
		PositionSide: positionSide,
		EntryMargin:  margin,
		EntryTime:    time.Now(),
	}
	// 更新 CurrentPositions：添加或更新持仓信息
	if t.engine.CurrentPositions == nil {
		t.engine.CurrentPositions = make([]*exchange.Position, 0)
	}
	// 查找是否已有该方向的持仓
	found := false
	for i, p := range t.engine.CurrentPositions {
		if p.PositionSide == positionSide {
			// 更新现有持仓
			t.engine.CurrentPositions[i].PositionAmt = quantity
			if side == "SELL" {
				t.engine.CurrentPositions[i].PositionAmt = -quantity
			}
			t.engine.CurrentPositions[i].EntryPrice = entryPrice
			t.engine.CurrentPositions[i].IsolatedMargin = margin
			t.engine.CurrentPositions[i].UnrealizedPnl = 0
			found = true
			break
		}
	}
	if !found {
		// 添加新持仓
		positionAmt := quantity
		if side == "SELL" {
			positionAmt = -quantity
		}
		t.engine.CurrentPositions = append(t.engine.CurrentPositions, &exchange.Position{
			Symbol:         robot.Symbol,
			PositionSide:   positionSide,
			PositionAmt:    positionAmt,
			EntryPrice:     entryPrice,
			IsolatedMargin: margin,
			UnrealizedPnl:  0,
		})
	}

	// 【内存优化】新订单时重置 PositionTracker（清除旧订单的监控数据）
	t.engine.PositionTrackers[positionSide] = &PositionTracker{
		PositionSide:      positionSide,
		EntryMargin:       margin,
		EntryTime:         time.Now(),
		HighestProfit:     0,     // 重置最高盈利
		TakeProfitEnabled: false, // 重置止盈回撤开关
	}
	g.Log().Infof(ctx, "[RobotTrader] robotId=%d 新订单已重置监控数据: positionSide=%s", robot.Id, positionSide)

	t.engine.mu.Unlock()

	// 【订单日志2】订单成功 - 记录订单成功信息
	t.saveExecutionLog(ctx, signalLogId, localOrderId, "order_success", "success", fmt.Sprintf("订单成功: 交易所订单ID=%s, %s方向, 数量%.4f, 成交价%.2f, 杠杆%dx", order.OrderId, positionSide, quantity, entryPrice, leverage), map[string]interface{}{
		"exchange_order_id":      order.OrderId,
		"local_order_id":         localOrderId,
		"side":                   side,
		"position_side":          positionSide,
		"quantity":               quantity,
		"price":                  entryPrice,
		"avg_price":              order.AvgPrice,
		"filled_qty":             order.FilledQty,
		"leverage":               leverage,
		"margin":                 margin,
		"margin_percent":         marginPercent,
		"market_state":           marketState,
		"risk_preference":        riskPreference,
		"stop_loss_percent":      strategyParams.StopLossPercent,
		"auto_start_retreat":     strategyParams.AutoStartRetreatPercent,
		"profit_retreat_percent": strategyParams.ProfitRetreatPercent,
	})

	// 【优化】每1秒自动同步订单，无需手动触发
	go func() {
		t.engine.syncAccountDataIfNeeded(ctx, "after_trade")
		g.Log().Debugf(ctx, "[RobotTrader] robotId=%d 开仓成功，等待下次自动同步: side=%s", robot.Id, positionSide)
	}()

	return nil
}

// recordOrder 记录订单到数据库（保留用于外部持仓补全等场景）
// 【注意】新的下单流程使用 preCreateOrder + updateOrderStatus，此函数仅用于兼容旧代码
func (t *RobotTrader) recordOrder(ctx context.Context, order *exchange.Order, signal *RobotSignal, strategyParams *StrategyParams, leverage int, marginPercent float64, marketState, riskPreference string) (int64, error) {
	robot := t.engine.Robot

	// 【必须】使用传入的市场状态和风险偏好（来自实时映射关系），不允许降级
	if marketState == "" {
		errMsg := fmt.Sprintf("机器人ID=%d 记录订单时市场状态为空，无法保存订单", robot.Id)
		g.Log().Errorf(ctx, "[RobotTrader] %s", errMsg)
		return 0, gerror.New(errMsg)
	}

	if riskPreference == "" {
		errMsg := fmt.Sprintf("机器人ID=%d 记录订单时风险偏好为空，无法保存订单", robot.Id)
		g.Log().Errorf(ctx, "[RobotTrader] %s", errMsg)
		return 0, gerror.New(errMsg)
	}

	g.Log().Infof(ctx, "[RobotTrader] 记录订单: robotId=%d, orderId=%s, marketState=%s, riskPreference=%s(来自实时映射关系)", robot.Id, order.OrderId, marketState, riskPreference)

	// 【重要】将 PositionSide (LONG/SHORT) 转换为 direction (long/short) 以匹配数据库字段
	direction := "long"
	if order.PositionSide == "SHORT" {
		direction = "short"
	}

	// 【修复】市价单下单成功后立即设置为"持仓中"状态（OrderStatusOpen），避免重复下单
	var orderStatus int = OrderStatusPending // 默认未成交（使用统一的订单状态常量）
	if order.Type == "MARKET" {
		// 市价单立即成交，状态设置为 OrderStatusOpen（持仓中）
		orderStatus = OrderStatusOpen
		g.Log().Infof(ctx, "[RobotTrader] 市价单下单成功，立即设置状态为持仓中: robotId=%d, orderId=%s, direction=%s, status=%d",
			robot.Id, order.OrderId, direction, orderStatus)
	} else {
		// 非市价单，根据交易所返回的状态转换
		if order.Status == "FILLED" || order.Status == "PARTIALLY_FILLED" {
			orderStatus = OrderStatusOpen
		} else {
			orderStatus = OrderStatusPending
		}
	}

	// 【修复】计算开仓保证金（根据数量、价格和杠杆）
	openMargin := 0.0
	if order.Quantity > 0 && order.AvgPrice > 0 && leverage > 0 {
		openMargin = (order.Quantity * order.AvgPrice) / float64(leverage)
	} else if order.Quantity > 0 && order.Price > 0 && leverage > 0 {
		openMargin = (order.Quantity * order.Price) / float64(leverage)
	}

	// 【修复】确定订单类型详情
	orderTypeDetail := ""
	if order.Type == "MARKET" {
		if order.Side == "BUY" {
			orderTypeDetail = "market_open_long"
		} else {
			orderTypeDetail = "market_open_short"
		}
	} else if order.Type == "LIMIT" {
		if order.Side == "BUY" {
			orderTypeDetail = "limit_open_long"
		} else {
			orderTypeDetail = "limit_open_short"
		}
	}

	// 【新增】生成系统订单号（格式：TO + 时间戳 + 6位随机字符）
	orderSn := fmt.Sprintf("TO%s%s", gtime.Now().Format("20060102150405"), grand.S(6))

	// 【新增】获取策略组ID（从机器人获取）
	strategyGroupId := robot.StrategyGroupId
	if strategyGroupId == 0 {
		// 如果机器人没有绑定策略组，尝试从 CurrentStrategy 获取（兼容旧数据）
		if robot.CurrentStrategy != "" {
			var configData map[string]interface{}
			if err := json.Unmarshal([]byte(robot.CurrentStrategy), &configData); err == nil {
				if groupIdVal, ok := configData["groupId"].(float64); ok {
					strategyGroupId = int64(groupIdVal)
				}
			}
		}
	}

	// 【优化】使用交易所返回的创建时间（如果存在），否则使用本地时间
	orderCreateTime := gtime.Now()
	if order.CreateTime > 0 {
		// Order.CreateTime 是毫秒时间戳，转换为 gtime.Time
		orderCreateTime = gtime.NewFromTimeStamp(order.CreateTime / 1000)
	}

	// 构建订单数据
	orderData := g.Map{
		"user_id":           robot.UserId,
		"robot_id":          robot.Id,
		"strategy_group_id": strategyGroupId,
		"exchange":          t.engine.Platform,
		"symbol":            robot.Symbol,
		"order_sn":          orderSn,
		"exchange_order_id": order.OrderId,
		"direction":         direction,
		"open_price":        order.AvgPrice,
		"quantity":          order.Quantity,
		"leverage":          leverage,
		"margin":            openMargin,
		"status":            orderStatus,
		"order_type":        order.Type,
		"order_type_detail": orderTypeDetail,
		"exchange_side":     order.Side,
		"price":             order.Price,
		"avg_price":         order.AvgPrice,
		"filled_qty":        order.FilledQty,
		"open_margin":       openMargin,
		"open_time":         orderCreateTime,
		"created_at":        orderCreateTime,
		"updated_at":        gtime.Now(),
		"market_state":      marketState,
		"risk_level":        riskPreference,
		// 【关键】明确初始化止盈回撤相关字段，避免依赖数据库默认值
		"profit_retreat_started": 0, // 默认未开启止盈回撤
		"highest_profit":         0, // 初始最高盈利为0
		"unrealized_profit":      0, // 初始未实现盈亏为0
	}

	// 保存策略参数
	if strategyParams != nil {
		if strategyParams.StopLossPercent > 0 {
			orderData["stop_loss_percent"] = strategyParams.StopLossPercent
		}
		if strategyParams.AutoStartRetreatPercent > 0 {
			orderData["auto_start_retreat_percent"] = strategyParams.AutoStartRetreatPercent
		}
		if strategyParams.ProfitRetreatPercent > 0 {
			orderData["profit_retreat_percent"] = strategyParams.ProfitRetreatPercent
		}
		if strategyParams.LeverageMin > 0 {
			orderData["leverage_min"] = strategyParams.LeverageMin
		}
		if strategyParams.LeverageMax > 0 {
			orderData["leverage_max"] = strategyParams.LeverageMax
		}
		if strategyParams.MarginPercentMin > 0 {
			orderData["margin_percent_min"] = strategyParams.MarginPercentMin
		}
		if strategyParams.MarginPercentMax > 0 {
			orderData["margin_percent_max"] = strategyParams.MarginPercentMax
		}
	}

	// 【重要】尝试插入订单数据
	// 【PostgreSQL 兼容】使用 InsertAndGetId() 而不是 Insert() + LastInsertId()
	orderId, err := dao.TradingOrder.Ctx(ctx).Data(orderData).InsertAndGetId()
	if err != nil && strings.Contains(err.Error(), "LastInsertId is not supported") {
		tx, e := g.DB().Begin(ctx)
		if e != nil {
			return 0, e
		}
		defer tx.Rollback()
		if _, e = tx.Model("hg_trading_order").Data(orderData).Insert(); e != nil {
			return 0, e
		}
		v, e := tx.GetValue("SELECT LASTVAL()")
		if e != nil {
			return 0, e
		}
		if e = tx.Commit(); e != nil {
			return 0, e
		}
		orderId = v.Int64()
		err = nil
	}
	if err != nil {
		errMsg := fmt.Sprintf("保存订单记录失败: robotId=%d, exchangeOrderId=%s, err=%v", robot.Id, order.OrderId, err)
		errDetail := err.Error()
		g.Log().Errorf(ctx, "[RobotTrader] %s", errMsg)
		g.Log().Errorf(ctx, "[RobotTrader] 数据库错误详情: err=%+v, err.Error()=%s", err, errDetail)
		return 0, gerror.New(errMsg)
	}

	if orderId == 0 {
		errMsg := fmt.Sprintf("订单ID为0: robotId=%d, exchangeOrderId=%s（可能是数据库插入失败但未返回错误）", robot.Id, order.OrderId)
		g.Log().Errorf(ctx, "[RobotTrader] %s", errMsg)
		return 0, gerror.New(errMsg)
	}

	g.Log().Infof(ctx, "[RobotTrader] 订单记录已保存: robotId=%d, orderId=%d, exchangeOrderId=%s, direction=%s, status=%d",
		robot.Id, orderId, order.OrderId, order.PositionSide, orderStatus)
	return orderId, nil
}

// saveExecutedSignal 保存已执行的信号记录
func (t *RobotTrader) saveExecutedSignal(ctx context.Context, signal *RobotSignal, orderId string) {
	robot := t.engine.Robot
	marketState := ""
	if t.engine.LastAnalysis != nil {
		marketState = t.engine.LastAnalysis.MarketState
	}

	_, err := g.DB().Model("hg_trading_signal_log").Ctx(ctx).Insert(g.Map{
		"robot_id":         robot.Id,
		"strategy_id":      0,
		"symbol":           robot.Symbol,
		"signal_type":      signal.Direction,
		"signal_source":    "window_weighted",
		"signal_strength":  signal.Strength,
		"current_price":    signal.CurrentPrice,
		"window_min_price": signal.WindowMinPrice,
		"window_max_price": signal.WindowMaxPrice,
		"threshold":        signal.SignalThreshold,
		"market_state":     marketState,
		"risk_preference":  "", // 【已移除】不再使用 Robot.RiskPreference，统一从映射关系获取
		"target_price":     0,
		"stop_loss":        0,
		"take_profit":      0,
		"executed":         1, // 已执行
		"execute_result":   "下单成功: " + orderId,
		"reason":           signal.Reason,
		"indicators":       "{}",
	})
	if err != nil {
		g.Log().Warningf(ctx, "[RobotTrader] 保存已执行信号失败: robotId=%d, err=%v", robot.Id, err)
	}
}

// shouldCloseFromOrder 已删除 - 自动平仓逻辑已删除
// shouldClose 已删除 - 自动平仓逻辑已删除
// executeClose 已删除 - 自动平仓逻辑已删除

// loadVolatilityConfig 从量化管理加载波动率配置（简化版）
// 只加载市场状态阈值和5个时间周期权重（1m/5m/15m/30m/1h）
// 优先级：货币对特定配置 > 全局配置 > 默认值
// 【优化】添加60秒缓存，避免频繁数据库查询
func (e *RobotEngine) loadVolatilityConfig(ctx context.Context) {
	// 检查缓存是否有效（60秒内不重复查询）
	e.mu.RLock()
	lastUpdate := e.LastVolatilityConfigUpdate
	hasConfig := e.VolatilityConfig != nil
	e.mu.RUnlock()

	if hasConfig && time.Since(lastUpdate) < 60*time.Second {
		return // 使用缓存，不查询数据库
	}

	// 从量化管理配置表读取（货币对特定配置或全局配置）
	config, err := service.ToogoVolatilityConfig().GetBySymbol(ctx, e.Robot.Symbol)
	if err == nil && config != nil {
		e.mu.Lock()
		symbolName := "全局"
		if config.Symbol != nil && *config.Symbol != "" {
			symbolName = *config.Symbol
		}
		e.VolatilityConfig = &VolatilityConfig{
			HighVolatilityThreshold: config.HighVolatilityThreshold,
			LowVolatilityThreshold:  config.LowVolatilityThreshold,
			TrendStrengthThreshold:  config.TrendStrengthThreshold,
			Weight1m:                config.Weight1m,
			Weight5m:                config.Weight5m,
			Weight15m:               config.Weight15m,
			Weight30m:               config.Weight30m,
			Weight1h:                config.Weight1h,
			Symbol:                  symbolName,
		}
		e.LastVolatilityConfigUpdate = time.Now() // 更新缓存时间
		e.mu.Unlock()
		g.Log().Debugf(ctx, "[RobotEngine] robotId=%d 加载波动率配置: symbol=%s, configType=%s",
			e.Robot.Id, e.Robot.Symbol, symbolName)
		return
	}

	// 使用默认值
	e.mu.Lock()
	e.VolatilityConfig = &VolatilityConfig{
		HighVolatilityThreshold: highVolatilityThreshold,
		LowVolatilityThreshold:  lowVolatilityThreshold,
		TrendStrengthThreshold:  trendStrengthThreshold,
		Weight1m:                0.10, // 1分钟权重10%
		Weight5m:                0.15, // 5分钟权重15%
		Weight15m:               0.25, // 15分钟权重25%
		Weight30m:               0.25, // 30分钟权重25%
		Weight1h:                0.25, // 1小时权重25%
		Symbol:                  "默认",
	}
	e.LastVolatilityConfigUpdate = time.Now() // 更新缓存时间
	e.mu.Unlock()
}

// getFloatFromMap 从map中安全获取float64值
func getFloatFromMap(m map[string]interface{}, key string, defaultValue float64) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case float32:
			return float64(val)
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return defaultValue
}
