# 算力扣除幂等性修复方案

## 问题描述

当前系统存在**盈利单重复扣除算力**的风险：
1. **实时扣算力**：平仓时调用 `ConsumePower()` 成功，但更新 `power_consumed=1` 失败（如数据库连接断开）
2. **补偿扣算力**：定时任务扫描 `power_consumed=0` 的盈利单，再次调用 `ConsumePower()`，导致重复扣除

**根本原因**：`hg_toogo_power_consume` 表没有唯一键约束，同一笔订单可以插入多条消耗记录。

---

## 解决方案（三步走）

### 第一步：数据库加唯一键约束

#### 1.1 检查是否有重复数据

```sql
-- 查询重复的算力消耗记录
SELECT user_id, order_id, COUNT(*) as cnt, SUM(consume_power) as total_power
FROM hg_toogo_power_consume 
GROUP BY user_id, order_id 
HAVING cnt > 1
ORDER BY cnt DESC;
```

#### 1.2 清理重复数据（保留最新的记录）

```sql
-- 删除重复记录（保留 id 最大的那条）
DELETE t1 FROM hg_toogo_power_consume t1
INNER JOIN hg_toogo_power_consume t2 
WHERE t1.user_id = t2.user_id 
  AND t1.order_id = t2.order_id 
  AND t1.id < t2.id;

-- 验证清理结果（应该返回 0 行）
SELECT user_id, order_id, COUNT(*) as cnt 
FROM hg_toogo_power_consume 
GROUP BY user_id, order_id 
HAVING cnt > 1;
```

#### 1.3 添加唯一键约束

```sql
-- 添加唯一键约束（确保同一用户的同一订单只能扣一次算力）
ALTER TABLE hg_toogo_power_consume 
ADD UNIQUE KEY uk_user_order (user_id, order_id);

-- 验证唯一键是否创建成功
SHOW INDEX FROM hg_toogo_power_consume WHERE Key_name = 'uk_user_order';
```

---

### 第二步：代码层面实现幂等检查

#### 2.1 修改 `ConsumePower()` 函数（已完成 ✅）

文件：`hotgo_v2/server/internal/logic/toogo/wallet.go`

```go
// ConsumePower 消耗算力 (盈利订单扣除，只扣算力不扣积分)
func (s *sToogoWallet) ConsumePower(ctx context.Context, userId int64, robotId int64, orderId int64, orderSn string, profitAmount float64) error {
	// 【幂等检查】查询是否已经扣除过算力
	var existingLog *entity.ToogoPowerConsume
	err := dao.ToogoPowerConsume.Ctx(ctx).
		Where("user_id", userId).
		Where("order_id", orderId).
		Scan(&existingLog)
	if err != nil {
		g.Log().Warningf(ctx, "[ConsumePower] 查询算力消耗记录失败: userId=%d, orderId=%d, err=%v", userId, orderId, err)
		// 查询失败不影响后续逻辑（容错）
	}
	if existingLog != nil {
		// 已经扣除过，直接返回成功（幂等）
		g.Log().Infof(ctx, "[ConsumePower] 订单已扣除算力，跳过重复扣除: userId=%d, orderId=%d, consumePower=%.4f", 
			userId, orderId, existingLog.ConsumePower)
		return nil
	}

	// ... 后续扣算力逻辑保持不变 ...
}
```

**关键点**：
- 在扣算力之前，先查询 `hg_toogo_power_consume` 表是否已有记录
- 如果已有记录，直接返回成功（不重复扣）
- 配合数据库唯一键 `uk_user_order`，双重保障幂等性

#### 2.2 确保 `power_consumed` 字段一致性

文件：`hotgo_v2/server/internal/logic/toogo/order_status_sync.go`（已有实现 ✅）

```go
// 【优化】如果是盈利订单且未扣除算力，消耗算力并更新订单表的算力消耗字段
if realizedProfit > 0 {
	// 检查是否已扣除算力
	var powerConsumed int
	powerData, _ := dao.TradingOrder.Ctx(ctx).
		Where("id", order.Id).
		Fields("power_consumed").
		One()
	if powerData != nil && !powerData.IsEmpty() {
		powerConsumed = powerData["power_consumed"].Int()
	}

	if powerConsumed == 0 {
		// 调用 ConsumePower（内部已实现幂等）
		err = service.ToogoWallet().ConsumePower(ctx, order.UserId, order.RobotId, order.Id, order.OrderSn, realizedProfit)
		if err != nil {
			g.Log().Warningf(ctx, "[OrderStatusSync] 消耗算力失败（不影响平仓）: orderId=%d, err=%v", order.Id, err)
		} else {
			// 成功后更新 power_consumed=1
			_, _ = dao.TradingOrder.Ctx(ctx).
				Where("id", order.Id).
				Data(g.Map{"power_consumed": 1, "power_amount": powerAmount, "updated_at": gtime.Now()}).
				Update()
		}
	}
}
```

