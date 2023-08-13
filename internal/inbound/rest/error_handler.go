package rest

import (
	"errors"
	v0 "github.com/angusgmorrison/realworld-go/internal/inbound/rest/api/v0"
	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/go-multierror"
)

// newErrorHandler composes the global error handler for the server.
func newErrorHandler() fiber.ErrorHandler {
	return newTerminatingErrorHandler(
		newLoggingErrorHandler(
			responseWritingErrorHandler,
		),
	)
}

// newTerminatingErrorHandler returns an error handler that stops the propagation
// of handled errors. It must be the outermost error handler.
func newTerminatingErrorHandler(next fiber.ErrorHandler) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		_ = next(c, err)
		return nil
	}
}

// newLoggingErrorHandler logs errors propagated from inner error handlers
// before returning them to outer error handlers.
func newLoggingErrorHandler(next fiber.ErrorHandler) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		handledErr := next(c, err)
		if handledErr != nil {
			LoggerFrom(c).Printf("%v\n", handledErr)
		}
		return handledErr
	}
}

// responseWritingErrorHandler writes the response to a client in the event of
// an error. UserFacingError responses should not be written from any other location.
//
// The underlying causes of errors are propagated to instrumenting error handlers.
func responseWritingErrorHandler(c *fiber.Ctx, err error) error {
	var userFacingErr *v0.UserFacingError
	if ok := errors.As(err, &userFacingErr); ok {
		if writeRespErr := c.Status(userFacingErr.StatusCode).JSON(userFacingErr.Body()); writeRespErr != nil {
			return multierror.Append(err, writeRespErr)
		}
		if cause := userFacingErr.Cause(); cause != nil {
			return cause
		}
		return nil
	}

	c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
	if writeRespErr := c.SendStatus(fiber.StatusInternalServerError); writeRespErr != nil {
		return multierror.Append(err, writeRespErr)
	}

	return err
}
