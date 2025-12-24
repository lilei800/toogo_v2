// Package exchange Bitget交易所API
package exchange

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"math"
	"net"
	"net/http"
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
	"golang.org/x/net/proxy"
)

// Bitget Bitget交易所
type Bitget struct {
	config      *Config
	endpoint    string
	client      *gclient.Client // 复用HTTP客户端
	lastRequest time.Time       // 上次请求时间（用于限频）
	reqMu       sync.Mutex      // 请求互斥锁
}

// NewBitget 创建Bitget实例
// 【优化】增加连接池大小和 TLS 握手超时，减少 TLS handshake timeout 错误
func NewBitget(config *Config) *Bitget {
	endpoint := "https://api.bitget.com"
	if config.IsTestnet {
		endpoint = "https://api.bitget.com" // Bitget测试网同域名
	}

	// 【优化】配置更大的连接池和更长的超时时间
	transport := &http.Transport{
		MaxIdleConns:        100,              // 最大空闲连接数
		MaxIdleConnsPerHost: 20,               // 每个主机最大空闲连接数
		MaxConnsPerHost:     50,               // 每个主机最大连接数
		IdleConnTimeout:     90 * time.Second, // 空闲连接超时
		TLSHandshakeTimeout: 30 * time.Second, // TLS 握手超时（增加到30秒）
		DisableKeepAlives:   false,            // 启用 Keep-Alive
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second, // 连接超时
			KeepAlive: 30 * time.Second, // Keep-Alive 间隔
		}).DialContext,
	}

	// 配置代理
	if config.Proxy != nil && config.Proxy.Enabled {
		proxyAddr := config.Proxy.Host + ":" + strconv.Itoa(config.Proxy.Port)
		if config.Proxy.Type == "socks5" {
			// SOCKS5 代理
			dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
			if err == nil {
				transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
					return dialer.Dial(network, addr)
				}
			}
		} else {
			// HTTP 代理
			proxyURL, _ := url.Parse("http://" + proxyAddr)
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	// 创建并配置HTTP客户端
	client := gclient.New()
	client.SetTimeout(30 * time.Second) // 总超时时间30秒
	client.Transport = transport        // 使用自定义 Transport

	return &Bitget{
		config:   config,
		endpoint: endpoint,
		client:   client,
	}
}

func (b *Bitget) GetName() string {
	return "bitget"
}

// getHttpClient 获取HTTP客户端(复用客户端)
func (b *Bitget) getHttpClient() *gclient.Client {
	return b.client
}

// GetBalance 获取账户余额
func (b *Bitget) GetBalance(ctx context.Context) (*Balance, error) {
	resp, err := b.signedRequest(ctx, "GET", "/api/v2/mix/account/accounts", map[string]string{
		"productType": "USDT-FUTURES",
	})
	if err != nil {
		return nil, err
	}

	json := gjson.New(resp)
	if json.Get("code").String() != "00000" {
		return nil, gerror.Newf("API error: %s", json.Get("msg").String())
	}

	data := json.Get("data").Array()
	for _, item := range data {
		j := gjson.New(item)
		if j.Get("marginCoin").String() == "USDT" {
			return &Balance{
				TotalBalance:     j.Get("accountEquity").Float64(),
				AvailableBalance: j.Get("available").Float64(),
				FrozenBalance:    j.Get("frozen").Float64(),
				UnrealizedPnl:    j.Get("unrealizedPL").Float64(),
				Currency:         "USDT",
			}, nil
		}
	}
	return &Balance{Currency: "USDT"}, nil
}

// GetTicker 获取行情
func (b *Bitget) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	// Bitget 在部分环境/接口下，USDT-FUTURES 的 symbol 可能需要带后缀（如 BTCUSDT_UMCBL）。
	// 这里做一次轻量兜底：先用标准 BTCUSDT 请求，若返回“标的不存在/precision null”等错误，再尝试 BTCUSDT_UMCBL。
	requestSymbols := []string{b.formatSymbol(symbol)}
	if !strings.Contains(requestSymbols[0], "_") {
		requestSymbols = append(requestSymbols, requestSymbols[0]+"_UMCBL")
	}

	var lastErr error
	var resp string
	var json *gjson.Json
	for _, reqSym := range requestSymbols {
		params := map[string]string{
			"symbol":      reqSym,
			"productType": "USDT-FUTURES",
		}
		resp, lastErr = b.publicRequest(ctx, "GET", "/api/v2/mix/market/ticker", params)
		if lastErr != nil {
			continue
		}

		j := gjson.New(resp)
		code := j.Get("code").String()
		if code != "00000" {
			lastErr = gerror.Newf("API error: %s", j.Get("msg").String())
			continue
		}

		data := j.Get("data").Array()
		if len(data) == 0 {
			lastErr = gerror.New("No ticker data")
			continue
		}
		json = j
		break
	}
	if json == nil {
		if lastErr == nil {
			lastErr = gerror.New("No ticker data")
		}
		return nil, lastErr
	}

	data := json.Get("data").Array()
	j := gjson.New(data[0])
	priceChangePercent := j.Get("change24h").Float64() * 100
	return &Ticker{
		Symbol:             symbol,
		LastPrice:          j.Get("lastPr").Float64(),
		BidPrice:           j.Get("bidPr").Float64(),
		AskPrice:           j.Get("askPr").Float64(),
		High24h:            j.Get("high24h").Float64(),
		Low24h:             j.Get("low24h").Float64(),
		Volume24h:          j.Get("baseVolume").Float64(),
		Change24h:          priceChangePercent,
		PriceChangePercent: priceChangePercent,
		Timestamp:          j.Get("ts").Int64(),
	}, nil
}

