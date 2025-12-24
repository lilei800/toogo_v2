// Package exchange OKX交易所API（U本位永续 / 逐仓 / 双向持仓）
package exchange

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
)

// OKX OKX交易所（API v5）
// 约束：仅实现 USDT 永续（SWAP）+ 逐仓（isolated）+ 双向持仓（posSide long/short）
type OKX struct {
	config   *Config
	endpoint string

	mu            sync.Mutex
	instrumentCtV map[string]float64 // instId -> ctVal（合约面值，单位：基础币）
}

func NewOKX(config *Config) *OKX {
	// OKX 模拟盘仍使用同域名，通过账号权限区分
	return &OKX{
		config:        config,
		endpoint:      "https://www.okx.com",
		instrumentCtV: make(map[string]float64),
	}
}

func (o *OKX) GetName() string { return "okx" }

func (o *OKX) getHttpClient() *gclient.Client {
	client := gclient.New()
	client.SetTimeout(20 * time.Second)
	if o.config.Proxy != nil && o.config.Proxy.Enabled {
		client.SetProxy(o.config.Proxy.GetProxyURL())
	}
	return client
}

func (o *OKX) formatInstId(symbol string) string {
	// 使用统一的Symbol格式化器
	return Formatter.FormatForOKX(symbol) // BTC-USDT-SWAP
}

func (o *OKX) convertBar(interval string) string {
	switch strings.ToLower(interval) {
	case "1m":
		return "1m"
	case "3m":
		return "3m"
	case "5m":
		return "5m"
	case "15m":
		return "15m"
	case "30m":
		return "30m"
	case "1h", "60m":
		return "1H"
	case "2h":
		return "2H"
	case "4h":
		return "4H"
	case "6h":
		return "6H"
	case "12h":
		return "12H"
	case "1d", "1D":
		return "1D"
	default:
		return "1m"
	}
}

// sign OKX V5: Base64(HMAC_SHA256(secret, ts+method+requestPath+body))
func (o *OKX) sign(ts, method, requestPath, body string) string {
	prehash := ts + strings.ToUpper(method) + requestPath + body
	mac := hmac.New(sha256.New, []byte(o.config.SecretKey))
	mac.Write([]byte(prehash))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (o *OKX) signedRequest(ctx context.Context, method, path string, query url.Values, body any) (string, error) {
	requestPath := path
	if query != nil && len(query) > 0 {
		requestPath += "?" + query.Encode()
	}

	bodyStr := ""
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return "", err
		}
		bodyStr = string(b)
	}

	// OKX 对 OK-ACCESS-TIMESTAMP 的格式较严格，使用毫秒级 RFC3339 更兼容：
	// e.g. 2025-12-24T12:34:56.789Z
	ts := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	sign := o.sign(ts, method, requestPath, bodyStr)

	client := o.getHttpClient()
	client.SetHeader("OK-ACCESS-KEY", o.config.ApiKey)
	client.SetHeader("OK-ACCESS-SIGN", sign)
	client.SetHeader("OK-ACCESS-TIMESTAMP", ts)
	client.SetHeader("OK-ACCESS-PASSPHRASE", o.config.Passphrase)
	client.SetHeader("Content-Type", "application/json")

	reqURL := o.endpoint + requestPath
	var resp *gclient.Response
	var err error
	switch strings.ToUpper(method) {
	case "GET":
		resp, err = client.Get(ctx, reqURL)
	case "POST":
		resp, err = client.Post(ctx, reqURL, bodyStr)
	default:
		resp, err = client.Get(ctx, reqURL)
	}
	if err != nil {
		return "", gerror.Wrap(err, "OKX request failed")
	}
	defer resp.Close()

	raw := resp.ReadAllString()
	if resp.StatusCode != 200 {
		return "", gerror.Wrapf(WrapAsAPIError("okx", resp.StatusCode, raw, nil), "[okx] http status=%d path=%s", resp.StatusCode, requestPath)
	}

	j := gjson.New(raw)
	if j.Get("code").String() != "0" {
		return "", gerror.Newf("OKX API error: code=%s msg=%s", j.Get("code").String(), j.Get("msg").String())
	}
	return raw, nil
}

