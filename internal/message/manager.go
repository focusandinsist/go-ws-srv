package message

import (
	"log"

	"websocket-server/internal/connection"
)

// MessageManager 负责消息管理
type MessageManager struct {
	clients *connection.ConnectionManager
}

// NewMessageManager 创建消息管理器
func NewMessageManager() *MessageManager {
	return &MessageManager{}
}

// BroadcastMessage 发送广播消息
func (mm *MessageManager) BroadcastMessage(msg *Message) {
	for _, client := range mm.clients.GetAllClients() {
		err := client.Conn.WriteJSON(msg)
		if err != nil {
			log.Println("发送消息失败:", err)
		}
	}
}

// Shutdown 执行消息管理器的清理任务
func (mm *MessageManager) Shutdown() {
	// 执行清理操作，比如停止后台任务、关闭消息队列等
}
