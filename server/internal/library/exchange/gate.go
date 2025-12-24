// Package exchange Gate.io 交易所API（U本位永续 / 逐仓 / 双向持仓）
package exchange

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
)

// Gate Gate.io（API v4, futures/usdt）
// 约束：仅实现 USDT 永续（futures/usdt） + 逐仓（isolated） + 双向持仓（long/short）
type Gate struct {
	config   *Config
	endpoint string

	mu sync.Mutex
	// contract -> quanto_multiplier（合约面值，单位：基础币），用于将 size(张) 转为基础币数量
	contractMultiplier map[string]float64
	dualModeEnsured    bool
}

func NewGate(config *Config) *Gate {
	return &Gate{
		config:             config,
		endpoint:           "https://api.gateio.ws",
		contractMultiplier: make(map[string]float64),
	}
}

func (gt *Gate) GetName() string { return "gate" }

func (gt *Gate) getHttpClient() *gclient.Client {
	client := gclient.New()
	client.SetTimeout(20 * time.Second)
	if gt.config.Proxy != nil && gt.config.Proxy.Enabled {
		client.SetProxy(gt.config.Proxy.GetProxyURL())
	}
	return client
}

func (gt *Gate) formatContract(symbol string) string {
	// 使用统一的Symbol格式化器
	return Formatter.FormatForGate(symbol) // BTC_USDT
}

func (gt *Gate) convertInterval(interval string) string {
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
		return "1h"
	case "2h":
		return "2h"
	case "4h":
		return "4h"
	case "6h":
		return "6h"
	case "12h":
		return "12h"
	case "1d":
		return "1d"
	default:
		return "1m"
	}
}

// Gate v4 签名（常见格式）：
// SIGN = HMAC_SHA512(secret, stringToSign)
// stringToSign = method+"\n"+requestPath+"\n"+queryString+"\n"+hashedPayload+"\n"+timestamp
// hashedPayload = SHA512(body)
func (gt *Gate) sign(method, requestPath, queryString, body, timestamp string) string {
	h := sha512.Sum512([]byte(body))
	hashedPayload := hex.EncodeToString(h[:])
	stringToSign := strings.ToUpper(method) + "\n" + requestPath + "\n" + queryString + "\n" + hashedPayload + "\n" + timestamp
	mac := hmac.New(sha512.New, []byte(gt.config.SecretKey))
	mac.Write([]byte(stringToSign))
	return hex.EncodeToString(mac.Sum(nil))
}

func (gt *Gate) signedRequest(ctx context.Context, method, path string, query url.Values, body any) (string, error) {
	requestPath := "/api/v4" + path
	queryString := ""
	if query != nil && len(query) > 0 {
		// 【重要】签名必须使用稳定顺序的 queryString，并且与实际请求 URL 完全一致
		queryString = buildQueryWithStableOrder(query)
	}

	bodyStr := ""
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return "", err
		}
		bodyStr = string(b)
	}

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sign := gt.sign(method, requestPath, queryString, bodyStr, ts)

	client := gt.getHttpClient()
	client.SetHeader("KEY", gt.config.ApiKey)
	client.SetHeader("Timestamp", ts)
	client.SetHeader("SIGN", sign)
	client.SetHeader("Content-Type", "application/json")

	reqURL := gt.endpoint + requestPath
	if queryString != "" {
		reqURL += "?" + queryString
	}

	var resp *gclient.Response
	var err error
	switch strings.ToUpper(method) {
	case "GET":
		resp, err = client.Get(ctx, reqURL)
	case "POST":
		resp, err = client.Post(ctx, reqURL, bodyStr)
	case "DELETE":
		resp, err = client.Delete(ctx, reqURL)
	default:
		resp, err = client.Get(ctx, reqURL)
	}
	if err != nil {
		return "", gerror.Wrap(err, "Gate request failed")
	}
	defer resp.Close()

	raw := resp.ReadAllString()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", gerror.Wrapf(WrapAsAPIError("gate", resp.StatusCode, raw, nil), "[gate] http status=%d path=%s", resp.StatusCode, requestPath)
	}
	return raw, nil
}

