-- =============================================
-- 修复菜单组件路径
-- =============================================

-- Trading 菜单组件路径修复
UPDATE `hg_admin_menu` SET `component` = '/trading/api-config/index' WHERE `name` = 'trading_api_config';
UPDATE `hg_admin_menu` SET `component` = '/trading/proxy-config/index' WHERE `name` = 'trading_proxy_config';
UPDATE `hg_admin_menu` SET `component` = '/trading/robot/index' WHERE `name` = 'trading_robot';
UPDATE `hg_admin_menu` SET `component` = '/trading/robot/create' WHERE `name` = 'trading_robot_create_page';
UPDATE `hg_admin_menu` SET `component` = '/trading/robot/detail' WHERE `name` = 'trading_robot_detail_page';

-- Payment 菜单组件路径修复
UPDATE `hg_admin_menu` SET `component` = '/payment/balance/index' WHERE `name` = 'payment_balance';
UPDATE `hg_admin_menu` SET `component` = '/payment/deposit/index' WHERE `name` = 'payment_deposit';
UPDATE `hg_admin_menu` SET `component` = '/payment/withdraw/index' WHERE `name` = 'payment_withdraw';
UPDATE `hg_admin_menu` SET `component` = '/payment/admin/withdraw-audit' WHERE `name` = 'payment_admin_withdraw_audit';

-- 顶级菜单使用LAYOUT
UPDATE `hg_admin_menu` SET `component` = 'LAYOUT' WHERE `name` IN ('trading', 'payment');

SELECT '✅ 组件路径修复完成！' as result;
SELECT '请刷新浏览器后查看效果' as tip;

