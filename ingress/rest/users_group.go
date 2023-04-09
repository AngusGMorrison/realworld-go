package rest

import (
	"errors"
	"fmt"

	"github.com/angusgmorrison/realworld/entity/validate"
	"github.com/angusgmorrison/realworld/service/user"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type loginResponse struct {
	Token    string            `json:"token"`
	Email    user.EmailAddress `json:"email"`
	Username string            `json:"username"`
	Bio      string            `json:"bio"`
	Image    string            `json:"image"`
}

func newLoginResponseFromDomain(authUser *user.AuthenticatedUser) *loginResponse {
	return &loginResponse{
		Token:    authUser.Token(),
		Email:    authUser.Email(),
		Username: authUser.Username(),
		Bio:      authUser.Bio(),
		Image:    authUser.ImageURL(),
	}
}

type usersGroup struct {
	service user.Service
}

func (users *usersGroup) loginHandler(c *fiber.Ctx) error {
	var authReq user.AuthRequest
	if err := c.BodyParser(&authReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "request body is not a valid JSON string",
		})
	}

	authenticatedUser, err := users.service.Authenticate(c.Context(), &authReq)
	if err != nil {
		return formatUserServiceError(c, err)
	}

	res := newLoginResponseFromDomain(authenticatedUser)
	return c.Status(fiber.StatusOK).JSON(res)
}

// formatUserServiceError maps user service errors to HTTP errors. Panics if it encounters
// an unhandled error, which MUST be handled by recovery middleware.
func formatUserServiceError(c *fiber.Ctx, err error) error {
	var authErr *user.AuthError
	if errors.As(err, &authErr) {
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	if errors.Is(err, user.ErrUserNotFound) {
		return fiber.NewError(fiber.StatusNotFound)
	}

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		return formatValidationErrors(c, validationErrs)
	}

	panic(fmt.Errorf("unhandled user service error: %w", err))
}

func formatValidationErrors(c *fiber.Ctx, errs validator.ValidationErrors) error {
	fieldErrs := make(map[string][]string)
	for _, err := range errs {
		fieldErrs[err.Field()] = append(fieldErrs[err.Field()], validate.Translate(err))
	}

	return c.Status(fiber.StatusUnprocessableEntity).JSON(
		fiber.Map{
			"errors": fieldErrs,
		},
	)
}
