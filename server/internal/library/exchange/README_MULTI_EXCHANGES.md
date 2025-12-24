## 多交易所接入说明（U本位永续 / 逐仓 / 双向持仓）

本目录实现了统一交易所接口 `Exchange`，用于机器人引擎对接不同平台。

### 快速自测（推荐）

新增了一个**数据库直连**的联调工具：`internal/cmd/exchange_smoketest`。

- **只读模式（默认）**：不会下单、不改杠杆、不改逐仓

```powershell
cd D:\go\src\hotgo_v2\server
go run internal/cmd/exchange_smoketest/main.go --apiConfigId=3 --symbol=BTCUSDT --interval=5m --limit=10
```

- **真实交易模式（可选，双保险）**：必须同时 `--trade=1 --yes=1` 才会执行；会尝试 `SetMarginType(ISOLATED)`、`SetLeverage`，然后 **开多+开空** 再 **reduceOnly 平仓**

```powershell
cd D:\go\src\hotgo_v2\server
go run internal/cmd/exchange_smoketest/main.go --apiConfigId=3 --symbol=BTCUSDT --trade=1 --yes=1 --qty=0.0001 --leverage=3 --sleepSec=2
```

> 建议先用模拟盘/小额账号测试，且只在你确认账户允许交易时使用真实交易模式。

### 当前支持的交易所

- **bitget**：实现完整（USDT-FUTURES + isolated + long/short）
- **binance**：USDT 永续（fapi），已增加安全补丁：
  - **首次下单**尝试开启 **双向持仓（Hedge Mode）**
  - **首次下单**尝试设置 **逐仓（ISOLATED）**
  - 平仓单默认带 **reduceOnly=true**（避免误开反向单/误平仓）
- **okx**：API v5，SWAP + `tdMode=isolated` + `posSide=long/short`，内部将“基础币数量”折算为 OKX `sz(合约张数)`（通过 `ctVal` 缓存）
- **gate**：API v4，futures/usdt，内部将“基础币数量”折算为 `size(合约张数)`（通过 `quanto_multiplier/contract_size` 缓存）

### 重要约束（与机器人逻辑一致）

- **只支持 U本位永续**
- **只支持逐仓**
- **必须支持双向持仓（多/空同时存在）**

### 符号格式（系统内部入参）

系统层面统一用：`BTCUSDT`（不带分隔符）。

- **bitget**：内部会格式化为 `BTCUSDT`
- **binance**：内部会格式化为 `BTCUSDT`
- **okx**：内部会格式化为 `BTC-USDT-SWAP`
- **gate**：内部会格式化为 `BTC_USDT`

### 关键字段映射

- **Binance**
  - symbol：`BTCUSDT`
  - positionSide：`LONG/SHORT`（要求 Hedge Mode）
  - close：通过 `reduceOnly=true` 的市价单实现

- **OKX**
  - instId：`BTC-USDT-SWAP`
  - `tdMode=isolated`
  - `posSide=long/short`
  - quantity：系统内使用基础币数量；OKX 下单使用 `sz`（合约张数），通过 `ctVal` 折算：`sz = floor(qty / ctVal)`
  - **重要**：如果 `sz=0` 会报错（数量过小），需要提高 `qty` 或做最小下单量校验

- **Gate**
  - contract：`BTC_USDT`
  - quantity：系统内使用基础币数量；Gate 下单使用 `size`（合约张数），通过 `quanto_multiplier/contract_size` 折算：`size = floor(qty / multiplier)`，买为正、卖为负
  - **重要**：如果 `size=0` 会报错（数量过小），需要提高 `qty` 或做最小下单量校验
  - close：通过 `reduce_only=true` 的市价单实现

### 常见问题

- **下单数量过小**
  - OKX/Gate 需要先折算为合约张数；如果折算为 0，会返回错误。应提高最小下单数量或在策略侧做数量精度/最小名义价值判断。

- **签名失败 / 权限不足**
  - 请确认 API Key 权限：合约交易、读取余额/持仓、下单/撤单权限齐全
  - 如使用代理，确保能访问交易所域名且时间同步正常（签名通常依赖时间戳）


