// Package exchange WebSocket基础连接管理
package exchange

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gorilla/websocket"
)

// WebSocketState 连接状态
type WebSocketState int

const (
	WSStateDisconnected WebSocketState = iota
	WSStateConnecting
	WSStateConnected
	WSStateReconnecting
)

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	URL               string                                       // WebSocket地址
	PingInterval      time.Duration                                // 心跳间隔
	PongTimeout       time.Duration                                // 心跳超时
	ReconnectDelay    time.Duration                                // 重连延迟
	MaxReconnects     int                                          // 最大重连次数
	MessageBufferSize int                                          // 消息缓冲区大小
	ProxyDialer       func(network, addr string) (net.Conn, error) // 代理拨号器（可选）
}

// DefaultWebSocketConfig 默认配置
func DefaultWebSocketConfig() *WebSocketConfig {
	return &WebSocketConfig{
		PingInterval:      15 * time.Second,
		PongTimeout:       10 * time.Second,
		ReconnectDelay:    1 * time.Second,
		MaxReconnects:     100,
		MessageBufferSize: 1000,
	}
}

// WebSocketConnection WebSocket连接管理器
type WebSocketConnection struct {
	mu      sync.RWMutex
	writeMu sync.Mutex // 写入互斥锁，防止并发写入导致 panic
	config  *WebSocketConfig

	conn    *websocket.Conn
	state   WebSocketState
	stopCh  chan struct{}
	msgChan chan []byte

	// 回调
	onMessage         func([]byte)
	onConnected       func()
	onDisconnected    func(error)
	onReconnectFailed func(error) // 重连失败回调（重连次数超过上限时触发）

	// 重连控制
	reconnectCount  int
	lastConnectTime time.Time

	// 订阅管理
	subscriptions map[string]interface{} // 保存订阅信息用于重连后恢复
}

// NewWebSocketConnection 创建WebSocket连接
func NewWebSocketConnection(config *WebSocketConfig) *WebSocketConnection {
	if config == nil {
		config = DefaultWebSocketConfig()
	}
	return &WebSocketConnection{
		config:        config,
		state:         WSStateDisconnected,
		stopCh:        make(chan struct{}),
		msgChan:       make(chan []byte, config.MessageBufferSize),
		subscriptions: make(map[string]interface{}),
	}
}

// SetCallbacks 设置回调
func (c *WebSocketConnection) SetCallbacks(onMessage func([]byte), onConnected func(), onDisconnected func(error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onMessage = onMessage
	c.onConnected = onConnected
	c.onDisconnected = onDisconnected
}

// SetReconnectFailedCallback 设置重连失败回调
// 当重连次数超过上限时触发，可用于通知上层应用进行处理
func (c *WebSocketConnection) SetReconnectFailedCallback(callback func(error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onReconnectFailed = callback
}

// Connect 建立连接
func (c *WebSocketConnection) Connect(ctx context.Context) error {
	c.mu.Lock()
	if c.state == WSStateConnected || c.state == WSStateConnecting {
		c.mu.Unlock()
		return nil
	}
	c.state = WSStateConnecting
	c.mu.Unlock()

	dialer := websocket.Dialer{
		HandshakeTimeout: 30 * time.Second,
	}

	// 如果配置了代理拨号器，使用代理连接
	if c.config.ProxyDialer != nil {
		dialer.NetDial = c.config.ProxyDialer
		dialer.NetDialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return c.config.ProxyDialer(network, addr)
		}
		g.Log().Info(ctx, "[WebSocket] 使用代理连接:", c.config.URL)
	}

	conn, _, err := dialer.DialContext(ctx, c.config.URL, nil)
	if err != nil {
		c.mu.Lock()
		c.state = WSStateDisconnected
		c.mu.Unlock()
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.state = WSStateConnected
	c.lastConnectTime = time.Now()
	c.reconnectCount = 0
	c.mu.Unlock()

	// 启动消息读取和心跳
	go c.readLoop(ctx)
	go c.pingLoop(ctx)
	go c.processLoop(ctx)

	// 触发连接成功回调
	if c.onConnected != nil {
		c.onConnected()
	}

	g.Log().Info(ctx, "[WebSocket] 连接成功:", c.config.URL)
	return nil
}

// Disconnect 断开连接
func (c *WebSocketConnection) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == WSStateDisconnected {
		return
	}

	c.state = WSStateDisconnected
	close(c.stopCh)

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

// IsConnected 检查是否已连接
func (c *WebSocketConnection) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state == WSStateConnected
}

// GetState 获取连接状态
func (c *WebSocketConnection) GetState() WebSocketState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// Send 发送消息（使用写入锁防止并发写入）
func (c *WebSocketConnection) Send(data interface{}) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return websocket.ErrCloseSent
	}

	var msg []byte
	var err error

	switch v := data.(type) {
	case []byte:
		msg = v
	case string:
		msg = []byte(v)
	default:
		msg, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}

	// 使用写入锁保护 WriteMessage，防止 concurrent write to websocket connection
	c.writeMu.Lock()
	err = conn.WriteMessage(websocket.TextMessage, msg)
	c.writeMu.Unlock()
	return err
}

