# 📋 菜单配置完成指南

## ✅ 菜单已成功导入！

### 📊 菜单结构

#### 💰 量化交易（Trading）

```
量化交易 /trading
├─ API配置 /trading/api-config
│  ├─ 新增
│  ├─ 编辑
│  ├─ 删除
│  ├─ 查看
│  ├─ 测试连接
│  └─ 设为默认
│
├─ 代理配置 /trading/proxy-config
│  ├─ 保存配置
│  └─ 测试连接
│
├─ 机器人管理 /trading/robot
│  ├─ 创建机器人 (隐藏路由)
│  ├─ 机器人详情 (隐藏路由)
│  ├─ 编辑
│  ├─ 删除
│  ├─ 启动
│  ├─ 暂停
│  ├─ 停止
│  └─ 统计
│
└─ 订单管理 /trading/order (隐藏菜单)
   ├─ 查看详情
   ├─ 手动平仓
   └─ 订单统计
```

#### 💵 USDT管理（Payment）

```
USDT管理 /payment
├─ 我的余额 /payment/balance
│  └─ 资金流水
│
├─ USDT充值 /payment/deposit
│  ├─ 创建充值
│  ├─ 查看详情
│  ├─ 检查状态
│  └─ 取消订单
│
├─ USDT提现 /payment/withdraw
│  ├─ 申请提现
│  ├─ 查看详情
│  └─ 取消提现
│
└─ 提现审核 /payment/admin/withdraw-audit
   └─ 审核（管理员权限）
```

---

## 🎯 查看新菜单

### 方式1：刷新浏览器

```
1. 在浏览器中按 Ctrl + F5 强制刷新
2. 重新登录系统
3. 查看左侧菜单栏
```

### 方式2：清除缓存

```
1. 浏览器 F12 打开开发者工具
2. 右键点击刷新按钮
3. 选择 "清空缓存并硬性重新加载"
4. 重新登录
```

### 预期效果

登录后，左侧菜单应该显示：

```
🏠 工作台
   └─ 控制台

💰 量化交易 ← 新增
   ├─ API配置
   ├─ 代理配置
   └─ 机器人管理

💵 USDT管理 ← 新增
   ├─ 我的余额
   ├─ USDT充值
   ├─ USDT提现
   └─ 提现审核

⚙️ 系统管理
   └─ ...
```

---

## 📝 菜单详细说明

### Trading 量化交易

#### API配置
- **路径**: `/trading/api-config`
- **权限**: `/trading/api-config/list`
- **组件**: `view.trading.api-config`
- **功能**: 管理交易所API接口（Binance/OKX/Bitget）

#### 代理配置
- **路径**: `/trading/proxy-config`
- **权限**: `/trading/proxy-config/get`
- **组件**: `view.trading.proxy-config`
- **功能**: 配置SOCKS5代理（开发调试用）

#### 机器人管理
- **路径**: `/trading/robot`
- **权限**: `/trading/robot/list`
- **组件**: `view.trading.robot`
- **功能**: 管理量化交易机器人
- **子页面**:
  - 创建机器人: `/trading/robot/create`
  - 机器人详情: `/trading/robot/detail/:id`

### Payment USDT管理

#### 我的余额
- **路径**: `/payment/balance`
- **权限**: `/payment/balance/view`
- **组件**: `view.payment.balance`
- **功能**: 查看USDT余额和资金流水

#### USDT充值
- **路径**: `/payment/deposit`
- **权限**: `/payment/deposit/list`
- **组件**: `view.payment.deposit`
- **功能**: USDT充值（二维码支付）

#### USDT提现
- **路径**: `/payment/withdraw`
- **权限**: `/payment/withdraw/list`
- **组件**: `view.payment.withdraw`
- **功能**: USDT提现申请

#### 提现审核
- **路径**: `/payment/admin/withdraw-audit`
- **权限**: `/payment/withdraw/audit`
- **组件**: `view.payment.admin.withdraw-audit`
- **功能**: 管理员审核提现申请

---

## 🔧 权限配置

### 给角色分配权限

1. **登录管理后台**
2. **进入 系统管理 → 角色管理**
3. **编辑对应角色**
4. **勾选新增的菜单权限**:
   - ✅ 量化交易
   - ✅ USDT管理
5. **保存**

### 默认权限

- **超级管理员**: 自动拥有所有权限
- **普通用户**: 需要手动分配

---

## 🎨 菜单图标

### Trading图标
```
WalletOutlined (钱包图标)
```

### Payment图标
```
DollarOutlined (美元符号图标)
```

如需修改图标，可以在数据库中修改 `hg_admin_menu` 表的 `icon` 字段。

可用图标参考：
- https://www.naiveui.com/zh-CN/os-theme/components/icon
- https://icons.ant.design/

---

## 🐛 故障排查

### Q1: 菜单不显示

**原因**: 缓存未清除

**解决**:
```
1. Ctrl + F5 强制刷新浏览器
2. 退出登录并重新登录
3. 清除浏览器缓存
```

### Q2: 没有权限访问

**原因**: 角色权限未配置

**解决**:
```
1. 进入 系统管理 → 角色管理
2. 编辑当前用户的角色
3. 勾选 Trading 和 Payment 相关权限
4. 保存后重新登录
```

### Q3: 点击菜单报404

**原因**: 路由组件路径不匹配

**解决**:
```
1. 检查前端路由配置
2. 确认组件文件存在于 web/src/views/ 目录
3. 检查路由路径是否正确
```

### Q4: 菜单图标不显示

**原因**: 图标名称不正确

**解决**:
```sql
-- 修改图标
UPDATE hg_admin_menu 
SET icon = 'WalletOutlined' 
WHERE name = '量化交易';

UPDATE hg_admin_menu 
SET icon = 'DollarOutlined' 
WHERE name = 'USDT管理';
```

---

## 📊 菜单统计

### 总计
- **顶级菜单**: 2个（Trading + Payment）
- **二级菜单**: 8个
- **三级菜单（操作）**: 24个
- **总计**: 34个菜单项

### 菜单类型
1. **目录菜单** (type=1): 2个
2. **页面菜单** (type=2): 8个
3. **按钮权限** (type=3): 24个

---

## 🎯 下一步

### 1. 验证菜单显示

```
✅ 刷新浏览器
✅ 重新登录
✅ 查看左侧菜单
```

### 2. 测试功能

**Trading测试**:
1. 点击 "量化交易" → "API配置"
2. 添加一个测试API配置
3. 点击 "机器人管理"
4. 创建一个测试机器人

**Payment测试**:
1. 点击 "USDT管理" → "我的余额"
2. 查看余额信息
3. 尝试 "USDT充值"
4. 查看充值二维码

### 3. 配置权限

```
系统管理 → 角色管理 → 编辑角色 → 勾选新菜单
```

---

## 📚 相关文档

- `COMPLETE_MIGRATION_SUMMARY.md` - 完整迁移总结
- `SYSTEM_STARTED.md` - 系统启动指南
- `BACKEND_MIGRATION_COMPLETED.md` - 后端迁移详情
- `TOOGO_MIGRATION_COMPLETED.md` - 前端迁移详情

---

## ✅ 验收标准

- [ ] 菜单在左侧导航栏正常显示
- [ ] 点击菜单能正常跳转到对应页面
- [ ] 页面功能正常工作
- [ ] 权限控制正常
- [ ] 图标正常显示

---

**菜单配置完成！刷新浏览器即可查看！** 🎊

