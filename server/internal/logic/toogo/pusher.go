// Package toogo Toogo实时推送服务
package toogo

import (
	"context"
	"strings"
	"sync"
	"time"

	"hotgo/internal/dao"
	"hotgo/internal/library/market"
	"hotgo/internal/library/exchange"
	"hotgo/internal/library/websocket"
	"hotgo/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// Pusher 实时推送服务
type Pusher struct {
	exchanges map[string]exchange.Exchange // 交易所实例缓存

	// tickerRESTFallback 用于兜底 REST 拉取 ticker 的限流（避免秒级推送导致平台限流）
	tickerRESTFallbackMu sync.Mutex
	tickerRESTFallbackAt map[string]time.Time // key: platform:symbol
}

// MarketSignal 市场信号（用于推送）
type MarketSignal struct {
	Direction   string  // 方向 LONG/SHORT/NEUTRAL
	Strength    float64 // 信号强度 0-100
	MarketState string  // 市场状态 trend/volatile/high_vol/low_vol
	RiskLevel   string  // 风险等级 low/medium/high
	Reason      string  // 信号原因
}

var pusherInstance *Pusher

// GetPusher 获取推送服务单例
func GetPusher() *Pusher {
	if pusherInstance == nil {
		pusherInstance = &Pusher{
			exchanges: make(map[string]exchange.Exchange),
			tickerRESTFallbackAt: make(map[string]time.Time),
		}
	}
	return pusherInstance
}

// PushTicker 推送行情数据
func (p *Pusher) PushTicker(ctx context.Context, symbols []string) {
	for _, symbol := range symbols {
		// 获取各交易所行情
		for _, platform := range []string{"binance", "okx", "bitget", "gate"} {
			channel := "ticker:" + platform + ":" + symbol
			// 没人订阅该频道则不做任何工作（避免无意义循环）
			if websocket.GetHub().GetChannelSubscribers(channel) <= 0 {
				continue
			}

			ex := p.getExchange(platform)
			if ex == nil {
				continue
			}

			// 优先用全局行情缓存（WS优先，降级HTTP缓存），避免每秒REST拉行情导致限流
			// 对于被订阅的频道，这里顺带订阅一次，让缓存能持续更新（引用计数在 market 内部管理）
			market.GetMarketServiceManager().Subscribe(ctx, platform, symbol, ex)
			ticker := market.GetMarketServiceManager().GetTicker(platform, symbol)
			if ticker == nil {
				// 缓存缺失时，才做低频REST兜底（platform:symbol 维度限流）
				if !p.allowTickerRESTFallback(platform, symbol, 10*time.Second) {
					continue
				}
				apiTicker, err := ex.GetTicker(ctx, symbol)
			if err != nil {
				continue
				}
				ticker = apiTicker
			}

			// 推送到对应频道
			websocket.GetHub().SendToChannel(channel, &websocket.Message{
				Type:      websocket.MsgTypeTicker,
				Channel:   channel,
				Data:      ticker,
				Timestamp: time.Now().UnixMilli(),
			})
		}
	}
}

type tickerSubscription struct {
	channel  string
	platform string
	symbol   string
}

// getTickerSubscriptionsFromHub 从 Hub 的订阅频道中提取 ticker 订阅
// 支持两种格式：
// 1) ticker:BTCUSDT                 -> 默认平台（bitget）
// 2) ticker:bitget:BTCUSDT          -> 指定平台
func (p *Pusher) getTickerSubscriptionsFromHub(ctx context.Context) []tickerSubscription {
	hub := websocket.GetHub()
	channels := hub.GetAllChannels()
	if len(channels) == 0 {
		return nil
	}

	const defaultPlatform = "bitget"
	out := make([]tickerSubscription, 0, len(channels))
	for _, ch := range channels {
		if !strings.HasPrefix(ch, "ticker:") {
			continue
		}
		// 频道没有订阅者就跳过（channelClients 可能会残留空map）
		if hub.GetChannelSubscribers(ch) <= 0 {
			continue
		}

		parts := strings.Split(ch, ":")
		// ticker:symbol
		if len(parts) == 2 && parts[1] != "" {
			out = append(out, tickerSubscription{
				channel:  ch,
				platform: defaultPlatform,
				symbol:   parts[1],
			})
			continue
		}
		// ticker:platform:symbol
		if len(parts) == 3 && parts[1] != "" && parts[2] != "" {
			out = append(out, tickerSubscription{
				channel:  ch,
				platform: parts[1],
				symbol:   parts[2],
			})
			continue
		}
		// 其他格式忽略
		g.Log().Debugf(ctx, "[Pusher] 忽略未知ticker频道格式: %s", ch)
	}
	return out
}

// pushSubscribedTickers 按实际订阅驱动推送行情
func (p *Pusher) pushSubscribedTickers(ctx context.Context) int {
	subs := p.getTickerSubscriptionsFromHub(ctx)
	if len(subs) == 0 {
		return 0
	}

	// 去重：同一个 platform:symbol 只拉取/读取一次，然后推送到多个频道
	type key struct {
		platform string
		symbol   string
	}
	channelsByKey := make(map[key][]string, len(subs))
	for _, s := range subs {
		k := key{platform: s.platform, symbol: s.symbol}
		channelsByKey[k] = append(channelsByKey[k], s.channel)
	}

	for k, channels := range channelsByKey {
		ex := p.getExchange(k.platform)
		if ex == nil {
			continue
		}

		// 让 market 缓存开始工作（WS优先/HTTP缓存降级）
		market.GetMarketServiceManager().Subscribe(ctx, k.platform, k.symbol, ex)

		ticker := market.GetMarketServiceManager().GetTicker(k.platform, k.symbol)
		if ticker == nil {
			// 缓存缺失，低频REST兜底
			if !p.allowTickerRESTFallback(k.platform, k.symbol, 10*time.Second) {
				continue
			}
			apiTicker, err := ex.GetTicker(ctx, k.symbol)
			if err != nil {
				continue
			}
			ticker = apiTicker
		}

		for _, ch := range channels {
			websocket.GetHub().SendToChannel(ch, &websocket.Message{
				Type:      websocket.MsgTypeTicker,
				Channel:   ch,
				Data:      ticker,
				Timestamp: time.Now().UnixMilli(),
			})
		}
	}

	return len(subs)
}

// calcTickerPushInterval 根据订阅量动态调整推送频率
func (p *Pusher) calcTickerPushInterval(subCount int) time.Duration {
	// 无订阅：不需要高频扫描
	if subCount <= 0 {
		return 5 * time.Second
	}
	// 订阅少：1s 推送体验最好（数据来自缓存，压力很小）
	if subCount <= 20 {
		return 1 * time.Second
	}
	// 订阅多：稍微降频，降低网络与CPU开销
	if subCount <= 100 {
		return 2 * time.Second
	}
	return 3 * time.Second
}

// PushRobotStatus 推送机器人状态
func (p *Pusher) PushRobotStatus(ctx context.Context, robot *entity.TradingRobot) {
	if robot == nil {
		return
	}

	// 【优化】使用引擎缓存的持仓数据，避免调用API
	var positions []*exchange.Position
	var currentPnl float64
	if engine := GetRobotTaskManager().GetEngine(robot.Id); engine != nil {
		engine.mu.RLock()
		// 使用缓存的持仓数据
		positions = engine.CurrentPositions
		// 计算当前盈亏
		for _, pos := range positions {
			currentPnl += pos.UnrealizedPnl
		}
		engine.mu.RUnlock()
	}

	statusText := "未启动"
	switch robot.Status {
	case 1:
		statusText = "未启动"
	case 2:
		statusText = "运行中"
	case 3:
		statusText = "已暂停"
	case 4:
		statusText = "已停用"
	}

	data := map[string]interface{}{
		"robotId":     robot.Id,
		"name":        robot.RobotName,
		"symbol":      robot.Symbol,
		"status":      robot.Status,
		"statusText":  statusText,
		"totalProfit": robot.TotalProfit,
		"currentPnl":  currentPnl,
		"positions":   positions,
		"maxProfit":   robot.MaxProfitTarget,
		"maxLoss":     robot.MaxLossAmount,
	}

	// 推送给用户
	websocket.GetHub().SendToUser(robot.UserId, &websocket.Message{
		Type:      websocket.MsgTypeRobot,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	})
}

// PushPosition 推送持仓更新
func (p *Pusher) PushPosition(ctx context.Context, userId int64, positions []*exchange.Position) {
	websocket.GetHub().SendToUser(userId, &websocket.Message{
		Type:      websocket.MsgTypePosition,
		Data:      positions,
		Timestamp: time.Now().UnixMilli(),
	})
}

// PushOrder 推送订单更新
func (p *Pusher) PushOrder(ctx context.Context, userId int64, order *exchange.Order) {
	websocket.GetHub().SendToUser(userId, &websocket.Message{
		Type:      websocket.MsgTypeOrder,
		Data:      order,
		Timestamp: time.Now().UnixMilli(),
	})
}

// PushSignal 推送交易信号
func (p *Pusher) PushSignal(ctx context.Context, userId int64, signal *MarketSignal, robotId int64) {
	data := map[string]interface{}{
		"robotId":     robotId,
		"direction":   signal.Direction,
		"strength":    signal.Strength,
		"marketState": signal.MarketState,
		"riskLevel":   signal.RiskLevel,
		"reason":      signal.Reason,
	}

	websocket.GetHub().SendToUser(userId, &websocket.Message{
		Type:      websocket.MsgTypeSignal,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	})
}

// PushPnL 推送盈亏更新
func (p *Pusher) PushPnL(ctx context.Context, userId int64, robotId int64, realizedPnl, unrealizedPnl float64) {
	data := map[string]interface{}{
		"robotId":       robotId,
		"realizedPnl":   realizedPnl,
		"unrealizedPnl": unrealizedPnl,
		"totalPnl":      realizedPnl + unrealizedPnl,
	}

	websocket.GetHub().SendToUser(userId, &websocket.Message{
		Type:      websocket.MsgTypePnL,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	})
}

// PushSystemNotice 推送系统通知
func (p *Pusher) PushSystemNotice(ctx context.Context, userId int64, title, content string, level string) {
	data := map[string]interface{}{
		"title":   title,
		"content": content,
		"level":   level, // info, warning, error, success
	}

	if userId > 0 {
		// 推送给指定用户
		websocket.GetHub().SendToUser(userId, &websocket.Message{
			Type:      websocket.MsgTypeSystem,
			Data:      data,
			Timestamp: time.Now().UnixMilli(),
		})
	} else {
		// 广播给所有用户
		websocket.GetHub().Broadcast(&websocket.Message{
			Type:      websocket.MsgTypeSystem,
			Data:      data,
			Timestamp: time.Now().UnixMilli(),
		})
	}
}

// PushError 推送错误消息
func (p *Pusher) PushError(ctx context.Context, userId int64, code int, message string) {
	data := map[string]interface{}{
		"code":    code,
		"message": message,
	}

	websocket.GetHub().SendToUser(userId, &websocket.Message{
		Type:      websocket.MsgTypeError,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	})
}

// getExchange 获取交易所实例
func (p *Pusher) getExchange(platform string) exchange.Exchange {
	if ex, ok := p.exchanges[platform]; ok {
		return ex
	}

	// 创建公共API实例(无需认证)
	ex, err := exchange.NewExchange(&exchange.Config{
		Platform: platform,
	})
	if err != nil {
		return nil
	}

	p.exchanges[platform] = ex
	return ex
}

// StartTickerPusher 启动行情推送定时任务
func StartTickerPusher(ctx context.Context) {
	g.Log().Info(ctx, "[Pusher] 行情推送服务启动")

	pusher := GetPusher()
	for {
		// 无在线用户直接降频休眠
		if websocket.GetHub().GetOnlineCount() <= 0 {
			select {
			case <-ctx.Done():
				g.Log().Info(ctx, "[Pusher] 行情推送服务停止")
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

		subCount := pusher.pushSubscribedTickers(ctx)
		interval := pusher.calcTickerPushInterval(subCount)
		select {
		case <-ctx.Done():
			g.Log().Info(ctx, "[Pusher] 行情推送服务停止")
			return
		case <-time.After(interval):
		}
	}
}

// StartRobotStatusPusher 启动机器人状态推送
func StartRobotStatusPusher(ctx context.Context) {
	g.Log().Info(ctx, "[Pusher] 机器人状态推送服务启动")

	// 每3秒推送一次机器人状态
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	pusher := GetPusher()
	for {
		select {
		case <-ctx.Done():
			g.Log().Info(ctx, "[Pusher] 机器人状态推送服务停止")
			return
		case <-ticker.C:
			// 无在线用户则跳过（避免频繁扫库）
			if websocket.GetHub().GetOnlineCount() <= 0 {
				continue
			}
			// 查询所有运行中的机器人
			var robots []*entity.TradingRobot
			_ = dao.TradingRobot.Ctx(ctx).Where("status", 2).Scan(&robots)

			for _, robot := range robots {
				// 检查用户是否在线
				if websocket.GetHub().GetUserOnline(robot.UserId) {
					pusher.PushRobotStatus(ctx, robot)
				}
			}
		}
	}
}

// allowTickerRESTFallback 判断是否允许对某个 platform:symbol 执行一次REST兜底拉取
func (p *Pusher) allowTickerRESTFallback(platform, symbol string, minInterval time.Duration) bool {
	key := platform + ":" + symbol
	now := time.Now()

	p.tickerRESTFallbackMu.Lock()
	defer p.tickerRESTFallbackMu.Unlock()

	if last, ok := p.tickerRESTFallbackAt[key]; ok {
		if now.Sub(last) < minInterval {
			return false
		}
	}
	p.tickerRESTFallbackAt[key] = now
	return true
}

