-- 为策略组表添加 is_visible 字段（用于控制策略组是否对用户可见）
-- 如果字段已存在则跳过

-- 检查字段是否存在，如果不存在则添加
SET @dbname = DATABASE();
SET @tablename = 'hg_trading_strategy_group';
SET @columnname = 'is_visible';
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE
      (TABLE_SCHEMA = @dbname)
      AND (TABLE_NAME = @tablename)
      AND (COLUMN_NAME = @columnname)
  ) > 0,
  'SELECT 1', -- 字段已存在，不执行任何操作
  CONCAT('ALTER TABLE ', @tablename, ' ADD COLUMN ', @columnname, ' TINYINT(1) NOT NULL DEFAULT 1 COMMENT ''是否可见: 0=隐藏, 1=显示''')
));
PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- 为现有记录设置默认值（可见）
UPDATE hg_trading_strategy_group SET is_visible = 1 WHERE is_visible IS NULL OR is_visible = 0;

