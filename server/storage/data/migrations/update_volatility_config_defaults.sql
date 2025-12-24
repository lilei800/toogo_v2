-- ============================================================
-- 更新波动率配置默认值
-- 1. 全局配置使用BTCUSDT默认阈值
-- 2. 为BTCUSDT和ETHUSDT创建/更新特定配置
-- ============================================================

-- 1. 更新全局配置（symbol为NULL的记录）使用BTCUSDT默认阈值
UPDATE `hg_toogo_volatility_config`
SET 
    `low_volatility_threshold` = 0.9000,      -- LowV = 0.9
    `trend_strength_threshold` = 1.2000,       -- TrendV = 1.2
    `high_volatility_threshold` = 2.0000,      -- HighV = 2.0
    `d_threshold` = 0.7000,                   -- DThreshold = 0.7
    `weight_1m` = 0.18,                        -- 1m权重 = 0.18
    `weight_5m` = 0.25,                        -- 5m权重 = 0.25
    `weight_15m` = 0.27,                       -- 15m权重 = 0.27
    `weight_30m` = 0.20,                       -- 30m权重 = 0.20
    `weight_1h` = 0.10,                        -- 1h权重 = 0.10
    `delta_1m` = COALESCE(`delta_1m`, 2.0000),
    `delta_5m` = COALESCE(`delta_5m`, 2.0000),
    `delta_15m` = COALESCE(`delta_15m`, 3.0000),
    `delta_30m` = COALESCE(`delta_30m`, 3.0000),
    `delta_1h` = COALESCE(`delta_1h`, 5.0000),
    `updated_at` = NOW()
WHERE `symbol` IS NULL;

-- 2. 为BTCUSDT创建或更新特定配置
INSERT INTO `hg_toogo_volatility_config` (
    `symbol`,
    `low_volatility_threshold`,
    `trend_strength_threshold`,
    `high_volatility_threshold`,
    `d_threshold`,
    `weight_1m`,
    `weight_5m`,
    `weight_15m`,
    `weight_30m`,
    `weight_1h`,
    `delta_1m`,
    `delta_5m`,
    `delta_15m`,
    `delta_30m`,
    `delta_1h`,
    `is_active`,
    `created_at`,
    `updated_at`
)
VALUES (
    'BTCUSDT',
    0.9000,  -- LowV = 0.9
    1.2000,  -- TrendV = 1.2
    2.0000,  -- HighV = 2.0
    0.7000,  -- DThreshold = 0.7
    0.18,    -- 1m权重 = 0.18
    0.25,    -- 5m权重 = 0.25
    0.27,    -- 15m权重 = 0.27
    0.20,    -- 30m权重 = 0.20
    0.10,    -- 1h权重 = 0.10
    2.0000,  -- delta_1m
    2.0000,  -- delta_5m
    3.0000,  -- delta_15m
    3.0000,  -- delta_30m
    5.0000,  -- delta_1h
    1,       -- is_active
    NOW(),
    NOW()
)
ON DUPLICATE KEY UPDATE
    `low_volatility_threshold` = 0.9000,
    `trend_strength_threshold` = 1.2000,
    `high_volatility_threshold` = 2.0000,
    `d_threshold` = 0.7000,
    `weight_1m` = 0.18,
    `weight_5m` = 0.25,
    `weight_15m` = 0.27,
    `weight_30m` = 0.20,
    `weight_1h` = 0.10,
    `delta_1m` = COALESCE(`delta_1m`, 2.0000),
    `delta_5m` = COALESCE(`delta_5m`, 2.0000),
    `delta_15m` = COALESCE(`delta_15m`, 3.0000),
    `delta_30m` = COALESCE(`delta_30m`, 3.0000),
    `delta_1h` = COALESCE(`delta_1h`, 5.0000),
    `updated_at` = NOW();

-- 3. 为ETHUSDT创建或更新特定配置
INSERT INTO `hg_toogo_volatility_config` (
    `symbol`,
    `low_volatility_threshold`,
    `trend_strength_threshold`,
    `high_volatility_threshold`,
    `d_threshold`,
    `weight_1m`,
    `weight_5m`,
    `weight_15m`,
    `weight_30m`,
    `weight_1h`,
    `delta_1m`,
    `delta_5m`,
    `delta_15m`,
    `delta_30m`,
    `delta_1h`,
    `is_active`,
    `created_at`,
    `updated_at`
)
VALUES (
    'ETHUSDT',
    0.8500,  -- LowV = 0.85
    1.3000,  -- TrendV = 1.3
    2.2000,  -- HighV = 2.2
    0.7000,  -- DThreshold = 0.7
    0.16,    -- 1m权重 = 0.16
    0.24,    -- 5m权重 = 0.24
    0.28,    -- 15m权重 = 0.28
    0.22,    -- 30m权重 = 0.22
    0.10,    -- 1h权重 = 0.10
    2.0000,  -- delta_1m
    2.0000,  -- delta_5m
    3.0000,  -- delta_15m
    3.0000,  -- delta_30m
    5.0000,  -- delta_1h
    1,       -- is_active
    NOW(),
    NOW()
)
ON DUPLICATE KEY UPDATE
    `low_volatility_threshold` = 0.8500,
    `trend_strength_threshold` = 1.3000,
    `high_volatility_threshold` = 2.2000,
    `d_threshold` = 0.7000,
    `weight_1m` = 0.16,
    `weight_5m` = 0.24,
    `weight_15m` = 0.28,
    `weight_30m` = 0.22,
    `weight_1h` = 0.10,
    `delta_1m` = COALESCE(`delta_1m`, 2.0000),
    `delta_5m` = COALESCE(`delta_5m`, 2.0000),
    `delta_15m` = COALESCE(`delta_15m`, 3.0000),
    `delta_30m` = COALESCE(`delta_30m`, 3.0000),
    `delta_1h` = COALESCE(`delta_1h`, 5.0000),
    `updated_at` = NOW();

-- 4. 验证更新结果
SELECT 
    '配置更新完成' AS message,
    COUNT(*) AS total_configs,
    COUNT(CASE WHEN `symbol` IS NULL THEN 1 END) AS global_configs,
    COUNT(CASE WHEN `symbol` = 'BTCUSDT' THEN 1 END) AS btcusdt_configs,
    COUNT(CASE WHEN `symbol` = 'ETHUSDT' THEN 1 END) AS ethusdt_configs
FROM `hg_toogo_volatility_config`
WHERE `is_active` = 1;

-- 5. 显示更新后的配置详情
SELECT 
    COALESCE(`symbol`, '全局配置') AS symbol,
    `low_volatility_threshold` AS LowV,
    `trend_strength_threshold` AS TrendV,
    `high_volatility_threshold` AS HighV,
    `d_threshold` AS DThreshold,
    `weight_1m` AS weight_1m,
    `weight_5m` AS weight_5m,
    `weight_15m` AS weight_15m,
    `weight_30m` AS weight_30m,
    `weight_1h` AS weight_1h,
    (`weight_1m` + `weight_5m` + `weight_15m` + `weight_30m` + `weight_1h`) AS total_weight
FROM `hg_toogo_volatility_config`
WHERE (`symbol` IS NULL OR `symbol` IN ('BTCUSDT', 'ETHUSDT'))
  AND `is_active` = 1
ORDER BY 
    CASE WHEN `symbol` IS NULL THEN 0 ELSE 1 END,
    `symbol`;

