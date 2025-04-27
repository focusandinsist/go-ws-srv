package message

import "websocket-server/protocol"

// MessageManager 负责消息管理
type MessageManager struct {
	messages []*protocol.Message
}

// NewMessageManager 创建消息管理器
func NewMessageManager() *MessageManager {
	return &MessageManager{}
}

// Shutdown 执行消息管理器的清理任务
func (mm *MessageManager) Shutdown() {
	// 执行清理操作，比如停止后台任务、关闭消息队列等
}

// 可选：将消息存储到内部数组中
func (mm *MessageManager) StoreMessage(msg *protocol.Message) {
	mm.messages = append(mm.messages, msg)
}
