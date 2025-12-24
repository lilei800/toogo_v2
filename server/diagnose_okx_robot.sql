-- OKX机器人下单问题诊断SQL
-- 查询预警记录、执行日志、订单状态

-- 1. 查询OKX运行中的机器人
SELECT 
    id AS robot_id,
    robot_name,
    symbol,
    platform,
    status,
    auto_trade_enabled,
    auto_close_enabled,
    dual_side_position
FROM hg_trading_robot
WHERE platform = 'okx' 
  AND status = 2  -- 运行中
  AND deleted_at IS NULL
ORDER BY id DESC;

-- 2. 查询最近的预警记录（最近24小时）
SELECT 
    id AS signal_log_id,
    robot_id,
    direction,
    action,
    signal_type,
    is_processed,
    executed,
    execute_result,
    created_at
FROM hg_trading_signal_log
WHERE robot_id IN (
    SELECT id FROM hg_trading_robot 
    WHERE platform = 'okx' AND status = 2 AND deleted_at IS NULL
)
  AND created_at > NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC
LIMIT 50;

-- 3. 查询执行日志（最近24小时）
SELECT 
    id,
    signal_log_id,
    robot_id,
    order_id,
    event_type,
    status,
    message,
    event_data,
    created_at
FROM hg_trading_execution_log
WHERE robot_id IN (
    SELECT id FROM hg_trading_robot 
    WHERE platform = 'okx' AND status = 2 AND deleted_at IS NULL
)
  AND created_at > NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC
LIMIT 100;

-- 4. 统计执行日志的失败原因
SELECT 
    message AS failure_reason,
    COUNT(*) AS count,
    MAX(created_at) AS last_occurrence
FROM hg_trading_execution_log
WHERE robot_id IN (
    SELECT id FROM hg_trading_robot 
    WHERE platform = 'okx' AND status = 2 AND deleted_at IS NULL
)
  AND status = 'failed'
  AND created_at > NOW() - INTERVAL '24 hours'
GROUP BY message
ORDER BY count DESC;

-- 5. 查询最近的订单状态
SELECT 
    id AS order_id,
    robot_id,
    direction,
    status,
    exchange_order_id,
    entry_price,
    quantity,
    realized_profit,
    created_at
FROM hg_trading_order
WHERE robot_id IN (
    SELECT id FROM hg_trading_robot 
    WHERE platform = 'okx' AND status = 2 AND deleted_at IS NULL
)
ORDER BY created_at DESC
LIMIT 20;

-- 6. 查询预警记录和执行日志的关联情况
SELECT 
    s.id AS signal_log_id,
    s.robot_id,
    s.direction,
    s.is_processed,
    s.executed,
    s.execute_result,
    s.created_at AS signal_created_at,
    e.id AS execution_log_id,
    e.event_type,
    e.status,
    e.message,
    e.created_at AS execution_created_at
FROM hg_trading_signal_log s
LEFT JOIN hg_trading_execution_log e ON s.id = e.signal_log_id
WHERE s.robot_id IN (
    SELECT id FROM hg_trading_robot 
    WHERE platform = 'okx' AND status = 2 AND deleted_at IS NULL
)
  AND s.created_at > NOW() - INTERVAL '24 hours'
ORDER BY s.created_at DESC
LIMIT 30;

-- 7. 查询有预警记录但没有执行日志的情况
SELECT 
    s.id AS signal_log_id,
    s.robot_id,
    s.direction,
    s.action,
    s.is_processed,
    s.executed,
    s.created_at
FROM hg_trading_signal_log s
WHERE s.robot_id IN (
    SELECT id FROM hg_trading_robot 
    WHERE platform = 'okx' AND status = 2 AND deleted_at IS NULL
)
  AND s.created_at > NOW() - INTERVAL '24 hours'
  AND NOT EXISTS (
      SELECT 1 FROM hg_trading_execution_log e 
      WHERE e.signal_log_id = s.id
  )
ORDER BY s.created_at DESC
LIMIT 20;

-- 8. 查询执行日志中的详细失败原因（解析event_data）
SELECT 
    id,
    robot_id,
    signal_log_id,
    event_type,
    status,
    message,
    event_data::jsonb->>'step' AS failure_step,
    event_data::jsonb->>'autoTradeEnabled' AS auto_trade_enabled,
    event_data::jsonb->>'dualSidePosition' AS dual_side_position,
    event_data::jsonb->>'positionSide' AS position_side,
    created_at
FROM hg_trading_execution_log
WHERE robot_id IN (
    SELECT id FROM hg_trading_robot 
    WHERE platform = 'okx' AND status = 2 AND deleted_at IS NULL
)
  AND status = 'failed'
  AND created_at > NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC
LIMIT 50;

