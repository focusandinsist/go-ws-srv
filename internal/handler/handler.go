// 事件处理 (handler.go)
// 职责：处理 WebSocket 消息事件的路由，负责根据消息类型分发不同的处理函数。
// 建议：将处理不同消息类型的函数封装成不同的处理器，使用策略模式或函数映射进行动态分发。这些事件处理函数会根据接收到的消息类型（如私聊、群聊等）进行不同的处理。
package handler

import (
	"log"
	"net/http"

	"github.com/focusandinsist/go-ws-srv/internal/auth"
	"github.com/focusandinsist/go-ws-srv/internal/broker"
	"github.com/focusandinsist/go-ws-srv/internal/connection"
	"github.com/focusandinsist/go-ws-srv/internal/event"
	"github.com/focusandinsist/go-ws-srv/internal/message"
	"github.com/focusandinsist/go-ws-srv/internal/room"
	"github.com/focusandinsist/go-ws-srv/internal/storage"
	"github.com/focusandinsist/go-ws-srv/protocol"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Handler 处理 WebSocket 消息
type Handler struct {
	connMgr      *connection.ConnectionManager
	msgMgr       *message.MessageManager
	roomMgr      *room.RoomManager
	authMgr      *auth.AuthManager
	kafkaBroker  *broker.KafkaBroker
	redisStorage *storage.RedisStorage
	eventMgr     *event.EventManager
	mongoStorage *storage.MongoStorage
}

// NewHandler 创建 Handler 实例
func NewHandler(connMgr *connection.ConnectionManager, msgMgr *message.MessageManager, authMgr *auth.AuthManager, roomMgr *room.RoomManager, kafkaBroker *broker.KafkaBroker, redisStorage *storage.RedisStorage, mongoStorage *storage.MongoStorage) *Handler {
	eventMgr := event.NewEventManager()
	return &Handler{
		connMgr:      connMgr,
		msgMgr:       msgMgr,
		roomMgr:      roomMgr,
		authMgr:      authMgr,
		kafkaBroker:  kafkaBroker,
		redisStorage: redisStorage,
		eventMgr:     eventMgr,
		mongoStorage: mongoStorage,
	}
}

// RegisterEventHandler 注册事件处理器
func (h *Handler) RegisterEventHandler(eventType string, handler func(*connection.Client, *protocol.Message)) {
	h.eventMgr.Register(eventType, handler)
}

// HandleMessage 处理客户端发送的消息
func (h *Handler) HandleMessage(client *connection.Client, data []byte) {
	msg, err := protocol.Decode(data)
	if err != nil {
		log.Println("解析消息失败:", err)
		return
	}

	// 存储消息到 MongoDB
	h.mongoStorage.StoreMessage(msg)

	// 将消息发送到 Kafka
	h.kafkaBroker.SendMessage(string(msg.Data))

	// 如果接收者不在线，存储到 Redis
	if h.connMgr.GetClient(msg.ReceiverID) == nil {
		h.redisStorage.AddOfflineMessage(msg.ReceiverID, string(msg.Data))
	}

	h.eventMgr.Trigger(msg.Event, client, msg)
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

// 新增 ReadPump，让它监听 WebSocket 消息，并调用 HandleMessage
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

		// 收到消息后调用 HandleMessage
		h.HandleMessage(client, msg)
	}
}

// Handler 中负责转发的部分：使用 ConnectionManager 来获取目标连接，然后发送消息
func (h *Handler) BroadcastMessage(client *connection.Client, msg *protocol.Message) {
	// 获取所有连接（这部分由 ConnectionManager 提供接口）
	for _, client := range h.connMgr.GetAllClients() {
		err := client.Conn.WriteMessage(websocket.TextMessage, []byte(msg.Data))
		if err != nil {
			log.Printf("发送消息给用户 %s 失败: %v", client.UserID, err)
		}
	}
}

func (h *Handler) SendDirectMessage(client *connection.Client, msg *protocol.Message) {
	// 例如，假设 msg 中的 Data 或者另有字段指定接收者 ID
	target := h.connMgr.GetClient(msg.ReceiverID)
	if target == nil {
		log.Printf("找不到目标用户: %s", msg.ReceiverID)
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

	offlineMessages, err := h.redisStorage.GetOfflineMessages(client.UserID)
	if err != nil {
		log.Printf("Error getting offline messages for client %s: %v", client.UserID, err)
		return
	}

	for _, msg := range offlineMessages {
		client.SendMessage(websocket.TextMessage, []byte(msg))
	}

	h.redisStorage.ClearOfflineMessages(client.UserID)
}

// 系统3的代码，似乎有点问题？
func (h *Handler) OnMessage(c *connection.Client, rawData []byte) {
	msg, err := protocol.Decode(rawData)
	if err != nil {
		log.Println("协议解析失败:", err)
		return
	}

	// ack
	if msg.Event == "__ack__" && msg.AckID != "" {
		protocol.AckManager.Receive(msg.AckID)
		return
	}

	switch msg.Event {
	case "chat":
		// 正常业务逻辑
	}
}

// 系统3的代码，似乎有点问题？
func (h *Handler) SendDirectMessage2(client *connection.Client, msg *protocol.Message) {
	data := map[string]any{
		"user": "tom",
		"text": "hello",
	}

	// 生成 ackID 并发送消息
	ackID := uuid.NewString()
	msgByte, _ := protocol.Encode("chat", data, false, ackID)

	// 注册 ack 等待
	go func() {
		_, err := protocol.AckManager.Wait()
		if err != nil {
			log.Printf("等待 ack 超时: %v", err)
		} else {
			log.Println("收到 ack 确认")
		}
	}()

	client.SendMessage(websocket.TextMessage, msgByte)
}
