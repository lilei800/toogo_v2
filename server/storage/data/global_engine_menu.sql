-- =====================================================
-- 全局引擎菜单配置
-- 在量化管理中添加全局引擎监控页面
-- =====================================================

-- 查找量化管理的父菜单ID
-- SELECT id, title, path FROM hg_admin_menu WHERE path = '/toogo' OR title LIKE '%量化%';

-- 假设量化管理的菜单已存在，获取其ID
SET @parent_id = (SELECT id FROM hg_admin_menu WHERE path = '/toogo/admin' LIMIT 1);

-- 如果没有找到，则使用0作为顶级菜单
SET @parent_id = IFNULL(@parent_id, 0);

-- 删除已存在的全局引擎菜单（如果有）
DELETE FROM `hg_admin_menu` WHERE `path` = '/toogo/admin/global-engine';

-- 创建全局引擎菜单
INSERT INTO `hg_admin_menu` 
(`pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`)
VALUES 
(@parent_id, '全局引擎', 'GlobalEngine', '/toogo/admin/global-engine', 'DashboardOutlined', 2, '', '', '', '/toogo/admin/global-engine/index', 0, '', 0, 0, '', 1, 0, 0, 2, '', 1, '全局引擎监控中心 - 管理行情数据服务、市场分析引擎、方向信号服务、机器人任务管理器等全局服务', 1, NOW(), NOW());

-- 完成提示
SELECT '全局引擎菜单创建完成!' AS message, @parent_id AS parent_id;

