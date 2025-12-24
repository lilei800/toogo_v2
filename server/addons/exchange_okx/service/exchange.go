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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hotgo/addons/exchange"

	"github.com/gogf/gf/v2/errors/gerror"
)

// OKXExchange OKX交易所实现
type OKXExchange struct {
	BaseURL    string
	ApiKey     string
	SecretKey  string
	Passphrase string
	Client     *http.Client
}

// NewOKXExchange 创建OKX交易所实例
func NewOKXExchange(apiKey, secretKey, passphrase string, proxyTransport *http.Transport) *OKXExchange {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	if proxyTransport != nil {
		client.Transport = proxyTransport
	}

	return &OKXExchange{
		BaseURL:    "https://www.okx.com",
		ApiKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
		Client:     client,
	}
}

// GetName 获取交易所名称
func (e *OKXExchange) GetName() string {
	return "okx"
}

// TestConnection 测试连接
func (e *OKXExchange) TestConnection(ctx context.Context) (balance string, err error) {
	// 调用获取账户余额接口测试连接
	endpoint := "/api/v5/account/balance"

	resp, err := e.request(ctx, "GET", endpoint, "")
	if err != nil {
		return "", err
	}

	// 解析响应
	var result struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			Details []struct {
				Ccy      string `json:"ccy"`
				AvailBal string `json:"availBal"`
			} `json:"details"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", gerror.Wrap(err, "解析响应失败")
	}

	if result.Code != "0" {
		return "", gerror.Newf("API返回错误: %s", result.Msg)
	}

	// 获取USDT余额
	if len(result.Data) > 0 {
		for _, detail := range result.Data[0].Details {
			if detail.Ccy == "USDT" {
				balance = detail.AvailBal + " USDT"
				return balance, nil
			}
		}
	}

	balance = "0 USDT"
	return balance, nil
}

// GetTicker 获取行情
func (e *OKXExchange) GetTicker(ctx context.Context, symbol string) (*exchange.Ticker, error) {
	// OKX symbol格式: BTC-USDT-SWAP
	okxSymbol := convertSymbolToOKX(symbol)

	endpoint := "/api/v5/market/ticker?instId=" + okxSymbol

	resp, err := e.request(ctx, "GET", endpoint, "")
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			InstId  string `json:"instId"`
			Last    string `json:"last"`
			High24h string `json:"high24h"`
			Low24h  string `json:"low24h"`
			Vol24h  string `json:"vol24h"`
			SodUtc0 string `json:"sodUtc0"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, gerror.Wrap(err, "解析响应失败")
	}

	if result.Code != "0" {
		return nil, gerror.Newf("API返回错误: %s", result.Msg)
	}

	if len(result.Data) == 0 {
		return nil, gerror.New("未获取到行情数据")
	}

	data := result.Data[0]

	// 转换数据类型
	lastPrice, _ := strconv.ParseFloat(data.Last, 64)
	high24h, _ := strconv.ParseFloat(data.High24h, 64)
	low24h, _ := strconv.ParseFloat(data.Low24h, 64)
	volume24h, _ := strconv.ParseFloat(data.Vol24h, 64)
	sodUtc0, _ := strconv.ParseFloat(data.SodUtc0, 64)

	// 计算涨跌幅
	var change24h float64
	if sodUtc0 > 0 {
		change24h = ((lastPrice - sodUtc0) / sodUtc0) * 100
	}

	ticker := &exchange.Ticker{
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
func (e *OKXExchange) GetKline(ctx context.Context, symbol string, interval string, limit int) ([]*exchange.Kline, error) {
	// 转换周期格式
	okxInterval := convertIntervalToOKX(interval)

	// 设置默认限制
	if limit <= 0 || limit > 300 {
		limit = 100
	}

	endpoint := fmt.Sprintf("/api/v5/market/candles?instId=%s&bar=%s&limit=%d",
		convertSymbolToOKX(symbol), okxInterval, limit)

	resp, err := e.request(ctx, "GET", endpoint, "")
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

	if result.Code != "0" {
		return nil, gerror.Newf("获取K线失败: %s", result.Msg)
	}

	// 转换数据格式
	klines := make([]*exchange.Kline, 0, len(result.Data))

	for _, item := range result.Data {
		if len(item) < 6 {
			continue
		}

		// OKX返回格式: [timestamp, open, high, low, close, volume, ...]
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
			CloseTime: time.UnixMilli(timestamp + getIntervalMillisOKX(interval)),
		}

		klines = append(klines, kline)
	}

	return klines, nil
}

