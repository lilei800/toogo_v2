// Package market 市场分析引擎（改进版）
// 优化点：
// 1. 多维度综合波动率计算
// 2. 趋势持续性判断
// 3. 币种特性自适应学习
// 4. 按需计算（只计算有机器人的币种）
// 5. 市场状态综合评分机制
package market

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"

	"hotgo/internal/dao"
	"hotgo/internal/library/exchange"
	"hotgo/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
)

// ComprehensiveVolatility 综合波动率
type ComprehensiveVolatility struct {
	RangeVolatility        float64 // 价格范围波动率
	FrequencyVolatility    float64 // 变化频率波动率
	ATRVolatility          float64 // ATR波动率
	AccelerationVolatility float64 // 加速度波动率
	FinalVolatility        float64 // 最终综合波动率
}

// TrendAnalysis 趋势分析
type TrendAnalysis struct {
	Consistency float64 // 多周期一致性 0-1
	Duration    float64 // 趋势持续时间占比 0-1
	Strength    float64 // 趋势强度 0-1
	Direction   string  // up/down/sideways
	IsTrend     bool    // 是否为趋势市场
}

// MarketStateScore 市场状态评分
type MarketStateScore struct {
	TrendScore    float64
	VolatileScore float64
	HighVolScore  float64
	LowVolScore   float64
	FinalState    string
	Confidence    float64
}

// SymbolCharacteristics 币种特性
type SymbolCharacteristics struct {
	Symbol           string
	Platform         string
	NormalVolatility float64 // 正常波动率（中位数）
	HighVolThreshold float64 // 高波动阈值（85分位数）
	LowVolThreshold  float64 // 低波动阈值（15分位数）
	TrendThreshold   float64 // 趋势阈值
	LastUpdated      time.Time
	SampleCount      int // 样本数量
}

// SymbolCharacteristicsCache 币种特性缓存
type SymbolCharacteristicsCache struct {
	mu    sync.RWMutex
	cache map[string]*SymbolCharacteristics // key: platform:symbol
}

var symbolCharacteristicsCache = &SymbolCharacteristicsCache{
	cache: make(map[string]*SymbolCharacteristics),
}

// GetActiveSymbols 获取所有运行中机器人使用的币种
func GetActiveSymbols(ctx context.Context) (map[string]map[string]bool, error) {
	var robots []*entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).
		Where("status", 2). // 只查询运行中的机器人
		Fields("exchange", "symbol").
		Scan(&robots)

	if err != nil {
		return nil, err
	}

	// platform -> symbol set
	activeSymbols := make(map[string]map[string]bool)

	for _, robot := range robots {
		// 关键：MarketAnalyzer 的 key 是 platform:symbol。
		// 当机器人配置了 analysisSourceOverrides（例如 gate/bitget -> okx），
		// 我们必须把“分析平台”加入 activeSymbols，否则一旦没有 OKX 机器人在跑，
		// MarketAnalyzer 就会停止分析 OKX，导致 Gate/Bitget “市场状态/多周期K线”断更。
		execPlatform := NormalizePlatform(robot.Exchange)
		platform := ResolveAnalysisPlatform(ctx, execPlatform)
		if platform == "" {
			platform = execPlatform
		}
		symbol := NormalizeSymbol(robot.Symbol)

		if activeSymbols[platform] == nil {
			activeSymbols[platform] = make(map[string]bool)
		}
		activeSymbols[platform][symbol] = true
	}

	return activeSymbols, nil
}

