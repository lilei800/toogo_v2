# 机器人连接失败问题修复

## 问题分析

### 原因
`engine.go` 中的 `getOrCreateRunner` 方法直接调用 `exchange.NewExchange` 创建交易所实例，而不是使用 `ExchangeManager.GetExchangeFromConfig`，导致：

1. **没有连接复用**：每次创建新连接，而不是复用已有连接
2. **代理配置不一致**：虽然获取了代理配置，但可能与应用不一致
3. **资源浪费**：重复创建连接，增加服务器负担

### 修复方案

**修改前**：
```go
// 直接创建交易所实例
ex, err := exchange.NewExchange(&exchange.Config{
    Platform:   apiConfig.Platform,
    ApiKey:     apiConfig.ApiKey,
    SecretKey:  apiConfig.SecretKey,
    Passphrase: apiConfig.Passphrase,
    IsTestnet:  false,
    Proxy:      proxyConfig,
})
```

**修改后**：
```go
// 使用 ExchangeManager 获取交易所实例（复用连接，统一代理配置）
ex, err := GetExchangeManager().GetExchangeFromConfig(ctx, apiConfig)
```

## 修复内容

### 1. 统一使用 ExchangeManager

**文件**：`internal/logic/toogo/engine.go`

**修改**：
- 移除直接创建交易所实例的代码
- 使用 `GetExchangeManager().GetExchangeFromConfig()` 获取实例
- 确保连接复用和代理配置统一

### 2. 连接复用机制

**优势**：
- 相同 `apiConfigId` 的机器人共享同一个交易所连接
- 减少连接数，提高性能
- 统一管理连接生命周期

### 3. 全局代理配置

**确保**：
- 所有机器人使用相同的全局代理配置
- 代理配置从数据库读取（`user_id=0, tenant_id=0`）
- 配置变更后自动生效

## 连接失败排查步骤

### 1. 检查 API 配置

```sql
-- 检查机器人的 API 配置是否存在
SELECT r.id, r.robot_name, r.api_config_id, a.platform, a.api_key
FROM hg_trading_robot r
LEFT JOIN hg_trading_api_config a ON r.api_config_id = a.id
WHERE r.status = 2;
```

### 2. 检查代理配置

```sql
-- 检查全局代理配置
SELECT * FROM hg_trading_proxy_config 
WHERE user_id = 0 AND tenant_id = 0 AND enabled = 1;
```

### 3. 检查定时任务日志

查看定时任务执行日志，确认是否有错误：
- 日志位置：`storage/log/cron/`
- 查看 `toogo_robot_engine` 任务的执行日志

### 4. 测试 API 连接

在后台管理界面：
1. 进入"API管理"
2. 找到机器人使用的 API 配置
3. 点击"测试"按钮
4. 确认连接成功

### 5. 检查机器人状态

```sql
-- 检查运行中的机器人
SELECT id, robot_name, status, api_config_id, symbol
FROM hg_trading_robot
WHERE status = 2;
```

## 常见连接失败原因

### 1. API 配置错误
- **问题**：API Key、Secret Key 或 Passphrase 错误
- **解决**：重新配置并测试 API 连接

### 2. 代理配置错误
- **问题**：代理服务器未启动或配置错误
- **解决**：检查代理服务器状态，测试代理连接

### 3. 网络问题
- **问题**：服务器无法访问交易所 API
- **解决**：检查网络连接，确认防火墙设置

### 4. 定时任务未运行
- **问题**：定时任务服务未启动
- **解决**：启动定时任务服务 `go run main.go cron`

### 5. API 配置不存在
- **问题**：机器人关联的 API 配置已被删除
- **解决**：重新关联有效的 API 配置

## 验证修复

### 1. 重启服务

```bash
# 重启主服务
go run main.go server

# 重启定时任务服务
go run main.go cron
```

### 2. 检查日志

查看定时任务日志，确认：
- 机器人执行成功
- 没有连接错误
- 成功获取行情和持仓

### 3. 前端验证

在前端"我的机器人"页面：
- 查看机器人连接状态
- 确认实时数据正常加载
- 检查持仓和订单数据

## 后续优化建议

1. **连接池管理**：考虑实现连接池，限制最大连接数
2. **错误重试**：添加连接失败后的自动重试机制
3. **健康检查**：定期检查连接健康状态，自动重建失效连接
4. **监控告警**：添加连接失败的监控和告警机制

