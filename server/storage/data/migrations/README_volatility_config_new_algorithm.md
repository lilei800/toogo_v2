# 波动率配置表迁移指南：适配新算法

## 迁移内容

本次迁移为 `hg_toogo_volatility_config` 表添加以下字段，以支持新的市场状态算法：

### 新增字段

1. **Delta值字段**（各周期波动点数阈值）：
   - `delta_1m` - 1分钟周期delta值（默认：2.0000）
   - `delta_5m` - 5分钟周期delta值（默认：2.0000）
   - `delta_15m` - 15分钟周期delta值（默认：3.0000）
   - `delta_30m` - 30分钟周期delta值（默认：3.0000）
   - `delta_1h` - 1小时周期delta值（默认：5.0000）

2. **方向一致性阈值**：
   - `d_threshold` - 方向一致性阈值，用于判断趋势市场（默认：0.7000，范围：0-1）

## 执行方式

### 方式1：使用批处理脚本（推荐）

1. 编辑 `execute_volatility_config_new_algorithm.bat`，修改数据库连接信息：
   ```bat
   set DB_HOST=127.0.0.1
   set DB_PORT=3306
   set DB_USER=root
   set DB_PASS=your_password
   set DB_NAME=hotgo
   ```

2. 双击运行 `execute_volatility_config_new_algorithm.bat`

### 方式2：使用MySQL命令行

```bash
mysql -h127.0.0.1 -P3306 -uroot -pyour_password hotgo < volatility_config_new_algorithm_fixed.sql
```

### 方式3：使用Go执行工具

```bash
cd server
go run internal/cmd/execsql/main.go storage/data/migrations/volatility_config_new_algorithm_fixed.sql
```

### 方式4：手动执行SQL

1. 打开MySQL客户端（如Navicat、DBeaver、MySQL Workbench等）
2. 连接到数据库
3. 执行 `volatility_config_new_algorithm_fixed.sql` 文件中的所有SQL语句

## 验证迁移结果

执行以下SQL查询验证字段是否添加成功：

```sql
-- 检查字段是否存在
SELECT 
    COLUMN_NAME,
    COLUMN_TYPE,
    COLUMN_DEFAULT,
    COLUMN_COMMENT
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME = 'hg_toogo_volatility_config'
  AND COLUMN_NAME IN ('delta_1m', 'delta_5m', 'delta_15m', 'delta_30m', 'delta_1h', 'd_threshold')
ORDER BY ORDINAL_POSITION;

-- 检查现有配置的默认值
SELECT 
    id,
    symbol,
    delta_1m,
    delta_5m,
    delta_15m,
    delta_30m,
    delta_1h,
    d_threshold
FROM hg_toogo_volatility_config
LIMIT 10;
```

## 注意事项

1. **备份数据**：执行迁移前，请先备份数据库
2. **字段已存在**：如果字段已存在，迁移脚本会跳过（使用 `COALESCE` 处理）
3. **默认值**：所有新增字段都有默认值，现有记录会自动填充默认值
4. **兼容性**：迁移后，旧代码仍可正常运行（新字段有默认值）

## 新算法说明

新算法使用单根K线（O, H, L, P）和delta值计算市场状态：

- **波动强度 (V)** = (最高价 - 最低价) / delta
- **方向一致性 (D)** = (收盘价 - 最低价) / (最高价 - 最低价) [上涨时]
- **方向一致性 (D)** = (最高价 - 收盘价) / (最高价 - 最低价) [下跌时]

市场状态判断规则：
- **低波动**：V < LowV
- **高波动**：V >= HighV 且 D < 0.4
- **趋势**：V >= TrendV 且 D >= DThreshold
- **震荡**：其他情况

## 回滚方案

如果需要回滚迁移，执行以下SQL：

```sql
ALTER TABLE `hg_toogo_volatility_config` 
DROP COLUMN IF EXISTS `delta_1m`,
DROP COLUMN IF EXISTS `delta_5m`,
DROP COLUMN IF EXISTS `delta_15m`,
DROP COLUMN IF EXISTS `delta_30m`,
DROP COLUMN IF EXISTS `delta_1h`,
DROP COLUMN IF EXISTS `d_threshold`;
```

## 技术支持

如有问题，请查看：
- 代码文件：`server/internal/library/market/new_algorithm.go`
- 配置管理：`server/internal/library/config/volatility_config_manager.go`

