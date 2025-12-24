-- 清理并重新导入Toogo菜单

-- 删除已存在的Toogo菜单
DELETE FROM `hg_admin_menu` WHERE `id` >= 2000 AND `id` < 2200;

-- 删除已存在的Toogo定时任务
DELETE FROM `hg_sys_cron` WHERE `name` LIKE 'toogo_%';

-- 重新插入菜单
-- 一级菜单：Toogo量化
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2000, 0, 'Toogo量化', 'ToogoRoot', '/toogo', 'HardwareChipOutline', 1, '/toogo/dashboard', '', '', 'LAYOUT', 1, '', 0, 0, '', 0, 0, 0, 1, 'tr_2000', 50, 'Toogo量化交易', 1, NOW(), NOW()),
(2001, 2000, '控制台', 'ToogoDashboard', '/toogo/dashboard', 'HomeOutline', 1, '', '/toogo/dashboard', '控制台', '/toogo/dashboard/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2000_2001', 1, '控制台', 1, NOW(), NOW()),
(2002, 2000, '订阅套餐', 'ToogoSubscription', '/toogo/subscription', 'CardOutline', 1, '', '/toogo/subscription', '订阅套餐', '/toogo/subscription/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2000_2002', 2, '订阅套餐', 1, NOW(), NOW()),
(2003, 2000, '我的机器人', 'ToogoRobot', '/toogo/robot', 'HardwareChipOutline', 1, '', '/toogo/robot', '机器人', '/toogo/robot/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2000_2003', 3, '机器人管理', 1, NOW(), NOW()),
(2004, 2003, '创建机器人', 'ToogoRobotCreate', '/toogo/robot/create', '', 1, '', '/toogo/robot/create', '创建', '/toogo/robot/create', 0, '/toogo/robot', 0, 0, '', 1, 1, 0, 3, 'tr_2000_2003_2004', 1, '创建', 1, NOW(), NOW()),
(2005, 2000, '我的团队', 'ToogoTeam', '/toogo/team', 'PeopleOutline', 1, '', '/toogo/team', '团队', '/toogo/team/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2000_2005', 4, '团队管理', 1, NOW(), NOW()),
(2006, 2000, '佣金明细', 'ToogoCommission', '/toogo/commission', 'TrendingUpOutline', 1, '', '/toogo/commission', '佣金', '/toogo/commission/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2000_2006', 5, '佣金管理', 1, NOW(), NOW());

-- 后台管理菜单
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2100, 0, 'Toogo管理', 'ToogoAdmin', '/toogo-admin', 'SettingsOutline', 1, '/toogo-admin/user', '', '', 'LAYOUT', 1, '', 0, 0, '', 0, 0, 0, 1, 'tr_2100', 51, '后台管理', 1, NOW(), NOW()),
(2101, 2100, '用户管理', 'ToogoAdminUser', '/toogo-admin/user', 'PersonOutline', 1, '', '/toogo-admin/user', '用户', '/toogo/admin/user/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2101', 1, '用户管理', 1, NOW(), NOW()),
(2102, 2100, '套餐管理', 'ToogoAdminPlan', '/toogo-admin/plan', 'CardOutline', 1, '', '/toogo-admin/plan', '套餐', '/toogo/admin/plan/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2102', 2, '套餐管理', 1, NOW(), NOW()),
(2103, 2100, 'VIP等级', 'ToogoAdminVipLevel', '/toogo-admin/vip-level', 'StarOutline', 1, '', '/toogo-admin/vip-level', 'VIP', '/toogo/admin/vip-level/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2103', 3, 'VIP等级', 1, NOW(), NOW()),
(2104, 2100, '代理等级', 'ToogoAdminAgentLevel', '/toogo-admin/agent-level', 'PeopleOutline', 1, '', '/toogo-admin/agent-level', '代理', '/toogo/admin/agent-level/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2104', 4, '代理等级', 1, NOW(), NOW()),
(2105, 2100, '策略模板', 'ToogoAdminStrategy', '/toogo-admin/strategy', 'BulbOutline', 1, '', '/toogo-admin/strategy', '策略', '/toogo/admin/strategy/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2105', 5, '策略模板', 1, NOW(), NOW()),
(2106, 2100, '提现审核', 'ToogoAdminWithdraw', '/toogo-admin/withdraw', 'WalletOutline', 1, '', '/toogo-admin/withdraw', '提现', '/toogo/admin/withdraw/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2106', 6, '提现审核', 1, NOW(), NOW()),
(2107, 2100, '充值记录', 'ToogoAdminDeposit', '/toogo-admin/deposit', 'CashOutline', 1, '', '/toogo-admin/deposit', '充值', '/toogo/admin/deposit/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2107', 7, '充值记录', 1, NOW(), NOW()),
(2108, 2100, '订阅记录', 'ToogoAdminSubscription', '/toogo-admin/subscription', 'DocumentTextOutline', 1, '', '/toogo-admin/subscription', '订阅', '/toogo/admin/subscription/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2108', 8, '订阅记录', 1, NOW(), NOW()),
(2109, 2100, '佣金记录', 'ToogoAdminCommissionLog', '/toogo-admin/commission-log', 'TrendingUpOutline', 1, '', '/toogo-admin/commission-log', '佣金', '/toogo/admin/commission-log/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2109', 9, '佣金记录', 1, NOW(), NOW()),
(2110, 2100, '系统配置', 'ToogoAdminConfig', '/toogo-admin/config', 'SettingsOutline', 1, '', '/toogo-admin/config', '配置', '/toogo/admin/config/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2110', 10, '系统配置', 1, NOW(), NOW());

-- 定时任务
INSERT INTO `hg_sys_cron` (`group_id`, `title`, `name`, `params`, `pattern`, `policy`, `count`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(1, '机器人引擎', 'toogo_robot_engine', '', '*/10 * * * * *', 1, 0, 100, '每10秒执行一次', 1, NOW(), NOW()),
(1, '订阅检查', 'toogo_subscription_check', '', '0 0 * * * *', 1, 0, 101, '每小时执行一次', 1, NOW(), NOW());

SELECT 'Toogo菜单导入完成' AS result;

