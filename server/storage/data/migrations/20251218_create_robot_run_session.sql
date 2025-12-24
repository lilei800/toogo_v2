-- 机器人运行区间记录表
-- 用于精确记录每次启动/停止的时间区间，便于按区间汇总交易所盈亏数据

CREATE TABLE IF NOT EXISTS `hg_trading_robot_run_session` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `robot_id` bigint(20) NOT NULL COMMENT '机器人ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `exchange` varchar(32) NOT NULL DEFAULT '' COMMENT '交易所',
  `symbol` varchar(32) NOT NULL DEFAULT '' COMMENT '交易对',
  `start_time` datetime NOT NULL COMMENT '启动时间',
  `end_time` datetime DEFAULT NULL COMMENT '结束时间（NULL表示仍在运行）',
  `end_reason` varchar(32) DEFAULT '' COMMENT '结束原因：pause/stop/auto_stop/error',
  `runtime_seconds` int(11) NOT NULL DEFAULT 0 COMMENT '运行时长(秒)',
  `total_pnl` decimal(20,8) DEFAULT NULL COMMENT '区间总盈亏(USDT)，从交易所同步',
  `total_fee` decimal(20,8) DEFAULT NULL COMMENT '区间总手续费(USDT)，从交易所同步',
  `trade_count` int(11) DEFAULT 0 COMMENT '区间成交笔数',
  `synced_at` datetime DEFAULT NULL COMMENT '最后同步时间',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_robot_id` (`robot_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_start_time` (`start_time`),
  KEY `idx_end_time` (`end_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='机器人运行区间记录表';
