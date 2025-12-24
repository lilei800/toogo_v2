-- ========================================
-- 修正永久邀请码 SQL脚本
-- ========================================
-- 说明：此脚本用于查询不符合规范的永久邀请码
-- 符合规范的格式：8位（4位大写字母 + 4位数字，数字不含4）
-- 
-- 使用方法：
-- 1. 先执行查询，查看需要修正的用户
-- 2. 使用 fix_invite_codes.bat 工具自动修正
-- 或
-- 2. 手动为这些用户更新邀请码
-- ========================================

-- 查询所有不符合规范的邀请码
-- （长度不是8位，或者包含小写字母、特殊字符、数字4等）
SELECT 
    id,
    username,
    realname,
    invite_code,
    LENGTH(invite_code) as code_length,
    created_at
FROM 
    hg_admin_member
WHERE 
    invite_code IS NOT NULL 
    AND invite_code != ''
    AND (
        -- 长度不是8位
        LENGTH(invite_code) != 8
        -- 或包含小写字母
        OR invite_code REGEXP '[a-z]'
        -- 或包含数字4
        OR invite_code LIKE '%4%'
        -- 或包含特殊字符
        OR invite_code REGEXP '[^A-Z0-9]'
        -- 或前4位不全是大写字母
        OR SUBSTRING(invite_code, 1, 4) REGEXP '[^A-Z]'
        -- 或后4位不全是数字
        OR SUBSTRING(invite_code, 5, 4) REGEXP '[^0-9]'
    )
ORDER BY 
    id;

-- ========================================
-- 统计信息
-- ========================================

-- 统计需要修正的用户数量
SELECT 
    COUNT(*) as need_fix_count
FROM 
    hg_admin_member
WHERE 
    invite_code IS NOT NULL 
    AND invite_code != ''
    AND (
        LENGTH(invite_code) != 8
        OR invite_code REGEXP '[a-z]'
        OR invite_code LIKE '%4%'
        OR invite_code REGEXP '[^A-Z0-9]'
        OR SUBSTRING(invite_code, 1, 4) REGEXP '[^A-Z]'
        OR SUBSTRING(invite_code, 5, 4) REGEXP '[^0-9]'
    );

-- 查看所有用户的邀请码（用于验证）
SELECT 
    id,
    username,
    invite_code,
    LENGTH(invite_code) as code_length,
    CASE 
        WHEN LENGTH(invite_code) = 8 
            AND SUBSTRING(invite_code, 1, 4) REGEXP '^[A-Z]{4}$'
            AND SUBSTRING(invite_code, 5, 4) REGEXP '^[0-9]{4}$'
            AND invite_code NOT LIKE '%4%'
        THEN '✅ 符合规范'
        ELSE '❌ 不符合规范'
    END as status
FROM 
    hg_admin_member
WHERE 
    invite_code IS NOT NULL 
    AND invite_code != ''
ORDER BY 
    status DESC, id;

-- ========================================
-- 备份表（执行修正前建议先备份）
-- ========================================
-- CREATE TABLE hg_admin_member_backup_invite_codes AS 
-- SELECT id, username, invite_code, updated_at 
-- FROM hg_admin_member 
-- WHERE invite_code IS NOT NULL AND invite_code != '';

