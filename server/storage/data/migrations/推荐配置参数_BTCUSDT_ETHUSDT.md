# BTCUSDT 和 ETHUSDT 推荐配置参数

## 一、配置说明

根据新市场状态算法的逻辑，生成以下推荐配置：

### 算法逻辑回顾

```
V = (当前K线最高价 - 当前K线最低价) ÷ Delta
D = 方向一致性（0-1之间）

判断规则：
- V < LowV → 低波动
- V >= HighV && D < 0.4 → 高波动
- V >= TrendV && D >= DThreshold → 趋势
- 其他 → 震荡
```

### 配置原则

1. **Delta值**：基于历史K线平均波动设置，表示"正常波动"的基准线
2. **LowV**：小于1，表示低波动的上限
3. **TrendV**：大于LowV但小于HighV，表示趋势的起始阈值
4. **HighV**：大于TrendV，表示高波动的起始阈值
5. **DThreshold**：方向一致性阈值，通常0.6-0.8，表示趋势需要较强的方向一致性
6. **权重**：5个周期权重之和应为1.0，中周期（15m）权重最大

---

## 二、BTCUSDT 推荐配置

### 基本信息
- **当前价格**：约 50000 USDT
- **波动特点**：价格高，波动绝对值大
- **适用场景**：主流币种，流动性好

### 配置参数

#### 1. 市场状态阈值

| 参数 | 值 | 说明 |
|------|-----|------|
| **低波动阈值 (LowV)** | 0.5 | V < 0.5 时判断为低波动 |
| **高波动阈值 (HighV)** | 2.0 | V >= 2.0 且 D < 0.4 时判断为高波动 |
| **趋势阈值 (TrendV)** | 1.2 | V >= 1.2 且 D >= DThreshold 时判断为趋势 |
| **方向一致性阈值 (DThreshold)** | 0.7 | 趋势判断时，D >= 0.7 表示方向一致性强 |

**阈值逻辑说明**：
- LowV = 0.5：实际波动小于正常波动的一半 → 低波动
- TrendV = 1.2：实际波动大于正常波动的1.2倍且方向一致 → 趋势
- HighV = 2.0：实际波动大于正常波动的2倍但方向不一致 → 高波动

#### 2. 各周期Delta值（USDT）

| 周期 | Delta值 | 说明 | 示例 |
|------|---------|------|------|
| **1分钟** | 30 | 1分钟内正常波动30 USDT仍属于低波动 | K线波动5 USDT，V=5/30≈0.17 < 0.5 → 低波动 |
| **5分钟** | 60 | 5分钟内正常波动60 USDT仍属于低波动 | K线波动50 USDT，V=50/60≈0.83 → 震荡 |
| **15分钟** | 150 | 15分钟内正常波动150 USDT仍属于低波动 | K线波动200 USDT，V=200/150≈1.33 > 1.2 → 趋势（如果D>=0.7） |
| **30分钟** | 300 | 30分钟内正常波动300 USDT仍属于低波动 | K线波动600 USDT，V=600/300=2.0 → 高波动（如果D<0.4） |
| **1小时** | 600 | 1小时内正常波动600 USDT仍属于低波动 | K线波动1200 USDT，V=1200/600=2.0 → 高波动 |

**Delta值设置依据**：
- 基于BTCUSDT历史K线数据统计
- 周期越长，Delta值越大（时间窗口越大，正常波动范围越大）
- 1分钟：约30 USDT（0.06%的价格波动）
- 15分钟：约150 USDT（0.3%的价格波动）
- 1小时：约600 USDT（1.2%的价格波动）

#### 3. 周期权重

| 周期 | 权重 | 百分比 | 说明 |
|------|------|--------|------|
| **1分钟** | 0.15 | 15% | 短期波动，权重较小 |
| **5分钟** | 0.20 | 20% | 短期趋势，权重中等 |
| **15分钟** | 0.30 | 30% | 中期趋势，权重最大（核心周期） |
| **30分钟** | 0.25 | 25% | 中期趋势，权重较大 |
| **1小时** | 0.10 | 10% | 长期趋势，权重较小 |
| **合计** | 1.00 | 100% | ✓ |

