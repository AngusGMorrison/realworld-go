package rest

import (
	"time"

	"github.com/angusgmorrison/realworld/service/user"
	"github.com/gofiber/fiber/v2"
)

type Server struct {
	innerServer *fiber.App
	userService user.Service
}

func NewServer() *Server {
	return &Server{
		innerServer: fiber.New(),
	}
}

func (s *Server) Listen(addr string) error {
	return s.innerServer.Listen(addr)
}

func (s *Server) ShutdownWithTimeout(timeout time.Duration) error {
	return s.innerServer.ShutdownWithTimeout(timeout)
}
