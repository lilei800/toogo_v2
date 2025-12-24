-- 代理商层级解锁功能迁移脚本
-- 作者: Toogo Team
-- 日期: 2024-12-21
-- 说明: 
--   1. 添加代理商申请状态字段 (agent_status)
--   2. 添加层级解锁字段 (agent_unlock_level)
--   3. 添加申请相关时间和备注字段

-- 添加代理商申请状态字段
ALTER TABLE `hg_toogo_user` 
ADD COLUMN `agent_status` tinyint(1) NOT NULL DEFAULT 0 COMMENT '代理商状态: 0=未申请, 1=待审批, 2=已通过, 3=已拒绝' AFTER `is_agent`;

-- 添加层级解锁字段
ALTER TABLE `hg_toogo_user` 
ADD COLUMN `agent_unlock_level` tinyint(1) NOT NULL DEFAULT 0 COMMENT '层级解锁: 0=仅一级佣金, 1=无限级佣金' AFTER `agent_status`;

-- 添加代理商申请备注
ALTER TABLE `hg_toogo_user` 
ADD COLUMN `agent_apply_remark` varchar(500) DEFAULT '' COMMENT '代理商申请备注' AFTER `power_discount`;

-- 添加代理商申请时间
ALTER TABLE `hg_toogo_user` 
ADD COLUMN `agent_apply_at` datetime DEFAULT NULL COMMENT '代理商申请时间' AFTER `agent_apply_remark`;

-- 添加代理商审批时间
ALTER TABLE `hg_toogo_user` 
ADD COLUMN `agent_approved_at` datetime DEFAULT NULL COMMENT '代理商审批时间' AFTER `agent_apply_at`;

-- 添加审批人ID
ALTER TABLE `hg_toogo_user` 
ADD COLUMN `agent_approved_by` bigint(20) DEFAULT 0 COMMENT '审批人ID' AFTER `agent_approved_at`;

-- 将已有的代理商状态同步为"已通过"
UPDATE `hg_toogo_user` SET `agent_status` = 2 WHERE `is_agent` = 1;

-- 索引
ALTER TABLE `hg_toogo_user` ADD INDEX `idx_agent_status` (`agent_status`);
ALTER TABLE `hg_toogo_user` ADD INDEX `idx_agent_unlock_level` (`agent_unlock_level`);

