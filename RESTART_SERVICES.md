# 🔄 重启前后端服务

## ✅ 已完成的修复

1. ✅ 创建前端环境变量配置 `.env.development`
2. ✅ 批量修改Trading和Payment API路径（去掉重复的/admin）
3. ✅ 修复菜单组件路径格式
4. ✅ 添加Payment模块路由注册到后端

---

## 🚀 重启服务步骤

### 1. 停止当前服务

在当前运行的终端中按 `Ctrl + C` 停止服务。

#### Terminal 7 或 5 (前端)
```
按 Ctrl + C 停止前端服务
```

#### Terminal 8 或 4 (后端)
```
按 Ctrl + C 停止后端服务
```

---

### 2. 重新启动后端服务

**Terminal 8 或新终端**:

```powershell
cd D:\go\src\hotgo_v2\server
go run main.go --args "all"
```

等待看到类似以下输出：
```
listening on :8000
admin module started successfully
```

---

### 3. 重新启动前端服务

**Terminal 7 或新终端**:

```powershell
cd D:\go\src\hotgo_v2\web
pnpm dev
```

等待看到类似以下输出：
```
VITE v5.x.x  ready in xxx ms

➜  Local:   http://localhost:8001/
➜  Network: http://192.168.1.2:8001/
```

---

## 🎯 验证修复成功

### 1. 打开浏览器

访问：`http://localhost:8001/`

### 2. 登录系统

- 用户名：`admin`
- 密码：`123456`

### 3. 检查控制台

打开浏览器开发者工具（F12），查看Network标签：

**修复前❌**:
```
GET http://192.168.1.2:8003/admin/admin/trading/api-config/list 404
```

**修复后✅**:
```
GET http://localhost:8000/admin/trading/api-config/list 200
GET http://localhost:8000/admin/payment/balance/view 200
```

### 4. 测试功能

点击左侧菜单：

#### Trading模块
- ✅ 量化交易 → API配置
- ✅ 量化交易 → 代理配置
- ✅ 量化交易 → 机器人管理

#### Payment模块
- ✅ USDT管理 → 我的余额
- ✅ USDT管理 → USDT充值
- ✅ USDT管理 → USDT提现
- ✅ USDT管理 → 提现审核

---

## 📊 修复总结

### 问题1：菜单在新窗口打开 ✅ 已修复
**原因**: 组件路径格式不正确  
**解决**: 修改为HotGo v2.0兼容的格式（`/trading/robot/index`）

### 问题2：API请求404错误 ✅ 已修复
**原因**: API路径重复了`/admin`前缀  
**解决**: 
- 创建`.env.development`配置
- 批量修改API路径去掉`/admin`前缀
- 添加Payment路由注册

### 问题3：Dashboard路径警告 ℹ️ 正常
**原因**: HotGo v2.0默认重定向到dashboard  
**影响**: 不影响功能使用

---

## 🛠️ 如果还有问题

### API仍然404

1. 确认后端服务已启动（检查Terminal 8）
2. 确认监听端口是8000
3. 检查路由注册是否生效：
   ```powershell
   cd D:\go\src\hotgo_v2\server
   go run main.go --args "all"
   ```

### 菜单点击无反应

1. 确认前端服务已重启
2. 强制刷新浏览器（Ctrl + F5）
3. 清除浏览器缓存并重新登录

### 组件路径错误

检查数据库中的component字段：
```sql
SELECT title, name, component FROM hg_admin_menu 
WHERE name LIKE 'trading%' OR name LIKE 'payment%';
```

应该是：
- 顶级菜单: `LAYOUT`
- 子菜单: `/trading/api-config/index`

---

## 📝 修改的文件清单

### 前端文件
1. `web/.env.development` ← **新建**
2. `web/src/api/trading/api-config.ts` ← 修改
3. `web/src/api/trading/proxy-config.ts` ← 修改
4. `web/src/api/trading/robot.ts` ← 修改
5. `web/src/api/trading/order.ts` ← 修改
6. `web/src/api/trading/monitor.ts` ← 修改
7. `web/src/api/payment/deposit.ts` ← 修改
8. `web/src/api/payment/withdraw.ts` ← 修改
9. `web/src/api/payment/balance.ts` ← 修改

### 后端文件
1. `server/internal/router/admin.go` ← 修改（添加Payment导入和路由）

### 数据库
1. `hg_admin_menu` 表 ← 修复component字段

---

**现在请重启前后端服务！** 🚀

重启完成后，系统应该能完全正常工作！✨

