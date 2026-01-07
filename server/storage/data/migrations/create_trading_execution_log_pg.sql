-- ============================================================
-- 创建交易执行日志表 - PostgreSQL
-- 用途：前端“订单日志/执行日志”读取本表（/trading/robot/executionLogs）。
-- 说明：代码写入可能包含 failure_category/failure_reason 字段；旧库缺字段时会降级写入。
-- ============================================================

CREATE TABLE IF NOT EXISTS hg_trading_execution_log (
  id BIGSERIAL PRIMARY KEY,
  signal_log_id BIGINT NOT NULL DEFAULT 0,
  robot_id BIGINT NOT NULL DEFAULT 0,
  order_id BIGINT NOT NULL DEFAULT 0,
  event_type VARCHAR(50) NOT NULL DEFAULT '',
  event_data TEXT,
  status VARCHAR(20) NOT NULL DEFAULT '',
  message TEXT NOT NULL DEFAULT '',
  -- 可选：用于前端更好展示失败原因（代码会兼容缺字段降级写入）
  failure_category VARCHAR(50) NOT NULL DEFAULT '',
  failure_reason TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_trading_execution_log_signal_log_id
  ON hg_trading_execution_log(signal_log_id);

CREATE INDEX IF NOT EXISTS idx_trading_execution_log_robot_time
  ON hg_trading_execution_log(robot_id, created_at);


