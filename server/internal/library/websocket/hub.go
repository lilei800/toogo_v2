// Package websocket Toogo WebSocket Hub
// 管理所有WebSocket连接和消息广播
package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gorilla/websocket"
)

// WebSocket配置常量
const (
	// 心跳间隔
	PingInterval = 30 * time.Second
	// Pong超时时间
	PongTimeout = 60 * time.Second
	// 写入超时
	WriteTimeout = 10 * time.Second
	// 最大消息大小
	MaxMessageSize = 512 * 1024 // 512KB
	// 发送缓冲区大小
	SendBufferSize = 256
	// 重连最大次数
	MaxReconnectAttempts = 5
	// 重连间隔
	ReconnectInterval = 5 * time.Second
)

// MessageType 消息类型
type MessageType string

const (
	// 行情消息
	MsgTypeTicker   MessageType = "ticker"   // 实时价格
	MsgTypeKline    MessageType = "kline"    // K线数据
	MsgTypeDepth    MessageType = "depth"    // 深度数据
	
	// 订单消息
	MsgTypePosition MessageType = "position" // 持仓更新
	MsgTypeOrder    MessageType = "order"    // 订单更新
	MsgTypeTrade    MessageType = "trade"    // 成交更新
	
	// 机器人消息
	MsgTypeRobot    MessageType = "robot"    // 机器人状态
	MsgTypeSignal   MessageType = "signal"   // 交易信号
	MsgTypePnL      MessageType = "pnl"      // 盈亏更新
	
	// 系统消息
	MsgTypeSystem   MessageType = "system"   // 系统通知
	MsgTypeError    MessageType = "error"    // 错误消息
)

// Message WebSocket消息结构
type Message struct {
	Type      MessageType `json:"type"`       // 消息类型
	Channel   string      `json:"channel"`    // 频道/交易对
	Data      interface{} `json:"data"`       // 消息数据
	Timestamp int64       `json:"timestamp"`  // 时间戳
}

// Client WebSocket客户端
type Client struct {
	Hub          *Hub
	Conn         *websocket.Conn
	UserID       int64             // 用户ID
	Send         chan []byte       // 发送通道
	Subscribed   map[string]bool   // 订阅的频道
	mu           sync.RWMutex
	
	// 心跳相关
	LastPing     time.Time         // 最后一次Ping时间
	LastPong     time.Time         // 最后一次Pong时间
	IsAlive      bool              // 是否存活
	
	// 连接信息
	ConnectedAt  time.Time         // 连接时间
	RemoteAddr   string            // 远程地址
	UserAgent    string            // 用户代理
	
	// 统计信息
	MessagesSent int64             // 发送消息数
	MessagesRecv int64             // 接收消息数
}

// ClientInfo 客户端信息（用于监控）
type ClientInfo struct {
	UserID       int64     `json:"userId"`
	RemoteAddr   string    `json:"remoteAddr"`
	ConnectedAt  time.Time `json:"connectedAt"`
	LastActivity time.Time `json:"lastActivity"`
	IsAlive      bool      `json:"isAlive"`
	Subscribed   []string  `json:"subscribed"`
	MessagesSent int64     `json:"messagesSent"`
	MessagesRecv int64     `json:"messagesRecv"`
}

// Hub WebSocket连接管理中心
type Hub struct {
	// 所有连接的客户端
	clients map[*Client]bool
	
	// 按用户ID索引的客户端
	userClients map[int64][]*Client
	
	// 按频道订阅的客户端
	channelClients map[string]map[*Client]bool
	
	// 广播消息通道
	broadcast chan *Message
	
	// 注册客户端
	register chan *Client
	
	// 注销客户端
	unregister chan *Client
	
	// 锁
	mu sync.RWMutex
	
	// 心跳检测
	pingTicker *time.Ticker
	
	// 统计信息
	totalConnections    int64
	totalMessages       int64
	startTime           time.Time
	
	// 是否运行中
	running bool
	stopCh  chan struct{}
}

