package exchange

import (
	"context"
	"net"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

// PrivateEventType 私有流事件类型（订单/持仓/账户）
type PrivateEventType string

const (
	PrivateEventUnknown  PrivateEventType = "unknown"
	PrivateEventOrder    PrivateEventType = "order"
	PrivateEventPosition PrivateEventType = "position"
	PrivateEventAccount  PrivateEventType = "account"
)

// PrivateEvent 私有WS事件（标准化后的最小信息）
type PrivateEvent struct {
	Platform    string           `json:"platform"`
	ApiConfigId int64            `json:"apiConfigId"`
	Type        PrivateEventType `json:"type"`
	Symbol      string           `json:"symbol"`
	Raw         []byte           `json:"raw"`
	ReceivedAt  int64            `json:"receivedAt"`
}

// PrivateStream 私有WS流（订单/持仓/余额变更）
// 设计目标：统一接口 + 事件驱动，上层按事件触发轻量同步，轮询仅做最终一致性兜底。
type PrivateStream interface {
	Start(ctx context.Context) error
	Stop()
	IsRunning() bool

	// AddSymbol / RemoveSymbol：部分交易所私有流需要按 symbol 订阅（如 Gate），没有需要的可忽略
	AddSymbol(symbol string) error
	RemoveSymbol(symbol string) error

	SetProxyDialer(dialer func(network, addr string) (net.Conn, error))
	SetOnEvent(cb func(ev *PrivateEvent))
}

// PrivateStreamStatusProvider is an optional interface implemented by some private streams
// to expose health/status signals for upper-layer reconciliation logic (e.g. WS silence fallback).
//
// NOTE: This is intentionally optional to avoid forcing every exchange implementation to support it.
type PrivateStreamStatusProvider interface {
	// LastMessageAt returns last time ANY WS message was received (including ping/pong/ack/data).
	LastMessageAt() time.Time
	// LastEventAt returns last time a business event was emitted to upper layer (order/position/account).
	LastEventAt() time.Time
}

// NewPrivateStream 创建私有流实例（按交易所平台）
func NewPrivateStream(cfg *Config) (PrivateStream, error) {
	if cfg == nil {
		return nil, gerror.New("exchange config is nil")
	}
	switch cfg.Platform {
	case "binance":
		return NewBinancePrivateStream(cfg), nil
	case "okx":
		return NewOKXPrivateStream(cfg), nil
	case "gate":
		return NewGatePrivateStream(cfg), nil
	default:
		return nil, gerror.Newf("unsupported exchange: %s", cfg.Platform)
	}
}
