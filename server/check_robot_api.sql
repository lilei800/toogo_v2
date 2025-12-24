-- 检查机器人ID=50的配置
SELECT 
    r.id AS robot_id,
    r.name AS robot_name,
    r.platform,
    r.symbol,
    r.status,
    a.id AS api_config_id,
    a.api_key,
    a.status AS api_status
FROM hg_trading_robot r
LEFT JOIN hg_trading_api_config a ON r.api_config_id = a.id
WHERE r.id = 50;

