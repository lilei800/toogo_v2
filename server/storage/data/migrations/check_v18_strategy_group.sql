-- 检查V18策略组的ID和模板数据
SELECT 
    g.id,
    g.group_name,
    g.group_key,
    COUNT(t.id) AS template_count
FROM hg_trading_strategy_group g
LEFT JOIN hg_trading_strategy_template t ON t.group_id = g.id
WHERE g.group_key = 'btc_usdt_v18_official'
GROUP BY g.id;

-- 检查V18策略组的所有策略模板
SELECT 
    id,
    group_id,
    strategy_key,
    strategy_name,
    market_state,
    risk_preference
FROM hg_trading_strategy_template
WHERE group_id = (SELECT id FROM hg_trading_strategy_group WHERE group_key = 'btc_usdt_v18_official' LIMIT 1)
ORDER BY market_state, risk_preference;