// GetKlines 获取K线数据
func (b *Bitget) GetKlines(ctx context.Context, symbol, interval string, limit int) ([]*Kline, error) {
	// 转换时间周期
	granularity := b.convertInterval(interval)

	requestSymbols := []string{b.formatSymbol(symbol)}
	if !strings.Contains(requestSymbols[0], "_") {
		requestSymbols = append(requestSymbols, requestSymbols[0]+"_UMCBL")
	}

	var lastErr error
	var json *gjson.Json
	for _, reqSym := range requestSymbols {
		params := map[string]string{
			"symbol":      reqSym,
			"productType": "USDT-FUTURES",
			"granularity": granularity,
			"limit":       strconv.Itoa(limit),
		}
		resp, err := b.publicRequest(ctx, "GET", "/api/v2/mix/market/candles", params)
		if err != nil {
			lastErr = err
			continue
		}
		j := gjson.New(resp)
		code := j.Get("code").String()
		if code != "00000" {
			lastErr = gerror.Newf("API error: %s", j.Get("msg").String())
			continue
		}
		json = j
		break
	}
	if json == nil {
		if lastErr == nil {
			lastErr = gerror.New("No kline data")
		}
		return nil, lastErr
	}

	var klines []*Kline
	items := json.Get("data").Array()
	for _, item := range items {
		arr := gjson.New(item).Array()
		if len(arr) >= 6 {
			klines = append(klines, &Kline{
				OpenTime:  g.NewVar(arr[0]).Int64(),
				Open:      g.NewVar(arr[1]).Float64(),
				High:      g.NewVar(arr[2]).Float64(),
				Low:       g.NewVar(arr[3]).Float64(),
				Close:     g.NewVar(arr[4]).Float64(),
				Volume:    g.NewVar(arr[5]).Float64(),
				CloseTime: g.NewVar(arr[0]).Int64() + b.getIntervalMs(interval),
			})
		}
	}
	return klines, nil
}

