package httpapi

import (
	proto_chat "websocket-server/api/rest"

	"github.com/gin-gonic/gin"
)

// HandleCGAddOneUser 注册
func HandleCGAddOneUser(c *gin.Context) {
	var (
		input proto_chat.RegisterRequest
		// output *proto_chat.RegisterRequest
		err error
	)
	if err = c.Bind(&input); err != nil {
		return
	}
	_, err = serv.HandleCGAddOneUser(c, &input)
}

// HandleCGLogin 登陆
func HandleCGLogin(c *gin.Context) {
	var (
		input proto_chat.LoginRequest
		// output *proto_chat.LoginResponse
		err error
	)
	if err = c.Bind(&input); err != nil {
		return
	}
	serv.HandleCGLogin(c, &input)
}