func (gt *Gate) publicRequest(ctx context.Context, path string, query url.Values) (string, error) {
	requestPath := "/api/v4" + path
	reqURL := gt.endpoint + requestPath
	if query != nil && len(query) > 0 {
		reqURL += "?" + query.Encode()
	}
	client := gt.getHttpClient()
	resp, err := client.Get(ctx, reqURL)
	if err != nil {
		return "", gerror.Wrap(err, "Gate request failed")
	}
	defer resp.Close()
	raw := resp.ReadAllString()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", gerror.Wrapf(WrapAsAPIError("gate", resp.StatusCode, raw, nil), "[gate] http status=%d path=%s", resp.StatusCode, requestPath)
	}
	return raw, nil
}

// ensureDualMode 尽量确保双向持仓（对冲）模式
// Gate 的双向模式在不同账号配置下可能默认开启；这里做一次尝试，不阻断交易。
func (gt *Gate) ensureDualMode(ctx context.Context) {
	gt.mu.Lock()
	if gt.dualModeEnsured {
		gt.mu.Unlock()
		return
	}
	gt.mu.Unlock()

	// 尝试调用双向模式设置接口（如果接口不存在/权限不足会失败，但不阻断）
	_, _ = gt.signedRequest(ctx, "POST", "/futures/usdt/dual_mode", nil, map[string]any{
		"dual_mode": true,
	})

	gt.mu.Lock()
	gt.dualModeEnsured = true
	gt.mu.Unlock()
}

func (gt *Gate) getMultiplier(ctx context.Context, contract string) (float64, error) {
	gt.mu.Lock()
	if v, ok := gt.contractMultiplier[contract]; ok && v > 0 {
		gt.mu.Unlock()
		return v, nil
	}
	gt.mu.Unlock()

	raw, err := gt.publicRequest(ctx, "/futures/usdt/contracts/"+url.PathEscape(contract), nil)
	if err != nil {
		return 0, err
	}
	j := gjson.New(raw)
	// 常见字段：quanto_multiplier
	m := j.Get("quanto_multiplier").Float64()
	if m <= 0 {
		// 兜底：部分字段命名可能为 "contract_size"
		m = j.Get("contract_size").Float64()
	}
	if m <= 0 {
		return 0, gerror.New("Gate contract multiplier invalid")
	}

	gt.mu.Lock()
	gt.contractMultiplier[contract] = m
	gt.mu.Unlock()
	return m, nil
}

// GetBalance 获取账户余额（USDT）
func (gt *Gate) GetBalance(ctx context.Context) (*Balance, error) {
	raw, err := gt.signedRequest(ctx, "GET", "/futures/usdt/accounts", nil, nil)
	if err != nil {
		return nil, err
	}
	j := gjson.New(raw)
	// 常见字段：total / available / unrealised_pnl
	total := j.Get("total").Float64()
	avail := j.Get("available").Float64()
	upl := j.Get("unrealised_pnl").Float64()
	return &Balance{
		TotalBalance:     total,
		AvailableBalance: avail,
		UnrealizedPnl:    upl,
		Currency:         "USDT",
	}, nil
}

// GetTicker 获取行情
func (gt *Gate) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	contract := gt.formatContract(symbol)
	q := url.Values{}
	q.Set("contract", contract)
	raw, err := gt.publicRequest(ctx, "/futures/usdt/tickers", q)
	if err != nil {
		return nil, err
	}
	// tickers 返回数组
	items := gjson.New(raw).Array()
	if len(items) == 0 {
		return nil, gerror.New("Gate ticker empty")
	}
	j := gjson.New(items[0])
	last := j.Get("last").Float64()
	bid := j.Get("highest_bid").Float64()
	ask := j.Get("lowest_ask").Float64()
	high := j.Get("high_24h").Float64()
	low := j.Get("low_24h").Float64()
	vol := j.Get("volume_24h").Float64()
	return &Ticker{
		Symbol:             symbol,
		LastPrice:          last,
		BidPrice:           bid,
		AskPrice:           ask,
		High24h:            high,
		Low24h:             low,
		Volume24h:          vol,
		Change24h:          0,
		PriceChangePercent: 0,
		Timestamp:          time.Now().UnixMilli(),
	}, nil
}

