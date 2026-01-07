-- ============================================================
-- Binance BTC-USDT 官方策略 V3.0（Fee-aware，MySQL版本）
-- 创建时间: 2026-01-01
--
-- 目标：
-- - 在 V2 的基础上，把“手续费 + 滑点”的盈亏平衡点显式纳入参数设计。
-- - 仍保持：4种市场状态 × 3种风险偏好 = 12套策略模板。
--
-- 链路说明（项目实现口径）：
-- - WS行情(Binance Futures) → MarketServiceManager 缓存 → MarketAnalyzer(全局) 产出 market_state
-- - RobotEngine.OnPriceUpdate → 窗口阈值触发（monitor_window + volatility_threshold）→ 方向预警(signal_log)
-- - RobotTrader.executeOpen：按模板计算 leverage/margin_percent → 市价开仓(CreateOrder)
-- - 平仓：止损/追踪止盈触发 → ClosePosition → 成交聚合/手续费汇总 → CloseOrder 落库结算
--
-- 手续费/滑点假设（务必根据账户等级校准）：
-- - Binance U本位合约市价单一般为 Taker，常见档位：0.04%/边（0.0004）
-- - 额外滑点/冲击成本：保守按 0.02%/边（0.0002）
-- - round-trip 总成本（名义价值口径）≈ 0.0004*2 + 0.0002*2 = 0.0012 = 0.12%
--
-- 盈亏平衡点（保证金口径）：
-- - 费用按“名义价值”计费，而策略参数的盈亏百分比按“保证金”口径展示
-- - 费用(保证金%) ≈ leverage × 0.12%
-- - 为避免“刚启动止盈就被回撤触发但净利润被费用吞噬”，本策略约束：
--     auto_start_retreat_percent × (1 - profit_retreat_percent/100)  >=  leverage × 0.12%  + buffer
--   其中 buffer 默认取 0.6%（保证金口径），用于覆盖偶发滑点/延迟/成交偏差。
--
-- 重要声明：
-- - 市场存在不确定性，本文件仅提供参数模板，无法保证实际盈利。
-- ============================================================

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET NAMES utf8mb4;

-- 清理旧数据（可重复导入）
DELETE FROM `hg_trading_strategy_template` WHERE `group_id` IN (
  SELECT `id` FROM `hg_trading_strategy_group` WHERE `group_key` = 'official_binance_btcusdt_v3'
);
DELETE FROM `hg_trading_strategy_group` WHERE `group_key` = 'official_binance_btcusdt_v3';

-- 插入官方策略组 V3.0
INSERT INTO `hg_trading_strategy_group` (
  `group_name`, `group_key`, `exchange`, `symbol`, `order_type`, `margin_mode`,
  `is_official`, `user_id`, `description`, `is_active`, `sort`
) VALUES (
  '🔥 Binance BTC-USDT 官方策略 V3.0（Fee-aware）',
  'official_binance_btcusdt_v3',
  'binance',
  'BTCUSDT',
  'market',
  'isolated',
  1,
  0,
  '币安BTCUSDT官方策略V3.0（手续费/滑点盈亏平衡点纳入参数约束）。闭环：WS行情→市场状态→模板选择→方向预警→市价开仓→止损/追踪止盈平仓→成交流水聚合补齐平仓价/手续费/已实现盈亏。推荐映射：trend->balanced, volatile->balanced, high_vol->conservative, low_vol->balanced。',
  1,
  5
);

SET @group_id = LAST_INSERT_ID();

-- 统一的费用模型（写入 config_json，仅用于解释；运行时主要使用硬字段）
-- roundTripNotionalCost=0.0012 (0.12% notional)
-- bufferMarginPct=0.6 (0.6% of margin)

-- ==================== 🛡️ 保守型（更低杠杆/更低保证金/更高阈值） ====================
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`,
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `auto_start_retreat_percent`, `profit_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES
(@group_id, 'binance_btc_v3_conservative_trend', '🛡️ 保守-趋势跟踪 (V3)',
 'conservative', 'trend', 480, 130.00, 3, 7.00, 3.20, 2.00, 25.00,
 '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
 '趋势：长窗口+中高阈值过滤噪音。止盈启动/回撤组合满足费用平衡点约束（含0.6%保证金缓冲）。', 1, 101),
(@group_id, 'binance_btc_v3_conservative_volatile', '🛡️ 保守-区间震荡 (V3)',
 'conservative', 'volatile', 240, 90.00, 2, 6.00, 2.60, 1.60, 22.00,
 '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
 '震荡：降杠杆/降仓位/阈值中等偏高，减少磨损。止盈启动更高，降低小利润被费用吞噬概率。', 1, 102),
(@group_id, 'binance_btc_v3_conservative_high_vol', '🛡️ 保守-高波动防守 (V3)',
 'conservative', 'high_vol', 120, 240.00, 2, 4.50, 5.20, 2.60, 30.00,
 '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
 '高波动：更高阈值+更低仓位，减少乱扫与手续费消耗；止盈启动上移提高净收益概率。', 1, 103),
