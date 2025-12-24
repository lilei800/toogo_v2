# 平台API调用汇总

## 一、全局引擎（Global Engine）

### 1. MarketServiceManager（市场服务管理器）
**位置**: `internal/library/market/market_service_manager.go`

**数据获取方式**:
- **Ticker（实时行情）**: 
  - **优先WebSocket**（如果配置启用 `toogo.websocketEnabled=true`）
    - Bitget: `bitgetWS.GetTicker()`
    - Binance: `binanceWS.GetTicker()`
  - **降级HTTP**: WebSocket不可用时，使用HTTP缓存
  - 调用位置: `GetTicker()` (line 283)
  
- **K线数据**: 
  - **全部使用HTTP轮询**（`ex.GetKlines()`）
  - 调用位置: `GetKlines()` (line 320)

### 2. MarketDataService（全局行情数据服务）
**位置**: `internal/library/market/market_data_service.go`

**数据获取方式**:
- **Ticker**: **HTTP轮询**（每秒调用 `ex.GetTicker()`）
  - 调用位置: `runTickerUpdater()` → `updateAllTickers()` (line 370)
  - 用途: 为所有机器人提供统一的行情数据缓存
  
- **K线数据**: **HTTP轮询**（每5秒更新）
  - 周期: 1m(100根), 5m(200根), 15m(200根), 30m(100根), 1h(200根), 1d(30根)
  - 调用位置: `runKlineUpdater()` → `updateAllKlines()` (line 385)
  - 调用时机: 订阅时主动获取历史K线，后续定时更新
  - 用途: 为市场分析引擎提供多周期K线数据

### 3. ExchangeMarketService（交易所市场服务）
**位置**: `internal/library/market/market_service_manager.go`

**数据获取方式**:
- **Ticker**: **HTTP轮询**（每秒调用 `ex.GetTicker()`）
  - 调用位置: `runTickerUpdater()` → `updateAllTickers()` (line 640)
  
- **K线数据**: **HTTP轮询**（每5秒更新）
  - 调用位置: `runKlineUpdater()` → `updateAllKlines()` → `fetchAllKlines()` (line 580)

### 2. MarketServiceManager（市场服务管理器）
**位置**: `internal/library/market/market_service_manager.go`

**使用的API**:
- `GetTicker(platform, symbol)` - 获取实时行情（优先WebSocket，降级HTTP）
- `GetKlines(platform, symbol, interval)` - 获取K线数据

### 3. ExchangeMarketService（交易所市场服务）
**位置**: `internal/library/market/market_service_manager.go`

**使用的API**:
- `GetTicker(ctx, symbol)` - 获取Ticker（从缓存）
- `GetKlines(ctx, symbol, interval)` - 获取K线数据（从缓存）

---

## 二、机器人引擎（Robot Engine）

### 1. RobotEngine（机器人引擎核心）
**位置**: `internal/logic/toogo/robot_engine.go`

**使用的API**:

#### 账户相关
- `GetBalance(ctx)` - 获取账户余额
  - 调用位置: `GetBalanceSmart()` (line 1259)
  - 缓存策略: 使用缓存 + singleflight 模式，避免并发重复请求
  - 调用时机: 下单前获取可用余额

- `GetPositions(ctx, symbol)` - 获取持仓
  - 调用位置: 
    - `GetPositionsSmart()` (line 1217) - 智能获取（带缓存）
    - `ForceRefreshPositions()` (line 1292) - 强制刷新（平仓等关键操作）
  - 缓存策略: 使用缓存 + singleflight 模式
  - 调用时机: 定期同步账户数据、平仓前强制刷新

#### 订单相关
- `GetOrderHistory(ctx, symbol, limit)` - 获取历史订单
  - 调用位置: `syncAccountData()` (line 1337)
  - 参数: limit=50
  - 用途: 同步订单状态到数据库

- `CreateOrder(ctx, req)` - 创建订单
  - 调用位置: `RobotTrader.executeTrade()` (line 5174)
  - 用途: 执行开仓下单

- `ClosePosition(ctx, symbol, positionSide, quantity)` - 平仓
  - 调用位置:
    - `executeStopLossCloseByPosition()` (line 1938) - 止损平仓
    - `executeTakeProfitCloseByPosition()` (line 2282) - 止盈平仓
    - `executeManualClose()` (line 2021, 2333) - 手动平仓
  - 用途: 执行平仓操作

#### 配置相关
- `SetLeverage(ctx, symbol, leverage)` - 设置杠杆
  - 调用位置: `RobotTrader.executeTrade()` (line 5139)
  - 用途: 下单前设置杠杆倍数

#### 行情相关
- `GetTicker(ctx, symbol)` - 获取行情
  - 调用位置: `detectManualClose()` (line 1540)
  - 用途: 检测手动平仓时获取最新价格

### 2. OrderStatusSync（订单状态同步）
**位置**: `internal/logic/toogo/order_status_sync.go`

