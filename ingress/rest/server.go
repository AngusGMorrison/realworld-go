package rest

import (
	"time"

	"github.com/angusgmorrison/realworld/service/user"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Server struct {
	innerServer *fiber.App
}

func NewServer(userService *user.Service) *Server {
	app := fiber.New()
	app.Use(
		logger.New(),
		recover.New(recover.Config{
			EnableStackTrace: true,
		}),
	)

	api := app.Group("/api")

	usersGroup := &usersGroup{service: *userService}
	usersRouter := api.Group("/users")
	usersRouter.Post("login", usersGroup.loginHandler)

	return &Server{
		innerServer: app,
	}
}

func (s *Server) Listen(addr string) error {
	return s.innerServer.Listen(addr)
}

func (s *Server) ShutdownWithTimeout(timeout time.Duration) error {
	return s.innerServer.ShutdownWithTimeout(timeout)
}
