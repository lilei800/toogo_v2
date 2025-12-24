-- =============================================
-- 删除从赠送算力扣除的相关字段
-- 执行时间：2025-12-05
-- =============================================

-- 删除 hg_toogo_power_consume 表中的 from_gift_power 字段
ALTER TABLE `hg_toogo_power_consume` DROP COLUMN IF EXISTS `from_gift_power`;

-- 说明：
-- 1. 此迁移删除了算力消耗记录表中的 from_gift_power 字段
-- 2. 现在所有算力消耗都直接从 power 账户扣除
-- 3. 如果需要回滚，可以重新添加该字段：
--    ALTER TABLE `hg_toogo_power_consume` ADD COLUMN `from_gift_power` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '从赠送算力扣除' AFTER `from_power`;
