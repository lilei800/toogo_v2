-- ============================================================
-- 检查策略模板是否存在
-- 用于排查机器人找不到策略模板的问题
-- ============================================================

-- 1. 检查策略组18的所有策略模板
SELECT 
    id,
    group_id,
    strategy_name,
    market_state,
    risk_preference,
    is_active,
    monitor_window,
    volatility_threshold,
    leverage,
    margin_percent,
    stop_loss_percent,
    auto_start_retreat_percent,
    profit_retreat_percent
FROM hg_trading_strategy_template
WHERE group_id = 18
ORDER BY market_state, risk_preference;

-- 2. 检查是否存在 volatile + balanced 的组合
SELECT 
    id,
    group_id,
    strategy_name,
    market_state,
    risk_preference,
    is_active
FROM hg_trading_strategy_template
WHERE group_id = 18
  AND market_state = 'volatile'
  AND risk_preference = 'balanced'
  AND is_active = 1;

-- 3. 检查策略组18是否存在
SELECT 
    id,
    group_name,
    is_official,
    is_active
FROM hg_trading_strategy_group
WHERE id = 18;

-- 4. 检查机器人25的配置
SELECT 
    id,
    robot_name,
    strategy_group_id,
    market_state,
    risk_preference,
    status
FROM hg_trading_robot
WHERE id = 25 OR id = 14;

-- 5. 列出策略组18的所有市场状态和风险偏好组合
SELECT 
    market_state,
    risk_preference,
    COUNT(*) as count,
    GROUP_CONCAT(strategy_name) as strategy_names
FROM hg_trading_strategy_template
WHERE group_id = 18
GROUP BY market_state, risk_preference;

