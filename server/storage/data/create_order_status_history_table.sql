-- ============================================================
-- 创建订单状态历史表
-- 用于记录订单在不同节点的状态变更
-- ============================================================

-- 订单状态历史表
CREATE TABLE IF NOT EXISTS `hg_trading_order_status_history` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID（关联hg_trading_order.id）',
  `order_sn` VARCHAR(50) NOT NULL COMMENT '订单号（关联hg_trading_order.order_sn）',
  `exchange_order_id` VARCHAR(100) DEFAULT NULL COMMENT '交易所订单ID',
  `status` INT NOT NULL COMMENT '订单状态：0=未成交,1=持仓中,2=已平仓,3=已取消,4=下单失败',
  `status_text` VARCHAR(50) DEFAULT NULL COMMENT '状态文本',
  `node_type` VARCHAR(50) NOT NULL COMMENT '节点类型：pre_create=预创建,exchange_submit=交易所下单,exchange_success=交易所成功,sync_detail=同步详情,sync_pnl=同步盈亏,close=平仓',
  `node_description` VARCHAR(200) DEFAULT NULL COMMENT '节点描述',
  
  -- 下单时字段
  `quantity` DECIMAL(15,8) DEFAULT NULL COMMENT '数量',
  `price` DECIMAL(15,4) DEFAULT NULL COMMENT '委托价格',
  `avg_price` DECIMAL(15,4) DEFAULT NULL COMMENT '成交均价',
  `open_price` DECIMAL(15,4) DEFAULT NULL COMMENT '开仓价格',
  `filled_qty` DECIMAL(15,8) DEFAULT NULL COMMENT '已成交数量',
  `leverage` INT DEFAULT NULL COMMENT '杠杆倍数',
  `margin` DECIMAL(15,4) DEFAULT NULL COMMENT '保证金',
  `open_margin` DECIMAL(15,4) DEFAULT NULL COMMENT '开仓保证金',
  
  -- 同步时字段
  `mark_price` DECIMAL(15,4) DEFAULT NULL COMMENT '标记价格',
  `unrealized_profit` DECIMAL(15,4) DEFAULT NULL COMMENT '未实现盈亏',
  `highest_profit` DECIMAL(15,4) DEFAULT NULL COMMENT '最高盈利',
  `fee` DECIMAL(15,8) DEFAULT NULL COMMENT '手续费',
  `fee_coin` VARCHAR(20) DEFAULT NULL COMMENT '手续费币种',
  
  -- 平仓时字段
  `close_price` DECIMAL(15,4) DEFAULT NULL COMMENT '平仓价格',
  `realized_profit` DECIMAL(15,4) DEFAULT NULL COMMENT '已实现盈亏',
  `close_reason` VARCHAR(50) DEFAULT NULL COMMENT '平仓原因',
  
  -- 市场状态和风险偏好（下单时）
  `market_state` VARCHAR(50) DEFAULT NULL COMMENT '市场状态',
  `risk_level` VARCHAR(50) DEFAULT NULL COMMENT '风险偏好',
  
  -- 策略参数（下单时）
  `stop_loss_percent` DECIMAL(10,4) DEFAULT NULL COMMENT '止损百分比',
  `auto_start_retreat_percent` DECIMAL(10,4) DEFAULT NULL COMMENT '启动止盈百分比',
  `profit_retreat_percent` DECIMAL(10,4) DEFAULT NULL COMMENT '止盈回撤百分比',
  
  -- 时间字段
  `node_time` DATETIME NOT NULL COMMENT '节点时间（该状态变更的时间）',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  
  PRIMARY KEY (`id`),
  INDEX `idx_order_id` (`order_id`),
  INDEX `idx_order_sn` (`order_sn`),
  INDEX `idx_exchange_order_id` (`exchange_order_id`),
  INDEX `idx_status` (`status`),
  INDEX `idx_node_type` (`node_type`),
  INDEX `idx_node_time` (`node_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单状态历史表';

