# HotGo V2 / Toogo.Ai 技术架构详细分析

## 📐 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        前端层 (Vue3)                          │
├─────────────────────────────────────────────────────────────┤
│  Admin管理端  │  Home前台  │  API接口  │  WebSocket客户端   │
└─────────────────────────────────────────────────────────────┘
                            ↕ HTTP/WebSocket
┌─────────────────────────────────────────────────────────────┐
│                      API网关层 (GoFrame)                      │
├─────────────────────────────────────────────────────────────┤
│  路由分发  │  中间件  │  认证授权  │  请求限流  │  日志记录   │
└─────────────────────────────────────────────────────────────┘
                            ↕
┌─────────────────────────────────────────────────────────────┐
│                     控制器层 (Controller)                     │
├─────────────────────────────────────────────────────────────┤
│  Admin控制器  │  Trading控制器  │  Toogo控制器  │  WebSocket   │
└─────────────────────────────────────────────────────────────┘
                            ↕
┌─────────────────────────────────────────────────────────────┐
│                     业务逻辑层 (Logic)                        │
├─────────────────────────────────────────────────────────────┤
│  Toogo业务  │  交易业务  │  系统管理  │  支付业务  │  推送服务  │
└─────────────────────────────────────────────────────────────┘
                            ↕
┌─────────────────────────────────────────────────────────────┐
│                     服务接口层 (Service)                      │
├─────────────────────────────────────────────────────────────┤
│  接口定义  │  依赖注入  │  服务注册  │  接口实现              │
└─────────────────────────────────────────────────────────────┘
                            ↕
┌─────────────────────────────────────────────────────────────┐
│                     数据访问层 (DAO)                          │
├─────────────────────────────────────────────────────────────┤
│  ORM操作  │  查询构建  │  事务管理  │  数据验证              │
└─────────────────────────────────────────────────────────────┘
                            ↕
┌─────────────────────────────────────────────────────────────┐
│                     数据存储层                                │
├─────────────────────────────────────────────────────────────┤
│  MySQL/PostgreSQL  │  Redis  │  消息队列  │  文件存储        │
└─────────────────────────────────────────────────────────────┘
                            ↕
┌─────────────────────────────────────────────────────────────┐
│                     外部服务层                                │
├─────────────────────────────────────────────────────────────┤
│  交易所API  │  支付网关  │  短信服务  │  邮件服务            │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔄 核心业务流程

### 1. 机器人交易流程

```
用户创建机器人
    ↓
配置API密钥和策略
    ↓
启动机器人 (status=2)
    ↓
定时任务每10秒执行
    ↓
┌─────────────────────────────────────┐
│  机器人引擎执行流程                  │
├─────────────────────────────────────┤
│ 1. 获取交易所连接 (ExchangeManager) │
│ 2. 获取实时行情 (GetTicker)         │
│ 3. 获取当前持仓 (GetPositions)      │
│ 4. 检查盈亏限制                      │
│ 5. 分析市场状态                      │
│    ├─ 趋势分析                       │
│    ├─ 波动分析                       │
│    └─ 信号生成                       │
│ 6. 执行交易决策                      │
│    ├─ 开仓条件检查                   │
│    ├─ 平仓条件检查                   │
│    └─ 反向下单策略                   │
│ 7. 记录订单和盈亏                    │
│ 8. WebSocket推送状态                 │
└─────────────────────────────────────┘
    ↓
继续下一轮循环
```

### 2. 支付充值流程

```
用户发起充值
    ↓
创建充值订单 (hg_toogo_deposit)
    ↓
调用支付网关 (NOWPayments)
    ↓
生成支付链接
    ↓
用户完成支付
    ↓
支付网关回调
    ↓
验证支付签名
    ↓
更新订单状态
    ↓
更新钱包余额 (hg_toogo_wallet)
    ↓
记录钱包流水 (hg_toogo_wallet_log)
    ↓
WebSocket推送通知
```

