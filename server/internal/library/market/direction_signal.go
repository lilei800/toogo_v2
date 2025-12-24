// Package market 方向信号服务（精简版）
// 只保留核心功能：趋势方向信号
package market

import (
	"context"
	"math"
	"sync"
	"time"

	"hotgo/internal/library/exchange"

	"github.com/gogf/gf/v2/frame/g"
)

// DirectionSignalService 方向信号服务（单例）
type DirectionSignalService struct {
	mu sync.RWMutex

	// 信号缓存 key: platform:symbol
	signals map[string]*DirectionSignal

	// 运行状态
	running bool
	stopCh  chan struct{}
}

// DirectionSignal 方向信号（精简版）
type DirectionSignal struct {
	Platform  string
	Symbol    string
	UpdatedAt time.Time

	// 信号方向
	Direction Direction
	Strength  float64 // 信号强度 0-100

	// 信号置信度
	Confidence float64 // 0-100

	// 趋势信号
	TrendSignal Direction

	// 各周期信号
	TimeframeSignals  map[string]Direction
	AlignedTimeframes int // 多周期一致的数量

	// 建议操作
	Action     SignalAction
	EntryPrice float64 // 建议入场价
	StopLoss   float64 // 建议止损价

	// 信号原因
	Reason string
}

// Direction 方向枚举
type Direction string

const (
	DirectionLong    Direction = "LONG"
	DirectionShort   Direction = "SHORT"
	DirectionNeutral Direction = "NEUTRAL"
)

// SignalAction 信号动作
type SignalAction string

const (
	ActionOpenLong   SignalAction = "OPEN_LONG"
	ActionOpenShort  SignalAction = "OPEN_SHORT"
	ActionCloseLong  SignalAction = "CLOSE_LONG"
	ActionCloseShort SignalAction = "CLOSE_SHORT"
	ActionHold       SignalAction = "HOLD"
	ActionWait       SignalAction = "WAIT"
)

// StrategyConfig 策略配置
type StrategyConfig struct {
	MonitorWindow       int     // 监控窗口（秒）
	VolatilityThreshold float64 // 波动点数阈值
	ConfidenceThreshold float64 // 置信度阈值
	EnableReverseOrder  bool    // 启用反向下单
}

var (
	directionSignalService     *DirectionSignalService
	directionSignalServiceOnce sync.Once
)

// GetDirectionSignalService 获取方向信号服务单例
func GetDirectionSignalService() *DirectionSignalService {
	directionSignalServiceOnce.Do(func() {
		directionSignalService = &DirectionSignalService{
			signals: make(map[string]*DirectionSignal),
			stopCh:  make(chan struct{}),
		}
	})
	return directionSignalService
}

// Start 启动方向信号服务
func (s *DirectionSignalService) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	g.Log().Info(ctx, "[DirectionSignalService] 方向信号服务启动")
	go s.runSignalLoop(ctx)
}

// Stop 停止服务
func (s *DirectionSignalService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}
	s.running = false
	close(s.stopCh)
}

// runSignalLoop 信号生成循环
func (s *DirectionSignalService) runSignalLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.updateAllSignals(ctx)
		}
	}
}

// updateAllSignals 更新所有信号
func (s *DirectionSignalService) updateAllSignals(ctx context.Context) {
	msm := GetMarketServiceManager()

	allServices := msm.GetAllServices()
	for platform, svc := range allServices {
		subscriptions := svc.GetAllSubscriptions()
		for symbol := range subscriptions {
			analysis := GetMarketAnalyzer().GetAnalysis(platform, symbol)
			ticker := svc.GetTicker(symbol)
			klineCache := svc.GetMultiTimeframeKlines(symbol)

			if analysis == nil || ticker == nil || klineCache == nil {
				continue
			}

			signal := s.generateSignal(ctx, platform, symbol, ticker, analysis, klineCache, nil)
			if signal != nil {
				key := platform + ":" + symbol
				s.mu.Lock()
				s.signals[key] = signal
				s.mu.Unlock()
			}
		}
	}
}

