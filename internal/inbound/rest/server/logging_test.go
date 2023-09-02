package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

func Test_requestScopedLogger_Printf(t *testing.T) {
	t.Parallel()

	innerLogger := &mockLogger{buf: &bytes.Buffer{}}
	logger := &requestScopedLogger{
		reqID:       "reqID",
		method:      "GET",
		path:        "path",
		innerLogger: innerLogger,
	}
	logger.Printf("%d %s", 1, "two")
	want := "| reqID |  GET      | path | 1 two"

	gotLogBytes, gotErr := io.ReadAll(innerLogger.buf)
	assert.NoError(t, gotErr)
	assert.Equal(t, want, string(gotLogBytes))
}

func Test_RequestScopedLogging(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{buf: &bytes.Buffer{}}
	method := http.MethodGet
	path := "/"
	message := "test"
	wantLogs := fmt.Sprintf("| %s |  %-7s  | %s | %s", uuid.Nil.String(), method, path, message)

	app := fiber.New()
	app.Get(path, requestScopedLogging(logger), func(c *fiber.Ctx) error {
		l := LoggerFrom(c)
		assert.IsType(t, &requestScopedLogger{}, l)

		l.Printf(message)

		gotLogs, gotErr := io.ReadAll(logger.buf)
		require.NoError(t, gotErr)
		assert.Equal(t, wantLogs, string(gotLogs))

		return nil
	})

	req, err := http.NewRequestWithContext(context.Background(), method, path, http.NoBody)
	require.NoError(t, err)

	_, err = app.Test(req)
	require.NoError(t, err)
}

type mockLogger struct {
	buf io.ReadWriter
}

func (m *mockLogger) Printf(format string, v ...interface{}) {
	_, _ = fmt.Fprintf(m.buf, format, v...)
}
