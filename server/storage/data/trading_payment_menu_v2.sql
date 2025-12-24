-- =============================================
-- Trading & Payment System Menu Configuration
-- For HotGo v2.0
-- =============================================

SET NAMES utf8mb4;

-- 删除旧菜单（如果存在）
DELETE FROM `hg_admin_menu` WHERE `name` IN ('trading', 'payment') OR `title` IN ('量化交易', 'USDT管理');

-- =============================================
-- 1. Trading 量化交易菜单
-- =============================================

-- 1.1 量化交易 - 顶级菜单
INSERT INTO `hg_admin_menu` (`pid`, `level`, `tree`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`, `remark`, `status`, `created_at`, `updated_at`) 
VALUES 
(0, 1, '', '量化交易', 'trading', '/trading', 'WalletOutlined', 1, '/trading/robot', '', '', 'layout.base$view.blank', 1, '', 0, 0, '', 0, 0, 0, 30, '量化交易系统', 1, NOW(), NOW());

SET @trading_menu_id = LAST_INSERT_ID();

-- 1.2 API配置
INSERT INTO `hg_admin_menu` (`pid`, `level`, `tree`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`, `remark`, `status`, `created_at`, `updated_at`) 
VALUES 
(@trading_menu_id, 2, '', 'API配置', 'trading_api_config', '/trading/api-config', '', 2, '', '/trading/api-config/list', '列表', 'view.trading.api-config', 0, '', 0, 0, '', 1, 0, 0, 10, '交易所API接口配置管理', 1, NOW(), NOW());

SET @trading_api_config_menu_id = LAST_INSERT_ID();

-- 1.2.1 API配置子菜单
INSERT INTO `hg_admin_menu` (`pid`, `level`, `tree`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`, `remark`, `status`, `created_at`, `updated_at`) 
VALUES 
(@trading_api_config_menu_id, 3, '', '新增', 'trading_api_config_create', '', '', 3, '', '/trading/api-config/create', '新增', '', 0, '', 0, 0, '', 0, 0, 0, 1, '', 1, NOW(), NOW()),
(@trading_api_config_menu_id, 3, '', '编辑', 'trading_api_config_update', '', '', 3, '', '/trading/api-config/update', '编辑', '', 0, '', 0, 0, '', 0, 0, 0, 2, '', 1, NOW(), NOW()),
(@trading_api_config_menu_id, 3, '', '删除', 'trading_api_config_delete', '', '', 3, '', '/trading/api-config/delete', '删除', '', 0, '', 0, 0, '', 0, 0, 0, 3, '', 1, NOW(), NOW()),
(@trading_api_config_menu_id, 3, '', '查看', 'trading_api_config_view', '', '', 3, '', '/trading/api-config/view', '查看', '', 0, '', 0, 0, '', 0, 0, 0, 4, '', 1, NOW(), NOW()),
(@trading_api_config_menu_id, 3, '', '测试', 'trading_api_config_test', '', '', 3, '', '/trading/api-config/test', '测试', '', 0, '', 0, 0, '', 0, 0, 0, 5, '', 1, NOW(), NOW());

-- 1.3 代理配置
INSERT INTO `hg_admin_menu` (`pid`, `level`, `tree`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`, `remark`, `status`, `created_at`, `updated_at`) 
VALUES 
(@trading_menu_id, 2, '', '代理配置', 'trading_proxy_config', '/trading/proxy-config', '', 2, '', '/trading/proxy-config/get', '查看', 'view.trading.proxy-config', 0, '', 0, 0, '', 1, 0, 0, 20, 'SOCKS5代理配置', 1, NOW(), NOW());

SET @trading_proxy_config_menu_id = LAST_INSERT_ID();

-- 1.3.1 代理配置子菜单
INSERT INTO `hg_admin_menu` (`pid`, `level`, `tree`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`, `remark`, `status`, `created_at`, `updated_at`) 
VALUES 
(@trading_proxy_config_menu_id, 3, '', '保存', 'trading_proxy_config_save', '', '', 3, '', '/trading/proxy-config/save', '保存', '', 0, '', 0, 0, '', 0, 0, 0, 1, '', 1, NOW(), NOW()),
(@trading_proxy_config_menu_id, 3, '', '测试', 'trading_proxy_config_test', '', '', 3, '', '/trading/proxy-config/test', '测试', '', 0, '', 0, 0, '', 0, 0, 0, 2, '', 1, NOW(), NOW());

-- 1.4 机器人管理
INSERT INTO `hg_admin_menu` (`pid`, `level`, `tree`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`, `remark`, `status`, `created_at`, `updated_at`) 
VALUES 
(@trading_menu_id, 2, '', '机器人管理', 'trading_robot', '/trading/robot', '', 2, '', '/trading/robot/list', '列表', 'view.trading.robot', 0, '', 0, 0, '', 1, 0, 0, 30, '交易机器人管理', 1, NOW(), NOW());

