package exchange

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const OKXWSPrivateURL = "wss://ws.okx.com:8443/ws/v5/private"

// OKXPrivateStream OKX v5 私有WS（orders/positions/account）
type OKXPrivateStream struct {
	mu sync.RWMutex

	cfg         *Config
	proxyDialer func(network, addr string) (net.Conn, error)
	conn        *WebSocketConnection
	ctx         context.Context
	cancel      context.CancelFunc
	running     bool

	symbols map[string]int
	onEvent func(ev *PrivateEvent)

	loggedIn bool
}

func NewOKXPrivateStream(cfg *Config) *OKXPrivateStream {
	return &OKXPrivateStream{
		cfg:     cfg,
		symbols: make(map[string]int),
	}
}

func (s *OKXPrivateStream) SetProxyDialer(dialer func(network, addr string) (net.Conn, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.proxyDialer = dialer
}

func (s *OKXPrivateStream) SetOnEvent(cb func(ev *PrivateEvent)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onEvent = cb
}

func (s *OKXPrivateStream) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running && s.conn != nil && s.conn.IsConnected()
}

func (s *OKXPrivateStream) AddSymbol(symbol string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil
	}
	s.symbols[symbol]++
	return nil
}

func (s *OKXPrivateStream) RemoveSymbol(symbol string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if n, ok := s.symbols[symbol]; ok {
		n--
		if n <= 0 {
			delete(s.symbols, symbol)
		} else {
			s.symbols[symbol] = n
		}
	}
	return nil
}

func (s *OKXPrivateStream) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.loggedIn = false
	s.ctx, s.cancel = context.WithCancel(ctx)
	proxyDialer := s.proxyDialer
	s.mu.Unlock()

	cfg := DefaultWebSocketConfig()
	cfg.URL = OKXWSPrivateURL
	cfg.PingInterval = 25 * time.Second
	cfg.ProxyDialer = proxyDialer

	s.conn = NewWebSocketConnection(cfg)
	s.conn.SetCallbacks(s.onMessage, s.onConnected, s.onDisconnected)

	if err := s.conn.Connect(s.ctx); err != nil {
		s.Stop()
		return err
	}
	g.Log().Info(s.ctx, "[OKXPrivateWS] started")
	return nil
}

func (s *OKXPrivateStream) Stop() {
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
	s.loggedIn = false
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if conn != nil {
		conn.Disconnect()
	}
}

func (s *OKXPrivateStream) onConnected() {
	// login then subscribe
	// 【关键修复】私有WS登录对 timestamp 很敏感：先同步 serverTime offset，再登录
	_, _ = SyncServerTimeOffset(s.ctx, s.cfg)
	_ = s.login()
}

func (s *OKXPrivateStream) onDisconnected(err error) {
	g.Log().Warningf(s.ctx, "[OKXPrivateWS] disconnected: %v", err)
}

func (s *OKXPrivateStream) emit(tp PrivateEventType, symbol string, raw []byte) {
	s.mu.RLock()
	cb := s.onEvent
	s.mu.RUnlock()
	if cb == nil {
		return
	}
	cb(&PrivateEvent{
		Platform:   "okx",
		Type:       tp,
		Symbol:     symbol,
		Raw:        raw,
		ReceivedAt: time.Now().UnixMilli(),
	})
}

func (s *OKXPrivateStream) login() error {
	s.mu.RLock()
	conn := s.conn
	s.mu.RUnlock()
	if conn == nil {
		return nil
	}

	ts := nowSecWithOffset(s.cfg)
	tsStr := g.NewVar(ts).String()
	prehash := tsStr + "GET" + "/users/self/verify"
	mac := hmac.New(sha256.New, []byte(s.cfg.SecretKey))
	mac.Write([]byte(prehash))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	msg := map[string]any{
		"op": "login",
		"args": []map[string]string{
			{
				"apiKey":     s.cfg.ApiKey,
				"passphrase": s.cfg.Passphrase,
				"timestamp":  tsStr,
				"sign":       sign,
			},
		},
	}
	return conn.Send(msg)
}

func (s *OKXPrivateStream) subscribeAll() {
	s.mu.RLock()
	conn := s.conn
	s.mu.RUnlock()
	if conn == nil {
		return
	}

	// 订阅订单/持仓/账户（USDT-SWAP）
	msg := map[string]any{
		"op": "subscribe",
		"args": []map[string]string{
			{"channel": "orders", "instType": "SWAP"},
			{"channel": "positions", "instType": "SWAP"},
			{"channel": "account"},
		},
	}
	_ = conn.Send(msg)
}

func (s *OKXPrivateStream) onMessage(msg []byte) {
	var data map[string]any
	if err := json.Unmarshal(msg, &data); err != nil {
		return
	}

	// event responses
	if ev, ok := data["event"].(string); ok {
		if ev == "login" {
			if code, _ := data["code"].(string); code == "0" {
				s.mu.Lock()
				s.loggedIn = true
				s.mu.Unlock()
				s.subscribeAll()
			} else {
				g.Log().Warningf(s.ctx, "[OKXPrivateWS] login failed: %s", string(msg))
			}
			return
		}
		if ev == "subscribe" || ev == "unsubscribe" {
			return
		}
		if ev == "error" {
			g.Log().Warningf(s.ctx, "[OKXPrivateWS] error msg: %s", string(msg))
			// 常见：{"event":"error","msg":"Timestamp request expired","code":"60006"}
			if IsTimestampExpiredError(nil, string(msg)) {
				_, _ = SyncServerTimeOffset(s.ctx, s.cfg)
				_ = s.login()
			}
			return
		}
	}

	arg, _ := data["arg"].(map[string]any)
	if arg == nil {
		return
	}
	ch, _ := arg["channel"].(string)
	switch ch {
	case "orders":
		// symbol: instId
		sym, _ := arg["instId"].(string)
		s.emit(PrivateEventOrder, okxNormalizeSymbol(sym), msg)
	case "positions":
		sym, _ := arg["instId"].(string)
		s.emit(PrivateEventPosition, okxNormalizeSymbol(sym), msg)
	case "account":
		s.emit(PrivateEventAccount, "", msg)
	}
}


