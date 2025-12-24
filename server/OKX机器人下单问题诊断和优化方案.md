# OKX机器人下单问题诊断和优化方案

> 生成时间：2024-12-24
> 
> 问题：OKX运行中的机器人有预警记录和大量执行日志，但没有实际下单

---

## 一、问题诊断步骤 🔍

### 1.1 查询诊断SQL

已创建诊断SQL文件：`diagnose_okx_robot.sql`

**关键查询：**

```sql
-- 1. 查询OKX运行中的机器人配置
SELECT 
    id, robot_name, symbol, platform, status,
    auto_trade_enabled,  -- 【关键】自动交易开关
    auto_close_enabled,
    dual_side_position   -- 【关键】持仓模式
FROM hg_trading_robot
WHERE platform = 'okx' AND status = 2 AND deleted_at IS NULL;

-- 2. 查询预警记录状态
SELECT 
    id, robot_id, direction, action,
    is_processed,  -- 【关键】是否已处理
    executed,      -- 【关键】是否已执行
    execute_result,
    created_at
FROM hg_trading_signal_log
WHERE robot_id IN (SELECT id FROM hg_trading_robot WHERE platform = 'okx' AND status = 2)
  AND created_at > NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC LIMIT 50;

-- 3. 查询执行日志失败原因
SELECT 
    message AS failure_reason,
    COUNT(*) AS count,
    MAX(created_at) AS last_occurrence
FROM hg_trading_execution_log
WHERE robot_id IN (SELECT id FROM hg_trading_robot WHERE platform = 'okx' AND status = 2)
  AND status = 'failed'
  AND created_at > NOW() - INTERVAL '24 hours'
GROUP BY message
ORDER BY count DESC;

-- 4. 查询执行日志详细失败步骤
SELECT 
    id, robot_id, signal_log_id,
    event_type, status, message,
    event_data::jsonb->>'step' AS failure_step,
    event_data::jsonb->>'autoTradeEnabled' AS auto_trade_enabled,
    event_data::jsonb->>'dualSidePosition' AS dual_side_position,
    created_at
FROM hg_trading_execution_log
WHERE robot_id IN (SELECT id FROM hg_trading_robot WHERE platform = 'okx' AND status = 2)
  AND status = 'failed'
ORDER BY created_at DESC LIMIT 50;
```

### 1.2 常见失败原因及解决方案

#### ❌ 原因1：自动交易未开启

**症状：**
```
message: "自动下单未开启"
step: "auto_trade_check"
autoTradeEnabled: 0
```

**解决方案：**
```sql
-- 检查自动交易开关
SELECT id, robot_name, auto_trade_enabled 
FROM hg_trading_robot 
WHERE platform = 'okx' AND status = 2;

-- 开启自动交易
UPDATE hg_trading_robot 
SET auto_trade_enabled = 1 
WHERE platform = 'okx' AND status = 2;
```

#### ❌ 原因2：单向持仓模式限制

**症状：**
```
message: "单向持仓模式：已有持仓（多头），持仓内只能有一单，拒绝新开仓（目标=多头）"
step: "single_position_check" 或 "single_position_check_after_lock"
dualSidePosition: 0
```

**解决方案：**
```sql
-- 方案A：切换到双向持仓模式
UPDATE hg_trading_robot 
SET dual_side_position = 1 
WHERE platform = 'okx' AND status = 2;

-- 方案B：先平仓现有持仓，再下新单
-- 需要手动平仓或等待自动平仓
```

#### ❌ 原因3：双向持仓同方向限制

**症状：**
```
message: "双向持仓模式：多头方向已有持仓，同方向只能一单（禁止加仓），拒绝新开仓"
step: "dual_side_same_direction_check" 或 "dual_side_same_direction_check_after_lock"
positionSide: "LONG"
```

**解决方案：**
- 等待当前持仓平仓后，才能开新的同方向仓位
- 或者开反方向的仓位（如果是双向模式）

#### ❌ 原因4：防重复下单机制

**症状：**
- 预警记录的 `is_processed = 1`
- 但没有对应的订单记录

**日志信息：**
```
"预警记录logId=xxx已被其他goroutine处理（is_processed=1），跳过重复下单"
```

**可能原因：**
- 并发触发导致重复标记
- 下单失败但已标记为已处理

