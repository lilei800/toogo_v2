// Package market 新算法实现
package market

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	configlib "hotgo/internal/library/config"
	"hotgo/internal/library/exchange"
	"math"
	"time"
)

// MarketStateThresholds 定义每个周期的状态阈值
type MarketStateThresholds struct {
	LowV       float64 // V小于LowV -> 低波动
	HighV      float64 // V大于HighV且D小 -> 高波动
	TrendV     float64 // V大于TrendV且D大 -> 趋势
	DThreshold float64 // 趋势判断时方向一致性阈值
}

// DetectMarketStateSingle 根据单根K线计算市场状态（新算法）
func DetectMarketStateSingle(O, H, L, P, delta float64, thresh MarketStateThresholds) string {
	// 1. 计算波动强度 V = (最高价-最低价)/delta
	if delta <= 0 {
		return "volatile" // delta无效，返回默认值
	}
	V := (H - L) / delta

	// 2. 计算方向一致性 D
	var D float64
	priceRange := H - L
	if priceRange <= 0 {
		return "volatile" // 价格范围无效
	}

	if P >= O {
		// 上涨：收盘价接近最高价的程度
		D = (P - L) / priceRange
	} else {
		// 下跌：收盘价接近最低价的程度
		D = (H - P) / priceRange
	}

	// 3. 根据阈值判定市场状态
	switch {
	case V < thresh.LowV:
		return "low_vol" // 低波动
	case V >= thresh.HighV && D < 0.4:
		return "high_vol" // 高波动
	case V >= thresh.TrendV && D >= thresh.DThreshold:
		return "trend" // 趋势
	default:
		return "volatile" // 震荡
	}
}

// DetectMarketStateMultiCycle 多周期计算市场状态（新算法）
func DetectMarketStateMultiCycle(
	O, H, L, P, delta []float64,
	thresh []MarketStateThresholds,
	weight []float64,
) string {
	// 参数长度检查
	if len(O) != len(H) || len(H) != len(L) || len(L) != len(P) ||
		len(P) != len(delta) || len(delta) != len(thresh) || len(thresh) != len(weight) {
		return "volatile" // 参数不一致，返回默认值
	}

	// 初始化投票表
	votes := map[string]float64{
		"low_vol":  0,
		"volatile": 0,
		"high_vol": 0,
		"trend":    0,
	}

	// 遍历每个周期，计算单周期市场状态
	for i := 0; i < len(O); i++ {
		state := DetectMarketStateSingle(O[i], H[i], L[i], P[i], delta[i], thresh[i])
		votes[state] += weight[i] // 按权重投票
	}

	// 选择权重最大状态作为最终市场状态
	finalState := "volatile"
	maxVote := float64(0)
	for k, v := range votes {
		if v > maxVote {
			maxVote = v
			finalState = k
		}
	}

	return finalState
}

// SmoothState 旧版平滑函数（保留用于兼容/参考）
// 注意：新版“新算法 + 平滑机制”不再使用该函数，而是使用按 platform:symbol 的滑动窗口平滑器（见 state_smoother.go）
func SmoothState(currentState string, history []string, threshold int) string {
	if len(history) == 0 {
		return currentState
	}
	stateCount := make(map[string]int)
	for _, state := range history {
		stateCount[state]++
	}
	for state, count := range stateCount {
		if count >= threshold {
			return state
		}
	}
	return currentState
}