// GenerateSignalForRobot 为特定机器人生成信号
func (s *DirectionSignalService) GenerateSignalForRobot(ctx context.Context, platform, symbol string, strategyConfig *StrategyConfig) *DirectionSignal {
	msm := GetMarketServiceManager()
	svc := msm.GetService(platform)
	if svc == nil {
		return nil
	}

	analysis := GetMarketAnalyzer().GetAnalysis(platform, symbol)
	ticker := svc.GetTicker(symbol)
	klineCache := svc.GetMultiTimeframeKlines(symbol)

	if analysis == nil || ticker == nil || klineCache == nil {
		return nil
	}

	return s.generateSignal(ctx, platform, symbol, ticker, analysis, klineCache, strategyConfig)
}

// generateSignal 生成方向信号（精简版）
func (s *DirectionSignalService) generateSignal(ctx context.Context, platform, symbol string, ticker *exchange.Ticker, analysis *MarketAnalysis, klineCache *KlineCache, strategyConfig *StrategyConfig) *DirectionSignal {
	signal := &DirectionSignal{
		Platform:         platform,
		Symbol:           symbol,
		UpdatedAt:        time.Now(),
		TimeframeSignals: make(map[string]Direction),
	}

	// 1. 各周期信号分析
	for interval, tf := range analysis.TimeframeAnalysis {
		signal.TimeframeSignals[interval] = s.analyzeTimeframeDirection(tf)
	}

	// 2. 趋势信号
	signal.TrendSignal = s.analyzeTrendSignal(analysis)

	// 3. 综合判断方向
	signal.Direction, signal.Strength, signal.Confidence = s.synthesizeDirection(signal)

	// 4. 应用策略配置（如果有）
	if strategyConfig != nil {
		s.applyStrategyConfig(signal, ticker, klineCache, strategyConfig)
	}

	// 5. 生成建议操作
	signal.Action = s.determineAction(signal, analysis)

	// 6. 计算入场价和止损
	s.calculatePriceTargets(signal, ticker, analysis)

	// 7. 生成信号原因
	signal.Reason = s.generateSignalReason(signal)

	return signal
}

// analyzeTimeframeDirection 分析单周期方向（精简：只看MACD）
func (s *DirectionSignalService) analyzeTimeframeDirection(tf *TimeframeResult) Direction {
	// 简化判断：EMA12 > EMA26 且 MACD > 0 → 做多
	if tf.EMA12 > tf.EMA26 && tf.MACD > 0 {
		return DirectionLong
	}
	if tf.EMA12 < tf.EMA26 && tf.MACD < 0 {
		return DirectionShort
	}
	return DirectionNeutral
}

// analyzeTrendSignal 分析趋势信号
func (s *DirectionSignalService) analyzeTrendSignal(analysis *MarketAnalysis) Direction {
	if analysis.Indicators == nil {
		return DirectionNeutral
	}

	if analysis.Indicators.TrendScore > 30 {
		return DirectionLong
	} else if analysis.Indicators.TrendScore < -30 {
		return DirectionShort
	}
	return DirectionNeutral
}

// synthesizeDirection 综合判断方向（精简版）
func (s *DirectionSignalService) synthesizeDirection(signal *DirectionSignal) (Direction, float64, float64) {
	var longCount, shortCount int

	// 统计各周期信号
	for _, dir := range signal.TimeframeSignals {
		switch dir {
		case DirectionLong:
			longCount++
		case DirectionShort:
			shortCount++
		}
	}

	// 趋势信号加权
	if signal.TrendSignal == DirectionLong {
		longCount += 2
	} else if signal.TrendSignal == DirectionShort {
		shortCount += 2
	}

	total := len(signal.TimeframeSignals) + 2 // 周期数 + 趋势信号权重

	// 计算信号强度和置信度
	var direction Direction
	var strength, confidence float64

	diff := longCount - shortCount
	if diff > 2 {
		direction = DirectionLong
		strength = float64(diff) / float64(total) * 100
		confidence = float64(longCount) / float64(total) * 100
		signal.AlignedTimeframes = longCount
	} else if diff < -2 {
		direction = DirectionShort
		strength = float64(-diff) / float64(total) * 100
		confidence = float64(shortCount) / float64(total) * 100
		signal.AlignedTimeframes = shortCount
	} else {
		direction = DirectionNeutral
		strength = 30
		confidence = 30
	}

	return direction, strength, confidence
}

