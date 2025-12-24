# Bitget API 接口分析与系统交易逻辑文档

## 一、Bitget API 接口分析

### 1.1 开仓接口（CreateOrder）

**接口路径**: `POST /api/v2/mix/order/place-order`

**参数说明**:
```go
params := map[string]string{
    "symbol":      "BTCUSDT",           // 交易对（已格式化）
    "productType": "USDT-FUTURES",      // 产品类型：USDT合约
    "marginMode":  "isolated",          // 保证金模式：逐仓
    "marginCoin":  "USDT",              // 保证金币种
    "side":        "buy" | "sell",      // 买卖方向：buy=买入, sell=卖出
    "tradeSide":   "open",              // 交易方向：open=开仓, close=平仓
    "orderType":   "market" | "limit",  // 订单类型：market=市价, limit=限价
    "size":        "0.001",            // 数量
    "price":       "50000",            // 价格（限价单必填）
}
```

**方向映射逻辑**:
- **做多（LONG）**: `side = "buy"`, `tradeSide = "open"`
  - 含义：买入开仓，建立多头持仓
  - 本地系统：`Side = "BUY"`, `PositionSide = "LONG"`

- **做空（SHORT）**: `side = "sell"`, `tradeSide = "open"`
  - 含义：卖出开仓，建立空头持仓
  - 本地系统：`Side = "SELL"`, `PositionSide = "SHORT"`

**代码实现** (`bitget.go:270-325`):
```go
func (b *Bitget) CreateOrder(ctx context.Context, req *OrderRequest) (*Order, error) {
    side := "buy"
    if req.Side == "SELL" {
        side = "sell"
    }
    
    tradeSide := "open"
    if req.ReduceOnly {
        tradeSide = "close"
    }
    
    // ... 构建参数并调用API
}
```

### 1.2 平仓接口（ClosePosition）

**接口路径**: `POST /api/v2/mix/order/place-order`

**参数说明**:
```go
params := map[string]string{
    "symbol":      "BTCUSDT",           // 交易对
    "productType": "USDT-FUTURES",      // 产品类型
    "marginMode":  "isolated",          // 保证金模式：逐仓
    "marginCoin":  "USDT",              // 保证金币种
    "side":        "buy" | "sell",      // 【关键】持仓方向：buy=多头方向, sell=空头方向
    "tradeSide":   "close",             // 交易方向：close=平仓
    "orderType":   "market",            // 订单类型：market=市价
    "size":        "0.001",             // 平仓数量
    "holdSide":    "long" | "short",    // 【关键】要平仓的持仓方向：long=多头, short=空头
}
```

**方向映射逻辑（Hedge-Mode 双向持仓模式）**:
- **平多（LONG）**: `side = "buy"`, `holdSide = "long"`, `tradeSide = "close"`
  - 含义：在多头方向进行平仓操作，平掉多头持仓
  - 本地系统：`PositionSide = "LONG"` → Bitget `side = "buy"`, `holdSide = "long"`

- **平空（SHORT）**: `side = "sell"`, `holdSide = "short"`, `tradeSide = "close"`
  - 含义：在空头方向进行平仓操作，平掉空头持仓
  - 本地系统：`PositionSide = "SHORT"` → Bitget `side = "sell"`, `holdSide = "short"`

**重要说明**:
- Bitget 使用 **hedge-mode（双向持仓模式）**，可以同时持有多空两个方向的持仓
- `side` 参数表示**持仓方向**（buy=多头方向，sell=空头方向）
- `holdSide` 参数表示**要平仓的具体持仓**（long=多头持仓，short=空头持仓）
- `tradeSide = "close"` 表示这是平仓操作

**代码实现** (`bitget.go:352-415`):
```go
func (b *Bitget) ClosePosition(ctx context.Context, symbol, positionSide string, quantity float64) (*Order, error) {
    side := "buy"  // 多头方向
    holdSide := "long"
    if positionSide == "SHORT" {
        side = "sell"  // 空头方向
        holdSide = "short"
    }
    
    params := map[string]string{
        "side":     side,
        "holdSide": holdSide,
        "tradeSide": "close",
        // ... 其他参数
    }
}
```

### 1.3 获取持仓接口（GetPositions）

**接口路径**: `GET /api/v2/mix/position/single-position` 或 `/api/v2/mix/position/all-position`

**返回数据结构**:
```json
{
    "code": "00000",
    "data": [
        {
            "symbol": "BTCUSDT",
            "holdSide": "long" | "short",  // 持仓方向：long=多头, short=空头
            "total": "0.001",              // 持仓数量（正数）
            "openPriceAvg": "50000",       // 开仓均价
            "markPrice": "51000",          // 标记价格
            "unrealizedPL": "1.0",         // 未实现盈亏
            "leverage": "10",              // 杠杆倍数
            "margin": "5.0",               // 保证金
            "marginMode": "isolated",      // 保证金模式
            "liquidationPrice": "45000"    // 强平价格
        }
    ]
}
```

