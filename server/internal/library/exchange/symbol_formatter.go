// Package exchange 统一的Symbol格式管理
package exchange

import (
	"strings"
)

// SymbolFormatter 交易所Symbol格式转换器
type SymbolFormatter struct{}

var Formatter = &SymbolFormatter{}

// NormalizeSymbol 标准化Symbol为统一格式 BTCUSDT (无分隔符，大写)
// 输入支持: BTCUSDT, BTC/USDT, BTC-USDT, BTC_USDT, BTC-USDT-SWAP等
// 输出: BTCUSDT
func (f *SymbolFormatter) NormalizeSymbol(symbol string) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	
	// 移除所有后缀
	s = strings.ReplaceAll(s, "-SWAP", "")
	s = strings.ReplaceAll(s, "_SWAP", "")
	s = strings.ReplaceAll(s, "SWAP", "")
	s = strings.ReplaceAll(s, "-PERP", "")
	s = strings.ReplaceAll(s, "_PERP", "")
	s = strings.ReplaceAll(s, "PERP", "")
	s = strings.ReplaceAll(s, "_UMCBL", "")
	s = strings.ReplaceAll(s, "-UMCBL", "")
	s = strings.ReplaceAll(s, "UMCBL", "")
	
	// 移除所有分隔符
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, " ", "")
	
	return s
}

// FormatForBinance 格式化为Binance格式: BTCUSDT
func (f *SymbolFormatter) FormatForBinance(symbol string) string {
	return f.NormalizeSymbol(symbol)
}

// FormatForOKX 格式化为OKX格式: BTC-USDT-SWAP
func (f *SymbolFormatter) FormatForOKX(symbol string) string {
	s := f.NormalizeSymbol(symbol)
	
	if strings.HasSuffix(s, "USDT") {
		base := strings.TrimSuffix(s, "USDT")
		return base + "-USDT-SWAP"
	}
	if strings.HasSuffix(s, "USDC") {
		base := strings.TrimSuffix(s, "USDC")
		return base + "-USDC-SWAP"
	}
	
	return s
}

// FormatForGate 格式化为Gate格式: BTC_USDT
func (f *SymbolFormatter) FormatForGate(symbol string) string {
	s := f.NormalizeSymbol(symbol)
	
	if strings.HasSuffix(s, "USDT") {
		base := strings.TrimSuffix(s, "USDT")
		return base + "_USDT"
	}
	if strings.HasSuffix(s, "USDC") {
		base := strings.TrimSuffix(s, "USDC")
		return base + "_USDC"
	}
	
	return s
}

// FormatForPlatform 根据平台名称格式化Symbol
func (f *SymbolFormatter) FormatForPlatform(platform, symbol string) string {
	switch strings.ToLower(platform) {
	case "binance":
		return f.FormatForBinance(symbol)
	case "okx":
		return f.FormatForOKX(symbol)
	case "gate":
		return f.FormatForGate(symbol)
	default:
		return f.NormalizeSymbol(symbol)
	}
}

// ParseSymbol 从任意格式解析出基础币种和计价币种
// 例如: BTC-USDT-SWAP -> ("BTC", "USDT")
func (f *SymbolFormatter) ParseSymbol(symbol string) (base, quote string) {
	normalized := f.NormalizeSymbol(symbol)
	
	// 支持常见的计价币种
	quotes := []string{"USDT", "USDC", "USD", "BUSD", "DAI"}
	
	for _, q := range quotes {
		if strings.HasSuffix(normalized, q) {
			return strings.TrimSuffix(normalized, q), q
		}
	}
	
	// 默认假设最后4个字符是计价币种
	if len(normalized) > 4 {
		return normalized[:len(normalized)-4], normalized[len(normalized)-4:]
	}
	
	return normalized, ""
}

