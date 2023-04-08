package rest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/angusgmorrison/realworld/service/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_UsersGroup_LoginHandler(t *testing.T) {
	t.Parallel()

	t.Run("when the request is malformed it responds 400 Bad Request", func(t *testing.T) {
		t.Parallel()

		s := NewServer(&mockUserService{}, testOptions(t)...)
		req := httptest.NewRequest(http.MethodPost, "/api/users/login", strings.NewReader(`{`))
		req.Header.Add("Content-Type", "application/json")

		resp, err := s.innerServer.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("when the user service responds with an auth error it responds 401 Unauthorized", func(t *testing.T) {
		t.Parallel()

		mockService := &mockUserService{
			AuthenticateFn: func(c context.Context, req *user.AuthRequest) (*user.User, string, error) {
				return nil, "", &user.AuthError{}
			},
		}
		s := NewServer(mockService, testOptions(t)...)
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/users/login",
			strings.NewReader(`{"email":"test@test.com","password":"test"}`),
		)
		req.Header.Add("Content-Type", "application/json")

		resp, err := s.innerServer.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("when the user service responds with a validation error it responds 422 Unprocessable Entity", func(t *testing.T) {
		t.Parallel()

		mockService := &mockUserService{
			AuthenticateFn: func(c context.Context, req *user.AuthRequest) (*user.User, string, error) {
				validationErr := &user.ValidationError{}
				validationErr.Push(user.ErrEmailAddressUnparseable, user.ErrPasswordTooLong)
				return nil, "", validationErr
			},
		}
		s := NewServer(mockService, testOptions(t)...)
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/users/login",
			strings.NewReader(`{"email":"test@test.com","password":"test"}`),
		)
		req.Header.Add("Content-Type", "application/json")

		resp, err := s.innerServer.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

		gotBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.JSONEq(
			t,
			`{"errors":{"email":["is invalid"],"password":["length exceeds 72 bytes"]}}`,
			string(gotBody),
		)
	})

	t.Run("when the user service responds with an unhandled error it responds 500 Internal Server Error", func(t *testing.T) {
		t.Parallel()

		mockService := &mockUserService{
			AuthenticateFn: func(c context.Context, req *user.AuthRequest) (*user.User, string, error) {
				return nil, "", errors.New("unhandled error")
			},
		}
		s := NewServer(mockService, testOptions(t)...)
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/users/login",
			strings.NewReader(`{"email":"test@test.com","password":"test"}`),
		)
		req.Header.Add("Content-Type", "application/json")

		resp, err := s.innerServer.Test(req)

		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// testOptions returns a set of server option for suppressing logging and stack
// traces during tests.
func testOptions(t *testing.T) []Option {
	t.Helper()

	return []Option{
		&LogOutputOption{W: io.Discard},
		&StackTraceOption{Enable: false},
	}
}

// Mock implementation of the user.Service interface. Only the Authenticate method is implemented.
type mockUserService struct {
	AuthenticateFn func(c context.Context, req *user.AuthRequest) (*user.User, string, error)
}

var _ user.Service = (*mockUserService)(nil)

func (m *mockUserService) Authenticate(c context.Context, req *user.AuthRequest) (*user.User, string, error) {
	return m.AuthenticateFn(c, req)
}

func (m *mockUserService) Register(c context.Context, req *user.RegisterRequest) (*user.User, error) {
	panic("not implemented")
}

func (m *mockUserService) Get(c context.Context, id uuid.UUID) (*user.User, error) {
	panic("not implemented")
}

func (m *mockUserService) Update(c context.Context, req *user.UpdateRequest) (*user.User, error) {
	panic("not implemented")
}
