# API测试连接失败诊断指南

## 当前错误
**错误信息**: "连接失败 请求失败"

## 诊断步骤

### 1. 检查后端日志
查看最新的日志文件，搜索以下关键词：
- `[Bitget]`
- `[ExchangeFactory]`
- `请求失败`
- `解密`

**命令**:
```powershell
# 在 server 目录下执行
Get-ChildItem storage\logs -Filter "*.log" | Sort-Object LastWriteTime -Descending | Select-Object -First 1 | Get-Content -Tail 100
```

### 2. 检查数据库中的API配置
```sql
-- 查看API配置详情
SELECT 
    id, 
    user_id, 
    tenant_id, 
    platform, 
    api_name, 
    status,
    LEFT(api_key, 20) as api_key_preview,
    LEFT(secret_key, 20) as secret_key_preview,
    CASE 
        WHEN passphrase IS NULL OR passphrase = '' THEN '空'
        ELSE LEFT(passphrase, 20)
    END as passphrase_preview,
    verify_status,
    verify_message,
    last_verify_time,
    deleted_at
FROM hg_trading_api_config 
WHERE deleted_at IS NULL
ORDER BY id DESC;
```

### 3. 常见问题排查

#### 问题1: Passphrase为空
**症状**: Bitget API需要Passphrase，如果为空会导致签名失败
**检查**: 
```sql
SELECT id, platform, api_name, 
       CASE WHEN passphrase IS NULL OR passphrase = '' THEN '空' ELSE '有值' END as passphrase_status
FROM hg_trading_api_config 
WHERE platform = 'bitget' AND deleted_at IS NULL;
```
**解决**: 确保Bitget API配置中包含Passphrase

#### 问题2: 网络连接问题
**症状**: 无法连接到 api.bitget.com
**检查**: 
```powershell
# 测试网络连接
Test-NetConnection api.bitget.com -Port 443
```
**解决**: 
- 检查防火墙设置
- 检查是否需要配置代理
- 检查DNS解析是否正常

#### 问题3: 代理配置问题
**症状**: 配置了代理但代理不可用
**检查**: 
```sql
SELECT id, user_id, enabled, proxy_type, proxy_address, auth_enabled
FROM hg_trading_proxy_config
WHERE enabled = 1;
```
**解决**: 
- 检查代理地址和端口是否正确
- 检查代理认证信息
- 临时禁用代理测试直连

#### 问题4: API密钥解密失败
**症状**: 解密API密钥时出错
**检查**: 查看日志中的 `解密API Key失败` 或 `解密Secret Key失败`
**解决**: 
- 重新添加API配置
- 检查加密密钥配置

#### 问题5: API密钥无效
**症状**: API密钥格式错误或已过期
**检查**: 
- 确认API Key格式正确（通常是32位字符串）
- 确认Secret Key格式正确
- 确认Passphrase正确（Bitget创建API时设置的）

### 4. 手动测试Bitget API连接

使用curl测试（需要替换为实际的API密钥）:
```bash
# 注意：这需要手动计算签名，仅用于参考
curl -X GET "https://api.bitget.com/api/mix/v1/account/accounts?productType=umcbl" \
  -H "ACCESS-KEY: your_api_key" \
  -H "ACCESS-SIGN: calculated_signature" \
  -H "ACCESS-TIMESTAMP: timestamp_in_milliseconds" \
  -H "ACCESS-PASSPHRASE: your_passphrase" \
  -H "Content-Type: application/json" \
  -H "locale: zh-CN"
```

### 5. 检查配置文件

检查 `manifest/config/config.yaml` 或 `manifest/config/exchange.yaml`:
```yaml
exchange:
  debug: true  # 启用调试模式，查看详细日志
  proxy:
    enabled: false  # 如果网络需要代理，设置为true
    type: http
    host: 127.0.0.1
    port: 7890
```

## 修复后的改进

1. **增强的错误日志**: 
   - 记录详细的请求URL、方法、错误信息
   - 记录API Key的前4位（用于调试）
   - 记录网络错误的具体类型

2. **更友好的错误提示**:
   - 超时错误: "请求超时，请检查网络连接或代理设置"
   - DNS错误: "DNS解析失败，无法连接到Bitget服务器"
   - 连接拒绝: "连接被拒绝，请检查代理设置或网络连接"
   - 代理错误: "代理连接失败"

3. **参数验证**:
   - 检查API Key、Secret Key、Passphrase是否为空
   - 记录密钥长度（不记录实际内容）

## 下一步操作

1. **重启后端服务**以应用修复
2. **查看日志**获取详细的错误信息
3. **根据日志中的错误信息**按照上述步骤排查
4. **如果问题仍然存在**，请提供日志中的具体错误信息

