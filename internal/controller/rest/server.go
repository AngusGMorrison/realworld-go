package rest

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/angusgmorrison/realworld/internal/controller/rest/api/v0/users"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/angusgmorrison/realworld/internal/controller/rest/middleware"
	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

type config struct {
	appName          string
	readTimeout      time.Duration
	writeTimeout     time.Duration
	logPrefix        string
	logFlags         int
	logOutput        io.Writer
	enableStackTrace bool
}

func (c config) applyOptions(opts ...Option) config {
	for _, opt := range opts {
		opt.apply(&c)
	}
	return c
}

var defaultConfig = config{
	appName:          "realworld-hexagonal",
	readTimeout:      5 * time.Second,
	writeTimeout:     10 * time.Second,
	logPrefix:        "",
	logFlags:         log.LstdFlags,
	logOutput:        os.Stdout,
	enableStackTrace: true,
}

// Server encapsulates a Fiber app and exposes methods for starting and stopping
// the server.
type Server struct {
	fiber *fiber.App
	cfg   config
}

// NewServer configures an application server with the injected dependencies. Options may be
// passed to override the server defaults.
func NewServer(
	userService user.Service,
	jwtVerificationKey *rsa.PublicKey,
	opts ...Option,
) *Server {
	cfg := defaultConfig.applyOptions(opts...)
	server := &Server{
		cfg: cfg,
		fiber: fiber.New(fiber.Config{
			AppName:      cfg.appName,
			ReadTimeout:  cfg.readTimeout,
			WriteTimeout: cfg.writeTimeout,
			ErrorHandler: globalErrorHandler,
		}),
	}

	server.applyRoutes(userService, jwtVerificationKey)

	return server
}

// Listen on the given address.
func (s *Server) Listen(addr string) error {
	if err := s.fiber.Listen(addr); err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	return nil
}

// ShutdownWithTimeout gracefully shuts down the server, closing open
// connections at `timeout`.
func (s *Server) ShutdownWithTimeout(timeout time.Duration) error {
	if err := s.fiber.ShutdownWithTimeout(timeout); err != nil {
		return fmt.Errorf("shutdown with timeout %s: %w", timeout, err)
	}
	return nil
}

func (s *Server) applyRoutes(userService user.Service, jwtVerificationKey *rsa.PublicKey) {
	s.useGlobalMiddleware()

	authMW := middleware.NewRS256Auth(jwtVerificationKey)

	// /api
	api := s.fiber.Group("/api/v0")

	// /api/users
	usersHandler := users.NewHandler(userService)
	usersGroup := api.Group("/users")
	usersGroup.Post("/", usersHandler.Register)
	usersGroup.Post("/login", usersHandler.Login)

	// /api/user
	authenticatedUserGroup := api.Group("/user", authMW)
	authenticatedUserGroup.Get("/", usersHandler.GetCurrent)
	authenticatedUserGroup.Put("/", usersHandler.UpdateCurrent)
}

func (s *Server) newLogger() *log.Logger {
	return log.New(s.cfg.logOutput, s.cfg.logPrefix, s.cfg.logFlags)
}

// useGlobalMiddleware applies essential middleware to all routes.
func (s *Server) useGlobalMiddleware() {
	// Order of middleware is important.
	s.fiber.Use(
		// Add a UUID to each request.
		requestid.New(),
		// Add a logger to the context for each request that automatically logs
		// the request's ID.
		middleware.RequestScopedLogging(s.newLogger()),
		// Log request stats.
		middleware.RequestStatsLogging(s.cfg.logOutput),
		// Recover from panics.
		recover.New(recover.Config{
			EnableStackTrace: s.cfg.enableStackTrace,
		}),
	)
}

// The top-level error api for the server.
func globalErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	var e *fiber.Error
	if ok := errors.As(err, &e); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error": http.StatusText(code),
	})
}
