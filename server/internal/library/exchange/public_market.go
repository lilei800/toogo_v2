// Package exchange 公共行情服务
// 支持多家交易所：Binance, OKX, Gate.io
// 用于获取实时报价、K线等公开数据，无需用户API Key
package exchange

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
	"github.com/gogf/gf/v2/os/gcache"
)

// 支持的交易所
const (
	PlatformBinance = "binance"
	PlatformOKX     = "okx"
	PlatformGate    = "gate"
)

// PublicMarketService 公共行情服务（多交易所）
type PublicMarketService struct {
	exchanges map[string]*PublicExchange
	proxy     *ProxyConfig
	cache     *gcache.Cache
	cacheTTL  time.Duration
	mu        sync.RWMutex
}

// PublicExchange 单个交易所的公共API
type PublicExchange struct {
	platform string
	baseURL  string
	enabled  bool
}

var (
	publicMarketService *PublicMarketService
	publicMarketOnce    sync.Once
)

// GetPublicMarketService 获取公共行情服务单例
func GetPublicMarketService() *PublicMarketService {
	publicMarketOnce.Do(func() {
		publicMarketService = newPublicMarketService()
	})
	return publicMarketService
}

// newPublicMarketService 创建公共行情服务
func newPublicMarketService() *PublicMarketService {
	pms := &PublicMarketService{
		exchanges: make(map[string]*PublicExchange),
		cache:     gcache.New(),
		cacheTTL:  1 * time.Second, // 默认1秒缓存，更实时
	}

	// 注册所有交易所
	pms.exchanges[PlatformBinance] = &PublicExchange{
		platform: PlatformBinance,
		baseURL:  "https://fapi.binance.com",
		enabled:  true,
	}
	pms.exchanges[PlatformOKX] = &PublicExchange{
		platform: PlatformOKX,
		baseURL:  "https://www.okx.com",
		enabled:  true,
	}
	pms.exchanges[PlatformGate] = &PublicExchange{
		platform: PlatformGate,
		baseURL:  "https://api.gateio.ws",
		enabled:  true,
	}

	return pms
}

// SetProxy 设置代理
func (pms *PublicMarketService) SetProxy(proxy *ProxyConfig) {
	pms.mu.Lock()
	defer pms.mu.Unlock()
	pms.proxy = proxy
}

// SetCacheTTL 设置缓存时间
func (pms *PublicMarketService) SetCacheTTL(ttl time.Duration) {
	pms.cacheTTL = ttl
}

// getHttpClient 获取HTTP客户端
func (pms *PublicMarketService) getHttpClient() *gclient.Client {
	client := gclient.New()
	client.SetTimeout(10 * time.Second)

	pms.mu.RLock()
	proxy := pms.proxy
	pms.mu.RUnlock()

	if proxy != nil && proxy.Enabled {
		proxyAddr := proxy.Host + ":" + strconv.Itoa(proxy.Port)
		if proxy.Type == "socks5" {
			client.SetProxy("socks5://" + proxyAddr)
		} else {
			client.SetProxy("http://" + proxyAddr)
		}
	}

	return client
}

// GetTicker 获取指定交易所的实时行情
func (pms *PublicMarketService) GetTicker(ctx context.Context, platform, symbol string) (*Ticker, error) {
	exchange, ok := pms.exchanges[platform]
	if !ok || !exchange.enabled {
		return nil, gerror.Newf("不支持的交易所: %s", platform)
	}

	// 缓存key
	cacheKey := "pub_ticker:" + platform + ":" + symbol

	// 尝试从缓存获取
	if cached, err := pms.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		if ticker, ok := cached.Val().(*Ticker); ok {
			return ticker, nil
		}
	}

	// 调用API获取
	var ticker *Ticker
	var err error

	switch platform {
	case PlatformBinance:
		ticker, err = pms.fetchBinanceTicker(ctx, symbol)
	case PlatformOKX:
		ticker, err = pms.fetchOKXTicker(ctx, symbol)
	case PlatformGate:
		ticker, err = pms.fetchGateTicker(ctx, symbol)
	default:
		return nil, gerror.Newf("不支持的交易所: %s", platform)
	}

	if err != nil {
		return nil, err
	}

	// 存入缓存
	_ = pms.cache.Set(ctx, cacheKey, ticker, pms.cacheTTL)

	return ticker, nil
}

