// Package exchange OKX交易所API（U本位永续 / 逐仓 / 双向持仓）
package exchange

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
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
	instrumentMin map[string]okxInstrumentInfo
}

type okxInstrumentInfo struct {
	CtVal float64 // 合约面值（基础币）
	MinSz float64 // 最小下单张数（合约张数）
	LotSz float64 // 张数步进（合约张数）
}

func NewOKX(config *Config) *OKX {
	// OKX 模拟盘仍使用同域名，通过账号权限区分
	return &OKX{
		config:        config,
		endpoint:      "https://www.okx.com",
		instrumentCtV: make(map[string]float64),
		instrumentMin: make(map[string]okxInstrumentInfo),
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
	if len(query) > 0 {
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

	// 【关键修复】OKX 对 timestamp 很敏感：本机时间漂移会导致 60006/401。
	// 这里引入 serverTime offset（REST 与 私有WS 共用），并在检测到 timestamp expired 时自动校准重试一次。
	if lastTimeSyncMs(o.config) == 0 || time.Since(time.UnixMilli(lastTimeSyncMs(o.config))) > 10*time.Minute {
		_, _ = SyncServerTimeOffset(ctx, o.config)
	}

	maxRetries := 1
	for retry := 0; retry <= maxRetries; retry++ {
		// OKX 对 OK-ACCESS-TIMESTAMP 的格式较严格，使用毫秒级 RFC3339 更兼容：
		// e.g. 2025-12-24T12:34:56.789Z
		tsMs := nowMsWithOffset(o.config)
		ts := time.UnixMilli(tsMs).UTC().Format("2006-01-02T15:04:05.000Z")
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
			// 网络错误不重试时间同步
			return "", gerror.Wrap(err, "OKX request failed")
		}

		raw := resp.ReadAllString()
		status := resp.StatusCode
		resp.Close()

		// http status != 200：尝试识别 timestamp expired（常见 401）
		if status != 200 {
			if retry < maxRetries && IsTimestampExpiredError(nil, raw) {
				_, _ = SyncServerTimeOffset(ctx, o.config)
				continue
			}
			return "", gerror.Wrapf(WrapAsAPIError("okx", status, raw, nil), "[okx] http status=%d path=%s", status, requestPath)
		}

		j := gjson.New(raw)
		if j.Get("code").String() != "0" {
			if retry < maxRetries && IsTimestampExpiredError(nil, raw) {
				_, _ = SyncServerTimeOffset(ctx, o.config)
				continue
			}
			// 提供更完整上下文，避免只有 “code=1 msg=All operations failed” 难以定位
			rawShort := raw
			if len(rawShort) > 600 {
				rawShort = rawShort[:600] + "...(truncated)"
			}
			bodyShort := bodyStr
			if len(bodyShort) > 400 {
				bodyShort = bodyShort[:400] + "...(truncated)"
			}
			if bodyShort != "" && strings.ToUpper(method) != "GET" {
				return "", gerror.Newf("OKX API error: method=%s path=%s code=%s msg=%s body=%s raw=%s",
					strings.ToUpper(method), requestPath, j.Get("code").String(), j.Get("msg").String(), bodyShort, rawShort)
			}
			return "", gerror.Newf("OKX API error: method=%s path=%s code=%s msg=%s raw=%s",
				strings.ToUpper(method), requestPath, j.Get("code").String(), j.Get("msg").String(), rawShort)
		}
		return raw, nil
	}

	return "", gerror.New("OKX request failed after retries")
}

func (o *OKX) publicRequest(ctx context.Context, path string, query url.Values) (string, error) {
	reqURL := o.endpoint + path
	if len(query) > 0 {
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
		rawShort := raw
		if len(rawShort) > 600 {
			rawShort = rawShort[:600] + "...(truncated)"
		}
		return "", gerror.Newf("OKX API error: method=GET path=%s code=%s msg=%s raw=%s",
			path, j.Get("code").String(), j.Get("msg").String(), rawShort)
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

func (o *OKX) getInstrumentInfo(ctx context.Context, instId string) (okxInstrumentInfo, error) {
	o.mu.Lock()
	if v, ok := o.instrumentMin[instId]; ok && v.CtVal > 0 {
		o.mu.Unlock()
		return v, nil
	}
	o.mu.Unlock()

	q := url.Values{}
	q.Set("instType", "SWAP")
	q.Set("instId", instId)
	raw, err := o.publicRequest(ctx, "/api/v5/public/instruments", q)
	if err != nil {
		return okxInstrumentInfo{}, err
	}
	data := gjson.New(raw).Get("data").Array()
	if len(data) == 0 {
		return okxInstrumentInfo{}, gerror.New("OKX instruments empty")
	}
	j := gjson.New(data[0])
	ctVal := j.Get("ctVal").Float64()
	if ctVal <= 0 {
		return okxInstrumentInfo{}, gerror.New("OKX ctVal invalid")
	}
	minSz := j.Get("minSz").Float64()
	lotSz := j.Get("lotSz").Float64()
	// 兜底：没有返回时，按 1 张处理（OKX 大多为整张）
	if minSz <= 0 {
		minSz = 1
	}
	if lotSz <= 0 {
		lotSz = 1
	}
	info := okxInstrumentInfo{CtVal: ctVal, MinSz: minSz, LotSz: lotSz}

	o.mu.Lock()
	o.instrumentCtV[instId] = ctVal
	o.instrumentMin[instId] = info
	o.mu.Unlock()
	return info, nil
}

// AdjustBaseQtyToMinContracts 将“基础币数量”换算为合约张数，并按 OKX 的 minSz/lotSz 向上取整。
// 返回：
// - adjQtyBase：调整后的基础币数量（= contracts * ctVal）
// - contracts：最终下单张数
func (o *OKX) AdjustBaseQtyToMinContracts(ctx context.Context, symbol string, qtyBase float64) (adjQtyBase float64, contracts float64, ctVal float64, minSz float64, lotSz float64, err error) {
	instId := o.formatInstId(symbol)
	info, err := o.getInstrumentInfo(ctx, instId)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	ctVal = info.CtVal
	minSz = info.MinSz
	lotSz = info.LotSz
	if ctVal <= 0 {
		return 0, 0, 0, minSz, lotSz, gerror.New("OKX ctVal invalid")
	}
	if qtyBase <= 0 {
		qtyBase = 0
	}
	c := qtyBase / ctVal
	if c < minSz {
		c = minSz
	}
	if lotSz > 0 {
		steps := math.Ceil(c / lotSz)
		c = steps * lotSz
	}
	if c <= 0 {
		return 0, 0, ctVal, minSz, lotSz, gerror.New("OKX contracts invalid")
	}
	return c * ctVal, c, ctVal, minSz, lotSz, nil
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

	last := d.Get("last").Float64()
	open24h := d.Get("open24h").Float64()
	// 兼容：部分接口/场景可能没有 open24h，尝试用当日UTC 0点开盘价兜底（不严格等同于24h，但比0好）
	if open24h <= 0 {
		open24h = d.Get("sodUtc0").Float64()
	}
	changePercent := 0.0
	if open24h > 0 && last > 0 {
		changePercent = (last - open24h) / open24h * 100.0
	}

	return &Ticker{
		Symbol:    symbol,
		LastPrice: last,
		BidPrice:  d.Get("bidPx").Float64(),
		AskPrice:  d.Get("askPx").Float64(),
		High24h:   d.Get("high24h").Float64(),
		Low24h:    d.Get("low24h").Float64(),
		Volume24h: d.Get("vol24h").Float64(),
		// 统一口径：这里直接返回“百分比”数值（例如 +3.25 表示 +3.25%）
		Change24h:          changePercent,
		PriceChangePercent: changePercent,
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
		// OKX positions 接口未直接返回统一口径的保证金字段，这里按“逐仓价值/杠杆”兜底计算，
		// 保证后端止损/止盈逻辑与前端血条口径一致（保证金=持仓价值/杠杆）。
		var margin float64
		if entry > 0 && lev > 0 && qtyBase > 0 {
			margin = (qtyBase * entry) / float64(lev)
		}
		out = append(out, &Position{
			Symbol:         symbol,
			PositionSide:   positionSide,
			PositionAmt:    qtyBase,
			EntryPrice:     entry,
			MarkPrice:      mark,
			UnrealizedPnl:  upl,
			Leverage:       lev,
			Margin:         margin,
			IsolatedMargin: margin,
			MarginType:     "ISOLATED",
		})
	}
	return out, nil
}

func (o *OKX) CreateOrder(ctx context.Context, req *OrderRequest) (*Order, error) {
	instId := o.formatInstId(req.Symbol)
	info, err := o.getInstrumentInfo(ctx, instId)
	if err != nil {
		return nil, err
	}
	// OKX 下单 sz 单位是“合约张数”，这里把基础币数量换算后按 minSz/lotSz 向上取整
	c := req.Quantity / info.CtVal
	if c < info.MinSz {
		c = info.MinSz
	}
	if info.LotSz > 0 {
		c = math.Ceil(c/info.LotSz) * info.LotSz
	}
	if c <= 0 {
		return nil, gerror.Newf("OKX 下单数量无效: qty=%.8f, ctVal=%.8f", req.Quantity, info.CtVal)
	}
	contracts := strconv.FormatFloat(c, 'f', -1, 64)

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
		"sz":         contracts,
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
		// Quantity 统一用基础币数量（与 GetPositions/Trade 解析一致）
		Quantity:   c * info.CtVal,
		Status:     d.Get("sCode").String(),
		CreateTime: time.Now().UnixMilli(),
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
	// OKX /api/v5/trade/fills:
	// - 单次 limit 有上限（官方为 100），之前传 1000 会触发参数错误 → “获取成交历史失败”
	// - 这里按需分页拉取，直到拿到期望条数或无更多数据
	want := limit
	if want <= 0 {
		want = 100
	}
	perPage := want
	if perPage > 100 {
		perPage = 100
	}
	maxPages := (want + perPage - 1) / perPage
	if maxPages < 1 {
		maxPages = 1
	}
	if maxPages > 20 {
		maxPages = 20 // 安全上限，避免异常分页导致无限循环
	}

	instId := ""
	if symbol != "" {
		instId = o.formatInstId(symbol)
	}

	// OKX fills 的 fillSz 是“合约张数”，需要折算成基础币数量（与其它交易所 quantity 口径一致）
	ctVal := float64(0)
	if instId != "" {
		ctVal, _ = o.getCtVal(ctx, instId)
	}

	out := make([]*Trade, 0, want)
	after := "" // pagination cursor (billId)
	for page := 0; page < maxPages && len(out) < want; page++ {
		q := url.Values{}
		q.Set("instType", "SWAP")
		if instId != "" {
			q.Set("instId", instId)
		}
		q.Set("limit", strconv.Itoa(perPage))
		if after != "" {
			q.Set("after", after)
		}

		raw, err := o.signedRequest(ctx, "GET", "/api/v5/trade/fills", q, nil)
		if err != nil {
			return nil, err
		}
		arr := gjson.New(raw).Get("data").Array()
		if len(arr) == 0 {
			break
		}

		nextAfter := after
		for i, it := range arr {
			j := gjson.New(it)
			tradeId := j.Get("fillId").String()
			billId := j.Get("billId").String()
			if tradeId == "" {
				tradeId = billId
			}
			fee := j.Get("fee").Float64()
			// OKX fills 的实现盈亏字段在不同接口/版本里可能为：pnl / fillPnl / realizedPnl
			rp := j.Get("pnl").Float64()
			if rp == 0 {
				rp = j.Get("fillPnl").Float64()
			}
			if rp == 0 {
				rp = j.Get("realizedPnl").Float64()
			}
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
				PositionSide:    strings.ToUpper(j.Get("posSide").String()),
				Price:           j.Get("fillPx").Float64(),
				Quantity:        qty,
				RealizedPnl:     rp,
				Commission:      math.Abs(fee),
				CommissionAsset: j.Get("feeCcy").String(),
				Time:            j.Get("ts").Int64(),
			})
			if len(out) >= want {
				break
			}
			// cursor: use last item's billId if present
			if i == len(arr)-1 && billId != "" {
				nextAfter = billId
			}
		}

		// no cursor progress → stop to avoid infinite loop
		if nextAfter == "" || nextAfter == after {
			break
		}
		after = nextAfter
	}

	// Fallback: if exchange does not provide pnl on fills, compute realized pnl by avg-cost model (Bitget-like).
	FillRealizedPnlByAvgCost(out)

	return out, nil
}
