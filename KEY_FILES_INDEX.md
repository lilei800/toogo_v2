# HotGo V2 / Toogo.Ai 关键文件索引

## 📋 快速导航

本文档列出了项目中的关键文件及其作用，方便快速定位和理解代码。

---

## 🎯 核心业务文件

### Toogo业务逻辑

#### 交易引擎
- **`server/internal/logic/toogo/engine.go`** - 交易引擎核心，机器人执行逻辑
- **`server/internal/logic/toogo/robot_engine.go`** - 机器人引擎实现
- **`server/internal/logic/toogo/exchange_manager.go`** - 交易所管理器
- **`server/internal/logic/toogo/robot.go`** - 机器人业务逻辑
- **`server/internal/logic/toogo/strategy_group.go`** - 策略组管理

#### 用户与钱包
- **`server/internal/logic/toogo/user.go`** - 用户管理逻辑
- **`server/internal/logic/toogo/wallet.go`** - 钱包管理逻辑
- **`server/internal/logic/toogo/subscription.go`** - 订阅管理逻辑
- **`server/internal/logic/toogo/commission.go`** - 佣金计算逻辑

#### 支付与财务
- **`server/internal/logic/toogo/finance.go`** - 财务管理逻辑
- **`server/internal/logic/toogo/finance_nowpayments.go`** - NOWPayments支付集成

#### 配置与推送
- **`server/internal/logic/toogo/config.go`** - 系统配置管理
- **`server/internal/logic/toogo/pusher.go`** - WebSocket推送服务

### 交易业务逻辑

- **`server/internal/logic/trading/robot.go`** - 交易机器人管理
- **`server/internal/logic/trading/monitor.go`** - 交易监控
- **`server/internal/logic/trading/alert_log.go`** - 告警日志
- **`server/internal/logic/trading/proxy_config.go`** - 代理配置管理

---

## 🔌 交易所对接

### 交易所接口
- **`server/internal/library/exchange/exchange.go`** - 交易所接口定义
- **`server/internal/library/exchange/manager.go`** - 交易所管理器
- **`server/internal/library/exchange/binance.go`** - Binance交易所实现
- **`server/internal/library/exchange/okx.go`** - OKX交易所实现
- **`server/internal/library/exchange/bitget.go`** - Bitget交易所实现
- **`server/internal/library/exchange/gate.go`** - Gate.io交易所实现

### 交易所配置
- **`server/manifest/config/exchange.example.yaml`** - 交易所配置示例

---

## 📡 WebSocket服务

- **`server/internal/library/websocket/hub.go`** - WebSocket连接管理
- **`server/internal/library/websocket/client.go`** - WebSocket客户端
- **`server/internal/library/websocket/handler.go`** - WebSocket处理器
- **`server/internal/library/websocket/pusher.go`** - 消息推送服务
- **`server/internal/websocket/router.go`** - WebSocket路由
- **`server/internal/controller/websocket/handler/`** - WebSocket控制器

---

## ⏰ 定时任务

- **`server/internal/crons/toogo_robot_engine.go`** - 机器人引擎定时任务（每10秒）
- **`server/internal/crons/toogo_robot_engine_v2.go`** - 机器人引擎V2版本
- **`server/internal/crons/toogo_engine.go`** - Toogo引擎定时任务
- **`server/internal/crons/close_order.go`** - 订单关闭定时任务

---

## 🎮 控制器层

### Toogo控制器
- **`server/internal/controller/admin/admin_toogo.go`** - Toogo通用控制器
- **`server/internal/controller/admin/admin_toogo_config.go`** - Toogo配置控制器
- **`server/internal/controller/admin/admin_payment_callback.go`** - 支付回调控制器

### 交易控制器
- **`server/internal/controller/admin/trading/robot.go`** - 交易机器人控制器
- **`server/internal/controller/admin/trading/alert.go`** - 告警控制器
- **`server/internal/controller/admin/trading/monitor.go`** - 监控控制器

---

## 📊 数据模型

### Entity实体
- **`server/internal/model/entity/toogo_user.go`** - 用户实体
- **`server/internal/model/entity/toogo_wallet.go`** - 钱包实体
- **`server/internal/model/entity/toogo_plan.go`** - 套餐实体
- **`server/internal/model/entity/toogo_subscription.go`** - 订阅实体
- **`server/internal/model/entity/toogo_strategy_template.go`** - 策略模板实体
- **`server/internal/model/entity/trading_robot.go`** - 交易机器人实体
- **`server/internal/model/entity/trading_api_config.go`** - API配置实体
- **`server/internal/model/entity/trading_order.go`** - 交易订单实体

### Input输入模型
- **`server/internal/model/input/toogoin/user.go`** - 用户输入模型
- **`server/internal/model/input/toogoin/wallet.go`** - 钱包输入模型
- **`server/internal/model/input/toogoin/robot.go`** - 机器人输入模型
- **`server/internal/model/input/toogoin/strategy.go`** - 策略输入模型

---

## 🗄️ 数据访问层

### DAO文件
- **`server/internal/dao/toogo_user.go`** - 用户DAO
- **`server/internal/dao/toogo_wallet.go`** - 钱包DAO
- **`server/internal/dao/toogo_robot.go`** - 机器人DAO（如果存在）
- **`server/internal/dao/trading_robot.go`** - 交易机器人DAO
- **`server/internal/dao/trading_api_config.go`** - API配置DAO
- **`server/internal/dao/trading_order.go`** - 交易订单DAO

