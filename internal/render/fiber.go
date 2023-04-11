package render

import (
	"errors"
	"fmt"

	usershandler "github.com/angusgmorrison/realworld/internal/controller/rest/api/users"
	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/angusgmorrison/realworld/pkg/validate"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// FiberRenderer implements handler Renderer interfaces for Fiber.
type FiberRenderer struct{}

var _ usershandler.Renderer = (*FiberRenderer)(nil)

func (r *FiberRenderer) BadRequest(c *fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"message": "request body is not a valid JSON string",
	})
}

// UserError maps user service errors to HTTP errors. Panics if it encounters an
// unhandled error, which MUST be handled by recovery middleware.
func (r *FiberRenderer) UserError(c *fiber.Ctx, err error) error {
	var authErr *user.AuthError
	if errors.As(err, &authErr) {
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	if errors.Is(err, user.ErrUserNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(
			newJsonErrors(map[string][]string{
				"email": {"user not found"},
			}),
		)
	}

	if errors.Is(err, user.ErrUserExists) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(
			newJsonErrors(map[string][]string{
				"email": {"user already registered"},
			}),
		)
	}

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		return formatValidationErrors(c, validationErrs)
	}

	panic(fmt.Errorf("unhandled user service error: %w", err))
}

// RegisterSuccess renders an authorized user as JSON with status 201.
func (r *FiberRenderer) RegisterSuccess(c *fiber.Ctx, user *user.User, token string) error {
	return renderUserWithToken(c, fiber.StatusCreated, user, token)
}

// LoginSuccess renders an authorized user as JSON with status 200.
func (r *FiberRenderer) LoginSuccess(c *fiber.Ctx, user *user.User, token string) error {
	return renderUserWithToken(c, fiber.StatusOK, user, token)
}

// GetCurrentUserSuccess renders an authorized user as JSON with status 200.
func (r *FiberRenderer) GetCurrentUserSuccess(c *fiber.Ctx, user *user.User, token string) error {
	return renderUserWithToken(c, fiber.StatusOK, user, token)
}

// UpdateCurrentCurrentUserSuccess renders an authorized user as JSON with status 200.
func (r *FiberRenderer) UpdateCurrentCurrentUserSuccess(c *fiber.Ctx, user *user.User, token string) error {
	return renderUserWithToken(c, fiber.StatusOK, user, token)
}

func renderUserWithToken(c *fiber.Ctx, status int, user *user.User, token string) error {
	res := newUserResponseFromDomain(user).withToken(token)
	return c.Status(status).JSON(res)
}

func newJsonErrors(errs map[string][]string) fiber.Map {
	return fiber.Map{
		"errors": errs,
	}
}

func formatValidationErrors(c *fiber.Ctx, errs validator.ValidationErrors) error {
	fieldErrs := make(map[string][]string)
	for _, err := range errs {
		fieldErrs[err.Field()] = append(fieldErrs[err.Field()], validate.Translate(err))
	}

	return c.Status(fiber.StatusUnprocessableEntity).JSON(newJsonErrors(fieldErrs))
}

type userResponse struct {
	User userFields `json:"user"`
}

type userFields struct {
	Token    string            `json:"token"`
	Email    user.EmailAddress `json:"email"`
	Username string            `json:"username"`
	Bio      string            `json:"bio"`
	Image    string            `json:"image"`
}

func newUserResponseFromDomain(u *user.User) *userResponse {
	return &userResponse{
		User: userFields{
			Email:    u.Email,
			Username: u.Username,
			Bio:      u.Bio,
			Image:    u.ImageURL,
		},
	}
}

func (u *userResponse) withToken(token string) *userResponse {
	u.User.Token = token
	return u
}