// GetTickerRealtime 获取实时行情（无缓存）
func (pms *PublicMarketService) GetTickerRealtime(ctx context.Context, platform, symbol string) (*Ticker, error) {
	switch platform {
	case PlatformBinance:
		return pms.fetchBinanceTicker(ctx, symbol)
	case PlatformOKX:
		return pms.fetchOKXTicker(ctx, symbol)
	case PlatformGate:
		return pms.fetchGateTicker(ctx, symbol)
	default:
		return nil, gerror.Newf("不支持的交易所: %s", platform)
	}
}

// GetAllExchangesTicker 从所有交易所获取同一交易对的行情
func (pms *PublicMarketService) GetAllExchangesTicker(ctx context.Context, symbol string) (map[string]*Ticker, error) {
	result := make(map[string]*Ticker)
	var wg sync.WaitGroup
	var mu sync.Mutex

	platforms := []string{PlatformBinance, PlatformOKX, PlatformGate}

	for _, platform := range platforms {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			ticker, err := pms.GetTicker(ctx, p, symbol)
			if err != nil {
				g.Log().Debugf(ctx, "获取 %s %s 行情失败: %v", p, symbol, err)
				return
			}
			mu.Lock()
			result[p] = ticker
			mu.Unlock()
		}(platform)
	}

	wg.Wait()
	return result, nil
}

// GetBestPrice 获取最优价格（比较所有交易所）
func (pms *PublicMarketService) GetBestPrice(ctx context.Context, symbol string) (*BestPriceResult, error) {
	tickers, err := pms.GetAllExchangesTicker(ctx, symbol)
	if err != nil {
		return nil, err
	}

	if len(tickers) == 0 {
		return nil, gerror.New("无法获取任何交易所行情")
	}

	result := &BestPriceResult{
		Symbol:    symbol,
		Tickers:   tickers,
		Timestamp: time.Now().UnixMilli(),
	}

	// 找最低卖价和最高买价
	for platform, ticker := range tickers {
		if result.BestAsk == 0 || ticker.AskPrice < result.BestAsk {
			result.BestAsk = ticker.AskPrice
			result.BestAskExchange = platform
		}
		if ticker.BidPrice > result.BestBid {
			result.BestBid = ticker.BidPrice
			result.BestBidExchange = platform
		}
		// 使用第一个有效价格作为参考
		if result.ReferencePrice == 0 {
			result.ReferencePrice = ticker.LastPrice
		}
	}

	// 计算价差
	if result.BestAsk > 0 && result.BestBid > 0 {
		result.Spread = (result.BestAsk - result.BestBid) / result.BestBid * 100
	}

	return result, nil
}

// BestPriceResult 最优价格结果
type BestPriceResult struct {
	Symbol          string             `json:"symbol"`
	BestBid         float64            `json:"bestBid"`         // 最高买价
	BestBidExchange string             `json:"bestBidExchange"` // 最高买价交易所
	BestAsk         float64            `json:"bestAsk"`         // 最低卖价
	BestAskExchange string             `json:"bestAskExchange"` // 最低卖价交易所
	Spread          float64            `json:"spread"`          // 价差百分比
	ReferencePrice  float64            `json:"referencePrice"`  // 参考价格
	Tickers         map[string]*Ticker `json:"tickers"`         // 各交易所行情
	Timestamp       int64              `json:"timestamp"`
}

// ========== Binance ==========

