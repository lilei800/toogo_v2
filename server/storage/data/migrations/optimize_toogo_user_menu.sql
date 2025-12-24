-- ============================================================
-- 优化量化交易用户菜单结构
-- 执行前请先备份数据库
-- ============================================================

-- 获取量化交易父菜单ID
SET @toogo_parent_id = (SELECT id FROM `hg_admin_menu` WHERE `path` = '/toogo' LIMIT 1);

-- 如果量化交易父菜单不存在，先创建
INSERT INTO `hg_admin_menu` (
    `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`,
    `permission_name`, `component`, `always_show`, `active_menu`, `is_root`,
    `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`,
    `remark`, `status`, `created_at`, `updated_at`
)
SELECT 
    0 as pid,
    '量化交易' as title,
    'ToogoRoot' as name,
    '/toogo' as path,
    'BarChartOutlined' as icon,
    1 as type,
    '/toogo/dashboard' as redirect,
    '' as permissions,
    '量化交易' as permission_name,
    'LAYOUT' as component,
    0 as always_show,
    '' as active_menu,
    0 as is_root,
    0 as is_frame,
    '' as frame_src,
    0 as keep_alive,
    0 as hidden,
    0 as affix,
    2 as sort,
    '用户量化交易菜单' as remark,
    1 as status,
    NOW() as created_at,
    NOW() as updated_at
WHERE NOT EXISTS (SELECT 1 FROM `hg_admin_menu` WHERE `path` = '/toogo');

-- 重新获取父菜单ID
SET @toogo_parent_id = (SELECT id FROM `hg_admin_menu` WHERE `path` = '/toogo' LIMIT 1);

-- ============================================================
-- 1. 控制台（保持不变）
-- ============================================================
UPDATE `hg_admin_menu` SET 
    `sort` = 10,
    `title` = '控制台'
WHERE `path` = '/toogo/dashboard' AND `pid` = @toogo_parent_id;

-- ============================================================
-- 2. 我的机器人（核心功能，排序提前）
-- ============================================================
UPDATE `hg_admin_menu` SET 
    `sort` = 20,
    `title` = '我的机器人'
WHERE `path` = '/toogo/robot' AND `pid` = @toogo_parent_id;

-- 2.1 创建机器人（隐藏页面）
INSERT INTO `hg_admin_menu` (
    `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`,
    `permission_name`, `component`, `always_show`, `active_menu`, `is_root`,
    `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`,
    `remark`, `status`, `created_at`, `updated_at`
)
SELECT 
    @toogo_parent_id as pid,
    '创建机器人' as title,
    'ToogoRobotCreate' as name,
    '/toogo/robot/create' as path,
    '' as icon,
    2 as type,
    '' as redirect,
    '' as permissions,
    '创建机器人' as permission_name,
    '/toogo/robot/create' as component,
    0 as always_show,
    '/toogo/robot' as active_menu,
    0 as is_root,
    0 as is_frame,
    '' as frame_src,
    0 as keep_alive,
    1 as hidden,
    0 as affix,
    21 as sort,
    '创建机器人页面（隐藏）' as remark,
    1 as status,
    NOW() as created_at,
    NOW() as updated_at
WHERE NOT EXISTS (SELECT 1 FROM `hg_admin_menu` WHERE `path` = '/toogo/robot/create');

-- ============================================================
-- 3. 策略模板（简化为单页）
-- ============================================================
-- 先删除旧的策略管理子菜单结构
DELETE FROM `hg_admin_menu` WHERE `path` IN (
    '/toogo/strategy/my',
    '/toogo/strategy/official', 
    '/toogo/strategy/ranking'
);

-- 更新或创建策略模板菜单
UPDATE `hg_admin_menu` SET 
    `sort` = 30,
    `title` = '策略模板',
    `redirect` = '',
    `component` = '/toogo/strategy/index'
WHERE `path` = '/toogo/strategy' AND `pid` = @toogo_parent_id;

-- 如果不存在，插入策略模板菜单
INSERT INTO `hg_admin_menu` (
    `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`,
    `permission_name`, `component`, `always_show`, `active_menu`, `is_root`,
    `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`,
    `remark`, `status`, `created_at`, `updated_at`
)
SELECT 
    @toogo_parent_id as pid,
    '策略模板' as title,
    'ToogoStrategy' as name,
    '/toogo/strategy' as path,
    'SettingOutlined' as icon,
    2 as type,
    '' as redirect,
    '' as permissions,
    '策略模板' as permission_name,
    '/toogo/strategy/index' as component,
    0 as always_show,
    '' as active_menu,
    0 as is_root,
    0 as is_frame,
    '' as frame_src,
    1 as keep_alive,
    0 as hidden,
    0 as affix,
    30 as sort,
    '策略模板管理（单页Tab）' as remark,
    1 as status,
    NOW() as created_at,
    NOW() as updated_at
