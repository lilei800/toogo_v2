# 全局引擎架构说明

## 📋 架构概述

全局引擎采用**管理器 + 多实例**的架构模式：

```
MarketServiceManager (全局单例)
    ├── ExchangeMarketService (Binance)
    │   ├── Tickers 缓存 (按 symbol)
    │   ├── Klines 缓存 (按 symbol)
    │   └── Subscriptions 订阅管理
    ├── ExchangeMarketService (Bitget)
    │   ├── Tickers 缓存
    │   ├── Klines 缓存
    │   └── Subscriptions 订阅管理
    ├── ExchangeMarketService (OKX)
    │   └── ...
    └── ExchangeMarketService (Gate.io)
        └── ...
```

## 🏗️ 架构层次

### 1. MarketServiceManager（全局管理器）

**角色**：全局单例，统一管理所有交易所的行情服务

**职责**：
- 管理多个 `ExchangeMarketService` 实例
- 提供统一的访问接口
- 协调各交易所服务的生命周期

**代码位置**：
```go
// server/internal/library/market/market_service_manager.go
type MarketServiceManager struct {
    mu sync.RWMutex
    
    // 每个交易所一个行情服务 key: platform (binance/bitget/okx/gate)
    services map[string]*ExchangeMarketService
    
    running bool
    stopCh  chan struct{}
}
```

**关键方法**：
- `GetOrCreateService()` - 获取或创建交易所服务（按需创建）
- `GetService()` - 获取已存在的交易所服务
- `Subscribe()` - 订阅交易对（自动创建服务）
- `GetMultiTimeframeKlines()` - 获取多周期K线（从对应交易所服务）

### 2. ExchangeMarketService（交易所行情服务）

**角色**：每个交易所一个独立实例，管理该交易所的所有行情数据

**职责**：
- 管理该交易所的 Ticker 缓存
- 管理该交易所的 K线缓存（按 symbol 存储）
- 管理该交易所的订阅（引用计数）
- 定时更新该交易所的行情数据

**代码位置**：
```go
// server/internal/library/market/market_service_manager.go
type ExchangeMarketService struct {
    mu sync.RWMutex
    
    Platform string            // 交易所名称
    Exchange exchange.Exchange // 交易所API实例
    
    // 行情数据缓存 key: symbol
    Tickers    map[string]*TickerCache
    Klines     map[string]*KlineCache
    OrderBooks map[string]*OrderBookCache
    
    // 订阅的交易对 key: symbol, value: 引用计数
    Subscriptions map[string]int
    
    running bool
    stopCh  chan struct{}
}
```

**关键方法**：
- `Start()` - 启动服务（启动定时更新任务）
- `Subscribe()` - 订阅交易对（引用计数+1）
- `Unsubscribe()` - 取消订阅（引用计数-1）
- `GetMultiTimeframeKlines()` - 获取多周期K线缓存
- `runKlineUpdater()` - 定时更新K线（每5秒）

## 🔄 数据流程

### 订阅流程

```
机器人启动
    ↓
调用 MarketServiceManager.Subscribe(platform, symbol, exchange)
    ↓
MarketServiceManager.GetOrCreateService(platform) 
    ↓ (如果不存在)
创建 ExchangeMarketService 实例
    ↓
启动 ExchangeMarketService.Start()
    ↓
启动定时更新任务（runTickerUpdater, runKlineUpdater）
    ↓
ExchangeMarketService.Subscribe(symbol)
    ↓
首次订阅，立即获取数据（fetchInitialData）
    ↓
并发获取多周期K线（fetchAllKlines）
    ↓
存储到 Klines[symbol] 缓存
```

### 读取流程

```
机器人需要K线数据
    ↓
调用 MarketServiceManager.GetMultiTimeframeKlines(platform, symbol)
    ↓
获取对应的 ExchangeMarketService
    ↓
从 ExchangeMarketService.Klines[symbol] 读取缓存
    ↓
返回 KlineCache（包含 1m/5m/15m/30m/1h）
```

