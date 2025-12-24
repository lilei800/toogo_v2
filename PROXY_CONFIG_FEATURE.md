# 代理配置功能说明

## 功能概述

在量化管理的系统配置页面中新增了"代理配置"功能，允许用户配置SOCKS5和HTTP代理，用于访问交易所API。

## 实现内容

### 1. 前端页面

#### 文件位置
- `web/src/views/toogo/admin/config/index.vue` - 系统配置主页面
- `web/src/views/toogo/admin/config/components/ProxyConfig.vue` - 代理配置组件

#### 功能特性
- ✅ 代理开关（启用/禁用）
- ✅ 代理类型选择（SOCKS5 / HTTP）
- ✅ 代理地址配置（IP:端口格式）
- ✅ 认证配置（用户名/密码）
- ✅ 连接测试功能
- ✅ 测试结果显示（成功/失败、外网IP、延迟）

### 2. 后端API

#### 已有接口（无需修改）
- `GET /trading/proxyConfig/get` - 获取代理配置
- `POST /trading/proxyConfig/save` - 保存代理配置
- `POST /trading/proxyConfig/test` - 测试代理连接
- `POST /trading/proxyConfig/toggle` - 切换启用状态

#### 后端逻辑
- `server/internal/logic/trading/proxy_config.go` - 代理配置业务逻辑
- `server/internal/controller/admin/trading/proxy_config.go` - 控制器
- `server/api/admin/trading/proxy_config.go` - API定义

## 使用说明

### 1. 访问配置页面
1. 登录后台管理系统
2. 进入"量化管理" -> "系统配置"
3. 点击"代理配置"标签页

### 2. 配置代理
1. **启用代理**：打开"启用代理"开关
2. **选择类型**：选择 SOCKS5 或 HTTP
3. **填写地址**：输入代理服务器地址，格式：`IP:端口`，例如：`127.0.0.1:10808`
4. **配置认证**（如需要）：
   - 打开"需要认证"开关
   - 填写用户名和密码
5. **测试连接**：点击"测试连接"按钮验证配置是否正确
6. **保存配置**：确认无误后点击"保存配置"

### 3. 配置示例

#### SOCKS5代理（无认证）
```
启用代理: ✅
代理类型: SOCKS5
代理地址: 127.0.0.1:10808
需要认证: ❌
```

#### SOCKS5代理（有认证）
```
启用代理: ✅
代理类型: SOCKS5
代理地址: 127.0.0.1:10808
需要认证: ✅
用户名: myuser
密码: mypassword
```

#### HTTP代理
```
启用代理: ✅
代理类型: HTTP
代理地址: 127.0.0.1:7890
需要认证: ❌
```

## 技术细节

### 1. 数据存储
- 代理配置存储在 `hg_trading_proxy_config` 表中
- 密码使用AES加密存储
- **全局配置**：使用 `user_id=0` 表示全局配置，所有用户共享

### 2. 代理使用
- 配置保存后，系统会自动使用代理访问交易所API
- 代理配置在创建交易所实例时自动应用
- 支持SOCKS5和HTTP两种代理类型
- **全局生效**：所有用户的API请求都会使用此全局代理配置

### 3. 测试功能
- 测试连接会尝试通过代理访问外网IP检测服务
- 显示连接结果、外网IP和延迟信息
- 测试结果会保存到数据库

## 注意事项

1. **全局配置**：此代理配置为项目全局配置，所有用户共享，修改后会影响所有用户的API请求
2. **代理服务器**：确保代理服务器正在运行且可访问
3. **地址格式**：代理地址必须是 `IP:端口` 格式，不要包含协议前缀
4. **密码安全**：密码会加密存储，留空则不修改现有密码
5. **测试建议**：建议先测试连接，确认配置正确后再启用
6. **网络环境**：如果服务器可以直接访问交易所API，则无需配置代理

## 相关文件

### 前端
- `web/src/views/toogo/admin/config/index.vue` - 系统配置主页面
- `web/src/views/toogo/admin/config/components/ProxyConfig.vue` - 代理配置组件

### 后端
- `server/internal/logic/trading/proxy_config.go` - 代理配置业务逻辑
- `server/internal/controller/admin/trading/proxy_config.go` - 控制器
- `server/api/admin/trading/proxy_config.go` - API定义
- `server/internal/model/input/trading_proxy_config.go` - 输入模型
- `server/internal/model/entity/trading_proxy_config.go` - 实体模型

## 后续优化建议

1. 支持更多代理类型（如SOCKS4）
2. 添加代理健康检查定时任务
3. 支持多个代理配置（主备切换）
4. 添加代理使用统计和监控

