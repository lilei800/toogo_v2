# ✅ Payment后端代码创建完成

## 🎉 已完成的工作

### 1. Logic层（业务逻辑）✅

创建了3个业务逻辑文件：

#### `internal/logic/payment/deposit.go`
- `Create()` - 创建充值订单
- `List()` - 充值订单列表
- `View()` - 查看充值订单详情
- `Check()` - 检查充值状态
- `Cancel()` - 取消充值订单
- `UpdateStatus()` - 更新订单状态（内部调用）

#### `internal/logic/payment/withdraw.go`
- `Apply()` - 申请提现
- `List()` - 提现订单列表
- `View()` - 查看提现订单详情
- `Audit()` - 审核提现（管理员）
- `Cancel()` - 取消提现
- `AdminList()` - 管理员查看提现列表

#### `internal/logic/payment/balance.go`
- `View()` - 查看余额
- `LogList()` - 资金流水列表

---

### 2. API定义层 ✅

创建了3个API定义文件：

#### `api/admin/payment/deposit.go`
- `DepositCreateReq/Res` - 创建充值订单
- `DepositListReq/Res` - 充值订单列表
- `DepositViewReq/Res` - 查看充值订单
- `DepositCheckReq/Res` - 检查充值状态
- `DepositCancelReq/Res` - 取消充值订单

#### `api/admin/payment/withdraw.go`
- `WithdrawApplyReq/Res` - 申请提现
- `WithdrawListReq/Res` - 提现订单列表
- `WithdrawViewReq/Res` - 查看提现订单
- `WithdrawAuditReq/Res` - 审核提现
- `WithdrawCancelReq/Res` - 取消提现

#### `api/admin/payment/balance.go`
- `BalanceViewReq/Res` - 查看余额
- `BalanceLogListReq/Res` - 资金流水列表

---

### 3. Controller层（控制器）✅

创建了3个控制器文件：

#### `internal/controller/admin/payment/deposit.go`
- 实现所有充值相关的HTTP请求处理

#### `internal/controller/admin/payment/withdraw.go`
- 实现所有提现相关的HTTP请求处理

#### `internal/controller/admin/payment/withdraw.go`
- 实现余额和流水相关的HTTP请求处理

---

### 4. Input模型 ✅

创建了输入模型文件：

#### `internal/model/input/payment.go`
- `DepositCreateInp` - 创建充值输入
- `DepositListInp` - 充值列表输入
- `WithdrawApplyInp` - 申请提现输入
- `WithdrawListInp` - 提现列表输入
- `WithdrawAuditInp` - 审核提现输入
- `BalanceLogListInp` - 资金流水列表输入
- `PageReq` - 分页请求

---

### 5. 路由注册 ✅

已在 `internal/router/admin.go` 中注册：

```go
import (
    "hotgo/internal/controller/admin/payment"
)

group.Bind(
    payment.Deposit,    // USDT充值
    payment.Withdraw,   // USDT提现
    payment.Balance,    // 余额管理
)
```

---

## 📊 完整的Payment API列表

### 充值相关
- POST `/admin/payment/deposit/create` - 创建充值订单
- GET `/admin/payment/deposit/list` - 充值订单列表
- GET `/admin/payment/deposit/view` - 查看充值订单
- POST `/admin/payment/deposit/check` - 检查充值状态
- POST `/admin/payment/deposit/cancel` - 取消充值订单

### 提现相关
- POST `/admin/payment/withdraw/apply` - 申请提现
- GET `/admin/payment/withdraw/list` - 提现订单列表
- GET `/admin/payment/withdraw/view` - 查看提现订单
- POST `/admin/payment/withdraw/audit` - 审核提现（管理员）
- POST `/admin/payment/withdraw/cancel` - 取消提现

### 余额相关
- GET `/admin/payment/balance/view` - 查看余额
- GET `/admin/payment/balance/logs` - 资金流水列表

---

## ⚠️ 需要的数据库表

