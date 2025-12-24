// Package service
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hotgo/addons/exchange"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// BinanceExchange Binance交易所实现
type BinanceExchange struct {
	BaseURL   string
	ApiKey    string
	SecretKey string
	Client    *http.Client
}

// NewBinanceExchange 创建Binance交易所实例
func NewBinanceExchange(apiKey, secretKey string, proxyTransport *http.Transport) *BinanceExchange {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	if proxyTransport != nil {
		client.Transport = proxyTransport
		// 注意：这里无法直接打印日志，因为没有 context
	}

	return &BinanceExchange{
		BaseURL:   "https://fapi.binance.com", // USDT合约
		ApiKey:    apiKey,
		SecretKey: secretKey,
		Client:    client,
	}
}

// GetName 获取交易所名称
func (e *BinanceExchange) GetName() string {
	return "binance"
}

// TestConnection 测试连接
func (e *BinanceExchange) TestConnection(ctx context.Context) (balance string, err error) {
	// 调用获取账户余额接口测试连接
	endpoint := "/fapi/v2/balance"

	resp, err := e.request(ctx, "GET", endpoint, nil, true)
	if err != nil {
		return "", err
	}

	// 解析响应
	var result []struct {
		Asset             string `json:"asset"`
		Balance           string `json:"balance"`
		AvailableBalance  string `json:"availableBalance"`
		MaxWithdrawAmount string `json:"maxWithdrawAmount"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", gerror.Wrap(err, "解析响应失败")
	}

	// 获取USDT余额
	for _, account := range result {
		if account.Asset == "USDT" {
			balance = account.AvailableBalance + " USDT"
			return balance, nil
		}
	}

	balance = "0 USDT"
	return balance, nil
}

// GetTicker 获取行情
func (e *BinanceExchange) GetTicker(ctx context.Context, symbol string) (*exchange.Ticker, error) {
	// Binance symbol格式: BTCUSDT
	binanceSymbol := convertSymbolToBinance(symbol)

	endpoint := "/fapi/v1/ticker/24hr"
	params := map[string]string{
		"symbol": binanceSymbol,
	}

	resp, err := e.request(ctx, "GET", endpoint, params, false)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result struct {
		Symbol             string `json:"symbol"`
		LastPrice          string `json:"lastPrice"`
		HighPrice          string `json:"highPrice"`
		LowPrice           string `json:"lowPrice"`
		Volume             string `json:"volume"`
		PriceChangePercent string `json:"priceChangePercent"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, gerror.Wrap(err, "解析响应失败")
	}

	// 转换数据类型
	lastPrice, _ := strconv.ParseFloat(result.LastPrice, 64)
	highPrice, _ := strconv.ParseFloat(result.HighPrice, 64)
	lowPrice, _ := strconv.ParseFloat(result.LowPrice, 64)
	volume, _ := strconv.ParseFloat(result.Volume, 64)
	priceChange, _ := strconv.ParseFloat(result.PriceChangePercent, 64)

	ticker := &exchange.Ticker{
		Symbol:    symbol,
		LastPrice: lastPrice,
		High24h:   highPrice,
		Low24h:    lowPrice,
		Volume24h: volume,
		Change24h: priceChange,
		Timestamp: time.Now(),
	}

	return ticker, nil
}

// GetKline 获取K线数据
func (e *BinanceExchange) GetKline(ctx context.Context, symbol string, interval string, limit int) ([]*exchange.Kline, error) {
	endpoint := "/fapi/v1/klines"

	// 转换周期格式
	binanceInterval := convertIntervalToBinance(interval)

	// 设置默认限制
	if limit <= 0 || limit > 1500 {
		limit = 100
	}

	params := map[string]string{
		"symbol":   convertSymbolToBinance(symbol),
		"interval": binanceInterval,
		"limit":    strconv.Itoa(limit),
	}

	resp, err := e.request(ctx, "GET", endpoint, params, false)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result [][]interface{}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, gerror.Wrap(err, "解析响应失败")
	}

	// 转换数据格式
	klines := make([]*exchange.Kline, 0, len(result))

	for _, item := range result {
		if len(item) < 6 {
			continue
		}

		// Binance返回格式: [openTime, open, high, low, close, volume, closeTime, ...]
		openTime := int64(item[0].(float64))
		open, _ := strconv.ParseFloat(item[1].(string), 64)
		high, _ := strconv.ParseFloat(item[2].(string), 64)
		low, _ := strconv.ParseFloat(item[3].(string), 64)
		closePrice, _ := strconv.ParseFloat(item[4].(string), 64)
		volume, _ := strconv.ParseFloat(item[5].(string), 64)
		closeTime := int64(item[6].(float64))

		kline := &exchange.Kline{
			OpenTime:  time.UnixMilli(openTime),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     closePrice,
			Volume:    volume,
			CloseTime: time.UnixMilli(closeTime),
		}

		klines = append(klines, kline)
	}

	return klines, nil
}

