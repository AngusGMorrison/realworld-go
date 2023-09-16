package server

import (
	"fmt"
	"io"

	"github.com/google/uuid"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// Logger represents the minimal set of methods required to log messages.
type Logger interface {
	Printf(format string, v ...interface{})
}

type requestScopedLogger struct {
	reqID       string
	method      string
	path        string
	innerLogger Logger
}

// Printf prepends the request ID to the log message. The formatting is designed
// to align with the request log format used by Fiber's logger middleware.
func (l *requestScopedLogger) Printf(format string, v ...interface{}) {
	formatWithReqFields := fmt.Sprintf("| %s |  %-7s  | %s | %s", l.reqID, l.method, l.path, format)
	l.innerLogger.Printf(formatWithReqFields, v...)
}

type loggerKeyT int

const loggerKey loggerKeyT = 0

// requestScopedLogging is Fiber middleware that adds a request-scoped logger
// containing the current request ID to the Fiber context.
func requestScopedLogging(logger Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqID, ok := c.Locals("requestid").(string)
		if !ok {
			reqID = uuid.Nil.String()
		}
		reqLogger := &requestScopedLogger{
			reqID:       reqID,
			method:      c.Method(),
			path:        c.Path(),
			innerLogger: logger,
		}
		c.Locals(loggerKey, reqLogger)
		return c.Next()
	}
}

// requestStatsLogging wraps the standard Fiber logger middleware, specifying a
// log format that can be reused consistently across the application (e.g.
// between the application server and test servers).
func requestStatsLogging(out io.Writer) fiber.Handler {
	return logger.New(logger.Config{
		Output:     out,
		Format:     "${time} | ${locals:requestid} | ${method} | ${path} | ${status} | ${latency}\n",
		TimeFormat: "2006/01/02 15:04:05",
		TimeZone:   "UTC",
	})
}

// LoggerFrom returns the request-scoped logger from the Fiber context, or
// [noOpLogger] if no logger is present.
func LoggerFrom(c *fiber.Ctx) Logger {
	l, _ := c.Locals(loggerKey).(Logger)
	if l == nil {
		return noOpLogger{}
	}

	return l
}

type noOpLogger struct{}

func (l noOpLogger) Printf(_ string, _ ...interface{}) {}
