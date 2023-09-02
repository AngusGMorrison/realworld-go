package v0

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/angusgmorrison/realworld-go/internal/domain/user"
	"github.com/angusgmorrison/realworld-go/internal/testutil"
	"github.com/angusgmorrison/realworld-go/pkg/option"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

func Test_NewUsersHandler(t *testing.T) {
	t.Parallel()

	service := &testutil.MockUserService{}
	jwtProvider := &mockJWTProvider{}
	wantHandler := &UsersHandler{
		service:     service,
		jwtProvider: jwtProvider,
	}

	gotHandler := NewUsersHandler(service, jwtProvider)
	assert.Equal(t, wantHandler, gotHandler)
}

func Test_UsersHandler_Register(t *testing.T) {
	t.Parallel()

	validUsernameCandidate := user.RandomUsernameCandidate()
	validEmailCandidate := user.RandomEmailAddressCandidate()
	validPasswordCandidate := user.RandomPasswordCandidate()

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name        string
			requestBody string
			setupMocks  func(service *testutil.MockUserService, jwtProvider *mockJWTProvider)
			assertError func(t *testing.T, err error)
			assertMocks func(t *testing.T, service *testutil.MockUserService, jwtProvider *mockJWTProvider)
		}{
			{
				name:        "JSON syntax error",
				requestBody: `{`,
				setupMocks:  func(service *testutil.MockUserService, jwtProvider *mockJWTProvider) {},
				assertError: func(t *testing.T, err error) {
					t.Helper()
					var syntaxErr *json.SyntaxError
					assert.ErrorAs(t, err, &syntaxErr)
				},
				assertMocks: func(t *testing.T, service *testutil.MockUserService, jwtProvider *mockJWTProvider) {
					t.Helper()
					service.AssertNotCalled(t, "Register")
					jwtProvider.AssertNotCalled(t, "TokenFor")
				},
			},
			{
				name: "parse domain model error",
				requestBody: fmt.Sprintf(`{"user": {"username": "", "email": %q, "password": %q}}`,
					validEmailCandidate, validPasswordCandidate),
				setupMocks: func(service *testutil.MockUserService, jwtProvider *mockJWTProvider) {},
				assertError: func(t *testing.T, err error) {
					t.Helper()
					var validationErrs user.ValidationErrors
					assert.ErrorAs(t, err, &validationErrs)
				},
				assertMocks: func(t *testing.T, service *testutil.MockUserService, jwtProvider *mockJWTProvider) {
					t.Helper()
					service.AssertNotCalled(t, "Register")
					jwtProvider.AssertNotCalled(t, "TokenFor")
				},
			},
			{
				name: "service error",
				requestBody: fmt.Sprintf(`{"user": {"username": %q, "email": %q, "password": %q}}`,
					validUsernameCandidate, validEmailCandidate, validPasswordCandidate.Expose()),
				setupMocks: func(service *testutil.MockUserService, jwtProvider *mockJWTProvider) {
					wantRegistrationReq, err := user.ParseRegistrationRequest(
						validUsernameCandidate,
						validEmailCandidate,
						validPasswordCandidate,
					)
					require.NoError(t, err)

					service.On(
						"Register",
						mock.AnythingOfType("*fasthttp.RequestCtx"),
						mock.MatchedBy(testutil.NewRegistrationRequestMatcher(
							t,
							wantRegistrationReq,
							validPasswordCandidate,
						)),
					).Return((*user.User)(nil), assert.AnError)
				},
				assertError: func(t *testing.T, err error) {
					t.Helper()
					assert.ErrorIs(t, err, assert.AnError)
				},
				assertMocks: func(t *testing.T, service *testutil.MockUserService, jwtProvider *mockJWTProvider) {
					t.Helper()
					service.AssertExpectations(t)
					jwtProvider.AssertNotCalled(t, "TokenFor")
				},
			},
			{
				name: "JWTProvider error",
				requestBody: fmt.Sprintf(`{"user": {"username": %q, "email": %q, "password": %q}}`,
					validUsernameCandidate, validEmailCandidate, validPasswordCandidate.Expose()),
				setupMocks: func(service *testutil.MockUserService, jwtProvider *mockJWTProvider) {
					wantRegistrationReq, err := user.ParseRegistrationRequest(
						validUsernameCandidate,
						validEmailCandidate,
						validPasswordCandidate,
					)
					require.NoError(t, err)

					service.On(
						"Register",
						mock.AnythingOfType("*fasthttp.RequestCtx"),
						mock.MatchedBy(testutil.NewRegistrationRequestMatcher(
							t,
							wantRegistrationReq,
							validPasswordCandidate,
						)),
					).Return(&user.User{}, nil)

					jwtProvider.On("TokenFor", uuid.Nil).Return(JWT(""), assert.AnError)
				},
				assertError: func(t *testing.T, err error) {
					t.Helper()
					assert.ErrorIs(t, err, assert.AnError)
				},
				assertMocks: func(t *testing.T, service *testutil.MockUserService, jwtProvider *mockJWTProvider) {
					t.Helper()
					service.AssertExpectations(t)
					jwtProvider.AssertExpectations(t)
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				service := &testutil.MockUserService{}
				jwtProvider := &mockJWTProvider{}
				handler := NewUsersHandler(service, jwtProvider)

				app := fiber.New(fiber.Config{
					ErrorHandler: func(ctx *fiber.Ctx, err error) error {
						tc.assertError(t, err)
						return nil
					},
				})
				app.Post("/", handler.Register)

				req, err := http.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/",
					bytes.NewBufferString(tc.requestBody),
				)
				require.NoError(t, err)
				req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

				tc.setupMocks(service, jwtProvider)

				_, err = app.Test(req)
				require.NoError(t, err)

				tc.assertMocks(t, service, jwtProvider)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service := &testutil.MockUserService{}
		jwtProvider := &mockJWTProvider{}
		handler := NewUsersHandler(service, jwtProvider)

		app := fiber.New()
		app.Post("/", handler.Register)

		requestBody := fmt.Sprintf(`{"user": {"username": %q, "email": %q, "password": %q}}`,
			validUsernameCandidate, validEmailCandidate, validPasswordCandidate.Expose())
		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodPost,
			"/",
			bytes.NewBufferString(requestBody),
		)
		require.NoError(t, err)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		wantRegistrationReq, err := user.ParseRegistrationRequest(
			validUsernameCandidate,
			validEmailCandidate,
			validPasswordCandidate,
		)
		require.NoError(t, err)
		username, err := user.ParseUsername(validUsernameCandidate)
		require.NoError(t, err)
		email, err := user.ParseEmailAddress(validEmailCandidate)
		require.NoError(t, err)
		hash, err := user.ParsePassword(validPasswordCandidate)
		require.NoError(t, err)
		wantUser := user.NewUser(
			uuid.New(),
			username,
			email,
			hash,
			option.None[user.Bio](),
			option.None[user.URL](),
		)
		wantToken := JWT("abc")
		wantStatusCode := fiber.StatusCreated
		wantBody := fmt.Sprintf(`{"user": {"token": %q, "email": %q, "username": %q, "bio": "", "image": ""}}`,
			wantToken, validEmailCandidate, validUsernameCandidate)

		service.On(
			"Register",
			mock.AnythingOfType("*fasthttp.RequestCtx"),
			mock.MatchedBy(testutil.NewRegistrationRequestMatcher(
				t,
				wantRegistrationReq,
				validPasswordCandidate,
			)),
		).Return(wantUser, nil)

		jwtProvider.On("TokenFor", wantUser.ID()).Return(wantToken, nil)

		res, err := app.Test(req)
		require.NoError(t, err)

		gotBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Equal(t, wantStatusCode, res.StatusCode)
		assert.JSONEq(t, wantBody, string(gotBody))
		service.AssertExpectations(t)
		jwtProvider.AssertExpectations(t)
	})
}