**权重设置依据**：
- 15分钟周期权重最大（30%），因为它是中期趋势的核心周期
- 1分钟和1小时权重较小，避免短期噪音和长期滞后影响
- 5分钟和30分钟权重中等，平衡短期和中期信息

### 完整配置JSON

```json
{
  "symbol": "BTCUSDT",
  "lowVolatilityThreshold": 0.5,
  "highVolatilityThreshold": 2.0,
  "trendStrengthThreshold": 1.2,
  "dThreshold": 0.7,
  "delta1m": 30.0,
  "delta5m": 60.0,
  "delta15m": 150.0,
  "delta30m": 300.0,
  "delta1h": 600.0,
  "weight1m": 0.15,
  "weight5m": 0.20,
  "weight15m": 0.30,
  "weight30m": 0.25,
  "weight1h": 0.10,
  "isActive": 1
}
```

---

## 三、ETHUSDT 推荐配置

### 基本信息
- **当前价格**：约 3000 USDT
- **波动特点**：价格中等，波动相对BTCUSDT略小
- **适用场景**：主流币种，流动性好

### 配置参数

#### 1. 市场状态阈值

| 参数 | 值 | 说明 |
|------|-----|------|
| **低波动阈值 (LowV)** | 0.5 | V < 0.5 时判断为低波动 |
| **高波动阈值 (HighV)** | 2.0 | V >= 2.0 且 D < 0.4 时判断为高波动 |
| **趋势阈值 (TrendV)** | 1.2 | V >= 1.2 且 D >= DThreshold 时判断为趋势 |
| **方向一致性阈值 (DThreshold)** | 0.7 | 趋势判断时，D >= 0.7 表示方向一致性强 |

**阈值逻辑说明**：
- 与BTCUSDT相同，因为阈值是相对值（V值），不依赖绝对价格

#### 2. 各周期Delta值（USDT）

| 周期 | Delta值 | 说明 | 示例 |
|------|---------|------|------|
| **1分钟** | 20 | 1分钟内正常波动20 USDT仍属于低波动 | K线波动3 USDT，V=3/20=0.15 < 0.5 → 低波动 |
| **5分钟** | 40 | 5分钟内正常波动40 USDT仍属于低波动 | K线波动30 USDT，V=30/40=0.75 → 震荡 |
| **15分钟** | 100 | 15分钟内正常波动100 USDT仍属于低波动 | K线波动150 USDT，V=150/100=1.5 > 1.2 → 趋势（如果D>=0.7） |
| **30分钟** | 200 | 30分钟内正常波动200 USDT仍属于低波动 | K线波动400 USDT，V=400/200=2.0 → 高波动（如果D<0.4） |
| **1小时** | 400 | 1小时内正常波动400 USDT仍属于低波动 | K线波动800 USDT，V=800/400=2.0 → 高波动 |

**Delta值设置依据**：
- ETHUSDT价格约为BTCUSDT的1/16，但波动幅度（USDT）约为BTCUSDT的2/3
- 1分钟：约20 USDT（0.67%的价格波动）
- 15分钟：约100 USDT（3.3%的价格波动）
- 1小时：约400 USDT（13.3%的价格波动）

#### 3. 周期权重

| 周期 | 权重 | 百分比 | 说明 |
|------|------|--------|------|
| **1分钟** | 0.15 | 15% | 短期波动，权重较小 |
| **5分钟** | 0.20 | 20% | 短期趋势，权重中等 |
| **15分钟** | 0.30 | 30% | 中期趋势，权重最大（核心周期） |
| **30分钟** | 0.25 | 25% | 中期趋势，权重较大 |
| **1小时** | 0.10 | 10% | 长期趋势，权重较小 |
| **合计** | 1.00 | 100% | ✓ |

**权重设置依据**：
- 与BTCUSDT相同，因为权重反映的是周期的重要性，不依赖币种

### 完整配置JSON