**解决方案：**
```sql
-- 重置已处理标记（谨慎操作）
UPDATE hg_trading_signal_log 
SET is_processed = 0 
WHERE id = <signal_log_id>;
```

#### ❌ 原因5：余额不足

**症状：**
```
message: "余额不足（交易所余额为0或负数）" 或 "余额不足（无法获取余额: xxx）"
step: "balance_check"
available_balance: 0 或 负数
```

**解决方案：**
- 检查交易所账户余额
- 充值或调整保证金比例

#### ❌ 原因6：策略参数获取失败

**症状：**
```
message: "获取策略参数失败: 市场状态=xxx 在映射关系中未找到对应的风险偏好"
step: "strategy_params"
```

**解决方案：**
```sql
-- 检查机器人的风险配置映射
SELECT id, robot_name, remark 
FROM hg_trading_robot 
WHERE platform = 'okx' AND status = 2;

-- remark字段应包含映射关系，如：
-- {"high_vol":"aggressive","low_vol":"conservative","trend":"balanced"}
```

#### ❌ 原因7：交易所API下单失败

**症状：**
```
message: "交易所下单失败: xxx"
step: "exchange_api"
error: "具体API错误信息"
```

**常见API错误：**
- `-1021 INVALID_TIMESTAMP`: 时间戳错误，检查服务器时间同步
- `-2010 NEW_ORDER_REJECTED`: 订单被拒绝，检查杠杆、数量、余额
- `-2015 INVALID_ORDER`: 无效订单，检查订单参数
- `-2019 MARGIN_NOT_SUFFICENT`: 保证金不足

---

## 二、优化方案：改进执行日志展示 ✨

### 2.1 当前问题

1. **message字段信息不够结构化**
   - 前端难以解析和分类显示
   - 用户难以快速找到失败原因

2. **event_data是JSON字符串**
   - 前端需要解析JSON
   - 关键字段可能被隐藏

3. **缺少失败原因分类**
   - 所有失败都用相同的 `event_type: "order_failed"`
   - 无法区分不同类型的失败

### 2.2 优化方案

#### 优化1：增加失败原因分类字段

**数据库表结构优化：**

```sql
-- 增加失败原因分类字段
ALTER TABLE hg_trading_execution_log 
ADD COLUMN IF NOT EXISTS failure_category VARCHAR(50) DEFAULT NULL 
COMMENT '失败分类：config/balance/position/exchange/strategy/system';

ALTER TABLE hg_trading_execution_log 
ADD COLUMN IF NOT EXISTS failure_reason TEXT DEFAULT NULL 
COMMENT '失败原因详情（结构化文本）';

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_failure_category 
ON hg_trading_execution_log(failure_category, created_at);
```

**失败分类定义：**

| 分类 | 说明 | 示例 |
|------|------|------|
| `config` | 配置问题 | 自动交易未开启、策略参数缺失 |
| `balance` | 余额问题 | 余额不足、无法获取余额 |
| `position` | 持仓问题 | 已有持仓、持仓模式限制 |
| `exchange` | 交易所API问题 | API调用失败、订单被拒绝 |
| `strategy` | 策略问题 | 策略参数获取失败、市场状态映射缺失 |
| `system` | 系统问题 | 获取锁超时、系统繁忙 |

#### 优化2：改进saveExecutionLog方法

**文件：** `robot_engine.go`

