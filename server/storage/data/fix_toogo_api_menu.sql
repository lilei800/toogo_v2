-- 修复Toogo API管理菜单的component路径
-- 确保菜单的component指向正确的Vue组件

-- 查看现有菜单
-- SELECT * FROM hg_admin_menu WHERE path LIKE '%api%' OR title LIKE '%API%';

-- 更新API管理菜单的component路径
UPDATE hg_admin_menu 
SET component = '/toogo/api/index'
WHERE path = '/toogo/api' OR path = 'api' AND pid IN (SELECT id FROM (SELECT id FROM hg_admin_menu WHERE path LIKE '%toogo%') AS t);

-- 如果菜单不存在，插入新菜单
INSERT INTO hg_admin_menu (pid, level, tree, title, name, path, icon, type, redirect, permissions, permission_name, component, always_show, active_menu, is_root, is_frame, frame_src, keep_alive, hidden, affix, sort, remark, status, updated_at, created_at)
SELECT 
    2000, 2, '', 'API管理', 'ToogoApi', '/toogo/api', 'ApiOutlined', 1, '', 'toogo:api:list', 'API管理', '/toogo/api/index', 0, '', 0, 1, '', 0, 0, 0, 6, '交易所API密钥管理', 1, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM hg_admin_menu WHERE path = '/toogo/api');

-- 验证结果
SELECT id, pid, title, path, component FROM hg_admin_menu WHERE title LIKE '%API%' OR path LIKE '%api%';

