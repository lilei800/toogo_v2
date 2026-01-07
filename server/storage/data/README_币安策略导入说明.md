# 币安交易所官方策略组导入说明

## 📋 文件说明

### PostgreSQL版本SQL文件
- `official_binance_btcusdt_v1_pg.sql` - 币安BTCUSDT策略组（12套策略）
- `official_binance_ethusdt_v1_pg.sql` - 币安ETHUSDT策略组（12套策略）

### 导入脚本
- `import_binance_strategies_pg.bat` - Windows批处理脚本
- `import_binance_strategies_pg.sh` - Linux/Mac Shell脚本

## 🚀 导入方法

### 方法一：使用导入脚本（推荐）

#### Windows系统
```bash
# 1. 修改批处理脚本中的数据库连接参数
# 编辑 import_binance_strategies_pg.bat，修改以下变量：
#   DB_HOST=localhost
#   DB_PORT=5432
#   DB_USER=postgres
#   DB_NAME=hotgo
#   PGPASSWORD=your_password_here（如果需要）

# 2. 双击运行或在命令行执行
cd D:\go\src\hotgo_v2\server\storage\data
import_binance_strategies_pg.bat
```

#### Linux/Mac系统
```bash
# 1. 修改Shell脚本中的数据库连接参数
# 编辑 import_binance_strategies_pg.sh，修改以下变量：
#   DB_HOST="localhost"
#   DB_PORT="5432"
#   DB_USER="postgres"
#   DB_NAME="hotgo"
#   export PGPASSWORD="your_password_here"（如果需要）

# 2. 添加执行权限并运行
chmod +x import_binance_strategies_pg.sh
./import_binance_strategies_pg.sh
```

### 方法二：手动导入SQL文件

```bash
# 导入BTCUSDT策略组
psql -h localhost -p 5432 -U postgres -d hotgo -f official_binance_btcusdt_v1_pg.sql

# 导入ETHUSDT策略组
psql -h localhost -p 5432 -U postgres -d hotgo -f official_binance_ethusdt_v1_pg.sql
```

### 方法三：使用数据库管理工具

1. 打开PostgreSQL管理工具（如pgAdmin、DBeaver、Navicat等）
2. 连接到数据库
3. 打开SQL文件并执行

## ✅ 验证导入结果

导入完成后，执行以下SQL验证：

```sql
-- 查看策略组
SELECT 
  group_name AS "策略组名称",
  exchange AS "交易所",
  symbol AS "交易对",
  is_official AS "是否官方"
FROM hg_trading_strategy_group 
WHERE group_key IN ('official_binance_btcusdt_v1', 'official_binance_ethusdt_v1');

-- 查看策略详情
SELECT 
  g.group_name AS "策略组",
  s.strategy_name AS "策略名称",
  s.risk_preference AS "风险偏好",
  s.market_state AS "市场状态",
  s.leverage || 'x' AS "杠杆",
  s.margin_percent || '%' AS "保证金",
  s.stop_loss_percent || '%' AS "止损",
  s.volatility_threshold AS "波动阈值(USDT)"
FROM hg_trading_strategy_group g
JOIN hg_trading_strategy_template s ON s.group_id = g.id
WHERE g.group_key IN ('official_binance_btcusdt_v1', 'official_binance_ethusdt_v1')
ORDER BY g.id, s.sort;
```

## 📊 策略组详情

### 币安BTCUSDT策略组
- **策略组名称**: 🔥 Binance BTC-USDT 官方策略 V1.0
- **交易所**: binance
- **交易对**: BTCUSDT
- **策略数量**: 12套
- **波动阈值**: 50-300 USDT

### 币安ETHUSDT策略组
- **策略组名称**: 🔥 Binance ETH-USDT 官方策略 V1.0
- **交易所**: binance
- **交易对**: ETHUSDT
- **策略数量**: 12套
- **波动阈值**: 4-25 USDT

## 💡 注意事项

1. **数据库连接**: 确保PostgreSQL服务已启动，并且有足够的权限
2. **密码设置**: 如果设置了密码，需要在脚本中配置PGPASSWORD环境变量
3. **重复导入**: SQL文件会自动清理旧数据，可以安全重复导入
4. **事务处理**: 所有操作都在事务中执行，失败会自动回滚

## 🔧 故障排查

### 问题1: 连接失败
```
错误: could not connect to server
解决: 检查PostgreSQL服务是否启动，连接参数是否正确
```

### 问题2: 权限不足
```
错误: permission denied
解决: 确保数据库用户有足够的权限（INSERT, DELETE, SELECT）
```

### 问题3: 表不存在
```
错误: relation "hg_trading_strategy_group" does not exist
解决: 确保数据库表已创建，检查表名前缀是否正确
```

## 📞 技术支持

如有问题，请检查：
1. PostgreSQL版本 >= 14
2. 数据库表结构是否正确
3. 用户权限是否足够
4. SQL文件编码是否为UTF-8

