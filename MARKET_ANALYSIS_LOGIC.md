# 机器人市场分析逻辑详解

## 📋 概述

机器人系统中有**两套市场分析逻辑**，分别用于不同的场景：

1. **简化版市场分析** (`engine.go` 的 `analyzeMarket` 函数) - 用于实时信号生成
2. **完整版市场分析** (`robot_engine.go` 的 `RobotAnalyzer`) - 多周期综合分析

---

## 🔍 简化版市场分析（实时信号生成）

### 位置
`server/internal/logic/toogo/engine.go` - `analyzeMarket()` 函数

### 分析流程

#### 1. 获取K线数据
```go
klineCountForAnalysis := 100 // 获取100根1分钟K线
klinesForAnalysis, err := runner.Exchange.GetKlines(ctx, runner.Robot.Symbol, "1m", klineCountForAnalysis)
```

#### 2. 计算波动率
使用收盘价序列计算标准差作为波动率：
```go
closes := make([]float64, len(klinesForAnalysis))
for i, k := range klinesForAnalysis {
    closes[i] = k.Close
}
volatility := calculateVolatility(closes, len(closes))
```

**波动率计算公式**：
- 计算收益率序列：`returns[i] = (price[i] - price[i-1]) / price[i-1]`
- 计算收益率的标准差：`volatility = sqrt(variance(returns))`

#### 3. 计算价格范围
```go
analysisHigh := 0.0      // 最高价
analysisLow := math.MaxFloat64  // 最低价
for _, k := range klinesForAnalysis {
    if k.High > analysisHigh {
        analysisHigh = k.High
    }
    if k.Low < analysisLow {
        analysisLow = k.Low
    }
}
priceRange := analysisHigh - analysisLow
avgPrice := (analysisHigh + analysisLow) / 2
volatilityPercent := priceRange / avgPrice * 100  // 价格波动百分比
```

#### 4. 判断市场状态

根据波动率百分比和波动率值判断：

```go
if volatilityPercent > 3 {
    signal.MarketState = "high_vol"      // 高波动：波动百分比 > 3%
} else if volatilityPercent < 0.5 {
    signal.MarketState = "low_vol"       // 低波动：波动百分比 < 0.5%
} else if volatility > 0.01 {
    signal.MarketState = "trend"         // 趋势市场：波动率 > 0.01
} else {
    signal.MarketState = "volatile"      // 震荡市场：其他情况
}
```

**市场状态分类**：
- **high_vol** (高波动): 价格波动百分比 > 3%
- **low_vol** (低波动): 价格波动百分比 < 0.5%
- **trend** (趋势市场): 波动百分比在0.5%-3%之间，且波动率 > 0.01
- **volatile** (震荡市场): 其他情况

---

## 🔬 完整版市场分析（多周期综合分析）

### 位置
`server/internal/logic/toogo/robot_engine.go` - `RobotAnalyzer` 模块

### 分析周期

系统分析**3个核心周期**，每个周期有不同的权重：

| 周期 | 权重 | 说明 |
|------|------|------|
| 5m   | 20%  | 短期周期 |
| 15m  | 35%  | 中期周期（主要） |
| 1h   | 45%  | 长期周期（最重要） |

### 单周期分析流程

#### 1. 计算技术指标

**EMA指标**：
```go
score.EMA12 = calculateEMA(closes, 12)  // 12周期EMA
score.EMA26 = calculateEMA(closes, 26)  // 26周期EMA
score.MACD = score.EMA12 - score.EMA26  // MACD = EMA12 - EMA26
```

**趋势强度**：
使用线性回归计算趋势斜率：
```go
// 线性回归计算斜率
slope = (n*ΣXY - ΣX*ΣY) / (n*ΣX² - (ΣX)²)
// 归一化到0-1范围
normalizedSlope = |slope| / avgPrice * 100
trendStrength = min(1, normalizedSlope)
```

**波动率（ATR）**：
```go
// 计算真实波动幅度（TR）
for i := 1; i < len(klines); i++ {
    tr := max(high-low, max(|high-prevClose|, |low-prevClose|))
    atr += tr
}
atr /= len(klines) - 1
// 相对波动率
volatility = (atr / currentPrice) * 100
```