---

## 🌐 API定义

- **`server/api/admin/toogo.go`** - Toogo API定义
- **`server/api/admin/toogo_config.go`** - Toogo配置API定义
- **`server/api/admin/payment_callback.go`** - 支付回调API定义
- **`server/api/admin/trading/robot.go`** - 交易机器人API定义

---

## 🛣️ 路由配置

- **`server/internal/router/admin.go`** - 后台管理路由
- **`server/internal/router/api.go`** - 对外API路由
- **`server/internal/router/home.go`** - 前台路由
- **`server/internal/router/websocket.go`** - WebSocket路由

---

## ⚙️ 服务接口

- **`server/internal/service/toogo.go`** - Toogo服务接口
- **`server/internal/service/trading.go`** - 交易服务接口（如果存在）

---

## 🎨 前端关键文件

### Toogo业务页面

#### 用户端
- **`web/src/views/toogo/dashboard/index.vue`** - 控制台首页
- **`web/src/views/toogo/robot/index.vue`** - 机器人列表（当前打开的文件）
- **`web/src/views/toogo/robot/create.vue`** - 创建机器人
- **`web/src/views/toogo/subscription/index.vue`** - 订阅套餐
- **`web/src/views/toogo/team/index.vue`** - 我的团队
- **`web/src/views/toogo/commission/index.vue`** - 佣金明细
- **`web/src/views/toogo/finance/index.vue`** - 财务管理
- **`web/src/views/toogo/strategy/index.vue`** - 策略管理

#### 管理端
- **`web/src/views/toogo/admin/user/index.vue`** - 用户管理
- **`web/src/views/toogo/admin/plan/index.vue`** - 套餐管理
- **`web/src/views/toogo/admin/robot/index.vue`** - 机器人管理
- **`web/src/views/toogo/admin/config/index.vue`** - 系统配置
- **`web/src/views/toogo/admin/withdraw/index.vue`** - 提现审核

### API封装
- **`web/src/api/toogo/index.ts`** - Toogo API封装
- **`web/src/api/trading/`** - 交易相关API

### WebSocket客户端
- **`web/src/utils/websocket.ts`** - WebSocket工具函数
- **`web/src/utils/websocket/`** - WebSocket相关工具

---

## 📦 配置文件

### 后端配置
- **`server/manifest/config/config.yaml`** - 主配置文件
- **`server/manifest/config/exchange.example.yaml`** - 交易所配置示例
- **`server/manifest/config/nowpayments.example.yaml`** - NOWPayments配置示例

### 前端配置
- **`web/vite.config.ts`** - Vite构建配置
- **`web/tsconfig.json`** - TypeScript配置
- **`web/package.json`** - 依赖配置

---

## 🗃️ 数据库文件

- **`server/storage/data/toogo_system.sql`** - Toogo系统表结构
- **`server/storage/data/toogo_menu.sql`** - Toogo菜单数据
- **`server/storage/data/toogo_install.sql`** - Toogo安装脚本
- **`server/storage/data/trading_system.sql`** - 交易系统表结构

---

## 📚 文档文件

- **`README.md`** - 项目主文档
- **`PROJECT_ANALYSIS.md`** - 项目全面分析（本文档）
- **`TECHNICAL_ARCHITECTURE.md`** - 技术架构详细分析
- **`server/TOOGO_ROBOT_ARCHITECTURE.md`** - 机器人架构文档
- **`TOOGO_COMPLETE_STATUS.md`** - Toogo完成状态报告
- **`docs/guide-zh-CN/`** - 中文开发文档

---

## 🔧 工具文件

### 加密工具
- **`server/utility/encrypt/aes.go`** - AES加密工具

### 数据库工具
- **`server/utility/db/`** - 数据库工具

---

## 🚀 启动文件

- **`server/main.go`** - 后端入口文件
- **`server/internal/cmd/http.go`** - HTTP服务启动
- **`server/internal/cmd/cron.go`** - 定时任务启动
- **`server/internal/cmd/queue.go`** - 消息队列启动
- **`web/src/main.ts`** - 前端入口文件

---

## 📝 常量定义

- **`server/internal/consts/toogo.go`** - Toogo常量定义
- **`server/internal/consts/trading.go`** - 交易常量定义
- **`server/internal/consts/app.go`** - 应用常量定义

---

## 🔍 快速查找指南

### 查找业务逻辑
1. 查看 `server/internal/logic/toogo/` 目录
2. 根据业务模块查找对应文件

### 查找API接口
1. 查看 `server/api/admin/` 目录
2. 查看 `server/internal/controller/admin/` 目录

### 查找前端页面
1. 查看 `web/src/views/toogo/` 目录
2. 根据页面功能查找对应文件

### 查找数据库表结构
1. 查看 `server/storage/data/` 目录
2. 查看 `server/internal/model/entity/` 目录

### 查找配置
1. 查看 `server/manifest/config/` 目录
2. 查看 `server/internal/model/config.go`

---

## 💡 开发建议

1. **新增功能**：按照分层架构，从API → Controller → Logic → DAO依次开发
2. **修改业务逻辑**：优先查看 `logic/toogo/` 目录下的文件
3. **添加前端页面**：参考 `web/src/views/toogo/` 下的现有页面
4. **数据库变更**：修改entity和dao文件，并更新SQL文件
5. **API变更**：修改api定义和controller文件

---

**最后更新**：2024年
**维护者**：开发团队

