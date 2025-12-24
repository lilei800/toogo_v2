-- ============================================================
-- 机器人交易系统 V2 - 预警日志表结构
-- ============================================================

-- 1. 市场状态预警日志表
DROP TABLE IF EXISTS `hg_trading_market_state_log`;
CREATE TABLE `hg_trading_market_state_log` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `platform` varchar(50) NOT NULL DEFAULT '' COMMENT '交易所平台',
    `symbol` varchar(50) NOT NULL DEFAULT '' COMMENT '交易对',
    `prev_state` varchar(20) NOT NULL DEFAULT '' COMMENT '之前状态',
    `new_state` varchar(20) NOT NULL DEFAULT '' COMMENT '新状态',
    `confidence` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '置信度',
    `reason` varchar(500) NOT NULL DEFAULT '' COMMENT '变化原因',
    `trend_strength` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '趋势强度',
    `volatility` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '波动率',
    `trend_score` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '趋势评分',
    `momentum_score` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '动量评分',
    `indicators` json DEFAULT NULL COMMENT '技术指标数据JSON',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_platform_symbol` (`platform`, `symbol`),
    KEY `idx_new_state` (`new_state`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='市场状态预警日志';

-- 2. 风险偏好预警日志表
DROP TABLE IF EXISTS `hg_trading_risk_preference_log`;
CREATE TABLE `hg_trading_risk_preference_log` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `robot_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '机器人ID',
    `user_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '用户ID',
    `platform` varchar(50) NOT NULL DEFAULT '' COMMENT '交易所平台',
    `symbol` varchar(50) NOT NULL DEFAULT '' COMMENT '交易对',
    `prev_preference` varchar(20) NOT NULL DEFAULT '' COMMENT '之前偏好',
    `new_preference` varchar(20) NOT NULL DEFAULT '' COMMENT '新偏好',
    `win_probability` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '胜算概率',
    `market_score` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '市场评分',
    `technical_score` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '技术评分',
    `account_score` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '账户评分',
    `history_score` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '历史评分',
    `volatility_risk` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '波动风险',
    `suggested_leverage` int(11) NOT NULL DEFAULT '10' COMMENT '建议杠杆',
    `suggested_margin_percent` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '建议保证金比例',
    `suggested_stop_loss` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '建议止损比例',
    `reason` varchar(500) NOT NULL DEFAULT '' COMMENT '评估原因',
    `factors` json DEFAULT NULL COMMENT '评估因素JSON',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_robot_id` (`robot_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_new_preference` (`new_preference`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='风险偏好预警日志';

-- 3. 方向预警日志表
DROP TABLE IF EXISTS `hg_trading_direction_log`;
CREATE TABLE `hg_trading_direction_log` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `platform` varchar(50) NOT NULL DEFAULT '' COMMENT '交易所平台',
    `symbol` varchar(50) NOT NULL DEFAULT '' COMMENT '交易对',
    `prev_direction` varchar(20) NOT NULL DEFAULT '' COMMENT '之前方向',
    `new_direction` varchar(20) NOT NULL DEFAULT '' COMMENT '新方向',
    `strength` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '信号强度',
    `confidence` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '置信度',
    `action` varchar(20) NOT NULL DEFAULT '' COMMENT '建议操作',
    `entry_price` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '建议入场价',
    `stop_loss` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '建议止损价',
    `take_profit1` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '止盈目标1',
    `take_profit2` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '止盈目标2',
    `trend_signal` varchar(20) NOT NULL DEFAULT '' COMMENT '趋势信号',
    `momentum_signal` varchar(20) NOT NULL DEFAULT '' COMMENT '动量信号',
    `pattern_signal` varchar(20) NOT NULL DEFAULT '' COMMENT '形态信号',
    `reason` varchar(500) NOT NULL DEFAULT '' COMMENT '信号原因',
    `indicators` json DEFAULT NULL COMMENT '指标数据JSON',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_platform_symbol` (`platform`, `symbol`),
    KEY `idx_new_direction` (`new_direction`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='方向预警日志';

-- 4. 机器人实时状态表（缓存表，用于快速查询）
DROP TABLE IF EXISTS `hg_trading_robot_status`;
CREATE TABLE `hg_trading_robot_status` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `robot_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '机器人ID',
    `user_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '用户ID',
    `platform` varchar(50) NOT NULL DEFAULT '' COMMENT '交易所平台',
    `symbol` varchar(50) NOT NULL DEFAULT '' COMMENT '交易对',
    `current_price` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '当前价格',
    `market_state` varchar(20) NOT NULL DEFAULT '' COMMENT '市场状态',
    `risk_preference` varchar(20) NOT NULL DEFAULT '' COMMENT '风险偏好',
    `win_probability` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '胜算概率',
    `direction_signal` varchar(20) NOT NULL DEFAULT '' COMMENT '方向信号',
    `signal_strength` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '信号强度',
    `signal_confidence` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '信号置信度',
    `has_position` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否有持仓',
    `position_side` varchar(20) NOT NULL DEFAULT '' COMMENT '持仓方向',
    `position_pnl` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '持仓盈亏',
    `account_balance` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '账户余额',
    `available_balance` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '可用余额',
    `last_order_time` datetime DEFAULT NULL COMMENT '最后下单时间',
    `last_close_time` datetime DEFAULT NULL COMMENT '最后平仓时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_robot_id` (`robot_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_market_state` (`market_state`),
    KEY `idx_direction_signal` (`direction_signal`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='机器人实时状态';

-- 5. 添加V2引擎定时任务配置
INSERT INTO `hg_sys_cron` (`group_id`, `title`, `name`, `params`, `pattern`, `policy`, `count`, `sort`, `remark`, `status`, `created_at`, `updated_at`)
SELECT 1, 'Toogo机器人引擎V2', 'toogo_robot_engine_v2', '', '@every 10s', 1, 0, 100, '新版机器人交易引擎，支持多周期分析和智能风险评估', 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM `hg_sys_cron` WHERE `name` = 'toogo_robot_engine_v2');

-- 6. 禁用旧版V1引擎（如果存在）
UPDATE `hg_sys_cron` SET `status` = 2 WHERE `name` = 'toogo_robot_engine' AND `status` = 1;

-- 完成提示
SELECT 'V2 机器人交易系统数据库表创建完成！' AS message;

