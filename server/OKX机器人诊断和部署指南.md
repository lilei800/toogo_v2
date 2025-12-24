# OKX机器人诊断和部署指南

> 生成时间：2024-12-25
> 
> 目的：快速诊断OKX机器人下单问题并部署优化

---

## 一、快速诊断（使用数据库工具）🔍

### 方法1：使用Navicat/DBeaver/pgAdmin

#### 步骤1：连接数据库
```
主机：127.0.0.1
端口：5432
数据库：hotgo
用户名：postgres
密码：postgres
```

#### 步骤2：执行诊断查询

**查询1：查看OKX机器人配置**
```sql
SELECT 
    id, 
    robot_name, 
    auto_trade_enabled,  -- 0=未开启，1=已开启
    auto_close_enabled,
    dual_side_position   -- 0=单向，1=双向
FROM hg_trading_robot
WHERE platform = 'okx' 
  AND status = 2  -- 运行中
  AND deleted_at IS NULL
ORDER BY id DESC;
```

**查询2：查看最近失败原因统计（最重要）**
```sql
SELECT 
    message AS failure_reason,
    COUNT(*) AS count,
    MAX(created_at) AS last_occurrence
FROM hg_trading_execution_log
WHERE robot_id IN (
    SELECT id FROM hg_trading_robot 
    WHERE platform = 'okx' AND status = 2 AND deleted_at IS NULL
)
  AND status = 'failed'
  AND created_at > NOW() - INTERVAL '24 hours'
GROUP BY message
ORDER BY count DESC
LIMIT 10;
```

**查询3：查看最近10条失败日志详情**
```sql
SELECT 
    id,
    robot_id,
    event_type,
    status,
    message,
    event_data,
    created_at
FROM hg_trading_execution_log
WHERE robot_id IN (
    SELECT id FROM hg_trading_robot 
    WHERE platform = 'okx' AND status = 2
)
  AND status = 'failed'
ORDER BY created_at DESC
LIMIT 10;
```

**查询4：查看有预警但无执行日志的情况**
```sql
SELECT 
    s.id AS signal_log_id,
    s.robot_id,
    s.direction,
    s.action,
    s.is_processed,
    s.executed,
    s.created_at
FROM hg_trading_signal_log s
WHERE s.robot_id IN (
    SELECT id FROM hg_trading_robot 
    WHERE platform = 'okx' AND status = 2
)
  AND s.created_at > NOW() - INTERVAL '24 hours'
  AND NOT EXISTS (
      SELECT 1 FROM hg_trading_execution_log e 
      WHERE e.signal_log_id = s.id
  )
ORDER BY s.created_at DESC
LIMIT 20;
```

---

## 二、常见问题快速修复 🔧

### 问题1：自动交易未开启

**症状：** `message = "自动下单未开启"`

**解决方案：**
```sql
-- 开启自动交易
UPDATE hg_trading_robot 
SET auto_trade_enabled = 1 
WHERE platform = 'okx' AND status = 2;

-- 验证
SELECT id, robot_name, auto_trade_enabled 
FROM hg_trading_robot 
WHERE platform = 'okx' AND status = 2;
```

### 问题2：单向持仓模式限制

**症状：** `message = "单向持仓模式：已有持仓..."`

**解决方案A：切换到双向持仓模式**
```sql
UPDATE hg_trading_robot 
SET dual_side_position = 1 
WHERE platform = 'okx' AND status = 2;
```

**解决方案B：查看并平仓现有持仓**
```sql
-- 查看当前持仓
SELECT * FROM hg_trading_order 
WHERE robot_id IN (SELECT id FROM hg_trading_robot WHERE platform = 'okx' AND status = 2)
  AND status = 1  -- status=1表示开仓中
ORDER BY created_at DESC;

-- 需要手动平仓或等待自动平仓
```

### 问题3：策略配置缺失

**症状：** `message = "市场状态=xxx 在映射关系中未找到对应的风险偏好"`

**解决方案：**
```sql
-- 查看当前配置
SELECT id, robot_name, remark 
FROM hg_trading_robot 
WHERE platform = 'okx' AND status = 2;

-- 如果remark为空或格式错误，更新配置
UPDATE hg_trading_robot 
SET remark = '{"high_vol":"aggressive","low_vol":"conservative","trend":"balanced"}'
WHERE platform = 'okx' AND status = 2 AND (remark IS NULL OR remark = '{}');
```

### 问题4：余额不足

**症状：** `message = "余额不足..."`

