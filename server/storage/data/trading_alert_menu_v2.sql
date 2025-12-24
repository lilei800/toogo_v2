-- =====================================================
-- 预警日志菜单配置 V2
-- 直接使用固定的菜单ID
-- =====================================================

-- 删除旧菜单（如果存在）
DELETE FROM `hg_admin_menu` WHERE `path` LIKE '/toogo/alert%';

-- 获取toogo父菜单ID
-- SELECT id FROM hg_admin_menu WHERE path = '/toogo';
-- 假设toogo父菜单ID为特定值，如果不存在则创建为顶级菜单

-- 1. 预警日志父菜单 (作为顶级菜单)
INSERT INTO `hg_admin_menu` (`pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`)
VALUES 
(0, '预警日志', 'AlertLogs', '/toogo/alert', 'AlertOutline', 1, '/toogo/alert/engine-status', '', '', 'LAYOUT', 1, '', 0, 0, '', 1, 0, 0, 1, '', 95, '预警日志管理', 1, NOW(), NOW());

-- 完成提示
SELECT '预警日志父菜单创建完成!' AS message;

