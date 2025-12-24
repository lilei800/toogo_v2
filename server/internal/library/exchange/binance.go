// Package exchange Binance交易所API
package exchange

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

// Binance Binance交易所
type Binance struct {
	config   *Config
	endpoint string

	// 运行期缓存：避免每次下单都调用“设置逐仓/双向持仓模式”接口
	mu                  sync.Mutex
	hedgeModeEnsured    bool
	isolatedEnsuredBySy map[string]bool // key: formatted symbol
}

// NewBinance 创建Binance实例
func NewBinance(config *Config) *Binance {
	endpoint := "https://fapi.binance.com"
	if config.IsTestnet {
		endpoint = "https://testnet.binancefuture.com"
	}
	return &Binance{
		config:              config,
		endpoint:            endpoint,
		isolatedEnsuredBySy: make(map[string]bool),
	}
}

func (b *Binance) GetName() string {
	return "binance"
}

// ensureHedgeMode 确保账户为双向持仓模式（Hedge Mode）
// Binance 合约的 positionSide 参数只有在双向持仓模式下才生效。
func (b *Binance) ensureHedgeMode(ctx context.Context) {
	b.mu.Lock()
	if b.hedgeModeEnsured {
		b.mu.Unlock()
		return
	}
	b.mu.Unlock()

	// POST /fapi/v1/positionSide/dual?dualSidePosition=true
	_, err := b.signedRequest(ctx, "POST", "/fapi/v1/positionSide/dual", map[string]string{
		"dualSidePosition": "true",
	})
	if err != nil {
		// 容错：如果已是目标模式，接口可能返回“无需修改”等错误，这里不阻断
		msg := err.Error()
		if strings.Contains(msg, "No need to change") || strings.Contains(msg, "not modified") {
			err = nil
		}
	}

	b.mu.Lock()
	// 不论是否成功，都只尝试一次，避免高频重试（失败会在后续下单报错暴露）
	b.hedgeModeEnsured = true
	b.mu.Unlock()
}

// ensureIsolatedMargin 确保该交易对为逐仓（ISOLATED）
func (b *Binance) ensureIsolatedMargin(ctx context.Context, symbol string) {
	sym := b.formatSymbol(symbol)

	b.mu.Lock()
	if b.isolatedEnsuredBySy[sym] {
		b.mu.Unlock()
		return
	}
	b.mu.Unlock()

	_, err := b.signedRequest(ctx, "POST", "/fapi/v1/marginType", map[string]string{
		"symbol":     sym,
		"marginType": "ISOLATED",
	})
	if err != nil {
		// 容错：已是逐仓时可能返回“无需修改”
		msg := err.Error()
		if strings.Contains(msg, "No need to change") || strings.Contains(msg, "margin type") {
			err = nil
		}
	}

	b.mu.Lock()
	b.isolatedEnsuredBySy[sym] = true
	b.mu.Unlock()
}

// GetBalance 获取账户余额
func (b *Binance) GetBalance(ctx context.Context) (*Balance, error) {
	resp, err := b.signedRequest(ctx, "GET", "/fapi/v2/balance", nil)
	if err != nil {
		return nil, err
	}

	json := gjson.New(resp)
	balances := json.Array()
	for _, item := range balances {
		j := gjson.New(item)
		asset := j.Get("asset").String()
		if asset == "USDT" {
			return &Balance{
				TotalBalance:     j.Get("balance").Float64(),
				AvailableBalance: j.Get("availableBalance").Float64(),
				UnrealizedPnl:    j.Get("crossUnPnl").Float64(),
				Currency:         "USDT",
			}, nil
		}
	}
	return &Balance{Currency: "USDT"}, nil
}

// GetTicker 获取行情
func (b *Binance) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	params := map[string]string{"symbol": b.formatSymbol(symbol)}
	resp, err := b.publicRequest(ctx, "GET", "/fapi/v1/ticker/24hr", params)
	if err != nil {
		return nil, err
	}

	j := gjson.New(resp)
	priceChangePercent := j.Get("priceChangePercent").Float64()
	return &Ticker{
		Symbol:             symbol,
		LastPrice:          j.Get("lastPrice").Float64(),
		BidPrice:           j.Get("bidPrice").Float64(),
		AskPrice:           j.Get("askPrice").Float64(),
		High24h:            j.Get("highPrice").Float64(),
		Low24h:             j.Get("lowPrice").Float64(),
		Volume24h:          j.Get("volume").Float64(),
		Change24h:          priceChangePercent,
		PriceChangePercent: priceChangePercent,
		Timestamp:          j.Get("closeTime").Int64(),
	}, nil
}

