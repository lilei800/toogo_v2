# 📋 创建所有页面文件 - 批量创建指南

## 由于页面组件代码量较大，我将分批创建

### 方式1：直接使用命令行批量创建（最快）

运行以下PowerShell脚本创建所有目录结构：

```powershell
# 创建Trading目录
New-Item -ItemType Directory -Force -Path "D:\go\src\hotgo_v2\web\src\views\trading\api-config"
New-Item -ItemType Directory -Force -Path "D:\go\src\hotgo_v2\web\src\views\trading\proxy-config"
New-Item -ItemType Directory -Force -Path "D:\go\src\hotgo_v2\web\src\views\trading\robot"

# 创建Payment目录
New-Item -ItemType Directory -Force -Path "D:\go\src\hotgo_v2\web\src\views\payment\deposit"
New-Item -ItemType Directory -Force -Path "D:\go\src\hotgo_v2\web\src\views\payment\withdraw"
New-Item -ItemType Directory -Force -Path "D:\go\src\hotgo_v2\web\src\views\payment\balance"
New-Item -ItemType Directory -Force -Path "D:\go\src\hotgo_v2\web\src\views\payment\admin"
```

### 方式2：让我逐个创建每个页面（推荐）

我将创建每个页面，保证质量和功能完整性。

---

## 📄 需要创建的文件清单

### Trading页面（5个）

1. **api-config/index.vue** - API配置管理
   - 功能：列表、新增、编辑、删除、测试连接、设为默认
   - 预计代码：~400行

2. **proxy-config/index.vue** - 代理配置  
   - 功能：配置SOCKS5代理、测试连接、启用/禁用
   - 预计代码：~250行

3. **robot/index.vue** - 机器人列表
   - 功能：列表、创建、启动、暂停、停止、统计
   - 预计代码：~500行

4. **robot/create.vue** - 创建机器人向导
   - 功能：5步向导式创建机器人
   - 预计代码：~600行

5. **robot/detail.vue** - 机器人详情
   - 功能：实时监控、持仓列表、平仓日志、统计图表
   - 预计代码：~700行

### Payment页面（4个）

6. **deposit/index.vue** - USDT充值
   - 功能：创建订单、二维码、订单列表
   - 预计代码：~400行

7. **withdraw/index.vue** - USDT提现
   - 功能：申请提现、地址验证、订单列表
   - 预计代码：~400行

8. **balance/index.vue** - 余额查看
   - 功能：余额展示、流水记录、统计图表
   - 预计代码：~350行

9. **admin/withdraw-audit.vue** - 提现审核
   - 功能：审核列表、批量审核、统计
   - 预计代码：~450行

### 路由配置（2个）

10. **router/routes/modules/trading.ts**
11. **router/routes/modules/payment.ts**

---

## 🎯 创建策略

### 优先级1：核心功能页面

1. robot/index.vue
2. robot/create.vue
3. deposit/index.vue
4. withdraw/index.vue

### 优先级2：配置页面

5. api-config/index.vue
6. proxy-config/index.vue

### 优先级3：详情和管理页面

7. robot/detail.vue
8. balance/index.vue
9. admin/withdraw-audit.vue

---

## 准备就绪！

**请确认：是否立即开始创建所有9个页面组件？**

我将逐个创建，确保每个页面都是完整可用的Vue3组件。

预计总时间：30-45分钟完成所有页面创建。