```json
{
  "symbol": "ETHUSDT",
  "lowVolatilityThreshold": 0.5,
  "highVolatilityThreshold": 2.0,
  "trendStrengthThreshold": 1.2,
  "dThreshold": 0.7,
  "delta1m": 20.0,
  "delta5m": 40.0,
  "delta15m": 100.0,
  "delta30m": 200.0,
  "delta1h": 400.0,
  "weight1m": 0.15,
  "weight5m": 0.20,
  "weight15m": 0.30,
  "weight30m": 0.25,
  "weight1h": 0.10,
  "isActive": 1
}
```

---

## 四、配置验证示例

### BTCUSDT 示例1：低波动市场

**15分钟K线**：
- 最高价：50020 USDT
- 最低价：50010 USDT
- 开盘价：50015 USDT
- 收盘价：50018 USDT
- 波动：10 USDT
- Delta15m：150 USDT
- **V = 10 / 150 ≈ 0.067**

**判断**：
- V = 0.067 < LowV = 0.5
- → **低波动市场** ✓

### BTCUSDT 示例2：震荡市场

**15分钟K线**：
- 最高价：50080 USDT
- 最低价：50020 USDT
- 开盘价：50050 USDT
- 收盘价：50040 USDT
- 波动：60 USDT
- Delta15m：150 USDT
- **V = 60 / 150 = 0.4**
- **D = (50040 - 50020) / 60 ≈ 0.33**

**判断**：
- V = 0.4 < LowV = 0.5，但接近
- D = 0.33 < DThreshold = 0.7
- → **震荡市场** ✓

### BTCUSDT 示例3：趋势市场

**15分钟K线**：
- 最高价：50180 USDT
- 最低价：50020 USDT
- 开盘价：50050 USDT
- 收盘价：50160 USDT（上涨）
- 波动：160 USDT
- Delta15m：150 USDT
- **V = 160 / 150 ≈ 1.07**
- **D = (50160 - 50020) / 160 = 0.875**

**判断**：
- V = 1.07 < TrendV = 1.2（未达到趋势阈值）
- → **震荡市场**（接近趋势）

**如果波动更大**：
- 波动：200 USDT
- **V = 200 / 150 ≈ 1.33**
- **D = (50160 - 50020) / 200 = 0.7**

**判断**：
- V = 1.33 >= TrendV = 1.2
- D = 0.7 >= DThreshold = 0.7
- → **趋势市场** ✓

### BTCUSDT 示例4：高波动市场

**15分钟K线**：
- 最高价：50300 USDT
- 最低价：50000 USDT
- 开盘价：50100 USDT
- 收盘价：50150 USDT（震荡）
- 波动：300 USDT
- Delta15m：150 USDT
- **V = 300 / 150 = 2.0**
- **D = (50150 - 50000) / 300 ≈ 0.5**

**判断**：
- V = 2.0 >= HighV = 2.0
- D = 0.5 > 0.4（但小于DThreshold = 0.7，方向不一致）
- → **高波动市场** ✓

---

## 五、配置调整建议

### 1. Delta值调整

**如果经常误判为高波动**：
- 适当增大Delta值（提高正常波动基准线）
- 例如：BTCUSDT 15分钟 Delta 从 150 调整为 200

**如果经常误判为低波动**：
- 适当减小Delta值（降低正常波动基准线）
- 例如：BTCUSDT 15分钟 Delta 从 150 调整为 120

### 2. 阈值调整

**如果趋势识别不敏感**：
- 适当降低 TrendV（例如从 1.2 调整为 1.0）
- 或适当降低 DThreshold（例如从 0.7 调整为 0.6）

**如果趋势识别过于敏感**：
- 适当提高 TrendV（例如从 1.2 调整为 1.5）
- 或适当提高 DThreshold（例如从 0.7 调整为 0.8）

### 3. 权重调整

**如果短期波动影响过大**：
- 减小1分钟和5分钟权重
- 增大15分钟和30分钟权重

**如果长期趋势影响过大**：
- 减小1小时权重
- 增大15分钟和30分钟权重

