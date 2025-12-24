# hotgo_v2 完整交易逻辑分析报告

> 生成时间：2024-12-24
> 
> 分析范围：方向信号生成 → 自动下单 → 自动平仓 完整流程

---

## 一、方向信号生成逻辑 📊

### 1.1 核心算法：窗口价格监控

**位置：** `robot_engine.go` - `EvaluateWindowSignal()` 函数（第2844-3060行）

**核心思想：**
在时间窗口内监控最高价和最低价，当价格偏离达到阈值时触发信号。

#### 信号判断规则：

```go
// 计算距离
distanceFromMax := maxPrice - currentPrice  // 最高价 - 实时价格
distanceFromMin := currentPrice - minPrice  // 实时价格 - 最低价

// 做空信号：价格从高点回落
shortTriggered := distanceFromMax >= threshold

// 做多信号：价格从低点反弹
longTriggered := distanceFromMin >= threshold
```

#### 信号类型：

1. **做多信号（LONG）**
   - 条件：`实时价格 - 窗口最低价 >= 波动阈值`
   - 含义：价格从低点反弹，上涨趋势
   - Action: `OPEN_LONG`

2. **做空信号（SHORT）**
   - 条件：`窗口最高价 - 实时价格 >= 波动阈值`
   - 含义：价格从高点回落，下跌趋势
   - Action: `OPEN_SHORT`

3. **中性信号（NEUTRAL）**
   - 条件：未达到波动阈值
   - Action: `HOLD`

4. **特殊情况：双向触发**
   - 条件：做多和做空同时触发（窗口范围 >= 2×阈值）
   - 处理：重置预警标记，继续监控
   - 原因：价格剧烈波动，避免误判

### 1.2 信号优化机制

#### 防重复触发：
```go
// 只在信号方向变化时保存预警记录
isNewDirection := newSignal != e.LastWindowSignal

if isNewDirection {
    // 检查是否已有该方向持仓
    if e.HasActivePosition(positionSide) {
        // 已有持仓，跳过
    } else {
        // 保存预警记录并触发下单
        logId := e.saveSignalAlertSimple(&signalCopy)
        if logId > 0 {
            go e.Trader.TryAutoTradeAndUpdate(ctx, &signalCopy, logId)
        }
    }
}
```

#### 关键优化点：

1. **每个信号只保存一次**
   - LONG信号持续存在时，不会重复保存预警记录
   - 只在 `NEUTRAL → LONG` 或 `SHORT → LONG` 时触发

2. **内存缓存持仓检查**
   - 在保存预警记录前检查内存中是否已有持仓
   - 解决 `LONG → NEUTRAL → LONG` 导致的重复下单问题

3. **窗口基准价变化检测**
   - 窗口最低价/最高价变化时，重置预警标记
   - 允许在新的价格区间重新触发信号

### 1.3 信号进度计算

对于未触发的信号，计算接近度：

```go
longProgress := (distanceFromMin / threshold) * 100   // 做多进度
shortProgress := (distanceFromMax / threshold) * 100  // 做空进度

signal.SignalProgress = math.Min(100, max(longProgress, shortProgress))
```

**用途：** 前端展示"血条"，让用户了解距离触发还有多远

---

## 二、自动下单逻辑 💰

### 2.1 触发流程

```
信号检测 → 保存预警记录 → 异步触发下单
    ↓
EvaluateWindowSignal()
    ↓
saveSignalAlertSimple()  // 保存到 hg_trading_signal_log
    ↓
go TryAutoTradeAndUpdate()  // 异步执行下单逻辑
```

### 2.2 下单执行逻辑

**位置：** `robot_engine.go` - `TryAutoTradeAndUpdate()` 函数（第4248-4800行）

#### 执行步骤：

##### **步骤1：基础检查**
```go
// 1. 机器人存在性检查
if robot == nil { return }

// 2. 信号有效性检查
if signal == nil || signal.Direction == "NEUTRAL" { return }

// 3. 只处理开仓信号
if signal.Action != "OPEN_LONG" && signal.Action != "OPEN_SHORT" { return }
```

##### **步骤2：防重复下单（原子操作）**
```go
// 使用数据库原子更新标记预警记录为已处理
result := g.DB().Model("hg_trading_signal_log").
    Where("id", logId).
    Where("(is_processed = 0 OR is_processed IS NULL)").
    Update(g.Map{"is_processed": 1})

rowsAffected, _ := result.RowsAffected()
if rowsAffected == 0 {
    // 已被其他goroutine处理，跳过
    return
}
```

