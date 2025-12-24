-- 检查所有策略组和策略模板的market_state格式
SELECT 
    g.id AS group_id,
    g.group_name,
    g.group_key,
    t.id AS template_id,
    t.strategy_key,
    t.strategy_name,
    t.market_state,
    t.risk_preference,
    t.is_active
FROM hg_trading_strategy_group g
LEFT JOIN hg_trading_strategy_template t ON t.group_id = g.id
WHERE g.is_official = 1
ORDER BY g.id, t.market_state, t.risk_preference;

-- 检查market_state的所有可能值
SELECT DISTINCT market_state
FROM hg_trading_strategy_template
ORDER BY market_state;

