-- 检查机器人ID=50的详细状态
SELECT 
    r.id,
    r.name,
    r.platform,
    r.symbol,
    r.status,  -- 必须是1才会启动
    r.api_config_id,
    r.created_at,
    r.updated_at,
    a.id AS api_id,
    a.platform AS api_platform,
    a.status AS api_status  -- API配置也必须是1
FROM hg_trading_robot r
LEFT JOIN hg_trading_api_config a ON r.api_config_id = a.id
WHERE r.id = 50;

-- 检查所有运行中的机器人
SELECT 
    id,
    name,
    platform,
    symbol,
    status,
    api_config_id
FROM hg_trading_robot
WHERE status = 1
ORDER BY id;

