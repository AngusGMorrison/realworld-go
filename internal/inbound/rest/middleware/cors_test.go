package middleware

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CORS(t *testing.T) {
	t.Parallel()

	t.Run("OPTIONS request", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		app.Use(CORS("*"))

		req, err := http.NewRequestWithContext(context.Background(), http.MethodOptions, "/", http.NoBody)
		require.NoError(t, err)

		resp, err := app.Test(req)
		require.NoError(t, err)

		assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
		assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "DELETE,GET,OPTIONS,PATCH,POST,PUT", resp.Header.Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Origin,Authorization", resp.Header.Get("Access-Control-Allow-Headers"))
	})
}