### 更新流程

```
定时任务（每5秒）
    ↓
ExchangeMarketService.runKlineUpdater()
    ↓
遍历所有订阅的交易对（Subscriptions）
    ↓
并发获取多周期K线（fetchAllKlines）
    ↓
更新 Klines[symbol] 缓存
```

## 📊 数据结构

### KlineCache（K线缓存）

```go
type KlineCache struct {
    Klines1m  []*exchange.Kline // 1分钟K线
    Klines5m  []*exchange.Kline // 5分钟K线
    Klines15m []*exchange.Kline // 15分钟K线
    Klines30m []*exchange.Kline // 30分钟K线
    Klines1h  []*exchange.Kline // 1小时K线
    UpdatedAt time.Time         // 更新时间
}
```

**存储位置**：
- `ExchangeMarketService.Klines[symbol] = KlineCache`

**更新频率**：
- 每5秒更新一次（`runKlineUpdater`）

## 🎯 关键特性

### 1. 按需创建

- 只有当有机器人订阅某个交易所时，才创建对应的 `ExchangeMarketService`
- 避免不必要的资源占用

### 2. 引用计数

- 每个交易对（symbol）使用引用计数管理订阅
- 多个机器人可以共享同一个交易对的K线缓存
- 当引用计数为0时，自动清理缓存

### 3. 独立更新

- 每个交易所服务独立运行定时更新任务
- 互不干扰，提高稳定性

### 4. 并发安全

- 使用 `sync.RWMutex` 保护共享数据
- 支持并发读取，互斥写入

## 🔍 使用示例

### 获取K线缓存

```go
// 从全局引擎获取K线缓存
platform := "binance"
symbol := "BTCUSDT"
klineCache := market.GetMarketServiceManager().GetMultiTimeframeKlines(platform, symbol)

if klineCache != nil {
    klines1m := klineCache.Klines1m
    klines5m := klineCache.Klines5m
    klines15m := klineCache.Klines15m
    klines30m := klineCache.Klines30m
    klines1h := klineCache.Klines1h
}
```

### 订阅行情服务

```go
// 机器人启动时订阅
platform := "binance"
symbol := "BTCUSDT"
exchange := binanceExchange // 交易所实例

market.GetMarketServiceManager().Subscribe(ctx, platform, symbol, exchange)
```

### 取消订阅

```go
// 机器人停止时取消订阅
market.GetMarketServiceManager().Unsubscribe(platform, symbol)
```

## 📈 优势

### 1. 资源隔离

- 每个交易所独立管理，互不影响
- 某个交易所出问题，不影响其他交易所

### 2. 性能优化

- 统一缓存，避免重复API调用
- 多个机器人共享同一份K线数据
- 定时更新，数据保持新鲜

### 3. 易于扩展

- 新增交易所只需创建新的 `ExchangeMarketService`
- 不影响现有交易所服务

### 4. 自动管理

- 引用计数自动管理订阅
- 定时任务自动更新数据
- 无需手动管理

## 🎯 总结

**架构确认**：
- ✅ **1个全局管理器**：`MarketServiceManager`（单例）
- ✅ **N个交易所服务**：每个交易所一个 `ExchangeMarketService` 实例
- ✅ **按需创建**：只有当有机器人订阅时才创建服务
- ✅ **独立管理**：每个交易所服务独立管理自己的缓存和更新任务

**数据存储**：
- 每个 `ExchangeMarketService` 维护自己的 `Klines map[string]*KlineCache`
- key 是 symbol（交易对），value 是 K线缓存
- 每个 symbol 的缓存包含 5 个周期的K线数据

**更新机制**：
- 每个 `ExchangeMarketService` 独立运行定时更新任务
- 每5秒更新一次该交易所所有订阅交易对的K线数据

---

**文档版本**：v1.0  
**最后更新**：2024年

