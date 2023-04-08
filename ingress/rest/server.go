package rest

import (
	"io"
	"os"
	"time"

	"github.com/angusgmorrison/realworld/service/user"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type config struct {
	readTimeout      time.Duration
	writeTimeout     time.Duration
	logOutput        io.Writer
	enableStackTrace bool
}

var defaultConfig = config{
	readTimeout:      5 * time.Second,
	writeTimeout:     10 * time.Second,
	logOutput:        os.Stdout,
	enableStackTrace: true,
}

// Server encapsulates a Fiber app and exposes methods for starting and stopping
// the server.
type Server struct {
	innerServer *fiber.App
}

func NewServer(userService user.Service, opts ...Option) *Server {
	cfg := defaultConfig
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	app := fiber.New(fiber.Config{
		AppName:      "realworld-hexagonal",
		ReadTimeout:  cfg.readTimeout,
		WriteTimeout: cfg.writeTimeout,
	})
	app.Use(
		logger.New(logger.Config{
			Output: cfg.logOutput,
		}),
		recover.New(recover.Config{
			EnableStackTrace: cfg.enableStackTrace,
		}),
	)

	api := app.Group("/api")

	usersGroup := &usersGroup{service: userService}
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
