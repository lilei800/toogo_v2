# 策略模板系统优化报告

## 📅 优化日期: 2024-11-29
## 🔄 最后更新: 2024-11-29 (支持完整手动配置)

---

## 📋 手动策略配置参数

| 参数 | 字段 | 说明 |
|------|------|------|
| 交易平台 | `exchange` | bitget/binance/okx/gateio |
| 交易对 | `symbol` | BTC-USDT 等 |
| 订单类型 | `orderType` | market(市价)/limit(限价) |
| 保证金模式 | `marginMode` | isolated(逐仓)/cross(全仓) |
| 时间窗口 | `monitorWindow` | 监控时间窗口(秒) |
| 波动值 | `volatilityThreshold` | 触发交易的波动阈值(USDT) |
| 杠杆倍数 | `leverage` | 推荐杠杆倍数 |
| 保证金比例 | `marginPercent` | 推荐仓位比例(%) |
| 止损百分比 | `stopLossPercent` | 止损触发比例(%) |
| 启动回撤百分比 | `autoStartRetreatPercent` | 启动止盈回撤的盈利比例(%) |
| 止盈回撤百分比 | `profitRetreatPercent` | 止盈回撤触发比例(%) |

---

## ✅ 已完成的优化

### 1. 创建官方 BTC-USDT 策略模板

**文件**: `server/storage/data/official_btc_usdt_strategies.sql`

创建了一套完整的官方 BTC-USDT 策略模板，包含：

- **1个策略组**: `🔥 BTC-USDT 官方推荐策略`
- **12种策略**: 4种市场状态 × 3种风险偏好

#### 策略矩阵

| 市场状态 | 保守型 🛡️ | 平衡型 ⚖️ | 激进型 🚀 |
|---------|----------|----------|----------|
| 趋势市场 (trend) | 3-5x 杠杆<br>5-10% 仓位<br>3% 止损 | 5-10x 杠杆<br>8-15% 仓位<br>5% 止损 | 10-20x 杠杆<br>10-20% 仓位<br>8% 止损 |
| 震荡市场 (volatile) | 3-4x 杠杆<br>4-8% 仓位<br>2.5% 止损 | 5-8x 杠杆<br>6-12% 仓位<br>4% 止损 | 8-15x 杠杆<br>8-16% 仓位<br>6% 止损 |
| 高波动 (high_vol) | 2-3x 杠杆<br>3-6% 仓位<br>5% 止损 | 4-7x 杠杆<br>5-10% 仓位<br>6% 止损 | 8-12x 杠杆<br>6-12% 仓位<br>10% 止损 |
| 低波动 (low_vol) | 4-6x 杠杆<br>6-12% 仓位<br>2% 止损 | 6-10x 杠杆<br>10-18% 仓位<br>3% 止损 | 12-20x 杠杆<br>15-25% 仓位<br>5% 止损 |

#### 完整策略配置 (config_json)

每个策略包含完整的手动配置参数（`config_json`字段）：

```json
{
  // === 交易配置 ===
  "exchange": "bitget",           // 交易平台
  "symbol": "BTC-USDT",           // 交易对
  "orderType": "market",          // 订单类型
  "marginMode": "isolated",       // 保证金模式
  
  // === 行情监控 ===
  "monitorWindow": 240,           // 时间窗口(秒)
  "volatilityThreshold": 100,     // 波动值(USDT)
  
  // === 杠杆仓位 ===
  "leverage": 8,                  // 推荐杠杆倍数
  "marginPercent": 12,            // 推荐仓位比例(%)
  
  // === 止损止盈 ===
  "stopLossPercent": 5,           // 止损百分比(%)
  "autoStartRetreatPercent": 3,   // 启动回撤百分比(%)
  "profitRetreatPercent": 25,     // 止盈回撤百分比(%)
  
  // === 反向策略 ===
  "reverseEnabled": true,         // 是否启用反向单
  "reverseLossRatio": 50,         // 亏损订单回撤比例(%)
  "reverseProfitRatio": 100,      // 盈利订单回撤比例(%)
  "trailingStop": false,          // 是否启用移动止盈
  
  "remark": "策略说明"
}
```

---

### 2. 完善策略应用功能

**文件**: `server/internal/controller/admin/trading/strategy_template.go`

更新了 `Apply` 函数，现在会完整复制以下参数到机器人：

- ✅ `leverage` - 杠杆倍数
- ✅ `margin_percent` - 保证金比例
- ✅ `risk_preference` - 风险偏好
- ✅ `market_state` - 市场状态
- ✅ `stop_loss_percent` - 止损百分比
- ✅ `profit_retreat_percent` - 止盈回撤百分比
- ✅ `auto_start_retreat_percent` - 启动回撤百分比
- ✅ `enable_reverse_order` - 反向下单开关
- ✅ `current_strategy` - 完整策略配置(JSON)

---

### 3. 优化交易引擎反向策略配置

**文件**: `server/internal/logic/toogo/engine.go`

改进了反向下单逻辑：

1. **新增 `ReverseConfig` 结构体**
   ```go
   type ReverseConfig struct {
       Enabled       bool    `json:"reverseEnabled"`
       LossRatio     float64 `json:"reverseLossRatio"`
       ProfitRatio   float64 `json:"reverseProfitRatio"`
       TrailingStop  bool    `json:"trailingStop"`
   }
   ```

2. **新增 `getReverseConfig` 方法**
   - 优先从机器人的 `current_strategy` JSON 配置中读取
   - 如果没有配置，回退到默认值（根据市场状态）

