-- ============================================================
-- 优化 hg_trading_order 表结构
-- 添加缺失字段、唯一约束、索引
-- ============================================================

-- ============================================================
-- 一、添加缺失字段
-- ============================================================

-- 1. market_state 字段（市场状态）
ALTER TABLE `hg_trading_order` 
ADD COLUMN IF NOT EXISTS `market_state` VARCHAR(50) DEFAULT NULL COMMENT '市场状态（创建订单时）' AFTER `remark`;

-- 2. risk_level 字段（风险偏好）
ALTER TABLE `hg_trading_order` 
ADD COLUMN IF NOT EXISTS `risk_level` VARCHAR(50) DEFAULT NULL COMMENT '风险偏好（创建订单时）' AFTER `market_state`;

-- 3. strategy_group_id 字段（策略组ID）
ALTER TABLE `hg_trading_order` 
ADD COLUMN IF NOT EXISTS `strategy_group_id` BIGINT DEFAULT NULL COMMENT '策略组ID' AFTER `robot_id`;

-- 4. order_type_detail 字段（订单类型详情）
ALTER TABLE `hg_trading_order` 
ADD COLUMN IF NOT EXISTS `order_type_detail` VARCHAR(50) DEFAULT NULL COMMENT '订单类型详情：market_open_long/market_open_short等' AFTER `order_type`;

-- 5. exchange_side 字段（交易所买卖方向）
ALTER TABLE `hg_trading_order` 
ADD COLUMN IF NOT EXISTS `exchange_side` VARCHAR(10) DEFAULT NULL COMMENT '交易所买卖方向：BUY/SELL' AFTER `order_type_detail`;

-- 6. tenant_id 字段（租户ID，如果不存在）
ALTER TABLE `hg_trading_order` 
ADD COLUMN IF NOT EXISTS `tenant_id` BIGINT DEFAULT 0 COMMENT '租户ID' AFTER `user_id`;

-- 7. price 字段（委托价格）
ALTER TABLE `hg_trading_order` 
ADD COLUMN IF NOT EXISTS `price` DECIMAL(15,4) DEFAULT NULL COMMENT '委托价格' AFTER `open_price`;

-- 8. avg_price 字段（平均成交价格）
ALTER TABLE `hg_trading_order` 
ADD COLUMN IF NOT EXISTS `avg_price` DECIMAL(15,4) DEFAULT NULL COMMENT '平均成交价格' AFTER `price`;

-- 9. filled_qty 字段（已成交数量）
ALTER TABLE `hg_trading_order` 
ADD COLUMN IF NOT EXISTS `filled_qty` DECIMAL(15,8) DEFAULT NULL COMMENT '已成交数量' AFTER `quantity`;

-- 10. open_margin 字段（开仓保证金）
ALTER TABLE `hg_trading_order` 
ADD COLUMN IF NOT EXISTS `open_margin` DECIMAL(15,4) DEFAULT NULL COMMENT '开仓保证金(USDT)' AFTER `margin`;

-- 11. mark_price 字段（标记价格）
ALTER TABLE `hg_trading_order` 
ADD COLUMN IF NOT EXISTS `mark_price` DECIMAL(15,4) DEFAULT NULL COMMENT '标记价格（用于计算未实现盈亏）' AFTER `open_price`;

-- ============================================================
-- 二、添加唯一约束（防止重复订单）
-- ============================================================

-- 检查是否存在唯一索引
SET @dbname = DATABASE();
SET @tablename = 'hg_trading_order';
SET @indexname = 'uk_exchange_order_id';

SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
    WHERE
      (TABLE_SCHEMA = @dbname)
      AND (TABLE_NAME = @tablename)
      AND (INDEX_NAME = @indexname)
  ) > 0,
  'SELECT 1',
  CONCAT('ALTER TABLE `', @tablename, '` ADD UNIQUE INDEX `', @indexname, '` (`exchange_order_id`)')
));

PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- ============================================================
-- 三、添加必要索引（优化查询性能）
-- ============================================================

-- 1. robot_id + status 索引（查询机器人持仓订单）
SET @indexname = 'idx_robot_status';
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
    WHERE
      (TABLE_SCHEMA = @dbname)
      AND (TABLE_NAME = @tablename)
      AND (INDEX_NAME = @indexname)
  ) > 0,
  'SELECT 1',
  CONCAT('ALTER TABLE `', @tablename, '` ADD INDEX `', @indexname, '` (`robot_id`, `status`)')
));

PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- 2. user_id + status 索引（查询用户订单）
SET @indexname = 'idx_user_status';
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
    WHERE
      (TABLE_SCHEMA = @dbname)
      AND (TABLE_NAME = @tablename)
      AND (INDEX_NAME = @indexname)
  ) > 0,
  'SELECT 1',
  CONCAT('ALTER TABLE `', @tablename, '` ADD INDEX `', @indexname, '` (`user_id`, `status`)')
));

PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- 3. symbol + status 索引（查询交易对订单）
SET @indexname = 'idx_symbol_status';
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
    WHERE
      (TABLE_SCHEMA = @dbname)
      AND (TABLE_NAME = @tablename)
      AND (INDEX_NAME = @indexname)
  ) > 0,
  'SELECT 1',
  CONCAT('ALTER TABLE `', @tablename, '` ADD INDEX `', @indexname, '` (`symbol`, `status`)')
));

PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- 4. created_at 索引（按时间排序）
SET @indexname = 'idx_created_at';
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
    WHERE
      (TABLE_SCHEMA = @dbname)
      AND (TABLE_NAME = @tablename)
      AND (INDEX_NAME = @indexname)
  ) > 0,
  'SELECT 1',
  CONCAT('ALTER TABLE `', @tablename, '` ADD INDEX `', @indexname, '` (`created_at`)')
));

PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- 5. exchange_order_id 索引（查询交易所订单，如果唯一索引已存在则跳过）
SET @indexname = 'idx_exchange_order_id';
SET @preparedStatement = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
    WHERE
      (TABLE_SCHEMA = @dbname)
      AND (TABLE_NAME = @tablename)
      AND (INDEX_NAME = @indexname)
  ) > 0,
  'SELECT 1',
  CONCAT('ALTER TABLE `', @tablename, '` ADD INDEX `', @indexname, '` (`exchange_order_id`)')
));

PREPARE alterIfNotExists FROM @preparedStatement;
EXECUTE alterIfNotExists;
DEALLOCATE PREPARE alterIfNotExists;

-- ============================================================
-- 四、验证结果
-- ============================================================

-- 检查字段
SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_COMMENT
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'hg_trading_order'
  AND COLUMN_NAME IN ('market_state', 'risk_level', 'strategy_group_id', 'order_type_detail', 'exchange_side', 'tenant_id', 'price', 'avg_price', 'filled_qty', 'open_margin', 'mark_price')
ORDER BY ORDINAL_POSITION;

-- 检查索引
SELECT INDEX_NAME, COLUMN_NAME, NON_UNIQUE, SEQ_IN_INDEX
FROM INFORMATION_SCHEMA.STATISTICS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'hg_trading_order'
  AND INDEX_NAME IN ('uk_exchange_order_id', 'idx_robot_status', 'idx_user_status', 'idx_symbol_status', 'idx_created_at', 'idx_exchange_order_id')
ORDER BY INDEX_NAME, SEQ_IN_INDEX;

