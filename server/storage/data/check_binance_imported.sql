-- ============================================================
-- 检查币安策略组是否已导入
-- ============================================================

-- 1. 检查币安策略组是否存在（不管is_official值）
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
   OR exchange = 'binance'
ORDER BY id;

-- 2. 检查币安策略组的策略数量
SELECT 
  g.id,
  g.group_name AS "策略组名称",
  g.group_key AS "标识",
  g.exchange AS "交易所",
  g.symbol AS "交易对",
  g.is_official AS "是否官方",
  COUNT(s.id) AS "策略数量"
FROM hg_trading_strategy_group g
LEFT JOIN hg_trading_strategy_template s ON s.group_id = g.id
WHERE g.group_key IN ('official_binance_btcusdt_v1', 'official_binance_ethusdt_v1')
   OR (g.exchange = 'binance' AND g.is_official = 1)
GROUP BY g.id, g.group_name, g.group_key, g.exchange, g.symbol, g.is_official
ORDER BY g.id;

-- 3. 如果存在但is_official不是1，显示需要修复的记录
SELECT 
  id,
  group_name AS "策略组名称",
  group_key AS "标识",
  exchange AS "交易所",
  is_official AS "当前是否官方",
  is_active AS "是否启用"
FROM hg_trading_strategy_group 
WHERE group_key IN ('official_binance_btcusdt_v1', 'official_binance_ethusdt_v1')
  AND is_official != 1;

