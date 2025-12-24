-- 双向开单功能测试SQL脚本
-- 用于验证字段和功能状态

-- 1. 检查字段是否存在
SELECT 
    '字段检查' as 测试项,
    CASE 
        WHEN COUNT(*) > 0 THEN '✅ 字段存在'
        ELSE '❌ 字段不存在'
    END as 结果,
    COUNT(*) as 字段数量
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
  AND TABLE_NAME = 'hg_trading_robot' 
  AND COLUMN_NAME = 'allow_bidirectional_order';

-- 2. 查看字段详细信息
SELECT 
    '字段详情' as 测试项,
    COLUMN_NAME as 字段名,
    COLUMN_TYPE as 字段类型,
    COLUMN_DEFAULT as 默认值,
    IS_NULLABLE as 允许空值,
    COLUMN_COMMENT as 注释
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
  AND TABLE_NAME = 'hg_trading_robot' 
  AND COLUMN_NAME = 'allow_bidirectional_order';

-- 3. 统计机器人双向开单状态
SELECT 
    '机器人统计' as 测试项,
    COUNT(*) as 总机器人数量,
    SUM(CASE WHEN allow_bidirectional_order = 1 THEN 1 ELSE 0 END) as 开启双向开单,
    SUM(CASE WHEN allow_bidirectional_order = 0 THEN 1 ELSE 0 END) as 关闭双向开单,
    SUM(CASE WHEN allow_bidirectional_order IS NULL THEN 1 ELSE 0 END) as 未设置
FROM hg_trading_robot
WHERE deleted_at IS NULL;

-- 4. 查看运行中机器人的状态
SELECT 
    '运行中机器人' as 测试项,
    id,
    robot_name,
    status,
    auto_trade_enabled as 自动下单,
    allow_bidirectional_order as 双向开单,
    auto_close_enabled as 自动平仓
FROM hg_trading_robot
WHERE deleted_at IS NULL 
  AND status = 2  -- 运行中
ORDER BY id DESC
LIMIT 10;

-- 5. 测试更新操作（可选，用于测试）
-- UPDATE hg_trading_robot 
-- SET allow_bidirectional_order = 0 
-- WHERE id = [测试机器人ID];

-- 6. 验证更新后的状态
-- SELECT id, robot_name, allow_bidirectional_order 
-- FROM hg_trading_robot 
-- WHERE id = [测试机器人ID];