// PlaceOrder 下单
func (e *OKXExchange) PlaceOrder(ctx context.Context, order *exchange.OrderRequest) (string, error) {
	endpoint := "/api/v5/trade/order"

	// 构建请求体
	body := map[string]interface{}{
		"instId":  convertSymbolToOKX(order.Symbol),
		"tdMode":  convertMarginModeToOKX(order.MarginMode),
		"side":    convertSideToOKX(order.Side),
		"ordType": convertOrderTypeToOKX(order.OrderType),
		"sz":      strconv.FormatFloat(order.Quantity, 'f', -1, 64),
	}

	// 限价单需要价格
	if order.OrderType == "limit" {
		body["px"] = strconv.FormatFloat(order.Price, 'f', -1, 64)
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return "", gerror.Wrap(err, "构建请求体失败")
	}

	resp, err := e.request(ctx, "POST", endpoint, string(bodyJSON))
	if err != nil {
		return "", err
	}

	// 解析响应
	var result struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			OrdId   string `json:"ordId"`
			ClOrdId string `json:"clOrdId"`
			SCode   string `json:"sCode"`
			SMsg    string `json:"sMsg"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", gerror.Wrap(err, "解析响应失败")
	}

	if result.Code != "0" {
		return "", gerror.Newf("下单失败: %s", result.Msg)
	}

	if len(result.Data) == 0 || result.Data[0].SCode != "0" {
		return "", gerror.Newf("下单失败: %s", result.Data[0].SMsg)
	}

	return result.Data[0].OrdId, nil
}

// CloseOrder 平仓
func (e *OKXExchange) CloseOrder(ctx context.Context, orderId string) error {
	// 1. 查询订单信息
	orderInfo, err := e.GetOrder(ctx, orderId)
	if err != nil {
		return gerror.Wrap(err, "查询订单失败")
	}

	// 2. 如果订单已完成，返回
	if orderInfo.Status == "filled" || orderInfo.Status == "canceled" {
		return nil
	}

	// 3. 平仓：关闭持仓
	endpoint := "/api/v5/trade/close-position"

	body := map[string]interface{}{
		"instId":  convertSymbolToOKX(orderInfo.Symbol),
		"mgnMode": "isolated", // 逐仓
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return gerror.Wrap(err, "构建平仓请求失败")
	}

	resp, err := e.request(ctx, "POST", endpoint, string(bodyJSON))
	if err != nil {
		return err
	}

	// 解析响应
	var result struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			InstId  string `json:"instId"`
			PosSide string `json:"posSide"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return gerror.Wrap(err, "解析平仓响应失败")
	}

	if result.Code != "0" {
		return gerror.Newf("平仓失败: %s", result.Msg)
	}

	return nil
}

// GetOrder 获取订单
func (e *OKXExchange) GetOrder(ctx context.Context, orderId string) (*exchange.OrderInfo, error) {
	endpoint := "/api/v5/trade/order?instId=BTC-USDT-SWAP&ordId=" + orderId // TODO: 从订单ID获取symbol

	resp, err := e.request(ctx, "GET", endpoint, "")
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			OrdId     string `json:"ordId"`
			InstId    string `json:"instId"`
			Side      string `json:"side"`
			OrdType   string `json:"ordType"`
			Px        string `json:"px"`
			Sz        string `json:"sz"`
			AccFillSz string `json:"accFillSz"`
			AvgPx     string `json:"avgPx"`
			State     string `json:"state"`
			CTime     string `json:"cTime"`
			UTime     string `json:"uTime"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, gerror.Wrap(err, "解析响应失败")
	}

	if result.Code != "0" {
		return nil, gerror.Newf("查询订单失败: %s", result.Msg)
	}

	if len(result.Data) == 0 {
		return nil, gerror.New("订单不存在")
	}

	data := result.Data[0]

	price, _ := strconv.ParseFloat(data.Px, 64)
	quantity, _ := strconv.ParseFloat(data.Sz, 64)
	filledQty, _ := strconv.ParseFloat(data.AccFillSz, 64)
	avgPrice, _ := strconv.ParseFloat(data.AvgPx, 64)
	cTime, _ := strconv.ParseInt(data.CTime, 10, 64)
	uTime, _ := strconv.ParseInt(data.UTime, 10, 64)

	order := &exchange.OrderInfo{
		OrderId:    data.OrdId,
		Symbol:     convertSymbolFromOKX(data.InstId),
		Side:       strings.ToLower(data.Side),
		OrderType:  strings.ToLower(data.OrdType),
		Quantity:   quantity,
		Price:      price,
		AvgPrice:   avgPrice,
		Status:     data.State,
		FilledQty:  filledQty,
		CreateTime: time.UnixMilli(cTime),
		UpdateTime: time.UnixMilli(uTime),
	}

	return order, nil
}

// GetPositions 获取持仓
func (e *OKXExchange) GetPositions(ctx context.Context, symbol string) ([]*exchange.Position, error) {
	endpoint := "/api/v5/account/positions"

	if symbol != "" {
		endpoint += "?instId=" + convertSymbolToOKX(symbol)
	}

	resp, err := e.request(ctx, "GET", endpoint, "")
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			InstId  string `json:"instId"`
			PosSide string `json:"posSide"`
			Pos     string `json:"pos"`
			AvgPx   string `json:"avgPx"`
			Lever   string `json:"lever"`
			Upl     string `json:"upl"`
			LiqPx   string `json:"liqPx"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, gerror.Wrap(err, "解析响应失败")
	}

	if result.Code != "0" {
		return nil, gerror.Newf("查询持仓失败: %s", result.Msg)
	}

	positions := make([]*exchange.Position, 0)

	for _, item := range result.Data {
		pos, _ := strconv.ParseFloat(item.Pos, 64)

		// 只返回有持仓的
		if pos == 0 {
			continue
		}

		avgPx, _ := strconv.ParseFloat(item.AvgPx, 64)
		upl, _ := strconv.ParseFloat(item.Upl, 64)
		lever, _ := strconv.Atoi(item.Lever)
		liqPx, _ := strconv.ParseFloat(item.LiqPx, 64)

		side := "long"
		if item.PosSide == "short" {
			side = "short"
		}

		position := &exchange.Position{
			Symbol:           convertSymbolFromOKX(item.InstId),
			Side:             side,
			Size:             pos,
			AvgPrice:         avgPx,
			Leverage:         lever,
			UnrealizedProfit: upl,
			LiquidationPrice: liqPx,
		}

		positions = append(positions, position)
	}

	return positions, nil
}

