-- 为 hg_toogo_power_consume 增加幂等唯一索引，避免同一订单重复扣算力
-- 说明：
-- 1) 如历史上已经产生重复数据，请先执行清理 SQL（本文件已包含清理语句）。
-- 2) 建议在低峰期执行，并提前备份相关表。

-- 0. 预检查：是否存在重复（返回 0 表示没有重复）
SELECT COUNT(*) AS dup_groups
FROM (
  SELECT user_id, order_id, COUNT(*) AS cnt
  FROM hg_toogo_power_consume
  GROUP BY user_id, order_id
  HAVING cnt > 1
) d;

-- 1. 清理重复数据（保留 id 最大的那条）
DELETE t1
FROM hg_toogo_power_consume t1
INNER JOIN hg_toogo_power_consume t2
WHERE t1.user_id = t2.user_id
  AND t1.order_id = t2.order_id
  AND t1.id < t2.id;

-- 2. 添加唯一索引（同一用户同一订单只能有一条消耗记录）
-- 幂等：若索引已存在则跳过
SET @idx_exists := (
  SELECT COUNT(1)
  FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'hg_toogo_power_consume'
    AND INDEX_NAME = 'uk_user_order'
);
SET @ddl := IF(@idx_exists > 0,
  'SELECT ''[SKIP] uk_user_order already exists'' AS msg',
  'ALTER TABLE hg_toogo_power_consume ADD UNIQUE KEY uk_user_order (user_id, order_id)'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;