// GetPositions 获取持仓
func (b *Bitget) GetPositions(ctx context.Context, symbol string) ([]*Position, error) {
	// Bitget V2 API 要求: 如果指定symbol，使用single-position；否则使用all-position
	var apiPath string
	params := map[string]string{
		"productType": "USDT-FUTURES",
		"marginCoin":  "USDT",
	}

	if symbol != "" {
		// 使用单个持仓接口
		apiPath = "/api/v2/mix/position/single-position"
		params["symbol"] = b.formatSymbol(symbol)
	} else {
		// 使用全部持仓接口
		apiPath = "/api/v2/mix/position/all-position"
	}

	g.Log().Debugf(ctx, "[Bitget] GetPositions: path=%s, symbol=%s, params=%+v", apiPath, symbol, params)

	resp, err := b.signedRequest(ctx, "GET", apiPath, params)
	if err != nil {
		g.Log().Warningf(ctx, "[Bitget] GetPositions request failed: %v", err)
		return nil, err
	}

	g.Log().Debugf(ctx, "[Bitget] GetPositions response: %s", resp)

	json := gjson.New(resp)
	code := json.Get("code").String()
	if code != "00000" {
		errMsg := json.Get("msg").String()
		g.Log().Warningf(ctx, "[Bitget] GetPositions API error: code=%s, msg=%s, path=%s", code, errMsg, apiPath)
		return nil, gerror.Newf("API error: code=%s, msg=%s", code, errMsg)
	}

	var positions []*Position
	items := json.Get("data").Array()

	// 兜底：Bitget 指定 symbol 走 single-position 时，symbol 格式可能需要后缀（如 *_UMCBL），导致返回空。
	// 若返回空且用户确实有持仓，改用 all-position 再在本地过滤。
	if symbol != "" && len(items) == 0 {
		fallbackPath := "/api/v2/mix/position/all-position"
		fallbackParams := map[string]string{
			"productType": "USDT-FUTURES",
			"marginCoin":  "USDT",
		}
		fallbackResp, fallbackErr := b.signedRequest(ctx, "GET", fallbackPath, fallbackParams)
		if fallbackErr == nil {
			fj := gjson.New(fallbackResp)
			fcode := fj.Get("code").String()
			if fcode == "00000" {
				resp = fallbackResp
				apiPath = fallbackPath
				json = fj
				items = fj.Get("data").Array()
				g.Log().Debugf(ctx, "[Bitget] GetPositions fallback to all-position: items=%d, symbol=%s", len(items), symbol)
			} else {
				g.Log().Warningf(ctx, "[Bitget] GetPositions fallback API error: code=%s, msg=%s", fcode, fj.Get("msg").String())
			}
		} else {
			g.Log().Warningf(ctx, "[Bitget] GetPositions fallback request failed: %v", fallbackErr)
		}
	}

	// 【重要】记录原始API响应，用于调试（避免在高频轮询/WS推送时刷屏，改为 Debug）
	g.Log().Debugf(ctx, "[Bitget] GetPositions API响应: code=%s, data数组长度=%d, 请求symbol=%s, 完整响应=%s", code, len(items), symbol, resp)
	if len(items) == 0 {
		g.Log().Warningf(ctx, "[Bitget] GetPositions 返回空数组，可能没有持仓或API响应格式异常: symbol=%s", symbol)
	}

	// 规范化 symbol，用于匹配（兼容带后缀的合约标识，例如 BTCUSDT_UMCBL）
	normalizeSymbolKey := func(s string) string {
		if s == "" {
			return ""
		}
		s = strings.ToUpper(strings.TrimSpace(s))
		s = strings.ReplaceAll(s, "/", "")
		s = strings.ReplaceAll(s, "_", "")
		s = strings.ReplaceAll(s, "-", "")
		return s
	}
	requestSymbolKey := normalizeSymbolKey(symbol)

	for i, item := range items {
		j := gjson.New(item)
		// Bitget 持仓数量字段在不同接口/版本可能存在差异：
		// - 常见字段：total / available
		// - 若数量字段为0但保证金/杠杆/开仓均价存在，则用 margin*leverage/openPriceAvg 反推一个近似数量（用于展示与风控判断）
		posAmt := j.Get("total").Float64()
		if posAmt == 0 {
			posAmt = j.Get("available").Float64()
		}
		holdSideRaw := j.Get("holdSide").String()
		returnedSymbol := j.Get("symbol").String()
		returnedSymbolKey := normalizeSymbolKey(returnedSymbol)

		// 【重要】记录原始数据，用于调试（高频场景降为 Debug）
		itemJson := gjson.New(item)
		matched := requestSymbolKey == "" || returnedSymbolKey == requestSymbolKey || strings.HasPrefix(returnedSymbolKey, requestSymbolKey)
		g.Log().Debugf(ctx, "[Bitget] GetPositions 解析持仓[%d]: 请求symbol=%s, 返回symbol=%s, symbol匹配=%v, holdSide=%s, total=%.6f, openPriceAvg=%.2f, 完整数据=%s",
			i, requestSymbolKey, returnedSymbolKey, matched, holdSideRaw, posAmt, j.Get("openPriceAvg").Float64(), itemJson.String())

		// 【修复】如果指定了 symbol，只返回匹配的持仓
		if symbol != "" && !matched {
			g.Log().Debugf(ctx, "[Bitget] GetPositions 跳过不匹配的持仓: 请求symbol=%s, 返回symbol=%s", requestSymbolKey, returnedSymbolKey)
			continue
		}

		// 即使没有持仓也返回，让调用方判断
		// if posAmt == 0 {
		// 	continue
		// }

		// 【修复】规范化 holdSide，确保大小写不敏感的比较
		posSide := "LONG"
		holdSideLower := strings.ToLower(strings.TrimSpace(holdSideRaw))
		if holdSideLower == "short" {
			posSide = "SHORT"
			// 空单的持仓数量应该是负数（或者绝对值，取决于交易所返回格式）
			// Bitget 返回的 total 字段，空单时应该已经是负数或需要取反
			if posAmt > 0 {
				posAmt = -posAmt
			}
		} else if holdSideLower != "long" {
			// 如果 holdSide 不是 "long" 也不是 "short"，记录警告
			g.Log().Warningf(ctx, "[Bitget] GetPositions 未知的 holdSide 值: symbol=%s, holdSide=%s, 默认使用 LONG", symbol, holdSideRaw)
		}

		// 如果数量仍为0，但有保证金/杠杆/开仓价，反推数量（避免被上层过滤导致“明明有仓位但显示为空”）
		if math.Abs(posAmt) == 0 {
			entryPriceTmp := j.Get("openPriceAvg").Float64()
			levTmp := j.Get("leverage").Float64()
			marginTmp := j.Get("margin").Float64()
			if entryPriceTmp > 0 && levTmp > 0 && marginTmp > 0 {
				derived := (marginTmp * levTmp) / entryPriceTmp
				if posSide == "SHORT" {
					derived = -derived
				}
				posAmt = derived
				g.Log().Debugf(ctx, "[Bitget] GetPositions 通过 margin/leverage/openPriceAvg 反推持仓数量: symbol=%s, posSide=%s, margin=%.6f, leverage=%.2f, entry=%.6f, qty=%.8f",
					returnedSymbol, posSide, marginTmp, levTmp, entryPriceTmp, posAmt)
			}
		}

		g.Log().Debugf(ctx, "[Bitget] GetPositions 解析结果[%d]: symbol=%s, holdSideRaw=%s, holdSideLower=%s, posSide=%s, posAmt=%.6f",
			i, returnedSymbol, holdSideRaw, holdSideLower, posSide, posAmt)

		// 计算保证金：持仓价值 / 杠杆
		entryPrice := j.Get("openPriceAvg").Float64()
		leverage := j.Get("leverage").Int()
		margin := 0.0
		if leverage > 0 {
			margin = (math.Abs(posAmt) * entryPrice) / float64(leverage)
		}

		positions = append(positions, &Position{
			Symbol:           returnedSymbol, // 使用返回的原始 symbol，保持一致性
			PositionSide:     posSide,
			PositionAmt:      posAmt,
			EntryPrice:       entryPrice,
			MarkPrice:        j.Get("markPrice").Float64(),
			UnrealizedPnl:    j.Get("unrealizedPL").Float64(),
			Leverage:         leverage,
			Margin:           margin,
			MarginType:       j.Get("marginMode").String(),
			IsolatedMargin:   j.Get("margin").Float64(),
			LiquidationPrice: j.Get("liquidationPrice").Float64(),
		})
	}

	g.Log().Infof(ctx, "[Bitget] GetPositions 解析完成: 原始数据项=%d, 解析后持仓数=%d", len(items), len(positions))
	if len(positions) > 0 {
		for i, pos := range positions {
			g.Log().Infof(ctx, "[Bitget] GetPositions 最终持仓[%d]: symbol=%s, PositionSide=%s, PositionAmt=%.6f, UnrealizedPnl=%.4f",
				i, pos.Symbol, pos.PositionSide, pos.PositionAmt, pos.UnrealizedPnl)
		}
	} else {
		g.Log().Warningf(ctx, "[Bitget] GetPositions 未解析到任何持仓，原始数据项数=%d", len(items))
	}
	return positions, nil
}

