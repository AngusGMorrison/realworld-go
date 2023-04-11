package users

import (
	"errors"
	"fmt"

	"github.com/angusgmorrison/realworld/internal/ingress/rest/middleware"
	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/angusgmorrison/realworld/pkg/validate"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
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

// Register creates and returns a new user, along with an auth token.
func (users *Handler) Register(c *fiber.Ctx) error {
	var regReq user.RegistrationRequest
	if err := c.BodyParser(&regReq); err != nil {
		return badRequest(c)
	}

	authenticatedUser, err := users.service.Register(c.Context(), &regReq)
	if err != nil {
		return formatUserServiceError(c, err)
	}

	res := newUserResponseFromDomain(&authenticatedUser.User).withToken(authenticatedUser.Token)
	return c.Status(fiber.StatusCreated).JSON(res)
}

// Login authenticates a user and returns the user and token if successful.
func (users *Handler) Login(c *fiber.Ctx) error {
	var authReq user.AuthRequest
	if err := c.BodyParser(&authReq); err != nil {
		return badRequest(c)
	}

	authenticatedUser, err := users.service.Authenticate(c.Context(), &authReq)
	if err != nil {
		return formatUserServiceError(c, err)
	}

	res := newUserResponseFromDomain(&authenticatedUser.User).withToken(authenticatedUser.Token)
	return c.Status(fiber.StatusOK).JSON(res)
}

// GetCurrentUser returns the user corresponding to the ID contained in the
// request JWT.
func (users *Handler) GetCurrentUser(c *fiber.Ctx) error {
	user, err := users.service.Get(c.Context(), middleware.CurrentUser(c))
	if err != nil {
		return formatUserServiceError(c, err)
	}

	res := newUserResponseFromDomain(user).withToken(tokenFromRequest(c))
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

func badRequest(c *fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"message": "request body is not a valid JSON string",
	})
}

func tokenFromRequest(c *fiber.Ctx) string {
	token, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return ""
	}

	return token.Raw
}
