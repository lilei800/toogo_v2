-- Add profit lock switch: when enabled and take-profit retreat started, block new auto orders until position closed.
-- MySQL version
ALTER TABLE `hg_trading_robot`
  ADD COLUMN `profit_lock_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '锁定盈利开关：0=关闭,1=开启（止盈启动后禁止自动开新仓）' AFTER `auto_close_enabled`;


