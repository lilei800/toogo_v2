package market

import (
	"math"

	"hotgo/internal/library/exchange"
)

// calculateMarketStateVolatilities 根据历史K线识别不同市场状态模式，计算每种状态的波动率特征值
// 【动态波动率优化】适应趋势向上、趋势向下、上下插针、高波动、低波动、震荡等行情
// 【动态波动率优化】适应趋势向上、趋势向下、上下插针、高波动、低波动、震荡等行情
func (a *MarketAnalyzer) calculateMarketStateVolatilities(klineCache *KlineCache) *MarketStateVolatilities {
	result := &MarketStateVolatilities{}

	// 【周期优化】使用多周期K线数据识别市场状态模式
	// 使用1m、5m、15m、30m、1h五个周期，去掉1天周期，更适合超短线交易
	timeframes := []struct {
		klines    []*exchange.Kline
		weight    float64
		minKlines int
		maxKlines int
	}{
		{klineCache.Klines1m, 0.30, 8, 15},   // 1分钟：权重30%，8-15根（高权重增加K线，提高数据支撑）
		{klineCache.Klines5m, 0.40, 20, 30}, // 5分钟：权重40%，20-30根（最高权重，大幅增加K线）
		{klineCache.Klines15m, 0.20, 12, 20}, // 15分钟：权重20%，12-20根（中权重保持适中）
		{klineCache.Klines30m, 0.07, 6, 10},  // 30分钟：权重7%，6-10根（低权重减少K线）
		{klineCache.Klines1h, 0.03, 5, 8},  // 1小时：权重3%，5-8根（最低权重，大幅减少K线）
	}

	var trendUpVols, trendDownVols, spikeVols, highVolVols, lowVolVols, volatileVols []float64
	var trendUpWeights, trendDownWeights, spikeWeights, highVolWeights, lowVolWeights, volatileWeights float64

	for _, tf := range timeframes {
		if len(tf.klines) < tf.minKlines {
			continue
		}

		// 只使用最近maxKlines根K线
		var klinesToUse []*exchange.Kline
		if len(tf.klines) > tf.maxKlines {
			klinesToUse = tf.klines[len(tf.klines)-tf.maxKlines:]
		} else {
			klinesToUse = tf.klines
		}

		// 识别市场状态模式并计算波动率
		patternVols := a.identifyMarketPatterns(klinesToUse, tf.weight)
		
		// 收集各模式的波动率
		if patternVols.TrendUpVolatility > 0 {
			trendUpVols = append(trendUpVols, patternVols.TrendUpVolatility)
			trendUpWeights += tf.weight
		}
		if patternVols.TrendDownVolatility > 0 {
			trendDownVols = append(trendDownVols, patternVols.TrendDownVolatility)
			trendDownWeights += tf.weight
		}
		if patternVols.SpikeVolatility > 0 {
			spikeVols = append(spikeVols, patternVols.SpikeVolatility)
			spikeWeights += tf.weight
		}
		if patternVols.HighVolVolatility > 0 {
			highVolVols = append(highVolVols, patternVols.HighVolVolatility)
			highVolWeights += tf.weight
		}
		if patternVols.LowVolVolatility > 0 {
			lowVolVols = append(lowVolVols, patternVols.LowVolVolatility)
			lowVolWeights += tf.weight
		}
		if patternVols.VolatileVolatility > 0 {
			volatileVols = append(volatileVols, patternVols.VolatileVolatility)
			volatileWeights += tf.weight
		}
	}

	// 计算加权平均波动率
	if len(trendUpVols) > 0 && trendUpWeights > 0 {
		var sum float64
		for _, v := range trendUpVols {
			sum += v
		}
		result.TrendUpVolatility = sum / float64(len(trendUpVols))
	}
	if len(trendDownVols) > 0 && trendDownWeights > 0 {
		var sum float64
		for _, v := range trendDownVols {
			sum += v
		}
		result.TrendDownVolatility = sum / float64(len(trendDownVols))
	}
	if len(spikeVols) > 0 && spikeWeights > 0 {
		var sum float64
		for _, v := range spikeVols {
			sum += v
		}
		result.SpikeVolatility = sum / float64(len(spikeVols))
	}
	if len(highVolVols) > 0 && highVolWeights > 0 {
		var sum float64
		for _, v := range highVolVols {
			sum += v
		}
		result.HighVolVolatility = sum / float64(len(highVolVols))
	}
	if len(lowVolVols) > 0 && lowVolWeights > 0 {
		var sum float64
		for _, v := range lowVolVols {
			sum += v
		}
		result.LowVolVolatility = sum / float64(len(lowVolVols))
	}
	if len(volatileVols) > 0 && volatileWeights > 0 {
		var sum float64
		for _, v := range volatileVols {
			sum += v
		}
		result.VolatileVolatility = sum / float64(len(volatileVols))
	}

	// 如果没有识别到任何模式，使用默认值
	if result.TrendUpVolatility == 0 && result.TrendDownVolatility == 0 &&
		result.SpikeVolatility == 0 && result.HighVolVolatility == 0 &&
		result.LowVolVolatility == 0 && result.VolatileVolatility == 0 {
		// 使用基准波动率作为默认值
		baseline := a.calculateMultiTimeframeBaselineVolatility(klineCache)
		if baseline <= 0 {
			baseline = 0.8 // 默认基准：0.8%
		}
		result.TrendUpVolatility = baseline * 1.2    // 趋势向上：基准 × 1.2
		result.TrendDownVolatility = baseline * 1.2   // 趋势向下：基准 × 1.2
		result.SpikeVolatility = baseline * 3.0       // 上下插针：基准 × 3.0
		result.HighVolVolatility = baseline * 2.0      // 高波动：基准 × 2.0
		result.LowVolVolatility = baseline * 0.3       // 低波动：基准 × 0.3
		result.VolatileVolatility = baseline * 0.8     // 震荡：基准 × 0.8
	}

	return result
}

