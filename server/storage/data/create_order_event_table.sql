-- ============================================================
-- 订单事件表：记录订单生命周期中的每个节点事件
-- 用于追踪订单的完整生命周期，便于审计和调试
-- ============================================================

-- 创建订单事件表
CREATE TABLE IF NOT EXISTS `hg_trading_order_event` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT DEFAULT 0 COMMENT '租户ID',
  `order_id` BIGINT NOT NULL COMMENT '订单ID（关联 hg_trading_order.id）',
  `exchange_order_id` VARCHAR(100) DEFAULT NULL COMMENT '交易所订单ID',
  `event_type` VARCHAR(50) NOT NULL COMMENT '事件类型：signal_generated/check_started/pre_created/exchange_ordered/order_filled/position_updated/order_closed/order_failed',
  `event_status` VARCHAR(20) DEFAULT NULL COMMENT '事件状态：success/failed/pending',
  `event_data` JSON DEFAULT NULL COMMENT '事件数据（JSON格式，存储事件相关的详细信息）',
  `event_message` TEXT DEFAULT NULL COMMENT '事件消息（人类可读的描述）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  INDEX `idx_order_id` (`order_id`),
  INDEX `idx_exchange_order_id` (`exchange_order_id`),
  INDEX `idx_event_type` (`event_type`),
  INDEX `idx_created_at` (`created_at`),
  INDEX `idx_order_event_type` (`order_id`, `event_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单事件表（订单生命周期追踪）';

-- ============================================================
-- 事件类型说明
-- ============================================================
-- signal_generated: 信号生成（记录信号信息）
-- check_started: 开仓检查开始（记录检查条件）
-- pre_created: 预创建订单记录（记录订单基本信息）
-- exchange_ordered: 交易所下单（记录API请求和响应）
-- order_filled: 订单成交（记录成交详情）
-- position_updated: 持仓更新（记录未实现盈亏等）
-- order_closed: 订单平仓（记录平仓详情）
-- order_failed: 订单失败（记录失败原因）
-- ============================================================

-- 验证表结构
SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_COMMENT
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'hg_trading_order_event'
ORDER BY ORDINAL_POSITION;

