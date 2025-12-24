// Package market 市场分析引擎（精简版）
// 只保留核心功能：趋势方向、波动率、多周期一致性
// 集成新的计算逻辑到现有的MarketAnalyzer
package market

import (
	"context"
	"hotgo/utility/simple"
	"sync"
	"time"

	"hotgo/internal/library/exchange"

	"github.com/gogf/gf/v2/frame/g"
)

// ============ 类型定义 ============

// MarketAnalyzer 市场分析引擎（单例）
type MarketAnalyzer struct {
	// 【性能优化】使用sync.Map替代map+mutex，减少锁竞争，提高并发性能
	// sync.Map适合读多写少的场景，无需加锁即可读取
	analysisCache sync.Map // key: platform:symbol, value: *MarketAnalysis

	// 【平滑机制】市场状态时间序列平滑缓存（按 platform:symbol）
	// 由于分析是并发执行的，每个key内部使用自己的mutex保护历史更新
	stateSmoothers sync.Map // key: platform:symbol, value: *marketStateSmoother

	// 运行状态（需要mutex保护）
	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}
}

// MarketAnalysis 市场分析结果（精简版）
type MarketAnalysis struct {
	Platform  string
	Symbol    string
	UpdatedAt time.Time

	// 当前价格
	CurrentPrice float64

	// 市场状态
	MarketState     MarketState
	MarketStateConf float64 // 置信度 0-1（综合投票占比和时间序列稳定度）
	VoteRatio       float64 // 投票占比 0-1（最大投票权重/总权重）

	// 多周期分析结果
	TimeframeAnalysis map[string]*TimeframeResult

	// 综合技术指标
	Indicators *TechnicalIndicators

	// 趋势强度 -1到1 (-1强空, 0中性, 1强多)
	TrendStrength float64

	// 波动率（相对波动率，百分比）
	Volatility float64

	// 调整后的阈值（基于市场状态优化）
	AdjustedHighThreshold float64 // 调整后的高波动阈值
	AdjustedLowThreshold  float64 // 调整后的低波动阈值

	// 支撑阻力位
	SupportLevel    float64
	ResistanceLevel float64
}

// MarketState 市场状态枚举
type MarketState string

const (
	MarketStateTrend    MarketState = "trend"    // 趋势市场
	MarketStateVolatile MarketState = "range"    // 震荡市场（兼容旧代码，值为"range"）
	MarketStateHighVol  MarketState = "high_vol" // 高波动
	MarketStateLowVol   MarketState = "low_vol"  // 低波动
)

// TimeframeResult 单周期分析结果（精简版）
type TimeframeResult struct {
	Interval string
	Weight   float64 // 权重

	// 趋势判断
	Trend         string  // up/down/sideways
	TrendStrength float64 // 0-1

	// 【实时性优化】移除技术指标，直接使用价格数据计算市场状态
	// 技术指标（EMA/MACD）有滞后性，直接使用价格数据更实时、更精准
	// 保留字段用于兼容性，但不再计算和使用
	EMA8  float64 // 已废弃，不再使用
	EMA12 float64 // 已废弃，不再使用
	EMA26 float64 // 已废弃，不再使用
	MACD  float64 // 已废弃，不再使用

	// 波动率（直接使用价格波动，不使用ATR）
	PriceVolatility   float64 // 价格波动率（百分比）：(最高价-最低价)/当前价*100
	PriceChangeRate   float64 // 价格变化率（百分比）：(当前价-起始价)/起始价*100
	PriceAcceleration float64 // 价格加速度（百分比）：价格变化率的变化率，捕捉趋势变化速度

	// 市场状态（单周期独立计算）
	MarketState MarketState // 该周期的市场状态（trend/volatile/high_vol/low_vol）

	// ============ 新算法实时细节（用于前端“播报/明细”展示）============
	// 最新一根K线价格
	Open  float64
	High  float64
	Low   float64
	Close float64

	// 配置与中间量
	Delta float64 // 当前周期使用的delta
	V     float64 // V = (H-L)/delta
	D     float64 // 方向一致性 D（0-1）

	// 平滑信息
	RawState      string  // 原始状态（未平滑）
	SmoothedState string  // 平滑后状态（参与投票）
	SmoothedConf  float64 // 平滑置信度（0-1）
}

// TechnicalIndicators 综合技术指标（精简版）
type TechnicalIndicators struct {
	TrendScore      float64 // 趋势综合评分 -100到100
	VolatilityScore float64 // 波动评分 0-100
}

// MarketStateVolatilities 不同市场状态的波动率特征值
// 【动态波动率优化】根据历史K线识别不同市场状态模式，计算每种状态的波动率特征值
type MarketStateVolatilities struct {
	TrendUpVolatility   float64 // 趋势向上状态的波动率
	TrendDownVolatility float64 // 趋势向下状态的波动率
	SpikeVolatility     float64 // 上下插针状态的波动率
	HighVolVolatility   float64 // 高波动状态的波动率
	LowVolVolatility    float64 // 低波动状态的波动率
	VolatileVolatility  float64 // 震荡状态的波动率
}

