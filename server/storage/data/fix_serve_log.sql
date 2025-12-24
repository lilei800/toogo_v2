-- 修复 hg_sys_serve_log 表字段默认值问题
ALTER TABLE `hg_sys_serve_log` 
MODIFY COLUMN `line` INT DEFAULT 0 COMMENT '行号';

-- 如果字段不存在，添加字段
-- ALTER TABLE `hg_sys_serve_log` ADD COLUMN IF NOT EXISTS `line` INT DEFAULT 0 COMMENT '行号';

