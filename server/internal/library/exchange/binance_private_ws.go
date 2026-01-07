package exchange

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Binance USDT-M Futures User Data Stream
// - create listenKey: POST /fapi/v1/listenKey (X-MBX-APIKEY)
// - keepalive:        PUT  /fapi/v1/listenKey (listenKey, X-MBX-APIKEY)
// - close:            DELETE /fapi/v1/listenKey
//
// WS: wss://fstream.binance.com/ws/<listenKey>
type BinancePrivateStream struct {
	mu sync.RWMutex

	cfg         *Config
	apiKey      string
	endpoint    string
	wsBase      string
	listenKey   string
	httpClient  *http.Client
	proxyDialer func(network, addr string) (net.Conn, error)

	conn   *WebSocketConnection
	ctx    context.Context
	cancel context.CancelFunc
	running bool

	symbols map[string]int // 仅用于记录关注的symbol，上层按需筛选事件

	onEvent func(ev *PrivateEvent)
}

func NewBinancePrivateStream(cfg *Config) *BinancePrivateStream {
	endpoint := "https://fapi.binance.com"
	wsBase := "wss://fstream.binance.com"
	if cfg != nil && cfg.IsTestnet {
		endpoint = "https://testnet.binancefuture.com"
		wsBase = "wss://stream.binancefuture.com"
	}
	return &BinancePrivateStream{
		cfg:        cfg,
		apiKey:     cfg.ApiKey,
		endpoint:   endpoint,
		wsBase:     wsBase,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		symbols:    make(map[string]int),
	}
}

func (s *BinancePrivateStream) SetProxyDialer(dialer func(network, addr string) (net.Conn, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.proxyDialer = dialer
}

func (s *BinancePrivateStream) SetOnEvent(cb func(ev *PrivateEvent)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onEvent = cb
}

func (s *BinancePrivateStream) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running && s.conn != nil && s.conn.IsConnected()
}

func (s *BinancePrivateStream) AddSymbol(symbol string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil
	}
	s.symbols[symbol]++
	return nil
}

func (s *BinancePrivateStream) RemoveSymbol(symbol string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil
	}
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

func (s *BinancePrivateStream) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.mu.Unlock()

	// 1) create listenKey
	lk, err := s.createListenKey(s.ctx)
	if err != nil {
		s.Stop()
		return err
	}

	s.mu.Lock()
	s.listenKey = lk
	s.mu.Unlock()

	// 2) connect ws
	cfg := DefaultWebSocketConfig()
	cfg.URL = s.wsBase + "/ws/" + lk
	cfg.PingInterval = 3 * time.Minute
	cfg.ProxyDialer = s.proxyDialer

	s.conn = NewWebSocketConnection(cfg)
	s.conn.SetCallbacks(s.onMessage, s.onConnected, s.onDisconnected)

	if err := s.conn.Connect(s.ctx); err != nil {
		s.Stop()
		return err
	}

	// 3) keepalive loop
	go s.keepaliveLoop()

	g.Log().Infof(s.ctx, "[BinancePrivateWS] started: listenKey=%s", lk)
	return nil
}

func (s *BinancePrivateStream) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	cancel := s.cancel
	conn := s.conn
	listenKey := s.listenKey
	s.cancel = nil
	s.conn = nil
	s.listenKey = ""
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if conn != nil {
		conn.Disconnect()
	}
	if listenKey != "" {
		_ = s.closeListenKey(context.Background(), listenKey)
	}
}

func (s *BinancePrivateStream) onConnected() {
	// no-op
}

func (s *BinancePrivateStream) onDisconnected(err error) {
	g.Log().Warningf(s.ctx, "[BinancePrivateWS] disconnected: %v", err)
}

func (s *BinancePrivateStream) emit(tp PrivateEventType, symbol string, raw []byte) {
	s.mu.RLock()
	cb := s.onEvent
	s.mu.RUnlock()
	if cb == nil {
		return
	}
	cb(&PrivateEvent{
		Platform:   "binance",
		Type:       tp,
		Symbol:     symbol,
		Raw:        raw,
		ReceivedAt: time.Now().UnixMilli(),
	})
}

