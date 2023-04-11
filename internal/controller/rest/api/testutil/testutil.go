package testutil

import (
	"crypto"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/angusgmorrison/realworld/internal/controller/rest/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// NewMockAuthMiddleware sets the request token and current user ID, bypassing authentication.
func NewMockAuthMiddleware(t *testing.T, userID uuid.UUID, rawToken string) fiber.Handler {
	t.Helper()

	return func(c *fiber.Ctx) error {
		c.Locals(middleware.UserIDKey, userID)
		c.Locals("user", &jwt.Token{Raw: rawToken})
		return c.Next()
	}
}

func NewRS256AuthMiddleware(t *testing.T, key crypto.PublicKey) fiber.Handler {
	t.Helper()

	return jwtware.New(jwtware.Config{
		SigningKey:    key,
		SigningMethod: "RS256",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Respond with the error details for debugging.
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
		SuccessHandler: middleware.DefaultRS256AuthSuccessHandler,
	})
}

type ServerConfig struct {
	PrintLogs               bool
	PrintRecoveryStackTrace bool
}

var defaultServerConfig = ServerConfig{
	PrintLogs:               false,
	PrintRecoveryStackTrace: false,
}

func NewServer(t *testing.T, cfgOverride ...ServerConfig) *fiber.App {
	t.Helper()

	cfg := defaultServerConfig
	if len(cfgOverride) > 0 {
		cfg = cfgOverride[0]
	}

	app := fiber.New(fiber.Config{
		AppName:      "realworld-hexagonal-test",
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	})

	// Apply middleware.
	app.Use(requestid.New())
	if cfg.PrintLogs {
		app.Use(
			middleware.RequestScopedLogging(
				log.New(os.Stdout, "", log.LstdFlags),
			),
			middleware.RequestStatsLogging(os.Stdout),
		)
	}
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.PrintRecoveryStackTrace,
	}))

	return app
}
