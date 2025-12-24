-- 确保 from_official_id 字段存在
-- 如果字段不存在则添加

-- 检查并添加 from_official_id 字段
SET @dbname = DATABASE();
SET @tablename = "hg_trading_strategy_group";
SET @columnname = "from_official_id";

SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = @dbname
    AND TABLE_NAME = @tablename
    AND COLUMN_NAME = @columnname
  ) > 0,
  "SELECT 'Column already exists'",
  CONCAT("ALTER TABLE ", @tablename, " ADD COLUMN `", @columnname, "` bigint DEFAULT 0 COMMENT '来源官方模板ID（复制自哪个官方模板）' AFTER `is_official`")
));

PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 添加索引（如果不存在）
SET @indexname = "idx_from_official_id";
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
    WHERE TABLE_SCHEMA = @dbname
    AND TABLE_NAME = @tablename
    AND INDEX_NAME = @indexname
  ) > 0,
  "SELECT 'Index already exists'",
  CONCAT("ALTER TABLE ", @tablename, " ADD INDEX `", @indexname, "` (`from_official_id`)")
));

PREPARE stmt FROM @preparedStatement;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 查看表结构确认
DESCRIBE hg_trading_strategy_group;

-- 查看当前的数据情况
SELECT id, group_name, is_official, from_official_id 
FROM hg_trading_strategy_group 
ORDER BY id;

