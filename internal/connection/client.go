package connection

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Ping 和 Pong 都是控制帧（control frames），它们不暴露给应用层，
// 也就是你用 conn.ReadMessage() 只会读到 TextMessage 或 BinaryMessage，永远也读不到 Ping 和 Pong。

// Client 代表单个 WebSocket 连接及其状态
type Client struct {
	Conn     *websocket.Conn // WebSocket 连接
	UserID   int64           // 用户ID
	Mu       sync.Mutex      // 保护并发写入和状态更新
	LastPong time.Time       // 上次收到 pong 的时间
}

// NewClient 创建一个新的 Client 实例
func NewClient(conn *websocket.Conn, userID int64) *Client {
	return &Client{
		Conn:     conn,
		UserID:   userID,
		LastPong: time.Now(),
	}
}

// StartHeartbeat 开启心跳检测，定期发送 ping 消息并检查 pong 响应
// 问：我服务器这样for循环监听客户端消息，对于一个社交服务器来说，是不是太耗费性能了？其他IM也是这么做的吗
// 答案是：这是业界标准写法，不算耗费性能。
// 原因：
// ✅ ReadMessage() 是阻塞式 I/O，底层 epoll/kqueue，等待网络事件，不是忙轮询，不浪费 CPU
// ✅ goroutine 调度代价在 Go 中很低（每个连接一个 goroutine 是常态）
// ✅ 微信、Telegram 都是每个长连接维护一个读 goroutine，这就是标准架构
// ✅ 真正限制服务器规模的往往是网络带宽、消息处理逻辑、序列化/反序列化，不是这个监听 goroutine
func (c *Client) StartHeartbeat(connMgr *ConnectionManager) {
	defer func() {
		connMgr.RemoveClient(c)
		c.Conn.Close()
	}()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 检查是否超时（例如超过 60 秒未收到 pong）
			c.Mu.Lock()
			if time.Since(c.LastPong) > 60*time.Second {
				c.Mu.Unlock()
				log.Printf("Heartbeat timeout for client %d", c.UserID)
				c.Conn.Close()
				return
			}
			c.Mu.Unlock()
		}
	}
}

// SendMessage 通过 WebSocket 发送消息给客户端
func (c *Client) SendMessage(messageType int, data []byte) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	return c.Conn.WriteMessage(messageType, data)
}

// ---------------------------- 无用代码 ----------------------------

// ReadPump 持续读取客户端消息，并调用传入的处理函数
// handleFunc 参数允许调用者处理读取到的消息（例如调用 Handler.HandleMessage）
func (c *Client) ReadPump(handleFunc func(messageType int, data []byte)) {
	defer func() {
		c.Conn.Close()
	}()

	for {
		msgType, data, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from client %d: %v", c.UserID, err)
			break
		}
		if msgType == websocket.PingMessage {
			c.Mu.Lock()
			c.LastPong = time.Now()
			c.Mu.Unlock()
			return
		}
		// 调用传入的消息处理函数
		handleFunc(msgType, data)
	}
}

// 什么时候关闭连接？是for循环读的失败？还是心跳超时？
// 两者结合使用，相辅相成：
// | 情况    | 关闭时机       | 作用       |
// | -----   | ---------- | -------- |
// | 读消息失败 | 立即关闭       | 快速发现连接断开 |
// | 心跳超时  | 定时检测到超时后关闭 | 发现“假死”连接 |
