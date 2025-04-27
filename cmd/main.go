package main

import "go-ws-srv/internal/server"

func main() {
	srv := server.NewServer()
	srv.Start(":8080")
}
