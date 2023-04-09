package users

import (
	"errors"
	"fmt"

	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/angusgmorrison/realworld/pkg/validate"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service user.Service
}

func NewHandler(service user.Service) *Handler {
	return &Handler{
		service: service,
	}
}

type userResponse struct {
	Token    string            `json:"token"`
	Email    user.EmailAddress `json:"email"`
	Username string            `json:"username"`
	Bio      string            `json:"bio"`
	Image    string            `json:"image"`
}

func newUserResponseFromDomain(authUser *user.AuthenticatedUser) *userResponse {
	return &userResponse{
		Token:    authUser.Token(),
		Email:    authUser.Email(),
		Username: authUser.Username(),
		Bio:      authUser.Bio(),
		Image:    authUser.ImageURL(),
	}
}

func (users *Handler) Register(c *fiber.Ctx) error {
	var regReq user.RegistrationRequest
	if err := c.BodyParser(&regReq); err != nil {
		return badRequest(c)
	}

	authenticatedUser, err := users.service.Register(c.Context(), &regReq)
	if err != nil {
		return formatUserServiceError(c, err)
	}

	res := newUserResponseFromDomain(authenticatedUser)
	return c.Status(fiber.StatusCreated).JSON(res)
}

func (users *Handler) Login(c *fiber.Ctx) error {
	var authReq user.AuthRequest
	if err := c.BodyParser(&authReq); err != nil {
		return badRequest(c)
	}

	authenticatedUser, err := users.service.Authenticate(c.Context(), &authReq)
	if err != nil {
		return formatUserServiceError(c, err)
	}

	res := newUserResponseFromDomain(authenticatedUser)
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

func badRequest(c *fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"message": "request body is not a valid JSON string",
	})
}
