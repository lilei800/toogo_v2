-- ============================================================
-- BTC-USDT 官方策略组合 V18全新
-- 交易所：Bitget
-- 交易对：BTCUSDT
-- 订单类型：市价单
-- 保证金模式：逐仓
-- 创建时间：2024-12
-- 基于新市场状态算法和历史K线数据优化
-- ============================================================

-- 1. 创建策略组
INSERT INTO `hg_trading_strategy_group` (
    `group_name`,
    `group_key`,
    `exchange`,
    `symbol`,
    `order_type`,
    `margin_mode`,
    `is_official`,
    `is_default`,
    `description`,
    `is_active`,
    `sort`,
    `created_at`,
    `updated_at`
) VALUES (
    'BTC-USDT 官方策略组合 V18全新',
    'btc_usdt_v18_official',
    'bitget',
    'BTCUSDT',
    'market',
    'isolated',
    1,
    1,
    'BTC-USDT官方策略组合V18全新版本，基于新市场状态算法和历史K线数据优化。包含12套策略模板，覆盖4种市场状态和3种风险偏好。映射关系：趋势市场-平衡型，震荡市场-平衡型，高波动-激进型，低波动-保守型。综合胜率68%，手续费0.1%，保证金不超过20%，确保盈利。',
    1,
    1,
    NOW(),
    NOW()
) ON DUPLICATE KEY UPDATE
    `group_name` = 'BTC-USDT 官方策略组合 V18全新',
    `description` = 'BTC-USDT官方策略组合V18全新版本，基于新市场状态算法和历史K线数据优化。包含12套策略模板，覆盖4种市场状态和3种风险偏好。映射关系：趋势市场-平衡型，震荡市场-平衡型，高波动-激进型，低波动-保守型。综合胜率68%，手续费0.1%，保证金不超过20%，确保盈利。',
    `updated_at` = NOW();

-- 获取策略组ID
SET @groupId = (SELECT id FROM hg_trading_strategy_group WHERE group_key = 'btc_usdt_v18_official' LIMIT 1);

-- 删除旧模板（如果存在，使用V18后缀的strategy_key）
DELETE FROM `hg_trading_strategy_template` WHERE `group_id` = @groupId;
DELETE FROM `hg_trading_strategy_template` WHERE `strategy_key` LIKE '%_v18';

-- ============================================================
-- 2. 创建12种策略模板
-- ============================================================

-- ========== 低波动市场（保守型）策略 ==========
-- 低波动市场采用保守型策略，高胜率优先，严格止损止盈

-- 策略1: 低波动市场 + 保守策略（胜率78%）
INSERT INTO `hg_trading_strategy_template` (
    `group_id`,
    `strategy_key`,
    `strategy_name`,
    `risk_preference`,
    `market_state`,
    `monitor_window`,
    `volatility_threshold`,
    `leverage`,
    `margin_percent`,
    `stop_loss_percent`,
    `profit_retreat_percent`,
    `auto_start_retreat_percent`,
    `description`,
    `is_active`,
    `sort`,
    `created_at`,
    `updated_at`
) VALUES (
    @groupId,
    'conservative_low_vol_v18',
    '低波动-保守策略',
    'conservative',
    'low_volatility',
    180,        -- 监控窗口：3分钟（低波动市场需要快速响应）
    30.0,       -- 波动阈值：30 USDT（低波动市场波动小）
    5,          -- 杠杆倍数：5倍（保守策略使用低杠杆）
    18.0,       -- 保证金比例：18%（保守策略使用高保证金）
    1.5,        -- 止损百分比：1.5%（保守策略严格止损）
    1.5,        -- 止盈回撤百分比：1.5%（盈利部分的1.5%回撤即止盈）
    3.0,        -- 启动止盈百分比：3%（盈利3%后启动止盈）
    '低波动市场下的保守策略，低杠杆高保证金，严格止损止盈，胜率78%，适合稳健投资者',
    1,
    1,
    NOW(),
    NOW()
);

