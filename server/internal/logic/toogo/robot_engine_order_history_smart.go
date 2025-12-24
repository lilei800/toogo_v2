package toogo

import (
	"context"
	"time"

	"hotgo/internal/library/exchange"

	"github.com/gogf/gf/v2/errors/gerror"
)

// GetOrderHistorySmart 智能获取订单历史（统一入口）
// - 兼容 OrderStatusSync 的调用：engine.GetOrderHistorySmart(ctx, 0, 50)
// - 内部带简单节流：120秒内优先返回缓存，避免重复打交易所API
// - startId 当前版本暂不使用（保留签名兼容）
func (e *RobotEngine) GetOrderHistorySmart(ctx context.Context, startId int64, limit int) ([]*exchange.Order, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if limit <= 0 {
		limit = 50
	}

	// 120s 内优先返回缓存
	e.mu.RLock()
	cached := e.OrderHistory
	last := e.LastOrderHistoryUpdate
	robot := e.Robot
	ex := e.Exchange
	e.mu.RUnlock()

	if cached != nil && time.Since(last) < 120*time.Second {
		return cached, nil
	}

	if robot == nil || ex == nil {
		// 没有引擎信息时，返回缓存（可能为nil）并给出错误
		if cached != nil {
			return cached, nil
		}
		return nil, gerror.New("机器人引擎未就绪，无法获取订单历史")
	}

	orders, err := ex.GetOrderHistory(ctx, robot.Symbol, limit)
	if err != nil {
		// API失败则回退缓存
		if cached != nil {
			return cached, nil
		}
		return nil, err
	}

	e.mu.Lock()
	e.OrderHistory = orders
	e.LastOrderHistoryUpdate = time.Now()
	e.mu.Unlock()

	return orders, nil
}


