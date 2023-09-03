package v0

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/angusgmorrison/logfusc"
	"github.com/angusgmorrison/realworld-go/internal/domain/user"
	"github.com/angusgmorrison/realworld-go/internal/testutil"
	"github.com/angusgmorrison/realworld-go/pkg/option"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"strings"
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

	t.Run("errors", func(t *testing.T) {
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
						mock.MatchedBy(testutil.NewUserRegistrationRequestMatcher(
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
						mock.MatchedBy(testutil.NewUserRegistrationRequestMatcher(
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
			mock.MatchedBy(testutil.NewUserRegistrationRequestMatcher(
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

	t.Run("errors", func(t *testing.T) {
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

func Test_UsersHandler_GetCurrent(t *testing.T) {
	t.Parallel()

	userID := uuid.New()

	t.Run("panics", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name         string
			setupContext func(c *fiber.Ctx) error
			setupMocks   func(service *testutil.MockUserService)
		}{
			{
				name: "current user ID missing from context",
				setupContext: func(c *fiber.Ctx) error {
					return c.Next()
				},
				setupMocks: func(service *testutil.MockUserService) {},
			},
			{
				name: "current JWT missing from context",
				setupContext: func(c *fiber.Ctx) error {
					c.Locals(userIDKey, userID)
					return c.Next()
				},
				setupMocks: func(service *testutil.MockUserService) {
					service.On(
						"GetUser",
						mock.AnythingOfType("*fasthttp.RequestCtx"),
						userID,
					).Return(user.RandomUser(t), nil)
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				service := &testutil.MockUserService{}
				handler := &UsersHandler{
					service: service,
				}
				app := fiber.New()
				assertPanics := func(c *fiber.Ctx) error {
					assert.Panics(t, func() {
						_ = c.Next()
					})
					return nil
				}
				app.Get("/", tc.setupContext, assertPanics, handler.GetCurrent)

				req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
				require.NoError(t, err)

				tc.setupMocks(service)

				_, _ = app.Test(req)

				service.AssertExpectations(t)
			})
		}
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name        string
			setupMocks  func(service *testutil.MockUserService)
			assertError func(t *testing.T, err error)
			assertMocks func(t *testing.T, service *testutil.MockUserService)
		}{
			{
				name: "service error",
				setupMocks: func(service *testutil.MockUserService) {
					service.On(
						"GetUser",
						mock.AnythingOfType("*fasthttp.RequestCtx"),
						userID,
					).Return((*user.User)(nil), assert.AnError)
				},
				assertError: func(t *testing.T, err error) {
					t.Helper()
					assert.ErrorIs(t, err, assert.AnError)
				},
				assertMocks: func(t *testing.T, service *testutil.MockUserService) {
					t.Helper()
					service.AssertExpectations(t)
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				service := &testutil.MockUserService{}
				handler := &UsersHandler{
					service: service,
				}

				app := fiber.New(fiber.Config{
					ErrorHandler: func(ctx *fiber.Ctx, err error) error {
						tc.assertError(t, err)
						return nil
					},
				})
				setUserIDOnContext := func(c *fiber.Ctx) error {
					c.Locals(userIDKey, userID)
					return c.Next()
				}
				app.Get("/", setUserIDOnContext, handler.GetCurrent)

				req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
				require.NoError(t, err)

				tc.setupMocks(service)

				_, err = app.Test(req)
				require.NoError(t, err)

				tc.assertMocks(t, service)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service := &testutil.MockUserService{}
		handler := &UsersHandler{
			service: service,
		}
		requestJWT := &jwt.Token{Raw: "abc"}

		app := fiber.New()
		setUserIDAndJWTOnContext := func(c *fiber.Ctx) error {
			c.Locals(userIDKey, userID)
			c.Locals(requestJWTKey, requestJWT)
			return c.Next()
		}
		app.Get("/", setUserIDAndJWTOnContext, handler.GetCurrent)

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
		require.NoError(t, err)

		wantUser := user.RandomUser(t)
		wantStatusCode := fiber.StatusOK
		wantResponseBody := fmt.Sprintf(`{"user": {"token": %q, "email": %q, "username": %q, "bio": %q, "image": %q}}`,
			requestJWT.Raw, wantUser.Email(), wantUser.Username(), wantUser.Bio().UnwrapOrZero(), wantUser.ImageURL().UnwrapOrZero())

		service.On(
			"GetUser",
			mock.AnythingOfType("*fasthttp.RequestCtx"),
			userID,
		).Return(wantUser, nil)

		res, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, res.StatusCode, wantStatusCode)

		gotResponseBodyBytes, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(t, wantResponseBody, string(gotResponseBodyBytes))

		service.AssertExpectations(t)
	})
}

func Test_UsersHandler_UpdateCurrent(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	emailOption := user.RandomOptionFromInstance(user.RandomEmailAddressCandidate())
	passwordOption := user.RandomOptionFromInstance(user.RandomPasswordCandidate())
	bioOption := user.RandomOptionFromInstance(string(user.RandomBio()))
	urlOption := user.RandomOptionFromInstance(user.RandomURLCandidate())

	t.Run("panics", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name         string
			setupContext func(c *fiber.Ctx) error
			setupMocks   func(service *testutil.MockUserService)
			assertMocks  func(t *testing.T, service *testutil.MockUserService)
		}{
			{
				name: "current user ID missing from context",
				setupContext: func(c *fiber.Ctx) error {
					return c.Next()
				},
				setupMocks: func(service *testutil.MockUserService) {},
				assertMocks: func(t *testing.T, service *testutil.MockUserService) {
					t.Helper()
					service.AssertNotCalled(t, "UpdateUser")
				},
			},
			{
				name: "current JWT missing from context",
				setupContext: func(c *fiber.Ctx) error {
					c.Locals(userIDKey, userID)
					return c.Next()
				},
				setupMocks: func(service *testutil.MockUserService) {
					wantUpdateReq, err := user.ParseUpdateRequest(
						userID,
						emailOption,
						passwordOption,
						bioOption,
						urlOption,
					)
					require.NoError(t, err)

					service.On(
						"UpdateUser",
						mock.AnythingOfType("*fasthttp.RequestCtx"),
						mock.MatchedBy(testutil.NewUserUpdateRequestMatcher(t, wantUpdateReq, passwordOption)),
					).Return(user.RandomUser(t), nil)
				},
				assertMocks: func(t *testing.T, service *testutil.MockUserService) {
					t.Helper()
					service.AssertExpectations(t)
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				service := &testutil.MockUserService{}
				handler := &UsersHandler{
					service: service,
				}
				app := fiber.New()
				assertPanics := func(c *fiber.Ctx) error {
					assert.Panics(t, func() {
						_ = c.Next()
					})
					return nil
				}
				app.Put("/", tc.setupContext, assertPanics, handler.UpdateCurrent)

				requestBody := updateRequestBodyFromOptions(emailOption, bioOption, urlOption, passwordOption)
				req, err := http.NewRequestWithContext(
					context.Background(),
					http.MethodPut,
					"/",
					bytes.NewBufferString(requestBody),
				)
				require.NoError(t, err)
				req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

				tc.setupMocks(service)

				_, err = app.Test(req)
				require.NoError(t, err)

				tc.assertMocks(t, service)
			})
		}
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name        string
			requestBody string
			setupMocks  func(service *testutil.MockUserService)
			assertError func(t *testing.T, err error)
			assertMocks func(t *testing.T, service *testutil.MockUserService)
		}{
			{
				name:        "JSON syntax error",
				requestBody: `{`,
				setupMocks:  func(service *testutil.MockUserService) {},
				assertError: func(t *testing.T, err error) {
					t.Helper()
					var syntaxErr *json.SyntaxError
					assert.ErrorAs(t, err, &syntaxErr)
				},
				assertMocks: func(t *testing.T, service *testutil.MockUserService) {
					t.Helper()
					service.AssertNotCalled(t, "UpdateUser")
				},
			},
			{
				name: "parse domain model error",
				requestBody: updateRequestBodyFromOptions(
					option.Some("invalid email"),
					bioOption,
					urlOption,
					passwordOption,
				),
				setupMocks: func(service *testutil.MockUserService) {},
				assertError: func(t *testing.T, err error) {
					t.Helper()
					var validationErrs user.ValidationErrors
					assert.ErrorAs(t, err, &validationErrs)
				},
				assertMocks: func(t *testing.T, service *testutil.MockUserService) {
					t.Helper()
					service.AssertNotCalled(t, "UpdateUser")
				},
			},
			{
				name:        "service error",
				requestBody: updateRequestBodyFromOptions(emailOption, bioOption, urlOption, passwordOption),
				setupMocks: func(service *testutil.MockUserService) {
					wantUpdateRequest, err := user.ParseUpdateRequest(
						userID,
						emailOption,
						passwordOption,
						bioOption,
						urlOption,
					)
					require.NoError(t, err)

					service.On(
						"UpdateUser",
						mock.AnythingOfType("*fasthttp.RequestCtx"),
						mock.MatchedBy(testutil.NewUserUpdateRequestMatcher(t, wantUpdateRequest, passwordOption)),
					).Return((*user.User)(nil), assert.AnError)
				},
				assertError: func(t *testing.T, err error) {
					t.Helper()
					assert.ErrorIs(t, err, assert.AnError)
				},
				assertMocks: func(t *testing.T, service *testutil.MockUserService) {
					t.Helper()
					service.AssertExpectations(t)
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				service := &testutil.MockUserService{}
				handler := &UsersHandler{
					service: service,
				}

				app := fiber.New(fiber.Config{
					ErrorHandler: func(ctx *fiber.Ctx, err error) error {
						tc.assertError(t, err)
						return nil
					},
				})
				setUserIDAndJWTOnContext := func(c *fiber.Ctx) error {
					c.Locals(userIDKey, userID)
					c.Locals(requestJWTKey, &jwt.Token{Raw: "abc"})
					return c.Next()
				}
				app.Put("/", setUserIDAndJWTOnContext, handler.UpdateCurrent)

				req, err := http.NewRequestWithContext(
					context.Background(),
					http.MethodPut,
					"/",
					bytes.NewBufferString(tc.requestBody),
				)
				require.NoError(t, err)
				req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

				tc.setupMocks(service)

				_, err = app.Test(req)
				require.NoError(t, err)

				tc.assertMocks(t, service)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		service := &testutil.MockUserService{}
		handler := &UsersHandler{
			service: service,
		}

		app := fiber.New()
		requestJWT := &jwt.Token{Raw: "abc"}
		setUserIDAndJWTOnContext := func(c *fiber.Ctx) error {
			c.Locals(userIDKey, userID)
			c.Locals(requestJWTKey, requestJWT)
			return c.Next()
		}
		app.Put("/", setUserIDAndJWTOnContext, handler.UpdateCurrent)

		reqBody := updateRequestBodyFromOptions(emailOption, bioOption, urlOption, passwordOption)
		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodPut,
			"/",
			bytes.NewBufferString(reqBody),
		)
		require.NoError(t, err)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		wantUser := user.RandomUser(t)
		wantUpdateRequest, err := user.ParseUpdateRequest(
			userID,
			emailOption,
			passwordOption,
			bioOption,
			urlOption,
		)
		require.NoError(t, err)
		wantStatusCode := fiber.StatusOK
		wantResponseBody := fmt.Sprintf(`{"user": {"token": %q, "email": %q, "username": %q, "bio": %q, "image": %q}}`,
			requestJWT.Raw, wantUser.Email(), wantUser.Username(), wantUser.Bio().UnwrapOrZero(), wantUser.ImageURL().UnwrapOrZero())

		service.On(
			"UpdateUser",
			mock.AnythingOfType("*fasthttp.RequestCtx"),
			mock.MatchedBy(testutil.NewUserUpdateRequestMatcher(t, wantUpdateRequest, passwordOption)),
		).Return(wantUser, nil)

		res, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, res.StatusCode, wantStatusCode)

		gotResponseBodyBytes, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.JSONEq(t, wantResponseBody, string(gotResponseBodyBytes))

		service.AssertExpectations(t)
	})
}

func Test_UsersErrorHandler(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input error
		want  error
	}{
		{
			name:  "*json.SyntaxError",
			input: &json.SyntaxError{},
			want:  NewBadRequestError(&json.SyntaxError{}),
		},
		{
			name: "*user.AuthError",
			input: &user.AuthError{
				Cause: assert.AnError,
			},
			want: NewUnauthorizedError("invalid credentials", assert.AnError),
		},
		{
			name: "*user.NotFoundError",
			input: &user.NotFoundError{
				IDType:  user.UUIDFieldType,
				IDValue: uuid.Nil.String(),
			},
			want: NewNotFoundError(
				"user",
				fmt.Sprintf("user with %s %q not found", user.UUIDFieldType, uuid.Nil.String()),
			),
		},
		{
			name: "user.ValidationErrors",
			input: user.ValidationErrors{
				{FieldType: user.EmailFieldType, Message: "invalid"},
				{FieldType: user.PasswordFieldType, Message: "invalid"},
				{FieldType: user.UsernameFieldType, Message: "invalid"},
				{FieldType: user.URLFieldType, Message: "invalid"},
				{FieldType: user.URLFieldType, Message: "another URL error"},
			},
			want: NewUnprocessableEntityError(
				userFacingValidationErrorMessages{
					"email":    []string{"invalid"},
					"password": []string{"invalid"},
					"username": []string{"invalid"},
					"image":    []string{"invalid", "another URL error"},
				},
			),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			errorSource := func(c *fiber.Ctx) error {
				return tc.input
			}
			errorAsserter := func(c *fiber.Ctx) error {
				err := c.Next()
				assert.Equal(t, tc.want, err)
				return nil
			}
			app.Get("/", errorAsserter, UsersErrorHandler, errorSource)

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			require.NoError(t, err)

			_, err = app.Test(req)
			require.NoError(t, err)
		})
	}
}

func updateRequestBodyFromOptions(
	emailOption, bioOption, urlOption option.Option[string],
	passwordOption option.Option[logfusc.Secret[string]],
) string {
	var requestFields []string
	if emailOption.IsSome() {
		requestFields = append(requestFields, fmt.Sprintf(`"email": %q`, emailOption.UnwrapOrZero()))
	}
	if passwordOption.IsSome() {
		requestFields = append(requestFields, fmt.Sprintf(`"password": %q`, passwordOption.UnwrapOrZero().Expose()))
	}
	if bioOption.IsSome() {
		requestFields = append(requestFields, fmt.Sprintf(`"bio": %q`, bioOption.UnwrapOrZero()))
	}
	if urlOption.IsSome() {
		requestFields = append(requestFields, fmt.Sprintf(`"image": %q`, urlOption.UnwrapOrZero()))
	}

	return fmt.Sprintf(`{"user": {%s}}`, strings.Join(requestFields, ", "))
}

type mockJWTProvider struct {
	mock.Mock
}

func (m *mockJWTProvider) TokenFor(subject uuid.UUID) (JWT, error) {
	args := m.Called(subject)
	return args.Get(0).(JWT), args.Error(1)
}
