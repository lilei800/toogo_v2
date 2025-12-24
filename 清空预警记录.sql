-- 清空预警记录 SQL 命令

-- 方式1：清空所有预警记录（谨慎使用）
TRUNCATE TABLE hg_trading_signal_log;

-- 方式2：删除所有预警记录（保留自增ID）
DELETE FROM hg_trading_signal_log;

-- 方式3：删除指定机器人的预警记录
-- DELETE FROM hg_trading_signal_log WHERE robot_id = ?;

-- 方式4：删除指定时间之前的预警记录（例如：删除7天前的记录）
-- DELETE FROM hg_trading_signal_log WHERE created_at < DATE_SUB(NOW(), INTERVAL 7 DAY);

-- 方式5：只删除未执行的预警记录
-- DELETE FROM hg_trading_signal_log WHERE executed = 0;