#### 2. 判断方向

根据EMA和MACD判断：
```go
if EMA12 > EMA26 && MACD > 0 {
    direction = "up"      // 上涨
    strength = min(100, 50 + trendStrength * 50)
} else if EMA12 < EMA26 && MACD < 0 {
    direction = "down"    // 下跌
    strength = min(100, 50 + trendStrength * 50)
} else {
    direction = "neutral" // 中性
    strength = 30
}
```

#### 3. 判断单周期市场状态

```go
// 阈值配置
highVolatilityThreshold = 2.0   // 高波动阈值
lowVolatilityThreshold = 0.5    // 低波动阈值
trendStrengthThreshold = 0.35    // 趋势强度阈值

// 判断逻辑
if trendStrength > 0.35 && volatility >= 0.5 && volatility <= 3.0 {
    marketState = "trend"        // 趋势市场
} else if volatility >= 2.0 {
    marketState = "high_vol"     // 高波动
} else if volatility <= 0.5 {
    marketState = "low_vol"      // 低波动
} else {
    marketState = "volatile"     // 震荡市场
}
```

### 综合判断市场状态

#### 1. 统计各周期状态

```go
var trendCount, highVolCount, lowVolCount int
for _, score := range timeframeScores {
    switch score.MarketState {
    case "trend": trendCount++
    case "high_vol": highVolCount++
    case "low_vol": lowVolCount++
    }
}
```

#### 2. 综合判断（优先级顺序）

```go
// 1. 高波动优先：2个以上周期为高波动
if highVolCount >= 2 {
    marketState = "high_vol"
    confidence = 0.8
    return
}

// 2. 趋势市场：2个以上周期为趋势
if trendCount >= 2 {
    marketState = "trend"
    confidence = trendCount / total
    return
}

// 3. 低波动：2个以上周期为低波动
if lowVolCount >= 2 {
    marketState = "low_vol"
    confidence = 0.7
    return
}

// 4. 默认震荡
marketState = "volatile"
confidence = 0.6
```

#### 3. 判断趋势方向

```go
var upCount, downCount int
for _, score := range timeframeScores {
    if score.Direction == "up" {
        upCount++
    } else if score.Direction == "down" {
        downCount++
    }
}

if upCount >= 2 {
    trendDirection = "up"
    trendStrength = upCount / total * 100
} else if downCount >= 2 {
    trendDirection = "down"
    trendStrength = downCount / total * 100
} else {
    trendDirection = "neutral"
    trendStrength = 30
}
```

#### 4. 计算综合指标

**趋势评分**：
```go
// 加权趋势评分
for tf, score := range timeframeScores {
    weight := timeframeWeights[tf]
    if score.Direction == "up" {
        weightedTrendSum += score.Strength * weight
    } else if score.Direction == "down" {
        weightedTrendSum -= score.Strength * weight
    }
}
trendScore = weightedTrendSum / totalWeight  // -100 ~ 100
```

**波动评分**：
```go
avgVolatility = sum(volatility) / count
volatilityScore = min(100, avgVolatility * 20)  // 0-100
```

---

## 📊 市场状态说明

### 四种市场状态

| 状态 | 说明 | 特征 | 适用策略 |
|------|------|------|----------|
| **trend** | 趋势市场 | 趋势强度高，波动率适中 | 趋势跟踪策略 |
| **volatile** | 震荡市场 | 波动频繁，无明显趋势 | 区间交易策略 |
| **high_vol** | 高波动市场 | 价格波动剧烈 | 保守策略，降低杠杆 |
| **low_vol** | 低波动市场 | 价格波动很小 | 平衡策略 |

### 风险偏好映射

根据市场状态自动映射风险偏好：

| 市场状态 | 风险偏好 | 说明 |
|----------|----------|------|
| trend | balanced | 趋势市场，平衡策略 |
| volatile | conservative | 震荡市场，保守策略 |
| high_vol | conservative | 高波动，保守策略 |
| low_vol | balanced | 低波动，平衡策略 |

