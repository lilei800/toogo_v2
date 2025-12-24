# robot_engine.go 和 robot.go 文件说明

## 📋 文件概述

### 1. `robot_engine.go` - 机器人核心引擎

**文件类型**: 核心业务逻辑层（运行时对象）

**作用**: 每个机器人独立的运行引擎，负责机器人的核心交易逻辑

**特点**:
- 每个机器人一个独立的 `RobotEngine` 实例
- 运行在后台，持续监控市场并执行交易
- 包含市场分析、信号生成、交易执行等核心功能

---

### 2. `robot.go` - 机器人服务层

**文件类型**: HTTP API 服务层（业务接口）

**作用**: 处理用户的 HTTP 请求，提供机器人管理的业务接口

**特点**:
- 单例服务 `sToogoRobot`
- 处理用户的启动、停止、查询等操作
- 作为用户请求和核心引擎之间的桥梁

---

## 🔍 详细对比

### 职责对比

| 维度 | robot_engine.go | robot.go |
|------|----------------|----------|
| **层级** | 核心业务逻辑层 | HTTP API 服务层 |
| **实例化** | 每个机器人一个实例 | 全局单例服务 |
| **运行方式** | 后台持续运行 | 按需响应请求 |
| **主要功能** | 市场分析、信号生成、自动交易 | 机器人管理、查询、手动操作 |
| **生命周期** | 机器人启动时创建，停止时销毁 | 服务启动时创建，一直存在 |

---

## 📝 功能详解

### robot_engine.go 核心功能

#### 1. 市场分析 (`RobotAnalyzer`)
```go
// 分析市场状态、趋势、波动率等
engine.Analyzer.Analyze(ctx)
```

#### 2. 信号生成 (`RobotSignalGen`)
```go
// 生成做多/做空信号
engine.SignalGen.GenerateSignal(ctx)
```

#### 3. 交易执行 (`RobotTrader`)
```go
// 执行开仓、平仓操作
engine.Trader.CheckAndOpenPositionWithSignal(ctx, signal)
engine.Trader.CheckAndClosePosition(ctx)
```

#### 4. 状态管理
- 持仓跟踪 (`PositionTrackers`)
- 价格窗口监控 (`PriceWindow`)
- 账户数据同步 (`syncAccountData`)

#### 5. 主循环 (`runMainLoop`)
```go
// 持续运行的主循环
// - 每2秒执行市场分析和信号生成
// - 每10秒同步账户数据
// - 事件驱动：价格更新时立即检查平仓
```

**关键方法**:
- `Start()` - 启动引擎
- `Stop()` - 停止引擎
- `doAnalysis()` - 执行市场分析
- `doSignalGeneration()` - 生成交易信号
- `checkClosePosition()` - 检查平仓条件
- `syncAccountData()` - 同步账户数据

---

### robot.go 核心功能

#### 1. 机器人管理
```go
// 启动机器人
StartRobot(ctx, in *toogoin.StartRobotInp)

// 停止机器人
StopRobot(ctx, in *toogoin.StopRobotInp)
```

#### 2. 查询功能
```go
// 获取机器人列表
RobotList(ctx, in *toogoin.RobotListInp)

// 获取机器人持仓
GetRobotPositions(ctx, robotId int64)

// 获取机器人订单
GetRobotOpenOrders(ctx, robotId int64)
GetRobotOrderHistory(ctx, robotId int64, limit int)
```

#### 3. 手动操作
```go
// 手动平仓
CloseRobotPosition(ctx, in *toogoin.ClosePositionInp)

// 撤销挂单
CancelRobotOrder(ctx, robotId int64, orderId string)
```

#### 4. 数据清理
```go
// 清理信号日志
ClearSignalLogs(ctx, robotId int64, keepExecuted bool)
```

**关键方法**:
- `StartRobot()` - 启动机器人（更新数据库状态）
- `StopRobot()` - 停止机器人（更新数据库状态）
- `GetRobotPositions()` - 查询机器人持仓
- `CloseRobotPosition()` - 手动平仓
- `RobotList()` - 查询机器人列表

---

## 🔄 两者关系

### 调用关系

