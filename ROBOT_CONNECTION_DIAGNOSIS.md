# 🚨 机器人连接失败诊断报告

## 📊 问题现状

**机器人ID:** 7 (名称: "12")  
**状态:** 运行中(2)  
**错误:** 
1. ❌ `API error: Apikey 不存在`
2. ❌ `context deadline exceeded (Client.Timeout exceeded while awaiting headers)`

---

## 🔍 根本原因

### ❌ 问题1：代理连接超时（主要问题）

**日志证据：**
```
Get "https://api.bitget.com/api/mix/v1/market/ticker?symbol=BTCUSDT": 
context deadline exceeded (Client.Timeout exceeded while awaiting headers)
```

**影响：**
- 后端无法访问Bitget API
- 前端获取行情超时（10秒+）
- 机器人无法获取持仓信息

**当前配置：**
```yaml
# server/manifest/config/config.yaml
exchange:
  proxy:
    enabled: true
    type: "socks5"
    host: "127.0.0.1"
    port: 10808
```

---

### ⚠️ 问题2：API Key 验证失败（次要问题）

**日志证据：**
```
[TradingEngine] 机器人执行失败: robotId=7, 
err=获取持仓失败: symbol=BTCUSDT: API error: Apikey 不存在
```

**API配置信息：**
```
ID: 3
Platform: bitget
Status: 1 (启用)
Verify Status: 2 (已验证)
User ID: 1
```

**可能原因：**
1. API Key已在Bitget平台被撤销
2. API Key权限不足（未开启合约交易权限）
3. API Key IP白名单限制

---

## ✅ 解决方案（3个方案）

### 🎯 方案1：禁用代理（推荐，如果国内可直连）

如果您的服务器在国内且可以直接访问Bitget API：

**步骤：**

1. **修改配置文件**

```bash
# 编辑配置
D:\go\src\hotgo_v2\server\manifest\config\config.yaml
```

修改为：

```yaml
exchange:
  proxy:
    enabled: false  # 改为 false
    type: "socks5"
    host: "127.0.0.1"
    port: 10808
```

2. **重启后端服务**

在终端73中：
- 按 `Ctrl+C` 停止服务
- 重新运行：`cd D:\go\src\hotgo_v2\server; go run main.go`

3. **验证**

刷新浏览器，查看机器人行情是否正常加载

---

### 🎯 方案2：修复代理配置（如果需要代理）

如果确实需要代理访问Bitget：

**检查代理服务：**

1. **确认代理正在运行**

```powershell
# 检查端口10808是否被监听
netstat -ano | findstr "10808"
```

如果没有输出，说明代理未运行。

2. **常见代理软件端口**

| 软件 | 默认端口 |
|------|---------|
| V2RayN | SOCKS: 10808, HTTP: 10809 |
| Clash | SOCKS: 7891, HTTP: 7890 |
| SSR | 1080 |

3. **修改配置为正确端口**

```yaml
exchange:
  proxy:
    enabled: true
    type: "socks5"
    host: "127.0.0.1"
    port: 7891  # 改为您的实际代理端口
```

4. **重启后端服务**

---

### 🎯 方案3：更新API Key

如果API Key确实失效：

**步骤：**

1. **登录Bitget平台**
   - 进入 API管理
   - 检查API Key状态
   - 确认权限包含：合约交易（Futures）

2. **如果需要，创建新的API Key**
   - 权限：读取 + 交易（合约）
   - IP白名单：添加服务器IP或设为不限制

3. **在系统中更新API配置**
   - 前端：量化交易 → API管理
   - 编辑API配置ID=3
   - 输入新的API Key和Secret Key
   - 点击"验证连接"

4. **重启机器人7**
   - 暂停机器人
   - 重新启动

---

## 🚀 快速修复步骤（推荐执行）

### Step 1: 禁用代理（最快）

```powershell
# 1. 停止后端服务（在终端73中按 Ctrl+C）

# 2. 修改配置
notepad D:\go\src\hotgo_v2\server\manifest\config\config.yaml

# 3. 找到 exchange.proxy.enabled，改为 false

# 4. 保存文件

# 5. 重新启动后端
cd D:\go\src\hotgo_v2\server
go run main.go

# 6. 刷新浏览器
```

### Step 2: 验证修复

1. **查看后端日志**

应该看到：
```
✅ 获取行情成功
✅ 获取持仓成功
```

不应再看到：
```
❌ context deadline exceeded
❌ Apikey 不存在
```

2. **前端验证**

刷新浏览器 → 机器人7 应显示：
- ✅ 实时价格
- ✅ 连接状态：正常
- ✅ 无错误提示

---

## 📋 检查清单

- [ ] 确认是否需要代理访问Bitget
- [ ] 如果不需要，禁用代理（enabled: false）
- [ ] 如果需要，确认代理端口正确且服务运行中
- [ ] 检查Bitget API Key是否有效
- [ ] 确认API Key权限包含合约交易
- [ ] 重启后端服务
- [ ] 刷新前端验证

---

## ⚡ 我建议的操作

**基于您的环境（Windows + 国内网络），我推荐：**

1. **先尝试禁用代理**
   - Bitget在国内通常可以直连
   - 避免代理配置复杂性

2. **如果禁用后仍失败**
   - 说明确实需要代理
   - 检查代理软件是否运行
   - 确认代理端口号

3. **最后检查API Key**
   - 如果前两步都正常但仍报错"Apikey 不存在"
   - 需要在Bitget平台重新配置API Key

---

## 🛠️ 立即执行

**我现在可以帮您：**

1. ✅ 自动禁用代理配置
2. ✅ 重启后端服务
3. ✅ 验证修复结果

**是否允许我执行修复？**

回复 **"是"** 或 **"执行修复"**，我将立即开始修复。

---

**诊断日期：** 2025-11-30  
**后端日志文件：** 终端73.txt  
**相关配置：** `server/manifest/config/config.yaml`


























