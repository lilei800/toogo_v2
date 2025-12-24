-- =============================================
-- Toogo Trading System Database Tables
-- 创建时间: 2024-11-26
-- 说明: 量化交易机器人SaaS平台数据库表
-- =============================================

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

-- --------------------------------------------------------

--
-- 表的结构 `hg_trading_api_config`
-- API接口配置表
--

CREATE TABLE IF NOT EXISTS `hg_trading_api_config` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '租户ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  
  -- 基础信息
  `api_name` varchar(100) NOT NULL COMMENT 'API接口名称',
  `platform` varchar(50) NOT NULL COMMENT '平台名称：bitget/binance/okx',
  `base_url` varchar(255) NOT NULL COMMENT 'API地址',
  
  -- 密钥信息（加密存储）
  `api_key` varchar(500) NOT NULL COMMENT 'API Key（加密）',
  `secret_key` varchar(500) NOT NULL COMMENT 'Secret Key（加密）',
  `passphrase` varchar(500) DEFAULT NULL COMMENT 'Passphrase（加密，可选）',
  
  -- 状态
  `is_default` tinyint(1) DEFAULT '0' COMMENT '是否默认：0=否,1=是',
  `status` tinyint(2) DEFAULT '1' COMMENT '状态：1=正常,2=禁用',
  
  -- 验证信息
  `last_verify_time` datetime DEFAULT NULL COMMENT '最后验证时间',
  `verify_status` tinyint(2) DEFAULT '0' COMMENT '验证状态：0=未验证,1=成功,2=失败',
  `verify_message` varchar(500) DEFAULT NULL COMMENT '验证消息',
  
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  
  PRIMARY KEY (`id`),
  KEY `idx_tenant_user` (`tenant_id`, `user_id`),
  KEY `idx_platform` (`platform`),
  KEY `idx_is_default` (`is_default`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API接口配置表';

-- --------------------------------------------------------

--
-- 表的结构 `hg_trading_proxy_config`
-- 代理配置表
--

CREATE TABLE IF NOT EXISTS `hg_trading_proxy_config` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '租户ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  
  -- 代理配置
  `enabled` tinyint(1) DEFAULT '0' COMMENT '是否启用：0=禁用,1=启用',
  `proxy_type` varchar(20) DEFAULT 'socks5' COMMENT '代理类型：socks5/http',
  `proxy_address` varchar(255) DEFAULT '127.0.0.1:10808' COMMENT '代理地址',
  
  -- 认证（可选）
  `auth_enabled` tinyint(1) DEFAULT '0' COMMENT '是否需要认证',
  `username` varchar(100) DEFAULT NULL COMMENT '用户名',
  `password` varchar(500) DEFAULT NULL COMMENT '密码（加密）',
  
  -- 测试状态
  `last_test_time` datetime DEFAULT NULL COMMENT '最后测试时间',
  `test_status` tinyint(2) DEFAULT '0' COMMENT '测试状态：0=未测试,1=成功,2=失败',
  `test_message` varchar(500) DEFAULT NULL COMMENT '测试消息',
  
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_user` (`tenant_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='代理配置表';

-- --------------------------------------------------------

--
-- 表的结构 `hg_trading_robot`
-- 交易机器人表
--

CREATE TABLE IF NOT EXISTS `hg_trading_robot` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '租户ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  
  -- ①基础信息
  `robot_name` varchar(100) NOT NULL COMMENT '机器人名称',
  `api_config_id` bigint(20) NOT NULL COMMENT 'API接口ID',
  `max_profit_target` decimal(15,2) DEFAULT '0.00' COMMENT '最大盈利目标(USDT)',
  `max_loss_amount` decimal(15,2) DEFAULT '0.00' COMMENT '最大亏损额(USDT)',
  `max_runtime` int(11) DEFAULT '0' COMMENT '最大运行时长(秒)',
  
  -- ②风险偏好
  `risk_preference` varchar(20) NOT NULL DEFAULT 'balanced' COMMENT '风险偏好：conservative/balanced/aggressive',
  `auto_risk_preference` tinyint(1) DEFAULT '0' COMMENT '自动风险偏好：0=手动,1=自动',
  
  -- ③市场行情
  `market_state` varchar(20) NOT NULL DEFAULT 'trend' COMMENT '市场状态：trend/volatile/high-volatility/low-volatility',
  `auto_market_state` tinyint(1) DEFAULT '0' COMMENT '自动市场状态：0=手动,1=自动',
  
  -- ④下单配置
  `exchange` varchar(20) NOT NULL DEFAULT 'bitget' COMMENT '交易所',
  `symbol` varchar(20) NOT NULL DEFAULT 'BTC_USDT' COMMENT '交易对',
  `order_type` varchar(20) DEFAULT 'market' COMMENT '订单类型：market/limit',
  `margin_mode` varchar(20) DEFAULT 'isolated' COMMENT '保证金模式：isolated/cross',
  `leverage` int(11) NOT NULL DEFAULT '10' COMMENT '杠杆倍数',
  `margin_percent` decimal(5,2) NOT NULL DEFAULT '30.00' COMMENT '使用保证金比例(%)',
  `use_monitor_signal` tinyint(1) DEFAULT '1' COMMENT '采用方向预警信号：0=否,1=是',
  `enable_reverse_order` tinyint(1) DEFAULT '0' COMMENT '启用反方向下单：0=否,1=是',
  
  -- ⑤自动平仓配置
  `stop_loss_percent` decimal(5,2) NOT NULL DEFAULT '10.00' COMMENT '止损百分比(%)',
  `profit_retreat_percent` decimal(5,2) NOT NULL DEFAULT '18.00' COMMENT '止盈回撤百分比(%)',
  `auto_start_retreat_percent` decimal(5,2) NOT NULL DEFAULT '8.00' COMMENT '启动回撤百分比(%)',
  
  -- 实时策略参数（根据风险偏好和市场状态动态加载）
  `current_strategy` text COMMENT '当前策略配置(JSON)',
  
  -- 状态信息
  `status` tinyint(2) DEFAULT '1' COMMENT '状态：1=未启动,2=运行中,3=暂停,4=停用',
  `start_time` datetime DEFAULT NULL COMMENT '启动时间',
  `pause_time` datetime DEFAULT NULL COMMENT '暂停时间',
  `stop_time` datetime DEFAULT NULL COMMENT '停止时间',
  
  -- 统计数据
  `long_count` int(11) DEFAULT '0' COMMENT '多单数',
  `short_count` int(11) DEFAULT '0' COMMENT '空单数',
  `total_profit` decimal(15,4) DEFAULT '0.0000' COMMENT '总盈亏(USDT)',
  `runtime_seconds` int(11) DEFAULT '0' COMMENT '已运行时长(秒)',
  
  -- 开关
  `auto_trade_enabled` tinyint(1) DEFAULT '0' COMMENT '全自动下单：0=否,1=是',
  `auto_close_enabled` tinyint(1) DEFAULT '1' COMMENT '全自动平仓：0=否,1=是',
  
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  
  PRIMARY KEY (`id`),
  KEY `idx_tenant_user` (`tenant_id`, `user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_api_config` (`api_config_id`),
  KEY `idx_symbol` (`symbol`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='交易机器人表';

-- --------------------------------------------------------

--
-- 表的结构 `hg_trading_order`
-- 交易订单表
--

CREATE TABLE IF NOT EXISTS `hg_trading_order` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '租户ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `robot_id` bigint(20) NOT NULL COMMENT '机器人ID',
  
  -- 订单信息
  `order_sn` varchar(64) NOT NULL COMMENT '订单号',
  `exchange_order_id` varchar(100) DEFAULT NULL COMMENT '交易所订单ID',
  `symbol` varchar(20) NOT NULL COMMENT '交易对',
  `direction` varchar(10) NOT NULL COMMENT '方向：long/short',
  
  -- 交易详情
  `open_price` decimal(15,4) NOT NULL COMMENT '开仓价格',
  `close_price` decimal(15,4) DEFAULT NULL COMMENT '平仓价格',
  `quantity` decimal(15,8) NOT NULL COMMENT '数量',
  `leverage` int(11) NOT NULL COMMENT '杠杆倍数',
  `margin` decimal(15,4) NOT NULL COMMENT '保证金(USDT)',
  
  -- 盈亏信息
  `realized_profit` decimal(15,4) DEFAULT '0.0000' COMMENT '已实现盈亏',
  `unrealized_profit` decimal(15,4) DEFAULT '0.0000' COMMENT '未实现盈亏',
  `highest_profit` decimal(15,4) DEFAULT '0.0000' COMMENT '最高盈利',
  
  -- 风控信息
  `stop_loss_price` decimal(15,4) DEFAULT NULL COMMENT '止损价格',
  `profit_retreat_started` tinyint(1) DEFAULT '0' COMMENT '止盈回撤已启动',
  `profit_retreat_percent` decimal(5,2) DEFAULT NULL COMMENT '止盈回撤百分比',
  
  -- 时间信息
  `open_time` datetime NOT NULL COMMENT '开仓时间',
  `close_time` datetime DEFAULT NULL COMMENT '平仓时间',
  `hold_duration` int(11) DEFAULT '0' COMMENT '持仓时长(秒)',
  
  -- 状态
  `status` tinyint(2) DEFAULT '1' COMMENT '状态：1=持仓中,2=已平仓,3=已取消',
  `close_reason` varchar(50) DEFAULT NULL COMMENT '平仓原因：stop_loss/take_profit/manual/timeout',
  
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_sn` (`order_sn`),
  KEY `idx_tenant_user` (`tenant_id`, `user_id`),
  KEY `idx_robot_id` (`robot_id`),
  KEY `idx_status` (`status`),
  KEY `idx_symbol` (`symbol`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='交易订单表';

-- --------------------------------------------------------

--
-- 表的结构 `hg_trading_close_log`
-- 平仓日志表
--

CREATE TABLE IF NOT EXISTS `hg_trading_close_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '租户ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `robot_id` bigint(20) NOT NULL COMMENT '机器人ID',
  `order_id` bigint(20) NOT NULL COMMENT '订单ID',
  
  -- 订单信息
  `order_sn` varchar(64) NOT NULL COMMENT '订单号',
  `symbol` varchar(20) NOT NULL COMMENT '交易对',
  `direction` varchar(10) NOT NULL COMMENT '方向：long/short',
  
  -- 交易详情
  `open_price` decimal(15,4) NOT NULL COMMENT '开仓价格',
  `close_price` decimal(15,4) NOT NULL COMMENT '平仓价格',
  `quantity` decimal(15,8) NOT NULL COMMENT '数量',
  `leverage` int(11) NOT NULL COMMENT '杠杆倍数',
  `margin` decimal(15,4) NOT NULL COMMENT '保证金(USDT)',
  
  -- 盈亏信息
  `realized_profit` decimal(15,4) NOT NULL COMMENT '已实现盈亏',
  `highest_profit` decimal(15,4) DEFAULT '0.0000' COMMENT '最高盈利',
  `profit_percent` decimal(8,4) DEFAULT '0.0000' COMMENT '盈利百分比',
  
  -- 平仓原因
  `close_reason` varchar(50) NOT NULL COMMENT '平仓原因',
  `close_detail` text COMMENT '平仓详情(JSON)',
  
  -- 费用信息
  `open_fee` decimal(15,4) DEFAULT '0.0000' COMMENT '开仓费用',
  `hold_fee` decimal(15,4) DEFAULT '0.0000' COMMENT '持仓费用',
  `close_fee` decimal(15,4) DEFAULT '0.0000' COMMENT '平仓费用',
  `total_fee` decimal(15,4) DEFAULT '0.0000' COMMENT '总费用',
  
  -- 佣金信息
  `commission_amount` decimal(15,4) DEFAULT '0.0000' COMMENT '佣金金额',
  `commission_percent` decimal(5,2) DEFAULT '0.00' COMMENT '佣金比例',
  
  -- 净利润
  `net_profit` decimal(15,4) DEFAULT '0.0000' COMMENT '净利润',
  
  -- 时间信息
  `open_time` datetime NOT NULL COMMENT '开仓时间',
  `close_time` datetime NOT NULL COMMENT '平仓时间',
  `hold_duration` int(11) DEFAULT '0' COMMENT '持仓时长(秒)',
  
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  
  PRIMARY KEY (`id`),
  KEY `idx_tenant_user` (`tenant_id`, `user_id`),
  KEY `idx_robot_id` (`robot_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_close_time` (`close_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='平仓日志表';

-- --------------------------------------------------------

--
-- 表的结构 `hg_trading_strategy_template`
-- 策略模板表
--

CREATE TABLE IF NOT EXISTS `hg_trading_strategy_template` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  
  -- 策略标识
  `strategy_key` varchar(50) NOT NULL COMMENT '策略KEY：conservative_trend',
  `strategy_name` varchar(100) NOT NULL COMMENT '策略名称',
  `risk_preference` varchar(20) NOT NULL COMMENT '风险偏好：conservative/balanced/aggressive',
  `market_state` varchar(20) NOT NULL COMMENT '市场状态：trend/volatile/high-volatility/low-volatility',
  
  -- 下单参数
  `monitor_window` int(11) NOT NULL COMMENT '监控时间窗口(秒)',
  `volatility_threshold` decimal(10,2) NOT NULL COMMENT '波动阈值(USDT)',
  `leverage_min` int(11) NOT NULL COMMENT '杠杆倍数最小值',
  `leverage_max` int(11) NOT NULL COMMENT '杠杆倍数最大值',
  `margin_percent_min` decimal(5,2) NOT NULL COMMENT '保证金比例最小值(%)',
  `margin_percent_max` decimal(5,2) NOT NULL COMMENT '保证金比例最大值(%)',
  
  -- 平仓参数
  `stop_loss_percent` decimal(5,2) NOT NULL COMMENT '止损百分比(%)',
  `profit_retreat_percent` decimal(5,2) NOT NULL COMMENT '止盈回撤百分比(%)',
  `auto_start_retreat_percent` decimal(5,2) NOT NULL COMMENT '启动回撤百分比(%)',
  
  -- 其他配置
  `config_json` text COMMENT '其他配置(JSON)',
  `description` varchar(500) DEFAULT NULL COMMENT '策略描述',
  
  -- 状态
  `is_active` tinyint(1) DEFAULT '1' COMMENT '是否激活',
  `sort` int(11) DEFAULT '100' COMMENT '排序',
  
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_strategy_key` (`strategy_key`),
  KEY `idx_risk_market` (`risk_preference`, `market_state`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='策略模板表';

-- --------------------------------------------------------

--
-- 表的结构 `hg_trading_monitor_log`
-- 市场监控日志表
--

CREATE TABLE IF NOT EXISTS `hg_trading_monitor_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '租户ID',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `robot_id` bigint(20) NOT NULL COMMENT '机器人ID',
  
  -- 监控信息
  `symbol` varchar(20) NOT NULL COMMENT '交易对',
  `current_price` decimal(15,4) NOT NULL COMMENT '当前价格',
  `window_high` decimal(15,4) DEFAULT NULL COMMENT '窗口最高价',
  `window_low` decimal(15,4) DEFAULT NULL COMMENT '窗口最低价',
  `volatility` decimal(10,2) DEFAULT NULL COMMENT '波动值',
  
  -- 信号信息
  `signal_type` varchar(20) DEFAULT NULL COMMENT '信号类型：buy/sell/hold',
  `signal_strength` decimal(5,2) DEFAULT NULL COMMENT '信号强度(0-100)',
  `market_state` varchar(20) DEFAULT NULL COMMENT '市场状态',
  
  `signal_detail` text COMMENT '信号详情(JSON)',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  
  PRIMARY KEY (`id`),
  KEY `idx_robot_id` (`robot_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='市场监控日志表';

-- --------------------------------------------------------

--
-- 插入默认策略模板数据
--

INSERT INTO `hg_trading_strategy_template` (`strategy_key`, `strategy_name`, `risk_preference`, `market_state`, `monitor_window`, `volatility_threshold`, `leverage_min`, `leverage_max`, `margin_percent_min`, `margin_percent_max`, `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`, `description`, `is_active`, `sort`) VALUES
-- 保守型策略
('conservative_trend', '保守型-趋势市场', 'conservative', 'trend', 180, 100.00, 6, 8, 20.00, 25.00, 10.00, 20.00, 8.00, '适合新手，稳健盈利，风险较低', 1, 10),
('conservative_volatile', '保守型-震荡市场', 'conservative', 'volatile', 300, 120.00, 6, 8, 20.00, 25.00, 10.00, 22.00, 8.00, '震荡市场保守策略，降低交易频率', 1, 11),
('conservative_high_vol', '保守型-高波动', 'conservative', 'high-volatility', 360, 150.00, 5, 7, 18.00, 22.00, 12.00, 25.00, 10.00, '高波动市场保守策略，风控优先', 1, 12),
('conservative_low_vol', '保守型-低波动', 'conservative', 'low-volatility', 120, 80.00, 7, 9, 22.00, 27.00, 8.00, 18.00, 6.00, '低波动市场保守策略，小额多次', 1, 13),

-- 平衡型策略
('balanced_trend', '平衡型-趋势市场', 'balanced', 'trend', 300, 75.00, 10, 10, 30.00, 35.00, 8.00, 18.00, 8.00, '平衡风险收益，适合大多数用户', 1, 20),
('balanced_volatile', '平衡型-震荡市场', 'balanced', 'volatile', 240, 90.00, 10, 10, 30.00, 35.00, 8.00, 20.00, 8.00, '震荡市场平衡策略，灵活应对', 1, 21),
('balanced_high_vol', '平衡型-高波动', 'balanced', 'high-volatility', 300, 120.00, 8, 10, 28.00, 33.00, 10.00, 22.00, 10.00, '高波动市场平衡策略', 1, 22),
('balanced_low_vol', '平衡型-低波动', 'balanced', 'low-volatility', 180, 60.00, 10, 12, 32.00, 37.00, 7.00, 15.00, 6.00, '低波动市场平衡策略', 1, 23),

-- 激进型策略
('aggressive_trend', '激进型-趋势市场', 'aggressive', 'trend', 120, 50.00, 12, 15, 40.00, 45.00, 8.00, 15.00, 5.00, '高风险高收益，适合专业用户', 1, 30),
('aggressive_volatile', '激进型-震荡市场', 'aggressive', 'volatile', 180, 70.00, 12, 15, 40.00, 45.00, 8.00, 18.00, 5.00, '震荡市场激进策略，快进快出', 1, 31),
('aggressive_high_vol', '激进型-高波动', 'aggressive', 'high-volatility', 240, 100.00, 10, 13, 35.00, 40.00, 10.00, 20.00, 8.00, '高波动市场激进策略，控制风险', 1, 32),
('aggressive_low_vol', '激进型-低波动', 'aggressive', 'low-volatility', 90, 40.00, 13, 16, 42.00, 48.00, 6.00, 12.00, 4.00, '低波动市场激进策略，最大化收益', 1, 33);

-- =============================================
-- 数据库表创建完成
-- =============================================

