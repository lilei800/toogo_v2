// Package websocket WebSocket客户端处理
package websocket

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gorilla/websocket"
)

const (
	// 写超时
	writeWait = 10 * time.Second
	// Pong超时
	pongWait = 60 * time.Second
	// Ping间隔
	pingPeriod = (pongWait * 9) / 10
	// 最大消息大小
	maxMessageSize = 512 * 1024
)

// ClientMessage 客户端发送的消息
type ClientMessage struct {
	Action  string `json:"action"`  // subscribe/unsubscribe/ping
	Channel string `json:"channel"` // 频道名称
}

// NewClient 创建新客户端
func NewClient(hub *Hub, conn *websocket.Conn, userID int64) *Client {
	return &Client{
		Hub:        hub,
		Conn:       conn,
		UserID:     userID,
		Send:       make(chan []byte, 256),
		Subscribed: make(map[string]bool),
	}
}

// ReadPump 读取消息泵
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()
	
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				g.Log().Errorf(context.Background(), "[WebSocket] Read error: %v", err)
			}
			break
		}
		
		// 处理客户端消息
		c.handleMessage(message)
	}
}

// WritePump 写入消息泵
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub关闭了通道
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			
			// 批量发送队列中的消息
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}
			
			if err := w.Close(); err != nil {
				return
			}
			
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理客户端消息
func (c *Client) handleMessage(data []byte) {
	var msg ClientMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		g.Log().Errorf(context.Background(), "[WebSocket] Parse message error: %v", err)
		return
	}
	
	switch msg.Action {
	case "subscribe":
		if msg.Channel != "" {
			c.Hub.Subscribe(c, msg.Channel)
			// 发送订阅确认
			c.sendResponse("subscribed", msg.Channel)
		}
		
	case "unsubscribe":
		if msg.Channel != "" {
			c.Hub.Unsubscribe(c, msg.Channel)
			c.sendResponse("unsubscribed", msg.Channel)
		}
		
	case "ping":
		c.sendResponse("pong", "")
		
	default:
		g.Log().Debugf(context.Background(), "[WebSocket] Unknown action: %s", msg.Action)
	}
}

// sendResponse 发送响应
func (c *Client) sendResponse(action, channel string) {
	response := map[string]interface{}{
		"type":      "response",
		"action":    action,
		"channel":   channel,
		"timestamp": time.Now().UnixMilli(),
	}
	
	data, _ := json.Marshal(response)
	select {
	case c.Send <- data:
	default:
	}
}

