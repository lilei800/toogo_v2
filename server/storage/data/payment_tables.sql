-- =============================================
-- Payment System Database Tables
-- USDT充值、提现、余额管理
-- =============================================

SET NAMES utf8mb4;

-- 1. USDT余额表
CREATE TABLE IF NOT EXISTS `hg_usdt_balance` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `balance` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '可用余额',
  `frozen_balance` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '冻结余额',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_id` (`user_id`),
  KEY `idx_balance` (`balance`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='USDT余额表';

-- 2. USDT充值订单表
CREATE TABLE IF NOT EXISTS `hg_usdt_deposit` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `order_sn` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '订单号',
  `amount` decimal(20,8) NOT NULL COMMENT '充值金额',
  `network` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '网络(TRC20/ERC20)',
  `payment_id` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '第三方支付ID',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态：1待支付 2已完成 3已超时 4已退款 5已取消',
  `paid_at` datetime DEFAULT NULL COMMENT '支付时间',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_sn` (`order_sn`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='USDT充值订单表';

-- 3. USDT提现订单表
CREATE TABLE IF NOT EXISTS `hg_usdt_withdraw` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `order_sn` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '订单号',
  `amount` decimal(20,8) NOT NULL COMMENT '提现金额',
  `to_address` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '提现地址',
  `network` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '网络(TRC20/ERC20)',
  `tx_hash` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '交易哈希',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态：1待审核 2审核通过 3审核拒绝 4已完成 5已取消',
  `audit_remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '审核备注',
  `audited_by` bigint unsigned DEFAULT NULL COMMENT '审核人ID',
  `audited_at` datetime DEFAULT NULL COMMENT '审核时间',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_sn` (`order_sn`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='USDT提现订单表';

-- 4. USDT资金流水表
CREATE TABLE IF NOT EXISTS `hg_usdt_balance_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `type` tinyint NOT NULL COMMENT '类型：1充值 2提现 3支付 4退款',
  `amount` decimal(20,8) NOT NULL COMMENT '变动金额',
  `balance` decimal(20,8) NOT NULL COMMENT '变动后余额',
  `order_sn` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '关联订单号',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_type` (`type`),
  KEY `idx_order_sn` (`order_sn`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='USDT资金流水表';

-- 完成
SELECT '✅ Payment表创建完成！' as result;



