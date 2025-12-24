// Package trading 公共行情API（支持多交易所）
package trading

import "github.com/gogf/gf/v2/frame/g"

// PublicTickerReq 获取公共行情请求
type PublicTickerReq struct {
	g.Meta   `path:"/trading/public/ticker" method:"get" tags:"公共行情" summary:"获取实时行情" dc:"获取交易对实时行情（无需API Key）"`
	Platform string `json:"platform" d:"bitget" dc:"交易所：binance/bitget/okx/gate"`
	Symbol   string `json:"symbol" v:"required#交易对不能为空" dc:"交易对，如BTCUSDT"`
}

// PublicTickerRes 获取公共行情响应
type PublicTickerRes struct {
	Platform           string  `json:"platform" dc:"交易所"`
	Symbol             string  `json:"symbol" dc:"交易对"`
	LastPrice          float64 `json:"lastPrice" dc:"最新价格"`
	BidPrice           float64 `json:"bidPrice" dc:"买一价"`
	AskPrice           float64 `json:"askPrice" dc:"卖一价"`
	High24h            float64 `json:"high24h" dc:"24小时最高价"`
	Low24h             float64 `json:"low24h" dc:"24小时最低价"`
	Volume24h          float64 `json:"volume24h" dc:"24小时成交量"`
	QuoteVolume24h     float64 `json:"quoteVolume24h" dc:"24小时成交额"`
	PriceChangePercent float64 `json:"priceChangePercent" dc:"24小时涨跌幅"`
	Timestamp          int64   `json:"timestamp" dc:"时间戳"`
}

// PublicTickerRealtimeReq 获取实时行情请求（无缓存）
type PublicTickerRealtimeReq struct {
	g.Meta   `path:"/trading/public/ticker/realtime" method:"get" tags:"公共行情" summary:"获取实时行情(无缓存)" dc:"获取最新行情，不使用缓存"`
	Platform string `json:"platform" d:"bitget" dc:"交易所：binance/bitget/okx/gate"`
	Symbol   string `json:"symbol" v:"required#交易对不能为空" dc:"交易对，如BTCUSDT"`
}

// PublicTickerRealtimeRes 获取实时行情响应
type PublicTickerRealtimeRes = PublicTickerRes

// PublicKlinesReq 获取K线数据请求
type PublicKlinesReq struct {
	g.Meta   `path:"/trading/public/klines" method:"get" tags:"公共行情" summary:"获取K线数据" dc:"获取交易对K线数据（无需API Key）"`
	Platform string `json:"platform" d:"bitget" dc:"交易所：binance/bitget/okx/gate"`
	Symbol   string `json:"symbol" v:"required#交易对不能为空" dc:"交易对，如BTCUSDT"`
	Interval string `json:"interval" d:"15m" dc:"K线周期：1m/5m/15m/30m/1h/4h/1d"`
	Limit    int    `json:"limit" d:"100" dc:"数量，最大500"`
}

// PublicKlinesRes 获取K线数据响应
type PublicKlinesRes struct {
	Platform string       `json:"platform" dc:"交易所"`
	Symbol   string       `json:"symbol" dc:"交易对"`
	Interval string       `json:"interval" dc:"K线周期"`
	List     []*KlineItem `json:"list" dc:"K线列表"`
}

// KlineItem K线数据项
type KlineItem struct {
	OpenTime  int64   `json:"openTime" dc:"开盘时间"`
	Open      float64 `json:"open" dc:"开盘价"`
	High      float64 `json:"high" dc:"最高价"`
	Low       float64 `json:"low" dc:"最低价"`
	Close     float64 `json:"close" dc:"收盘价"`
	Volume    float64 `json:"volume" dc:"成交量"`
	CloseTime int64   `json:"closeTime" dc:"收盘时间"`
}

// PublicMultiTickersReq 批量获取行情请求
type PublicMultiTickersReq struct {
	g.Meta   `path:"/trading/public/multiTickers" method:"get" tags:"公共行情" summary:"批量获取行情" dc:"批量获取多个交易对行情（无需API Key）"`
	Platform string `json:"platform" d:"bitget" dc:"交易所：binance/bitget/okx/gate"`
	Symbols  string `json:"symbols" v:"required#交易对不能为空" dc:"交易对列表，逗号分隔，如BTCUSDT,ETHUSDT"`
}

// PublicMultiTickersRes 批量获取行情响应
type PublicMultiTickersRes struct {
	Platform string                      `json:"platform" dc:"交易所"`
	List     map[string]*PublicTickerRes `json:"list" dc:"行情列表"`
}

// PublicAllExchangesTickerReq 获取所有交易所行情请求
type PublicAllExchangesTickerReq struct {
	g.Meta `path:"/trading/public/allExchanges" method:"get" tags:"公共行情" summary:"获取所有交易所行情" dc:"同时获取多个交易所的同一交易对行情"`
	Symbol string `json:"symbol" v:"required#交易对不能为空" dc:"交易对，如BTCUSDT"`
}

// PublicAllExchangesTickerRes 获取所有交易所行情响应
type PublicAllExchangesTickerRes struct {
	Symbol  string                      `json:"symbol" dc:"交易对"`
	Tickers map[string]*PublicTickerRes `json:"tickers" dc:"各交易所行情"`
}

// PublicBestPriceReq 获取最优价格请求
type PublicBestPriceReq struct {
	g.Meta `path:"/trading/public/bestPrice" method:"get" tags:"公共行情" summary:"获取最优价格" dc:"比较所有交易所，获取最优买卖价"`
	Symbol string `json:"symbol" v:"required#交易对不能为空" dc:"交易对，如BTCUSDT"`
}

// PublicBestPriceRes 获取最优价格响应
type PublicBestPriceRes struct {
	Symbol          string                      `json:"symbol" dc:"交易对"`
	BestBid         float64                     `json:"bestBid" dc:"最高买价"`
	BestBidExchange string                      `json:"bestBidExchange" dc:"最高买价交易所"`
	BestAsk         float64                     `json:"bestAsk" dc:"最低卖价"`
	BestAskExchange string                      `json:"bestAskExchange" dc:"最低卖价交易所"`
	Spread          float64                     `json:"spread" dc:"价差百分比"`
	ReferencePrice  float64                     `json:"referencePrice" dc:"参考价格"`
	Tickers         map[string]*PublicTickerRes `json:"tickers" dc:"各交易所行情"`
	Timestamp       int64                       `json:"timestamp" dc:"时间戳"`
}

// PublicPlatformsReq 获取支持的交易所列表请求
type PublicPlatformsReq struct {
	g.Meta `path:"/trading/public/platforms" method:"get" tags:"公共行情" summary:"获取支持的交易所" dc:"获取系统支持的交易所列表"`
}

// PublicPlatformsRes 获取支持的交易所列表响应
type PublicPlatformsRes struct {
	List []string `json:"list" dc:"交易所列表"`
}
