// 连接管理 (connection.go)
// 职责：负责处理 WebSocket 连接的建立、维护和关闭。通常也负责心跳检测和存储在线用户。
// 建议：可以封装成一个连接管理器（ConnectionManager），负责管理所有活动连接。这个管理器会在 server.go 中被创建并管理。
package connection

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Client 代表一个 WebSocket 连接
type Client struct {
	Conn   *websocket.Conn
	UserID string
}

// ConnectionManager 管理所有连接的 WebSocket 客户端
type ConnectionManager struct {
	clients map[string]*Client
	mu      sync.Mutex
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		clients: make(map[string]*Client),
	}
}

// AddClient 添加新客户端
func (cm *ConnectionManager) AddClient(client *Client) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.clients[client.UserID] = client
}

// RemoveClient 移除客户端
func (cm *ConnectionManager) RemoveClient(client *Client) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.clients, client.UserID)
}

// GetClient 获取特定用户的连接
func (cm *ConnectionManager) GetClient(userID string) *Client {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.clients[userID]
}

// GetAllClients 获取所有连接的客户端
func (cm *ConnectionManager) GetAllClients() []*Client {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	clients := make([]*Client, 0, len(cm.clients))
	for _, client := range cm.clients {
		clients = append(clients, client)
	}
	return clients
}

// CloseConnection 关闭单个连接
func (cm *ConnectionManager) CloseConnection(userID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 检查连接是否存在
	targetClient, exists := cm.clients[userID]
	if !exists {
		return fmt.Errorf("connection not found")
	}

	// 关闭连接
	err := targetClient.Conn.Close()
	if err != nil {
		log.Printf("Error closing connection: %v", err)
		return err
	}

	// 移除该连接
	delete(cm.clients, userID)
	log.Printf("Connection closed: %s", userID)
	return nil
}

// CloseAllConnections 关闭所有连接
func (cm *ConnectionManager) CloseAllConnections() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, client := range cm.clients {
		err := client.Conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}
	clear(cm.clients) // 清空所有连接
}
