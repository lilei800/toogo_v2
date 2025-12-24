-- ============================================================
-- 快速诊断：预警不下单问题
-- 执行此脚本可以快速定位问题
-- ============================================================

\echo '========================================='
\echo '诊断1: 检查 is_processed 字段是否存在'
\echo '========================================='

SELECT 
    CASE 
        WHEN COUNT(*) > 0 THEN '✓ is_processed 字段存在'
        ELSE '✗ is_processed 字段不存在 - 需要执行修复脚本'
    END as result,
    MAX(data_type) as data_type,
    MAX(column_default::TEXT) as default_value,
    MAX(is_nullable) as nullable
FROM information_schema.columns
WHERE table_name = 'hg_trading_signal_log' 
  AND column_name = 'is_processed';

\echo ''
\echo '========================================='
\echo '诊断2: 检查其他必需字段'
\echo '========================================='

SELECT 
    column_name,
    data_type,
    CASE WHEN is_nullable = 'YES' THEN '允许NULL' ELSE '不允许NULL' END as nullable,
    COALESCE(column_default::TEXT, '无') as default_value
FROM information_schema.columns
WHERE table_name = 'hg_trading_signal_log' 
  AND column_name IN ('is_processed', 'window_min_price', 'window_max_price', 'threshold', 'market_state')
ORDER BY 
    CASE column_name
        WHEN 'is_processed' THEN 1
        WHEN 'window_min_price' THEN 2
        WHEN 'window_max_price' THEN 3
        WHEN 'threshold' THEN 4
        WHEN 'market_state' THEN 5
    END;

\echo ''
\echo '========================================='
\echo '诊断3: 检查最近1小时的预警记录'
\echo '========================================='

SELECT 
    COUNT(*) as total_alerts,
    SUM(CASE WHEN is_processed = 0 THEN 1 ELSE 0 END) as unprocessed,
    SUM(CASE WHEN is_processed = 1 THEN 1 ELSE 0 END) as processed,
    SUM(CASE WHEN executed = 1 THEN 1 ELSE 0 END) as executed,
    COUNT(*) - SUM(CASE WHEN executed = 1 THEN 1 ELSE 0 END) as not_executed
FROM hg_trading_signal_log
WHERE created_at >= NOW() - INTERVAL '1 hour';

\echo ''
\echo '========================================='
\echo '诊断4: 最近10条预警记录详情'
\echo '========================================='

SELECT 
    id,
    robot_id,
    signal_type,
    ROUND(current_price::NUMERIC, 2) as price,
    CASE 
        WHEN is_processed = 0 THEN '未处理'
        WHEN is_processed = 1 THEN '已处理/已读'
        ELSE '未知'
    END as process_status,
    CASE 
        WHEN executed = 0 THEN '未执行'
        WHEN executed = 1 THEN '已执行'
        ELSE '未知'
    END as execute_status,
    TO_CHAR(created_at, 'HH24:MI:SS') as time,
    LEFT(reason, 60) as reason
FROM hg_trading_signal_log
WHERE created_at >= NOW() - INTERVAL '2 hour'
ORDER BY id DESC
LIMIT 10;

\echo ''
\echo '========================================='
\echo '诊断5: 检查机器人配置'
\echo '========================================='

SELECT 
    id,
    name,
    symbol,
    CASE 
        WHEN auto_trade_enabled = 1 THEN '✓ 已开启'
        WHEN auto_trade_enabled = 0 THEN '✗ 未开启'
        ELSE '未知'
    END as auto_trade,
    CASE 
        WHEN dual_side_position = 1 THEN '双向模式'
        WHEN dual_side_position = 0 THEN '单向模式'
        ELSE '未知'
    END as position_mode,
    CASE 
        WHEN status = 1 THEN '运行中'
        WHEN status = 0 THEN '已停止'
        ELSE '未知'
    END as status
FROM hg_trading_robot
WHERE id IN (
    SELECT DISTINCT robot_id 
    FROM hg_trading_signal_log 
    WHERE created_at >= NOW() - INTERVAL '2 hour'
)
LIMIT 5;

\echo ''
\echo '========================================='
\echo '诊断6: 检查交易执行日志'
\echo '========================================='

-- 检查表是否存在
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'hg_trading_execution_log') THEN
        RAISE NOTICE '✓ hg_trading_execution_log 表存在';
    ELSE
        RAISE NOTICE '✗ hg_trading_execution_log 表不存在 - 需要创建';
    END IF;
END $$;

-- 如果表存在，查询最近的执行日志
SELECT 
    el.id,
    el.signal_log_id,
    el.robot_id,
    el.event_type,
    el.status,
    LEFT(el.message, 80) as message,
    TO_CHAR(el.created_at, 'HH24:MI:SS') as time
FROM hg_trading_execution_log el
WHERE el.created_at >= NOW() - INTERVAL '2 hour'
ORDER BY el.id DESC
LIMIT 10;

\echo ''
\echo '========================================='
\echo '诊断7: 问题分析'
\echo '========================================='

DO $$
DECLARE
    has_is_processed BOOLEAN;
    unprocessed_count INTEGER;
    processed_not_executed INTEGER;
    auto_trade_disabled INTEGER;
