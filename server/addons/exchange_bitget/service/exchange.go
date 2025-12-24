// Package service
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"hotgo/addons/exchange"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// 使用共享的exchange包中的接口和数据结构定义

// BitgetExchange Bitget交易所实现
type BitgetExchange struct {
	BaseURL    string
	ApiKey     string
	SecretKey  string
	Passphrase string
	Client     *http.Client
}

// NewBitgetExchange 创建Bitget交易所实例
func NewBitgetExchange(apiKey, secretKey, passphrase string, proxyTransport *http.Transport) *BitgetExchange {
	// 增加超时时间到30秒，避免网络延迟导致失败
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if proxyTransport != nil {
		client.Transport = proxyTransport
	}

	return &BitgetExchange{
		BaseURL:    "https://api.bitget.com",
		ApiKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
		Client:     client,
	}
}

// GetName 获取交易所名称
func (e *BitgetExchange) GetName() string {
	return "bitget"
}

// TestConnection 测试连接
func (e *BitgetExchange) TestConnection(ctx context.Context) (balance string, err error) {
	// 调用获取账户余额接口测试连接（使用V2 API）
	// Bitget V2 API路径格式：/api/v2/mix/account/accounts
	endpoint := "/api/v2/mix/account/accounts"

	// V2 API的productType值：USDT-FUTURES 表示USDT合约
	params := map[string]string{
		"productType": "USDT-FUTURES", // V2 API格式（参考 internal/library/exchange/bitget.go）
	}

	g.Log().Infof(ctx, "[Bitget] 开始测试连接: endpoint=%s, productType=USDT-FUTURES, BaseURL=%s",
		endpoint, e.BaseURL)

	// 检查必要的配置
	if e.ApiKey == "" {
		return "", gerror.New("API Key为空")
	}
	if e.SecretKey == "" {
		return "", gerror.New("Secret Key为空")
	}
	if e.Passphrase == "" {
		return "", gerror.New("Passphrase为空（Bitget必填）")
	}

	resp, err := e.request(ctx, "GET", endpoint, params)
	if err != nil {
		g.Log().Errorf(ctx, "[Bitget] 测试连接请求失败: %v", err)
		return "", err // 直接返回错误，让上层处理
	}

	// 解析响应
	var result struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			MarginCoin string `json:"marginCoin"`
			Available  string `json:"available"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		g.Log().Errorf(ctx, "[Bitget] 解析响应失败: %v, Response: %s", err, string(resp))
		return "", gerror.Wrapf(err, "解析响应失败: %s", string(resp))
	}

	if result.Code != "00000" {
		g.Log().Errorf(ctx, "[Bitget] API返回错误: code=%s, msg=%s", result.Code, result.Msg)
		return "", gerror.Newf("API返回错误 [%s]: %s", result.Code, result.Msg)
	}

	// 获取USDT余额
	for _, account := range result.Data {
		if account.MarginCoin == "USDT" {
			balance = account.Available + " USDT"
			g.Log().Infof(ctx, "[Bitget] 测试连接成功: balance=%s", balance)
			return balance, nil
		}
	}

	balance = "0 USDT"
	g.Log().Infof(ctx, "[Bitget] 测试连接成功: 未找到USDT余额，返回0")
	return balance, nil
}

// GetTicker 获取行情
func (e *BitgetExchange) GetTicker(ctx context.Context, symbol string) (ticker *exchange.Ticker, err error) {
	endpoint := "/api/v2/mix/market/ticker"
	params := map[string]string{
		"symbol":      symbol,
		"productType": "USDT-FUTURES",
	}

	resp, err := e.request(ctx, "GET", endpoint, params)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Symbol     string `json:"symbol"`
			Last       string `json:"last"`
			High24h    string `json:"high24h"`
			Low24h     string `json:"low24h"`
			BaseVolume string `json:"baseVolume"`
			Change24h  string `json:"change24h"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, gerror.Wrap(err, "解析响应失败")
	}

	if result.Code != "00000" {
		return nil, gerror.Newf("API返回错误: %s", result.Msg)
	}

	// 转换数据类型
	lastPrice, _ := strconv.ParseFloat(result.Data.Last, 64)
	high24h, _ := strconv.ParseFloat(result.Data.High24h, 64)
	low24h, _ := strconv.ParseFloat(result.Data.Low24h, 64)
	volume24h, _ := strconv.ParseFloat(result.Data.BaseVolume, 64)
	change24h, _ := strconv.ParseFloat(result.Data.Change24h, 64)

	ticker = &exchange.Ticker{
		Symbol:    symbol,
		LastPrice: lastPrice,
		High24h:   high24h,
		Low24h:    low24h,
		Volume24h: volume24h,
		Change24h: change24h,
		Timestamp: time.Now(),
	}

	return ticker, nil
}

