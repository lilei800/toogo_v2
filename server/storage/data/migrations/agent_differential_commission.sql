-- 代理商级差制佣金系统升级
-- 1. 给 toogo_user 添加佣金比例字段

ALTER TABLE `hg_toogo_user` 
ADD COLUMN `subscribe_rate` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT '订阅返佣比例(%)' AFTER `agent_level`,
ADD COLUMN `power_rate` decimal(5,2) NOT NULL DEFAULT '0.00' COMMENT '算力消耗佣金比例(%)' AFTER `subscribe_rate`;

-- 2. 不再需要代理商等级表的分级佣金字段，但保留表用于其他配置
-- 可以选择删除旧的代理商等级表或保留

-- 3. 更新佣金记录表的备注信息（可选）
-- ALTER TABLE `hg_toogo_commission_log` MODIFY COLUMN `level` int DEFAULT 0 COMMENT '级差层级(从消费者往上数)';

