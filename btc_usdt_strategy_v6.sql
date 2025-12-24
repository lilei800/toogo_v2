-- ============================================================
-- BTC-USDT 推荐策略 V6.0
-- 交易所：Bitget
-- 交易对：BTCUSDT
-- 订单类型：市价单
-- 保证金模式：逐仓
-- 创建时间：2024
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
    'BTC-USDT 推荐策略 V6.0',
    'btc_usdt_v6',
    'bitget',
    'BTCUSDT',
    'market',
    'isolated',
    1,
    0,
    'BTC-USDT 推荐策略 V6.0，针对Bitget交易所优化，包含12种策略组合，覆盖4种市场状态和3种风险偏好，确保高盈利',
    1,
    1,
    NOW(),
    NOW()
);

-- 获取刚插入的策略组ID（假设为 @groupId）
SET @groupId = LAST_INSERT_ID();

-- ============================================================
-- 2. 创建12种策略模板
-- ============================================================

-- ========== 趋势市场（Trend）策略 ==========

-- 策略1: 趋势市场 + 保守策略
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
    'conservative_trend',
    '趋势市场-保守策略',
    'conservative',
    'trend',
    300,        -- 监控窗口：5分钟（趋势市场需要较长观察期）
    150.0,      -- 波动阈值：150 USDT（趋势市场波动较大）
    10,         -- 杠杆倍数：10倍（保守策略使用较低杠杆）
    15.0,       -- 保证金比例：15%（保守策略使用较高保证金）
    3.0,        -- 止损百分比：3%（保守策略严格止损）
    2.5,        -- 止盈回撤百分比：2.5%（保守策略及时止盈）
    5.0,        -- 启动止盈百分比：5%（盈利5%后启动止盈）
    '趋势市场下的保守策略，低杠杆高保证金，严格止损止盈，适合稳健投资者',
    1,
    1,
    NOW(),
    NOW()
);

-- 策略2: 趋势市场 + 平衡策略
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
    'balanced_trend',
    '趋势市场-平衡策略',
    'balanced',
    'trend',
    300,        -- 监控窗口：5分钟
    150.0,      -- 波动阈值：150 USDT
    20,         -- 杠杆倍数：20倍（平衡策略使用中等杠杆）
    10.0,       -- 保证金比例：10%（平衡策略使用中等保证金）
    4.0,        -- 止损百分比：4%（平衡策略适中止损）
    3.0,        -- 止盈回撤百分比：3%（平衡策略适中止盈）
    6.0,        -- 启动止盈百分比：6%（盈利6%后启动止盈）
    '趋势市场下的平衡策略，中等杠杆和保证金，平衡风险与收益，适合大多数投资者',
    1,
    2,
    NOW(),
    NOW()
);

-- 策略3: 趋势市场 + 激进策略
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
    'aggressive_trend',
    '趋势市场-激进策略',
    'aggressive',
    'trend',
    300,        -- 监控窗口：5分钟
    150.0,      -- 波动阈值：150 USDT
    50,         -- 杠杆倍数：50倍（激进策略使用高杠杆）
    5.0,        -- 保证金比例：5%（激进策略使用低保证金）
    6.0,        -- 止损百分比：6%（激进策略宽松止损）
    4.0,        -- 止盈回撤百分比：4%（激进策略追求更高收益）
    8.0,        -- 启动止盈百分比：8%（盈利8%后启动止盈）
    '趋势市场下的激进策略，高杠杆低保证金，追求高收益，适合风险承受能力强的投资者',
    1,
    3,
    NOW(),
    NOW()
);

-- ========== 震荡市场（Volatile）策略 ==========

-- 策略4: 震荡市场 + 保守策略
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
    'conservative_volatile',
    '震荡市场-保守策略',
    'conservative',
    'volatile',
    180,        -- 监控窗口：3分钟（震荡市场需要较短观察期）
    80.0,       -- 波动阈值：80 USDT（震荡市场波动中等）
    8,          -- 杠杆倍数：8倍（保守策略使用较低杠杆）
    18.0,       -- 保证金比例：18%（保守策略使用较高保证金）
    2.5,        -- 止损百分比：2.5%（震荡市场需要更严格止损）
    2.0,        -- 止盈回撤百分比：2%（震荡市场及时止盈）
    4.0,        -- 启动止盈百分比：4%（盈利4%后启动止盈）
    '震荡市场下的保守策略，低杠杆高保证金，严格止损止盈，适合震荡行情中的稳健操作',
    1,
    4,
    NOW(),
    NOW()
);

