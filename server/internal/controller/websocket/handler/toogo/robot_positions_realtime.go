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
	RobotPositionsRealtime = cRobotPositionsRealtime{}
)

// 说明：
// - 这是“按机器人ID订阅”的推送：前端传 robotIds（逗号分隔），服务端每 interval 推送一次 positions snapshot
// - 推送事件固定为：toogo/robot/positions/push
//
// 推送 payload:
// {
//   "list": [
//     { "robotId": 1, "list": [ ...positions... ], "error": "" },
//     { "robotId": 2, "list": [], "error": "xxx" }
//   ]
// }

type cRobotPositionsRealtime struct{}

type robotPositionsRealtimeSub struct {
	stopCh   chan struct{}
	robotIds string
	interval time.Duration
}

var robotPositionsRealtimeSubs sync.Map // key: client.ID, value: *robotPositionsRealtimeSub

// Subscribe 订阅机器人持仓实时推送
// req.Data:
// - robotIds: "1,2,3"
// - intervalMs: (optional) 推送间隔毫秒，默认1000
func (c *cRobotPositionsRealtime) Subscribe(client *websocket.Client, req *websocket.WRequest) {
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
	if v, ok := robotPositionsRealtimeSubs.Load(client.ID); ok {
		if sub, ok2 := v.(*robotPositionsRealtimeSub); ok2 && sub != nil {
			close(sub.stopCh)
		}
		robotPositionsRealtimeSubs.Delete(client.ID)
	}

	sub := &robotPositionsRealtimeSub{
		stopCh:   make(chan struct{}),
		robotIds: robotIds,
		interval: interval,
	}
	robotPositionsRealtimeSubs.Store(client.ID, sub)

	// ack
	websocket.SendSuccess(client, req.Event, g.Map{
		"robotIds":    robotIds,
		"intervalMs":  intervalMs,
		"pushEvent":   "toogo/robot/positions/push",
		"description": "已订阅机器人持仓实时推送（positions snapshot）",
	})

	// 立即推一次 + 定时推
	go func() {
		// 订阅维度缓存：避免“获取失败时推空数组”导致前端把已有持仓清掉产生闪烁/丢失
		// - key: robotId
		// - value: 上一次成功返回的 positions snapshot（尽量保持引用隔离，避免外部修改）
		lastOk := make(map[int64]any)

		push := func() {
			// 客户端断开/超时后停止
			if client == nil || client.SendClose || !websocket.Manager().InClient(client) {
				return
			}

			idStrs := strings.Split(sub.robotIds, ",")
			items := make([]g.Map, 0, len(idStrs))

			for _, idStr := range idStrs {
				robotId, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
				if err != nil || robotId <= 0 {
					continue
				}

				ctx := client.Context()
				if ctx == nil {
					ctx = context.Background()
				}
				list, err := service.ToogoRobot().GetRobotPositions(ctx, robotId)
				item := g.Map{
					"robotId": robotId,
					"list":    list,
					"error":   "",
				}
				if err != nil {
					// 失败时：优先复用上一帧成功结果，避免前端闪烁/误判为“已平仓”
					if prev, ok := lastOk[robotId]; ok && prev != nil {
						item["list"] = prev
						item["stale"] = true
					} else {
						item["list"] = g.Slice{}
					}
					item["error"] = err.Error()
				} else if list == nil {
					// 统一输出空数组，避免前端收到 null 误判为“无数据/未推送”
					item["list"] = g.Slice{}
					lastOk[robotId] = item["list"]
				} else {
					// 成功：记录最新快照（注意：这里存的是当前返回的 slice 引用；后续如需更强隔离可深拷贝）
					lastOk[robotId] = list
				}
				items = append(items, item)
			}

			websocket.SendSuccess(client, "toogo/robot/positions/push", g.Map{
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
				// 客户端断开/超时后停止
				if client == nil || client.SendClose || !websocket.Manager().InClient(client) {
					return
				}
				push()
			}
		}
	}()
}

// Unsubscribe 取消订阅
func (c *cRobotPositionsRealtime) Unsubscribe(client *websocket.Client, req *websocket.WRequest) {
	if v, ok := robotPositionsRealtimeSubs.Load(client.ID); ok {
		if sub, ok2 := v.(*robotPositionsRealtimeSub); ok2 && sub != nil {
			close(sub.stopCh)
		}
		robotPositionsRealtimeSubs.Delete(client.ID)
	}
	websocket.SendSuccess(client, req.Event, g.Map{"ok": true})
}


