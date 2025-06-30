// 事件处理 (handler.go)
// 职责：处理 WebSocket 消息事件的路由，负责根据消息类型分发不同的处理函数。
// 建议：将处理不同消息类型的函数封装成不同的处理器，使用策略模式或函数映射进行动态分发。这些事件处理函数会根据接收到的消息类型（如私聊、群聊等）进行不同的处理。
package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"websocket-server/internal/auth"
	"websocket-server/internal/broker"
	"websocket-server/internal/connection"
	"websocket-server/internal/event"
	"websocket-server/internal/message"
	"websocket-server/internal/room"
	"websocket-server/internal/storage"
	"websocket-server/protocol"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var GHandler *Handler

// Handler 处理 WebSocket 消息
type Handler struct {
	ConnMgr      *connection.ConnectionManager
	MsgMgr       *message.MessageManager
	RoomMgr      *room.RoomManager
	AuthMgr      *auth.AuthManager
	KafkaBroker  *broker.KafkaBroker
	RedisStorage *storage.RedisStorage
	EventMgr     *event.EventManager
	MongoStorage *storage.MongoStorage
}

// NewHandler 创建 Handler 实例
func NewHandler(connMgr *connection.ConnectionManager, msgMgr *message.MessageManager, authMgr *auth.AuthManager, roomMgr *room.RoomManager, kafkaBroker *broker.KafkaBroker, redisStorage *storage.RedisStorage, mongoStorage *storage.MongoStorage) *Handler {
	eventMgr := event.NewEventManager()
	return &Handler{
		ConnMgr:      connMgr,
		MsgMgr:       msgMgr,
		RoomMgr:      roomMgr,
		AuthMgr:      authMgr,
		KafkaBroker:  kafkaBroker,
		RedisStorage: redisStorage,
		EventMgr:     eventMgr,
		MongoStorage: mongoStorage,
	}
}

// RegisterEventHandler 注册事件处理器
func (h *Handler) RegisterEventHandler(eventType string, handler func(*connection.Client, *protocol.Message)) {
	h.EventMgr.Register(eventType, handler)
}

// HandleMessage 处理客户端发送的消息
func (h *Handler) HandleMessage(client *connection.Client, data []byte) {
	msg, err := protocol.Decode(data)
	if err != nil {
		log.Println("解析消息失败:", err)
		return
	}

	if msg.SenderID != client.UserID {
		log.Println("mismatch userID")
		return
	}

	// 存储消息到 MongoDB
	tableName := "message"
	h.MongoStorage.StoreMessage(tableName, msg)

	// 将消息发送到 Kafka
	h.KafkaBroker.MessageProduce(msg)

	// 如果接收者不在线，存储到 Redis
	if h.ConnMgr.GetClient(msg.ReceiverID) == nil {
		h.RedisStorage.AddOfflineMessage(msg.ReceiverID, msg)
	}

	h.EventMgr.Trigger(msg.Event, client, msg)
}

// HandleWebSocket 处理 WebSocket 请求
var secret = []byte("chatsecret") // 定义签名密钥
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, err := ExtractUserIDFromRequest(r, secret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	log.Printf("User %d connected\n", userID)
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

	reconnect := r.URL.Query().Get("reconnect") // "true" 表示重连
	newClient := connection.NewClient(conn, userID)

	// 在连接管理器中注册新的连接
	h.ConnMgr.AddClient(newClient)

	// 🔥 在这里设置 PingHandler，这样客户端发 Ping 时我们更新 LastPong
	// 当服务器收到客户端 Ping 时，它自动返回 Pong，前提是你没有用 SetPingHandler() 把默认行为覆盖掉。
	// 但是你服务端现在用 SetPingHandler() 来更新时间戳，这会覆盖默认返回 Pong 的逻辑！也就是导致服务端根本没有返回 Pong，所以客户端 SetPongHandler() 永远触发不到，70 秒超时。
	conn.SetPingHandler(func(appData string) error {
		newClient.Mu.Lock()
		newClient.LastPong = time.Now()
		newClient.Mu.Unlock()
		// 🔥 必须手动回复 Pong
		err := conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(time.Second))
		if err != nil {
			log.Println("send pong failed:", err)
		}
		fmt.Println("received ping and pong sent, updated LastPong")
		return nil
	})
	// conn.SetPongHandler(func(appData string) error {
	// 	newClient.Mu.Lock()
	// 	newClient.LastPong = time.Now()
	// 	newClient.Mu.Unlock()
	// 	fmt.Println("set pong handler")
	// 	return nil
	// })
	log.Println("WebSocket connection established")

	// 如果是断线重连，则恢复之前状态
	if reconnect == "true" {
		h.RestoreClientState(newClient)
	}

	// 启动心跳检测
	go newClient.StartHeartbeat(h.ConnMgr)

	// **启动 ReadPump，让它监听消息**
	go h.ReadPump(newClient)
}

