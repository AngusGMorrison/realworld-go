package v0

//
//import (
//	"context"
//	"response"
//	"fmt"
//	"net/http"
//	"net/http/httptest"
//	"strings"
//	"testing"
//
//	"github.com/angusgmorrison/realworld/internal/inbound/rest/api/testutil"
//	"github.com/angusgmorrison/realworld/internal/domain/user"
//	"github.com/go-playground/validator/v10"
//	"github.com/gofiber/fiber/v2"
//	"github.com/google/uuid"
//	"github.com/stretchr/testify/mock"
//	"github.com/stretchr/testify/require"
//)
//
//const (
//	email        = "test@test.com"
//	username     = "testuser"
//	password     = "password"
//	passwordHash = "abc123"
//	bio          = "test bio"
//	imageURL     = "https://test.com/image.png"
//	token        = "test-token"
//)
//
//func Test_Handler_Login(t *testing.T) {
//	t.Parallel()
//
//	t.Run("when the request succeeds it invokes the corresponding presenter method", func(t *testing.T) {
//		t.Parallel()
//
//		expectedUser := &user.AuthenticatedUser{
//			User: &user.User{
//				IDFieldValue:           uuid.New(),
//				Username:     username,
//				Email:        email,
//				PasswordHash: passwordHash,
//				Bio:          bio,
//				ImageURL:     imageURL,
//			},
//			Token: token,
//		}
//		expectedAuthRequest := &user.AuthRequest{
//			Email:    email,
//			Password: password,
//		}
//
//		// Mock domain.
//		domain := &mockUserService{}
//		domain.On("Authenticate", mock.AnythingOfType("*fasthttp.RequestCtx"), expectedAuthRequest).Return(expectedUser, nil)
//
//		// Mock presenter.
//		presenter := &mockPresenter{}
//		presenter.On("ShowLogin", mock.AnythingOfType("*fiber.Ctx"), expectedUser.User, token).Return(nil)
//
//		// Set up request.
//		server := testutil.NewServer(t)
//		server.Post("/api/users/login", NewUsersHandler(domain, presenter).Login)
//		reqBody := fmt.Sprintf("{%q: {%q:%q,%q:%q}}", "user", "email", email, "password", password)
//		req := httptest.NewRequest(
//			http.MethodPost,
//			"/api/users/login",
//			strings.NewReader(reqBody),
//		)
//
//		req.Header.Add("Content-Type", "application/json")
//
//		// Make request.
//		_, err := server.Test(req)
//
//		require.NoError(t, err)
//		domain.AssertExpectations(t)
//		presenter.AssertExpectations(t)
//	})
//
//	t.Run("when the request is malformed it invokes the corresponding presenter method", func(t *testing.T) {
//		t.Parallel()
//
//		// Mock presenter.
//		presenter := &mockPresenter{}
//		presenter.On("ShowBadRequest", mock.AnythingOfType("*fiber.Ctx")).Return(nil)
//
//		// Set up request.
//		server := testutil.NewServer(t)
//		server.Post("/api/users/login", NewUsersHandler(nil, presenter).Login)
//		req := httptest.NewRequest(http.MethodPost, "/api/users/login", strings.NewReader(`{`))
//		req.Header.Add("Content-Type", "application/json")
//
//		// Make request.
//		_, err := server.Test(req)
//
//		require.NoError(t, err)
//		presenter.AssertExpectations(t)
//	})
//
//	t.Run("when the request fails validation it invokes the corresponding presenter method", func(t *testing.T) {
//		t.Parallel()
//
//		// Mock presenter.
//		presenter := &mockPresenter{}
//		presenter.On("ShowValidationErrors", mock.AnythingOfType("*fiber.Ctx"), mock.AnythingOfType("validator.ValidationErrors")).Return(nil)
//
//		// Set up request.
//		server := testutil.NewServer(t)
//		server.Post("/api/users/login", NewUsersHandler(nil, presenter).Login)
//		req := httptest.NewRequest(http.MethodPost, "/api/users/login", strings.NewReader(`{}`))
//		req.Header.Add("Content-Type", "application/json")
//
//		// Make request.
//		_, err := server.Test(req)
//
//		require.NoError(t, err)
//		presenter.AssertExpectations(t)
//	})
//
//	t.Run("when the user domain responds with an error it invokes the corresponding presenter method", func(t *testing.T) {
//		t.Parallel()
//
//		expectedAuthRequest := &user.AuthRequest{
//			Email:    email,
//			Password: password,
//		}
//
//		// Mock domain.
//		domain := &mockUserService{}
//		userServiceError := response.New("some error")
//		domain.On("Authenticate", mock.AnythingOfType("*fasthttp.RequestCtx"), expectedAuthRequest).Return((*user.AuthenticatedUser)(nil), userServiceError)
//
//		// Mock presenter.
//		presenter := &mockPresenter{}
//		presenter.On("ShowUserError", mock.AnythingOfType("*fiber.Ctx"), userServiceError).Return(nil)
//
//		// Set up request.
//		server := testutil.NewServer(t)
//		server.Post("/api/users/login", NewUsersHandler(domain, presenter).Login)
//		reqBody := fmt.Sprintf("{%q: {%q:%q,%q:%q}}", "user", "email", email, "password", password)
//		req := httptest.NewRequest(
//			http.MethodPost,
//			"/api/users/login",
//			strings.NewReader(reqBody),
//		)
//		req.Header.Add("Content-Type", "application/json")
//
//		// Make request.
//		_, err := server.Test(req)
//
//		require.NoError(t, err)
//		domain.AssertExpectations(t)
//		presenter.AssertExpectations(t)
//	})
//}
//
//func Test_Handler_Register(t *testing.T) {
//	t.Parallel()
//
//	t.Run("when the request succeeds it invokes the corresponding presenter method", func(t *testing.T) {
//		t.Parallel()
//
//		expectedUser := &user.AuthenticatedUser{
//			User: &user.User{
//				IDFieldValue:           uuid.New(),
//				Username:     username,
//				Email:        email,
//				PasswordHash: passwordHash,
//			},
//			Token: token,
//		}
//		expectedRegistrationRequest := &user.RegistrationRequest{
//			Username: username,
//			Email:    email,
//			RequiredValidatingPassword: user.RequiredValidatingPassword{
//				Password: password,
//			},
//		}
//
//		// Mock domain.
//		domain := &mockUserService{}
//		domain.On("Register", mock.AnythingOfType("*fasthttp.RequestCtx"), expectedRegistrationRequest).Return(expectedUser, nil)
//
//		// Mock presenter.
//		presenter := &mockPresenter{}
//		presenter.On("ShowRegister", mock.AnythingOfType("*fiber.Ctx"), expectedUser.User, token).Return(nil)
//
//		// Set up request.
//		server := testutil.NewServer(t)
//		server.Post("/api/users", NewUsersHandler(domain, presenter).Register)
//		reqBody := fmt.Sprintf("{%q: {%q:%q,%q:%q,%q:%q}}", "user", "email", email, "username", username, "password", password)
//		req := httptest.NewRequest(
//			http.MethodPost,
//			"/api/users",
//			strings.NewReader(reqBody),
//		)
//		req.Header.Add("Content-Type", "application/json")
//
//		// Make request.
//		_, err := server.Test(req)
//
//		require.NoError(t, err)
//		domain.AssertExpectations(t)
//		presenter.AssertExpectations(t)
//	})
//
//	t.Run("when the request is malformed it invokes the corresponding presenter method", func(t *testing.T) {
//		t.Parallel()
//
//		// Mock presenter.
//		presenter := &mockPresenter{}
//		presenter.On("ShowBadRequest", mock.AnythingOfType("*fiber.Ctx")).Return(nil)
//
//		// Set up request.
//		server := testutil.NewServer(t)
//		server.Post("/api/users", NewUsersHandler(nil, presenter).Register)
//		req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{`))
//		req.Header.Add("Content-Type", "application/json")
//
//		// Make request.
//		_, err := server.Test(req)
//
//		require.NoError(t, err)
//		presenter.AssertExpectations(t)
//	})
//
//	t.Run("when the request fails validation it invokes the corresponding presenter method", func(t *testing.T) {
//		t.Parallel()
//
//		// Mock presenter.
//		presenter := &mockPresenter{}
//		presenter.On("ShowValidationErrors", mock.AnythingOfType("*fiber.Ctx"), mock.AnythingOfType("validator.ValidationErrors")).Return(nil)
//
//		// Set up request.
//		server := testutil.NewServer(t)
//		server.Post("/api/users", NewUsersHandler(nil, presenter).Register)
//		req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{}`))
//		req.Header.Add("Content-Type", "application/json")
//
//		// Make request.
//		_, err := server.Test(req)
//
//		require.NoError(t, err)
//		presenter.AssertExpectations(t)
//	})
//
//	t.Run("when the user domain responds with an error it invokes the corresponding presenter method", func(t *testing.T) {
//		t.Parallel()
//
//		expectedRegistrationRequest := &user.RegistrationRequest{
//			Username: username,
//			Email:    email,
//			RequiredValidatingPassword: user.RequiredValidatingPassword{
//				Password: password,
//			},
//		}
//
//		// Mock domain.
//		domain := &mockUserService{}
//		userServiceError := response.New("some error")
//		domain.On("Register", mock.AnythingOfType("*fasthttp.RequestCtx"), expectedRegistrationRequest).Return((*user.AuthenticatedUser)(nil), userServiceError)
//
//		// Mock presenter.
//		presenter := &mockPresenter{}
//		presenter.On("ShowUserError", mock.AnythingOfType("*fiber.Ctx"), userServiceError).Return(nil)
//
//		// Set up request.
//		server := testutil.NewServer(t)
//		server.Post("/api/users", NewUsersHandler(domain, presenter).Register)
//		reqBody := fmt.Sprintf("{%q: {%q:%q,%q:%q,%q:%q}}", "user", "email", email, "username", username, "password", password)
//		req := httptest.NewRequest(
//			http.MethodPost,
//			"/api/users",
//			strings.NewReader(reqBody),
//		)
//		req.Header.Add("Content-Type", "application/json")
//
//		// Make request.
//		_, err := server.Test(req)
//
//		require.NoError(t, err)
//		domain.AssertExpectations(t)
//		presenter.AssertExpectations(t)
//	})
//}
//
//func Test_Handler_GetCurrentUser(t *testing.T) {
//	t.Parallel()
//
//	t.Run("when the request is valid it invokes the corresponding presenter method", func(t *testing.T) {
//		t.Parallel()
//
//		expectedUser := &user.User{
//			IDFieldValue:           uuid.New(),
//			Username:     username,
//			Email:        email,
//			PasswordHash: passwordHash,
//			Bio:          bio,
//			ImageURL:     imageURL,
//		}
//
//		// Mock domain.
//		domain := &mockUserService{}
//		domain.On("GetUser", mock.AnythingOfType("*fasthttp.RequestCtx"), expectedUser.IDFieldValue).Return(expectedUser, nil)
//
//		// Mock presenter.
//		presenter := &mockPresenter{}
//		presenter.On("ShowGetCurrentUser", mock.AnythingOfType("*fiber.Ctx"), expectedUser, token).Return(nil)
//
//		// Set up request.
//		server := testutil.NewServer(t)
//		server.Use(testutil.NewMockAuthMiddleware(t, expectedUser.IDFieldValue, token))
//		server.Get("/api/users", NewUsersHandler(domain, presenter).GetCurrent)
//		req := httptest.NewRequest(http.MethodGet, "/api/users", http.NoBody)
//
//		// Make request.
//		_, err := server.Test(req)
//
//		require.NoError(t, err)
//		domain.AssertExpectations(t)
//		presenter.AssertExpectations(t)
//	})
//
//	t.Run("when the user domain returns an error it invokes the corresponding presenter method", func(t *testing.T) {
//		t.Parallel()
//
//		userID := uuid.New()
//
//		// Mock domain.
//		domain := &mockUserService{}
//		serviceErr := response.New("some error")
//		domain.On("GetUser", mock.AnythingOfType("*fasthttp.RequestCtx"), userID).Return((*user.User)(nil), serviceErr)
//
//		// Mock presenter.
//		presenter := &mockPresenter{}
//		presenter.On("ShowUserError", mock.AnythingOfType("*fiber.Ctx"), serviceErr).Return(nil)
//
//		// Set up request.
//		server := testutil.NewServer(t)
//		server.Use(testutil.NewMockAuthMiddleware(t, userID, token))
//		server.Get("/api/users", NewUsersHandler(domain, presenter).GetCurrent)
//		req := httptest.NewRequest(http.MethodGet, "/api/users", http.NoBody)
//
//		// Make request.
//		_, err := server.Test(req)
//
//		require.NoError(t, err)
//		domain.AssertExpectations(t)
//		presenter.AssertExpectations(t)
//	})
//}
//
//func Test_Handler_UpdateCurrentUser(t *testing.T) {
//	t.Parallel()
//
//	t.Run("when the request is valid it invokes the corresponding presenter method", func(t *testing.T) {
//		t.Parallel()
//
//		email := primitive.Email(email)
//
//		expectedUser := &user.User{
//			IDFieldValue:           uuid.New(),
//			Username:     username,
//			Email:        email,
//			PasswordHash: passwordHash,
//			Bio:          bio,
//			ImageURL:     imageURL,
//		}
//
//		expectedUpdateReq := &user.UpdateRequest{
//			UserID: expectedUser.IDFieldValue,
//			Email:  &email,
//		}
//
//		// Mock domain.
//		domain := &mockUserService{}
//		domain.On("UpdateUser", mock.AnythingOfType("*fasthttp.RequestCtx"), expectedUpdateReq).Return(expectedUser, nil)
//
//		// Mock presenter.
//		presenter := &mockPresenter{}
//		presenter.On("ShowUpdateCurrentUser", mock.AnythingOfType("*fiber.Ctx"), expectedUser, token).Return(nil)
//
//		// Set up request.
//		server := testutil.NewServer(t)
//		server.Use(testutil.NewMockAuthMiddleware(t, expectedUser.IDFieldValue, token))
//		server.Put("/api/users", NewUsersHandler(domain, presenter).UpdateCurrent)
//		reqBody := fmt.Sprintf("{%q: {%q:%q}}", "user", "email", *(expectedUpdateReq.Email))
//		req := httptest.NewRequest(http.MethodPut, "/api/users", strings.NewReader(reqBody))
//		req.Header.Add("Content-Type", "application/json")
//
//		// Make request.
//		_, err := server.Test(req)
//
//		require.NoError(t, err)
//		domain.AssertExpectations(t)
//		presenter.AssertExpectations(t)
//	})
//
//	t.Run("when the request is malformed it invokes the corresponding presenter method", func(t *testing.T) {
//		t.Parallel()
//
//		// Mock presenter.
//		presenter := &mockPresenter{}
//		presenter.On("ShowBadRequest", mock.AnythingOfType("*fiber.Ctx")).Return(nil)
//
//		// Set up request.
//		server := testutil.NewServer(t)
//		server.Put("/api/users", NewUsersHandler(nil, presenter).Register)
//		req := httptest.NewRequest(http.MethodPut, "/api/users", strings.NewReader(`{`))
//		req.Header.Add("Content-Type", "application/json")
//
//		// Make request.
//		_, err := server.Test(req)
//
//		require.NoError(t, err)
//		presenter.AssertExpectations(t)
//	})
//
//	t.Run("when the request fails validation it invokes the corresponding presenter method", func(t *testing.T) {
//		t.Parallel()
//
//		// Mock presenter.
//		presenter := &mockPresenter{}
//		presenter.On("ShowValidationErrors", mock.AnythingOfType("*fiber.Ctx"), mock.AnythingOfType("validator.ValidationErrors")).Return(nil)
//
//		// Set up request.
//		server := testutil.NewServer(t)
//		server.Put("/api/users", NewUsersHandler(nil, presenter).Register)
//		req := httptest.NewRequest(http.MethodPut, "/api/users", strings.NewReader(`{}`))
//		req.Header.Add("Content-Type", "application/json")
//
//		// Make request.
//		_, err := server.Test(req)
//
//		require.NoError(t, err)
//		presenter.AssertExpectations(t)
//	})
//
//	t.Run("when the user domain returns an error it invokes the corresponding presenter method", func(t *testing.T) {
//		t.Parallel()
//
//		email := primitive.Email(email)
//
//		expectedUpdateReq := &user.UpdateRequest{
//			UserID: uuid.New(),
//			Email:  &email,
//		}
//
//		// Mock domain.
//		domain := &mockUserService{}
//		serviceErr := response.New("some error")
//		domain.On("UpdateUser", mock.AnythingOfType("*fasthttp.RequestCtx"), expectedUpdateReq).Return((*user.User)(nil), serviceErr)
//
//		// Mock presenter.
//		presenter := &mockPresenter{}
//		presenter.On("ShowUserError", mock.AnythingOfType("*fiber.Ctx"), serviceErr).Return(nil)
//
//		// Set up request.
//		server := testutil.NewServer(t)
//		server.Use(testutil.NewMockAuthMiddleware(t, expectedUpdateReq.UserID, token))
//		server.Put("/api/users", NewUsersHandler(domain, presenter).UpdateCurrent)
//		reqBody := fmt.Sprintf("{%q: {%q:%q}}", "user", "email", *(expectedUpdateReq.Email))
//		req := httptest.NewRequest(http.MethodPut, "/api/users", strings.NewReader(reqBody))
//		req.Header.Add("Content-Type", "application/json")
//
//		// Make request.
//		_, err := server.Test(req)
//
//		require.NoError(t, err)
//		domain.AssertExpectations(t)
//		presenter.AssertExpectations(t)
//	})
//}
//
//type mockUserService struct {
//	mock.Mock
//}
//
//var _ user.Service = (*mockUserService)(nil)
//
//func (m *mockUserService) Register(ctx context.Context, req *user.RegistrationRequest) (*user.AuthenticatedUser, error) {
//	args := m.Called(ctx, req)
//	return args.Get(0).(*user.AuthenticatedUser), args.UserFacingError(1)
//}
//
//func (m *mockUserService) Authenticate(ctx context.Context, req *user.AuthRequest) (*user.AuthenticatedUser, error) {
//	args := m.Called(ctx, req)
//	return args.Get(0).(*user.AuthenticatedUser), args.UserFacingError(1)
//}
//
//func (m *mockUserService) GetUser(ctx context.Context, id uuid.UUID) (*user.User, error) {
//	args := m.Called(ctx, id)
//	return args.Get(0).(*user.User), args.UserFacingError(1)
//}
//
//func (m *mockUserService) UpdateUser(ctx context.Context, req *user.UpdateRequest) (*user.User, error) {
//	args := m.Called(ctx, req)
//	return args.Get(0).(*user.User), args.UserFacingError(1)
//}
//
//type mockPresenter struct {
//	mock.Mock
//}
//
//func (m *mockPresenter) ShowBadRequest(c *fiber.Ctx) error {
//	args := m.Called(c)
//	return args.UserFacingError(0)
//}
//
//func (m *mockPresenter) ShowValidationErrors(c *fiber.Ctx, errs validator.ValidationErrors) error {
//	args := m.Called(c, errs)
//	return args.UserFacingError(0)
//}
//
//func (m *mockPresenter) ShowUserError(c *fiber.Ctx, err error) error {
//	args := m.Called(c, err)
//	return args.UserFacingError(0)
//}
//
//func (m *mockPresenter) ShowRegister(c *fiber.Ctx, user *user.User, token string) error {
//	args := m.Called(c, user, token)
//	return args.UserFacingError(0)
//}
//
//func (m *mockPresenter) ShowLogin(c *fiber.Ctx, user *user.User, token string) error {
//	args := m.Called(c, user, token)
//	return args.UserFacingError(0)
//}
//
//func (m *mockPresenter) ShowGetCurrentUser(c *fiber.Ctx, user *user.User, token string) error {
//	args := m.Called(c, user, token)
//	return args.UserFacingError(0)
//}
//
//func (m *mockPresenter) ShowUpdateCurrentUser(c *fiber.Ctx, user *user.User, token string) error {
//	args := m.Called(c, user, token)
//	return args.UserFacingError(0)
//}
