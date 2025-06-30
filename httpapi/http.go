package httpapi

import (
	"websocket-server/handler"
	"websocket-server/service"

	"github.com/gin-gonic/gin"
)

// serv 服务,给全局用的
var serv *service.Service

// SetupRouter 普通请求
func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/online", HandleCGGetAllOnLineUsers) // 获取所有在线用户
	r.POST("/send", HandleCGSend)                // 给某个用户发送消息
	r.POST("/addoneuser", HandleCGAddOneUser)    // 临时注册
	r.POST("/login", HandleCGLogin)              // 登陆

	// 关于消息，先落库，再推送
	// http.HandleFunc("/ws", s.handler.HandleWebSocket)
	// WebSocket路由，利用Gin处理升级
	r.GET("/ws", func(c *gin.Context) {
		handler.GHandler.HandleWebSocket(c.Writer, c.Request) // 用一个全局的GHandler做这事似乎不太好
	})
	return r
}

func Init(s *service.Service) {
	serv = s
}

// Gin POST /send
//       │
//       ▼
// 调用 eventMgr.Emit("send_message", req)
//       │
//       ▼
// eventMgr 调用已注册的 send_message handler：
//     → 消息存库
//     → 调用 connMgr.SendMessageToUser
//     → 打点、记录日志
//     → 推送给其他服务
