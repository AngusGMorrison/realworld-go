package v0

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/angusgmorrison/logfusc"

	"github.com/gofiber/fiber/v2"

	"github.com/angusgmorrison/realworld-go/internal/domain/user"
	"github.com/angusgmorrison/realworld-go/pkg/option"
)

// UsersHandler holds dependencies for users endpoints.
type UsersHandler struct {
	service     user.Service
	jwtProvider JWTProvider
}

func NewUsersHandler(service user.Service, jwtProvider JWTProvider) *UsersHandler {
	return &UsersHandler{
		service:     service,
		jwtProvider: jwtProvider,
	}
}

// Register creates and returns a new user, along with an auth token.
func (h *UsersHandler) Register(c *fiber.Ctx) error {
	registrationReq, err := parseRegistrationRequest(c)
	if err != nil {
		return err
	}

	registeredUser, err := h.service.Register(c.Context(), registrationReq)
	if err != nil {
		return err
	}

	token, err := h.jwtProvider.TokenFor(registeredUser.ID())
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(
		newUserResponseBody(registeredUser, token),
	)
}

// Login authenticates a user and returns the user and token if successful.
func (h *UsersHandler) Login(c *fiber.Ctx) error {
	authReq, err := parseAuthRequest(c)
	if err != nil {
		return err
	}

	authenticatedUser, err := h.service.Authenticate(c.Context(), authReq)
	if err != nil {
		return err
	}

	token, err := h.jwtProvider.TokenFor(authenticatedUser.ID())
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(
		newUserResponseBody(authenticatedUser, token),
	)
}

// GetCurrent returns the user corresponding to the ID contained in the
// request JWT.
func (h *UsersHandler) GetCurrent(c *fiber.Ctx) error {
	currentUser, err := h.service.GetUser(c.Context(), mustGetCurrentUserIDFromContext(c))
	if err != nil {
		return err
	}

	token := mustGetCurrentJWTFromContext(c)

	return c.Status(fiber.StatusOK).JSON(
		newUserResponseBody(currentUser, token),
	)
}

// UpdateCurrent updates the user corresponding to the ID contained in the
// request JWT.
func (h *UsersHandler) UpdateCurrent(c *fiber.Ctx) error {
	updateReq, err := parseUpdateRequest(c)
	if err != nil {
		return err
	}

	updatedUser, err := h.service.UpdateUser(c.Context(), updateReq)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(
		newUserResponseBody(updatedUser, mustGetCurrentJWTFromContext(c)),
	)
}

type registrationRequestBody struct {
	User registrationRequestBodyUser `json:"user"`
}

type registrationRequestBodyUser struct {
	Username string                 `json:"username"`
	Email    string                 `json:"email"`
	Password logfusc.Secret[string] `json:"password"`
}

func parseRegistrationRequest(c *fiber.Ctx) (*user.RegistrationRequest, error) {
	var body registrationRequestBody
	if err := c.BodyParser(&body); err != nil {
		return nil, fmt.Errorf("parse request body: %w", err)
	}

	return user.ParseRegistrationRequest(body.User.Username, body.User.Email, body.User.Password)
}

type loginRequestBody struct {
	User loginRequestBodyUser `json:"user"`
}

type loginRequestBodyUser struct {
	Email    string                 `json:"email"`
	Password logfusc.Secret[string] `json:"password"`
}

func parseAuthRequest(c *fiber.Ctx) (*user.AuthRequest, error) {
	var body loginRequestBody
	if err := c.BodyParser(&body); err != nil {
		return nil, fmt.Errorf("parse request body: %w", err)
	}

	return user.ParseAuthRequest(body.User.Email, body.User.Password)
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

func parseUpdateRequest(c *fiber.Ctx) (*user.UpdateRequest, error) {
	var body updateRequestBody
	if err := c.BodyParser(&body); err != nil {
		return nil, fmt.Errorf("parse request body: %w", err)
	}

	return user.ParseUpdateRequest(
		mustGetCurrentUserIDFromContext(c),
		body.User.Email,
		body.User.Password,
		body.User.Bio,
		body.User.ImageURL,
	)
}

type successResponseBody struct {
	User successResponseBodyUser `json:"user"`
}

type successResponseBodyUser struct {
	Token    JWT    `json:"token"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Bio      string `json:"bio"`
	ImageURL string `json:"image"`
}

func newUserResponseBody(u *user.User, token JWT) *successResponseBody {
	return &successResponseBody{
		User: successResponseBodyUser{
			Token:    token,
			Email:    u.Email().String(),
			Username: u.Username().String(),
			Bio:      string(u.Bio().UnwrapOrZero()),
			ImageURL: u.ImageURL().UnwrapOrZero().String(),
		},
	}
}

// UsersErrorHandler is middleware that handles all users endpoint-related response.
func UsersErrorHandler(c *fiber.Ctx) error {
	err := c.Next()
	if err == nil {
		return nil
	}

	var (
		syntaxErr      *json.SyntaxError
		authErr        *user.AuthError
		notFoundErr    *user.NotFoundError
		validationErrs user.ValidationErrors
	)

	switch {
	case errors.As(err, &syntaxErr):
		return NewBadRequestError(syntaxErr)
	case errors.As(err, &authErr):
		return NewUnauthorizedError("invalid credentials", authErr.Cause)
	case errors.As(err, &notFoundErr):
		return NewNotFoundError(
			"user",
			fmt.Sprintf("user with %s %q not found", notFoundErr.IDType, notFoundErr.IDValue),
		)
	case errors.As(err, &validationErrs):
		return handleValidationErrors(validationErrs...)
	default:
		return err // pass to next error handler
	}
}

func handleValidationErrors(errs ...*user.ValidationError) error {
	userFacingMessages := make(userFacingValidationErrorMessages)
	for _, err := range errs {
		switch err.FieldType {
		case user.EmailFieldType:
			userFacingMessages["email"] = append(userFacingMessages["email"], err.Message)
		case user.PasswordFieldType:
			userFacingMessages["password"] = append(userFacingMessages["password"], err.Message)
		case user.UsernameFieldType:
			userFacingMessages["username"] = append(userFacingMessages["username"], err.Message)
		case user.URLFieldType:
			userFacingMessages["image"] = append(userFacingMessages["image"], err.Message)
		default:
			panic(fmt.Errorf("unhandled validation error field type %q: %w", err.FieldType, err))
		}
	}

	return NewUnprocessableEntityError(userFacingMessages)
}