// applyStrategyConfig 应用策略配置
func (s *DirectionSignalService) applyStrategyConfig(signal *DirectionSignal, ticker *exchange.Ticker, klineCache *KlineCache, config *StrategyConfig) {
	if config == nil {
		return
	}

	monitorWindow := config.MonitorWindow
	volatilityThreshold := config.VolatilityThreshold

	// 获取对应周期的K线
	var klines []*exchange.Kline
	switch {
	case monitorWindow <= 300:
		klines = klineCache.Klines5m
	case monitorWindow <= 900:
		klines = klineCache.Klines15m
	default:
		klines = klineCache.Klines1h
	}

	if len(klines) < 5 {
		return
	}

	// 计算窗口内最高最低价
	klineCount := monitorWindow / 60
	if klineCount > len(klines) {
		klineCount = len(klines)
	}
	if klineCount < 5 {
		klineCount = 5
	}

	windowHigh := 0.0
	windowLow := math.MaxFloat64
	for i := len(klines) - klineCount; i < len(klines); i++ {
		if klines[i].High > windowHigh {
			windowHigh = klines[i].High
		}
		if klines[i].Low < windowLow {
			windowLow = klines[i].Low
		}
	}

	currentPrice := ticker.LastPrice

	// 策略方向判断
	shortTrigger := windowHigh - volatilityThreshold
	longTrigger := windowLow + volatilityThreshold

	if currentPrice <= shortTrigger {
		signal.Direction = DirectionShort
		signal.Strength = math.Min(100, (windowHigh-currentPrice)/volatilityThreshold*100)
		signal.Reason = "窗口做空信号：价格从高点回落"
	} else if currentPrice >= longTrigger {
		signal.Direction = DirectionLong
		signal.Strength = math.Min(100, (currentPrice-windowLow)/volatilityThreshold*100)
		signal.Reason = "窗口做多信号：价格从低点反弹"
	}

	// 置信度阈值过滤
	if config.ConfidenceThreshold > 0 && signal.Confidence < config.ConfidenceThreshold {
		signal.Direction = DirectionNeutral
		signal.Action = ActionWait
		signal.Reason = "信号置信度未达阈值"
	}
}

// determineAction 确定建议操作
func (s *DirectionSignalService) determineAction(signal *DirectionSignal, analysis *MarketAnalysis) SignalAction {
	// 高波动市场谨慎操作
	if analysis.MarketState == MarketStateHighVol && signal.Confidence < 70 {
		return ActionWait
	}

	// 信号强度不足
	if signal.Strength < 30 {
		return ActionHold
	}

	// 置信度不足
	if signal.Confidence < 50 {
		return ActionWait
	}

	switch signal.Direction {
	case DirectionLong:
		return ActionOpenLong
	case DirectionShort:
		return ActionOpenShort
	default:
		return ActionHold
	}
}

// calculatePriceTargets 计算价格目标
func (s *DirectionSignalService) calculatePriceTargets(signal *DirectionSignal, ticker *exchange.Ticker, analysis *MarketAnalysis) {
	currentPrice := ticker.LastPrice
	atr := analysis.Volatility

	if atr <= 0 {
		atr = currentPrice * 0.01 // 默认1%
	}

	signal.EntryPrice = currentPrice

	switch signal.Direction {
	case DirectionLong:
		signal.StopLoss = currentPrice - atr*2
	case DirectionShort:
		signal.StopLoss = currentPrice + atr*2
	}
}

// generateSignalReason 生成信号原因
func (s *DirectionSignalService) generateSignalReason(signal *DirectionSignal) string {
	if signal.Reason != "" {
		return signal.Reason
	}

	switch signal.Direction {
	case DirectionLong:
		return "多周期指标看多"
	case DirectionShort:
		return "多周期指标看空"
	default:
		return "信号不明确，观望"
	}
}

// GetSignal 获取方向信号
func (s *DirectionSignalService) GetSignal(platform, symbol string) *DirectionSignal {
	key := platform + ":" + symbol

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.signals[key]
}

// GetAllSignals 获取所有信号
func (s *DirectionSignalService) GetAllSignals() []*DirectionSignal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*DirectionSignal, 0, len(s.signals))
	for _, signal := range s.signals {
		result = append(result, signal)
	}
	return result
}
