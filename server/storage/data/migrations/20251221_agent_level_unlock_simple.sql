-- Agent Level Unlock Migration
-- Add agent_status field
ALTER TABLE `hg_toogo_user` 
ADD COLUMN `agent_status` tinyint(1) NOT NULL DEFAULT 0 COMMENT 'Agent status: 0=not applied, 1=pending, 2=approved, 3=rejected' AFTER `is_agent`;

-- Add agent_unlock_level field
ALTER TABLE `hg_toogo_user` 
ADD COLUMN `agent_unlock_level` tinyint(1) NOT NULL DEFAULT 0 COMMENT 'Level unlock: 0=first level only, 1=unlimited levels' AFTER `agent_status`;

-- Add agent_apply_remark field
ALTER TABLE `hg_toogo_user` 
ADD COLUMN `agent_apply_remark` varchar(500) DEFAULT '' COMMENT 'Agent apply remark' AFTER `power_discount`;

-- Add agent_apply_at field
ALTER TABLE `hg_toogo_user` 
ADD COLUMN `agent_apply_at` datetime DEFAULT NULL COMMENT 'Agent apply time' AFTER `agent_apply_remark`;

-- Add agent_approved_at field
ALTER TABLE `hg_toogo_user` 
ADD COLUMN `agent_approved_at` datetime DEFAULT NULL COMMENT 'Agent approved time' AFTER `agent_apply_at`;

-- Add agent_approved_by field
ALTER TABLE `hg_toogo_user` 
ADD COLUMN `agent_approved_by` bigint(20) DEFAULT 0 COMMENT 'Approved by user ID' AFTER `agent_approved_at`;

-- Update existing agents to approved status
UPDATE `hg_toogo_user` SET `agent_status` = 2 WHERE `is_agent` = 1;

-- Add indexes
ALTER TABLE `hg_toogo_user` ADD INDEX `idx_agent_status` (`agent_status`);
ALTER TABLE `hg_toogo_user` ADD INDEX `idx_agent_unlock_level` (`agent_unlock_level`);

