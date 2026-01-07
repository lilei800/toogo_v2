-- ============================================================
-- 检查币安策略组导入情况
-- ============================================================

-- 1. 检查策略组是否存在
SELECT 
  id,
  group_name AS "策略组名称",
  group_key AS "标识",
  exchange AS "交易所",
  symbol AS "交易对",
  is_official AS "是否官方",
  is_active AS "是否启用",
  user_id AS "用户ID",
  created_at AS "创建时间"
FROM hg_trading_strategy_group 
WHERE group_key IN ('official_binance_btcusdt_v1', 'official_binance_ethusdt_v1')
ORDER BY id;

-- 2. 检查策略数量
SELECT 
  g.group_name AS "策略组",
  g.group_key AS "标识",
  g.is_official AS "是否官方",
  COUNT(s.id) AS "策略数量"
FROM hg_trading_strategy_group g
LEFT JOIN hg_trading_strategy_template s ON s.group_id = g.id
WHERE g.group_key IN ('official_binance_btcusdt_v1', 'official_binance_ethusdt_v1')
GROUP BY g.id, g.group_name, g.group_key, g.is_official;

-- 3. 检查所有官方策略组
SELECT 
  id,
  group_name AS "策略组名称",
  group_key AS "标识",
  exchange AS "交易所",
  symbol AS "交易对",
  is_official AS "是否官方",
  is_active AS "是否启用"
FROM hg_trading_strategy_group 
WHERE is_official = 1
ORDER BY sort ASC, id ASC;

-- 4. 如果看不到，检查是否有数据但is_official字段不对
SELECT 
  id,
  group_name AS "策略组名称",
  group_key AS "标识",
  exchange AS "交易所",
  symbol AS "交易对",
  is_official AS "是否官方",
  is_active AS "是否启用"
FROM hg_trading_strategy_group 
WHERE group_key LIKE '%binance%'
ORDER BY id;

