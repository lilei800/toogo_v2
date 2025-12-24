# API连接失败问题修复指南

## 问题现象
点击"测试"按钮后，弹窗提示：**"连接失败 请求失败"**

## 根本原因
网络测试显示：**无法连接到 `api.bitget.com`**（TCP连接失败）

```
警告: TCP connect to (104.244.43.248 : 443) failed
警告: Ping to 104.244.43.248 failed with status: TimedOut
```

这说明服务器无法直接访问Bitget API服务器，可能需要：
1. 配置代理服务器
2. 检查防火墙设置
3. 检查网络连接

## 已修复的问题

### 1. 查询参数排序
- **问题**: Bitget API要求查询参数按字母顺序排序
- **修复**: 使用 `sort.Strings()` 对参数键进行排序
- **位置**: `addons/exchange_bitget/service/exchange.go:request()`

### 2. 错误信息优化
- **问题**: 错误信息不够详细，难以定位问题
- **修复**: 
  - 添加详细的错误日志
  - 根据错误类型提供友好的错误提示
  - 区分超时、DNS、代理等不同错误
- **位置**: `exchange.go:request()`, `factory.go:CreateExchange()`

### 3. 超时时间增加
- **问题**: 默认10秒超时可能不够
- **修复**: 增加到30秒
- **位置**: `exchange.go:NewBitgetExchange()`

### 4. API密钥验证
- **问题**: 解密后未验证是否为空
- **修复**: 添加空值检查
- **位置**: `factory.go:CreateExchange()`

## 解决方案

### 方案1: 配置代理服务器（推荐）

如果您的服务器无法直接访问Bitget API，需要配置代理：

#### 方式1: 通过配置文件（全局代理）

编辑 `manifest/config/config.yaml`，添加：

```yaml
exchange:
  proxy:
    enabled: true                    # 启用代理
    type: "socks5"                   # 代理类型: socks5 或 http
    host: "127.0.0.1"                # 代理地址
    port: 10808                      # 代理端口
    # username: ""                   # 代理用户名(可选)
    # password: ""                   # 代理密码(可选)
```

#### 方式2: 通过数据库配置（用户级代理）

在数据库中配置用户代理：

```sql
-- 查看现有代理配置
SELECT * FROM hg_trading_proxy_config WHERE user_id = ?;

-- 添加代理配置
INSERT INTO hg_trading_proxy_config 
  (user_id, tenant_id, enabled, proxy_type, proxy_address, auth_enabled, username, password)
VALUES 
  (?, ?, 1, 'socks5', '127.0.0.1:10808', 0, '', '');
```

### 方案2: 检查网络连接

1. **测试DNS解析**:
   ```powershell
   nslookup api.bitget.com
   ```

2. **测试端口连接**:
   ```powershell
   Test-NetConnection -ComputerName api.bitget.com -Port 443
   ```

3. **检查防火墙**:
   - 确保443端口未被阻止
   - 检查Windows防火墙设置
   - 检查企业防火墙/网关设置

### 方案3: 使用VPN或更换网络

如果服务器在受限网络环境中，考虑：
- 使用VPN连接
- 更换服务器网络
- 使用云服务器（通常网络更稳定）

## 调试步骤

### 1. 查看详细日志

重启服务后，查看日志文件：
```powershell
Get-Content "D:\go\src\hotgo_v2\server\storage\logs\*.log" -Tail 50 | Select-String -Pattern "Bitget|ExchangeFactory|API|error|Error"
```

关键日志标识：
- `[ExchangeFactory]` - 工厂创建过程
- `[Bitget]` - API请求过程
- `请求失败` - 网络请求错误

### 2. 检查API配置

```sql
-- 查看API配置
SELECT id, user_id, platform, api_name, status, 
       verify_status, verify_message, last_verify_time
FROM hg_trading_api_config 
WHERE id = ? AND deleted_at IS NULL;

-- 检查密钥是否已加密（前几位应该是加密后的字符串）
SELECT id, LEFT(api_key, 20) as api_key_preview
FROM hg_trading_api_config 
WHERE id = ?;
```

### 3. 测试代理连接

如果有代理，先测试代理是否可用：
```powershell
# 测试SOCKS5代理
curl --socks5 127.0.0.1:10808 https://api.bitget.com/api/mix/v1/market/ticker?symbol=BTCUSDT

# 测试HTTP代理
curl --proxy http://127.0.0.1:7890 https://api.bitget.com/api/mix/v1/market/ticker?symbol=BTCUSDT
```

## 常见错误信息

### 1. "请求超时，请检查网络连接或代理设置"
- **原因**: 网络连接超时
- **解决**: 配置代理或检查网络

### 2. "DNS解析失败，无法连接到Bitget服务器"
- **原因**: DNS无法解析域名
- **解决**: 检查DNS设置，或使用IP直连（不推荐）

### 3. "连接被拒绝，请检查代理设置或网络连接"
- **原因**: 端口被阻止或代理配置错误
- **解决**: 检查防火墙和代理配置

### 4. "解密API Key失败"
- **原因**: 加密密钥不匹配或数据损坏
- **解决**: 重新添加API配置

### 5. "API Key解密后为空"
- **原因**: 解密成功但值为空
- **解决**: 检查API配置，确保密钥正确

## 验证修复

修复后，重启服务并测试：

1. **重启后端服务**:
   ```powershell
   cd D:\go\src\hotgo_v2\server
   go run main.go
   ```

2. **在前端测试连接**:
   - 进入"API管理"页面
   - 点击"测试"按钮
   - 查看错误提示（现在应该更详细）

3. **查看日志**:
   - 如果配置了代理，应该看到 `[ExchangeFactory] 已配置代理`
   - 如果请求失败，会看到详细的错误信息

## 下一步

1. ✅ 已修复代码中的问题
2. ⏳ 需要配置代理（如果网络受限）
3. ⏳ 重启服务并测试
4. ⏳ 查看详细日志定位具体问题

如果配置代理后仍然失败，请提供日志中的详细错误信息，以便进一步诊断。

