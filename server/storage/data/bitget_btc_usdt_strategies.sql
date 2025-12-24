-- ============================================================
-- Bitget BTC-USDT 官方策略模板（12种）
-- 覆盖：4种市场状态 × 3种风险偏好
-- 市场状态：trend(趋势)、volatile(震荡)、high_vol(高波动)、low_vol(低波动)
-- 风险偏好：conservative(保守)、balanced(平衡)、aggressive(激进)
-- ============================================================

-- 先清理旧的官方策略
DELETE FROM hg_trading_strategy_template WHERE strategy_key LIKE 'bitget_btc_%';

-- ============================================================
-- 一、趋势市场（trend）策略 - 适合单边行情
-- 特点：顺势而为，持仓时间较长，回撤容忍度较高
-- ============================================================

-- 1. 趋势市场 - 保守型
INSERT INTO hg_trading_strategy_template (
    strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort
) VALUES (
    'bitget_btc_trend_conservative',
    'BTC趋势-保守型',
    'conservative',
    'trend',
    300, 10.00,
    3, 5, 5.00, 10.00,
    3.00, 30.00, 5.00,
    '{"exchange":"bitget","symbol":"BTC-USDT","orderType":"market","marginMode":"isolated","reverseEnabled":true,"reverseLossRatio":50,"reverseProfitRatio":100}',
    '适合趋势行情，低杠杆(3-5x)稳健操作。止损3%保护本金，盈利5%启动追踪，回撤30%止盈。允许反向单(亏损50%/盈利100%回撤后)。',
    1, 101
);

-- 2. 趋势市场 - 平衡型
INSERT INTO hg_trading_strategy_template (
    strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort
) VALUES (
    'bitget_btc_trend_balanced',
    'BTC趋势-平衡型',
    'balanced',
    'trend',
    300, 15.00,
    5, 10, 8.00, 15.00,
    5.00, 25.00, 8.00,
    '{"exchange":"bitget","symbol":"BTC-USDT","orderType":"market","marginMode":"isolated","reverseEnabled":true,"reverseLossRatio":50,"reverseProfitRatio":100}',
    '适合趋势行情，中等杠杆(5-10x)平衡收益与风险。止损5%，盈利8%启动追踪，回撤25%止盈。',
    1, 102
);

-- 3. 趋势市场 - 激进型
INSERT INTO hg_trading_strategy_template (
    strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort
) VALUES (
    'bitget_btc_trend_aggressive',
    'BTC趋势-激进型',
    'aggressive',
    'trend',
    300, 20.00,
    10, 20, 10.00, 20.00,
    8.00, 20.00, 10.00,
    '{"exchange":"bitget","symbol":"BTC-USDT","orderType":"market","marginMode":"isolated","reverseEnabled":true,"reverseLossRatio":50,"reverseProfitRatio":100}',
    '适合强趋势行情，高杠杆(10-20x)追求高收益。止损8%，盈利10%启动追踪，回撤20%止盈。',
    1, 103
);

-- ============================================================
-- 二、震荡市场（volatile）策略 - 适合区间波动
-- 特点：高抛低吸，持仓时间短，不开反向单
-- ============================================================

-- 4. 震荡市场 - 保守型
INSERT INTO hg_trading_strategy_template (
    strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort
) VALUES (
    'bitget_btc_volatile_conservative',
    'BTC震荡-保守型',
    'conservative',
    'volatile',
    120, 5.00,
    2, 4, 3.00, 8.00,
    2.00, 40.00, 3.00,
    '{"exchange":"bitget","symbol":"BTC-USDT","orderType":"market","marginMode":"isolated","reverseEnabled":false,"reverseLossRatio":0,"reverseProfitRatio":0}',
    '适合震荡行情，超低杠杆(2-4x)安全操作。止损2%极度保守，盈利3%启动追踪，不开反向单。',
    1, 104
);

-- 5. 震荡市场 - 平衡型
INSERT INTO hg_trading_strategy_template (
    strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort
) VALUES (
    'bitget_btc_volatile_balanced',
    'BTC震荡-平衡型',
    'balanced',
    'volatile',
    120, 8.00,
    4, 8, 5.00, 12.00,
    4.00, 35.00, 5.00,
    '{"exchange":"bitget","symbol":"BTC-USDT","orderType":"market","marginMode":"isolated","reverseEnabled":false,"reverseLossRatio":0,"reverseProfitRatio":0}',
    '适合震荡行情，中低杠杆(4-8x)稳健获利。止损4%，盈利5%启动追踪，不开反向单。',
    1, 105
);

-- 6. 震荡市场 - 激进型
INSERT INTO hg_trading_strategy_template (
    strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort
) VALUES (
    'bitget_btc_volatile_aggressive',
    'BTC震荡-激进型',
    'aggressive',
    'volatile',
    120, 12.00,
    8, 15, 8.00, 18.00,
    6.00, 30.00, 8.00,
    '{"exchange":"bitget","symbol":"BTC-USDT","orderType":"market","marginMode":"isolated","reverseEnabled":false,"reverseLossRatio":0,"reverseProfitRatio":0}',
    '适合震荡行情，中等杠杆(8-15x)快进快出。止损6%，盈利8%启动追踪，不开反向单。',
    1, 106
);

