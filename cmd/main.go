package main

import "websocket-server/internal/server"

func main() {
	srv := server.NewServer()
	srv.Start(":8080")
}