### 3. WebSocket实时推送流程

```
客户端连接WebSocket
    ↓
服务端创建Client连接
    ↓
注册到Hub管理器
    ↓
客户端订阅主题
    ├─ robot (机器人状态)
    ├─ ticker (实时行情)
    └─ position (持仓信息)
    ↓
定时任务/业务逻辑触发推送
    ↓
Hub广播消息到订阅的客户端
    ↓
客户端接收并更新UI
```

---

## 🏛️ 核心组件详解

### 1. TradingEngine (交易引擎)

**位置**：`internal/logic/toogo/engine.go`

**职责**：
- 管理所有运行中的机器人
- 执行机器人交易逻辑
- 缓存机器人运行器
- 市场分析和信号生成

**关键结构**：
```go
type TradingEngine struct {
    mu sync.RWMutex
    robots map[int64]*RobotRunner  // 机器人运行器缓存
}

type RobotRunner struct {
    Robot         *entity.TradingRobot
    Exchange      exchange.Exchange
    Wallet        *entity.ToogoWallet
    User          *entity.ToogoUser
    LastTicker    *exchange.Ticker
    Positions     map[string]*PositionTracker
    IsClosing     bool
    StrategyConfig *StrategyConfig
}
```

**核心方法**：
- `RunAllRobots()`: 运行所有机器人
- `runSingleRobot()`: 运行单个机器人
- `analyzeMarket()`: 分析市场生成信号
- `executeTrade()`: 执行交易

### 2. ExchangeManager (交易所管理器)

**位置**：`internal/library/exchange/manager.go`

**职责**：
- 管理交易所连接实例
- 连接复用和缓存
- 线程安全的连接管理

**关键结构**：
```go
type Manager struct {
    exchanges sync.Map  // map[int64]Exchange, key为apiConfigId
}
```

**核心方法**：
- `GetExchange()`: 获取交易所实例（缓存优先）
- `RemoveExchange()`: 移除交易所实例
- `TestConnection()`: 测试API连接

**连接复用机制**：
```go
// 第一次使用某个API配置时创建连接
ex, err := GetManager().GetExchange(ctx, apiConfigId)

// 后续使用相同API配置时，直接返回缓存的连接
// 避免重复创建连接，提高性能
```

### 3. WebSocketPusher (推送服务)

**位置**：`internal/library/websocket/pusher.go`

**职责**：
- 管理WebSocket连接
- 消息推送
- 订阅管理

**关键结构**：
```go
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}

type Client struct {
    hub    *Hub
    conn   *websocket.Conn
    send   chan []byte
    topics map[string]bool  // 订阅的主题
}
```

**推送类型**：
- `PushRobotStatus()`: 推送机器人状态
- `PushTicker()`: 推送实时行情
- `PushPosition()`: 推送持仓信息

### 4. StrategyConfig (策略配置)

**位置**：`internal/logic/toogo/engine.go`

**职责**：
- 解析策略配置JSON
- 策略参数管理
- 反向下单配置

**关键结构**：
```go
type StrategyConfig struct {
    StrategyId           int64
    StrategyKey          string
    StrategyName         string
    MonitorWindow        int      // 时间窗口(秒)
    VolatilityThreshold  float64  // 波动点数(USDT)
    LeverageMin          int
    LeverageMax          int
    MarginPercentMin     float64
    MarginPercentMax     float64
    Config               *ReverseConfig
}

type ReverseConfig struct {
    Enabled       bool
    LossRatio     float64
    ProfitRatio   float64
    TrailingStop  bool
}
```

---

## 🔌 交易所对接架构

### Exchange接口定义

**位置**：`internal/library/exchange/exchange.go`

