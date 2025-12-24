# 完整订单流程文档

## 📋 目录
1. [开仓流程](#开仓流程)
2. [持仓管理流程](#持仓管理流程)
3. [平仓流程](#平仓流程)
4. [同步流程](#同步流程)
5. [事件记录](#事件记录)
6. [数据一致性保证](#数据一致性保证)

---

## 🚀 开仓流程

### 流程概览
```
信号生成 → 开仓检查 → 预创建订单 → 交易所下单 → 更新订单状态 → 更新内存缓存 → 触发同步
```

### 详细步骤

#### 1. 信号生成（Signal Generation）
**位置**: `robot_engine.go` - `doSignalGeneration()`

- **触发条件**: 价格窗口数据达到阈值
- **信号类型**: `OPEN_LONG` / `OPEN_SHORT`
- **事件记录**: `signal_generated`
  ```go
  RecordSignalGenerated(ctx, robot.Id, signal)
  ```

#### 2. 开仓检查（Open Position Check）
**位置**: `robot_engine.go` - `CheckAndOpenPositionWithSignal()`

**检查项**:
- ✅ **条件1**: 自动交易是否开启 (`AutoTradeEnabled == 1`)
- ✅ **条件2**: 双重验证机制（内存 + 数据库）
  - 快速检查内存：`hasPositionInMemory(positionSide)`
  - 准确检查数据库：`checkOpenPositionInDB(direction)`
  - 如果数据库有持仓但内存没有，同步内存：`syncPositionFromDB()`
- ✅ **条件3**: 算力是否充足 (`checkPower()`)
- ✅ **条件4**: 信号是否已处理（去重机制：`LastProcessedSignalTime`）

**事件记录**: `check_started`
```go
RecordCheckStarted(ctx, robot.Id, direction, checkResult)
```

#### 3. 预创建订单记录（Pre-Create Order）
**位置**: `robot_engine.go` - `preCreateOrder()`

**操作**:
- 计算订单参数：
  - 杠杆 (`leverage`)
  - 保证金比例 (`marginPercent`)
  - 数量 (`quantity`)
  - 开仓价格 (`entryPrice`)
  - 保证金 (`margin`)
- 从全局市场分析器获取市场状态：`market.GetMarketAnalyzer().GetAnalysis()`
- 从映射关系获取风险偏好：`MarketRiskMapping[marketState]`
- 加载策略参数：`loadStrategyParams()`
- **创建订单记录**（状态 = `PENDING`）
  ```sql
  INSERT INTO hg_trading_order (
    status = OrderStatusPending,  -- 0: 待处理
    market_state, risk_level, strategy_params, ...
  )
  ```

**事件记录**: `pre_created`
```go
RecordPreCreated(ctx, orderId, orderEventData)
```

#### 4. 交易所下单（Exchange Order）
**位置**: `robot_engine.go` - `executeOpen()`

**操作**:
- 设置杠杆：`Exchange.SetLeverage()`
- 调用交易所API：`Exchange.CreateOrder()`
  ```go
  order, err := t.engine.Exchange.CreateOrder(ctx, &exchange.OrderRequest{
      Symbol:       robot.Symbol,
      Side:         side,          // BUY/SELL
      PositionSide: positionSide,  // LONG/SHORT
      Type:         "MARKET",
      Quantity:     quantity,
  })
  ```

**成功处理**:
- 事件记录：`exchange_ordered` (success)
- 继续下一步

**失败处理**:
- 事件记录：`exchange_ordered` (failed) + `order_failed`
- 更新订单状态为 `FAILED`：`updateOrderStatus(OrderStatusFailed)`
- 返回错误

**事件记录**: `exchange_ordered`
```go
RecordExchangeOrdered(ctx, localOrderId, exchangeOrderId, requestData, responseData, success)
```

#### 5. 更新订单状态（Update Order Status）
**位置**: `robot_engine.go` - `updateOrderStatus()`

**操作**:
- 更新订单状态：`PENDING` → `OPEN`
- 更新交易所订单ID：`exchange_order_id`
- 更新成交价格：`avg_price`, `open_price`
- 更新已成交数量：`filled_qty`
- 更新创建时间：`open_time`, `created_at`

**事件记录**: `order_filled`
```go
RecordOrderFilled(ctx, localOrderId, exchangeOrderId, fillData)
```

#### 6. 更新内存缓存（Update Memory Cache）
**位置**: `robot_engine.go` - `executeOpen()`

**操作**:
- 更新 `PositionTrackers[positionSide]`：
  ```go
  PositionTrackers[positionSide] = &PositionTracker{
      PositionSide: positionSide,
      EntryMargin:  margin,
      EntryTime:    time.Now(),
  }
  ```
- 更新 `CurrentPositions`：
  ```go
  CurrentPositions = append(CurrentPositions, &exchange.Position{
      Symbol:         robot.Symbol,
      PositionSide:   positionSide,
      PositionAmt:    positionAmt,
      EntryPrice:     entryPrice,
      IsolatedMargin: margin,
      UnrealizedPnl:  0,
  })
  ```

#### 7. 触发同步服务（Trigger Sync）
**位置**: `robot_engine.go` - `executeOpen()` (goroutine)

**操作**:
- 立即同步账户数据：`syncAccountDataIfNeeded(ctx, "after_trade")`
- 立即同步订单状态：`SyncSingleRobot(ctx, robot.Id)`
- **无延迟**：收到API返回后立即同步

---

## 📊 持仓管理流程

### 流程概览
```
价格变动 → 计算未实现盈亏 → 更新数据库 → 更新内存 → 检查平仓条件
```

### 详细步骤

#### 1. 价格变动触发（Price Change）
**位置**: `robot_engine.go` - `doAnalysis()`

- **事件驱动**：实时价格更新触发
- 获取最新价格：`ticker.LastPrice`
- 触发持仓更新检查

#### 2. 计算未实现盈亏（Calculate Unrealized PnL）
**位置**: `robot_engine.go` - `updateOrdersUnrealizedPnl()`

**计算逻辑**:
```go
// 多头：未实现盈亏 = (当前价格 - 开仓价格) × 数量 × 杠杆
// 空头：未实现盈亏 = (开仓价格 - 当前价格) × 数量 × 杠杆
unrealizedPnl = (currentPrice - entryPrice) * quantity * leverage
```

#### 3. 更新数据库（Update Database）
**位置**: `order_status_sync.go` - `updateOrderUnrealizedPnl()`

**操作**:
- 更新 `unrealized_profit`
- 更新 `highest_profit`（只增不减）
- 更新 `mark_price`

**事件记录**: `position_updated`
```go
RecordPositionUpdated(ctx, order.Id, order.ExchangeOrderId, updateData)
```

#### 4. 更新内存（Update Memory）
**位置**: `robot_engine.go` - `updateOrdersUnrealizedPnl()`

**操作**:
- 更新 `CurrentPositions[i].UnrealizedPnl`
- 更新 `PositionTrackers[positionSide].HighestProfit`

#### 5. 检查平仓条件（Check Close Conditions）
**位置**: `robot_engine.go` - `shouldCloseFromOrder()`

**检查项**:
- ✅ **止损检查**：
  ```go
  lossPercent = |unrealizedPnl| / margin * 100
  if lossPercent >= stopLossPercent {
      return true  // 触发止损
  }
  ```
- ✅ **止盈检查**：
  - 阶段1：启动止盈回撤
    ```go
    profitPercent = unrealizedPnl / margin * 100
    if profitPercent >= takeProfitStart {
        tracker.TakeProfitEnabled = true
    }
    ```
  - 阶段2：止盈回撤触发
    ```go
    retreatPercent = (highestProfit - unrealizedPnl) / highestProfit * 100
    if retreatPercent >= profitRetreatPercent {
        return true  // 触发止盈
    }
    ```

---

## 🔄 平仓流程

### 自动平仓流程

#### 1. 触发平仓（Trigger Close）
**位置**: `robot_engine.go` - `doAnalysis()`

- **事件驱动**：价格变动触发
- 检查平仓条件：`shouldCloseFromOrder()`
- 如果满足条件，调用：`executeClose()`

#### 2. 执行平仓（Execute Close）
**位置**: `robot_engine.go` - `executeClose()`

**操作**:
- 调用交易所API：`Exchange.ClosePosition()`
  ```go
  order, err := t.engine.Exchange.ClosePosition(ctx, symbol, positionSide, quantity)
  ```

**成功处理**:
- 查找本地订单：通过 `exchange_order_id`
- 更新机器人盈亏：`Increment("total_profit", unrealizedPnl)`
- 清除内存持仓：`ClearPosition(ctx, positionSide)`
- 触发同步：`SyncSingleRobot(ctx, robot.Id)`

**失败处理**:
- 事件记录：`order_failed`
- 记录错误日志

#### 3. 同步订单状态（Sync Order Status）
**位置**: `order_status_sync.go` - `CloseOrder()`

**操作**:
- 查询当前订单状态
- 计算已实现盈亏：`realizedProfit`
- 计算平仓价格：`closePrice`
- 更新订单状态：`OPEN` → `CLOSED`
- 补全平仓信息：
  - `close_price`
  - `close_time`
  - `realized_profit`
  - `hold_duration`
  - `close_reason`
  - `close_order_id`
  - `close_fee`
- 扣除算力：`ConsumePower()`（如果盈利）

**事件记录**: `order_closed`
```go
RecordOrderClosed(ctx, order.Id, order.ExchangeOrderId, closeEventData)
```

### 手动平仓流程

#### 1. 接收平仓请求（Receive Close Request）
**位置**: `robot.go` - `CloseRobotPosition()`

- 用户手动触发平仓
- 参数：`robotId`, `positionSide`, `quantity`

#### 2. 验证持仓（Validate Position）
**位置**: `robot.go` - `CloseRobotPosition()`

**检查项**:
- ✅ 查询交易所持仓：`Exchange.GetPositions()`
- ✅ 验证持仓方向：`positionSide` 匹配
- ✅ 验证持仓数量：`actualQuantity > 0`

#### 3. 执行平仓（Execute Close）
**位置**: `robot.go` - `CloseRobotPosition()`

**操作**:
- 调用交易所API：`Exchange.ClosePosition()`
- 查询本地订单：`dao.TradingOrder.Where("status", OrderStatusOpen)`
- 计算已实现盈亏和平仓价格
- 调用 `CloseOrder()` 补全信息

#### 4. 清除内存（Clear Memory）
**位置**: `robot.go` - `CloseRobotPosition()`

**操作**:
- 清除内存持仓：`robotEngine.ClearPosition(ctx, positionSide)`
- 触发同步：`SyncSingleRobot(ctx, robot.Id)`

---

## 🔁 同步流程

### 事件驱动同步

#### 1. 开仓后同步（After Open）
**位置**: `robot_engine.go` - `executeOpen()` (goroutine)

**触发时机**: 交易所下单成功后立即触发

**操作**:
- 同步账户数据：`syncAccountDataIfNeeded(ctx, "after_trade")`
- 同步订单状态：`SyncSingleRobot(ctx, robot.Id)`

#### 2. 平仓后同步（After Close）
**位置**: `robot_engine.go` - `executeClose()` (goroutine)

**触发时机**: 交易所平仓成功后立即触发

**操作**:
- 同步账户数据：`syncAccountDataIfNeeded(ctx, "after_trade")`
- 同步订单状态：`SyncSingleRobot(ctx, robot.Id)`

#### 3. 手动平仓后同步（After Manual Close）
**位置**: `robot.go` - `CloseRobotPosition()` (goroutine)

**触发时机**: 手动平仓成功后立即触发

**操作**:
- 同步账户数据：`syncAccountDataIfNeeded(ctx, "after_trade")`
- 同步订单状态：`SyncSingleRobot(ctx, robot.Id)`

### 定期同步

#### 1. 外部持仓检测（External Position Detection）
**位置**: `order_status_sync.go` - `SyncExternalPositions()`

**触发时机**: 每3秒执行一次

**操作**:
- 查询所有机器人的交易所持仓
- 检查本地是否有对应订单
- 如果没有，创建订单记录：`createOrderFromExchange()`

#### 2. 订单状态同步（Order Status Sync）
**位置**: `order_status_sync.go` - `syncLocalOrders()`

**操作**:
- 查询订单历史：`Exchange.GetOrderHistory()`
- 匹配本地订单：通过 `exchange_order_id`
- 补全缺失字段（去重机制）：
  - `avg_price`（如果缺失）
  - `filled_qty`（如果缺失）
  - `exchange_order_id`（如果缺失）
  - `open_time`（如果缺失）
  - `market_state`（如果缺失）
  - `risk_level`（如果缺失）
  - 策略参数（如果缺失）

#### 3. 手动平仓检测（Manual Close Detection）
**位置**: `order_status_sync.go` - `syncLocalOrders()`

**操作**:
- 检查本地订单状态：`status = OrderStatusOpen`
- 检查交易所持仓：是否存在对应持仓
- 如果交易所无持仓但本地有订单 → 手动平仓
- 调用 `CloseOrder()` 补全信息

---

## 📝 事件记录

### 事件类型

| 事件类型 | 触发时机 | 状态 |
|---------|---------|------|
| `signal_generated` | 信号生成 | success |
| `check_started` | 开仓检查开始 | success |
| `pre_created` | 预创建订单记录 | success |
| `exchange_ordered` | 交易所下单 | success/failed |
| `order_filled` | 订单成交 | success |
| `position_updated` | 持仓更新 | success |
| `order_closed` | 订单平仓 | success |
| `order_failed` | 订单失败 | failed |

### 事件记录位置

- **信号生成**: `robot_engine.go` - `CheckAndOpenPositionWithSignal()`
- **开仓检查**: `robot_engine.go` - `CheckAndOpenPositionWithSignal()`
- **预创建订单**: `robot_engine.go` - `preCreateOrder()`
- **交易所下单**: `robot_engine.go` - `executeOpen()`
- **订单成交**: `robot_engine.go` - `executeOpen()`
- **持仓更新**: `order_status_sync.go` - `updateOrderUnrealizedPnl()`
- **订单平仓**: `order_status_sync.go` - `CloseOrder()`
- **订单失败**: `robot_engine.go` - `executeOpen()` / `executeClose()`

---

## 🔒 数据一致性保证

### 三层数据同步

1. **内存层** (`RobotEngine`)
   - `CurrentPositions`: 当前持仓列表
   - `PositionTrackers`: 持仓跟踪器

2. **数据库层** (`hg_trading_order`)
   - 订单主表：存储订单完整信息
   - 订单事件表：存储订单生命周期事件

3. **交易所层** (`Exchange API`)
   - 实时持仓数据
   - 订单历史记录

### 一致性策略

1. **双重验证机制**
   - 开仓前：内存 + 数据库双重检查
   - 以数据库为准，内存为辅

2. **去重机制**
   - 同步服务只更新缺失字段
   - 避免重复更新已存在的有效数据

3. **立即同步**
   - 开仓/平仓后立即同步
   - 无延迟，确保数据实时性

4. **定期同步**
   - 每3秒检测外部持仓
   - 确保外部持仓能够及时补全

5. **内存同步**
   - 数据库有持仓但内存没有 → 同步内存
   - 平仓后立即清除内存

---

## 📊 完整流程图

```
┌─────────────────────────────────────────────────────────────────┐
│                        开仓流程                                  │
└─────────────────────────────────────────────────────────────────┘
    信号生成 (signal_generated)
         ↓
    开仓检查 (check_started)
    ├─ 自动交易开启？
    ├─ 内存/数据库双重验证
    ├─ 算力充足？
    └─ 信号已处理？
         ↓
    预创建订单 (pre_created)
    ├─ 状态: PENDING
    ├─ 保存到数据库
    └─ 记录事件
         ↓
    交易所下单 (exchange_ordered)
    ├─ 成功 → 继续
    └─ 失败 → 更新状态为FAILED，记录事件，返回
         ↓
    更新订单状态 (order_filled)
    ├─ 状态: PENDING → OPEN
    ├─ 更新交易所订单ID
    ├─ 更新成交价格
    └─ 记录事件
         ↓
    更新内存缓存
    ├─ PositionTrackers
    └─ CurrentPositions
         ↓
    触发同步 (立即，无延迟)
    ├─ syncAccountDataIfNeeded()
    └─ SyncSingleRobot()

┌─────────────────────────────────────────────────────────────────┐
│                        持仓管理流程                              │
└─────────────────────────────────────────────────────────────────┘
    价格变动 (事件驱动)
         ↓
    计算未实现盈亏
         ↓
    更新数据库 (position_updated)
    ├─ unrealized_profit
    ├─ highest_profit (只增不减)
    └─ mark_price
         ↓
    更新内存
    ├─ CurrentPositions[].UnrealizedPnl
    └─ PositionTrackers[].HighestProfit
         ↓
    检查平仓条件
    ├─ 止损检查
    └─ 止盈检查

┌─────────────────────────────────────────────────────────────────┐
│                        平仓流程                                  │
└─────────────────────────────────────────────────────────────────┘
    触发平仓 (事件驱动)
         ↓
    执行平仓
    ├─ 调用交易所API
    ├─ 成功 → 继续
    └─ 失败 → 记录事件，返回
         ↓
    同步订单状态 (order_closed)
    ├─ 状态: OPEN → CLOSED
    ├─ 计算已实现盈亏
    ├─ 补全平仓信息
    ├─ 扣除算力（如果盈利）
    └─ 记录事件
         ↓
    清除内存
    ├─ ClearPosition()
    └─ 更新 CurrentPositions
         ↓
    触发同步 (立即，无延迟)
    ├─ syncAccountDataIfNeeded()
    └─ SyncSingleRobot()

┌─────────────────────────────────────────────────────────────────┐
│                        同步流程                                  │
└─────────────────────────────────────────────────────────────────┘
    事件驱动同步 (立即)
    ├─ 开仓后同步
    ├─ 平仓后同步
    └─ 手动平仓后同步
    
    定期同步 (每3秒)
    ├─ 外部持仓检测
    │   └─ 创建订单记录 (如果缺失)
    ├─ 订单状态同步
    │   └─ 补全缺失字段 (去重机制)
    └─ 手动平仓检测
        └─ 补全平仓信息
```

---

## 🎯 关键特性

1. **事件驱动架构**
   - 开仓：信号生成触发
   - 平仓：价格变动触发
   - 同步：开仓/平仓后立即触发

2. **无延迟同步**
   - 收到API返回后立即同步
   - 不等待延迟，确保数据实时性

3. **去重机制**
   - 同步服务只更新缺失字段
   - 避免重复更新已存在的有效数据

4. **双重验证**
   - 内存 + 数据库双重检查
   - 以数据库为准，内存为辅

5. **完整的事件追踪**
   - 每个订单的每个节点都有事件记录
   - 便于问题排查和审计

6. **完善的错误处理**
   - 所有错误都记录日志和事件
   - 不阻塞主流程，确保订单能够正常创建

---

## 📚 相关文件

- `robot_engine.go` - 机器人核心引擎（开仓、平仓逻辑）
- `robot.go` - 机器人服务层（手动平仓）
- `order_status_sync.go` - 订单状态同步服务
- `order_event.go` - 订单事件记录
- `order_event_monitor.go` - 订单事件监控和分析
- `order_status.go` - 订单状态常量定义