// GetKline 获取K线数据
func (e *BitgetExchange) GetKline(ctx context.Context, symbol string, interval string, limit int) (klines []*exchange.Kline, err error) {
	endpoint := "/api/v2/mix/market/candles"

	// 转换周期格式
	bitgetInterval := convertIntervalToBitget(interval)

	// 设置默认限制
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	params := map[string]string{
		"symbol":      symbol,
		"productType": "USDT-FUTURES",
		"granularity": bitgetInterval,
		"limit":       strconv.Itoa(limit),
	}

	resp, err := e.request(ctx, "GET", endpoint, params)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result struct {
		Code string     `json:"code"`
		Msg  string     `json:"msg"`
		Data [][]string `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, gerror.Wrap(err, "解析响应失败")
	}

	if result.Code != "00000" {
		return nil, gerror.Newf("获取K线失败: %s", result.Msg)
	}

	// 转换数据格式
	klines = make([]*exchange.Kline, 0, len(result.Data))

	for _, item := range result.Data {
		if len(item) < 6 {
			continue
		}

		// Bitget返回格式: [timestamp, open, high, low, close, volume, ...]
		timestamp, _ := strconv.ParseInt(item[0], 10, 64)
		open, _ := strconv.ParseFloat(item[1], 64)
		high, _ := strconv.ParseFloat(item[2], 64)
		low, _ := strconv.ParseFloat(item[3], 64)
		closePrice, _ := strconv.ParseFloat(item[4], 64)
		volume, _ := strconv.ParseFloat(item[5], 64)

		kline := &exchange.Kline{
			OpenTime:  time.UnixMilli(timestamp),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     closePrice,
			Volume:    volume,
			CloseTime: time.UnixMilli(timestamp + getIntervalMillis(interval)),
		}

		klines = append(klines, kline)
	}

	return klines, nil
}

// PlaceOrder 下单
func (e *BitgetExchange) PlaceOrder(ctx context.Context, order *exchange.OrderRequest) (orderId string, err error) {
	endpoint := "/api/v2/mix/order/place-order" // V2 API使用连字符格式

	// 构建请求体
	body := map[string]interface{}{
		"symbol":      order.Symbol,
		"productType": "USDT-FUTURES",
		"marginCoin":  "USDT",
		"marginMode":  "isolated",
		"side":        convertSideToBitget(order.Side),
		"orderType":   convertOrderTypeToBitget(order.OrderType),
		"timeInForce": "normal",
	}

	// 计算交易数量
	if order.OrderType == "market" {
		// 市价单使用资金数量
		body["size"] = strconv.FormatFloat(order.Quantity, 'f', -1, 64)
	} else {
		// 限价单使用合约数量
		body["size"] = strconv.FormatFloat(order.Quantity, 'f', -1, 64)
		body["price"] = strconv.FormatFloat(order.Price, 'f', -1, 64)
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return "", gerror.Wrap(err, "构建请求体失败")
	}

	resp, err := e.requestWithBody(ctx, "POST", endpoint, bodyJSON)
	if err != nil {
		return "", err
	}

	// 解析响应
	var result struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			OrderId   string `json:"orderId"`
			ClientOid string `json:"clientOid"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", gerror.Wrap(err, "解析响应失败")
	}

	if result.Code != "00000" {
		return "", gerror.Newf("下单失败: %s", result.Msg)
	}

	g.Log().Infof(ctx, "Bitget下单成功: OrderId=%s, Symbol=%s, Side=%s", result.Data.OrderId, order.Symbol, order.Side)

	return result.Data.OrderId, nil
}