// GetKlines 获取K线
func (gt *Gate) GetKlines(ctx context.Context, symbol, interval string, limit int) ([]*Kline, error) {
	contract := gt.formatContract(symbol)
	q := url.Values{}
	q.Set("contract", contract)
	q.Set("interval", gt.convertInterval(interval))
	q.Set("limit", strconv.Itoa(limit))
	raw, err := gt.publicRequest(ctx, "/futures/usdt/candlesticks", q)
	if err != nil {
		return nil, err
	}
	var out []*Kline
	// 返回二维数组： [t, v, c, h, l, o] 或 [t, o, h, l, c, v]（不同文档版本）
	for _, it := range gjson.New(raw).Array() {
		arr := gjson.New(it).Array()
		if len(arr) < 6 {
			continue
		}
		// 尝试两种格式
		t0 := g.NewVar(arr[0]).Int64()
		// Gate 通常为秒，转换为毫秒
		openTime := t0 * 1000

		// 格式A: [t, v, c, h, l, o]
		vA := g.NewVar(arr[1]).Float64()
		cA := g.NewVar(arr[2]).Float64()
		hA := g.NewVar(arr[3]).Float64()
		lA := g.NewVar(arr[4]).Float64()
		oA := g.NewVar(arr[5]).Float64()

		// 如果 close 为 0 而 open 非 0，可能是格式B
		open, high, low, close, vol := oA, hA, lA, cA, vA
		if close == 0 && g.NewVar(arr[1]).Float64() > 0 && g.NewVar(arr[5]).Float64() == 0 {
			// 格式B: [t, o, h, l, c, v]
			open = g.NewVar(arr[1]).Float64()
			high = g.NewVar(arr[2]).Float64()
			low = g.NewVar(arr[3]).Float64()
			close = g.NewVar(arr[4]).Float64()
			vol = g.NewVar(arr[5]).Float64()
		}

		out = append(out, &Kline{
			OpenTime:  openTime,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    vol,
			CloseTime: openTime,
		})
	}
	return out, nil
}

