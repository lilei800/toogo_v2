-- =============================================
-- 添加订单算力消耗标记字段
-- 执行时间：2025-12-05
-- 用途：防止手动平仓逃避算力消耗
-- =============================================

-- 1. 添加算力消耗标记字段
ALTER TABLE `hg_trading_order` 
ADD COLUMN `power_consumed` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已消耗算力：0=未消耗，1=已消耗' AFTER `close_reason`,
ADD COLUMN `power_amount` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '消耗的算力数量' AFTER `power_consumed`;

-- 2. 添加索引以提高查询性能
ALTER TABLE `hg_trading_order` 
ADD INDEX idx_power_consumed (power_consumed, status, realized_profit);

-- 说明：
-- 1. power_consumed：标记订单是否已经消耗算力，防止重复扣费
-- 2. power_amount：记录实际消耗的算力数量，用于对账和统计
-- 3. 索引：加速查询未消耗算力的盈利订单
-- 4. 历史订单默认为0，会被定时任务检测并补扣

-- 验证
SELECT COUNT(*) as total_orders,
       SUM(CASE WHEN power_consumed = 0 THEN 1 ELSE 0 END) as not_consumed,
       SUM(CASE WHEN power_consumed = 1 THEN 1 ELSE 0 END) as consumed
FROM hg_trading_order 
WHERE status = 2 AND realized_profit > 0;