SET @trading_robot_menu_id = LAST_INSERT_ID();

-- 1.4.1 机器人管理子菜单
INSERT INTO `hg_admin_menu` (`pid`, `level`, `tree`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`, `remark`, `status`, `created_at`, `updated_at`) 
VALUES 
(@trading_robot_menu_id, 3, '', '创建', 'trading_robot_create_page', '/trading/robot/create', '', 2, '', '/trading/robot/create', '创建', 'view.trading.robot.create', 0, 'trading_robot', 0, 0, '', 0, 1, 0, 1, '', 1, NOW(), NOW()),
(@trading_robot_menu_id, 3, '', '详情', 'trading_robot_detail_page', '/trading/robot/detail/:id', '', 2, '', '/trading/robot/view', '详情', 'view.trading.robot.detail', 0, 'trading_robot', 0, 0, '', 0, 1, 0, 2, '', 1, NOW(), NOW()),
(@trading_robot_menu_id, 3, '', '编辑', 'trading_robot_update', '', '', 3, '', '/trading/robot/update', '编辑', '', 0, '', 0, 0, '', 0, 0, 0, 3, '', 1, NOW(), NOW()),
(@trading_robot_menu_id, 3, '', '删除', 'trading_robot_delete', '', '', 3, '', '/trading/robot/delete', '删除', '', 0, '', 0, 0, '', 0, 0, 0, 4, '', 1, NOW(), NOW()),
(@trading_robot_menu_id, 3, '', '启动', 'trading_robot_start', '', '', 3, '', '/trading/robot/start', '启动', '', 0, '', 0, 0, '', 0, 0, 0, 5, '', 1, NOW(), NOW()),
(@trading_robot_menu_id, 3, '', '暂停', 'trading_robot_pause', '', '', 3, '', '/trading/robot/pause', '暂停', '', 0, '', 0, 0, '', 0, 0, 0, 6, '', 1, NOW(), NOW()),
(@trading_robot_menu_id, 3, '', '停止', 'trading_robot_stop', '', '', 3, '', '/trading/robot/stop', '停止', '', 0, '', 0, 0, '', 0, 0, 0, 7, '', 1, NOW(), NOW());

-- =============================================
-- 2. Payment USDT管理菜单
-- =============================================

-- 2.1 USDT管理 - 顶级菜单
INSERT INTO `hg_admin_menu` (`pid`, `level`, `tree`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`, `remark`, `status`, `created_at`, `updated_at`) 
VALUES 
(0, 1, '', 'USDT管理', 'payment', '/payment', 'DollarOutlined', 1, '/payment/balance', '', '', 'layout.base$view.blank', 1, '', 0, 0, '', 0, 0, 0, 40, 'USDT充值提现管理', 1, NOW(), NOW());

SET @payment_menu_id = LAST_INSERT_ID();

-- 2.2 我的余额
INSERT INTO `hg_admin_menu` (`pid`, `level`, `tree`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`, `remark`, `status`, `created_at`, `updated_at`) 
VALUES 
(@payment_menu_id, 2, '', '我的余额', 'payment_balance', '/payment/balance', '', 2, '', '/payment/balance/view', '查看', 'view.payment.balance', 0, '', 0, 0, '', 1, 0, 0, 10, '查看USDT余额', 1, NOW(), NOW());

SET @payment_balance_menu_id = LAST_INSERT_ID();

-- 2.2.1 余额子菜单
INSERT INTO `hg_admin_menu` (`pid`, `level`, `tree`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`, `remark`, `status`, `created_at`, `updated_at`) 
VALUES 
(@payment_balance_menu_id, 3, '', '流水', 'payment_balance_logs', '', '', 3, '', '/payment/balance/logs', '流水', '', 0, '', 0, 0, '', 0, 0, 0, 1, '', 1, NOW(), NOW());