// SaveSubscription 保存订阅信息（用于重连恢复）
func (c *WebSocketConnection) SaveSubscription(key string, sub interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subscriptions[key] = sub
}

// RemoveSubscription 移除订阅信息
func (c *WebSocketConnection) RemoveSubscription(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.subscriptions, key)
}

// GetSubscriptions 获取所有订阅
func (c *WebSocketConnection) GetSubscriptions() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]interface{}, len(c.subscriptions))
	for k, v := range c.subscriptions {
		result[k] = v
	}
	return result
}

// readLoop 读取消息循环
func (c *WebSocketConnection) readLoop(ctx context.Context) {
	defer func() {
		c.handleDisconnect(ctx, nil)
	}()

	for {
		select {
		case <-c.stopCh:
			return
		default:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				return
			}

			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					g.Log().Warningf(ctx, "[WebSocket] 读取错误: %v", err)
				}
				c.handleDisconnect(ctx, err)
				return
			}

			// 发送到消息通道
			select {
			case c.msgChan <- message:
			default:
				// 队列满，丢弃旧消息
				select {
				case <-c.msgChan:
				default:
				}
				c.msgChan <- message
			}
		}
	}
}

// processLoop 处理消息循环
func (c *WebSocketConnection) processLoop(ctx context.Context) {
	for {
		select {
		case <-c.stopCh:
			return
		case msg := <-c.msgChan:
			if c.onMessage != nil {
				c.onMessage(msg)
			}
		}
	}
}

// pingLoop 心跳循环（使用写入锁防止并发写入）
func (c *WebSocketConnection) pingLoop(ctx context.Context) {
	ticker := time.NewTicker(c.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				return
			}

			// 使用写入锁保护 WriteMessage，防止 concurrent write to websocket connection
			c.writeMu.Lock()
			err := conn.WriteMessage(websocket.PingMessage, nil)
			c.writeMu.Unlock()

			if err != nil {
				g.Log().Warningf(ctx, "[WebSocket] 发送心跳失败: %v", err)
				c.handleDisconnect(ctx, err)
				return
			}
		}
	}
}

// handleDisconnect 处理断开连接
func (c *WebSocketConnection) handleDisconnect(ctx context.Context, err error) {
	c.mu.Lock()
	if c.state == WSStateDisconnected || c.state == WSStateReconnecting {
		c.mu.Unlock()
		return
	}
	c.state = WSStateReconnecting
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()

	// 触发断开回调
	if c.onDisconnected != nil {
		c.onDisconnected(err)
	}

	// 尝试重连
	go c.reconnect(ctx)
}

// reconnect 重连
func (c *WebSocketConnection) reconnect(ctx context.Context) {
	for {
		c.mu.Lock()
		if c.state == WSStateDisconnected {
			c.mu.Unlock()
			return
		}
		c.reconnectCount++
		count := c.reconnectCount
		c.mu.Unlock()

		if count > c.config.MaxReconnects {
			g.Log().Errorf(ctx, "[WebSocket] 重连次数超过上限 (%d), 停止重连", c.config.MaxReconnects)
			c.mu.Lock()
			c.state = WSStateDisconnected
			callback := c.onReconnectFailed
			c.mu.Unlock()

			// 触发重连失败回调，通知上层应用
			if callback != nil {
				reconnectErr := gerror.Newf("WebSocket重连失败: 已尝试%d次, 超过上限%d", count, c.config.MaxReconnects)
				callback(reconnectErr)
			}
			return
		}

		// 指数退避
		delay := c.config.ReconnectDelay * time.Duration(1<<uint(min(count-1, 6)))
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}

		g.Log().Infof(ctx, "[WebSocket] 第 %d 次重连, 等待 %v", count, delay)
		time.Sleep(delay)

		// 重新创建stopCh和msgChan
		c.mu.Lock()
		c.stopCh = make(chan struct{})
		c.msgChan = make(chan []byte, c.config.MessageBufferSize)
		c.mu.Unlock()

		err := c.Connect(ctx)
		if err == nil {
			g.Log().Info(ctx, "[WebSocket] 重连成功")
			return
		}

		g.Log().Warningf(ctx, "[WebSocket] 重连失败: %v", err)
	}
}

// min 返回较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
