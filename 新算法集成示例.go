package market

import (
	"math"
	"hotgo/internal/library/exchange"
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
	V := (H - L) / delta
	if delta <= 0 {
		return "震荡" // delta无效，返回默认值
	}

	// 2. 计算方向一致性 D
	var D float64
	priceRange := H - L
	if priceRange <= 0 {
		return "震荡" // 价格范围无效
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

// CalculateDynamicDelta 动态计算delta（改进版：自适应delta）
func CalculateDynamicDelta(klines []*exchange.Kline) float64 {
	if len(klines) < 5 {
		return 1.0 // 默认值
	}

	// 计算历史平均波动
	var totalVolatility float64
	for i := 1; i < len(klines); i++ {
		priceRange := klines[i].High - klines[i].Low
		if klines[i].Close > 0 {
			volatility := priceRange / klines[i].Close
			totalVolatility += volatility
		}
	}

	if len(klines) > 1 {
		avgVolatility := totalVolatility / float64(len(klines)-1)
		// delta = 平均波动率 * 当前价格
		if len(klines) > 0 && klines[len(klines)-1].Close > 0 {
			return avgVolatility * klines[len(klines)-1].Close
		}
	}

	return 1.0 // 默认值
}

// SmoothState 平滑状态（改进版：增加平滑机制）
func SmoothState(currentState string, history []string, threshold int) string {
	if len(history) == 0 {
		return currentState
	}

	// 统计最近N个状态
	stateCount := make(map[string]int)
	for _, state := range history {
		stateCount[state]++
	}

	// 如果某个状态出现次数 >= 阈值，使用该状态
	for state, count := range stateCount {
		if count >= threshold {
			return state
		}
	}

	// 否则使用当前状态
	return currentState
}

// AnalyzeMarketWithNewAlgorithm 使用新算法分析市场（集成版）
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

	// 多周期配置
	timeframes := []struct {
		interval  string
		klines    []*exchange.Kline
		weight    float64
		minKlines int
		maxKlines int
	}{
		{"1m", klineCache.Klines1m, 0.20, 1, 1},
		{"5m", klineCache.Klines5m, 0.25, 1, 1},
		{"15m", klineCache.Klines15m, 0.25, 1, 1},
		{"30m", klineCache.Klines30m, 0.20, 1, 1},
		{"1h", klineCache.Klines1h, 0.10, 1, 1},
	}

	// 准备数据
	var O, H, L, P, delta []float64
	var thresh []MarketStateThresholds
	var weight []float64
	var stateHistory []string // 用于平滑

	for _, tf := range timeframes {
		if len(tf.klines) < tf.minKlines {
			continue
		}

		// 获取最新一根K线
		latestKline := tf.klines[len(tf.klines)-1]

		O = append(O, latestKline.Open)
		H = append(O, latestKline.High)
		L = append(L, latestKline.Low)
		P = append(P, latestKline.Close)

		// 动态计算delta（改进版）
		deltaValue := CalculateDynamicDelta(tf.klines)
		delta = append(delta, deltaValue)

		// 阈值配置（可以根据币种特性调整）
		thresh = append(thresh, MarketStateThresholds{
			LowV:       1.0,
			HighV:      2.0,
			TrendV:     1.2,
			DThreshold: 0.7,
		})

		weight = append(weight, tf.weight)

		// 记录历史状态（用于平滑）
		if previousAnalysis != nil {
			if tfResult, ok := previousAnalysis.TimeframeAnalysis[tf.interval]; ok {
				stateHistory = append(stateHistory, string(tfResult.MarketState))
			}
		}
	}

	if len(O) == 0 {
		return nil
	}

	// 使用新算法计算市场状态
	finalState := DetectMarketStateMultiCycle(O, H, L, P, delta, thresh, weight)

	// 平滑处理（改进版）
	if len(stateHistory) > 0 {
		finalState = SmoothState(finalState, stateHistory, 2)
	}

	// 映射到MarketState类型
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
	analysis.MarketStateConf = 0.8 // 新算法置信度

	// 计算波动率（用于兼容）
	var totalVolatility float64
	var totalWeight float64
	for i, tf := range timeframes {
		if i < len(delta) && delta[i] > 0 {
			volatility := (H[i] - L[i]) / delta[i] * 100 // 转换为百分比
			totalVolatility += volatility * weight[i]
			totalWeight += weight[i]
		}
	}
	if totalWeight > 0 {
		analysis.Volatility = totalVolatility / totalWeight
	}

	// 计算趋势强度（用于兼容）
	var totalTrendStrength float64
	for i := range O {
		if P[i] >= O[i] {
			totalTrendStrength += 0.5 * weight[i] // 上涨
		} else {
			totalTrendStrength -= 0.5 * weight[i] // 下跌
		}
	}
	analysis.TrendStrength = totalTrendStrength

	return analysis
}

// 对比函数：展示两种算法的差异
func CompareAlgorithms(
	klines []*exchange.Kline,
	newAlgorithm bool,
) (MarketState, float64) {
	if newAlgorithm {
		// 使用新算法
		// ... 实现新算法逻辑
		return MarketStateTrend, 0.8
	} else {
		// 使用当前算法
		// ... 实现当前算法逻辑
		return MarketStateTrend, 0.8
	}
}

