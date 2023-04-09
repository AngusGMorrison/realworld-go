package users

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/angusgmorrison/realworld/pkg/validate"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Handler_Login(t *testing.T) {
	t.Parallel()

	t.Run("when the request is valid it responds 200 OK with the user", func(t *testing.T) {

	})

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

func Test_Handler_GetCurrentUser(t *testing.T) {
	t.Parallel()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "generate RSA keypair")
	authMW := jwtware.New(jwtware.Config{
		SigningKey:    key.PublicKey,
		SigningMethod: "RS256",
	})

	t.Run("when the request is valid it responds 200 OK with the user", func(t *testing.T) {

	})

	t.Run("when the auth token is invalid it responds 401 Unauthorized", func(t *testing.T) {
		t.Parallel()

		server := newTestServer(t)
		server.Use(authMW)
		server.Get("/api/users", NewHandler(nil).GetCurrentUser)
		claims := jwt.MapClaims{
			"userID": "invalid-uuid",
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		signedToken, err := token.SignedString(key)
		require.NoError(t, err, "sign token")
		req := httptest.NewRequest(http.MethodGet, "/api/users", http.NoBody)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", signedToken))

		resp, err := server.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("when the user service returns AuthError it responds 401 Unauthorized", func(t *testing.T) {
	})

	t.Run("when the user service returns an unhandled error it responds 500 Internal Server Error", func(t *testing.T) {
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

func (m *mockUserService) Get(c context.Context, id uuid.UUID) (*user.AuthenticatedUser, error) {
	panic("not implemented")
}

func (m *mockUserService) Update(c context.Context, req *user.UpdateRequest) (*user.AuthenticatedUser, error) {
	panic("not implemented")
}
