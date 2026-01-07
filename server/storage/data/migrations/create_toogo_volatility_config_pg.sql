-- =============================================
-- Toogo Volatility Config (PostgreSQL)
-- Table: hg_toogo_volatility_config
-- Purpose: Provide global/per-symbol volatility params for the new algorithm
--
-- Run:
--   PGPASSWORD='你的密码' psql -h 127.0.0.1 -U hotgo_user -d hotgo -f storage/data/migrations/create_toogo_volatility_config_pg.sql
-- =============================================

BEGIN;

CREATE TABLE IF NOT EXISTS hg_toogo_volatility_config (
  id                        BIGSERIAL PRIMARY KEY,
  symbol                    VARCHAR(50) NULL, -- NULL 表示全局配置

  high_volatility_threshold  NUMERIC(20,8) NOT NULL DEFAULT 2.0,
  low_volatility_threshold   NUMERIC(20,8) NOT NULL DEFAULT 0.9,
  trend_strength_threshold   NUMERIC(20,8) NOT NULL DEFAULT 1.2,
  d_threshold                NUMERIC(20,8) NOT NULL DEFAULT 0.7,
  range_volatility_threshold NUMERIC(20,8) NOT NULL DEFAULT 0,

  delta_1m                   NUMERIC(20,8) NOT NULL DEFAULT 2.0,
  delta_5m                   NUMERIC(20,8) NOT NULL DEFAULT 2.0,
  delta_15m                  NUMERIC(20,8) NOT NULL DEFAULT 3.0,
  delta_30m                  NUMERIC(20,8) NOT NULL DEFAULT 3.0,
  delta_1h                   NUMERIC(20,8) NOT NULL DEFAULT 5.0,

  weight_1m                  NUMERIC(20,8) NOT NULL DEFAULT 0.18,
  weight_5m                  NUMERIC(20,8) NOT NULL DEFAULT 0.25,
  weight_15m                 NUMERIC(20,8) NOT NULL DEFAULT 0.27,
  weight_30m                 NUMERIC(20,8) NOT NULL DEFAULT 0.20,
  weight_1h                  NUMERIC(20,8) NOT NULL DEFAULT 0.10,

  is_active                  SMALLINT NOT NULL DEFAULT 1,
  created_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at                 TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 对 symbol 建唯一索引（PG 允许多个 NULL；用于保证每个交易对最多一条配置）
CREATE UNIQUE INDEX IF NOT EXISTS uk_hg_toogo_volatility_config_symbol
  ON hg_toogo_volatility_config(symbol);

CREATE INDEX IF NOT EXISTS idx_hg_toogo_volatility_config_active
  ON hg_toogo_volatility_config(is_active);

-- 插入默认“全局配置”（仅在不存在 symbol=NULL 的记录时插入）
INSERT INTO hg_toogo_volatility_config (
  symbol,
  high_volatility_threshold,
  low_volatility_threshold,
  trend_strength_threshold,
  d_threshold,
  range_volatility_threshold,
  delta_1m, delta_5m, delta_15m, delta_30m, delta_1h,
  weight_1m, weight_5m, weight_15m, weight_30m, weight_1h,
  is_active,
  created_at,
  updated_at
)
SELECT
  NULL,
  2.0,
  0.9,
  1.2,
  0.7,
  0,
  2.0, 2.0, 3.0, 3.0, 5.0,
  0.18, 0.25, 0.27, 0.20, 0.10,
  1,
  NOW(),
  NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM hg_toogo_volatility_config WHERE symbol IS NULL
);

COMMIT;

