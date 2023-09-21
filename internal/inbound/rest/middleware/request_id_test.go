package middleware

import (
	"context"
	"net/http"
	"testing"

	"github.com/angusgmorrison/realworld-go/internal/testutil"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_RequestIDInjection(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(RequestIDInjection())
	app.Get("/", func(c *fiber.Ctx) error {
		requestID := c.Locals(RequestIDKey).(string)
		require.NotNil(t, requestID, "request ID was not set on context")

		_, err := uuid.Parse(requestID)
		require.NoError(t, err, "request ID was not a valid UUID")

		return nil
	})

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
	require.NoError(t, err)

	_, err = app.Test(req, testutil.FiberTestTimeoutMillis)
	require.NoError(t, err)
}
