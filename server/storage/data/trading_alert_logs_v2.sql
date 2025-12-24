-- =====================================================
-- Toogo V2 预警日志表结构
-- 创建时间: 2024
-- =====================================================

-- 1. 市场状态预警日志表
DROP TABLE IF EXISTS `hg_trading_market_state_log`;
CREATE TABLE `hg_trading_market_state_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `platform` varchar(32) NOT NULL COMMENT '交易所平台',
  `symbol` varchar(32) NOT NULL COMMENT '交易对',
  `prev_state` varchar(32) DEFAULT NULL COMMENT '之前的市场状态',
  `new_state` varchar(32) NOT NULL COMMENT '新的市场状态: trend/volatile/high_vol/low_vol/breakout/reversion',
  `confidence` decimal(10,4) DEFAULT '0.0000' COMMENT '置信度 0-1',
  `trend_strength` decimal(10,4) DEFAULT '0.0000' COMMENT '趋势强度 -1到1',
  `volatility` decimal(10,4) DEFAULT '0.0000' COMMENT '波动率',
  `trend_score` decimal(10,4) DEFAULT '0.0000' COMMENT '趋势评分',
  `momentum_score` decimal(10,4) DEFAULT '0.0000' COMMENT '动量评分',
  `reason` varchar(500) DEFAULT NULL COMMENT '预警原因',
  `indicators` json DEFAULT NULL COMMENT '技术指标JSON',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_platform_symbol` (`platform`, `symbol`),
  KEY `idx_new_state` (`new_state`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='市场状态预警日志';

-- 2. 风险偏好预警日志表
DROP TABLE IF EXISTS `hg_trading_risk_preference_log`;
CREATE TABLE `hg_trading_risk_preference_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `robot_id` bigint(20) NOT NULL COMMENT '机器人ID',
  `user_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '用户ID',
  `platform` varchar(32) NOT NULL COMMENT '交易所平台',
  `symbol` varchar(32) NOT NULL COMMENT '交易对',
  `prev_preference` varchar(32) DEFAULT NULL COMMENT '之前的风险偏好',
  `new_preference` varchar(32) NOT NULL COMMENT '新的风险偏好: conservative/balanced/aggressive',
  `win_probability` decimal(10,4) DEFAULT '0.0000' COMMENT '胜算概率 0-100',
  `market_score` decimal(10,4) DEFAULT '0.0000' COMMENT '市场状态评分',
  `technical_score` decimal(10,4) DEFAULT '0.0000' COMMENT '技术指标评分',
  `account_score` decimal(10,4) DEFAULT '0.0000' COMMENT '账户状况评分',
  `history_score` decimal(10,4) DEFAULT '0.0000' COMMENT '历史表现评分',
  `volatility_risk` decimal(10,4) DEFAULT '0.0000' COMMENT '波动风险',
  `suggested_leverage` int(11) DEFAULT '10' COMMENT '建议杠杆',
  `suggested_margin_percent` decimal(10,4) DEFAULT '10.0000' COMMENT '建议保证金比例',
  `suggested_stop_loss` decimal(10,4) DEFAULT '10.0000' COMMENT '建议止损比例',
  `suggested_take_profit` decimal(10,4) DEFAULT '10.0000' COMMENT '建议止盈比例',
  `reason` varchar(500) DEFAULT NULL COMMENT '预警原因',
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
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `platform` varchar(32) NOT NULL COMMENT '交易所平台',
  `symbol` varchar(32) NOT NULL COMMENT '交易对',
  `prev_direction` varchar(32) DEFAULT NULL COMMENT '之前的方向',
  `new_direction` varchar(32) NOT NULL COMMENT '新的方向: LONG/SHORT/NEUTRAL',
  `strength` decimal(10,4) DEFAULT '0.0000' COMMENT '信号强度 0-100',
  `confidence` decimal(10,4) DEFAULT '0.0000' COMMENT '置信度 0-100',
  `action` varchar(32) DEFAULT NULL COMMENT '建议操作: OPEN_LONG/OPEN_SHORT/CLOSE_LONG/CLOSE_SHORT/HOLD/WAIT',
  `trend_signal` varchar(32) DEFAULT NULL COMMENT '趋势信号',
  `momentum_signal` varchar(32) DEFAULT NULL COMMENT '动量信号',
  `pattern_signal` varchar(32) DEFAULT NULL COMMENT '形态信号',
  `near_support` tinyint(1) DEFAULT '0' COMMENT '是否接近支撑位',
  `near_resistance` tinyint(1) DEFAULT '0' COMMENT '是否接近阻力位',
  `entry_price` decimal(20,8) DEFAULT '0.00000000' COMMENT '建议入场价',
  `stop_loss` decimal(20,8) DEFAULT '0.00000000' COMMENT '建议止损价',
  `take_profit_1` decimal(20,8) DEFAULT '0.00000000' COMMENT '止盈目标1',
  `take_profit_2` decimal(20,8) DEFAULT '0.00000000' COMMENT '止盈目标2',
  `reason` varchar(500) DEFAULT NULL COMMENT '预警原因',
  `timeframe_signals` json DEFAULT NULL COMMENT '各周期信号JSON',
  `indicators` json DEFAULT NULL COMMENT '技术指标JSON',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_platform_symbol` (`platform`, `symbol`),
  KEY `idx_new_direction` (`new_direction`),
  KEY `idx_action` (`action`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='方向预警日志';

-- 4. 机器人实时状态表（用于缓存当前状态，供API快速查询）
DROP TABLE IF EXISTS `hg_trading_robot_realtime`;
CREATE TABLE `hg_trading_robot_realtime` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `robot_id` bigint(20) NOT NULL COMMENT '机器人ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `platform` varchar(32) NOT NULL COMMENT '交易所平台',
  `symbol` varchar(32) NOT NULL COMMENT '交易对',
  
  -- 当前价格
  `current_price` decimal(20,8) DEFAULT '0.00000000' COMMENT '当前价格',
  `price_change_24h` decimal(10,4) DEFAULT '0.0000' COMMENT '24h涨跌幅',
  
  -- 市场状态
  `market_state` varchar(32) DEFAULT NULL COMMENT '市场状态',
  `market_state_conf` decimal(10,4) DEFAULT '0.0000' COMMENT '市场状态置信度',
  `trend_strength` decimal(10,4) DEFAULT '0.0000' COMMENT '趋势强度',
  `volatility` decimal(10,4) DEFAULT '0.0000' COMMENT '波动率',
  
  -- 风险评估
  `risk_preference` varchar(32) DEFAULT NULL COMMENT '风险偏好',
  `win_probability` decimal(10,4) DEFAULT '0.0000' COMMENT '胜算概率',
  `risk_level` int(11) DEFAULT '0' COMMENT '风险等级1-5',
  
  -- 方向信号
  `direction` varchar(32) DEFAULT NULL COMMENT '方向信号',
  `direction_strength` decimal(10,4) DEFAULT '0.0000' COMMENT '方向强度',
  `direction_confidence` decimal(10,4) DEFAULT '0.0000' COMMENT '方向置信度',
  `suggested_action` varchar(32) DEFAULT NULL COMMENT '建议操作',
  
  -- 持仓信息
  `has_position` tinyint(1) DEFAULT '0' COMMENT '是否有持仓',
  `position_side` varchar(32) DEFAULT NULL COMMENT '持仓方向',
  `position_amt` decimal(20,8) DEFAULT '0.00000000' COMMENT '持仓数量',
  `position_pnl` decimal(20,8) DEFAULT '0.00000000' COMMENT '持仓盈亏',
  `position_pnl_percent` decimal(10,4) DEFAULT '0.0000' COMMENT '持仓盈亏比例',
  
  -- 账户余额
  `account_balance` decimal(20,8) DEFAULT '0.00000000' COMMENT '账户余额',
  `available_balance` decimal(20,8) DEFAULT '0.00000000' COMMENT '可用余额',
  
  -- 更新时间
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_robot_id` (`robot_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_platform_symbol` (`platform`, `symbol`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='机器人实时状态';

-- 5. 添加定时任务配置
INSERT INTO `hg_sys_cron` (`group_id`, `title`, `name`, `params`, `pattern`, `policy`, `count`, `sort`, `remark`, `status`, `created_at`, `updated_at`) 
VALUES 
(1, 'Toogo机器人引擎V2', 'toogo_robot_engine_v2', '', '@every 60s', 1, 0, 100, '新架构机器人引擎，只需启动一次，内部自动循环', 2, NOW(), NOW())
ON DUPLICATE KEY UPDATE `title`='Toogo机器人引擎V2', `params`='', `pattern`='@every 60s', `remark`='新架构机器人引擎，只需启动一次，内部自动循环';

-- 6. 禁用旧的V1引擎（如果存在）
UPDATE `hg_sys_cron` SET `status` = 1 WHERE `name` = 'toogo_robot_engine' AND `name` != 'toogo_robot_engine_v2';

-- 完成提示
SELECT 'V2 预警日志表创建完成!' AS message;

