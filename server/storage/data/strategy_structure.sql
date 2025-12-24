-- ============================================================
-- 策略管理重构：模板 + 策略 两层结构
-- ============================================================

-- 1. 创建策略模板表（顶层）
CREATE TABLE IF NOT EXISTS `hg_trading_strategy_group` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `group_name` varchar(100) NOT NULL COMMENT '模板名称，如：BTC-USDT官方策略',
  `group_key` varchar(50) NOT NULL COMMENT '模板标识，唯一',
  `exchange` varchar(20) NOT NULL DEFAULT 'bitget' COMMENT '交易平台',
  `symbol` varchar(20) NOT NULL COMMENT '交易对，如BTC-USDT',
  `order_type` varchar(20) NOT NULL DEFAULT 'market' COMMENT '订单类型',
  `margin_mode` varchar(20) NOT NULL DEFAULT 'isolated' COMMENT '保证金模式',
  `is_official` tinyint(1) DEFAULT 0 COMMENT '是否官方模板',
  `from_official_id` bigint DEFAULT 0 COMMENT '来源官方模板ID（复制自哪个官方模板）',
  `is_default` tinyint(1) DEFAULT 0 COMMENT '是否默认策略',
  `user_id` bigint DEFAULT 0 COMMENT '创建用户ID，0为系统',
  `description` varchar(500) DEFAULT NULL COMMENT '模板描述',
  `is_active` tinyint(1) DEFAULT 1 COMMENT '是否启用',
  `sort` int DEFAULT 100 COMMENT '排序',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_key` (`group_key`),
  KEY `idx_exchange_symbol` (`exchange`, `symbol`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_from_official` (`from_official_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='策略模板表';

-- 2. 修改策略表，添加 group_id 关联
ALTER TABLE `hg_trading_strategy_template` 
ADD COLUMN `group_id` bigint DEFAULT 0 COMMENT '所属策略模板ID' AFTER `id`,
ADD KEY `idx_group_id` (`group_id`);

-- 3. 插入官方BTC-USDT策略模板
INSERT INTO `hg_trading_strategy_group` (
  `group_name`, `group_key`, `exchange`, `symbol`, `order_type`, `margin_mode`,
  `is_official`, `user_id`, `description`, `is_active`, `sort`
) VALUES (
  'BTC-USDT 官方策略模板',
  'official_bitget_btc_usdt',
  'bitget',
  'BTC-USDT',
  'market',
  'isolated',
  1,
  0,
  'Bitget平台BTC-USDT官方推荐策略模板，包含12种市场状态与风险偏好组合，由专业量化团队调优。',
  1,
  1
);

-- 4. 更新现有策略关联到官方模板
SET @group_id = (SELECT id FROM hg_trading_strategy_group WHERE group_key = 'official_bitget_btc_usdt');
UPDATE hg_trading_strategy_template SET group_id = @group_id WHERE strategy_key LIKE 'conservative_%' OR strategy_key LIKE 'balanced_%' OR strategy_key LIKE 'aggressive_%';

-- 5. 查询验证
SELECT g.group_name, g.symbol, COUNT(s.id) as strategy_count
FROM hg_trading_strategy_group g
LEFT JOIN hg_trading_strategy_template s ON s.group_id = g.id
GROUP BY g.id;