**关键点：**
- ✅ 使用 `WHERE is_processed=0` 确保只更新未处理的记录
- ✅ 检查 `RowsAffected` 判断是否成功标记
- ✅ PostgreSQL 完全兼容
- ✅ 防止并发重复下单

##### **步骤3：自动交易开关检查**
```go
if robot.AutoTradeEnabled != 1 {
    return  // 自动下单未开启
}
```

##### **步骤4：持仓检查（核心逻辑）**

```go
// 使用智能缓存获取交易所实时持仓（1秒内缓存有效）
positions, err := t.engine.GetPositionsSmart(ctx, 1*time.Second)

// 检查持仓情况
hasAnyPosition := false        // 是否有任何持仓
hasSameSidePosition := false   // 是否有同方向持仓

for _, pos := range positions {
    if math.Abs(pos.PositionAmt) > 0.0001 {
        hasAnyPosition = true
        if pos.PositionSide == positionSide {
            hasSameSidePosition = true
        }
    }
}
```

**持仓规则：**

**单向模式（`DualSidePosition = 0`）：**
```go
if robot.DualSidePosition == 0 && hasAnyPosition {
    // 已有任何方向的持仓，拒绝新开仓
    // 原因：持仓内只能有一单
    return
}
```

**双向模式（`DualSidePosition = 1`）：**
```go
if robot.DualSidePosition == 1 && hasSameSidePosition {
    // 同方向已有持仓，拒绝新开仓（禁止加仓）
    // 反方向可以开仓（允许多空同持）
    return
}
```

##### **步骤5：下单锁机制**
```go
// TryLock 机制，最多重试5次
for i := 0; i < 5; i++ {
    if t.engine.orderLock.TryLock() {
        locked = true
        break
    }
    time.Sleep(10 * time.Millisecond)
}

if !locked {
    return  // 系统繁忙
}
defer t.engine.orderLock.Unlock()
```

##### **步骤6：获取锁后二次持仓检查**
```go
// 防止并发下单穿透
// 使用刚刚更新的内存缓存，不再调用API
positionsAgain := t.engine.CurrentPositions

// 再次检查持仓规则（逻辑同步骤4）
```

**为什么需要二次检查？**
- 时间线：线程A检查持仓（无）→ 线程B检查持仓（无）→ 线程A下单 → 线程B下单（重复！）
- 锁机制：线程A获取锁 → 下单 → 释放锁 → 线程B获取锁 → **二次检查发现已有持仓** → 取消下单 ✅

##### **步骤7：执行下单**
```go
// 计算下单参数（策略模板参数）
// 调用交易所API下单
// 保存订单到数据库
// 更新预警记录执行状态
```

### 2.3 下单参数计算

**策略参数来源：** 根据市场状态（`high_vol/low_vol/trend`）从策略模板加载

**关键参数：**
- `OrderMargin`: 下单保证金（USDT）
- `Leverage`: 杠杆倍数
- `StopLossPercent`: 止损百分比
- `AutoStartRetreatPercent`: 启动止盈百分比
- `ProfitRetreatPercent`: 止盈回撤百分比

### 2.4 执行日志记录

**保存位置：** `hg_trading_execution_log` 表

**关键字段：**
- `alert_log_id`: 关联预警记录ID
- `order_id`: 订单ID（成功时）
- `event_type`: 事件类型（`order_created`, `order_failed`）
- `status`: 执行状态（`success`, `failed`）
- `message`: 执行消息
- `event_data`: 详细数据（JSON）

---

## 三、自动平仓逻辑 🎯

### 3.1 平仓功能状态

根据代码检查：

❌ **旧的自动平仓逻辑已删除**
```go
// 位置：engine/core.go 第243、261、263行
// doTradingCheck 交易检查
func (e *RobotEngine) doTradingCheck(ctx context.Context) {
    // 【已删除】自动平仓检查已删除
    // go e.checkClosePosition(ctx)
}

// checkClosePosition 检查是否应该平仓
// 【已删除】自动平仓功能已删除，此函数不再执行任何操作
func (e *RobotEngine) checkClosePosition(ctx context.Context) {
    // 自动平仓功能已删除
    return
}
```

✅ **新的平仓机制：止损 + 止盈回撤**

### 3.2 止损逻辑

