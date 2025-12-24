-- ====================================
-- 策略模板字段优化：将区间值改为固定值
-- 创建时间: 2025-12-03
-- ====================================

-- 1. 备份现有数据
CREATE TABLE IF NOT EXISTS `hg_trading_strategy_template_backup_20251203` AS 
SELECT * FROM `hg_trading_strategy_template`;

-- 2. 添加新的固定值字段
ALTER TABLE `hg_trading_strategy_template` 
  ADD COLUMN `leverage` int(11) DEFAULT NULL COMMENT '杠杆倍数' AFTER `volatility_threshold`,
  ADD COLUMN `margin_percent` decimal(5,2) DEFAULT NULL COMMENT '保证金比例(%)' AFTER `leverage`;

-- 3. 数据迁移：使用区间的最小值（更保守）
UPDATE `hg_trading_strategy_template` 
SET 
  `leverage` = COALESCE(`leverage_min`, 5),
  `margin_percent` = COALESCE(`margin_percent_min`, 10.00)
WHERE `leverage` IS NULL OR `margin_percent` IS NULL;

-- 4. 设置默认值
ALTER TABLE `hg_trading_strategy_template` 
  MODIFY COLUMN `leverage` int(11) DEFAULT '5' COMMENT '杠杆倍数',
  MODIFY COLUMN `margin_percent` decimal(5,2) DEFAULT '10.00' COMMENT '保证金比例(%)';

-- 5. 删除旧的区间字段
ALTER TABLE `hg_trading_strategy_template` 
  DROP COLUMN `leverage_min`,
  DROP COLUMN `leverage_max`,
  DROP COLUMN `margin_percent_min`,
  DROP COLUMN `margin_percent_max`;

-- 6. 验证迁移结果
SELECT 
  COUNT(*) as total_count,
  COUNT(CASE WHEN leverage IS NULL THEN 1 END) as null_leverage_count,
  COUNT(CASE WHEN margin_percent IS NULL THEN 1 END) as null_margin_count,
  MIN(leverage) as min_leverage,
  MAX(leverage) as max_leverage,
  MIN(margin_percent) as min_margin,
  MAX(margin_percent) as max_margin
FROM `hg_trading_strategy_template`;

-- 7. 查看样本数据
SELECT id, strategy_name, leverage, margin_percent, stop_loss_percent 
FROM `hg_trading_strategy_template` 
LIMIT 10;

-- 回滚SQL（如需要）：
-- DROP TABLE `hg_trading_strategy_template`;
-- RENAME TABLE `hg_trading_strategy_template_backup_20251203` TO `hg_trading_strategy_template`;

















