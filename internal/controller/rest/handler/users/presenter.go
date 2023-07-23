package users

import (
	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Presenter is an abstraction over the logic for rendering responses.
type Presenter interface {
	ShowBadRequest(c *fiber.Ctx) error
	ShowValidationErrors(c *fiber.Ctx, errs validator.ValidationErrors) error
	ShowUserError(c *fiber.Ctx, err error) error
	ShowRegister(c *fiber.Ctx, user *user.User, token string) error
	ShowLogin(c *fiber.Ctx, user *user.User, token string) error
	ShowGetCurrentUser(c *fiber.Ctx, user *user.User, token string) error
	ShowUpdateCurrentUser(c *fiber.Ctx, user *user.User, token string) error
}
