package users

import (
	"fmt"

	"github.com/angusgmorrison/realworld/internal/controller/rest/middleware"
	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

type Handler struct {
	service   user.Service
	presenter Presenter
}

func NewHandler(service user.Service, presenter Presenter) *Handler {
	return &Handler{
		service:   service,
		presenter: presenter,
	}
}

// Register creates and returns a new user, along with an auth token.
func (h *Handler) Register(c *fiber.Ctx) error {
	var regReq user.RegistrationRequest
	if err := c.BodyParser(&regReq); err != nil {
		return h.presenter.ShowBadRequest(c)
	}

	authenticatedUser, err := h.service.Register(c.Context(), &regReq)
	if err != nil {
		return h.presenter.ShowUserError(c, err)
	}

	return h.presenter.ShowRegister(c, authenticatedUser.User, authenticatedUser.Token)
}

// Login authenticates a user and returns the user and token if successful.
func (h *Handler) Login(c *fiber.Ctx) error {
	var authReq user.AuthRequest
	if err := c.BodyParser(&authReq); err != nil {
		return h.presenter.ShowBadRequest(c)
	}

	fmt.Println("parsed")
	authenticatedUser, err := h.service.Authenticate(c.Context(), &authReq)
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

// UpdateCurrentUser updates the user corresponding to the ID contained in the
// request JWT.
func (h *Handler) UpdateCurrentUser(c *fiber.Ctx) error {
	var updateReq user.UpdateRequest
	if err := c.BodyParser(&updateReq); err != nil {
		fmt.Println(err)
		return h.presenter.ShowBadRequest(c)
	}

	updateReq.UserID = middleware.CurrentUser(c)
	user, err := h.service.UpdateUser(c.Context(), &updateReq)
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
