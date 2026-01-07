# 甯佸畨浜ゆ槗鎵€瀹樻柟绛栫暐缁勫鍏ヨ鏄庯紙V2锛?
> 鏈鏄庨潰鍚?**PostgreSQL** 鐜锛堜笌浣犲綋鍓?`*_pg.sql` 鏂囦欢涓€鑷达級銆?>
> V2 閫傞厤浜嗙郴缁熺殑鈥滄柊甯傚満鐘舵€佺畻娉?娉㈠姩鐜囬厤缃?ToogoVolatilityConfig)/榛樿椋庨櫓鍋忓ソ鏄犲皠(v2)鈥濓紝骞朵笖鏈哄櫒浜鸿繍琛屾椂浼氱洿鎺ヤ娇鐢ㄧ瓥鐣ユā鏉胯〃閲岀殑鍙傛暟锛堟潬鏉?淇濊瘉閲?姝㈡崯/姝㈢泩/绐楀彛/闃堝€硷級銆?
## 馃搵 鏂囦欢璇存槑

### PostgreSQL鐗堟湰SQL鏂囦欢
- `official_binance_btcusdt_v3_pg.sql` - Binance BTCUSDT 瀹樻柟绛栫暐 V3锛堟墜缁垂/婊戠偣 盈浜忓钩琛＄偣绾︽潫锛?
- `official_binance_btcusdt_v2_pg.sql` - Binance BTCUSDT 瀹樻柟绛栫暐 V2锛堟柊绠楁硶锛?
- `official_binance_ethusdt_v2_pg.sql` - Binance ETHUSDT 瀹樻柟绛栫暐 V2锛堟柊绠楁硶锛?

### 鍏煎鏃х増锛堝彲閫夛級
- `official_binance_btcusdt_v1_pg.sql`
- `official_binance_ethusdt_v1_pg.sql`

### 瀵煎叆鑴氭湰锛圴2锛?- `import_binance_strategies_pg_v2.bat` - Windows
- `import_binance_strategies_pg_v2.sh` - Linux/Mac

## 馃殌 瀵煎叆鏂规硶

### 鏂规硶涓€锛氫娇鐢ㄥ鍏ヨ剼鏈紙鎺ㄨ崘锛?
#### Windows绯荤粺（V3）

```bash
cd D:\go\src\hotgo_v2\server\storage\data
import_binance_strategies_pg_v3.bat
```

#### Linux/Mac绯荤粺（V3）

```bash
cd /path/to/hotgo_v2/server/storage/data
chmod +x import_binance_strategies_pg_v2.sh
chmod +x import_binance_strategies_pg_v3.sh
./import_binance_strategies_pg_v3.sh
```

### 鏂规硶浜岋細鎵嬪姩瀵煎叆SQL鏂囦欢

```bash
psql -h localhost -p 5432 -U postgres -d hotgo -f official_binance_btcusdt_v3_pg.sql
psql -h localhost -p 5432 -U postgres -d hotgo -f official_binance_ethusdt_v2_pg.sql
```

## 鉁?楠岃瘉瀵煎叆缁撴灉

```sql
-- 鏌ョ湅绛栫暐缁?SELECT
  group_name AS "绛栫暐缁勫悕绉?,
  group_key  AS "鏍囪瘑",
  exchange   AS "浜ゆ槗鎵€",
  symbol     AS "浜ゆ槗瀵?,
  is_official AS "鏄惁瀹樻柟",
  is_active   AS "鏄惁鍚敤"
FROM hg_trading_strategy_group
WHERE group_key IN ('official_binance_btcusdt_v3', 'official_binance_ethusdt_v2')
ORDER BY id;

-- 鏌ョ湅绛栫暐璇︽儏
SELECT
  g.group_key AS "绛栫暐缁勬爣璇?,
  s.sort AS "鎺掑簭",
  s.strategy_name AS "绛栫暐鍚嶇О",
  s.market_state AS "甯傚満鐘舵€?,
  s.risk_preference AS "椋庨櫓鍋忓ソ",
  s.monitor_window AS "绐楀彛(s)",
  s.volatility_threshold AS "闃堝€?USDT)",
  s.leverage AS "鏉犳潌",
  s.margin_percent AS "淇濊瘉閲?%)",
  s.stop_loss_percent AS "姝㈡崯(%)",
  s.auto_start_retreat_percent AS "鍚姩姝㈢泩(%)",
  s.profit_retreat_percent AS "姝㈢泩鍥炴挙(%)"
FROM hg_trading_strategy_group g
JOIN hg_trading_strategy_template s ON s.group_id = g.id
WHERE g.group_key IN ('official_binance_btcusdt_v3', 'official_binance_ethusdt_v2')
ORDER BY g.id, s.sort;
```
