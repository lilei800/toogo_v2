-- 更新用户dong的上级为邀请码RSOW2235对应的用户
-- 步骤1: 查找邀请码RSOW2235对应的用户
SELECT 
    tu.id,
    tu.member_id,
    am.username,
    tu.invite_code
FROM hg_toogo_user tu
LEFT JOIN hg_admin_member am ON tu.member_id = am.id
WHERE tu.invite_code = 'RSOW2235';

-- 步骤2: 查找用户dong的当前信息
SELECT 
    tu.id,
    tu.member_id,
    am.username,
    tu.inviter_id,
    tu.invite_count
FROM hg_toogo_user tu
LEFT JOIN hg_admin_member am ON tu.member_id = am.id
WHERE am.username = 'dong';

-- 步骤3: 开始事务，更新dong的上级
START TRANSACTION;

-- 获取新上级的member_id
SET @new_inviter_member_id = (
    SELECT member_id 
    FROM hg_toogo_user 
    WHERE invite_code = 'RSOW2235'
);

-- 获取dong的member_id和旧上级ID
SET @dong_member_id = (
    SELECT member_id 
    FROM hg_toogo_user tu
    LEFT JOIN hg_admin_member am ON tu.member_id = am.id
    WHERE am.username = 'dong'
);

SET @old_inviter_id = (
    SELECT inviter_id 
    FROM hg_toogo_user 
    WHERE member_id = @dong_member_id
);

-- 如果有旧上级，减少旧上级的邀请计数
UPDATE hg_toogo_user 
SET invite_count = GREATEST(0, invite_count - 1)
WHERE member_id = @old_inviter_id AND @old_inviter_id > 0;

-- 更新dong的inviter_id
UPDATE hg_toogo_user 
SET inviter_id = @new_inviter_member_id,
    updated_at = NOW()
WHERE member_id = @dong_member_id;

-- 增加新上级的邀请计数
UPDATE hg_toogo_user 
SET invite_count = invite_count + 1
WHERE member_id = @new_inviter_member_id;

-- 验证更新结果
SELECT 
    tu.id,
    tu.member_id,
    am.username,
    tu.inviter_id,
    (SELECT username FROM hg_admin_member WHERE id = tu.inviter_id) as inviter_username,
    (SELECT invite_code FROM hg_toogo_user WHERE member_id = tu.inviter_id) as inviter_code
FROM hg_toogo_user tu
LEFT JOIN hg_admin_member am ON tu.member_id = am.id
WHERE am.username = 'dong';

-- 如果一切正常，提交事务
COMMIT;

-- 如果有问题，可以回滚
-- ROLLBACK;