// ReadPump 读泵，让它监听 WebSocket 消息，并调用 HandleMessage
func (h *Handler) ReadPump(client *connection.Client) {
	defer func() {
		h.ConnMgr.RemoveClient(client)
		client.Conn.Close()
	}()

	for {
		_, msg, err := client.Conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		// if msgType == websocket.PingMessage {
		// 	client.Mu.Lock()
		// 	client.LastPong = time.Now()
		// 	client.Mu.Unlock()
		// 	continue
		// }

		// 收到消息后调用 HandleMessage
		h.HandleMessage(client, msg)
	}
}

// BroadcastMessage Handler 中负责转发的部分：使用 ConnectionManager 来获取目标连接，然后发送消息
func (h *Handler) BroadcastMessage(client *connection.Client, msg *protocol.Message) {
	// 获取所有连接（这部分由 ConnectionManager 提供接口）
	jsonBytes, err := json.Marshal(msg) // 传msg或者*msg都可以，因为Marshal会自动解引用指针
	if err != nil {
		return
	}
	for _, client := range h.ConnMgr.GetAllClients() {
		err := client.Conn.WriteMessage(websocket.TextMessage, jsonBytes)
		if err != nil {
			log.Printf("发送消息给用户 %d 失败: %v", client.UserID, err)
		}
	}
}

// SendDirectMessage 单发消息
func (h *Handler) SendDirectMessage(client *connection.Client, msg *protocol.Message) {
	// 例如，假设 msg 中的 Data 或者另有字段指定接收者 ID
	target := h.ConnMgr.GetClient(msg.ReceiverID)
	if target == nil {
		log.Printf("找不到目标用户: %d", msg.ReceiverID)
		return
	}

	raw, err := json.Marshal(msg)
	if err != nil {
		log.Printf("序列化消息失败: %v", err)
		return
	}
	// 这两种写法都可以，回家取舍一下
	// err := target.Conn.WriteMessage(websocket.TextMessage, []byte(msg.Data))
	err = target.SendMessage(websocket.TextMessage, raw)
	if err != nil {
		log.Printf("发送消息给用户 %d 失败: %v", target.UserID, err)
	}
}

// RestoreClientState 是个伪函数，用于恢复客户端状态
func (h *Handler) RestoreClientState(client *connection.Client) {
	// 示例：从存储中获取该用户之前订阅的房间、离线消息等
	// 这里的具体实现需要你根据业务逻辑来编写
	log.Printf("Restoring state for client %d", client.UserID)
	// 例如：重新加入房间
	// room := h.roomMgr.GetRoom("exampleRoom")
	// room.AddMember(client.UserID)
	// 发送离线消息
	// offlineMessages := h.msgMgr.GetOfflineMessages(client.UserID)
	// for _, m := range offlineMessages {
	//     client.Conn.WriteMessage(websocket.TextMessage, m.Data)
	// }

	offlineMessages, err := h.RedisStorage.GetOfflineMessages(client.UserID)
	if err != nil {
		log.Printf("Error getting offline messages for client %d: %v", client.UserID, err)
		return
	}

	for _, msg := range offlineMessages {
		client.SendMessage(websocket.TextMessage, []byte(msg)) // 这里有问题，不能直接传msg
	}

	h.RedisStorage.ClearOfflineMessages(client.UserID)
}

// ExtractUserIDFromRequest 从请求中取 JWT Token 并解析 user_id
func ExtractUserIDFromRequest(r *http.Request, secret []byte) (int64, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return 0, fmt.Errorf("missing auth token")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid claims")
	}

	// 类型转换成 int64
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("user_id not found or not a number")
	}
	return int64(userIDFloat), nil
}

// ------------------------------ 无用代码 ------------------------------

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
