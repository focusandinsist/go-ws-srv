package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// Message 代表 WebSocket 消息
type Message struct {
	Type     string `json:"type"`        // 消息类型
	SenderID string `json:"sender_id"`   // 发送者 ID
	Receiver string `json:"receiver_id"` // 接收者 ID
	RoomID   string `json:"room_id"`     // 房间 ID（可选）
	Data     string `json:"data"`        // 消息内容
}

func main() {
	// 连接到 WebSocket 服务端
	serverURL := "ws://localhost:8080/ws" // WebSocket 服务器地址
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatal("Dial failed:", err)
	}
	defer conn.Close()

	// 发送一条消息到服务器
	msg := &Message{
		Type:     "broadcast",
		SenderID: "LHM",
		RoomID:   "123",
		Data:     "Hello from client",
	}
	// 将结构体转换为 JSON 格式的 []byte
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("JSON 编码失败:", err)
		return
	}
	// message := []byte("Hello from client")
	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Fatal("Write failed:", err)
	}

	// 接收服务器的消息
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Fatal("Read failed:", err)
				return
			}
			fmt.Printf("Received from srv: %s\n", string(msg))
		}
	}()

	// 模拟持续发送消息
	for {
		time.Sleep(5 * time.Second) // 每隔 5 秒发送一次消息

		pingMsg := &Message{
			Type:     "broadcast",
			SenderID: "LHM",
			RoomID:   "123",
			Data:     "Ping",
		}
		// 将结构体转换为 JSON 格式的 []byte
		data, err := json.Marshal(pingMsg)
		if err != nil {
			fmt.Println("JSON 编码失败:", err)
			return
		}

		err = conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Fatal("Write failed:", err)
		}
	}
}
