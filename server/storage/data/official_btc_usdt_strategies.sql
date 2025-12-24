-- ============================================================
-- 官方 BTC-USDT 策略模板 (Bitget)
-- 创建时间: 2024-11-29
-- 更新时间: 2024-11-29
-- 说明: 经过专业量化团队调优的12种策略组合
-- 包含完整的手动策略配置参数
-- ============================================================

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET NAMES utf8mb4;

-- ============================================================
-- 1. 确保策略组表存在
-- ============================================================

CREATE TABLE IF NOT EXISTS `hg_trading_strategy_group` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `group_name` varchar(100) NOT NULL COMMENT '模板名称',
  `group_key` varchar(50) NOT NULL COMMENT '模板标识',
  `exchange` varchar(20) NOT NULL DEFAULT 'bitget' COMMENT '交易平台',
  `symbol` varchar(20) NOT NULL COMMENT '交易对',
  `order_type` varchar(20) NOT NULL DEFAULT 'market' COMMENT '订单类型',
  `margin_mode` varchar(20) NOT NULL DEFAULT 'isolated' COMMENT '保证金模式',
  `is_official` tinyint(1) DEFAULT 0 COMMENT '是否官方模板：0=否,1=是',
  `user_id` bigint DEFAULT 0 COMMENT '创建用户ID，0=系统',
  `description` varchar(500) DEFAULT NULL COMMENT '模板描述',
  `is_active` tinyint(1) DEFAULT 1 COMMENT '是否启用',
  `sort` int DEFAULT 100 COMMENT '排序',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_key` (`group_key`),
  KEY `idx_exchange_symbol` (`exchange`, `symbol`),
  KEY `idx_is_official` (`is_official`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='策略模板表';

-- ============================================================
-- 2. 确保策略表有 group_id 字段
-- ============================================================

-- 添加 group_id 字段（如果不存在）
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
     WHERE TABLE_SCHEMA = DATABASE() 
     AND TABLE_NAME = 'hg_trading_strategy_template' 
     AND COLUMN_NAME = 'group_id') = 0,
    'ALTER TABLE `hg_trading_strategy_template` ADD COLUMN `group_id` bigint DEFAULT 0 COMMENT "所属策略模板ID" AFTER `id`, ADD KEY `idx_group_id` (`group_id`)',
    'SELECT 1'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================================
-- 3. 清理旧的官方策略数据（如果存在）
-- ============================================================

DELETE FROM `hg_trading_strategy_template` WHERE `strategy_key` LIKE 'official_btc_%';
DELETE FROM `hg_trading_strategy_group` WHERE `group_key` = 'official_bitget_btc_usdt';

-- ============================================================
-- 4. 插入官方 BTC-USDT 策略组
-- ============================================================

INSERT INTO `hg_trading_strategy_group` (
  `group_name`, `group_key`, `exchange`, `symbol`, `order_type`, `margin_mode`,
  `is_official`, `user_id`, `description`, `is_active`, `sort`
) VALUES (
  '🔥 BTC-USDT 官方推荐策略',
  'official_bitget_btc_usdt',
  'bitget',
  'BTC-USDT',
  'market',
  'isolated',
  1,
  0,
  '由Toogo专业量化团队精心调优的BTC-USDT策略模板，覆盖4种市场状态×3种风险偏好共12种策略组合。适用于Bitget合约交易，参数经过大量历史数据回测验证。',
  1,
  1
);

SET @group_id = LAST_INSERT_ID();

-- ============================================================
-- 5. 插入12种官方策略（包含完整手动配置参数）
-- ============================================================

-- =========================================
-- 🛡️ 保守型策略 (Conservative)
-- 特点：低杠杆、小仓位、宽止损、适合新手
-- =========================================

-- 【保守型-趋势市场】
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_conservative_trend', '🛡️ 保守型-趋势市场',
  'conservative', 'trend',
  300, 80.00,
  3, 5,
  5.00, 10.00,
  3.00, 30.00, 2.00,
  '{
    "exchange": "bitget",
    "symbol": "BTC-USDT",
    "orderType": "market",
    "marginMode": "isolated",
    "monitorWindow": 300,
    "volatilityThreshold": 80,
    "leverage": 4,
    "marginPercent": 8,
    "stopLossPercent": 3,
    "autoStartRetreatPercent": 2,
    "profitRetreatPercent": 30,
    "reverseEnabled": true,
    "reverseLossRatio": 50,
    "reverseProfitRatio": 100,
    "trailingStop": false,
    "remark": "趋势明确时顺势而为，严格止损保护本金"
  }',
  '适合趋势明确的单边行情，顺势交易，止损严格。推荐新手使用，日收益预期0.5-2%。',
  1, 101
);

-- 【保守型-震荡市场】
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_conservative_volatile', '🛡️ 保守型-震荡市场',
  'conservative', 'volatile',
  180, 50.00,
  3, 4,
  4.00, 8.00,
  2.50, 25.00, 1.50,
  '{
    "exchange": "bitget",
    "symbol": "BTC-USDT",
    "orderType": "market",
    "marginMode": "isolated",
    "monitorWindow": 180,
    "volatilityThreshold": 50,
    "leverage": 3,
    "marginPercent": 6,
    "stopLossPercent": 2.5,
    "autoStartRetreatPercent": 1.5,
    "profitRetreatPercent": 25,
    "reverseEnabled": false,
    "reverseLossRatio": 0,
    "reverseProfitRatio": 0,
    "trailingStop": false,
    "remark": "震荡市场减少交易，等待明确信号"
  }',
  '震荡市场减少交易频率，等待箱体突破。不开启反向单，避免来回止损。',
  1, 102
);

-- 【保守型-高波动】
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_conservative_high_vol', '🛡️ 保守型-高波动',
  'conservative', 'high_vol',
  120, 150.00,
  2, 3,
  3.00, 6.00,
  5.00, 35.00, 3.00,
  '{
    "exchange": "bitget",
    "symbol": "BTC-USDT",
    "orderType": "market",
    "marginMode": "isolated",
    "monitorWindow": 120,
    "volatilityThreshold": 150,
    "leverage": 2,
    "marginPercent": 5,
    "stopLossPercent": 5,
    "autoStartRetreatPercent": 3,
    "profitRetreatPercent": 35,
    "reverseEnabled": true,
    "reverseLossRatio": 100,
    "reverseProfitRatio": 100,
    "trailingStop": true,
    "remark": "高波动时期最小仓位，快进快出"
  }',
  '高波动市场风险极高，使用最小杠杆和仓位。启用反向单捕捉双向波动。',
  1, 103
);

-- 【保守型-低波动】
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_conservative_low_vol', '🛡️ 保守型-低波动',
  'conservative', 'low_vol',
  600, 30.00,
  4, 6,
  6.00, 12.00,
  2.00, 20.00, 1.00,
  '{
    "exchange": "bitget",
    "symbol": "BTC-USDT",
    "orderType": "market",
    "marginMode": "isolated",
    "monitorWindow": 600,
    "volatilityThreshold": 30,
    "leverage": 5,
    "marginPercent": 10,
    "stopLossPercent": 2,
    "autoStartRetreatPercent": 1,
    "profitRetreatPercent": 20,
    "reverseEnabled": false,
    "reverseLossRatio": 0,
    "reverseProfitRatio": 0,
    "trailingStop": false,
    "remark": "低波动时可适当加仓，小额多次"
  }',
  '低波动市场波动有限，可适当增加杠杆。主要赚取窄幅波动收益。',
  1, 104
);

-- =========================================
-- ⚖️ 平衡型策略 (Balanced)
-- 特点：中等杠杆、中等仓位、适合大多数用户
-- =========================================

-- 【平衡型-趋势市场】★推荐★
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_balanced_trend', '⚖️ 平衡型-趋势市场 ★推荐',
  'balanced', 'trend',
  240, 100.00,
  5, 10,
  8.00, 15.00,
  5.00, 25.00, 3.00,
  '{
    "exchange": "bitget",
    "symbol": "BTC-USDT",
    "orderType": "market",
    "marginMode": "isolated",
    "monitorWindow": 240,
    "volatilityThreshold": 100,
    "leverage": 8,
    "marginPercent": 12,
    "stopLossPercent": 5,
    "autoStartRetreatPercent": 3,
    "profitRetreatPercent": 25,
    "reverseEnabled": true,
    "reverseLossRatio": 50,
    "reverseProfitRatio": 100,
    "trailingStop": false,
    "remark": "趋势行情的标准配置，平衡风险与收益"
  }',
  '【推荐】趋势市场最佳策略，平衡风险收益。日收益预期1-5%。',
  1, 201
);

-- 【平衡型-震荡市场】
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_balanced_volatile', '⚖️ 平衡型-震荡市场',
  'balanced', 'volatile',
  180, 60.00,
  5, 8,
  6.00, 12.00,
  4.00, 22.00, 2.50,
  '{
    "exchange": "bitget",
    "symbol": "BTC-USDT",
    "orderType": "market",
    "marginMode": "isolated",
    "monitorWindow": 180,
    "volatilityThreshold": 60,
    "leverage": 6,
    "marginPercent": 10,
    "stopLossPercent": 4,
    "autoStartRetreatPercent": 2.5,
    "profitRetreatPercent": 22,
    "reverseEnabled": false,
    "reverseLossRatio": 0,
    "reverseProfitRatio": 0,
    "trailingStop": false,
    "remark": "震荡市场谨慎操作，等待突破"
  }',
  '震荡区间高抛低吸，不追涨杀跌。适合有一定经验的用户。',
  1, 202
);

-- 【平衡型-高波动】
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_balanced_high_vol', '⚖️ 平衡型-高波动',
  'balanced', 'high_vol',
  90, 180.00,
  4, 7,
  5.00, 10.00,
  6.00, 28.00, 4.00,
  '{
    "exchange": "bitget",
    "symbol": "BTC-USDT",
    "orderType": "market",
    "marginMode": "isolated",
    "monitorWindow": 90,
    "volatilityThreshold": 180,
    "leverage": 5,
    "marginPercent": 8,
    "stopLossPercent": 6,
    "autoStartRetreatPercent": 4,
    "profitRetreatPercent": 28,
    "reverseEnabled": true,
    "reverseLossRatio": 100,
    "reverseProfitRatio": 100,
    "trailingStop": true,
    "remark": "高波动需快速反应，启用移动止盈"
  }',
  '高波动市场机会与风险并存，适度降低仓位，快速止盈止损。',
  1, 203
);

-- 【平衡型-低波动】
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_balanced_low_vol', '⚖️ 平衡型-低波动',
  'balanced', 'low_vol',
  360, 40.00,
  6, 10,
  10.00, 18.00,
  3.00, 18.00, 2.00,
  '{
    "exchange": "bitget",
    "symbol": "BTC-USDT",
    "orderType": "market",
    "marginMode": "isolated",
    "monitorWindow": 360,
    "volatilityThreshold": 40,
    "leverage": 8,
    "marginPercent": 15,
    "stopLossPercent": 3,
    "autoStartRetreatPercent": 2,
    "profitRetreatPercent": 18,
    "reverseEnabled": false,
    "reverseLossRatio": 0,
    "reverseProfitRatio": 0,
    "trailingStop": false,
    "remark": "低波动加大仓位，耐心等待行情"
  }',
  '低波动市场可适当放大仓位，赚取稳定的小幅收益。',
  1, 204
);

-- =========================================
-- 🚀 激进型策略 (Aggressive)
-- 特点：高杠杆、大仓位、适合专业用户
-- =========================================

-- 【激进型-趋势市场】
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_aggressive_trend', '🚀 激进型-趋势市场',
  'aggressive', 'trend',
  180, 120.00,
  10, 20,
  10.00, 20.00,
  8.00, 20.00, 5.00,
  '{
    "exchange": "bitget",
    "symbol": "BTC-USDT",
    "orderType": "market",
    "marginMode": "isolated",
    "monitorWindow": 180,
    "volatilityThreshold": 120,
    "leverage": 15,
    "marginPercent": 18,
    "stopLossPercent": 8,
    "autoStartRetreatPercent": 5,
    "profitRetreatPercent": 20,
    "reverseEnabled": true,
    "reverseLossRatio": 100,
    "reverseProfitRatio": 100,
    "trailingStop": true,
    "remark": "趋势明确时重仓出击，追求高收益"
  }',
  '⚠️ 高风险策略！趋势明确时可获得高收益，但亏损也会放大。仅限专业用户。',
  1, 301
);

-- 【激进型-震荡市场】
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_aggressive_volatile', '🚀 激进型-震荡市场',
  'aggressive', 'volatile',
  120, 80.00,
  8, 15,
  8.00, 16.00,
  6.00, 18.00, 4.00,
  '{
    "exchange": "bitget",
    "symbol": "BTC-USDT",
    "orderType": "market",
    "marginMode": "isolated",
    "monitorWindow": 120,
    "volatilityThreshold": 80,
    "leverage": 12,
    "marginPercent": 14,
    "stopLossPercent": 6,
    "autoStartRetreatPercent": 4,
    "profitRetreatPercent": 18,
    "reverseEnabled": true,
    "reverseLossRatio": 50,
    "reverseProfitRatio": 50,
    "trailingStop": false,
    "remark": "震荡市场双向操作，频繁交易"
  }',
  '⚠️ 高风险策略！震荡市场双向开单，需要较强的市场判断能力。',
  1, 302
);

-- 【激进型-高波动】
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_aggressive_high_vol', '🚀 激进型-高波动',
  'aggressive', 'high_vol',
  60, 200.00,
  8, 12,
  6.00, 12.00,
  10.00, 22.00, 6.00,
  '{
    "exchange": "bitget",
    "symbol": "BTC-USDT",
    "orderType": "market",
    "marginMode": "isolated",
    "monitorWindow": 60,
    "volatilityThreshold": 200,
    "leverage": 10,
    "marginPercent": 10,
    "stopLossPercent": 10,
    "autoStartRetreatPercent": 6,
    "profitRetreatPercent": 22,
    "reverseEnabled": true,
    "reverseLossRatio": 100,
    "reverseProfitRatio": 100,
    "trailingStop": true,
    "remark": "高波动双向博弈，极高风险极高收益"
  }',
  '⚠️ 极高风险！高波动市场博取超额收益，可能快速盈利也可能快速爆仓。',
  1, 303
);

-- 【激进型-低波动】
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`, 
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage_min`, `leverage_max`,
  `margin_percent_min`, `margin_percent_max`,
  `stop_loss_percent`, `profit_retreat_percent`, `auto_start_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES (
  @group_id, 'official_btc_aggressive_low_vol', '🚀 激进型-低波动',
  'aggressive', 'low_vol',
  300, 50.00,
  12, 20,
  15.00, 25.00,
  5.00, 15.00, 3.00,
  '{
    "exchange": "bitget",
    "symbol": "BTC-USDT",
    "orderType": "market",
    "marginMode": "isolated",
    "monitorWindow": 300,
    "volatilityThreshold": 50,
    "leverage": 18,
    "marginPercent": 22,
    "stopLossPercent": 5,
    "autoStartRetreatPercent": 3,
    "profitRetreatPercent": 15,
    "reverseEnabled": false,
    "reverseLossRatio": 0,
    "reverseProfitRatio": 0,
    "trailingStop": false,
    "remark": "低波动时重仓等待突破，博取大行情"
  }',
  '⚠️ 高风险策略！低波动时重仓布局，等待大行情突破。',
  1, 304
);

