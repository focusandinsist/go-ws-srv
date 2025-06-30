// 👉 简单说：eventMgr 是一层业务抽象，Kafka 是可能的底层实现。
// 👉 eventMgr 的职责可以总结为：
// 注册和管理事件及其处理器
// 监听事件发生，分发给对应处理器
// 解耦模块间依赖（事件驱动）
// 支持异步处理（提升性能）
// 方便扩展、测试

package event

import (
	"sync"

	"websocket-server/internal/connection"
	"websocket-server/protocol"
)

// EventManager 事件管理器
type EventManager struct {
	handlers map[string]func(*connection.Client, *protocol.Message)
	mu       sync.RWMutex
}

// NewEventManager .
func NewEventManager() *EventManager {
	return &EventManager{
		handlers: make(map[string]func(*connection.Client, *protocol.Message)),
	}
}

// Register 注册
func (em *EventManager) Register(eventType string, handler func(*connection.Client, *protocol.Message)) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.handlers[eventType] = handler
}

// Trigger 触发
func (em *EventManager) Trigger(eventType string, client *connection.Client, msg *protocol.Message) {
	em.mu.RLock()
	handler, exists := em.handlers[eventType]
	em.mu.RUnlock()

	if exists {
		handler(client, msg)
	}
}
