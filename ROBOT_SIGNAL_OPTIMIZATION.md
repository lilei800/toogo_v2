# 机器人实时信号优化说明

## 📋 优化概述

本次优化主要针对机器人实时信号生成逻辑，实现了完整的信号分析流程，包括市场状态分析、风险偏好映射、策略配置获取、方向判断和预警日志生成。

## 🎯 优化目标

1. **分析市场状态获取逻辑** - 自动或手动获取市场状态
2. **根据风险配置的映射获得风险偏好** - 根据市场状态映射风险偏好
3. **从策略管理获取时间窗口和波动值** - 根据市场状态和风险偏好获取策略配置
4. **优化方向判断逻辑** - 根据时间窗口和波动点数判断交易方向
5. **生成方向预警日志** - 记录方向信号变化

## 🔧 实现细节

### 1. 市场状态分析

**位置**: `server/internal/logic/toogo/engine.go` - `analyzeMarket()` 函数

**逻辑**:
- 获取K线数据进行分析（默认100根1分钟K线）
- 计算波动率和价格范围
- 根据波动率百分比自动判断市场状态：
  - `high_vol`: 波动率 > 3%
  - `low_vol`: 波动率 < 0.5%
  - `trend`: 波动率在0.5%-3%且趋势明显
  - `volatile`: 其他情况

**代码片段**:
```go
// 自动判断市场状态
if runner.Robot.AutoMarketState == 1 || runner.Robot.MarketState == "" {
    if volatilityPercent > 3 {
        signal.MarketState = "high_vol"
    } else if volatilityPercent < 0.5 {
        signal.MarketState = "low_vol"
    } else if volatility > 0.01 {
        signal.MarketState = "trend"
    } else {
        signal.MarketState = "volatile"
    }
}
```

### 2. 风险偏好映射

**逻辑**:
- 根据市场状态自动映射风险偏好：
  - `trend` → `balanced`
  - `volatile` → `conservative`
  - `high_vol` → `conservative`
  - `low_vol` → `balanced`

**代码片段**:
```go
// 根据市场状态映射风险偏好
switch signal.MarketState {
case "trend":
    signal.RiskLevel = "balanced"
case "volatile":
    signal.RiskLevel = "conservative"
case "high_vol":
    signal.RiskLevel = "conservative"
case "low_vol":
    signal.RiskLevel = "balanced"
}
```

### 3. 策略配置获取

**优先级顺序**:
1. **策略模板** - 根据市场状态和风险偏好从 `hg_toogo_strategy_template` 表获取
2. **机器人配置** - 从机器人的 `current_strategy` JSON配置中获取
3. **无配置处理** - 如果都没有配置，记录警告并返回NONE信号（不生成交易信号）

**注意**: 不再使用默认值，必须配置策略模板或机器人策略才能生成交易信号。

**代码片段**:
```go
// 优先从策略模板中获取（根据市场状态和风险偏好）
strategyTemplate, err := service.ToogoStrategy().GetByCondition(ctx, &toogoin.GetStrategyByConditionInp{
    MarketState:     signal.MarketState,
    RiskPreference:  signal.RiskLevel,
})
if err == nil && strategyTemplate != nil {
    monitorWindow = strategyTemplate.TimeWindow
    volatilityThreshold = strategyTemplate.VolatilityPoints
}

// 如果策略模板中没有，尝试从机器人的current_strategy JSON配置中获取
if monitorWindow == 0 || volatilityThreshold == 0 {
    strategyConfig := e.parseStrategyConfig(ctx, runner)
    // ... 从机器人配置获取
}

// 如果还是没有配置，记录警告并返回NONE信号（不生成交易信号）
if monitorWindow == 0 || volatilityThreshold == 0 {
    g.Log().Warningf(ctx, "[Signal] 策略配置缺失: robot=%d, 无法生成信号", runner.Robot.Id)
    signal.Direction = "NONE"
    signal.Reason = "策略配置缺失,请配置策略模板或机器人策略"
    return signal
}
```

### 4. 方向判断逻辑（核心优化）

**判断规则**:
- 在时间窗口的时间范围内保持最高价和最低价
- **做空条件**: 最高价减去实时价格的值 ≥ 波动值
- **做多条件**: 实时价格减去最低价的值 ≥ 波动值

