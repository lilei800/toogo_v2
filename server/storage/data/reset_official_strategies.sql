-- ============================================
-- 重置官方策略模板 - BTC-USDT V4.0 (完整版)
-- 删除所有旧官方策略，创建新的12种策略
-- ============================================

-- 1. 删除旧的官方策略模板
DELETE FROM hg_trading_strategy_template WHERE group_id IN (SELECT id FROM hg_trading_strategy_group WHERE is_official = 1);
DELETE FROM hg_trading_strategy_group WHERE is_official = 1;

-- 2. 创建新的官方策略模板组
INSERT INTO hg_trading_strategy_group (
    group_name, group_key, exchange, symbol, order_type, margin_mode,
    is_official, user_id, description, is_active, sort, created_at, updated_at
) VALUES (
    '🚀 BTC-USDT 官方策略 V4.0', 
    'official_btc_usdt_v4',
    'bitget',
    'BTCUSDT',
    'market',
    'isolated',
    1,
    0,
    '专业团队精心调优的BTC-USDT策略组合，基于实时行情分析、智能方向判断、多时间周期综合分析，包含12种市场状态和风险偏好组合，支持全自动下单和平仓。',
    1,
    1,
    NOW(),
    NOW()
);

SET @group_id = LAST_INSERT_ID();

-- ============================================
-- 策略参数说明：
-- 
-- 【基础参数】
-- - monitor_window: 时间窗口(秒) - 分析多长时间内的行情
-- - volatility_threshold: 波动阈值(USDT) - 价格偏离多少触发信号
-- - leverage_min/max: 杠杆范围
-- - margin_percent_min/max: 仓位比例范围
-- - stop_loss_percent: 止损百分比
-- - auto_start_retreat_percent: 启动止盈百分比
-- - profit_retreat_percent: 止盈回撤百分比
--
-- 【config_json包含】
-- - exchange: 交易平台
-- - symbol: 交易对
-- - orderType: 订单类型
-- - marginMode: 保证金模式
-- - leverage: 推荐杠杆
-- - marginPercent: 推荐仓位
-- - reverseEnabled: 是否启用反向单
-- - reverseLossRatio: 亏损回撤反向比例
-- - reverseProfitRatio: 盈利回撤反向比例
-- - aiWeightEnabled: 是否启用AI权重
-- ============================================

-- ========== 趋势市场策略 (TREND) ==========
-- 特点：价格持续单向移动，需要较长时间窗口确认趋势，跟随趋势获利

-- 1. 趋势市场 - 保守型
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, CONCAT(@group_id, '_trend_conservative'), '🛡️ 趋势-保守型', 'conservative', 'trend',
    300, 180,
    3, 5, 5, 10,
    3.0, 2.0, 35,
    JSON_OBJECT(
        'exchange', 'bitget',
        'symbol', 'BTCUSDT',
        'orderType', 'market',
        'marginMode', 'isolated',
        'leverage', 4,
        'marginPercent', 8,
        'stopLossPercent', 3.0,
        'autoStartRetreatPercent', 2.0,
        'profitRetreatPercent', 35,
        'monitorWindow', 300,
        'volatilityThreshold', 180,
        'reverseEnabled', true,
        'reverseLossRatio', 70,
        'reverseProfitRatio', 90,
        'aiWeightEnabled', true,
        'trendFollowStrength', 0.8
    ),
    '趋势市场保守策略：5分钟窗口确认趋势，180U波动触发信号，低杠杆稳健操作，严格止损保护本金',
    1, 1, NOW(), NOW()
);

-- 2. 趋势市场 - 平衡型 ★推荐
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, CONCAT(@group_id, '_trend_balanced'), '⭐ 趋势-平衡型', 'balanced', 'trend',
    240, 150,
    5, 10, 8, 15,
    5.0, 3.0, 30,
    JSON_OBJECT(
        'exchange', 'bitget',
        'symbol', 'BTCUSDT',
        'orderType', 'market',
        'marginMode', 'isolated',
        'leverage', 8,
        'marginPercent', 12,
        'stopLossPercent', 5.0,
        'autoStartRetreatPercent', 3.0,
        'profitRetreatPercent', 30,
        'monitorWindow', 240,
        'volatilityThreshold', 150,
        'reverseEnabled', true,
        'reverseLossRatio', 60,
        'reverseProfitRatio', 80,
        'aiWeightEnabled', true,
        'trendFollowStrength', 0.85
    ),
    '【推荐】趋势市场平衡策略：4分钟窗口快速捕捉趋势，平衡收益与风险，适合大多数用户',
    1, 2, NOW(), NOW()
);