// HubStats Hub统计信息
type HubStats struct {
	OnlineClients      int           `json:"onlineClients"`
	TotalConnections   int64         `json:"totalConnections"`
	TotalMessages      int64         `json:"totalMessages"`
	Uptime             time.Duration `json:"uptime"`
	ChannelCount       int           `json:"channelCount"`
	UniqueUsers        int           `json:"uniqueUsers"`
}

var (
	hubInstance *Hub
	once        sync.Once
)

// GetHub 获取Hub单例
func GetHub() *Hub {
	once.Do(func() {
		hubInstance = &Hub{
			clients:        make(map[*Client]bool),
			userClients:    make(map[int64][]*Client),
			channelClients: make(map[string]map[*Client]bool),
			broadcast:      make(chan *Message, SendBufferSize),
			register:       make(chan *Client),
			unregister:     make(chan *Client),
			pingTicker:     time.NewTicker(PingInterval),
			startTime:      time.Now(),
			stopCh:         make(chan struct{}),
		}
		go hubInstance.Run()
	})
	return hubInstance
}

// Run 运行Hub主循环
func (h *Hub) Run() {
	h.running = true
	g.Log().Info(context.Background(), "[WebSocket] Hub started")
	
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
			
		case client := <-h.unregister:
			h.unregisterClient(client)
			
		case message := <-h.broadcast:
			h.broadcastMessage(message)
			h.totalMessages++
			
		case <-h.pingTicker.C:
			h.checkHeartbeat()
			
		case <-h.stopCh:
			h.running = false
			h.pingTicker.Stop()
			h.closeAllClients()
			g.Log().Info(context.Background(), "[WebSocket] Hub stopped")
			return
		}
	}
}

// Stop 停止Hub
func (h *Hub) Stop() {
	close(h.stopCh)
}

// checkHeartbeat 检查心跳
func (h *Hub) checkHeartbeat() {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	now := time.Now()
	deadClients := make([]*Client, 0)
	
	for client := range h.clients {
		// 检查Pong超时
		if now.Sub(client.LastPong) > PongTimeout {
			client.IsAlive = false
			deadClients = append(deadClients, client)
			continue
		}
		
		// 发送Ping
		client.LastPing = now
		if err := h.sendPing(client); err != nil {
			g.Log().Warningf(context.Background(), "[WebSocket] Send ping error: UserID=%d, err=%v", 
				client.UserID, err)
			deadClients = append(deadClients, client)
		}
	}
	
	// 清理死亡连接
	for _, client := range deadClients {
		g.Log().Infof(context.Background(), "[WebSocket] Client timeout, disconnecting: UserID=%d", 
			client.UserID)
		go h.Unregister(client)
	}
	
	if len(deadClients) > 0 {
		g.Log().Infof(context.Background(), "[WebSocket] Heartbeat check: %d clients removed, %d remaining", 
			len(deadClients), len(h.clients)-len(deadClients))
	}
}

// sendPing 发送Ping
func (h *Hub) sendPing(client *Client) error {
	client.Conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
	return client.Conn.WriteMessage(websocket.PingMessage, nil)
}

// closeAllClients 关闭所有客户端
func (h *Hub) closeAllClients() {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	for client := range h.clients {
		close(client.Send)
		client.Conn.Close()
	}
	h.clients = make(map[*Client]bool)
	h.userClients = make(map[int64][]*Client)
	h.channelClients = make(map[string]map[*Client]bool)
}