// identifyMarketPatterns 识别单个K线窗口的市场状态模式
func (a *MarketAnalyzer) identifyMarketPatterns(klines []*exchange.Kline, weight float64) *MarketStateVolatilities {
	result := &MarketStateVolatilities{}

	if len(klines) < 3 {
		return result
	}

	// 提取价格数据
	closes := make([]float64, len(klines))
	highs := make([]float64, len(klines))
	lows := make([]float64, len(klines))
	for i, k := range klines {
		closes[i] = k.Close
		highs[i] = k.High
		lows[i] = k.Low
	}

	currentPrice := closes[len(closes)-1]
	startPrice := closes[0]

	// 计算基础指标
	maxPrice := highs[0]
	minPrice := lows[0]
	for i := 1; i < len(klines); i++ {
		if highs[i] > maxPrice {
			maxPrice = highs[i]
		}
		if lows[i] < minPrice {
			minPrice = lows[i]
		}
	}

	if currentPrice <= 0 {
		return result
	}

	// 1. 计算价格波动率
	priceVolatility := ((maxPrice - minPrice) / currentPrice) * 100

	// 2. 计算价格变化率
	priceChangeRate := 0.0
	if startPrice > 0 {
		priceChangeRate = ((currentPrice - startPrice) / startPrice) * 100
	}

	// 3. 计算上下影线比例（识别插针）
	var upperShadowRatio, lowerShadowRatio float64
	for i := 0; i < len(klines); i++ {
		body := math.Abs(closes[i] - (highs[i]+lows[i])/2)
		upperShadow := highs[i] - math.Max(closes[i], (highs[i]+lows[i])/2)
		lowerShadow := math.Min(closes[i], (highs[i]+lows[i])/2) - lows[i]
		
		if body > 0 {
			upperShadowRatio += (upperShadow / currentPrice) * 100
			lowerShadowRatio += (lowerShadow / currentPrice) * 100
		}
	}
	upperShadowRatio /= float64(len(klines))
	lowerShadowRatio /= float64(len(klines))
	totalShadowRatio := upperShadowRatio + lowerShadowRatio

	// 4. 计算价格变化的一致性（识别趋势）
	priceConsistency := 0.0
	if len(closes) >= 3 {
		upCount := 0
		for i := 1; i < len(closes); i++ {
			if closes[i] > closes[i-1] {
				upCount++
			}
		}
		priceConsistency = float64(upCount) / float64(len(closes)-1)
		// 如果一致性接近0或1，说明趋势明显
		if priceConsistency < 0.3 {
			priceConsistency = 1.0 - priceConsistency // 下跌趋势
		}
	}

	// 5. 识别市场状态模式并计算波动率

	// 模式1：上下插针（上下影线比例大）
	if totalShadowRatio > priceVolatility*0.6 {
		result.SpikeVolatility = priceVolatility
	}

	// 模式2：趋势向上（价格变化率 > 0.2%，一致性高）
	if priceChangeRate > 0.2 && priceConsistency > 0.6 {
		result.TrendUpVolatility = priceVolatility
	}

	// 模式3：趋势向下（价格变化率 < -0.2%，一致性高）
	if priceChangeRate < -0.2 && priceConsistency > 0.6 {
		result.TrendDownVolatility = priceVolatility
	}

	// 模式4：高波动（波动率大，且变化率不明显）
	if priceVolatility > 1.5 && math.Abs(priceChangeRate) < 0.5 {
		result.HighVolVolatility = priceVolatility
	}

	// 模式5：低波动（波动率小）
	// 【优化】提高低波动识别阈值，从0.3%提高到0.5%，更准确识别低波动状态
	// 0.3%太严格，很多实际低波动（0.4%-0.5%）无法被识别
	if priceVolatility < 0.5 {
		result.LowVolVolatility = priceVolatility
	}

	// 模式6：震荡（波动率中等，变化率小，一致性低）
	// 【优化】调整震荡识别条件，避免与低波动重叠
	// 震荡应该是波动率在0.5%-1.5%之间，且不是低波动
	if priceVolatility >= 0.5 && priceVolatility <= 1.5 &&
		math.Abs(priceChangeRate) < 0.3 && priceConsistency < 0.6 {
		result.VolatileVolatility = priceVolatility
	}

	return result
}

