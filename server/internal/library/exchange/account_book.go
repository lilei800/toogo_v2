package exchange

// AccountBookItem represents an account ledger (资金流水/账本) record.
// For some exchanges (e.g. Gate futures), realized PnL and fees are more reliably available here
// than on "my_trades" (fills) APIs.
type AccountBookItem struct {
	// Time: timestamp in milliseconds (ms).
	Time int64 `json:"time"`
	// Type: record type, e.g. "pnl", "fee", "trade_fee", etc.
	Type string `json:"type"`
	// Change: balance change amount. Profit is typically positive; fee is typically negative.
	Change float64 `json:"change"`
	// Currency: quote currency, usually "USDT" for USDT-settled futures.
	Currency string `json:"currency"`
	// Symbol: unified symbol like BTCUSDT (best-effort; some exchanges only provide contract like BTC_USDT).
	Symbol string `json:"symbol"`
	// Contract: exchange-native contract name like BTC_USDT.
	Contract string `json:"contract"`
	// OrderId: exchange order id if present.
	OrderId string `json:"orderId"`
	// Text: extra description.
	Text string `json:"text"`
}