**实现**:
```go
// 在时间窗口的时间范围内保持最高价和最低价
windowHigh := 0.0
windowLow := math.MaxFloat64
for _, k := range klines {
    if k.High > windowHigh {
        windowHigh = k.High
    }
    if k.Low < windowLow {
        windowLow = k.Low
    }
}

// 计算距离最高价和最低价的差值
distanceFromHigh := windowHigh - currentPrice // 距离最高价的距离
distanceFromLow := currentPrice - windowLow   // 距离最低价的距离

// 判断方向
if distanceFromHigh >= volatilityThreshold {
    // 最高价减去实时价格的值≥波动值 → 做空
    signal.Direction = "SHORT"
    signal.Strength = math.Min(1.0, distanceFromHigh/volatilityThreshold)
} else if distanceFromLow >= volatilityThreshold {
    // 实时价格减去最低价的值≥波动值 → 做多
    signal.Direction = "LONG"
    signal.Strength = math.Min(1.0, distanceFromLow/volatilityThreshold)
}
```

### 5. 方向预警日志生成

**功能**:
- 当方向信号产生时，自动记录方向预警日志
- 日志包含：机器人ID、用户ID、平台、交易对、方向、强度、置信度、建议操作等

**实现**:
```go
// logDirectionAlert 记录方向预警日志
func (e *TradingEngine) logDirectionAlert(ctx context.Context, runner *RobotRunner, signal *MarketSignal, 
    currentPrice, windowHigh, windowLow, distanceFromHigh, distanceFromLow float64) {
    
    alertLogger := market.GetAlertLogger()
    entry := &market.DirectionLogEntry{
        RobotId:       runner.Robot.Id,
        UserId:        runner.Robot.UserId,
        Platform:      runner.Exchange.GetName(),
        Symbol:        runner.Robot.Symbol,
        NewDirection:  signal.Direction,
        Strength:      signal.Strength * 100,
        Confidence:    confidence,
        Action:        action,
        EntryPrice:    currentPrice,
        Reason:        signal.Reason,
        CreatedAt:     time.Now(),
    }
    alertLogger.LogDirection(entry)
}
```

## 📊 数据流

```
获取K线数据
    ↓
分析市场状态
    ↓
映射风险偏好
    ↓
获取策略配置（时间窗口、波动值）
    ↓
计算窗口内最高价和最低价
    ↓
判断方向（距离最高/最低价是否≥波动值）
    ↓
生成方向信号
    ↓
记录预警日志
    ↓
返回信号（用于自动下单）
```

## 🔍 关键改进点

1. **完整的信号分析流程** - 从市场状态到方向判断的完整链路
2. **策略配置动态获取** - 根据市场状态和风险偏好动态获取策略参数
3. **精确的方向判断** - 基于时间窗口内的最高/最低价和波动值进行判断
4. **预警日志记录** - 自动记录方向信号变化，便于分析和追踪
5. **代码结构优化** - 清晰的步骤划分，易于维护和扩展

## 📝 使用说明

### 配置要求

1. **机器人配置**:
   - `use_monitor_signal = 1` - 启用信号监控
   - `auto_market_state = 1` - 自动市场状态（可选）
   - `auto_risk_preference = 1` - 自动风险偏好（可选）

2. **策略模板配置**:
   - 在 `hg_toogo_strategy_template` 表中配置策略模板
   - 包含：市场状态、风险偏好、时间窗口、波动点数等

3. **预警日志**:
   - 方向预警日志自动记录到 `hg_trading_direction_log` 表
   - 可通过预警日志接口查询

## 🚀 后续优化建议

1. **信号强度计算优化** - 可以根据距离波动值的倍数计算更精确的信号强度
2. **多周期分析** - 可以结合多个时间周期的信号进行综合判断
3. **信号过滤** - 可以添加信号过滤机制，避免频繁交易
4. **历史信号追踪** - 可以记录历史信号，用于回测和优化

## 📚 相关文件

- `server/internal/logic/toogo/engine.go` - 交易引擎核心逻辑
- `server/internal/logic/toogo/strategy.go` - 策略管理逻辑
- `server/internal/library/market/alert_logger.go` - 预警日志记录器
- `server/internal/model/entity/toogo_strategy_template.go` - 策略模板实体
- `server/internal/model/entity/trading_alert_logs.go` - 预警日志实体

---

**优化完成时间**: 2024年
**优化版本**: v1.0
**优化人员**: AI Assistant

