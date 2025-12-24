-- 创建交易执行日志表
-- 用于记录完整的交易执行流程：信号检测、下单尝试、下单成功/失败、持仓监控、平仓执行等

CREATE TABLE IF NOT EXISTS `hg_trading_execution_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `signal_log_id` bigint(20) DEFAULT NULL COMMENT '关联的预警日志ID（可选）',
  `robot_id` bigint(20) NOT NULL COMMENT '机器人ID',
  `order_id` bigint(20) DEFAULT NULL COMMENT '关联的订单ID（可选）',
  `event_type` varchar(50) NOT NULL COMMENT '事件类型：signal_detected/order_attempt/order_success/order_failed/position_monitor/position_close/stop_loss/take_profit',
  `event_data` json DEFAULT NULL COMMENT '事件数据（JSON格式，包含详细信息）',
  `status` varchar(20) NOT NULL DEFAULT 'pending' COMMENT '状态：pending/success/failed',
  `message` text COMMENT '消息（详细说明，TEXT类型）',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_robot_id` (`robot_id`),
  KEY `idx_signal_log_id` (`signal_log_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_event_type` (`event_type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='交易执行日志表';

