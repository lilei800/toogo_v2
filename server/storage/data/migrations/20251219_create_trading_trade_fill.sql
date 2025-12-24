-- 创建成交流水表（交易所一笔成交一条记录）
-- 表名：hg_trading_trade_fill
-- 目标：交易明细/盈亏/手续费以交易所成交为准；trading_order 仅作为“订单壳/意图”用于关联展示
-- 幂等：用 uk_api_exchange_trade (api_config_id, exchange, trade_id) 防重复落库

CREATE TABLE IF NOT EXISTS `hg_trading_trade_fill` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',

  `api_config_id` BIGINT NOT NULL DEFAULT 0 COMMENT 'API配置ID',
  `exchange` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '交易所',

  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `robot_id` BIGINT NOT NULL DEFAULT 0 COMMENT '机器人ID',
  `session_id` BIGINT DEFAULT NULL COMMENT '运行区间ID(可选)',

  `symbol` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '交易对',

  `order_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '交易所订单ID',
  `client_order_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '客户端订单ID(可选)',
  `trade_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '成交ID',

  `side` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '方向: BUY/SELL 或 OPEN/CLOSE',
  `qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '成交数量',
  `price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '成交价格',

  `fee` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '手续费(正数)',
  `fee_coin` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '手续费币种',

  `realized_pnl` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '已实现盈亏(USDT口径，按交易所返回)',

  `ts` BIGINT NOT NULL DEFAULT 0 COMMENT '成交时间戳(毫秒)',

  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  PRIMARY KEY (`id`),

  UNIQUE KEY `uk_api_exchange_trade` (`api_config_id`, `exchange`, `trade_id`),

  KEY `idx_user_ts` (`user_id`, `ts`),
  KEY `idx_robot_ts` (`robot_id`, `ts`),
  KEY `idx_symbol_ts` (`symbol`, `ts`),
  KEY `idx_order_ts` (`order_id`, `ts`),
  KEY `idx_session_ts` (`session_id`, `ts`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='成交流水表（交易所成交记录）';
