-- ============================================================
-- 创建交易所订单事实表（挂单/订单）- PostgreSQL
-- 目标：前端“挂单列表/订单状态”统一只读 DB；私有WS增量 upsert；REST 低频兜底对账。
-- ============================================================

CREATE TABLE IF NOT EXISTS hg_trading_exchange_order (
  id BIGSERIAL PRIMARY KEY,
  tenant_id BIGINT NOT NULL DEFAULT 0,
  user_id BIGINT NOT NULL DEFAULT 0,
  robot_id BIGINT NOT NULL DEFAULT 0,
  api_config_id BIGINT NOT NULL DEFAULT 0,
  platform VARCHAR(20) NOT NULL,
  symbol VARCHAR(50) NOT NULL,

  exchange_order_id VARCHAR(120) NOT NULL,
  client_order_id VARCHAR(120),

  side VARCHAR(10),
  position_side VARCHAR(10),
  order_type VARCHAR(30),
  reduce_only BOOLEAN NOT NULL DEFAULT FALSE,

  price NUMERIC(20,8),
  quantity NUMERIC(28,12),
  filled_qty NUMERIC(28,12),
  avg_price NUMERIC(20,8),

  status VARCHAR(30),
  raw_status VARCHAR(50),
  is_open BOOLEAN NOT NULL DEFAULT TRUE,

  create_time BIGINT,
  update_time BIGINT,
  last_event_time BIGINT,
  raw TEXT,

  created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_platform_api_order
  ON hg_trading_exchange_order(platform, api_config_id, exchange_order_id);

CREATE INDEX IF NOT EXISTS idx_robot_open
  ON hg_trading_exchange_order(robot_id, is_open);

CREATE INDEX IF NOT EXISTS idx_robot_symbol_open
  ON hg_trading_exchange_order(robot_id, symbol, is_open);

CREATE INDEX IF NOT EXISTS idx_user_open
  ON hg_trading_exchange_order(user_id, is_open);

CREATE INDEX IF NOT EXISTS idx_update_time
  ON hg_trading_exchange_order(update_time);