**位置：** `robot_engine.go` - `checkStopLossAndClose()` 函数（第1924-2020行）

#### 触发时机：
```go
// 在价格更新时调用
func (e *RobotEngine) OnPriceUpdate(ctx context.Context, price float64) {
    // ...
    e.checkStopLossAndClose(ctx, price)
    e.checkTakeProfitAndClose(ctx, price)
}
```

#### 止损检查流程：

```go
// 1. 检查自动平仓开关
if robot.AutoCloseEnabled != 1 {
    return  // 自动平仓未开启
}

// 2. 检查是否有持仓
if len(positions) == 0 {
    return
}

// 3. 获取止损参数（从策略模板）
stopLossPercent := strategyParams.StopLossPercent
if stopLossPercent <= 0 {
    return  // 未设置止损
}

// 4. 遍历每个持仓
for _, pos := range positions {
    // 4.1 基于实时价格计算未实现盈亏
    if pos.PositionSide == "LONG" {
        realTimeUnrealizedPnl = (currentPrice - pos.EntryPrice) * |pos.PositionAmt|
    } else {
        realTimeUnrealizedPnl = (pos.EntryPrice - currentPrice) * |pos.PositionAmt|
    }
    
    // 4.2 只有亏损时才检查止损
    if realTimeUnrealizedPnl >= 0 {
        continue
    }
    
    // 4.3 计算保证金
    margin = |pos.PositionAmt| × pos.EntryPrice / robot.Leverage
    
    // 4.4 计算止损进度
    stopLossAmount = margin × (stopLossPercent / 100)
    progress = (|realTimeUnrealizedPnl| / stopLossAmount) × 100%
    
    // 4.5 如果止损进度达到100%，立即执行平仓
    if progress >= 100.0 {
        e.executeStopLossCloseByPosition(ctx, pos)
    }
}
```

#### 止损计算示例：

假设：
- 保证金：1000 USDT
- 止损百分比：10%
- 当前未实现盈亏：-120 USDT

计算：
- 止损金额 = 1000 × 0.1 = 100 USDT
- 止损进度 = 120 / 100 × 100% = **120%** ✅ 触发平仓

### 3.3 止盈回撤逻辑

**位置：** `robot_engine.go` - `checkTakeProfitAndClose()` 函数（第2230-2431行）

#### 核心机制：**追踪最高盈利 + 回撤平仓**

```
开仓 → 持续盈利 → 达到启动条件 → 启动止盈回撤 → 追踪最高盈利 → 回撤触发 → 平仓
```

#### 详细流程：

##### **阶段1：监控盈利（未启动止盈）**

```go
// 获取或创建持仓跟踪器（内存）
tracker := e.PositionTrackers[pos.PositionSide]
if tracker == nil {
    tracker = &PositionTracker{
        PositionSide:      pos.PositionSide,
        EntryMargin:       margin,
        EntryTime:         time.Now(),
        HighestProfit:     0,
        TakeProfitEnabled: false,  // 初始未启动
    }
    e.PositionTrackers[pos.PositionSide] = tracker
}

// 持续更新最高盈利
if realTimeUnrealizedPnl > tracker.HighestProfit {
    tracker.HighestProfit = realTimeUnrealizedPnl
}
```

##### **阶段2：自动启动止盈回撤**

```go
// 检查是否满足启动条件
if !tracker.TakeProfitEnabled && autoStartPercent > 0 {
    currentProfitPercent := (realTimeUnrealizedPnl / margin) × 100%
    
    if currentProfitPercent >= autoStartPercent {
        // 自动启动止盈回撤
        tracker.TakeProfitEnabled = true
        tracker.HighestProfit = realTimeUnrealizedPnl
        
        // 持久化到数据库（支持服务重启后继续）
        go e.persistProfitRetreatStarted(ctx, pos.PositionSide, tracker.HighestProfit)
        
        g.Log().Infof("【自动启动】止盈回撤: currentProfitPercent=%.2f%% >= autoStartPercent=%.2f%%",
            currentProfitPercent, autoStartPercent)
    }
}
```

**启动条件示例：**
- 保证金：1000 USDT
- 启动止盈百分比：5%
- 当前盈利：60 USDT
- 盈利百分比 = 60 / 1000 × 100% = 6% >= 5% ✅ **自动启动**

##### **阶段3：追踪最高盈利**

