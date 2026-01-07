// Package exchange Gate.io 交易所API（U本位永续 / 逐仓 / 双向持仓）
package exchange

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"regexp"
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

var gateZeroParseLogAt sync.Map // key: string, value: time.Time

var gateOrderIDRegex = regexp.MustCompile(`\d{16,}`)

func gateShouldLog(key string, every time.Duration) bool {
	if strings.TrimSpace(key) == "" {
		return false
	}
	now := time.Now()
	if v, ok := gateZeroParseLogAt.Load(key); ok {
		if t0, ok2 := v.(time.Time); ok2 && now.Sub(t0) < every {
			return false
		}
	}
	gateZeroParseLogAt.Store(key, now)
	return true
}

func gateExtractOrderIDFromText(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// try to find a long digit sequence (Gate order ids are long integers)
	m := gateOrderIDRegex.FindString(s)
	return strings.TrimSpace(m)
}

// Gate Gate.io（API v4, futures/usdt）
// 约束：仅实现 USDT 永续（futures/usdt） + 逐仓（isolated） + 双向持仓（long/short）
type Gate struct {
	config   *Config
	endpoint string

	mu sync.Mutex
	// contract -> quanto_multiplier（合约面值，单位：基础币），用于将 size(张) 转为基础币数量
	contractMultiplier map[string]float64
	// contract -> 下单张数最小值/步进（用于对齐 size 精度，避免 Gate 报错）
	contractOrderSizeMin   map[string]int64
	contractOrderSizeRound map[string]int64
	dualModeEnsured        bool
}

func NewGate(config *Config) *Gate {
	return &Gate{
		config:                 config,
		endpoint:               "https://api.gateio.ws",
		contractMultiplier:     make(map[string]float64),
		contractOrderSizeMin:   make(map[string]int64),
		contractOrderSizeRound: make(map[string]int64),
	}
}

func (gt *Gate) GetName() string { return "gate" }

