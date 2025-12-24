-- =========================================
-- Toogo.Ai 土狗 - 菜单数据
-- 量化交易系统菜单配置
-- =========================================

-- 注意：请根据实际的menu id进行调整
-- 假设从 ID 2000 开始

-- 一级菜单：Toogo量化
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2000, 0, 'Toogo量化', 'ToogoRoot', '/toogo', 'HardwareChipOutline', 1, '/toogo/dashboard', '', '', 'LAYOUT', 1, '', 0, 0, '', 0, 0, 0, 1, 'tr_2000', 50, 'Toogo量化交易系统', 1, NOW(), NOW());

-- 二级菜单：控制台
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2001, 2000, '控制台', 'ToogoDashboard', '/toogo/dashboard', 'HomeOutline', 1, '', '/toogo/dashboard', '控制台', '/toogo/dashboard/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2000_2001', 1, 'Toogo控制台', 1, NOW(), NOW());

-- 二级菜单：订阅套餐
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2002, 2000, '订阅套餐', 'ToogoSubscription', '/toogo/subscription', 'CardOutline', 1, '', '/toogo/subscription', '订阅套餐', '/toogo/subscription/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2000_2002', 2, '套餐订阅管理', 1, NOW(), NOW());

-- 二级菜单：我的机器人
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2003, 2000, '我的机器人', 'ToogoRobot', '/toogo/robot', 'HardwareChipOutline', 1, '', '/toogo/robot', '我的机器人', '/toogo/robot/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2000_2003', 3, '云机器人管理', 1, NOW(), NOW());

-- 三级菜单：创建机器人（隐藏）
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2004, 2003, '创建机器人', 'ToogoRobotCreate', '/toogo/robot/create', '', 1, '', '/toogo/robot/create', '创建机器人', '/toogo/robot/create', 0, '/toogo/robot', 0, 0, '', 1, 1, 0, 3, 'tr_2000_2003_2004', 1, '创建机器人页面', 1, NOW(), NOW());

-- 二级菜单：我的团队
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2005, 2000, '我的团队', 'ToogoTeam', '/toogo/team', 'PeopleOutline', 1, '', '/toogo/team', '我的团队', '/toogo/team/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2000_2005', 4, '团队推广管理', 1, NOW(), NOW());

-- 二级菜单：佣金明细
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2006, 2000, '佣金明细', 'ToogoCommission', '/toogo/commission', 'TrendingUpOutline', 1, '', '/toogo/commission', '佣金明细', '/toogo/commission/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2000_2006', 5, '佣金收入管理', 1, NOW(), NOW());

-- ===========================================
-- 后台管理菜单
-- ===========================================

-- 一级菜单：Toogo管理
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2100, 0, 'Toogo管理', 'ToogoAdmin', '/toogo-admin', 'SettingsOutline', 1, '/toogo-admin/user', '', '', 'LAYOUT', 1, '', 0, 0, '', 0, 0, 0, 1, 'tr_2100', 51, 'Toogo后台管理', 1, NOW(), NOW());

-- 二级菜单：用户管理
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2101, 2100, '用户管理', 'ToogoAdminUser', '/toogo-admin/user', 'PersonOutline', 1, '', '/toogo-admin/user', '用户管理', '/toogo/admin/user/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2101', 1, 'Toogo用户管理', 1, NOW(), NOW());

-- 二级菜单：套餐管理
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2102, 2100, '套餐管理', 'ToogoAdminPlan', '/toogo-admin/plan', 'CardOutline', 1, '', '/toogo-admin/plan', '套餐管理', '/toogo/admin/plan/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2102', 2, '订阅套餐管理', 1, NOW(), NOW());

-- 二级菜单：VIP等级
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2103, 2100, 'VIP等级', 'ToogoAdminVipLevel', '/toogo-admin/vip-level', 'StarOutline', 1, '', '/toogo-admin/vip-level', 'VIP等级', '/toogo/admin/vip-level/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2103', 3, 'VIP等级配置', 1, NOW(), NOW());

-- 二级菜单：代理商等级
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2104, 2100, '代理商等级', 'ToogoAdminAgentLevel', '/toogo-admin/agent-level', 'PeopleOutline', 1, '', '/toogo-admin/agent-level', '代理商等级', '/toogo/admin/agent-level/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2104', 4, '代理商等级配置', 1, NOW(), NOW());

-- 二级菜单：策略模板
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2105, 2100, '策略模板', 'ToogoAdminStrategy', '/toogo-admin/strategy', 'BulbOutline', 1, '', '/toogo-admin/strategy', '策略模板', '/toogo/admin/strategy/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2105', 5, '策略模板管理', 1, NOW(), NOW());

-- 二级菜单：提现审核
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2106, 2100, '提现审核', 'ToogoAdminWithdraw', '/toogo-admin/withdraw', 'WalletOutline', 1, '', '/toogo-admin/withdraw', '提现审核', '/toogo/admin/withdraw/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2106', 6, '提现申请审核', 1, NOW(), NOW());

-- 二级菜单：充值记录
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2107, 2100, '充值记录', 'ToogoAdminDeposit', '/toogo-admin/deposit', 'CashOutline', 1, '', '/toogo-admin/deposit', '充值记录', '/toogo/admin/deposit/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2107', 7, '充值记录管理', 1, NOW(), NOW());

-- 二级菜单：订阅记录
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2108, 2100, '订阅记录', 'ToogoAdminSubscription', '/toogo-admin/subscription', 'DocumentTextOutline', 1, '', '/toogo-admin/subscription', '订阅记录', '/toogo/admin/subscription/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2108', 8, '订阅记录管理', 1, NOW(), NOW());

-- 二级菜单：佣金记录
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2109, 2100, '佣金记录', 'ToogoAdminCommissionLog', '/toogo-admin/commission-log', 'TrendingUpOutline', 1, '', '/toogo-admin/commission-log', '佣金记录', '/toogo/admin/commission-log/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2109', 9, '佣金记录管理', 1, NOW(), NOW());

-- 二级菜单：系统配置
INSERT INTO `hg_admin_menu` (`id`, `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(2110, 2100, '系统配置', 'ToogoAdminConfig', '/toogo-admin/config', 'SettingsOutline', 1, '', '/toogo-admin/config', '系统配置', '/toogo/admin/config/index', 0, '', 0, 0, '', 1, 0, 0, 2, 'tr_2100_2110', 10, 'Toogo系统配置', 1, NOW(), NOW());

-- ===========================================
-- 定时任务配置
-- ===========================================
INSERT INTO `hg_sys_cron` (`group_id`, `title`, `name`, `params`, `pattern`, `policy`, `count`, `sort`, `remark`, `status`, `created_at`, `updated_at`) VALUES
(1, 'Toogo机器人引擎', 'toogo_robot_engine', '', '*/10 * * * * *', 1, 0, 100, '每10秒执行一次，检查并执行所有运行中的机器人交易逻辑', 1, NOW(), NOW()),
(1, 'Toogo订阅过期检查', 'toogo_subscription_check', '', '0 0 * * * *', 1, 0, 101, '每小时执行一次，检查并处理过期的订阅', 1, NOW(), NOW());
