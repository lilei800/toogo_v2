-- 更新所有策略的config_json和名称

-- 趋势市场策略
UPDATE hg_trading_strategy_template 
SET config_json = '{"exchange":"bitget","symbol":"BTC-USDT","orderType":"market","marginMode":"isolated","reverseEnabled":true,"reverseLossRatio":50,"reverseProfitRatio":100}',
    strategy_name = CONCAT('BTC趋势-', CASE risk_preference WHEN 'conservative' THEN '保守型' WHEN 'balanced' THEN '平衡型' ELSE '激进型' END)
WHERE market_state = 'trend';

-- 震荡市场策略 (不开反向单)
UPDATE hg_trading_strategy_template 
SET config_json = '{"exchange":"bitget","symbol":"BTC-USDT","orderType":"market","marginMode":"isolated","reverseEnabled":false,"reverseLossRatio":0,"reverseProfitRatio":0}',
    strategy_name = CONCAT('BTC震荡-', CASE risk_preference WHEN 'conservative' THEN '保守型' WHEN 'balanced' THEN '平衡型' ELSE '激进型' END)
WHERE market_state = 'volatile';

-- 高波动市场策略 (允许反向单，100%回撤)
UPDATE hg_trading_strategy_template 
SET config_json = '{"exchange":"bitget","symbol":"BTC-USDT","orderType":"market","marginMode":"isolated","reverseEnabled":true,"reverseLossRatio":100,"reverseProfitRatio":100}',
    strategy_name = CONCAT('BTC高波动-', CASE risk_preference WHEN 'conservative' THEN '保守型' WHEN 'balanced' THEN '平衡型' ELSE '激进型' END),
    market_state = 'high_vol'
WHERE market_state = 'high-volatility';

-- 低波动市场策略 (不开反向单)
UPDATE hg_trading_strategy_template 
SET config_json = '{"exchange":"bitget","symbol":"BTC-USDT","orderType":"market","marginMode":"isolated","reverseEnabled":false,"reverseLossRatio":0,"reverseProfitRatio":0}',
    strategy_name = CONCAT('BTC低波动-', CASE risk_preference WHEN 'conservative' THEN '保守型' WHEN 'balanced' THEN '平衡型' ELSE '激进型' END),
    market_state = 'low_vol'
WHERE market_state = 'low-volatility';

-- 更新描述
UPDATE hg_trading_strategy_template SET description = '适合趋势行情，低杠杆(3-5x)稳健操作。止损3%保护本金，盈利5%启动追踪，回撤30%止盈。允许反向单(亏损50%/盈利100%回撤后)。' WHERE strategy_key = 'conservative_trend';
UPDATE hg_trading_strategy_template SET description = '适合趋势行情，中等杠杆(5-10x)平衡收益与风险。止损5%，盈利8%启动追踪，回撤25%止盈。允许反向单。' WHERE strategy_key = 'balanced_trend';
UPDATE hg_trading_strategy_template SET description = '适合强趋势行情，高杠杆(10-20x)追求高收益。止损8%，盈利10%启动追踪，回撤20%止盈。允许反向单。' WHERE strategy_key = 'aggressive_trend';

UPDATE hg_trading_strategy_template SET description = '适合震荡行情，超低杠杆(2-4x)安全操作。止损2%极度保守，盈利3%启动追踪，不开反向单。' WHERE strategy_key = 'conservative_volatile';
UPDATE hg_trading_strategy_template SET description = '适合震荡行情，中低杠杆(4-8x)稳健获利。止损4%，盈利5%启动追踪，不开反向单。' WHERE strategy_key = 'balanced_volatile';
UPDATE hg_trading_strategy_template SET description = '适合震荡行情，中等杠杆(8-15x)快进快出。止损6%，盈利8%启动追踪，不开反向单。' WHERE strategy_key = 'aggressive_volatile';

UPDATE hg_trading_strategy_template SET description = '适合高波动行情，超低仓位(2-5%)控制风险。止损2%快速离场，允许反向单对冲(100%回撤后)。' WHERE strategy_key = 'conservative_high_vol';
UPDATE hg_trading_strategy_template SET description = '适合高波动行情，小仓位(4-8%)捕捉机会。止损4%控制风险，允许反向单灵活操作。' WHERE strategy_key = 'balanced_high_vol';
UPDATE hg_trading_strategy_template SET description = '适合高波动行情，中等仓位(6-12%)博取收益。止损6%，允许双向持仓最大化收益。' WHERE strategy_key = 'aggressive_high_vol';

UPDATE hg_trading_strategy_template SET description = '适合低波动行情，耐心等待突破。低止损2%防假突破，高仓位(10-20%)捕捉真突破，不开反向单。' WHERE strategy_key = 'conservative_low_vol';
UPDATE hg_trading_strategy_template SET description = '适合低波动行情，中等杠杆(8-12x)等待机会。止损3%，盈利3%启动追踪，回撤15%止盈。' WHERE strategy_key = 'balanced_low_vol';
UPDATE hg_trading_strategy_template SET description = '适合低波动突破行情，高杠杆(12-20x)高仓位(20-30%)。止损5%防止假突破，回撤10%快速止盈。' WHERE strategy_key = 'aggressive_low_vol';

-- 验证结果
SELECT strategy_key, strategy_name, market_state, risk_preference, config_json IS NOT NULL as has_config FROM hg_trading_strategy_template ORDER BY sort;

