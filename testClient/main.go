package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	// 连接到 WebSocket 服务端
	serverURL := "ws://localhost:8080/ws" // WebSocket 服务器地址
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatal("Dial failed:", err)
	}
	defer conn.Close()

	// 发送一条消息到服务器
	message := []byte("Hello from client")
	err = conn.WriteMessage(websocket.TextMessage, message)
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
			fmt.Printf("Received: %s\n", msg)
		}
	}()

	// 模拟持续发送消息
	for {
		time.Sleep(5 * time.Second) // 每隔 5 秒发送一次消息
		err = conn.WriteMessage(websocket.TextMessage, []byte("Ping"))
		if err != nil {
			log.Fatal("Write failed:", err)
		}
	}
}
