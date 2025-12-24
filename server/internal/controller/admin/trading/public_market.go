// Package trading 公共行情控制器（支持多交易所）
package trading

import (
	"context"
	"strings"

	"hotgo/api/admin/trading"
	"hotgo/internal/library/exchange"
)

// PublicMarket 公共行情控制器（导出）
var PublicMarket = cPublicMarket{}

// cPublicMarket 公共行情控制器
type cPublicMarket struct{}

// Ticker 获取实时行情
func (c *cPublicMarket) Ticker(ctx context.Context, req *trading.PublicTickerReq) (res *trading.PublicTickerRes, err error) {
	pms := exchange.InitPublicMarketService(ctx)
	
	platform := req.Platform
	if platform == "" {
		platform = "bitget"
	}
	
	ticker, err := pms.GetTicker(ctx, platform, req.Symbol)
	if err != nil {
		return nil, err
	}

	res = &trading.PublicTickerRes{
		Platform:           platform,
		Symbol:             ticker.Symbol,
		LastPrice:          ticker.LastPrice,
		BidPrice:           ticker.BidPrice,
		AskPrice:           ticker.AskPrice,
		High24h:            ticker.High24h,
		Low24h:             ticker.Low24h,
		Volume24h:          ticker.Volume24h,
		QuoteVolume24h:     ticker.QuoteVolume24h,
		PriceChangePercent: ticker.PriceChangePercent,
		Timestamp:          ticker.Timestamp,
	}
	return
}

// TickerRealtime 获取实时行情（无缓存）
func (c *cPublicMarket) TickerRealtime(ctx context.Context, req *trading.PublicTickerRealtimeReq) (res *trading.PublicTickerRealtimeRes, err error) {
	pms := exchange.InitPublicMarketService(ctx)
	
	platform := req.Platform
	if platform == "" {
		platform = "bitget"
	}
	
	ticker, err := pms.GetTickerRealtime(ctx, platform, req.Symbol)
	if err != nil {
		return nil, err
	}

	res = &trading.PublicTickerRealtimeRes{
		Platform:           platform,
		Symbol:             ticker.Symbol,
		LastPrice:          ticker.LastPrice,
		BidPrice:           ticker.BidPrice,
		AskPrice:           ticker.AskPrice,
		High24h:            ticker.High24h,
		Low24h:             ticker.Low24h,
		Volume24h:          ticker.Volume24h,
		QuoteVolume24h:     ticker.QuoteVolume24h,
		PriceChangePercent: ticker.PriceChangePercent,
		Timestamp:          ticker.Timestamp,
	}
	return
}

// Klines 获取K线数据
func (c *cPublicMarket) Klines(ctx context.Context, req *trading.PublicKlinesReq) (res *trading.PublicKlinesRes, err error) {
	pms := exchange.InitPublicMarketService(ctx)

	platform := req.Platform
	if platform == "" {
		platform = "bitget"
	}

	limit := req.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	klines, err := pms.GetKlines(ctx, platform, req.Symbol, req.Interval, limit)
	if err != nil {
		return nil, err
	}

	res = &trading.PublicKlinesRes{
		Platform: platform,
		Symbol:   req.Symbol,
		Interval: req.Interval,
		List:     make([]*trading.KlineItem, len(klines)),
	}

	for i, k := range klines {
		res.List[i] = &trading.KlineItem{
			OpenTime:  k.OpenTime,
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
			CloseTime: k.CloseTime,
		}
	}

	return
}

