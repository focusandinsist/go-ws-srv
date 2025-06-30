package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

// Message 代表 WebSocket 消息
type Message struct {
	Event    string `json:"event"`       // 事件类型
	SenderID int64  `json:"sender_id"`   // 发送者 ID
	Receiver int64  `json:"receiver_id"` // 接收者 ID（可选）
	RoomID   string `json:"room_id"`     // 房间 ID（可选）
	Data     string `json:"data"`        // 消息内容
	AckID    string `json:"ack_id"`      // ACK ID（可选，用于接收时回传 ACK）
}

var secret = []byte("chatsecret") // 定义签名密钥,这个临时写法，TODO：需要从客户端发过来

func main() {
	fmt.Println("start")
	// header设置用户身份
	userID := int64(12345)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
	})
	tokenString, err := token.SignedString(secret)
	if err != nil {
		log.Fatal(err)
	}
	header := http.Header{}
	header.Set("Authorization", "Bearer "+tokenString)

	serverURL := "ws://localhost:8080/ws" // WebSocket 服务器地址
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, header)
	if err != nil {
		log.Fatal("Dial failed:", err)
	}
	defer conn.Close()

	// 设置 PongHandler 来维持连接，收到 Pong 时刷新读超时
	conn.SetPongHandler(func(appData string) error {
		fmt.Println("pong received")
		return conn.SetReadDeadline(time.Now().Add(70 * time.Second))
	})

	// 初始化读超时，防止连接闲置被关闭
	err = conn.SetReadDeadline(time.Now().Add(70 * time.Second))
	if err != nil {
		log.Fatal("SetReadDeadline failed:", err)
	}

	// 初始发送一条消息
	msg := &Message{
		Event:    "first",
		SenderID: userID,
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
					Event:    "__ack__",
					SenderID: userID,         // 用同一个 sender
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

	// 持续发送心跳
	go func() {
		for {
			time.Sleep(20 * time.Second) // 心跳时间是服务器的三分之一到二分之一
			// WebSocket 协议的 Ping 帧是二进制帧（control frame），不能带太长或者复杂数据，通常最多是125字节，且一般是简短的应用数据（比如空或者少量字节）。
			err = conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Fatal("Write failed:", err)
			}
		}
	}()

	// 持续发送消息
	go func() {
		for {
			time.Sleep(5 * time.Second)

			pingMsg := &Message{
				Event:    "normal",
				SenderID: userID,
				RoomID:   "123",
				Data:     "客户端用来测试，每五秒发一个消息看",
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
	}()
	fmt.Println("finish")
	select {}
}

//🎯 客户端读超时必须大于 Ping 周期
//🎯 服务端默认会回应 Pong，所以你只用客户端发 Ping 就能维持连接
//🎯 没有服务端主动 Ping 时，这就是标准用法