---

### 第三步：补偿扣算力逻辑保持不变

文件：`hotgo_v2/server/internal/logic/toogo/order_sync.go`（无需修改 ✅）

```go
// SyncClosedOrders 同步已平仓订单并补扣算力
func (s *sToogoRobot) SyncClosedOrders(ctx context.Context) error {
	// 1. 查询所有状态为"已平仓"且未消耗算力的盈利订单
	var orders []*entity.TradingOrder
	err := dao.TradingOrder.Ctx(ctx).
		Where("status", 2).                                         // 已平仓
		Where("power_consumed", 0).                                 // 未消耗算力
		Where("realized_profit > 0").                               // 盈利订单
		Where("close_time >= ?", gtime.Now().Add(-7*24*time.Hour)). // 最近7天
		OrderDesc("close_time").
		Scan(&orders)

	// 2. 逐个订单补扣算力
	for _, order := range orders {
		// 调用 ConsumePower（内部已实现幂等，不会重复扣）
		_, err := s.consumeOrderPower(ctx, order)
		if err != nil {
			g.Log().Errorf(ctx, "[OrderSync] 订单 %s 补扣算力失败: %v", order.OrderSn, err)
			continue
		}
		// 成功后更新 power_consumed=1
		_, _ = dao.TradingOrder.Ctx(ctx).
			Where("id", order.Id).
			Data(g.Map{"power_consumed": 1, "power_amount": powerAmount, "updated_at": gtime.Now()}).
			Update()
	}
	return nil
}
```

**关键点**：
- 即使"实时扣算力"时 `ConsumePower()` 成功但 `UPDATE power_consumed=1` 失败
- 补偿扣算力时，`ConsumePower()` 会检测到 `hg_toogo_power_consume` 表已有记录，直接返回成功（不重复扣）
- 然后更新 `power_consumed=1`，修复数据一致性

---

## 实施步骤

### 1. 备份数据（必须！）

```bash
# 备份 hg_toogo_power_consume 表
mysqldump -u root -p hotgo_db hg_toogo_power_consume > backup_power_consume_$(date +%Y%m%d_%H%M%S).sql

# 备份 hg_trading_order 表（power_consumed 字段）
mysqldump -u root -p hotgo_db hg_trading_order --where="power_consumed=1" > backup_trading_order_power_$(date +%Y%m%d_%H%M%S).sql
```

### 2. 执行数据库迁移（按顺序）

```sql
-- 2.1 检查重复数据
SELECT user_id, order_id, COUNT(*) as cnt, SUM(consume_power) as total_power
FROM hg_toogo_power_consume 
GROUP BY user_id, order_id 
HAVING cnt > 1
ORDER BY cnt DESC;

-- 2.2 清理重复数据（如果有）
DELETE t1 FROM hg_toogo_power_consume t1
INNER JOIN hg_toogo_power_consume t2 
WHERE t1.user_id = t2.user_id 
  AND t1.order_id = t2.order_id 
  AND t1.id < t2.id;

-- 2.3 添加唯一键约束
ALTER TABLE hg_toogo_power_consume 
ADD UNIQUE KEY uk_user_order (user_id, order_id);

-- 2.4 验证唯一键
SHOW INDEX FROM hg_toogo_power_consume WHERE Key_name = 'uk_user_order';
```

### 3. 重启后端服务

```bash
cd D:\go\src\hotgo_v2\server
go build -o hotgo.exe main.go
# 停止旧进程
taskkill /F /IM hotgo.exe
# 启动新进程
start hotgo.exe
```

### 4. 验证修复效果

#### 4.1 测试幂等性

```sql
-- 查询某个订单的算力消耗记录（应该只有1条）
SELECT * FROM hg_toogo_power_consume WHERE order_id = 123;

-- 尝试手动触发补偿扣算力（应该跳过已扣除的订单）
-- 观察日志：应该输出 "订单已扣除算力，跳过重复扣除"
```

