package v0

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/angusgmorrison/realworld-go/internal/inbound/rest/middleware"
	"github.com/gofiber/fiber/v2"
)

// SupportedContentTypes lists the content types supported by the v0 API.
var supportedContentTypes = []string{fiber.MIMEApplicationJSON}

// ContentTypeValidation is middleware that validates the request Content-Type
// header.
func ContentTypeValidation(c *fiber.Ctx) error {
	contentType := c.Get(fiber.HeaderContentType)
	isSupported := slices.ContainsFunc(supportedContentTypes, func(supportedType string) bool {
		return strings.HasPrefix(contentType, supportedType)
	})

	if !isSupported {
		requestID, err := middleware.RequestIDFrom(c)
		if err != nil {
			return fmt.Errorf("unable to respond to request with unsupported content type: %w", err)
		}

		return NewUnsupportedContentTypeError(requestID, contentType, supportedContentTypes)
	}

	return c.Next()
}

// ErrorHandling is middleware that handles errors returned by the v0 API
// handlers.
func ErrorHandling(c *fiber.Ctx) error {
	err := c.Next()
	if err == nil {
		return nil
	}

	var jsonErr *JSONError
	if errors.As(err, &jsonErr) {
		middleware.LoggerFrom(c).Printf("%v\n", err)
		return c.Status(jsonErr.Status).JSON(jsonErr)
	}

	requestID, requestIDErr := middleware.RequestIDFrom(c)
	if requestIDErr != nil {
		return fmt.Errorf("unhandlable v0 API error: %w: %w", requestIDErr, err)
	}

	middleware.LoggerFrom(c).Printf("%v\n", err)
	return c.Status(fiber.StatusInternalServerError).JSON(NewInternalServerError(requestID, err))
}
