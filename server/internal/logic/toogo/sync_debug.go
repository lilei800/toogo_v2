package toogo

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// ===== 订单/持仓同步诊断开关（选择性放开终端输出）=====
//
// 目标：
// - 默认保持 warning 级别 + 节流，避免终端洪流
// - 需要排障时，仅打开“订单/持仓同步”链路关键日志（仍带节流）
//
// 配置：
//   toogo:
//     debug:
//       orderPositionSync: true

var orderPositionSyncLogAt sync.Map // key: string -> time.Time

func isOrderPositionSyncDebugEnabled(ctx context.Context) bool {
	v, err := g.Cfg().Get(ctx, "toogo.debug.orderPositionSync")
	if err != nil {
		return false
	}
	return v.Bool()
}

func shouldLogOrderPositionSync(key string, every time.Duration) bool {
	if strings.TrimSpace(key) == "" {
		return false
	}
	now := time.Now()
	if v, ok := orderPositionSyncLogAt.Load(key); ok {
		if t0, ok2 := v.(time.Time); ok2 && now.Sub(t0) < every {
			return false
		}
	}
	orderPositionSyncLogAt.Store(key, now)
	return true
}