-- 3. 趋势市场 - 激进型
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, CONCAT(@group_id, '_trend_aggressive'), '🚀 趋势-激进型', 'aggressive', 'trend',
    180, 120,
    10, 20, 12, 20,
    8.0, 5.0, 25,
    JSON_OBJECT(
        'exchange', 'bitget',
        'symbol', 'BTCUSDT',
        'orderType', 'market',
        'marginMode', 'isolated',
        'leverage', 15,
        'marginPercent', 18,
        'stopLossPercent', 8.0,
        'autoStartRetreatPercent', 5.0,
        'profitRetreatPercent', 25,
        'monitorWindow', 180,
        'volatilityThreshold', 120,
        'reverseEnabled', true,
        'reverseLossRatio', 50,
        'reverseProfitRatio', 70,
        'aiWeightEnabled', true,
        'trendFollowStrength', 0.9
    ),
    '趋势市场激进策略：3分钟快速响应，高杠杆追求高收益，适合有经验的交易者',
    1, 3, NOW(), NOW()
);

-- ========== 震荡市场策略 (RANGE) ==========
-- 特点：价格在区间内波动，需要短时间窗口捕捉波动，高抛低吸

-- 4. 震荡市场 - 保守型
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, CONCAT(@group_id, '_range_conservative'), '🛡️ 震荡-保守型', 'conservative', 'range',
    180, 100,
    2, 4, 4, 8,
    2.5, 1.5, 40,
    JSON_OBJECT(
        'exchange', 'bitget',
        'symbol', 'BTCUSDT',
        'orderType', 'market',
        'marginMode', 'isolated',
        'leverage', 3,
        'marginPercent', 6,
        'stopLossPercent', 2.5,
        'autoStartRetreatPercent', 1.5,
        'profitRetreatPercent', 40,
        'monitorWindow', 180,
        'volatilityThreshold', 100,
        'reverseEnabled', true,
        'reverseLossRatio', 40,
        'reverseProfitRatio', 60,
        'aiWeightEnabled', true,
        'rangeTradeStrength', 0.75
    ),
    '震荡市场保守策略：3分钟窗口捕捉区间波动，低杠杆高抛低吸，稳健获利',
    1, 4, NOW(), NOW()
);

-- 5. 震荡市场 - 平衡型
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, CONCAT(@group_id, '_range_balanced'), '⚖️ 震荡-平衡型', 'balanced', 'range',
    150, 80,
    4, 8, 6, 12,
    4.0, 2.0, 35,
    JSON_OBJECT(
        'exchange', 'bitget',
        'symbol', 'BTCUSDT',
        'orderType', 'market',
        'marginMode', 'isolated',
        'leverage', 6,
        'marginPercent', 10,
        'stopLossPercent', 4.0,
        'autoStartRetreatPercent', 2.0,
        'profitRetreatPercent', 35,
        'monitorWindow', 150,
        'volatilityThreshold', 80,
        'reverseEnabled', true,
        'reverseLossRatio', 30,
        'reverseProfitRatio', 50,
        'aiWeightEnabled', true,
        'rangeTradeStrength', 0.8
    ),
    '震荡市场平衡策略：快速捕捉区间波动，平衡仓位控制风险',
    1, 5, NOW(), NOW()
);