// registerClient 注册客户端
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// 初始化客户端状态
	client.ConnectedAt = time.Now()
	client.LastPong = time.Now()
	client.IsAlive = true
	if client.Subscribed == nil {
		client.Subscribed = make(map[string]bool)
	}
	
	h.clients[client] = true
	h.totalConnections++
	
	// 添加到用户索引
	if client.UserID > 0 {
		h.userClients[client.UserID] = append(h.userClients[client.UserID], client)
	}
	
	g.Log().Infof(context.Background(), "[WebSocket] Client registered: UserID=%d, RemoteAddr=%s, Total=%d", 
		client.UserID, client.RemoteAddr, len(h.clients))
}

// unregisterClient 注销客户端
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)
		
		// 从用户索引移除
		if client.UserID > 0 {
			clients := h.userClients[client.UserID]
			for i, c := range clients {
				if c == client {
					h.userClients[client.UserID] = append(clients[:i], clients[i+1:]...)
					break
				}
			}
			if len(h.userClients[client.UserID]) == 0 {
				delete(h.userClients, client.UserID)
			}
		}
		
		// 从频道订阅移除
		client.mu.RLock()
		for channel := range client.Subscribed {
			if clients, ok := h.channelClients[channel]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.channelClients, channel)
				}
			}
		}
		client.mu.RUnlock()
		
		g.Log().Debugf(context.Background(), "[WebSocket] Client unregistered: UserID=%d, Total=%d", 
			client.UserID, len(h.clients))
	}
}

// broadcastMessage 广播消息
func (h *Hub) broadcastMessage(message *Message) {
	data, err := json.Marshal(message)
	if err != nil {
		g.Log().Errorf(context.Background(), "[WebSocket] Marshal message error: %v", err)
		return
	}
	
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	// 如果有频道，只发送给订阅了该频道的客户端
	if message.Channel != "" {
		if clients, ok := h.channelClients[message.Channel]; ok {
			for client := range clients {
				select {
				case client.Send <- data:
				default:
					// 发送失败，可能客户端已断开
				}
			}
		}
		return
	}
	
	// 无频道，广播给所有客户端
	for client := range h.clients {
		select {
		case client.Send <- data:
		default:
		}
	}
}

// Subscribe 订阅频道
func (h *Hub) Subscribe(client *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	client.mu.Lock()
	client.Subscribed[channel] = true
	client.mu.Unlock()
	
	if h.channelClients[channel] == nil {
		h.channelClients[channel] = make(map[*Client]bool)
	}
	h.channelClients[channel][client] = true
	
	g.Log().Debugf(context.Background(), "[WebSocket] Client subscribed: UserID=%d, Channel=%s", 
		client.UserID, channel)
}

// Unsubscribe 取消订阅
func (h *Hub) Unsubscribe(client *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	client.mu.Lock()
	delete(client.Subscribed, channel)
	client.mu.Unlock()
	
	if clients, ok := h.channelClients[channel]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.channelClients, channel)
		}
	}
}

// SendToUser 发送消息给指定用户
func (h *Hub) SendToUser(userID int64, message *Message) {
	data, err := json.Marshal(message)
	if err != nil {
		g.Log().Errorf(context.Background(), "[WebSocket] Marshal message error: %v", err)
		return
	}
	
	h.mu.RLock()
	clients := h.userClients[userID]
	h.mu.RUnlock()
	
	for _, client := range clients {
		select {
		case client.Send <- data:
		default:
		}
	}
}

// SendToChannel 发送消息到频道
func (h *Hub) SendToChannel(channel string, message *Message) {
	message.Channel = channel
	h.broadcast <- message
}

// Broadcast 广播消息给所有客户端
func (h *Hub) Broadcast(message *Message) {
	h.broadcast <- message
}

// Register 注册客户端
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 注销客户端
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// GetOnlineCount 获取在线人数
func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetUserOnline 检查用户是否在线
func (h *Hub) GetUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.userClients[userID]) > 0
}

// GetStats 获取Hub统计信息
func (h *Hub) GetStats() *HubStats {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	uniqueUsers := len(h.userClients)
	
	return &HubStats{
		OnlineClients:    len(h.clients),
		TotalConnections: h.totalConnections,
		TotalMessages:    h.totalMessages,
		Uptime:           time.Since(h.startTime),
		ChannelCount:     len(h.channelClients),
		UniqueUsers:      uniqueUsers,
	}
}