// CreateOrder 创建订单
// 【重要修复】添加 holdSide 参数，支持双向持仓模式(hedge-mode)
// 在双向持仓模式下，不传 holdSide 会导致开仓时自动平掉对面方向的仓位！
func (b *Bitget) CreateOrder(ctx context.Context, req *OrderRequest) (*Order, error) {
	side := "buy"
	if req.Side == "SELL" {
		side = "sell"
	}

	tradeSide := "open"
	if req.ReduceOnly {
		tradeSide = "close"
	}

	orderType := "market"
	if req.Type == "LIMIT" {
		orderType = "limit"
	}

	// 【关键】根据 PositionSide 确定 holdSide（双向持仓模式必需）
	// - PositionSide=LONG → holdSide=long（开多/平多）
	// - PositionSide=SHORT → holdSide=short（开空/平空）
	holdSide := "long"
	if req.PositionSide == "SHORT" {
		holdSide = "short"
	}

	params := map[string]string{
		"symbol":      b.formatSymbol(req.Symbol),
		"productType": "USDT-FUTURES",
		"marginMode":  "isolated",
		"marginCoin":  "USDT",
		"side":        side,
		"tradeSide":   tradeSide,
		"orderType":   orderType,
		"size":        strconv.FormatFloat(req.Quantity, 'f', -1, 64),
		"holdSide":    holdSide, // 【关键修复】添加 holdSide，防止开仓时平掉对面方向的仓位
	}

	g.Log().Infof(ctx, "[Bitget] CreateOrder: symbol=%s, side=%s, tradeSide=%s, holdSide=%s, positionSide=%s, quantity=%f",
		req.Symbol, side, tradeSide, holdSide, req.PositionSide, req.Quantity)

	if req.Price > 0 {
		params["price"] = strconv.FormatFloat(req.Price, 'f', -1, 64)
	}

	resp, err := b.signedRequest(ctx, "POST", "/api/v2/mix/order/place-order", params)
	if err != nil {
		return nil, err
	}

	json := gjson.New(resp)
	if json.Get("code").String() != "00000" {
		return nil, gerror.Newf("API error: %s", json.Get("msg").String())
	}

	data := gjson.New(json.Get("data").Val())

	// 【优化】优先使用交易所返回的创建时间，如果不存在则使用本地时间
	createTime := time.Now().UnixMilli() // 默认使用本地时间
	if cTime := data.Get("cTime").Int64(); cTime > 0 {
		// Bitget API 返回的时间可能是秒或毫秒，需要判断
		// 如果小于 1e12，说明是秒级时间戳，需要转换为毫秒
		if cTime < 1e12 {
			createTime = cTime * 1000
		} else {
			createTime = cTime
		}
	} else if orderTime := data.Get("orderTime").Int64(); orderTime > 0 {
		// 尝试其他可能的时间字段名
		if orderTime < 1e12 {
			createTime = orderTime * 1000
		} else {
			createTime = orderTime
		}
	} else if ts := json.Get("ts").Int64(); ts > 0 {
		// 使用响应中的时间戳（如果存在）
		if ts < 1e12 {
			createTime = ts * 1000
		} else {
			createTime = ts
		}
	}

	return &Order{
		OrderId:      data.Get("orderId").String(),
		ClientId:     data.Get("clientOid").String(),
		Symbol:       req.Symbol,
		Side:         req.Side,
		PositionSide: req.PositionSide,
		Type:         req.Type,
		Quantity:     req.Quantity,
		Price:        req.Price,
		Status:       "NEW",
		CreateTime:   createTime, // 【优化】使用交易所返回的时间（如果存在）
	}, nil
}

// CancelOrder 取消订单
func (b *Bitget) CancelOrder(ctx context.Context, symbol, orderId string) (*Order, error) {
	params := map[string]string{
		"symbol":      b.formatSymbol(symbol),
		"productType": "USDT-FUTURES",
		"orderId":     orderId,
	}

	resp, err := b.signedRequest(ctx, "POST", "/api/v2/mix/order/cancel-order", params)
	if err != nil {
		return nil, err
	}

	json := gjson.New(resp)
	if json.Get("code").String() != "00000" {
		return nil, gerror.Newf("API error: %s", json.Get("msg").String())
	}

	return &Order{
		OrderId: orderId,
		Symbol:  symbol,
		Status:  "CANCELED",
	}, nil
}

// ClosePosition 平仓
func (b *Bitget) ClosePosition(ctx context.Context, symbol, positionSide string, quantity float64) (*Order, error) {
	// 【重要】规范化 positionSide，确保大小写一致
	positionSide = strings.ToUpper(strings.TrimSpace(positionSide))
	if positionSide != "LONG" && positionSide != "SHORT" {
		return nil, gerror.Newf("持仓方向无效: %s，必须是 LONG 或 SHORT", positionSide)
	}

	// 【修复】Bitget hedge-mode 平仓参数（2025-12-08）
	// 根据 Bitget API 文档：
	// - hedge-mode 下 side 表示持仓方向：buy=多头方向, sell=空头方向
	// - tradeSide=close 表示平仓操作
	// - 平仓多头(LONG): side=buy, tradeSide=close, holdSide=long
	// - 平仓空头(SHORT): side=sell, tradeSide=close, holdSide=short
	side := "buy" // 多头方向
	holdSide := "long"
	if positionSide == "SHORT" {
		side = "sell" // 空头方向
		holdSide = "short"
	}

	// 【修复】使用 isolated 模式（逐仓），因为机器人通常使用逐仓模式
	// 注意：如果机器人使用全仓模式，需要从配置中获取
	params := map[string]string{
		"symbol":      b.formatSymbol(symbol),
		"productType": "USDT-FUTURES",
		"marginMode":  "isolated", // 使用逐仓模式（机器人通常使用逐仓）
		"marginCoin":  "USDT",
		"side":        side,
		"tradeSide":   "close",
		"orderType":   "market",
		"size":        strconv.FormatFloat(quantity, 'f', -1, 64),
		"holdSide":    holdSide, // 【关键】holdSide 必须正确匹配要平仓的持仓方向
	}

	g.Log().Infof(ctx, "[Bitget] ClosePosition: symbol=%s, positionSide=%s, side=%s, holdSide=%s, quantity=%f, params=%+v",
		symbol, positionSide, side, holdSide, quantity, params)

	resp, err := b.signedRequest(ctx, "POST", "/api/v2/mix/order/place-order", params)
	if err != nil {
		g.Log().Errorf(ctx, "[Bitget] ClosePosition request error: %v", err)
		return nil, err
	}

	g.Log().Debugf(ctx, "[Bitget] ClosePosition response: %s", resp)

	json := gjson.New(resp)
	if json.Get("code").String() != "00000" {
		errMsg := json.Get("msg").String()
		g.Log().Errorf(ctx, "[Bitget] ClosePosition API error: %s", errMsg)
		return nil, gerror.Newf("平仓失败: %s", errMsg)
	}

	data := gjson.New(json.Get("data").Val())
	return &Order{
		OrderId:      data.Get("orderId").String(),
		Symbol:       symbol,
		Side:         strings.ToUpper(side),
		PositionSide: positionSide,
		Type:         "MARKET",
		Quantity:     quantity,
		Status:       "FILLED",
	}, nil
}

