-- ============================================================
-- 量化管理波动率配置表
-- 支持为每个货币对（交易对）配置独立的波动率阈值
-- ============================================================

-- 创建波动率配置表
CREATE TABLE IF NOT EXISTS `hg_toogo_volatility_config` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `symbol` varchar(50) DEFAULT NULL COMMENT '交易对（NULL表示全局配置，如：BTCUSDT表示BTCUSDT特定配置）',
  `high_volatility_threshold` decimal(10,4) DEFAULT 2.0000 COMMENT '高波动阈值',
  `low_volatility_threshold` decimal(10,4) DEFAULT 0.5000 COMMENT '低波动阈值',
  `default_volatility_threshold` decimal(10,4) DEFAULT 1.0000 COMMENT '默认波动阈值',
  `is_active` tinyint(1) DEFAULT 1 COMMENT '是否启用: 0=否, 1=是',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_symbol` (`symbol`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='量化管理波动率配置表（支持每个货币对独立配置）';

-- 插入默认全局配置
INSERT INTO `hg_toogo_volatility_config` 
(`symbol`, `high_volatility_threshold`, `low_volatility_threshold`, `default_volatility_threshold`, `is_active`, `created_at`, `updated_at`) 
SELECT NULL, 2.0000, 0.5000, 1.0000, 1, NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM `hg_toogo_volatility_config` WHERE `symbol` IS NULL
);

-- 验证表创建成功
SELECT 
    'Table created successfully!' AS message,
    COUNT(*) AS total_configs,
    (SELECT COUNT(*) FROM `hg_toogo_volatility_config` WHERE `symbol` IS NULL) AS global_configs,
    (SELECT COUNT(*) FROM `hg_toogo_volatility_config` WHERE `symbol` IS NOT NULL) AS symbol_specific_configs
FROM `hg_toogo_volatility_config`;

