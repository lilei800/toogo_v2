-- ============================================================
-- 清理波动率配置菜单（保留唯一的简化版）
-- ============================================================

-- 1. 删除旧的波动率配置菜单（如果存在）
DELETE FROM `hg_admin_menu` WHERE `path` = '/toogo-admin/volatility';
DELETE FROM `hg_admin_menu` WHERE `name` = 'ToogoAdminVolatility';

-- 2. 更新或插入新的波动率配置菜单
-- 先尝试更新已存在的 volatility-config 菜单
UPDATE `hg_admin_menu` SET 
    `title` = '波动率配置',
    `name` = 'ToogoAdminVolatilityConfig',
    `path` = '/toogo-admin/volatility-config',
    `icon` = 'LineChartOutlined',
    `component` = '/toogo/admin/volatility-config/index',
    `is_hide` = 0,
    `status` = 1,
    `updated_at` = NOW()
WHERE `path` = '/toogo-admin/volatility-config' OR `name` = 'ToogoAdminVolatilityConfig';

-- 3. 如果不存在，插入新菜单
INSERT INTO `hg_admin_menu` (
    `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`,
    `permission_name`, `component`, `always_show`, `active_menu`, `is_root`,
    `is_frame`, `frame_src`, `keep_alive`, `is_hide`, `is_affix`, `sort`,
    `remark`, `status`, `created_at`, `updated_at`
)
SELECT 
    (SELECT id FROM `hg_admin_menu` WHERE `path` = '/toogo-admin' LIMIT 1) as pid,
    '波动率配置' as title,
    'ToogoAdminVolatilityConfig' as name,
    '/toogo-admin/volatility-config' as path,
    'LineChartOutlined' as icon,
    1 as type,
    '' as redirect,
    '/volatility/config/list,/volatility/config/create,/volatility/config/update,/volatility/config/delete,/volatility/config/batch-edit' as permissions,
    '波动率配置' as permission_name,
    '/toogo/admin/volatility-config/index' as component,
    0 as always_show,
    '' as active_menu,
    0 as is_root,
    0 as is_frame,
    '' as frame_src,
    1 as keep_alive,
    0 as is_hide,
    0 as is_affix,
    350 as sort,
    '波动率配置管理（简化版）' as remark,
    1 as status,
    NOW() as created_at,
    NOW() as updated_at
WHERE NOT EXISTS (
    SELECT 1 FROM `hg_admin_menu` 
    WHERE `path` = '/toogo-admin/volatility-config' OR `name` = 'ToogoAdminVolatilityConfig'
);

-- 4. 验证结果
SELECT id, pid, title, name, path, icon, status 
FROM `hg_admin_menu` 
WHERE `path` LIKE '%volatility%' OR `name` LIKE '%Volatility%'
ORDER BY sort;

-- ============================================================
-- 说明：
-- 此脚本将清理旧的波动率配置菜单，只保留新的简化版波动率配置
-- 新页面路径：/toogo-admin/volatility-config
-- 功能：每个货币对独立配置市场状态阈值和5个时间周期权重
-- ============================================================