```go
// saveExecutionLog 保存交易执行日志（记录完整的交易执行过程）
// 【优化】增加失败分类和结构化失败原因
func (t *RobotTrader) saveExecutionLog(ctx context.Context, signalLogId int64, orderId int64, eventType string, status string, message string, eventData map[string]interface{}) {
	robot := t.engine.Robot
	if robot == nil {
		return
	}

	// 序列化事件数据为JSON
	eventDataJSON := "{}"
	if len(eventData) > 0 {
		data, err := json.Marshal(eventData)
		if err == nil {
			eventDataJSON = string(data)
		}
	}

	// 【新增】分析失败原因，提取分类和详情
	failureCategory, failureReason := t.analyzeFailureReason(eventType, message, eventData)

	// 写入交易日志
	_, err := g.DB().Model("hg_trading_execution_log").Ctx(ctx).Insert(g.Map{
		"signal_log_id":    signalLogId,
		"robot_id":         robot.Id,
		"order_id":         orderId,
		"event_type":       eventType,
		"event_data":       eventDataJSON,
		"status":           status,
		"message":          message,
		"failure_category": failureCategory, // 【新增】
		"failure_reason":   failureReason,   // 【新增】
		"created_at":       time.Now(),
	})
	if err != nil {
		g.Log().Warningf(ctx, "[RobotTrader] 保存交易日志失败: robotId=%d, eventType=%s, err=%v", robot.Id, eventType, err)
	} else {
		g.Log().Debugf(ctx, "[RobotTrader] 交易日志已保存: robotId=%d, eventType=%s, status=%s, category=%s", robot.Id, eventType, status, failureCategory)
	}

	// 兼容前端"信号日志/执行结果"展示
	if signalLogId > 0 && (eventType == "order_failed" || eventType == "order_success") {
		result := message
		if len(result) > 200 {
			result = result[:200]
		}
		_, _ = g.DB().Model("hg_trading_signal_log").Ctx(ctx).
			Where("id", signalLogId).
			Data(g.Map{
				"executed":       1,
				"execute_result": result,
			}).
			Update()
	}
}

// analyzeFailureReason 分析失败原因，提取分类和详情
// 【新增】自动分析失败原因，便于前端展示
func (t *RobotTrader) analyzeFailureReason(eventType string, message string, eventData map[string]interface{}) (category string, reason string) {
	// 只处理失败事件
	if eventType != "order_failed" {
		return "", ""
	}

	step, _ := eventData["step"].(string)
	
	switch step {
	case "robot_check", "signal_check":
		category = "system"
		reason = formatFailureReason("系统检查", message, eventData)
		
	case "auto_trade_check":
		category = "config"
		autoTradeEnabled, _ := eventData["autoTradeEnabled"].(int)
		if autoTradeEnabled == 0 {
			reason = "自动交易开关未开启，请在机器人设置中开启自动交易"
		} else {
			reason = formatFailureReason("自动交易检查", message, eventData)
		}
		
	case "position_check", "single_position_check", "dual_side_same_direction_check",
		 "single_position_check_after_lock", "dual_side_same_direction_check_after_lock":
		category = "position"
		dualSidePosition, _ := eventData["dualSidePosition"].(int)
		positionSide, _ := eventData["positionSide"].(string)
		existingPositionSide, _ := eventData["existingPositionSide"].(string)
		
		if dualSidePosition == 0 {
			// 单向持仓模式
			reason = fmt.Sprintf("单向持仓模式限制：当前已有%s方向持仓，持仓内只能有一单。建议：1) 等待当前持仓平仓后再下单，或 2) 切换到双向持仓模式", 
				translatePositionSide(existingPositionSide))
		} else {
			// 双向持仓模式
			reason = fmt.Sprintf("双向持仓模式限制：%s方向已有持仓，同方向不允许加仓。建议：1) 等待当前%s持仓平仓后再下单，或 2) 开反方向的%s仓位",
				translatePositionSide(positionSide),
				translatePositionSide(positionSide),
				translateOppositePositionSide(positionSide))
		}
		
	case "balance_check":
		category = "balance"
		availableBalance, _ := eventData["available_balance"].(float64)
		if availableBalance <= 0 {
			reason = "账户余额不足或为0。请：1) 充值到交易所账户，或 2) 降低保证金比例"
		} else {
			reason = formatFailureReason("余额检查", message, eventData)
		}
		
	case "ticker_check":
		category = "system"
		reason = "无法获取实时行情数据。请检查：1) 网络连接是否正常，2) WebSocket服务是否运行"
		
	case "strategy_params":
		category = "strategy"
		errorMsg, _ := eventData["error"].(string)
		if strings.Contains(errorMsg, "未找到对应的风险偏好") {
			reason = "策略配置缺失：市场状态与风险偏好映射关系未配置。请：1) 检查机器人的风险配置映射，2) 重新创建机器人并设置完整的映射关系"
		} else {
			reason = fmt.Sprintf("策略参数获取失败：%s。请检查策略模板配置是否完整", errorMsg)
		}
		
	case "pre_create_order":
		category = "system"
		reason = formatFailureReason("预创建订单", message, eventData)
		
	case "exchange_api":
		category = "exchange"
		errorMsg, _ := eventData["error"].(string)
		reason = formatExchangeAPIError(errorMsg)
		
	case "order_status_update":
		category = "system"
		reason = formatFailureReason("订单状态更新", message, eventData)
		
	case "lock_acquire":
		category = "system"
		reason = "系统繁忙，无法获取下单锁。建议：稍后再试或联系技术支持"
		
	default:
		category = "system"
		reason = formatFailureReason("未知错误", message, eventData)
	}
	
	return category, reason
}

// translatePositionSide 翻译持仓方向
func translatePositionSide(positionSide string) string {
	switch positionSide {
	case "LONG":
		return "多头"
	case "SHORT":
		return "空头"
	default:
		return positionSide
	}
}

// translateOppositePositionSide 获取反向持仓方向
func translateOppositePositionSide(positionSide string) string {
	switch positionSide {
	case "LONG":
		return "空头"
	case "SHORT":
		return "多头"
	default:
		return positionSide
	}
}

// formatFailureReason 格式化失败原因
func formatFailureReason(context string, message string, eventData map[string]interface{}) string {
	return fmt.Sprintf("%s失败：%s", context, message)
}

// formatExchangeAPIError 格式化交易所API错误
func formatExchangeAPIError(errorMsg string) string {
	// 常见错误码映射
	errorMappings := map[string]string{
		"-1021": "时间戳错误，请检查服务器时间同步",
		"-2010": "订单被交易所拒绝，请检查：1) 账户余额是否充足，2) 杠杆设置是否正确，3) 订单数量是否符合要求",
		"-2015": "无效订单参数，请检查订单配置",
		"-2019": "保证金不足，请充值或降低杠杆倍数",
		"insufficient balance": "余额不足，请充值",
		"position not found": "持仓不存在，可能已被平仓",
	}
	
	// 查找匹配的错误码
	for code, description := range errorMappings {
		if strings.Contains(errorMsg, code) {
			return fmt.Sprintf("交易所API错误 [%s]：%s", code, description)
		}
	}
	
	// 未匹配到具体错误码，返回原始错误信息
	return fmt.Sprintf("交易所API错误：%s", errorMsg)
}
```

