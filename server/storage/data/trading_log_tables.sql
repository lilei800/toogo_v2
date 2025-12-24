-- ===========================================
-- 交易监控日志表
-- ===========================================

-- 交易操作日志表
CREATE TABLE IF NOT EXISTS `hg_trading_operation_log` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `robot_id` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '机器人ID',
    `user_id` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '用户ID',
    `operation` varchar(50) NOT NULL DEFAULT '' COMMENT '操作类型: OPEN/CLOSE/MODIFY/CANCEL',
    `symbol` varchar(50) NOT NULL DEFAULT '' COMMENT '交易对',
    `side` varchar(20) NOT NULL DEFAULT '' COMMENT '方向: BUY/SELL',
    `position_side` varchar(20) NOT NULL DEFAULT '' COMMENT '持仓方向: LONG/SHORT',
    `order_type` varchar(20) NOT NULL DEFAULT '' COMMENT '订单类型: MARKET/LIMIT',
    `quantity` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '数量',
    `price` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '价格',
    `order_id` varchar(100) NOT NULL DEFAULT '' COMMENT '交易所订单ID',
    `status` varchar(20) NOT NULL DEFAULT '' COMMENT '状态: SUCCESS/FAILED/PENDING',
    `error_code` varchar(50) NOT NULL DEFAULT '' COMMENT '错误码',
    `error_msg` text COMMENT '错误信息',
    `request_data` text COMMENT '请求数据JSON',
    `response_data` text COMMENT '响应数据JSON',
    `execute_time` int(11) NOT NULL DEFAULT 0 COMMENT '执行耗时(ms)',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_robot_time` (`robot_id`, `created_at`),
    KEY `idx_user_time` (`user_id`, `created_at`),
    KEY `idx_symbol_time` (`symbol`, `created_at`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='交易操作日志表';

-- 机器人日统计表
CREATE TABLE IF NOT EXISTS `hg_trading_daily_stats` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `robot_id` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '机器人ID',
    `user_id` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '用户ID',
    `date` date NOT NULL COMMENT '统计日期',
    `symbol` varchar(50) NOT NULL DEFAULT '' COMMENT '交易对',
    `total_trades` int(11) NOT NULL DEFAULT 0 COMMENT '总交易次数',
    `win_trades` int(11) NOT NULL DEFAULT 0 COMMENT '盈利次数',
    `loss_trades` int(11) NOT NULL DEFAULT 0 COMMENT '亏损次数',
    `total_volume` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '总交易量',
    `total_pnl` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '总盈亏',
    `realized_pnl` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '已实现盈亏',
    `commission` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '总手续费',
    `max_profit` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '最大单笔盈利',
    `max_loss` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '最大单笔亏损',
    `max_drawdown` decimal(10,4) NOT NULL DEFAULT 0.0000 COMMENT '最大回撤比例',
    `win_rate` decimal(10,4) NOT NULL DEFAULT 0.0000 COMMENT '胜率',
    `profit_factor` decimal(10,4) NOT NULL DEFAULT 0.0000 COMMENT '盈亏比',
    `avg_holding_time` int(11) NOT NULL DEFAULT 0 COMMENT '平均持仓时间(秒)',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_robot_date` (`robot_id`, `date`),
    KEY `idx_user_date` (`user_id`, `date`),
    KEY `idx_date` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='机器人日统计表';

-- 用户交易汇总表
CREATE TABLE IF NOT EXISTS `hg_trading_user_summary` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '用户ID',
    `total_robots` int(11) NOT NULL DEFAULT 0 COMMENT '机器人总数',
    `active_robots` int(11) NOT NULL DEFAULT 0 COMMENT '活跃机器人数',
    `total_trades` int(11) NOT NULL DEFAULT 0 COMMENT '总交易次数',
    `total_volume` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '总交易量',
    `total_pnl` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '总盈亏',
    `total_commission` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '总手续费',
    `overall_win_rate` decimal(10,4) NOT NULL DEFAULT 0.0000 COMMENT '总体胜率',
    `best_robot_id` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '最佳机器人ID',
    `best_robot_pnl` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '最佳机器人盈亏',
    `first_trade_time` datetime DEFAULT NULL COMMENT '首次交易时间',
    `last_trade_time` datetime DEFAULT NULL COMMENT '最后交易时间',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户交易汇总表';

