package v0

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/angusgmorrison/realworld-go/pkg/etag"

	"github.com/angusgmorrison/realworld-go/internal/inbound/rest/middleware"

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
	currentUserID, err := currentUserIDFrom(c)
	if err != nil {
		return err
	}

	currentUser, err := h.service.GetUser(c.Context(), currentUserID)
	if err != nil {
		return err
	}

	token, err := currentJWTFrom(c)
	if err != nil {
		return err
	}

	c.Set(fiber.HeaderETag, currentUser.ETag().String())

	// If any If-None-Match header matches the ETag of the retrieved resource,
	// the client's cached resource is up-to-date.
	if c.Get(fiber.HeaderIfNoneMatch) == currentUser.ETag().String() {
		return c.Status(fiber.StatusNotModified).Send(nil)
	}

	return c.Status(fiber.StatusOK).JSON(newUserResponseBody(currentUser, token))
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

	currentJWT, err := currentJWTFrom(c)
	if err != nil {
		return err
	}

	c.Set(fiber.HeaderETag, updatedUser.ETag().String())

	return c.Status(fiber.StatusOK).JSON(newUserResponseBody(updatedUser, currentJWT))
}

type registrationRequestBody struct {
	User registrationRequestBodyUser `json:"user"`
}

type registrationRequestBodyUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
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
	Email    string `json:"email"`
	Password string `json:"password"`
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
	Email    option.Option[string] `json:"email"`
	Password option.Option[string] `json:"password"`
	Bio      option.Option[string] `json:"bio"`
	ImageURL option.Option[string] `json:"image"`
}

func parseUpdateRequest(c *fiber.Ctx) (*user.UpdateRequest, error) {
	ifMatch := c.Get(fiber.HeaderIfMatch)
	eTag, err := etag.Parse(ifMatch)
	if err != nil {
		return nil, fmt.Errorf("parse If-Match header: %w", err)
	}

	var body updateRequestBody
	if err := c.BodyParser(&body); err != nil {
		return nil, fmt.Errorf("parse request body: %w", err)
	}

	currentUserID, err := currentUserIDFrom(c)
	if err != nil {
		return nil, err
	}

	return user.ParseUpdateRequest(
		currentUserID,
		eTag,
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

// UsersErrorHandling is middleware that handles all users endpoint-related response.
func UsersErrorHandling(c *fiber.Ctx) error {
	err := c.Next()
	if err == nil {
		return nil
	}

	requestID, ok := c.Locals(middleware.RequestIDKey).(string)
	if !ok {
		return fmt.Errorf("unhandled users error: request ID not set on context: %w", err)
	}

	var (
		syntaxErr                 *json.SyntaxError
		authErr                   *user.AuthError
		notFoundErr               *user.NotFoundError
		concurrentModificationErr *user.ConcurrentModificationError
		validationErrs            user.ValidationErrors
	)

	switch {
	case errors.As(err, &syntaxErr):
		return NewBadRequestError(requestID, syntaxErr)
	case errors.As(err, &authErr):
		return NewUnauthorizedError(requestID, authErr)
	case errors.As(err, &notFoundErr):
		return NewNotFoundError(
			requestID,
			missingResource{
				name:   "user",
				idType: notFoundErr.IDType.String(),
				id:     notFoundErr.IDValue,
			},
		)
	case errors.As(err, &concurrentModificationErr):
		return NewPreconditionFailedError(
			requestID,
			"user",
			concurrentModificationErr.ETag,
			concurrentModificationErr,
		)
	case errors.As(err, &validationErrs):
		return NewUnprocessableEntityError(requestID, validationErrs)
	default:
		return err // pass to next error handler
	}
}
