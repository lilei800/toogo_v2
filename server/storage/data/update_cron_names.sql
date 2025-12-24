-- 更新定时任务name以匹配代码中的注册名称
UPDATE `hg_sys_cron` SET `name`='toogo_robot_engine' WHERE `name`='ToogoEngine';
UPDATE `hg_sys_cron` SET `name`='toogo_subscription_check' WHERE `name`='ToogoSubscriptionChecker';
SELECT id, title, name, pattern, status FROM `hg_sys_cron` WHERE `group_id`=10;