```go
// 持续更新最高盈利（只增不减）
if realTimeUnrealizedPnl > tracker.HighestProfit {
    tracker.HighestProfit = realTimeUnrealizedPnl
    g.Log().Infof("更新最高盈利: %.4f", tracker.HighestProfit)
}
```

##### **阶段4：检查回撤触发**

```go
// 计算当前回撤百分比
currentRetreatPercent = ((tracker.HighestProfit - realTimeUnrealizedPnl) / tracker.HighestProfit) × 100%

// 计算血条百分比（供前端展示）
bloodBarPercent = 100% - (currentRetreatPercent / profitRetreatPercent) × 100%

// 触发平仓条件
if currentRetreatPercent >= profitRetreatPercent {
    g.Log().Warningf("【触发止盈】止盈回撤达到阈值，立即执行平仓")
    e.executeTakeProfitCloseByPosition(ctx, pos, "take_profit")
}
```

#### 止盈回撤计算示例：

**场景：**
- 保证金：1000 USDT
- 启动止盈：5%（50 USDT）
- 止盈回撤：30%
- 最高盈利：100 USDT
- 当前盈利：70 USDT

**计算：**
1. 盈利百分比 = 100 / 1000 × 100% = 10% ✅ 已启动止盈（>5%）
2. 回撤百分比 = (100 - 70) / 100 × 100% = **30%** ✅ **触发平仓**
3. 血条 = 100% - (30% / 30%) × 100% = **0%** ⚠️ 血条耗尽

**血条理解：**
- 100%：当前盈利 = 最高盈利（没有回撤）
- 50%：回撤了一半的允许回撤量
- 0%：回撤达到阈值，触发平仓

### 3.4 持仓跟踪器（PositionTracker）

**作用：** 内存中追踪每个持仓的止盈状态

```go
type PositionTracker struct {
    PositionSide      string    // 持仓方向（LONG/SHORT）
    EntryMargin       float64   // 开仓保证金
    EntryTime         time.Time // 开仓时间
    HighestProfit     float64   // 最高盈利（USDT）
    TakeProfitEnabled bool      // 是否已启动止盈回撤
}
```

**持久化机制：**
- 启动止盈时：写入数据库 `hg_trading_order` 表的 `take_profit_enabled` 字段
- 服务重启时：从数据库恢复状态（`initTrackerFromDB()`）
- **不可关闭原则：** 一旦启动止盈回撤，无法手动关闭（只能平仓）

### 3.5 平仓执行

**止损平仓：**
```go
func (e *RobotEngine) executeStopLossCloseByPosition(ctx context.Context, pos *exchange.Position) {
    // 1. 防重复平仓检查（查询订单状态）
    // 2. 调用交易所API平仓
    // 3. 更新订单状态为"已平仓"
    // 4. 扣除算力
    // 5. 保存平仓日志
    // 6. 清除持仓跟踪器
}
```

**止盈平仓：**
```go
func (e *RobotEngine) executeTakeProfitCloseByPosition(ctx context.Context, pos *exchange.Position, reason string) {
    // 逻辑同止损平仓
    // reason: "take_profit" 或 "manual"
}
```

### 3.6 防重复平仓机制

```go
// 查询数据库中订单状态
var localOrder struct {
    Status int
}
err := dao.TradingOrder.Ctx(ctx).
    Where("robot_id", robot.Id).
    Where("direction", direction).
    Fields("status").
    Scan(&localOrder)

// 如果订单状态已经是"平仓中"或"已平仓"，跳过
if localOrder.Status == OrderStatusClosing || localOrder.Status == OrderStatusClosed {
    g.Log().Warningf("订单已在平仓中或已平仓，跳过重复平仓")
    return
}

// 先更新订单状态为"平仓中"
g.DB().Model("hg_trading_order").
    Where("id", localOrderId).
    Where("status", OrderStatusOpen).  // 只更新状态为"开仓"的订单
    Update(g.Map{"status": OrderStatusClosing})
```

---

## 四、完整流程图 🔄

### 4.1 开仓流程

```
价格更新 (WebSocket)
    ↓
OnPriceUpdate() - 更新价格窗口
    ↓
EvaluateWindowSignal() - 评估窗口信号
    ↓
检测信号方向变化？
    ↓ 是
检查内存是否已有该方向持仓？
    ↓ 否
saveSignalAlertSimple() - 保存预警记录
    ↓
TryAutoTradeAndUpdate() - 异步下单
    ↓
原子操作标记预警记录为已处理
    ↓
检查自动交易开关
    ↓
检查持仓规则（单向/双向模式）
    ↓
获取下单锁（TryLock，重试5次）
    ↓
二次持仓检查（防并发）
    ↓
计算下单参数（策略模板）
    ↓
调用交易所API下单
    ↓
保存订单到数据库
    ↓
保存执行日志
```

