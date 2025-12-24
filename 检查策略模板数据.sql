-- 检查策略模板数据
-- 用于排查机器人参数显示错误问题

-- 1. 查找策略组 "BTC-USDT 官方策略 V5.0 (我的副本)"
SELECT 
    id,
    group_name,
    is_official,
    is_active,
    created_at
FROM hg_trading_strategy_group
WHERE group_name LIKE '%BTC-USDT 官方策略 V5.0%'
   OR group_name LIKE '%V5.0%我的副本%'
ORDER BY created_at DESC;

-- 2. 查看该策略组下的所有策略模板
SELECT 
    t.id,
    t.group_id,
    g.group_name,
    t.strategy_key,
    t.strategy_name,
    t.market_state,
    t.risk_preference,
    t.monitor_window AS 时间窗口秒,
    t.volatility_threshold AS 波动值USDT,
    t.leverage AS 杠杆,
    t.margin_percent AS 保证金百分比,
    t.stop_loss_percent AS 止损百分比,
    t.auto_start_retreat_percent AS 启动止盈百分比,
    t.profit_retreat_percent AS 止盈回撤百分比,
    t.is_active,
    t.created_at
FROM hg_trading_strategy_template t
LEFT JOIN hg_trading_strategy_group g ON t.group_id = g.id
WHERE g.group_name LIKE '%BTC-USDT 官方策略 V5.0%'
   OR g.group_name LIKE '%V5.0%我的副本%'
ORDER BY t.market_state, t.risk_preference;

-- 3. 重点检查：震荡(volatile) + 平衡(balanced) 的策略模板
SELECT 
    t.id,
    t.group_id,
    g.group_name,
    t.strategy_key,
    t.strategy_name,
    t.market_state AS 市场状态,
    t.risk_preference AS 风险偏好,
    t.monitor_window AS 时间窗口秒,
    t.volatility_threshold AS 波动值USDT,
    t.leverage AS 杠杆,
    t.margin_percent AS 保证金百分比,
    t.stop_loss_percent AS 止损百分比,
    t.auto_start_retreat_percent AS 启动止盈百分比,
    t.profit_retreat_percent AS 止盈回撤百分比,
    t.is_active AS 是否激活,
    t.config_json AS 其他配置JSON
FROM hg_trading_strategy_template t
LEFT JOIN hg_trading_strategy_group g ON t.group_id = g.id
WHERE (g.group_name LIKE '%BTC-USDT 官方策略 V5.0%' OR g.group_name LIKE '%V5.0%我的副本%')
  AND t.market_state = 'volatile'
  AND t.risk_preference = 'balanced'
  AND t.is_active = 1;

-- 4. 检查机器人的策略组ID配置
SELECT 
    r.id AS 机器人ID,
    r.robot_name AS 机器人名称,
    r.strategy_group_id AS 策略组ID,
    g.group_name AS 策略组名称,
    r.risk_preference AS 机器人风险偏好,
    r.market_state AS 机器人市场状态,
    r.leverage AS 机器人杠杆,
    r.margin_percent AS 机器人保证金,
    r.stop_loss_percent AS 机器人止损,
    r.auto_start_retreat_percent AS 机器人启动止盈,
    r.profit_retreat_percent AS 机器人止盈回撤
FROM hg_trading_robot r
LEFT JOIN hg_trading_strategy_group g ON r.strategy_group_id = g.id
WHERE g.group_name LIKE '%BTC-USDT 官方策略 V5.0%'
   OR g.group_name LIKE '%V5.0%我的副本%'
ORDER BY r.id DESC
LIMIT 10;

-- 5. 检查所有策略组，查找包含 "V5.0" 或 "我的副本" 的
SELECT 
    id,
    group_name,
    is_official,
    is_active,
    created_at
FROM hg_trading_strategy_group
WHERE group_name LIKE '%V5.0%'
   OR group_name LIKE '%我的副本%'
ORDER BY created_at DESC;

-- 6. 检查市场状态为 volatile 的所有策略模板（看看是否有多个）
SELECT 
    t.id,
    t.group_id,
    g.group_name,
    t.market_state,
    t.risk_preference,
    t.leverage,
    t.margin_percent,
    t.stop_loss_percent,
    t.monitor_window,
    t.volatility_threshold,
    t.is_active
FROM hg_trading_strategy_template t
LEFT JOIN hg_trading_strategy_group g ON t.group_id = g.id
WHERE t.market_state = 'volatile'
  AND t.risk_preference = 'balanced'
  AND t.is_active = 1
ORDER BY g.group_name, t.id;

-- 7. 检查是否有 market_state 字段值不一致的情况（volatile vs range）
SELECT 
    market_state,
    COUNT(*) AS 数量
FROM hg_trading_strategy_template
GROUP BY market_state
ORDER BY 数量 DESC;

-- 8. 检查策略模板的 config_json 字段（可能包含反向下单阈值）
SELECT 
    id,
    group_id,
    market_state,
    risk_preference,
    leverage,
    margin_percent,
    config_json
FROM hg_trading_strategy_template
WHERE market_state = 'volatile'
  AND risk_preference = 'balanced'
  AND is_active = 1
LIMIT 5;