// GetContractMultiplier returns Gate futures contract multiplier (quanto_multiplier) for the given symbol.
// This is useful for converting "contracts" to "base quantity" in other subsystems (WS/order store).
func (gt *Gate) GetContractMultiplier(ctx context.Context, symbol string) (float64, error) {
	contract := gt.formatContract(symbol)
	return gt.getMultiplier(ctx, contract)
}

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
	// 限流：统一通过全局 RateLimiter 控制请求速率，避免 Gate 429/风控封禁
	_ = WaitForRateLimit(ctx, "gate")

	requestPath := "/api/v4" + path
	queryString := ""
	if len(query) > 0 {
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

// gateCheckBizError checks Gate's common "HTTP 200 but error JSON" responses.
// Typical error payload:
// {"label":"INVALID_KEY","message":"..."}
func gateCheckBizError(raw string) error {
	j := gjson.New(raw)
	label := strings.TrimSpace(j.Get("label").String())
	if label == "" {
		return nil
	}
	msg := strings.TrimSpace(j.Get("message").String())
	if msg == "" {
		msg = strings.TrimSpace(j.Get("detail").String())
	}
	if msg == "" {
		msg = raw
	}
	return gerror.Newf("[gate] api error: %s: %s", label, msg)
}

func (gt *Gate) publicRequest(ctx context.Context, path string, query url.Values) (string, error) {
	// 限流：公共接口也纳入限流（避免全局行情/多机器人同时拉取造成风控）
	_ = WaitForRateLimit(ctx, "gate")

	requestPath := "/api/v4" + path
	reqURL := gt.endpoint + requestPath
	if len(query) > 0 {
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

// getOrderSizeRules returns (minSize, sizeStep) for Gate futures orders (in contracts).
// Best-effort: if fields are missing or request fails, fallback to (1, 1).
func (gt *Gate) getOrderSizeRules(ctx context.Context, contract string) (minSize int64, step int64) {
	gt.mu.Lock()
	mn, okMin := gt.contractOrderSizeMin[contract]
	rd, okRound := gt.contractOrderSizeRound[contract]
	gt.mu.Unlock()
	if okMin && okRound && mn > 0 && rd > 0 {
		return mn, rd
	}

	raw, err := gt.publicRequest(ctx, "/futures/usdt/contracts/"+url.PathEscape(contract), nil)
	if err != nil {
		return 1, 1
	}
	j := gjson.New(raw)

	// 常见字段：order_size_min / order_size_round
	mn = j.Get("order_size_min").Int64()
	if mn <= 0 {
		mn = 1
	}
	rd = j.Get("order_size_round").Int64()
	if rd <= 0 {
		rd = 1
	}

	gt.mu.Lock()
	gt.contractOrderSizeMin[contract] = mn
	gt.contractOrderSizeRound[contract] = rd
	gt.mu.Unlock()
	return mn, rd
}

// GetBalance 获取账户余额（USDT）
func (gt *Gate) GetBalance(ctx context.Context) (*Balance, error) {
	raw, err := gt.signedRequest(ctx, "GET", "/futures/usdt/accounts", nil, nil)
	if err != nil {
		return nil, err
	}
	if bizErr := gateCheckBizError(raw); bizErr != nil {
		return nil, bizErr
	}
	// 兼容：部分网关/SDK 可能返回数组或包一层 data/result
	j := gjson.New(raw)
	trim := strings.TrimSpace(raw)
	if strings.HasPrefix(trim, "[") {
		if arr := j.Array(); len(arr) > 0 {
			j = gjson.New(arr[0])
		}
	} else {
		// data/result/records: 可能是数组，也可能是对象（不同代理/网关/版本差异）
		if arr := j.Get("data").Array(); len(arr) > 0 {
			j = gjson.New(arr[0])
		} else if arr := j.Get("result").Array(); len(arr) > 0 {
			j = gjson.New(arr[0])
		} else if arr := j.Get("records").Array(); len(arr) > 0 {
			j = gjson.New(arr[0])
		} else if m := j.Get("data").Map(); len(m) > 0 {
			j = gjson.New(m)
		} else if m := j.Get("result").Map(); len(m) > 0 {
			j = gjson.New(m)
		}
	}
	// 常见字段：total / available / unrealised_pnl
	// 注意：Gate 的 total 往往是“钱包余额”(不含未实现盈亏)，unrealised_pnl 为未实现盈亏；
	// 列表页需要的“账户权益”口径应为 WalletBalance + UnrealizedPnl。
	total := j.Get("total").Float64()
	if total == 0 {
		// 兼容可能字段
		total = j.Get("total_wallet_balance").Float64()
	}
	if total == 0 {
		total = j.Get("wallet_balance").Float64()
	}
	avail := j.Get("available").Float64()
	if avail == 0 {
		// 兼容可能的字段名差异
		avail = j.Get("available_margin").Float64()
	}
	if avail == 0 {
		avail = j.Get("available_balance").Float64()
	}
	if avail == 0 {
		avail = j.Get("available_funds").Float64()
	}
	if avail == 0 {
		avail = j.Get("availableBalance").Float64()
	}
	upl := j.Get("unrealised_pnl").Float64()
	if upl == 0 {
		upl = j.Get("unrealized_pnl").Float64()
	}
	if upl == 0 {
		upl = j.Get("unrealizedPnl").Float64()
	}
	equity := j.Get("equity").Float64()
	if equity <= 0 {
		equity = j.Get("account_equity").Float64()
	}
	if equity <= 0 {
		equity = j.Get("accountEquity").Float64()
	}
	if equity <= 0 {
		// 兜底：用 total + upl 作为账户权益
		equity = total + upl
	}

	// 如果仍然全为0：大概率是返回结构/字段名变更或权限导致的“成功响应但无字段”
	// 这里低频输出一次 top-level keys，方便现场对照官方文档/抓包
	if equity == 0 && avail == 0 && total == 0 && upl == 0 {
		if gateShouldLog("balance_zero", 30*time.Second) {
			keys := make([]string, 0, 16)
			for k := range j.Map() {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			preview := raw
			if len(preview) > 400 {
				preview = preview[:400]
			}
			g.Log().Warningf(ctx, "[gate][GetBalance] parsed all zeros; topKeys=%v rawPreview=%s", keys, preview)
		}
		// 不返回“全0成功结果”，否则上层会缓存 0.00 导致页面长期不刷新
		return nil, gerror.New("[gate] GetBalance parsed all zeros (invalid response or permission issue)")
	}

	return &Balance{
		TotalBalance:     equity,
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
	mark := j.Get("mark_price").Float64()
	if mark <= 0 {
		mark = j.Get("markPrice").Float64()
	}
	// Gate futures tickers 常见字段 change_percentage（可能为 string/number）
	changePct := j.Get("change_percentage").Float64()
	if changePct == 0 {
		changePct = j.Get("changePercent").Float64()
	}
	// time 字段通常为秒级时间戳
	ts := j.Get("time").Int64()
	if ts == 0 {
		ts = j.Get("t").Int64()
	}
	if ts > 0 && ts < 1e12 {
		ts = ts * 1000
	}
	if ts <= 0 {
		ts = time.Now().UnixMilli()
	}
	return &Ticker{
		Symbol:             symbol,
		LastPrice:          last,
		MarkPrice:          mark,
		BidPrice:           bid,
		AskPrice:           ask,
		High24h:            high,
		Low24h:             low,
		Volume24h:          vol,
		Change24h:          changePct,
		PriceChangePercent: changePct,
		Timestamp:          ts,
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

	mul, mErr := gt.getMultiplier(ctx, contract)
	if mul <= 0 {
		// 不阻断：仍允许返回持仓，但会影响 qty 折算/保证金展示；尽量记录原因由上层排查
		_ = mErr
		mul = 1
	}

	var out []*Position
	for _, it := range items {
		j := gjson.New(it)
		size := j.Get("size").Float64() // 合约张数（对冲模式下可能总为正，方向在 side/pos_side 字段里）
		if size == 0 {
			continue
		}
		// 方向：优先从 side/pos_side/position_side 判断（Gate 双向模式常见 size 为正）
		positionSide := ""
		sideRaw := strings.ToLower(strings.TrimSpace(j.Get("side").String()))
		if sideRaw == "" {
			sideRaw = strings.ToLower(strings.TrimSpace(j.Get("pos_side").String()))
		}
		if sideRaw == "" {
			sideRaw = strings.ToLower(strings.TrimSpace(j.Get("position_side").String()))
		}
		switch sideRaw {
		case "long":
			positionSide = "LONG"
		case "short":
			positionSide = "SHORT"
		}
		if positionSide == "" {
			// 兜底：按符号判断
			positionSide = "LONG"
			if size < 0 {
				positionSide = "SHORT"
			}
		}

		qtyBase := absFloat(size) * mul
		entry := j.Get("entry_price").Float64()
		if entry == 0 {
			entry = j.Get("entryPrice").Float64()
		}
		mark := j.Get("mark_price").Float64()
		if mark == 0 {
			mark = j.Get("markPrice").Float64()
		}
		upl := j.Get("unrealised_pnl").Float64()
		if upl == 0 {
			upl = j.Get("unrealized_pnl").Float64()
		}
		lev := int(j.Get("leverage").Float64())
		if lev <= 0 {
			lev = int(j.Get("cross_leverage").Float64())
		}
		margin := j.Get("margin").Float64()
		if margin <= 0 {
			// 兼容字段：不同返回可能用 position_margin/initial_margin
			margin = j.Get("position_margin").Float64()
		}
		if margin <= 0 {
			margin = j.Get("initial_margin").Float64()
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
			MarginType:     "ISOLATED",
			IsolatedMargin: margin, // Gate futures/usdt 默认逐仓，这里保持一致供前端展示兜底
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
	desired := req.Quantity / mul
	minSize, step := gt.getOrderSizeRules(ctx, contract)
	if minSize <= 0 {
		minSize = 1
	}
	if step <= 0 {
		step = 1
	}
	if desired <= 0 {
		return nil, gerror.New("Gate 下单数量无效（desired contracts <= 0）")
	}

	// 关键优化：
	// - 开仓（reduceOnly=false）：向上取整，避免“向下取整导致实际下单数量过小/长期偏离”
	// - 平仓（reduceOnly=true）：优先向下取整；若小于最小张数则兜底为最小张数，尽量避免“无法平仓/留残仓”
	var contracts int64
	if req.ReduceOnly {
		contracts = int64(desired) // floor
		if step > 1 && contracts > 0 {
			contracts = (contracts / step) * step
		}
		if contracts < minSize {
			contracts = minSize
		}
	} else {
		contracts = int64(math.Ceil(desired))
		if step > 1 && contracts > 0 {
			contracts = ((contracts + step - 1) / step) * step // round up to step
		}
		if contracts < minSize {
			contracts = minSize
		}
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

	// size/left are in contracts; convert to base qty using multiplier for consistency with other exchanges.
	sizeContracts := j.Get("size").Float64()
	leftContracts := j.Get("left").Float64()
	qtyContractsAbs := absFloat(sizeContracts)
	filledContracts := 0.0
	if qtyContractsAbs > 0 {
		filledContracts = qtyContractsAbs - absFloat(leftContracts)
		if filledContracts < 0 {
			filledContracts = 0
		}
	}
	qtyBase := qtyContractsAbs * mul
	filledBase := filledContracts * mul

	return &Order{
		OrderId:      j.Get("id").String(),
		ClientId:     j.Get("text").String(),
		Symbol:       req.Symbol,
		Side:         strings.ToUpper(req.Side),
		PositionSide: strings.ToUpper(req.PositionSide),
		Type:         strings.ToUpper(req.Type),
		Price:        j.Get("price").Float64(),
		// Quantity/FilledQty unified as base coin quantity.
		Quantity:   qtyBase,
		FilledQty:  filledBase,
		AvgPrice:   j.Get("fill_price").Float64(),
		Status:     j.Get("status").String(),
		CreateTime: time.Now().UnixMilli(),
		UpdateTime: time.Now().UnixMilli(),
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
	// Gate 文档/实际返回对 leverage 参数比较严格：
	// - 必须传 leverage（否则返回 MISSING_REQUIRED_PARAM）
	// - 部分环境下要求 query 与 body 都带（以确保服务端能识别）
	// - 值应为数值类型（非字符串）
	q := url.Values{}
	q.Set("leverage", strconv.Itoa(leverage))
	body := map[string]any{
		"leverage": leverage,
	}
	_, err := gt.signedRequest(ctx, "POST", "/futures/usdt/positions/"+url.PathEscape(contract)+"/leverage", q, body)
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
	mul, _ := gt.getMultiplier(ctx, contract)
	if mul <= 0 {
		mul = 1
	}
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
		sizeContracts := j.Get("size").Float64()
		leftContracts := j.Get("left").Float64()
		side := "BUY"
		if sizeContracts < 0 {
			side = "SELL"
		}
		typ := "LIMIT"
		if j.Get("price").String() == "0" {
			typ = "MARKET"
		}
		qtyContractsAbs := absFloat(sizeContracts)
		filledContracts := 0.0
		if qtyContractsAbs > 0 {
			filledContracts = qtyContractsAbs - absFloat(leftContracts)
			if filledContracts < 0 {
				filledContracts = 0
			}
		}
		// pos_side: long/short (optional)
		posSide := strings.ToUpper(strings.TrimSpace(j.Get("pos_side").String()))
		if posSide == "" {
			posSide = strings.ToUpper(strings.TrimSpace(j.Get("position_side").String()))
		}
		if posSide == "LONG" || posSide == "SHORT" {
			// ok
		} else if strings.EqualFold(posSide, "long") {
			posSide = "LONG"
		} else if strings.EqualFold(posSide, "short") {
			posSide = "SHORT"
		} else {
			posSide = ""
		}
		reduceOnly := j.Get("reduce_only").Bool()
		out = append(out, &Order{
			OrderId:      j.Get("id").String(),
			ClientId:     j.Get("text").String(),
			Symbol:       symbol,
			Side:         side,
			PositionSide: posSide,
			Type:         typ,
			ReduceOnly:   reduceOnly,
			Price:        j.Get("price").Float64(),
			Quantity:     qtyContractsAbs * mul,
			FilledQty:    filledContracts * mul,
			AvgPrice:     j.Get("fill_price").Float64(),
			Status:       j.Get("status").String(),
			CreateTime:   j.Get("create_time").Int64() * 1000,
			UpdateTime:   j.Get("finish_time").Int64() * 1000,
		})
	}
	return out, nil
}

func (gt *Gate) GetOrderHistory(ctx context.Context, symbol string, limit int) ([]*Order, error) {
	contract := gt.formatContract(symbol)
	mul, _ := gt.getMultiplier(ctx, contract)
	if mul <= 0 {
		mul = 1
	}
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
		sizeContracts := j.Get("size").Float64()
		leftContracts := j.Get("left").Float64()
		side := "BUY"
		if sizeContracts < 0 {
			side = "SELL"
		}
		typ := "LIMIT"
		if j.Get("price").String() == "0" {
			typ = "MARKET"
		}
		qtyContractsAbs := absFloat(sizeContracts)
		filledContracts := 0.0
		if qtyContractsAbs > 0 {
			filledContracts = qtyContractsAbs - absFloat(leftContracts)
			if filledContracts < 0 {
				filledContracts = 0
			}
		}
		posSide := strings.ToUpper(strings.TrimSpace(j.Get("pos_side").String()))
		if posSide == "" {
			posSide = strings.ToUpper(strings.TrimSpace(j.Get("position_side").String()))
		}
		if strings.EqualFold(posSide, "long") {
			posSide = "LONG"
		} else if strings.EqualFold(posSide, "short") {
			posSide = "SHORT"
		} else if posSide != "LONG" && posSide != "SHORT" {
			posSide = ""
		}
		reduceOnly := j.Get("reduce_only").Bool()
		out = append(out, &Order{
			OrderId:      j.Get("id").String(),
			ClientId:     j.Get("text").String(),
			Symbol:       symbol,
			Side:         side,
			PositionSide: posSide,
			Type:         typ,
			ReduceOnly:   reduceOnly,
			Price:        j.Get("price").Float64(),
			Quantity:     qtyContractsAbs * mul,
			FilledQty:    filledContracts * mul,
			AvgPrice:     j.Get("fill_price").Float64(),
			Status:       j.Get("status").String(),
			CreateTime:   j.Get("create_time").Int64() * 1000,
			UpdateTime:   j.Get("finish_time").Int64() * 1000,
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
	if bizErr := gateCheckBizError(raw); bizErr != nil {
		return nil, bizErr
	}

	mul, _ := gt.getMultiplier(ctx, contract)
	if mul <= 0 {
		mul = 1
	}

	var out []*Trade
	j := gjson.New(raw)
	items := j.Array()
	if len(items) == 0 {
		// 兼容包一层 data/result
		if arr := j.Get("data").Array(); len(arr) > 0 {
			items = arr
		} else if arr := j.Get("result").Array(); len(arr) > 0 {
			items = arr
		} else if arr := j.Get("records").Array(); len(arr) > 0 {
			items = arr
		}
	}
	for _, it := range items {
		j := gjson.New(it)
		size := j.Get("size").Float64() // 合约张数，正/负表示方向（常见约定）
		closeSize := j.Get("close_size").Float64() // 平仓张数（Gate: dual 模式下常见；正/负可用于推断平仓方向）
		side := strings.ToUpper(j.Get("side").String())
		// normalize side (buy/sell -> BUY/SELL)
		if strings.EqualFold(side, "buy") {
			side = "BUY"
		} else if strings.EqualFold(side, "sell") {
			side = "SELL"
		}
		if side == "" {
			if size < 0 {
				side = "SELL"
			} else {
				side = "BUY"
			}
		}
		// Gate 成交明细的持仓方向字段在不同模式/版本下可能不同：
		// - pos_side / position_side（常见，dual mode 时通常有 long/short）
		// - posSide / positionSide（部分实现为 camelCase）
		// - auto_size（下单接口字段，可能在某些回传结构里出现：close_long/close_short）
		// avg-cost fallback 需要 LONG/SHORT，否则会跳过回填。
		posSide := strings.ToUpper(strings.TrimSpace(j.Get("pos_side").String()))
		if posSide == "" {
			posSide = strings.ToUpper(strings.TrimSpace(j.Get("position_side").String()))
		}
		if posSide == "" {
			posSide = strings.ToUpper(strings.TrimSpace(j.Get("posSide").String()))
		}
		if posSide == "" {
			posSide = strings.ToUpper(strings.TrimSpace(j.Get("positionSide").String()))
		}
		if posSide == "" {
			autoSize := strings.ToUpper(strings.TrimSpace(j.Get("auto_size").String()))
			// e.g. close_long / close_short
			if strings.Contains(autoSize, "LONG") {
				posSide = "LONG"
			} else if strings.Contains(autoSize, "SHORT") {
				posSide = "SHORT"
			}
		}
		// normalize posSide (Gate commonly returns "long"/"short")
		if strings.EqualFold(posSide, "long") {
			posSide = "LONG"
		} else if strings.EqualFold(posSide, "short") {
			posSide = "SHORT"
		}
		// Gate dual 模式下部分记录不返回 pos_side，但会返回 close_size：
		// - 平仓 LONG：通常为 SELL，且 close_size 为负
		// - 平仓 SHORT：通常为 BUY，且 close_size 为正
		// 这是“账本兜底/avg-cost 回填”的必要条件之一，否则上层会因为缺少 PositionSide 过滤掉候选成交。
		if strings.TrimSpace(posSide) == "" && closeSize != 0 {
			if closeSize < 0 {
				posSide = "LONG"
			} else if closeSize > 0 {
				posSide = "SHORT"
			}
		}
		// 没有明确 posSide 时留空（上层/回填逻辑会跳过；我们会低频告警提示）
		orderID := j.Get("order_id").String()
		if orderID == "" {
			orderID = j.Get("orderId").String()
		}

		fee := j.Get("fee").Float64()
		if fee == 0 {
			fee = j.Get("commission").Float64()
		}
		feeCcy := j.Get("fee_currency").String()
		if feeCcy == "" {
			feeCcy = j.Get("feeCcy").String()
		}
		if feeCcy == "" {
			feeCcy = j.Get("commission_asset").String()
		}
		if feeCcy == "" {
			feeCcy = "USDT"
		}

		ts := j.Get("create_time").Int64()
		if ts > 0 && ts < 1e12 {
			ts = ts * 1000
		}

		out = append(out, &Trade{
			TradeId:      j.Get("id").String(),
			OrderId:      orderID,
			Symbol:       symbol,
			Side:         side,
			PositionSide: posSide,
			Price:        j.Get("price").Float64(),
			Quantity:     absFloat(size) * mul, // 折算成基础币数量
			RealizedPnl: func() float64 {
				if v := j.Get("pnl").Float64(); v != 0 {
					return v
				}
				// 兼容字段：部分实现使用 fill_pnl / realizedPnl
				if v := j.Get("fill_pnl").Float64(); v != 0 {
					return v
				}
				if v := j.Get("realizedPnl").Float64(); v != 0 {
					return v
				}
				if v := j.Get("realized_pnl").Float64(); v != 0 {
					return v
				}
				if v := j.Get("realised_pnl").Float64(); v != 0 {
					return v
				}
				return j.Get("profit").Float64()
			}(),
			Commission:      absFloat(fee),
			CommissionAsset: feeCcy,
			Time:            ts,
		})
	}

	// Gate 兜底：my_trades 经常不返回 pnl（或部分记录为0），但资金流水(account_book)里会有“盈亏/手续费”。
	// 为了让上层（运行区间同步/成交流水落库/交易明细）尽量与 Binance/OKX 统一，这里在 exchange 层做一次“账本优先”的回填：
	// - 只填补 RealizedPnl==0 / Commission==0 的记录，不覆盖交易所返回的非0值
	// - 优先按 order_id 精确匹配 account_book（若响应不带 order_id，则仅对 pnl 做“秒级弱匹配”，避免误归因）
	if len(out) > 0 {
		needPnl := 0
		needFee := 0
		var minNeedTs, maxNeedTs int64
		for _, t := range out {
			if t == nil {
				continue
			}
			if strings.TrimSpace(t.OrderId) == "" {
				continue
			}
			if t.RealizedPnl == 0 {
				needPnl++
			}
			if t.Commission == 0 {
				needFee++
			}
			if (t.RealizedPnl == 0 || t.Commission == 0) && t.Time > 0 {
				if minNeedTs == 0 || t.Time < minNeedTs {
					minNeedTs = t.Time
				}
				if t.Time > maxNeedTs {
					maxNeedTs = t.Time
				}
			}
		}
		if (needPnl > 0 || needFee > 0) && minNeedTs > 0 && maxNeedTs >= minNeedTs {
			// 账本写入有延迟：给一个可控窗口（默认: min-10m ~ max+10m；但最多不超过 6 小时）
			fromMs := minNeedTs - 10*60*1000
			toMs := maxNeedTs + 10*60*1000
			if toMs-fromMs > 6*60*60*1000 {
				fromMs = maxNeedTs - 6*60*60*1000
			}
			books, err := gt.GetAccountBook(ctx, symbol, fromMs, toMs, 500)
			if err == nil && len(books) > 0 {
				pnlByOrder := make(map[string]float64)
				feeByOrder := make(map[string]float64)
				feeCcyByOrder := make(map[string]string)
				hasOrderId := false
				for _, b := range books {
					if b == nil {
						continue
					}
					oid := strings.TrimSpace(b.OrderId)
					if oid == "" {
						continue
					}
					hasOrderId = true
					typ := strings.ToLower(strings.TrimSpace(b.Type))
					switch {
					case strings.Contains(typ, "pnl") || strings.Contains(typ, "profit") || strings.Contains(typ, "loss"):
						pnlByOrder[oid] += b.Change
					case strings.Contains(typ, "fee"):
						feeByOrder[oid] += b.Change // fee 通常为负数
						if feeCcyByOrder[oid] == "" && strings.TrimSpace(b.Currency) != "" {
							feeCcyByOrder[oid] = strings.TrimSpace(b.Currency)
					}
				}
				}

				if hasOrderId && (len(pnlByOrder) > 0 || len(feeByOrder) > 0) {
					// orderId -> totalQty, for proportional allocation across multiple fills
					totalQty := make(map[string]float64)
					for _, t := range out {
						if t == nil {
							continue
						}
						oid := strings.TrimSpace(t.OrderId)
						if oid == "" || t.Quantity <= 0 {
							continue
						}
						if pnlByOrder[oid] != 0 || feeByOrder[oid] != 0 {
						totalQty[oid] += t.Quantity
						}
					}
					for _, t := range out {
						if t == nil || t.Quantity <= 0 {
							continue
						}
						oid := strings.TrimSpace(t.OrderId)
						q := totalQty[oid]
						if oid == "" || q <= 0 {
							continue
						}
						// fill pnl if missing
						if t.RealizedPnl == 0 {
							if pnl := pnlByOrder[oid]; pnl != 0 {
							t.RealizedPnl = pnl * (t.Quantity / q)
						}
					}
						// fill fee if missing
						if t.Commission == 0 {
							if fee := feeByOrder[oid]; fee != 0 {
								t.Commission = math.Abs(fee) * (t.Quantity / q)
								if strings.TrimSpace(t.CommissionAsset) == "" {
									ccy := strings.TrimSpace(feeCcyByOrder[oid])
									if ccy == "" {
										ccy = "USDT"
									}
									t.CommissionAsset = ccy
								}
							}
						}
					}
				} else if needPnl > 0 {
					// 账本不带 order_id：尝试“秒级时间 + 合约”弱匹配（仅在同一秒内只出现一个订单ID时才归因，避免误配）
					pnlBySec := make(map[int64]float64) // sec -> pnlSum
					for _, b := range books {
						if b == nil || b.Time <= 0 {
							continue
						}
						typ := strings.ToLower(strings.TrimSpace(b.Type))
						if !(strings.Contains(typ, "pnl") || strings.Contains(typ, "profit") || strings.Contains(typ, "loss")) {
							continue
						}
						sec := b.Time / 1000
						pnlBySec[sec] += b.Change
					}
					if len(pnlBySec) > 0 {
						// group fills by sec
						type secBucket struct {
							sumQty   float64
							orderIDs map[string]bool
						}
						buckets := make(map[int64]*secBucket)
						for _, t := range out {
							if t == nil || t.Time <= 0 || t.Quantity <= 0 {
								continue
							}
							sec := t.Time / 1000
							b := buckets[sec]
							if b == nil {
								b = &secBucket{orderIDs: map[string]bool{}}
								buckets[sec] = b
							}
							b.sumQty += t.Quantity
							oid := strings.TrimSpace(t.OrderId)
							if oid != "" {
								b.orderIDs[oid] = true
							}
						}
						applied := 0
						skippedMulti := 0
						for _, t := range out {
							if t == nil || t.RealizedPnl != 0 || t.Time <= 0 || t.Quantity <= 0 {
								continue
							}
							sec := t.Time / 1000
							pnl, ok := pnlBySec[sec]
							if !ok || pnl == 0 {
								continue
							}
							b := buckets[sec]
							if b == nil || b.sumQty <= 0 {
								continue
							}
							// safety: only attribute when this second contains exactly one order id (or empty ids)
							if len(b.orderIDs) > 1 {
								skippedMulti++
								continue
							}
							t.RealizedPnl = pnl * (t.Quantity / b.sumQty)
							applied++
						}
						if applied == 0 && gateShouldLog("account_book_no_order_id:"+strings.ToUpper(strings.TrimSpace(symbol)), 60*time.Second) {
							g.Log().Warningf(ctx, "[gate][GetAccountBook] no order_id fields found; weak match not applied (maybe multi-orders per second): symbol=%s window=[%d,%d] items=%d skippedMulti=%d",
								symbol, fromMs, toMs, len(books), skippedMulti)
						}
					} else {
						if gateShouldLog("account_book_no_order_id:"+strings.ToUpper(strings.TrimSpace(symbol)), 60*time.Second) {
							g.Log().Warningf(ctx, "[gate][GetAccountBook] no order_id fields found; skip pnl fill: symbol=%s window=[%d,%d] items=%d",
								symbol, fromMs, toMs, len(books))
						}
					}
				}
			}
		}
	}

	// 低频诊断：成交不为空但 posSide 全缺失时，avg-cost 回填会全部跳过 → realized_pnl 全为 0
	if len(out) > 0 {
		missingPos := 0
		for _, t := range out {
			if t == nil || strings.TrimSpace(t.PositionSide) == "" {
				missingPos++
			}
		}
		if missingPos == len(out) && gateShouldLog("my_trades_no_pos_side:"+strings.ToUpper(strings.TrimSpace(symbol)), 30*time.Second) {
			preview := raw
			if len(preview) > 600 {
				preview = preview[:600]
			}
			g.Log().Warningf(ctx, "[gate][GetTradeHistory] pos_side missing for all trades: symbol=%s rawPreview=%s", symbol, preview)
		}
	}

	// Fallback: Gate 的 my_trades 在部分模式/账号下不返回 pnl，使用 avg-cost 模型按方向计算已实现盈亏。
	FillRealizedPnlByAvgCost(out)

	// 低频诊断：回填后仍全为0（可能是：全部为开仓成交 / 字段缺失 / 口径不匹配）
	if len(out) > 0 {
		nonZero := 0
		for _, t := range out {
			if t != nil && t.RealizedPnl != 0 {
				nonZero++
			}
		}
		if nonZero == 0 && gateShouldLog("my_trades_pnl_all_zero:"+strings.ToUpper(strings.TrimSpace(symbol)), 60*time.Second) {
			sample := 0
			sb := strings.Builder{}
			for _, t := range out {
				if t == nil {
					continue
				}
				sb.WriteString(fmt.Sprintf("{id=%s orderId=%s side=%s posSide=%s px=%.8f qty=%.8f pnl=%.8f ts=%d} ",
					t.TradeId, t.OrderId, t.Side, t.PositionSide, t.Price, t.Quantity, t.RealizedPnl, t.Time))
				sample++
				if sample >= 3 {
					break
				}
			}
			g.Log().Warningf(ctx, "[gate][GetTradeHistory] realized pnl still all zero after avg-cost fill: symbol=%s sample=%s", symbol, sb.String())
		}
	}

	// 低频提示：my_trades 空可能是“确实无成交”或“权限不足/结构变更”
	if len(out) == 0 && gateShouldLog("my_trades_empty:"+strings.ToUpper(strings.TrimSpace(symbol)), 30*time.Second) {
		preview := raw
		if len(preview) > 400 {
			preview = preview[:400]
		}
		g.Log().Warningf(ctx, "[gate][GetTradeHistory] empty result: symbol=%s rawPreview=%s", symbol, preview)
	}

	return out, nil
}

// GetAccountBook 获取期货资金流水（用于补齐“盈亏/手续费”等账本口径数据）
// Gate v4 futures: GET /futures/usdt/account_book
//
// Notes:
// - Gate 的 my_trades 在部分账号/模式下不稳定返回 pnl；但资金流水里会有 "盈亏"/"交易手续费"
// - fromMs/toMs: 毫秒时间戳（内部会转换为秒级参数）
func (gt *Gate) GetAccountBook(ctx context.Context, symbol string, fromMs, toMs int64, limit int) ([]*AccountBookItem, error) {
	contract := gt.formatContract(symbol)
	q := url.Values{}
	if contract != "" {
		q.Set("contract", contract)
	}
	// Gate docs typically uses seconds for from/to
	if fromMs > 0 {
		q.Set("from", strconv.FormatInt(fromMs/1000, 10))
	}
	if toMs > 0 {
		q.Set("to", strconv.FormatInt(toMs/1000, 10))
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	raw, err := gt.signedRequest(ctx, "GET", "/futures/usdt/account_book", q, nil)
	if err != nil {
		return nil, err
	}
	if bizErr := gateCheckBizError(raw); bizErr != nil {
		return nil, bizErr
	}

	j := gjson.New(raw)
	items := j.Array()
	if len(items) == 0 {
		// 兼容包一层 data/result/records
		if arr := j.Get("data").Array(); len(arr) > 0 {
			items = arr
		} else if arr := j.Get("result").Array(); len(arr) > 0 {
			items = arr
		} else if arr := j.Get("records").Array(); len(arr) > 0 {
			items = arr
		}
	}

	out := make([]*AccountBookItem, 0, len(items))
	for _, it := range items {
		x := gjson.New(it)
		ts := x.Get("time").Int64()
		if ts == 0 {
			ts = x.Get("create_time").Int64()
		}
		// normalize to ms
		if ts > 0 && ts < 1e12 {
			ts *= 1000
		}
		orderID := x.Get("order_id").String()
		if orderID == "" {
			orderID = x.Get("orderId").String()
		}
		if orderID == "" {
			orderID = x.Get("order").String()
		}
		typ := x.Get("type").String()
		if typ == "" {
			typ = x.Get("reason").String()
		}
		change := x.Get("change").Float64()
		if change == 0 {
			change = x.Get("amount").Float64()
		}
		ccy := x.Get("currency").String()
		if ccy == "" {
			ccy = x.Get("fee_currency").String()
		}
		if ccy == "" {
			ccy = "USDT"
		}
		ct := x.Get("contract").String()
		text := x.Get("text").String()
		if text == "" {
			text = x.Get("remark").String()
		}
		if orderID == "" && text != "" {
			orderID = gateExtractOrderIDFromText(text)
		}

		out = append(out, &AccountBookItem{
			Time:     ts,
			Type:     typ,
			Change:   change,
			Currency: ccy,
			Symbol:   symbol,
			Contract: ct,
			OrderId:  orderID,
			Text:     text,
		})
	}

	// 低频诊断：打印 account_book 的字段结构（便于现场确认 Gate 返回是否包含 order_id/是否在 text 里）
	if len(items) > 0 && gateShouldLog("account_book_schema:"+strings.ToUpper(strings.TrimSpace(symbol)), 60*time.Second) {
		x := gjson.New(items[0])
		keys := make([]string, 0)
		if m := x.Map(); len(m) > 0 {
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)
		}
		preview := x.String()
		if len(preview) > 600 {
			preview = preview[:600]
		}
		g.Log().Warningf(ctx, "[gate][GetAccountBook] schema: symbol=%s keys=%v firstItemPreview=%s", symbol, keys, preview)
	}

	// 低频提示：返回为空可能是“确实无记录”或“参数/权限不对”
	if len(out) == 0 && gateShouldLog("account_book_empty:"+strings.ToUpper(strings.TrimSpace(symbol)), 60*time.Second) {
		preview := raw
		if len(preview) > 400 {
			preview = preview[:400]
		}
		g.Log().Warningf(ctx, "[gate][GetAccountBook] empty result: symbol=%s rawPreview=%s", symbol, preview)
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