#### 4.2 监控日志

```bash
# 实时查看日志
tail -f D:\go\src\hotgo_v2\server\storage\logs\app.log | grep -E "ConsumePower|OrderSync"

# 关键日志示例：
# [ConsumePower] 订单已扣除算力，跳过重复扣除: userId=1001, orderId=12345, consumePower=0.0792
# [OrderSync] 订单 TO20241214120000ABCDEF 补扣算力成功，盈利: 0.7920 USDT，消耗算力: 0.0792
```

---

## 风险评估

| 风险项 | 影响 | 缓解措施 |
|--------|------|----------|
| 清理重复数据时误删 | 高 | ① 先备份数据<br>② 使用 `id < t2.id` 保留最新记录<br>③ 清理后验证 |
| 唯一键创建失败 | 中 | ① 检查表是否有其他唯一键冲突<br>② 确保 `user_id` 和 `order_id` 字段非空 |
| 代码幂等检查失败 | 低 | ① 查询失败时容错（不影响后续逻辑）<br>② 配合数据库唯一键双重保障 |
| 补偿扣算力遗漏 | 低 | ① 定时任务每天执行<br>② 扫描最近7天数据 |

---

## 回滚方案

如果修复后出现问题，可以按以下步骤回滚：

```sql
-- 1. 删除唯一键约束
ALTER TABLE hg_toogo_power_consume DROP INDEX uk_user_order;

-- 2. 恢复备份数据
mysql -u root -p hotgo_db < backup_power_consume_20241214_120000.sql

-- 3. 回滚代码（恢复旧版本的 wallet.go）
git checkout HEAD~1 hotgo_v2/server/internal/logic/toogo/wallet.go

-- 4. 重启后端服务
```

---

## 历史订单同步补充方案（可选）

如果需要补全"交易所手动平仓但本地数据库缺失"的历史订单，可以增加以下定时任务：

```go
// SyncHistoricalClosedOrders 补全历史已平仓订单（每天执行一次）
func (s *sToogoRobot) SyncHistoricalClosedOrders(ctx context.Context) error {
	// 1. 查询所有运行中的机器人
	var robots []*entity.TradingRobot
	err := dao.TradingRobot.Ctx(ctx).Where("status", 1).Scan(&robots)
	if err != nil {
		return err
	}

	// 2. 逐个机器人查询交易所的历史已平仓订单（最近7天）
	for _, robot := range robots {
		ex, err := GetExchangeManager().GetExchange(ctx, robot.Id)
		if err != nil {
			continue
		}

		// 查询最近7天的已平仓订单
		closedOrders, err := ex.GetClosedOrders(robot.Symbol, 7*24*time.Hour)
		if err != nil {
			g.Log().Warningf(ctx, "[SyncHistoricalClosedOrders] 查询已平仓订单失败: robotId=%d, err=%v", robot.Id, err)
			continue
		}

		// 3. 调用 SyncOrderHistoryToDB 同步到本地数据库
		err = service.ToogoRobot().SyncOrderHistoryToDB(ctx, robot.Id, robot, closedOrders)
		if err != nil {
			g.Log().Errorf(ctx, "[SyncHistoricalClosedOrders] 同步订单失败: robotId=%d, err=%v", robot.Id, err)
		}
	}

	return nil
}
```

**注意**：这个任务是可选的，因为现有的 `syncPositionsWithCache()` 已经能保证"持仓中的订单"一定会在本地数据库有记录。

---

## 总结

| 问题 | 解决方案 | 状态 |
|------|----------|------|
| 盈利单重复扣算力 | ① 数据库唯一键 `uk_user_order`<br>② 代码层面幂等检查<br>③ `power_consumed` 字段一致性 | ✅ 已修复 |
| 历史订单同步不完整 | ① 实时对账 `createOrderFromPosition()`<br>② 补全字段 `CloseOrder()`<br>③ 可选：补全历史已平仓订单 | ✅ 已基本解决 |

**关键收益**：
- **幂等性保障**：同一笔订单最多扣一次算力，即使补偿扣算力多次执行也不会重复扣除
- **数据一致性**：`power_consumed` 字段与 `hg_toogo_power_consume` 表保持一致
- **容错能力**：即使实时扣算力失败，补偿扣算力也能修复数据

