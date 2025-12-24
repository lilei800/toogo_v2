# 机器人详情页面数据来源说明

## 📍 位置
**运行中的机器人余额框下面，自动下单开关上面显示的数据**

## 🔍 数据来源

### API接口
- **接口**: `/admin/trading/robot/monitor` 或相关监控接口
- **实现文件**: `internal/logic/trading/monitor.go`
- **函数**: `buildConfigInfo()`

### 数据获取流程

```
1. 优先从 RobotEngine 获取实时数据
   ↓
2. 从全局市场分析器获取当前市场状态
   ↓
3. 从映射关系获取风险偏好
   ↓
4. 从数据库加载策略参数
```

## 📊 显示的数据字段

### 1. 市场状态 (MarketState)
**数据来源**: **全局市场分析器（实时）**

**获取方式**:
```go
// 【1】市场状态：从全局市场分析器获取（实时）
globalAnalysis := market.GetMarketAnalyzer().GetAnalysis(engine.Platform, robot.Symbol)
if globalAnalysis != nil {
    marketState := string(globalAnalysis.MarketState)
    if marketState != "" {
        config.MarketState = marketState
    }
}
```

**代码位置**: `monitor.go:510-518`

**说明**:
- ✅ 直接从全局市场分析器获取实时市场状态
- ✅ 每5秒更新一次，确保数据实时性
- ⚠️ 如果全局市场分析器没有数据，降级使用引擎状态

### 2. 风险偏好 (RiskPreference)
**数据来源**: **映射关系（根据市场状态从机器人创建时候保存的值中获取）**

**获取方式**:
```go
// 【2】风险偏好：从映射关系获取（根据市场状态从机器人创建时候保存的值中获取）
// 映射关系存储在 RobotEngine.MarketRiskMapping 中，是创建机器人时保存的
// 通过 GetStatus() 方法获取，该方法会从映射关系中根据当前市场状态获取风险偏好
engineStatus := engine.GetStatus()
riskPref := engineStatus.CurrentRiskPref

if riskPref == "" {
    // 【严格模式】映射关系中没有找到，记录错误但不阻塞显示
    // 降级：使用数据库的默认值
    riskPref = robot.RiskPreference
    if riskPref == "" {
        riskPref = "balanced" // 最后兜底
    }
}
```

**代码位置**: `monitor.go:520-540`

**说明**:
- ✅ 从机器人创建时保存的映射关系（`MarketRiskMapping`）中获取
- ✅ 根据当前市场状态动态查找对应的风险偏好
- ✅ 映射关系存储在 `Robot.Remark` 字段中（JSON格式）
- ⚠️ 如果映射关系中没有找到，降级使用数据库的默认值

### 3. 策略参数（杠杆、保证金、止损、止盈等）
**数据来源**: **策略模板表（根据市场状态+风险偏好查询机器人绑定的策略组的具体策略）**

**获取方式**:
```go
// 【3】策略参数：从策略模板表查询（根据市场状态+风险偏好查询机器人绑定的策略组的具体策略）
// LoadFullStrategyParams 会使用机器人绑定的策略组ID（robot.StrategyGroupId）查询策略模板表
params, err := engine.LoadFullStrategyParams(ctx, marketState, riskPref)
```

**代码位置**: `monitor.go:542-561`

**查询逻辑** (`robot_engine.go:833-948`):
```go
// 1. 获取策略组ID（优先级：机器人.StrategyGroupId > CurrentStrategy.group_id）
groupId := e.Robot.StrategyGroupId

// 2. 从策略模板表中查询对应的策略
// 查询条件：策略组ID + 市场状态 + 风险偏好
dao.TradingStrategyTemplate.Ctx(ctx).
    Where("group_id", groupId).
    Where("market_state", marketState).
    Where("risk_preference", riskPreference).
    Scan(&strategy)
```

**说明**:
- ✅ 使用机器人绑定的策略组ID（`robot.StrategyGroupId`）查询
- ✅ 根据市场状态和风险偏好查询策略模板表
- ✅ 每次请求时从数据库重新加载，确保获取最新的策略模板数据
- ⚠️ 如果加载失败，降级使用引擎缓存的参数