func (o *OKX) publicRequest(ctx context.Context, path string, query url.Values) (string, error) {
	reqURL := o.endpoint + path
	if query != nil && len(query) > 0 {
		reqURL += "?" + query.Encode()
	}
	client := o.getHttpClient()
	resp, err := client.Get(ctx, reqURL)
	if err != nil {
		return "", gerror.Wrap(err, "OKX request failed")
	}
	defer resp.Close()
	raw := resp.ReadAllString()
	if resp.StatusCode != 200 {
		return "", gerror.Wrapf(WrapAsAPIError("okx", resp.StatusCode, raw, nil), "[okx] http status=%d path=%s", resp.StatusCode, path)
	}
	j := gjson.New(raw)
	if j.Get("code").String() != "0" {
		return "", gerror.Newf("OKX API error: code=%s msg=%s", j.Get("code").String(), j.Get("msg").String())
	}
	return raw, nil
}

func (o *OKX) getCtVal(ctx context.Context, instId string) (float64, error) {
	o.mu.Lock()
	if v, ok := o.instrumentCtV[instId]; ok && v > 0 {
		o.mu.Unlock()
		return v, nil
	}
	o.mu.Unlock()

	q := url.Values{}
	q.Set("instType", "SWAP")
	q.Set("instId", instId)
	raw, err := o.publicRequest(ctx, "/api/v5/public/instruments", q)
	if err != nil {
		return 0, err
	}
	data := gjson.New(raw).Get("data").Array()
	if len(data) == 0 {
		return 0, gerror.New("OKX instruments empty")
	}
	ctVal := gjson.New(data[0]).Get("ctVal").Float64()
	if ctVal <= 0 {
		return 0, gerror.New("OKX ctVal invalid")
	}

	o.mu.Lock()
	o.instrumentCtV[instId] = ctVal
	o.mu.Unlock()
	return ctVal, nil
}

// GetBalance 获取账户余额（USDT）
func (o *OKX) GetBalance(ctx context.Context) (*Balance, error) {
	q := url.Values{}
	q.Set("ccy", "USDT")
	raw, err := o.signedRequest(ctx, "GET", "/api/v5/account/balance", q, nil)
	if err != nil {
		return nil, err
	}
	j := gjson.New(raw)
	data := j.Get("data").Array()
	if len(data) == 0 {
		return &Balance{Currency: "USDT"}, nil
	}
	item := gjson.New(data[0])
	totalEq := item.Get("totalEq").Float64()
	details := item.Get("details").Array()
	var availEq float64
	for _, d := range details {
		dd := gjson.New(d)
		if dd.Get("ccy").String() == "USDT" {
			availEq = dd.Get("availEq").Float64()
			break
		}
	}
	return &Balance{
		TotalBalance:     totalEq,
		AvailableBalance: availEq,
		UnrealizedPnl:    0,
		Currency:         "USDT",
	}, nil
}

// GetTicker 获取行情
func (o *OKX) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	instId := o.formatInstId(symbol)
	q := url.Values{}
	q.Set("instId", instId)
	raw, err := o.publicRequest(ctx, "/api/v5/market/ticker", q)
	if err != nil {
		return nil, err
	}
	data := gjson.New(raw).Get("data").Array()
	if len(data) == 0 {
		return nil, gerror.New("OKX ticker empty")
	}
	d := gjson.New(data[0])
	return &Ticker{
		Symbol:             symbol,
		LastPrice:          d.Get("last").Float64(),
		BidPrice:           d.Get("bidPx").Float64(),
		AskPrice:           d.Get("askPx").Float64(),
		High24h:            d.Get("high24h").Float64(),
		Low24h:             d.Get("low24h").Float64(),
		Volume24h:          d.Get("vol24h").Float64(),
		Change24h:          0,
		PriceChangePercent: 0,
		Timestamp:          d.Get("ts").Int64(),
	}, nil
}