// calculateDynamicThresholdsFromMarketStates 根据历史K线识别的市场状态波动率特征值计算动态阈值
// 【动态波动率优化】使用不同市场状态的波动率特征值，更精准地判断当前市场状态
func (a *MarketAnalyzer) calculateDynamicThresholdsFromMarketStates(baselineVolatility float64, marketStateVolatilities *MarketStateVolatilities) (highThreshold, lowThreshold float64) {
	if marketStateVolatilities == nil {
		// 降级：使用原来的方法
		return a.calculateDynamicThresholds(baselineVolatility)
	}

	// 【动态波动率优化】根据历史K线识别的市场状态波动率特征值计算阈值
	// 高波动阈值：取高波动、插针、趋势状态的波动率最大值
	highVolCandidates := []float64{
		marketStateVolatilities.HighVolVolatility,
		marketStateVolatilities.SpikeVolatility,
		marketStateVolatilities.TrendUpVolatility,
		marketStateVolatilities.TrendDownVolatility,
	}
	
	maxHighVol := 0.0
	for _, v := range highVolCandidates {
		if v > maxHighVol {
			maxHighVol = v
		}
	}
	
	// 【优化】低波动阈值：优先使用低波动特征值，如果不存在则使用震荡状态
	// 低波动和震荡是不同的状态，不应该取最小值，而应该优先识别低波动
	var lowThresholdFromPattern float64
	if marketStateVolatilities.LowVolVolatility > 0 {
		// 如果识别到了低波动模式，使用低波动特征值 × 1.5（提高阈值，更准确识别低波动）
		// 例如：低波动特征值0.3% → 阈值0.45%，实际波动率0.4%会被识别为低波动
		lowThresholdFromPattern = marketStateVolatilities.LowVolVolatility * 1.5
	} else if marketStateVolatilities.VolatileVolatility > 0 {
		// 如果没有识别到低波动模式，但识别到了震荡模式，使用震荡特征值 × 0.8（降低阈值）
		// 这样可以更准确地区分低波动和震荡
		lowThresholdFromPattern = marketStateVolatilities.VolatileVolatility * 0.8
	}

	// 如果识别到了有效的波动率特征值，使用它们
	if maxHighVol > 0 {
		// 高波动阈值 = 高波动特征值 × 0.8（稍微降低，避免过度敏感）
		highThreshold = maxHighVol * 0.8
	} else {
		// 降级：使用基准波动率
		highThreshold = baselineVolatility * 1.5
	}

	if lowThresholdFromPattern > 0 {
		// 使用从模式识别得到的低波动阈值
		lowThreshold = lowThresholdFromPattern
	} else {
		// 降级：使用基准波动率
		lowThreshold = baselineVolatility * 0.5
	}

	// 确保阈值在合理范围内
	if baselineVolatility > 0 {
		// 高波动阈值：基准 × 1.2 ~ 基准 × 3.0
		highThreshold = math.Max(baselineVolatility*1.2, math.Min(baselineVolatility*3.0, highThreshold))
		// 低波动阈值：基准 × 0.2 ~ 基准 × 0.8
		lowThreshold = math.Max(baselineVolatility*0.2, math.Min(baselineVolatility*0.8, lowThreshold))
	}

	// 最终限制：高波动阈值 0.6% ~ 10%，低波动阈值 0.1% ~ 2%
	highThreshold = math.Max(0.6, math.Min(10.0, highThreshold))
	lowThreshold = math.Max(0.1, math.Min(2.0, lowThreshold))

	return highThreshold, lowThreshold
}

