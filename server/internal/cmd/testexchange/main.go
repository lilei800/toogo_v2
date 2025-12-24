// Package main 交易所API连接测试
package main

import (
	"context"
	"fmt"
	"hotgo/internal/library/exchange"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func main() {
	ctx := context.Background()

	// 从环境变量读取配置（避免把密钥写进代码）
	// 必填：EXCHANGE_PLATFORM, EXCHANGE_API_KEY, EXCHANGE_SECRET_KEY
	// OKX/Bitget 可能需要：EXCHANGE_PASSPHRASE
	// 可选：EXCHANGE_PROXY_URL（如 socks5://127.0.0.1:10808 或 http://127.0.0.1:7890）
	platform := getEnv("EXCHANGE_PLATFORM", "bitget")
	apiKey := getEnv("EXCHANGE_API_KEY", "")
	secretKey := getEnv("EXCHANGE_SECRET_KEY", "")
	passphrase := getEnv("EXCHANGE_PASSPHRASE", "")
	symbol := getEnv("EXCHANGE_SYMBOL", "BTCUSDT")
	interval := getEnv("EXCHANGE_INTERVAL", "5m")

	if apiKey == "" || secretKey == "" {
		fmt.Println("❌ 缺少环境变量：EXCHANGE_API_KEY / EXCHANGE_SECRET_KEY")
		fmt.Println("   示例：")
		fmt.Println("   $env:EXCHANGE_PLATFORM='okx'")
		fmt.Println("   $env:EXCHANGE_API_KEY='...'")
		fmt.Println("   $env:EXCHANGE_SECRET_KEY='...'")
		fmt.Println("   $env:EXCHANGE_PASSPHRASE='...'  # okx/bitget")
		fmt.Println("   go run internal/cmd/testexchange/main.go")
		return
	}

	config := &exchange.Config{
		Platform:   platform,
		ApiKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
		IsTestnet:  false,
	}
	// 代理：EXCHANGE_PROXY_URL=socks5://127.0.0.1:10808 或 http://127.0.0.1:7890
	if proxyURL := os.Getenv("EXCHANGE_PROXY_URL"); proxyURL != "" {
		if pu, err := url.Parse(proxyURL); err == nil {
			host := pu.Hostname()
			port, _ := strconv.Atoi(pu.Port())
			if host != "" && port > 0 {
				config.Proxy = &exchange.ProxyConfig{
					Enabled: true,
					Type:    strings.ToLower(pu.Scheme),
					Host:    host,
					Port:    port,
				}
				if pu.User != nil {
					config.Proxy.Username = pu.User.Username()
					if pwd, ok := pu.User.Password(); ok {
						config.Proxy.Password = pwd
					}
				}
			}
		}
	}

	fmt.Println("========================================")
	fmt.Printf("%s API 连接测试\n", strings.ToUpper(config.Platform))
	fmt.Println("========================================")
	if len(config.ApiKey) > 8 {
		fmt.Printf("API Key: %s...\n", config.ApiKey[:8])
	}
	if config.Proxy != nil && config.Proxy.Enabled {
		fmt.Printf("代理: %s://%s:%d\n", config.Proxy.Type, config.Proxy.Host, config.Proxy.Port)
	}
	fmt.Println("----------------------------------------")

	// 创建交易所实例
	ex, err := exchange.NewExchange(config)
	if err != nil {
		fmt.Printf("❌ 创建交易所实例失败: %v\n", err)
		return
	}

	// 1. 测试获取行情 (公开接口)
	fmt.Printf("\n[1] 测试获取行情 (%s)...\n", symbol)
	ticker, err := ex.GetTicker(ctx, symbol)
	if err != nil {
		fmt.Printf("❌ 获取行情失败: %v\n", err)
	} else {
		fmt.Printf("✅ 获取行情成功!\n")
		fmt.Printf("   最新价: %.2f USDT\n", ticker.LastPrice)
		fmt.Printf("   24h高: %.2f USDT\n", ticker.High24h)
		fmt.Printf("   24h低: %.2f USDT\n", ticker.Low24h)
		fmt.Printf("   24h涨跌: %.2f%%\n", ticker.PriceChangePercent)
	}

	// 2. 测试获取K线
	fmt.Printf("\n[2] 测试获取K线 (%s %s)...\n", symbol, interval)
	klines, err := ex.GetKlines(ctx, symbol, interval, 10)
	if err != nil {
		fmt.Printf("❌ 获取K线失败: %v\n", err)
	} else {
		fmt.Printf("✅ 获取K线成功! 共 %d 条\n", len(klines))
		if len(klines) > 0 {
			k := klines[len(klines)-1]
			fmt.Printf("   最新K线: 开=%.2f 高=%.2f 低=%.2f 收=%.2f\n",
				k.Open, k.High, k.Low, k.Close)
		}
	}

	// 3. 测试获取账户余额 (私有接口)
	fmt.Println("\n[3] 测试获取账户余额...")
	balance, err := ex.GetBalance(ctx)
	if err != nil {
		fmt.Printf("❌ 获取余额失败: %v\n", err)
	} else {
		fmt.Printf("✅ 获取余额成功!\n")
		fmt.Printf("   总余额: %.4f USDT\n", balance.TotalBalance)
		fmt.Printf("   可用余额: %.4f USDT\n", balance.AvailableBalance)
		fmt.Printf("   冻结余额: %.4f USDT\n", balance.FrozenBalance)
		fmt.Printf("   未实现盈亏: %.4f USDT\n", balance.UnrealizedPnl)
	}

	// 4. 测试获取持仓
	fmt.Println("\n[4] 测试获取持仓...")
	positions, err := ex.GetPositions(ctx, "")
	if err != nil {
		fmt.Printf("❌ 获取持仓失败: %v\n", err)
	} else {
		fmt.Printf("✅ 获取持仓成功! 共 %d 个持仓\n", len(positions))
		for _, pos := range positions {
			fmt.Printf("   %s %s: 数量=%.4f 开仓价=%.2f 盈亏=%.4f\n",
				pos.Symbol, pos.PositionSide, pos.PositionAmt, pos.EntryPrice, pos.UnrealizedPnl)
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("测试完成!")
	fmt.Println("========================================")
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