func (pms *PublicMarketService) fetchBinanceTicker(ctx context.Context, symbol string) (*Ticker, error) {
	client := pms.getHttpClient()
	url := "https://fapi.binance.com/fapi/v1/ticker/24hr"

	resp, err := client.Get(ctx, url, g.Map{
		"symbol": formatBinanceSymbol(symbol),
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "Binance请求失败")
	}
	defer resp.Close()

	j := gjson.New(resp.ReadAllString())
	if j.Get("code").Int() != 0 && j.Get("code").String() != "" {
		return nil, gerror.Newf("Binance API error: %s", j.Get("msg").String())
	}

	return &Ticker{
		Symbol:             symbol,
		LastPrice:          j.Get("lastPrice").Float64(),
		BidPrice:           j.Get("bidPrice").Float64(),
		AskPrice:           j.Get("askPrice").Float64(),
		High24h:            j.Get("highPrice").Float64(),
		Low24h:             j.Get("lowPrice").Float64(),
		Volume24h:          j.Get("volume").Float64(),
		QuoteVolume24h:     j.Get("quoteVolume").Float64(),
		Change24h:          j.Get("priceChangePercent").Float64(),
		PriceChangePercent: j.Get("priceChangePercent").Float64(),
		Timestamp:          j.Get("closeTime").Int64(),
	}, nil
}

func (pms *PublicMarketService) fetchBinanceKlines(ctx context.Context, symbol, interval string, limit int) ([]*Kline, error) {
	client := pms.getHttpClient()
	url := "https://fapi.binance.com/fapi/v1/klines"

	resp, err := client.Get(ctx, url, g.Map{
		"symbol":   formatBinanceSymbol(symbol),
		"interval": interval,
		"limit":    limit,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "Binance K线请求失败")
	}
	defer resp.Close()

	json := gjson.New(resp.ReadAllString())
	data := json.Array()
	klines := make([]*Kline, 0, len(data))

	for _, item := range data {
		arr := gjson.New(item).Array()
		if len(arr) >= 11 {
			klines = append(klines, &Kline{
				OpenTime:  gjson.New(arr[0]).Get(".").Int64(),
				Open:      gjson.New(arr[1]).Get(".").Float64(),
				High:      gjson.New(arr[2]).Get(".").Float64(),
				Low:       gjson.New(arr[3]).Get(".").Float64(),
				Close:     gjson.New(arr[4]).Get(".").Float64(),
				Volume:    gjson.New(arr[5]).Get(".").Float64(),
				CloseTime: gjson.New(arr[6]).Get(".").Int64(),
			})
		}
	}

	return klines, nil
}

// ========== OKX ==========

func (pms *PublicMarketService) fetchOKXTicker(ctx context.Context, symbol string) (*Ticker, error) {
	client := pms.getHttpClient()
	url := "https://www.okx.com/api/v5/market/ticker"

	resp, err := client.Get(ctx, url, g.Map{
		"instId": formatOKXSymbol(symbol),
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "OKX请求失败")
	}
	defer resp.Close()

	json := gjson.New(resp.ReadAllString())
	if json.Get("code").String() != "0" {
		return nil, gerror.Newf("OKX API error: %s", json.Get("msg").String())
	}

	data := json.Get("data").Array()
	if len(data) == 0 {
		return nil, gerror.New("OKX: No ticker data")
	}

	j := gjson.New(data[0])
	open24h := j.Get("open24h").Float64()
	last := j.Get("last").Float64()
	changePercent := 0.0
	if open24h > 0 {
		changePercent = (last - open24h) / open24h * 100
	}

	return &Ticker{
		Symbol:             symbol,
		LastPrice:          last,
		BidPrice:           j.Get("bidPx").Float64(),
		AskPrice:           j.Get("askPx").Float64(),
		High24h:            j.Get("high24h").Float64(),
		Low24h:             j.Get("low24h").Float64(),
		Volume24h:          j.Get("vol24h").Float64(),
		QuoteVolume24h:     j.Get("volCcy24h").Float64(),
		Change24h:          changePercent,
		PriceChangePercent: changePercent,
		Timestamp:          j.Get("ts").Int64(),
	}, nil
}

