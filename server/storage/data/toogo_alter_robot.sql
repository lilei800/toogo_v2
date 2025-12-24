-- 为hg_trading_robot表添加Toogo相关字段
-- 如果字段已存在会报错，忽略即可

ALTER TABLE `hg_trading_robot` ADD COLUMN `plan_id` bigint(20) DEFAULT NULL COMMENT '套餐ID';
ALTER TABLE `hg_trading_robot` ADD COLUMN `consumed_power` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '已消耗算力';
ALTER TABLE `hg_trading_robot` ADD COLUMN `estimated_power` decimal(20,8) NOT NULL DEFAULT '0.00000000' COMMENT '预计消耗算力';
ALTER TABLE `hg_trading_robot` ADD COLUMN `schedule_start` datetime DEFAULT NULL COMMENT '定时开启时间';
ALTER TABLE `hg_trading_robot` ADD COLUMN `schedule_stop` datetime DEFAULT NULL COMMENT '定时关闭时间';
ALTER TABLE `hg_trading_robot` ADD COLUMN `auto_analyze_market` tinyint(1) DEFAULT '0' COMMENT '全自动分析行情';
ALTER TABLE `hg_trading_robot` ADD COLUMN `auto_signal_enabled` tinyint(1) DEFAULT '1' COMMENT '全自动方向预警信号';
ALTER TABLE `hg_trading_robot` ADD COLUMN `trade_type` varchar(20) DEFAULT 'perpetual' COMMENT '交易类型';