// SetLeverage 设置杠杆
func (b *Bitget) SetLeverage(ctx context.Context, symbol string, leverage int) error {
	params := map[string]string{
		"symbol":      b.formatSymbol(symbol),
		"productType": "USDT-FUTURES",
		"marginCoin":  "USDT",
		"leverage":    strconv.Itoa(leverage),
	}

	resp, err := b.signedRequest(ctx, "POST", "/api/v2/mix/account/set-leverage", params)
	if err != nil {
		return err
	}

	json := gjson.New(resp)
	if json.Get("code").String() != "00000" {
		return gerror.Newf("API error: %s", json.Get("msg").String())
	}
	return nil
}

// SetMarginType 设置保证金模式
func (b *Bitget) SetMarginType(ctx context.Context, symbol string, marginType string) error {
	mode := "isolated"
	if marginType == "CROSSED" {
		mode = "crossed"
	}

	params := map[string]string{
		"symbol":      b.formatSymbol(symbol),
		"productType": "USDT-FUTURES",
		"marginCoin":  "USDT",
		"marginMode":  mode,
	}

	resp, err := b.signedRequest(ctx, "POST", "/api/v2/mix/account/set-margin-mode", params)
	if err != nil {
		return err
	}

	json := gjson.New(resp)
	if json.Get("code").String() != "00000" {
		return gerror.Newf("API error: %s", json.Get("msg").String())
	}
	return nil
}

// formatSymbol 格式化交易对
func (b *Bitget) formatSymbol(symbol string) string {
	// 使用统一的Symbol格式化器
	return Formatter.FormatForBitget(symbol) // BTCUSDT
}

// convertInterval 转换K线周期
func (b *Bitget) convertInterval(interval string) string {
	mapping := map[string]string{
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
	}
	if v, ok := mapping[interval]; ok {
		return v
	}
	return "5m"
}

// getIntervalMs 获取K线周期毫秒数
func (b *Bitget) getIntervalMs(interval string) int64 {
	mapping := map[string]int64{
		"1m":  60000,
		"3m":  180000,
		"5m":  300000,
		"15m": 900000,
		"30m": 1800000,
		"1h":  3600000,
		"2h":  7200000,
		"4h":  14400000,
		"6h":  21600000,
		"12h": 43200000,
		"1d":  86400000,
		"1w":  604800000,
	}
	if v, ok := mapping[interval]; ok {
		return v
	}
	return 300000
}

// publicRequest 公开请求
func (b *Bitget) publicRequest(ctx context.Context, method, path string, params map[string]string) (string, error) {
	// 【限频优化】先计算需要等待的时间，然后释放锁，在锁外等待
	b.reqMu.Lock()
	elapsed := time.Since(b.lastRequest)
	waitTime := time.Duration(0)
	if elapsed < 100*time.Millisecond {
		waitTime = 100*time.Millisecond - elapsed
	}
	b.lastRequest = time.Now().Add(waitTime)
	b.reqMu.Unlock()

	if waitTime > 0 {
		time.Sleep(waitTime)
	}

	client := b.getHttpClient()

	reqUrl := b.endpoint + path
	if len(params) > 0 {
		reqUrl += "?" + b.buildQuery(params)
	}

	var resp *gclient.Response
	var err error
	if method == "GET" {
		resp, err = client.Get(ctx, reqUrl)
	} else {
		resp, err = client.Post(ctx, reqUrl)
	}
	if err != nil {
		return "", gerror.Wrap(err, "Request failed")
	}
	defer resp.Close()

	body := resp.ReadAllString()
	return body, nil
}