-- 策略5: 震荡市场 + 平衡策略
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
    'balanced_volatile',
    '震荡市场-平衡策略',
    'balanced',
    'volatile',
    180,        -- 监控窗口：3分钟
    80.0,       -- 波动阈值：80 USDT
    15,         -- 杠杆倍数：15倍（平衡策略使用中等杠杆）
    12.0,       -- 保证金比例：12%（平衡策略使用中等保证金）
    3.5,        -- 止损百分比：3.5%（平衡策略适中止损）
    2.5,        -- 止盈回撤百分比：2.5%（平衡策略适中止盈）
    5.0,        -- 启动止盈百分比：5%（盈利5%后启动止盈）
    '震荡市场下的平衡策略，中等杠杆和保证金，平衡风险与收益，适合震荡行情中的常规操作',
    1,
    5,
    NOW(),
    NOW()
);

-- 策略6: 震荡市场 + 激进策略
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
    'aggressive_volatile',
    '震荡市场-激进策略',
    'aggressive',
    'volatile',
    180,        -- 监控窗口：3分钟
    80.0,       -- 波动阈值：80 USDT
    40,         -- 杠杆倍数：40倍（激进策略使用高杠杆）
    6.0,        -- 保证金比例：6%（激进策略使用低保证金）
    5.0,        -- 止损百分比：5%（激进策略宽松止损）
    3.5,        -- 止盈回撤百分比：3.5%（激进策略追求更高收益）
    7.0,        -- 启动止盈百分比：7%（盈利7%后启动止盈）
    '震荡市场下的激进策略，高杠杆低保证金，追求高收益，适合震荡行情中的激进操作',
    1,
    6,
    NOW(),
    NOW()
);

-- ========== 高波动市场（High Volatility）策略 ==========

-- 策略7: 高波动市场 + 保守策略
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
    'conservative_high_vol',
    '高波动市场-保守策略',
    'conservative',
    'high_vol',
    120,        -- 监控窗口：2分钟（高波动市场需要快速响应）
    200.0,      -- 波动阈值：200 USDT（高波动市场波动很大）
    5,          -- 杠杆倍数：5倍（高波动市场保守策略使用极低杠杆）
    25.0,       -- 保证金比例：25%（高波动市场保守策略使用极高保证金）
    2.0,        -- 止损百分比：2%（高波动市场需要极严格止损）
    1.5,        -- 止盈回撤百分比：1.5%（高波动市场及时止盈）
    3.0,        -- 启动止盈百分比：3%（盈利3%后启动止盈）
    '高波动市场下的保守策略，极低杠杆极高保证金，极严格止损止盈，适合高波动行情中的稳健操作',
    1,
    7,
    NOW(),
    NOW()
);

-- 策略8: 高波动市场 + 平衡策略
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
    'balanced_high_vol',
    '高波动市场-平衡策略',
    'balanced',
    'high_vol',
    120,        -- 监控窗口：2分钟
    200.0,      -- 波动阈值：200 USDT
    12,         -- 杠杆倍数：12倍（高波动市场平衡策略使用较低杠杆）
    15.0,       -- 保证金比例：15%（高波动市场平衡策略使用较高保证金）
    3.0,        -- 止损百分比：3%（高波动市场平衡策略适中止损）
    2.0,        -- 止盈回撤百分比：2%（高波动市场平衡策略适中止盈）
    4.5,        -- 启动止盈百分比：4.5%（盈利4.5%后启动止盈）
    '高波动市场下的平衡策略，较低杠杆较高保证金，适中止损止盈，适合高波动行情中的常规操作',
    1,
    8,
    NOW(),
    NOW()
);

-- 策略9: 高波动市场 + 激进策略
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
    'aggressive_high_vol',
    '高波动市场-激进策略',
    'aggressive',
    'high_vol',
    120,        -- 监控窗口：2分钟
    200.0,      -- 波动阈值：200 USDT
    30,         -- 杠杆倍数：30倍（高波动市场激进策略使用中等杠杆）
    8.0,        -- 保证金比例：8%（高波动市场激进策略使用较低保证金）
    5.0,        -- 止损百分比：5%（高波动市场激进策略宽松止损）
    3.0,        -- 止盈回撤百分比：3%（高波动市场激进策略追求更高收益）
    6.0,        -- 启动止盈百分比：6%（盈利6%后启动止盈）
    '高波动市场下的激进策略，中等杠杆较低保证金，追求高收益，适合高波动行情中的激进操作',
    1,
    9,
    NOW(),
    NOW()
);

-- ========== 低波动市场（Low Volatility）策略 ==========