// GetPositions 获取持仓（统一返回基础币数量）
func (gt *Gate) GetPositions(ctx context.Context, symbol string) ([]*Position, error) {
	gt.ensureDualMode(ctx)
	contract := gt.formatContract(symbol)

	// 如果指定 contract，优先查单个持仓；否则查列表
	var raw string
	var err error
	if symbol != "" {
		raw, err = gt.signedRequest(ctx, "GET", "/futures/usdt/positions/"+url.PathEscape(contract), nil, nil)
		if err != nil {
			// 如果单个接口不可用，退回列表接口
			raw = ""
		}
	}
	if raw == "" {
		q := url.Values{}
		if symbol != "" {
			q.Set("contract", contract)
		}
		raw, err = gt.signedRequest(ctx, "GET", "/futures/usdt/positions", q, nil)
		if err != nil {
			return nil, err
		}
	}

	// 统一成数组处理（gf gjson 没有 IsArray/Value 之类方法，这里用文本前缀判断）
	var items []any
	trim := strings.TrimSpace(raw)
	if strings.HasPrefix(trim, "[") {
		items = gjson.New(raw).Array()
	} else {
		// 单对象响应：转为 map 再封装成 slice
		items = []any{gjson.New(raw).Map()}
	}

	mul, _ := gt.getMultiplier(ctx, contract)
	if mul <= 0 {
		// 不阻断，使用 1 兜底（会影响 qty 折算，但后续下单会根据最小量校验）
		mul = 1
	}

	var out []*Position
	for _, it := range items {
		j := gjson.New(it)
		size := j.Get("size").Float64() // 合约张数，正=多，负=空（常见约定）
		if size == 0 {
			continue
		}
		positionSide := "LONG"
		if size < 0 {
			positionSide = "SHORT"
		}
		qtyBase := absFloat(size) * mul
		entry := j.Get("entry_price").Float64()
		mark := j.Get("mark_price").Float64()
		upl := j.Get("unrealised_pnl").Float64()
		lev := int(j.Get("leverage").Float64())
		margin := j.Get("margin").Float64()
		out = append(out, &Position{
			Symbol:        symbol,
			PositionSide:  positionSide,
			PositionAmt:   qtyBase,
			EntryPrice:    entry,
			MarkPrice:     mark,
			UnrealizedPnl: upl,
			Leverage:      lev,
			Margin:        margin,
			MarginType:    "ISOLATED",
		})
	}
	return out, nil
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// CreateOrder 创建订单
func (gt *Gate) CreateOrder(ctx context.Context, req *OrderRequest) (*Order, error) {
	gt.ensureDualMode(ctx)

	contract := gt.formatContract(req.Symbol)
	mul, err := gt.getMultiplier(ctx, contract)
	if err != nil {
		return nil, err
	}
	if mul <= 0 {
		return nil, gerror.New("Gate contract multiplier invalid")
	}

	// Gate futures 下单 size 为合约张数，正=买，负=卖
	contracts := int64(req.Quantity / mul)
	if contracts <= 0 {
		return nil, gerror.Newf("Gate 下单数量过小: qty=%.8f, multiplier=%.8f，折算合约张数为0", req.Quantity, mul)
	}

	size := contracts
	if strings.ToUpper(req.Side) == "SELL" {
		size = -contracts
	}

	orderType := "market"
	if strings.ToUpper(req.Type) == "LIMIT" {
		orderType = "limit"
	}

	// 逐仓 + 双向：使用 reduce_only 保护平仓
	body := map[string]any{
		"contract":    contract,
		"size":        size,
		"price":       "0",   // 市价单：价格=0
		"tif":         "ioc", // 立即成交
		"reduce_only": req.ReduceOnly,
	}
	if orderType == "limit" && req.Price > 0 {
		body["price"] = strconv.FormatFloat(req.Price, 'f', -1, 64)
		body["tif"] = "gtc"
	}

	raw, err := gt.signedRequest(ctx, "POST", "/futures/usdt/orders", nil, body)
	if err != nil {
		return nil, err
	}
	j := gjson.New(raw)
	return &Order{
		OrderId:      j.Get("id").String(),
		ClientId:     j.Get("text").String(),
		Symbol:       req.Symbol,
		Side:         strings.ToUpper(req.Side),
		PositionSide: strings.ToUpper(req.PositionSide),
		Type:         strings.ToUpper(req.Type),
		Price:        j.Get("price").Float64(),
		Quantity:     req.Quantity,
		FilledQty:    j.Get("size").Float64(), // 合约张数（近似）
		AvgPrice:     j.Get("fill_price").Float64(),
		Status:       j.Get("status").String(),
		CreateTime:   time.Now().UnixMilli(),
		UpdateTime:   time.Now().UnixMilli(),
	}, nil
}

// CancelOrder 取消订单
func (gt *Gate) CancelOrder(ctx context.Context, symbol, orderId string) (*Order, error) {
	contract := gt.formatContract(symbol)
	q := url.Values{}
	q.Set("contract", contract)
	_, err := gt.signedRequest(ctx, "DELETE", "/futures/usdt/orders/"+url.PathEscape(orderId), q, nil)
	if err != nil {
		return nil, err
	}
	return &Order{OrderId: orderId, Symbol: symbol, Status: "CANCELED"}, nil
}

// ClosePosition 平仓（reduce_only 市价）
func (gt *Gate) ClosePosition(ctx context.Context, symbol, positionSide string, quantity float64) (*Order, error) {
	side := "SELL"
	if strings.ToUpper(positionSide) == "SHORT" {
		side = "BUY"
	}
	return gt.CreateOrder(ctx, &OrderRequest{
		Symbol:       symbol,
		Side:         side,
		PositionSide: positionSide,
		Type:         "MARKET",
		Quantity:     quantity,
		ReduceOnly:   true,
	})
}

func (gt *Gate) SetLeverage(ctx context.Context, symbol string, leverage int) error {
	contract := gt.formatContract(symbol)
	body := map[string]any{
		"leverage": strconv.Itoa(leverage),
	}
	// 兼容不同实现：有的接口是 /positions/{contract}/leverage
	_, err := gt.signedRequest(ctx, "POST", "/futures/usdt/positions/"+url.PathEscape(contract)+"/leverage", nil, body)
	return err
}

func (gt *Gate) SetMarginType(ctx context.Context, symbol, marginType string) error {
	// 本系统只支持逐仓
	if strings.ToUpper(marginType) == "ISOLATED" || strings.ToLower(marginType) == "isolated" {
		return nil
	}
	return gerror.New("Gate 仅支持逐仓模式（isolated）")
}

func (gt *Gate) GetOpenOrders(ctx context.Context, symbol string) ([]*Order, error) {
	contract := gt.formatContract(symbol)
	q := url.Values{}
	q.Set("contract", contract)
	q.Set("status", "open")
	raw, err := gt.signedRequest(ctx, "GET", "/futures/usdt/orders", q, nil)
	if err != nil {
		return nil, err
	}
	var out []*Order
	for _, it := range gjson.New(raw).Array() {
		j := gjson.New(it)
		side := "BUY"
		if j.Get("size").Int64() < 0 {
			side = "SELL"
		}
		typ := "LIMIT"
		if j.Get("price").String() == "0" {
			typ = "MARKET"
		}
		out = append(out, &Order{
			OrderId:    j.Get("id").String(),
			ClientId:   j.Get("text").String(),
			Symbol:     symbol,
			Side:       side,
			Type:       typ,
			Price:      j.Get("price").Float64(),
			Status:     j.Get("status").String(),
			CreateTime: j.Get("create_time").Int64() * 1000,
			UpdateTime: j.Get("finish_time").Int64() * 1000,
		})
	}
	return out, nil
}

func (gt *Gate) GetOrderHistory(ctx context.Context, symbol string, limit int) ([]*Order, error) {
	contract := gt.formatContract(symbol)
	q := url.Values{}
	q.Set("contract", contract)
	q.Set("status", "finished")
	q.Set("limit", strconv.Itoa(limit))
	raw, err := gt.signedRequest(ctx, "GET", "/futures/usdt/orders", q, nil)
	if err != nil {
		return nil, err
	}
	var out []*Order
	for _, it := range gjson.New(raw).Array() {
		j := gjson.New(it)
		side := "BUY"
		if j.Get("size").Int64() < 0 {
			side = "SELL"
		}
		typ := "LIMIT"
		if j.Get("price").String() == "0" {
			typ = "MARKET"
		}
		out = append(out, &Order{
			OrderId:    j.Get("id").String(),
			ClientId:   j.Get("text").String(),
			Symbol:     symbol,
			Side:       side,
			Type:       typ,
			Price:      j.Get("price").Float64(),
			Status:     j.Get("status").String(),
			CreateTime: j.Get("create_time").Int64() * 1000,
			UpdateTime: j.Get("finish_time").Int64() * 1000,
		})
	}
	return out, nil
}

// GetTradeHistory 获取成交记录（用于财务对账/已实现盈亏/手续费汇总）
// Gate v4 futures: GET /futures/usdt/my_trades
func (gt *Gate) GetTradeHistory(ctx context.Context, symbol string, limit int) ([]*Trade, error) {
	contract := gt.formatContract(symbol)
	q := url.Values{}
	q.Set("contract", contract)
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	raw, err := gt.signedRequest(ctx, "GET", "/futures/usdt/my_trades", q, nil)
	if err != nil {
		return nil, err
	}

	mul, _ := gt.getMultiplier(ctx, contract)
	if mul <= 0 {
		mul = 1
	}

	var out []*Trade
	for _, it := range gjson.New(raw).Array() {
		j := gjson.New(it)
		size := j.Get("size").Float64() // 合约张数，正/负表示方向（常见约定）
		side := strings.ToUpper(j.Get("side").String())
		if side == "" {
			if size < 0 {
				side = "SELL"
			} else {
				side = "BUY"
			}
		}
		posSide := strings.ToUpper(j.Get("pos_side").String())
		if posSide == "" {
			posSide = strings.ToUpper(j.Get("position_side").String())
		}
		// 没有明确 posSide 时留空（上层过滤时会容错）
		orderID := j.Get("order_id").String()
		if orderID == "" {
			orderID = j.Get("orderId").String()
		}

		fee := j.Get("fee").Float64()
		feeCcy := j.Get("fee_currency").String()
		if feeCcy == "" {
			feeCcy = j.Get("feeCcy").String()
		}
		if feeCcy == "" {
			feeCcy = "USDT"
		}

		ts := j.Get("create_time").Int64()
		if ts > 0 && ts < 1e12 {
			ts = ts * 1000
		}

		out = append(out, &Trade{
			TradeId:         j.Get("id").String(),
			OrderId:         orderID,
			Symbol:          symbol,
			Side:            side,
			PositionSide:    posSide,
			Price:           j.Get("price").Float64(),
			Quantity:        absFloat(size) * mul, // 折算成基础币数量
			RealizedPnl:     j.Get("pnl").Float64(),
			Commission:      absFloat(fee),
			CommissionAsset: feeCcy,
			Time:            ts,
		})
	}
	return out, nil
}

// buildQueryWithStableOrder：Gate 对 query 的排序较敏感，这里显式排序
func buildQueryWithStableOrder(values url.Values) string {
	if values == nil {
		return ""
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		for _, v := range values[k] {
			parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
		}
	}
	return strings.Join(parts, "&")
}
