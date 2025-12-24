SET NAMES utf8mb4;

SELECT group_name FROM hg_trading_strategy_group WHERE group_key = 'official_btc_usdt_v14_enhanced';

SELECT strategy_name, risk_preference, market_state 
FROM hg_trading_strategy_template 
WHERE strategy_key = 'v14_conservative_trend';

SELECT COUNT(*) as total_strategies 
FROM hg_trading_strategy_template 
WHERE group_id = (SELECT id FROM hg_trading_strategy_group WHERE group_key = 'official_btc_usdt_v14_enhanced');

