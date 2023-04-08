package rest

import (
	"errors"
	"fmt"

	"github.com/angusgmorrison/realworld/service/user"
	"github.com/gofiber/fiber/v2"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (lr *loginRequest) toDomain() (*user.AuthRequest, error) {
	return user.NewAuthRequest(lr.Email, lr.Password)
}

func (s *Server) LoginHandler(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	authReq, err := req.toDomain()
	if err != nil {
		return newUsersError(c, err)
	}

	user, token, err := s.userService.Authenticate(c.Context(), authReq)
	if err != nil {
		return newUsersError(c, err)
	}

}

// newUsersError maps user service errors to HTTP errors. Panics if it encounters
// an unhandled error, which MUST be handled by recovery middleware.
func newUsersError(c *fiber.Ctx, err error) error {
	var authErr *user.AuthError
	if errors.As(err, &authErr) {
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	fieldErrs := make(map[string][]string)
	var validationErr *user.ValidationError
	if errors.As(err, &validationErr) {
		for _, fieldErr := range validationErr.FieldErrors() {
			if errors.Is(fieldErr, user.ErrImageURLUnparseable) {
				fieldErrs["image"] = []string{"is invalid"}
			} else if errors.Is(fieldErr, user.ErrEmailAddressUnparseable) {
				fieldErrs["email"] = []string{"is invalid"}
			} else if errors.Is(fieldErr, user.ErrPasswordTooLong) {
				fieldErrs["password"] = []string{"length exceeds 72 bytes"}
			} else {
				panic(fmt.Errorf("unhandled validation error: %w", err))
			}
		}

		return c.Status(fiber.StatusUnprocessableEntity).JSON(
			fiber.Map{
				"errors": fieldErrs,
			},
		)
	}

	if errors.Is(err, user.ErrUserNotFound) {
		return fiber.NewError(fiber.StatusNotFound)
	}

	panic(fmt.Errorf("unhandled user service error: %w", err))
}
