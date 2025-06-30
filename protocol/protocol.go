package protocol

import (
	"encoding/json"
	"errors"
)

type Message struct {
	Event      string `bson:"event"       json:"event"`
	Namespace  string `bson:"namespace"   json:"namespace,omitempty"` // 可选
	Ack        bool   `bson:"ack"         json:"ack,omitempty"`
	AckID      string `bson:"ack_id"      json:"ack_id,omitempty"` // 用于确认机制
	SenderID   int64  `bson:"sender_id"   json:"sender_id,omitempty"`
	ReceiverID int64  `bson:"receiver_id" json:"receiver_id,omitempty"`
	Data       any    `bson:"data"        json:"data"`
}

// func Encode(event string, data any, ack bool, ackID string) ([]byte, error) {
// 	raw, err := json.Marshal(data)
// 	if err != nil {
// 		return nil, err
// 	}
// 	msg := Message{
// 		Event: event,
// 		Data:  raw,
// 		Ack:   ack,
// 		AckID: ackID,
// 	}
// 	return json.Marshal(msg)
// }

// Encode 把 event 和 data 包成完整消息，data 传结构体或 map 即可
func Encode(event string, data any, ack bool, ackID string) ([]byte, error) {
	msg := Message{
		Event: event,
		Data:  data,
		Ack:   ack,
		AckID: ackID,
	}
	return json.Marshal(msg)
}

// Decode 把 JSON 解成 Message
func Decode(input []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(input, &msg); err != nil {
		return nil, err
	}
	if msg.Event == "" {
		return nil, errors.New("missing event")
	}
	return &msg, nil
}
