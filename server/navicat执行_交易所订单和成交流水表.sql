-- ============================================================
-- 在 Navicat 中执行：创建交易所订单事实表和成交流水表
-- 使用方法：
--   1. 打开 Navicat，连接到 hotgo 数据库
--   2. 点击"查询" → "新建查询"
--   3. 复制本文件全部内容，粘贴到查询窗口
--   4. 点击"运行"按钮执行
-- ============================================================

-- ============================================================
-- 【表1】创建交易所订单事实表（挂单/订单）- PostgreSQL
-- 目标：前端"挂单列表/订单状态"统一只读 DB；私有WS增量 upsert；REST 低频兜底对账。
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

-- 创建索引
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

-- ============================================================
-- 【表2】创建成交流水表（交易所一笔成交一条记录）- PostgreSQL
-- 表名：hg_trading_trade_fill
-- 目标：交易明细/盈亏/手续费以交易所成交为准；trading_order 作为"订单壳/意图"
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

-- 创建索引
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

-- ============================================================
-- 执行完成！
-- ============================================================
-- 创建的表：
--   1. hg_trading_exchange_order - 交易所订单事实表
--   2. hg_trading_trade_fill - 成交流水表
--
-- 用途说明：
--   【hg_trading_exchange_order】
--   - WebSocket实时推送订单数据
--   - 供前端挂单列表展示
--   - REST API低频兜底对账
--
--   【hg_trading_trade_fill】
--   - 存储交易所成交记录（每笔成交一条记录）
--   - 精确的盈亏和手续费数据
--   - 订单历史查询的权威数据源
-- ============================================================

