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
	Receiver string `json:"receiver_id"` // 接收者 ID（可选）
	RoomID   string `json:"room_id"`     // 房间 ID（可选）
	Data     string `json:"data"`        // 消息内容
	AckID    string `json:"ack_id"`      // ACK ID（可选，用于接收时回传 ACK）
}

func main() {
	serverURL := "ws://localhost:8080/ws" // WebSocket 服务器地址
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatal("Dial failed:", err)
	}
	defer conn.Close()

	// 初始发送一条消息
	msg := &Message{
		Type:     "broadcast",
		SenderID: "LHM",
		RoomID:   "123",
		Data:     "Hello from client",
	}
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("JSON 编码失败:", err)
		return
	}
	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Fatal("Write failed:", err)
	}

	// 接收消息 + 自动 ACK
	go func() {
		for {
			_, msgBytes, err := conn.ReadMessage()
			if err != nil {
				log.Fatal("Read failed:", err)
				return
			}
			fmt.Printf("Received from srv: %s\n", string(msgBytes))

			var incoming Message
			if err := json.Unmarshal(msgBytes, &incoming); err != nil {
				fmt.Println("解码失败:", err)
				continue
			}

			// 如果包含 ack_id，自动回 ACK
			if incoming.AckID != "" {
				ackMsg := &Message{
					Type:     "__ack__",
					SenderID: "LHM",          // 用同一个 sender
					AckID:    incoming.AckID, // 原封不动回去
				}
				ackBytes, _ := json.Marshal(ackMsg)
				err := conn.WriteMessage(websocket.TextMessage, ackBytes)
				if err != nil {
					fmt.Println("发送 ACK 失败:", err)
				} else {
					fmt.Printf("发送 ACK: %s\n", incoming.AckID)
				}
			}
		}
	}()

	// 持续发送消息
	for {
		time.Sleep(5 * time.Second)

		pingMsg := &Message{
			Type:     "broadcast",
			SenderID: "LHM",
			RoomID:   "123",
			Data:     "Ping",
		}
		data, err := json.Marshal(pingMsg)
		if err != nil {
			fmt.Println("JSON 编码失败:", err)
			continue
		}

		err = conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Fatal("Write failed:", err)
		}
	}
}
