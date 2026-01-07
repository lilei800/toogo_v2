-- ============================================================
-- 检查币安策略组导入情况（V3）
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
WHERE group_key IN ('official_binance_btcusdt_v3')
ORDER BY id;

-- 2. 检查策略数量
SELECT
  g.group_name AS "策略组",
  g.group_key AS "标识",
  g.is_official AS "是否官方",
  COUNT(s.id) AS "策略数量"
FROM hg_trading_strategy_group g
LEFT JOIN hg_trading_strategy_template s ON s.group_id = g.id
WHERE g.group_key IN ('official_binance_btcusdt_v3')
GROUP BY g.id, g.group_name, g.group_key, g.is_official;

-- 3. 检查策略明细
SELECT
  g.group_key AS "策略组标识",
  s.sort AS "排序",
  s.strategy_name AS "策略名称",
  s.market_state AS "市场状态",
  s.risk_preference AS "风险偏好",
  s.monitor_window AS "窗口(s)",
  s.volatility_threshold AS "阈值(USDT)",
  s.leverage AS "杠杆",
  s.margin_percent AS "保证金(%)",
  s.stop_loss_percent AS "止损(%)",
  s.auto_start_retreat_percent AS "启动止盈(%)",
  s.profit_retreat_percent AS "止盈回撤(%)",
  s.strategy_key AS "策略key"
FROM hg_trading_strategy_group g
JOIN hg_trading_strategy_template s ON s.group_id = g.id
WHERE g.group_key IN ('official_binance_btcusdt_v3')
ORDER BY g.id, s.sort;


