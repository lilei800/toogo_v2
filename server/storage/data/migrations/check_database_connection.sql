-- 检查数据库连接和表结构
SELECT COUNT(*) AS total_templates FROM hg_trading_strategy_template;
SELECT COUNT(*) AS total_groups FROM hg_trading_strategy_group;

-- 查看最新的策略组（按ID倒序）
SELECT 
    id,
    group_name,
    group_key,
    symbol,
    is_official,
    is_active,
    created_at
FROM hg_trading_strategy_group
ORDER BY id DESC
LIMIT 5;

-- 查看最新的策略模板（按ID倒序）
SELECT 
    id,
    group_id,
    strategy_key,
    strategy_name,
    market_state,
    risk_preference
FROM hg_trading_strategy_template
ORDER BY id DESC
LIMIT 20;

-- 检查V18策略组是否存在
SELECT 
    id,
    group_name,
    group_key,
    symbol,
    is_official,
    is_active
FROM hg_trading_strategy_group
WHERE group_key LIKE '%v18%' OR group_name LIKE '%V18%'
ORDER BY id DESC;

