# 下单信号到API请求数据流分析

## 概述

本文档详细分析当系统检测到下单信号后，如何提取数据并向交易所平台提交API请求的完整流程。

---

## 一、信号触发阶段

### 1.1 信号生成 (`robot_engine.go:1295-1310`)

**信号数据结构 (`RobotSignal`)**：
```go
type RobotSignal struct {
    Direction       string  // 方向: "LONG" / "SHORT"
    Action          string  // 操作: "OPEN_LONG" / "OPEN_SHORT"
    CurrentPrice    float64 // 当前价格
    WindowMinPrice  float64 // 窗口最低价
    WindowMaxPrice  float64 // 窗口最高价
    SignalThreshold float64 // 信号阈值
    Strength        int     // 信号强度
    Reason          string  // 信号原因
}
```

**触发条件**：
- 做多信号：`当前价 - 窗口最低价 ≥ 阈值`
- 做空信号：`窗口最高价 - 当前价 ≥ 阈值`

---

## 二、下单前数据提取阶段 (`executeOpen`)

### 2.1 基础数据提取 (`robot_engine.go:2830-2841`)

**从引擎内存中提取**：
```go
robot := t.engine.Robot          // 机器人配置
balance := t.engine.AccountBalance  // 账户余额
ticker := t.engine.LastTicker    // 最新行情
```

**提取的数据**：
- `robot.Symbol` - 交易对（如 "BTCUSDT"）
- `balance.AvailableBalance` - 可用余额（USDT）
- `ticker.LastPrice` - 最新价格

---

### 2.2 策略参数提取 (`robot_engine.go:2843-2886`)

**策略参数来源**：
1. **优先**：从缓存中获取 `t.engine.CurrentStrategyParams`
2. **备用**：根据市场状态和风险偏好重新加载

**策略参数结构 (`StrategyParams`)**：
```go
type StrategyParams struct {
    LeverageMin           int     // 最小杠杆
    LeverageMax           int     // 最大杠杆
    MarginPercentMin      float64 // 最小保证金比例
    MarginPercentMax      float64 // 最大保证金比例
    StopLossPercent       float64 // 止损百分比
    AutoStartRetreatPercent float64 // 启动止盈百分比
    ProfitRetreatPercent  float64 // 止盈回撤百分比
}
```

**计算逻辑**：
```go
// 杠杆：使用范围的中值
leverage := (strategyParams.LeverageMin + strategyParams.LeverageMax) / 2
if leverage <= 0 {
    leverage = 10  // 默认10倍
}

// 保证金比例：使用范围的中值
marginPercent := (strategyParams.MarginPercentMin + strategyParams.MarginPercentMax) / 2
if marginPercent <= 0 {
    marginPercent = 10  // 默认10%
}
```

---

### 2.3 订单金额和数量计算 (`robot_engine.go:2921-2944`)

**计算步骤**：

1. **计算保证金**：
   ```go
   margin := balance.AvailableBalance * marginPercent / 100
   ```

2. **计算订单价值**：
   ```go
   orderValue := margin * float64(leverage)
   ```

3. **检查最小订单金额**：
   ```go
   minOrderValue := 5.0  // 最小订单价值 5 USDT
   if orderValue < minOrderValue {
       return error("订单金额不足")
   }
   ```

4. **计算下单数量**：
   ```go
   quantity := margin * float64(leverage) / ticker.LastPrice
   ```

5. **检查最小数量**：
   ```go
   minQuantity := 0.0001
   if quantity < minQuantity {
       return error("订单数量不足")
   }
   ```

---

### 2.4 方向确定 (`robot_engine.go:2949-2955`)

**根据信号方向确定**：
```go
side := "BUY"
positionSide := "LONG"
if signal.Direction == "SHORT" {
    side = "SELL"
    positionSide = "SHORT"
}
```

**字段说明**：
- `side` - 买卖方向：`BUY`（买入）/ `SELL`（卖出）
- `positionSide` - 持仓方向：`LONG`（做多）/ `SHORT`（做空）

---

### 2.5 设置杠杆 (`robot_engine.go:2947`)

**在提交订单前设置杠杆**：
```go
_ = t.engine.Exchange.SetLeverage(ctx, robot.Symbol, leverage)
```

**API调用**：
- **Binance**: `POST /fapi/v1/leverage`
  ```json
  {
    "symbol": "BTCUSDT",
    "leverage": 10
  }
  ```

---

## 三、构建API请求 (`robot_engine.go:2958-2964`)

### 3.1 构建订单请求 (`OrderRequest`)