**方向映射逻辑**:
- Bitget `holdSide = "long"` → 本地 `PositionSide = "LONG"`
- Bitget `holdSide = "short"` → 本地 `PositionSide = "SHORT"`
- 空单的持仓数量需要取反：`if holdSide == "short" { posAmt = -posAmt }`

**代码实现** (`bitget.go:208-268`):
```go
posSide := "LONG"
holdSideLower := strings.ToLower(strings.TrimSpace(holdSideRaw))
if holdSideLower == "short" {
    posSide = "SHORT"
    if posAmt > 0 {
        posAmt = -posAmt  // 空单数量取反
    }
}
```

## 二、系统全自动下单逻辑

### 2.1 触发流程

**入口**: `RobotEngine.TryAutoTradeAndUpdate()` (`robot_engine.go:2759-2812`)

**流程**:
1. **信号生成**: `doSignalGeneration()` 生成交易信号（LONG/SHORT/NEUTRAL）
2. **信号检查**: `checkSignalConditions()` 检查信号是否满足开仓条件
3. **保存信号日志**: `saveUnexecutedSignal()` 保存未执行的信号日志
4. **获取锁**: 尝试获取 `orderLock`（最多5次，共50ms）
5. **条件检查**: `checkTradingConditions()` 再次检查交易条件
6. **执行开仓**: `executeOpen()` 执行开仓操作

### 2.2 开仓条件检查

**函数**: `checkTradingConditions()` (`robot_engine.go:2900-2970`)

**检查项**:
1. ✅ 机器人状态：`status = 2`（运行中）
2. ✅ 自动下单开关：`AutoTradeEnabled = 1`
3. ✅ 信号方向：`signal.Direction != "NEUTRAL"`
4. ✅ 余额检查：`AvailableBalance > 0`
5. ✅ 持仓检查：该方向没有持仓（`direction = "long"` 或 `"short"`）
6. ✅ 最小订单金额：`orderValue >= 5 USDT`

### 2.3 执行开仓

**函数**: `executeOpen()` (`robot_engine.go:3244-3450`)

**步骤**:
1. **加载策略参数**: 从策略模板加载杠杆、保证金比例、止损止盈参数
2. **计算订单参数**:
   - 杠杆：`leverage = (LeverageMin + LeverageMax) / 2`
   - 保证金比例：`marginPercent = (MarginPercentMin + MarginPercentMax) / 2`
   - 保证金：`margin = AvailableBalance * marginPercent / 100`
   - 订单数量：`quantity = margin * leverage / ticker.LastPrice`
   - 最小数量：`quantity = max(quantity, 0.0001)`
3. **确定方向**:
   ```go
   side := "BUY"
   positionSide := "LONG"
   if signal.Direction == "SHORT" {
       side = "SELL"
       positionSide = "SHORT"
   }
   ```
4. **调用交易所API**: `Exchange.CreateOrder()`
5. **保存订单记录**: `recordOrder()` 保存到数据库
6. **更新内存状态**: 更新 `PositionTrackers` 和 `CurrentPositions`

### 2.4 数据库字段映射

**订单记录** (`robot_engine.go:3850-3895`):
```go
direction := "long"
if order.PositionSide == "SHORT" {
    direction = "short"
}

orderData := g.Map{
    "direction": direction,  // 数据库字段：long/short
    "side": order.Side,      // BUY/SELL
    // ... 其他字段
}
```

## 三、系统全自动平仓逻辑

### 3.1 触发流程

**入口**: `RobotEngine.checkClosePosition()` (`robot_engine.go:1139-1147`)

**频率**: 每500ms执行一次（`doTradingCheck()` 每500ms调用）

**流程**:
1. **获取锁**: `closeLock.TryLock()` 防止并发
2. **调用平仓检查**: `Trader.CheckAndClosePosition()`

### 3.2 平仓检查

**函数**: `CheckAndClosePosition()` (`robot_engine.go:3153-3242`)

**步骤**:
1. **检查开关**: `AutoCloseEnabled = 1`
2. **查询持仓订单**: 从数据库查询 `status = 1`（持仓中）的订单
3. **获取交易所持仓**: 
   - 优先使用缓存的 `CurrentPositions`（每10秒更新）
   - 缓存为空时才调用 `Exchange.GetPositions()`
4. **遍历订单**: 对每个持仓订单检查是否需要平仓
5. **方向映射**: `order.Direction`（long/short）→ `positionSide`（LONG/SHORT）
6. **匹配持仓**: 在交易所持仓中查找对应方向的持仓
7. **判断平仓**: `shouldCloseFromOrder()` 判断是否应该平仓
8. **执行平仓**: `executeClose()` 执行平仓操作

### 3.3 平仓条件判断

**函数**: `shouldCloseFromOrder()` (`robot_engine.go:3473-3650`)

**判断逻辑**:

#### 3.3.1 止损检查
```go
if margin > 0 && unrealizedPnl < 0 {
    lossPercent := math.Abs(unrealizedPnl) / margin * 100
    if lossPercent >= stopLossPercent {
        return true  // 触发止损
    }
}
```

**公式**: `亏损比例 = |未实现盈亏| / 保证金 × 100%`

