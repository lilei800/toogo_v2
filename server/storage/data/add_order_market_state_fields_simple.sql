-- ============================================================
-- 为 hg_trading_order 表添加市场状态和风险偏好字段
-- 用于保存订单创建时的市场状态和风险偏好
-- ============================================================

-- 检查并添加 market_state 字段（如果不存在）
ALTER TABLE `hg_trading_order` 
ADD COLUMN IF NOT EXISTS `market_state` VARCHAR(50) DEFAULT NULL COMMENT '市场状态（创建订单时）' AFTER `remark`;

-- 检查并添加 risk_level 字段（如果不存在）
ALTER TABLE `hg_trading_order` 
ADD COLUMN IF NOT EXISTS `risk_level` VARCHAR(50) DEFAULT NULL COMMENT '风险偏好（创建订单时）' AFTER `market_state`;

-- 检查字段
SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_COMMENT
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'hg_trading_order'
  AND COLUMN_NAME IN ('market_state', 'risk_level');

