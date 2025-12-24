package toogo

import (
	"context"
	"math"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

// CloseAllPositionsAndCancelOrders 强制撤销当前挂单并平掉所有持仓，并等待持仓归零（用于“停止前全平/达标后全平”）。
//
// 说明：
// - 这个方法是 RobotTaskManager.CloseAllAndWait 的运行中引擎分支所需的兼容实现。
// - 会尽力刷新引擎内存缓存（CurrentPositions/LastPositionUpdate），避免继续使用旧状态。
func (e *RobotEngine) CloseAllPositionsAndCancelOrders(ctx context.Context, reason string, timeout time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}

	e.mu.RLock()
	robot := e.Robot
	ex := e.Exchange
	e.mu.RUnlock()

	if robot == nil {
		return gerror.New("机器人配置为空，无法执行全平")
	}
	if ex == nil {
		return gerror.New("交易所实例不存在，无法执行全平")
	}

	start := time.Now()

	// 1) 撤销挂单
	openOrders, _ := ex.GetOpenOrders(ctx, robot.Symbol)
	for _, o := range openOrders {
		if o == nil || o.OrderId == "" {
			continue
		}
		_, _ = ex.CancelOrder(ctx, robot.Symbol, o.OrderId)
	}

	// 2) 平掉所有持仓
	positions, err := ex.GetPositions(ctx, robot.Symbol)
	if err != nil {
		return err
	}
	for _, pos := range positions {
		if pos == nil || math.Abs(pos.PositionAmt) <= positionAmtEpsilon {
			continue
		}
		_, _ = ex.ClosePosition(ctx, robot.Symbol, pos.PositionSide, math.Abs(pos.PositionAmt))
	}

	// 3) 等待持仓归零（直到超时）
	deadline := start.Add(timeout)
	for time.Now().Before(deadline) {
		ps, _ := ex.GetPositions(ctx, robot.Symbol)
		any := false
		for _, p := range ps {
			if p != nil && math.Abs(p.PositionAmt) > positionAmtEpsilon {
				any = true
				break
			}
		}

		// 刷新引擎内存缓存
		e.mu.Lock()
		e.CurrentPositions = ps
		e.LastPositionUpdate = time.Now()
		e.mu.Unlock()

		if !any {
			return nil
		}
		time.Sleep(400 * time.Millisecond)
	}

	return gerror.New("全部平仓超时，仍检测到持仓未归零")
}


