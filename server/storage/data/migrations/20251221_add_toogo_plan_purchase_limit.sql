-- ============================================================
-- 为 hg_toogo_plan 增加套餐级别“购买次数限制”字段
-- 说明：
-- - purchase_limit: 购买次数限制（0为不限）
-- - 套餐管理页面将使用该字段做“是否限购/限购次数”
-- ============================================================

SET NAMES utf8mb4;

-- 兼容旧库：可能不存在 default_period 字段，因此不依赖 AFTER 子句
-- MySQL 8.0+ 支持 ADD COLUMN IF NOT EXISTS（本机 mysql 8.0.37 已支持）
ALTER TABLE `hg_toogo_plan`
  ADD COLUMN IF NOT EXISTS `purchase_limit` INT NOT NULL DEFAULT 0 COMMENT '购买次数限制(0为不限)';


