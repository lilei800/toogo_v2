-- ============================================================
-- 为 hg_trading_robot 表添加 strategy_group_id 字段
-- 如果字段已存在，此脚本不会报错（使用 IF NOT EXISTS）
-- ============================================================

-- 1. 添加 strategy_group_id 字段（如果不存在）
SET @dbname = DATABASE();
SET @tablename = 'hg_trading_robot';
SET @columnname = 'strategy_group_id';
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE
      (TABLE_SCHEMA = @dbname)
      AND (TABLE_NAME = @tablename)
      AND (COLUMN_NAME = @columnname)
  ) > 0,
  'SELECT 1', -- 字段已存在，不执行任何操作
  CONCAT('ALTER TABLE ', @tablename, ' ADD COLUMN ', @columnname, ' bigint(20) DEFAULT NULL COMMENT ''策略组ID'' AFTER current_strategy')
));
PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- 2. 添加索引（如果不存在）
SET @indexname = 'idx_strategy_group';
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
    WHERE
      (TABLE_SCHEMA = @dbname)
      AND (TABLE_NAME = @tablename)
      AND (INDEX_NAME = @indexname)
  ) > 0,
  'SELECT 1', -- 索引已存在，不执行任何操作
  CONCAT('ALTER TABLE ', @tablename, ' ADD INDEX ', @indexname, ' (strategy_group_id)')
));
PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- ============================================================
-- 更新现有机器人的策略组ID示例
-- ============================================================

-- 示例1: 更新单个机器人的策略组ID
-- UPDATE `hg_trading_robot` 
-- SET `strategy_group_id` = 18, 
--     `updated_at` = NOW()
-- WHERE `id` = <机器人ID>;

-- 示例2: 批量更新所有未设置策略组ID的机器人（设置为默认策略组）
-- UPDATE `hg_trading_robot` 
-- SET `strategy_group_id` = 18, 
--     `updated_at` = NOW()
-- WHERE `strategy_group_id` IS NULL 
--   AND `deleted_at` IS NULL;

-- 示例3: 根据备注中的策略组ID信息更新（如果备注格式为 "策略组ID: 18"）
-- UPDATE `hg_trading_robot` 
-- SET `strategy_group_id` = CAST(SUBSTRING_INDEX(SUBSTRING_INDEX(`remark`, '策略组ID: ', -1), ' ', 1) AS UNSIGNED),
--     `updated_at` = NOW()
-- WHERE `remark` LIKE '%策略组ID: %'
--   AND `strategy_group_id` IS NULL
--   AND `deleted_at` IS NULL;

-- ============================================================
-- 查询检查
-- ============================================================

-- 检查字段是否存在
-- SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_COMMENT
-- FROM INFORMATION_SCHEMA.COLUMNS
-- WHERE TABLE_SCHEMA = DATABASE()
--   AND TABLE_NAME = 'hg_trading_robot'
--   AND COLUMN_NAME = 'strategy_group_id';

-- 查看所有机器人的策略组ID情况
-- SELECT 
--     id,
--     robot_name,
--     strategy_group_id,
--     remark,
--     status,
--     created_at
-- FROM hg_trading_robot
-- WHERE deleted_at IS NULL
-- ORDER BY id DESC
-- LIMIT 20;

-- 查看没有设置策略组ID的机器人
-- SELECT 
--     id,
--     robot_name,
--     strategy_group_id,
--     remark,
--     status
-- FROM hg_trading_robot
-- WHERE deleted_at IS NULL
--   AND strategy_group_id IS NULL
-- ORDER BY id DESC;

