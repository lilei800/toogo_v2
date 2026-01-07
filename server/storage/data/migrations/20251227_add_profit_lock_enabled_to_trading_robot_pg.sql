-- Add profit lock switch: when enabled and take-profit retreat started, block new auto orders until position closed.
-- PostgreSQL version
ALTER TABLE hg_trading_robot
  ADD COLUMN IF NOT EXISTS profit_lock_enabled SMALLINT NOT NULL DEFAULT 1;

COMMENT ON COLUMN hg_trading_robot.profit_lock_enabled IS '锁定盈利开关：0=关闭,1=开启（止盈启动后禁止自动开新仓）';


