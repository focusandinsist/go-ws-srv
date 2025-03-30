package connection

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client 代表单个 WebSocket 连接及其状态
type Client struct {
	Conn     *websocket.Conn // WebSocket 连接
	UserID   string          // 用户 ID
	lastPong time.Time       // 上次收到 pong 的时间
	mu       sync.Mutex      // 保护并发写入和状态更新
}

// NewClient 创建一个新的 Client 实例
func NewClient(conn *websocket.Conn, userID string) *Client {
	return &Client{
		Conn:     conn,
		UserID:   userID,
		lastPong: time.Now(),
	}
}

// StartHeartbeat 开启心跳检测，定期发送 ping 消息并检查 pong 响应
func (c *Client) StartHeartbeat() {
	// 设置 pong 处理函数，更新 lastPong 时间
	c.Conn.SetPongHandler(func(appData string) error {
		c.mu.Lock()
		c.lastPong = time.Now()
		c.mu.Unlock()
		return nil
	})

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 检查是否超时（例如超过 60 秒未收到 pong）
			c.mu.Lock()
			if time.Since(c.lastPong) > 60*time.Second {
				c.mu.Unlock()
				log.Printf("Heartbeat timeout for client %s", c.UserID)
				c.Conn.Close()
				return
			}
			c.mu.Unlock()

			// 发送 ping 消息
			if err := c.Conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				log.Printf("Error sending ping to client %s: %v", c.UserID, err)
				c.Conn.Close()
				return
			}
		}
	}
}

// ReadPump 持续读取客户端消息，并调用传入的处理函数
// handleFunc 参数允许调用者处理读取到的消息（例如调用 Handler.HandleMessage）
func (c *Client) ReadPump(handleFunc func(messageType int, data []byte)) {
	defer func() {
		c.Conn.Close()
	}()

	for {
		msgType, data, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from client %s: %v", c.UserID, err)
			break
		}
		// 调用传入的消息处理函数
		handleFunc(msgType, data)
	}
}

// SendMessage 通过 WebSocket 发送消息给客户端
func (c *Client) SendMessage(messageType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteMessage(messageType, data)
}
