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

const BitgetWSPrivateURL = "wss://ws.bitget.com/v2/ws/private"

// BitgetPrivateStream Bitget 私有WS（orders/positions/account）
// 说明：频道名称/字段会随版本略有差异，这里按 Bitget v2/ws/private 常见格式实现。
type BitgetPrivateStream struct {
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

func NewBitgetPrivateStream(cfg *Config) *BitgetPrivateStream {
	return &BitgetPrivateStream{
		cfg:     cfg,
		symbols: make(map[string]int),
	}
}

func (s *BitgetPrivateStream) SetProxyDialer(dialer func(network, addr string) (net.Conn, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.proxyDialer = dialer
}

func (s *BitgetPrivateStream) SetOnEvent(cb func(ev *PrivateEvent)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onEvent = cb
}

func (s *BitgetPrivateStream) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running && s.conn != nil && s.conn.IsConnected()
}

func (s *BitgetPrivateStream) AddSymbol(symbol string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil
	}
	s.symbols[symbol]++
	if s.conn != nil && s.conn.IsConnected() && s.loggedIn {
		s.subscribeSymbolLocked(symbol)
	}
	return nil
}

func (s *BitgetPrivateStream) RemoveSymbol(symbol string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if n, ok := s.symbols[symbol]; ok {
		n--
		if n <= 0 {
			delete(s.symbols, symbol)
			if s.conn != nil && s.conn.IsConnected() && s.loggedIn {
				s.unsubscribeSymbolLocked(symbol)
			}
		} else {
			s.symbols[symbol] = n
		}
	}
	return nil
}

func (s *BitgetPrivateStream) Start(ctx context.Context) error {
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
	cfg.URL = BitgetWSPrivateURL
	cfg.PingInterval = 25 * time.Second
	cfg.ProxyDialer = proxyDialer

	s.conn = NewWebSocketConnection(cfg)
	s.conn.SetCallbacks(s.onMessage, s.onConnected, s.onDisconnected)
	if err := s.conn.Connect(s.ctx); err != nil {
		s.Stop()
		return err
	}
	g.Log().Info(s.ctx, "[BitgetPrivateWS] started")
	return nil
}

func (s *BitgetPrivateStream) Stop() {
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

func (s *BitgetPrivateStream) onConnected() {
	_ = s.login()
}

func (s *BitgetPrivateStream) onDisconnected(err error) {
	g.Log().Warningf(s.ctx, "[BitgetPrivateWS] disconnected: %v", err)
}

func (s *BitgetPrivateStream) emit(tp PrivateEventType, symbol string, raw []byte) {
	s.mu.RLock()
	cb := s.onEvent
	s.mu.RUnlock()
	if cb == nil {
		return
	}
	cb(&PrivateEvent{
		Platform:   "bitget",
		Type:       tp,
		Symbol:     symbol,
		Raw:        raw,
		ReceivedAt: time.Now().UnixMilli(),
	})
}

func (s *BitgetPrivateStream) login() error {
	s.mu.RLock()
	conn := s.conn
	s.mu.RUnlock()
	if conn == nil {
		return nil
	}

	ts := time.Now().Unix()
	tsStr := g.NewVar(ts).String()
	// 常见 bitget ws login preSign: ts + "GET" + "/user/verify"
	preSign := tsStr + "GET" + "/user/verify"
	mac := hmac.New(sha256.New, []byte(s.cfg.SecretKey))
	mac.Write([]byte(preSign))
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

func (s *BitgetPrivateStream) subscribeSymbolLocked(symbol string) {
	if s.conn == nil {
		return
	}
	// 使用统一的Symbol格式化器，确保格式正确
	instId := Formatter.FormatForBitget(symbol) // BTCUSDT (不带任何后缀)
	
	// Bitget v2私有WS：订阅订单和持仓（USDT-FUTURES）
	// 注意：instId格式为 BTCUSDT，不需要 _UMCBL 后缀
	_ = s.conn.Send(map[string]any{
		"op": "subscribe",
		"args": []map[string]string{
			{"instType": "USDT-FUTURES", "channel": "orders", "instId": instId},
			{"instType": "USDT-FUTURES", "channel": "positions", "instId": instId},
		},
	})
}

func (s *BitgetPrivateStream) unsubscribeSymbolLocked(symbol string) {
	if s.conn == nil {
		return
	}
	// 使用统一的Symbol格式化器
	instId := Formatter.FormatForBitget(symbol) // BTCUSDT
	
	_ = s.conn.Send(map[string]any{
		"op": "unsubscribe",
		"args": []map[string]string{
			{"instType": "USDT-FUTURES", "channel": "orders", "instId": instId},
			{"instType": "USDT-FUTURES", "channel": "positions", "instId": instId},
		},
	})
}

func (s *BitgetPrivateStream) subscribeAllLocked() {
	for sym := range s.symbols {
		s.subscribeSymbolLocked(sym)
	}
}

func (s *BitgetPrivateStream) onMessage(msg []byte) {
	var data map[string]any
	if err := json.Unmarshal(msg, &data); err != nil {
		return
	}

	// event responses
	if ev, ok := data["event"].(string); ok {
		if ev == "login" {
			// code 可能是 string 或 number（你日志里是 {"code":0}）
			loginOK := false
			switch v := data["code"].(type) {
			case string:
				loginOK = (v == "0" || v == "00000")
			case float64:
				loginOK = (v == 0)
			case int:
				loginOK = (v == 0)
			case int64:
				loginOK = (v == 0)
			default:
				// 有些版本可能用 "success":true
				if b, ok2 := data["success"].(bool); ok2 {
					loginOK = b
				}
			}

			if loginOK {
				s.mu.Lock()
				s.loggedIn = true
				s.mu.Unlock()
				s.mu.Lock()
				s.subscribeAllLocked()
				s.mu.Unlock()
			} else {
				g.Log().Warningf(s.ctx, "[BitgetPrivateWS] login failed: %s", string(msg))
			}
			return
		}
		if ev == "subscribe" || ev == "unsubscribe" {
			return
		}
		if ev == "error" {
			g.Log().Warningf(s.ctx, "[BitgetPrivateWS] error msg: %s", string(msg))
			return
		}
	}

	// data push: arg.channel
	arg, _ := data["arg"].(map[string]any)
	if arg == nil {
		return
	}
	ch, _ := arg["channel"].(string)
	instId, _ := arg["instId"].(string)
	symbol := strings.ToUpper(strings.TrimSpace(instId))
	// 归一化：BTCUSDT_UMCBL -> BTCUSDT（方便与系统内部订阅/缓存 key 对齐）
	if idx := strings.Index(symbol, "_"); idx > 0 {
		symbol = symbol[:idx]
	}
	switch ch {
	case "orders":
		s.emit(PrivateEventOrder, symbol, msg)
	case "positions":
		s.emit(PrivateEventPosition, symbol, msg)
	case "account":
		s.emit(PrivateEventAccount, "", msg)
	}
}