func (s *BinancePrivateStream) onMessage(msg []byte) {
	var m map[string]any
	if err := json.Unmarshal(msg, &m); err != nil {
		return
	}

	evt, _ := m["e"].(string)
	switch evt {
	case "ORDER_TRADE_UPDATE":
		// order update symbol: o.s
		symbol := ""
		if o, ok := m["o"].(map[string]any); ok {
			if sym, ok := o["s"].(string); ok {
				symbol = strings.ToUpper(sym)
			}
		}
		s.emit(PrivateEventOrder, symbol, msg)
	case "ACCOUNT_UPDATE":
		// account update contains positions under a.P
		// Binance 没有独立的 positions WS channel，这里把 ACCOUNT_UPDATE 解析为“按 symbol 的 Position 事件”，
		// 以便上层触发 positions/delta（页面秒级刷新）与对账逻辑。
		s.emit(PrivateEventAccount, "", msg)
		// Best-effort: extract symbols from a.P and emit PrivateEventPosition per symbol.
		// Payload 仍沿用原始 msg（上层只需要 symbol 做路由/触发同步）。
		if a, ok := m["a"].(map[string]any); ok {
			if ps, ok := a["P"].([]any); ok && len(ps) > 0 {
				seen := make(map[string]struct{}, len(ps))
				for _, it := range ps {
					p, ok := it.(map[string]any)
					if !ok {
						continue
					}
					sym, _ := p["s"].(string)
					sym = strings.ToUpper(strings.TrimSpace(sym))
					if sym == "" {
						continue
					}
					if _, dup := seen[sym]; dup {
						continue
					}
					seen[sym] = struct{}{}
					s.emit(PrivateEventPosition, sym, msg)
				}
			}
		}
	default:
		// ignore
	}
}

func (s *BinancePrivateStream) keepaliveLoop() {
	t := time.NewTicker(30 * time.Minute)
	defer t.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-t.C:
			s.mu.RLock()
			lk := s.listenKey
			s.mu.RUnlock()
			if lk == "" {
				continue
			}
			if err := s.keepaliveListenKey(s.ctx, lk); err != nil {
				g.Log().Warningf(s.ctx, "[BinancePrivateWS] keepalive listenKey failed: %v", err)
			}
		}
	}
}

func (s *BinancePrivateStream) createListenKey(ctx context.Context) (string, error) {
	u := s.endpoint + "/fapi/v1/listenKey"
	req, _ := http.NewRequestWithContext(ctx, "POST", u, nil)
	req.Header.Set("X-MBX-APIKEY", s.apiKey)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("create listenKey failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	var out struct {
		ListenKey string `json:"listenKey"`
	}
	_ = json.Unmarshal(body, &out)
	if out.ListenKey == "" {
		return "", fmt.Errorf("create listenKey: empty response: %s", string(body))
	}
	return out.ListenKey, nil
}

func (s *BinancePrivateStream) keepaliveListenKey(ctx context.Context, listenKey string) error {
	u := s.endpoint + "/fapi/v1/listenKey"
	form := url.Values{}
	form.Set("listenKey", listenKey)
	req, _ := http.NewRequestWithContext(ctx, "PUT", u, bytes.NewBufferString(form.Encode()))
	req.Header.Set("X-MBX-APIKEY", s.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("keepalive listenKey failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	return nil
}

func (s *BinancePrivateStream) closeListenKey(ctx context.Context, listenKey string) error {
	u := s.endpoint + "/fapi/v1/listenKey"
	form := url.Values{}
	form.Set("listenKey", listenKey)
	req, _ := http.NewRequestWithContext(ctx, "DELETE", u, bytes.NewBufferString(form.Encode()))
	req.Header.Set("X-MBX-APIKEY", s.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
	return nil
}


