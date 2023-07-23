package users

import (
	"github.com/angusgmorrison/logfusc"
	"github.com/angusgmorrison/realworld/internal/controller/rest/handler"
	"github.com/angusgmorrison/realworld/internal/controller/rest/middleware"
	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/angusgmorrison/realworld/pkg/option"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// Handler holds dependencies for users endpoints.
type Handler struct {
	service      user.Service
	errorHandler handler.ErrorHandler
}

func NewHandler(service user.Service) *Handler {
	return &Handler{
		service:      service,
		errorHandler: errorHandler{},
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

func (body *registrationRequestBody) toDomain() (*user.RegistrationRequest, []error) {
	var parseErrors []error
	username, usernameErr := user.ParseUsername(body.User.Username)
	if usernameErr != nil {
		parseErrors = append(parseErrors, usernameErr)
	}
	email, emailErr := user.ParseEmailAddress(body.User.Email)
	if emailErr != nil {
		parseErrors = append(parseErrors, emailErr)
	}
	passwordHash, passwordHashErr := user.ParsePassword(body.User.Password)
	if passwordHashErr != nil {
		parseErrors = append(parseErrors, passwordHashErr)
	}
	if len(parseErrors) > 0 {
		return nil, parseErrors
	}

	return user.NewRegistrationRequest(username, email, passwordHash), nil
}

// Register creates and returns a new user, along with an auth token.
func (h *Handler) Register(c *fiber.Ctx) error {
	var body registrationRequestBody
	if err := c.BodyParser(&body); err != nil {
		return h.errorHandler.BadRequest(c)
	}

	registrationReq, errs := body.toDomain()
	if errs != nil {
		return h.errorHandler.Handle(c, errs...)
	}

	authenticatedUser, err := h.service.Register(c.Context(), registrationReq)
	if err != nil {
		return h.errorHandler.Handle(c, err)
	}

	userResponseBody := newUserResponseBodyFromDomain(authenticatedUser.User(), authenticatedUser.Token())
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
	email, emailErr := user.ParseEmailAddress(body.User.Email)
	if emailErr != nil {
		return nil, emailErr
	}

	return user.NewAuthRequest(email, body.User.Password), nil
}

// Login authenticates a user and returns the user and token if successful.
func (h *Handler) Login(c *fiber.Ctx) error {
	var body loginRequestBody
	if err := c.BodyParser(&body); err != nil {
		return h.errorHandler.BadRequest(c)
	}

	authReq, err := body.toDomain()
	if err != nil {
		return h.errorHandler.Handle(c, err)
	}

	authenticatedUser, err := h.service.Authenticate(c.Context(), authReq)
	if err != nil {
		return h.errorHandler.Handle(c, err)
	}

	userResponseBody := newUserResponseBodyFromDomain(authenticatedUser.User(), authenticatedUser.Token())
	return c.Status(fiber.StatusOK).JSON(userResponseBody)
}

// GetCurrent returns the user corresponding to the ID contained in the
// request JWT.
func (h *Handler) GetCurrent(c *fiber.Ctx) error {
	currentUser, err := h.service.GetUser(c.Context(), middleware.CurrentUserID(c))
	if err != nil {
		return h.errorHandler.Handle(c, err)
	}

	userResponseBody := newUserResponseBodyFromDomain(currentUser, tokenFromRequest(c))
	return c.Status(fiber.StatusOK).JSON(userResponseBody)
}

type updateRequestBody struct {
	User updateRequestBodyUser `json:"user"`
}

type updateRequestBodyUser struct {
	Email    *string                 `json:"email"`
	Password *logfusc.Secret[string] `json:"password"`
	Bio      *string                 `json:"bio"`
	ImageURL *string                 `json:"image"`
}

func (body *updateRequestBody) toDomain(userID uuid.UUID) (*user.UpdateRequest, []error) {
	var errors []error
	emailOpt := option.None[user.EmailAddress]()
	if body.User.Email != nil {
		email, err := user.ParseEmailAddress(*body.User.Email)
		if err != nil {
			errors = append(errors, err)
		} else {
			emailOpt = option.Some(email)
		}

	}

	passwordOpt := option.None[user.PasswordHash]()
	if body.User.Password != nil {
		hash, err := user.ParsePassword(*body.User.Password)
		if err != nil {
			errors = append(errors, err)
		} else {
			passwordOpt = option.Some(hash)
		}
	}

	bioOpt := option.None[user.Bio]()
	if body.User.Bio != nil {
		bioOpt = option.Some(user.Bio(*body.User.Bio))
	}

	imageURLOpt := option.None[user.URL]()
	if body.User.ImageURL != nil {
		url, err := user.ParseURL(*body.User.ImageURL)
		if err != nil {
			errors = append(errors, err)
		} else {
			imageURLOpt = option.Some(url)
		}
	}
	if len(errors) > 0 {
		return nil, errors
	}

	return user.NewUpdateRequest(userID, emailOpt, passwordOpt, bioOpt, imageURLOpt), nil
}

// UpdateCurrent updates the user corresponding to the ID contained in the
// request JWT.
func (h *Handler) UpdateCurrent(c *fiber.Ctx) error {
	var body updateRequestBody
	if err := c.BodyParser(&body); err != nil {
		return h.errorHandler.BadRequest(c)
	}

	updateReq, errs := body.toDomain(middleware.CurrentUserID(c))
	if errs != nil {
		return h.errorHandler.Handle(c, errs...)
	}

	updatedUser, err := h.service.UpdateUser(c.Context(), updateReq)
	if err != nil {
		return h.errorHandler.Handle(c, err)
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

func newUserResponseBodyFromDomain(u *user.User, token user.JWT) *successResponseBody {
	return &successResponseBody{
		User: successResponseBodyUser{
			Token:    token.String(),
			Email:    u.Email().String(),
			Username: u.Username().String(),
			Bio:      string(u.Bio().ValueOrZero()),
			ImageURL: u.ImageURL().ValueOrZero().String(),
		},
	}
}

func badRequest(c *fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": "request body is not a valid JSON string",
	})
}
