package protocol

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ackEntry struct {
	ch   chan struct{}
	once sync.Once
}

type ackManager struct {
	mu   sync.Mutex
	acks map[string]*ackEntry
	ttl  time.Duration
	// acks sync.Map // key: string -> chan *Message // TODO 这个是合理方案
}

var AckManager = NewAckManager(5 * time.Second)

func NewAckManager(ttl time.Duration) *ackManager {
	return &ackManager{
		acks: make(map[string]*ackEntry),
		ttl:  ttl,
	}
}

// Wait 生成一个 ackID 并阻塞等待 ack 返回，超时返回错误
func (m *ackManager) Wait() (string, error) {
	ackID := uuid.NewString()

	m.mu.Lock()
	entry := &ackEntry{ch: make(chan struct{})}
	m.acks[ackID] = entry
	m.mu.Unlock()

	timer := time.NewTimer(m.ttl)
	defer timer.Stop()

	select {
	case <-entry.ch:
		return ackID, nil
	case <-timer.C:
		m.mu.Lock()
		delete(m.acks, ackID)
		m.mu.Unlock()
		return "", errors.New("ack timeout")
	}
}

// Receive 表示收到了 ack，释放等待的协程
func (m *ackManager) Receive(ackID string) {
	m.mu.Lock()
	entry, ok := m.acks[ackID]
	if ok {
		delete(m.acks, ackID)
	}
	m.mu.Unlock()

	if ok {
		entry.once.Do(func() {
			close(entry.ch)
		})
	}
}