**使用的API**:
- `GetOrderHistory(ctx, symbol, limit)` - 获取历史订单
  - 调用位置: `syncOrderStatus()` (line 387)
  - 参数: limit=50
  - 用途: 同步订单状态

---

## 三、Exchange接口定义

**位置**: `internal/library/exchange/exchange.go`

### 基础接口（Exchange）
```go
type Exchange interface {
    GetName() string
    GetBalance(ctx context.Context) (*Balance, error)
    GetTicker(ctx context.Context, symbol string) (*Ticker, error)
    GetKlines(ctx context.Context, symbol, interval string, limit int) ([]*Kline, error)
    GetPositions(ctx context.Context, symbol string) ([]*Position, error)
    CreateOrder(ctx context.Context, req *OrderRequest) (*Order, error)
    CancelOrder(ctx context.Context, symbol, orderId string) (*Order, error)
    ClosePosition(ctx context.Context, symbol, positionSide string, quantity float64) (*Order, error)
    SetLeverage(ctx context.Context, symbol string, leverage int) error
    SetMarginType(ctx context.Context, symbol, marginType string) error
    GetOpenOrders(ctx context.Context, symbol string) ([]*Order, error)
    GetOrderHistory(ctx context.Context, symbol string, limit int) ([]*Order, error)
}
```

### 高级接口（ExchangeAdvanced，可选实现）
```go
type ExchangeAdvanced interface {
    SetStopLoss(ctx context.Context, req *StopLossRequest) (*Order, error)
    SetTakeProfit(ctx context.Context, req *TakeProfitRequest) (*Order, error)
    SetStopLossAndTakeProfit(ctx context.Context, req *SLTPRequest) (*SLTPResponse, error)
    CancelStopLoss(ctx context.Context, symbol, orderId string) error
    CancelTakeProfit(ctx context.Context, symbol, orderId string) error
    BatchClosePositions(ctx context.Context, symbols []string) ([]*CloseResult, error)
    CloseAllPositions(ctx context.Context) ([]*CloseResult, error)
    GetAccountInfo(ctx context.Context) (*AccountInfo, error)
    GetSymbolInfo(ctx context.Context, symbol string) (*SymbolInfo, error)
    GetFundingRate(ctx context.Context, symbol string) (*FundingRate, error)
    ModifyOrder(ctx context.Context, symbol, orderId string, price, quantity float64) (*Order, error)
    GetTradeHistory(ctx context.Context, symbol string, limit int) ([]*Trade, error)
}
```

---

## 四、API调用统计

### 全局引擎调用
| API | 数据源 | 调用频率 | 用途 |
|-----|--------|---------|------|
| GetTicker | **WebSocket优先**（Bitget/Binance），降级HTTP | 1秒/次 | MarketServiceManager全局行情 |
| GetTicker | HTTP轮询 | 1秒/次 | MarketDataService行情缓存 |
| GetTicker | HTTP轮询 | 1秒/次 | ExchangeMarketService行情缓存 |
| GetKlines | **HTTP轮询**（全部） | 订阅时+每5秒更新 | 多周期K线数据 |

### 机器人引擎调用
| API | 调用频率 | 用途 |
|-----|---------|------|
| GetBalance | 下单前 | 获取可用余额 |
| GetPositions | 定期同步（带缓存） | 持仓状态同步 |
| GetOrderHistory | 定期同步 | 订单状态同步 |
| CreateOrder | 触发开仓时 | 执行下单 |
| ClosePosition | 触发平仓时 | 执行平仓 |
| SetLeverage | 下单前 | 设置杠杆 |
| GetTicker | 检测手动平仓时 | 获取最新价格 |

---

## 五、优化策略

### 1. 缓存机制
- **GetBalance**: 使用缓存 + singleflight，避免并发重复请求
- **GetPositions**: 使用缓存 + singleflight，定期刷新
- **GetTicker**: 全局服务统一缓存，所有机器人共享

### 2. 调用合并
- 使用 singleflight 模式，同一时间多个请求合并为一个API调用
- 全局行情服务统一管理，避免每个机器人独立调用

### 3. 数据获取策略
- **Ticker（实时行情）**:
  - MarketServiceManager: **优先WebSocket**（Bitget/Binance），失败时降级HTTP缓存
  - MarketDataService: HTTP轮询（每秒）
  - ExchangeMarketService: HTTP轮询（每秒）
  
- **K线数据**:
  - **全部使用HTTP轮询**（WebSocket暂不支持K线流）
  - 订阅时主动获取历史K线，后续每5秒更新

### 4. 降级策略
- MarketServiceManager 优先使用 WebSocket，失败时降级到 HTTP
- API 失败时返回缓存数据，保证系统稳定性
- WebSocket仅支持Bitget和Binance的Ticker数据，其他平台和K线数据使用HTTP

---

## 六、支持的平台

当前支持的交易所平台：
- Binance（币安）
- Bitget（币格）
- OKX（欧易）
- Gate（Gate.io）

每个平台都实现了 `Exchange` 接口，通过 `ExchangeManager` 统一管理。