// GetKlines 获取K线数据
func (b *Binance) GetKlines(ctx context.Context, symbol, interval string, limit int) ([]*Kline, error) {
	params := map[string]string{
		"symbol":   b.formatSymbol(symbol),
		"interval": interval,
		"limit":    strconv.Itoa(limit),
	}
	resp, err := b.publicRequest(ctx, "GET", "/fapi/v1/klines", params)
	if err != nil {
		return nil, err
	}

	var klines []*Kline
	items := gjson.New(resp).Array()
	for _, item := range items {
		arr := gjson.New(item).Array()
		if len(arr) >= 7 {
			klines = append(klines, &Kline{
				OpenTime:  g.NewVar(arr[0]).Int64(),
				Open:      g.NewVar(arr[1]).Float64(),
				High:      g.NewVar(arr[2]).Float64(),
				Low:       g.NewVar(arr[3]).Float64(),
				Close:     g.NewVar(arr[4]).Float64(),
				Volume:    g.NewVar(arr[5]).Float64(),
				CloseTime: g.NewVar(arr[6]).Int64(),
			})
		}
	}
	return klines, nil
}

// GetPositions 获取持仓
func (b *Binance) GetPositions(ctx context.Context, symbol string) ([]*Position, error) {
	resp, err := b.signedRequest(ctx, "GET", "/fapi/v2/positionRisk", nil)
	if err != nil {
		return nil, err
	}

	var positions []*Position
	items := gjson.New(resp).Array()
	for _, item := range items {
		j := gjson.New(item)
		sym := j.Get("symbol").String()
		if symbol != "" && sym != b.formatSymbol(symbol) {
			continue
		}
		posAmt := j.Get("positionAmt").Float64()
		if posAmt == 0 {
			continue
		}
		positions = append(positions, &Position{
			Symbol:           sym,
			PositionSide:     j.Get("positionSide").String(),
			PositionAmt:      posAmt,
			EntryPrice:       j.Get("entryPrice").Float64(),
			MarkPrice:        j.Get("markPrice").Float64(),
			UnrealizedPnl:    j.Get("unRealizedProfit").Float64(),
			Leverage:         j.Get("leverage").Int(),
			MarginType:       j.Get("marginType").String(),
			IsolatedMargin:   j.Get("isolatedMargin").Float64(),
			LiquidationPrice: j.Get("liquidationPrice").Float64(),
		})
	}
	return positions, nil
}

// CreateOrder 创建订单
func (b *Binance) CreateOrder(ctx context.Context, req *OrderRequest) (*Order, error) {
	// 逐仓 + 双向持仓：在首次下单前尽量确保一次，降低误平仓/positionSide无效概率
	b.ensureHedgeMode(ctx)
	b.ensureIsolatedMargin(ctx, req.Symbol)

	params := map[string]string{
		"symbol":       b.formatSymbol(req.Symbol),
		"side":         req.Side,
		"positionSide": req.PositionSide,
		"type":         req.Type,
		"quantity":     strconv.FormatFloat(req.Quantity, 'f', -1, 64),
	}

	if req.Price > 0 {
		params["price"] = strconv.FormatFloat(req.Price, 'f', -1, 64)
		params["timeInForce"] = "GTC"
	}

	// 只减仓：用于平仓，避免反向开仓导致“越开越大/误操作”
	if req.ReduceOnly {
		params["reduceOnly"] = "true"
	}

	resp, err := b.signedRequest(ctx, "POST", "/fapi/v1/order", params)
	if err != nil {
		return nil, err
	}

	j := gjson.New(resp)
	return &Order{
		OrderId:      j.Get("orderId").String(),
		ClientId:     j.Get("clientOrderId").String(),
		Symbol:       req.Symbol,
		Side:         j.Get("side").String(),
		PositionSide: j.Get("positionSide").String(),
		Type:         j.Get("type").String(),
		Price:        j.Get("price").Float64(),
		Quantity:     j.Get("origQty").Float64(),
		FilledQty:    j.Get("executedQty").Float64(),
		AvgPrice:     j.Get("avgPrice").Float64(),
		Status:       j.Get("status").String(),
		CreateTime:   j.Get("updateTime").Int64(),
	}, nil
}

