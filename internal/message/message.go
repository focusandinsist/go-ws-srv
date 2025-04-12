// 消息管理 (message.go)
// 职责：处理消息的发送、广播、私聊、频道消息等。可以定义消息队列，处理消息的发送和接收。
// 建议：将消息的广播、私聊等功能封装成 MessageManager，并将其引入到 server.go 中。
package message

// import "encoding/json"

// // Message 代表 WebSocket 消息
// type Message struct {
// 	Type     string `json:"type"`        // 消息类型
// 	SenderID string `json:"sender_id"`   // 发送者 ID
// 	Receiver string `json:"receiver_id"` // 接收者 ID
// 	RoomID   string `json:"room_id"`     // 房间 ID（可选）
// 	Data     string `json:"data"`        // 消息内容
// }

// // ParseMessage 解析 JSON 消息
// func ParseMessage(data []byte) (*Message, error) {
// 	var msg Message
// 	err := json.Unmarshal(data, &msg)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &msg, nil
// }


