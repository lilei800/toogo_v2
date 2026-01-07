package market

import (
	"strings"

	"hotgo/internal/library/exchange"
)

// NormalizePlatform normalizes platform names (binance/bitget/okx/gate...).
// Keep this as the single source of truth to avoid key mismatches across
// MarketServiceManager / MarketAnalyzer / RobotEngine.
func NormalizePlatform(platform string) string {
	p := strings.ToLower(strings.TrimSpace(platform))
	switch p {
	// Gate 的历史/别名口径：UI/DB/配置可能出现 gateio / gate.io
	case "gateio", "gate.io", "gate-io", "gate_io":
		return "gate"
	// OKX 的历史别名
	case "okex", "okex-swap", "okexswap":
		return "okx"
	default:
		return p
	}
}

// NormalizeSymbol normalizes symbols/instIds.
// IMPORTANT:
// - This is used as the internal cache key across MarketServiceManager / MarketAnalyzer / RobotEngine.
// - Must unify common UI/DB formats like BTC_USDT / BTC-USDT / BTCUSDT into BTCUSDT.
// - Also strips common suffixes like -SWAP / PERP to avoid key mismatches.
func NormalizeSymbol(symbol string) string {
	return exchange.Formatter.NormalizeSymbol(symbol)
}


