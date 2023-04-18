package presenter

import (
	"errors"
	"fmt"

	usershandler "github.com/angusgmorrison/realworld/internal/controller/rest/api/users"
	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Fiber implements handler Presenter interfaces for the Fiber HTTP framework.
type Fiber struct{}

// NewFiberPresenter returns a new Fiber presenter.
func NewFiberPresenter() *Fiber {
	return &Fiber{}
}

var _ usershandler.Presenter = (*Fiber)(nil)

func (r *Fiber) ShowBadRequest(c *fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": "request body is not a valid JSON string",
	})
}

// UserError maps user service errors to HTTP errors. Panics if it encounters an
// unhandled error, which MUST be handled by recovery middleware.
func (r *Fiber) ShowUserError(c *fiber.Ctx, err error) error {
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

	if errors.Is(err, user.ErrEmailRegistered) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(
			newJsonErrors(map[string][]string{
				"email": {"is already registered"},
			}),
		)
	}

	if errors.Is(err, user.ErrUsernameTaken) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(
			newJsonErrors(map[string][]string{
				"username": {"is taken"},
			}),
		)
	}

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		return r.ShowValidationErrors(c, validationErrs)
	}

	panic(fmt.Errorf("unhandled user service error: %w", err))
}

// ShowRegister renders an authorized user as JSON with status 201.
func (r *Fiber) ShowRegister(c *fiber.Ctx, user *user.User, token string) error {
	return showUserWithToken(c, fiber.StatusCreated, user, token)
}

// ShowLogin renders an authorized user as JSON with status 200.
func (r *Fiber) ShowLogin(c *fiber.Ctx, user *user.User, token string) error {
	return showUserWithToken(c, fiber.StatusOK, user, token)
}

// ShowGetCurrentUser renders an authorized user as JSON with status 200.
func (r *Fiber) ShowGetCurrentUser(c *fiber.Ctx, user *user.User, token string) error {
	return showUserWithToken(c, fiber.StatusOK, user, token)
}

// ShowUpdateCurrentCurrentUser renders an authorized user as JSON with status 200.
func (r *Fiber) ShowUpdateCurrentUser(c *fiber.Ctx, user *user.User, token string) error {
	return showUserWithToken(c, fiber.StatusOK, user, token)
}

func (r *Fiber) ShowValidationErrors(c *fiber.Ctx, errs validator.ValidationErrors) error {
	fieldErrs := make(map[string][]string)
	for _, err := range errs {
		fieldName := requestFieldName(err.Field())
		fieldErrs[fieldName] = append(fieldErrs[fieldName], userFriendlyErrMessage(err.Tag(), err.Param()))
	}

	return c.Status(fiber.StatusUnprocessableEntity).JSON(newJsonErrors(fieldErrs))
}

func showUserWithToken(c *fiber.Ctx, status int, user *user.User, token string) error {
	res := newUserResponseFromDomain(user, token)
	return c.Status(status).JSON(res)
}

func newJsonErrors(errs map[string][]string) fiber.Map {
	return fiber.Map{
		"errors": errs,
	}
}

func requestFieldName(modelFieldName string) string {
	if name, ok := modelFieldToRequestField[modelFieldName]; ok {
		return name
	}

	return modelFieldName
}

func userFriendlyErrMessage(tag string, param string) string {
	if format, ok := validationTagToErrMessage[tag]; ok {
		if param == "" {
			return format
		}
		return fmt.Sprintf(format, param)
	}

	return "is invalid"
}

var modelFieldToRequestField = map[string]string{
	"Email":    "email",
	"ImageURL": "image",
	"Password": "password",
	"User":     "user",
	"Username": "username",
}

var validationTagToErrMessage = map[string]string{
	"required": "is required",
	"email":    "is invalid",
	"min":      "must be at least %s characters",
	"max":      "must be at most %s bytes",
	"url":      "is invalid",
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

func newUserResponseFromDomain(u *user.User, token string) *userResponse {
	return &userResponse{
		User: userFields{
			Token:    token,
			Email:    u.Email,
			Username: u.Username,
			Bio:      u.Bio,
			Image:    u.ImageURL,
		},
	}
}