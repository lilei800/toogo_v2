-- 检查策略组49的策略模板数据
SELECT 
    id,
    group_id,
    strategy_key,
    strategy_name,
    market_state,
    risk_preference,
    leverage,
    margin_percent,
    stop_loss_percent,
    profit_retreat_percent,
    auto_start_retreat_percent,
    is_active
FROM hg_trading_strategy_template
WHERE group_id = 49
ORDER BY market_state, risk_preference;

-- 检查策略组49的信息
SELECT 
    id,
    group_name,
    group_key,
    symbol,
    description
FROM hg_trading_strategy_group
WHERE id = 49;

-- 检查所有策略组中的market_state值格式
SELECT DISTINCT market_state
FROM hg_trading_strategy_template
WHERE group_id = 49;

