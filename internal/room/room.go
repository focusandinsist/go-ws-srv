package room

import "sync"

// Room 代表一个聊天房间
type Room struct {
	Name    string   // 房间名称
	Members []string // 房间成员（用户 ID）
	mu      sync.Mutex
}

// NewRoom 创建一个新的房间
func NewRoom(name string) *Room {
	return &Room{
		Name:    name,
		Members: make([]string, 0),
	}
}

// AddMember 添加成员到房间
func (r *Room) AddMember(userID string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Members = append(r.Members, userID)
}

// RemoveMember 从房间移除成员
func (r *Room) RemoveMember(userID string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, member := range r.Members {
		if member == userID {
			r.Members = append(r.Members[:i], r.Members[i+1:]...)
			break
		}
	}
}

// GetMembers 获取房间成员
func (r *Room) GetMembers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Members
}