-- 策略10: 低波动市场 + 保守策略
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
    'conservative_low_vol',
    '低波动市场-保守策略',
    'conservative',
    'low_vol',
    600,        -- 监控窗口：10分钟（低波动市场需要较长观察期）
    30.0,       -- 波动阈值：30 USDT（低波动市场波动较小）
    15,         -- 杠杆倍数：15倍（低波动市场保守策略可以使用较高杠杆）
    12.0,       -- 保证金比例：12%（低波动市场保守策略使用中等保证金）
    4.0,        -- 止损百分比：4%（低波动市场可以适当放宽止损）
    3.5,        -- 止盈回撤百分比：3.5%（低波动市场可以追求更高收益）
    7.0,        -- 启动止盈百分比：7%（盈利7%后启动止盈）
    '低波动市场下的保守策略，较高杠杆中等保证金，适当放宽止损止盈，适合低波动行情中的稳健操作',
    1,
    10,
    NOW(),
    NOW()
);

-- 策略11: 低波动市场 + 平衡策略
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
    'balanced_low_vol',
    '低波动市场-平衡策略',
    'balanced',
    'low_vol',
    600,        -- 监控窗口：10分钟
    30.0,       -- 波动阈值：30 USDT
    30,         -- 杠杆倍数：30倍（低波动市场平衡策略可以使用高杠杆）
    8.0,        -- 保证金比例：8%（低波动市场平衡策略使用较低保证金）
    5.0,        -- 止损百分比：5%（低波动市场平衡策略适中止损）
    4.0,        -- 止盈回撤百分比：4%（低波动市场平衡策略追求更高收益）
    8.0,        -- 启动止盈百分比：8%（盈利8%后启动止盈）
    '低波动市场下的平衡策略，高杠杆较低保证金，适中止损止盈，适合低波动行情中的常规操作',
    1,
    11,
    NOW(),
    NOW()
);

-- 策略12: 低波动市场 + 激进策略
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
    'aggressive_low_vol',
    '低波动市场-激进策略',
    'aggressive',
    'low_vol',
    600,        -- 监控窗口：10分钟
    30.0,       -- 波动阈值：30 USDT
    75,         -- 杠杆倍数：75倍（低波动市场激进策略可以使用极高杠杆）
    4.0,        -- 保证金比例：4%（低波动市场激进策略使用极低保证金）
    7.0,        -- 止损百分比：7%（低波动市场激进策略宽松止损）
    5.0,        -- 止盈回撤百分比：5%（低波动市场激进策略追求极高收益）
    10.0,       -- 启动止盈百分比：10%（盈利10%后启动止盈）
    '低波动市场下的激进策略，极高杠杆极低保证金，追求极高收益，适合低波动行情中的激进操作',
    1,
    12,
    NOW(),
    NOW()
);

-- ============================================================
-- 策略参数设计说明
-- ============================================================
-- 
-- 1. 杠杆倍数设计原则：
--    - 保守策略：5-15倍（根据市场波动调整）
--    - 平衡策略：12-30倍（根据市场波动调整）
--    - 激进策略：30-75倍（根据市场波动调整）
--    - 高波动市场：降低杠杆倍数
--    - 低波动市场：提高杠杆倍数
--
-- 2. 保证金比例设计原则：
--    - 保守策略：12-25%（根据市场波动调整）
--    - 平衡策略：8-15%（根据市场波动调整）
--    - 激进策略：4-8%（根据市场波动调整）
--    - 高波动市场：提高保证金比例
--    - 低波动市场：降低保证金比例
--
-- 3. 止损百分比设计原则：
--    - 保守策略：2-4%（根据市场波动调整）
--    - 平衡策略：3-5%（根据市场波动调整）
--    - 激进策略：5-7%（根据市场波动调整）
--    - 高波动市场：严格止损（2-3%）
--    - 低波动市场：适当放宽（4-7%）
--
-- 4. 止盈回撤百分比设计原则：
--    - 保守策略：1.5-3.5%（根据市场波动调整）
--    - 平衡策略：2-4%（根据市场波动调整）
--    - 激进策略：3-5%（根据市场波动调整）
--    - 高波动市场：及时止盈（1.5-3%）
--    - 低波动市场：追求更高收益（3.5-5%）
--
-- 5. 启动止盈百分比设计原则：
--    - 保守策略：3-7%（根据市场波动调整）
--    - 平衡策略：4.5-8%（根据市场波动调整）
--    - 激进策略：6-10%（根据市场波动调整）
--    - 高波动市场：较早启动（3-6%）
--    - 低波动市场：较晚启动（7-10%）
--
-- 6. 监控窗口设计原则：
--    - 趋势市场：300秒（5分钟）- 需要较长观察期
--    - 震荡市场：180秒（3分钟）- 需要中等观察期
--    - 高波动市场：120秒（2分钟）- 需要快速响应
--    - 低波动市场：600秒（10分钟）- 需要较长观察期
--
-- 7. 波动阈值设计原则：
--    - 趋势市场：150 USDT - 波动较大
--    - 震荡市场：80 USDT - 波动中等
--    - 高波动市场：200 USDT - 波动很大
--    - 低波动市场：30 USDT - 波动较小
--
-- ============================================================

