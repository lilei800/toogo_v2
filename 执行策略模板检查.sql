-- ============================================
-- 策略模板数据检查脚本
-- 用于排查机器人参数显示错误问题
-- ============================================

-- 1. 查找策略组 "BTC-USDT 官方策略 V5.0 (我的副本)"
SELECT 
    id AS 策略组ID,
    group_name AS 策略组名称,
    is_official AS 是否官方,
    is_active AS 是否激活,
    created_at AS 创建时间
FROM hg_trading_strategy_group
WHERE group_name LIKE '%BTC-USDT 官方策略 V5.0%'
   OR group_name LIKE '%V5.0%我的副本%'
ORDER BY created_at DESC;

-- 2. 检查 volatile + balanced 的策略模板（重点检查）
SELECT 
    t.id AS 模板ID,
    t.group_id AS 策略组ID,
    g.group_name AS 策略组名称,
    t.strategy_key AS 策略KEY,
    t.strategy_name AS 策略名称,
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
    t.config_json AS 其他配置JSON,
    CASE 
        WHEN t.leverage = 10 AND t.margin_percent = 10.0 AND t.monitor_window = 60 
             AND t.volatility_threshold = 50.0 AND t.stop_loss_percent = 10.0 
             AND t.auto_start_retreat_percent = 5.0 AND t.profit_retreat_percent = 10.0 
        THEN '✅ 参数正确'
        ELSE '❌ 参数错误'
    END AS 参数检查
FROM hg_trading_strategy_template t
LEFT JOIN hg_trading_strategy_group g ON t.group_id = g.id
WHERE (g.group_name LIKE '%BTC-USDT 官方策略 V5.0%' OR g.group_name LIKE '%V5.0%我的副本%')
  AND t.market_state = 'volatile'
  AND t.risk_preference = 'balanced'
  AND t.is_active = 1;

-- 3. 如果上面查询没有结果，检查是否有其他市场状态名称
SELECT 
    t.id AS 模板ID,
    g.group_name AS 策略组名称,
    t.market_state AS 市场状态,
    t.risk_preference AS 风险偏好,
    t.leverage AS 杠杆,
    t.margin_percent AS 保证金百分比,
    t.monitor_window AS 时间窗口秒,
    t.volatility_threshold AS 波动值USDT
FROM hg_trading_strategy_template t
LEFT JOIN hg_trading_strategy_group g ON t.group_id = g.id
WHERE (g.group_name LIKE '%BTC-USDT 官方策略 V5.0%' OR g.group_name LIKE '%V5.0%我的副本%')
  AND t.risk_preference = 'balanced'
  AND t.is_active = 1
ORDER BY t.market_state;

-- 4. 检查机器人的策略组ID配置
SELECT 
    r.id AS 机器人ID,
    r.robot_name AS 机器人名称,
    r.strategy_group_id AS 策略组ID,
    g.group_name AS 策略组名称,
    r.risk_preference AS 机器人风险偏好,
    r.market_state AS 机器人市场状态
FROM hg_trading_robot r
LEFT JOIN hg_trading_strategy_group g ON r.strategy_group_id = g.id
WHERE g.group_name LIKE '%V5.0%我的副本%'
   OR g.group_name LIKE '%BTC-USDT 官方策略 V5.0%'
ORDER BY r.id DESC
LIMIT 10;

-- 5. 检查所有包含 'V5.0' 或 '我的副本' 的策略组
SELECT 
    id AS 策略组ID,
    group_name AS 策略组名称,
    is_official AS 是否官方,
    is_active AS 是否激活
FROM hg_trading_strategy_group
WHERE group_name LIKE '%V5.0%'
   OR group_name LIKE '%我的副本%'
ORDER BY created_at DESC;

-- 6. 检查市场状态值分布（看看是否有命名不一致的问题）
SELECT 
    market_state AS 市场状态,
    COUNT(*) AS 数量
FROM hg_trading_strategy_template
GROUP BY market_state
ORDER BY 数量 DESC;

-- 7. 如果发现参数错误，使用以下SQL更新（请先备份数据！）
-- UPDATE hg_trading_strategy_template t
-- LEFT JOIN hg_trading_strategy_group g ON t.group_id = g.id
-- SET 
--     t.monitor_window = 60,
--     t.volatility_threshold = 50.0,
--     t.leverage = 10,
--     t.margin_percent = 10.0,
--     t.stop_loss_percent = 10.0,
--     t.auto_start_retreat_percent = 5.0,
--     t.profit_retreat_percent = 10.0
-- WHERE (g.group_name LIKE '%BTC-USDT 官方策略 V5.0%' OR g.group_name LIKE '%V5.0%我的副本%')
--   AND t.market_state = 'volatile'
--   AND t.risk_preference = 'balanced'
--   AND t.is_active = 1;

