# Bitget API V2 实现说明

## 官方API文档
- Bitget API 文档：https://www.bitget.com/zh-CN/api-doc/common/intro
- 合约交易 API：https://www.bitget.com/zh-CN/api-doc/contract/intro

## 当前实现状态

### API路径格式
根据项目中的 `internal/library/exchange/bitget.go` 实现，Bitget V2 API使用以下格式：

**基础URL**: `https://api.bitget.com`

**API路径格式**: `/api/v2/mix/{endpoint}`

### 已实现的接口

1. **账户余额查询**
   - 路径: `/api/v2/mix/account/accounts`
   - 方法: `GET`
   - 参数: `productType: "USDT-FUTURES"`
   - 实现位置: `addons/exchange_bitget/service/exchange.go:68`

2. **获取行情**
   - 路径: `/api/v2/mix/market/ticker`
   - 方法: `GET`
   - 参数: `symbol`, `productType: "USDT-FUTURES"`
   - 实现位置: `addons/exchange_bitget/service/exchange.go:131`

3. **获取K线**
   - 路径: `/api/v2/mix/market/candles`
   - 方法: `GET`
   - 参数: `symbol`, `granularity`, `limit`, `productType: "USDT-FUTURES"`
   - 实现位置: `addons/exchange_bitget/service/exchange.go:186`

4. **下单**
   - 路径: `/api/v2/mix/order/place-order`
   - 方法: `POST`
   - 参数: `symbol`, `marginCoin`, `side`, `orderType`, `size`, `price`等
   - 实现位置: `addons/exchange_bitget/service/exchange.go:257`

5. **查询订单**
   - 路径: `/api/v2/mix/order/history`
   - 方法: `GET`
   - 参数: `productType: "USDT-FUTURES"`, `symbol`, `orderId`
   - 实现位置: `addons/exchange_bitget/service/exchange.go:381`

6. **获取持仓**
   - 路径: `/api/v2/mix/position/all-position`
   - 方法: `GET`
   - 参数: `productType: "USDT-FUTURES"`, `symbol`(可选)
   - 实现位置: `addons/exchange_bitget/service/exchange.go:448`

### 关键参数

**productType**: 统一使用 `"USDT-FUTURES"` 表示USDT合约

**路径格式**: 使用连字符格式（kebab-case），如：
- `/api/v2/mix/order/place-order` ✅
- `/api/v2/mix/position/all-position` ✅
- `/api/v2/mix/order/history` ✅

### 参考实现

项目中的 `internal/library/exchange/bitget.go` 提供了正确的实现参考：
- 所有接口都使用 `/api/v2/mix/` 前缀
- 所有接口都使用 `productType: "USDT-FUTURES"`
- 路径使用连字符格式

## 可能的问题

### 1. 404错误
如果遇到404错误，可能的原因：
- 代理配置问题（需要HTTP代理才能访问Bitget API）
- API路径格式错误（已修复为 `/api/v2/mix/` 格式）
- 参数格式错误（已修复为 `USDT-FUTURES`）

### 2. 代理配置
- HTTP代理（127.0.0.1:33210）已验证可用
- SOCKS5代理（127.0.0.1:33211）如果测试失败，说明该端口可能不是SOCKS5代理

### 3. 签名认证
Bitget API需要以下请求头：
- `ACCESS-KEY`: API Key
- `ACCESS-SIGN`: 签名（HMAC-SHA256 + Base64）
- `ACCESS-TIMESTAMP`: 时间戳（毫秒）
- `ACCESS-PASSPHRASE`: Passphrase
- `Content-Type`: application/json
- `locale`: zh-CN

签名算法：
```
signStr = timestamp + method + endpoint + queryString (GET) 或 + body (POST)
signature = Base64(HMAC-SHA256(signStr, secretKey))
```

## 签名算法验证

### 当前实现（addons/exchange_bitget/service/exchange.go）
```go
// GET请求签名
signStr := timestamp + method + endpoint
if queryStr != "" {
    signStr += "?" + queryStr  // queryStr不包含?，手动添加
}
signature := Base64(HMAC-SHA256(signStr, secretKey))
```

### 参考实现（internal/library/exchange/bitget.go）
```go
// GET请求签名
queryString = "?" + buildQuery(params)  // queryString包含?
preSign := timestamp + method + path + queryString + body
signature := Base64(HMAC-SHA256(preSign, secretKey))
```

**结论**：两种实现方式一致，签名算法正确。

## 验证步骤

1. ✅ 确认代理配置正确（HTTP代理可用）
2. ✅ 确认API路径格式正确（`/api/v2/mix/`）
3. ✅ 确认参数格式正确（`productType: "USDT-FUTURES"`）
4. ✅ 签名算法正确（与参考实现一致）
5. ⚠️ 需要确认：API Key、Secret Key、Passphrase是否正确
6. ⚠️ 需要确认：代理是否正常工作（HTTP代理已验证可用）

## 下一步

如果仍有问题，建议：
1. 查看Bitget官方API文档确认最新格式
2. 使用Postman等工具直接测试API
3. 检查代理是否正常工作
4. 查看后端日志获取详细错误信息