func (pms *PublicMarketService) fetchOKXKlines(ctx context.Context, symbol, interval string, limit int) ([]*Kline, error) {
	client := pms.getHttpClient()
	url := "https://www.okx.com/api/v5/market/candles"

	resp, err := client.Get(ctx, url, g.Map{
		"instId": formatOKXSymbol(symbol),
		"bar":    convertOKXInterval(interval),
		"limit":  limit,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "OKX K线请求失败")
	}
	defer resp.Close()

	json := gjson.New(resp.ReadAllString())
	if json.Get("code").String() != "0" {
		return nil, gerror.Newf("OKX API error: %s", json.Get("msg").String())
	}

	data := json.Get("data").Array()
	klines := make([]*Kline, 0, len(data))

	for _, item := range data {
		arr := gjson.New(item).Array()
		if len(arr) >= 6 {
			klines = append(klines, &Kline{
				OpenTime:  gjson.New(arr[0]).Get(".").Int64(),
				Open:      gjson.New(arr[1]).Get(".").Float64(),
				High:      gjson.New(arr[2]).Get(".").Float64(),
				Low:       gjson.New(arr[3]).Get(".").Float64(),
				Close:     gjson.New(arr[4]).Get(".").Float64(),
				Volume:    gjson.New(arr[5]).Get(".").Float64(),
				CloseTime: gjson.New(arr[0]).Get(".").Int64(),
			})
		}
	}

	return klines, nil
}

// ========== Gate.io ==========