Payment模块需要以下数据库表：

### 1. `hg_usdt_balance` - 余额表
```sql
CREATE TABLE `hg_usdt_balance` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `balance` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '可用余额',
  `frozen_balance` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '冻结余额',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='USDT余额表';
```

### 2. `hg_usdt_deposit` - 充值订单表
```sql
CREATE TABLE `hg_usdt_deposit` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `order_sn` varchar(64) NOT NULL COMMENT '订单号',
  `amount` decimal(20,8) NOT NULL COMMENT '充值金额',
  `network` varchar(32) NOT NULL COMMENT '网络(TRC20/ERC20)',
  `payment_id` varchar(128) DEFAULT NULL COMMENT '第三方支付ID',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态：1待支付 2已完成 3已超时 4已退款 5已取消',
  `paid_at` datetime DEFAULT NULL COMMENT '支付时间',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_sn` (`order_sn`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='USDT充值订单表';
```

### 3. `hg_usdt_withdraw` - 提现订单表
```sql
CREATE TABLE `hg_usdt_withdraw` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `order_sn` varchar(64) NOT NULL COMMENT '订单号',
  `amount` decimal(20,8) NOT NULL COMMENT '提现金额',
  `to_address` varchar(128) NOT NULL COMMENT '提现地址',
  `network` varchar(32) NOT NULL COMMENT '网络(TRC20/ERC20)',
  `tx_hash` varchar(128) DEFAULT NULL COMMENT '交易哈希',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态：1待审核 2审核通过 3审核拒绝 4已完成 5已取消',
  `audit_remark` varchar(255) DEFAULT NULL COMMENT '审核备注',
  `audited_by` bigint unsigned DEFAULT NULL COMMENT '审核人ID',
  `audited_at` datetime DEFAULT NULL COMMENT '审核时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_sn` (`order_sn`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='USDT提现订单表';
```

### 4. `hg_usdt_balance_log` - 资金流水表
```sql
CREATE TABLE `hg_usdt_balance_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `type` tinyint NOT NULL COMMENT '类型：1充值 2提现 3支付 4退款',
  `amount` decimal(20,8) NOT NULL COMMENT '变动金额',
  `balance` decimal(20,8) NOT NULL COMMENT '变动后余额',
  `order_sn` varchar(64) DEFAULT NULL COMMENT '关联订单号',
  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_type` (`type`),
  KEY `idx_order_sn` (`order_sn`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='USDT资金流水表';
```

---

## 🚀 下一步

### 1. 检查数据库表是否存在

```sql
SHOW TABLES LIKE 'hg_usdt_%';
```

### 2. 如果表不存在，需要创建

运行上面的4个CREATE TABLE语句。

### 3. 重启后端服务

```powershell
cd D:\go\src\hotgo_v2\server
go run main.go --args "all"
```

### 4. 刷新前端浏览器

按 `Ctrl + F5` 强制刷新

### 5. 测试Payment功能

点击 "USDT管理" 菜单，应该能正常显示数据了（即使是空数据）。

---

## 📝 文件清单

### Logic层（3个文件）
- ✅ `internal/logic/payment/deposit.go`
- ✅ `internal/logic/payment/withdraw.go`
- ✅ `internal/logic/payment/balance.go`

### API定义层（3个文件）
- ✅ `api/admin/payment/deposit.go`
- ✅ `api/admin/payment/withdraw.go`
- ✅ `api/admin/payment/balance.go`

### Controller层（3个文件）
- ✅ `internal/controller/admin/payment/deposit.go`
- ✅ `internal/controller/admin/payment/withdraw.go`
- ✅ `internal/controller/admin/payment/balance.go`

### Input模型（1个文件）
- ✅ `internal/model/input/payment.go`

### 路由注册
- ✅ `internal/router/admin.go` - 已添加Payment路由

---

**Payment模块后端代码已全部创建完成！** 🎊

**下一步：检查数据库表并重启服务！** 🚀



