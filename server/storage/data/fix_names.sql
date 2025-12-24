-- 修复策略名称
UPDATE hg_trading_strategy_template SET strategy_name = 'BTC趋势-保守型' WHERE strategy_key = 'conservative_trend';
UPDATE hg_trading_strategy_template SET strategy_name = 'BTC趋势-平衡型' WHERE strategy_key = 'balanced_trend';
UPDATE hg_trading_strategy_template SET strategy_name = 'BTC趋势-激进型' WHERE strategy_key = 'aggressive_trend';

UPDATE hg_trading_strategy_template SET strategy_name = 'BTC震荡-保守型' WHERE strategy_key = 'conservative_volatile';
UPDATE hg_trading_strategy_template SET strategy_name = 'BTC震荡-平衡型' WHERE strategy_key = 'balanced_volatile';
UPDATE hg_trading_strategy_template SET strategy_name = 'BTC震荡-激进型' WHERE strategy_key = 'aggressive_volatile';

UPDATE hg_trading_strategy_template SET strategy_name = 'BTC高波动-保守型' WHERE strategy_key = 'conservative_high_vol';
UPDATE hg_trading_strategy_template SET strategy_name = 'BTC高波动-平衡型' WHERE strategy_key = 'balanced_high_vol';
UPDATE hg_trading_strategy_template SET strategy_name = 'BTC高波动-激进型' WHERE strategy_key = 'aggressive_high_vol';

UPDATE hg_trading_strategy_template SET strategy_name = 'BTC低波动-保守型' WHERE strategy_key = 'conservative_low_vol';
UPDATE hg_trading_strategy_template SET strategy_name = 'BTC低波动-平衡型' WHERE strategy_key = 'balanced_low_vol';
UPDATE hg_trading_strategy_template SET strategy_name = 'BTC低波动-激进型' WHERE strategy_key = 'aggressive_low_vol';

