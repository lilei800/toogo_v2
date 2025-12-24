-- ============================================
-- BTCUSDT 和 ETHUSDT 推荐配置更新
-- 根据新市场状态算法生成
-- ============================================

-- BTCUSDT 配置
INSERT INTO hg_toogo_volatility_config (
    symbol,
    low_volatility_threshold,
    high_volatility_threshold,
    trend_strength_threshold,
    d_threshold,
    delta_1m, delta_5m, delta_15m, delta_30m, delta_1h,
    weight_1m, weight_5m, weight_15m, weight_30m, weight_1h,
    is_active,
    created_at,
    updated_at
) VALUES (
    'BTCUSDT',
    0.5,   -- LowV: V < 0.5 时判断为低波动
    2.0,   -- HighV: V >= 2.0 且 D < 0.4 时判断为高波动
    1.2,   -- TrendV: V >= 1.2 且 D >= DThreshold 时判断为趋势
    0.7,   -- DThreshold: 趋势判断时，D >= 0.7 表示方向一致性强
    30.0,  -- Delta1m: 1分钟内正常波动30 USDT仍属于低波动
    60.0,  -- Delta5m: 5分钟内正常波动60 USDT仍属于低波动
    150.0, -- Delta15m: 15分钟内正常波动150 USDT仍属于低波动
    300.0, -- Delta30m: 30分钟内正常波动300 USDT仍属于低波动
    600.0, -- Delta1h: 1小时内正常波动600 USDT仍属于低波动
    0.15,  -- Weight1m: 15%
    0.20,  -- Weight5m: 20%
    0.30,  -- Weight15m: 30%（核心周期）
    0.25,  -- Weight30m: 25%
    0.10,  -- Weight1h: 10%
    1,     -- 启用
    NOW(),
    NOW()
) ON DUPLICATE KEY UPDATE
    low_volatility_threshold = 0.5,
    high_volatility_threshold = 2.0,
    trend_strength_threshold = 1.2,
    d_threshold = 0.7,
    delta_1m = 30.0,
    delta_5m = 60.0,
    delta_15m = 150.0,
    delta_30m = 300.0,
    delta_1h = 600.0,
    weight_1m = 0.15,
    weight_5m = 0.20,
    weight_15m = 0.30,
    weight_30m = 0.25,
    weight_1h = 0.10,
    updated_at = NOW();

-- ETHUSDT 配置
INSERT INTO hg_toogo_volatility_config (
    symbol,
    low_volatility_threshold,
    high_volatility_threshold,
    trend_strength_threshold,
    d_threshold,
    delta_1m, delta_5m, delta_15m, delta_30m, delta_1h,
    weight_1m, weight_5m, weight_15m, weight_30m, weight_1h,
    is_active,
    created_at,
    updated_at
) VALUES (
    'ETHUSDT',
    0.5,   -- LowV: V < 0.5 时判断为低波动
    2.0,   -- HighV: V >= 2.0 且 D < 0.4 时判断为高波动
    1.2,   -- TrendV: V >= 1.2 且 D >= DThreshold 时判断为趋势
    0.7,   -- DThreshold: 趋势判断时，D >= 0.7 表示方向一致性强
    20.0,  -- Delta1m: 1分钟内正常波动20 USDT仍属于低波动
    40.0,  -- Delta5m: 5分钟内正常波动40 USDT仍属于低波动
    100.0, -- Delta15m: 15分钟内正常波动100 USDT仍属于低波动
    200.0, -- Delta30m: 30分钟内正常波动200 USDT仍属于低波动
    400.0, -- Delta1h: 1小时内正常波动400 USDT仍属于低波动
    0.15,  -- Weight1m: 15%
    0.20,  -- Weight5m: 20%
    0.30,  -- Weight15m: 30%（核心周期）
    0.25,  -- Weight30m: 25%
    0.10,  -- Weight1h: 10%
    1,     -- 启用
    NOW(),
    NOW()
) ON DUPLICATE KEY UPDATE
    low_volatility_threshold = 0.5,
    high_volatility_threshold = 2.0,
    trend_strength_threshold = 1.2,
    d_threshold = 0.7,
    delta_1m = 20.0,
    delta_5m = 40.0,
    delta_15m = 100.0,
    delta_30m = 200.0,
    delta_1h = 400.0,
    weight_1m = 0.15,
    weight_5m = 0.20,
    weight_15m = 0.30,
    weight_30m = 0.25,
    weight_1h = 0.10,
    updated_at = NOW();

-- 验证配置
SELECT 
    symbol,
    low_volatility_threshold AS LowV,
    high_volatility_threshold AS HighV,
    trend_strength_threshold AS TrendV,
    d_threshold AS DThreshold,
    delta_1m, delta_5m, delta_15m, delta_30m, delta_1h,
    weight_1m, weight_5m, weight_15m, weight_30m, weight_1h,
    (weight_1m + weight_5m + weight_15m + weight_30m + weight_1h) AS weight_sum,
    is_active,
    updated_at
FROM hg_toogo_volatility_config
WHERE symbol IN ('BTCUSDT', 'ETHUSDT')
ORDER BY symbol;

