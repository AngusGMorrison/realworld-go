package rest

import (
	"io"
	"os"
	"time"

	"github.com/angusgmorrison/realworld/internal/ingress/rest/api/users"
	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	jwtware "github.com/gofiber/jwt/v3"
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

func NewServer(
	userService user.Service,
	signingKey string,
	signingAlg string,
	opts ...Option,
) *Server {
	cfg := defaultConfig
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	app := fiber.New(fiber.Config{
		AppName:      "realworld-hexagonal",
		ReadTimeout:  cfg.readTimeout,
		WriteTimeout: cfg.writeTimeout,
	})

	applyRoutes(app, cfg, userService, signingKey, signingAlg)

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

func applyRoutes(
	app *fiber.App,
	cfg config,
	userService user.Service,
	signingKey string,
	signingAlg string,
) {
	authMW := jwtware.New(jwtware.Config{
		SigningKey:    signingKey,
		SigningMethod: signingAlg,
	})

	app.Use(
		logger.New(logger.Config{
			Output: cfg.logOutput,
		}),
		recover.New(recover.Config{
			EnableStackTrace: cfg.enableStackTrace,
		}),
	)

	// /api
	api := app.Group("/api")

	// /api/usersGroup
	usersGroup := api.Group("/users")
	usersHandler := users.NewHandler(userService)
	// Unauthenticated routes.
	usersGroup.Post("/", usersHandler.Register)
	usersGroup.Post("/login", usersHandler.Login)
	// Authenticated routes.
	usersGroup.Use(authMW)
	// usersGroup.Get("/", usersHandler.GetCurrentUser)
}
