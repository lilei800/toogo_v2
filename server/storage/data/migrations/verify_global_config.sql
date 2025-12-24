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
  AND deleted_at IS NULL
ORDER BY updated_at DESC
LIMIT 1;