// CancelOrder 取消订单
func (b *Binance) CancelOrder(ctx context.Context, symbol, orderId string) (*Order, error) {
	params := map[string]string{
		"symbol":  b.formatSymbol(symbol),
		"orderId": orderId,
	}

	resp, err := b.signedRequest(ctx, "DELETE", "/fapi/v1/order", params)
	if err != nil {
		return nil, err
	}

	j := gjson.New(resp)
	return &Order{
		OrderId: j.Get("orderId").String(),
		Symbol:  symbol,
		Status:  j.Get("status").String(),
	}, nil
}

// ClosePosition 平仓
func (b *Binance) ClosePosition(ctx context.Context, symbol, positionSide string, quantity float64) (*Order, error) {
	side := "SELL"
	if positionSide == "SHORT" {
		side = "BUY"
	}

	return b.CreateOrder(ctx, &OrderRequest{
		Symbol:       symbol,
		Side:         side,
		PositionSide: positionSide,
		Type:         "MARKET",
		Quantity:     quantity,
		ReduceOnly:   true,
	})
}

// SetLeverage 设置杠杆
func (b *Binance) SetLeverage(ctx context.Context, symbol string, leverage int) error {
	params := map[string]string{
		"symbol":   b.formatSymbol(symbol),
		"leverage": strconv.Itoa(leverage),
	}
	_, err := b.signedRequest(ctx, "POST", "/fapi/v1/leverage", params)
	return err
}

// SetMarginType 设置保证金模式
func (b *Binance) SetMarginType(ctx context.Context, symbol string, marginType string) error {
	params := map[string]string{
		"symbol":     b.formatSymbol(symbol),
		"marginType": marginType,
	}
	_, err := b.signedRequest(ctx, "POST", "/fapi/v1/marginType", params)
	return err
}

// formatSymbol 格式化交易对
func (b *Binance) formatSymbol(symbol string) string {
	// 使用统一的Symbol格式化器
	return Formatter.FormatForBinance(symbol) // BTCUSDT
}

// getHttpClient 获取HTTP客户端(支持代理)
func (b *Binance) getHttpClient() *gclient.Client {
	client := gclient.New()
	client.SetTimeout(15 * time.Second)

	// 配置代理
	if b.config.Proxy != nil && b.config.Proxy.Enabled {
		proxyAddr := b.config.Proxy.Host + ":" + strconv.Itoa(b.config.Proxy.Port)

		if b.config.Proxy.Type == "socks5" {
			// SOCKS5代理 - 使用http代理格式
			proxyURL := "socks5://" + proxyAddr
			client.SetProxy(proxyURL)
		} else {
			// HTTP代理
			proxyURL := b.config.Proxy.GetProxyURL()
			client.SetProxy(proxyURL)
		}
	}

	return client
}

// publicRequest 公开请求
func (b *Binance) publicRequest(ctx context.Context, method, path string, params map[string]string) (string, error) {
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
	if resp.StatusCode != 200 {
		return "", WrapAsAPIError("binance", resp.StatusCode, body, nil)
	}
	return body, nil
}

// signedRequest 签名请求
func (b *Binance) signedRequest(ctx context.Context, method, path string, params map[string]string) (string, error) {
	if params == nil {
		params = make(map[string]string)
	}
	params["timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	params["recvWindow"] = "5000"

	query := b.buildQuery(params)
	signature := b.sign(query)
	query += "&signature=" + signature

	client := b.getHttpClient()
	client.SetHeader("X-MBX-APIKEY", b.config.ApiKey)

	reqUrl := b.endpoint + path + "?" + query

	var resp *gclient.Response
	var err error
	switch method {
	case "GET":
		resp, err = client.Get(ctx, reqUrl)
	case "POST":
		resp, err = client.Post(ctx, reqUrl)
	case "DELETE":
		resp, err = client.Delete(ctx, reqUrl)
	default:
		resp, err = client.Get(ctx, reqUrl)
	}

	if err != nil {
		return "", gerror.Wrap(err, "Request failed")
	}
	defer resp.Close()

	body := resp.ReadAllString()
	if resp.StatusCode != 200 {
		return "", WrapAsAPIError("binance", resp.StatusCode, body, nil)
	}
	return body, nil
}

// buildQuery 构建查询字符串
func (b *Binance) buildQuery(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(params[k]))
	}
	return strings.Join(parts, "&")
}

