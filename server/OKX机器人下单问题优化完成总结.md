# OKX机器人下单问题优化完成总结

> 完成时间：2024-12-24
> 
> 问题：OKX运行中的机器人有预警记录和执行日志，但没有实际下单
> 
> 解决：优化执行日志，增加失败分类和结构化失败原因

---

## 一、已完成的优化 ✅

### 1.1 数据库优化

**文件：** `add_failure_category_fields.sql`

**新增字段：**
```sql
ALTER TABLE hg_trading_execution_log 
ADD COLUMN IF NOT EXISTS failure_category VARCHAR(50) DEFAULT NULL;

ALTER TABLE hg_trading_execution_log 
ADD COLUMN IF NOT EXISTS failure_reason TEXT DEFAULT NULL;
```

**新增索引：**
```sql
CREATE INDEX IF NOT EXISTS idx_failure_category 
ON hg_trading_execution_log(failure_category, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_robot_status_category 
ON hg_trading_execution_log(robot_id, status, failure_category, created_at DESC);
```

### 1.2 代码优化

**文件：** `internal/logic/toogo/robot_engine.go`

**优化内容：**

1. **增强 `saveExecutionLog()` 方法**
   - 自动分析失败原因
   - 保存失败分类（`failure_category`）
   - 保存结构化失败原因（`failure_reason`）

2. **新增 `analyzeFailureReason()` 方法**
   - 根据 `step` 和 `eventData` 智能分析失败原因
   - 提取关键信息（持仓状态、余额、配置等）
   - 生成用户友好的错误说明和解决建议

3. **新增辅助方法**
   - `translatePositionSide()`: 翻译持仓方向（英文→中文）
   - `translateOppositePositionSide()`: 获取反向持仓方向
   - `formatExchangeAPIError()`: 格式化交易所API错误

### 1.3 失败分类定义

| 分类 | 说明 | 常见原因 | 解决方案 |
|------|------|----------|----------|
| **config** | 配置问题 | 自动交易未开启 | 开启自动交易开关 |
| **balance** | 余额问题 | 余额不足、无法获取余额 | 充值或降低保证金比例 |
| **position** | 持仓问题 | 已有持仓、持仓模式限制 | 等待平仓或切换持仓模式 |
| **exchange** | 交易所API问题 | API调用失败、订单被拒绝 | 检查网络、余额、杠杆 |
| **strategy** | 策略问题 | 策略参数缺失、映射关系错误 | 检查策略配置 |
| **system** | 系统问题 | 获取锁超时、行情服务未就绪 | 稍后再试或联系技术支持 |

---

## 二、诊断工具 🔧

### 2.1 诊断SQL脚本

**文件：** `diagnose_okx_robot.sql`

**关键查询：**

1. 查询OKX运行中的机器人配置
2. 查询预警记录状态
3. 统计执行日志失败原因
4. 查询详细失败步骤
5. 查询订单状态
6. 查询预警记录和执行日志的关联

### 2.2 使用方法

```bash
# 方法1：在PostgreSQL中执行
psql -U postgres -d hotgo -f diagnose_okx_robot.sql

# 方法2：通过数据库工具执行
# 打开pgAdmin、DBeaver等工具，加载并执行 diagnose_okx_robot.sql
```

---

## 三、部署步骤 📋

### 3.1 数据库更新（必须先执行）

```bash
# 1. 备份数据库（重要！）
pg_dump -U postgres -d hotgo > backup_$(date +%Y%m%d_%H%M%S).sql

# 2. 执行数据库升级脚本
psql -U postgres -d hotgo -f add_failure_category_fields.sql

# 3. 验证字段已添加
psql -U postgres -d hotgo -c "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'hg_trading_execution_log' AND column_name IN ('failure_category', 'failure_reason');"
```

### 3.2 代码部署

```bash
# 1. 停止服务
systemctl stop hotgo

# 2. 备份当前版本
cp /path/to/hotgo /path/to/hotgo.backup

# 3. 编译新版本
cd D:\go\src\hotgo_v2\server
go build -o hotgo main.go

# 4. 启动服务
systemctl start hotgo

# 5. 查看日志
tail -f /var/log/hotgo/app.log
```

### 3.3 验证部署

```sql
-- 1. 触发一个失败场景（例如关闭自动交易）
UPDATE hg_trading_robot SET auto_trade_enabled = 0 WHERE id = <robot_id>;

-- 2. 等待信号触发（或手动触发）

-- 3. 查看执行日志
SELECT 
    id, event_type, status, message,
    failure_category, failure_reason,
    created_at
FROM hg_trading_execution_log
WHERE robot_id = <robot_id>
ORDER BY created_at DESC LIMIT 5;

-- 4. 验证failure_category和failure_reason是否有值
-- 预期：failure_category='config', failure_reason='自动交易开关未开启...'
```

---

## 四、常见问题排查 🔍

### 4.1 快速诊断命令