func (pms *PublicMarketService) fetchGateTicker(ctx context.Context, symbol string) (*Ticker, error) {
	client := pms.getHttpClient()
	url := "https://api.gateio.ws/api/v4/futures/usdt/tickers"

	resp, err := client.Get(ctx, url, g.Map{
		"contract": formatGateSymbol(symbol),
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "Gate.io请求失败")
	}
	defer resp.Close()

	raw := resp.ReadAllString()
	// Gate 公共接口可能在受限网络/地区返回 403/451 或 HTML 拦截页；若不检查 status，会被误解析成“空数组”
	if resp.StatusCode != 200 {
		preview := raw
		if len(preview) > 400 {
			preview = preview[:400]
		}
		return nil, gerror.Newf("Gate.io http status=%d, url=%s, preview=%s", resp.StatusCode, url, preview)
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || (!strings.HasPrefix(trimmed, "[") && !strings.HasPrefix(trimmed, "{")) {
		preview := trimmed
		if len(preview) > 400 {
			preview = preview[:400]
		}
		return nil, gerror.Newf("Gate.io ticker unexpected body, url=%s, preview=%s", url, preview)
	}

	json := gjson.New(raw)
	data := json.Array()
	if len(data) == 0 {
		return nil, gerror.Newf("Gate.io: No ticker data, contract=%s", formatGateSymbol(symbol))
	}

	j := gjson.New(data[0])
	priceChangePercent := j.Get("change_percentage").Float64()
	return &Ticker{
		Symbol:             symbol,
		LastPrice:          j.Get("last").Float64(),
		BidPrice:           j.Get("highest_bid").Float64(),
		AskPrice:           j.Get("lowest_ask").Float64(),
		High24h:            j.Get("high_24h").Float64(),
		Low24h:             j.Get("low_24h").Float64(),
		Volume24h:          j.Get("volume_24h").Float64(),
		QuoteVolume24h:     j.Get("volume_24h_quote").Float64(),
		Change24h:          priceChangePercent,
		PriceChangePercent: priceChangePercent,
		Timestamp:          time.Now().UnixMilli(),
	}, nil
}

func (pms *PublicMarketService) fetchGateKlines(ctx context.Context, symbol, interval string, limit int) ([]*Kline, error) {
	client := pms.getHttpClient()
	url := "https://api.gateio.ws/api/v4/futures/usdt/candlesticks"

	resp, err := client.Get(ctx, url, g.Map{
		"contract": formatGateSymbol(symbol),
		"interval": convertGateInterval(interval),
		"limit":    limit,
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "Gate.io K线请求失败")
	}
	defer resp.Close()

	raw := resp.ReadAllString()
	// Gate 公共接口可能在受限网络/地区返回 403/451 或 HTML 拦截页；若不检查 status，会被误解析成“空数组”
	if resp.StatusCode != 200 {
		preview := raw
		if len(preview) > 400 {
			preview = preview[:400]
		}
		return nil, gerror.Newf("Gate.io K线 http status=%d, url=%s, preview=%s", resp.StatusCode, url, preview)
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || (!strings.HasPrefix(trimmed, "[") && !strings.HasPrefix(trimmed, "{")) {
		preview := trimmed
		if len(preview) > 400 {
			preview = preview[:400]
		}
		return nil, gerror.Newf("Gate.io K线 unexpected body, url=%s, contract=%s, interval=%s, limit=%d, preview=%s",
			url, formatGateSymbol(symbol), convertGateInterval(interval), limit, preview)
	}

	json := gjson.New(raw)
	data := json.Array()
	if len(data) == 0 {
		return nil, gerror.Newf("Gate.io K线: empty data, contract=%s, interval=%s, limit=%d",
			formatGateSymbol(symbol), convertGateInterval(interval), limit)
	}
	klines := make([]*Kline, 0, len(data))

	for _, item := range data {
		arr := gjson.New(item).Array()
		if len(arr) < 6 {
			continue
		}
		t0 := g.NewVar(arr[0]).Int64()
		openTime := t0
		if openTime > 0 && openTime < 1e12 {
			openTime = openTime * 1000
		}

		type cand struct {
			open, high, low, close, vol float64
		}
		cands := []cand{
			// [t, v, c, h, l, o]
			{open: g.NewVar(arr[5]).Float64(), high: g.NewVar(arr[3]).Float64(), low: g.NewVar(arr[4]).Float64(), close: g.NewVar(arr[2]).Float64(), vol: g.NewVar(arr[1]).Float64()},
			// [t, o, h, l, c, v]
			{open: g.NewVar(arr[1]).Float64(), high: g.NewVar(arr[2]).Float64(), low: g.NewVar(arr[3]).Float64(), close: g.NewVar(arr[4]).Float64(), vol: g.NewVar(arr[5]).Float64()},
			// [t, o, c, h, l, v]
			{open: g.NewVar(arr[1]).Float64(), high: g.NewVar(arr[3]).Float64(), low: g.NewVar(arr[4]).Float64(), close: g.NewVar(arr[2]).Float64(), vol: g.NewVar(arr[5]).Float64()},
			// [t, c, h, l, o, v]
			{open: g.NewVar(arr[4]).Float64(), high: g.NewVar(arr[2]).Float64(), low: g.NewVar(arr[3]).Float64(), close: g.NewVar(arr[1]).Float64(), vol: g.NewVar(arr[5]).Float64()},
		}
		choose := func(c cand) bool {
			if c.high <= 0 || c.low <= 0 || c.open <= 0 || c.close <= 0 {
				return false
			}
			if c.high < c.low {
				return false
			}
			if c.high < c.open || c.high < c.close {
				return false
			}
			if c.low > c.open || c.low > c.close {
				return false
			}
			return true
		}
		picked := cands[0]
		for _, c := range cands {
			if choose(c) {
				picked = c
				break
			}
		}
		if !choose(picked) {
			// 仍不合理则跳过（避免污染分析）
			continue
		}

		klines = append(klines, &Kline{
			OpenTime:  openTime,
			Open:      picked.open,
			High:      picked.high,
			Low:       picked.low,
			Close:     picked.close,
			Volume:    picked.vol,
			CloseTime: openTime,
		})
	}

	return klines, nil
}

// GetKlines 获取K线数据
func (pms *PublicMarketService) GetKlines(ctx context.Context, platform, symbol, interval string, limit int) ([]*Kline, error) {
	cacheKey := "pub_klines:" + platform + ":" + symbol + ":" + interval

	// 尝试从缓存获取
	if cached, err := pms.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		if klines, ok := cached.Val().([]*Kline); ok {
			return klines, nil
		}
	}

	var klines []*Kline
	var err error

	switch platform {
	case PlatformBinance:
		klines, err = pms.fetchBinanceKlines(ctx, symbol, interval, limit)
	case PlatformOKX:
		klines, err = pms.fetchOKXKlines(ctx, symbol, interval, limit)
	case PlatformGate:
		klines, err = pms.fetchGateKlines(ctx, symbol, interval, limit)
	default:
		return nil, gerror.Newf("不支持的交易所: %s", platform)
	}

	if err != nil {
		return nil, err
	}

	// K线缓存时间稍长
	_ = pms.cache.Set(ctx, cacheKey, klines, pms.cacheTTL*3)

	return klines, nil
}

// GetMultiTickers 批量获取同一交易所的多个交易对行情
func (pms *PublicMarketService) GetMultiTickers(ctx context.Context, platform string, symbols []string) (map[string]*Ticker, error) {
	result := make(map[string]*Ticker)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, symbol := range symbols {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			ticker, err := pms.GetTicker(ctx, platform, s)
			if err != nil {
				g.Log().Debugf(ctx, "获取 %s %s 行情失败: %v", platform, s, err)
				return
			}
			mu.Lock()
			result[s] = ticker
			mu.Unlock()
		}(symbol)
	}

	wg.Wait()
	return result, nil
}

