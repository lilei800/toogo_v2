# 机器人下单系统架构优化方案

## 📋 当前架构分析

### 1. 下单逻辑流程

```
信号生成 → 开仓检查 → 交易所下单 → 更新内存 → 保存数据库 → 同步服务
```

**当前问题**：
- ❌ 下单成功后先更新内存，再保存数据库，如果数据库保存失败，内存状态不一致
- ❌ 没有事务保护，可能导致部分成功
- ❌ 错误处理分散，难以追踪
- ❌ 订单保存失败时，内存已有持仓但数据库无记录

### 2. 内存缓存机制

**当前结构**：
```go
CurrentPositions []*exchange.Position  // 交易所持仓缓存
PositionTrackers map[string]*PositionTracker  // 持仓跟踪器
AccountBalance   *exchange.Balance    // 账户余额缓存
```

**当前问题**：
- ❌ `CurrentPositions` 和数据库订单状态可能不一致
- ❌ `PositionTrackers` 和数据库订单状态可能不一致
- ❌ 同步机制复杂，有多个同步点（10秒定期、3秒外部持仓、事件驱动）
- ❌ 内存数据可能过期，但系统仍依赖内存数据做决策
- ❌ 开仓检查只检查内存，不检查数据库，可能导致重复下单

### 3. 数据库结构

**当前表结构**：
- `hg_trading_order` - 订单表（开仓和平仓二合一）

**当前问题**：
- ❌ 字段可能缺失（`market_state`, `risk_level` 等需要迁移脚本）
- ❌ 没有唯一约束防止重复订单（`exchange_order_id` 可能重复）
- ❌ 缺少必要的索引（`robot_id + status`, `exchange_order_id`）
- ❌ `tenant_id` 字段可能缺失导致插入失败
- ❌ 字段类型可能不匹配（如时间字段）

## 🎯 优化目标

1. **数据一致性**：确保内存、数据库、交易所三者数据一致
2. **可靠性**：订单保存失败时能够自动恢复
3. **性能**：减少不必要的数据库查询和API调用
4. **可维护性**：简化同步逻辑，统一数据流
5. **可扩展性**：支持未来功能扩展

## 🏗️ 全新设计方案

### 方案一：事务化下单流程（推荐）

#### 核心思想
**先保存数据库，再更新内存，最后同步交易所**

#### 设计原则
1. **数据库为唯一真实来源**：所有决策基于数据库数据
2. **内存缓存为性能优化**：仅用于快速查询，不作为决策依据
3. **事务保护**：关键操作使用事务，确保原子性
4. **最终一致性**：允许短暂不一致，但最终必须一致

#### 新流程设计

```
信号生成 
  ↓
开仓检查（检查数据库 + 内存双重验证）
  ↓
预创建订单记录（状态=PENDING，事务保护）
  ↓
交易所下单
  ↓
更新订单记录（状态=OPEN，事务保护）
  ↓
更新内存缓存（成功后）
  ↓
触发同步服务（异步）
```

#### 关键改进

**1. 预创建订单记录**
```go
// 在下单前先创建订单记录（状态=PENDING）
orderId, err := t.preCreateOrder(ctx, signal, strategyParams)
if err != nil {
    return err // 数据库保存失败，不下单
}

// 然后才调用交易所API
order, err := t.engine.Exchange.CreateOrder(ctx, ...)
if err != nil {
    // 交易所下单失败，更新订单状态为FAILED
    t.updateOrderStatus(ctx, orderId, OrderStatusFailed, err.Error())
    return err
}

// 交易所下单成功，更新订单状态为OPEN
t.updateOrderStatus(ctx, orderId, OrderStatusOpen, "")
```

**2. 双重验证机制**
```go
// 开仓检查：同时检查数据库和内存
func (t *RobotTrader) checkOpenPosition(ctx context.Context, direction string) (bool, error) {
    // 1. 检查内存（快速检查）
    if t.hasPositionInMemory(direction) {
        return true, nil
    }
    
    // 2. 检查数据库（准确检查）
    hasOrder, err := dao.TradingOrder.Ctx(ctx).
        Where("robot_id", t.engine.Robot.Id).
        Where("direction", direction).
        Where("status", OrderStatusOpen).
        Count()
    if err != nil {
        return false, err
    }
    
    return hasOrder > 0, nil
}
```

