# WebSocket 连接失败排查指南

## 问题现象

前端控制台显示：
```
WebSocket connection to 'ws://192.168.1.2:8000/socket?authorization=...' failed
[WebSocket] 发生错误
[WebSocket] 已关闭
```

## 可能原因

### 1. WebSocket 服务器未启动

**检查**：确认主服务（HTTP服务）是否运行

```bash
# 检查服务是否运行
# Windows: 查看任务管理器或使用 netstat
netstat -ano | findstr :8000

# Linux: 
ps aux | grep "main.go http"
netstat -tlnp | grep :8000
```

**解决**：启动主服务
```bash
cd D:\go\src\hotgo_v2\server
go run main.go http
```

### 2. WebSocket 路由路径问题

**配置位置**：`manifest/config/config.yaml`

```yaml
router:
  websocket:
    prefix: "/socket"  # WebSocket路径前缀
```

**前端连接地址**：`ws://192.168.1.2:8000/socket`

**注意**：
- WebSocket 路由注册在 `group.GET("/", websocket.WsPage)`
- 完整路径应该是 `/socket/`（带末尾斜杠）
- 但前端连接时使用 `/socket` 也可以（GoFrame会自动处理）

### 3. JWT Token 过期或无效

**检查**：
- 查看浏览器控制台中的完整错误信息
- 检查 token 是否过期
- 尝试重新登录获取新 token

**解决**：
- 重新登录获取新的 token
- 检查 token 的有效期设置

### 4. 认证中间件问题

**检查**：WebSocket 连接需要经过认证中间件 `WebSocketAuth`

**可能问题**：
- Token 格式不正确
- Token 验证失败
- 用户权限不足

### 5. 网络或防火墙问题

**检查**：
- 服务器防火墙是否允许 WebSocket 连接
- 网络是否可达 `192.168.1.2:8000`
- 是否有代理或 NAT 配置问题

**测试**：
```bash
# 测试端口是否开放
telnet 192.168.1.2 8000

# 或使用 curl 测试 WebSocket
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: test" \
  http://192.168.1.2:8000/socket
```

### 6. 配置问题

**检查 WebSocket 地址配置**：

1. **后端配置**：`manifest/config/config.yaml`
   ```yaml
   router:
     websocket:
       prefix: "/socket"
   ```

2. **前端配置**：检查 `useUserStore.config.wsAddr` 的值
   - 如果是本地IP访问，会自动生成：`ws://IP:端口/socket`
   - 否则使用系统配置中的 `basicWsAddr`

3. **系统配置**：后台 → 系统配置 → 基础配置 → WebSocket地址
   - 如果配置了 `basicWsAddr`，会优先使用配置的地址
   - 如果未配置，会使用本地IP模式

## 排查步骤

### 步骤1：确认服务运行状态

```bash
# 检查主服务是否运行
ps aux | grep "main.go http"  # Linux
tasklist | findstr "main.go"   # Windows

# 检查端口是否监听
netstat -tlnp | grep :8000     # Linux
netstat -ano | findstr :8000   # Windows
```

### 步骤2：查看服务日志

查看主服务日志，确认：
- WebSocket 路由是否注册成功
- 是否有连接错误
- 是否有认证错误

日志位置：`storage/log/`

### 步骤3：测试 WebSocket 连接

**使用浏览器控制台测试**：
```javascript
const ws = new WebSocket('ws://192.168.1.2:8000/socket?authorization=YOUR_TOKEN');
ws.onopen = () => console.log('Connected');
ws.onerror = (e) => console.error('Error:', e);
ws.onclose = (e) => console.log('Closed:', e.code, e.reason);
```

### 步骤4：检查认证

1. 确认 token 有效：
   - 检查 token 是否过期
   - 检查 token 格式是否正确
   - 尝试重新登录获取新 token

2. 查看认证中间件日志：
   - 检查是否有认证失败的错误
   - 检查用户权限是否足够

### 步骤5：检查配置

1. **检查 WebSocket 路径配置**：
   ```yaml
   # manifest/config/config.yaml
   router:
     websocket:
       prefix: "/socket"
   ```

2. **检查系统配置**：
   - 后台 → 系统配置 → 基础配置
   - 查看 `basicWsAddr` 配置
   - 如果配置了，确保地址正确

## 常见解决方案

### 方案1：使用本地IP模式（推荐）

如果是在本地开发环境，系统会自动检测本地IP并使用：
```
ws://192.168.1.2:8000/socket
```

**确保**：
- 主服务正常运行
- 端口8000可访问
- 防火墙允许连接

### 方案2：配置系统WebSocket地址

1. 登录后台管理系统
2. 进入：系统配置 → 基础配置
3. 设置 `basicWsAddr` 为：`ws://192.168.1.2:8000/socket`
4. 保存配置
5. 刷新前端页面

### 方案3：检查并修复认证

1. **重新登录**：获取新的 token
2. **检查 token 有效期**：确认 token 未过期
3. **查看认证日志**：检查是否有认证错误

### 方案4：检查网络和防火墙

1. **检查防火墙**：
   ```bash
   # Windows: 检查防火墙规则
   # Linux: 检查 iptables 或 firewalld
   ```

2. **检查端口**：
   ```bash
   # 确认8000端口是否开放
   netstat -tlnp | grep :8000
   ```

3. **测试连接**：
   ```bash
   # 使用 telnet 或 curl 测试
   telnet 192.168.1.2 8000
   ```

## 调试技巧

### 1. 启用详细日志

在 `manifest/config/config.yaml` 中设置日志级别：
```yaml
logger:
  level: "debug"  # 设置为 debug 查看详细日志
```

### 2. 查看 WebSocket 连接日志

查看 `storage/log/` 目录下的日志文件，查找：
- WebSocket 连接请求
- 认证错误
- 路由匹配错误

### 3. 使用浏览器开发者工具

1. 打开浏览器开发者工具（F12）
2. 切换到 Network 标签
3. 筛选 WS（WebSocket）连接
4. 查看连接详情和错误信息

## 验证修复

修复后，检查以下内容：

1. ✅ 主服务正常运行
2. ✅ WebSocket 路由已注册
3. ✅ Token 有效且未过期
4. ✅ 网络连接正常
5. ✅ 防火墙允许连接
6. ✅ 前端成功连接 WebSocket

## 注意事项

1. **WebSocket 连接是自动的**：前端会在登录后自动尝试连接
2. **连接失败会自动重试**：前端有重连机制，每10秒重试一次
3. **不影响主要功能**：WebSocket 主要用于实时通知，连接失败不影响主要功能
4. **生产环境建议使用 WSS**：生产环境应使用 `wss://`（WebSocket Secure）

## 相关文件

- WebSocket 路由：`internal/router/websocket.go`
- WebSocket 处理器：`internal/controller/websocket/`
- WebSocket Hub：`internal/library/websocket/hub.go`
- 前端 WebSocket：`web/src/utils/websocket/index.ts`
- 配置文件：`manifest/config/config.yaml`





















