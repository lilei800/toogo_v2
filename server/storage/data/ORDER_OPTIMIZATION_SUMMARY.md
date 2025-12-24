# 订单流程优化和拆分方案总结

## ✅ 已完成的优化

### 1. 同步服务去重机制优化

**问题**：步骤5（`updateOrderStatus`）和步骤7（同步服务）可能重复更新数据库字段（`avg_price`、`filled_qty`）

**解决方案**：
- 在同步服务中添加字段检查，只更新缺失或需要更新的字段
- 如果字段已存在且有效，不重复更新
- 如果价格/数量差异较大，记录警告并更新为交易所数据

**代码位置**：`order_status_sync.go:778-800`

```go
// 【优化】补全成交信息（只更新缺失字段，避免重复更新）
if historyOrder.AvgPrice > 0 {
    // 只更新缺失的成交均价
    if localOrder.OpenPrice == 0 || localOrder.OpenPrice <= 0 {
        updateData["avg_price"] = historyOrder.AvgPrice
        updateData["open_price"] = historyOrder.AvgPrice
    } else if localOrder.OpenPrice > 0 && math.Abs(localOrder.OpenPrice-historyOrder.AvgPrice) > 0.01 {
        // 如果价格差异较大（超过0.01），可能是数据不一致，更新为交易所数据
        g.Log().Warningf(ctx, "[OrderStatusSync] 订单成交价格不一致: robotId=%d, orderId=%d, localPrice=%.4f, exchangePrice=%.4f, 更新为交易所价格",
            robot.Id, localOrder.Id, localOrder.OpenPrice, historyOrder.AvgPrice)
        updateData["avg_price"] = historyOrder.AvgPrice
        updateData["open_price"] = historyOrder.AvgPrice
    }
}
```

### 2. 完善错误处理

**问题**：步骤5失败时，订单状态可能不一致；步骤7失败时，订单详情可能不完整

**解决方案**：
- 步骤5失败时，记录错误日志，但不阻塞后续流程（因为交易所订单已成功）
- 步骤7失败时，记录警告日志，但不影响主流程
- 所有错误都记录到订单事件表中，便于追踪

**代码位置**：
- `robot_engine.go:4024-4028` - 订单状态更新失败处理
- `order_status_sync.go:1056-1058` - 未实现盈亏更新失败处理

### 3. 订单事件记录系统

**设计**：每个订单生命周期节点创建独立的事件记录，便于追踪和审计

**事件类型**：
- `signal_generated`: 信号生成
- `check_started`: 开仓检查开始
- `pre_created`: 预创建订单记录
- `exchange_ordered`: 交易所下单
- `order_filled`: 订单成交
- `position_updated`: 持仓更新
- `order_closed`: 订单平仓
- `order_failed`: 订单失败

**数据库表**：`hg_trading_order_event`

**代码位置**：
- `order_event.go` - 事件记录函数
- `robot_engine.go` - 在关键节点调用事件记录
- `order_status_sync.go` - 在同步服务中调用事件记录

### 4. 订单记录拆分方案

**设计理念**：每个节点创建独立记录，不更新现有记录

**节点类型**：
- `pre_create`: 预创建订单记录
- `exchange_submit`: 提交到交易所
- `exchange_success`: 交易所下单成功
- `sync_detail`: 同步订单详情
- `sync_pnl`: 同步未实现盈亏
- `close`: 订单平仓
- `failed`: 下单失败

**代码位置**：`order_status_history.go`

## 📊 优化后的流程