// ============ 单例 ============

var (
	marketAnalyzer     *MarketAnalyzer
	marketAnalyzerOnce sync.Once
)

// GetMarketAnalyzer 获取市场分析引擎单例
func GetMarketAnalyzer() *MarketAnalyzer {
	marketAnalyzerOnce.Do(func() {
		marketAnalyzer = &MarketAnalyzer{
			analysisCache:  sync.Map{}, // 【性能优化】使用sync.Map，无需初始化
			stateSmoothers: sync.Map{}, // 【平滑机制】状态平滑缓存
			stopCh:         make(chan struct{}),
		}
	})
	return marketAnalyzer
}

// ============ 基础方法 ============

// Start 启动分析引擎
func (a *MarketAnalyzer) Start(ctx context.Context) {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return
	}
	a.running = true
	a.mu.Unlock()

	g.Log().Info(ctx, "[MarketAnalyzer] 市场分析引擎启动")
	// 使用增强版分析循环（只分析有机器人的币种）
	simple.SafeGo(ctx, func(ctx context.Context) {
		a.RunAnalysisLoopEnhanced(ctx)
	})
}

// Stop 停止分析引擎
func (a *MarketAnalyzer) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return
	}

	a.running = false
	close(a.stopCh)
}

// GetAnalysis 获取市场分析结果
// 每个币种（platform+symbol）有独立的市场状态信号
// 例如：bitget:BTCUSDT 和 binance:BTCUSDT 返回不同的市场状态
// 【性能优化】使用sync.Map.Load，无需加锁，提高并发性能
func (a *MarketAnalyzer) GetAnalysis(platform, symbol string) *MarketAnalysis {
	if platform == "" || symbol == "" {
		return nil
	}

	// 使用 platform:symbol 作为唯一键
	key := platform + ":" + symbol

	// 【性能优化】使用sync.Map.Load，无需加锁
	if val, ok := a.analysisCache.Load(key); ok {
		analysis := val.(*MarketAnalysis)
		// 验证缓存的数据是否匹配请求的币种（确保数据一致性）
		if analysis.Platform == platform && analysis.Symbol == symbol {
			return analysis
		}
	}

	return nil
}

// GetAllAnalyses 获取所有分析结果
func (a *MarketAnalyzer) GetAllAnalyses() []*MarketAnalysis {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]*MarketAnalysis, 0)
	// 【性能优化】sync.Map需要使用Range方法遍历
	a.analysisCache.Range(func(key, value interface{}) bool {
		if analysis, ok := value.(*MarketAnalysis); ok {
			result = append(result, analysis)
		}
		return true // 继续遍历
	})
	return result
}

// ============ 兼容性方法 ============

// analyzeTimeframe 分析单个周期（兼容方法，实际使用增强版）
// 这个方法保留用于兼容性，但实际分析逻辑在 AnalyzeMarketEnhanced 中
func (a *MarketAnalyzer) analyzeTimeframe(interval string, klines []*exchange.Kline, weight float64) *TimeframeResult {
	// 这个方法在新版本中不再使用，保留用于兼容性
	// 实际分析逻辑在 AnalyzeMarketEnhanced 中实现
	result := &TimeframeResult{
		Interval: interval,
		Weight:   weight,
		Trend:    "sideways",
	}
	return result
}

// calculateTrendConsistency 计算多周期趋势一致性（兼容方法）
func (a *MarketAnalyzer) calculateTrendConsistency(tfResults map[string]*TimeframeResult) float64 {
	var upCount, downCount int

	for _, tf := range tfResults {
		switch tf.Trend {
		case "up":
			upCount++
		case "down":
			downCount++
		}
	}

	total := len(tfResults)
	if total == 0 {
		return 0
	}

	maxCount := upCount
	if downCount > upCount {
		maxCount = downCount
	}

	return float64(maxCount) / float64(total)
}

// ============ 增强版方法 ============

// AnalyzeMarketEnhanced 增强版市场分析（使用新算法替换旧算法）
func (a *MarketAnalyzer) AnalyzeMarketEnhanced(
	ctx context.Context,
	platform, symbol string,
	ticker *exchange.Ticker,
	klineCache *KlineCache,
	previousAnalysis *MarketAnalysis,
) *MarketAnalysis {
	// 直接使用新算法替换旧算法
	return a.AnalyzeMarketWithNewAlgorithm(ctx, platform, symbol, ticker, klineCache, previousAnalysis)
}

