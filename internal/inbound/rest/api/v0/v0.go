// Package v0 contains handlers, middleware and response formats for the v0 API.
package v0

import (
	"github.com/gofiber/fiber/v2"
)

// SupportedContentTypes lists the content types supported by the v0 API.
var SupportedContentTypes = []string{fiber.MIMEApplicationJSON}

// UnsupportedContentTypeHandler takes an unsupported content type and returns a formatted error.
func UnsupportedContentTypeHandler(unsupportedType string) error {
	return NewUnsupportedMediaTypeError(unsupportedType, SupportedContentTypes)
}
