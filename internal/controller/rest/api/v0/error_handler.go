package v0

import "github.com/gofiber/fiber/v2"

// ErrorHandler is an abstraction over the logic for handling errors that
// allows error handling code to be shared and mocked.
type ErrorHandler interface {
	Handle(c *fiber.Ctx, errs ...error) error
	BadRequest(c *fiber.Ctx) error
}

// CommonErrorHandler is an [ErrorHandler] providing common defaults.
// It is intended to be embedded in other ErrorHandlers.
type CommonErrorHandler struct{}

var _ ErrorHandler = &CommonErrorHandler{}

func (eh CommonErrorHandler) Handle(c *fiber.Ctx, errs ...error) error {
	panic("CommonErrorHandler.Handle() is abstract and must be implemented by the api that embeds it")
}

func (eh CommonErrorHandler) BadRequest(c *fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": "request body is not a valid JSON string",
	})
}