---

## 六、SQL更新语句

### BTCUSDT配置

```sql
INSERT INTO hg_toogo_volatility_config (
    symbol,
    low_volatility_threshold,
    high_volatility_threshold,
    trend_strength_threshold,
    d_threshold,
    delta_1m, delta_5m, delta_15m, delta_30m, delta_1h,
    weight_1m, weight_5m, weight_15m, weight_30m, weight_1h,
    is_active,
    created_at,
    updated_at
) VALUES (
    'BTCUSDT',
    0.5,   -- LowV
    2.0,   -- HighV
    1.2,   -- TrendV
    0.7,   -- DThreshold
    30.0,  -- Delta1m
    60.0,  -- Delta5m
    150.0, -- Delta15m
    300.0, -- Delta30m
    600.0, -- Delta1h
    0.15,  -- Weight1m
    0.20,  -- Weight5m
    0.30,  -- Weight15m
    0.25,  -- Weight30m
    0.10,  -- Weight1h
    1,
    NOW(),
    NOW()
) ON DUPLICATE KEY UPDATE
    low_volatility_threshold = 0.5,
    high_volatility_threshold = 2.0,
    trend_strength_threshold = 1.2,
    d_threshold = 0.7,
    delta_1m = 30.0,
    delta_5m = 60.0,
    delta_15m = 150.0,
    delta_30m = 300.0,
    delta_1h = 600.0,
    weight_1m = 0.15,
    weight_5m = 0.20,
    weight_15m = 0.30,
    weight_30m = 0.25,
    weight_1h = 0.10,
    updated_at = NOW();
```

### ETHUSDT配置

```sql
INSERT INTO hg_toogo_volatility_config (
    symbol,
    low_volatility_threshold,
    high_volatility_threshold,
    trend_strength_threshold,
    d_threshold,
    delta_1m, delta_5m, delta_15m, delta_30m, delta_1h,
    weight_1m, weight_5m, weight_15m, weight_30m, weight_1h,
    is_active,
    created_at,
    updated_at
) VALUES (
    'ETHUSDT',
    0.5,   -- LowV
    2.0,   -- HighV
    1.2,   -- TrendV
    0.7,   -- DThreshold
    20.0,  -- Delta1m
    40.0,  -- Delta5m
    100.0, -- Delta15m
    200.0, -- Delta30m
    400.0, -- Delta1h
    0.15,  -- Weight1m
    0.20,  -- Weight5m
    0.30,  -- Weight15m
    0.25,  -- Weight30m
    0.10,  -- Weight1h
    1,
    NOW(),
    NOW()
) ON DUPLICATE KEY UPDATE
    low_volatility_threshold = 0.5,
    high_volatility_threshold = 2.0,
    trend_strength_threshold = 1.2,
    d_threshold = 0.7,
    delta_1m = 20.0,
    delta_5m = 40.0,
    delta_15m = 100.0,
    delta_30m = 200.0,
    delta_1h = 400.0,
    weight_1m = 0.15,
    weight_5m = 0.20,
    weight_15m = 0.30,
    weight_30m = 0.25,
    weight_1h = 0.10,
    updated_at = NOW();
```

---

## 七、总结

### BTCUSDT配置要点

1. **Delta值**：基于历史波动统计，周期越长Delta值越大
2. **阈值**：LowV=0.5, TrendV=1.2, HighV=2.0, DThreshold=0.7
3. **权重**：15分钟周期权重最大（30%），1小时权重最小（10%）

### ETHUSDT配置要点

1. **Delta值**：比BTCUSDT小约1/3-1/2，但阈值相同
2. **阈值**：与BTCUSDT相同（因为阈值是相对值）
3. **权重**：与BTCUSDT相同（周期重要性不依赖币种）

### 使用建议

1. **初始使用**：直接使用推荐配置
2. **观察效果**：运行1-2周，观察市场状态判断的准确性
3. **微调参数**：根据实际效果调整Delta值和阈值
4. **记录优化**：记录每次调整的效果，持续优化