-- ============================================================
-- 6. 验证插入结果
-- ============================================================

SELECT 
  g.id AS group_id,
  g.group_name,
  g.is_official AS '官方',
  g.exchange AS '交易平台',
  g.symbol AS '交易对',
  COUNT(s.id) AS '策略数'
FROM hg_trading_strategy_group g
LEFT JOIN hg_trading_strategy_template s ON s.group_id = g.id
WHERE g.group_key = 'official_bitget_btc_usdt'
GROUP BY g.id;

-- 显示所有策略的详细参数
SELECT 
  strategy_name AS '策略名称',
  risk_preference AS '风险偏好',
  market_state AS '市场状态',
  monitor_window AS '时间窗口(秒)',
  volatility_threshold AS '波动值',
  CONCAT(leverage_min, '-', leverage_max, 'x') AS '杠杆范围',
  CONCAT(margin_percent_min, '-', margin_percent_max, '%') AS '仓位范围',
  CONCAT(stop_loss_percent, '%') AS '止损',
  CONCAT(auto_start_retreat_percent, '%') AS '启动回撤',
  CONCAT(profit_retreat_percent, '%') AS '止盈回撤'
FROM hg_trading_strategy_template
WHERE group_id = @group_id
ORDER BY sort;

-- ============================================================
-- 完成！
-- ============================================================
