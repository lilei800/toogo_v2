# ⚡ 快速修复指南（3分钟）

## 🔴 步骤1：修复数据库（必须执行）

### 方法A：使用批处理脚本（推荐，最简单）

```powershell
# 双击运行
D:\go\src\hotgo_v2\fix_database.bat

# 输入MySQL密码，回车
# 看到 "✅ 数据库修复成功！" 即完成
```

### 方法B：使用MySQL客户端（Navicat/Workbench）

1. 打开 Navicat 或 MySQL Workbench
2. 连接到 `hotgo` 数据库
3. 打开文件：`server/storage/data/EXECUTE_FIX_NOW.sql`
4. 执行SQL（点击运行按钮）
5. 看到 "✅ 数据库修复完成" 即可

### 方法C：命令行执行

```powershell
cd D:\go\src\hotgo_v2\server\storage\data
mysql -u root -p hotgo < EXECUTE_FIX_NOW.sql
# 输入密码后回车
```

---

## 🟡 步骤2：检查代理配置（已自动配置）

✅ **代理已启用：** `config.yaml` 中已配置

```yaml
exchange:
  proxy:
    enabled: true       # ✅ 已启用
    type: "socks5"      # ✅ SOCKS5
    host: "127.0.0.1"   # ✅ 本地
    port: 10808         # ✅ 端口10808
```

**请确认：**
- [ ] 代理软件正在运行（v2ray/clash/v2rayN等）
- [ ] 代理端口是 10808（如果不是，修改 config.yaml）

**测试代理：**
```powershell
curl -x socks5://127.0.0.1:10808 https://api.binance.com/api/v3/time
# 如果返回时间戳JSON，说明代理正常
```

---

## 🟢 步骤3：重启服务

```powershell
# 1. 停止当前服务（在运行 go run main.go 的终端按 Ctrl+C）

# 2. 重新启动
cd D:\go\src\hotgo_v2\server
go run main.go

# 3. 等待看到
# [INFO] HTTP Server started listening on [:8000]
```

---

## ✅ 验证修复

### 1. 检查日志无错误

```powershell
# 查看最新日志
cd D:\go\src\hotgo_v2\server\logs\logger
Get-Content -Path "$(Get-Date -Format 'yyyy-MM-dd').log" -Tail 20
```

**应该看到：**
```
✅ 没有 "Field 'line' doesn't have a default value"
✅ 没有持续的 [ERRO] 日志
✅ 看到 [INFO] [TradingEngine] 执行机器人任务
```

### 2. 前端测试

1. 打开浏览器：http://localhost:3000
2. 进入：Trading → 机器人管理
3. 点击某个机器人的"启动"按钮
4. 观察状态变化

**成功标志：**
```
✅ 状态显示"运行中"（绿色）
✅ 没有弹出错误提示
✅ 实时更新盈亏数据
✅ 控制台无报错（F12查看）
```

---

## 🔍 如果还是失败

### 情况1：代理连接失败

**症状：** `dial tcp i/o timeout` 或 `connection refused`

**解决：**
```powershell
# 1. 检查代理是否运行
netstat -ano | findstr "10808"
# 应该看到 LISTENING 状态

# 2. 测试代理
curl -x socks5://127.0.0.1:10808 https://www.google.com

# 3. 如果失败，检查代理软件配置
# - 确认代理端口是 10808
# - 确认允许局域网连接
# - 尝试重启代理软件
```

### 情况2：API密钥无效

**症状：** `Invalid API-key` 或 `signature invalid`

**解决：**
1. 进入后台 → Trading → API配置
2. 点击"测试连接"按钮
3. 如果失败，重新创建API密钥

**交易所API配置要求：**
```
✅ 开启"合约交易"权限
✅ 开启"读取"权限
❌ 不要开启"提现"权限（安全考虑）
✅ IP白名单添加服务器IP（可选）
```

### 情况3：算力不足

**症状：** 提示"算力不足"

**解决：**
```sql
-- 查询算力余额
SELECT user_id, power, gift_power, (power + gift_power) as total_power
FROM hg_toogo_wallet 
WHERE user_id = 你的用户ID;

-- 如果不足10，临时赠送测试算力（仅测试环境）
UPDATE hg_toogo_wallet 
SET gift_power = gift_power + 100 
WHERE user_id = 你的用户ID;
```

---

## 📊 完整诊断命令

如果以上都不行，执行以下诊断：

```powershell
# 1. 检查数据库修复状态
mysql -u root -p -e "USE hotgo; DESC hg_sys_serve_log;" | findstr "line"
# 应该看到: line | int | YES | | 0 |

# 2. 检查最新错误日志
cd D:\go\src\hotgo_v2\server\logs\logger
Select-String -Path "$(Get-Date -Format 'yyyy-MM-dd').log" -Pattern "ERRO|failed" | Select-Object -Last 10

# 3. 检查定时任务
cd D:\go\src\hotgo_v2\server\logs\cron
Get-Content -Path "$(Get-Date -Format 'yyyy-MM-dd').log" -Tail 10

# 4. 检查机器人状态
mysql -u root -p -e "USE hotgo; SELECT id, robot_name, status, api_config_id FROM hg_trading_robot ORDER BY id DESC LIMIT 5;"
```

---

## 📞 获取支持

将以下信息提供给我：

```powershell
# 收集诊断信息
$date = Get-Date -Format 'yyyy-MM-dd'

echo "=== 系统信息 ===" > D:\go\src\hotgo_v2\debug_info.txt
echo "日期: $date" >> D:\go\src\hotgo_v2\debug_info.txt
echo "" >> D:\go\src\hotgo_v2\debug_info.txt

echo "=== 最新错误日志 ===" >> D:\go\src\hotgo_v2\debug_info.txt
Get-Content "D:\go\src\hotgo_v2\server\logs\logger\$date.log" -Tail 50 | Select-String "ERRO" >> D:\go\src\hotgo_v2\debug_info.txt

echo "" >> D:\go\src\hotgo_v2\debug_info.txt
echo "=== 代理测试 ===" >> D:\go\src\hotgo_v2\debug_info.txt
curl -x socks5://127.0.0.1:10808 https://api.binance.com/api/v3/time >> D:\go\src\hotgo_v2\debug_info.txt 2>&1

notepad D:\go\src\hotgo_v2\debug_info.txt
# 将内容发送给我分析
```

---

## ✅ 修复成功确认

全部完成后，您应该看到：

1. ✅ 数据库修复成功（无 line 字段错误）
2. ✅ 后端启动无错误
3. ✅ 机器人可以正常启动
4. ✅ 状态显示"运行中"
5. ✅ 实时数据更新
6. ✅ 日志正常记录

**恭喜！系统已正常运行！** 🎊

---

**执行时间：** < 3 分钟  
**难度级别：** ⭐⭐ 简单

现在立即执行步骤1修复数据库！