-- 系统交易监控表
CREATE TABLE IF NOT EXISTS `hg_trading_system_monitor` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `date` date NOT NULL COMMENT '日期',
    `hour` tinyint(4) NOT NULL DEFAULT 0 COMMENT '小时(0-23)',
    `total_users` int(11) NOT NULL DEFAULT 0 COMMENT '活跃用户数',
    `total_robots` int(11) NOT NULL DEFAULT 0 COMMENT '运行机器人数',
    `total_orders` int(11) NOT NULL DEFAULT 0 COMMENT '订单数',
    `total_volume` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '交易量',
    `total_pnl` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '总盈亏',
    `api_calls` int(11) NOT NULL DEFAULT 0 COMMENT 'API调用次数',
    `api_errors` int(11) NOT NULL DEFAULT 0 COMMENT 'API错误次数',
    `avg_latency` int(11) NOT NULL DEFAULT 0 COMMENT '平均延迟(ms)',
    `max_latency` int(11) NOT NULL DEFAULT 0 COMMENT '最大延迟(ms)',
    `ws_connections` int(11) NOT NULL DEFAULT 0 COMMENT 'WebSocket连接数',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_date_hour` (`date`, `hour`),
    KEY `idx_date` (`date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统交易监控表';

-- 交易信号日志表
CREATE TABLE IF NOT EXISTS `hg_trading_signal_log` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `robot_id` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '机器人ID',
    `strategy_id` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '策略ID',
    `symbol` varchar(50) NOT NULL DEFAULT '' COMMENT '交易对',
    `signal_type` varchar(20) NOT NULL DEFAULT '' COMMENT '信号类型: OPEN_LONG/OPEN_SHORT/CLOSE',
    `signal_source` varchar(50) NOT NULL DEFAULT '' COMMENT '信号来源',
    `signal_strength` decimal(10,4) NOT NULL DEFAULT 0.0000 COMMENT '信号强度(0-1)',
    `current_price` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '当前价格',
    `target_price` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '目标价格',
    `stop_loss` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '止损价',
    `take_profit` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '止盈价',
    `executed` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否执行',
    `execute_result` varchar(20) NOT NULL DEFAULT '' COMMENT '执行结果',
    `reason` varchar(255) NOT NULL DEFAULT '' COMMENT '原因/备注',
    `indicators` text COMMENT '指标数据JSON',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_robot_time` (`robot_id`, `created_at`),
    KEY `idx_symbol_time` (`symbol`, `created_at`),
    KEY `idx_signal_type` (`signal_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='交易信号日志表';

-- 行情缓存表
CREATE TABLE IF NOT EXISTS `hg_trading_ticker_cache` (
    `symbol` varchar(50) NOT NULL COMMENT '交易对',
    `platform` varchar(20) NOT NULL DEFAULT 'binance' COMMENT '平台',
    `last_price` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '最新价',
    `bid_price` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '买一价',
    `ask_price` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '卖一价',
    `high_24h` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '24h最高',
    `low_24h` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '24h最低',
    `volume_24h` decimal(20,8) NOT NULL DEFAULT 0.00000000 COMMENT '24h成交量',
    `change_24h` decimal(10,4) NOT NULL DEFAULT 0.0000 COMMENT '24h涨跌幅',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`symbol`, `platform`),
    KEY `idx_updated` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='行情缓存表';

-- ===========================================
-- 审计日志表
-- ===========================================
CREATE TABLE IF NOT EXISTS `hg_audit_log` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '用户ID',
    `username` varchar(100) NOT NULL DEFAULT '' COMMENT '用户名',
    `module` varchar(50) NOT NULL DEFAULT '' COMMENT '模块',
    `action` varchar(50) NOT NULL DEFAULT '' COMMENT '操作',
    `target_type` varchar(50) NOT NULL DEFAULT '' COMMENT '目标类型',
    `target_id` bigint(20) unsigned NOT NULL DEFAULT 0 COMMENT '目标ID',
    `before_data` text COMMENT '操作前数据',
    `after_data` text COMMENT '操作后数据',
    `ip` varchar(50) NOT NULL DEFAULT '' COMMENT 'IP地址',
    `user_agent` varchar(500) NOT NULL DEFAULT '' COMMENT 'UserAgent',
    `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '状态: 1成功 0失败',
    `error_msg` varchar(500) NOT NULL DEFAULT '' COMMENT '错误信息',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_user_time` (`user_id`, `created_at`),
    KEY `idx_module_action` (`module`, `action`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审计日志表';

-- ===========================================
-- 完成
-- ===========================================
SELECT '监控日志表创建完成！' AS message;