**请求结构**：
```go
orderRequest := &exchange.OrderRequest{
    Symbol:       robot.Symbol,      // 交易对，如 "BTCUSDT"
    Side:         side,              // "BUY" 或 "SELL"
    PositionSide: positionSide,      // "LONG" 或 "SHORT"
    Type:         "MARKET",          // 订单类型：市价单
    Quantity:     quantity,          // 下单数量（计算得出）
}
```

**字段说明**：
- `Symbol` - 交易对（从机器人配置获取）
- `Side` - 买卖方向（`BUY`/`SELL`）
- `PositionSide` - 持仓方向（`LONG`/`SHORT`）
- `Type` - 订单类型（固定为 `MARKET` 市价单）
- `Quantity` - 下单数量（根据保证金和杠杆计算）

---

## 四、提交API请求

### 4.1 Binance 交易所 (`binance.go:156-190`)

**API端点**：`POST /fapi/v1/order`

**请求参数**：
```go
params := map[string]string{
    "symbol":       "BTCUSDT",           // 格式化后的交易对
    "side":         "BUY",               // BUY 或 SELL
    "positionSide": "LONG",              // LONG 或 SHORT
    "type":         "MARKET",            // MARKET 或 LIMIT
    "quantity":     "0.001",             // 数量（字符串格式）
}
```

**完整请求示例**：
```http
POST https://fapi.binance.com/fapi/v1/order
Headers:
  X-MBX-APIKEY: <api_key>
  Content-Type: application/x-www-form-urlencoded
Body:
  symbol=BTCUSDT&side=BUY&positionSide=LONG&type=MARKET&quantity=0.001&timestamp=1234567890&signature=<signature>
```

---

### 4.2 Bitget 交易所 (`bitget.go:250-304`)

**API端点**：`POST /api/v2/mix/order/place-order`

**请求参数转换**：
```go
// Side 转换
side := "buy"  // BUY -> buy, SELL -> sell

// TradeSide 转换
tradeSide := "open"  // 开仓

// OrderType 转换
orderType := "market"  // MARKET -> market, LIMIT -> limit
```

**请求参数**：
```go
params := map[string]string{
    "symbol":      "BTCUSDT",           // 格式化后的交易对
    "productType": "USDT-FUTURES",      // 产品类型（固定）
    "marginMode":  "isolated",          // 保证金模式（固定：逐仓）
    "marginCoin":  "USDT",              // 保证金币种（固定）
    "side":        "buy",                // buy 或 sell
    "tradeSide":   "open",               // open（开仓）或 close（平仓）
    "orderType":   "market",            // market 或 limit
    "size":        "0.001",             // 数量（字符串格式）
}
```

**完整请求示例**：
```http
POST https://api.bitget.com/api/v2/mix/order/place-order
Headers:
  ACCESS-KEY: <api_key>
  ACCESS-SIGN: <signature>
  ACCESS-TIMESTAMP: <timestamp>
  ACCESS-PASSPHRASE: <passphrase>
  Content-Type: application/json
Body:
{
  "symbol": "BTCUSDT",
  "productType": "USDT-FUTURES",
  "marginMode": "isolated",
  "marginCoin": "USDT",
  "side": "buy",
  "tradeSide": "open",
  "orderType": "market",
  "size": "0.001"
}
```

---

## 五、数据流总结

### 5.1 完整数据流图

```
信号生成
  ↓
RobotSignal {
  Direction: "LONG" / "SHORT"
  CurrentPrice: 50000.0
  ...
}
  ↓
executeOpen() 函数
  ↓
【数据提取】
  ├─ robot.Symbol              → "BTCUSDT"
  ├─ balance.AvailableBalance  → 1000.0 USDT
  ├─ ticker.LastPrice          → 50000.0
  └─ strategyParams            → {杠杆、保证金比例等}
  ↓
【计算】
  ├─ leverage = (LeverageMin + LeverageMax) / 2  → 10
  ├─ marginPercent = (MarginPercentMin + MarginPercentMax) / 2  → 10%
  ├─ margin = balance * marginPercent / 100  → 100 USDT
  ├─ orderValue = margin * leverage  → 1000 USDT
  ├─ quantity = orderValue / price  → 0.02 BTC
  ├─ side = signal.Direction == "LONG" ? "BUY" : "SELL"
  └─ positionSide = signal.Direction == "LONG" ? "LONG" : "SHORT"
  ↓
【设置杠杆】
  Exchange.SetLeverage(symbol, leverage)
  ↓
【构建请求】
  OrderRequest {
    Symbol: "BTCUSDT",
    Side: "BUY",
    PositionSide: "LONG",
    Type: "MARKET",
    Quantity: 0.02
  }
  ↓
【提交API】
  Exchange.CreateOrder(ctx, OrderRequest)
  ↓
【API请求】
  POST /fapi/v1/order (Binance)
  或
  POST /api/v2/mix/order/place-order (Bitget)
```

