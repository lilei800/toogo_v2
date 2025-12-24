# BTCUSDT 市场状态计算方式（实际代码）

## ⚠️ 重要说明

**实际使用的是 `MarketAnalyzer`（全局引擎），不是 `RobotAnalyzer`（本地计算，已废弃）**

---

## 一、实际使用的计算逻辑

### 1.1 数据来源

**BTCUSDT 市场状态计算使用的数据**：
- **K线数据**：从 `MarketServiceManager.GetMultiTimeframeKlines("bitget", "BTCUSDT")` 获取
- **多周期K线**：
  - **5分钟周期（5m）**：权重 0.20（短期）
  - **15分钟周期（15m）**：权重 0.35（中期，主要）
  - **1小时周期（1h）**：权重 0.45（长期，最重要）
- **最小K线要求**：每个周期至少需要 **26根K线**

**注意**：只使用3个周期，**不包括1m和1d周期**

---

## 二、技术指标计算

### 2.1 每个周期计算的技术指标

**对每个周期（5m、15m、1h）分别计算**：

| 指标 | 说明 | 参数 |
|------|------|------|
| EMA12 | 12周期指数移动平均 | 基于收盘价 |
| EMA26 | 26周期指数移动平均 | 基于收盘价 |
| MACD | MACD指标 | EMA12 - EMA26 |
| ATR | 平均真实波动范围 | 14周期 |

### 2.2 趋势判断

**使用3个条件判断趋势**：
1. **EMA排列**：EMA12 > EMA26 → +0.5分，否则 -0.5分
2. **价格与EMA关系**：当前价 > EMA平均值 → +0.3分，否则 -0.3分
3. **MACD方向**：MACD > 0 → +0.2分，否则 -0.2分

**判定**：
- 总分 > 0.3 → "up"（上涨趋势）
- 总分 < -0.3 → "down"（下跌趋势）
- 否则 → "sideways"（横盘）

---

## 三、综合指标计算

### 3.1 加权平均

**综合趋势强度**：
```
TrendStrength = Σ(各周期TrendStrength × 权重)
```

**综合波动率**：
```
Volatility = Σ(各周期ATR × 权重)
```

**权重分配**：
- 5m: 0.20
- 15m: 0.35
- 1h: 0.45

---

## 四、市场状态判定（实际代码）

### 4.1 判定逻辑

**代码位置**：`server/internal/library/market/market_analyzer.go:determineMarketState()`

```go
// 1. 优先判断波动率（使用固定阈值）
highVolThreshold := 2.0
lowVolThreshold := 0.5

if analysis.Volatility > highVolThreshold {
    return MarketStateHighVol  // 高波动
}
if analysis.Volatility < lowVolThreshold {
    return MarketStateLowVol  // 低波动
}

// 2. 判断趋势一致性
trendConsistency := calculateTrendConsistency(analysis.TimeframeAnalysis)
if trendConsistency > 0.6 {
    return MarketStateTrend  // 趋势市场
}

// 3. 默认震荡
return MarketStateVolatile  // 震荡市场
```

### 4.2 趋势一致性计算

**代码位置**：`server/internal/library/market/market_analyzer.go:calculateTrendConsistency()`

```go
// 统计各周期的趋势方向
var upCount, downCount int
for _, tf := range timeframeResults {
    switch tf.Trend {
    case "up":
        upCount++
    case "down":
        downCount++
    }
}

// 计算一致性 = max(upCount, downCount) / total
trendConsistency = float64(maxCount) / float64(total)
```

---

## 五、阈值设置（实际代码）

### 5.1 固定阈值

**当前使用固定阈值**（所有币种相同）：

| 阈值 | 值 | 说明 |
|------|-----|------|
| highVolThreshold | **2.0** | 高波动阈值（固定值） |
| lowVolThreshold | **0.5** | 低波动阈值（固定值） |
| trendConsistencyThreshold | **0.6** | 趋势一致性阈值 |

**⚠️ 注意**：这些是**固定阈值**，不是自适应阈值！

---

## 六、完整计算流程（BTCUSDT示例）

### 6.1 步骤1：获取K线数据