```
信号生成 
  ↓ 【事件记录】signal_generated
开仓检查（数据库 + 内存双重验证）
  ↓ 【事件记录】check_started
预创建订单记录（状态=PENDING，事务保护）
  ↓ 【事件记录】pre_created
交易所下单
  ↓ 【事件记录】exchange_ordered (success/failed)
更新订单记录（状态=OPEN，关键字段，事务保护）
  ├─ 订单状态：PENDING → OPEN
  ├─ 交易所订单ID
  ├─ 成交价格（如果API返回）
  └─ 已成交数量（如果API返回）
  ↓ 【事件记录】order_filled
更新内存缓存（成功后）
  ├─ 更新 CurrentPositions
  └─ 更新 PositionTrackers
  ↓
触发同步服务（补全订单详情，异步）
  ├─ 查询订单历史，获取完整成交信息
  ├─ 更新未实现盈亏（基于标记价格）
  ├─ 补全缺失字段（如果缺失）✅ 已有去重检查
  └─ 更新标记价格
  ↓ 【事件记录】position_updated
订单平仓
  ↓ 【事件记录】order_closed
```

## 🎯 关键改进点

1. **职责分工明确**：
   - 步骤5只更新关键字段（订单状态、交易所订单ID）
   - 步骤7只补全和更新实时字段（未实现盈亏、标记价格）

2. **去重机制**：
   - 同步服务只更新缺失或需要更新的字段
   - 避免重复更新已存在的有效数据

3. **错误处理**：
   - 所有错误都记录日志和事件
   - 不阻塞主流程，确保订单能够正常创建

4. **事件追踪**：
   - 每个关键节点都有事件记录
   - 便于问题排查和审计

5. **数据一致性**：
   - 双重验证机制（内存 + 数据库）
   - 以数据库为准，内存为辅

## 📝 数据库变更

### 新增表：`hg_trading_order_event`

```sql
CREATE TABLE IF NOT EXISTS `hg_trading_order_event` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT DEFAULT 0 COMMENT '租户ID',
  `order_id` BIGINT NOT NULL COMMENT '订单ID（关联 hg_trading_order.id）',
  `exchange_order_id` VARCHAR(100) DEFAULT NULL COMMENT '交易所订单ID',
  `event_type` VARCHAR(50) NOT NULL COMMENT '事件类型',
  `event_status` VARCHAR(20) DEFAULT NULL COMMENT '事件状态：success/failed/pending',
  `event_data` JSON DEFAULT NULL COMMENT '事件数据（JSON格式）',
  `event_message` TEXT DEFAULT NULL COMMENT '事件消息（人类可读的描述）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  INDEX `idx_order_id` (`order_id`),
  INDEX `idx_exchange_order_id` (`exchange_order_id`),
  INDEX `idx_event_type` (`event_type`),
  INDEX `idx_created_at` (`created_at`),
  INDEX `idx_order_event_type` (`order_id`, `event_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单事件表（订单生命周期追踪）';
```

## 🔍 使用示例

### 查询订单的所有事件

```sql
SELECT * FROM hg_trading_order_event 
WHERE order_id = 12345 
ORDER BY created_at ASC;
```

### 查询订单的生命周期

```sql
SELECT 
    event_type,
    event_status,
    event_message,
    created_at
FROM hg_trading_order_event 
WHERE order_id = 12345 
ORDER BY created_at ASC;
```

### 统计订单各阶段耗时

```sql
SELECT 
    e1.event_type AS from_event,
    e2.event_type AS to_event,
    TIMESTAMPDIFF(SECOND, e1.created_at, e2.created_at) AS duration_seconds
FROM hg_trading_order_event e1
JOIN hg_trading_order_event e2 ON e1.order_id = e2.order_id
WHERE e1.order_id = 12345
  AND e2.created_at > e1.created_at
ORDER BY e1.created_at ASC
LIMIT 1;
```

## 🚀 下一步建议

1. **性能优化**：
   - 订单事件表可以考虑分区（按时间）
   - 添加定期清理机制（保留最近N天的数据）

2. **监控告警**：
   - 监控订单事件失败率
   - 监控订单状态更新失败率
   - 监控同步服务延迟

3. **数据分析**：
   - 分析订单各阶段耗时
   - 分析订单失败原因
   - 分析同步服务补全字段的覆盖率