// sign 签名
func (b *Binance) sign(data string) string {
	h := hmac.New(sha256.New, []byte(b.config.SecretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// GetOpenOrders 获取当前挂单
func (b *Binance) GetOpenOrders(ctx context.Context, symbol string) ([]*Order, error) {
	params := make(map[string]string)
	if symbol != "" {
		params["symbol"] = b.formatSymbol(symbol)
	}

	resp, err := b.signedRequest(ctx, "GET", "/fapi/v1/openOrders", params)
	if err != nil {
		return nil, err
	}

	var orders []*Order
	items := gjson.New(resp).Array()
	for _, item := range items {
		j := gjson.New(item)
		orders = append(orders, &Order{
			OrderId:      j.Get("orderId").String(),
			ClientId:     j.Get("clientOrderId").String(),
			Symbol:       j.Get("symbol").String(),
			Side:         j.Get("side").String(),
			PositionSide: j.Get("positionSide").String(),
			Type:         j.Get("type").String(),
			Price:        j.Get("price").Float64(),
			Quantity:     j.Get("origQty").Float64(),
			FilledQty:    j.Get("executedQty").Float64(),
			AvgPrice:     j.Get("avgPrice").Float64(),
			Status:       j.Get("status").String(),
			CreateTime:   j.Get("time").Int64(),
			UpdateTime:   j.Get("updateTime").Int64(),
		})
	}
	return orders, nil
}

// GetOrderHistory 获取历史订单
func (b *Binance) GetOrderHistory(ctx context.Context, symbol string, limit int) ([]*Order, error) {
	params := map[string]string{
		"limit": strconv.Itoa(limit),
	}
	if symbol != "" {
		params["symbol"] = b.formatSymbol(symbol)
	}

	resp, err := b.signedRequest(ctx, "GET", "/fapi/v1/allOrders", params)
	if err != nil {
		return nil, err
	}

	var orders []*Order
	items := gjson.New(resp).Array()
	for _, item := range items {
		j := gjson.New(item)
		orders = append(orders, &Order{
			OrderId:      j.Get("orderId").String(),
			ClientId:     j.Get("clientOrderId").String(),
			Symbol:       j.Get("symbol").String(),
			Side:         j.Get("side").String(),
			PositionSide: j.Get("positionSide").String(),
			Type:         j.Get("type").String(),
			Price:        j.Get("price").Float64(),
			Quantity:     j.Get("origQty").Float64(),
			FilledQty:    j.Get("executedQty").Float64(),
			AvgPrice:     j.Get("avgPrice").Float64(),
			Status:       j.Get("status").String(),
			CreateTime:   j.Get("time").Int64(),
			UpdateTime:   j.Get("updateTime").Int64(),
		})
	}
	return orders, nil
}

// GetTradeHistory 获取成交记录（用于财务对账/已实现盈亏/手续费汇总）
// Binance USDT 永续：GET /fapi/v1/userTrades
// 注意：该接口返回的是“成交(fill)”级别数据，包含 realizedPnl 与 commission。
func (b *Binance) GetTradeHistory(ctx context.Context, symbol string, limit int) ([]*Trade, error) {
	params := map[string]string{}
	if symbol != "" {
		params["symbol"] = b.formatSymbol(symbol)
	}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	resp, err := b.signedRequest(ctx, "GET", "/fapi/v1/userTrades", params)
	if err != nil {
		return nil, err
	}

	var out []*Trade
	for _, it := range gjson.New(resp).Array() {
		j := gjson.New(it)
		out = append(out, &Trade{
			TradeId:         j.Get("id").String(),
			OrderId:         j.Get("orderId").String(),
			Symbol:          symbol,
			Side:            strings.ToUpper(j.Get("side").String()),
			PositionSide:    strings.ToUpper(j.Get("positionSide").String()),
			Price:           j.Get("price").Float64(),
			Quantity:        j.Get("qty").Float64(),
			RealizedPnl:     j.Get("realizedPnl").Float64(),
			Commission:      j.Get("commission").Float64(),
			CommissionAsset: j.Get("commissionAsset").String(),
			Time:            j.Get("time").Int64(),
		})
	}
	return out, nil
}