#### 优化3：前端展示优化

**查询API优化：**

```sql
-- 优化后的查询（包含失败分类）
SELECT 
    id,
    signal_log_id,
    robot_id,
    order_id,
    event_type,
    status,
    message,
    failure_category,  -- 【新增】失败分类
    failure_reason,    -- 【新增】结构化失败原因
    created_at
FROM hg_trading_execution_log
WHERE robot_id = ?
  AND status = 'failed'
ORDER BY created_at DESC
LIMIT 100;

-- 按分类统计失败次数
SELECT 
    failure_category,
    COUNT(*) AS count,
    MAX(created_at) AS last_occurrence
FROM hg_trading_execution_log
WHERE robot_id = ?
  AND status = 'failed'
  AND created_at > NOW() - INTERVAL '24 hours'
GROUP BY failure_category
ORDER BY count DESC;
```

**前端展示建议：**

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

// 失败分类图标映射
const categoryIcons = {
  config: 'settings',
  balance: 'account_balance_wallet',
  position: 'trending_up',
  exchange: 'swap_horiz',
  strategy: 'analytics',
  system: 'error',
};

// 展示失败日志
function renderFailureLog(log) {
  return (
    <div className="failure-log-item">
      <div className="category-tag" style={{ backgroundColor: categoryColors[log.failure_category] }}>
        <Icon name={categoryIcons[log.failure_category]} />
        {translateCategory(log.failure_category)}
      </div>
      <div className="failure-reason">
        {log.failure_reason || log.message}
      </div>
      <div className="timestamp">
        {formatTime(log.created_at)}
      </div>
    </div>
  );
}
```

---

## 三、实施步骤 📋

### 3.1 立即诊断（第一步）

1. **执行诊断SQL**
   ```bash
   psql -U postgres -d hotgo -f diagnose_okx_robot.sql > diagnosis_result.txt
   ```

2. **查看失败原因统计**
   - 重点关注 `failure_reason` 统计结果
   - 确认最常见的失败原因

3. **针对性解决**
   - 根据诊断结果，采取对应的解决方案
   - 优先解决出现频率最高的问题

### 3.2 数据库优化（第二步）

```sql
-- 1. 增加失败分类字段
ALTER TABLE hg_trading_execution_log 
ADD COLUMN IF NOT EXISTS failure_category VARCHAR(50) DEFAULT NULL;

