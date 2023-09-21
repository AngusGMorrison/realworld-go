package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/utils"
)

// RequestIDKey is the fiber context key for the request ID.
const RequestIDKey string = "requestid"

// RequestIDInjection returns middleware injects a UUID into the context of each
// request.
func RequestIDInjection() fiber.Handler {
	return requestid.New(
		requestid.Config{
			ContextKey: RequestIDKey,
			// The default generator is fast but leaks the number of requests to the server.
			// The UUIDv4 generator avoids this.
			Generator: utils.UUIDv4,
			Header:    "Request-ID",
		},
	)
}

// ErrMissingRequestID is returned when a request ID is not present in the context.
var ErrMissingRequestID = errors.New("no request ID set on request context")

// RequestIDFrom returns the request ID from the Fiber context, or
// [ErrMissingRequestID] if no request ID is present.
func RequestIDFrom(c *fiber.Ctx) (string, error) {
	id, ok := c.Locals(RequestIDKey).(string)
	if !ok {
		return "", ErrMissingRequestID
	}

	return id, nil
}
