# 机器人详情页面策略参数获取逻辑修复

## 🔍 问题分析

### 修改前的问题

**机器人详情页面（View接口）**：
- 直接从数据库查询机器人实体
- 返回数据库中的静态字段：`Leverage`, `MarginPercent`, `StopLossPercent`, `RiskPreference`, `MarketState`
- **问题**：显示的是创建机器人时保存的静态值，不是实时计算的策略参数

**机器人页面（监控页面，buildConfigInfo）**：
- 从全局市场分析器获取市场状态（实时）
- 从映射关系获取风险偏好（根据市场状态动态获取）
- 从策略模板表查询策略参数（根据市场状态+风险偏好查询机器人绑定的策略组）
- **正确**：显示的是实时计算的策略参数

### 不一致的原因

1. **详情页面**：使用数据库静态字段
2. **监控页面**：使用实时计算的策略参数

导致两个页面显示的策略参数不一致。

## ✅ 修复方案

### 修改内容

**文件**: `internal/logic/trading/robot.go`
**函数**: `View()`

**修改逻辑**：
- 如果机器人正在运行（`status == 2`），使用和监控页面相同的逻辑获取实时策略参数
- 确保详情页面和监控页面显示的策略参数一致

### 修改后的数据来源

#### 1. 市场状态：全局市场分析器（实时）
```go
// 【1】市场状态：从全局市场分析器获取（实时）
globalAnalysis := market.GetMarketAnalyzer().GetAnalysis(engine.Platform, out.Symbol)
if globalAnalysis != nil {
    marketState := string(globalAnalysis.MarketState)
    if marketState != "" {
        out.MarketState = marketState
    }
}
```

#### 2. 风险偏好：映射关系（根据市场状态从机器人创建时候保存的值中获取）
```go
// 【2】风险偏好：从映射关系获取（根据市场状态从机器人创建时候保存的值中获取）
engineStatus := engine.GetStatus()
riskPref := engineStatus.CurrentRiskPref  // 从映射关系中获取
out.RiskPreference = riskPref
```

#### 3. 策略参数：策略模板表（根据市场状态+风险偏好查询机器人绑定的策略组的具体策略）
```go
// 【3】策略参数：从策略模板表查询（根据市场状态+风险偏好查询机器人绑定的策略组的具体策略）
params, err := engine.LoadFullStrategyParams(ctx, marketState, riskPref)
if err == nil && params != nil {
    // 更新策略参数
    out.Leverage = (params.LeverageMin + params.LeverageMax) / 2
    out.MarginPercent = (params.MarginPercentMin + params.MarginPercentMax) / 2
    out.StopLossPercent = params.StopLossPercent
    out.AutoStartRetreatPercent = params.AutoStartRetreatPercent
    out.ProfitRetreatPercent = params.ProfitRetreatPercent
}
```

## 📊 对比

| 页面 | 修改前 | 修改后 |
|------|--------|--------|
| **详情页面** | 数据库静态字段 | ✅ 实时计算的策略参数（与监控页面一致） |
| **监控页面** | 实时计算的策略参数 | ✅ 实时计算的策略参数（保持不变） |

## 🎯 修复效果

### 修改前
- ❌ 详情页面显示：数据库静态字段（创建时保存的值）
- ✅ 监控页面显示：实时计算的策略参数
- ❌ **不一致**

### 修改后
- ✅ 详情页面显示：实时计算的策略参数（运行中时）
- ✅ 监控页面显示：实时计算的策略参数
- ✅ **一致**

## 📝 代码位置

### 修改的文件
- `internal/logic/trading/robot.go` - `View()` 函数（504-600行）

### 参考的实现
- `internal/logic/trading/monitor.go` - `buildConfigInfo()` 函数（477-596行）

## 🔄 数据流向

```
机器人详情页面 View 接口
    ↓
查询数据库（获取基础信息）
    ↓
如果机器人正在运行（status == 2）
    ↓
【1】从全局市场分析器获取市场状态（实时）
    ↓
【2】从映射关系获取风险偏好（根据市场状态）
    ↓
【3】从策略模板表查询策略参数（根据市场状态+风险偏好+策略组ID）
    ↓
更新返回的字段值
    ↓
返回给前端（与监控页面一致）
```

## ✅ 验证

1. ✅ 代码已编译通过
2. ✅ 使用和监控页面相同的逻辑
3. ✅ 确保数据一致性

## 🎯 总结

**修复内容**：
- 机器人详情页面（View接口）现在使用和监控页面相同的逻辑获取策略参数
- 确保两个页面显示的策略参数一致
- 数据来源统一：
  1. 市场状态：全局市场分析器（实时）
  2. 风险偏好：映射关系（根据市场状态从机器人创建时候保存的值中获取）
  3. 策略参数：策略模板表（根据市场状态+风险偏好查询机器人绑定的策略组的具体策略）

