// =================================================================================
// This file is auto-generated for the table hg_trading_order_status_history.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// tradingOrderStatusHistoryDao is the data access object for the table hg_trading_order_status_history.
type tradingOrderStatusHistoryDao struct {
	*internal.TradingOrderStatusHistoryDao
}

var (
	// TradingOrderStatusHistory is a globally accessible object for table hg_trading_order_status_history operations.
	TradingOrderStatusHistory = tradingOrderStatusHistoryDao{internal.NewTradingOrderStatusHistoryDao()}
)

