-- 预警日志表结构
-- 执行命令: mysql -uroot -proot hotgo < create_alert_logs_tables.sql

-- 1. 市场状态预警日志表
DROP TABLE IF EXISTS `hg_trading_market_state_log`;
CREATE TABLE `hg_trading_market_state_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `platform` varchar(50) NOT NULL DEFAULT '' COMMENT '交易所平台',
  `symbol` varchar(50) NOT NULL DEFAULT '' COMMENT '交易对',
  `prev_state` varchar(30) NOT NULL DEFAULT '' COMMENT '之前状态',
  `new_state` varchar(30) NOT NULL DEFAULT '' COMMENT '新状态',
  `confidence` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '置信度',
  `reason` varchar(500) NOT NULL DEFAULT '' COMMENT '变化原因',
  `trend_strength` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '趋势强度',
  `volatility` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '波动率',
  `trend_score` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '趋势评分',
  `momentum_score` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '动量评分',
  `indicators` json DEFAULT NULL COMMENT '技术指标JSON',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_platform_symbol` (`platform`, `symbol`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_new_state` (`new_state`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='市场状态预警日志';

-- 2. 风险偏好预警日志表
DROP TABLE IF EXISTS `hg_trading_risk_preference_log`;
CREATE TABLE `hg_trading_risk_preference_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `robot_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '机器人ID',
  `user_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '用户ID',
  `platform` varchar(50) NOT NULL DEFAULT '' COMMENT '交易所平台',
  `symbol` varchar(50) NOT NULL DEFAULT '' COMMENT '交易对',
  `prev_preference` varchar(30) NOT NULL DEFAULT '' COMMENT '之前偏好',
  `new_preference` varchar(30) NOT NULL DEFAULT '' COMMENT '新偏好',
  `win_probability` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '胜算概率',
  `market_score` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '市场评分',
  `technical_score` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '技术评分',
  `account_score` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '账户评分',
  `history_score` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '历史评分',
  `volatility_risk` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '波动风险',
  `reason` varchar(500) NOT NULL DEFAULT '' COMMENT '变化原因',
  `suggested_leverage` int(11) NOT NULL DEFAULT '0' COMMENT '建议杠杆',
  `suggested_margin_percent` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '建议保证金比例',
  `factors` json DEFAULT NULL COMMENT '评估因子JSON',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_robot_id` (`robot_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_new_preference` (`new_preference`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='风险偏好预警日志';

-- 3. 方向预警日志表
DROP TABLE IF EXISTS `hg_trading_direction_log`;
CREATE TABLE `hg_trading_direction_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `platform` varchar(50) NOT NULL DEFAULT '' COMMENT '交易所平台',
  `symbol` varchar(50) NOT NULL DEFAULT '' COMMENT '交易对',
  `prev_direction` varchar(30) NOT NULL DEFAULT '' COMMENT '之前方向',
  `new_direction` varchar(30) NOT NULL DEFAULT '' COMMENT '新方向',
  `strength` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '信号强度',
  `confidence` decimal(10,4) NOT NULL DEFAULT '0.0000' COMMENT '置信度',
  `action` varchar(30) NOT NULL DEFAULT '' COMMENT '建议操作',
  `reason` varchar(500) NOT NULL DEFAULT '' COMMENT '信号原因',
  `trend_signal` varchar(30) NOT NULL DEFAULT '' COMMENT '趋势信号',
  `momentum_signal` varchar(30) NOT NULL DEFAULT '' COMMENT '动量信号',
  `pattern_signal` varchar(30) NOT NULL DEFAULT '' COMMENT '形态信号',
  `entry_price` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '建议入场价',
  `stop_loss` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '建议止损价',
  `take_profit` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '建议止盈价',
  `indicators` json DEFAULT NULL COMMENT '指标JSON',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_platform_symbol` (`platform`, `symbol`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_new_direction` (`new_direction`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='方向预警日志';

-- 4. 添加V2引擎定时任务配置
INSERT INTO `hg_sys_cron` (`group_id`, `title`, `name`, `params`, `pattern`, `policy`, `count`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(1, 'Toogo机器人引擎V2', 'toogo_robot_engine_v2', '', '@every 60s', 1, 0, 100, '新架构机器人引擎，启动后自动循环执行', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE `title` = VALUES(`title`), `updated_at` = NOW();

-- 5. 禁用旧版引擎（可选，保留作为备用）
-- UPDATE `hg_sys_cron` SET `status` = 2 WHERE `name` = 'toogo_robot_engine';

SELECT '预警日志表创建完成!' AS message;

