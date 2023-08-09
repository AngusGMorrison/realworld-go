package users

import (
	"errors"
	"fmt"
	"github.com/angusgmorrison/realworld/internal/domain/user"
	"github.com/angusgmorrison/realworld/internal/inbound/rest/api/v0"
	"github.com/gofiber/fiber/v2"
)

// HandleErrors is middleware that handles all users endpoint-related errors.
func HandleErrors(c *fiber.Ctx) error {
	err := c.Next()
	if err == nil {
		return nil
	}

	var (
		authErr        *user.AuthError
		notFoundErr    *user.NotFoundError
		validationErr  *user.ValidationError
		validationErrs user.ValidationErrors
	)

	switch {
	case errors.As(err, &authErr):
		return v0.Unauthorized(c)
	case errors.As(err, &notFoundErr):
		return v0.NotFound(c, "user", notFoundErr.Error())
	case errors.As(err, &validationErr):
		jsonErrs := make(fiber.Map)
		setValidationError(jsonErrs, validationErr)
		return v0.UnprocessableEntity(c, jsonErrs)
	case errors.As(err, &validationErrs):
		jsonErrs := make(fiber.Map)
		for _, subErr := range validationErrs {
			setValidationError(jsonErrs, subErr)
		}
		return v0.UnprocessableEntity(c, jsonErrs)
	default:
		return err // pass to next error handler
	}
}

func setValidationError(jsonErrs fiber.Map, err *user.ValidationError) {
	switch err.FieldType {
	case user.EmailFieldType:
		jsonErrs["email"] = err.Message
	case user.PasswordFieldType:
		jsonErrs["password"] = err.Message
	case user.UsernameFieldType:
		jsonErrs["username"] = err.Message
	case user.URLFieldType:
		jsonErrs["image"] = err.Message
	default:
		panic(fmt.Errorf("unhandled validation error field type %q: %w", err.FieldType, err))
	}
}
