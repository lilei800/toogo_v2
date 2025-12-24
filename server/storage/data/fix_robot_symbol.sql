-- ================================================
-- 🔧 修复机器人Symbol字段
-- ================================================
-- 功能: 为缺失Symbol的机器人补充默认值
-- ================================================

USE hotgo;

-- 查看当前机器人数据
SELECT id, robot_name, symbol, exchange, api_config_id, status
FROM hg_trading_robot
ORDER BY id DESC
LIMIT 20;

-- 修复缺失的Symbol字段
UPDATE hg_trading_robot
SET symbol = 'BTCUSDT'
WHERE symbol IS NULL OR symbol = '' OR TRIM(symbol) = '';

-- 修复缺失的Exchange字段
UPDATE hg_trading_robot  
SET exchange = 'binance'
WHERE exchange IS NULL OR exchange = '' OR TRIM(exchange) = '';

-- 验证修复结果
SELECT 
    COUNT(*) as total_robots,
    SUM(CASE WHEN symbol IS NULL OR symbol = '' THEN 1 ELSE 0 END) as missing_symbol,
    SUM(CASE WHEN exchange IS NULL OR exchange = '' THEN 1 ELSE 0 END) as missing_exchange
FROM hg_trading_robot;

SELECT '✅ 机器人Symbol字段修复完成！' AS status;

