# 代理配置故障排查指南

## 错误：unexpected protocol version 72

### 错误原因
这个错误通常表示：
1. **代理类型配置错误**：实际是HTTP代理，但配置为SOCKS5
2. **代理地址/端口错误**：代理服务器不在指定端口
3. **代理服务器未运行**：代理服务未启动或不可访问

### 排查步骤

#### 1. 检查代理服务器状态
```powershell
# 检查端口是否在监听
netstat -ano | findstr "10808"
```

#### 2. 测试代理连接
```powershell
# 测试SOCKS5代理（如果配置的是SOCKS5）
# 需要安装curl或使用其他工具
curl --socks5 127.0.0.1:10808 https://api.ipify.org?format=text

# 测试HTTP代理（如果配置的是HTTP）
curl --proxy http://127.0.0.1:10808 https://api.ipify.org?format=text
```

#### 3. 检查代理类型
- 如果您的代理是 **V2Ray/Clash**，通常使用 **SOCKS5**
- 如果您的代理是 **HTTP代理**，使用 **HTTP** 类型
- 如果您的代理是 **Shadowsocks**，使用 **SOCKS5** 类型

#### 4. 常见代理端口
- **V2Ray**: 通常 10808 (SOCKS5) 或 10809 (HTTP)
- **Clash**: 通常 7890 (HTTP) 或 7891 (SOCKS5)
- **Shadowsocks**: 通常 1080 (SOCKS5)

### 解决方案

#### 方案1：修改代理类型
如果实际是HTTP代理，请：
1. 进入"系统配置" -> "代理配置"
2. 将"代理类型"改为 **HTTP**
3. 保存并重新测试

#### 方案2：检查代理地址
1. 确认代理服务器正在运行
2. 确认代理地址和端口正确
3. 如果端口不对，修改为正确的端口

#### 方案3：使用HTTP代理
如果SOCKS5不工作，尝试使用HTTP代理：
1. 将代理类型改为 **HTTP**
2. 如果代理支持HTTP，使用HTTP端口（如7890）
3. 保存并测试

### 配置示例

#### V2Ray (SOCKS5)
```
代理类型: SOCKS5
代理地址: 127.0.0.1:10808
需要认证: 否
```

#### Clash (HTTP)
```
代理类型: HTTP
代理地址: 127.0.0.1:7890
需要认证: 否
```

#### Clash (SOCKS5)
```
代理类型: SOCKS5
代理地址: 127.0.0.1:7891
需要认证: 否
```

### 调试建议

1. **先测试直连**：关闭代理，测试是否能直接访问交易所API
2. **测试代理**：使用命令行工具测试代理是否可用
3. **检查日志**：查看后端日志获取详细错误信息
4. **逐步排查**：先确认代理服务器可用，再配置到系统中

### 相关文件

- `server/internal/logic/trading/proxy_config.go` - 代理配置逻辑
- `server/addons/exchange_bitget/service/factory.go` - 代理应用逻辑