// AnalyzeMarketWithNewAlgorithm 使用新算法分析市场（替换旧算法）
func (a *MarketAnalyzer) AnalyzeMarketWithNewAlgorithm(
	ctx context.Context,
	platform, symbol string,
	ticker *exchange.Ticker,
	klineCache *KlineCache,
	previousAnalysis *MarketAnalysis,
) *MarketAnalysis {
	analysis := &MarketAnalysis{
		Platform:          platform,
		Symbol:            symbol,
		UpdatedAt:         time.Now(),
		CurrentPrice:      ticker.LastPrice,
		TimeframeAnalysis: make(map[string]*TimeframeResult),
	}

	// 1. 获取波动率配置（从全局配置管理器）
	volatilityConfig := configlib.GetVolatilityConfigManager().GetConfig(symbol)
	if volatilityConfig == nil {
		g.Log().Warningf(ctx, "[MarketAnalyzer] 获取波动率配置失败: symbol=%s，使用默认值", symbol)
		// 使用默认配置
		volatilityConfig = &configlib.VolatilityConfig{
			HighVolatilityThreshold: 2.0,
			LowVolatilityThreshold:  1.0,
			TrendStrengthThreshold:  1.2,
			DThreshold:              0.7,
			Delta1m:                 2.0,
			Delta5m:                 2.0,
			Delta15m:                3.0,
			Delta30m:                3.0,
			Delta1h:                 5.0,
			Weight1m:                0.20,
			Weight5m:                0.25,
			Weight15m:               0.25,
			Weight30m:               0.20,
			Weight1h:                0.10,
		}
	}

	// 2. 多周期配置（使用配置中的权重）
	timeframes := []struct {
		interval  string
		klines    []*exchange.Kline
		weight    float64
		delta     float64
		threshold MarketStateThresholds
	}{
		{
			"1m", klineCache.Klines1m, volatilityConfig.Weight1m, volatilityConfig.Delta1m,
			MarketStateThresholds{
				LowV:       volatilityConfig.LowVolatilityThreshold,
				HighV:      volatilityConfig.HighVolatilityThreshold,
				TrendV:     volatilityConfig.TrendStrengthThreshold,
				DThreshold: volatilityConfig.DThreshold,
			},
		},
		{
			"5m", klineCache.Klines5m, volatilityConfig.Weight5m, volatilityConfig.Delta5m,
			MarketStateThresholds{
				LowV:       volatilityConfig.LowVolatilityThreshold,
				HighV:      volatilityConfig.HighVolatilityThreshold,
				TrendV:     volatilityConfig.TrendStrengthThreshold,
				DThreshold: volatilityConfig.DThreshold,
			},
		},
		{
			"15m", klineCache.Klines15m, volatilityConfig.Weight15m, volatilityConfig.Delta15m,
			MarketStateThresholds{
				LowV:       volatilityConfig.LowVolatilityThreshold,
				HighV:      volatilityConfig.HighVolatilityThreshold,
				TrendV:     volatilityConfig.TrendStrengthThreshold,
				DThreshold: volatilityConfig.DThreshold,
			},
		},
		{
			"30m", klineCache.Klines30m, volatilityConfig.Weight30m, volatilityConfig.Delta30m,
			MarketStateThresholds{
				LowV:       volatilityConfig.LowVolatilityThreshold,
				HighV:      volatilityConfig.HighVolatilityThreshold,
				TrendV:     volatilityConfig.TrendStrengthThreshold,
				DThreshold: volatilityConfig.DThreshold,
			},
		},
		{
			"1h", klineCache.Klines1h, volatilityConfig.Weight1h, volatilityConfig.Delta1h,
			MarketStateThresholds{
				LowV:       volatilityConfig.LowVolatilityThreshold,
				HighV:      volatilityConfig.HighVolatilityThreshold,
				TrendV:     volatilityConfig.TrendStrengthThreshold,
				DThreshold: volatilityConfig.DThreshold,
			},
		},
	}

	// 3. 准备数据（单根K线） + 平滑机制（按 platform:symbol 的时间序列）
	key := platform + ":" + symbol
	smoother := a.getOrCreateStateSmoother(key)

	// 平滑参数（默认值：不改数据库/页面即可生效）
	// - timeframeWindow: 单周期状态平滑窗口（1秒一轮，5表示约5秒）
	// - finalWindow: 最终状态平滑窗口
	// - minRatio: 多数占比阈值，>=60% 才切换到多数态
	const timeframeWindow = 5
	const finalWindow = 7
	const minRatio = 0.6

	var O, H, L, P, delta []float64
	var thresh []MarketStateThresholds
	var weight []float64

	// 预先构建单周期结果（用于调试/兼容）并将“平滑后的状态”作为投票输入
	analysis.TimeframeAnalysis = make(map[string]*TimeframeResult)
	votes := map[string]float64{
		"low_vol":  0,
		"volatile": 0,
		"high_vol": 0,
		"trend":    0,
	}
	var totalWeight float64

	for _, tf := range timeframes {
		if len(tf.klines) < 1 {
			continue // 至少需要1根K线
		}

		// 获取最新一根K线
		latestKline := tf.klines[len(tf.klines)-1]

		O = append(O, latestKline.Open)
		H = append(H, latestKline.High)
		L = append(L, latestKline.Low)
		P = append(P, latestKline.Close)
		delta = append(delta, tf.delta)
		thresh = append(thresh, tf.threshold)
		weight = append(weight, tf.weight)

		// 单周期：先算 raw，再做时间序列平滑（每个周期独立平滑）
		rawState := DetectMarketStateSingle(latestKline.Open, latestKline.High, latestKline.Low, latestKline.Close, tf.delta, tf.threshold)
		smoothedState, smoothedConf := smoother.pushAndSmoothTimeframe(tf.interval, rawState, timeframeWindow, minRatio)

		votes[smoothedState] += tf.weight
		totalWeight += tf.weight

		// 写入单周期结果（使用“平滑后的状态”）
		var tfMarketState MarketState
		switch smoothedState {
		case "low_vol":
			tfMarketState = MarketStateLowVol
		case "high_vol":
			tfMarketState = MarketStateHighVol
		case "trend":
			tfMarketState = MarketStateTrend
		default:
			tfMarketState = MarketStateVolatile
		}

		priceRange := latestKline.High - latestKline.Low
		var vVal float64
		var priceVol float64
		if tf.delta > 0 {
			vVal = priceRange / tf.delta
			priceVol = vVal * 100
		}
		var priceChangeRate float64
		if latestKline.Open != 0 {
			priceChangeRate = ((latestKline.Close - latestKline.Open) / latestKline.Open) * 100
		}

		// 方向一致性 D（0-1）
		var dVal float64
		if priceRange > 0 {
			if latestKline.Close >= latestKline.Open {
				dVal = (latestKline.Close - latestKline.Low) / priceRange
			} else {
				dVal = (latestKline.High - latestKline.Close) / priceRange
			}
		}

		analysis.TimeframeAnalysis[tf.interval] = &TimeframeResult{
			Interval:        tf.interval,
			Weight:          tf.weight,
			PriceVolatility: priceVol,
			PriceChangeRate: priceChangeRate,
			MarketState:     tfMarketState,

			Open:  latestKline.Open,
			High:  latestKline.High,
			Low:   latestKline.Low,
			Close: latestKline.Close,

			Delta: tf.delta,
			V:     vVal,
			D:     dVal,

			RawState:      rawState,
			SmoothedState: smoothedState,
			SmoothedConf:  smoothedConf,
		}
	}

	if len(O) == 0 {
		g.Log().Warningf(ctx, "[MarketAnalyzer] 数据不足: symbol=%s", symbol)
		return nil
	}

	// 4. 多周期融合：使用“平滑后的单周期状态”做加权投票
	finalStateRaw := "volatile"
	maxVote := 0.0
	for st, v := range votes {
		if v > maxVote {
			maxVote = v
			finalStateRaw = st
		}
	}

	// 投票置信度：最大投票权重占比
	voteConf := 0.0
	if totalWeight > 0 {
		voteConf = maxVote / totalWeight
	}

	// 5. 最终状态再做一次时间序列平滑（抗抖）
	finalState, smoothConf := smoother.pushAndSmoothFinal(finalStateRaw, finalWindow, minRatio)

	// 6. 映射到MarketState类型
	var marketState MarketState
	switch finalState {
	case "low_vol":
		marketState = MarketStateLowVol
	case "high_vol":
		marketState = MarketStateHighVol
	case "trend":
		marketState = MarketStateTrend
	default:
		marketState = MarketStateVolatile
	}

	analysis.MarketState = marketState

	// 保存投票占比（用于前端显示）
	analysis.VoteRatio = voteConf

	// 置信度综合：取“投票占比”和“时间序列稳定度”的较大者，并做合理区间限制
	conf := voteConf
	if smoothConf > conf {
		conf = smoothConf
	}
	// 下限：避免前端显示过低；上限：避免误导
	conf = math.Max(0.55, math.Min(0.95, conf))
	analysis.MarketStateConf = conf

	// 7. 计算波动率（用于兼容）
	var totalVolatility float64
	var totalWeightForVol float64
	for i := range O {
		if i < len(delta) && delta[i] > 0 {
			V := (H[i] - L[i]) / delta[i]
			totalVolatility += V * weight[i]
			totalWeightForVol += weight[i]
		}
	}
	if totalWeightForVol > 0 {
		analysis.Volatility = totalVolatility / totalWeightForVol * 100 // 转换为百分比
	}

	// 8. 计算趋势强度（用于兼容）
	var totalTrendStrength float64
	for i := range O {
		if P[i] >= O[i] {
			// 上涨：计算方向一致性
			priceRange := H[i] - L[i]
			if priceRange > 0 {
				D := (P[i] - L[i]) / priceRange
				totalTrendStrength += D * weight[i] // 使用D作为趋势强度
			}
		} else {
			// 下跌：计算方向一致性
			priceRange := H[i] - L[i]
			if priceRange > 0 {
				D := (H[i] - P[i]) / priceRange
				totalTrendStrength -= D * weight[i] // 下跌为负
			}
		}
	}
	analysis.TrendStrength = totalTrendStrength

	// 9. 更新调整阈值（用于兼容）
	analysis.AdjustedHighThreshold = volatilityConfig.HighVolatilityThreshold
	analysis.AdjustedLowThreshold = volatilityConfig.LowVolatilityThreshold

	// 10. 计算综合技术指标（用于兼容）
	analysis.Indicators = &TechnicalIndicators{
		TrendScore:      analysis.TrendStrength * 100,
		VolatilityScore: analysis.Volatility,
	}

	return analysis
}
