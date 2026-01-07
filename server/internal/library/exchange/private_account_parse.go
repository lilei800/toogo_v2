package exchange

import (
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
)

// ParseBalanceFromPrivateWS parses a private WS "account" event to a normalized Balance.
// Returns (bal, true) if parsed successfully (at least one key field is present).
func ParseBalanceFromPrivateWS(platform string, raw []byte) (*Balance, bool) {
	platform = strings.ToLower(strings.TrimSpace(platform))
	if len(raw) == 0 {
		return nil, false
	}
	switch platform {
	case "okx":
		return parseOKXAccountWS(raw)
	case "gate":
		return parseGateAccountWS(raw)
	default:
		return nil, false
	}
}

func parseOKXAccountWS(raw []byte) (*Balance, bool) {
	// OKX private ws "account":
	// { "arg":{...}, "data":[{ "totalEq":"...", "details":[{"availEq":"..."}] ... }] }
	j := gjson.New(string(raw))
	data := j.Get("data").Array()
	if len(data) == 0 {
		return nil, false
	}
	item := gjson.New(data[0])
	totalEq := item.Get("totalEq").Float64()
	availEq := float64(0)
	// prefer details[0].availEq (USDT margin)
	details := item.Get("details").Array()
	if len(details) > 0 {
		dd := gjson.New(details[0])
		availEq = dd.Get("availEq").Float64()
		if availEq == 0 {
			availEq = dd.Get("availBal").Float64()
		}
	}
	if totalEq == 0 && availEq == 0 {
		return nil, false
	}
	return &Balance{
		TotalBalance:     totalEq,
		AvailableBalance: availEq,
		Currency:         "USDT",
	}, true
}

func parseGateAccountWS(raw []byte) (*Balance, bool) {
	// Gate private ws futures.account:
	// Usually: { "channel":"futures.account", "event":"update", "result":{...} }
	// We'll accept result/data/result[0] variants.
	j := gjson.New(string(raw))

	// result/data may be map or array; normalize to a Json object for reading fields.
	obj := j
	if r := j.Get("result"); !r.IsNil() {
		if arr := r.Array(); len(arr) > 0 {
			obj = gjson.New(arr[0])
		} else if m := r.Map(); len(m) > 0 {
			obj = gjson.New(m)
		}
	} else if d := j.Get("data"); !d.IsNil() {
		if arr := d.Array(); len(arr) > 0 {
			obj = gjson.New(arr[0])
		} else if m := d.Map(); len(m) > 0 {
			obj = gjson.New(m)
		}
	}

	total := obj.Get("total").Float64()
	if total == 0 {
		total = obj.Get("total_wallet_balance").Float64()
	}
	if total == 0 {
		total = obj.Get("wallet_balance").Float64()
	}
	avail := obj.Get("available").Float64()
	if avail == 0 {
		avail = obj.Get("available_margin").Float64()
	}
	if avail == 0 {
		avail = obj.Get("available_balance").Float64()
	}
	if avail == 0 {
		avail = obj.Get("available_funds").Float64()
	}

	upl := obj.Get("unrealised_pnl").Float64()
	if upl == 0 {
		upl = obj.Get("unrealized_pnl").Float64()
	}

	equity := obj.Get("equity").Float64()
	if equity <= 0 {
		equity = obj.Get("account_equity").Float64()
	}
	if equity <= 0 {
		equity = total + upl
	}

	if equity == 0 && avail == 0 && total == 0 && upl == 0 {
		return nil, false
	}
	return &Balance{
		TotalBalance:     equity,
		AvailableBalance: avail,
		UnrealizedPnl:    upl,
		Currency:         "USDT",
	}, true
}

// NOTE: Bitget 已被移除，不再解析其私有WS account 事件