-- 6. 震荡市场 - 激进型
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, CONCAT(@group_id, '_range_aggressive'), '🚀 震荡-激进型', 'aggressive', 'range',
    120, 60,
    8, 15, 10, 18,
    6.0, 3.0, 30,
    JSON_OBJECT(
        'exchange', 'bitget',
        'symbol', 'BTCUSDT',
        'orderType', 'market',
        'marginMode', 'isolated',
        'leverage', 12,
        'marginPercent', 15,
        'stopLossPercent', 6.0,
        'autoStartRetreatPercent', 3.0,
        'profitRetreatPercent', 30,
        'monitorWindow', 120,
        'volatilityThreshold', 60,
        'reverseEnabled', true,
        'reverseLossRatio', 20,
        'reverseProfitRatio', 40,
        'aiWeightEnabled', true,
        'rangeTradeStrength', 0.85
    ),
    '震荡市场激进策略：2分钟快速响应，高频捕捉波动机会',
    1, 6, NOW(), NOW()
);

-- ========== 高波动市场策略 (HIGH_VOL) ==========
-- 特点：价格剧烈波动，需要更宽的止损防止被洗出，快速锁定利润

-- 7. 高波动 - 保守型
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, CONCAT(@group_id, '_high_vol_conservative'), '🛡️ 高波动-保守型', 'conservative', 'high_vol',
    90, 250,
    2, 3, 3, 6,
    5.0, 2.0, 25,
    JSON_OBJECT(
        'exchange', 'bitget',
        'symbol', 'BTCUSDT',
        'orderType', 'market',
        'marginMode', 'isolated',
        'leverage', 2,
        'marginPercent', 5,
        'stopLossPercent', 5.0,
        'autoStartRetreatPercent', 2.0,
        'profitRetreatPercent', 25,
        'monitorWindow', 90,
        'volatilityThreshold', 250,
        'reverseEnabled', false,
        'reverseLossRatio', 100,
        'reverseProfitRatio', 100,
        'aiWeightEnabled', true,
        'volatilityAdaptive', true
    ),
    '高波动保守策略：降低杠杆应对剧烈波动，快速止盈锁定利润，不开启反向单',
    1, 7, NOW(), NOW()
);

-- 8. 高波动 - 平衡型
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, CONCAT(@group_id, '_high_vol_balanced'), '⚖️ 高波动-平衡型', 'balanced', 'high_vol',
    60, 200,
    3, 5, 5, 10,
    7.0, 3.0, 22,
    JSON_OBJECT(
        'exchange', 'bitget',
        'symbol', 'BTCUSDT',
        'orderType', 'market',
        'marginMode', 'isolated',
        'leverage', 4,
        'marginPercent', 8,
        'stopLossPercent', 7.0,
        'autoStartRetreatPercent', 3.0,
        'profitRetreatPercent', 22,
        'monitorWindow', 60,
        'volatilityThreshold', 200,
        'reverseEnabled', true,
        'reverseLossRatio', 70,
        'reverseProfitRatio', 90,
        'aiWeightEnabled', true,
        'volatilityAdaptive', true
    ),
    '高波动平衡策略：适度杠杆抓住大波动机会，快速止盈防止回撤',
    1, 8, NOW(), NOW()
);

-- 9. 高波动 - 激进型
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, CONCAT(@group_id, '_high_vol_aggressive'), '🚀 高波动-激进型', 'aggressive', 'high_vol',
    45, 150,
    5, 10, 8, 15,
    10.0, 5.0, 20,
    JSON_OBJECT(
        'exchange', 'bitget',
        'symbol', 'BTCUSDT',
        'orderType', 'market',
        'marginMode', 'isolated',
        'leverage', 8,
        'marginPercent', 12,
        'stopLossPercent', 10.0,
        'autoStartRetreatPercent', 5.0,
        'profitRetreatPercent', 20,
        'monitorWindow', 45,
        'volatilityThreshold', 150,
        'reverseEnabled', true,
        'reverseLossRatio', 60,
        'reverseProfitRatio', 80,
        'aiWeightEnabled', true,
        'volatilityAdaptive', true
    ),
    '高波动激进策略：快速响应大波动，宽止损防止被洗出，快速锁定利润',
    1, 9, NOW(), NOW()
);

-- ========== 低波动市场策略 (LOW_VOL) ==========
-- 特点：价格波动小，需要更高杠杆放大收益，更长时间窗口等待机会