WHERE NOT EXISTS (SELECT 1 FROM `hg_admin_menu` WHERE `path` = '/toogo/strategy' AND `pid` = @toogo_parent_id);

-- 策略列表（隐藏页面）
INSERT INTO `hg_admin_menu` (
    `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`,
    `permission_name`, `component`, `always_show`, `active_menu`, `is_root`,
    `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`,
    `remark`, `status`, `created_at`, `updated_at`
)
SELECT 
    @toogo_parent_id as pid,
    '策略列表' as title,
    'StrategyList' as name,
    '/toogo/strategy/list' as path,
    '' as icon,
    2 as type,
    '' as redirect,
    '' as permissions,
    '策略列表' as permission_name,
    '/toogo/strategy/list' as component,
    0 as always_show,
    '/toogo/strategy' as active_menu,
    0 as is_root,
    0 as is_frame,
    '' as frame_src,
    0 as keep_alive,
    1 as hidden,
    0 as affix,
    31 as sort,
    '策略列表详情页面（隐藏）' as remark,
    1 as status,
    NOW() as created_at,
    NOW() as updated_at
WHERE NOT EXISTS (SELECT 1 FROM `hg_admin_menu` WHERE `path` = '/toogo/strategy/list');

-- ============================================================
-- 4. API配置
-- ============================================================
UPDATE `hg_admin_menu` SET 
    `sort` = 40,
    `title` = 'API配置'
WHERE `path` = '/toogo/api' AND `pid` = @toogo_parent_id;

-- ============================================================
-- 5. 钱包中心（合并财务中心和订阅套餐）
-- ============================================================
-- 创建钱包中心父菜单
INSERT INTO `hg_admin_menu` (
    `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`,
    `permission_name`, `component`, `always_show`, `active_menu`, `is_root`,
    `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`,
    `remark`, `status`, `created_at`, `updated_at`
)
SELECT 
    @toogo_parent_id as pid,
    '钱包中心' as title,
    'ToogoWallet' as name,
    '/toogo/wallet' as path,
    'WalletOutlined' as icon,
    1 as type,
    '/toogo/wallet/overview' as redirect,
    '' as permissions,
    '钱包中心' as permission_name,
    'LAYOUT' as component,
    0 as always_show,
    '' as active_menu,
    0 as is_root,
    0 as is_frame,
    '' as frame_src,
    0 as keep_alive,
    0 as hidden,
    0 as affix,
    50 as sort,
    '钱包中心（资产+订阅）' as remark,
    1 as status,
    NOW() as created_at,
    NOW() as updated_at
WHERE NOT EXISTS (SELECT 1 FROM `hg_admin_menu` WHERE `path` = '/toogo/wallet');

SET @wallet_parent_id = (SELECT id FROM `hg_admin_menu` WHERE `path` = '/toogo/wallet' LIMIT 1);

-- 5.1 资产总览
INSERT INTO `hg_admin_menu` (
    `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`,
    `permission_name`, `component`, `always_show`, `active_menu`, `is_root`,
    `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`,
    `remark`, `status`, `created_at`, `updated_at`
)
SELECT 
    @wallet_parent_id as pid,
    '资产总览' as title,
    'WalletOverview' as name,
    '/toogo/wallet/overview' as path,
    'DashboardOutlined' as icon,
    2 as type,
    '' as redirect,
    '' as permissions,
    '资产总览' as permission_name,
    '/toogo/finance/index' as component,
    0 as always_show,
    '' as active_menu,
    0 as is_root,
    0 as is_frame,
    '' as frame_src,
    1 as keep_alive,
    0 as hidden,
    0 as affix,
    51 as sort,
    '资产总览（原财务中心）' as remark,
    1 as status,
    NOW() as created_at,
    NOW() as updated_at
WHERE NOT EXISTS (SELECT 1 FROM `hg_admin_menu` WHERE `path` = '/toogo/wallet/overview');

-- 5.2 订阅套餐
INSERT INTO `hg_admin_menu` (
    `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`,
    `permission_name`, `component`, `always_show`, `active_menu`, `is_root`,
    `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`,
    `remark`, `status`, `created_at`, `updated_at`
)
SELECT 
    @wallet_parent_id as pid,
    '订阅套餐' as title,
    'WalletSubscription' as name,
    '/toogo/wallet/subscription' as path,
    'CreditCardOutlined' as icon,
    2 as type,
    '' as redirect,
    '' as permissions,
    '订阅套餐' as permission_name,
    '/toogo/subscription/index' as component,
    0 as always_show,
    '' as active_menu,
    0 as is_root,
    0 as is_frame,
    '' as frame_src,
    1 as keep_alive,
    0 as hidden,
    0 as affix,
    52 as sort,
    '订阅套餐' as remark,
    1 as status,
    NOW() as created_at,
    NOW() as updated_at
WHERE NOT EXISTS (SELECT 1 FROM `hg_admin_menu` WHERE `path` = '/toogo/wallet/subscription');

