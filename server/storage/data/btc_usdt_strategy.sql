-- BTC-USDT 官方推荐策略模板
-- 覆盖12种组合：4种市场状态 × 3种风险偏好

-- 清空原有数据
TRUNCATE TABLE hg_trading_strategy_template;

-- ========== 趋势市场 (trend) ==========
-- 趋势市场特点：价格沿一个方向持续运动，适合顺势交易

-- 保守型-趋势市场
INSERT INTO hg_trading_strategy_template 
(strategy_key, strategy_name, risk_preference, market_state, monitor_window, volatility_threshold, leverage_min, leverage_max, margin_percent_min, margin_percent_max, stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent, description, is_active, sort) VALUES
('conservative_trend', '保守型-趋势市场', 'conservative', 'trend', 300, 50.00, 3, 5, 5.00, 10.00, 3.00, 30.00, 2.00, 'BTC-USDT趋势市场保守策略：低杠杆顺势而为，严格止损，稳健获利', 1, 1);

-- 平衡型-趋势市场
INSERT INTO hg_trading_strategy_template 
(strategy_key, strategy_name, risk_preference, market_state, monitor_window, volatility_threshold, leverage_min, leverage_max, margin_percent_min, margin_percent_max, stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent, description, is_active, sort) VALUES
('balanced_trend', '平衡型-趋势市场', 'balanced', 'trend', 300, 80.00, 5, 10, 8.00, 15.00, 5.00, 35.00, 3.00, 'BTC-USDT趋势市场平衡策略：中等杠杆，追踪趋势，平衡风险收益', 1, 2);

-- 激进型-趋势市场
INSERT INTO hg_trading_strategy_template 
(strategy_key, strategy_name, risk_preference, market_state, monitor_window, volatility_threshold, leverage_min, leverage_max, margin_percent_min, margin_percent_max, stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent, description, is_active, sort) VALUES
('aggressive_trend', '激进型-趋势市场', 'aggressive', 'trend', 300, 120.00, 10, 20, 10.00, 20.00, 8.00, 40.00, 5.00, 'BTC-USDT趋势市场激进策略：高杠杆捕捉大趋势，追求高收益', 1, 3);

-- ========== 震荡市场 (volatile) ==========
-- 震荡市场特点：价格在区间内波动，适合高抛低吸

-- 保守型-震荡市场
INSERT INTO hg_trading_strategy_template 
(strategy_key, strategy_name, risk_preference, market_state, monitor_window, volatility_threshold, leverage_min, leverage_max, margin_percent_min, margin_percent_max, stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent, description, is_active, sort) VALUES
('conservative_volatile', '保守型-震荡市场', 'conservative', 'volatile', 180, 30.00, 2, 4, 5.00, 8.00, 2.00, 25.00, 1.50, 'BTC-USDT震荡市场保守策略：小仓位区间操作，快进快出', 1, 4);

-- 平衡型-震荡市场
INSERT INTO hg_trading_strategy_template 
(strategy_key, strategy_name, risk_preference, market_state, monitor_window, volatility_threshold, leverage_min, leverage_max, margin_percent_min, margin_percent_max, stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent, description, is_active, sort) VALUES
('balanced_volatile', '平衡型-震荡市场', 'balanced', 'volatile', 180, 50.00, 4, 8, 8.00, 12.00, 4.00, 30.00, 2.00, 'BTC-USDT震荡市场平衡策略：区间交易，及时止盈止损', 1, 5);

-- 激进型-震荡市场
INSERT INTO hg_trading_strategy_template 
(strategy_key, strategy_name, risk_preference, market_state, monitor_window, volatility_threshold, leverage_min, leverage_max, margin_percent_min, margin_percent_max, stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent, description, is_active, sort) VALUES
('aggressive_volatile', '激进型-震荡市场', 'aggressive', 'volatile', 180, 80.00, 8, 15, 10.00, 18.00, 6.00, 35.00, 3.00, 'BTC-USDT震荡市场激进策略：高频区间交易，快速获利', 1, 6);