```go
// 重新从数据库加载策略参数（确保获取最新数据）
params, err := engine.LoadFullStrategyParams(ctx, marketState, riskPref)
if err != nil {
    g.Log().Errorf(ctx, "[Monitor] robotId=%d 策略参数加载失败: %v", robot.Id, err)
    // 如果加载失败，尝试使用引擎缓存的参数（降级方案）
    if engine.CurrentStrategyParams != nil {
        params = engine.CurrentStrategyParams
    }
}

if params != nil {
    // 使用杠杆范围的中间值
    if params.LeverageMax > 0 {
        config.Leverage = (params.LeverageMin + params.LeverageMax) / 2
    } else if params.LeverageMin > 0 {
        config.Leverage = params.LeverageMin
    }
    // 使用保证金比例范围的中间值
    if params.MarginPercentMax > 0 {
        config.MarginPercent = (params.MarginPercentMin + params.MarginPercentMax) / 2
    } else if params.MarginPercentMin > 0 {
        config.MarginPercent = params.MarginPercentMin
    }
    // 止损和止盈回撤
    if params.StopLossPercent > 0 {
        config.StopLossPercent = params.StopLossPercent
    }
    if params.AutoStartRetreatPercent > 0 {
        config.AutoStartRetreat = params.AutoStartRetreatPercent
    }
    if params.ProfitRetreatPercent > 0 {
        config.TakeProfitPercent = params.ProfitRetreatPercent
    }
}
```

## 🔄 数据更新机制

### 实时更新
- **市场状态**: 从全局市场分析器实时获取（每5秒更新一次）
- **风险偏好**: 从映射关系实时获取（根据当前市场状态）
- **策略参数**: 每次请求时从数据库重新加载（确保最新）

### 数据流向

```
全局市场分析器 (实时)
    ↓
RobotEngine.GetStatus()
    ↓
buildConfigInfo() 函数
    ↓
前端显示
```

## 📝 关键代码位置

### 1. 数据构建函数
**文件**: `internal/logic/trading/monitor.go`
**函数**: `buildConfigInfo()`
**行数**: 477-570

### 2. 引擎状态获取
**文件**: `internal/logic/toogo/robot_engine.go`
**函数**: `GetStatus()`
**行数**: 2132-2252

### 3. 策略参数加载
**文件**: `internal/logic/toogo/robot_engine.go`
**函数**: `LoadFullStrategyParams()`
**行数**: 约 600-900

## 🎯 总结

### 数据来源（已优化）

**显示的数据来源**:
1. ✅ **市场状态**: 全局市场分析器（实时）
   - 直接从 `market.GetMarketAnalyzer().GetAnalysis()` 获取
   - 每5秒更新一次，确保数据实时性

2. ✅ **风险偏好**: 映射关系（根据市场状态从机器人创建时候保存的值中获取）
   - 从 `RobotEngine.MarketRiskMapping` 中获取
   - 映射关系是创建机器人时保存的，存储在 `Robot.Remark` 字段中
   - 根据当前市场状态动态查找对应的风险偏好

3. ✅ **策略参数**: 策略模板表（根据市场状态+风险偏好查询机器人绑定的策略组的具体策略）
   - 使用机器人绑定的策略组ID（`robot.StrategyGroupId`）查询
   - 根据市场状态和风险偏好查询策略模板表
   - 每次请求时从数据库重新加载，确保获取最新的策略模板数据

### 特点

- ✅ **实时性**: 市场状态从全局市场分析器实时获取
- ✅ **准确性**: 风险偏好从创建时保存的映射关系中获取，策略参数每次从数据库重新加载
- ✅ **一致性**: 使用机器人绑定的策略组ID，确保查询的是正确的策略组
- ✅ **降级机制**: 如果实时数据获取失败，使用数据库缓存值

### 重要提示

- 这些数据**不是**从机器人表的静态字段获取的
- 市场状态是**实时计算**的，从全局市场分析器获取
- 风险偏好是**动态查找**的，根据市场状态从创建时保存的映射关系中获取
- 策略参数来自**策略模板表**，使用机器人绑定的策略组ID查询

