# 🚀 Toogo.Ai 部署指南

## 📋 部署清单

### 前置要求
- [x] Go 1.20+
- [x] Node.js 18+
- [x] MySQL 8.0+
- [x] Redis 6.0+

---

## 🗄️ 步骤一：数据库部署

### 方式一：分步导入（推荐）

```bash
# 1. 进入数据目录
cd D:\go\src\hotgo_v2\server\storage\data

# 2. 先确保HotGo基础表已导入
mysql -u root -p your_database < hotgo.sql

# 3. 导入Toogo系统表
mysql -u root -p your_database < toogo_system.sql

# 4. 导入Toogo菜单
mysql -u root -p your_database < toogo_menu.sql
```

### 方式二：使用MySQL客户端

1. 打开Navicat/DBeaver/HeidiSQL等工具
2. 连接到目标数据库
3. 依次执行以下SQL文件：
   - `toogo_system.sql` - 创建15张系统表 + 扩展机器人表
   - `toogo_menu.sql` - 导入菜单和定时任务

### 导入后验证

```sql
-- 检查表是否创建成功
SHOW TABLES LIKE 'hg_toogo_%';

-- 预期结果应该有以下表：
-- hg_toogo_user
-- hg_toogo_vip_level
-- hg_toogo_plan
-- hg_toogo_subscription
-- hg_toogo_wallet
-- hg_toogo_wallet_log
-- hg_toogo_deposit
-- hg_toogo_withdraw
-- hg_toogo_transfer
-- hg_toogo_agent_level
-- hg_toogo_commission_log
-- hg_toogo_strategy_template
-- hg_toogo_power_consume
-- hg_toogo_config
-- hg_toogo_ai_learning

-- 检查默认数据
SELECT COUNT(*) FROM hg_toogo_vip_level;        -- 应该 = 10
SELECT COUNT(*) FROM hg_toogo_plan;             -- 应该 = 5
SELECT COUNT(*) FROM hg_toogo_agent_level;      -- 应该 = 5
SELECT COUNT(*) FROM hg_toogo_strategy_template; -- 应该 = 12
SELECT COUNT(*) FROM hg_toogo_config;           -- 应该 > 15

-- 检查菜单导入
SELECT COUNT(*) FROM hg_admin_menu WHERE id >= 2000 AND id < 2200;  -- 应该 >= 16
```

---

## ⚙️ 步骤二：配置文件

### 1. 数据库配置
编辑 `server/manifest/config/config.yaml`:

```yaml
database:
  default:
    type: mysql
    host: 127.0.0.1
    port: 3306
    user: root
    pass: your_password
    name: your_database
    charset: utf8mb4
```

### 2. Redis配置

```yaml
redis:
  default:
    address: 127.0.0.1:6379
    db: 0
    pass: ""
```

### 3. 交易所配置
复制并编辑 `server/manifest/config/exchange.example.yaml` 为 `exchange.yaml`:

```yaml
exchange:
  binance:
    api_key: "YOUR_BINANCE_API_KEY"
    secret_key: "YOUR_BINANCE_SECRET_KEY"
    testnet: true  # 测试网

  okx:
    api_key: "YOUR_OKX_API_KEY"
    secret_key: "YOUR_OKX_SECRET_KEY"
    passphrase: "YOUR_OKX_PASSPHRASE"
    testnet: true

  bitget:
    api_key: "YOUR_BITGET_API_KEY"
    secret_key: "YOUR_BITGET_SECRET_KEY"
    passphrase: "YOUR_BITGET_PASSPHRASE"

  gate:
    api_key: "YOUR_GATE_API_KEY"
    secret_key: "YOUR_GATE_SECRET_KEY"
```

### 4. NOWPayments配置
复制并编辑 `server/manifest/config/nowpayments.example.yaml` 为 `nowpayments.yaml`:

```yaml
nowpayments:
  api_key: "YOUR_NOWPAYMENTS_API_KEY"
  ipn_secret: "YOUR_IPN_SECRET"
  sandbox: true  # 沙盒测试
  callback_url: "https://your-domain.com/api/admin/toogo/payment/callback"
```

---

## 🖥️ 步骤三：启动服务

### 启动后端

```bash
# 进入后端目录
cd D:\go\src\hotgo_v2\server

# 首次运行，下载依赖
go mod tidy

# 启动服务
go run main.go

# 或使用Air热重载(开发推荐)
air
```

后端启动成功后访问: `http://localhost:8000`

### 启动前端

```bash
# 进入前端目录
cd D:\go\src\hotgo_v2\web

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端启动成功后访问: `http://localhost:3100`

---

## ✅ 步骤四：功能测试

### 测试用例清单

#### 1. 用户模块测试
- [ ] 使用邀请码注册新用户
- [ ] 登录系统查看控制台
- [ ] 查看个人钱包余额

#### 2. 订阅模块测试
- [ ] 查看套餐列表
- [ ] 购买A/B/C/D套餐
- [ ] 验证订阅生效、到期时间

#### 3. 机器人模块测试
- [ ] 创建API接口配置
- [ ] 创建云机器人
- [ ] 启动/停止机器人
- [ ] 查看机器人详情

#### 4. 财务模块测试
- [ ] 充值测试(NOWPayments)
- [ ] 余额转算力测试
- [ ] 提现申请测试

#### 5. 后台管理测试
- [ ] 用户管理列表
- [ ] 套餐管理(增删改)
- [ ] VIP等级配置
- [ ] 提现审核

---

## 🔧 常见问题

### Q1: 菜单不显示？
检查菜单ID是否冲突，可手动修改toogo_menu.sql中的ID

### Q2: 定时任务不执行？
检查sys_cron表中的定时任务状态是否为启用(status=1)

### Q3: 交易所API报错？
1. 检查API Key/Secret是否正确
2. 确认IP是否已加入交易所白名单
3. 测试环境请开启testnet

### Q4: 数据库连接失败？
1. 检查MySQL服务是否启动
2. 验证用户名密码
3. 确认数据库已创建

---

## 📂 文件结构

```
hotgo_v2/
├── server/
│   ├── api/admin/toogo.go              # API定义
│   ├── internal/
│   │   ├── controller/admin/           # 控制器
│   │   ├── logic/toogo/                # 业务逻辑
│   │   ├── model/entity/               # 实体模型
│   │   ├── model/input/toogoin/        # 输入模型
│   │   ├── dao/                        # 数据访问
│   │   ├── service/toogo.go            # 服务接口
│   │   ├── crons/                      # 定时任务
│   │   └── library/exchange/           # 交易所封装
│   ├── storage/data/
│   │   ├── toogo_system.sql            # 系统表
│   │   └── toogo_menu.sql              # 菜单数据
│   └── manifest/config/
│       ├── exchange.example.yaml       # 交易所配置示例
│       └── nowpayments.example.yaml    # 支付配置示例
│
└── web/
    ├── src/api/toogo/                  # 前端API
    └── src/views/toogo/                # 前端页面
        ├── dashboard/                  # 控制台
        ├── subscription/               # 订阅
        ├── robot/                      # 机器人
        ├── team/                       # 团队
        ├── commission/                 # 佣金
        └── admin/                      # 后台管理
```

---

## 📞 技术支持

如有问题，请检查以下日志：
- 后端日志: `server/resource/log/`
- 前端控制台: 浏览器F12

---

**🎉 部署完成后，即可开始使用 Toogo.Ai 全自动量化交易系统！**