```sql
-- 1. 查看OKX机器人配置
SELECT 
    id, robot_name, 
    auto_trade_enabled,  -- 0=未开启，1=已开启
    auto_close_enabled,
    dual_side_position   -- 0=单向，1=双向
FROM hg_trading_robot
WHERE platform = 'okx' AND status = 2;

-- 2. 查看最近失败原因统计
SELECT 
    failure_category,
    COUNT(*) AS count,
    array_agg(DISTINCT failure_reason) AS reasons
FROM hg_trading_execution_log
WHERE robot_id IN (SELECT id FROM hg_trading_robot WHERE platform = 'okx' AND status = 2)
  AND status = 'failed'
  AND created_at > NOW() - INTERVAL '1 hour'
GROUP BY failure_category
ORDER BY count DESC;

-- 3. 查看最近10条失败日志
SELECT 
    id,
    robot_id,
    failure_category,
    failure_reason,
    created_at
FROM hg_trading_execution_log
WHERE robot_id IN (SELECT id FROM hg_trading_robot WHERE platform = 'okx' AND status = 2)
  AND status = 'failed'
ORDER BY created_at DESC
LIMIT 10;
```

### 4.2 常见问题快速修复

#### 问题1：自动交易未开启（failure_category='config'）

```sql
-- 开启自动交易
UPDATE hg_trading_robot 
SET auto_trade_enabled = 1 
WHERE id = <robot_id>;
```

#### 问题2：单向持仓模式限制（failure_category='position'）

```sql
-- 方案A：切换到双向持仓模式
UPDATE hg_trading_robot 
SET dual_side_position = 1 
WHERE id = <robot_id>;

-- 方案B：查看当前持仓，手动平仓
SELECT * FROM hg_trading_order 
WHERE robot_id = <robot_id> AND status = 1;
```

#### 问题3：策略配置缺失（failure_category='strategy'）

```sql
-- 查看机器人的风险配置映射
SELECT id, robot_name, remark 
FROM hg_trading_robot 
WHERE id = <robot_id>;

-- remark字段应包含类似：
-- {"high_vol":"aggressive","low_vol":"conservative","trend":"balanced"}

-- 如果为空或格式错误，需要重新配置
UPDATE hg_trading_robot 
SET remark = '{"high_vol":"aggressive","low_vol":"conservative","trend":"balanced"}'
WHERE id = <robot_id>;
```

#### 问题4：余额不足（failure_category='balance'）

```sql
-- 查看当前余额配置
SELECT 
    r.id, r.robot_name,
    s.margin_percent_min, s.margin_percent_max,
    s.leverage_min, s.leverage_max
FROM hg_trading_robot r
JOIN hg_trading_strategy_template s ON r.strategy_id = s.id
WHERE r.id = <robot_id>;

-- 降低保证金比例（临时方案）
UPDATE hg_trading_strategy_template 
SET margin_percent_min = 5, margin_percent_max = 10
WHERE id = (SELECT strategy_id FROM hg_trading_robot WHERE id = <robot_id>);
```

---

## 五、前端展示建议 💡

### 5.1 查询API

```go
// API接口：查询执行日志
func GetExecutionLogs(robotId int64, limit int) ([]*ExecutionLog, error) {
    var logs []*ExecutionLog
    err := g.DB().Model("hg_trading_execution_log").
        Where("robot_id", robotId).
        Order("created_at DESC").
        Limit(limit).
        Scan(&logs)
    return logs, err
}

// API接口：按分类统计失败次数
func GetFailureStatistics(robotId int64, hours int) (map[string]int, error) {
    var results []struct {
        FailureCategory string `json:"failure_category"`
        Count          int    `json:"count"`
    }
    
    err := g.DB().Model("hg_trading_execution_log").
        Where("robot_id", robotId).
        Where("status", "failed").
        Where("created_at > ?", time.Now().Add(-time.Duration(hours)*time.Hour)).
        Group("failure_category").
        Fields("failure_category", "COUNT(*) as count").
        Scan(&results)
    
    stats := make(map[string]int)
    for _, r := range results {
        stats[r.FailureCategory] = r.Count
    }
    return stats, err
}
```

### 5.2 前端展示组件

