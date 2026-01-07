-- ============================================================
-- Binance BTC-USDT 官方策略 V3.0（Fee-aware，PostgreSQL版本）
-- 创建时间: 2026-01-01
--
-- 与 `official_binance_btcusdt_v3.sql` 参数一致，仅SQL方言不同。
-- ============================================================

DO $$
DECLARE
    v_group_id BIGINT;
BEGIN
    -- 清理旧数据（可重复导入）
    SELECT id INTO v_group_id FROM hg_trading_strategy_group WHERE group_key = 'official_binance_btcusdt_v3';
    IF v_group_id IS NOT NULL THEN
        DELETE FROM hg_trading_strategy_template WHERE group_id = v_group_id;
        DELETE FROM hg_trading_strategy_group WHERE id = v_group_id;
    END IF;

    -- 插入官方策略组 V3.0
    INSERT INTO hg_trading_strategy_group (
      group_name, group_key, exchange, symbol, order_type, margin_mode,
      is_official, user_id, description, is_active, sort, created_at, updated_at
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
      5,
      NOW(),
      NOW()
    ) RETURNING id INTO v_group_id;

    -- 统一费用模型（写入 config_json，仅用于解释；运行时主要使用硬字段）
    -- - taker: 0.0004/side, slippage: 0.0002/side, roundTripNotionalCost: 0.0012
    -- - bufferMarginPct: 0.6

    -- ==================== 🛡️ 保守型 ====================
    INSERT INTO hg_trading_strategy_template (
      group_id, strategy_key, strategy_name,
      risk_preference, market_state,
      monitor_window, volatility_threshold,
      leverage, margin_percent,
      stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
      config_json, description, is_active, sort, created_at, updated_at
    ) VALUES
    (v_group_id, 'binance_btc_v3_conservative_trend', '🛡️ 保守-趋势跟踪 (V3)',
     'conservative', 'trend', 480, 130.00, 3, 7.00, 3.20, 2.00, 25.00,
     '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
     '趋势：长窗口+中高阈值过滤噪音。止盈启动/回撤组合满足费用平衡点约束（含0.6%保证金缓冲）。', 1, 101, NOW(), NOW()),
    (v_group_id, 'binance_btc_v3_conservative_volatile', '🛡️ 保守-区间震荡 (V3)',
     'conservative', 'volatile', 240, 90.00, 2, 6.00, 2.60, 1.60, 22.00,
     '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
     '震荡：降杠杆/降仓位/阈值中等偏高，减少磨损。止盈启动更高，降低小利润被费用吞噬概率。', 1, 102, NOW(), NOW()),
    (v_group_id, 'binance_btc_v3_conservative_high_vol', '🛡️ 保守-高波动防守 (V3)',
     'conservative', 'high_vol', 120, 240.00, 2, 4.50, 5.20, 2.60, 30.00,
     '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
     '高波动：更高阈值+更低仓位，减少乱扫与手续费消耗；止盈启动上移提高净收益概率。', 1, 103, NOW(), NOW()),
    (v_group_id, 'binance_btc_v3_conservative_low_vol', '🛡️ 保守-低波动蓄力 (V3)',
     'conservative', 'low_vol', 720, 70.00, 4, 9.00, 2.40, 1.80, 18.00,
     '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
     '低波动：长窗口+较低阈值，适配慢行情。止盈启动适中，确保覆盖费用并保留一定获利空间。', 1, 104, NOW(), NOW());

    -- ==================== ⚖️ 平衡型 ====================
    INSERT INTO hg_trading_strategy_template (
      group_id, strategy_key, strategy_name,
      risk_preference, market_state,
      monitor_window, volatility_threshold,
      leverage, margin_percent,
      stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
      config_json, description, is_active, sort, created_at, updated_at
    ) VALUES
    (v_group_id, 'binance_btc_v3_balanced_trend', '⚖️ 平衡-趋势跟踪 ⭐推荐 (V3)',
     'balanced', 'trend', 360, 150.00, 5, 11.00, 4.80, 2.80, 25.00,
     '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
     '推荐映射：trend->balanced。中窗口/中高阈值，提高触发质量；止盈启动更偏“净利润”导向。', 1, 201, NOW(), NOW()),
    (v_group_id, 'binance_btc_v3_balanced_volatile', '⚖️ 平衡-区间套利 ⭐推荐 (V3)',
     'balanced', 'volatile', 240, 105.00, 4, 9.50, 4.20, 2.20, 22.00,
     '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
     '推荐映射：volatile->balanced。阈值更偏保守，降低频率；止盈启动提高，减少“赚了个手续费”。', 1, 202, NOW(), NOW()),
    (v_group_id, 'binance_btc_v3_balanced_high_vol', '⚖️ 平衡-波动捕捉 (V3)',
     'balanced', 'high_vol', 90, 280.00, 6, 7.50, 6.80, 3.40, 28.00,
     '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
     '高波动：在控制仓位前提下参与机会；止盈启动更高，避免高频成本侵蚀。', 1, 203, NOW(), NOW()),
    (v_group_id, 'binance_btc_v3_balanced_low_vol', '⚖️ 平衡-低波动突破 ⭐推荐 (V3)',
     'balanced', 'low_vol', 600, 85.00, 6, 12.50, 3.40, 2.40, 18.00,
     '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
     '推荐映射：low_vol->balanced。低波动更容易“磨”，提高净利润门槛避免费用反噬。', 1, 204, NOW(), NOW());

    -- ==================== 🚀 激进型 ====================
    INSERT INTO hg_trading_strategy_template (
      group_id, strategy_key, strategy_name,
      risk_preference, market_state,
      monitor_window, volatility_threshold,
      leverage, margin_percent,
      stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
      config_json, description, is_active, sort, created_at, updated_at
    ) VALUES
    (v_group_id, 'binance_btc_v3_aggressive_trend', '🚀 激进-趋势冲锋 (V3)',
     'aggressive', 'trend', 240, 170.00, 10, 14.00, 8.50, 4.50, 18.00,
     '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
     '高杠杆趋势追击：止盈启动与回撤组合显著高于费用门槛，避免净利润为负。风险极高。', 1, 301, NOW(), NOW()),
    (v_group_id, 'binance_btc_v3_aggressive_volatile', '🚀 激进-双向博弈 (V3)',
     'aggressive', 'volatile', 180, 130.00, 8, 12.00, 7.50, 3.80, 18.00,
     '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
     '震荡高频容易被费用吃掉，因此抬高净利润门槛；仍偏快进快出。风险很高。', 1, 302, NOW(), NOW()),
    (v_group_id, 'binance_btc_v3_aggressive_high_vol', '🚀 激进-极速博弈 (V3)',
     'aggressive', 'high_vol', 60, 340.00, 12, 9.00, 11.50, 6.00, 22.00,
     '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
     '高波动激进：超短窗口+高阈值减少乱扫；止盈启动更高，尽量覆盖高杠杆费用与滑点。', 1, 303, NOW(), NOW()),
    (v_group_id, 'binance_btc_v3_aggressive_low_vol', '🚀 激进-低波动狙击 (V3)',
     'aggressive', 'low_vol', 420, 95.00, 15, 12.00, 6.50, 4.80, 15.00,
     '{"version":"3.0","fee":{"maker":0.0002,"taker":0.0004},"slippage":0.0002,"roundTripNotionalCost":0.0012,"bufferMarginPct":0.6}',
     '低波动高杠杆最易被费用侵蚀：抬高止盈启动门槛，降低回撤阈值以锁净利润；风险极高。', 1, 304, NOW(), NOW());

END $$;

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