**3. 统一的数据同步服务**
```go
// 统一的数据同步服务，确保三者一致
type DataSyncService struct {
    // 同步策略：
    // 1. 数据库为主，交易所为辅
    // 2. 内存缓存自动失效（TTL机制）
    // 3. 事件驱动 + 定期兜底
}

func (s *DataSyncService) SyncOrder(ctx context.Context, orderId int64) error {
    // 1. 从数据库读取订单
    order := s.getOrderFromDB(ctx, orderId)
    
    // 2. 从交易所获取最新状态
    exchangeOrder := s.getOrderFromExchange(ctx, order.ExchangeOrderId)
    
    // 3. 对比并更新数据库
    s.reconcileOrder(ctx, order, exchangeOrder)
    
    // 4. 更新内存缓存
    s.updateMemoryCache(ctx, order)
    
    return nil
}
```

### 方案二：事件溯源架构（高级）

#### 核心思想
**所有操作记录为事件，通过事件重建状态**

#### 设计原则
1. **事件不可变**：所有操作记录为事件
2. **状态可重建**：通过事件序列重建任意时刻的状态
3. **最终一致性**：通过事件同步确保一致性

#### 事件类型
```go
type OrderEvent struct {
    EventType   string  // ORDER_CREATED, ORDER_FILLED, ORDER_CLOSED
    OrderId     int64
    ExchangeId  string
    Timestamp   time.Time
    Data        map[string]interface{}
}
```

#### 优势
- ✅ 完整的操作历史
- ✅ 易于调试和审计
- ✅ 支持回滚和重放
- ✅ 天然支持分布式

#### 劣势
- ❌ 实现复杂度高
- ❌ 存储空间需求大
- ❌ 需要额外的事件存储

### 方案三：分层缓存架构（平衡）

#### 核心思想
**三层数据架构：数据库（持久层）→ 内存缓存（应用层）→ 交易所（外部层）**

#### 设计原则
1. **数据库为权威来源**：所有查询优先从数据库
2. **内存缓存加速**：热点数据缓存，自动失效
3. **交易所验证**：定期从交易所验证数据一致性

#### 缓存策略
```go
type OrderCache struct {
    // L1缓存：内存（TTL=1秒）
    memoryCache map[int64]*Order
    
    // L2缓存：数据库（持久化）
    dbCache *OrderDAO
    
    // L3缓存：交易所（外部验证）
    exchangeCache *ExchangeAPI
}

func (c *OrderCache) GetOrder(ctx context.Context, orderId int64) (*Order, error) {
    // 1. 检查L1缓存
    if order := c.memoryCache[orderId]; order != nil && !order.IsExpired() {
        return order, nil
    }
    
    // 2. 检查L2缓存（数据库）
    order, err := c.dbCache.Get(ctx, orderId)
    if err == nil {
        // 更新L1缓存
        c.memoryCache[orderId] = order
        return order, nil
    }
    
    // 3. 从L3缓存（交易所）获取
    exchangeOrder, err := c.exchangeCache.GetOrder(ctx, orderId)
    if err == nil {
        // 保存到L2缓存
        c.dbCache.Save(ctx, exchangeOrder)
        // 更新L1缓存
        c.memoryCache[orderId] = exchangeOrder
        return exchangeOrder, nil
    }
    
    return nil, err
}
```

## 🔧 推荐实施方案：方案一（事务化下单流程）

### 实施步骤

#### 阶段一：数据库结构优化