-- ============================================================
-- 6. 推广中心（合并我的团队和佣金记录）
-- ============================================================
-- 创建推广中心父菜单
INSERT INTO `hg_admin_menu` (
    `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`,
    `permission_name`, `component`, `always_show`, `active_menu`, `is_root`,
    `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`,
    `remark`, `status`, `created_at`, `updated_at`
)
SELECT 
    @toogo_parent_id as pid,
    '推广中心' as title,
    'ToogoPromote' as name,
    '/toogo/promote' as path,
    'TeamOutlined' as icon,
    1 as type,
    '/toogo/promote/team' as redirect,
    '' as permissions,
    '推广中心' as permission_name,
    'LAYOUT' as component,
    0 as always_show,
    '' as active_menu,
    0 as is_root,
    0 as is_frame,
    '' as frame_src,
    0 as keep_alive,
    0 as hidden,
    0 as affix,
    60 as sort,
    '推广中心（团队+佣金）' as remark,
    1 as status,
    NOW() as created_at,
    NOW() as updated_at
WHERE NOT EXISTS (SELECT 1 FROM `hg_admin_menu` WHERE `path` = '/toogo/promote');

SET @promote_parent_id = (SELECT id FROM `hg_admin_menu` WHERE `path` = '/toogo/promote' LIMIT 1);

-- 6.1 我的团队
INSERT INTO `hg_admin_menu` (
    `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`,
    `permission_name`, `component`, `always_show`, `active_menu`, `is_root`,
    `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`,
    `remark`, `status`, `created_at`, `updated_at`
)
SELECT 
    @promote_parent_id as pid,
    '我的团队' as title,
    'PromoteTeam' as name,
    '/toogo/promote/team' as path,
    'TeamOutlined' as icon,
    2 as type,
    '' as redirect,
    '' as permissions,
    '我的团队' as permission_name,
    '/toogo/team/index' as component,
    0 as always_show,
    '' as active_menu,
    0 as is_root,
    0 as is_frame,
    '' as frame_src,
    1 as keep_alive,
    0 as hidden,
    0 as affix,
    61 as sort,
    '我的团队（原团队页面）' as remark,
    1 as status,
    NOW() as created_at,
    NOW() as updated_at
WHERE NOT EXISTS (SELECT 1 FROM `hg_admin_menu` WHERE `path` = '/toogo/promote/team');

-- 6.2 佣金记录
INSERT INTO `hg_admin_menu` (
    `pid`, `title`, `name`, `path`, `icon`, `type`, `redirect`, `permissions`,
    `permission_name`, `component`, `always_show`, `active_menu`, `is_root`,
    `is_frame`, `frame_src`, `keep_alive`, `hidden`, `affix`, `sort`,
    `remark`, `status`, `created_at`, `updated_at`
)
SELECT 
    @promote_parent_id as pid,
    '佣金记录' as title,
    'PromoteCommission' as name,
    '/toogo/promote/commission' as path,
    'RiseOutlined' as icon,
    2 as type,
    '' as redirect,
    '' as permissions,
    '佣金记录' as permission_name,
    '/toogo/commission/index' as component,
    0 as always_show,
    '' as active_menu,
    0 as is_root,
    0 as is_frame,
    '' as frame_src,
    1 as keep_alive,
    0 as hidden,
    0 as affix,
    62 as sort,
    '佣金记录' as remark,
    1 as status,
    NOW() as created_at,
    NOW() as updated_at
WHERE NOT EXISTS (SELECT 1 FROM `hg_admin_menu` WHERE `path` = '/toogo/promote/commission');

-- ============================================================
-- 7. 隐藏旧的一级菜单（可选：保留兼容性）
-- ============================================================
-- 将旧菜单设置为隐藏，保持向后兼容
UPDATE `hg_admin_menu` SET `hidden` = 1, `sort` = 999 
WHERE `path` IN ('/toogo/subscription', '/toogo/team', '/toogo/commission', '/toogo/finance')
  AND `pid` = @toogo_parent_id;

-- ============================================================
-- 8. 验证结果
-- ============================================================
SELECT id, pid, title, name, path, sort, hidden, status 
FROM `hg_admin_menu` 
WHERE `path` LIKE '/toogo%'
ORDER BY sort, pid;

-- ============================================================
-- 新菜单结构：
-- 量化交易
-- ├── 控制台 (sort: 10)
-- ├── 我的机器人 (sort: 20)
-- │   └── 创建机器人 (hidden, sort: 21)
-- ├── 策略模板 (sort: 30)
-- │   └── 策略列表 (hidden, sort: 31)
-- ├── API配置 (sort: 40)
-- ├── 钱包中心 (sort: 50)
-- │   ├── 资产总览 (sort: 51)
-- │   └── 订阅套餐 (sort: 52)
-- └── 推广中心 (sort: 60)
--     ├── 我的团队 (sort: 61)
--     └── 佣金记录 (sort: 62)
-- ============================================================