// mapToCompatibleMarketState 映射到兼容的市场状态格式
// 确保新算法返回的状态值能被现有代码正确处理
// 注意：旧代码使用 "range" 表示震荡市场，但 normalizeMarketState 会将其转换为 "volatile"
func mapToCompatibleMarketState(state string) MarketState {
	switch state {
	case "trend":
		return MarketStateTrend
	case "volatile":
		// 新算法返回 "volatile"，但为了兼容旧代码，映射到 MarketStateVolatile
		// MarketStateVolatile 的值是 "range"，但 normalizeMarketState 会将其转换为 "volatile"
		return MarketStateVolatile // 值为 "range"，但会被 normalizeMarketState 转换为 "volatile"
	case "high_vol":
		return MarketStateHighVol
	case "low_vol":
		return MarketStateLowVol
	default:
		// 未知状态，返回默认值
		return MarketStateTrend
	}
}

// AnalyzeAllMarketsEnhanced 增强版分析所有市场（只分析有机器人的币种）
func (a *MarketAnalyzer) AnalyzeAllMarketsEnhanced(ctx context.Context) {
	// 1. 获取所有运行中机器人使用的币种
	activeSymbols, err := GetActiveSymbols(ctx)
	if err != nil {
		g.Log().Errorf(ctx, "[MarketAnalyzer] 获取活跃币种失败: %v", err)
		return
	}

	if len(activeSymbols) == 0 {
		// 没有运行中的机器人，跳过分析
		g.Log().Debugf(ctx, "[MarketAnalyzer] 没有运行中的机器人，跳过市场分析")
		return
	}

	msm := GetMarketServiceManager()
	allServices := msm.GetAllServices()

	// 2. 只分析活跃币种
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // 限制并发数

	for platform, svc := range allServices {
		symbols := activeSymbols[platform]
		if symbols == nil {
			continue // 该平台没有活跃币种
		}

		for symbol := range symbols {
			// 检查是否有K线数据
			klineCache := svc.GetMultiTimeframeKlines(symbol)
			ticker := svc.GetTicker(symbol)

			if klineCache == nil || ticker == nil {
				continue
			}

			wg.Add(1)
			semaphore <- struct{}{}

			p := platform
			s := symbol
			simple.SafeGo(ctx, func(ctx context.Context) {
				defer wg.Done()
				defer func() { <-semaphore }()

				// 获取上一次的分析结果
				key := p + ":" + s
				var previousAnalysis *MarketAnalysis
				if val, ok := a.analysisCache.Load(key); ok {
					if pa, ok2 := val.(*MarketAnalysis); ok2 {
						previousAnalysis = pa
					} else {
						// 类型不匹配，清理异常缓存
						a.analysisCache.Delete(key)
					}
				}

				// 使用增强版分析
				analysis := a.AnalyzeMarketEnhanced(ctx, p, s, ticker, klineCache, previousAnalysis)
				if analysis != nil {
					a.analysisCache.Store(key, analysis)
				}
			})
		}
	}

	wg.Wait()
}

// RunAnalysisLoopEnhanced 增强版分析循环
func (a *MarketAnalyzer) RunAnalysisLoopEnhanced(ctx context.Context) {
	// 【实时性优化】降低分析频率到1秒，满足超短线交易对实时性的需求
	ticker := time.NewTicker(1 * time.Second)
	cleanupTicker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	defer cleanupTicker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			// 使用增强版分析（只分析有机器人的币种）
			a.AnalyzeAllMarketsEnhanced(ctx)
		case <-cleanupTicker.C:
			// 定期清理不再活跃/过期的缓存，防止长期运行只增不减
			activeSymbols, err := GetActiveSymbols(ctx)
			if err != nil {
				g.Log().Warningf(ctx, "[MarketAnalyzer] 获取活跃币种失败(用于清理缓存): %v", err)
			}
			removed := a.cleanupAnalysisCache(ctx, activeSymbols, 10*time.Minute)
			if removed > 0 {
				g.Log().Infof(ctx, "[MarketAnalyzer] 已清理分析缓存条目数=%d", removed)
			}
		}
	}
}

// cleanupAnalysisCache 清理分析缓存：
// - 非活跃机器人使用的币种（activeSymbols中不存在）
// - 或者超过maxAge的过期数据
// 返回本次清理删除的条目数
func (a *MarketAnalyzer) cleanupAnalysisCache(ctx context.Context, activeSymbols map[string]map[string]bool, maxAge time.Duration) (removed int) {
	now := time.Now()
	a.analysisCache.Range(func(key, value interface{}) bool {
		analysis, ok := value.(*MarketAnalysis)
		if !ok || analysis == nil {
			a.analysisCache.Delete(key)
			// 同步清理平滑缓存
			a.stateSmoothers.Delete(key)
			removed++
			return true
		}

		// 过期清理
		if maxAge > 0 && now.Sub(analysis.UpdatedAt) > maxAge {
			a.analysisCache.Delete(key)
			// 同步清理平滑缓存
			a.stateSmoothers.Delete(key)
			removed++
			return true
		}

		// 非活跃清理（只有在activeSymbols可用时才执行）
		if activeSymbols != nil {
			symbols := activeSymbols[analysis.Platform]
			if symbols == nil || !symbols[analysis.Symbol] {
				a.analysisCache.Delete(key)
				// 同步清理平滑缓存
				a.stateSmoothers.Delete(key)
				removed++
				return true
			}
		}
		return true
	})
	return removed
}

