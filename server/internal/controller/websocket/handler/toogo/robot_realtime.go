// Package toogo WebSocket handlers for toogo realtime data.
package toogo

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	tradingLogic "hotgo/internal/logic/trading"
	"hotgo/internal/websocket"
)

var (
	RobotRealtime = cRobotRealtime{}
)

// 说明：
// - 这是“按机器人ID订阅”的推送：前端传 robotIds（逗号分隔），服务端每 interval 推送一次 batchRobotAnalysis 的结果
// - 推送事件固定为：toogo/robot/realtime/push

type cRobotRealtime struct{}

type robotRealtimeSub struct {
	stopCh   chan struct{}
	robotIds string
	interval time.Duration
}

var robotRealtimeSubs sync.Map // key: client.ID, value: *robotRealtimeSub

// Subscribe 订阅机器人实时数据推送
// req.Data:
// - robotIds: "1,2,3"
// - intervalMs: (optional) 推送间隔毫秒，默认1000
func (c *cRobotRealtime) Subscribe(client *websocket.Client, req *websocket.WRequest) {
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
	if v, ok := robotRealtimeSubs.Load(client.ID); ok {
		if sub, ok2 := v.(*robotRealtimeSub); ok2 && sub != nil {
			close(sub.stopCh)
		}
		robotRealtimeSubs.Delete(client.ID)
	}

	sub := &robotRealtimeSub{
		stopCh:   make(chan struct{}),
		robotIds: robotIds,
		interval: interval,
	}
	robotRealtimeSubs.Store(client.ID, sub)

	g.Log().Warningf(client.Context(), "[WS][RobotRealtime] subscribe ok: client=%s robotIds=%s intervalMs=%d event=%s", client.ID, robotIds, intervalMs, req.Event)

	// ack
	websocket.SendSuccess(client, req.Event, g.Map{
		"robotIds":    robotIds,
		"intervalMs":  intervalMs,
		"pushEvent":   "toogo/robot/realtime/push",
		"description": "已订阅机器人实时数据推送（批量实时分析）",
	})

	// 立即推一次 + 定时推
	go func() {
		push := func() {
			// 客户端断开/超时后停止
			if client == nil || client.SendClose || !websocket.Manager().InClient(client) {
				return
			}

			out, err := tradingLogic.Monitor.GetBatchRobotAnalysis(client.Context(), sub.robotIds)
			if err != nil {
				// 不中断，继续下一轮
				g.Log().Warningf(client.Context(), "[WS][RobotRealtime] push failed: client=%s robotIds=%s err=%v", client.ID, sub.robotIds, err)
				return
			}
			websocket.SendSuccess(client, "toogo/robot/realtime/push", out)
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
func (c *cRobotRealtime) Unsubscribe(client *websocket.Client, req *websocket.WRequest) {
	if v, ok := robotRealtimeSubs.Load(client.ID); ok {
		if sub, ok2 := v.(*robotRealtimeSub); ok2 && sub != nil {
			close(sub.stopCh)
		}
		robotRealtimeSubs.Delete(client.ID)
	}
	websocket.SendSuccess(client, req.Event, g.Map{"ok": true})
}
