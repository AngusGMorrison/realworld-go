package users

import (
	"github.com/angusgmorrison/logfusc"
	"github.com/angusgmorrison/realworld/internal/domain/user"
	"github.com/angusgmorrison/realworld/internal/inbound/rest/api/v0"
	"github.com/angusgmorrison/realworld/internal/inbound/rest/middleware"
	"github.com/angusgmorrison/realworld/pkg/option"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// Handler holds dependencies for users endpoints.
type Handler struct {
	service     user.Service
	jwtProvider JWTProvider
}

func NewHandler(service user.Service, jwtProvider JWTProvider) *Handler {
	return &Handler{
		service:     service,
		jwtProvider: jwtProvider,
	}
}

type registrationRequestBody struct {
	User registrationRequestBodyUser `json:"user"`
}

type registrationRequestBodyUser struct {
	Username string                 `json:"username"`
	Email    string                 `json:"email"`
	Password logfusc.Secret[string] `json:"password"`
}

func (body *registrationRequestBody) toDomain() (*user.RegistrationRequest, error) {
	return user.ParseRegistrationRequest(body.User.Username, body.User.Email, body.User.Password)
}

// Register creates and returns a new user, along with an auth token.
func (h *Handler) Register(c *fiber.Ctx) error {
	var body registrationRequestBody
	if err := c.BodyParser(&body); err != nil {
		return v0.BadRequest(c)
	}

	registrationReq, err := body.toDomain()
	if err != nil {
		return err
	}

	registeredUser, err := h.service.Register(c.Context(), registrationReq)
	if err != nil {
		return err
	}

	token, err := h.jwtProvider.FromSubject(registeredUser.ID())
	if err != nil {
		return err
	}

	userResponseBody := newUserResponseBodyFromDomain(registeredUser, token)
	return c.Status(fiber.StatusCreated).JSON(userResponseBody)
}

type loginRequestBody struct {
	User loginRequestBodyUser `json:"user"`
}

type loginRequestBodyUser struct {
	Email    string                 `json:"email"`
	Password logfusc.Secret[string] `json:"password"`
}

func (body *loginRequestBody) toDomain() (*user.AuthRequest, error) {
	return user.ParseAuthRequest(body.User.Email, body.User.Password)
}

// Login authenticates a user and returns the user and token if successful.
func (h *Handler) Login(c *fiber.Ctx) error {
	var body loginRequestBody
	if err := c.BodyParser(&body); err != nil {
		return v0.BadRequest(c)
	}

	authReq, err := body.toDomain()
	if err != nil {
		return err
	}

	authenticatedUser, err := h.service.Authenticate(c.Context(), authReq)
	if err != nil {
		return err
	}

	token, err := h.jwtProvider.FromSubject(authenticatedUser.ID())
	if err != nil {
		return err
	}

	userResponseBody := newUserResponseBodyFromDomain(authenticatedUser, token)
	return c.Status(fiber.StatusOK).JSON(userResponseBody)
}

// GetCurrent returns the user corresponding to the ID contained in the
// request JWT.
func (h *Handler) GetCurrent(c *fiber.Ctx) error {
	currentUser, err := h.service.GetUser(c.Context(), middleware.CurrentUserID(c))
	if err != nil {
		return err
	}

	userResponseBody := newUserResponseBodyFromDomain(currentUser, tokenFromRequest(c))
	return c.Status(fiber.StatusOK).JSON(userResponseBody)
}

type updateRequestBody struct {
	User updateRequestBodyUser `json:"user"`
}

type updateRequestBodyUser struct {
	Email    option.Option[string]                 `json:"email"`
	Password option.Option[logfusc.Secret[string]] `json:"password"`
	Bio      option.Option[string]                 `json:"bio"`
	ImageURL option.Option[string]                 `json:"image"`
}

func (body *updateRequestBody) toDomain(userID uuid.UUID) (*user.UpdateRequest, error) {
	return user.ParseUpdateRequest(
		userID,
		body.User.Email,
		body.User.Password,
		body.User.Bio,
		body.User.ImageURL,
	)
}

// UpdateCurrent updates the user corresponding to the ID contained in the
// request JWT.
func (h *Handler) UpdateCurrent(c *fiber.Ctx) error {
	var body updateRequestBody
	if err := c.BodyParser(&body); err != nil {
		return v0.BadRequest(c)
	}

	updateReq, errs := body.toDomain(middleware.CurrentUserID(c))
	if errs != nil {
		return errs
	}

	updatedUser, err := h.service.UpdateUser(c.Context(), updateReq)
	if err != nil {
		return err
	}

	userResponseBody := newUserResponseBodyFromDomain(updatedUser, tokenFromRequest(c))
	return c.Status(fiber.StatusOK).JSON(userResponseBody)
}

func tokenFromRequest(c *fiber.Ctx) logfusc.Secret[string] {
	token, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return logfusc.NewSecret("")
	}

	return logfusc.NewSecret(token.Raw)
}

type successResponseBody struct {
	User successResponseBodyUser `json:"user"`
}

type successResponseBodyUser struct {
	Token    string `json:"token"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Bio      string `json:"bio"`
	ImageURL string `json:"image"`
}

func newUserResponseBodyFromDomain(u *user.User, token JWT) *successResponseBody {
	return &successResponseBody{
		User: successResponseBodyUser{
			Token:    token.Expose(),
			Email:    u.Email().String(),
			Username: u.Username().String(),
			Bio:      string(u.Bio().ValueOrZero()),
			ImageURL: u.ImageURL().ValueOrZero().String(),
		},
	}
}