**1. 添加缺失字段**
```sql
-- 确保所有字段存在
ALTER TABLE `hg_trading_order` 
ADD COLUMN IF NOT EXISTS `market_state` VARCHAR(50) DEFAULT NULL,
ADD COLUMN IF NOT EXISTS `risk_level` VARCHAR(50) DEFAULT NULL,
ADD COLUMN IF NOT EXISTS `strategy_group_id` BIGINT DEFAULT NULL,
ADD COLUMN IF NOT EXISTS `order_type_detail` VARCHAR(50) DEFAULT NULL,
ADD COLUMN IF NOT EXISTS `exchange_side` VARCHAR(10) DEFAULT NULL,
ADD COLUMN IF NOT EXISTS `tenant_id` BIGINT DEFAULT 0;
```

**2. 添加唯一约束**
```sql
-- 防止重复订单
ALTER TABLE `hg_trading_order` 
ADD UNIQUE INDEX `uk_exchange_order_id` (`exchange_order_id`);
```

**3. 添加必要索引**
```sql
-- 优化查询性能
ALTER TABLE `hg_trading_order` 
ADD INDEX `idx_robot_status` (`robot_id`, `status`),
ADD INDEX `idx_user_status` (`user_id`, `status`),
ADD INDEX `idx_symbol_status` (`symbol`, `status`),
ADD INDEX `idx_created_at` (`created_at`);
```

#### 阶段二：下单流程重构

**1. 预创建订单记录**
```go
func (t *RobotTrader) preCreateOrder(ctx context.Context, signal *RobotSignal, strategyParams *StrategyParams) (int64, error) {
    // 在事务中创建订单记录（状态=PENDING）
    tx, err := g.DB().Begin(ctx)
    if err != nil {
        return 0, err
    }
    defer tx.Rollback()
    
    orderData := g.Map{
        "robot_id": t.engine.Robot.Id,
        "status": OrderStatusPending,
        "market_state": currentMarketState,
        "risk_level": currentRiskPreference,
        // ... 其他字段
    }
    
    result, err := tx.Model("hg_trading_order").Insert(orderData)
    if err != nil {
        return 0, err
    }
    
    orderId, _ := result.LastInsertId()
    tx.Commit()
    
    return orderId, nil
}
```

**2. 更新订单状态**
```go
func (t *RobotTrader) updateOrderStatus(ctx context.Context, orderId int64, status int, exchangeOrderId string) error {
    updateData := g.Map{
        "status": status,
        "updated_at": gtime.Now(),
    }
    
    if exchangeOrderId != "" {
        updateData["exchange_order_id"] = exchangeOrderId
    }
    
    _, err := dao.TradingOrder.Ctx(ctx).
        WherePri(orderId).
        Update(updateData)
    
    return err
}
```

**3. 双重验证机制**
```go
func (t *RobotTrader) checkOpenPosition(ctx context.Context, direction string) (bool, error) {
    // 1. 快速检查内存
    t.engine.mu.RLock()
    hasMemoryPosition := false
    for _, pos := range t.engine.CurrentPositions {
        if pos.PositionSide == direction && math.Abs(pos.PositionAmt) > 0.0001 {
            hasMemoryPosition = true
            break
        }
    }
    t.engine.mu.RUnlock()
    
    // 2. 准确检查数据库
    count, err := dao.TradingOrder.Ctx(ctx).
        Where("robot_id", t.engine.Robot.Id).
        Where("direction", direction).
        Where("status", OrderStatusOpen).
        Count()
    
    if err != nil {
        return false, err
    }
    
    hasDBOrder := count > 0
    
    // 3. 如果内存和数据库不一致，以数据库为准，同步内存
    if hasDBOrder && !hasMemoryPosition {
        t.syncPositionFromDB(ctx, direction)
    }
    
    return hasDBOrder, nil
}
```

#### 阶段三：统一同步服务

