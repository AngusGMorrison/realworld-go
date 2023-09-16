// Package server provides the application's HTTP server, wired with routes
// according to the structure of the api package.
package server

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"github.com/angusgmorrison/realworld-go/internal/inbound/rest/api/v0"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"github.com/angusgmorrison/realworld-go/internal/domain/user"
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

// New configures an application server with the injected configuration and
// dependencies.
func New(
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
	router.Use(
		requestid.New(),
		requestScopedLogging(log.New(os.Stdout, "", log.LstdFlags)),
		requestStatsLogging(os.Stdout),
		recover.New(recover.Config{EnableStackTrace: cfg.EnableStackTrace}),
	)

	router.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	router.Route("/api", func(api fiber.Router) {
		api.Use(validateAPIContentType)

		api.Route("/v0", func(apiV0 fiber.Router) {
			usersHandler := v0.NewUsersHandler(
				userService,
				v0.NewJWTProvider(cfg.JwtCfg.RS265PrivateKey, cfg.JwtCfg.TTL, cfg.JwtCfg.Issuer),
			)

			apiV0.Route("/users", func(users fiber.Router) {
				users.Use(v0.UsersErrorHandler)
				users.Post("/", usersHandler.Register)
				users.Post("/login", usersHandler.Login)
			})

			apiV0.Route("/user", func(user fiber.Router) {
				user.Use(v0.NewRS256JWTAuthMiddleware(cfg.JwtCfg.PublicKey()), v0.UsersErrorHandler)
				user.Get("/", usersHandler.GetCurrent)
				user.Put("/", usersHandler.UpdateCurrent)
			})
		})
	})
}

func strictDecoder(b []byte, v any) error {
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.DisallowUnknownFields()
	return decoder.Decode(v) // nolint:wrapcheck
}

var supportedMediaTypes = []string{fiber.MIMEApplicationJSON}

func validateAPIContentType(c *fiber.Ctx) error {
	mediaType := c.Get(fiber.HeaderContentType)
	if !slices.Contains(supportedMediaTypes, mediaType) {
		return v0.NewUnsupportedMediaTypeError(mediaType, supportedMediaTypes)
	}
	return c.Next()
}
