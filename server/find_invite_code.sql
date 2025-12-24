-- 查找邀请码RSOW2235
-- 在toogo_user表中查找
SELECT 
    '在toogo_user表' as source,
    tu.id,
    tu.member_id,
    am.username,
    tu.invite_code as code,
    tu.invite_code_expire
FROM hg_toogo_user tu
LEFT JOIN hg_admin_member am ON tu.member_id = am.id
WHERE tu.invite_code = 'RSOW2235';

-- 在admin_member表中查找永久邀请码
SELECT 
    '在admin_member表(永久邀请码)' as source,
    am.id as member_id,
    am.username,
    am.invite_code as code,
    'permanent' as type
FROM hg_admin_member am
WHERE am.invite_code = 'RSOW2235';

-- 查找用户dong的信息
SELECT 
    'dong的信息' as info,
    am.id as member_id,
    am.username,
    am.invite_code as permanent_code,
    tu.invite_code as temp_code,
    tu.inviter_id,
    (SELECT username FROM hg_admin_member WHERE id = tu.inviter_id) as inviter_username
FROM hg_admin_member am
LEFT JOIN hg_toogo_user tu ON am.id = tu.member_id
WHERE am.username = 'dong';