// GetClientInfo 获取客户端信息
func (h *Hub) GetClientInfo(userID int64) []*ClientInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	clients := h.userClients[userID]
	infos := make([]*ClientInfo, 0, len(clients))
	
	for _, client := range clients {
		client.mu.RLock()
		subscribed := make([]string, 0, len(client.Subscribed))
		for ch := range client.Subscribed {
			subscribed = append(subscribed, ch)
		}
		client.mu.RUnlock()
		
		infos = append(infos, &ClientInfo{
			UserID:       client.UserID,
			RemoteAddr:   client.RemoteAddr,
			ConnectedAt:  client.ConnectedAt,
			LastActivity: client.LastPong,
			IsAlive:      client.IsAlive,
			Subscribed:   subscribed,
			MessagesSent: client.MessagesSent,
			MessagesRecv: client.MessagesRecv,
		})
	}
	
	return infos
}

// GetAllClientsInfo 获取所有客户端信息
func (h *Hub) GetAllClientsInfo() []*ClientInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	infos := make([]*ClientInfo, 0, len(h.clients))
	
	for client := range h.clients {
		client.mu.RLock()
		subscribed := make([]string, 0, len(client.Subscribed))
		for ch := range client.Subscribed {
			subscribed = append(subscribed, ch)
		}
		client.mu.RUnlock()
		
		infos = append(infos, &ClientInfo{
			UserID:       client.UserID,
			RemoteAddr:   client.RemoteAddr,
			ConnectedAt:  client.ConnectedAt,
			LastActivity: client.LastPong,
			IsAlive:      client.IsAlive,
			Subscribed:   subscribed,
			MessagesSent: client.MessagesSent,
			MessagesRecv: client.MessagesRecv,
		})
	}
	
	return infos
}

// KickUser 踢出用户
func (h *Hub) KickUser(userID int64, reason string) {
	h.mu.RLock()
	clients := make([]*Client, len(h.userClients[userID]))
	copy(clients, h.userClients[userID])
	h.mu.RUnlock()
	
	// 发送踢出消息
	kickMsg := &Message{
		Type: MsgTypeSystem,
		Data: map[string]interface{}{
			"action": "kick",
			"reason": reason,
		},
		Timestamp: time.Now().Unix(),
	}
	
	for _, client := range clients {
		h.SendToUser(userID, kickMsg)
		h.Unregister(client)
	}
	
	g.Log().Infof(context.Background(), "[WebSocket] User kicked: UserID=%d, Reason=%s", userID, reason)
}

// GetChannelSubscribers 获取频道订阅者数量
func (h *Hub) GetChannelSubscribers(channel string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if clients, ok := h.channelClients[channel]; ok {
		return len(clients)
	}
	return 0
}

// GetAllChannels 获取所有频道
func (h *Hub) GetAllChannels() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	channels := make([]string, 0, len(h.channelClients))
	for ch := range h.channelClients {
		channels = append(channels, ch)
	}
	return channels
}

// IsRunning 检查Hub是否运行中
func (h *Hub) IsRunning() bool {
	return h.running
}

// HandlePong 处理Pong响应
func (client *Client) HandlePong() {
	client.mu.Lock()
	client.LastPong = time.Now()
	client.IsAlive = true
	client.mu.Unlock()
}

// IncrementMessagesSent 增加发送消息计数
func (client *Client) IncrementMessagesSent() {
	client.mu.Lock()
	client.MessagesSent++
	client.mu.Unlock()
}

// IncrementMessagesRecv 增加接收消息计数
func (client *Client) IncrementMessagesRecv() {
	client.mu.Lock()
	client.MessagesRecv++
	client.mu.Unlock()
}

