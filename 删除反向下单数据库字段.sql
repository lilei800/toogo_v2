-- 删除反向下单相关的数据库字段
-- 执行前请备份数据库！

-- 1. 删除 hg_trading_robot 表中的 enable_reverse_order 字段
ALTER TABLE `hg_trading_robot` DROP COLUMN IF EXISTS `enable_reverse_order`;

-- 2. 删除 hg_trading_strategy_template 表中的反向下单相关字段（如果存在）
ALTER TABLE `hg_trading_strategy_template` DROP COLUMN IF EXISTS `reverse_loss_retreat`;
ALTER TABLE `hg_trading_strategy_template` DROP COLUMN IF EXISTS `reverse_profit_retreat`;

-- 3. 检查并清理 config_json 中的反向下单配置（可选，如果需要保留历史数据可以跳过）
-- UPDATE `hg_trading_strategy_template` 
-- SET `config_json` = JSON_REMOVE(`config_json`, '$.reverse_loss_retreat', '$.reverse_profit_retreat')
-- WHERE JSON_EXTRACT(`config_json`, '$.reverse_loss_retreat') IS NOT NULL 
--    OR JSON_EXTRACT(`config_json`, '$.reverse_profit_retreat') IS NOT NULL;