// calculateComprehensiveVolatility 计算综合波动率
func calculateComprehensiveVolatility(klines []*exchange.Kline) *ComprehensiveVolatility {
	if len(klines) < 10 {
		return nil
	}

	closes := make([]float64, len(klines))
	highs := make([]float64, len(klines))
	lows := make([]float64, len(klines))
	for i, k := range klines {
		closes[i] = k.Close
		highs[i] = k.High
		lows[i] = k.Low
	}

	currentPrice := closes[len(closes)-1]
	if currentPrice <= 0 {
		return nil
	}

	// 1. 价格范围波动率
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
	rangeVol := (maxPrice - minPrice) / currentPrice * 100

	// 2. 变化频率波动率（统计价格变化次数）
	changeCount := 0
	threshold := currentPrice * 0.001 // 0.1%阈值
	for i := 1; i < len(closes); i++ {
		if math.Abs(closes[i]-closes[i-1]) > threshold {
			changeCount++
		}
	}
	freqVol := float64(changeCount) / float64(len(closes)-1) * 100

	// 3. ATR波动率
	atr := calculateATR(highs, lows, closes, 14)
	atrVol := (atr / currentPrice) * 100

	// 4. 加速度波动率（价格变化率的变化率）
	priceChanges := make([]float64, len(closes)-1)
	for i := 1; i < len(closes); i++ {
		if closes[i-1] > 0 {
			priceChanges[i-1] = (closes[i] - closes[i-1]) / closes[i-1] * 100
		}
	}
	acceleration := 0.0
	if len(priceChanges) >= 2 {
		acceleration = math.Abs(priceChanges[len(priceChanges)-1] - priceChanges[len(priceChanges)-2])
	}

	// 5. 加权平均
	finalVol := rangeVol*0.3 + freqVol*0.2 + atrVol*0.3 + acceleration*0.2

	return &ComprehensiveVolatility{
		RangeVolatility:        rangeVol,
		FrequencyVolatility:    freqVol,
		ATRVolatility:          atrVol,
		AccelerationVolatility: acceleration,
		FinalVolatility:        finalVol,
	}
}

// calculateATR 计算ATR
func calculateATR(highs, lows, closes []float64, period int) float64 {
	if len(highs) < period+1 {
		return 0
	}

	var trSum float64
	startIdx := len(highs) - period
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx + 1; i < len(highs); i++ {
		tr := math.Max(highs[i]-lows[i],
			math.Max(math.Abs(highs[i]-closes[i-1]),
				math.Abs(lows[i]-closes[i-1])))
		trSum += tr
	}

	return trSum / float64(len(highs)-startIdx-1)
}

// determineSingleTimeframeTrend 判断单周期趋势
func determineSingleTimeframeTrend(klines []*exchange.Kline) string {
	if len(klines) < 10 {
		return "sideways"
	}

	closes := make([]float64, len(klines))
	for i, k := range klines {
		closes[i] = k.Close
	}

	startPrice := closes[0]
	endPrice := closes[len(closes)-1]

	changeRate := (endPrice - startPrice) / startPrice * 100

	threshold := 0.1 // 0.1%阈值
	if changeRate > threshold {
		return "up"
	} else if changeRate < -threshold {
		return "down"
	}
	return "sideways"
}

// calculateTrendDuration 计算趋势持续时间
func calculateTrendDuration(klines []*exchange.Kline, direction string) float64 {
	if len(klines) < 10 {
		return 0
	}

	consecutiveCount := 0
	maxConsecutive := 0

	for i := 1; i < len(klines); i++ {
		klineTrend := "sideways"
		if klines[i].Close > klines[i-1].Close*1.001 {
			klineTrend = "up"
		} else if klines[i].Close < klines[i-1].Close*0.999 {
			klineTrend = "down"
		}

		if klineTrend == direction {
			consecutiveCount++
			if consecutiveCount > maxConsecutive {
				maxConsecutive = consecutiveCount
			}
		} else {
			consecutiveCount = 0
		}
	}

	return float64(maxConsecutive) / float64(len(klines))
}