**接口方法**：
```go
type Exchange interface {
    // 基础信息
    GetPlatform() string
    GetTestnet() bool
    
    // 市场数据
    GetTicker(ctx context.Context, symbol string) (*Ticker, error)
    GetKlines(ctx context.Context, symbol string, interval string, limit int) ([]*Kline, error)
    
    // 账户信息
    GetBalance(ctx context.Context) (*Balance, error)
    GetPositions(ctx context.Context, symbol string) ([]*Position, error)
    
    // 交易操作
    PlaceOrder(ctx context.Context, req *OrderRequest) (*Order, error)
    CancelOrder(ctx context.Context, symbol string, orderId string) error
    GetOrder(ctx context.Context, symbol string, orderId string) (*Order, error)
}
```

### 交易所实现

#### Binance实现
**位置**：`internal/library/exchange/binance.go`
- 支持现货和合约交易
- 支持测试网和主网
- WebSocket行情支持

#### OKX实现
**位置**：`internal/library/exchange/okx.go`
- 支持合约交易
- Passphrase认证
- 限流处理

#### Bitget实现
**位置**：`internal/library/exchange/bitget.go`
- 支持合约交易
- 签名算法实现
- 错误处理

#### Gate.io实现
**位置**：`internal/library/exchange/gate.go`
- 支持合约交易
- API认证
- 请求签名

---

## 📡 WebSocket架构

### 连接管理

```
客户端连接
    ↓
创建Client实例
    ↓
注册到Hub
    ↓
启动读写Goroutine
    ├─ readPump: 接收客户端消息
    └─ writePump: 发送消息到客户端
    ↓
订阅主题
    ↓
接收推送消息
    ↓
断开连接时清理
```

### 消息格式

```json
{
    "type": "robot_status|ticker|position",
    "data": {
        // 具体数据
    },
    "timestamp": 1234567890
}
```

### 订阅机制

```javascript
// 前端订阅
ws.send(JSON.stringify({
    action: "subscribe",
    topic: "robot",
    robotId: 123
}));

// 取消订阅
ws.send(JSON.stringify({
    action: "unsubscribe",
    topic: "robot",
    robotId: 123
}));
```

---

## 🗄️ 数据库设计模式

### 表命名规范
- `hg_` 前缀：HotGo系统表
- `toogo_` 前缀：Toogo业务表
- `trading_` 前缀：交易相关表

### 关键表关系

```
hg_admin_member (用户)
    ↓
hg_toogo_user (用户扩展)
    ↓
hg_toogo_wallet (钱包)
    ↓
hg_toogo_subscription (订阅)
    ↓
hg_trading_robot (机器人)
    ↓
hg_trading_order (订单)
```

### 索引设计
- 主键索引：所有表都有主键
- 外键索引：关联字段建立索引
- 查询索引：常用查询字段建立索引
- 复合索引：多字段组合查询

---

## 🔐 安全机制

### 1. API密钥加密

**位置**：`utility/encrypt/aes.go`

**加密流程**：
```
明文API密钥
    ↓
AES加密
    ↓
Base64编码
    ↓
存储到数据库
```

**解密流程**：
```
数据库读取
    ↓
Base64解码
    ↓
AES解密
    ↓
使用明文密钥
```

### 2. JWT认证

**Token结构**：
```json
{
    "user_id": 123,
    "username": "admin",
    "exp": 1234567890,
    "iat": 1234567800
}
```

**验证流程**：
```
请求携带Token
    ↓
中间件验证Token
    ↓
解析用户信息
    ↓
设置到Context
    ↓
业务逻辑使用
```

### 3. Casbin权限

**权限模型**：
```
用户 → 角色 → 权限 → 资源
```

**权限检查**：
```go
// 检查用户是否有权限
enforcer.Enforce(userId, resource, action)
```

---

## ⚡ 性能优化策略

### 1. 连接复用
- **交易所连接**：相同API配置共享连接
- **数据库连接**：连接池管理
- **Redis连接**：连接池复用

