// Package server provides the application's HTTP server, wired with routes
// according to the structure of the api package.
package server

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/angusgmorrison/realworld-go/internal/inbound/rest/middleware"

	"github.com/angusgmorrison/realworld-go/internal/inbound/rest/api/v0"

	"github.com/angusgmorrison/realworld-go/internal/domain/user"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
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
	AllowOrigins     string
}

// Server encapsulates a Fiber app and exposes methods for starting and stopping
// the server.
type Server struct {
	app *fiber.App
	cfg Config
}

// New configures an application server with the injected configuration and
// dependencies.
func New(
	cfg Config,
	userService user.Service,
) *Server {
	app := fiber.New(fiber.Config{
		AppName:      cfg.AppName,
		ErrorHandler: globalErrorHandler,
		JSONDecoder:  decodeStrict,
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
	router.Use(
		middleware.RequestIDInjection(),
		middleware.RequestScopedLoggerInjection(log.New(os.Stdout, "", log.LstdFlags)),
		middleware.RequestStatsLogging(os.Stdout),
		recover.New(recover.Config{
			EnableStackTrace: cfg.EnableStackTrace,
		}),
		middleware.CORS(cfg.AllowOrigins),
	)

	router.Head("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusMethodNotAllowed)
	})

	router.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	router.Route("/api", func(api fiber.Router) {
		api.Route("/v0", func(apiV0 fiber.Router) {
			apiV0.Use(v0.ErrorHandling, v0.ContentTypeValidation)

			usersHandler := v0.NewUsersHandler(
				userService,
				v0.NewJWTProvider(cfg.JwtCfg.RS265PrivateKey, cfg.JwtCfg.TTL, cfg.JwtCfg.Issuer),
			)

			apiV0.Route("/users", func(users fiber.Router) {
				users.Use(v0.UsersErrorHandling)
				users.Post("/", usersHandler.Register)
				users.Post("/login", usersHandler.Login)
			})

			apiV0.Route("/user", func(user fiber.Router) {
				user.Use(v0.Authentication(cfg.JwtCfg.PublicKey()), v0.UsersErrorHandling)
				user.Get("/", usersHandler.GetCurrent)
				user.Put("/", usersHandler.UpdateCurrent)
			})
		})
	})

	router.Use(notFoundHandler)
}

func decodeStrict(b []byte, v any) error {
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()
	return decoder.Decode(v) // nolint:wrapcheck
}

// globalErrorHandler is the top-level error handler for the application. It
// mops up all errors that API error handlers were unable to handle.
func globalErrorHandler(c *fiber.Ctx, err error) error {
	middleware.LoggerFrom(c).Printf("%v\n", err)

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"status":  fiber.StatusInternalServerError,
		"message": http.StatusText(fiber.StatusInternalServerError),
	})
}

func notFoundHandler(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"status":  fiber.StatusNotFound,
		"message": fmt.Sprintf("Endpoint %q not found.", c.Path()),
	})
}
