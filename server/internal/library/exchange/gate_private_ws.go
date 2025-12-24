package exchange

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Gate futures(usdt) 私有WS与公共WS同域：
// wss://fx-ws.gateio.ws/v4/ws/usdt
//
// 认证参考 Gate WS v4 常见签名方式：
// SIGN = HMAC_SHA512(secret, "channel=<channel>&event=<event>&time=<time>")
// 请求带 auth: { "method":"api_key", "KEY":apiKey, "SIGN":sign }
type GatePrivateStream struct {
	mu sync.RWMutex

	cfg         *Config
	proxyDialer func(network, addr string) (net.Conn, error)
	conn        *WebSocketConnection
	ctx         context.Context
	cancel      context.CancelFunc
	running     bool

	symbols map[string]int // contract: BTC_USDT
	onEvent func(ev *PrivateEvent)
}

func NewGatePrivateStream(cfg *Config) *GatePrivateStream {
	return &GatePrivateStream{
		cfg:     cfg,
		symbols: make(map[string]int),
	}
}

func (s *GatePrivateStream) SetProxyDialer(dialer func(network, addr string) (net.Conn, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.proxyDialer = dialer
}

func (s *GatePrivateStream) SetOnEvent(cb func(ev *PrivateEvent)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onEvent = cb
}

func (s *GatePrivateStream) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running && s.conn != nil && s.conn.IsConnected()
}

func (s *GatePrivateStream) AddSymbol(symbol string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	contract := gateFormatContract(symbol)
	if contract == "" {
		return nil
	}
	s.symbols[contract]++
	if s.conn != nil && s.conn.IsConnected() {
		s.subscribeLocked(contract)
	}
	return nil
}

func (s *GatePrivateStream) RemoveSymbol(symbol string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	contract := gateFormatContract(symbol)
	if n, ok := s.symbols[contract]; ok {
		n--
		if n <= 0 {
			delete(s.symbols, contract)
			if s.conn != nil && s.conn.IsConnected() {
				s.unsubscribeLocked(contract)
			}
		} else {
			s.symbols[contract] = n
		}
	}
	return nil
}

func (s *GatePrivateStream) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.ctx, s.cancel = context.WithCancel(ctx)
	proxyDialer := s.proxyDialer
	s.mu.Unlock()

	cfg := DefaultWebSocketConfig()
	cfg.URL = GateWSFuturesUSDTURL
	cfg.PingInterval = 20 * time.Second
	cfg.ProxyDialer = proxyDialer

	s.conn = NewWebSocketConnection(cfg)
	s.conn.SetCallbacks(s.onMessage, s.onConnected, s.onDisconnected)
	if err := s.conn.Connect(s.ctx); err != nil {
		s.Stop()
		return err
	}
	g.Log().Info(s.ctx, "[GatePrivateWS] started")
	return nil
}

func (s *GatePrivateStream) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	cancel := s.cancel
	conn := s.conn
	s.cancel = nil
	s.conn = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if conn != nil {
		conn.Disconnect()
	}
}

func (s *GatePrivateStream) onConnected() {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 重连后恢复订阅
	for contract := range s.symbols {
		s.subscribeLocked(contract)
	}
}

func (s *GatePrivateStream) onDisconnected(err error) {
	g.Log().Warningf(s.ctx, "[GatePrivateWS] disconnected: %v", err)
}

func (s *GatePrivateStream) emit(tp PrivateEventType, symbol string, raw []byte) {
	s.mu.RLock()
	cb := s.onEvent
	s.mu.RUnlock()
	if cb == nil {
		return
	}
	cb(&PrivateEvent{
		Platform:   "gate",
		Type:       tp,
		Symbol:     symbol,
		Raw:        raw,
		ReceivedAt: time.Now().UnixMilli(),
	})
}

func (s *GatePrivateStream) sign(channel, event string, ts int64) string {
	// channel=<channel>&event=<event>&time=<ts>
	payload := "channel=" + channel + "&event=" + event + "&time=" + g.NewVar(ts).String()
	m := hmac.New(sha512.New, []byte(s.cfg.SecretKey))
	m.Write([]byte(payload))
	return hex.EncodeToString(m.Sum(nil))
}

func (s *GatePrivateStream) auth(channel, event string, ts int64) map[string]string {
	return map[string]string{
		"method": "api_key",
		"KEY":    s.cfg.ApiKey,
		"SIGN":   s.sign(channel, event, ts),
	}
}

func (s *GatePrivateStream) subscribeLocked(contract string) {
	if s.conn == nil {
		return
	}
	ts := time.Now().Unix()
	// orders
	_ = s.conn.Send(map[string]any{
		"time":    ts,
		"channel": "futures.orders",
		"event":   "subscribe",
		"payload": []string{contract},
		"auth":    s.auth("futures.orders", "subscribe", ts),
	})
	// positions
	_ = s.conn.Send(map[string]any{
		"time":    ts,
		"channel": "futures.positions",
		"event":   "subscribe",
		"payload": []string{contract},
		"auth":    s.auth("futures.positions", "subscribe", ts),
	})
}

func (s *GatePrivateStream) unsubscribeLocked(contract string) {
	if s.conn == nil {
		return
	}
	ts := time.Now().Unix()
	_ = s.conn.Send(map[string]any{
		"time":    ts,
		"channel": "futures.orders",
		"event":   "unsubscribe",
		"payload": []string{contract},
		"auth":    s.auth("futures.orders", "unsubscribe", ts),
	})
	_ = s.conn.Send(map[string]any{
		"time":    ts,
		"channel": "futures.positions",
		"event":   "unsubscribe",
		"payload": []string{contract},
		"auth":    s.auth("futures.positions", "unsubscribe", ts),
	})
}

func (s *GatePrivateStream) onMessage(msg []byte) {
	var data map[string]any
	if err := json.Unmarshal(msg, &data); err != nil {
		return
	}

	channel, _ := data["channel"].(string)
	if channel == "" {
		return
	}
	if ev, ok := data["event"].(string); ok && (ev == "subscribe" || ev == "unsubscribe") {
		return
	}
	if ev, ok := data["event"].(string); ok && ev == "error" {
		g.Log().Warningf(s.ctx, "[GatePrivateWS] error msg: %s", string(msg))
		return
	}

	// best-effort symbol derivation from result.contract
	symbol := ""
	if result, ok := data["result"].(map[string]any); ok {
		if c, ok := result["contract"].(string); ok {
			symbol = gateNormalizeSymbol(c)
		}
	}

	switch channel {
	case "futures.orders":
		s.emit(PrivateEventOrder, symbol, msg)
	case "futures.positions":
		s.emit(PrivateEventPosition, symbol, msg)
	case "futures.account":
		s.emit(PrivateEventAccount, "", msg)
	}
}