```typescript
// 失败分类颜色映射
const categoryColors = {
  config: '#FF9800',    // 橙色 - 配置问题
  balance: '#F44336',   // 红色 - 余额问题
  position: '#2196F3',  // 蓝色 - 持仓问题
  exchange: '#9C27B0',  // 紫色 - 交易所问题
  strategy: '#00BCD4',  // 青色 - 策略问题
  system: '#607D8B',    // 灰色 - 系统问题
};

// 失败分类中文名称
const categoryNames = {
  config: '配置问题',
  balance: '余额问题',
  position: '持仓问题',
  exchange: '交易所问题',
  strategy: '策略问题',
  system: '系统问题',
};

// 展示组件
interface ExecutionLog {
  id: number;
  eventType: string;
  status: string;
  message: string;
  failureCategory?: string;
  failureReason?: string;
  createdAt: string;
}

function ExecutionLogItem({ log }: { log: ExecutionLog }) {
  // 失败日志
  if (log.status === 'failed' && log.failureCategory) {
    return (
      <div className="execution-log-item failed">
        <div 
          className="category-tag" 
          style={{ backgroundColor: categoryColors[log.failureCategory] }}
        >
          {categoryNames[log.failureCategory]}
        </div>
        <div className="reason">
          {log.failureReason || log.message}
        </div>
        <div className="timestamp">
          {formatTime(log.createdAt)}
        </div>
      </div>
    );
  }
  
  // 成功日志
  return (
    <div className="execution-log-item success">
      <div className="status-tag">成功</div>
      <div className="message">{log.message}</div>
      <div className="timestamp">{formatTime(log.createdAt)}</div>
    </div>
  );
}

// 失败统计图表
function FailureStatistics({ stats }: { stats: Record<string, number> }) {
  const data = Object.entries(stats).map(([category, count]) => ({
    category: categoryNames[category],
    count,
    color: categoryColors[category],
  }));
  
  return (
    <div className="failure-statistics">
      <h3>失败原因统计（最近24小时）</h3>
      <div className="chart">
        {data.map(item => (
          <div key={item.category} className="chart-item">
            <div 
              className="bar" 
              style={{ 
                width: `${(item.count / Math.max(...data.map(d => d.count))) * 100}%`,
                backgroundColor: item.color 
              }}
            />
            <span className="label">{item.category}</span>
            <span className="count">{item.count}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
```

---

## 六、测试用例 ✅

### 6.1 配置问题测试

```sql
-- 测试：自动交易未开启
UPDATE hg_trading_robot SET auto_trade_enabled = 0 WHERE id = 1;
-- 触发信号，观察执行日志
-- 预期：failure_category='config', failure_reason='自动交易开关未开启...'
```

### 6.2 持仓问题测试

```sql
-- 测试：单向持仓模式限制
UPDATE hg_trading_robot SET dual_side_position = 0 WHERE id = 1;
-- 在有持仓的情况下触发同方向信号
-- 预期：failure_category='position', failure_reason='单向持仓模式限制...'

-- 测试：双向持仓同方向限制
UPDATE hg_trading_robot SET dual_side_position = 1 WHERE id = 1;
-- 在有多头持仓时触发多头信号
-- 预期：failure_category='position', failure_reason='双向持仓模式限制...'
```

### 6.3 余额问题测试

```sql
-- 测试：余额不足
UPDATE hg_trading_strategy_template 
SET margin_percent_min = 100 
WHERE id = (SELECT strategy_id FROM hg_trading_robot WHERE id = 1);
-- 触发信号
-- 预期：failure_category='balance', failure_reason='账户余额不足...'
```

### 6.4 策略问题测试

```sql
-- 测试：策略配置缺失
UPDATE hg_trading_robot SET remark = NULL WHERE id = 1;
-- 或设置错误的映射关系
UPDATE hg_trading_robot SET remark = '{}' WHERE id = 1;
-- 触发信号
-- 预期：failure_category='strategy', failure_reason='策略配置缺失...'
```

---

## 七、效果对比 📊

### 7.1 优化前

**执行日志：**
```json
{
  "id": 12345,
  "event_type": "order_failed",
  "status": "failed",
  "message": "自动下单未开启",
  "event_data": "{\"step\":\"auto_trade_check\",\"autoTradeEnabled\":0}",
  "failure_category": null,
  "failure_reason": null
}
```

**问题：**
- ❌ 失败原因隐藏在JSON中
- ❌ 前端需要解析event_data
- ❌ 无法分类统计
- ❌ 用户难以理解

### 7.2 优化后

**执行日志：**
```json
{
  "id": 12345,
  "event_type": "order_failed",
  "status": "failed",
  "message": "自动下单未开启",
  "event_data": "{\"step\":\"auto_trade_check\",\"autoTradeEnabled\":0}",
  "failure_category": "config",
  "failure_reason": "自动交易开关未开启。解决方案：在机器人设置中开启自动交易开关"
}
```

**改进：**
- ✅ 失败分类清晰（config）
- ✅ 结构化失败原因（包含解决方案）
- ✅ 前端可以直接展示
- ✅ 支持分类统计
- ✅ 用户友好的说明

---

## 八、总结 📝

### 8.1 优化成果

✅ **数据库优化**
- 增加 `failure_category` 字段（失败分类）
- 增加 `failure_reason` 字段（结构化失败原因）
- 创建索引提升查询性能

✅ **代码优化**
- 智能分析失败原因
- 生成用户友好的错误说明
- 提供具体的解决建议

✅ **诊断工具**
- 完整的诊断SQL脚本
- 快速定位问题原因
- 提供修复方案

### 8.2 下一步

1. **立即诊断**
   - 执行 `diagnose_okx_robot.sql`
   - 查看失败原因统计
   - 针对性解决问题

2. **部署更新**
   - 执行数据库升级脚本
   - 部署优化后的代码
   - 验证效果

3. **前端优化**
   - 使用新增的字段展示失败原因
   - 增加分类统计图表
   - 提升用户体验

---

**优化完成 ✅**

通过这套优化方案，您可以：
1. 快速定位OKX机器人没有下单的具体原因
2. 获得清晰、结构化的失败原因说明
3. 得到具体的解决建议
4. 大幅提升系统可维护性和用户体验

