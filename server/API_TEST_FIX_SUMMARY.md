# API测试连接失败修复总结

## 问题诊断

### 发现的错误
1. **网络连接超时**: `context deadline exceeded (Client.Timeout exceeded while awaiting headers)`
2. **无法连接Bitget API**: 服务器无法直接访问 `api.bitget.com:443`
3. **代理配置未生效**: 虽然配置了SOCKS5代理，但Go标准库不支持SOCKS5

## 修复内容

### 1. 添加SOCKS5代理支持
**问题**: Go标准库 `net/http` 只支持HTTP/HTTPS代理，不支持SOCKS5
**修复**: 
- 添加 `golang.org/x/net/proxy` 依赖
- 在 `factory.go` 中添加SOCKS5代理支持
- 根据代理类型（socks5/http）分别处理

**代码位置**: `addons/exchange_bitget/service/factory.go:143-171`

### 2. 增强错误日志
**改进**:
- 记录详细的请求URL、方法、错误信息
- 记录API Key前4位（用于调试，不泄露完整密钥）
- 根据错误类型提供友好的错误提示

**代码位置**: `addons/exchange_bitget/service/exchange.go:562-579`

### 3. 参数验证
**改进**:
- 检查API Key、Secret Key、Passphrase是否为空
- 记录密钥长度（不记录实际内容）

**代码位置**: `addons/exchange_bitget/service/exchange.go:62-75`

### 4. 查询参数排序
**问题**: Bitget API要求查询参数按字母顺序排序
**修复**: 使用 `sort.Strings()` 对参数键进行排序

**代码位置**: `addons/exchange_bitget/service/exchange.go:497-524`

## 使用说明

### 1. 安装依赖
```bash
cd D:\go\src\hotgo_v2\server
go get golang.org/x/net/proxy
```

### 2. 重启服务
重启后端服务以应用修复

### 3. 检查代理配置
确保代理服务器正在运行：
- SOCKS5代理: `127.0.0.1:10808`
- 如果代理未运行，需要启动代理服务或配置正确的代理地址

### 4. 测试连接
1. 打开前端页面：API管理
2. 点击"测试"按钮
3. 查看错误信息（如果仍有错误）

## 常见问题

### Q1: 仍然显示"请求失败"
**A**: 检查：
1. 代理服务器是否正在运行（`127.0.0.1:10808`）
2. 代理地址和端口是否正确
3. 查看后端日志获取详细错误信息

### Q2: 如何查看详细日志？
**A**: 
```powershell
# 查看最新日志
Get-ChildItem storage\logs -Filter "*.log" | Sort-Object LastWriteTime -Descending | Select-Object -First 1 | Get-Content -Tail 100
```

### Q3: 如何测试代理是否可用？
**A**: 
```powershell
# 测试SOCKS5代理连接
Test-NetConnection 127.0.0.1 -Port 10808
```

### Q4: 如果不想使用代理怎么办？
**A**: 
1. 在数据库中禁用代理：
```sql
UPDATE hg_trading_proxy_config SET enabled = 0 WHERE user_id = 1;
```
2. 或者确保网络可以直接访问 `api.bitget.com`

## 下一步

1. **重启后端服务**
2. **测试API连接**
3. **如果仍有问题，查看日志并按照错误信息排查**

## 相关文件

- `addons/exchange_bitget/service/factory.go` - 代理配置和交易所实例创建
- `addons/exchange_bitget/service/exchange.go` - HTTP请求和错误处理
- `internal/logic/trading/api_config.go` - API测试逻辑
- `DIAGNOSE_API_TEST.md` - 详细诊断指南

