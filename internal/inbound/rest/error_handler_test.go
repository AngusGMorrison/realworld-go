package rest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	v0 "github.com/angusgmorrison/realworld-go/internal/inbound/rest/api/v0"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

func Test_newErrorHandler(t *testing.T) {
	t.Parallel()

	t.Run("unhandled errors map to 500 Internal Server Error and the error is logged", func(t *testing.T) {
		t.Parallel()

		logger := &mockLogger{buf: &bytes.Buffer{}}
		handlerErr := errors.New("unhandled")
		expectedLogEntry := fmt.Sprintf("%d %s caused by %v",
			http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), handlerErr.Error())
		app := fiber.New(fiber.Config{
			ErrorHandler: newErrorHandler(),
		})
		app.Use(
			RequestScopedLogging(logger),
		)
		app.Get("/", func(c *fiber.Ctx) error {
			return handlerErr
		})

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
		require.NoError(t, err)

		res, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = res.Body.Close() }()

		// Assert error is mapped to the correct response.
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		assert.Equal(t, fiber.MIMETextPlainCharsetUTF8, res.Header.Get(fiber.HeaderContentType))

		bodyBytes, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Equal(t, string(bodyBytes), http.StatusText(http.StatusInternalServerError))

		// Assert error is logged.
		logEntry, err := io.ReadAll(logger.buf)
		require.NoError(t, err)
		assert.Contains(t, string(logEntry), expectedLogEntry)
	})

	t.Run(
		"user-facing errors map to the appropriate status code and the error is logged",
		func(t *testing.T) {
			t.Parallel()

			logger := &mockLogger{buf: &bytes.Buffer{}}
			handlerErr := v0.NewBadRequestError(errors.New("some cause")).(*v0.UserFacingError)
			wantBody := `{"errors": {"bad_request": ["request body was invalid JSON or contained unknown fields"]}}`
			app := fiber.New(fiber.Config{
				ErrorHandler: newErrorHandler(),
			})
			app.Use(
				RequestScopedLogging(logger),
			)
			app.Get("/", func(c *fiber.Ctx) error {
				return handlerErr
			})

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			require.NoError(t, err)

			res, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = res.Body.Close() }()

			// Assert error is mapped to the correct response.
			assert.Equal(t, handlerErr.StatusCode, res.StatusCode)
			assert.Equal(t, fiber.MIMEApplicationJSON, res.Header.Get(fiber.HeaderContentType))

			bodyBytes, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.JSONEq(t, string(bodyBytes), wantBody)

			// Assert error is logged.
			logEntry, err := io.ReadAll(logger.buf)
			require.NoError(t, err)
			assert.Contains(t, string(logEntry), handlerErr.Error())
		},
	)
}