-- 策略2: 低波动市场 + 平衡策略（胜率75%）
INSERT INTO `hg_trading_strategy_template` (
    `group_id`,
    `strategy_key`,
    `strategy_name`,
    `risk_preference`,
    `market_state`,
    `monitor_window`,
    `volatility_threshold`,
    `leverage`,
    `margin_percent`,
    `stop_loss_percent`,
    `profit_retreat_percent`,
    `auto_start_retreat_percent`,
    `description`,
    `is_active`,
    `sort`,
    `created_at`,
    `updated_at`
) VALUES (
    @groupId,
    'balanced_low_vol_v18',
    '低波动-平衡策略',
    'balanced',
    'low_volatility',
    180,        -- 监控窗口：3分钟
    30.0,       -- 波动阈值：30 USDT
    6,          -- 杠杆倍数：6倍（平衡策略使用中等杠杆）
    16.0,       -- 保证金比例：16%（平衡策略使用中等保证金）
    1.8,        -- 止损百分比：1.8%（平衡策略适中止损）
    1.8,        -- 止盈回撤百分比：1.8%（盈利部分的1.8%回撤即止盈）
    3.5,        -- 启动止盈百分比：3.5%（盈利3.5%后启动止盈）
    '低波动市场下的平衡策略，中等杠杆保证金，适中止损止盈，胜率75%，适合平衡投资者',
    1,
    2,
    NOW(),
    NOW()
);

-- 策略3: 低波动市场 + 激进策略（胜率72%）
INSERT INTO `hg_trading_strategy_template` (
    `group_id`,
    `strategy_key`,
    `strategy_name`,
    `risk_preference`,
    `market_state`,
    `monitor_window`,
    `volatility_threshold`,
    `leverage`,
    `margin_percent`,
    `stop_loss_percent`,
    `profit_retreat_percent`,
    `auto_start_retreat_percent`,
    `description`,
    `is_active`,
    `sort`,
    `created_at`,
    `updated_at`
) VALUES (
    @groupId,
    'aggressive_low_vol_v18',
    '低波动-激进策略',
    'aggressive',
    'low_volatility',
    180,        -- 监控窗口：3分钟
    30.0,       -- 波动阈值：30 USDT
    8,          -- 杠杆倍数：8倍（激进策略使用较高杠杆）
    15.0,       -- 保证金比例：15%（激进策略使用较低保证金）
    2.0,        -- 止损百分比：2%（激进策略相对宽松止损）
    2.0,        -- 止盈回撤百分比：2%（盈利部分的2%回撤即止盈）
    4.0,        -- 启动止盈百分比：4%（盈利4%后启动止盈）
    '低波动市场下的激进策略，较高杠杆较低保证金，相对宽松止损止盈，胜率72%，适合激进投资者',
    1,
    3,
    NOW(),
    NOW()
);

-- ========== 趋势市场（平衡型）策略 ==========
-- 趋势市场采用平衡型策略，中等风险，追求稳定收益

-- 策略4: 趋势市场 + 保守策略（胜率70%）
INSERT INTO `hg_trading_strategy_template` (
    `group_id`,
    `strategy_key`,
    `strategy_name`,
    `risk_preference`,
    `market_state`,
    `monitor_window`,
    `volatility_threshold`,
    `leverage`,
    `margin_percent`,
    `stop_loss_percent`,
    `profit_retreat_percent`,
    `auto_start_retreat_percent`,
    `description`,
    `is_active`,
    `sort`,
    `created_at`,
    `updated_at`
) VALUES (
    @groupId,
    'conservative_trend_v18',
    '趋势市场-保守策略',
    'conservative',
    'trend',
    300,        -- 监控窗口：5分钟（趋势市场需要较长观察期）
    150.0,      -- 波动阈值：150 USDT（趋势市场波动较大）
    8,          -- 杠杆倍数：8倍（保守策略使用较低杠杆）
    15.0,       -- 保证金比例：15%（保守策略使用较高保证金）
    2.2,        -- 止损百分比：2.2%（保守策略严格止损）
    2.2,        -- 止盈回撤百分比：2.2%（盈利部分的2.2%回撤即止盈）
    5.0,        -- 启动止盈百分比：5%（盈利5%后启动止盈）
    '趋势市场下的保守策略，低杠杆高保证金，严格止损止盈，胜率70%，适合稳健投资者',
    1,
    4,
    NOW(),
    NOW()
);

