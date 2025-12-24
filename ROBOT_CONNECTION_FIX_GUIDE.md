# 🔧 机器人连接失败修复指南

## 📊 问题诊断

**症状：** 点击机器人后显示"连接失败，正在分析市场"

**根本原因：** 数据库表结构错误 + 可能的配置问题

---

## ✅ 解决方案（按顺序执行）

### 🔴 步骤1：修复数据库表结构【必须执行】

**问题：** `hg_sys_serve_log` 表的 `line` 字段缺少默认值

**执行SQL修复：**

```powershell
# 方法1：通过命令行（推荐）
cd D:\go\src\hotgo_v2\server\storage\data
mysql -u root -p hotgo < fix_database.sql

# 方法2：通过MySQL客户端
# 打开 Navicat/MySQL Workbench 等工具，连接数据库，执行以下SQL：
```

```sql
ALTER TABLE `hg_sys_serve_log` 
MODIFY COLUMN `line` INT DEFAULT 0 COMMENT '行号';
```

**验证修复：**
```sql
DESC hg_sys_serve_log;
-- 查看 line 字段是否有 Default 值为 0
```

---

### 🟡 步骤2：检查代理配置

**检查文件：** `server/manifest/config/config.yaml`

```yaml
# 如果需要访问交易所，必须配置代理（中国大陆）
exchange:
  proxy:
    enabled: true                    # ✅ 确保为 true
    type: "socks5"                   # socks5 或 http
    host: "127.0.0.1"                # 代理地址
    port: 10808                      # 代理端口
```

**验证代理是否可用：**
```powershell
# 测试代理连接
curl -x socks5://127.0.0.1:10808 https://api.binance.com/api/v3/time

# 如果返回 JSON 数据，说明代理正常
# 如果超时或报错，说明代理不可用
```

---

### 🟡 步骤3：检查API配置

**位置：** 后台管理 → Trading → API配置

**必须检查项：**

| 项目 | 要求 | 说明 |
|------|------|------|
| **API Key** | ✅ 有效 | 从交易所复制的正确Key |
| **Secret Key** | ✅ 有效 | 从交易所复制的正确Secret |
| **Passphrase** | ✅ 填写 | OKX/Bitget必填 |
| **交易所平台** | ✅ 正确 | binance/okx/bitget/gate |
| **API权限** | ✅ 开启 | 必须开启"合约交易"权限 |
| **IP白名单** | ⚠️ 配置 | 如果设置了，添加服务器IP |

**测试API连接：**

1. 进入后台 → Trading → API配置
2. 点击某个API配置的"测试连接"按钮
3. 如果显示"连接成功"并返回余额，说明API配置正确

---

### 🟡 步骤4：检查机器人配置

**必须完整配置项：**

```
✅ 机器人名称
✅ 选择API配置（必须先创建API配置）
✅ 选择交易对（如 BTC/USDT）
✅ 选择策略模板
✅ 设置杠杆倍数（1-125）
✅ 设置保证金比例（1-100%）
✅ 止损百分比（建议10-30%）
✅ 止盈回撤百分比（建议20-50%）
```

**可选配置：**
```
⚪ 最大盈利目标（USDT）
⚪ 最大亏损限制（USDT）
⚪ 最大运行时长（秒）
⚪ 启用反向下单
⚪ 启用信号监控
```

---

### 🟡 步骤5：检查算力余额

**位置：** 后台管理 → Toogo → 我的钱包

**检查项：**
- **算力余额 ≥ 10** （启动机器人最低要求）
- **可用算力 = 付费算力 + 赠送算力**

**如果算力不足：**
1. 充值USDT
2. 购买算力
3. 或者联系管理员赠送测试算力

---

## 🔍 详细诊断步骤

### 方法1：查看实时日志

```powershell
# 1. 进入日志目录
cd D:\go\src\hotgo_v2\server\logs\logger

# 2. 实时查看最新日志
Get-Content -Path "$(Get-Date -Format 'yyyy-MM-dd').log" -Tail 50 -Wait

# 3. 搜索错误信息
Select-String -Path "*.log" -Pattern "ERROR|ERRO|failed|失败" | Select-Object -Last 20
```

### 方法2：检查定时任务

```powershell
# 查看定时任务日志
cd D:\go\src\hotgo_v2\server\logs\cron
Get-Content -Path "$(Get-Date -Format 'yyyy-MM-dd').log" -Tail 20
```

### 方法3：检查数据库

```sql
-- 1. 检查机器人状态
SELECT id, robot_name, status, api_config_id, created_at 
FROM hg_trading_robot 
WHERE user_id = 你的用户ID
ORDER BY id DESC 
LIMIT 10;

-- 2. 检查API配置
SELECT id, api_name, platform, status, verify_status, verify_message 
FROM hg_trading_api_config 
WHERE user_id = 你的用户ID;

-- 3. 检查算力余额
SELECT user_id, power, gift_power, (power + gift_power) as total_power
FROM hg_toogo_wallet 
WHERE user_id = 你的用户ID;

-- 4. 检查最近错误日志
SELECT level_format, content, created_at 
FROM hg_sys_serve_log 
WHERE level_format IN ('ERRO', 'FATA', 'PANI')
ORDER BY created_at DESC 
LIMIT 20;
```

---

## 📋 常见错误及解决方案

### ❌ 错误1：Field 'line' doesn't have a default value

**原因：** 数据库表结构问题  
**解决：** 执行步骤1的SQL修复

---

### ❌ 错误2：Connection timeout / dial tcp i/o timeout