(@group_id, 'binance_btc_v3_conservative_low_vol', '🛡️ 保守-低波动蓄力 (V3)',
 'conservative', 'low_vol', 720, 70.00, 4, 9.00, 2.40, 1.80, 18.00,
 '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
 '低波动：长窗口+较低阈值，适配慢行情。止盈启动适中，确保覆盖费用并保留一定获利空间。', 1, 104);

-- ==================== ⚖️ 平衡型（默认推荐：趋势/震荡/低波动） ====================
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`,
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `auto_start_retreat_percent`, `profit_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES
(@group_id, 'binance_btc_v3_balanced_trend', '⚖️ 平衡-趋势跟踪 ⭐推荐 (V3)',
 'balanced', 'trend', 360, 150.00, 5, 11.00, 4.80, 2.80, 25.00,
 '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
 '推荐映射：trend->balanced。中窗口/中高阈值，提高触发质量；止盈启动更偏“净利润”导向。', 1, 201),
(@group_id, 'binance_btc_v3_balanced_volatile', '⚖️ 平衡-区间套利 ⭐推荐 (V3)',
 'balanced', 'volatile', 240, 105.00, 4, 9.50, 4.20, 2.20, 22.00,
 '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
 '推荐映射：volatile->balanced。阈值更偏保守，降低频率；止盈启动提高，减少“赚了个手续费”。', 1, 202),
(@group_id, 'binance_btc_v3_balanced_high_vol', '⚖️ 平衡-波动捕捉 (V3)',
 'balanced', 'high_vol', 90, 280.00, 6, 7.50, 6.80, 3.40, 28.00,
 '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
 '高波动：在控制仓位前提下参与机会；止盈启动更高，避免高频成本侵蚀。', 1, 203),
(@group_id, 'binance_btc_v3_balanced_low_vol', '⚖️ 平衡-低波动突破 ⭐推荐 (V3)',
 'balanced', 'low_vol', 600, 85.00, 6, 12.50, 3.40, 2.40, 18.00,
 '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
 '推荐映射：low_vol->balanced。低波动更容易“磨”，提高净利润门槛避免费用反噬。', 1, 204);

-- ==================== 🚀 激进型（高杠杆，高费用门槛，更适合小仓位） ====================
INSERT INTO `hg_trading_strategy_template` (
  `group_id`, `strategy_key`, `strategy_name`,
  `risk_preference`, `market_state`,
  `monitor_window`, `volatility_threshold`,
  `leverage`, `margin_percent`,
  `stop_loss_percent`, `auto_start_retreat_percent`, `profit_retreat_percent`,
  `config_json`, `description`, `is_active`, `sort`
) VALUES
(@group_id, 'binance_btc_v3_aggressive_trend', '🚀 激进-趋势冲锋 (V3)',
 'aggressive', 'trend', 240, 170.00, 10, 14.00, 8.50, 4.50, 18.00,
 '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
 '高杠杆趋势追击：止盈启动与回撤组合显著高于费用门槛，避免净利润为负。风险极高。', 1, 301),
(@group_id, 'binance_btc_v3_aggressive_volatile', '🚀 激进-双向博弈 (V3)',
 'aggressive', 'volatile', 180, 130.00, 8, 12.00, 7.50, 3.80, 18.00,
 '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
 '震荡高频容易被费用吃掉，因此抬高净利润门槛；仍偏快进快出。风险很高。', 1, 302),
(@group_id, 'binance_btc_v3_aggressive_high_vol', '🚀 激进-极速博弈 (V3)',
 'aggressive', 'high_vol', 60, 340.00, 12, 9.00, 11.50, 6.00, 22.00,
 '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
 '高波动激进：超短窗口+高阈值减少乱扫；止盈启动更高，尽量覆盖高杠杆费用与滑点。', 1, 303),
(@group_id, 'binance_btc_v3_aggressive_low_vol', '🚀 激进-低波动狙击 (V3)',
 'aggressive', 'low_vol', 420, 95.00, 15, 12.00, 6.50, 4.80, 15.00,
 '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
 '低波动高杠杆最易被费用侵蚀：抬高止盈启动门槛，降低回撤阈值以锁净利润；风险极高。', 1, 304);

-- 验证（可选）
SELECT
  g.id AS group_id, g.group_key, g.group_name, g.exchange, g.symbol, g.is_official, g.is_active,
  t.market_state, t.risk_preference,
  t.monitor_window, t.volatility_threshold,
  t.leverage, t.margin_percent,
  t.stop_loss_percent, t.auto_start_retreat_percent, t.profit_retreat_percent,
  t.strategy_key
FROM hg_trading_strategy_group g
JOIN hg_trading_strategy_template t ON t.group_id = g.id
WHERE g.group_key = 'official_binance_btcusdt_v3'
ORDER BY t.sort;


