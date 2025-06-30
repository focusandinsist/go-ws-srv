package service

import (
	"websocket-server/handler"
)

type Service struct {
	h *handler.Handler
}

func NewService() *Service {
	s := &Service{
		h: handler.GHandler,
	}
	return s
}
