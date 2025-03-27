// 事件处理 (handler.go)
// 职责：处理 WebSocket 消息事件的路由，负责根据消息类型分发不同的处理函数。
// 建议：将处理不同消息类型的函数封装成不同的处理器，使用策略模式或函数映射进行动态分发。这些事件处理函数会根据接收到的消息类型（如私聊、群聊等）进行不同的处理。
package handler

import (
	"log"
	"net/http"

	"websocket-server/internal/auth"
	"websocket-server/internal/connection"
	"websocket-server/internal/message"
	"websocket-server/internal/room"

	"github.com/gorilla/websocket"
)

// Handler 处理 WebSocket 消息
type Handler struct {
	connMgr *connection.ConnectionManager
	msgMgr  *message.MessageManager
	roomMgr *room.RoomManager
	authMgr *auth.AuthManager
}

// NewHandler 创建 Handler 实例
func NewHandler(connMgr *connection.ConnectionManager, msgMgr *message.MessageManager, authMgr *auth.AuthManager, roomMgr *room.RoomManager) *Handler {
	return &Handler{
		connMgr: connMgr,
		msgMgr:  msgMgr,
		roomMgr: roomMgr,
		authMgr: authMgr,
	}
}

// HandleMessage 处理客户端发送的消息
func (h *Handler) HandleMessage(client *connection.Client, data []byte) {
	msg, err := message.ParseMessage(data)
	if err != nil {
		log.Println("解析消息失败:", err)
		return
	}

	switch msg.Type {
	case "broadcast":
		h.msgMgr.BroadcastMessage(msg)
	case "join_room":
		h.roomMgr.GetRoom(msg.RoomID).AddMember(client.UserID)
	case "leave_room":
		h.roomMgr.GetRoom(msg.RoomID).RemoveMember(client.UserID)
	default:
		log.Println("未知消息类型:", msg.Type)
	}
}

// HandleWebSocket 处理 WebSocket 请求
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 这里是 WebSocket 处理的逻辑
	log.Println("Handling WebSocket connection...")

	// 示例：如果是 WebSocket 连接，升级协议并处理连接
	// 你可以在这里进行身份验证，连接管理等操作

	// 使用 gorilla/websocket 库来升级连接
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	newClinet := &connection.Client{
		Conn:   conn,
		UserID: "test", // get userID from http head
	}

	// 在连接管理器中注册新的连接
	h.connMgr.AddClient(newClinet)
	log.Println("WebSocket connection established")
}