func Test_UsersHandlerLogin(t *testing.T) {
	t.Parallel()

	validEmailCandidate := user.RandomEmailAddressCandidate()
	validPasswordCandidate := user.RandomPasswordCandidate()

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name        string
			requestBody string
			setupMocks  func(service *testutil.MockUserService, jwtProvider *mockJWTProvider)
			assertError func(t *testing.T, err error)
			assertMocks func(t *testing.T, service *testutil.MockUserService, jwtProvider *mockJWTProvider)
		}{
			{
				name:        "JSON syntax error",
				requestBody: `{`,
				setupMocks:  func(service *testutil.MockUserService, jwtProvider *mockJWTProvider) {},
				assertError: func(t *testing.T, err error) {
					t.Helper()
					var syntaxErr *json.SyntaxError
					assert.ErrorAs(t, err, &syntaxErr)
				},
				assertMocks: func(t *testing.T, service *testutil.MockUserService, jwtProvider *mockJWTProvider) {
					t.Helper()
					service.AssertNotCalled(t, "Register")
					jwtProvider.AssertNotCalled(t, "TokenFor")
				},
			},
			{
				name:        "parse domain model error",
				requestBody: fmt.Sprintf(`{"user": {"email": "", "password": %q}}`, validPasswordCandidate),
				setupMocks:  func(service *testutil.MockUserService, jwtProvider *mockJWTProvider) {},
				assertError: func(t *testing.T, err error) {
					t.Helper()
					var validationErrs user.ValidationErrors
					assert.ErrorAs(t, err, &validationErrs)
				},
				assertMocks: func(t *testing.T, service *testutil.MockUserService, jwtProvider *mockJWTProvider) {
					t.Helper()
					service.AssertNotCalled(t, "Register")
					jwtProvider.AssertNotCalled(t, "TokenFor")
				},
			},
			{
				name: "service error",
				requestBody: fmt.Sprintf(`{"user": {"email": %q, "password": %q}}`,
					validEmailCandidate, validPasswordCandidate.Expose()),
				setupMocks: func(service *testutil.MockUserService, jwtProvider *mockJWTProvider) {
					wantAuthReq, err := user.ParseAuthRequest(
						validEmailCandidate,
						validPasswordCandidate,
					)
					require.NoError(t, err)

					service.On(
						"Authenticate",
						mock.AnythingOfType("*fasthttp.RequestCtx"),
						wantAuthReq,
					).Return((*user.User)(nil), assert.AnError)
				},
				assertError: func(t *testing.T, err error) {
					t.Helper()
					assert.ErrorIs(t, err, assert.AnError)
				},
				assertMocks: func(t *testing.T, service *testutil.MockUserService, jwtProvider *mockJWTProvider) {
					t.Helper()
					service.AssertExpectations(t)
					jwtProvider.AssertNotCalled(t, "TokenFor")
				},
			},
			{
				name: "JWTProvider error",
				requestBody: fmt.Sprintf(`{"user": {"email": %q, "password": %q}}`,
					validEmailCandidate, validPasswordCandidate.Expose()),
				setupMocks: func(service *testutil.MockUserService, jwtProvider *mockJWTProvider) {
					wantAuthReq, err := user.ParseAuthRequest(
						validEmailCandidate,
						validPasswordCandidate,
					)
					require.NoError(t, err)

					service.On(
						"Authenticate",
						mock.AnythingOfType("*fasthttp.RequestCtx"),
						wantAuthReq,
					).Return(&user.User{}, nil)

					jwtProvider.On("TokenFor", uuid.Nil).Return(JWT(""), assert.AnError)
				},
				assertError: func(t *testing.T, err error) {
					t.Helper()
					assert.ErrorIs(t, err, assert.AnError)
				},
				assertMocks: func(t *testing.T, service *testutil.MockUserService, jwtProvider *mockJWTProvider) {
					t.Helper()
					service.AssertExpectations(t)
					jwtProvider.AssertExpectations(t)
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				service := &testutil.MockUserService{}
				jwtProvider := &mockJWTProvider{}
				handler := NewUsersHandler(service, jwtProvider)

				app := fiber.New(fiber.Config{
					ErrorHandler: func(ctx *fiber.Ctx, err error) error {
						tc.assertError(t, err)
						return nil
					},
				})
				app.Post("/", handler.Login)

				req, err := http.NewRequestWithContext(
					context.Background(),
					http.MethodPost,
					"/",
					bytes.NewBufferString(tc.requestBody),
				)
				require.NoError(t, err)
				req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

				tc.setupMocks(service, jwtProvider)

				_, err = app.Test(req)
				require.NoError(t, err)

				tc.assertMocks(t, service, jwtProvider)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service := &testutil.MockUserService{}
		jwtProvider := &mockJWTProvider{}
		handler := NewUsersHandler(service, jwtProvider)

		app := fiber.New()
		app.Post("/", handler.Login)

		requestBody := fmt.Sprintf(`{"user": {"email": %q, "password": %q}}`,
			validEmailCandidate, validPasswordCandidate.Expose())
		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodPost,
			"/",
			bytes.NewBufferString(requestBody),
		)
		require.NoError(t, err)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		wantAuthReq, err := user.ParseAuthRequest(
			validEmailCandidate,
			validPasswordCandidate,
		)
		require.NoError(t, err)
		email, err := user.ParseEmailAddress(validEmailCandidate)
		require.NoError(t, err)
		hash, err := user.ParsePassword(validPasswordCandidate)
		require.NoError(t, err)
		username := user.RandomUsername(t)
		bio := user.RandomOption[user.Bio](t)
		imageURL := user.RandomOption[user.URL](t)
		wantUser := user.NewUser(
			uuid.New(),
			username,
			email,
			hash,
			bio,
			imageURL,
		)
		wantToken := JWT("abc")
		wantStatusCode := fiber.StatusOK
		wantBody := fmt.Sprintf(`{"user": {"token": %q, "email": %q, "username": %q, "bio": %q, "image": %q}}`,
			wantToken, validEmailCandidate, username, bio.UnwrapOrZero(), imageURL.UnwrapOrZero())

		service.On(
			"Authenticate",
			mock.AnythingOfType("*fasthttp.RequestCtx"),
			wantAuthReq,
		).Return(wantUser, nil)

		jwtProvider.On("TokenFor", wantUser.ID()).Return(wantToken, nil)

		res, err := app.Test(req)
		require.NoError(t, err)

		gotBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Equal(t, wantStatusCode, res.StatusCode)
		assert.JSONEq(t, wantBody, string(gotBody))
		service.AssertExpectations(t)
		jwtProvider.AssertExpectations(t)
	})
}

type mockJWTProvider struct {
	mock.Mock
}

func (m *mockJWTProvider) TokenFor(subject uuid.UUID) (JWT, error) {
	args := m.Called(subject)
	return args.Get(0).(JWT), args.Error(1)
}