// PlaceOrder 下单
func (e *BinanceExchange) PlaceOrder(ctx context.Context, order *exchange.OrderRequest) (string, error) {
	endpoint := "/fapi/v1/order"

	// 构建参数
	params := map[string]string{
		"symbol":   convertSymbolToBinance(order.Symbol),
		"side":     convertSideToBinance(order.Side),
		"type":     strings.ToUpper(order.OrderType),
		"quantity": strconv.FormatFloat(order.Quantity, 'f', -1, 64),
	}

	// 限价单需要价格
	if order.OrderType == "limit" {
		params["price"] = strconv.FormatFloat(order.Price, 'f', -1, 64)
		params["timeInForce"] = "GTC"
	}

	resp, err := e.request(ctx, "POST", endpoint, params, true)
	if err != nil {
		return "", err
	}

	// 解析响应
	var result struct {
		OrderId       int64  `json:"orderId"`
		Symbol        string `json:"symbol"`
		Status        string `json:"status"`
		ClientOrderId string `json:"clientOrderId"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", gerror.Wrap(err, "解析响应失败")
	}

	return strconv.FormatInt(result.OrderId, 10), nil
}

// CloseOrder 平仓
func (e *BinanceExchange) CloseOrder(ctx context.Context, orderId string) error {
	// 1. 查询订单信息
	orderInfo, err := e.GetOrder(ctx, orderId)
	if err != nil {
		return gerror.Wrap(err, "查询订单失败")
	}

	// 2. 如果订单已完成，返回
	if orderInfo.Status == "FILLED" || orderInfo.Status == "CANCELED" {
		return nil
	}

	// 3. 平仓：开反向单
	endpoint := "/fapi/v1/order"

	// 反向方向
	closeSide := "SELL"
	if orderInfo.Side == "sell" {
		closeSide = "BUY"
	}

	params := map[string]string{
		"symbol":     convertSymbolToBinance(orderInfo.Symbol),
		"side":       closeSide,
		"type":       "MARKET",
		"quantity":   strconv.FormatFloat(orderInfo.FilledQty, 'f', -1, 64),
		"reduceOnly": "true", // 只减仓
	}

	_, err = e.request(ctx, "POST", endpoint, params, true)
	if err != nil {
		return err
	}

	return nil
}

// GetOrder 获取订单
func (e *BinanceExchange) GetOrder(ctx context.Context, orderId string) (*exchange.OrderInfo, error) {
	endpoint := "/fapi/v1/order"

	params := map[string]string{
		"symbol":  "BTCUSDT", // TODO: 从订单ID获取symbol
		"orderId": orderId,
	}

	resp, err := e.request(ctx, "GET", endpoint, params, true)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result struct {
		OrderId     int64  `json:"orderId"`
		Symbol      string `json:"symbol"`
		Side        string `json:"side"`
		Type        string `json:"type"`
		Price       string `json:"price"`
		OrigQty     string `json:"origQty"`
		ExecutedQty string `json:"executedQty"`
		AvgPrice    string `json:"avgPrice"`
		Status      string `json:"status"`
		Time        int64  `json:"time"`
		UpdateTime  int64  `json:"updateTime"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, gerror.Wrap(err, "解析响应失败")
	}

	price, _ := strconv.ParseFloat(result.Price, 64)
	quantity, _ := strconv.ParseFloat(result.OrigQty, 64)
	filledQty, _ := strconv.ParseFloat(result.ExecutedQty, 64)
	avgPrice, _ := strconv.ParseFloat(result.AvgPrice, 64)

	order := &exchange.OrderInfo{
		OrderId:    strconv.FormatInt(result.OrderId, 10),
		Symbol:     result.Symbol,
		Side:       strings.ToLower(result.Side),
		OrderType:  strings.ToLower(result.Type),
		Quantity:   quantity,
		Price:      price,
		AvgPrice:   avgPrice,
		Status:     result.Status,
		FilledQty:  filledQty,
		CreateTime: time.UnixMilli(result.Time),
		UpdateTime: time.UnixMilli(result.UpdateTime),
	}

	return order, nil
}