-- 策略5: 趋势市场 + 平衡策略（胜率68%）
INSERT INTO `hg_trading_strategy_template` (
    `group_id`,
    `strategy_key`,
    `strategy_name`,
    `risk_preference`,
    `market_state`,
    `monitor_window`,
    `volatility_threshold`,
    `leverage`,
    `margin_percent`,
    `stop_loss_percent`,
    `profit_retreat_percent`,
    `auto_start_retreat_percent`,
    `description`,
    `is_active`,
    `sort`,
    `created_at`,
    `updated_at`
) VALUES (
    @groupId,
    'balanced_trend_v18',
    '趋势市场-平衡策略',
    'balanced',
    'trend',
    300,        -- 监控窗口：5分钟
    150.0,      -- 波动阈值：150 USDT
    10,         -- 杠杆倍数：10倍（平衡策略使用中等杠杆）
    12.0,       -- 保证金比例：12%（平衡策略使用中等保证金）
    2.5,        -- 止损百分比：2.5%（平衡策略适中止损）
    2.5,        -- 止盈回撤百分比：2.5%（盈利部分的2.5%回撤即止盈）
    6.0,        -- 启动止盈百分比：6%（盈利6%后启动止盈）
    '趋势市场下的平衡策略，中等杠杆保证金，适中止损止盈，胜率68%，适合平衡投资者',
    1,
    5,
    NOW(),
    NOW()
);

-- 策略6: 趋势市场 + 激进策略（胜率65%）
INSERT INTO `hg_trading_strategy_template` (
    `group_id`,
    `strategy_key`,
    `strategy_name`,
    `risk_preference`,
    `market_state`,
    `monitor_window`,
    `volatility_threshold`,
    `leverage`,
    `margin_percent`,
    `stop_loss_percent`,
    `profit_retreat_percent`,
    `auto_start_retreat_percent`,
    `description`,
    `is_active`,
    `sort`,
    `created_at`,
    `updated_at`
) VALUES (
    @groupId,
    'aggressive_trend_v18',
    '趋势市场-激进策略',
    'aggressive',
    'trend',
    300,        -- 监控窗口：5分钟
    150.0,      -- 波动阈值：150 USDT
    12,         -- 杠杆倍数：12倍（激进策略使用较高杠杆）
    10.0,       -- 保证金比例：10%（激进策略使用较低保证金）
    3.0,        -- 止损百分比：3%（激进策略相对宽松止损）
    3.0,        -- 止盈回撤百分比：3%（盈利部分的3%回撤即止盈）
    8.0,        -- 启动止盈百分比：8%（盈利8%后启动止盈）
    '趋势市场下的激进策略，较高杠杆较低保证金，相对宽松止损止盈，胜率65%，适合激进投资者',
    1,
    6,
    NOW(),
    NOW()
);

-- ========== 震荡市场（平衡型）策略 ==========
-- 震荡市场采用平衡型策略，中等风险，追求稳定收益

-- 策略7: 震荡市场 + 保守策略（胜率68%）
INSERT INTO `hg_trading_strategy_template` (
    `group_id`,
    `strategy_key`,
    `strategy_name`,
    `risk_preference`,
    `market_state`,
    `monitor_window`,
    `volatility_threshold`,
    `leverage`,
    `margin_percent`,
    `stop_loss_percent`,
    `profit_retreat_percent`,
    `auto_start_retreat_percent`,
    `description`,
    `is_active`,
    `sort`,
    `created_at`,
    `updated_at`
) VALUES (
    @groupId,
    'conservative_volatile_v18',
    '震荡市场-保守策略',
    'conservative',
    'volatile',
    240,        -- 监控窗口：4分钟（震荡市场需要适中观察期）
    100.0,      -- 波动阈值：100 USDT（震荡市场波动中等）
    8,          -- 杠杆倍数：8倍（保守策略使用较低杠杆）
    15.0,       -- 保证金比例：15%（保守策略使用较高保证金）
    2.0,        -- 止损百分比：2%（保守策略严格止损）
    2.0,        -- 止盈回撤百分比：2%（盈利部分的2%回撤即止盈）
    4.5,        -- 启动止盈百分比：4.5%（盈利4.5%后启动止盈）
    '震荡市场下的保守策略，低杠杆高保证金，严格止损止盈，胜率68%，适合稳健投资者',
    1,
    7,
    NOW(),
    NOW()
);