-- 2.3 USDT充值
INSERT INTO `hg_admin_menu` (`pid`, `level`, `tree`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`, `remark`, `status`, `created_at`, `updated_at`) 
VALUES 
(@payment_menu_id, 2, '', 'USDT充值', 'payment_deposit', '/payment/deposit', '', 2, '', '/payment/deposit/list', '列表', 'view.payment.deposit', 0, '', 0, 0, '', 1, 0, 0, 20, 'USDT充值管理', 1, NOW(), NOW());

SET @payment_deposit_menu_id = LAST_INSERT_ID();

-- 2.3.1 充值子菜单
INSERT INTO `hg_admin_menu` (`pid`, `level`, `tree`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`, `remark`, `status`, `created_at`, `updated_at`) 
VALUES 
(@payment_deposit_menu_id, 3, '', '创建', 'payment_deposit_create', '', '', 3, '', '/payment/deposit/create', '创建', '', 0, '', 0, 0, '', 0, 0, 0, 1, '', 1, NOW(), NOW()),
(@payment_deposit_menu_id, 3, '', '详情', 'payment_deposit_view', '', '', 3, '', '/payment/deposit/view', '详情', '', 0, '', 0, 0, '', 0, 0, 0, 2, '', 1, NOW(), NOW()),
(@payment_deposit_menu_id, 3, '', '检查', 'payment_deposit_check', '', '', 3, '', '/payment/deposit/check', '检查', '', 0, '', 0, 0, '', 0, 0, 0, 3, '', 1, NOW(), NOW()),
(@payment_deposit_menu_id, 3, '', '取消', 'payment_deposit_cancel', '', '', 3, '', '/payment/deposit/cancel', '取消', '', 0, '', 0, 0, '', 0, 0, 0, 4, '', 1, NOW(), NOW());

-- 2.4 USDT提现
INSERT INTO `hg_admin_menu` (`pid`, `level`, `tree`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`, `remark`, `status`, `created_at`, `updated_at`) 
VALUES 
(@payment_menu_id, 2, '', 'USDT提现', 'payment_withdraw', '/payment/withdraw', '', 2, '', '/payment/withdraw/list', '列表', 'view.payment.withdraw', 0, '', 0, 0, '', 1, 0, 0, 30, 'USDT提现管理', 1, NOW(), NOW());

SET @payment_withdraw_menu_id = LAST_INSERT_ID();

-- 2.4.1 提现子菜单
INSERT INTO `hg_admin_menu` (`pid`, `level`, `tree`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`, `remark`, `status`, `created_at`, `updated_at`) 
VALUES 
(@payment_withdraw_menu_id, 3, '', '申请', 'payment_withdraw_apply', '', '', 3, '', '/payment/withdraw/apply', '申请', '', 0, '', 0, 0, '', 0, 0, 0, 1, '', 1, NOW(), NOW()),
(@payment_withdraw_menu_id, 3, '', '详情', 'payment_withdraw_view', '', '', 3, '', '/payment/withdraw/view', '详情', '', 0, '', 0, 0, '', 0, 0, 0, 2, '', 1, NOW(), NOW()),
(@payment_withdraw_menu_id, 3, '', '取消', 'payment_withdraw_cancel', '', '', 3, '', '/payment/withdraw/cancel', '取消', '', 0, '', 0, 0, '', 0, 0, 0, 3, '', 1, NOW(), NOW());

-- 2.5 提现审核
INSERT INTO `hg_admin_menu` (`pid`, `level`, `tree`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`, `remark`, `status`, `created_at`, `updated_at`) 
VALUES 
(@payment_menu_id, 2, '', '提现审核', 'payment_admin_withdraw_audit', '/payment/admin/withdraw-audit', '', 2, '', '/payment/withdraw/audit', '审核', 'view.payment.admin.withdraw-audit', 0, '', 0, 0, '', 1, 0, 0, 40, '管理员提现审核', 1, NOW(), NOW());

-- =============================================
-- 更新tree字段
-- =============================================
UPDATE `hg_admin_menu` SET `tree` = CONCAT('0-', `id`) WHERE `pid` = 0 AND `name` IN ('trading', 'payment');
UPDATE `hg_admin_menu` SET `tree` = CONCAT((SELECT `tree` FROM (SELECT * FROM `hg_admin_menu`) as tmp WHERE `id` = `pid`), '-', `id`) WHERE `pid` IN (@trading_menu_id, @payment_menu_id);
UPDATE `hg_admin_menu` SET `tree` = CONCAT((SELECT `tree` FROM (SELECT * FROM `hg_admin_menu`) as tmp WHERE `id` = `pid`), '-', `id`) WHERE `pid` IN (@trading_api_config_menu_id, @trading_proxy_config_menu_id, @trading_robot_menu_id);
UPDATE `hg_admin_menu` SET `tree` = CONCAT((SELECT `tree` FROM (SELECT * FROM `hg_admin_menu`) as tmp WHERE `id` = `pid`), '-', `id`) WHERE `pid` IN (@payment_balance_menu_id, @payment_deposit_menu_id, @payment_withdraw_menu_id);

-- =============================================
-- 完成
-- =============================================
SELECT '✅ 菜单配置完成！' as result;