---

## 🔧 技术指标说明

### EMA (指数移动平均线)

**计算公式**：
```
EMA(today) = (Price(today) - EMA(yesterday)) * Multiplier + EMA(yesterday)
Multiplier = 2 / (Period + 1)
```

**用途**：
- EMA12：短期趋势
- EMA26：长期趋势
- EMA12 > EMA26：上涨趋势
- EMA12 < EMA26：下跌趋势

### MACD (移动平均收敛散度)

**计算公式**：
```
MACD = EMA12 - EMA26
```

**用途**：
- MACD > 0：上涨趋势
- MACD < 0：下跌趋势
- MACD绝对值越大，趋势越强

### ATR (平均真实波动幅度)

**计算公式**：
```
TR = max(High - Low, |High - PrevClose|, |Low - PrevClose|)
ATR = average(TR)
```

**用途**：
- 衡量市场波动性
- ATR越大，波动越大
- 用于判断市场状态

### 趋势强度（线性回归斜率）

**计算公式**：
```
slope = (n*ΣXY - ΣX*ΣY) / (n*ΣX² - (ΣX)²)
normalizedSlope = |slope| / avgPrice * 100
trendStrength = min(1, normalizedSlope)
```

**用途**：
- 衡量趋势的强度
- 值越大，趋势越明显
- 用于判断是否为趋势市场

---

## 📈 分析结果结构

### RobotMarketAnalysis（市场分析结果）

```go
type RobotMarketAnalysis struct {
    Timestamp       time.Time
    MarketState     string              // trend/volatile/high_vol/low_vol
    MarketStateConf float64             // 市场状态置信度 0-1
    TrendDirection  string              // up/down/neutral
    TrendStrength   float64             // 趋势强度 0-100
    Volatility      float64             // 波动率
    VolatilityLevel string              // low/normal/high
    TimeframeScores map[string]*TimeframeScore  // 各周期评分
    Indicators      *TechnicalIndicators         // 技术指标汇总
}
```

### TimeframeScore（单周期评分）

```go
type TimeframeScore struct {
    Timeframe     string   // 周期：5m/15m/1h
    Direction     string   // 方向：up/down/neutral
    Strength      float64  // 方向强度 0-100
    TrendStrength float64  // 趋势强度 0-1
    Volatility    float64  // 波动率
    MarketState   string   // 市场状态
    MACD          float64  // MACD值
    EMA12         float64  // EMA12值
    EMA26         float64  // EMA26值
    KlinesCount   int      // K线数量
}
```

---

## 🎯 使用场景

### 简化版市场分析
- **用途**：实时信号生成
- **频率**：每10秒执行一次（定时任务）
- **特点**：快速、轻量级
- **输出**：市场状态、风险偏好、交易信号

### 完整版市场分析
- **用途**：深度市场分析
- **频率**：每1秒执行一次（分析循环）
- **特点**：多周期、多指标综合分析
- **输出**：详细的市场分析结果、技术指标、趋势判断

---

## 🔄 数据流

```
获取K线数据
    ↓
计算技术指标（EMA、MACD、ATR、趋势强度）
    ↓
单周期分析（5m/15m/1h）
    ↓
综合判断（加权平均）
    ↓
判断市场状态（trend/volatile/high_vol/low_vol）
    ↓
判断趋势方向（up/down/neutral）
    ↓
计算置信度
    ↓
输出分析结果
```

---

## 📝 注意事项

1. **数据要求**：
   - 单周期分析至少需要26根K线
   - 简化版分析至少需要3根K线

2. **市场状态判断**：
   - 优先判断高波动（最危险）
   - 其次判断趋势市场（最有利）
   - 最后判断低波动和震荡

3. **置信度**：
   - 多个周期一致时，置信度更高
   - 置信度用于评估分析结果的可靠性

4. **性能考虑**：
   - 简化版分析速度快，适合高频调用
   - 完整版分析计算量大，适合深度分析

---

**文档版本**: v1.0  
**最后更新**: 2024年