// signedRequest 签名请求
func (b *Bitget) signedRequest(ctx context.Context, method, path string, params map[string]string) (string, error) {
	// 【限频优化】先计算需要等待的时间，然后释放锁，在锁外等待
	// 避免在持锁情况下 sleep 导致其他请求长时间阻塞
	b.reqMu.Lock()
	elapsed := time.Since(b.lastRequest)
	waitTime := time.Duration(0)
	if elapsed < 100*time.Millisecond {
		waitTime = 100*time.Millisecond - elapsed
	}
	b.lastRequest = time.Now().Add(waitTime) // 预先记录时间，避免其他请求同时进入
	b.reqMu.Unlock()

	// 在锁外等待，不阻塞其他请求获取锁
	if waitTime > 0 {
		time.Sleep(waitTime)
	}

	// 【修复时间戳过期问题】在发送请求前重新生成时间戳，确保时间戳是最新的
	// 避免在限频等待期间时间戳过期
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	var body string
	queryString := ""

	if method == "GET" && len(params) > 0 {
		queryString = "?" + b.buildQuery(params)
	} else if method == "POST" && len(params) > 0 {
		bodyJson, _ := gjson.New(params).ToJsonString()
		body = bodyJson
	}

	// 签名字符串: timestamp + method + path + queryString + body
	preSign := timestamp + method + path + queryString + body
	signature := b.sign(preSign)

	// 调试日志
	g.Log().Debugf(ctx, "[Bitget] Request: method=%s, path=%s, queryString=%s", method, path, queryString)
	g.Log().Debugf(ctx, "[Bitget] PreSign string: %s", preSign)

	client := b.getHttpClient()
	client.SetHeader("ACCESS-KEY", b.config.ApiKey)
	client.SetHeader("ACCESS-SIGN", signature)
	client.SetHeader("ACCESS-TIMESTAMP", timestamp)
	client.SetHeader("ACCESS-PASSPHRASE", b.config.Passphrase)
	client.SetHeader("Content-Type", "application/json")
	client.SetHeader("locale", "zh-CN")

	reqUrl := b.endpoint + path + queryString

	// 【修复时间戳过期问题】添加重试机制，如果遇到时间戳过期错误，重新生成时间戳并重试
	maxRetries := 1
	var respBody string
	for retry := 0; retry <= maxRetries; retry++ {
		if retry > 0 {
			// 重试前重新生成时间戳和签名
			timestamp = strconv.FormatInt(time.Now().UnixMilli(), 10)
			preSign = timestamp + method + path + queryString + body
			signature = b.sign(preSign)
			client.SetHeader("ACCESS-SIGN", signature)
			client.SetHeader("ACCESS-TIMESTAMP", timestamp)
			g.Log().Debugf(ctx, "[Bitget] Retry request with new timestamp: %s", timestamp)
		}

		var resp *gclient.Response
		var err error
		if method == "GET" {
			resp, err = client.Get(ctx, reqUrl)
		} else {
			resp, err = client.Post(ctx, reqUrl, body)
		}

		if err != nil {
			g.Log().Warningf(ctx, "[Bitget] HTTP request failed: %v", err)
			if retry < maxRetries {
				continue // 重试
			}
			return "", gerror.Wrap(err, "Request failed")
		}

		respBody = resp.ReadAllString()
		resp.Close()

		// 检查是否有错误
		json := gjson.New(respBody)
		code := json.Get("code").String()
		if code != "" && code != "00000" {
			msg := json.Get("msg").String()
			g.Log().Warningf(ctx, "[Bitget] API Error: code=%s, msg=%s, requestId=%s",
				code, msg, json.Get("requestId").String())

			// 如果是时间戳过期错误（40008），且还有重试次数，则重试
			if code == "40008" && strings.Contains(msg, "时间戳") && retry < maxRetries {
				g.Log().Infof(ctx, "[Bitget] Timestamp expired, retrying with new timestamp...")
				continue // 重试
			}
		}

		return respBody, nil
	}

	return "", gerror.New("Request failed after retries")
}

// buildQuery 构建查询字符串
func (b *Bitget) buildQuery(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	return strings.Join(parts, "&")
}