-- ============================================================
-- 三、高波动市场（high_vol）策略 - 适合剧烈波动
-- 特点：快速止损止盈，允许双向持仓，风险控制严格
-- ============================================================

-- 7. 高波动 - 保守型
INSERT INTO hg_trading_strategy_template (
    strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort
) VALUES (
    'bitget_btc_high_vol_conservative',
    'BTC高波动-保守型',
    'conservative',
    'high_vol',
    60, 20.00,
    2, 3, 2.00, 5.00,
    2.00, 50.00, 2.00,
    '{"exchange":"bitget","symbol":"BTC-USDT","orderType":"market","marginMode":"isolated","reverseEnabled":true,"reverseLossRatio":100,"reverseProfitRatio":100}',
    '适合高波动行情，超低仓位(2-5%)控制风险。止损2%快速离场，允许反向单对冲(100%回撤后)。',
    1, 107
);

-- 8. 高波动 - 平衡型
INSERT INTO hg_trading_strategy_template (
    strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort
) VALUES (
    'bitget_btc_high_vol_balanced',
    'BTC高波动-平衡型',
    'balanced',
    'high_vol',
    60, 30.00,
    3, 6, 4.00, 8.00,
    4.00, 45.00, 4.00,
    '{"exchange":"bitget","symbol":"BTC-USDT","orderType":"market","marginMode":"isolated","reverseEnabled":true,"reverseLossRatio":100,"reverseProfitRatio":100}',
    '适合高波动行情，小仓位(4-8%)捕捉机会。止损4%控制风险，允许反向单灵活操作。',
    1, 108
);

-- 9. 高波动 - 激进型
INSERT INTO hg_trading_strategy_template (
    strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort
) VALUES (
    'bitget_btc_high_vol_aggressive',
    'BTC高波动-激进型',
    'aggressive',
    'high_vol',
    60, 40.00,
    5, 10, 6.00, 12.00,
    6.00, 40.00, 6.00,
    '{"exchange":"bitget","symbol":"BTC-USDT","orderType":"market","marginMode":"isolated","reverseEnabled":true,"reverseLossRatio":100,"reverseProfitRatio":100}',
    '适合高波动行情，中等仓位(6-12%)博取收益。止损6%，允许双向持仓最大化收益。',
    1, 109
);

-- ============================================================
-- 四、低波动市场（low_vol）策略 - 适合盘整行情
-- 特点：等待突破，耐心持仓，不开反向单
-- ============================================================

-- 10. 低波动 - 保守型
INSERT INTO hg_trading_strategy_template (
    strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort
) VALUES (
    'bitget_btc_low_vol_conservative',
    'BTC低波动-保守型',
    'conservative',
    'low_vol',
    300, 2.00,
    5, 8, 10.00, 20.00,
    2.00, 20.00, 2.00,
    '{"exchange":"bitget","symbol":"BTC-USDT","orderType":"market","marginMode":"isolated","reverseEnabled":false,"reverseLossRatio":0,"reverseProfitRatio":0}',
    '适合低波动行情，耐心等待突破。低止损2%防假突破，高仓位(10-20%)捕捉真突破，不开反向单。',
    1, 110
);

-- 11. 低波动 - 平衡型
INSERT INTO hg_trading_strategy_template (
    strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort
) VALUES (
    'bitget_btc_low_vol_balanced',
    'BTC低波动-平衡型',
    'balanced',
    'low_vol',
    300, 3.00,
    8, 12, 15.00, 25.00,
    3.00, 15.00, 3.00,
    '{"exchange":"bitget","symbol":"BTC-USDT","orderType":"market","marginMode":"isolated","reverseEnabled":false,"reverseLossRatio":0,"reverseProfitRatio":0}',
    '适合低波动行情，中等杠杆(8-12x)等待机会。止损3%，盈利3%启动追踪，回撤15%止盈。',
    1, 111
);

-- 12. 低波动 - 激进型
INSERT INTO hg_trading_strategy_template (
    strategy_key, strategy_name, risk_preference, market_state,
    monitor_window, volatility_threshold,
    leverage_min, leverage_max, margin_percent_min, margin_percent_max,
    stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent,
    config_json, description, is_active, sort
) VALUES (
    'bitget_btc_low_vol_aggressive',
    'BTC低波动-激进型',
    'aggressive',
    'low_vol',
    300, 5.00,
    12, 20, 20.00, 30.00,
    5.00, 10.00, 5.00,
    '{"exchange":"bitget","symbol":"BTC-USDT","orderType":"market","marginMode":"isolated","reverseEnabled":false,"reverseLossRatio":0,"reverseProfitRatio":0}',
    '适合低波动突破行情，高杠杆(12-20x)高仓位(20-30%)。止损5%防止假突破，回撤10%快速止盈。',
    1, 112
);