### 4.2 平仓流程（止损）

```
价格更新 (WebSocket)
    ↓
OnPriceUpdate() - 获取实时价格
    ↓
checkStopLossAndClose() - 检查止损
    ↓
检查自动平仓开关
    ↓
遍历持仓
    ↓
基于实时价格计算未实现盈亏
    ↓
只处理亏损持仓
    ↓
计算保证金和止损进度
    ↓
止损进度 >= 100%？
    ↓ 是
executeStopLossCloseByPosition()
    ↓
防重复平仓检查（订单状态）
    ↓
更新订单状态为"平仓中"
    ↓
调用交易所API平仓
    ↓
更新订单状态为"已平仓"
    ↓
扣除算力
    ↓
保存平仓日志
    ↓
清除持仓跟踪器
```

### 4.3 平仓流程（止盈回撤）

```
价格更新 (WebSocket)
    ↓
OnPriceUpdate() - 获取实时价格
    ↓
checkTakeProfitAndClose() - 检查止盈回撤
    ↓
检查自动平仓开关
    ↓
遍历持仓
    ↓
获取或创建持仓跟踪器（内存）
    ↓
基于实时价格计算未实现盈亏
    ↓
更新最高盈利
    ↓
未启动止盈？
    ↓ 是
计算当前盈利百分比
    ↓
当前盈利百分比 >= 启动止盈百分比？
    ↓ 是
【自动启动】止盈回撤
持久化到数据库
    ↓
已启动止盈 → 计算回撤百分比
    ↓
回撤百分比 = (最高盈利 - 当前盈利) / 最高盈利 × 100%
    ↓
回撤百分比 >= 止盈回撤百分比？
    ↓ 是
executeTakeProfitCloseByPosition()
    ↓
（逻辑同止损平仓）
```

---

## 五、关键特性总结 ⭐

### 5.1 方向信号

✅ **优点：**
1. 纯窗口逻辑，算法简单清晰
2. 防重复触发机制（信号方向变化时才触发）
3. 内存缓存持仓检查（解决重复下单问题）
4. 双向触发自动重置（避免剧烈波动误判）
5. 信号进度计算（前端展示"接近度"）

⚠️ **注意事项：**
1. 依赖窗口大小和波动阈值配置
2. 窗口过小：信号频繁，可能过度交易
3. 窗口过大：信号延迟，可能错过机会

### 5.2 自动下单

✅ **优点：**
1. **原子操作防重复下单**（数据库级别）
2. **双重持仓检查**（锁前 + 锁后）
3. **智能持仓缓存**（1秒内缓存有效，减少API调用）
4. **完整的执行日志**（每个步骤都有记录）
5. **PostgreSQL完全兼容**

✅ **持仓规则清晰：**
- 单向模式：持仓内只能有一单
- 双向模式：同方向只能一单（禁止加仓），允许多空同持

⚠️ **注意事项：**
1. 依赖 `AutoTradeEnabled` 开关
2. 依赖交易所API稳定性
3. 策略参数需要正确配置

### 5.3 自动平仓

✅ **优点：**
1. **基于实时价格计算盈亏**（不依赖交易所API数据，避免2分钟延迟）
2. **止盈回撤自动启动**（达到盈利阈值自动启动，无需手动）
3. **持仓跟踪器持久化**（支持服务重启后继续止盈回撤）
4. **防重复平仓机制**（订单状态 + 数据库原子更新）
5. **完整的平仓日志**（止损/止盈/手动平仓统一记录）

✅ **两层保护：**
- **第一层：止损**（防止超额亏损）
- **第二层：止盈回撤**（保护已有利润）

⚠️ **注意事项：**
1. 依赖 `AutoCloseEnabled` 开关
2. 止盈回撤一旦启动，无法手动关闭（设计原则）
3. 需要合理设置止盈回撤参数（避免过早平仓）

### 5.4 并发安全

✅ **多层保护机制：**
1. **预警记录原子标记**（`is_processed`字段 + WHERE条件）
2. **下单锁**（`orderLock.TryLock()`，最多重试5次）
3. **二次持仓检查**（获取锁后再次检查）
4. **平仓状态检查**（`OrderStatusClosing` 防重复平仓）

