-- =====================================================
-- 预警日志菜单配置
-- =====================================================

-- 获取父菜单ID (Toogo量化交易)
SET @parent_id = (SELECT id FROM hg_admin_menu WHERE `path` = '/toogo' LIMIT 1);

-- 如果没找到，使用默认值
SET @parent_id = IFNULL(@parent_id, 0);

-- 1. 预警日志父菜单
INSERT INTO `hg_admin_menu` (`pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`)
VALUES 
(@parent_id, '预警日志', 'AlertLogs', '/toogo/alert', 'AlertOutline', 1, '/toogo/alert/engine-status', '', '', 'LAYOUT', 1, '', 0, 0, '', 1, 0, 0, 2, '', 95, '预警日志管理', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE `title`='预警日志', `icon`='AlertOutline';

-- 获取预警日志父菜单ID
SET @alert_parent_id = (SELECT id FROM hg_admin_menu WHERE `path` = '/toogo/alert' ORDER BY id DESC LIMIT 1);

-- 2. 引擎状态
INSERT INTO `hg_admin_menu` (`pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`)
VALUES 
(@alert_parent_id, '引擎状态', 'EngineStatus', '/toogo/alert/engine-status', 'Dashboard', 2, '', '/trading/alert/engine/status', '引擎状态', '/toogo/alert/engine-status', 0, '', 0, 0, '', 1, 0, 0, 3, '', 10, '查看引擎运行状态', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE `title`='引擎状态';

-- 3. 市场状态预警
INSERT INTO `hg_admin_menu` (`pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`)
VALUES 
(@alert_parent_id, '市场状态预警', 'MarketStateLog', '/toogo/alert/market-state', 'TrendingUp', 2, '', '/trading/alert/marketState/list', '市场状态预警', '/toogo/alert/market-state', 0, '', 0, 0, '', 1, 0, 0, 3, '', 20, '市场状态变化预警日志', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE `title`='市场状态预警';

-- 4. 风险偏好预警
INSERT INTO `hg_admin_menu` (`pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`)
VALUES 
(@alert_parent_id, '风险偏好预警', 'RiskPreferenceLog', '/toogo/alert/risk-preference', 'Shield', 2, '', '/trading/alert/riskPreference/list', '风险偏好预警', '/toogo/alert/risk-preference', 0, '', 0, 0, '', 1, 0, 0, 3, '', 30, '风险偏好变化预警日志', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE `title`='风险偏好预警';

-- 5. 方向预警
INSERT INTO `hg_admin_menu` (`pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`, `permission_name`, `component`, `always_show`, `active_menu`, `is_root`, `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `level`, `tree`, `sort`, `remark`, `status`, `created_at`, `updated_at`)
VALUES 
(@alert_parent_id, '方向预警', 'DirectionLog', '/toogo/alert/direction', 'Compass', 2, '', '/trading/alert/direction/list', '方向预警', '/toogo/alert/direction', 0, '', 0, 0, '', 1, 0, 0, 3, '', 40, '交易方向变化预警日志', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE `title`='方向预警';

-- 6. 为超级管理员角色添加菜单权限
INSERT IGNORE INTO `hg_admin_role_menu` (`role_id`, `menu_id`)
SELECT 1, id FROM `hg_admin_menu` WHERE `path` LIKE '/toogo/alert%';

SELECT '预警日志菜单创建完成!' AS message;

