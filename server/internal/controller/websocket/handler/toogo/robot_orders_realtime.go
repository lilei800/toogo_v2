// Package toogo WebSocket handlers for toogo realtime data.
package toogo

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	"hotgo/internal/service"
	"hotgo/internal/websocket"
)

var (
	RobotOrdersRealtime = cRobotOrdersRealtime{}
)

// 说明：
// - 这是“按机器人ID订阅”的推送：前端传 robotIds（逗号分隔），服务端每 interval 推送一次 openOrders snapshot（来自 DB 事实表）
// - 推送事件固定为：toogo/robot/orders/push
//
// 推送 payload:
// {
//   "list": [
//     { "robotId": 1, "list": [ ...orders... ], "error": "", "stale": false },
//     { "robotId": 2, "list": [], "error": "xxx", "stale": true }
//   ]
// }
//
// 设计目标：
// - WS 失败时不推空数组覆盖前端（避免“挂单闪烁/丢失”）
// - 以 DB 事实表为主（私有WS增量写入 + REST 低频兜底对账），减少直连交易所压力
type cRobotOrdersRealtime struct{}

type robotOrdersRealtimeSub struct {
	stopCh   chan struct{}
	robotIds string
	interval time.Duration
}

var robotOrdersRealtimeSubs sync.Map // key: client.ID, value: *robotOrdersRealtimeSub

// Subscribe 订阅机器人挂单实时推送
// req.Data:
// - robotIds: "1,2,3"
// - intervalMs: (optional) 推送间隔毫秒，默认1000
func (c *cRobotOrdersRealtime) Subscribe(client *websocket.Client, req *websocket.WRequest) {
	robotIds := strings.TrimSpace(gconv.String(req.Data["robotIds"]))
	if robotIds == "" {
		websocket.SendError(client, req.Event, errors.New("robotIds 不能为空"))
		return
	}

	intervalMs := gconv.Int(req.Data["intervalMs"])
	if intervalMs <= 0 {
		intervalMs = 1000
	}
	interval := time.Duration(intervalMs) * time.Millisecond

	// 重置旧订阅
	if v, ok := robotOrdersRealtimeSubs.Load(client.ID); ok {
		if sub, ok2 := v.(*robotOrdersRealtimeSub); ok2 && sub != nil {
			close(sub.stopCh)
		}
		robotOrdersRealtimeSubs.Delete(client.ID)
	}

	sub := &robotOrdersRealtimeSub{
		stopCh:   make(chan struct{}),
		robotIds: robotIds,
		interval: interval,
	}
	robotOrdersRealtimeSubs.Store(client.ID, sub)

	// ack
	websocket.SendSuccess(client, req.Event, g.Map{
		"robotIds":    robotIds,
		"intervalMs":  intervalMs,
		"pushEvent":   "toogo/robot/orders/push",
		"description": "已订阅机器人挂单实时推送（open orders snapshot）",
	})

	go func() {
		// 订阅维度缓存：避免“获取失败时推空数组”导致前端闪烁/误判为“已撤单/已成交”
		lastOk := make(map[int64]any)

		push := func() {
			if client == nil || client.SendClose || !websocket.Manager().InClient(client) {
				return
			}
			ctx := client.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			idStrs := strings.Split(sub.robotIds, ",")
			items := make([]g.Map, 0, len(idStrs))
			for _, idStr := range idStrs {
				robotId, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
				if err != nil || robotId <= 0 {
					continue
				}

				list, err := service.ToogoRobot().GetRobotOpenOrders(ctx, robotId)
				item := g.Map{
					"robotId": robotId,
					"list":    list,
					"error":   "",
					"stale":   false,
				}

				if err != nil {
					if prev, ok := lastOk[robotId]; ok && prev != nil {
						item["list"] = prev
						item["stale"] = true
					} else {
						item["list"] = g.Slice{}
						item["stale"] = true
					}
					item["error"] = err.Error()
				} else if list == nil {
					item["list"] = g.Slice{}
					lastOk[robotId] = item["list"]
				} else {
					lastOk[robotId] = list
				}

				items = append(items, item)
			}

			websocket.SendSuccess(client, "toogo/robot/orders/push", g.Map{
				"list": items,
			})
		}

		push()
		tk := time.NewTicker(sub.interval)
		defer tk.Stop()
		for {
			select {
			case <-sub.stopCh:
				return
			case <-tk.C:
				if client == nil || client.SendClose || !websocket.Manager().InClient(client) {
					return
				}
				push()
			}
		}
	}()
}

// Unsubscribe 取消订阅
func (c *cRobotOrdersRealtime) Unsubscribe(client *websocket.Client, req *websocket.WRequest) {
	if v, ok := robotOrdersRealtimeSubs.Load(client.ID); ok {
		if sub, ok2 := v.(*robotOrdersRealtimeSub); ok2 && sub != nil {
			close(sub.stopCh)
		}
		robotOrdersRealtimeSubs.Delete(client.ID)
	}
	websocket.SendSuccess(client, req.Event, g.Map{"ok": true})
}