**解决方案：**
```sql
-- 查看当前保证金配置
SELECT 
    r.id, 
    r.robot_name,
    s.margin_percent_min, 
    s.margin_percent_max,
    s.leverage_min
FROM hg_trading_robot r
JOIN hg_trading_strategy_template s ON r.strategy_id = s.id
WHERE r.platform = 'okx' AND r.status = 2;

-- 临时降低保证金比例（谨慎操作）
UPDATE hg_trading_strategy_template 
SET margin_percent_min = 5, margin_percent_max = 10
WHERE id IN (
    SELECT strategy_id FROM hg_trading_robot 
    WHERE platform = 'okx' AND status = 2
);
```

---

## 三、部署优化（数据库升级）📦

### 步骤1：备份数据库（重要！）

在数据库工具中执行：
```sql
-- 或使用数据库工具的备份功能
```

### 步骤2：执行数据库升级脚本

**在数据库工具中打开并执行：** `add_failure_category_fields.sql`

或直接执行以下SQL：

```sql
-- 1. 增加失败分类字段
ALTER TABLE hg_trading_execution_log 
ADD COLUMN IF NOT EXISTS failure_category VARCHAR(50) DEFAULT NULL;

-- 2. 增加结构化失败原因字段
ALTER TABLE hg_trading_execution_log 
ADD COLUMN IF NOT EXISTS failure_reason TEXT DEFAULT NULL;

-- 3. 创建索引
CREATE INDEX IF NOT EXISTS idx_failure_category 
ON hg_trading_execution_log(failure_category, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_robot_status_category 
ON hg_trading_execution_log(robot_id, status, failure_category, created_at DESC);

-- 4. 验证字段已添加
SELECT column_name, data_type, character_maximum_length
FROM information_schema.columns
WHERE table_name = 'hg_trading_execution_log'
  AND column_name IN ('failure_category', 'failure_reason')
ORDER BY ordinal_position;
```

### 步骤3：编译和部署代码

```powershell
# 1. 停止服务（如果正在运行）
# 在任务管理器中结束 hotgo.exe 进程，或：
taskkill /F /IM hotgo.exe

# 2. 进入项目目录
cd D:\go\src\hotgo_v2\server

# 3. 编译
go build -o hotgo.exe main.go

# 4. 启动服务
.\hotgo.exe

# 或者后台运行：
Start-Process -FilePath ".\hotgo.exe" -WindowStyle Hidden
```

---

## 四、验证部署效果 ✅

### 验证1：检查字段是否添加

```sql
SELECT column_name, data_type 
FROM information_schema.columns
WHERE table_name = 'hg_trading_execution_log'
  AND column_name IN ('failure_category', 'failure_reason');
```

**预期结果：**
```
column_name       | data_type
------------------+-----------
failure_category  | character varying
failure_reason    | text
```

### 验证2：触发一个失败场景测试

```sql
-- 关闭自动交易（触发失败）
UPDATE hg_trading_robot 
SET auto_trade_enabled = 0 
WHERE id = <your_robot_id>;

-- 等待信号触发（约1-2分钟）

-- 查看最新的执行日志
SELECT 
    id,
    event_type,
    status,
    message,
    failure_category,  -- 应该有值
    failure_reason,    -- 应该有值
    created_at
FROM hg_trading_execution_log
WHERE robot_id = <your_robot_id>
ORDER BY created_at DESC
LIMIT 5;

-- 恢复自动交易
UPDATE hg_trading_robot 
SET auto_trade_enabled = 1 
WHERE id = <your_robot_id>;
```

**预期结果：**
- `failure_category` = `'config'`
- `failure_reason` = `'自动交易开关未开启。解决方案：在机器人设置中开启自动交易开关'`

### 验证3：查看失败分类统计

```sql
SELECT 
    failure_category,
    COUNT(*) AS count,
    MAX(created_at) AS last_occurrence
FROM hg_trading_execution_log
WHERE robot_id IN (
    SELECT id FROM hg_trading_robot 
    WHERE platform = 'okx' AND status = 2
)
  AND status = 'failed'
  AND created_at > NOW() - INTERVAL '1 hour'
GROUP BY failure_category
ORDER BY count DESC;
```

---

## 五、监控和维护 📊

### 日常监控查询

**1. 查看最近1小时的失败统计**
```sql
SELECT 
    failure_category,
    COUNT(*) AS count
FROM hg_trading_execution_log
WHERE status = 'failed'
  AND created_at > NOW() - INTERVAL '1 hour'
GROUP BY failure_category
ORDER BY count DESC;
```

**2. 查看特定机器人的失败历史**
```sql
SELECT 
    id,
    failure_category,
    failure_reason,
    created_at
FROM hg_trading_execution_log
WHERE robot_id = <robot_id>
  AND status = 'failed'
ORDER BY created_at DESC
LIMIT 20;
```

