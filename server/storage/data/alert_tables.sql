-- 预警日志数据库表
-- 创建时间: 2024

-- 1. 市场状态预警日志表
CREATE TABLE IF NOT EXISTS `hg_trading_market_alert` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `robot_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '机器人ID',
    `symbol` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '交易对',
    `level` VARCHAR(20) NOT NULL DEFAULT 'INFO' COMMENT '预警级别: INFO/WARNING/DANGER/CRITICAL',
    `title` VARCHAR(200) NOT NULL DEFAULT '' COMMENT '预警标题',
    `message` TEXT COMMENT '预警消息',
    `previous_state` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '之前的市场状态',
    `current_state` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '当前市场状态',
    `trend_score` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '趋势评分 -100到100',
    `volatility` DECIMAL(10,4) NOT NULL DEFAULT 0 COMMENT '波动率',
    `volatility_level` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '波动等级',
    `confidence` DECIMAL(5,4) NOT NULL DEFAULT 0 COMMENT '置信度 0-1',
    `suggest_action` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '建议操作',
    `time_frame_signals` JSON COMMENT '各周期信号',
    `technical_summary` TEXT COMMENT '技术面总结',
    `recommendation` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '操作建议',
    `is_read` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已读',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    INDEX `idx_robot_id` (`robot_id`),
    INDEX `idx_symbol` (`symbol`),
    INDEX `idx_level` (`level`),
    INDEX `idx_current_state` (`current_state`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_robot_created` (`robot_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='市场状态预警日志表';

-- 2. 风险偏好预警日志表
CREATE TABLE IF NOT EXISTS `hg_trading_risk_alert` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `robot_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '机器人ID',
    `symbol` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '交易对',
    `level` VARCHAR(20) NOT NULL DEFAULT 'INFO' COMMENT '预警级别: INFO/WARNING/DANGER/CRITICAL',
    `title` VARCHAR(200) NOT NULL DEFAULT '' COMMENT '预警标题',
    `message` TEXT COMMENT '预警消息',
    `previous_preference` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '之前的风险偏好',
    `current_preference` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '当前风险偏好',
    `win_probability` DECIMAL(5,2) NOT NULL DEFAULT 0 COMMENT '胜算概率 0-100',
    `risk_score` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '风险评分 0-100',
    `account_health` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '账户健康度',
    `suggest_leverage` INT NOT NULL DEFAULT 1 COMMENT '建议杠杆',
    `suggest_position` DECIMAL(5,4) NOT NULL DEFAULT 0 COMMENT '建议仓位比例 0-1',
    `suggest_stop_loss` DECIMAL(5,2) NOT NULL DEFAULT 0 COMMENT '建议止损比例',
    `reasons` JSON COMMENT '判断理由',
    `action_required` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '需要的操作',
    `is_read` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已读',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    INDEX `idx_robot_id` (`robot_id`),
    INDEX `idx_symbol` (`symbol`),
    INDEX `idx_level` (`level`),
    INDEX `idx_current_preference` (`current_preference`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_robot_created` (`robot_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='风险偏好预警日志表';

-- 3. 下单方向预警日志表
CREATE TABLE IF NOT EXISTS `hg_trading_direction_alert` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `robot_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '机器人ID',
    `symbol` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '交易对',
    `level` VARCHAR(20) NOT NULL DEFAULT 'INFO' COMMENT '预警级别: INFO/WARNING/DANGER/CRITICAL',
    `title` VARCHAR(200) NOT NULL DEFAULT '' COMMENT '预警标题',
    `message` TEXT COMMENT '预警消息',
    `direction` VARCHAR(20) NOT NULL DEFAULT 'WAIT' COMMENT '建议方向: LONG/SHORT/WAIT',
    `direction_score` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '方向评分 -100到100',
    `signal_strength` DECIMAL(5,2) NOT NULL DEFAULT 0 COMMENT '信号强度 0-100',
    `entry_price` DECIMAL(20,8) NOT NULL DEFAULT 0 COMMENT '建议入场价',
    `stop_loss_price` DECIMAL(20,8) NOT NULL DEFAULT 0 COMMENT '建议止损价',
    `take_profit_price` DECIMAL(20,8) NOT NULL DEFAULT 0 COMMENT '建议止盈价',
    `risk_reward_ratio` DECIMAL(10,4) NOT NULL DEFAULT 0 COMMENT '风险收益比',
    `time_window` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '时间窗口',
    `volatility_points` DECIMAL(20,8) NOT NULL DEFAULT 0 COMMENT '波动点数',
    `confidence` DECIMAL(5,4) NOT NULL DEFAULT 0 COMMENT '置信度 0-1',
    `market_condition` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '市场状况',
    `technical_signals` JSON COMMENT '技术信号',
    `recommendation` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '操作建议',
    `is_executed` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已执行',
    `execution_result` VARCHAR(50) DEFAULT NULL COMMENT '执行结果',
    `is_read` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已读',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    INDEX `idx_robot_id` (`robot_id`),
    INDEX `idx_symbol` (`symbol`),
    INDEX `idx_level` (`level`),
    INDEX `idx_direction` (`direction`),
    INDEX `idx_signal_strength` (`signal_strength`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_robot_created` (`robot_id`, `created_at`),
    INDEX `idx_executed` (`is_executed`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='下单方向预警日志表';

-- 4. 交易信号综合记录表
CREATE TABLE IF NOT EXISTS `hg_trading_signal` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `robot_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '机器人ID',
    `symbol` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '交易对',
    `final_direction` VARCHAR(20) NOT NULL DEFAULT 'WAIT' COMMENT '最终方向',
    `final_confidence` DECIMAL(5,4) NOT NULL DEFAULT 0 COMMENT '最终置信度',
    `should_trade` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否应该交易',
    `market_state` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '市场状态',
    `risk_preference` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '风险偏好',
    `trend_score` DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '趋势评分',
    `win_probability` DECIMAL(5,2) NOT NULL DEFAULT 0 COMMENT '胜算概率',
    `signal_strength` DECIMAL(5,2) NOT NULL DEFAULT 0 COMMENT '信号强度',
    `market_alert_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '市场预警ID',
    `risk_alert_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '风险预警ID',
    `direction_alert_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '方向预警ID',
    `analysis_data` JSON COMMENT '完整分析数据',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    INDEX `idx_robot_id` (`robot_id`),
    INDEX `idx_symbol` (`symbol`),
    INDEX `idx_final_direction` (`final_direction`),
    INDEX `idx_should_trade` (`should_trade`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_robot_created` (`robot_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='交易信号综合记录表';

-- 5. K线数据缓存表 (用于多周期分析)
CREATE TABLE IF NOT EXISTS `hg_trading_kline_cache` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `symbol` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '交易对',
    `platform` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '平台',
    `time_frame` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '时间周期: 1m/5m/15m/30m/1h',
    `timestamp` BIGINT NOT NULL DEFAULT 0 COMMENT 'K线时间戳',
    `open` DECIMAL(20,8) NOT NULL DEFAULT 0 COMMENT '开盘价',
    `high` DECIMAL(20,8) NOT NULL DEFAULT 0 COMMENT '最高价',
    `low` DECIMAL(20,8) NOT NULL DEFAULT 0 COMMENT '最低价',
    `close` DECIMAL(20,8) NOT NULL DEFAULT 0 COMMENT '收盘价',
    `volume` DECIMAL(30,8) NOT NULL DEFAULT 0 COMMENT '成交量',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_symbol_tf_ts` (`symbol`, `platform`, `time_frame`, `timestamp`),
    INDEX `idx_symbol` (`symbol`),
    INDEX `idx_time_frame` (`time_frame`),
    INDEX `idx_timestamp` (`timestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='K线数据缓存表';