✅ **缓存优化：**
1. **智能持仓缓存**（`GetPositionsSmart`，1秒内有效）
2. **策略参数缓存**（60秒内有效）
3. **波动率配置缓存**（60秒内有效）

### 5.5 数据一致性

✅ **事件驱动架构：**
1. 价格更新 → WebSocket推送 → 立即检查止损/止盈
2. 不依赖定时轮询（避免延迟和资源浪费）

✅ **订单同步机制：**
1. 开仓后：异步同步订单状态
2. 平仓后：更新订单信息 + 扣除算力 + 保存日志
3. 手动平仓：交易所检测 + 同步到本地数据库

✅ **日志完整性：**
1. **预警日志**（`hg_trading_signal_log`）：记录信号
2. **执行日志**（`hg_trading_execution_log`）：记录下单/平仓执行
3. **订单日志**（`hg_trading_order`）：记录完整订单信息

---

## 六、配置参数说明 ⚙️

### 6.1 信号参数

| 参数 | 说明 | 来源 | 示例 |
|------|------|------|------|
| `window` | 时间窗口大小（秒） | 策略模板 | 300（5分钟） |
| `threshold` | 波动阈值（价格变动） | 策略模板 | 50（USDT） |

### 6.2 下单参数

| 参数 | 说明 | 来源 | 示例 |
|------|------|------|------|
| `OrderMargin` | 下单保证金（USDT） | 策略模板 | 100 |
| `Leverage` | 杠杆倍数 | 机器人配置 | 10 |
| `DualSidePosition` | 双向持仓模式 | 机器人配置 | 0/1 |
| `AutoTradeEnabled` | 自动交易开关 | 机器人配置 | 0/1 |

### 6.3 平仓参数

| 参数 | 说明 | 来源 | 示例 |
|------|------|------|------|
| `StopLossPercent` | 止损百分比 | 策略模板 | 10（10%） |
| `AutoStartRetreatPercent` | 启动止盈百分比 | 策略模板 | 5（5%） |
| `ProfitRetreatPercent` | 止盈回撤百分比 | 策略模板 | 30（30%） |
| `AutoCloseEnabled` | 自动平仓开关 | 机器人配置 | 0/1 |

### 6.4 市场状态映射

**位置：** 机器人 `remark` 字段（JSON格式）

```json
{
    "high_vol": "aggressive",   // 高波动 → 激进策略
    "low_vol": "conservative",  // 低波动 → 保守策略
    "trend": "balanced"         // 趋势 → 平衡策略
}
```

**作用：** 根据当前市场状态自动切换策略参数

---

## 七、常见问题与排查 🔍

### 7.1 信号检测到但没有下单

**可能原因：**
1. ❌ `AutoTradeEnabled` = 0（自动交易未开启）
2. ❌ 已有该方向的持仓（内存缓存检查）
3. ❌ 预警记录保存失败（`logId = 0`）
4. ❌ 原子标记失败（`RowsAffected = 0`，已被其他线程处理）
5. ❌ 持仓检查未通过（单向/双向模式限制）
6. ❌ 获取锁超时（系统繁忙）

**排查步骤：**
```sql
-- 1. 查看预警记录
SELECT * FROM hg_trading_signal_log 
WHERE robot_id = ? 
ORDER BY created_at DESC LIMIT 10;

-- 2. 查看执行日志
SELECT * FROM hg_trading_execution_log 
WHERE alert_log_id IN (
    SELECT id FROM hg_trading_signal_log WHERE robot_id = ?
) 
ORDER BY created_at DESC LIMIT 10;

-- 3. 查看机器人配置
SELECT auto_trade_enabled, dual_side_position 
FROM hg_trading_robot 
WHERE id = ?;
```

### 7.2 止损/止盈没有触发

**可能原因：**
1. ❌ `AutoCloseEnabled` = 0（自动平仓未开启）
2. ❌ 策略参数未正确加载（`CurrentStrategyParams` 为 nil）
3. ❌ 止损/止盈参数为0或未设置
4. ❌ 价格更新未触发（WebSocket断开）
5. ❌ 持仓数据未更新

