package exchange

import (
	"sort"
	"strings"
)

// FillRealizedPnlByAvgCost tries to fill Trade.RealizedPnl for exchanges that do not provide it
// (or provide empty/0) on "fill" APIs.
//
// It uses an average-cost position model per (symbol, positionSide):
// - LONG:
//   - BUY increases position and updates avg entry
//   - SELL decreases position and realizes PnL = (sellPx - avgEntryPx) * qtyClosed
// - SHORT:
//   - SELL increases position and updates avg entry
//   - BUY decreases position and realizes PnL = (avgEntryPx - buyPx) * qtyClosed
//
// Notes:
// - Only fills trades with RealizedPnl == 0 (won't overwrite exchange-provided non-zero values).
// - Requires Trade.PositionSide in {LONG, SHORT} and Trade.Side in {BUY, SELL}.
// - Quantity is assumed to be "base quantity" (consistent with our unified Trade/Order conventions).
func FillRealizedPnlByAvgCost(trades []*Trade) {
	if len(trades) == 0 {
		return
	}

	type posState struct {
		qty float64 // absolute position qty in base coin (always >= 0)
		avg float64 // avg entry price in quote currency (e.g. USDT)
	}

	// iterate in chronological order to build correct avg cost
	idx := make([]int, 0, len(trades))
	for i := range trades {
		idx = append(idx, i)
	}
	sort.SliceStable(idx, func(i, j int) bool {
		ti := int64(0)
		tj := int64(0)
		if trades[idx[i]] != nil {
			ti = trades[idx[i]].Time
		}
		if trades[idx[j]] != nil {
			tj = trades[idx[j]].Time
		}
		return ti < tj
	})

	st := make(map[string]*posState)
	keyOf := func(t *Trade) string {
		return strings.ToUpper(strings.TrimSpace(t.Symbol)) + ":" + strings.ToUpper(strings.TrimSpace(t.PositionSide))
	}

	for _, i := range idx {
		t := trades[i]
		if t == nil {
			continue
		}
		if t.RealizedPnl != 0 {
			continue // keep exchange-provided
		}
		ps := strings.ToUpper(strings.TrimSpace(t.PositionSide))
		side := strings.ToUpper(strings.TrimSpace(t.Side))
		if ps != "LONG" && ps != "SHORT" {
			continue
		}
		if side != "BUY" && side != "SELL" {
			continue
		}
		if t.Price <= 0 || t.Quantity <= 0 {
			continue
		}

		k := keyOf(t)
		s := st[k]
		if s == nil {
			s = &posState{}
			st[k] = s
		}

		switch ps {
		case "LONG":
			if side == "BUY" {
				// increase long position
				newQty := s.qty + t.Quantity
				if newQty > 0 {
					s.avg = (s.avg*s.qty + t.Price*t.Quantity) / newQty
				}
				s.qty = newQty
				continue
			}
			// SELL: close long
			if s.qty <= 0 || s.avg <= 0 {
				continue
			}
			closeQty := t.Quantity
			if closeQty > s.qty {
				closeQty = s.qty
			}
			if closeQty <= 0 {
				continue
			}
			t.RealizedPnl = (t.Price - s.avg) * closeQty
			s.qty -= closeQty
			if s.qty <= 0 {
				s.qty = 0
				s.avg = 0
			}
		case "SHORT":
			if side == "SELL" {
				// increase short position (store avg entry)
				newQty := s.qty + t.Quantity
				if newQty > 0 {
					s.avg = (s.avg*s.qty + t.Price*t.Quantity) / newQty
				}
				s.qty = newQty
				continue
			}
			// BUY: close short
			if s.qty <= 0 || s.avg <= 0 {
				continue
			}
			closeQty := t.Quantity
			if closeQty > s.qty {
				closeQty = s.qty
			}
			if closeQty <= 0 {
				continue
			}
			t.RealizedPnl = (s.avg - t.Price) * closeQty
			s.qty -= closeQty
			if s.qty <= 0 {
				s.qty = 0
				s.avg = 0
			}
		}
	}
}