// GetKlines 获取K线数据
func (o *OKX) GetKlines(ctx context.Context, symbol, interval string, limit int) ([]*Kline, error) {
	instId := o.formatInstId(symbol)
	q := url.Values{}
	q.Set("instId", instId)
	q.Set("bar", o.convertBar(interval))
	q.Set("limit", strconv.Itoa(limit))
	raw, err := o.publicRequest(ctx, "/api/v5/market/candles", q)
	if err != nil {
		return nil, err
	}
	var klines []*Kline
	items := gjson.New(raw).Get("data").Array()
	for _, it := range items {
		arr := gjson.New(it).Array()
		// [ts, o, h, l, c, vol, volCcy, volCcyQuote, confirm]
		if len(arr) >= 6 {
			openTime := g.NewVar(arr[0]).Int64()
			klines = append(klines, &Kline{
				OpenTime:  openTime,
				Open:      g.NewVar(arr[1]).Float64(),
				High:      g.NewVar(arr[2]).Float64(),
				Low:       g.NewVar(arr[3]).Float64(),
				Close:     g.NewVar(arr[4]).Float64(),
				Volume:    g.NewVar(arr[5]).Float64(),
				CloseTime: openTime,
			})
		}
	}
	return klines, nil
}

// GetPositions 获取持仓（统一返回基础币数量）
func (o *OKX) GetPositions(ctx context.Context, symbol string) ([]*Position, error) {
	q := url.Values{}
	q.Set("instType", "SWAP")
	if symbol != "" {
		q.Set("instId", o.formatInstId(symbol))
	}
	raw, err := o.signedRequest(ctx, "GET", "/api/v5/account/positions", q, nil)
	if err != nil {
		return nil, err
	}
	items := gjson.New(raw).Get("data").Array()
	var out []*Position
	for _, it := range items {
		j := gjson.New(it)
		instId := j.Get("instId").String()
		if instId == "" {
			continue
		}
		posContracts := j.Get("pos").Float64()
		if posContracts == 0 {
			continue
		}
		posSide := strings.ToLower(j.Get("posSide").String()) // long/short
		positionSide := "LONG"
		if posSide == "short" {
			positionSide = "SHORT"
		}
		ctVal := j.Get("ctVal").Float64()
		if ctVal <= 0 {
			ctVal, _ = o.getCtVal(ctx, instId)
		}
		if ctVal <= 0 {
			continue
		}
		qtyBase := posContracts * ctVal
		entry := j.Get("avgPx").Float64()
		mark := j.Get("markPx").Float64()
		upl := j.Get("upl").Float64()
		lev := int(j.Get("lever").Float64())
		out = append(out, &Position{
			Symbol:        symbol,
			PositionSide:  positionSide,
			PositionAmt:   qtyBase,
			EntryPrice:    entry,
			MarkPrice:     mark,
			UnrealizedPnl: upl,
			Leverage:      lev,
			MarginType:    "ISOLATED",
		})
	}
	return out, nil
}

func (o *OKX) CreateOrder(ctx context.Context, req *OrderRequest) (*Order, error) {
	instId := o.formatInstId(req.Symbol)
	ctVal, err := o.getCtVal(ctx, instId)
	if err != nil {
		return nil, err
	}
	contracts := int64(req.Quantity / ctVal)
	if contracts <= 0 {
		return nil, gerror.Newf("OKX 下单数量过小: qty=%.8f, ctVal=%.8f，折算合约张数为0", req.Quantity, ctVal)
	}

	posSide := "long"
	if strings.ToUpper(req.PositionSide) == "SHORT" {
		posSide = "short"
	}
	side := "buy"
	if strings.ToUpper(req.Side) == "SELL" {
		side = "sell"
	}
	ordType := "market"
	if strings.ToUpper(req.Type) == "LIMIT" {
		ordType = "limit"
	}

	body := map[string]any{
		"instId":     instId,
		"tdMode":     "isolated",
		"side":       side,
		"posSide":    posSide,
		"ordType":    ordType,
		"sz":         fmt.Sprintf("%d", contracts),
		"reduceOnly": req.ReduceOnly,
	}
	if ordType == "limit" && req.Price > 0 {
		body["px"] = strconv.FormatFloat(req.Price, 'f', -1, 64)
	}

	raw, err := o.signedRequest(ctx, "POST", "/api/v5/trade/order", nil, body)
	if err != nil {
		return nil, err
	}
	data := gjson.New(raw).Get("data").Array()
	if len(data) == 0 {
		return nil, gerror.New("OKX order response empty")
	}
	d := gjson.New(data[0])
	// OKX 下单错误也会返回 200 + code=0，但 data[0].sCode != "0"
	if d.Get("sCode").String() != "" && d.Get("sCode").String() != "0" {
		return nil, gerror.Newf("OKX order failed: sCode=%s sMsg=%s", d.Get("sCode").String(), d.Get("sMsg").String())
	}
	return &Order{
		OrderId:      d.Get("ordId").String(),
		ClientId:     d.Get("clOrdId").String(),
		Symbol:       req.Symbol,
		Side:         strings.ToUpper(req.Side),
		PositionSide: strings.ToUpper(req.PositionSide),
		Type:         strings.ToUpper(req.Type),
		Quantity:     req.Quantity,
		Status:       d.Get("sCode").String(),
		CreateTime:   time.Now().UnixMilli(),
	}, nil
}