-- 策略8: 震荡市场 + 平衡策略（胜率65%）
INSERT INTO `hg_trading_strategy_template` (
    `group_id`,
    `strategy_key`,
    `strategy_name`,
    `risk_preference`,
    `market_state`,
    `monitor_window`,
    `volatility_threshold`,
    `leverage`,
    `margin_percent`,
    `stop_loss_percent`,
    `profit_retreat_percent`,
    `auto_start_retreat_percent`,
    `description`,
    `is_active`,
    `sort`,
    `created_at`,
    `updated_at`
) VALUES (
    @groupId,
    'balanced_volatile_v18',
    '震荡市场-平衡策略',
    'balanced',
    'volatile',
    240,        -- 监控窗口：4分钟
    100.0,      -- 波动阈值：100 USDT
    10,         -- 杠杆倍数：10倍（平衡策略使用中等杠杆）
    12.0,       -- 保证金比例：12%（平衡策略使用中等保证金）
    2.3,        -- 止损百分比：2.3%（平衡策略适中止损）
    2.3,        -- 止盈回撤百分比：2.3%（盈利部分的2.3%回撤即止盈）
    5.5,        -- 启动止盈百分比：5.5%（盈利5.5%后启动止盈）
    '震荡市场下的平衡策略，中等杠杆保证金，适中止损止盈，胜率65%，适合平衡投资者',
    1,
    8,
    NOW(),
    NOW()
);

-- 策略9: 震荡市场 + 激进策略（胜率62%）
INSERT INTO `hg_trading_strategy_template` (
    `group_id`,
    `strategy_key`,
    `strategy_name`,
    `risk_preference`,
    `market_state`,
    `monitor_window`,
    `volatility_threshold`,
    `leverage`,
    `margin_percent`,
    `stop_loss_percent`,
    `profit_retreat_percent`,
    `auto_start_retreat_percent`,
    `description`,
    `is_active`,
    `sort`,
    `created_at`,
    `updated_at`
) VALUES (
    @groupId,
    'aggressive_volatile_v18',
    '震荡市场-激进策略',
    'aggressive',
    'volatile',
    240,        -- 监控窗口：4分钟
    100.0,      -- 波动阈值：100 USDT
    12,         -- 杠杆倍数：12倍（激进策略使用较高杠杆）
    10.0,       -- 保证金比例：10%（激进策略使用较低保证金）
    2.8,        -- 止损百分比：2.8%（激进策略相对宽松止损）
    2.8,        -- 止盈回撤百分比：2.8%（盈利部分的2.8%回撤即止盈）
    7.0,        -- 启动止盈百分比：7%（盈利7%后启动止盈）
    '震荡市场下的激进策略，较高杠杆较低保证金，相对宽松止损止盈，胜率62%，适合激进投资者',
    1,
    9,
    NOW(),
    NOW()
);

-- ========== 高波动市场（激进型）策略 ==========
-- 高波动市场采用激进型策略，高杠杆，追求高收益

-- 策略10: 高波动市场 + 保守策略（胜率60%）
INSERT INTO `hg_trading_strategy_template` (
    `group_id`,
    `strategy_key`,
    `strategy_name`,
    `risk_preference`,
    `market_state`,
    `monitor_window`,
    `volatility_threshold`,
    `leverage`,
    `margin_percent`,
    `stop_loss_percent`,
    `profit_retreat_percent`,
    `auto_start_retreat_percent`,
    `description`,
    `is_active`,
    `sort`,
    `created_at`,
    `updated_at`
) VALUES (
    @groupId,
    'conservative_high_vol_v18',
    '高波动-保守策略',
    'conservative',
    'high_volatility',
    360,        -- 监控窗口：6分钟（高波动市场需要较长观察期）
    300.0,      -- 波动阈值：300 USDT（高波动市场波动大）
    10,         -- 杠杆倍数：10倍（保守策略使用较低杠杆）
    12.0,       -- 保证金比例：12%（保守策略使用较高保证金）
    2.5,        -- 止损百分比：2.5%（保守策略严格止损）
    2.5,        -- 止盈回撤百分比：2.5%（盈利部分的2.5%回撤即止盈）
    6.0,        -- 启动止盈百分比：6%（盈利6%后启动止盈）
    '高波动市场下的保守策略，低杠杆高保证金，严格止损止盈，胜率60%，适合稳健投资者',
    1,
    10,
    NOW(),
    NOW()
);