```
bitget:BTCUSDT
├── 5m: 100根K线（权重0.20）
├── 15m: 100根K线（权重0.35）
└── 1h: 50根K线（权重0.45）
```

### 6.2 步骤2：计算各周期技术指标

**5m周期**：
- EMA12 = 89200.5
- EMA26 = 89150.3
- MACD = 50.2
- ATR = 0.8
- Trend = "up" (score=0.6)

**15m周期**：
- EMA12 = 89180.2
- EMA26 = 89120.1
- MACD = 60.1
- ATR = 1.2
- Trend = "up" (score=0.7)

**1h周期**：
- EMA12 = 89100.8
- EMA26 = 89050.5
- MACD = 50.3
- ATR = 1.5
- Trend = "up" (score=0.65)

### 6.3 步骤3：计算综合指标

```
TrendStrength = 0.6×0.20 + 0.7×0.35 + 0.65×0.45 = 0.66
Volatility = 0.8×0.20 + 1.2×0.35 + 1.5×0.45 = 1.26
```

### 6.4 步骤4：判定市场状态

```
1. 判断波动率：
   Volatility = 1.26
   1.26 > 2.0? → 否
   1.26 < 0.5? → 否
   → 继续判断趋势

2. 判断趋势一致性：
   upCount = 3, downCount = 0
   trendConsistency = 3/3 = 1.0
   1.0 > 0.6? → 是
   → 返回 "trend"
```

**最终结果**：`MarketState = "trend"`

---

## 七、与 RobotAnalyzer 的区别

### 7.1 MarketAnalyzer（实际使用）

| 特性 | 值 |
|------|-----|
| 周期数量 | 3个（5m, 15m, 1h） |
| 权重分配 | 5m:0.20, 15m:0.35, 1h:0.45 |
| 阈值类型 | **固定阈值**（2.0, 0.5） |
| 判定方式 | 先波动率，后趋势一致性 |
| 使用场景 | 全局引擎，所有机器人共享 |

### 7.2 RobotAnalyzer（已废弃）

| 特性 | 值 |
|------|-----|
| 周期数量 | 5个（1m, 5m, 15m, 1h, 1d） |
| 权重分配 | 1m:0.20, 5m:0.25, 15m:0.25, 1h:0.20, 1d:0.10 |
| 阈值类型 | **自适应阈值**（基于基准波动率） |
| 判定方式 | 加权投票机制 |
| 使用场景 | 本地计算（已废弃，不再使用） |

---

## 八、代码位置

### 8.1 实际使用的代码

**文件**：`server/internal/library/market/market_analyzer.go`

- `analyzeMarket()` - 分析单个市场（第180行）
- `analyzeTimeframe()` - 分析单个周期（第235行）
- `determineMarketState()` - 判定市场状态（第338行）
- `calculateTrendConsistency()` - 计算趋势一致性（第360行）

### 8.2 调用链

```
RobotEngine.AnalyzeMarket()
  ↓
MarketServiceManager.GetMarketState()
  ↓
MarketAnalyzer.GetAnalysis()
  ↓
MarketAnalyzer.analyzeMarket()  ← 实际计算在这里
  ↓
MarketAnalyzer.determineMarketState()  ← 判定市场状态
```

---

## 九、总结

### 9.1 实际计算方式

✅ **周期**：只使用3个周期（5m, 15m, 1h）
✅ **技术指标**：EMA12, EMA26, MACD, ATR
✅ **阈值**：**固定阈值**（highVol=2.0, lowVol=0.5）
✅ **判定逻辑**：先判断波动率，再判断趋势一致性

### 9.2 关键差异

❌ **不是**：5个周期（1m, 5m, 15m, 1h, 1d）
❌ **不是**：自适应阈值（基于基准波动率）
❌ **不是**：加权投票机制

### 9.3 文档修正

**之前的文档可能描述的是 `RobotAnalyzer` 的逻辑，但实际使用的是 `MarketAnalyzer`**。

**实际代码使用的是**：
- 3个周期（5m, 15m, 1h）
- 固定阈值（2.0, 0.5）
- 先波动率后趋势一致性的判定逻辑