// ============ 兼容性方法（用于 market_state_pattern.go）============

// calculateMultiTimeframeBaselineVolatility 计算多周期基准波动率（兼容方法）
// 这个方法用于 market_state_pattern.go 的兼容性
func (a *MarketAnalyzer) calculateMultiTimeframeBaselineVolatility(klineCache *KlineCache) float64 {
	timeframes := []struct {
		klines    []*exchange.Kline
		weight    float64
		minLen    int
		maxKlines int
	}{
		{klineCache.Klines1m, 0.30, 12, 20},
		{klineCache.Klines5m, 0.40, 25, 40},
		{klineCache.Klines15m, 0.20, 15, 25},
		{klineCache.Klines30m, 0.07, 8, 12},
		{klineCache.Klines1h, 0.03, 5, 10},
	}

	var totalBaseline float64
	var totalWeight float64

	for _, tf := range timeframes {
		if len(tf.klines) < tf.minLen {
			continue
		}

		var klinesToUse []*exchange.Kline
		if len(tf.klines) > tf.maxKlines {
			klinesToUse = tf.klines[len(tf.klines)-tf.maxKlines:]
		} else {
			klinesToUse = tf.klines
		}

		baseline := calculateBaselineVolatilityForPattern(klinesToUse)
		if baseline > 0 {
			totalBaseline += baseline * tf.weight
			totalWeight += tf.weight
		}
	}

	if totalWeight == 0 {
		return 0.8 // 默认基准波动率：0.8%（基于BTCUSDT正常市场）
	}

	return totalBaseline / totalWeight
}

// calculateBaselineVolatilityForPattern 计算单周期基准波动率（用于模式识别）
func calculateBaselineVolatilityForPattern(klines []*exchange.Kline) float64 {
	if len(klines) < 20 {
		return 0
	}

	// 使用最近的数据计算基准波动率
	windowSize := 20
	if len(klines) > 200 {
		klines = klines[len(klines)-200:]
	}

	var volatilities []float64
	for i := windowSize; i < len(klines); i += windowSize {
		endIdx := i
		if endIdx > len(klines) {
			endIdx = len(klines)
		}
		windowKlines := klines[i-windowSize : endIdx]
		vol := calculateWindowVolatility(windowKlines)
		if vol > 0 {
			volatilities = append(volatilities, vol)
		}
	}

	if len(volatilities) == 0 {
		return 0
	}

	var sum float64
	for _, v := range volatilities {
		sum += v
	}

	return sum / float64(len(volatilities))
}

// calculateWindowVolatility 计算窗口波动率（用于模式识别）
// 注意：此函数已在 market_analyzer_v2.go 中定义，这里只是引用

// calculateDynamicThresholds 计算动态阈值（兼容方法）
// 这个方法用于 market_state_pattern.go 的兼容性
func (a *MarketAnalyzer) calculateDynamicThresholds(baselineVolatility float64) (highThreshold, lowThreshold float64) {
	// 如果基准波动率无效，使用默认值
	if baselineVolatility <= 0 || baselineVolatility > 30.0 {
		baselineVolatility = 0.8 // 默认基准：0.8%（BTCUSDT正常市场）
	}

	// 根据基准波动率动态调整倍数
	var highMultiplier, lowMultiplier float64
	if baselineVolatility < 0.5 {
		// 低波动市场：倍数增大，更敏感地捕捉波动变化
		highMultiplier = 2.0
		lowMultiplier = 0.4
	} else if baselineVolatility > 2.0 {
		// 高波动市场：倍数减小，避免过度敏感
		highMultiplier = 1.2
		lowMultiplier = 0.6
	} else {
		// 正常市场（0.5% - 2.0%）：使用标准倍数
		highMultiplier = 1.5
		lowMultiplier = 0.5
	}

	// 动态阈值计算（单位：百分比）
	highThreshold = baselineVolatility * highMultiplier
	lowThreshold = baselineVolatility * lowMultiplier

	// 确保阈值在合理范围内
	if highThreshold < 0.6 {
		highThreshold = 0.6
	}
	if lowThreshold < 0.2 {
		lowThreshold = 0.2
	}
	if highThreshold > 10.0 {
		highThreshold = 10.0
	}

	return highThreshold, lowThreshold
}
