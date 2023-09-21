package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/angusgmorrison/realworld-go/internal/inbound/rest/middleware"
	"github.com/gofiber/fiber/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_JWTConfig_PublicKey(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	cfg := JWTConfig{RS265PrivateKey: privateKey}

	got := cfg.PublicKey()
	assert.Equal(t, &privateKey.PublicKey, got)
}

func Test_globalErrorHandler(t *testing.T) {
	t.Parallel()

	logger := &middleware.MockLogger{Buf: &bytes.Buffer{}}
	unhandledErr := errors.New("unhandled")
	expectedLogEntry := unhandledErr.Error()
	expectedResBody := `{"status":500,"message":"Internal Server Error"}`
	app := fiber.New(fiber.Config{
		ErrorHandler: globalErrorHandler,
	})
	app.Use(middleware.RequestScopedLoggerInjection(logger))
	app.Get("/", func(c *fiber.Ctx) error {
		return unhandledErr
	})

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
	require.NoError(t, err)

	res, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	// Assert error is mapped to the correct response.
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	bodyBytes, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.JSONEq(t, expectedResBody, string(bodyBytes))

	// Assert error is logged.
	logEntry, err := io.ReadAll(logger.Buf)
	require.NoError(t, err)
	assert.Contains(t, string(logEntry), expectedLogEntry)
}
