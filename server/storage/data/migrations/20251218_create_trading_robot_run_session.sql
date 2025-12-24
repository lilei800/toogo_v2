-- 创建机器人运行区间盈亏汇总表
-- 说明：
-- - 仅记录机器人“运行区间”(start/end)，用于后续按区间同步交易所成交数据做盈亏/手续费汇总
-- - 表名与 DAO/entity 保持一致：hg_trading_robot_run_session

CREATE TABLE IF NOT EXISTS `hg_trading_robot_run_session` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `robot_id` BIGINT NOT NULL COMMENT '机器人ID',
  `user_id` BIGINT NOT NULL COMMENT '用户ID',
  `exchange` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '交易所',
  `symbol` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '交易对',
  `start_time` DATETIME NOT NULL COMMENT '启动时间',
  `end_time` DATETIME DEFAULT NULL COMMENT '结束时间（为空表示仍在运行）',
  `end_reason` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '结束原因：pause/stop/auto_stop/restart/error 等',
  `runtime_seconds` INT NOT NULL DEFAULT 0 COMMENT '运行时长(秒)',
  `total_pnl` DECIMAL(20,8) DEFAULT NULL COMMENT '区间总盈亏(USDT)',
  `total_fee` DECIMAL(20,8) DEFAULT NULL COMMENT '区间总手续费(USDT)',
  `trade_count` INT NOT NULL DEFAULT 0 COMMENT '区间成交笔数',
  `synced_at` DATETIME DEFAULT NULL COMMENT '最后同步时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_start` (`user_id`, `start_time`),
  KEY `idx_robot_end` (`robot_id`, `end_time`),
  KEY `idx_robot_start` (`robot_id`, `start_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='机器人运行区间记录（用于按区间同步交易所盈亏/手续费）';


