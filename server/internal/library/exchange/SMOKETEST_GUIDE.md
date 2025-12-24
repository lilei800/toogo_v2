## Exchange SmokeTest 使用指南（只读/真实交易）

本指南用于你“自己测试”多交易所接入是否可用。

### 1. 前置准备

- 在后台创建对应交易所的 API 配置（`量化管理 -> API配置`），确保：
  - **U本位永续权限**（合约交易）
  - **读取余额/持仓权限**
  - 若要真实下单：需要**下单/撤单**权限
- 记住该条记录的 `apiConfigId`（即 `hg_trading_api_config.id`）
- 交易对建议先用：`BTCUSDT`

### 2. 只读模式（推荐先跑）

只读模式不会改任何交易所配置，也不会下单：

```powershell
cd D:\go\src\hotgo_v2\server
go run internal/cmd/exchange_smoketest/main.go --apiConfigId=3 --symbol=BTCUSDT --interval=5m --limit=10
```

期望输出：
- GetTicker / GetKlines 成功
- GetBalance 返回非 0（或至少不报错）
- GetPositions / GetOpenOrders / GetOrderHistory 不报错

### 3. 真实交易模式（双保险）

必须同时满足：
- `--trade=1`
- `--yes=1`

该模式会：
1) 尝试设置逐仓 `ISOLATED`（失败不阻断）  
2) 尝试设置杠杆 `--leverage`（失败不阻断）  
3) 市价开多 + 市价开空（验证双向持仓）  
4) 读取持仓并 `reduceOnly` 平仓  

示例（BTC 最小量建议从 `0.0001` 起）：

```powershell
cd D:\go\src\hotgo_v2\server
go run internal/cmd/exchange_smoketest/main.go --apiConfigId=3 --symbol=BTCUSDT --trade=1 --yes=1 --qty=0.0001 --leverage=3 --sleepSec=2
```

### 4. 常见失败与处理

- **数量过小（OKX/Gate 常见）**
  - 这两家下单需要把“基础币数量”折算为“合约张数”，如果折算为 0 会直接失败。
  - 处理：提高 `--qty`，或者换价格更低的币种做联调（更容易满足最小张数）。

- **权限错误 / 签名错误**
  - 检查 API Key 权限是否包含：合约交易、读取余额/持仓、下单/撤单
  - 检查服务器时间是否准确（NTP）
  - 如果使用代理，确保代理可访问交易所域名

### 5. 交易所符号映射（供参考）

系统统一传入：`BTCUSDT`

- Binance：`BTCUSDT`
- Bitget：`BTCUSDT`
- OKX：内部转换为 `BTC-USDT-SWAP`
- Gate：内部转换为 `BTC_USDT`