**3. 查看所有OKX机器人的健康状态**
```sql
SELECT 
    r.id,
    r.robot_name,
    r.auto_trade_enabled,
    r.dual_side_position,
    COUNT(e.id) FILTER (WHERE e.status = 'failed' AND e.created_at > NOW() - INTERVAL '1 hour') AS failures_last_hour,
    COUNT(e.id) FILTER (WHERE e.status = 'success' AND e.created_at > NOW() - INTERVAL '1 hour') AS success_last_hour
FROM hg_trading_robot r
LEFT JOIN hg_trading_execution_log e ON r.id = e.robot_id
WHERE r.platform = 'okx' AND r.status = 2
GROUP BY r.id, r.robot_name, r.auto_trade_enabled, r.dual_side_position
ORDER BY failures_last_hour DESC;
```

---

## 六、故障排查流程 🔍

### 流程图

```
1. 查看失败原因统计
   ↓
2. 识别最常见的失败类型
   ↓
3. 根据失败分类执行对应的修复方案
   ↓
4. 验证修复效果
   ↓
5. 持续监控
```

### 快速排查命令

```sql
-- 一键诊断：查看所有关键信息
WITH robot_info AS (
    SELECT id, robot_name, auto_trade_enabled, dual_side_position
    FROM hg_trading_robot
    WHERE platform = 'okx' AND status = 2
),
failure_stats AS (
    SELECT 
        robot_id,
        failure_category,
        COUNT(*) AS count
    FROM hg_trading_execution_log
    WHERE status = 'failed'
      AND created_at > NOW() - INTERVAL '24 hours'
    GROUP BY robot_id, failure_category
)
SELECT 
    r.id,
    r.robot_name,
    r.auto_trade_enabled,
    r.dual_side_position,
    f.failure_category,
    f.count AS failure_count
FROM robot_info r
LEFT JOIN failure_stats f ON r.id = f.robot_id
ORDER BY r.id, f.count DESC NULLS LAST;
```

---

## 七、常见问题FAQ ❓

### Q1：数据库升级后，旧的执行日志会有新字段吗？

**A：** 不会。`failure_category` 和 `failure_reason` 字段只会在新记录中填充。旧记录这两个字段为NULL。

### Q2：如何批量开启所有OKX机器人的自动交易？

**A：**
```sql
UPDATE hg_trading_robot 
SET auto_trade_enabled = 1 
WHERE platform = 'okx' AND status = 2;
```

### Q3：如何查看某个机器人为什么一直不下单？

**A：**
```sql
-- 查看最近的失败日志
SELECT 
    failure_category,
    failure_reason,
    created_at
FROM hg_trading_execution_log
WHERE robot_id = <robot_id>
  AND status = 'failed'
ORDER BY created_at DESC
LIMIT 10;
```

### Q4：部署后服务无法启动怎么办？

**A：**
1. 查看日志文件：`logs/server/latest.log`
2. 检查数据库连接是否正常
3. 检查端口是否被占用
4. 回滚到备份版本

### Q5：如何回滚数据库升级？

**A：**
```sql
-- 删除新增的字段（谨慎操作）
ALTER TABLE hg_trading_execution_log DROP COLUMN IF EXISTS failure_category;
ALTER TABLE hg_trading_execution_log DROP COLUMN IF EXISTS failure_reason;

-- 删除新增的索引
DROP INDEX IF EXISTS idx_failure_category;
DROP INDEX IF EXISTS idx_robot_status_category;
```

---

## 八、总结 📝

### 部署检查清单

- [ ] 数据库已备份
- [ ] 执行诊断SQL，了解当前问题
- [ ] 执行数据库升级脚本
- [ ] 验证字段已添加
- [ ] 编译新版本代码
- [ ] 停止旧服务
- [ ] 启动新服务
- [ ] 验证部署效果
- [ ] 监控执行日志

### 关键文件

| 文件 | 用途 |
|------|------|
| `diagnose_okx_robot.sql` | 诊断SQL脚本 |
| `add_failure_category_fields.sql` | 数据库升级脚本 |
| `OKX机器人下单问题诊断和优化方案.md` | 详细技术方案 |
| `OKX机器人下单问题优化完成总结.md` | 实施总结 |
| `OKX机器人诊断和部署指南.md` | 本文档 |

### 技术支持

如遇到问题，请提供以下信息：
1. 诊断SQL的执行结果
2. 失败日志的详细内容
3. 机器人配置信息
4. 服务日志文件

---

**部署指南完成 ✅**

按照本指南操作，您可以快速诊断和解决OKX机器人的下单问题！

