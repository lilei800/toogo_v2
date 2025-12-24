-- ===========================================
-- 数据库修复脚本
-- 解决机器人连接失败问题
-- ===========================================

-- 1. 修复 hg_sys_serve_log 表的 line 字段
ALTER TABLE `hg_sys_serve_log` 
MODIFY COLUMN `line` INT DEFAULT 0 COMMENT '行号';

-- 2. 如果 line 字段不存在，添加字段（可选）
-- ALTER TABLE `hg_sys_serve_log` ADD COLUMN IF NOT EXISTS `line` INT DEFAULT 0 COMMENT '行号';

SELECT '✅ 数据库修复完成！' AS status;