// GetPositions 获取持仓
func (e *BinanceExchange) GetPositions(ctx context.Context, symbol string) ([]*exchange.Position, error) {
	endpoint := "/fapi/v2/positionRisk"

	params := map[string]string{}
	if symbol != "" {
		params["symbol"] = convertSymbolToBinance(symbol)
	}

	resp, err := e.request(ctx, "GET", endpoint, params, true)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result []struct {
		Symbol           string `json:"symbol"`
		PositionAmt      string `json:"positionAmt"`
		EntryPrice       string `json:"entryPrice"`
		UnRealizedProfit string `json:"unRealizedProfit"`
		Leverage         string `json:"leverage"`
		LiquidationPrice string `json:"liquidationPrice"`
		PositionSide     string `json:"positionSide"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, gerror.Wrap(err, "解析响应失败")
	}

	positions := make([]*exchange.Position, 0)

	for _, item := range result {
		posAmt, _ := strconv.ParseFloat(item.PositionAmt, 64)

		// 只返回有持仓的
		if posAmt == 0 {
			continue
		}

		entryPrice, _ := strconv.ParseFloat(item.EntryPrice, 64)
		unrealizedPL, _ := strconv.ParseFloat(item.UnRealizedProfit, 64)
		leverage, _ := strconv.Atoi(item.Leverage)
		liquidationPrice, _ := strconv.ParseFloat(item.LiquidationPrice, 64)

		side := "long"
		if posAmt < 0 {
			side = "short"
			posAmt = -posAmt
		}

		position := &exchange.Position{
			Symbol:           item.Symbol,
			Side:             side,
			Size:             posAmt,
			AvgPrice:         entryPrice,
			Leverage:         leverage,
			UnrealizedProfit: unrealizedPL,
			LiquidationPrice: liquidationPrice,
		}

		positions = append(positions, position)
	}

	return positions, nil
}

// request 发送HTTP请求
func (e *BinanceExchange) request(ctx context.Context, method, endpoint string, params map[string]string, needSign bool) ([]byte, error) {
	// 日志：记录请求信息
	hasProxy := e.Client.Transport != nil
	
	// 构建请求URL
	url := e.BaseURL + endpoint

	// 添加时间戳（签名必需）
	if needSign {
		if params == nil {
			params = make(map[string]string)
		}
		params["timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	}

	// 构建查询参数
	if len(params) > 0 {
		url += "?"
		query := ""
		for k, v := range params {
			if query != "" {
				query += "&"
			}
			query += k + "=" + v
		}
		url += query

		// 签名
		if needSign {
			signature := e.sign(query)
			url += "&signature=" + signature
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("X-MBX-APIKEY", e.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// 日志：记录请求详情
	var proxyInfo string
	if hasProxy {
		proxyInfo = "✅ 使用代理"
	} else {
		proxyInfo = "⚠️ 直连（无代理）"
	}
	g.Log().Infof(ctx, "[Binance] 发送请求: %s %s [%s]", method, endpoint, proxyInfo)

	// 发送请求
	resp, err := e.Client.Do(req)
	if err != nil {
		return nil, gerror.Wrapf(err, "[Binance] 请求失败 [%s]", proxyInfo)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, gerror.Wrap(err, "读取响应失败")
	}

	if resp.StatusCode != 200 {
		// 解析错误信息
		var errResp struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}
		json.Unmarshal(body, &errResp)
		return nil, gerror.Newf("API错误[%d]: %s", errResp.Code, errResp.Msg)
	}

	return body, nil
}

// sign 签名
func (e *BinanceExchange) sign(message string) string {
	mac := hmac.New(sha256.New, []byte(e.SecretKey))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// convertSymbolToBinance 转换交易对格式
// BTC_USDT -> BTCUSDT
func convertSymbolToBinance(symbol string) string {
	// 移除下划线
	binanceSymbol := ""
	for _, char := range symbol {
		if char != '_' {
			binanceSymbol += string(char)
		}
	}
	return binanceSymbol
}

// convertSideToBinance 转换方向到Binance格式
func convertSideToBinance(side string) string {
	switch strings.ToLower(side) {
	case "buy", "long":
		return "BUY"
	case "sell", "short":
		return "SELL"
	default:
		return strings.ToUpper(side)
	}
}

// convertIntervalToBinance 转换K线周期到Binance格式
func convertIntervalToBinance(interval string) string {
	// Binance格式: 1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d, 3d, 1w, 1M
	intervalMap := map[string]string{
		"1m":  "1m",
		"3m":  "3m",
		"5m":  "5m",
		"15m": "15m",
		"30m": "30m",
		"1h":  "1h",
		"2h":  "2h",
		"4h":  "4h",
		"6h":  "6h",
		"8h":  "8h",
		"12h": "12h",
		"1d":  "1d",
		"3d":  "3d",
		"1w":  "1w",
		"1M":  "1M",
	}

	if binanceInterval, ok := intervalMap[interval]; ok {
		return binanceInterval
	}
	return "15m" // 默认15分钟
}
