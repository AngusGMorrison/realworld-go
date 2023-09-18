package middleware

import (
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// UnsupportedContentTypeHandler takes an unsupported content type and returns a formatted error.
type UnsupportedContentTypeHandler func(unsupportedType string) error

// ContentTypeValidation is middleware that validates that the request Content-Type header
// is application/json.
func ContentTypeValidation(supportedTypes []string, handleUnsupported UnsupportedContentTypeHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		mediaType := c.Get(fiber.HeaderContentType)
		isSupported := slices.ContainsFunc(supportedTypes, func(supportedType string) bool {
			return strings.HasPrefix(mediaType, supportedType)
		})

		if !isSupported {
			return handleUnsupported(mediaType)
		}

		return c.Next()
	}
}
