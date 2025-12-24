-- 验证波动率配置默认值更新结果

-- 显示更新后的配置详情
SELECT 
    COALESCE(`symbol`, '全局配置') AS symbol,
    `low_volatility_threshold` AS LowV,
    `trend_strength_threshold` AS TrendV,
    `high_volatility_threshold` AS HighV,
    `d_threshold` AS DThreshold,
    CONCAT(`weight_1m`, ', ', `weight_5m`, ', ', `weight_15m`, ', ', `weight_30m`, ', ', `weight_1h`) AS weights,
    (`weight_1m` + `weight_5m` + `weight_15m` + `weight_30m` + `weight_1h`) AS total_weight,
    `updated_at`
FROM `hg_toogo_volatility_config`
WHERE (`symbol` IS NULL OR `symbol` IN ('BTCUSDT', 'ETHUSDT'))
  AND `is_active` = 1
ORDER BY 
    CASE WHEN `symbol` IS NULL THEN 0 ELSE 1 END,
    `symbol`;