// sign 签名
func (b *Bitget) sign(data string) string {
	h := hmac.New(sha256.New, []byte(b.config.SecretKey))
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// GetOrderHistory 获取成交明细
func (b *Bitget) GetOrderHistory(ctx context.Context, symbol string, limit int) ([]*Order, error) {
	// Bitget V2 API 成交明细接口需要 symbol 参数
	if symbol == "" {
		return nil, gerror.New("symbol is required for fills")
	}

	params := map[string]string{
		"productType": "USDT-FUTURES",
		"symbol":      b.formatSymbol(symbol),
	}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}

	g.Log().Debugf(ctx, "[Bitget] GetFills: symbol=%s, limit=%d, params=%+v", symbol, limit, params)

	// Bitget V2 API 成交明细接口
	resp, err := b.signedRequest(ctx, "GET", "/api/v2/mix/order/fills", params)
	if err != nil {
		g.Log().Warningf(ctx, "[Bitget] GetFills request failed: %v", err)
		return nil, err
	}

	g.Log().Debugf(ctx, "[Bitget] GetFills response: %s", resp)

	json := gjson.New(resp)
	code := json.Get("code").String()
	if code != "00000" {
		errMsg := json.Get("msg").String()
		g.Log().Warningf(ctx, "[Bitget] GetFills API error: code=%s, msg=%s", code, errMsg)
		return nil, gerror.Newf("API error: code=%s, msg=%s", code, errMsg)
	}

	// Bitget 的 fills 是“成交(fill)级”，同一个订单可能返回多条成交明细。
	// 为了让上层（订单同步/历史查询）能按订单维度稳定匹配，这里按 orderId 聚合为“订单级”数据：
	// - OrderId = orderId（交易所订单ID）
	// - ClientId = clientOid（若有）
	// - AvgPrice/Quantity/Fee/Time 做汇总
	type fillAgg struct {
		orderID     string
		clientID    string
		symbol      string
		side        string
		orderType   string
		tradeScope  string
		sumQty      float64
		sumPriceQty float64
		sumFee      float64
		feeCoin     string
		minTs       int64
		maxTs       int64
	}
	aggMap := make(map[string]*fillAgg)
	// Bitget V2 成交明细返回的数据结构是 data.fillList
	items := json.Get("data.fillList").Array()
	if len(items) == 0 {
		items = json.Get("data").Array()
	}
	g.Log().Debugf(ctx, "[Bitget] GetFills: found %d fills", len(items))

	for _, item := range items {
		j := gjson.New(item)

		// Bitget V2 成交明细字段映射:
		// tradeId - 成交ID
		// orderId - 订单ID
		// symbol - 交易对
		// side - 买卖方向
		// price - 成交价格
		// baseVolume - 成交数量
		// quoteVolume - 成交金额
		// tradeSide - 交易方向 (open/close)
		// tradeScope - 流动性方向 (maker/taker)
		// feeDetail - 手续费详情
		// cTime - 成交时间

		price := j.Get("price").Float64()
		quantity := j.Get("baseVolume").Float64()
		if quantity == 0 {
			quantity = j.Get("size").Float64()
		}

		// 成交明细中 side 可能是 buy/sell
		side := strings.ToUpper(j.Get("side").String())

		// tradeSide 表示开仓/平仓
		tradeSide := j.Get("tradeSide").String()
		orderType := "MARKET"
		if tradeSide == "open" {
			orderType = "开仓"
		} else if tradeSide == "close" {
			orderType = "平仓"
		}

		// 流动性方向 (maker/taker)
		tradeScope := j.Get("tradeScope").String()

		// 手续费解析 - feeDetail 是一个数组，取第一个
		var fee float64
		var feeCoin string
		feeDetails := j.Get("feeDetail").Array()
		if len(feeDetails) > 0 {
			feeJ := gjson.New(feeDetails[0])
			fee = feeJ.Get("totalFee").Float64()
			feeCoin = feeJ.Get("feeCoin").String()
		}

		orderID := strings.TrimSpace(j.Get("orderId").String())
		if orderID == "" {
			// 没有 orderId 无法对齐本地订单，跳过
			continue
		}
		a := aggMap[orderID]
		if a == nil {
			a = &fillAgg{orderID: orderID}
			aggMap[orderID] = a
		}
		if a.clientID == "" {
			a.clientID = strings.TrimSpace(j.Get("clientOid").String())
		}
		if a.symbol == "" {
			a.symbol = j.Get("symbol").String()
		}
		if a.side == "" {
			a.side = side
		}
		if a.orderType == "" {
			a.orderType = orderType
		}
		if a.tradeScope == "" {
			a.tradeScope = tradeScope
		}
		if a.feeCoin == "" && feeCoin != "" {
			a.feeCoin = feeCoin
		}
		if quantity > 0 && price > 0 {
			a.sumQty += quantity
			a.sumPriceQty += price * quantity
		}
		a.sumFee += fee
		ts := j.Get("cTime").Int64()
		if ts > 0 {
			if a.minTs == 0 || ts < a.minTs {
				a.minTs = ts
			}
			if ts > a.maxTs {
				a.maxTs = ts
			}
		}
	}
	orders := make([]*Order, 0, len(aggMap))
	for _, a := range aggMap {
		if a == nil {
			continue
		}
		avg := 0.0
		if a.sumQty > 0 {
			avg = a.sumPriceQty / a.sumQty
		}
		createTs := a.minTs
		updateTs := a.maxTs
		if createTs == 0 {
			createTs = updateTs
		}
		orders = append(orders, &Order{
			OrderId:    a.orderID,  // ✅ 订单ID（用于对齐本地 exchange_order_id/close_order_id）
			ClientId:   a.clientID, // 客户端订单ID（可能为空）
			Symbol:     a.symbol,
			Side:       a.side,
			Type:       a.orderType,
			Price:      avg,
			Quantity:   a.sumQty,
			FilledQty:  a.sumQty,
			AvgPrice:   avg,
			Status:     "FILLED",
			TradeScope: a.tradeScope,
			Fee:        a.sumFee,
			FeeCoin:    a.feeCoin,
			CreateTime: createTs,
			UpdateTime: updateTs,
		})
	}
	return orders, nil
}

// GetTradeHistory 获取成交记录（用于财务对账/已实现盈亏/手续费汇总）
// Bitget V2：GET /api/v2/mix/order/fills
func (b *Bitget) GetTradeHistory(ctx context.Context, symbol string, limit int) ([]*Trade, error) {
	// Bitget V2 API 成交明细接口需要 symbol 参数
	if symbol == "" {
		return nil, gerror.New("symbol is required for fills")
	}

	params := map[string]string{
		"productType": "USDT-FUTURES",
		"symbol":      b.formatSymbol(symbol),
	}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}

	resp, err := b.signedRequest(ctx, "GET", "/api/v2/mix/order/fills", params)
	if err != nil {
		return nil, err
	}

	json := gjson.New(resp)
	code := json.Get("code").String()
	if code != "00000" {
		return nil, gerror.Newf("API error: code=%s, msg=%s", code, json.Get("msg").String())
	}

	items := json.Get("data.fillList").Array()
	if len(items) == 0 {
		items = json.Get("data").Array()
	}

	var out []*Trade
	for _, it := range items {
		j := gjson.New(it)

		price := j.Get("price").Float64()
		qty := j.Get("baseVolume").Float64()
		if qty == 0 {
			qty = j.Get("size").Float64()
		}

		// side: buy/sell
		side := strings.ToUpper(j.Get("side").String())

		// 尝试识别持仓方向（部分返回包含 holdSide/posSide/positionSide）
		posSide := strings.ToUpper(j.Get("holdSide").String())
		if posSide == "" {
			posSide = strings.ToUpper(j.Get("posSide").String())
		}
		if posSide == "" {
			posSide = strings.ToUpper(j.Get("positionSide").String())
		}

		// 手续费解析 - feeDetail 是一个数组，取第一个
		var fee float64
		var feeCoin string
		feeDetails := j.Get("feeDetail").Array()
		if len(feeDetails) > 0 {
			feeJ := gjson.New(feeDetails[0])
			fee = feeJ.Get("totalFee").Float64()
			feeCoin = feeJ.Get("feeCoin").String()
		}

		// 已实现盈亏：不同接口字段可能不同，尽量兼容
		rp := j.Get("realizedPnl").Float64()
		if rp == 0 {
			rp = j.Get("pnl").Float64()
		}
		if rp == 0 {
			rp = j.Get("profit").Float64()
		}

		out = append(out, &Trade{
			TradeId:         j.Get("tradeId").String(),
			OrderId:         j.Get("orderId").String(),
			Symbol:          symbol,
			Side:            side,
			PositionSide:    posSide,
			Price:           price,
			Quantity:        qty,
			RealizedPnl:     rp,
			Commission:      fee,
			CommissionAsset: feeCoin,
			Time:            j.Get("cTime").Int64(),
		})
	}
	return out, nil
}

