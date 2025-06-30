// NewServer() 函数设计
// 在 NewServer() 中，可以进行这些组件的初始化：
// 创建并初始化 连接管理器（ConnectionManager）。
// 创建 消息管理器（MessageManager）等其他功能模块。
// 设置 WebSocket 服务器的事件处理逻辑，确保 WebSocket 的消息能够正确分发到相应的处理器。
// 将这些组件传递给 Server 结构体，以便后续管理和使用。
// NewServer() 不需要直接封装所有的具体业务逻辑，只需要负责将这些不同模块集成起来，保持结构清晰。
package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"websocket-server/handler"
	"websocket-server/httpapi"
	"websocket-server/internal/auth"
	"websocket-server/internal/broker"
	"websocket-server/internal/connection"
	"websocket-server/internal/message"
	"websocket-server/internal/room"
	"websocket-server/internal/snowflake"
	"websocket-server/internal/storage"
)

// Server .
type Server struct {
	connMgr      *connection.ConnectionManager
	msgMgr       *message.MessageManager
	authMgr      *auth.AuthManager
	roomMgr      *room.RoomManager
	handler      *handler.Handler
	server       *http.Server
	kafkaBroker  *broker.KafkaBroker
	redisStorage *storage.RedisStorage
	mongoStorage *storage.MongoStorage

	Handler *handler.Handler // 要把Handler放这对外公开
	Deps    any
}

// NewServer 相当于是newService,先把server当作service用
func NewServer() *Server {
	// 初始化各个管理器
	connMgr := connection.NewConnectionManager()
	msgMgr := message.NewMessageManager()
	authMgr := auth.NewAuthManager()
	roomMgr := room.NewRoomManager()

	// 主存
	mongoStorage, err := storage.NewMongoStorage("mongodb://localhost:27017", "chatDB", "messages")
	if err != nil {
		log.Fatalf("Failed to create MongoDB storage: %v", err)
	}
	// 缓存
	redisStorage := storage.NewRedisStorage("localhost:6379")
	// 消息队列
	kafkaBroker, err := broker.NewKafkaBroker([]string{"localhost:9092"}, "websocket-messages")
	if err != nil {
		log.Fatalf("Failed to create Kafka broker: %v", err)
	}
	// temp 启动临时消费者
	handleFunc := func(msg string, offset int64) {
		fmt.Printf("别烦,本消费者正在打印! offset=%d value=%s\n", offset, msg)
	}
	go kafkaBroker.MessageConsume(handleFunc)

	// 创建 WebSocket 处理器
	wsHandler := handler.NewHandler(connMgr, msgMgr, authMgr, roomMgr, kafkaBroker, redisStorage, mongoStorage)
	handler.GHandler = wsHandler

	// 注册事件处理器
	wsHandler.RegisterEventHandler("broadcast", wsHandler.BroadcastMessage)
	wsHandler.RegisterEventHandler("direct", wsHandler.SendDirectMessage)

	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// 雪花算法
	snowflake.GlabolSF, err = snowflake.NewSnowflake(1, 1)
	if err != nil {
		return nil
	}

	return &Server{
		connMgr:      connMgr,
		msgMgr:       msgMgr,
		authMgr:      authMgr,
		roomMgr:      roomMgr,
		handler:      wsHandler,
		server:       server,
		kafkaBroker:  kafkaBroker,
		redisStorage: redisStorage,
		mongoStorage: mongoStorage,
	}
}

// Start 开启服务
func (s *Server) Start(addr string) error {
	r := httpapi.SetupRouter()
	// 或者下面三行直接用 return r.Run(addr) 代替
	s.server.Handler = r
	s.server.Addr = addr
	return s.server.ListenAndServe()
}

// Shutdown 停机
func (s *Server) Shutdown() {
	log.Println("Closing all connections...")
	s.connMgr.CloseAllConnections()
	s.msgMgr.Shutdown()
}

// 【Server 是负责整个系统的初始化和优雅退出的，而 Handler 才是负责具体业务的。】

// Server 它的本质职责是：

// 统一管理依赖生命周期（初始化、关闭、重启等）

// 组装所有模块（Handler、Router、Broker等）

// 启动 HTTP / WebSocket 服务

// 统一 Shutdown 资源

// 它是你的 IM 系统大管家，但它并不负责业务逻辑处理。

// 🚀 Server 的流程应该是：
// func main() {
//     server := NewServer()

//     // 可以启动路由（可能通过 handler 或 router 传进去）
//     go func() {
//         if err := server.Start(":8080"); err != nil {
//             log.Fatalf("Server error: %v", err)
//         }
//     }()

//     // 优雅退出
//     waitForShutdownSignal()
//     server.Shutdown()
// }
// Server：

// 启动 HTTP Server / Gin / WebSocket

// 管理 Kafka、Redis、Mongo 等资源的连接生命周期

// 提供统一关闭接口（比如 Shutdown）

// 👉 一旦初始化好，你一般不会直接用它做业务逻辑，因为业务逻辑交给 Handler 或其他业务模块。

// ❓ 为什么 Server 不直接参与业务逻辑？
// 如果 Server 下去处理业务逻辑，Server 会变得臃肿、职责不清：

// Server 就会变成业务和系统管理混在一起的“上帝类”

// 不利于测试（你想单测逻辑时，不需要整个 Server 启动起来）

// 逻辑耦合太强，后期维护困难

// 💡 Server 的生命周期价值在哪？
// ✅ 初始化全局依赖（连接池、broker、storage、handler）
// ✅ 启动并持有 HTTP / WS server
// ✅ 管理资源关闭、重启、动态加载配置等
// ✅ 有时可以做健康检查、监控、指标暴露
