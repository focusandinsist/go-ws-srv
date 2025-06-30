package service

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"websocket-server/protocol"
)

// HandleCGGetAllOnLineUsers 获取所有在线用户
func (s *Service) HandleCGGetAllOnLineUsers(c *gin.Context) {
	senderID := int64(1231321)

	users := s.h.ConnMgr.GetAllOnlineUserIDs()
	c.JSON(http.StatusOK, gin.H{"online_users": users})

	// 协程1 发到eventMgr
	evtType := "getAllOnlines"
	client := s.h.ConnMgr.GetClient(senderID)
	msg := &protocol.Message{
		Event:      evtType,
		AckID:      "123",
		SenderID:   senderID,
		ReceiverID: 0, // 发给多人的，这个就临时0
		Data:       json.RawMessage("getAllOnlines"),
	}
	s.h.EventMgr.Trigger(evtType, client, msg)
	// 协程2 异步其他处理 Kafka(and 存储)
}

// HandleCGSend 给某个用户发送消息
// 先把一条线捋顺，就这个sendMsg,长连，kafka,
func (s *Service) HandleCGSend(c *gin.Context) {
	// 要从c中获取到sender
	senderID := int64(123213132423)
	var req struct {
		UserID  int64  `json:"user_id"`
		Message string `json:"message"`
	}
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 主任务，发消息
	err := s.h.ConnMgr.SendMessageToUser(req.UserID, []byte(req.Message))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 协程1 发到eventMgr
	evtType := "send"
	client := s.h.ConnMgr.GetClient(req.UserID)
	msg := &protocol.Message{
		Event:      evtType,
		AckID:      "123",
		SenderID:   senderID,
		ReceiverID: req.UserID,
		Data:       json.RawMessage(req.Message),
	}
	s.h.EventMgr.Trigger(evtType, client, msg)

	// 协程2 异步其他处理 Kafka(and 存储)
	// TODO
	c.JSON(http.StatusOK, gin.H{"status": "sent"})
}

// 一个理想化的框架写法（伪代码）
// func (h *Handler) HandleMessage(c *Client, raw []byte) {
//   msg := protocol.Decode(raw)

//   // 先触发本地 handler（即时响应）
//   h.eventMgr.Trigger(msg.Event, c, msg)

//   // 如果配置要求投递到 Kafka，投
//   if h.config.EnableKafka {
//     h.kafkaBroker.SendMessage(msg)
//   }

//   // 如果配置要求异步内部队列，投
//   if h.config.EnableInternalQueue {
//     h.queue.Push(msg)
//   }
// }