3. **配置优先级**
   1. 机器人的 `current_strategy.config` 中的反向配置
   2. 默认的 `defaultReverseStrategyConfig` 配置

---

## 📦 部署步骤

### 1. 导入官方策略SQL

```bash
# 进入 MySQL
mysql -u root -p

# 选择数据库
USE hotgo;

# 导入官方策略
SOURCE D:/go/src/hotgo_v2/server/storage/data/official_btc_usdt_strategies.sql;
```

### 2. 验证导入结果

```sql
-- 查看策略组
SELECT * FROM hg_trading_strategy_group WHERE is_official = 1;

-- 查看策略数量
SELECT 
  g.group_name, 
  COUNT(s.id) as strategy_count
FROM hg_trading_strategy_group g
LEFT JOIN hg_trading_strategy_template s ON s.group_id = g.id
WHERE g.is_official = 1
GROUP BY g.id;

-- 查看12种策略
SELECT 
  strategy_name, risk_preference, market_state,
  leverage_min, leverage_max, stop_loss_percent
FROM hg_trading_strategy_template
WHERE strategy_key LIKE 'official_btc_%'
ORDER BY sort;
```

### 3. 重新编译后端

```bash
cd server
go build -o hotgo.exe main.go
```

---

## 📊 策略参数说明

### 市场状态 (market_state)

| 状态 | 说明 | 特征 |
|------|------|------|
| `trend` | 趋势市场 | MA7 > MA25 或 MA7 < MA25，波动率适中 |
| `volatile` | 震荡市场 | 价格在区间内震荡，无明显趋势 |
| `high_vol` | 高波动市场 | 波动率 > 3%，市场剧烈波动 |
| `low_vol` | 低波动市场 | 波动率 < 1%，市场横盘整理 |

### 风险偏好 (risk_preference)

| 偏好 | 杠杆范围 | 仓位范围 | 适合人群 |
|------|---------|---------|---------|
| `conservative` | 2-6x | 3-12% | 新手、稳健型投资者 |
| `balanced` | 4-10x | 5-18% | 有一定经验的用户 |
| `aggressive` | 8-20x | 6-25% | 专业用户、高风险承受能力 |

### 方向判断逻辑（核心策略）

**时间窗口** (`monitorWindow`) 和 **波动点数** (`volatilityThreshold`) 是方向判断的核心参数：

```
方向判断规则：
┌────────────────────────────────────────────────────────────────┐
│  在时间窗口内监控价格，获取最高价和最低价                           │
│                                                                │
│  做空信号：实时价格 ≤ 窗口最高价 - 波动点数                        │
│           → 价格从高点回落超过波动点数，趋势可能反转               │
│                                                                │
│  做多信号：实时价格 ≥ 窗口最低价 + 波动点数                        │
│           → 价格从低点反弹超过波动点数，趋势可能反转               │
└────────────────────────────────────────────────────────────────┘
```

**参数设置建议：**

| 市场状态 | 时间窗口 | 波动点数 | 说明 |
|---------|---------|---------|------|
| 趋势市场 | 240-300秒 | 80-120 USDT | 中等窗口，捕捉趋势反转 |
| 震荡市场 | 180秒 | 50-80 USDT | 较短窗口，快速响应 |
| 高波动 | 60-120秒 | 150-200 USDT | 短窗口，高阈值避免假信号 |
| 低波动 | 300-600秒 | 30-50 USDT | 长窗口，低阈值捕捉小波动 |

---

### 反向下单策略

| 市场状态 | 亏损回撤 | 盈利回撤 | 说明 |
|---------|---------|---------|------|
| 趋势市场 | 50% | 100% | 顺势为主，亏损快速反手 |
| 震荡市场 | 0% | 0% | 不开反向单，避免来回止损 |
| 高波动 | 100% | 100% | 双向操作，捕捉剧烈波动 |
| 低波动 | 0% | 0% | 不开反向单，等待突破 |

---

## 🔧 后续优化建议

### 待完成

1. **添加更多交易所的官方策略**
   - [ ] Binance BTC-USDT 官方策略
   - [ ] OKX BTC-USDT 官方策略
   - [ ] ETH-USDT 官方策略

2. **策略回测功能**
   - [ ] 历史数据回测
   - [ ] 策略收益率统计
   - [ ] 最大回撤计算

3. **策略自动切换**
   - [ ] 根据市场状态自动切换策略
   - [ ] 智能参数调整

4. **用户自定义策略**
   - [ ] 支持用户创建私有策略
   - [ ] 策略分享功能

---

### 4. 优化前端策略编辑表单

**文件**: `web/src/views/toogo/strategy/list.vue`

新增字段：
- **推荐杠杆** (`leverage`) - 应用策略时实际使用的杠杆倍数
- **推荐仓位** (`marginPercent`) - 应用策略时实际使用的仓位比例

表单布局优化：
- 杠杆与仓位区域改为3列布局
- 添加说明提示：推荐值 vs 范围
- config_json 包含所有手动配置参数

---

## 📝 修改文件清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `server/storage/data/official_btc_usdt_strategies.sql` | 更新 | 官方策略SQL，config_json包含完整参数 |
| `server/internal/controller/admin/trading/strategy_template.go` | 更新 | Apply函数完整复制所有手动配置 |
| `server/internal/logic/toogo/engine.go` | 修改 | 优化反向策略配置读取 |
| `web/src/views/toogo/strategy/list.vue` | 更新 | 添加推荐杠杆/仓位字段 |

---

**完成时间**: 2024-11-29

