-- Toogo菜单导入 v2
SET NAMES utf8mb4;

-- 清理旧数据
DELETE FROM `hg_admin_menu` WHERE `id` >= 2000 AND `id` < 2200;
DELETE FROM `hg_sys_cron` WHERE `name` LIKE 'toogo_%';

-- 用户端菜单
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2000, 0, 'Toogo', 'ToogoRoot', '/toogo', 'HardwareChipOutline', 1, '/toogo/dashboard', '', '', 'LAYOUT', 1, '', 0, 0, '', 0, 0, 0, 1, 'tr_2000', 50, '', 1, NOW(), NOW());

INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `component`, `level`, `tree`, `sort`, `status`, `created_at`, `updated_at`) VALUES
(2001, 2000, 'Dashboard', 'ToogoDashboard', '/toogo/dashboard', 'HomeOutline', 1, '/toogo/dashboard/index', 2, 'tr_2000_2001', 1, 1, NOW(), NOW()),
(2002, 2000, 'Plan', 'ToogoSubscription', '/toogo/subscription', 'CardOutline', 1, '/toogo/subscription/index', 2, 'tr_2000_2002', 2, 1, NOW(), NOW()),
(2003, 2000, 'Robot', 'ToogoRobot', '/toogo/robot', 'HardwareChipOutline', 1, '/toogo/robot/index', 2, 'tr_2000_2003', 3, 1, NOW(), NOW()),
(2004, 2003, 'Create', 'ToogoRobotCreate', '/toogo/robot/create', '', 1, '/toogo/robot/create', 3, 'tr_2000_2003_2004', 1, 1, NOW(), NOW()),
(2005, 2000, 'Team', 'ToogoTeam', '/toogo/team', 'PeopleOutline', 1, '/toogo/team/index', 2, 'tr_2000_2005', 4, 1, NOW(), NOW()),
(2006, 2000, 'Commission', 'ToogoCommission', '/toogo/commission', 'TrendingUpOutline', 1, '/toogo/commission/index', 2, 'tr_2000_2006', 5, 1, NOW(), NOW());

-- 管理后台菜单
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2100, 0, 'ToogoAdmin', 'ToogoAdmin', '/toogo-admin', 'SettingsOutline', 1, '/toogo-admin/user', '', '', 'LAYOUT', 1, '', 0, 0, '', 0, 0, 0, 1, 'tr_2100', 51, '', 1, NOW(), NOW());

INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `component`, `level`, `tree`, `sort`, `status`, `created_at`, `updated_at`) VALUES
(2101, 2100, 'User', 'ToogoAdminUser', '/toogo-admin/user', 'PersonOutline', 1, '/toogo/admin/user/index', 2, 'tr_2100_2101', 1, 1, NOW(), NOW()),
(2102, 2100, 'Plan', 'ToogoAdminPlan', '/toogo-admin/plan', 'CardOutline', 1, '/toogo/admin/plan/index', 2, 'tr_2100_2102', 2, 1, NOW(), NOW()),
(2103, 2100, 'VIP', 'ToogoAdminVipLevel', '/toogo-admin/vip-level', 'StarOutline', 1, '/toogo/admin/vip-level/index', 2, 'tr_2100_2103', 3, 1, NOW(), NOW()),
(2104, 2100, 'Agent', 'ToogoAdminAgentLevel', '/toogo-admin/agent-level', 'PeopleOutline', 1, '/toogo/admin/agent-level/index', 2, 'tr_2100_2104', 4, 1, NOW(), NOW()),
(2105, 2100, 'Strategy', 'ToogoAdminStrategy', '/toogo-admin/strategy', 'BulbOutline', 1, '/toogo/admin/strategy/index', 2, 'tr_2100_2105', 5, 1, NOW(), NOW()),
(2106, 2100, 'Withdraw', 'ToogoAdminWithdraw', '/toogo-admin/withdraw', 'WalletOutline', 1, '/toogo/admin/withdraw/index', 2, 'tr_2100_2106', 6, 1, NOW(), NOW()),
(2107, 2100, 'Deposit', 'ToogoAdminDeposit', '/toogo-admin/deposit', 'CashOutline', 1, '/toogo/admin/deposit/index', 2, 'tr_2100_2107', 7, 1, NOW(), NOW()),
(2108, 2100, 'Subscription', 'ToogoAdminSubscription', '/toogo-admin/subscription', 'DocumentTextOutline', 1, '/toogo/admin/subscription/index', 2, 'tr_2100_2108', 8, 1, NOW(), NOW()),
(2109, 2100, 'CommLog', 'ToogoAdminCommissionLog', '/toogo-admin/commission-log', 'TrendingUpOutline', 1, '/toogo/admin/commission-log/index', 2, 'tr_2100_2109', 9, 1, NOW(), NOW()),
(2110, 2100, 'Config', 'ToogoAdminConfig', '/toogo-admin/config', 'SettingsOutline', 1, '/toogo/admin/config/index', 2, 'tr_2100_2110', 10, 1, NOW(), NOW());

-- 定时任务
INSERT INTO `hg_sys_cron` (`group_id`, `title`, `name`, `params`, `pattern`, `policy`, `count`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(1, 'RobotEngine', 'toogo_robot_engine', '', '*/10 * * * * *', 1, 0, 100, 'Run every 10s', 1, NOW(), NOW()),
(1, 'SubCheck', 'toogo_subscription_check', '', '0 0 * * * *', 1, 0, 101, 'Run every hour', 1, NOW(), NOW());

SELECT 'Done' AS result;

