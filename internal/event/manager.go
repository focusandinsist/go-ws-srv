package event

import (
	"sync"

	"go-ws-srv/internal/connection"
	"go-ws-srv/protocol"
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
