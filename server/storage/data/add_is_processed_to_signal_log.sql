-- ============================================================
-- 为 hg_trading_signal_log 表添加 is_processed 字段
-- 用于标记预警记录是否已被处理（防止并发重复下单）
-- ============================================================

-- 使用存储过程检查并添加字段（兼容MySQL 5.7+）
DELIMITER $$

DROP PROCEDURE IF EXISTS add_is_processed_column$$

CREATE PROCEDURE add_is_processed_column()
BEGIN
    DECLARE column_exists INT DEFAULT 0;
    
    -- 检查字段是否存在
    SELECT COUNT(*) INTO column_exists
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'hg_trading_signal_log'
      AND COLUMN_NAME = 'is_processed';
    
    -- 如果字段不存在，则添加
    IF column_exists = 0 THEN
        ALTER TABLE `hg_trading_signal_log` 
        ADD COLUMN `is_processed` TINYINT(1) NOT NULL DEFAULT 0 
        COMMENT '已读标识：0=未处理，1=已处理（用于防止重复下单）' 
        AFTER `executed`;
        
        SELECT '字段 is_processed 已成功添加' AS result;
    ELSE
        SELECT '字段 is_processed 已存在，跳过添加' AS result;
    END IF;
END$$

DELIMITER ;

-- 执行存储过程
CALL add_is_processed_column();

-- 删除存储过程
DROP PROCEDURE IF EXISTS add_is_processed_column;

-- 检查字段
SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_COMMENT
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'hg_trading_signal_log'
  AND COLUMN_NAME = 'is_processed';

-- 为已存在的记录设置默认值（确保所有记录都是未处理状态）
UPDATE `hg_trading_signal_log` SET `is_processed` = 0 WHERE `is_processed` IS NULL;

