package main

import (
	"websocket-server/httpapi"
	"websocket-server/server"
	"websocket-server/service"
)

var serv *server.Server

func main() {
	serv = server.NewServer()

	s := service.NewService()
	// 初始化路由
	httpapi.Init(s)
	serv.Start(":8080")
}