---

### 5.2 关键数据来源汇总

| 数据项 | 来源 | 说明 |
|--------|------|------|
| **交易对 (Symbol)** | `robot.Symbol` | 机器人配置的交易对 |
| **可用余额** | `balance.AvailableBalance` | 账户可用余额（USDT） |
| **当前价格** | `ticker.LastPrice` | 最新行情价格 |
| **杠杆** | `strategyParams.LeverageMin/Max` | 策略参数（范围中值） |
| **保证金比例** | `strategyParams.MarginPercentMin/Max` | 策略参数（范围中值） |
| **买卖方向 (Side)** | `signal.Direction` | 信号方向（LONG→BUY, SHORT→SELL） |
| **持仓方向 (PositionSide)** | `signal.Direction` | 信号方向（LONG/SHORT） |
| **订单类型 (Type)** | 固定值 | `MARKET`（市价单） |
| **下单数量 (Quantity)** | 计算得出 | `margin * leverage / price` |

---

### 5.3 计算公式汇总

```go
// 1. 杠杆（策略参数范围中值）
leverage = (LeverageMin + LeverageMax) / 2

// 2. 保证金比例（策略参数范围中值）
marginPercent = (MarginPercentMin + MarginPercentMax) / 2

// 3. 保证金金额
margin = AvailableBalance * marginPercent / 100

// 4. 订单价值
orderValue = margin * leverage

// 5. 下单数量
quantity = orderValue / LastPrice
        = (margin * leverage) / LastPrice
        = (AvailableBalance * marginPercent / 100 * leverage) / LastPrice
```

---

## 六、验证和限制

### 6.1 订单金额验证 (`robot_engine.go:2923-2931`)

```go
minOrderValue := 5.0  // 最小订单价值 5 USDT
orderValue := margin * float64(leverage)
if orderValue < minOrderValue {
    return error("订单金额不足")
}
```

### 6.2 订单数量验证 (`robot_engine.go:2938-2944`)

```go
minQuantity := 0.0001  // 最小下单数量
if quantity < minQuantity {
    return error("订单数量不足")
}
```

---

## 七、实际示例

### 7.1 示例场景

**输入数据**：
- 交易对：`BTCUSDT`
- 可用余额：`1000 USDT`
- 当前价格：`50000 USDT`
- 信号方向：`LONG`（做多）
- 策略参数：
  - 杠杆范围：`5-15` → 中值 `10`
  - 保证金比例：`5%-15%` → 中值 `10%`

**计算过程**：
1. 杠杆：`(5 + 15) / 2 = 10`
2. 保证金比例：`(5% + 15%) / 2 = 10%`
3. 保证金：`1000 * 10% = 100 USDT`
4. 订单价值：`100 * 10 = 1000 USDT`
5. 下单数量：`1000 / 50000 = 0.02 BTC`

**API请求**：
```json
{
  "symbol": "BTCUSDT",
  "side": "BUY",
  "positionSide": "LONG",
  "type": "MARKET",
  "quantity": "0.02"
}
```

---

## 八、总结

### 8.1 关键要点

1. **数据来源**：
   - 机器人配置（交易对）
   - 账户余额（可用余额）
   - 行情数据（当前价格）
   - 策略参数（杠杆、保证金比例）

2. **计算逻辑**：
   - 使用策略参数范围的中值
   - 保证金 = 余额 × 保证金比例
   - 订单价值 = 保证金 × 杠杆
   - 下单数量 = 订单价值 / 价格

3. **API请求**：
   - 固定使用市价单（`MARKET`）
   - 下单前先设置杠杆
   - 不同交易所参数格式不同

4. **验证机制**：
   - 最小订单价值：5 USDT
   - 最小下单数量：0.0001

### 8.2 代码位置

- **信号生成**：`robot_engine.go:1295-1310`
- **下单执行**：`robot_engine.go:2829-3016`
- **API请求（Binance）**：`binance.go:156-190`
- **API请求（Bitget）**：`bitget.go:250-304`
- **订单请求结构**：`exchange.go:154-165`

---

## 九、注意事项

1. **杠杆设置**：下单前必须先设置杠杆，否则可能使用默认杠杆
2. **余额检查**：确保可用余额足够，否则下单会失败
3. **价格精度**：不同交易所有不同的价格和数量精度要求
4. **最小订单**：订单价值必须 ≥ 5 USDT，数量必须 ≥ 0.0001
5. **市价单**：当前固定使用市价单，不设置价格参数