// CloseOrder 平仓（通过反向开单实现）
func (e *BitgetExchange) CloseOrder(ctx context.Context, orderId string) error {
	// 1. 先查询订单信息
	orderInfo, err := e.GetOrder(ctx, orderId)
	if err != nil {
		return gerror.Wrap(err, "查询订单失败")
	}

	// 2. 如果订单已平仓，直接返回
	if orderInfo.Status == "filled" || orderInfo.Status == "cancelled" {
		return nil
	}

	// 3. 平仓：开反向单
	endpoint := "/api/v2/mix/order/place-order" // V2 API使用连字符格式

	// 反向方向
	closeSide := "close_long"
	if orderInfo.Side == "sell" {
		closeSide = "close_short"
	}

	body := map[string]interface{}{
		"symbol":      orderInfo.Symbol,
		"productType": "USDT-FUTURES",
		"marginCoin":  "USDT",
		"marginMode":  "isolated",
		"side":        closeSide,
		"tradeSide":   "close",
		"orderType":   "market", // 市价平仓
		"size":        strconv.FormatFloat(orderInfo.FilledQty, 'f', -1, 64),
		"timeInForce": "normal",
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return gerror.Wrap(err, "构建平仓请求失败")
	}

	resp, err := e.requestWithBody(ctx, "POST", endpoint, bodyJSON)
	if err != nil {
		return err
	}

	// 解析响应
	var result struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			OrderId string `json:"orderId"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return gerror.Wrap(err, "解析平仓响应失败")
	}

	if result.Code != "00000" {
		return gerror.Newf("平仓失败: %s", result.Msg)
	}

	g.Log().Infof(ctx, "Bitget平仓成功: OrderId=%s, CloseOrderId=%s", orderId, result.Data.OrderId)

	return nil
}

// GetOrder 获取订单
func (e *BitgetExchange) GetOrder(ctx context.Context, orderId string) (order *exchange.OrderInfo, err error) {
	// V2 API使用历史订单接口查询单个订单
	endpoint := "/api/v2/mix/order/history" // 使用历史订单接口查询
	params := map[string]string{
		"productType": "USDT-FUTURES", // V2 API格式
		"symbol":      "BTCUSDT",      // TODO: 从订单ID获取symbol
		"orderId":     orderId,
	}

	resp, err := e.request(ctx, "GET", endpoint, params)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Symbol       string `json:"symbol"`
			OrderId      string `json:"orderId"`
			Side         string `json:"side"`
			OrderType    string `json:"orderType"`
			Price        string `json:"price"`
			Size         string `json:"size"`
			FilledQty    string `json:"filledQty"`
			AveragePrice string `json:"priceAvg"`
			Status       string `json:"status"`
			CTime        string `json:"cTime"`
			UTime        string `json:"uTime"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, gerror.Wrap(err, "解析响应失败")
	}

	if result.Code != "00000" {
		return nil, gerror.Newf("查询订单失败: %s", result.Msg)
	}

	data := result.Data

	// 转换数据
	price, _ := strconv.ParseFloat(data.Price, 64)
	quantity, _ := strconv.ParseFloat(data.Size, 64)
	filledQty, _ := strconv.ParseFloat(data.FilledQty, 64)
	avgPrice, _ := strconv.ParseFloat(data.AveragePrice, 64)

	// 转换时间
	cTime, _ := strconv.ParseInt(data.CTime, 10, 64)
	uTime, _ := strconv.ParseInt(data.UTime, 10, 64)

	order = &exchange.OrderInfo{
		OrderId:    data.OrderId,
		Symbol:     data.Symbol,
		Side:       convertSideFromBitget(data.Side),
		OrderType:  convertOrderTypeFromBitget(data.OrderType),
		Quantity:   quantity,
		Price:      price,
		AvgPrice:   avgPrice,
		Status:     data.Status,
		FilledQty:  filledQty,
		CreateTime: time.UnixMilli(cTime),
		UpdateTime: time.UnixMilli(uTime),
	}

	return order, nil
}

