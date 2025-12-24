# API测试连接失败诊断文档

## 项目结构

### 前端
- **文件**: `web/src/views/toogo/api/index.vue`
- **测试接口**: `/trading/apiConfig/test`
- **请求方式**: POST
- **请求参数**: `{ id: row.id }`

### 后端流程

1. **Controller层** (`internal/controller/admin/trading/api_config.go`)
   - 接收请求，调用Logic层

2. **Logic层** (`internal/logic/trading/api_config.go`)
   - `Test` 方法执行测试流程：
     - 验证用户登录状态
     - 查询API配置（需要匹配userId）
     - 清除缓存
     - 获取交易所实例
     - 调用 `TestConnection` 测试连接

3. **ExchangeManager** (`internal/logic/trading/exchange_manager.go`)
   - `GetExchange` 方法通过工厂创建交易所实例

4. **ExchangeFactory** (`addons/exchange_bitget/service/factory.go`)
   - `CreateExchange` 方法：
     - 查询API配置（需要匹配userId和tenantId）
     - 解密API密钥
     - 获取代理配置
     - 创建交易所实例

5. **BitgetExchange** (`addons/exchange_bitget/service/exchange.go`)
   - `TestConnection` 方法：
     - 调用 `/api/mix/v1/account/accounts` 接口
     - 解析响应，返回USDT余额

## 可能失败的原因

### 1. 用户权限问题
- **问题**: API配置查询时要求 `userId` 匹配
- **检查**: 确认当前登录用户ID与API配置的userId一致
- **位置**: `api_config.go:306`, `factory.go:36`

### 2. API配置不存在或已删除
- **问题**: 配置不存在或已被软删除
- **检查**: 
  ```sql
  SELECT id, user_id, tenant_id, platform, api_name, status, deleted_at 
  FROM hg_trading_api_config 
  WHERE id = ? AND deleted_at IS NULL;
  ```

### 3. API密钥解密失败
- **问题**: 加密密钥不匹配或数据损坏
- **检查**: 
  - 确认 `encrypt.AesDecrypt` 使用的密钥与加密时一致
  - 检查数据库中存储的加密数据格式
- **位置**: `factory.go:62-78`

### 4. 交易所平台不支持
- **问题**: 目前只支持 `bitget` 平台
- **检查**: 确认 `platform` 字段值为 `bitget`
- **位置**: `factory.go:88-99`

### 5. 网络连接问题
- **问题**: 无法连接到Bitget API服务器
- **检查**: 
  - 网络连接是否正常
  - 是否需要代理
  - API服务器是否可访问

### 6. API密钥无效或权限不足
- **问题**: API Key/Secret Key/Passphrase 错误
- **检查**: 
  - 确认API密钥是否正确
  - 确认API权限是否包含"读取账户信息"
  - Bitget API需要Passphrase

### 7. 签名计算错误
- **问题**: API请求签名不正确
- **检查**: 
  - 时间戳格式
  - 签名算法实现
  - Header设置
- **位置**: `exchange.go:497-555`

### 8. 代理配置问题
- **问题**: 代理配置错误导致请求失败
- **检查**: 
  - 代理地址、端口、认证信息
  - 代理类型（http/socks5）
- **位置**: `factory.go:102-146`

### 9. 响应解析错误
- **问题**: API返回格式不符合预期
- **检查**: 
  - 响应JSON格式
  - 错误码处理
- **位置**: `exchange.go:76-93`

## 调试步骤

### 1. 检查数据库记录
```sql
-- 查看API配置
SELECT id, user_id, tenant_id, platform, api_name, status, 
       verify_status, verify_message, last_verify_time, deleted_at
FROM hg_trading_api_config 
WHERE id = ?;

-- 查看加密后的密钥（前几位）
SELECT id, LEFT(api_key, 20) as api_key_preview, 
       LEFT(secret_key, 20) as secret_key_preview,
       LEFT(passphrase, 20) as passphrase_preview
FROM hg_trading_api_config 
WHERE id = ?;
```

### 2. 检查后端日志
- 查看 `storage/logs/` 目录下的日志文件
- 搜索关键词: `API测试`, `TestConnection`, `获取交易所实例失败`, `API连接失败`

### 3. 添加详细日志
在关键位置添加日志输出：
- `factory.go:CreateExchange` - 记录解密过程
- `exchange.go:TestConnection` - 记录请求和响应
- `exchange.go:request` - 记录签名和HTTP请求详情

### 4. 测试API密钥
使用curl直接测试Bitget API：
```bash
# 需要先计算签名，这里只是示例
curl -X GET "https://api.bitget.com/api/mix/v1/account/accounts?productType=umcbl" \
  -H "ACCESS-KEY: your_api_key" \
  -H "ACCESS-SIGN: your_signature" \
  -H "ACCESS-TIMESTAMP: timestamp" \
  -H "ACCESS-PASSPHRASE: your_passphrase" \
  -H "Content-Type: application/json"
```

## 常见错误信息

1. **"用户未登录"**
   - 原因: `memberId <= 0`
   - 解决: 确认用户已登录

2. **"配置不存在"**
   - 原因: 查询不到配置或userId不匹配
   - 解决: 检查配置ID和用户ID

3. **"获取交易所实例失败: ..."**
   - 原因: 创建交易所实例时出错
   - 可能: 解密失败、平台不支持、配置禁用

4. **"API连接失败: ..."**
   - 原因: 调用交易所API时出错
   - 可能: 网络问题、签名错误、API密钥无效

5. **"解密API Key失败"**
   - 原因: 加密密钥不匹配或数据损坏
   - 解决: 重新添加API配置

6. **"不支持的交易所平台: xxx"**
   - 原因: platform不是"bitget"
   - 解决: 使用Bitget平台

## 修复建议

1. **增强错误日志**: 在关键位置添加详细的错误日志，包括请求参数、响应内容等

2. **改进错误提示**: 将技术错误信息转换为用户友好的提示

3. **添加重试机制**: 对于网络错误，可以添加重试逻辑

4. **验证API密钥格式**: 在保存前验证API密钥的基本格式

5. **添加测试模式**: 支持测试网环境，避免使用真实API密钥测试

