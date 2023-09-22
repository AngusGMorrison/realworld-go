package v0

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/angusgmorrison/realworld-go/internal/inbound/rest/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ContentTypeValidation(t *testing.T) {
	t.Parallel()

	requestID := uuid.New().String()
	deterministicRequestIDSetter := func(c *fiber.Ctx) error {
		c.Locals(middleware.RequestIDKey, requestID)
		return c.Next()
	}

	testCases := []struct {
		name            string
		contentType     string
		requestIDSetter func(c *fiber.Ctx) error
		wantErr         error
	}{
		{
			name:            "supported content type",
			contentType:     fiber.MIMEApplicationJSON,
			requestIDSetter: deterministicRequestIDSetter,
			wantErr:         nil,
		},
		{
			name:            "supported content type with suffix",
			contentType:     fiber.MIMEApplicationJSONCharsetUTF8,
			requestIDSetter: deterministicRequestIDSetter,
			wantErr:         nil,
		},
		{
			name:            "unsupported content type",
			contentType:     fiber.MIMETextPlain,
			requestIDSetter: deterministicRequestIDSetter,
			wantErr:         NewUnsupportedContentTypeError(requestID, fiber.MIMETextPlain, supportedContentTypes),
		},
		{
			name:        "unsupported content type with missing request ID on context",
			contentType: fiber.MIMETextPlain,
			requestIDSetter: func(c *fiber.Ctx) error {
				return c.Next()
			},
			wantErr: middleware.ErrMissingRequestID,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{
				ErrorHandler: func(c *fiber.Ctx, err error) error {
					assert.ErrorIs(t, err, tc.wantErr)
					return nil
				},
			})
			app.Use(tc.requestIDSetter, ContentTypeValidation)
			app.Get("/", func(c *fiber.Ctx) error {
				return nil
			})

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			require.NoError(t, err)

			req.Header.Set(fiber.HeaderContentType, tc.contentType)

			_, err = app.Test(req)
			require.NoError(t, err)
		})
	}
}

func Test_ErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("when error is nil, returns nil and logs nothing", func(t *testing.T) {
		t.Parallel()

		app := fiber.New(fiber.Config{
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				panic(fmt.Errorf("received unexpected error: %v", err))
			},
		})
		logger := &middleware.MockLogger{Buf: &bytes.Buffer{}}
		app.Use(middleware.RequestScopedLoggerInjection(logger), ErrorHandling)
		app.Get("/", func(c *fiber.Ctx) error {
			return nil
		})

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
		require.NoError(t, err)

		_, err = app.Test(req)
		require.NoError(t, err)

		loggedBytes, err := io.ReadAll(logger.Buf)
		require.NoError(t, err)
		assert.Empty(t, loggedBytes)
	})

	t.Run("when error is *JSONError, renders and logs the error", func(t *testing.T) {
		t.Parallel()

		app := fiber.New(fiber.Config{
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				panic(fmt.Errorf("received unexpected error: %v", err))
			},
		})
		logger := &middleware.MockLogger{Buf: &bytes.Buffer{}}
		requestID := uuid.New().String()
		requestIDSetter := func(c *fiber.Ctx) error {
			c.Locals(middleware.RequestIDKey, requestID)
			return c.Next()
		}
		app.Use(
			requestIDSetter,
			middleware.RequestScopedLoggerInjection(logger),
			ErrorHandling,
		)
		inputErr := NewBadRequestError(requestID, assert.AnError)
		app.Get("/", func(c *fiber.Ctx) error {
			return inputErr
		})

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
		require.NoError(t, err)

		wantResBody, err := json.Marshal(inputErr)
		require.NoError(t, err)

		res, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = res.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, res.StatusCode)

		gotResBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(t, string(wantResBody), string(gotResBody))

		loggedBytes, err := io.ReadAll(logger.Buf)
		require.NoError(t, err)
		assert.Contains(t, string(loggedBytes), inputErr.Error())
	})

	t.Run("when error is unhandled and requestID is missing, wraps and returns the error", func(t *testing.T) {
		t.Parallel()

		app := fiber.New(fiber.Config{
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				assert.ErrorIs(t, err, middleware.ErrMissingRequestID)
				assert.ErrorIs(t, err, assert.AnError)
				return nil
			},
		})
		logger := &middleware.MockLogger{Buf: &bytes.Buffer{}}
		app.Use(
			middleware.RequestScopedLoggerInjection(logger),
			ErrorHandling,
		)
		app.Get("/", func(c *fiber.Ctx) error {
			return assert.AnError
		})

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
		require.NoError(t, err)

		_, err = app.Test(req)
		require.NoError(t, err)

		loggedBytes, err := io.ReadAll(logger.Buf)
		require.NoError(t, err)
		assert.Empty(t, loggedBytes)
	})

	t.Run(
		"when error is unhandled and request ID is present, logs the error and renders Internal Server Error",
		func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{
				ErrorHandler: func(c *fiber.Ctx, err error) error {
					panic(fmt.Errorf("received unexpected error: %v", err))
				},
			})
			logger := &middleware.MockLogger{Buf: &bytes.Buffer{}}
			requestID := uuid.New().String()
			requestIDSetter := func(c *fiber.Ctx) error {
				c.Locals(middleware.RequestIDKey, requestID)
				return c.Next()
			}
			app.Use(
				requestIDSetter,
				middleware.RequestScopedLoggerInjection(logger),
				ErrorHandling,
			)
			app.Get("/", func(c *fiber.Ctx) error {
				return assert.AnError
			})

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			require.NoError(t, err)

			wantResBody, err := json.Marshal(NewInternalServerError(requestID, assert.AnError))
			require.NoError(t, err)

			res, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = res.Body.Close() }()

			assert.Equal(t, fiber.StatusInternalServerError, res.StatusCode)

			gotResBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.JSONEq(t, string(wantResBody), string(gotResBody))

			loggedBytes, err := io.ReadAll(logger.Buf)
			require.NoError(t, err)
			assert.Contains(t, string(loggedBytes), assert.AnError.Error())
		},
	)
}