// GetPositions 获取持仓
func (e *BitgetExchange) GetPositions(ctx context.Context, symbol string) (positions []*exchange.Position, err error) {
	endpoint := "/api/v2/mix/position/all-position" // V2 API使用连字符格式
	params := map[string]string{
		"productType": "USDT-FUTURES", // V2 API格式
	}

	if symbol != "" {
		params["symbol"] = symbol
	}

	resp, err := e.request(ctx, "GET", endpoint, params)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			Symbol           string `json:"symbol"`
			MarginCoin       string `json:"marginCoin"`
			HoldSide         string `json:"holdSide"`
			OpenPriceAvg     string `json:"openPriceAvg"`
			Leverage         string `json:"leverage"`
			Total            string `json:"total"`
			Available        string `json:"available"`
			UnrealizedPL     string `json:"unrealizedPL"`
			LiquidationPrice string `json:"liquidationPrice"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, gerror.Wrap(err, "解析响应失败")
	}

	if result.Code != "00000" {
		return nil, gerror.Newf("查询持仓失败: %s", result.Msg)
	}

	positions = make([]*exchange.Position, 0, len(result.Data))

	for _, item := range result.Data {
		// 只返回有持仓的
		total, _ := strconv.ParseFloat(item.Total, 64)
		if total <= 0 {
			continue
		}

		avgPrice, _ := strconv.ParseFloat(item.OpenPriceAvg, 64)
		leverage, _ := strconv.Atoi(item.Leverage)
		unrealizedPL, _ := strconv.ParseFloat(item.UnrealizedPL, 64)
		liquidationPrice, _ := strconv.ParseFloat(item.LiquidationPrice, 64)

		// 计算保证金：持仓价值 / 杠杆
		positionValue := total * avgPrice
		margin := 0.0
		if leverage > 0 {
			margin = positionValue / float64(leverage)
		}

		position := &exchange.Position{
			Symbol:           item.Symbol,
			Side:             item.HoldSide,
			Size:             total,
			AvgPrice:         avgPrice,
			Leverage:         leverage,
			Margin:           margin,
			UnrealizedProfit: unrealizedPL,
			LiquidationPrice: liquidationPrice,
		}

		positions = append(positions, position)
	}

	return positions, nil
}

// request 发送HTTP请求
func (e *BitgetExchange) request(ctx context.Context, method, endpoint string, params map[string]string) ([]byte, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// 构建查询字符串（按key排序，Bitget API要求）
	var queryStr string
	if method == "GET" && len(params) > 0 {
		// 对参数key进行排序
		keys := make([]string, 0, len(params))
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys) // 使用标准库排序

		// 构建查询字符串
		for i, k := range keys {
			if i > 0 {
				queryStr += "&"
			}
			queryStr += k + "=" + params[k]
		}
	}

	// 构建请求URL
	url := e.BaseURL + endpoint
	if queryStr != "" {
		url += "?" + queryStr
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	// 计算签名（Bitget要求: timestamp + method + endpoint + queryString）
	signStr := timestamp + method + endpoint
	if queryStr != "" {
		signStr += "?" + queryStr
	}

	signature := e.sign(signStr)

	// 记录调试信息（仅在调试模式下）
	if g.Cfg().MustGet(ctx, "exchange.debug", false).Bool() {
		g.Log().Debugf(ctx, "[Bitget] Request: %s %s, SignStr: %s", method, url, signStr)
	}

	// 设置请求头
	req.Header.Set("ACCESS-KEY", e.ApiKey)
	req.Header.Set("ACCESS-SIGN", signature)
	req.Header.Set("ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("ACCESS-PASSPHRASE", e.Passphrase)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("locale", "zh-CN")

	// 发送请求
	g.Log().Infof(ctx, "[Bitget] 发送请求: %s %s", method, url)
	resp, err := e.Client.Do(req)
	if err != nil {
		// 详细记录错误信息
		errorDetail := err.Error()
		g.Log().Errorf(ctx, "[Bitget] 请求失败: %v, URL: %s, Method: %s, Error: %s, ApiKey前4位: %s",
			err, url, method, errorDetail,
			func() string {
				if len(e.ApiKey) > 4 {
					return e.ApiKey[:4]
				}
				return "N/A"
			}())

		// 根据错误类型提供更友好的错误信息
		errorMsg := "网络请求失败"
		if strings.Contains(errorDetail, "timeout") || strings.Contains(errorDetail, "deadline exceeded") {
			errorMsg = "请求超时，请检查网络连接或代理设置"
		} else if strings.Contains(errorDetail, "no such host") || strings.Contains(errorDetail, "DNS") {
			errorMsg = "DNS解析失败，无法连接到Bitget服务器"
		} else if strings.Contains(errorDetail, "connection refused") {
			errorMsg = "连接被拒绝，请检查代理设置或网络连接"
		} else if strings.Contains(errorDetail, "proxy") {
			errorMsg = "代理连接失败"
		} else if strings.Contains(errorDetail, "certificate") || strings.Contains(errorDetail, "TLS") {
			errorMsg = "SSL证书验证失败"
		} else if strings.Contains(errorDetail, "network is unreachable") {
			errorMsg = "网络不可达，请检查网络连接"
		}

		return nil, gerror.Newf("%s: %s", errorMsg, errorDetail)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		g.Log().Errorf(ctx, "[Bitget] 读取响应失败: %v", err)
		return nil, gerror.Wrap(err, "读取响应失败")
	}

	if resp.StatusCode != 200 {
		g.Log().Errorf(ctx, "[Bitget] HTTP错误: %d, Body: %s", resp.StatusCode, string(body))
		return nil, gerror.Newf("HTTP错误: %d, Body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// sign 签名
func (e *BitgetExchange) sign(message string) string {
	mac := hmac.New(sha256.New, []byte(e.SecretKey))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// requestWithBody 发送带Body的HTTP请求
func (e *BitgetExchange) requestWithBody(ctx context.Context, method, endpoint string, body []byte) ([]byte, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// 计算签名
	signStr := timestamp + method + endpoint + string(body)
	signature := e.sign(signStr)

	// 构建请求URL
	url := e.BaseURL + endpoint

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("ACCESS-KEY", e.ApiKey)
	req.Header.Set("ACCESS-SIGN", signature)
	req.Header.Set("ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("ACCESS-PASSPHRASE", e.Passphrase)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("locale", "zh-CN")

	// 发送请求
	resp, err := e.Client.Do(req)
	if err != nil {
		return nil, gerror.Wrap(err, "请求失败")
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, gerror.Wrap(err, "读取响应失败")
	}

	if resp.StatusCode != 200 {
		return nil, gerror.Newf("HTTP错误: %d, Body: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// convertSideToBitget 转换方向到Bitget格式
func convertSideToBitget(side string) string {
	switch side {
	case "buy", "long":
		return "open_long"
	case "sell", "short":
		return "open_short"
	default:
		return side
	}
}

// convertSideFromBitget 从Bitget格式转换方向
func convertSideFromBitget(side string) string {
	switch side {
	case "open_long", "close_short":
		return "buy"
	case "open_short", "close_long":
		return "sell"
	default:
		return side
	}
}

// convertOrderTypeToBitget 转换订单类型到Bitget格式
func convertOrderTypeToBitget(orderType string) string {
	switch orderType {
	case "market":
		return "market"
	case "limit":
		return "limit"
	default:
		return orderType
	}
}

// convertOrderTypeFromBitget 从Bitget格式转换订单类型
func convertOrderTypeFromBitget(orderType string) string {
	switch orderType {
	case "market":
		return "market"
	case "limit":
		return "limit"
	default:
		return orderType
	}
}

// convertIntervalToBitget 转换K线周期到Bitget格式
func convertIntervalToBitget(interval string) string {
	// Bitget格式: 1m, 5m, 15m, 30m, 1H, 4H, 12H, 1D, 1W
	intervalMap := map[string]string{
		"1m":  "1m",
		"5m":  "5m",
		"15m": "15m",
		"30m": "30m",
		"1h":  "1H",
		"4h":  "4H",
		"12h": "12H",
		"1d":  "1D",
		"1w":  "1W",
	}

	if bitgetInterval, ok := intervalMap[interval]; ok {
		return bitgetInterval
	}
	return "15m" // 默认15分钟
}

// getIntervalMillis 获取周期的毫秒数
func getIntervalMillis(interval string) int64 {
	intervalMap := map[string]int64{
		"1m":  60 * 1000,
		"5m":  5 * 60 * 1000,
		"15m": 15 * 60 * 1000,
		"30m": 30 * 60 * 1000,
		"1h":  60 * 60 * 1000,
		"4h":  4 * 60 * 60 * 1000,
		"12h": 12 * 60 * 60 * 1000,
		"1d":  24 * 60 * 60 * 1000,
		"1w":  7 * 24 * 60 * 60 * 1000,
	}

	if millis, ok := intervalMap[interval]; ok {
		return millis
	}
	return 15 * 60 * 1000 // 默认15分钟
}
