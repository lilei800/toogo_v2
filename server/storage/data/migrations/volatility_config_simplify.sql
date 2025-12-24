-- ============================================================
-- 波动率配置表简化迁移脚本
-- 移除: default_volatility_threshold
-- 新增: trend_strength_threshold, weight_1m, weight_5m, weight_15m, weight_30m, weight_1h
-- ============================================================

-- 1. 添加新字段（5个时间周期权重）
ALTER TABLE `hg_toogo_volatility_config` 
ADD COLUMN `trend_strength_threshold` DECIMAL(5,2) DEFAULT 0.35 COMMENT '趋势强度阈值（判断趋势市场）' AFTER `low_volatility_threshold`,
ADD COLUMN `weight_1m` DECIMAL(3,2) DEFAULT 0.10 COMMENT '1分钟周期权重' AFTER `trend_strength_threshold`,
ADD COLUMN `weight_5m` DECIMAL(3,2) DEFAULT 0.15 COMMENT '5分钟周期权重' AFTER `weight_1m`,
ADD COLUMN `weight_15m` DECIMAL(3,2) DEFAULT 0.25 COMMENT '15分钟周期权重' AFTER `weight_5m`,
ADD COLUMN `weight_30m` DECIMAL(3,2) DEFAULT 0.25 COMMENT '30分钟周期权重' AFTER `weight_15m`,
ADD COLUMN `weight_1h` DECIMAL(3,2) DEFAULT 0.25 COMMENT '1小时周期权重' AFTER `weight_30m`;

-- 2. 删除旧字段（可选，如需保留历史数据可跳过此步骤）
-- ALTER TABLE `hg_toogo_volatility_config` DROP COLUMN `default_volatility_threshold`;

-- 3. 更新已有数据的默认权重值
UPDATE `hg_toogo_volatility_config` SET 
    `trend_strength_threshold` = 0.35,
    `weight_1m` = 0.10,
    `weight_5m` = 0.15,
    `weight_15m` = 0.25,
    `weight_30m` = 0.25,
    `weight_1h` = 0.25
WHERE `trend_strength_threshold` IS NULL OR `trend_strength_threshold` = 0;

-- ============================================================
-- 验证SQL（执行后查看结果）
-- ============================================================
-- SELECT * FROM `hg_toogo_volatility_config`;

-- ============================================================
-- 字段说明：
-- high_volatility_threshold: 高波动阈值（波动率≥此值判定为高波动市场）
-- low_volatility_threshold: 低波动阈值（波动率≤此值判定为低波动市场）
-- trend_strength_threshold: 趋势强度阈值（趋势强度≥此值判定为趋势市场）
-- weight_1m: 1分钟K线周期权重（默认10%）
-- weight_5m: 5分钟K线周期权重（默认15%）
-- weight_15m: 15分钟K线周期权重（默认25%）
-- weight_30m: 30分钟K线周期权重（默认25%）
-- weight_1h: 1小时K线周期权重（默认25%）
-- ============================================================