-- ========== 高波动市场 (high-volatility) ==========
-- 高波动市场特点：价格剧烈波动，风险和机会并存

-- 保守型-高波动市场
INSERT INTO hg_trading_strategy_template 
(strategy_key, strategy_name, risk_preference, market_state, monitor_window, volatility_threshold, leverage_min, leverage_max, margin_percent_min, margin_percent_max, stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent, description, is_active, sort) VALUES
('conservative_high_vol', '保守型-高波动', 'conservative', 'high-volatility', 120, 100.00, 2, 3, 3.00, 6.00, 2.00, 20.00, 1.00, 'BTC-USDT高波动保守策略：极低杠杆，严格风控，等待确定性机会', 1, 7);

-- 平衡型-高波动市场
INSERT INTO hg_trading_strategy_template 
(strategy_key, strategy_name, risk_preference, market_state, monitor_window, volatility_threshold, leverage_min, leverage_max, margin_percent_min, margin_percent_max, stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent, description, is_active, sort) VALUES
('balanced_high_vol', '平衡型-高波动', 'balanced', 'high-volatility', 120, 150.00, 3, 6, 5.00, 10.00, 4.00, 25.00, 2.00, 'BTC-USDT高波动平衡策略：适度仓位，捕捉大波动', 1, 8);

-- 激进型-高波动市场
INSERT INTO hg_trading_strategy_template 
(strategy_key, strategy_name, risk_preference, market_state, monitor_window, volatility_threshold, leverage_min, leverage_max, margin_percent_min, margin_percent_max, stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent, description, is_active, sort) VALUES
('aggressive_high_vol', '激进型-高波动', 'aggressive', 'high-volatility', 120, 200.00, 5, 10, 8.00, 15.00, 6.00, 30.00, 3.00, 'BTC-USDT高波动激进策略：高风险高收益，适合经验丰富者', 1, 9);

-- ========== 低波动市场 (low-volatility) ==========
-- 低波动市场特点：价格波动小，需要更高杠杆放大收益

-- 保守型-低波动市场
INSERT INTO hg_trading_strategy_template 
(strategy_key, strategy_name, risk_preference, market_state, monitor_window, volatility_threshold, leverage_min, leverage_max, margin_percent_min, margin_percent_max, stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent, description, is_active, sort) VALUES
('conservative_low_vol', '保守型-低波动', 'conservative', 'low-volatility', 600, 20.00, 5, 8, 8.00, 12.00, 2.00, 20.00, 1.50, 'BTC-USDT低波动保守策略：适度杠杆，耐心等待，小幅获利', 1, 10);

-- 平衡型-低波动市场
INSERT INTO hg_trading_strategy_template 
(strategy_key, strategy_name, risk_preference, market_state, monitor_window, volatility_threshold, leverage_min, leverage_max, margin_percent_min, margin_percent_max, stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent, description, is_active, sort) VALUES
('balanced_low_vol', '平衡型-低波动', 'balanced', 'low-volatility', 600, 30.00, 8, 12, 10.00, 15.00, 3.00, 25.00, 2.00, 'BTC-USDT低波动平衡策略：中高杠杆，放大小波动收益', 1, 11);

-- 激进型-低波动市场
INSERT INTO hg_trading_strategy_template 
(strategy_key, strategy_name, risk_preference, market_state, monitor_window, volatility_threshold, leverage_min, leverage_max, margin_percent_min, margin_percent_max, stop_loss_percent, profit_retreat_percent, auto_start_retreat_percent, description, is_active, sort) VALUES
('aggressive_low_vol', '激进型-低波动', 'aggressive', 'low-volatility', 600, 50.00, 12, 20, 12.00, 20.00, 5.00, 30.00, 3.00, 'BTC-USDT低波动激进策略：高杠杆放大收益，需紧密监控', 1, 12);

