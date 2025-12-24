-- ============================================================
-- 修正「我的邀请码」菜单位置和权限
-- ============================================================

SET NAMES utf8mb4;

-- 1. 获取推广菜单ID
SET @promote_menu_id = (SELECT id FROM `hg_admin_menu` WHERE `path` = '/toogo/promote' LIMIT 1);

-- 2. 将"我的邀请码"移到 /toogo/promote (推广) 菜单下
UPDATE `hg_admin_menu`
SET 
  `pid` = @promote_menu_id,
  `path` = '/toogo/promote/invite-code',
  `updated_at` = NOW()
WHERE `path` = '/toogo/invite-code';

-- 3. 删除空的"我的推广中心"菜单（如果没有其他子菜单）
SET @promotion_menu_id = (SELECT id FROM `hg_admin_menu` WHERE `path` = '/toogo/promotion' LIMIT 1);
SET @has_children = (SELECT COUNT(*) FROM `hg_admin_menu` WHERE pid = @promotion_menu_id);

DELETE FROM `hg_admin_menu`
WHERE `path` = '/toogo/promotion' AND @has_children = 0;

-- 4. 为"用户"角色(id=210)分配"我的邀请码"菜单权限
INSERT IGNORE INTO `hg_admin_role_menu` (`role_id`, `menu_id`)
SELECT 210, id FROM `hg_admin_menu` WHERE `path` = '/toogo/promote/invite-code';

-- 5. 确保"用户"角色有"推广"父菜单权限
INSERT IGNORE INTO `hg_admin_role_menu` (`role_id`, `menu_id`)
SELECT 210, id FROM `hg_admin_menu` WHERE `path` = '/toogo/promote';

-- 6. 确保"用户"角色有"我的团队"菜单权限
INSERT IGNORE INTO `hg_admin_role_menu` (`role_id`, `menu_id`)
SELECT 210, id FROM `hg_admin_menu` WHERE `path` = '/toogo/promote/team' OR `path` = 'team';

-- 7. 确保"用户"角色有"佣金记录"菜单权限
INSERT IGNORE INTO `hg_admin_role_menu` (`role_id`, `menu_id`)
SELECT 210, id FROM `hg_admin_menu` WHERE `path` = '/toogo/promote/commission' OR `path` = 'commission';

SELECT 'OK' AS result;