### 2. 缓存策略
- **机器人运行器**：内存缓存
- **交易所实例**：内存缓存
- **系统配置**：Redis缓存
- **用户信息**：Redis缓存

### 3. 并发处理
- **Goroutine并发**：多个机器人并发执行
- **读写锁**：保证线程安全
- **消息队列**：异步处理耗时操作

### 4. 数据库优化
- **索引优化**：合理建立索引
- **查询优化**：避免N+1查询
- **批量操作**：批量插入/更新
- **分页查询**：大数据量分页

---

## 🧪 测试策略

### 1. 单元测试
- **Logic层测试**：业务逻辑测试
- **DAO层测试**：数据访问测试
- **工具函数测试**：工具函数测试

### 2. 集成测试
- **API测试**：接口集成测试
- **数据库测试**：数据库操作测试
- **第三方服务测试**：交易所API测试

### 3. 压力测试
- **并发测试**：多机器人并发执行
- **负载测试**：高负载场景测试
- **稳定性测试**：长时间运行测试

---

## 📦 部署架构

### 单机部署
```
┌─────────────────────────┐
│  前端 (Nginx)           │
├─────────────────────────┤
│  后端 (Go)              │
│  ├─ HTTP服务            │
│  ├─ Cron定时任务        │
│  └─ Queue消息队列       │
├─────────────────────────┤
│  数据库 (MySQL)         │
├─────────────────────────┤
│  缓存 (Redis)           │
└─────────────────────────┘
```

### 分布式部署（未来）
```
┌──────────────┐  ┌──────────────┐
│  前端集群    │  │  负载均衡    │
└──────────────┘  └──────────────┘
                        ↓
┌──────────────┐  ┌──────────────┐
│  后端实例1   │  │  后端实例2   │
└──────────────┘  └──────────────┘
                        ↓
┌──────────────┐  ┌──────────────┐
│  数据库主    │  │  数据库从    │
└──────────────┘  └──────────────┘
┌──────────────┐  ┌──────────────┐
│  Redis集群   │  │  消息队列    │
└──────────────┘  └──────────────┘
```

---

## 🔍 监控与日志

### 1. 日志系统
- **访问日志**：记录所有HTTP请求
- **操作日志**：记录用户操作
- **错误日志**：记录系统错误
- **交易日志**：记录交易操作

### 2. 监控指标
- **系统监控**：CPU、内存、磁盘、网络
- **业务监控**：机器人运行状态、订单数量
- **性能监控**：响应时间、并发数
- **错误监控**：错误率、异常告警

### 3. 告警机制
- **系统告警**：系统资源告警
- **业务告警**：业务异常告警
- **交易告警**：交易异常告警

---

## 📚 开发规范

### 1. 代码分层
```
API层 → Controller层 → Logic层 → DAO层 → 数据库
```

### 2. 错误处理
```go
// 统一错误处理
if err != nil {
    return nil, gerror.Wrap(err, "操作失败")
}
```

### 3. 日志记录
```go
// 关键操作记录日志
g.Log().Infof(ctx, "操作成功: %v", data)
g.Log().Errorf(ctx, "操作失败: %+v", err)
```

### 4. 注释规范
```go
// 包注释
package toogo

// 函数注释
// FunctionName 函数说明
// @param ctx 上下文
// @return result 返回结果
func FunctionName(ctx context.Context) (result string, err error) {
    // 实现代码
}
```

---

## 🎯 总结

**HotGo V2 / Toogo.Ai** 采用现代化的微服务架构设计，具有以下特点：

1. **分层清晰**：API → Controller → Logic → DAO，职责明确
2. **组件化设计**：交易引擎、交易所管理器、推送服务等独立组件
3. **高并发支持**：Goroutine并发、连接复用、缓存机制
4. **可扩展性强**：插件化架构、接口设计、配置化
5. **安全性高**：JWT认证、Casbin权限、数据加密

项目架构设计合理，代码组织清晰，具有良好的可维护性和扩展性。