ALTER TABLE hg_trading_execution_log 
ADD COLUMN IF NOT EXISTS failure_reason TEXT DEFAULT NULL;

-- 2. 创建索引
CREATE INDEX IF NOT EXISTS idx_failure_category 
ON hg_trading_execution_log(failure_category, created_at);

-- 3. 验证字段已添加
SELECT column_name, data_type, character_maximum_length
FROM information_schema.columns
WHERE table_name = 'hg_trading_execution_log'
  AND column_name IN ('failure_category', 'failure_reason');
```

### 3.3 代码优化（第三步）

1. **修改 `saveExecutionLog` 方法**
   - 增加 `analyzeFailureReason` 调用
   - 保存失败分类和详情

2. **添加辅助方法**
   - `analyzeFailureReason`
   - `translatePositionSide`
   - `formatExchangeAPIError`

3. **测试验证**
   - 触发各种失败场景
   - 验证日志记录是否正确

### 3.4 前端展示优化（第四步）

1. **更新API接口**
   - 返回 `failure_category` 和 `failure_reason`

2. **优化前端展示**
   - 分类标签展示
   - 颜色和图标映射
   - 结构化失败原因展示

---

## 四、验证和测试 ✅

### 4.1 测试场景

**场景1：自动交易未开启**
```sql
-- 关闭自动交易
UPDATE hg_trading_robot SET auto_trade_enabled = 0 WHERE id = <robot_id>;

-- 触发信号，观察执行日志
-- 预期：failure_category='config', failure_reason='自动交易开关未开启...'
```

**场景2：单向持仓模式限制**
```sql
-- 设置为单向模式
UPDATE hg_trading_robot SET dual_side_position = 0 WHERE id = <robot_id>;

-- 在有持仓的情况下触发同方向信号
-- 预期：failure_category='position', failure_reason='单向持仓模式限制...'
```

**场景3：余额不足**
```sql
-- 设置极高的保证金比例
UPDATE hg_trading_strategy_template SET margin_percent_min = 100 WHERE id = <strategy_id>;

-- 触发信号
-- 预期：failure_category='balance', failure_reason='账户余额不足...'
```

### 4.2 验证查询

```sql
-- 验证失败分类是否正确记录
SELECT 
    failure_category,
    COUNT(*) AS count,
    array_agg(DISTINCT failure_reason) AS reasons
FROM hg_trading_execution_log
WHERE robot_id = <robot_id>
  AND status = 'failed'
  AND created_at > NOW() - INTERVAL '1 hour'
GROUP BY failure_category;

-- 查看最新的失败日志详情
SELECT 
    id,
    event_type,
    status,
    message,
    failure_category,
    failure_reason,
    event_data,
    created_at
FROM hg_trading_execution_log
WHERE robot_id = <robot_id>
  AND status = 'failed'
ORDER BY created_at DESC
LIMIT 10;
```

---

## 五、总结 📝

### 5.1 优化效果

**优化前：**
- ❌ 失败原因隐藏在JSON中，难以查看
- ❌ 前端需要解析复杂的event_data
- ❌ 无法快速定位问题类型
- ❌ 用户难以理解失败原因

**优化后：**
- ✅ 失败原因分类清晰（6大类）
- ✅ 结构化的失败原因说明
- ✅ 前端可以直接展示分类和原因
- ✅ 提供具体的解决建议
- ✅ 支持按分类统计和过滤

### 5.2 关键改进

1. **新增失败分类字段**
   - `failure_category`: 6大分类（config/balance/position/exchange/strategy/system）
   - `failure_reason`: 结构化的失败原因和解决建议

2. **智能分析失败原因**
   - 根据 `step` 和 `eventData` 自动分析
   - 提取关键信息（如已有持仓方向、余额、错误码）
   - 生成用户友好的失败说明

3. **前端展示优化**
   - 颜色标签区分失败类型
   - 图标可视化
   - 提供具体的解决建议

### 5.3 下一步

1. **立即诊断**：执行诊断SQL，找出OKX机器人的具体失败原因
2. **数据库优化**：添加失败分类字段
3. **代码优化**：实现智能失败原因分析
4. **前端优化**：改进失败日志展示

---

**优化方案完成 ✅**

通过这套优化方案，您可以：
1. 快速定位OKX机器人没有下单的具体原因
2. 获得清晰、结构化的失败原因说明
3. 得到具体的解决建议
4. 提升用户体验和系统可维护性

