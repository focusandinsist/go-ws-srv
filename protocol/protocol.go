package protocol

import (
	"encoding/json"
	"errors"
)

type Message struct {
	Event      string          `json:"event"`
	Namespace  string          `json:"namespace,omitempty"` // 可选
	Ack        bool            `json:"ack,omitempty"`
	AckID      string          `json:"ack_id,omitempty"` // 用于确认机制
	SenderID   string          `json:"sender_id,omitempty"`
	ReceiverID string          `json:"receiver_id,omitempty"`
	Data       json.RawMessage `json:"data"`
}

func Encode(event string, data any, ack bool, ackID string) ([]byte, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	msg := Message{
		Event: event,
		Data:  raw,
		Ack:   ack,
		AckID: ackID,
	}
	return json.Marshal(msg)
}

func Decode(input []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(input, &msg)
	if err != nil {
		return nil, err
	}
	if msg.Event == "" {
		return nil, errors.New("missing event")
	}
	return &msg, nil
}