func (o *OKX) CancelOrder(ctx context.Context, symbol, orderId string) (*Order, error) {
	instId := o.formatInstId(symbol)
	body := map[string]any{
		"instId": instId,
		"ordId":  orderId,
	}
	_, err := o.signedRequest(ctx, "POST", "/api/v5/trade/cancel-order", nil, body)
	if err != nil {
		return nil, err
	}
	return &Order{OrderId: orderId, Symbol: symbol, Status: "CANCELED"}, nil
}

func (o *OKX) ClosePosition(ctx context.Context, symbol, positionSide string, quantity float64) (*Order, error) {
	// 优先使用 OKX 官方 close-position（更稳：不依赖 ctVal 换算/数量取整）
	if order, err := o.closePositionAll(ctx, symbol, positionSide); err == nil {
		return order, nil
	}

	// fallback：用 reduceOnly 市价单尝试（支持部分平仓）
	side := "SELL"
	if strings.ToUpper(positionSide) == "SHORT" {
		side = "BUY"
	}
	return o.CreateOrder(ctx, &OrderRequest{
		Symbol:       symbol,
		Side:         side,
		PositionSide: positionSide,
		Type:         "MARKET",
		Quantity:     quantity,
		ReduceOnly:   true,
	})
}

// closePositionAll 使用 OKX /api/v5/trade/close-position 一键全平（按方向）
func (o *OKX) closePositionAll(ctx context.Context, symbol, positionSide string) (*Order, error) {
	instId := o.formatInstId(symbol)
	posSide := "long"
	if strings.ToUpper(strings.TrimSpace(positionSide)) == "SHORT" {
		posSide = "short"
	}
	body := map[string]any{
		"instId":  instId,
		"mgnMode": "isolated",
		"posSide": posSide,
		"autoCxl": true, // 自动撤销相关挂单
	}
	raw, err := o.signedRequest(ctx, "POST", "/api/v5/trade/close-position", nil, body)
	if err != nil {
		return nil, err
	}
	data := gjson.New(raw).Get("data").Array()
	if len(data) == 0 {
		return nil, gerror.New("OKX close-position response empty")
	}
	d := gjson.New(data[0])
	if d.Get("sCode").String() != "" && d.Get("sCode").String() != "0" {
		return nil, gerror.Newf("OKX close-position failed: sCode=%s sMsg=%s", d.Get("sCode").String(), d.Get("sMsg").String())
	}
	return &Order{
		OrderId:      d.Get("ordId").String(),
		ClientId:     d.Get("clOrdId").String(),
		Symbol:       symbol,
		PositionSide: strings.ToUpper(positionSide),
		Type:         "MARKET",
		Status:       "CLOSED",
		CreateTime:   time.Now().UnixMilli(),
	}, nil
}

