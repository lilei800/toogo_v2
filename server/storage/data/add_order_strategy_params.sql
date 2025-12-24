-- ============================================================
-- 为 hg_trading_order 表添加策略参数字段
-- 用于保存订单创建时的策略参数（止损、启动止盈等）
-- ============================================================

-- 1. 添加 stop_loss_percent 字段（如果不存在）
SET @dbname = DATABASE();
SET @tablename = 'hg_trading_order';
SET @columnname = 'stop_loss_percent';
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE
      (TABLE_SCHEMA = @dbname)
      AND (TABLE_NAME = @tablename)
      AND (COLUMN_NAME = @columnname)
  ) > 0,
  'SELECT 1', -- 字段已存在，不执行任何操作
  CONCAT('ALTER TABLE ', @tablename, ' ADD COLUMN ', @columnname, ' decimal(5,2) DEFAULT NULL COMMENT ''止损百分比(%)'' AFTER profit_retreat_percent')
));
PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- 2. 添加 auto_start_retreat_percent 字段（如果不存在）
SET @columnname = 'auto_start_retreat_percent';
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE
      (TABLE_SCHEMA = @dbname)
      AND (TABLE_NAME = @tablename)
      AND (COLUMN_NAME = @columnname)
  ) > 0,
  'SELECT 1', -- 字段已存在，不执行任何操作
  CONCAT('ALTER TABLE ', @tablename, ' ADD COLUMN ', @columnname, ' decimal(5,2) DEFAULT NULL COMMENT ''启动止盈百分比(%)'' AFTER stop_loss_percent')
));
PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- 3. 添加 margin_percent 字段（如果不存在，用于保存创建订单时的保证金比例）
SET @columnname = 'margin_percent';
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE
      (TABLE_SCHEMA = @dbname)
      AND (TABLE_NAME = @tablename)
      AND (COLUMN_NAME = @columnname)
  ) > 0,
  'SELECT 1', -- 字段已存在，不执行任何操作
  CONCAT('ALTER TABLE ', @tablename, ' ADD COLUMN ', @columnname, ' decimal(5,2) DEFAULT NULL COMMENT ''保证金比例(%)'' AFTER leverage')
));
PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- ============================================================
-- 查询检查
-- ============================================================

-- 检查字段是否存在
-- SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_COMMENT
-- FROM INFORMATION_SCHEMA.COLUMNS
-- WHERE TABLE_SCHEMA = DATABASE()
--   AND TABLE_NAME = 'hg_trading_order'
--   AND COLUMN_NAME IN ('stop_loss_percent', 'auto_start_retreat_percent', 'margin_percent');