#### 3.3.2 启动止盈
```go
if !tracker.TakeProfitEnabled && margin > 0 && autoStartRetreatPercent > 0 {
    profitPercent := unrealizedPnl / margin * 100
    if profitPercent >= autoStartRetreatPercent {
        tracker.TakeProfitEnabled = true
        // 持久化到数据库
        dao.TradingOrder.Update("profit_retreat_started", 1)
    }
}
```

**公式**: `盈利比例 = 未实现盈亏 / 保证金 × 100%`

#### 3.3.3 止盈回撤
```go
if tracker.TakeProfitEnabled && tracker.HighestProfit > 0 {
    retreat := (tracker.HighestProfit - unrealizedPnl) / tracker.HighestProfit * 100
    if retreat >= profitRetreatPercent {
        return true  // 触发止盈
    }
}
```

**公式**: `回撤比例 = (最高盈利 - 当前盈亏) / 最高盈利 × 100%`

### 3.4 执行平仓

**函数**: `executeClose()` (`robot_engine.go:3816-3865`)

**步骤**:
1. **获取持仓数量**: `quantity = math.Abs(pos.PositionAmt)`
2. **调用交易所API**: `Exchange.ClosePosition(symbol, pos.PositionSide, quantity)`
3. **更新数据库**: 更新订单状态为已平仓（`status = 2`）
4. **清除内存状态**: 删除 `PositionTrackers` 和 `CurrentPositions` 中的对应持仓

## 四、手动平仓逻辑

### 4.1 接口入口

**函数**: `CloseRobotPosition()` (`robot.go:499-745`)

**参数**:
```go
type ClosePositionInp struct {
    RobotId      int64   // 机器人ID
    Symbol       string  // 交易对（可选，默认使用机器人配置）
    PositionSide string  // 持仓方向：LONG/SHORT
    Quantity     float64 // 平仓数量（可选，默认使用全部持仓）
}
```

### 4.2 处理流程

1. **权限验证**: 验证用户是否有权限操作该机器人
2. **参数规范化**: `positionSide = strings.ToUpper(strings.TrimSpace(in.PositionSide))`
3. **获取持仓**: `Exchange.GetPositions()` 获取交易所持仓
4. **匹配持仓**: 根据 `PositionSide` 匹配要平仓的持仓
5. **数量处理**: 
   - 如果 `Quantity <= 0`，使用实际持仓数量
   - 如果 `Quantity > 实际持仓数量`，调整为实际持仓数量
6. **方向验证**: 再次确认 `foundPosition.PositionSide == positionSide`
7. **执行平仓**: `Exchange.ClosePosition(symbol, positionSide, quantity)`
8. **更新数据库**: 更新订单状态

### 4.3 错误提示优化

**修复前**:
```go
return gerror.Newf("平仓失败: 未找到 %s 方向的持仓，当前可平仓持仓: %v", positionSide, availableSides)
```

**修复后**:
```go
requestSideCN := "多单"
if positionSide == "SHORT" {
    requestSideCN = "空单"
}
// ... 构建中文描述
return gerror.Newf("平仓失败: 未找到 %s(%s) 方向的持仓。当前可平仓持仓: %v", 
    requestSideCN, positionSide, availableSidesCN)
```

## 五、方向映射总结

### 5.1 开仓方向映射

| 本地系统 | Bitget API | 说明 |
|---------|-----------|------|
| `Side = "BUY"`<br>`PositionSide = "LONG"` | `side = "buy"`<br>`tradeSide = "open"` | 买入开仓，建立多头持仓 |
| `Side = "SELL"`<br>`PositionSide = "SHORT"` | `side = "sell"`<br>`tradeSide = "open"` | 卖出开仓，建立空头持仓 |

### 5.2 平仓方向映射

| 本地系统 | Bitget API | 说明 |
|---------|-----------|------|
| `PositionSide = "LONG"` | `side = "buy"`<br>`holdSide = "long"`<br>`tradeSide = "close"` | 平多头持仓 |
| `PositionSide = "SHORT"` | `side = "sell"`<br>`holdSide = "short"`<br>`tradeSide = "close"` | 平空头持仓 |

### 5.3 数据库字段映射

| 本地系统 | 数据库字段 | 说明 |
|---------|-----------|------|
| `PositionSide = "LONG"` | `direction = "long"` | 持仓方向：多单 |
| `PositionSide = "SHORT"` | `direction = "short"` | 持仓方向：空单 |

## 六、关键注意事项

1. **Bitget 使用 hedge-mode（双向持仓模式）**，可以同时持有多空两个方向的持仓
2. **平仓时的 `side` 参数表示持仓方向**，不是交易方向
3. **`holdSide` 参数必须正确匹配要平仓的持仓方向**
4. **数据库使用 `direction` 字段（long/short）**，不是 `position_side`（LONG/SHORT）
5. **空单的持仓数量需要取反**（`PositionAmt < 0`）
6. **自动平仓每500ms检查一次**，使用缓存的持仓数据（每10秒更新）
7. **止损止盈状态会持久化到数据库**，防止重启后状态丢失