func (o *OKX) SetLeverage(ctx context.Context, symbol string, leverage int) error {
	instId := o.formatInstId(symbol)
	lev := strconv.Itoa(leverage)
	for _, ps := range []string{"long", "short"} {
		body := map[string]any{
			"instId":  instId,
			"mgnMode": "isolated",
			"lever":   lev,
			"posSide": ps,
		}
		_, err := o.signedRequest(ctx, "POST", "/api/v5/account/set-leverage", nil, body)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *OKX) SetMarginType(ctx context.Context, symbol, marginType string) error {
	if strings.ToUpper(marginType) == "ISOLATED" || strings.ToLower(marginType) == "isolated" {
		return nil
	}
	return gerror.New("OKX 仅支持逐仓模式（isolated）")
}

func (o *OKX) GetOpenOrders(ctx context.Context, symbol string) ([]*Order, error) {
	q := url.Values{}
	q.Set("instType", "SWAP")
	if symbol != "" {
		q.Set("instId", o.formatInstId(symbol))
	}
	raw, err := o.signedRequest(ctx, "GET", "/api/v5/trade/orders-pending", q, nil)
	if err != nil {
		return nil, err
	}
	var out []*Order
	for _, it := range gjson.New(raw).Get("data").Array() {
		j := gjson.New(it)
		out = append(out, &Order{
			OrderId:      j.Get("ordId").String(),
			ClientId:     j.Get("clOrdId").String(),
			Symbol:       symbol,
			Side:         strings.ToUpper(j.Get("side").String()),
			PositionSide: strings.ToUpper(j.Get("posSide").String()),
			Type:         strings.ToUpper(j.Get("ordType").String()),
			Price:        j.Get("px").Float64(),
			Status:       j.Get("state").String(),
			CreateTime:   j.Get("cTime").Int64(),
			UpdateTime:   j.Get("uTime").Int64(),
		})
	}
	return out, nil
}

func (o *OKX) GetOrderHistory(ctx context.Context, symbol string, limit int) ([]*Order, error) {
	q := url.Values{}
	q.Set("instType", "SWAP")
	q.Set("limit", strconv.Itoa(limit))
	if symbol != "" {
		q.Set("instId", o.formatInstId(symbol))
	}
	raw, err := o.signedRequest(ctx, "GET", "/api/v5/trade/orders-history", q, nil)
	if err != nil {
		return nil, err
	}
	var out []*Order
	for _, it := range gjson.New(raw).Get("data").Array() {
		j := gjson.New(it)
		out = append(out, &Order{
			OrderId:      j.Get("ordId").String(),
			ClientId:     j.Get("clOrdId").String(),
			Symbol:       symbol,
			Side:         strings.ToUpper(j.Get("side").String()),
			PositionSide: strings.ToUpper(j.Get("posSide").String()),
			Type:         strings.ToUpper(j.Get("ordType").String()),
			Price:        j.Get("px").Float64(),
			Status:       j.Get("state").String(),
			CreateTime:   j.Get("cTime").Int64(),
			UpdateTime:   j.Get("uTime").Int64(),
		})
	}
	return out, nil
}

// GetTradeHistory 获取成交记录（用于财务对账/已实现盈亏/手续费汇总）
// OKX V5：GET /api/v5/trade/fills
func (o *OKX) GetTradeHistory(ctx context.Context, symbol string, limit int) ([]*Trade, error) {
	q := url.Values{}
	q.Set("instType", "SWAP")
	if symbol != "" {
		q.Set("instId", o.formatInstId(symbol))
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	raw, err := o.signedRequest(ctx, "GET", "/api/v5/trade/fills", q, nil)
	if err != nil {
		return nil, err
	}

	// OKX fills 的 fillSz 是“合约张数”，需要折算成基础币数量（与其它交易所 quantity 口径一致）
	ctVal := float64(0)
	if symbol != "" {
		ctVal, _ = o.getCtVal(ctx, o.formatInstId(symbol))
	}

	var out []*Trade
	for _, it := range gjson.New(raw).Get("data").Array() {
		j := gjson.New(it)
		tradeId := j.Get("fillId").String()
		if tradeId == "" {
			tradeId = j.Get("billId").String()
		}
		fee := j.Get("fee").Float64()
		fillSz := math.Abs(j.Get("fillSz").Float64())
		qty := fillSz
		if ctVal > 0 {
			qty = fillSz * ctVal
		}
		out = append(out, &Trade{
			TradeId:         tradeId,
			OrderId:         j.Get("ordId").String(),
			Symbol:          symbol,
			Side:            strings.ToUpper(j.Get("side").String()),
			PositionSide:    strings.ToUpper(j.Get("posSide").String()), // long/short -> LONG/SHORT
			Price:           j.Get("fillPx").Float64(),
			Quantity:        qty,
			RealizedPnl:     j.Get("pnl").Float64(),
			Commission:      math.Abs(fee),
			CommissionAsset: j.Get("feeCcy").String(),
			Time:            j.Get("ts").Int64(),
		})
	}
	return out, nil
}