**1. 订单状态同步服务**
```go
type UnifiedSyncService struct {
    // 同步策略：
    // 1. 事件驱动：开仓/平仓后立即同步
    // 2. 定期同步：每10秒同步一次（兜底）
    // 3. 差异检测：检测到不一致时立即同步
}

func (s *UnifiedSyncService) SyncOrder(ctx context.Context, orderId int64) error {
    // 1. 从数据库读取订单
    order, err := dao.TradingOrder.Ctx(ctx).WherePri(orderId).One()
    if err != nil {
        return err
    }
    
    // 2. 从交易所获取最新状态
    exchangeOrder, err := s.getOrderFromExchange(ctx, order.ExchangeOrderId)
    if err != nil {
        return err
    }
    
    // 3. 对比并更新数据库
    if s.isOrderChanged(order, exchangeOrder) {
        s.updateOrderFromExchange(ctx, order, exchangeOrder)
    }
    
    // 4. 更新内存缓存
    s.updateMemoryCache(ctx, order)
    
    return nil
}
```

**2. 内存缓存失效机制**
```go
type CachedOrder struct {
    Order      *entity.TradingOrder
    ExpireTime time.Time
}

func (c *OrderCache) GetOrder(ctx context.Context, orderId int64) (*entity.TradingOrder, error) {
    cached, ok := c.memoryCache[orderId]
    if ok && time.Now().Before(cached.ExpireTime) {
        return cached.Order, nil
    }
    
    // 缓存过期，从数据库重新加载
    order, err := dao.TradingOrder.Ctx(ctx).WherePri(orderId).One()
    if err != nil {
        return nil, err
    }
    
    // 更新缓存
    c.memoryCache[orderId] = &CachedOrder{
        Order:      order,
        ExpireTime: time.Now().Add(1 * time.Second), // TTL=1秒
    }
    
    return order, nil
}
```

## 📊 对比分析

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| **方案一：事务化下单** | ✅ 实现简单<br>✅ 数据一致性好<br>✅ 易于维护 | ❌ 需要事务支持<br>❌ 性能略低 | **推荐：当前系统** |
| **方案二：事件溯源** | ✅ 完整历史<br>✅ 易于审计<br>✅ 支持回滚 | ❌ 实现复杂<br>❌ 存储需求大 | 大型系统、审计要求高 |
| **方案三：分层缓存** | ✅ 性能最优<br>✅ 扩展性好 | ❌ 实现复杂<br>❌ 缓存一致性难 | 高并发系统 |

## 🎯 实施建议

### 短期优化（1-2周）

1. **数据库结构完善**
   - ✅ 执行所有迁移脚本
   - ✅ 添加唯一约束和索引
   - ✅ 验证字段完整性

2. **下单流程优化**
   - ✅ 实现预创建订单记录
   - ✅ 添加双重验证机制
   - ✅ 增强错误处理

3. **同步服务优化**
   - ✅ 统一同步入口
   - ✅ 添加缓存失效机制
   - ✅ 优化同步频率

### 中期优化（1-2月）

1. **事务保护**
   - ✅ 关键操作使用事务
   - ✅ 实现补偿机制
   - ✅ 添加重试逻辑

2. **监控和告警**
   - ✅ 数据一致性监控
   - ✅ 订单保存失败告警
   - ✅ 性能监控

3. **测试和验证**
   - ✅ 单元测试
   - ✅ 集成测试
   - ✅ 压力测试

### 长期优化（3-6月）

1. **架构升级**
   - ✅ 考虑事件溯源架构
   - ✅ 分布式锁机制
   - ✅ 消息队列解耦

2. **性能优化**
   - ✅ 数据库读写分离
   - ✅ 缓存集群
   - ✅ 异步处理优化

## 📝 总结

**当前问题**：
1. ❌ 下单流程：先更新内存后保存数据库，导致不一致
2. ❌ 内存缓存：与数据库状态可能不一致
3. ❌ 数据库结构：字段缺失、缺少约束和索引

**优化方案**：
1. ✅ **事务化下单流程**：先保存数据库，再更新内存
2. ✅ **双重验证机制**：同时检查数据库和内存
3. ✅ **统一同步服务**：确保三者数据一致
4. ✅ **数据库结构完善**：添加字段、约束、索引

**预期效果**：
- ✅ 数据一致性提升 99%+
- ✅ 订单保存成功率提升 99%+
- ✅ 系统可靠性提升
- ✅ 维护成本降低