// GetOpenOrders 获取当前挂单
func (b *Bitget) GetOpenOrders(ctx context.Context, symbol string) ([]*Order, error) {
	params := map[string]string{
		"productType": "USDT-FUTURES",
	}
	if symbol != "" {
		params["symbol"] = b.formatSymbol(symbol)
	}

	resp, err := b.signedRequest(ctx, "GET", "/api/v2/mix/order/orders-pending", params)
	if err != nil {
		return nil, err
	}

	json := gjson.New(resp)
	if json.Get("code").String() != "00000" {
		return nil, gerror.Newf("API error: %s", json.Get("msg").String())
	}

	var orders []*Order
	items := json.Get("data.entrustedList").Array()
	for _, item := range items {
		j := gjson.New(item)
		orders = append(orders, &Order{
			OrderId:    j.Get("orderId").String(),
			ClientId:   j.Get("clientOid").String(),
			Symbol:     j.Get("symbol").String(),
			Side:       strings.ToUpper(j.Get("side").String()),
			Type:       strings.ToUpper(j.Get("orderType").String()),
			Price:      j.Get("price").Float64(),
			Quantity:   j.Get("size").Float64(),
			FilledQty:  j.Get("filledQty").Float64(),
			Status:     "NEW",
			CreateTime: j.Get("cTime").Int64(),
		})
	}
	return orders, nil
}

// GetSymbolInfo 获取交易对信息
func (b *Bitget) GetSymbolInfo(ctx context.Context, symbol string) (*SymbolInfo, error) {
	params := map[string]string{
		"productType": "USDT-FUTURES",
		"symbol":      b.formatSymbol(symbol),
	}

	resp, err := b.publicRequest(ctx, "GET", "/api/v2/mix/market/contracts", params)
	if err != nil {
		return nil, err
	}

	json := gjson.New(resp)
	if json.Get("code").String() != "00000" {
		return nil, gerror.Newf("API error: %s", json.Get("msg").String())
	}

	data := json.Get("data").Array()
	if len(data) == 0 {
		return nil, gerror.New("交易对不存在")
	}

	j := gjson.New(data[0])
	return &SymbolInfo{
		Symbol:          j.Get("symbol").String(),
		BaseCoin:        j.Get("baseCoin").String(),
		QuoteCoin:       j.Get("quoteCoin").String(),
		PricePrecision:  j.Get("pricePlace").Int(),
		QtyPrecision:    j.Get("volumePlace").Int(),
		MinQty:          j.Get("minTradeNum").Float64(),
		MaxLeverage:     j.Get("maxLever").Int(),
		ContractSize:    j.Get("sizeMultiplier").Float64(),
		MinNotionalUSDT: j.Get("minTradeUSDT").Float64(),
	}, nil
}

// GetAllSymbols 获取所有交易对
func (b *Bitget) GetAllSymbols(ctx context.Context) ([]*SymbolInfo, error) {
	params := map[string]string{
		"productType": "USDT-FUTURES",
	}

	resp, err := b.publicRequest(ctx, "GET", "/api/v2/mix/market/contracts", params)
	if err != nil {
		return nil, err
	}

	json := gjson.New(resp)
	if json.Get("code").String() != "00000" {
		return nil, gerror.Newf("API error: %s", json.Get("msg").String())
	}

	var symbols []*SymbolInfo
	items := json.Get("data").Array()
	for _, item := range items {
		j := gjson.New(item)
		symbols = append(symbols, &SymbolInfo{
			Symbol:          j.Get("symbol").String(),
			BaseCoin:        j.Get("baseCoin").String(),
			QuoteCoin:       j.Get("quoteCoin").String(),
			PricePrecision:  j.Get("pricePlace").Int(),
			QtyPrecision:    j.Get("volumePlace").Int(),
			MinQty:          j.Get("minTradeNum").Float64(),
			MaxLeverage:     j.Get("maxLever").Int(),
			ContractSize:    j.Get("sizeMultiplier").Float64(),
			MinNotionalUSDT: j.Get("minTradeUSDT").Float64(),
		})
	}
	return symbols, nil
}

// parseOrderStatus 解析订单状态
func (b *Bitget) parseOrderStatus(state string) string {
	// Bitget V2 状态值可能是小写或大写
	lowerState := strings.ToLower(state)
	switch lowerState {
	case "live", "new", "init":
		return "NEW"
	case "partially_filled", "partial-fill":
		return "PARTIALLY_FILLED"
	case "filled", "full-fill":
		return "FILLED"
	case "cancelled", "canceled", "cancel":
		return "CANCELED"
	case "rejected":
		return "REJECTED"
	case "expired":
		return "EXPIRED"
	default:
		if state == "" {
			return "UNKNOWN"
		}
		return strings.ToUpper(state)
	}
}