```
用户请求 (HTTP API)
    ↓
robot.go (服务层)
    ↓
RobotTaskManager (任务管理器)
    ↓
robot_engine.go (核心引擎)
    ↓
交易所 API
```

### 数据流

```
1. 用户通过 HTTP API 调用 robot.go 的 StartRobot()
   ↓
2. robot.go 更新数据库状态，通知 RobotTaskManager
   ↓
3. RobotTaskManager 创建 RobotEngine 实例
   ↓
4. RobotEngine 开始运行，执行市场分析和交易
   ↓
5. 用户通过 robot.go 查询持仓、订单等信息
   ↓
6. robot.go 从数据库或 RobotEngine 获取数据返回给用户
```

---

## 💡 使用场景

### robot_engine.go 使用场景

**何时使用**:
- ✅ 需要访问机器人的运行时状态（持仓、信号等）
- ✅ 需要执行市场分析或信号生成
- ✅ 需要监控机器人的运行状态
- ✅ 需要获取价格窗口、信号历史等数据

**示例**:
```go
engine := GetRobotTaskManager().GetEngine(robotId)
if engine != nil {
    status := engine.GetStatus()
    positions := engine.CurrentPositions
}
```

---

### robot.go 使用场景

**何时使用**:
- ✅ 用户通过 HTTP API 启动/停止机器人
- ✅ 用户查询机器人列表、持仓、订单
- ✅ 用户手动平仓、撤销订单
- ✅ 需要更新数据库中的机器人状态

**示例**:
```go
// HTTP 请求处理
func (c *cToogoRobot) Start(ctx context.Context, req *toogoin.StartRobotReq) {
    service.ToogoRobot().StartRobot(ctx, &toogoin.StartRobotInp{
        RobotId: req.RobotId,
    })
}
```

---

## 🎯 关键区别总结

### 1. 生命周期

**robot_engine.go**:
- 机器人启动时创建
- 机器人停止时销毁
- 运行期间持续存在

**robot.go**:
- 服务启动时创建
- 服务停止时销毁
- 一直存在，处理所有机器人的请求

### 2. 数据访问

**robot_engine.go**:
- 访问运行时内存数据（最新状态）
- 缓存持仓、行情等数据
- 实时数据，性能高

**robot.go**:
- 访问数据库数据（持久化数据）
- 查询历史订单、持仓记录
- 数据持久化，可靠性高

### 3. 操作类型

**robot_engine.go**:
- 自动执行：市场分析、信号生成、自动交易
- 后台运行，无需用户干预

**robot.go**:
- 手动操作：启动、停止、查询、手动平仓
- 响应用户请求，需要用户触发

---

## 📚 相关文件

### 与 robot_engine.go 相关的文件
- `robot_task_manager.go` - 管理所有 RobotEngine 实例
- `order_status_sync.go` - 订单状态同步服务（使用 RobotEngine）
- `order_status.go` - 订单状态管理（使用 RobotEngine）

### 与 robot.go 相关的文件
- `controller/toogo/robot.go` - HTTP 控制器（调用 robot.go）
- `model/input/toogoin/robot.go` - 输入输出模型
- `dao/trading_robot.go` - 数据库访问层

---

## 🔧 代码示例

### 获取机器人引擎
```go
// 从 RobotTaskManager 获取引擎
engine := GetRobotTaskManager().GetEngine(robotId)
if engine != nil {
    // 访问引擎的运行时数据
    positions := engine.CurrentPositions
    status := engine.GetStatus()
}
```

### 调用机器人服务
```go
// 启动机器人
err := service.ToogoRobot().StartRobot(ctx, &toogoin.StartRobotInp{
    RobotId: robotId,
})

// 查询持仓
positions, err := service.ToogoRobot().GetRobotPositions(ctx, robotId)
```

---

## ✅ 总结

- **robot_engine.go**: 核心引擎，负责机器人的自动交易逻辑，每个机器人一个实例
- **robot.go**: 服务层，处理用户的 HTTP 请求，提供机器人管理接口，全局单例

两者配合工作：
- `robot.go` 处理用户请求，管理机器人的生命周期
- `robot_engine.go` 执行核心交易逻辑，实现自动化交易

