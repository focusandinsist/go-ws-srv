package service

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	proto_chat "websocket-server/api/rest"
	"websocket-server/internal/snowflake"
	"websocket-server/model"
	"websocket-server/protocol"
)

// HandleCGAddOneUser 新增一个用户
func (s *Service) HandleCGAddOneUser(c *gin.Context, input *proto_chat.RegisterRequest) (output *proto_chat.RegisterRequest, err error) {

	uid := snowflake.GlabolSF.Generate()
	user := &model.User{
		UID:      uid,
		NickName: input.NickName,
	}

	userTableName := "user"
	s.h.MongoStorage.StoreMessage(userTableName, user)

	users := s.h.ConnMgr.GetAllOnlineUserIDs()
	c.JSON(http.StatusOK, gin.H{"online_users": users})

	// 协程1 发到eventMgr
	evtType := "getAllOnlines"
	senderID := int64(0)
	client := s.h.ConnMgr.GetClient(senderID)
	msg := &protocol.Message{
		Event:      evtType,
		AckID:      "123",
		SenderID:   senderID,
		ReceiverID: 0, // 发给多人的，这个就临时填空
		Data:       json.RawMessage("getAllOnlines"),
	}
	s.h.EventMgr.Trigger(evtType, client, msg)
	// 协程2 异步其他处理 Kafka(and 存储)
	return
}

// HandleCGLogin 登陆
func (s *Service) HandleCGLogin(c *gin.Context, input *proto_chat.LoginRequest) (output *proto_chat.LoginResponse, err error) {

	// 更新登陆事件，协程同步到其他用户
	senderID := int64(0)

	users := s.h.ConnMgr.GetAllOnlineUserIDs()
	c.JSON(http.StatusOK, gin.H{"online_users": users})

	// 协程1 发到eventMgr
	evtType := "getAllOnlines"
	client := s.h.ConnMgr.GetClient(senderID)
	msg := &protocol.Message{
		Event:      evtType,
		AckID:      "123",
		SenderID:   senderID,
		ReceiverID: int64(0), // 发给多人的，这个就临时填空
		Data:       json.RawMessage("getAllOnlines"),
	}
	s.h.EventMgr.Trigger(evtType, client, msg)
	// 协程2 异步其他处理 Kafka(and 存储)

	return
}
