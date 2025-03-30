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

	// 可选：存储消息以便历史查询或离线消息处理
	h.msgMgr.StoreMessage(msg)

	switch msg.Type {
	case "broadcast":
		h.BroadcastMessage(msg)
	case "direct":
		h.SendDirectMessage(msg)
	// 根据需要增加更多消息类型的处理
	default:
		log.Println("未知消息类型:", msg.Type)
	}
}

// HandleWebSocket 处理 WebSocket 请求
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 这里是 WebSocket 处理的逻辑
	log.Println("Handling WebSocket connection...")

	// 示例：如果是 WebSocket 连接，升级协议并处理连接
	// 可以在这里进行身份验证，连接管理等操作

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

	// 从 URL 查询参数中获取 userID 和 reconnect 标记
	// userID := r.URL.Query().Get("user_id")
	reconnect := r.URL.Query().Get("reconnect") // "true" 表示重连

	newClient := &connection.Client{
		Conn:   conn,
		UserID: "test", // get userID from http head
	}

	// 在连接管理器中注册新的连接
	h.connMgr.AddClient(newClient)
	log.Println("WebSocket connection established")

	// 如果是断线重连，则恢复之前状态
	if reconnect == "true" {
		h.RestoreClientState(newClient)
	}

	// 启动心跳检测
	go newClient.StartHeartbeat()

	// **启动 ReadPump，让它监听消息**
	go h.ReadPump(newClient)
}

// **新增 ReadPump，让它监听 WebSocket 消息，并调用 HandleMessage**
func (h *Handler) ReadPump(client *connection.Client) {
	defer func() {
		h.connMgr.RemoveClient(client)
		client.Conn.Close()
	}()

	for {
		_, msg, err := client.Conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		// **收到消息后调用 HandleMessage**
		h.HandleMessage(client, msg)
	}
}

// Handler 中负责转发的部分：使用 ConnectionManager 来获取目标连接，然后发送消息
func (h *Handler) BroadcastMessage(msg *message.Message) {
	// 获取所有连接（这部分由 ConnectionManager 提供接口）
	for _, client := range h.connMgr.GetAllClients() {
		err := client.Conn.WriteMessage(websocket.TextMessage, []byte(msg.Data))
		if err != nil {
			log.Printf("发送消息给用户 %s 失败: %v", client.UserID, err)
		}
	}
}

func (h *Handler) SendDirectMessage(msg *message.Message) {
	// 例如，假设 msg 中的 Data 或者另有字段指定接收者 ID
	target := h.connMgr.GetClient(msg.Receiver)
	if target == nil {
		log.Printf("找不到目标用户: %s", msg.Receiver)
		return
	}

	// 这两种写法都可以，回家取舍一下
	// err := target.Conn.WriteMessage(websocket.TextMessage, []byte(msg.Data))
	err := target.SendMessage(websocket.TextMessage, []byte(msg.Data))
	if err != nil {
		log.Printf("发送消息给用户 %s 失败: %v", target.UserID, err)
	}
}

// RestoreClientState 是个伪函数，用于恢复客户端状态
func (h *Handler) RestoreClientState(client *connection.Client) {
	// 示例：从存储中获取该用户之前订阅的房间、离线消息等
	// 这里的具体实现需要你根据业务逻辑来编写
	log.Printf("Restoring state for client %s", client.UserID)
	// 例如：重新加入房间
	// room := h.roomMgr.GetRoom("exampleRoom")
	// room.AddMember(client.UserID)
	// 发送离线消息
	// offlineMessages := h.msgMgr.GetOfflineMessages(client.UserID)
	// for _, m := range offlineMessages {
	//     client.Conn.WriteMessage(websocket.TextMessage, m.Data)
	// }
}
