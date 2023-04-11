package users

import (
	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/gofiber/fiber/v2"
)

// Renderer is an abstraction over the logic for rendering responses.
type Renderer interface {
	BadRequest(c *fiber.Ctx) error
	UserError(c *fiber.Ctx, err error) error
	RegisterSuccess(c *fiber.Ctx, user *user.User, token string) error
	LoginSuccess(c *fiber.Ctx, user *user.User, token string) error
	GetCurrentUserSuccess(c *fiber.Ctx, user *user.User, token string) error
	UpdateCurrentCurrentUserSuccess(c *fiber.Ctx, user *user.User, token string) error
}
