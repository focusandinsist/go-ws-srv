// NewServer() 函数设计
// 在 NewServer() 中，你可以进行这些组件的初始化：
// 创建并初始化 连接管理器（ConnectionManager）。
// 创建 消息管理器（MessageManager）等其他功能模块。
// 设置 WebSocket 服务器的事件处理逻辑，确保 WebSocket 的消息能够正确分发到相应的处理器。
// 将这些组件传递给 Server 结构体，以便后续管理和使用。
// NewServer() 不需要直接封装所有的具体业务逻辑，只需要负责将这些不同模块集成起来，保持结构清晰。
package server

import (
	"log"
	"net/http"
	"time"

	"websocket-server/internal/auth"
	"websocket-server/internal/connection"
	"websocket-server/internal/handler"
	"websocket-server/internal/message"
	"websocket-server/internal/room"
)

type Server struct {
	connMgr *connection.ConnectionManager
	msgMgr  *message.MessageManager
	authMgr *auth.AuthManager
	roomMgr *room.RoomManager
	handler *handler.Handler
	server  *http.Server
}

func NewServer() *Server {
	// 初始化各个管理器
	connMgr := connection.NewConnectionManager()
	msgMgr := message.NewMessageManager()
	authMgr := auth.NewAuthManager()
	roomMgr := room.NewRoomManager()

	// 创建 WebSocket 处理器
	wsHandler := handler.NewHandler(connMgr, msgMgr, authMgr, roomMgr)

	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &Server{
		connMgr: connMgr,
		msgMgr:  msgMgr,
		authMgr: authMgr,
		roomMgr: roomMgr,
		handler: wsHandler,
		server:  server,
	}
}

func (s *Server) Start(addr string) error {
	http.HandleFunc("/ws", s.handler.HandleWebSocket)
	s.server.Addr = addr
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown() {
	log.Println("Closing all connections...")
	s.connMgr.CloseAllConnections()
	s.msgMgr.Shutdown()
}
