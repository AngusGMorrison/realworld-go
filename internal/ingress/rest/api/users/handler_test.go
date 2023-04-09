package users

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/angusgmorrison/realworld/pkg/validate"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_UsersGroup_Login(t *testing.T) {
	t.Parallel()

	t.Run("when the request is malformed it responds 400 Bad Request", func(t *testing.T) {
		t.Parallel()

		server := newTestServer(t)
		server.Post("/api/users/login", NewHandler(nil).Login)
		req := httptest.NewRequest(http.MethodPost, "/api/users/login", strings.NewReader(`{`))
		req.Header.Add("Content-Type", "application/json")

		resp, err := server.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("when the user service responds with an auth error it responds 401 Unauthorized", func(t *testing.T) {
		t.Parallel()

		mockService := &mockUserService{
			AuthenticateFn: func(c context.Context, req *user.AuthRequest) (*user.AuthenticatedUser, error) {
				return nil, &user.AuthError{}
			},
		}
		server := newTestServer(t)
		server.Post("/api/users/login", NewHandler(mockService).Login)
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/users/login",
			strings.NewReader(`{"email":"test@test.com","password":"test"}`),
		)
		req.Header.Add("Content-Type", "application/json")

		resp, err := server.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("when the user service responds with a validation error it responds 422 Unprocessable Entity", func(t *testing.T) {
		t.Parallel()

		invalidEmail := "invalid-email"
		mockService := &mockUserService{
			AuthenticateFn: func(c context.Context, req *user.AuthRequest) (*user.AuthenticatedUser, error) {
				err := validate.Struct(struct {
					Email string `json:"email" validate:"required,email"`
				}{
					Email: invalidEmail,
				})

				return nil, err
			},
		}
		server := newTestServer(t)
		server.Post("/api/users/login", NewHandler(mockService).Login)
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/users/login",
			strings.NewReader(`{"email":"test@test.com","password":"test"}`),
		)
		req.Header.Add("Content-Type", "application/json")

		resp, err := server.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

		gotBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.JSONEq(
			t,
			`{"errors":{"email":["\"invalid-email\" is not a valid email address"]}}`,
			string(gotBody),
		)
	})

	t.Run("when the user service responds with an unhandled error it responds 500 Internal Server Error", func(t *testing.T) {
		t.Parallel()

		mockService := &mockUserService{
			AuthenticateFn: func(c context.Context, req *user.AuthRequest) (*user.AuthenticatedUser, error) {
				return nil, errors.New("unhandled error")
			},
		}
		server := newTestServer(t)
		server.Post("/api/users/login", NewHandler(mockService).Login)
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/users/login",
			strings.NewReader(`{"email":"test@test.com","password":"test"}`),
		)
		req.Header.Add("Content-Type", "application/json")

		resp, err := server.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func newTestServer(t *testing.T) *fiber.App {
	t.Helper()

	return fiber.New(fiber.Config{
		AppName:      "realworld-hexagonal-test",
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	})
}

// Mock implementation of the user.Service interface. Only the Authenticate method is implemented.
type mockUserService struct {
	AuthenticateFn func(c context.Context, req *user.AuthRequest) (*user.AuthenticatedUser, error)
}

var _ user.Service = (*mockUserService)(nil)

func (m *mockUserService) Authenticate(c context.Context, req *user.AuthRequest) (*user.AuthenticatedUser, error) {
	return m.AuthenticateFn(c, req)
}

func (m *mockUserService) Register(c context.Context, req *user.RegistrationRequest) (*user.AuthenticatedUser, error) {
	panic("not implemented")
}

func (m *mockUserService) Get(c context.Context, id uuid.UUID) (*user.User, error) {
	panic("not implemented")
}

func (m *mockUserService) Update(c context.Context, req *user.UpdateRequest) (*user.User, error) {
	panic("not implemented")
}
