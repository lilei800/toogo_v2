# BTCUSDT和ETHUSDT波动率配置

## 配置参数说明

### BTCUSDT（比特币）

**特点：**
- 市值最大，流动性最好
- 波动相对稳定
- 适合作为主流币种的基准配置

**配置值：**
```json
{
  "symbol": "BTCUSDT",
  "highVolatilityThreshold": 2.0,    // 高波动阈值：2.0%
  "lowVolatilityThreshold": 0.4,     // 低波动阈值：0.4%
  "trendStrengthThreshold": 0.35,   // 趋势阈值：0.35
  "weight1m": 0.10,                  // 1分钟权重：10%
  "weight5m": 0.20,                 // 5分钟权重：20%
  "weight15m": 0.25,                // 15分钟权重：25%
  "weight30m": 0.25,                // 30分钟权重：25%
  "weight1h": 0.20,                 // 1小时权重：20%
  "isActive": 1                      // 启用状态
}
```

**配置理由：**
- 高波动阈值2.0%：比特币波动相对稳定，2.0%已经算是高波动
- 低波动阈值0.4%：比特币波动较小，0.4%以下算是低波动
- 权重分布：偏向中长期周期（15m、30m、1h），更稳定

### ETHUSDT（以太坊）

**特点：**
- 市值第二，流动性好
- 波动可能比BTC稍大
- 适合作为主流币种的参考配置

**配置值：**
```json
{
  "symbol": "ETHUSDT",
  "highVolatilityThreshold": 2.5,    // 高波动阈值：2.5%
  "lowVolatilityThreshold": 0.5,     // 低波动阈值：0.5%
  "trendStrengthThreshold": 0.35,   // 趋势阈值：0.35
  "weight1m": 0.10,                  // 1分钟权重：10%
  "weight5m": 0.20,                  // 5分钟权重：20%
  "weight15m": 0.25,                 // 15分钟权重：25%
  "weight30m": 0.25,                 // 30分钟权重：25%
  "weight1h": 0.20,                  // 1小时权重：20%
  "isActive": 1                      // 启用状态
}
```

**配置理由：**
- 高波动阈值2.5%：以太坊波动稍大，2.5%算是高波动
- 低波动阈值0.5%：以太坊波动稍大，0.5%以下算是低波动
- 权重分布：偏向中长期周期（15m、30m、1h），更稳定

## 使用方法

### 方法1：通过SQL脚本创建（推荐）

执行 `create_volatility_configs.sql` 文件中的SQL语句：

```bash
mysql -u用户名 -p数据库名 < create_volatility_configs.sql
```

### 方法2：通过API创建

在前端页面或通过API调用：

```javascript
// 创建BTCUSDT配置
await ToogoVolatilityConfigApi.create({
  symbol: 'BTCUSDT',
  highVolatilityThreshold: 2.0,
  lowVolatilityThreshold: 0.4,
  trendStrengthThreshold: 0.35,
  weight1m: 0.10,
  weight5m: 0.20,
  weight15m: 0.25,
  weight30m: 0.25,
  weight1h: 0.20,
  isActive: 1,
});

// 创建ETHUSDT配置
await ToogoVolatilityConfigApi.create({
  symbol: 'ETHUSDT',
  highVolatilityThreshold: 2.5,
  lowVolatilityThreshold: 0.5,
  trendStrengthThreshold: 0.35,
  weight1m: 0.10,
  weight5m: 0.20,
  weight15m: 0.25,
  weight30m: 0.25,
  weight1h: 0.20,
  isActive: 1,
});
```

### 方法3：通过管理界面创建

1. 访问波动率配置页面
2. 点击"新增配置"
3. 选择交易对 BTCUSDT 或 ETHUSDT
4. 填入对应的配置值
5. 保存

## 配置验证

创建后可以通过以下方式验证：

1. 查看波动率配置列表，确认BTCUSDT和ETHUSDT已创建
2. 检查配置值是否正确
3. 确认启用状态为"启用"

## 后续调整建议

根据实际交易情况，可以调整以下参数：

1. **高波动阈值**：如果发现频繁触发高波动，可以适当提高
2. **低波动阈值**：如果发现低波动判断不准确，可以适当调整
3. **趋势阈值**：如果趋势判断不准确，可以调整趋势阈值
4. **权重分布**：根据实际效果调整各时间周期的权重

