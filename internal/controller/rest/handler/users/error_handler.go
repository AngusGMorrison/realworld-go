package users

import (
	"errors"
	"github.com/angusgmorrison/realworld/internal/controller/rest/handler"
	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/gofiber/fiber/v2"
)

type errorHandler struct {
	handler.CommonErrorHandler
}

var _ handler.ErrorHandler = &errorHandler{}

func (eh errorHandler) Handle(c *fiber.Ctx, errs ...error) error {
	if len(errs) == 0 {
		panic(errors.New("ErrorHandler.Handle() invoked with no errors")) // 500
	}

	validationErrs := make(fiber.Map)
	for _, err := range errs {
		switch err := err.(type) {
		case *user.AuthError:
			return c.SendStatus(fiber.StatusUnauthorized)
		case *user.NotFoundError:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"user": err.Error()})
		case *user.ValidationError:
			switch err.Field {
			case user.EmailField:
				validationErrs["email"] = err.Message
			case user.PasswordField:
				validationErrs["password"] = err.Message
			case user.UsernameField:
				validationErrs["username"] = err.Message
			}
		default:
			panic(err) // 500
		}
	}

	return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"user": validationErrs})
}