// calculateTrendStrength 计算趋势强度（使用线性回归）
func calculateTrendStrength(klines []*exchange.Kline) float64 {
	if len(klines) < 10 {
		return 0
	}

	closes := make([]float64, len(klines))
	for i, k := range klines {
		closes[i] = k.Close
	}

	// 线性回归计算斜率
	n := len(closes)
	var sumX, sumY, sumXY, sumX2 float64
	for i := 0; i < n; i++ {
		x := float64(i)
		y := closes[i]
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

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
	return math.Min(1.0, normalizedSlope)
}

// analyzeTrend 分析趋势
func analyzeTrend(klinesMap map[string][]*exchange.Kline) *TrendAnalysis {
	// 1. 多周期趋势一致性
	trends := make(map[string]string) // timeframe -> trend
	for tf, klines := range klinesMap {
		if len(klines) >= 10 {
			trends[tf] = determineSingleTimeframeTrend(klines)
		}
	}

	if len(trends) == 0 {
		return &TrendAnalysis{
			Consistency: 0,
			Duration:    0,
			Strength:    0,
			Direction:   "sideways",
			IsTrend:     false,
		}
	}

	// 统计趋势方向
	upCount := 0
	downCount := 0
	total := len(trends)
	for _, trend := range trends {
		if trend == "up" {
			upCount++
		} else if trend == "down" {
			downCount++
		}
	}

	consistency := float64(max(upCount, downCount)) / float64(total)
	direction := "sideways"
	if upCount > downCount && consistency > 0.5 {
		direction = "up"
	} else if downCount > upCount && consistency > 0.5 {
		direction = "down"
	}

	// 2. 趋势持续时间（在主要周期中）
	mainKlines := klinesMap["5m"] // 使用5分钟周期
	if len(mainKlines) < 20 {
		mainKlines = klinesMap["15m"]
	}
	if len(mainKlines) < 10 {
		mainKlines = nil
	}

	var trendDuration float64
	if mainKlines != nil {
		trendDuration = calculateTrendDuration(mainKlines, direction)
	}

	// 3. 趋势强度
	var trendStrength float64
	if mainKlines != nil {
		trendStrength = calculateTrendStrength(mainKlines)
	}

	// 4. 综合判断
	isTrend := consistency > 0.6 && trendDuration > 0.5 && trendStrength > 0.4

	return &TrendAnalysis{
		Consistency: consistency,
		Duration:    trendDuration,
		Strength:    trendStrength,
		Direction:   direction,
		IsTrend:     isTrend,
	}
}

// calculateWindowVolatility 计算窗口波动率
func calculateWindowVolatility(klines []*exchange.Kline) float64 {
	if len(klines) < 5 {
		return 0
	}

	closes := make([]float64, len(klines))
	highs := make([]float64, len(klines))
	lows := make([]float64, len(klines))
	for i, k := range klines {
		closes[i] = k.Close
		highs[i] = k.High
		lows[i] = k.Low
	}

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

	currentPrice := closes[len(closes)-1]
	if currentPrice > 0 {
		return (maxPrice - minPrice) / currentPrice * 100
	}
	return 0
}

// learnSymbolCharacteristics 学习币种特性
func learnSymbolCharacteristics(ctx context.Context, platform, symbol string, klineCache *KlineCache) (*SymbolCharacteristics, error) {
	// 使用1小时周期的历史数据（过去7-30天）
	klines := klineCache.Klines1h
	if len(klines) < 24*7 { // 至少7天数据
		// 如果数据不足，使用其他周期
		if len(klineCache.Klines15m) >= 24*4*7 {
			klines = klineCache.Klines15m
		} else {
			return nil, gerror.New("历史数据不足，无法学习币种特性")
		}
	}

	// 只使用最近30天的数据
	maxDays := 30
	maxKlines := maxDays * 24
	if len(klines) > maxKlines {
		klines = klines[len(klines)-maxKlines:]
	}

	// 计算历史波动率序列
	volatilities := make([]float64, 0)
	windowSize := 24 // 每24根K线计算一个波动率

	for i := windowSize; i < len(klines); i += windowSize {
		windowKlines := klines[i-windowSize : i]
		vol := calculateWindowVolatility(windowKlines)
		if vol > 0 {
			volatilities = append(volatilities, vol)
		}
	}

	if len(volatilities) < 10 {
		return nil, gerror.New("有效样本不足，无法学习币种特性")
	}

	// 排序计算分位数
	sort.Float64s(volatilities)

	normalVol := volatilities[len(volatilities)/2]                // 中位数
	highVol := volatilities[int(float64(len(volatilities))*0.85)] // 85分位数
	lowVol := volatilities[int(float64(len(volatilities))*0.15)]  // 15分位数

	// 计算趋势阈值（使用价格变化率）
	priceChanges := make([]float64, 0)
	for i := 1; i < len(klines); i++ {
		if klines[i-1].Close > 0 {
			changeRate := math.Abs((klines[i].Close - klines[i-1].Close) / klines[i-1].Close * 100)
			priceChanges = append(priceChanges, changeRate)
		}
	}
	sort.Float64s(priceChanges)
	trendThreshold := priceChanges[int(float64(len(priceChanges))*0.7)] // 70分位数

	return &SymbolCharacteristics{
		Symbol:           symbol,
		Platform:         platform,
		NormalVolatility: normalVol,
		HighVolThreshold: highVol,
		LowVolThreshold:  lowVol,
		TrendThreshold:   trendThreshold,
		LastUpdated:      time.Now(),
		SampleCount:      len(volatilities),
	}, nil
}

// getDefaultCharacteristics 获取默认币种特性
func getDefaultCharacteristics(symbol string) *SymbolCharacteristics {
	// 根据币种名称判断（简单规则）
	normalVol := 0.8
	highVol := 1.5
	lowVol := 0.3
	trendThreshold := 0.1

	// BTC/ETH等主流币种
	if symbol == "BTCUSDT" || symbol == "ETHUSDT" {
		normalVol = 0.8
		highVol = 1.5
		lowVol = 0.3
		trendThreshold = 0.1
	} else {
		// 小币种，波动率更高
		normalVol = 2.0
		highVol = 4.0
		lowVol = 0.8
		trendThreshold = 0.3
	}

	return &SymbolCharacteristics{
		Symbol:           symbol,
		NormalVolatility: normalVol,
		HighVolThreshold: highVol,
		LowVolThreshold:  lowVol,
		TrendThreshold:   trendThreshold,
		LastUpdated:      time.Now(),
		SampleCount:      0,
	}
}

// GetCharacteristics 获取币种特性（带缓存）
func (c *SymbolCharacteristicsCache) GetCharacteristics(ctx context.Context, platform, symbol string, klineCache *KlineCache) (*SymbolCharacteristics, error) {
	key := platform + ":" + symbol

	// 1. 检查缓存
	c.mu.RLock()
	if chars, ok := c.cache[key]; ok {
		// 检查是否需要更新（每24小时更新一次）
		if time.Since(chars.LastUpdated) < 24*time.Hour {
			c.mu.RUnlock()
			return chars, nil
		}
	}
	c.mu.RUnlock()

	// 2. 重新学习
	chars, err := learnSymbolCharacteristics(ctx, platform, symbol, klineCache)
	if err != nil {
		// 如果学习失败，尝试使用缓存
		c.mu.RLock()
		if cached, ok := c.cache[key]; ok {
			c.mu.RUnlock()
			return cached, nil
		}
		c.mu.RUnlock()
		// 如果缓存也没有，使用默认值
		return getDefaultCharacteristics(symbol), nil
	}

	// 3. 更新缓存
	c.mu.Lock()
	c.cache[key] = chars
	c.mu.Unlock()

	return chars, nil
}

// calculateMarketStateScore 计算市场状态评分
func calculateMarketStateScore(
	comprehensiveVol *ComprehensiveVolatility,
	trendAnalysis *TrendAnalysis,
	characteristics *SymbolCharacteristics,
) *MarketStateScore {
	scores := &MarketStateScore{}

	volatility := comprehensiveVol.FinalVolatility

	// 1. 趋势市场得分
	if trendAnalysis.IsTrend {
		// 趋势市场：趋势强度高 + 波动率适中
		trendScore := trendAnalysis.Strength*0.4 +
			trendAnalysis.Consistency*0.3 +
			trendAnalysis.Duration*0.3

		// 波动率不能太高也不能太低
		volScore := 1.0
		if volatility > characteristics.HighVolThreshold {
			volScore = 0.3 // 波动率太高，不是趋势
		} else if volatility < characteristics.LowVolThreshold {
			volScore = 0.5 // 波动率太低，可能是低波动
		}

		scores.TrendScore = trendScore * volScore
	} else {
		scores.TrendScore = 0
	}

	// 2. 震荡市场得分
	// 【修复】排除低波动区域，避免低波动市场被误判为震荡
	lowVolUpperBound := characteristics.LowVolThreshold * 1.15 // 允许15%的容差，避免边界值误判
	if !trendAnalysis.IsTrend &&
		volatility > lowVolUpperBound && // 排除低波动区域
		volatility <= characteristics.HighVolThreshold {
		// 震荡市场：无趋势 + 波动率适中
		scores.VolatileScore = (1.0-trendAnalysis.Strength)*0.5 +
			(1.0-trendAnalysis.Consistency)*0.3 +
			0.2 // 波动率适中加分
	} else {
		scores.VolatileScore = 0
	}

	// 3. 高波动市场得分
	if volatility > characteristics.HighVolThreshold {
		// 高波动：波动率超过阈值
		highVolScore := math.Min(1.0, (volatility-characteristics.HighVolThreshold)/
			(characteristics.HighVolThreshold*0.5))
		scores.HighVolScore = highVolScore
	} else {
		scores.HighVolScore = 0
	}

	// 4. 低波动市场得分
	// 【修复】允许阈值附近的波动率也被识别为低波动，避免边界值被误判为震荡
	if volatility <= lowVolUpperBound {
		// 低波动：波动率低于或接近阈值
		if volatility < characteristics.LowVolThreshold {
			// 明显低于阈值，给予高分
			lowVolScore := math.Min(1.0, (characteristics.LowVolThreshold-volatility)/
				(characteristics.LowVolThreshold*0.5))
			scores.LowVolScore = lowVolScore
		} else {
			// 在阈值附近（阈值到1.15倍阈值之间），给予部分得分
			// 使用线性衰减：在阈值处得分为1.0，在1.15倍阈值处得分为0
			excessRatio := (volatility - characteristics.LowVolThreshold) /
				(characteristics.LowVolThreshold * 0.15)
			scores.LowVolScore = math.Max(0, 1.0-excessRatio)
		}
	} else {
		scores.LowVolScore = 0
	}

	// 5. 找到得分最高的状态
	maxScore := 0.0
	finalState := "volatile" // 默认值
	scoresMap := map[string]float64{
		"trend":    scores.TrendScore,
		"volatile": scores.VolatileScore,
		"high_vol": scores.HighVolScore,
		"low_vol":  scores.LowVolScore,
	}

	for state, score := range scoresMap {
		if score > maxScore {
			maxScore = score
			finalState = state
		}
	}

	// 6. 置信度检查
	confidence := maxScore
	minConfidence := 0.5

	if confidence < minConfidence {
		// 如果置信度不足，检查是否有明显的次优状态
		secondMaxScore := 0.0
		for state, score := range scoresMap {
			if state != finalState && score > secondMaxScore {
				secondMaxScore = score
			}
		}

		// 如果次优得分也很高，降低置信度
		if secondMaxScore > 0.3 && maxScore-secondMaxScore < 0.2 {
			confidence = maxScore * 0.8 // 降低置信度
		}

		// 【修复】如果置信度仍然不足，优先根据波动率判断，而不是强制使用震荡
		if confidence < minConfidence {
			// 如果波动率明显低于阈值，优先判断为低波动
			if volatility < characteristics.LowVolThreshold*0.8 {
				finalState = "low_vol"
				confidence = 0.6
			} else if volatility > characteristics.HighVolThreshold {
				// 如果波动率明显高于阈值，判断为高波动
				finalState = "high_vol"
				confidence = 0.6
			} else if trendAnalysis.IsTrend {
				// 如果有明显趋势，判断为趋势市场
				finalState = "trend"
				confidence = 0.6
			} else {
				// 最后才使用震荡作为默认
				finalState = "volatile"
				confidence = 0.6
			}
		}
	}

	scores.FinalState = finalState
	scores.Confidence = confidence

	return scores
}

// max 返回两个float64中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
