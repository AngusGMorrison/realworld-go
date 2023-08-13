package rest

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/angusgmorrison/realworld/internal/inbound/rest/api/v0"
	"log"
	"os"
	"time"

	"github.com/angusgmorrison/realworld/internal/domain/user"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

type JWTConfig struct {
	RS265PrivateKey *rsa.PrivateKey
	TTL             time.Duration
	Issuer          string
}

func (cfg JWTConfig) PublicKey() *rsa.PublicKey {
	return &cfg.RS265PrivateKey.PublicKey
}

type Config struct {
	AppName          string
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	JwtCfg           JWTConfig
	EnableStackTrace bool
}

// Server encapsulates a Fiber app and exposes methods for starting and stopping
// the server.
type Server struct {
	app *fiber.App
	cfg Config
}

// NewServer configures an application server with the injected dependencies. Options may be
// passed to override the server defaults.
func NewServer(
	cfg Config,
	userService user.Service,
) *Server {
	app := fiber.New(fiber.Config{
		AppName:      cfg.AppName,
		ErrorHandler: newErrorHandler(),
		JSONDecoder:  strictDecoder,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	initRouter(app, cfg, userService)

	return &Server{app: app, cfg: cfg}
}

// Listen on the given address.
func (s *Server) Listen(addr string) error {
	if err := s.app.Listen(addr); err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	return nil
}

// ShutdownWithTimeout gracefully shuts down the server, closing open
// connections at `timeout`.
func (s *Server) ShutdownWithTimeout(timeout time.Duration) error {
	if err := s.app.ShutdownWithTimeout(timeout); err != nil {
		return fmt.Errorf("shutdown with timeout %s: %w", timeout, err)
	}
	return nil
}

func initRouter(router fiber.Router, cfg Config, userService user.Service) {
	initGlobalMiddleware(router, cfg)

	// /api
	api := router.Group("/api")
	api.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	// /api/v0
	initV0Routes(api.Group("/v0"), cfg, userService)
}

func initGlobalMiddleware(router fiber.Router, cfg Config) {
	// Order of middleware is important.
	router.Use(
		// Add a UUID to each request.
		requestid.New(),
		// Add a logger to the context for each request that automatically logs
		// the request's ID.
		RequestScopedLogging(log.New(os.Stdout, "", log.LstdFlags)),
		// Log request stats.
		RequestStatsLogging(os.Stdout),
		// Recover from panics.
		recover.New(recover.Config{
			EnableStackTrace: cfg.EnableStackTrace,
		}),
	)
}

func initV0Routes(router fiber.Router, cfg Config, userService user.Service) {
	handler := v0.NewUsersHandler(
		userService,
		v0.NewJWTProvider(
			cfg.JwtCfg.RS265PrivateKey,
			cfg.JwtCfg.TTL,
			cfg.JwtCfg.Issuer,
		),
	)

	// /api/v0/users
	usersGroup := router.Group("/users", v0.UsersErrorHandler)
	usersGroup.Post("/", handler.Register)
	usersGroup.Post("/login", handler.Login)

	// /api/v0/user
	authenticatedUserGroup := router.Group(
		"/user",
		v0.NewRS256JWTAuthMiddleware(cfg.JwtCfg.PublicKey()),
		v0.UsersErrorHandler,
	)
	authenticatedUserGroup.Get("/", handler.GetCurrent)
	authenticatedUserGroup.Put("/", handler.UpdateCurrent)
}

func strictDecoder(b []byte, v any) error {
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}
