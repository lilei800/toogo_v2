package main

import (
	"flag"
	"fmt"
	"math"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "hotgo/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"
	"hotgo/internal/library/exchange"
)

// 只读联调：不下单、不改杠杆、不改逐仓模式
// 用法示例（PowerShell）：
//   cd D:\go\src\hotgo_v2\server
//   go run internal/cmd/exchange_smoketest/main.go --apiConfigId=3 --symbol=BTCUSDT --interval=5m --limit=10
func main() {
	var (
		apiConfigId = flag.Int64("apiConfigId", 0, "API配置ID（hg_trading_api_config.id）")
		symbol      = flag.String("symbol", "BTCUSDT", "交易对（如 BTCUSDT）")
		interval    = flag.String("interval", "5m", "K线周期（如 1m/5m/1h）")
		limit       = flag.Int("limit", 10, "K线条数")
		trade       = flag.Int("trade", 0, "是否执行真实下单/平仓联调：1=是（默认0只读）")
		yes         = flag.Int("yes", 0, "真实下单二次确认：必须同时 --trade=1 --yes=1 才会执行")
		leverage    = flag.Int("leverage", 3, "下单前设置的杠杆倍数（仅 trade=1 生效）")
		qty         = flag.Float64("qty", 0.0001, "每次开仓数量（基础币数量，例 BTC=0.0001）（仅 trade=1 生效）")
		sleepSec    = flag.Int("sleepSec", 2, "开仓/平仓后等待秒数（仅 trade=1 生效）")
	)
	flag.Parse()

	if *apiConfigId <= 0 {
		fmt.Println("❌ 缺少参数：--apiConfigId")
		return
	}

	ctx := gctx.New()
	ex, err := exchange.GetManager().GetExchange(ctx, *apiConfigId)
	if err != nil {
		fmt.Printf("❌ 获取交易所实例失败: %v\n", err)
		return
	}

	fmt.Println("========================================")
	fmt.Printf("Exchange SmokeTest (apiConfigId=%d, platform=%s)\n", *apiConfigId, ex.GetName())
	fmt.Println("========================================")

	// 1) Ticker
	fmt.Printf("\n[1] GetTicker: %s\n", *symbol)
	tk, err := ex.GetTicker(ctx, *symbol)
	if err != nil {
		fmt.Printf("❌ GetTicker失败: %v\n", err)
	} else {
		fmt.Printf("✅ last=%.4f bid=%.4f ask=%.4f ts=%d\n", tk.LastPrice, tk.BidPrice, tk.AskPrice, tk.Timestamp)
	}

	// 2) Klines
	fmt.Printf("\n[2] GetKlines: %s %s limit=%d\n", *symbol, *interval, *limit)
	ks, err := ex.GetKlines(ctx, *symbol, *interval, *limit)
	if err != nil {
		fmt.Printf("❌ GetKlines失败: %v\n", err)
	} else {
		fmt.Printf("✅ klines=%d\n", len(ks))
		if len(ks) > 0 {
			k := ks[len(ks)-1]
			fmt.Printf("   lastK: t=%d o=%.4f h=%.4f l=%.4f c=%.4f v=%.4f\n", k.OpenTime, k.Open, k.High, k.Low, k.Close, k.Volume)
		}
	}

	// 3) Balance
	fmt.Println("\n[3] GetBalance")
	bal, err := ex.GetBalance(ctx)
	if err != nil {
		fmt.Printf("❌ GetBalance失败: %v\n", err)
	} else {
		fmt.Printf("✅ total=%.4f avail=%.4f frozen=%.4f upl=%.4f %s\n", bal.TotalBalance, bal.AvailableBalance, bal.FrozenBalance, bal.UnrealizedPnl, bal.Currency)
	}

	// 4) Positions
	fmt.Println("\n[4] GetPositions (all)")
	pos, err := ex.GetPositions(ctx, "")
	if err != nil {
		fmt.Printf("❌ GetPositions失败: %v\n", err)
	} else {
		fmt.Printf("✅ positions=%d\n", len(pos))
		for _, p := range pos {
			fmt.Printf("   %s %s amt=%.8f entry=%.4f mark=%.4f pnl=%.4f lev=%d marginType=%s\n",
				p.Symbol, p.PositionSide, p.PositionAmt, p.EntryPrice, p.MarkPrice, p.UnrealizedPnl, p.Leverage, p.MarginType)
		}
	}

	// 5) OpenOrders
	fmt.Printf("\n[5] GetOpenOrders: %s\n", *symbol)
	oo, err := ex.GetOpenOrders(ctx, *symbol)
	if err != nil {
		fmt.Printf("❌ GetOpenOrders失败: %v\n", err)
	} else {
		fmt.Printf("✅ openOrders=%d\n", len(oo))
	}

	// 6) OrderHistory
	fmt.Printf("\n[6] GetOrderHistory: %s limit=20\n", *symbol)
	his, err := ex.GetOrderHistory(ctx, *symbol, 20)
	if err != nil {
		fmt.Printf("❌ GetOrderHistory失败: %v\n", err)
	} else {
		fmt.Printf("✅ history=%d\n", len(his))
	}

	// ===== 可选：真实下单/平仓联调 =====
	if *trade == 1 {
		if *yes != 1 {
			fmt.Println("\n⚠️ 已请求 trade 模式，但未提供 --yes=1，已安全退出（未下单）")
			return
		}
		if *qty <= 0 {
			fmt.Println("\n❌ trade 模式 qty 必须 > 0")
			return
		}
		if *leverage <= 0 {
			fmt.Println("\n❌ trade 模式 leverage 必须 > 0")
			return
		}

		fmt.Println("\n========================================")
		fmt.Println("⚠️ 真实下单/平仓联调开始（请确认该账号允许交易且风险自担）")
		fmt.Println("========================================")

		// 1) 尝试设置逐仓（某些交易所可能忽略/报错，这里不阻断）
		fmt.Println("\n[T1] SetMarginType: ISOLATED")
		if err := ex.SetMarginType(ctx, *symbol, "ISOLATED"); err != nil {
			fmt.Printf("⚠️ SetMarginType失败（忽略继续）：%v\n", err)
		} else {
			fmt.Println("✅ SetMarginType成功")
		}

		// 2) 设置杠杆
		fmt.Printf("\n[T2] SetLeverage: %dx\n", *leverage)
		if err := ex.SetLeverage(ctx, *symbol, *leverage); err != nil {
			fmt.Printf("⚠️ SetLeverage失败（忽略继续）：%v\n", err)
		} else {
			fmt.Println("✅ SetLeverage成功")
		}

		// 3) 开多
		fmt.Printf("\n[T3] Open LONG (market): qty=%.8f\n", *qty)
		longOrder, err := ex.CreateOrder(ctx, &exchange.OrderRequest{
			Symbol:       *symbol,
			Side:         "BUY",
			PositionSide: "LONG",
			Type:         "MARKET",
			Quantity:     *qty,
		})
		if err != nil {
			fmt.Printf("❌ 开多失败: %v\n", err)
		} else {
			fmt.Printf("✅ 开多提交成功: orderId=%s\n", longOrder.OrderId)
		}

		// 4) 开空
		fmt.Printf("\n[T4] Open SHORT (market): qty=%.8f\n", *qty)
		shortOrder, err := ex.CreateOrder(ctx, &exchange.OrderRequest{
			Symbol:       *symbol,
			Side:         "SELL",
			PositionSide: "SHORT",
			Type:         "MARKET",
			Quantity:     *qty,
		})
		if err != nil {
			fmt.Printf("❌ 开空失败: %v\n", err)
		} else {
			fmt.Printf("✅ 开空提交成功: orderId=%s\n", shortOrder.OrderId)
		}

		if *sleepSec > 0 {
			time.Sleep(time.Duration(*sleepSec) * time.Second)
		}

		// 5) 查持仓并按实际持仓量平仓（reduceOnly 由各交易所实现保障）
		fmt.Println("\n[T5] Close positions by current PositionAmt (reduceOnly)")
		posAll, err := ex.GetPositions(ctx, "")
		if err != nil {
			fmt.Printf("❌ GetPositions失败: %v\n", err)
		} else {
			closed := 0
			for _, p := range posAll {
				if p.Symbol != *symbol && p.Symbol != "" {
					continue
				}
				amt := math.Abs(p.PositionAmt)
				if amt <= 0 {
					continue
				}
				_, cerr := ex.ClosePosition(ctx, *symbol, p.PositionSide, amt)
				if cerr != nil {
					fmt.Printf("❌ 平仓失败: %s %s amt=%.8f err=%v\n", *symbol, p.PositionSide, amt, cerr)
					continue
				}
				fmt.Printf("✅ 平仓提交: %s %s amt=%.8f\n", *symbol, p.PositionSide, amt)
				closed++
			}
			if closed == 0 {
				fmt.Println("⚠️ 未检测到可平仓持仓（可能是未成交/接口限制/符号字段差异）")
			}
		}

		if *sleepSec > 0 {
			time.Sleep(time.Duration(*sleepSec) * time.Second)
		}

		// 6) 再查一遍持仓
		fmt.Println("\n[T6] GetPositions after close")
		posAfter, err := ex.GetPositions(ctx, "")
		if err != nil {
			fmt.Printf("❌ GetPositions失败: %v\n", err)
		} else {
			fmt.Printf("✅ positions=%d\n", len(posAfter))
			for _, p := range posAfter {
				fmt.Printf("   %s %s amt=%.8f entry=%.4f mark=%.4f pnl=%.4f\n", p.Symbol, p.PositionSide, p.PositionAmt, p.EntryPrice, p.MarkPrice, p.UnrealizedPnl)
			}
		}

		fmt.Println("\n✅ Trade SmokeTest 完成（已尝试开多+开空+平仓）")
		return
	}

	fmt.Println("\n✅ SmokeTest 完成（只读，不包含下单/平仓；如需真实联调请加 --trade=1 --yes=1）")
}


