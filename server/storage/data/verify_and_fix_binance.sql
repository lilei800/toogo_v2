-- ============================================================
-- 验证并修复币安官方策略组
-- ============================================================

-- 步骤1：检查我们创建的策略组是否存在
SELECT 
  id,
  group_name AS "策略组名称",
  group_key AS "标识",
  exchange AS "交易所",
  symbol AS "交易对",
  is_official AS "是否官方",
  is_active AS "是否启用"
FROM hg_trading_strategy_group 
WHERE group_key IN ('official_binance_btcusdt_v1', 'official_binance_ethusdt_v1');

-- 步骤2：如果上面查询为空，说明策略组没有导入成功
-- 请执行以下SQL文件重新导入：
-- 1. official_binance_btcusdt_v1_pg.sql
-- 2. official_binance_ethusdt_v1_pg.sql

-- 步骤3：如果上面查询有结果但is_official不是1，执行以下修复
UPDATE hg_trading_strategy_group 
SET 
    is_official = 1,
    user_id = 0,
    is_active = 1,
    updated_at = NOW()
WHERE group_key IN ('official_binance_btcusdt_v1', 'official_binance_ethusdt_v1')
  AND is_official != 1;

-- 步骤4：验证修复结果
SELECT 
  id,
  group_name AS "策略组名称",
  group_key AS "标识",
  exchange AS "交易所",
  symbol AS "交易对",
  is_official AS "是否官方",
  is_active AS "是否启用",
  (SELECT COUNT(*) FROM hg_trading_strategy_template WHERE group_id = g.id) AS "策略数量"
FROM hg_trading_strategy_group g
WHERE group_key IN ('official_binance_btcusdt_v1', 'official_binance_ethusdt_v1')
ORDER BY id;

-- 步骤5：查看所有官方策略组（应该能看到币安的）
SELECT 
  id,
  group_name AS "策略组名称",
  exchange AS "交易所",
  symbol AS "交易对",
  is_official AS "是否官方",
  (SELECT COUNT(*) FROM hg_trading_strategy_template WHERE group_id = g.id) AS "策略数量"
FROM hg_trading_strategy_group g
WHERE is_official = 1
ORDER BY exchange, sort ASC, id ASC;

