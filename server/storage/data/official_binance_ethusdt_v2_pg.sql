-- ============================================================
-- Binance ETH-USDT 官方策略 V2.0 (PostgreSQL版本)
-- 创建时间: 2026-01-01
-- 说明:
--   - 适配“新市场状态算法 + 波动率配置(ToogoVolatilityConfig) + 风险偏好映射(v2)”
--   - 机器人运行时仅使用策略模板硬字段（与BTC一致）
-- 手续费假设：
--   - Market为Taker：开仓0.04% + 平仓0.04% ≈ 0.08%（名义价值口径）
-- 重要声明：
--   - 市场存在不确定性，本文件仅提供参数模板，无法保证实际盈利。
-- ============================================================

DO $$
DECLARE
    v_group_id BIGINT;
BEGIN
    -- 清理旧数据（可重复导入）
    SELECT id INTO v_group_id FROM hg_trading_strategy_group WHERE group_key = 'official_binance_ethusdt_v2';
    IF v_group_id IS NOT NULL THEN
        DELETE FROM hg_trading_strategy_template WHERE group_id = v_group_id;
        DELETE FROM hg_trading_strategy_group WHERE id = v_group_id;
    END IF;

    -- ============================================================
    -- 插入官方策略组 V2.0
    -- ============================================================
    INSERT INTO hg_trading_strategy_group (
      group_name, group_key, exchange, symbol, order_type, margin_mode,
      is_official, user_id, description, is_active, sort, created_at, updated_at
    ) VALUES (
      '🔥 Binance ETH-USDT 官方策略 V2.0（新算法）',
      'official_binance_ethusdt_v2',
      'binance',
      'ETHUSDT',
      'market',
      'isolated',
      1,
      0,
      '币安ETHUSDT官方策略V2.0（适配新市场状态算法/波动率配置/风险偏好映射）。窗口阈值触发→方向预警→自动下单→止损/追踪止盈自动平仓。开仓保证金百分比严格等于平台余额百分比（按AvailableBalance计算）。手续费按Taker双边约0.08%纳入止盈启动阈值设计。',
      1,
      4,
      NOW(),
      NOW()
    ) RETURNING id INTO v_group_id;

    -- ============================================================
    -- 12套策略模板（4种市场状态 × 3种风险偏好）
    -- ETHUSDT 波动阈值(USDT)按常见盘面缩放：约为BTC的 1/12~1/18
    -- ============================================================

    -- ==================== 🛡️ 保守型 ====================

    -- 【1】保守-趋势跟踪
    INSERT INTO hg_trading_strategy_template (
      group_id, strategy_key, strategy_name,
      risk_preference, market_state,
      monitor_window, volatility_threshold,
      leverage, margin_percent,
      stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
      config_json, description, is_active, sort, created_at, updated_at
    ) VALUES (
      v_group_id, 'binance_eth_v2_conservative_trend', '🛡️ 保守-趋势跟踪 (V2)',
      'conservative', 'trend',
      480, 9.00,
      3, 8.00,
      3.00, 2.20, 28.00,
      '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004},"notes":"runtime uses template fields only"}',
      '顺势为主，阈值适中以过滤噪音；止盈启动2.2%覆盖手续费与滑点冗余。',
      1, 101, NOW(), NOW()
    );

    -- 【2】保守-区间震荡
    INSERT INTO hg_trading_strategy_template (
      group_id, strategy_key, strategy_name,
      risk_preference, market_state,
      monitor_window, volatility_threshold,
      leverage, margin_percent,
      stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
      config_json, description, is_active, sort, created_at, updated_at
    ) VALUES (
      v_group_id, 'binance_eth_v2_conservative_volatile', '🛡️ 保守-区间震荡 (V2)',
      'conservative', 'volatile',
      240, 6.00,
      2, 7.00,
      2.80, 1.80, 25.00,
      '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
      '震荡市降低杠杆/仓位，阈值更贴近盘面常见波幅；止损偏紧降低来回磨损。',
      1, 102, NOW(), NOW()
    );

    -- 【3】保守-高波动防守
    INSERT INTO hg_trading_strategy_template (
      group_id, strategy_key, strategy_name,
      risk_preference, market_state,
      monitor_window, volatility_threshold,
      leverage, margin_percent,
      stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
      config_json, description, is_active, sort, created_at, updated_at
    ) VALUES (
      v_group_id, 'binance_eth_v2_conservative_high_vol', '🛡️ 保守-高波动防守 (V2)',
      'conservative', 'high_vol',
      120, 18.00,
      2, 5.00,
      5.50, 3.50, 35.00,
      '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
      '高波动时提高触发阈值减少噪声，止盈启动提高到3.5%避免小利润被费用吞噬。',
      1, 103, NOW(), NOW()
    );

    -- 【4】保守-低波动蓄力
    INSERT INTO hg_trading_strategy_template (
      group_id, strategy_key, strategy_name,
      risk_preference, market_state,
      monitor_window, volatility_threshold,
      leverage, margin_percent,
      stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
      config_json, description, is_active, sort, created_at, updated_at
    ) VALUES (
      v_group_id, 'binance_eth_v2_conservative_low_vol', '🛡️ 保守-低波动蓄力 (V2)',
      'conservative', 'low_vol',
      720, 4.50,
      4, 10.00,
      2.20, 1.20, 20.00,
      '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
      '低波动用更长窗口捕捉慢行情，阈值更小；止盈启动更低以适配小波动收益。',
      1, 104, NOW(), NOW()
    );

    -- ==================== ⚖️ 平衡型 ====================

    -- 【5】平衡-趋势跟踪 ⭐默认映射推荐(trend->balanced)
    INSERT INTO hg_trading_strategy_template (
      group_id, strategy_key, strategy_name,
      risk_preference, market_state,
      monitor_window, volatility_threshold,
      leverage, margin_percent,
      stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
      config_json, description, is_active, sort, created_at, updated_at
    ) VALUES (
      v_group_id, 'binance_eth_v2_balanced_trend', '⚖️ 平衡-趋势跟踪 ⭐推荐 (V2)',
      'balanced', 'trend',
      360, 11.00,
      5, 12.00,
      5.00, 3.00, 25.00,
      '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
      '默认映射推荐：趋势→平衡。中等窗口/阈值，兼顾触发质量与频率。',
      1, 201, NOW(), NOW()
    );

    -- 【6】平衡-区间套利 ⭐默认映射推荐(volatile->balanced)
    INSERT INTO hg_trading_strategy_template (
      group_id, strategy_key, strategy_name,
      risk_preference, market_state,
      monitor_window, volatility_threshold,
      leverage, margin_percent,
      stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
      config_json, description, is_active, sort, created_at, updated_at
    ) VALUES (
      v_group_id, 'binance_eth_v2_balanced_volatile', '⚖️ 平衡-区间套利 ⭐推荐 (V2)',
      'balanced', 'volatile',
      240, 7.00,
      4, 10.00,
      4.50, 2.50, 22.00,
      '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
      '默认映射推荐：震荡→平衡。阈值适中，追踪止盈偏紧以提高净利润留存。',
      1, 202, NOW(), NOW()
    );

    -- 【7】平衡-波动捕捉
    INSERT INTO hg_trading_strategy_template (
      group_id, strategy_key, strategy_name,
      risk_preference, market_state,
      monitor_window, volatility_threshold,
      leverage, margin_percent,
      stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
      config_json, description, is_active, sort, created_at, updated_at
    ) VALUES (
      v_group_id, 'binance_eth_v2_balanced_high_vol', '⚖️ 平衡-波动捕捉 (V2)',
      'balanced', 'high_vol',
      90, 20.00,
      6, 8.00,
      7.00, 4.50, 28.00,
      '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
      '高波动下提高阈值降低误触发，止盈启动上移到4.5%覆盖高杠杆费用与滑点。',
      1, 203, NOW(), NOW()
    );

    -- 【8】平衡-低波动突破
    INSERT INTO hg_trading_strategy_template (
      group_id, strategy_key, strategy_name,
      risk_preference, market_state,
      monitor_window, volatility_threshold,
      leverage, margin_percent,
      stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
      config_json, description, is_active, sort, created_at, updated_at
    ) VALUES (
      v_group_id, 'binance_eth_v2_balanced_low_vol', '⚖️ 平衡-低波动突破 (V2)',
      'balanced', 'low_vol',
      600, 5.00,
      6, 14.00,
      3.20, 1.80, 18.00,
      '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
      '低波动中更强调“高质量触发”，阈值略高于保守；止损更宽避免小回撤打掉。',
      1, 204, NOW(), NOW()
    );

    -- ==================== 🚀 激进型 ====================

    -- 【9】激进-趋势冲锋
    INSERT INTO hg_trading_strategy_template (
      group_id, strategy_key, strategy_name,
      risk_preference, market_state,
      monitor_window, volatility_threshold,
      leverage, margin_percent,
      stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
      config_json, description, is_active, sort, created_at, updated_at
    ) VALUES (
      v_group_id, 'binance_eth_v2_aggressive_trend', '🚀 激进-趋势冲锋 (V2)',
      'aggressive', 'trend',
      240, 13.00,
      10, 18.00,
      8.00, 5.00, 20.00,
      '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
      '高杠杆趋势追击，止盈启动5%（保证金口径）显著高于费用平衡点。',
      1, 301, NOW(), NOW()
    );

    -- 【10】激进-双向博弈
    INSERT INTO hg_trading_strategy_template (
      group_id, strategy_key, strategy_name,
      risk_preference, market_state,
      monitor_window, volatility_threshold,
      leverage, margin_percent,
      stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
      config_json, description, is_active, sort, created_at, updated_at
    ) VALUES (
      v_group_id, 'binance_eth_v2_aggressive_volatile', '🚀 激进-双向博弈 (V2)',
      'aggressive', 'volatile',
      180, 9.00,
      8, 15.00,
      7.00, 4.00, 18.00,
      '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
      '震荡市更快窗口更高阈值，追踪止盈更紧，偏向快进快出以控制净手续费占比。',
      1, 302, NOW(), NOW()
    );

    -- 【11】激进-极速博弈 ⭐默认映射推荐(high_vol->aggressive)
    INSERT INTO hg_trading_strategy_template (
      group_id, strategy_key, strategy_name,
      risk_preference, market_state,
      monitor_window, volatility_threshold,
      leverage, margin_percent,
      stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
      config_json, description, is_active, sort, created_at, updated_at
    ) VALUES (
      v_group_id, 'binance_eth_v2_aggressive_high_vol', '🚀 激进-极速博弈 ⭐推荐 (V2)',
      'aggressive', 'high_vol',
      60, 24.00,
      12, 12.00,
      11.00, 6.50, 22.00,
      '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
      '默认映射推荐：高波动→激进。超短窗口+更高阈值降低噪声，止盈启动6.5%覆盖高杠杆费用与滑点。',
      1, 303, NOW(), NOW()
    );

    -- 【12】激进-低波动狙击
    INSERT INTO hg_trading_strategy_template (
      group_id, strategy_key, strategy_name,
      risk_preference, market_state,
      monitor_window, volatility_threshold,
      leverage, margin_percent,
      stop_loss_percent, auto_start_retreat_percent, profit_retreat_percent,
      config_json, description, is_active, sort, created_at, updated_at
    ) VALUES (
      v_group_id, 'binance_eth_v2_aggressive_low_vol', '🚀 激进-低波动狙击 (V2)',
      'aggressive', 'low_vol',
      420, 6.00,
      15, 22.00,
      5.50, 3.20, 15.00,
      '{"version":"2.0","fee":{"maker":0.0002,"taker":0.0004}}',
      '低波动重仓放大收益，风险极高；追踪止盈更紧以锁定利润，避免横盘磨损。',
      1, 304, NOW(), NOW()
    );

END $$;