BEGIN
    -- 检查 is_processed 字段
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'hg_trading_signal_log' 
          AND column_name = 'is_processed'
    ) INTO has_is_processed;
    
    -- 统计未处理的预警
    SELECT COUNT(*) INTO unprocessed_count
    FROM hg_trading_signal_log
    WHERE created_at >= NOW() - INTERVAL '1 hour'
      AND (is_processed = 0 OR is_processed IS NULL);
    
    -- 统计已处理但未执行的预警
    SELECT COUNT(*) INTO processed_not_executed
    FROM hg_trading_signal_log
    WHERE created_at >= NOW() - INTERVAL '1 hour'
      AND is_processed = 1
      AND executed = 0;
    
    -- 统计自动交易未开启的机器人
    SELECT COUNT(*) INTO auto_trade_disabled
    FROM hg_trading_robot
    WHERE id IN (
        SELECT DISTINCT robot_id 
        FROM hg_trading_signal_log 
        WHERE created_at >= NOW() - INTERVAL '1 hour'
    )
    AND auto_trade_enabled != 1;
    
    RAISE NOTICE '';
    RAISE NOTICE '问题分析结果：';
    RAISE NOTICE '----------------------------------------';
    
    IF NOT has_is_processed THEN
        RAISE NOTICE '❌ 严重问题：is_processed 字段不存在';
        RAISE NOTICE '   → 解决方案：执行 fix_is_processed_postgresql.sql';
    ELSE
        RAISE NOTICE '✓ is_processed 字段存在';
    END IF;
    
    IF unprocessed_count > 0 THEN
        RAISE NOTICE '⚠ 发现 % 条未处理的预警记录', unprocessed_count;
        RAISE NOTICE '   → 可能原因：保存预警记录失败（logId=0）';
        RAISE NOTICE '   → 或：重试机制尚未处理';
    ELSE
        RAISE NOTICE '✓ 没有未处理的预警记录';
    END IF;
    
    IF processed_not_executed > 0 THEN
        RAISE NOTICE '⚠ 发现 % 条已处理但未执行的预警', processed_not_executed;
        RAISE NOTICE '   → 可能原因1：自动交易开关未开启';
        RAISE NOTICE '   → 可能原因2：已有持仓，被持仓检查阻止';
        RAISE NOTICE '   → 可能原因3：获取锁失败';
        RAISE NOTICE '   → 建议：查看 hg_trading_execution_log 表的详细日志';
    ELSE
        RAISE NOTICE '✓ 没有已处理但未执行的预警';
    END IF;
    
    IF auto_trade_disabled > 0 THEN
        RAISE NOTICE '⚠ 发现 % 个机器人的自动交易开关未开启', auto_trade_disabled;
        RAISE NOTICE '   → 解决方案：在机器人管理页面开启自动交易';
    ELSE
        RAISE NOTICE '✓ 所有相关机器人的自动交易开关已开启';
    END IF;
    
    RAISE NOTICE '';
END $$;

\echo ''
\echo '========================================='
\echo '诊断8: 建议操作'
\echo '========================================='

DO $$
DECLARE
    has_is_processed BOOLEAN;
    unprocessed_count INTEGER;
    processed_not_executed INTEGER;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'hg_trading_signal_log' 
          AND column_name = 'is_processed'
    ) INTO has_is_processed;
    
    SELECT COUNT(*) INTO unprocessed_count
    FROM hg_trading_signal_log
    WHERE created_at >= NOW() - INTERVAL '1 hour'
      AND (is_processed = 0 OR is_processed IS NULL);
    
    SELECT COUNT(*) INTO processed_not_executed
    FROM hg_trading_signal_log
    WHERE created_at >= NOW() - INTERVAL '1 hour'
      AND is_processed = 1
      AND executed = 0;
    
    RAISE NOTICE '';
    RAISE NOTICE '建议操作步骤：';
    RAISE NOTICE '----------------------------------------';
    
    IF NOT has_is_processed THEN
        RAISE NOTICE '1. 【必须】执行修复脚本添加 is_processed 字段：';
        RAISE NOTICE '   psql -U your_user -d your_db -f fix_is_processed_postgresql.sql';
    ELSIF processed_not_executed > 0 THEN
        RAISE NOTICE '1. 查看执行日志，了解具体失败原因：';
        RAISE NOTICE '   SELECT * FROM hg_trading_execution_log';
        RAISE NOTICE '   WHERE created_at >= NOW() - INTERVAL ''1 hour''';
        RAISE NOTICE '   ORDER BY id DESC LIMIT 20;';
        RAISE NOTICE '';
        RAISE NOTICE '2. 检查机器人配置：';
        RAISE NOTICE '   - 自动交易开关是否开启';
        RAISE NOTICE '   - 是否已有持仓（单向/双向模式）';
        RAISE NOTICE '   - 持仓方向是否冲突';
    ELSIF unprocessed_count > 0 THEN
        RAISE NOTICE '1. 等待重试机制处理（每分钟自动重试）';
        RAISE NOTICE '2. 或者重启服务，触发立即重试';
    ELSE
        RAISE NOTICE '✓ 系统运行正常，没有发现问题';
        RAISE NOTICE '  等待下一个交易信号，观察是否正常下单';
    END IF;
    
    RAISE NOTICE '';
    RAISE NOTICE '详细诊断报告：预警不下单问题诊断报告.md';
    RAISE NOTICE '========================================';
END $$;

-- ============================================================
-- 诊断完成
-- ============================================================

