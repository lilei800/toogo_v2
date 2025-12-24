-- ============================================
-- 更新全局波动率配置
-- 使用BTCUSDT推荐配置作为全局默认配置
-- ============================================

-- 更新全局配置（symbol IS NULL 或 symbol = ''）
UPDATE hg_toogo_volatility_config
SET
    low_volatility_threshold = 0.5,   -- LowV: V < 0.5 时判断为低波动
    high_volatility_threshold = 2.0,  -- HighV: V >= 2.0 且 D < 0.4 时判断为高波动
    trend_strength_threshold = 1.2,   -- TrendV: V >= 1.2 且 D >= DThreshold 时判断为趋势
    d_threshold = 0.7,                -- DThreshold: 趋势判断时，D >= 0.7 表示方向一致性强
    delta_1m = 30.0,                  -- Delta1m: 1分钟内正常波动30 USDT仍属于低波动
    delta_5m = 60.0,                  -- Delta5m: 5分钟内正常波动60 USDT仍属于低波动
    delta_15m = 150.0,                -- Delta15m: 15分钟内正常波动150 USDT仍属于低波动
    delta_30m = 300.0,                -- Delta30m: 30分钟内正常波动300 USDT仍属于低波动
    delta_1h = 600.0,                 -- Delta1h: 1小时内正常波动600 USDT仍属于低波动
    weight_1m = 0.15,                 -- Weight1m: 15%
    weight_5m = 0.20,                 -- Weight5m: 20%
    weight_15m = 0.30,                -- Weight15m: 30%（核心周期）
    weight_30m = 0.25,                -- Weight30m: 25%
    weight_1h = 0.10,                 -- Weight1h: 10%
    updated_at = NOW()
WHERE (symbol IS NULL OR symbol = '');

-- 如果全局配置不存在，则创建
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
)
SELECT 
    NULL,  -- symbol为NULL表示全局配置
    0.5,   -- LowV
    2.0,   -- HighV
    1.2,   -- TrendV
    0.7,   -- DThreshold
    30.0,  -- Delta1m
    60.0,  -- Delta5m
    150.0, -- Delta15m
    300.0, -- Delta30m
    600.0, -- Delta1h
    0.15,  -- Weight1m
    0.20,  -- Weight5m
    0.30,  -- Weight15m
    0.25,  -- Weight30m
    0.10,  -- Weight1h
    1,     -- 启用
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 
    FROM hg_toogo_volatility_config 
    WHERE (symbol IS NULL OR symbol = '')
);

-- 验证全局配置
SELECT 
    CASE 
        WHEN symbol IS NULL THEN '全局配置'
        WHEN symbol = '' THEN '全局配置'
        ELSE symbol
    END AS config_type,
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
WHERE (symbol IS NULL OR symbol = '')
ORDER BY updated_at DESC
LIMIT 1;
