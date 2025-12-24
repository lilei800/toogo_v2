-- 查看所有策略模板的实际数据
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
    is_active,
    created_at
FROM hg_trading_strategy_template
ORDER BY group_id, market_state, risk_preference
LIMIT 50;

-- 查看策略组49的所有策略模板
SELECT 
    id,
    group_id,
    strategy_key,
    strategy_name,
    market_state,
    risk_preference,
    leverage,
    margin_percent,
    is_active
FROM hg_trading_strategy_template
WHERE group_id = 49
ORDER BY market_state, risk_preference;

-- 查看所有策略组
SELECT 
    id,
    group_name,
    group_key,
    symbol,
    is_official,
    is_active
FROM hg_trading_strategy_group
ORDER BY id DESC
LIMIT 10;

-- 查看V18策略组的数据
SELECT 
    g.id AS group_id,
    g.group_name,
    COUNT(t.id) AS template_count
FROM hg_trading_strategy_group g
LEFT JOIN hg_trading_strategy_template t ON t.group_id = g.id
WHERE g.group_key = 'btc_usdt_v18_official'
GROUP BY g.id;

-- 查看V18策略组的所有策略模板
SELECT 
    t.id,
    t.group_id,
    t.strategy_key,
    t.strategy_name,
    t.market_state,
    t.risk_preference,
    t.leverage,
    t.margin_percent
FROM hg_trading_strategy_template t
WHERE t.group_id = (SELECT id FROM hg_trading_strategy_group WHERE group_key = 'btc_usdt_v18_official' LIMIT 1)
ORDER BY t.market_state, t.risk_preference;

