-- ============================================================
-- 创建成交流水表（交易所一笔成交一条记录）- PostgreSQL
-- 表名：hg_trading_trade_fill
-- 目标：交易明细/盈亏/手续费以交易所成交为准；trading_order 作为“订单壳/意图”
-- 幂等：用 (api_config_id, exchange, trade_id) 防重复落库
-- ============================================================

CREATE TABLE IF NOT EXISTS hg_trading_trade_fill (
  id BIGSERIAL PRIMARY KEY,
  tenant_id BIGINT NOT NULL DEFAULT 0,

  api_config_id BIGINT NOT NULL DEFAULT 0,
  exchange VARCHAR(32) NOT NULL DEFAULT '',

  user_id BIGINT NOT NULL DEFAULT 0,
  robot_id BIGINT NOT NULL DEFAULT 0,
  session_id BIGINT NULL,

  symbol VARCHAR(64) NOT NULL DEFAULT '',

  order_id VARCHAR(128) NOT NULL DEFAULT '',
  client_order_id VARCHAR(128) NOT NULL DEFAULT '',
  trade_id VARCHAR(128) NOT NULL DEFAULT '',

  side VARCHAR(16) NOT NULL DEFAULT '',
  qty NUMERIC(32,16) NOT NULL DEFAULT 0,
  price NUMERIC(32,16) NOT NULL DEFAULT 0,

  fee NUMERIC(32,16) NOT NULL DEFAULT 0,
  fee_coin VARCHAR(32) NOT NULL DEFAULT '',
  realized_pnl NUMERIC(32,16) NOT NULL DEFAULT 0,

  ts BIGINT NOT NULL DEFAULT 0,

  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_api_exchange_trade
  ON hg_trading_trade_fill(api_config_id, exchange, trade_id);

CREATE INDEX IF NOT EXISTS idx_user_ts
  ON hg_trading_trade_fill(user_id, ts);

CREATE INDEX IF NOT EXISTS idx_robot_ts
  ON hg_trading_trade_fill(robot_id, ts);

CREATE INDEX IF NOT EXISTS idx_symbol_ts
  ON hg_trading_trade_fill(symbol, ts);

CREATE INDEX IF NOT EXISTS idx_order_ts
  ON hg_trading_trade_fill(order_id, ts);

CREATE INDEX IF NOT EXISTS idx_session_ts
  ON hg_trading_trade_fill(session_id, ts);


