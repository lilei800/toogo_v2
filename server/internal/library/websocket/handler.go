// Package websocket WebSocket HTTP处理器
package websocket

import (
	"net/http"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许跨域
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ServeWs WebSocket处理入口
func ServeWs(r *ghttp.Request) {
	ctx := r.Context()
	
	// 从上下文或请求中获取用户ID
	userID := int64(0)
	if v := r.Get("user_id"); v != nil {
		userID = v.Int64()
	}
	
	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(r.Response.Writer, r.Request, nil)
	if err != nil {
		g.Log().Errorf(ctx, "[WebSocket] Upgrade error: %v", err)
		return
	}
	
	// 创建客户端
	hub := GetHub()
	client := NewClient(hub, conn, userID)
	
	// 注册客户端
	hub.Register(client)
	
	// 启动读写协程
	go client.WritePump()
	go client.ReadPump()
	
	g.Log().Infof(ctx, "[WebSocket] New connection: UserID=%d, RemoteAddr=%s", 
		userID, r.GetClientIp())
}