**排查步骤：**
```sql
-- 1. 查看机器人配置
SELECT auto_close_enabled FROM hg_trading_robot WHERE id = ?;

-- 2. 查看策略参数
SELECT stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent 
FROM hg_trading_strategy_template 
WHERE id = (
    SELECT strategy_id FROM hg_trading_robot WHERE id = ?
);

-- 3. 查看持仓信息
SELECT * FROM hg_trading_order 
WHERE robot_id = ? AND status = 1;  -- status=1表示开仓中
```

### 7.3 重复下单

**如果发生重复下单，检查：**
1. ❌ 原子标记逻辑是否正确（`is_processed` 字段）
2. ❌ 二次持仓检查是否生效
3. ❌ 锁机制是否正常工作
4. ❌ 信号方向变化检查是否生效

**紧急处理：**
```sql
-- 手动标记预警记录为已处理
UPDATE hg_trading_signal_log 
SET is_processed = 1 
WHERE robot_id = ? AND is_processed = 0;
```

### 7.4 止盈回撤未启动

**可能原因：**
1. ❌ 盈利未达到启动条件（`autoStartPercent`）
2. ❌ 持仓跟踪器未创建
3. ❌ 策略参数中 `AutoStartRetreatPercent` = 0

**查看日志：**
```bash
# 搜索止盈相关日志
grep "止盈检查" robot_engine.log
grep "自动启动" robot_engine.log
```

---

## 八、优化建议 💡

### 8.1 性能优化

1. **持仓缓存优化**
   - ✅ 当前：1秒内缓存有效
   - 💡 建议：根据交易频率动态调整缓存时间

2. **数据库查询优化**
   - ✅ 当前：策略参数60秒缓存
   - 💡 建议：增加索引优化（`robot_id`, `status`, `direction`）

3. **日志优化**
   - ⚠️ 当前：大量 `Info` 级别日志
   - 💡 建议：生产环境降低日志级别为 `Warning`

### 8.2 功能增强

1. **信号确认机制**
   - 💡 增加信号持续时间要求（避免瞬间波动误判）
   - 💡 增加成交量确认（提高信号可靠性）

2. **风险控制**
   - 💡 增加单日最大亏损限制
   - 💡 增加连续止损次数限制（暂停交易）

3. **止盈策略**
   - 💡 支持多级止盈（部分平仓）
   - 💡 支持移动止损（跟随价格上涨）

### 8.3 监控告警

1. **异常监控**
   - 💡 API调用失败告警
   - 💡 订单状态异常告警
   - 💡 持仓数据不一致告警

2. **性能监控**
   - 💡 下单延迟监控
   - 💡 平仓延迟监控
   - 💡 缓存命中率监控

---

## 九、总结 📝

### 9.1 系统优点

✅ **架构清晰：** 信号生成 → 自动下单 → 自动平仓 三个模块独立
✅ **并发安全：** 多层保护机制，防止重复下单和重复平仓
✅ **实时性强：** 基于WebSocket价格推送，立即检查止损/止盈
✅ **日志完整：** 预警日志 + 执行日志 + 订单日志 三层记录
✅ **参数化配置：** 支持根据市场状态自动切换策略
✅ **数据库兼容：** PostgreSQL完全兼容

### 9.2 核心流程

```
价格更新（WebSocket）
    ↓
方向信号生成（EvaluateWindowSignal）
    ↓
自动下单（TryAutoTradeAndUpdate）
    ↓
持仓监控（PositionTracker）
    ↓
自动平仓（checkStopLossAndClose + checkTakeProfitAndClose）
```

### 9.3 关键防护

1. **防重复下单：** 原子标记 + 锁机制 + 二次检查
2. **防重复平仓：** 订单状态检查 + 数据库WHERE条件
3. **防数据延迟：** 实时价格计算盈亏 + 智能缓存
4. **防异常退出：** 持仓跟踪器持久化 + 服务重启恢复

### 9.4 使用建议

1. **合理配置参数：**
   - 窗口大小和波动阈值根据币种波动性调整
   - 止损百分比建议 5%-15%
   - 止盈回撤百分比建议 20%-40%

2. **监控关键指标：**
   - 信号触发频率
   - 下单成功率
   - 止损/止盈触发次数
   - 平均持仓时间

3. **定期检查：**
   - 预警记录是否有大量未处理
   - 执行日志是否有大量失败
   - 订单状态是否有异常

---

**报告完成 ✅**

本报告详细分析了hotgo_v2的完整交易逻辑，包括方向信号生成、自动下单、自动平仓三个核心模块，以及相关的并发安全、数据一致性、性能优化等方面。

