-- =============================================
-- Toogo/HotGo Trading Core Tables (PostgreSQL)
-- =============================================
-- 说明：
-- - 本文件用于 PostgreSQL 环境，补齐交易系统核心表。
-- - 你之前执行的 `trading_system.sql` 是 MySQL 版本，不能用于 PG。
-- - 已存在的表会被 IF NOT EXISTS 跳过。
--
-- 建议执行方式：
--   PGPASSWORD='你的密码' psql -h 127.0.0.1 -U hotgo_user -d hotgo -f storage/data/migrations/create_trading_core_pg.sql
-- =============================================

BEGIN;

-- -----------------------------
-- 交易 API 配置表
-- -----------------------------
CREATE TABLE IF NOT EXISTS hg_trading_api_config (
  id               BIGSERIAL PRIMARY KEY,
  tenant_id        BIGINT NOT NULL DEFAULT 0,
  user_id          BIGINT NOT NULL DEFAULT 0,
  api_name         VARCHAR(100) NOT NULL DEFAULT '',
  platform         VARCHAR(20)  NOT NULL DEFAULT '',
  base_url         VARCHAR(255) NOT NULL DEFAULT '',
  api_key          TEXT NOT NULL DEFAULT '',
  secret_key       TEXT NOT NULL DEFAULT '',
  passphrase       TEXT NOT NULL DEFAULT '',
  is_default       SMALLINT NOT NULL DEFAULT 0,
  status           SMALLINT NOT NULL DEFAULT 1,
  last_verify_time TIMESTAMPTZ NULL,
  verify_status    SMALLINT NOT NULL DEFAULT 0,
  verify_message   VARCHAR(500) NOT NULL DEFAULT '',
  remark           VARCHAR(500) NOT NULL DEFAULT '',
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at       TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_trading_api_config_tenant_user ON hg_trading_api_config(tenant_id, user_id);
CREATE INDEX IF NOT EXISTS idx_trading_api_config_platform    ON hg_trading_api_config(platform);
CREATE INDEX IF NOT EXISTS idx_trading_api_config_status      ON hg_trading_api_config(status);

-- -----------------------------
-- 交易代理配置表
-- -----------------------------
CREATE TABLE IF NOT EXISTS hg_trading_proxy_config (
  id             BIGSERIAL PRIMARY KEY,
  tenant_id      BIGINT NOT NULL DEFAULT 0,
  user_id        BIGINT NOT NULL DEFAULT 0,
  enabled        SMALLINT NOT NULL DEFAULT 0,
  proxy_type     VARCHAR(20) NOT NULL DEFAULT '',
  proxy_address  VARCHAR(255) NOT NULL DEFAULT '',
  auth_enabled   SMALLINT NOT NULL DEFAULT 0,
  username       VARCHAR(100) NOT NULL DEFAULT '',
  password       TEXT NOT NULL DEFAULT '',
  last_test_time TIMESTAMPTZ NULL,
  test_status    SMALLINT NOT NULL DEFAULT 0,
  test_message   VARCHAR(500) NOT NULL DEFAULT '',
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_trading_proxy_config_tenant_user ON hg_trading_proxy_config(tenant_id, user_id);

-- -----------------------------
-- 策略组表（你的代码大量使用）
-- -----------------------------
CREATE TABLE IF NOT EXISTS hg_trading_strategy_group (
  id              BIGSERIAL PRIMARY KEY,
  group_name      VARCHAR(100) NOT NULL DEFAULT '',
  group_key       VARCHAR(120) NOT NULL DEFAULT '',
  exchange        VARCHAR(20)  NOT NULL DEFAULT '',
  symbol          VARCHAR(30)  NOT NULL DEFAULT '',
  order_type      VARCHAR(20)  NOT NULL DEFAULT '',
  margin_mode     VARCHAR(20)  NOT NULL DEFAULT '',
  is_official     SMALLINT NOT NULL DEFAULT 0,
  from_official_id BIGINT NOT NULL DEFAULT 0,
  is_default      SMALLINT NOT NULL DEFAULT 0,
  user_id         BIGINT NOT NULL DEFAULT 0,
  description     VARCHAR(500) NOT NULL DEFAULT '',
  is_active       SMALLINT NOT NULL DEFAULT 1,
  sort            INT NOT NULL DEFAULT 100,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_trading_strategy_group_key ON hg_trading_strategy_group(group_key);
CREATE INDEX IF NOT EXISTS idx_trading_strategy_group_symbol    ON hg_trading_strategy_group(exchange, symbol);

-- -----------------------------
-- 策略模板表（对应 entity/trading_strategy_template.go）
-- -----------------------------
CREATE TABLE IF NOT EXISTS hg_trading_strategy_template (
  id                      BIGSERIAL PRIMARY KEY,
  group_id                BIGINT NOT NULL DEFAULT 0,
  strategy_key            VARCHAR(120) NOT NULL DEFAULT '',
  strategy_name           VARCHAR(120) NOT NULL DEFAULT '',
  risk_preference         VARCHAR(30)  NOT NULL DEFAULT '',
  market_state            VARCHAR(30)  NOT NULL DEFAULT '',
  monitor_window          INT NOT NULL DEFAULT 0,
  volatility_threshold    NUMERIC(20,8) NOT NULL DEFAULT 0,
  leverage                INT NOT NULL DEFAULT 0,
  margin_percent          NUMERIC(10,4) NOT NULL DEFAULT 0,
  stop_loss_percent       NUMERIC(10,4) NOT NULL DEFAULT 0,
  profit_retreat_percent  NUMERIC(10,4) NOT NULL DEFAULT 0,
  auto_start_retreat_percent NUMERIC(10,4) NOT NULL DEFAULT 0,
  config_json             TEXT NOT NULL DEFAULT '',
  description             VARCHAR(500) NOT NULL DEFAULT '',
  is_active               SMALLINT NOT NULL DEFAULT 1,
  sort                    INT NOT NULL DEFAULT 100,
  created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_trading_strategy_template_key ON hg_trading_strategy_template(strategy_key);
CREATE INDEX IF NOT EXISTS idx_trading_strategy_template_group     ON hg_trading_strategy_template(group_id);
CREATE INDEX IF NOT EXISTS idx_trading_strategy_template_risk_ms   ON hg_trading_strategy_template(risk_preference, market_state);

-- -----------------------------
-- 交易机器人表（核心缺失：MarketAnalyzer 会查它）
-- 对应 entity/trading_robot.go
-- -----------------------------
CREATE TABLE IF NOT EXISTS hg_trading_robot (
  id                      BIGSERIAL PRIMARY KEY,
  tenant_id               BIGINT NOT NULL DEFAULT 0,
  user_id                 BIGINT NOT NULL DEFAULT 0,
  robot_name              VARCHAR(120) NOT NULL DEFAULT '',
  api_config_id           BIGINT NOT NULL DEFAULT 0,
  max_profit_target       NUMERIC(20,8) NOT NULL DEFAULT 0,
  max_loss_amount         NUMERIC(20,8) NOT NULL DEFAULT 0,
  max_runtime             INT NOT NULL DEFAULT 0,
  risk_preference         VARCHAR(30) NOT NULL DEFAULT '',
  auto_risk_preference    SMALLINT NOT NULL DEFAULT 0,
  market_state            VARCHAR(30) NOT NULL DEFAULT '',
  auto_market_state       SMALLINT NOT NULL DEFAULT 0,
  exchange                VARCHAR(20) NOT NULL DEFAULT '',
  symbol                  VARCHAR(30) NOT NULL DEFAULT '',
  order_type              VARCHAR(20) NOT NULL DEFAULT '',
  margin_mode             VARCHAR(20) NOT NULL DEFAULT '',
  leverage                INT NOT NULL DEFAULT 0,
  margin_percent          NUMERIC(10,4) NOT NULL DEFAULT 0,
  use_monitor_signal      SMALLINT NOT NULL DEFAULT 0,
  stop_loss_percent       NUMERIC(10,4) NOT NULL DEFAULT 0,
  profit_retreat_percent  NUMERIC(10,4) NOT NULL DEFAULT 0,
  auto_start_retreat_percent NUMERIC(10,4) NOT NULL DEFAULT 0,
  current_strategy        TEXT NOT NULL DEFAULT '',
  strategy_group_id       BIGINT NOT NULL DEFAULT 0,
  status                  SMALLINT NOT NULL DEFAULT 1,
  start_time              TIMESTAMPTZ NULL,
  pause_time              TIMESTAMPTZ NULL,
  stop_time               TIMESTAMPTZ NULL,
  long_count              INT NOT NULL DEFAULT 0,
  short_count             INT NOT NULL DEFAULT 0,
  total_profit            NUMERIC(20,8) NOT NULL DEFAULT 0,
  runtime_seconds         INT NOT NULL DEFAULT 0,
  auto_trade_enabled      SMALLINT NOT NULL DEFAULT 0,
  auto_close_enabled      SMALLINT NOT NULL DEFAULT 1,
  profit_lock_enabled     SMALLINT NOT NULL DEFAULT 0,
  dual_side_position      SMALLINT NOT NULL DEFAULT 0,
  schedule_start          TIMESTAMPTZ NULL,
  schedule_stop           TIMESTAMPTZ NULL,
  remark                  VARCHAR(500) NOT NULL DEFAULT '',
  created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at              TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_trading_robot_tenant_user ON hg_trading_robot(tenant_id, user_id);
CREATE INDEX IF NOT EXISTS idx_trading_robot_status      ON hg_trading_robot(status);
CREATE INDEX IF NOT EXISTS idx_trading_robot_api_config  ON hg_trading_robot(api_config_id);
CREATE INDEX IF NOT EXISTS idx_trading_robot_symbol      ON hg_trading_robot(symbol);
CREATE INDEX IF NOT EXISTS idx_trading_robot_exchange_symbol ON hg_trading_robot(exchange, symbol);

-- -----------------------------
-- 交易订单表（对应 entity/trading_order.go）
-- -----------------------------
CREATE TABLE IF NOT EXISTS hg_trading_order (
  id                     BIGSERIAL PRIMARY KEY,
  tenant_id              BIGINT NOT NULL DEFAULT 0,
  user_id                BIGINT NOT NULL DEFAULT 0,
  robot_id               BIGINT NOT NULL DEFAULT 0,
  order_sn               VARCHAR(80) NOT NULL DEFAULT '',
  exchange_order_id      VARCHAR(120) NOT NULL DEFAULT '',
  symbol                 VARCHAR(30) NOT NULL DEFAULT '',
  direction              VARCHAR(10) NOT NULL DEFAULT '',
  open_price             NUMERIC(20,8) NOT NULL DEFAULT 0,
  close_price            NUMERIC(20,8) NOT NULL DEFAULT 0,
  quantity               NUMERIC(30,12) NOT NULL DEFAULT 0,
  leverage               INT NOT NULL DEFAULT 0,
  margin                 NUMERIC(20,8) NOT NULL DEFAULT 0,
  realized_profit        NUMERIC(20,8) NOT NULL DEFAULT 0,
  unrealized_profit      NUMERIC(20,8) NOT NULL DEFAULT 0,
  highest_profit         NUMERIC(20,8) NOT NULL DEFAULT 0,
  stop_loss_price        NUMERIC(20,8) NOT NULL DEFAULT 0,
  profit_retreat_started SMALLINT NOT NULL DEFAULT 0,
  profit_retreat_percent NUMERIC(10,4) NOT NULL DEFAULT 0,
  open_time              TIMESTAMPTZ NULL,
  close_time             TIMESTAMPTZ NULL,
  hold_duration          INT NOT NULL DEFAULT 0,
  status                 SMALLINT NOT NULL DEFAULT 1,
  close_reason           VARCHAR(50) NOT NULL DEFAULT '',
  remark                 VARCHAR(500) NOT NULL DEFAULT '',
  created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_trading_order_sn ON hg_trading_order(order_sn);
CREATE INDEX IF NOT EXISTS idx_trading_order_tenant_user ON hg_trading_order(tenant_id, user_id);
CREATE INDEX IF NOT EXISTS idx_trading_order_robot ON hg_trading_order(robot_id);
CREATE INDEX IF NOT EXISTS idx_trading_order_status ON hg_trading_order(status);
CREATE INDEX IF NOT EXISTS idx_trading_order_symbol ON hg_trading_order(symbol);

-- -----------------------------
-- 平仓日志表（对应 entity/trading_close_log.go）
-- -----------------------------
CREATE TABLE IF NOT EXISTS hg_trading_close_log (
  id                 BIGSERIAL PRIMARY KEY,
  tenant_id           BIGINT NOT NULL DEFAULT 0,
  user_id             BIGINT NOT NULL DEFAULT 0,
  robot_id            BIGINT NOT NULL DEFAULT 0,
  order_id            BIGINT NOT NULL DEFAULT 0,
  order_sn            VARCHAR(80) NOT NULL DEFAULT '',
  symbol              VARCHAR(30) NOT NULL DEFAULT '',
  direction           VARCHAR(10) NOT NULL DEFAULT '',
  open_price          NUMERIC(20,8) NOT NULL DEFAULT 0,
  close_price         NUMERIC(20,8) NOT NULL DEFAULT 0,
  quantity            NUMERIC(30,12) NOT NULL DEFAULT 0,
  leverage            INT NOT NULL DEFAULT 0,
  margin              NUMERIC(20,8) NOT NULL DEFAULT 0,
  realized_profit     NUMERIC(20,8) NOT NULL DEFAULT 0,
  highest_profit      NUMERIC(20,8) NOT NULL DEFAULT 0,
  profit_percent      NUMERIC(20,8) NOT NULL DEFAULT 0,
  close_reason        VARCHAR(80) NOT NULL DEFAULT '',
  close_detail        TEXT NOT NULL DEFAULT '',
  open_fee            NUMERIC(20,8) NOT NULL DEFAULT 0,
  hold_fee            NUMERIC(20,8) NOT NULL DEFAULT 0,
  close_fee           NUMERIC(20,8) NOT NULL DEFAULT 0,
  total_fee           NUMERIC(20,8) NOT NULL DEFAULT 0,
  commission_amount   NUMERIC(20,8) NOT NULL DEFAULT 0,
  commission_percent  NUMERIC(20,8) NOT NULL DEFAULT 0,
  net_profit          NUMERIC(20,8) NOT NULL DEFAULT 0,
  open_time           TIMESTAMPTZ NULL,
  close_time          TIMESTAMPTZ NULL,
  hold_duration       INT NOT NULL DEFAULT 0,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_trading_close_log_tenant_user ON hg_trading_close_log(tenant_id, user_id);
CREATE INDEX IF NOT EXISTS idx_trading_close_log_robot      ON hg_trading_close_log(robot_id);
CREATE INDEX IF NOT EXISTS idx_trading_close_log_order      ON hg_trading_close_log(order_id);
CREATE INDEX IF NOT EXISTS idx_trading_close_log_close_time ON hg_trading_close_log(close_time);

-- -----------------------------
-- 市场监控日志表（对应 entity/trading_monitor_log.go）
-- -----------------------------
CREATE TABLE IF NOT EXISTS hg_trading_monitor_log (
  id              BIGSERIAL PRIMARY KEY,
  tenant_id        BIGINT NOT NULL DEFAULT 0,
  user_id          BIGINT NOT NULL DEFAULT 0,
  robot_id         BIGINT NOT NULL DEFAULT 0,
  symbol           VARCHAR(30) NOT NULL DEFAULT '',
  current_price    NUMERIC(20,8) NOT NULL DEFAULT 0,
  window_high      NUMERIC(20,8) NOT NULL DEFAULT 0,
  window_low       NUMERIC(20,8) NOT NULL DEFAULT 0,
  volatility       NUMERIC(20,8) NOT NULL DEFAULT 0,
  signal_type      VARCHAR(30) NOT NULL DEFAULT '',
  signal_strength  NUMERIC(20,8) NOT NULL DEFAULT 0,
  market_state     VARCHAR(30) NOT NULL DEFAULT '',
  signal_detail    TEXT NOT NULL DEFAULT '',
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_trading_monitor_log_robot ON hg_trading_monitor_log(robot_id);
CREATE INDEX IF NOT EXISTS idx_trading_monitor_log_created_at ON hg_trading_monitor_log(created_at);

COMMIT;