**原因：** 代理未配置或不可用  
**解决：**
1. 检查代理软件是否运行（v2ray/clash等）
2. 确认端口号正确（通常10808或7890）
3. 配置文件中启用代理 `enabled: true`

**测试代理：**
```powershell
# 测试SOCKS5代理
curl -x socks5://127.0.0.1:10808 https://api.binance.com/api/v3/time

# 测试HTTP代理
curl -x http://127.0.0.1:7890 https://api.binance.com/api/v3/time
```

---

### ❌ 错误3：API signature invalid / Invalid API-key

**原因：** API密钥错误或过期  
**解决：**
1. 重新从交易所复制API Key和Secret
2. 确认Passphrase正确（OKX/Bitget）
3. 检查API权限是否开启"合约交易"
4. 确认IP白名单包含服务器IP

---

### ❌ 错误4：Insufficient balance / 余额不足

**原因：** 交易所账户余额不足  
**解决：**
1. 登录交易所，充值USDT到合约账户
2. 或者降低"保证金使用比例"

---

### ❌ 错误5：算力不足

**原因：** Toogo算力余额 < 10  
**解决：**
1. 充值USDT购买算力
2. 或使用邀请码获得赠送算力

---

### ❌ 错误6：机器人数量已达上限

**原因：** 超过当前VIP等级的机器人配额  
**解决：**
1. 停止不需要的机器人
2. 升级VIP套餐
3. 购买更高级的订阅

---

## 🚀 完整启动流程

### 第一步：准备环境

```powershell
# 1. 启动代理软件（如 v2ray/clash）
# 2. 确认代理端口（通常10808或7890）
# 3. 修复数据库
cd D:\go\src\hotgo_v2\server\storage\data
mysql -u root -p hotgo < fix_database.sql
```

### 第二步：启动服务

```powershell
# 1. 启动后端
cd D:\go\src\hotgo_v2\server
go run main.go

# 2. 启动前端（新终端）
cd D:\go\src\hotgo_v2\web
pnpm run dev
```

### 第三步：配置检查

1. ✅ 访问 http://localhost:3000
2. ✅ 登录系统（admin / 123456）
3. ✅ 进入 Trading → API配置
4. ✅ 测试API连接是否成功
5. ✅ 检查算力余额是否充足

### 第四步：创建机器人

1. ✅ 进入 Trading → 机器人管理
2. ✅ 点击"创建机器人"
3. ✅ 完整填写所有必填项
4. ✅ 保存机器人
5. ✅ 点击"启动"按钮

### 第五步：监控运行

```powershell
# 实时查看日志
cd D:\go\src\hotgo_v2\server\logs\logger
Get-Content -Path "$(Get-Date -Format 'yyyy-MM-dd').log" -Tail 50 -Wait
```

**正常日志示例：**
```
2025-11-30T14:00:00 [INFO] [TradingEngine] 执行机器人任务: 共1个运行中
2025-11-30T14:00:01 [INFO] [Signal] robot=1, direction=LONG, strength=0.75
2025-11-30T14:00:02 [INFO] [Open] 开仓成功: robot=1, orderId=12345, side=BUY
```

---

## 🎯 快速检查清单

执行修复前，请逐项检查：

- [ ] 数据库表结构已修复（执行fix_database.sql）
- [ ] 代理软件正在运行（v2ray/clash）
- [ ] 代理配置正确（config.yaml中enabled=true）
- [ ] 代理连接正常（curl测试成功）
- [ ] API配置已创建
- [ ] API连接测试成功
- [ ] 交易所账户有余额
- [ ] Toogo算力余额 ≥ 10
- [ ] 机器人配置完整
- [ ] 后端服务正常运行
- [ ] 前端服务正常运行
- [ ] 定时任务正常执行（每10秒）

---

## 📞 获取更多帮助

### 查看日志位置

```
server/logs/
├── logger/      # 主日志
├── cron/        # 定时任务日志
├── database/    # 数据库查询日志
└── server/      # HTTP服务日志
```

### 关键日志搜索

```powershell
# 搜索错误
cd D:\go\src\hotgo_v2\server\logs\logger
Select-String -Path "*.log" -Pattern "ERRO|失败|failed" | Select-Object -Last 20

# 搜索机器人相关
Select-String -Path "*.log" -Pattern "TradingEngine|Robot|机器人" | Select-Object -Last 20

# 搜索API连接
Select-String -Path "*.log" -Pattern "API|Exchange|连接" | Select-Object -Last 20
```

---

## ✅ 修复完成验证

执行所有步骤后，验证是否成功：

1. **后端日志无ERROR**
   ```
   tail logs/logger/$(date +%Y-%m-%d).log
   # 应该没有 [ERRO] 或 [FATA] 级别的日志
   ```

2. **机器人状态显示"运行中"**
   - 前端界面显示绿色"运行中"状态
   - 实时更新盈亏数据

3. **定时任务正常执行**
   ```powershell
   # 每10秒应该有类似日志：
   [INFO] [TradingEngine] 执行机器人任务: 共N个运行中
   ```

4. **WebSocket连接正常**
   - 浏览器F12控制台无WebSocket错误
   - 实时行情数据更新

---

## 🎊 修复成功！

如果以上检查都通过，说明系统已正常运行！

**下一步：**
- 观察机器人运行日志
- 监控盈亏变化
- 根据市场调整策略

**祝您交易顺利！** 🚀

---

**文档版本：** v1.0  
**更新日期：** 2025-11-30  
**适用版本：** HotGo V2 + Toogo.Ai