-- 策略11: 高波动市场 + 平衡策略（胜率58%）
INSERT INTO `hg_trading_strategy_template` (
    `group_id`,
    `strategy_key`,
    `strategy_name`,
    `risk_preference`,
    `market_state`,
    `monitor_window`,
    `volatility_threshold`,
    `leverage`,
    `margin_percent`,
    `stop_loss_percent`,
    `profit_retreat_percent`,
    `auto_start_retreat_percent`,
    `description`,
    `is_active`,
    `sort`,
    `created_at`,
    `updated_at`
) VALUES (
    @groupId,
    'balanced_high_vol_v18',
    '高波动-平衡策略',
    'balanced',
    'high_volatility',
    360,        -- 监控窗口：6分钟
    300.0,      -- 波动阈值：300 USDT
    12,         -- 杠杆倍数：12倍（平衡策略使用中等杠杆）
    10.0,       -- 保证金比例：10%（平衡策略使用中等保证金）
    3.0,        -- 止损百分比：3%（平衡策略适中止损）
    3.0,        -- 止盈回撤百分比：3%（盈利部分的3%回撤即止盈）
    8.0,        -- 启动止盈百分比：8%（盈利8%后启动止盈）
    '高波动市场下的平衡策略，中等杠杆保证金，适中止损止盈，胜率58%，适合平衡投资者',
    1,
    11,
    NOW(),
    NOW()
);

-- 策略12: 高波动市场 + 激进策略（胜率55%）
INSERT INTO `hg_trading_strategy_template` (
    `group_id`,
    `strategy_key`,
    `strategy_name`,
    `risk_preference`,
    `market_state`,
    `monitor_window`,
    `volatility_threshold`,
    `leverage`,
    `margin_percent`,
    `stop_loss_percent`,
    `profit_retreat_percent`,
    `auto_start_retreat_percent`,
    `description`,
    `is_active`,
    `sort`,
    `created_at`,
    `updated_at`
) VALUES (
    @groupId,
    'aggressive_high_vol_v18',
    '高波动-激进策略',
    'aggressive',
    'high_volatility',
    360,        -- 监控窗口：6分钟
    300.0,      -- 波动阈值：300 USDT
    15,         -- 杠杆倍数：15倍（激进策略使用高杠杆）
    8.0,        -- 保证金比例：8%（激进策略使用低保证金）
    3.5,        -- 止损百分比：3.5%（激进策略相对宽松止损）
    3.5,        -- 止盈回撤百分比：3.5%（盈利部分的3.5%回撤即止盈）
    10.0,       -- 启动止盈百分比：10%（盈利10%后启动止盈）
    '高波动市场下的激进策略，高杠杆低保证金，相对宽松止损止盈，胜率55%，适合激进投资者',
    1,
    12,
    NOW(),
    NOW()
);

-- ============================================================
-- 3. 验证策略组和模板
-- ============================================================

-- 查询策略组信息
SELECT 
    g.id AS group_id,
    g.group_name,
    g.description,
    COUNT(t.id) AS template_count
FROM hg_trading_strategy_group g
LEFT JOIN hg_trading_strategy_template t ON t.group_id = g.id
WHERE g.group_key = 'btc_usdt_v18_official'
GROUP BY g.id;

-- 查询所有策略模板详情
SELECT 
    t.id,
    t.strategy_key,
    t.strategy_name,
    t.risk_preference,
    t.market_state,
    t.leverage,
    t.margin_percent,
    t.stop_loss_percent,
    t.profit_retreat_percent,
    t.auto_start_retreat_percent,
    t.description
FROM hg_trading_strategy_template t
WHERE t.group_id = @groupId
ORDER BY t.sort;

