package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/angusgmorrison/realworld-go/internal/testutil"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_requestScopedLogger_Printf(t *testing.T) {
	t.Parallel()

	innerLogger := &MockLogger{Buf: &bytes.Buffer{}}
	logger := &requestScopedLogger{
		reqID:       "reqID",
		method:      "GET",
		path:        "path",
		innerLogger: innerLogger,
	}
	logger.Printf("%d %s", 1, "two")
	want := "| reqID | GET | path | 1 two"

	gotLogBytes, gotErr := io.ReadAll(innerLogger.Buf)
	assert.NoError(t, gotErr)
	assert.Equal(t, want, string(gotLogBytes))
}

func Test_RequestScopedLoggerInjection(t *testing.T) {
	t.Parallel()

	logger := &MockLogger{Buf: &bytes.Buffer{}}
	method := http.MethodGet
	path := "/"
	message := "test"
	wantLogs := fmt.Sprintf("| %s | %s | %s | %s", uuid.Nil.String(), method, path, message)

	app := fiber.New()
	app.Get(path, RequestScopedLoggerInjection(logger), func(c *fiber.Ctx) error {
		l := LoggerFrom(c)
		assert.IsType(t, &requestScopedLogger{}, l)

		l.Printf(message)

		gotLogs, gotErr := io.ReadAll(logger.Buf)
		require.NoError(t, gotErr)
		assert.Equal(t, wantLogs, string(gotLogs))

		return nil
	})

	req, err := http.NewRequestWithContext(context.Background(), method, path, http.NoBody)
	require.NoError(t, err)

	_, err = app.Test(req)
	require.NoError(t, err)
}

func Test_LoggerFrom(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		setLogger  func(c *fiber.Ctx) error
		wantLogger Logger
	}{
		{
			name: "logger is present",
			setLogger: func(c *fiber.Ctx) error {
				c.Locals(loggerKey, &requestScopedLogger{})
				return c.Next()
			},
			wantLogger: &requestScopedLogger{},
		},
		{
			name: "logger is not present",
			setLogger: func(c *fiber.Ctx) error {
				return c.Next()
			},
			wantLogger: noOpLogger{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			app.Use(tc.setLogger)
			app.Get("/", func(c *fiber.Ctx) error {
				l := LoggerFrom(c)
				assert.IsType(t, tc.wantLogger, l)
				return nil
			})

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			require.NoError(t, err)

			_, err = app.Test(req, testutil.FiberTestTimeoutMillis)
			require.NoError(t, err)
		})
	}
}
