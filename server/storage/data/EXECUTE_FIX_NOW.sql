-- ================================================
-- 🔧 立即执行修复脚本
-- ================================================
-- 功能: 修复 hg_sys_serve_log 表字段问题
-- 执行方式: 
--   1. 打开 Navicat/MySQL Workbench
--   2. 连接到 hotgo 数据库
--   3. 复制下面的SQL执行
-- ================================================

USE hotgo;

-- 修复 line 字段默认值
ALTER TABLE `hg_sys_serve_log` 
MODIFY COLUMN `line` INT DEFAULT 0 COMMENT '行号';

-- 验证修复结果
DESC hg_sys_serve_log;

-- 清理之前的错误日志（可选）
-- DELETE FROM hg_sys_serve_log WHERE level_format = 'ERRO' AND created_at < NOW();

SELECT '✅ 数据库修复完成！请重启服务。' AS status;

