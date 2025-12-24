// Package exchange
// @Description 交易所接口测试
package exchange

import (
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestCalculateStopLossPrice(t *testing.T) {
	testCases := []struct {
		name            string
		entryPrice      float64
		stopLossPercent float64
		positionSide    string
		expected        float64
	}{
		{"Long 5% stop loss", 100.0, 5.0, "LONG", 95.0},
		{"Short 5% stop loss", 100.0, 5.0, "SHORT", 105.0},
		{"Long 10% stop loss", 50000.0, 10.0, "LONG", 45000.0},
		{"Short 10% stop loss", 50000.0, 10.0, "SHORT", 55000.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CalculateStopLossPrice(tc.entryPrice, tc.stopLossPercent, tc.positionSide)
			if !almostEqual(result, tc.expected) {
				t.Errorf("Got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestCalculateTakeProfitPrice(t *testing.T) {
	testCases := []struct {
		name              string
		entryPrice        float64
		takeProfitPercent float64
		positionSide      string
		expected          float64
	}{
		{"Long 10% take profit", 100.0, 10.0, "LONG", 110.0},
		{"Short 10% take profit", 100.0, 10.0, "SHORT", 90.0},
		{"Long 20% take profit", 50000.0, 20.0, "LONG", 60000.0},
		{"Short 20% take profit", 50000.0, 20.0, "SHORT", 40000.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CalculateTakeProfitPrice(tc.entryPrice, tc.takeProfitPercent, tc.positionSide)
			if !almostEqual(result, tc.expected) {
				t.Errorf("Got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestCalculatePnLPercent(t *testing.T) {
	testCases := []struct {
		name         string
		entryPrice   float64
		currentPrice float64
		positionSide string
		leverage     int
		expected     float64
	}{
		{"Long profit", 100.0, 110.0, "LONG", 1, 10.0},
		{"Long loss", 100.0, 90.0, "LONG", 1, -10.0},
		{"Short profit", 100.0, 90.0, "SHORT", 1, 10.0},
		{"Short loss", 100.0, 110.0, "SHORT", 1, -10.0},
		{"Long profit 10x leverage", 100.0, 105.0, "LONG", 10, 50.0},
		{"Short profit 10x leverage", 100.0, 95.0, "SHORT", 10, 50.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CalculatePnLPercent(tc.entryPrice, tc.currentPrice, tc.positionSide, tc.leverage)
			if result != tc.expected {
				t.Errorf("Got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestCalculateLiquidationPrice(t *testing.T) {
	testCases := []struct {
		name         string
		entryPrice   float64
		leverage     int
		positionSide string
		marginType   string
	}{
		{"Long isolated 10x", 100.0, 10, "LONG", "isolated"},
		{"Short isolated 10x", 100.0, 10, "SHORT", "isolated"},
		{"Long cross 10x", 100.0, 10, "LONG", "cross"},
		{"Short cross 10x", 100.0, 10, "SHORT", "cross"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CalculateLiquidationPrice(tc.entryPrice, tc.leverage, tc.positionSide, tc.marginType)
			// 验证强平价格在合理范围内
			if tc.positionSide == "LONG" {
				if result >= tc.entryPrice {
					t.Errorf("Long liquidation price should be below entry price: got %v, entry %v", result, tc.entryPrice)
				}
			} else {
				if result <= tc.entryPrice {
					t.Errorf("Short liquidation price should be above entry price: got %v, entry %v", result, tc.entryPrice)
				}
			}
		})
	}
}

func TestCalculatePositionValue(t *testing.T) {
	testCases := []struct {
		quantity float64
		price    float64
		expected float64
	}{
		{1.0, 100.0, 100.0},
		{0.5, 50000.0, 25000.0},
		{0.001, 50000.0, 50.0},
	}

	for _, tc := range testCases {
		result := CalculatePositionValue(tc.quantity, tc.price)
		if result != tc.expected {
			t.Errorf("CalculatePositionValue(%v, %v) = %v, want %v", tc.quantity, tc.price, result, tc.expected)
		}
	}
}

func TestCalculateRequiredMargin(t *testing.T) {
	testCases := []struct {
		positionValue float64
		leverage      int
		expected      float64
	}{
		{1000.0, 10, 100.0},
		{5000.0, 20, 250.0},
		{10000.0, 1, 10000.0},
	}

	for _, tc := range testCases {
		result := CalculateRequiredMargin(tc.positionValue, tc.leverage)
		if result != tc.expected {
			t.Errorf("CalculateRequiredMargin(%v, %v) = %v, want %v", tc.positionValue, tc.leverage, result, tc.expected)
		}
	}
}

func TestValidateOrderRequest(t *testing.T) {
	testCases := []struct {
		name    string
		request *OrderRequest
		wantErr bool
	}{
		{
			name: "Valid market order",
			request: &OrderRequest{
				Symbol:   "BTCUSDT",
				Side:     "BUY",
				Type:     "MARKET",
				Quantity: 0.001,
			},
			wantErr: false,
		},
		{
			name: "Valid limit order",
			request: &OrderRequest{
				Symbol:   "BTCUSDT",
				Side:     "SELL",
				Type:     "LIMIT",
				Quantity: 0.001,
				Price:    50000,
			},
			wantErr: false,
		},
		{
			name: "Empty symbol",
			request: &OrderRequest{
				Symbol:   "",
				Side:     "BUY",
				Type:     "MARKET",
				Quantity: 0.001,
			},
			wantErr: true,
		},
		{
			name: "Invalid side",
			request: &OrderRequest{
				Symbol:   "BTCUSDT",
				Side:     "INVALID",
				Type:     "MARKET",
				Quantity: 0.001,
			},
			wantErr: true,
		},
		{
			name: "Invalid order type",
			request: &OrderRequest{
				Symbol:   "BTCUSDT",
				Side:     "BUY",
				Type:     "INVALID",
				Quantity: 0.001,
			},
			wantErr: true,
		},
		{
			name: "Zero quantity",
			request: &OrderRequest{
				Symbol:   "BTCUSDT",
				Side:     "BUY",
				Type:     "MARKET",
				Quantity: 0,
			},
			wantErr: true,
		},
		{
			name: "Limit order without price",
			request: &OrderRequest{
				Symbol:   "BTCUSDT",
				Side:     "BUY",
				Type:     "LIMIT",
				Quantity: 0.001,
				Price:    0,
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateOrderRequest(tc.request)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateOrderRequest() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestProxyConfig(t *testing.T) {
	testCases := []struct {
		name     string
		config   *ProxyConfig
		expected string
	}{
		{
			name:     "Nil config",
			config:   nil,
			expected: "",
		},
		{
			name: "Disabled proxy",
			config: &ProxyConfig{
				Enabled: false,
				Type:    "socks5",
				Host:    "127.0.0.1",
				Port:    1080,
			},
			expected: "",
		},
		{
			name: "Simple proxy",
			config: &ProxyConfig{
				Enabled: true,
				Type:    "socks5",
				Host:    "127.0.0.1",
				Port:    1080,
			},
			expected: "socks5://127.0.0.1:1080",
		},
		{
			name: "Proxy with auth",
			config: &ProxyConfig{
				Enabled:  true,
				Type:     "http",
				Host:     "proxy.example.com",
				Port:     8080,
				Username: "user",
				Password: "pass",
			},
			expected: "http://user:pass@proxy.example.com:8080",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.config.GetProxyURL()
			if result != tc.expected {
				t.Errorf("GetProxyURL() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func BenchmarkCalculateStopLossPrice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CalculateStopLossPrice(50000, 5, "LONG")
	}
}

func BenchmarkCalculatePnLPercent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CalculatePnLPercent(50000, 55000, "LONG", 10)
	}
}

