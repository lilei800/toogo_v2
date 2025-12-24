-- ============================================================
-- 波动率配置表优化：适配新算法（修复版，兼容MySQL）
-- 添加delta字段（各周期波动点数阈值）和DThreshold（方向一致性阈值）
-- ============================================================

-- 1. 添加各周期的delta字段（使用ALTER TABLE，兼容MySQL）
ALTER TABLE `hg_toogo_volatility_config` 
ADD COLUMN `delta_1m` decimal(10,4) DEFAULT 2.0000 COMMENT '1分钟周期波动点数阈值delta',
ADD COLUMN `delta_5m` decimal(10,4) DEFAULT 2.0000 COMMENT '5分钟周期波动点数阈值delta',
ADD COLUMN `delta_15m` decimal(10,4) DEFAULT 3.0000 COMMENT '15分钟周期波动点数阈值delta',
ADD COLUMN `delta_30m` decimal(10,4) DEFAULT 3.0000 COMMENT '30分钟周期波动点数阈值delta',
ADD COLUMN `delta_1h` decimal(10,4) DEFAULT 5.0000 COMMENT '1小时周期波动点数阈值delta';

-- 2. 添加方向一致性阈值DThreshold
ALTER TABLE `hg_toogo_volatility_config` 
ADD COLUMN `d_threshold` decimal(10,4) DEFAULT 0.7000 COMMENT '方向一致性阈值（用于判断趋势市场，0-1之间）';

-- 3. 更新现有记录的默认值
UPDATE `hg_toogo_volatility_config` 
SET 
    `delta_1m` = COALESCE(`delta_1m`, 2.0000),
    `delta_5m` = COALESCE(`delta_5m`, 2.0000),
    `delta_15m` = COALESCE(`delta_15m`, 3.0000),
    `delta_30m` = COALESCE(`delta_30m`, 3.0000),
    `delta_1h` = COALESCE(`delta_1h`, 5.0000),
    `d_threshold` = COALESCE(`d_threshold`, 0.7000)
WHERE `delta_1m` IS NULL OR `delta_1m` = 0;

-- 4. 验证字段添加成功
SELECT 
    'Fields added successfully!' AS message,
    COUNT(*) AS total_configs,
    COUNT(CASE WHEN `delta_1m` IS NOT NULL THEN 1 END) AS configs_with_delta,
    COUNT(CASE WHEN `d_threshold` IS NOT NULL THEN 1 END) AS configs_with_dthreshold
FROM `hg_toogo_volatility_config`;