// request 发送HTTP请求
func (e *OKXExchange) request(ctx context.Context, method, endpoint, body string) ([]byte, error) {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")

	// 构建签名字符串
	signStr := timestamp + method + endpoint + body

	// 计算签名
	signature := e.sign(signStr)

	// 构建请求URL
	url := e.BaseURL + endpoint

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("OK-ACCESS-KEY", e.ApiKey)
	req.Header.Set("OK-ACCESS-SIGN", signature)
	req.Header.Set("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("OK-ACCESS-PASSPHRASE", e.Passphrase)
	req.Header.Set("Content-Type", "application/json")

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

// sign 签名
func (e *OKXExchange) sign(message string) string {
	mac := hmac.New(sha256.New, []byte(e.SecretKey))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// convertSymbolToOKX 转换交易对格式
// BTC_USDT -> BTC-USDT-SWAP
func convertSymbolToOKX(symbol string) string {
	// 将下划线替换为横杠，并添加-SWAP后缀
	okxSymbol := ""
	for _, char := range symbol {
		if char == '_' {
			okxSymbol += "-"
		} else {
			okxSymbol += string(char)
		}
	}
	return okxSymbol + "-SWAP"
}

// convertSymbolFromOKX 从OKX格式转换交易对
// BTC-USDT-SWAP -> BTC_USDT
func convertSymbolFromOKX(symbol string) string {
	// 移除-SWAP后缀
	symbol = strings.TrimSuffix(symbol, "-SWAP")
	// 将横杠替换为下划线
	return strings.ReplaceAll(symbol, "-", "_")
}

// convertSideToOKX 转换方向到OKX格式
func convertSideToOKX(side string) string {
	switch strings.ToLower(side) {
	case "buy", "long":
		return "buy"
	case "sell", "short":
		return "sell"
	default:
		return strings.ToLower(side)
	}
}

// convertOrderTypeToOKX 转换订单类型到OKX格式
func convertOrderTypeToOKX(orderType string) string {
	switch strings.ToLower(orderType) {
	case "market":
		return "market"
	case "limit":
		return "limit"
	default:
		return strings.ToLower(orderType)
	}
}

// convertMarginModeToOKX 转换保证金模式到OKX格式
func convertMarginModeToOKX(marginMode string) string {
	switch strings.ToLower(marginMode) {
	case "isolated":
		return "isolated"
	case "cross":
		return "cross"
	default:
		return "isolated"
	}
}

// convertIntervalToOKX 转换K线周期到OKX格式
func convertIntervalToOKX(interval string) string {
	// OKX格式: 1m, 3m, 5m, 15m, 30m, 1H, 2H, 4H, 6H, 12H, 1D, 1W, 1M, 3M
	intervalMap := map[string]string{
		"1m":  "1m",
		"3m":  "3m",
		"5m":  "5m",
		"15m": "15m",
		"30m": "30m",
		"1h":  "1H",
		"2h":  "2H",
		"4h":  "4H",
		"6h":  "6H",
		"12h": "12H",
		"1d":  "1D",
		"1w":  "1W",
		"1M":  "1M",
		"3M":  "3M",
	}

	if okxInterval, ok := intervalMap[interval]; ok {
		return okxInterval
	}
	return "15m" // 默认15分钟
}

// getIntervalMillisOKX 获取周期的毫秒数
func getIntervalMillisOKX(interval string) int64 {
	intervalMap := map[string]int64{
		"1m":  60 * 1000,
		"3m":  3 * 60 * 1000,
		"5m":  5 * 60 * 1000,
		"15m": 15 * 60 * 1000,
		"30m": 30 * 60 * 1000,
		"1h":  60 * 60 * 1000,
		"2h":  2 * 60 * 60 * 1000,
		"4h":  4 * 60 * 60 * 1000,
		"6h":  6 * 60 * 60 * 1000,
		"12h": 12 * 60 * 60 * 1000,
		"1d":  24 * 60 * 60 * 1000,
		"1w":  7 * 24 * 60 * 60 * 1000,
		"1M":  30 * 24 * 60 * 60 * 1000,
		"3M":  90 * 24 * 60 * 60 * 1000,
	}

	if millis, ok := intervalMap[interval]; ok {
		return millis
	}
	return 15 * 60 * 1000 // 默认15分钟
}
