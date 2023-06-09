package testutil

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/angusgmorrison/realworld/internal/controller/rest/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
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

// ServerConfig overrides the default test server config, which can be helpful for debugging.
type ServerConfig struct {
	PrintLogs               bool
	PrintRecoveryStackTrace bool
}

var defaultServerConfig = ServerConfig{
	PrintLogs: false,
}

// NewServer requires a new Fiber server for testing handlers.
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

	return app
}
