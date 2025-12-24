-- ============================================================
-- 添加策略管理菜单（策略模板 + 策略列表）
-- ============================================================

-- 查找Toogo父菜单ID
SET @parent_id = (SELECT id FROM hg_admin_menu WHERE path = '/toogo' OR title LIKE '%Toogo%' OR title LIKE '%量化%' LIMIT 1);

-- 如果找不到父菜单，使用默认值
SET @parent_id = IFNULL(@parent_id, 0);

-- 删除旧的策略菜单
DELETE FROM hg_admin_menu WHERE path LIKE '/toogo/strategy%';

-- 添加策略管理父菜单
INSERT INTO hg_admin_menu (pid, level, tree, title, name, path, component, redirect, icon, permissions, type, sort, hidden, status, remark, created_at, updated_at)
VALUES (@parent_id, 2, '', '策略管理', 'ToogoStrategy', '/toogo/strategy', 'LAYOUT', '/toogo/strategy/group', 'SettingOutlined', '', 1, 30, 0, 1, '策略管理', NOW(), NOW());

SET @strategy_parent_id = LAST_INSERT_ID();

-- 添加策略模板菜单
INSERT INTO hg_admin_menu (pid, level, tree, title, name, path, component, redirect, icon, permissions, type, sort, hidden, status, remark, created_at, updated_at)
VALUES (@strategy_parent_id, 3, '', '策略模板', 'StrategyGroup', '/toogo/strategy/group', '/toogo/strategy/group', '', 'AppstoreOutlined', '', 1, 1, 0, 1, '策略模板管理', NOW(), NOW());

-- 添加策略列表菜单（隐藏，通过策略模板进入）
INSERT INTO hg_admin_menu (pid, level, tree, title, name, path, component, redirect, icon, permissions, type, sort, hidden, status, remark, created_at, updated_at)
VALUES (@strategy_parent_id, 3, '', '策略列表', 'StrategyList', '/toogo/strategy/list', '/toogo/strategy/list', '', 'UnorderedListOutlined', '', 1, 2, 1, 1, '策略列表（通过模板进入）', NOW(), NOW());

-- 分配权限给超级管理员
INSERT IGNORE INTO hg_admin_role_menu (role_id, menu_id) 
SELECT 1, id FROM hg_admin_menu WHERE path LIKE '/toogo/strategy%';

-- 查询验证
SELECT id, pid, title, path, hidden FROM hg_admin_menu WHERE path LIKE '/toogo/strategy%' ORDER BY pid, sort;
