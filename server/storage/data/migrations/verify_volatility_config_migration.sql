-- 验证波动率配置表迁移结果

-- 1. 检查字段是否存在
SELECT 
    COLUMN_NAME,
    COLUMN_TYPE,
    COLUMN_DEFAULT,
    COLUMN_COMMENT
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'hg_toogo_volatility_config'
  AND COLUMN_NAME IN ('delta_1m', 'delta_5m', 'delta_15m', 'delta_30m', 'delta_1h', 'd_threshold')
ORDER BY ORDINAL_POSITION;

-- 2. 检查现有配置的默认值
SELECT 
    id,
    symbol,
    delta_1m,
    delta_5m,
    delta_15m,
    delta_30m,
    delta_1h,
    d_threshold,
    high_volatility_threshold,
    low_volatility_threshold,
    trend_strength_threshold
FROM hg_toogo_volatility_config
LIMIT 10;

-- 3. 统计配置数量
SELECT 
    COUNT(*) AS total_configs,
    COUNT(CASE WHEN delta_1m IS NOT NULL THEN 1 END) AS configs_with_delta,
    COUNT(CASE WHEN d_threshold IS NOT NULL THEN 1 END) AS configs_with_dthreshold
FROM hg_toogo_volatility_config;

