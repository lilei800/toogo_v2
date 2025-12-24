-- ============================================================
-- 创建交易所订单事实表（挂单/订单）- MySQL
-- 目标：前端“挂单列表/订单状态”统一只读 DB；私有WS增量 upsert；REST 低频兜底对账。
-- ============================================================

CREATE TABLE IF NOT EXISTS `hg_trading_exchange_order` (
  `id` BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '用户ID',
  `robot_id` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '机器人ID',
  `api_config_id` BIGINT(20) NOT NULL DEFAULT 0 COMMENT 'API配置ID（同一账户内 orderId 才唯一）',
  `platform` VARCHAR(20) NOT NULL COMMENT '交易所平台：binance/okx/gate/bitget',
  `symbol` VARCHAR(50) NOT NULL COMMENT '交易对（使用机器人配置中的 symbol，保持前端一致）',

  `exchange_order_id` VARCHAR(120) NOT NULL COMMENT '交易所订单ID',
  `client_order_id` VARCHAR(120) DEFAULT NULL COMMENT '客户端订单ID（如有）',

  `side` VARCHAR(10) DEFAULT NULL COMMENT 'BUY/SELL',
  `position_side` VARCHAR(10) DEFAULT NULL COMMENT 'LONG/SHORT',
  `order_type` VARCHAR(30) DEFAULT NULL COMMENT 'MARKET/LIMIT/...',
  `reduce_only` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否只减仓',

  `price` DECIMAL(20,8) DEFAULT NULL COMMENT '委托价',
  `quantity` DECIMAL(28,12) DEFAULT NULL COMMENT '委托数量（统一为基础币数量口径，尽力归一化）',
  `filled_qty` DECIMAL(28,12) DEFAULT NULL COMMENT '已成交数量',
  `avg_price` DECIMAL(20,8) DEFAULT NULL COMMENT '成交均价',

  `status` VARCHAR(30) DEFAULT NULL COMMENT '标准化状态：NEW/PARTIALLY_FILLED/FILLED/CANCELED/REJECTED/EXPIRED',
  `raw_status` VARCHAR(50) DEFAULT NULL COMMENT '交易所原始状态（如 state/status）',
  `is_open` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否挂单中（用于挂单列表查询）',

  `create_time` BIGINT(20) DEFAULT NULL COMMENT '交易所创建时间(ms)',
  `update_time` BIGINT(20) DEFAULT NULL COMMENT '交易所更新时间(ms)',
  `last_event_time` BIGINT(20) DEFAULT NULL COMMENT '最近一次事件时间(ms)',
  `raw` TEXT DEFAULT NULL COMMENT '原始消息/响应（截断存储，便于审计）',

  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_platform_api_order` (`platform`, `api_config_id`, `exchange_order_id`),
  KEY `idx_robot_open` (`robot_id`, `is_open`),
  KEY `idx_robot_symbol_open` (`robot_id`, `symbol`, `is_open`),
  KEY `idx_user_open` (`user_id`, `is_open`),
  KEY `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='交易所订单事实表（挂单/订单，WS增量同步）';


