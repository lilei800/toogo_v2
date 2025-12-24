-- 修复策略组49的策略模板数据
-- 确保market_state使用正确的格式（low_volatility, high_volatility等）

-- 1. 查看策略组49的信息
SELECT 
    id,
    group_name,
    group_key,
    symbol,
    is_official,
    is_active
FROM hg_trading_strategy_group
WHERE id = 49;

-- 2. 查看策略组49的所有策略模板
SELECT 
    id,
    group_id,
    strategy_key,
    strategy_name,
    market_state,
    risk_preference,
    is_active
FROM hg_trading_strategy_template
WHERE group_id = 49
ORDER BY market_state, risk_preference;

-- 3. 如果策略组49存在但没有策略模板，检查是否需要从V18策略组复制
-- 或者直接为策略组49创建策略模板

-- 4. 检查market_state格式，统一为下划线格式
-- 如果策略组49使用的是旧格式（low_vol），需要更新为low_volatility
UPDATE hg_trading_strategy_template
SET market_state = 'low_volatility'
WHERE group_id = 49 
  AND market_state IN ('low_vol', 'low-volatility');

UPDATE hg_trading_strategy_template
SET market_state = 'high_volatility'
WHERE group_id = 49 
  AND market_state IN ('high_vol', 'high-volatility');

UPDATE hg_trading_strategy_template
SET market_state = 'volatile'
WHERE group_id = 49 
  AND market_state = 'range';

-- 5. 验证更新后的数据
SELECT 
    id,
    group_id,
    strategy_key,
    strategy_name,
    market_state,
    risk_preference,
    is_active
FROM hg_trading_strategy_template
WHERE group_id = 49
ORDER BY market_state, risk_preference;

