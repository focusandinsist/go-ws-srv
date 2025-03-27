package room

import "sync"

// RoomManager 管理多个房间
type RoomManager struct {
	rooms map[string]*Room
	mu    sync.Mutex
}

// NewRoomManager 创建房间管理器
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

// CreateRoom 创建新房间
func (rm *RoomManager) CreateRoom(name string) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.rooms[name]; exists {
		return nil // 房间已存在
	}
	room := NewRoom(name)
	rm.rooms[name] = room
	return room
}

// GetRoom 获取指定房间
func (rm *RoomManager) GetRoom(name string) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.rooms[name]
}

// DeleteRoom 删除房间
func (rm *RoomManager) DeleteRoom(name string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.rooms, name)
}
