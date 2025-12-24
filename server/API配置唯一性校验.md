# API配置唯一性校验

## 📋 功能说明

**规则**: 每个API配置只能绑定一个未删除的机器人

**目的**: 
- 避免多个机器人使用同一个API密钥导致冲突
- 保证每个交易账户的交易行为独立可控
- 防止订单管理混乱

---

## ✅ 实现逻辑

### 位置
`internal/logic/trading/robot.go` - `Create()` 函数

### 校验时机
在创建机器人时，验证API配置后立即校验

### 校验代码

```go
// 【新增】校验：每个API配置只能绑定一个未删除的机器人
var existingRobot *entity.TradingRobot
err = dao.TradingRobot.Ctx(ctx).
    Where(dao.TradingRobot.Columns().ApiConfigId, in.ApiConfigId).
    WhereNull(dao.TradingRobot.Columns().DeletedAt).
    Scan(&existingRobot)

if err != nil {
    return 0, gerror.Wrap(err, "检查API配置绑定失败")
}
if existingRobot != nil {
    return 0, gerror.Newf("该API配置已绑定机器人【%s】，每个API配置只能绑定一个机器人", existingRobot.RobotName)
}
```

---

## 🎯 校验规则

### ✅ 允许的情况

1. **首次绑定** - API配置从未绑定过机器人
2. **旧机器人已删除** - 之前的机器人已被删除（deleted_at 不为空）
3. **更换API配置** - 创建新机器人时使用不同的API配置

### ❌ 不允许的情况

1. **重复绑定** - API配置已绑定到一个未删除的机器人
2. **同时多机器人** - 尝试让同一个API配置同时运行多个机器人

---

## 💡 使用场景

### 场景1: 创建第一个机器人
```
用户选择 API配置A → ✅ 允许创建
结果: 机器人A 绑定 API配置A
```

### 场景2: 尝试用同一API创建第二个机器人
```
用户再次选择 API配置A → ❌ 提示错误
错误信息: "该API配置已绑定机器人【机器人A】，每个API配置只能绑定一个机器人"
```

### 场景3: 删除机器人后重新创建
```
1. 删除机器人A
2. 用户选择 API配置A → ✅ 允许创建
结果: 机器人B 绑定 API配置A（机器人A已被删除）
```

### 场景4: 使用不同的API配置
```
用户选择 API配置B（全新配置） → ✅ 允许创建
结果: 机器人B 绑定 API配置B
```

---

## 🔧 实际应用

### 推荐做法

1. **多账户交易**
   - 创建多个API配置（对应不同交易所账户）
   - 每个API配置创建一个机器人
   - 实现多账户同时交易

2. **单账户单策略**
   - 一个API配置 → 一个机器人
   - 专注一个策略，避免冲突

### 不推荐做法

❌ **同一API配置运行多个机器人**
- 会导致订单管理混乱
- 可能产生冲突
- 统计数据不准确

---

## 📊 数据库查询

### 查询某个API配置绑定的机器人

```sql
SELECT robot_name, status, created_at 
FROM hg_trading_robot 
WHERE api_config_id = ? 
  AND deleted_at IS NULL;
```

### 查询所有API配置的绑定情况

```sql
SELECT 
    a.id,
    a.api_name,
    r.robot_name,
    r.status
FROM hg_trading_api_config a
LEFT JOIN hg_trading_robot r 
    ON r.api_config_id = a.id 
    AND r.deleted_at IS NULL
WHERE a.deleted_at IS NULL;
```

---

## 🎯 错误处理

### 错误信息格式

```
该API配置已绑定机器人【{机器人名称}】，每个API配置只能绑定一个机器人
```

### 用户应对方式

1. **删除现有机器人** - 如果不再需要
2. **使用其他API配置** - 创建或选择另一个API配置
3. **等待现有机器人完成任务** - 停止并删除后再创建新的

---

## 🔄 与删除机制的配合

### 软删除设计

机器人使用软删除（`deleted_at` 字段），删除后：
- ✅ 不占用API配置绑定
- ✅ 可以创建新机器人使用该API
- ✅ 历史数据保留，可追溯

### 删除流程

```
机器人运行中 
    ↓
用户点击删除
    ↓
设置 deleted_at = 当前时间
    ↓
API配置释放，可绑定新机器人
```

---

## ⚠️ 注意事项

1. **删除检查** - 只检查 `deleted_at IS NULL` 的机器人
2. **用户隔离** - API配置已通过 `userId` 隔离，不同用户互不影响
3. **实时校验** - 每次创建机器人都会实时检查
4. **准确提示** - 错误信息中包含已绑定的机器人名称

---

## 📈 效果

### 修复前
```
❌ 同一API配置可创建多个机器人
❌ 订单管理混乱
❌ 难以追踪交易来源
```

### 修复后
```
✅ 一个API配置只能绑定一个机器人
✅ 订单归属清晰
✅ 交易行为可控
✅ 用户体验更好
```

---

## 🎉 总结

### ✅ 已实现

- ✅ 创建时校验API配置唯一性
- ✅ 排除已删除的机器人
- ✅ 友好的错误提示
- ✅ 完整的用户隔离

### 📝 建议

1. 前端在创建机器人时，可以提前显示API配置的绑定状态
2. 提供"解绑并创建新机器人"的快捷操作
3. 在API配置列表中显示绑定的机器人信息

---

**实现时间**: 2025-12-23  
**状态**: ✅ 已实现  
**测试**: ⏳ 等待用户验证