-- 10. 低波动 - 保守型
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, CONCAT(@group_id, '_low_vol_conservative'), '🛡️ 低波动-保守型', 'conservative', 'low_vol',
    600, 60,
    5, 8, 6, 12,
    2.0, 1.0, 45,
    JSON_OBJECT(
        'exchange', 'bitget',
        'symbol', 'BTCUSDT',
        'orderType', 'market',
        'marginMode', 'isolated',
        'leverage', 6,
        'marginPercent', 10,
        'stopLossPercent', 2.0,
        'autoStartRetreatPercent', 1.0,
        'profitRetreatPercent', 45,
        'monitorWindow', 600,
        'volatilityThreshold', 60,
        'reverseEnabled', true,
        'reverseLossRatio', 60,
        'reverseProfitRatio', 90,
        'aiWeightEnabled', true,
        'lowVolAmplify', true
    ),
    '低波动保守策略：长时间窗口等待机会，适度杠杆放大小波动收益',
    1, 10, NOW(), NOW()
);

-- 11. 低波动 - 平衡型
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, CONCAT(@group_id, '_low_vol_balanced'), '⚖️ 低波动-平衡型', 'balanced', 'low_vol',
    480, 50,
    8, 12, 10, 16,
    3.0, 1.5, 40,
    JSON_OBJECT(
        'exchange', 'bitget',
        'symbol', 'BTCUSDT',
        'orderType', 'market',
        'marginMode', 'isolated',
        'leverage', 10,
        'marginPercent', 13,
        'stopLossPercent', 3.0,
        'autoStartRetreatPercent', 1.5,
        'profitRetreatPercent', 40,
        'monitorWindow', 480,
        'volatilityThreshold', 50,
        'reverseEnabled', true,
        'reverseLossRatio', 50,
        'reverseProfitRatio', 80,
        'aiWeightEnabled', true,
        'lowVolAmplify', true
    ),
    '低波动平衡策略：中等杠杆放大收益，耐心等待确定性机会',
    1, 11, NOW(), NOW()
);

-- 12. 低波动 - 激进型
INSERT INTO hg_trading_strategy_template (
    group_id, strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
    config_json, description, is_active, sort, created_at, updated_at
) VALUES (
    @group_id, CONCAT(@group_id, '_low_vol_aggressive'), '🚀 低波动-激进型', 'aggressive', 'low_vol',
    360, 40,
    15, 25, 15, 25,
    4.0, 2.0, 35,
    JSON_OBJECT(
        'exchange', 'bitget',
        'symbol', 'BTCUSDT',
        'orderType', 'market',
        'marginMode', 'isolated',
        'leverage', 20,
        'marginPercent', 20,
        'stopLossPercent', 4.0,
        'autoStartRetreatPercent', 2.0,
        'profitRetreatPercent', 35,
        'monitorWindow', 360,
        'volatilityThreshold', 40,
        'reverseEnabled', true,
        'reverseLossRatio', 40,
        'reverseProfitRatio', 70,
        'aiWeightEnabled', true,
        'lowVolAmplify', true
    ),
    '低波动激进策略：高杠杆最大化放大小波动收益，适合低波动横盘行情',
    1, 12, NOW(), NOW()
);

-- 查看创建结果
SELECT 
    g.group_name as '策略组',
    COUNT(t.id) as '策略数量'
FROM hg_trading_strategy_group g
LEFT JOIN hg_trading_strategy_template t ON t.group_id = g.id
WHERE g.is_official = 1
GROUP BY g.id;

SELECT 
    t.strategy_name as '策略名称',
    t.market_state as '市场状态',
    t.risk_preference as '风险偏好',
    t.monitor_window as '时间窗口(s)',
    t.volatility_threshold as '波动阈值(U)',
    CONCAT(t.leverage_min, '-', t.leverage_max, 'x') as '杠杆范围',
    CONCAT(t.margin_percent_min, '-', t.margin_percent_max, '%') as '仓位范围',
    CONCAT(t.stop_loss_percent, '%') as '止损',
    CONCAT(t.auto_start_retreat_percent, '%') as '启动止盈',
    CONCAT(t.profit_retreat_percent, '%') as '止盈回撤'
FROM hg_trading_strategy_template t
JOIN hg_trading_strategy_group g ON t.group_id = g.id
WHERE g.is_official = 1
ORDER BY t.sort;

