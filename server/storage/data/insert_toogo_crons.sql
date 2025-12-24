-- =============================================
-- Toogo定时任务注册
-- =============================================

-- 先检查是否存在sys_cron_group
INSERT IGNORE INTO `hg_sys_cron_group` (`id`, `name`, `remark`, `status`, `created_at`, `updated_at`)
VALUES (10, 'toogo', 'Toogo量化交易定时任务', 1, NOW(), NOW());

-- 插入定时任务
INSERT INTO `hg_sys_cron` (`group_id`, `title`, `name`, `params`, `pattern`, `policy`, `count`, `sort`, `remark`, `status`, `created_at`, `updated_at`)
VALUES 
-- 机器人交易引擎 - 每5秒执行一次
(10, 'Robot Engine', 'ToogoEngine', '', '@every 5s', 1, 0, 10, 'Trading robot engine', 1, NOW(), NOW()),

-- 订阅到期检查 - 每小时检查一次
(10, 'Sub Check', 'ToogoSubscriptionChecker', '', '@every 1h', 1, 0, 20, 'Subscription check', 1, NOW(), NOW()),

-- 算力结算 - 每分钟检查
(10, 'Power Settle', 'ToogoPowerSettlement', '', '@every 1m', 1, 0, 30, 'Power settlement', 1, NOW(), NOW()),

-- 邀请码清理 - 每小时清理
(10, 'Code Cleanup', 'ToogoInviteCodeCleanup', '', '@every 1h', 1, 0, 40, 'Invite code cleanup', 1, NOW(), NOW()),

-- VIP等级检查 - 每天凌晨2点 (使用@every语法避免格式问题)
(10, 'VIP Check', 'ToogoVipLevelCheck', '', '@every 24h', 1, 0, 50, 'VIP level check', 1, NOW(), NOW())

ON DUPLICATE KEY UPDATE 
    `pattern` = VALUES(`pattern`),
    `remark` = VALUES(`remark`),
    `updated_at` = NOW();

SELECT '✅ Toogo定时任务已注册！' as result;
SELECT id, title, name, pattern, status FROM `hg_sys_cron` WHERE `group_id` = 10;

