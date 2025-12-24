-- 查询官方V30策略组信息
SELECT 
    g.id AS '策略组ID', 
    g.group_name AS '策略组名称', 
    g.group_key AS '唯一标识',
    g.exchange AS '交易所',
    g.symbol AS '交易对',
    g.margin_mode AS '保证金模式',
    g.is_official AS '是否官方',
    g.is_active AS '是否启用'
FROM hg_trading_strategy_group g 
WHERE g.group_key = 'official_v30';

-- 查询12套策略模板详情
SELECT 
    t.id AS 'ID',
    t.sort AS '序号',
    t.strategy_name AS '策略名称',
    t.market_state AS '市场状态',
    t.risk_preference AS '风险偏好',
    t.monitor_window AS '窗口(秒)',
    t.volatility_threshold AS '阈值(U)',
    t.leverage AS '杠杆',
    t.margin_percent AS '保证金(%)',
    t.stop_loss_percent AS '止损(%)',
    t.auto_start_retreat_percent AS '启动止盈(%)',
    t.profit_retreat_percent AS '止盈回撤(%)'
FROM hg_trading_strategy_template t
JOIN hg_trading_strategy_group g ON t.group_id = g.id
WHERE g.group_key = 'official_v30'
ORDER BY t.sort;

-- 统计总数
SELECT COUNT(*) AS '策略模板总数' FROM hg_trading_strategy_template t
JOIN hg_trading_strategy_group g ON t.group_id = g.id
WHERE g.group_key = 'official_v30';