// MultiTickers 批量获取行情
func (c *cPublicMarket) MultiTickers(ctx context.Context, req *trading.PublicMultiTickersReq) (res *trading.PublicMultiTickersRes, err error) {
	pms := exchange.InitPublicMarketService(ctx)

	platform := req.Platform
	if platform == "" {
		platform = "bitget"
	}

	symbols := strings.Split(req.Symbols, ",")
	tickers, err := pms.GetMultiTickers(ctx, platform, symbols)
	if err != nil {
		return nil, err
	}

	res = &trading.PublicMultiTickersRes{
		Platform: platform,
		List:     make(map[string]*trading.PublicTickerRes),
	}

	for symbol, ticker := range tickers {
		res.List[symbol] = &trading.PublicTickerRes{
			Platform:           platform,
			Symbol:             ticker.Symbol,
			LastPrice:          ticker.LastPrice,
			BidPrice:           ticker.BidPrice,
			AskPrice:           ticker.AskPrice,
			High24h:            ticker.High24h,
			Low24h:             ticker.Low24h,
			Volume24h:          ticker.Volume24h,
			QuoteVolume24h:     ticker.QuoteVolume24h,
			PriceChangePercent: ticker.PriceChangePercent,
			Timestamp:          ticker.Timestamp,
		}
	}

	return
}

// AllExchangesTicker 获取所有交易所行情
func (c *cPublicMarket) AllExchangesTicker(ctx context.Context, req *trading.PublicAllExchangesTickerReq) (res *trading.PublicAllExchangesTickerRes, err error) {
	pms := exchange.InitPublicMarketService(ctx)

	tickers, err := pms.GetAllExchangesTicker(ctx, req.Symbol)
	if err != nil {
		return nil, err
	}

	res = &trading.PublicAllExchangesTickerRes{
		Symbol:  req.Symbol,
		Tickers: make(map[string]*trading.PublicTickerRes),
	}

	for platform, ticker := range tickers {
		res.Tickers[platform] = &trading.PublicTickerRes{
			Platform:           platform,
			Symbol:             ticker.Symbol,
			LastPrice:          ticker.LastPrice,
			BidPrice:           ticker.BidPrice,
			AskPrice:           ticker.AskPrice,
			High24h:            ticker.High24h,
			Low24h:             ticker.Low24h,
			Volume24h:          ticker.Volume24h,
			QuoteVolume24h:     ticker.QuoteVolume24h,
			PriceChangePercent: ticker.PriceChangePercent,
			Timestamp:          ticker.Timestamp,
		}
	}

	return
}

// BestPrice 获取最优价格
func (c *cPublicMarket) BestPrice(ctx context.Context, req *trading.PublicBestPriceReq) (res *trading.PublicBestPriceRes, err error) {
	pms := exchange.InitPublicMarketService(ctx)

	result, err := pms.GetBestPrice(ctx, req.Symbol)
	if err != nil {
		return nil, err
	}

	res = &trading.PublicBestPriceRes{
		Symbol:          result.Symbol,
		BestBid:         result.BestBid,
		BestBidExchange: result.BestBidExchange,
		BestAsk:         result.BestAsk,
		BestAskExchange: result.BestAskExchange,
		Spread:          result.Spread,
		ReferencePrice:  result.ReferencePrice,
		Tickers:         make(map[string]*trading.PublicTickerRes),
		Timestamp:       result.Timestamp,
	}

	for platform, ticker := range result.Tickers {
		res.Tickers[platform] = &trading.PublicTickerRes{
			Platform:           platform,
			Symbol:             ticker.Symbol,
			LastPrice:          ticker.LastPrice,
			BidPrice:           ticker.BidPrice,
			AskPrice:           ticker.AskPrice,
			High24h:            ticker.High24h,
			Low24h:             ticker.Low24h,
			Volume24h:          ticker.Volume24h,
			QuoteVolume24h:     ticker.QuoteVolume24h,
			PriceChangePercent: ticker.PriceChangePercent,
			Timestamp:          ticker.Timestamp,
		}
	}

	return
}

// Platforms 获取支持的交易所列表
func (c *cPublicMarket) Platforms(ctx context.Context, req *trading.PublicPlatformsReq) (res *trading.PublicPlatformsRes, err error) {
	pms := exchange.InitPublicMarketService(ctx)
	res = &trading.PublicPlatformsRes{
		List: pms.GetSupportedPlatforms(),
	}
	return
}
