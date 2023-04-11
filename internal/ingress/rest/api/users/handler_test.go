package users

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/angusgmorrison/realworld/internal/ingress/rest/api/testutil"
	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/angusgmorrison/realworld/pkg/validate"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Handler_Login(t *testing.T) {
	t.Parallel()

	t.Run("when the request is valid it responds 200 OK with the user", func(t *testing.T) {
		t.Parallel()

		// Mock service.
		subject := &user.AuthenticatedUser{
			User: &user.User{
				ID:           uuid.New(),
				Username:     "testuser",
				Email:        "test@test.com",
				PasswordHash: "abc",
				Bio:          "Test bio",
				ImageURL:     "https://test.com/image.png",
			},
			Token: "test-token",
		}
		mockService := &mockUserService{
			AuthenticateFn: func(c context.Context, req *user.AuthRequest) (*user.AuthenticatedUser, error) {
				return subject, nil
			},
		}

		// Set up request.
		server := testutil.NewServer(t)
		server.Post("/api/users/login", NewHandler(mockService).Login)
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/users/login",
			strings.NewReader(`{"email":"test@test.com","password":"test"}`),
		)
		req.Header.Add("Content-Type", "application/json")

		// Expected output.
		wantUserRes := newUserResponseFromDomain(subject.User).withToken(subject.Token)
		wantBody, err := json.Marshal(wantUserRes)
		require.NoError(t, err, "marshal user response")

		res, err := server.Test(req)
		require.NoError(t, err)

		// Check status code.
		assert.Equal(t, http.StatusOK, res.StatusCode)

		// Check body.
		bodyBytes, err := io.ReadAll(res.Body)
		require.NoError(t, err, "read response body")
		assert.JSONEq(t, string(wantBody), string(bodyBytes))

	})

	t.Run("when the request is malformed it responds 400 Bad Request", func(t *testing.T) {
		t.Parallel()

		server := testutil.NewServer(t)
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
		server := testutil.NewServer(t)
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
		server := testutil.NewServer(t)
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
		server := testutil.NewServer(t)
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

func Test_Handler_Register(t *testing.T) {
	t.Parallel()

	t.Run("when the request is valid it responds 201 Created with the user", func(t *testing.T) {
		t.Parallel()

		// Mock service.
		subject := &user.AuthenticatedUser{
			User: &user.User{
				ID:           uuid.New(),
				Username:     "testuser",
				Email:        "test@test.com",
				PasswordHash: "abc",
				Bio:          "Test bio",
				ImageURL:     "https://test.com/image.png",
			},
			Token: "test-token",
		}
		mockService := &mockUserService{
			RegisterFn: func(c context.Context, req *user.RegistrationRequest) (*user.AuthenticatedUser, error) {
				return subject, nil
			},
		}

		// Set up request.
		server := testutil.NewServer(t)
		server.Post("/api/users", NewHandler(mockService).Register)
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/users",
			strings.NewReader(`{"email":"test@test.com", "username": "testuser", "password":"test"}`),
		)
		req.Header.Add("Content-Type", "application/json")

		// Expected output.
		wantUserRes := newUserResponseFromDomain(subject.User).withToken(subject.Token)
		wantBody, err := json.Marshal(wantUserRes)
		require.NoError(t, err, "marshal user response")

		res, err := server.Test(req)
		require.NoError(t, err)

		// Check status code.
		assert.Equal(t, http.StatusCreated, res.StatusCode)

		// Check body.
		bodyBytes, err := io.ReadAll(res.Body)
		require.NoError(t, err, "read response body")
		assert.JSONEq(t, string(wantBody), string(bodyBytes))
	})

	t.Run("when the request is malformed it responds 400 Bad Request", func(t *testing.T) {
		t.Parallel()

		server := testutil.NewServer(t)
		server.Post("/api/users", NewHandler(nil).Login)
		req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{`))
		req.Header.Add("Content-Type", "application/json")

		resp, err := server.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("when the user service responds with an error", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name           string
			mockServiceErr error
			wantStatus     int
		}{
			{
				name:           "when the error is ErrUserExists it responds 422 Unprocessable Entity",
				mockServiceErr: user.ErrUserExists,
				wantStatus:     fiber.StatusUnprocessableEntity,
			},
			{
				name:           "when the error is unhandled it responds 500 Internal Server Error",
				mockServiceErr: errors.New("unhandled error"),
				wantStatus:     fiber.StatusInternalServerError,
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				mockService := &mockUserService{
					RegisterFn: func(c context.Context, req *user.RegistrationRequest) (*user.AuthenticatedUser, error) {
						return nil, tc.mockServiceErr
					},
				}
				server := testutil.NewServer(t)
				server.Post("/api/users", NewHandler(mockService).Register)

				// Set up request.
				req := httptest.NewRequest(
					http.MethodPost,
					"/api/users", strings.NewReader(`{"email":"test@test.com", "username": "testuser", "password":"test"}`),
				)
				req.Header.Add("Content-Type", "application/json")

				// Make request.
				res, err := server.Test(req)

				require.NoError(t, err)
				assert.Equal(t, tc.wantStatus, res.StatusCode)
			})
		}
	})
}

func Test_Handler_GetCurrentUser(t *testing.T) {
	t.Parallel()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "generate RSA keypair")
	authMW := testutil.NewRS256AuthMiddleware(t, key.Public())

	t.Run("when the request is valid it responds 200 OK with the user", func(t *testing.T) {
		t.Parallel()

		subject := &user.User{
			ID:       uuid.New(),
			Username: "test",
			Email:    "test@test.com",
			Bio:      "I am a test user",
			ImageURL: "https://test.com/image.png",
		}

		// Set up server.
		mockService := &mockUserService{
			GetFn: func(c context.Context, id uuid.UUID) (*user.User, error) {
				return subject, nil
			},
		}
		server := testutil.NewServer(t)
		server.Use(authMW)
		server.Get("/api/users", NewHandler(mockService).GetCurrentUser)

		// Set up token.
		claims := jwt.MapClaims{
			"sub": subject.ID.String(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		signedToken, err := token.SignedString(key)
		require.NoError(t, err, "sign token")

		// Expected output.
		wantUserRes := newUserResponseFromDomain(subject).withToken(signedToken)
		wantBody, err := json.Marshal(wantUserRes)
		require.NoError(t, err, "marshal user response")

		// Set up request.
		req := httptest.NewRequest(http.MethodGet, "/api/users", http.NoBody)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", signedToken))

		// Make request.
		res, err := server.Test(req)

		require.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, res.StatusCode)
		bodyBytes, err := io.ReadAll(res.Body)
		require.NoError(t, err, "read response body")
		assert.JSONEq(t, string(wantBody), string(bodyBytes))
	})

	t.Run("when the request is invalid", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name           string
			mockServiceErr error
			wantStatus     int
		}{
			{
				name:           "when the user service responds with ErrUserNotFound it responds 404 Not Found",
				mockServiceErr: user.ErrUserNotFound,
				wantStatus:     fiber.StatusNotFound,
			},
			{
				name:           "when the user service responds with an unhandled error it responds 500 Internal Server Error",
				mockServiceErr: errors.New("unhandled error"),
				wantStatus:     fiber.StatusInternalServerError,
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				mockService := &mockUserService{
					GetFn: func(c context.Context, id uuid.UUID) (*user.User, error) {
						return nil, tc.mockServiceErr
					},
				}
				server := testutil.NewServer(t)
				server.Use(authMW)
				server.Get("/api/users", NewHandler(mockService).GetCurrentUser)

				// Set up token.
				claims := jwt.MapClaims{
					"sub": uuid.New().String(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				signedToken, err := token.SignedString(key)
				require.NoError(t, err, "sign token")

				// Set up request.
				req := httptest.NewRequest(http.MethodGet, "/api/users", http.NoBody)
				req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", signedToken))

				// Make request.
				res, err := server.Test(req)

				require.NoError(t, err)
				assert.Equal(t, tc.wantStatus, res.StatusCode)
			})
		}
	})
}

// Mock implementation of the user.Service interface. Only the Authenticate method is implemented.
type mockUserService struct {
	AuthenticateFn func(c context.Context, req *user.AuthRequest) (*user.AuthenticatedUser, error)
	RegisterFn     func(c context.Context, req *user.RegistrationRequest) (*user.AuthenticatedUser, error)
	GetFn          func(c context.Context, id uuid.UUID) (*user.User, error)
}

var _ user.Service = (*mockUserService)(nil)

func (m *mockUserService) Authenticate(c context.Context, req *user.AuthRequest) (*user.AuthenticatedUser, error) {
	return m.AuthenticateFn(c, req)
}

func (m *mockUserService) Register(c context.Context, req *user.RegistrationRequest) (*user.AuthenticatedUser, error) {
	return m.RegisterFn(c, req)
}

func (m *mockUserService) Get(c context.Context, id uuid.UUID) (*user.User, error) {
	return m.GetFn(c, id)
}

func (m *mockUserService) Update(c context.Context, req *user.UpdateRequest) (*user.User, error) {
	panic("not implemented")
}
