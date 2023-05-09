package users

import (
	"github.com/angusgmorrison/realworld/internal/controller/rest/middleware"
	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/angusgmorrison/realworld/pkg/primitive"
	"github.com/angusgmorrison/realworld/pkg/validate"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// Handler holds dependencies for users endpoints.
type Handler struct {
	service   user.Service
	presenter Presenter
}

// NewHandler returns a new Handler from the injected dependencies.
func NewHandler(service user.Service, presenter Presenter) *Handler {
	return &Handler{
		service:   service,
		presenter: presenter,
	}
}

type registrationRequestBody struct {
	User *user.RegistrationRequest `json:"user" validate:"required"`
}

// Register creates and returns a new user, along with an auth token.
func (h *Handler) Register(c *fiber.Ctx) error {
	var body registrationRequestBody
	if err := c.BodyParser(&body); err != nil {
		return h.presenter.ShowBadRequest(c)
	}

	if err := validate.Struct(&body); err != nil {
		return h.presenter.ShowValidationErrors(c, err.(validator.ValidationErrors))
	}

	authenticatedUser, err := h.service.Register(c.Context(), body.User)
	if err != nil {
		return h.presenter.ShowUserError(c, err)
	}

	return h.presenter.ShowRegister(c, authenticatedUser.User, authenticatedUser.Token)
}

type loginRequestBody struct {
	User *user.AuthRequest `json:"user" validate:"required"`
}

// Login authenticates a user and returns the user and token if successful.
func (h *Handler) Login(c *fiber.Ctx) error {
	var body loginRequestBody
	if err := c.BodyParser(&body); err != nil {
		return h.presenter.ShowBadRequest(c)
	}

	if err := validate.Struct(&body); err != nil {
		return h.presenter.ShowValidationErrors(c, err.(validator.ValidationErrors))
	}

	authenticatedUser, err := h.service.Authenticate(c.Context(), body.User)
	if err != nil {
		return h.presenter.ShowUserError(c, err)
	}

	return h.presenter.ShowLogin(c, authenticatedUser.User, authenticatedUser.Token)
}

// GetCurrentUser returns the user corresponding to the ID contained in the
// request JWT.
func (h *Handler) GetCurrentUser(c *fiber.Ctx) error {
	user, err := h.service.GetUser(c.Context(), middleware.CurrentUser(c))
	if err != nil {
		return h.presenter.ShowUserError(c, err)
	}

	return h.presenter.ShowGetCurrentUser(c, user, tokenFromRequest(c))
}

// For updates, we decouple the REST model from the domain model, since they are
// structurally disinct and require different validations (the REST request
// contains no user ID, which is derived from the request context and is
// required by the domain).
type updateCurrentUserRequestBody struct {
	User userUpdates `json:"user" validate:"required"`
}
type userUpdates struct {
	Email    *primitive.EmailAddress    `json:"email" validate:"omitempty,email"`
	Bio      *string                    `json:"bio"`
	ImageURL *string                    `json:"image" validate:"omitempty,url"`
	Password *primitive.SensitiveString `json:"password"`
}

func (body *updateCurrentUserRequestBody) toDomain(userID uuid.UUID) *user.UpdateRequest {
	return &user.UpdateRequest{
		UserID:   userID,
		Email:    body.User.Email,
		Bio:      body.User.Bio,
		ImageURL: body.User.ImageURL,
		OptionalValidatingPassword: user.OptionalValidatingPassword{
			Password: body.User.Password,
		},
	}
}

// UpdateCurrentUser updates the user corresponding to the ID contained in the
// request JWT.
func (h *Handler) UpdateCurrentUser(c *fiber.Ctx) error {
	var body updateCurrentUserRequestBody
	if err := c.BodyParser(&body); err != nil {
		return h.presenter.ShowBadRequest(c)
	}

	if err := validate.Struct(&body); err != nil {
		return h.presenter.ShowValidationErrors(c, err.(validator.ValidationErrors))
	}

	updateReq := body.toDomain(middleware.CurrentUser(c))
	user, err := h.service.UpdateUser(c.Context(), updateReq)
	if err != nil {
		return h.presenter.ShowUserError(c, err)
	}

	return h.presenter.ShowUpdateCurrentUser(c, user, tokenFromRequest(c))
}

func tokenFromRequest(c *fiber.Ctx) string {
	token, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return ""
	}

	return token.Raw
}
