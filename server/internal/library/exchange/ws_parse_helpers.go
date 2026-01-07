package exchange

// parseFloat/parseInt are small helpers used by some WS implementations (e.g. Binance).
// They delegate to the shared "Any" variants to handle string/float64/json.Number safely.

func parseFloat(v interface{}) float64 {
	return parseFloatAny(v)
}

func parseInt(v interface{}) int64 {
	return parseIntAny(v)
}