// GetSupportedPlatforms 获取支持的交易所列表
func (pms *PublicMarketService) GetSupportedPlatforms() []string {
	return []string{PlatformBinance, PlatformOKX, PlatformGate}
}

// ========== Symbol格式化 ==========

func formatBinanceSymbol(symbol string) string {
	// BTCUSDT -> BTCUSDT
	return symbol
}

func formatOKXSymbol(symbol string) string {
	// BTCUSDT -> BTC-USDT-SWAP
	if len(symbol) > 4 && symbol[len(symbol)-4:] == "USDT" {
		base := symbol[:len(symbol)-4]
		return base + "-USDT-SWAP"
	}
	return symbol + "-SWAP"
}

func formatGateSymbol(symbol string) string {
	// Gate futures contract: BTC_USDT
	// 兼容输入：BTCUSDT / BTC_USDT / BTC-USDT / BTC/USDT
	s := strings.ToUpper(strings.TrimSpace(symbol))
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "/", "")
	if strings.HasSuffix(s, "USDT") {
		base := strings.TrimSuffix(s, "USDT")
		return base + "_USDT"
	}
	// 兜底：如果输入本身就是带 "_" 的 contract
	if strings.Contains(symbol, "_") {
		return strings.ToUpper(strings.TrimSpace(symbol))
	}
	return s
}

// ========== Interval转换 ==========

func convertOKXInterval(interval string) string {
	mapping := map[string]string{
		"1m": "1m", "5m": "5m", "15m": "15m", "30m": "30m",
		"1h": "1H", "4h": "4H", "1d": "1D",
	}
	if v, ok := mapping[interval]; ok {
		return v
	}
	return "15m"
}

func convertGateInterval(interval string) string {
	mapping := map[string]string{
		"1m": "1m", "5m": "5m", "15m": "15m", "30m": "30m",
		"1h": "1h", "60m": "1h", "4h": "4h", "1d": "1d",
	}
	if v, ok := mapping[interval]; ok {
		return v
	}
	return "15m"
}

// InitPublicMarketService 初始化公共行情服务（从配置文件）
func InitPublicMarketService(ctx context.Context) *PublicMarketService {
	pms := GetPublicMarketService()

	// 读取代理配置
	proxyEnabled := g.Cfg().MustGet(ctx, "exchange.proxy.enabled", false).Bool()
	proxyType := g.Cfg().MustGet(ctx, "exchange.proxy.type", "socks5").String()
	proxyHost := g.Cfg().MustGet(ctx, "exchange.proxy.host", "127.0.0.1").String()
	proxyPort := g.Cfg().MustGet(ctx, "exchange.proxy.port", 10808).Int()

	if proxyEnabled {
		pms.SetProxy(&ProxyConfig{
			Enabled: true,
			Type:    proxyType,
			Host:    proxyHost,
			Port:    proxyPort,
		})
	}

	// 设置缓存时间（可配置）
	cacheTTL := g.Cfg().MustGet(ctx, "exchange.cacheTTL", 1).Int()
	pms.SetCacheTTL(time.Duration(cacheTTL) * time.Second)

	return pms
}

// 兼容旧接口
type PublicMarket = PublicMarketService
type PublicMarketConfig struct {
	Platform string
	Proxy    *ProxyConfig
	CacheTTL time.Duration
}

func GetPublicMarket(config *PublicMarketConfig) *PublicMarket {
	return GetPublicMarketService()
}

func InitPublicMarket(ctx context.Context) *PublicMarket {
	return InitPublicMarketService(ctx)
}
