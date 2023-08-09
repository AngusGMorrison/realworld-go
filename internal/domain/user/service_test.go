package user

// import (
// 	"context"
// 	"crypto/rand"
// 	"crypto/rsa"
// 	"errors"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/angusgmorrison/realworld/pkg/primitive"
// 	"github.com/go-playground/validator/v10"
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/require"
// )

// const (
// 	email    = "test@test.com"
// 	username = "testuser"
// 	password = "password"
// 	bio      = "test bio"
// 	imageURL = "https://test.com/image.png"
// )

// func Test_service_Register(t *testing.T) {
// 	t.Parallel()

// 	t.Run("when request validation fails it returns validation errors", func(t *testing.T) {
// 		t.Parallel()

// 		testCases := []struct {
// 			name string
// 			req  *RegistrationRequest
// 		}{
// 			{
// 				name: "missing email",
// 				req: &RegistrationRequest{
// 					Email:    "",
// 					username: username,
// 					RequiredValidatingPassword: RequiredValidatingPassword{
// 						Password: password,
// 					},
// 				},
// 			},
// 			{
// 				name: "invalid email",
// 				req: &RegistrationRequest{
// 					Email:    "invalid",
// 					username: username,
// 					RequiredValidatingPassword: RequiredValidatingPassword{
// 						Password: password,
// 					},
// 				},
// 			},
// 			{
// 				name: "missing password",
// 				req: &RegistrationRequest{
// 					Email:    email,
// 					username: username,
// 					RequiredValidatingPassword: RequiredValidatingPassword{
// 						Password: "1",
// 					},
// 				},
// 			},
// 			{
// 				name: "password too short",
// 				req: &RegistrationRequest{
// 					Email:    email,
// 					username: username,
// 					RequiredValidatingPassword: RequiredValidatingPassword{
// 						Password: "1",
// 					},
// 				},
// 			},
// 			{
// 				name: "password too long",
// 				req: &RegistrationRequest{
// 					Email:    email,
// 					username: username,
// 					RequiredValidatingPassword: RequiredValidatingPassword{
// 						Password: primitive.SensitiveString(strings.Repeat("1", 73)),
// 					},
// 				},
// 			},
// 			{
// 				name: "missing username",
// 				req: &RegistrationRequest{
// 					Email:    email,
// 					username: "",
// 					RequiredValidatingPassword: RequiredValidatingPassword{
// 						Password: password,
// 					},
// 				},
// 			},
// 			{
// 				name: "username too long",
// 				req: &RegistrationRequest{
// 					Email:    email,
// 					username: strings.Repeat("a", 33),
// 					RequiredValidatingPassword: RequiredValidatingPassword{
// 						Password: password,
// 					},
// 				},
// 			},
// 		}

// 		for _, tc := range testCases {
// 			tc := tc

// 			t.Run(tc.name, func(t *testing.T) {
// 				t.Parallel()

// 				s := NewService(nil, nil, 0)

// 				usr, err := s.Register(context.Background(), tc.req)

// 				var errs validator.ValidationErrors
// 				assert.ErrorAs(t, err, &errs)
// 				assert.Nil(t, usr)
// 			})
// 		}
// 	})

// 	t.Run("when the outbound returns an error it returns the error", func(t *testing.T) {
// 		t.Parallel()

// 		repo := &mockRepository{}
// 		req := &RegistrationRequest{
// 			Email:    email,
// 			username: username,
// 			RequiredValidatingPassword: RequiredValidatingPassword{
// 				Password: password,
// 			},
// 		}

// 		passwordHash, err := req.HashPassword()
// 		require.NoError(t, err, "hash password")

// 		expectedUser := &User{
// 			Email:        email,
// 			username:     username,
// 			passwordHash: passwordHash,
// 		}

// 		userMatcher := newUserMatcher(expectedUser, req.Password)
// 		repo.On("CreateUser", mock.AnythingOfType("*context.emptyCtx"), mock.MatchedBy(userMatcher)).Return((*User)(nil), errors.New("error"))
// 		s := NewService(repo, nil, 0)

// 		usr, err := s.Register(context.Background(), req)

// 		assert.Error(t, err)
// 		assert.Nil(t, usr)
// 	})

// 	t.Run("when the outbound call succeeds it returns an authenticated user", func(t *testing.T) {
// 		t.Parallel()

// 		key, err := rsa.GenerateKey(rand.Reader, 2048)
// 		require.NoError(t, err, "generate RSA key")

// 		req := &RegistrationRequest{
// 			Email:    email,
// 			username: username,
// 			RequiredValidatingPassword: RequiredValidatingPassword{
// 				Password: password,
// 			},
// 		}

// 		passwordHash, err := req.HashPassword()
// 		require.NoError(t, err, "hash password")

// 		expectedRepoUser := &User{
// 			Email:        email,
// 			username:     username,
// 			passwordHash: passwordHash,
// 		}

// 		repoUserMatcher := newUserMatcher(expectedRepoUser, req.Password)

// 		ID := uuid.New()
// 		jwtTTL := 1 * time.Hour
// 		jwt, err := newJWT(key, jwtTTL, ID.String())
// 		require.NoError(t, err, "generate JWT")

// 		expectedAuthUser := &AuthenticatedUser{
// 			User: &User{
// 				ID:           ID,
// 				Email:        email,
// 				username:     username,
// 				passwordHash: passwordHash,
// 				Bio:          bio,
// 				ImageURL:     imageURL,
// 			},
// 			Token: jwt,
// 		}

// 		repo := &mockRepository{}
// 		repo.On("CreateUser", mock.AnythingOfType("*context.emptyCtx"), mock.MatchedBy(repoUserMatcher)).Return(expectedAuthUser.User, nil)

// 		s := NewService(repo, key, jwtTTL)

// 		gotAuthUser, err := s.Register(context.Background(), req)

// 		require.NoError(t, err, "call domain")

// 		authUsersEqual := expectedAuthUser.Equals(gotAuthUser, &key.PublicKey)
// 		assert.Truef(t, authUsersEqual, "expected AutheticatedUser %#v, got %#v", expectedAuthUser, gotAuthUser)
// 	})
// }

// func Test_service_Authenticate(t *testing.T) {
// 	t.Parallel()

// 	t.Run("when request validation fails it returns validation errors", func(t *testing.T) {
// 		t.Parallel()

// 		testCases := []struct {
// 			name string
// 			req  *AuthRequest
// 		}{
// 			{
// 				name: "missing email",
// 				req: &AuthRequest{
// 					Email:    "",
// 					Password: password,
// 				},
// 			},
// 			{
// 				name: "invalid email",
// 				req: &AuthRequest{
// 					Email:    "invalid",
// 					Password: password,
// 				},
// 			},
// 			{
// 				name: "missing password",
// 				req: &AuthRequest{
// 					Email:    email,
// 					Password: "",
// 				},
// 			},
// 		}

// 		for _, tc := range testCases {
// 			tc := tc

// 			t.Run(tc.name, func(t *testing.T) {
// 				t.Parallel()

// 				s := NewService(nil, nil, 0)

// 				usr, err := s.Authenticate(context.Background(), tc.req)

// 				var errs validator.ValidationErrors
// 				assert.ErrorAs(t, err, &errs)
// 				assert.Nil(t, usr)
// 			})
// 		}
// 	})

// 	t.Run("when the outbound returns an error", func(t *testing.T) {
// 		t.Parallel()

// 		testCases := []struct {
// 			name string
// 			err  error
// 		}{
// 			{
// 				name: "when the error is ErrUserNotFound it returns an AuthError",
// 				err:  ErrUserNotFound,
// 			},
// 			{
// 				name: "when the error is not ErrUserNotFound it returns the error",
// 				err:  errors.New("error"),
// 			},
// 		}

// 		for _, tc := range testCases {
// 			tc := tc

// 			t.Run(tc.name, func(t *testing.T) {
// 				t.Parallel()

// 				repo := &mockRepository{}
// 				expectedAuthReq := &AuthRequest{
// 					Email:    email,
// 					Password: password,
// 				}
// 				repo.On("GetUserByEmail", mock.AnythingOfType("*context.emptyCtx"), expectedAuthReq.Email).Return((*User)(nil), tc.err)
// 				s := NewService(repo, nil, 0)

// 				usr, err := s.Authenticate(context.Background(), expectedAuthReq)

// 				assert.Error(t, err)
// 				assert.Nil(t, usr)
// 			})
// 		}
// 	})

// 	t.Run("when the request password doesn't match the saved password it returns an AuthError", func(t *testing.T) {
// 		t.Parallel()

// 		repo := &mockRepository{}
// 		expectedAuthReq := &AuthRequest{
// 			Email:    email,
// 			Password: "mismatch",
// 		}
// 		hashedPassword, err := bcryptHash(password)
// 		require.NoError(t, err, "hash password")
// 		savedUser := &User{
// 			ID:           uuid.New(),
// 			Email:        email,
// 			username:     username,
// 			passwordHash: hashedPassword,
// 		}

// 		repo.On("GetUserByEmail", mock.AnythingOfType("*context.emptyCtx"), expectedAuthReq.Email).Return(savedUser, nil)
// 		s := NewService(repo, nil, 0)

// 		usr, err := s.Authenticate(context.Background(), expectedAuthReq)

// 		var authErr *AuthError
// 		require.ErrorAs(t, err, &authErr)
// 		assert.ErrorIs(t, authErr.Cause, ErrPasswordMismatch)
// 		assert.Nil(t, usr)
// 	})

// 	t.Run("when the request succeeds it returns the authenticated user", func(t *testing.T) {
// 		t.Parallel()

// 		key, err := rsa.GenerateKey(rand.Reader, 2048)
// 		require.NoError(t, err, "generate RSA key")

// 		req := &AuthRequest{
// 			Email:    email,
// 			Password: password,
// 		}

// 		passwordHash, err := bcryptHash(req.Password)
// 		require.NoError(t, err, "hash password")

// 		ID := uuid.New()
// 		jwtTTL := 1 * time.Hour
// 		jwt, err := newJWT(key, jwtTTL, ID.String())
// 		require.NoError(t, err, "generate JWT")
// 		expectedAuthUser := &AuthenticatedUser{
// 			User: &User{
// 				ID:           ID,
// 				Email:        email,
// 				username:     username,
// 				passwordHash: passwordHash,
// 				Bio:          bio,
// 				ImageURL:     imageURL,
// 			},
// 			Token: jwt,
// 		}

// 		repo := &mockRepository{}
// 		repo.On("GetUserByEmail", mock.AnythingOfType("*context.emptyCtx"), req.Email).Return(expectedAuthUser.User, nil)

// 		s := NewService(repo, key, jwtTTL)

// 		gotAuthUser, err := s.Authenticate(context.Background(), req)

// 		require.NoError(t, err, "call domain")

// 		authUsersEqual := expectedAuthUser.Equals(gotAuthUser, &key.PublicKey)
// 		assert.Truef(t, authUsersEqual, "expected AutheticatedUser %#v, got %#v", expectedAuthUser, gotAuthUser)
// 	})
// }

// func Test_service_GetUser(t *testing.T) {
// 	t.Parallel()

// 	t.Run("when the outbound returns an error it returns the error", func(t *testing.T) {
// 		t.Parallel()

// 		ID := uuid.New()
// 		repo := &mockRepository{}
// 		repo.On("GetUserByID", mock.AnythingOfType("*context.emptyCtx"), ID).Return((*User)(nil), errors.New("error"))

// 		s := NewService(repo, nil, 0)

// 		usr, err := s.GetUser(context.Background(), ID)

// 		assert.Error(t, err)
// 		assert.Nil(t, usr)
// 	})

// 	t.Run("when the outbound call succeeds it returns the user", func(t *testing.T) {
// 		t.Parallel()

// 		ID := uuid.New()
// 		expectedUser := &User{
// 			ID:       ID,
// 			Email:    email,
// 			username: username,
// 			Bio:      bio,
// 			ImageURL: imageURL,
// 		}

// 		repo := &mockRepository{}
// 		repo.On("GetUserByID", mock.AnythingOfType("*context.emptyCtx"), ID).Return(expectedUser, nil)

// 		s := NewService(repo, nil, 0)

// 		usr, err := s.GetUser(context.Background(), ID)

// 		require.NoError(t, err, "call domain")
// 		assert.Equal(t, expectedUser, usr)
// 	})
// }

// func Test_service_UpdateUser(t *testing.T) {
// 	t.Parallel()

// 	t.Run("when request validation fails it returns validation errors", func(t *testing.T) {
// 		t.Parallel()

// 		var (
// 			email         = primitive.EmailAddress("invalid")
// 			imageURL      = "invalid"
// 			shortPassword = primitive.SensitiveString("a")
// 			longPassword  = primitive.SensitiveString(strings.Repeat("a", 73))
// 		)

// 		testCases := []struct {
// 			name string
// 			req  *UpdateRequest
// 		}{
// 			{
// 				name: "missing user ID",
// 				req:  &UpdateRequest{},
// 			},
// 			{
// 				name: "invalid email",
// 				req: &UpdateRequest{
// 					UserID: uuid.New(),
// 					Email:  &email,
// 				},
// 			},
// 			{
// 				name: "invalid image URL",
// 				req: &UpdateRequest{
// 					UserID:   uuid.New(),
// 					ImageURL: &imageURL,
// 				},
// 			},
// 			{
// 				name: "password too short",
// 				req: &UpdateRequest{
// 					UserID: uuid.New(),
// 					OptionalValidatingPassword: OptionalValidatingPassword{
// 						Password: &shortPassword,
// 					},
// 				},
// 			},
// 			{
// 				name: "password too long",
// 				req: &UpdateRequest{
// 					UserID: uuid.New(),
// 					OptionalValidatingPassword: OptionalValidatingPassword{
// 						Password: &longPassword,
// 					},
// 				},
// 			},
// 		}

// 		for _, tc := range testCases {
// 			tc := tc

// 			t.Run(tc.name, func(t *testing.T) {
// 				t.Parallel()

// 				s := NewService(nil, nil, 0)

// 				usr, err := s.UpdateUser(context.Background(), tc.req)

// 				var validationErrs validator.ValidationErrors
// 				require.ErrorAs(t, err, &validationErrs)
// 				assert.Nil(t, usr)
// 			})
// 		}
// 	})

// 	t.Run("when the outbound returns an error it returns the error", func(t *testing.T) {
// 		t.Parallel()

// 		ID := uuid.New()
// 		updateReq := &UpdateRequest{
// 			UserID: ID,
// 		}
// 		repo := &mockRepository{}
// 		repo.On("UpdateUser", mock.AnythingOfType("*context.emptyCtx"), updateReq).Return((*User)(nil), errors.New("error"))

// 		s := NewService(repo, nil, 0)

// 		usr, err := s.UpdateUser(context.Background(), updateReq)

// 		assert.Error(t, err)
// 		assert.Nil(t, usr)
// 	})

// 	t.Run("when the outbound call succeeds it returns the updated user", func(t *testing.T) {
// 		t.Parallel()

// 		ID := uuid.New()
// 		expectedUser := &User{
// 			ID:       ID,
// 			Email:    email,
// 			username: username,
// 			Bio:      bio,
// 			ImageURL: imageURL,
// 		}
// 		updateReq := &UpdateRequest{
// 			UserID: ID,
// 		}
// 		repo := &mockRepository{}
// 		repo.On("UpdateUser", mock.AnythingOfType("*context.emptyCtx"), updateReq).Return(expectedUser, nil)

// 		s := NewService(repo, nil, 0)

// 		usr, err := s.UpdateUser(context.Background(), updateReq)

// 		require.NoError(t, err, "call domain")
// 		assert.Equal(t, expectedUser, usr)
// 	})
// }

// type mockRepository struct {
// 	mock.Mock
// }

// func (m *mockRepository) GetUserByID(ctx context.Context, ID uuid.UUID) (*User, error) {
// 	args := m.Called(ctx, ID)
// 	return args.Get(0).(*User), args.Error(1)
// }

// func (m *mockRepository) GetUserByEmail(ctx context.Context, email primitive.EmailAddress) (*User, error) {
// 	args := m.Called(ctx, email)
// 	return args.Get(0).(*User), args.Error(1)
// }

// func (m *mockRepository) CreateUser(ctx context.Context, user *User) (*User, error) {
// 	args := m.Called(ctx, user)
// 	return args.Get(0).(*User), args.Error(1)
// }

// func (m *mockRepository) UpdateUser(ctx context.Context, req *UpdateRequest) (*User, error) {
// 	args := m.Called(ctx, req)
// 	return args.Get(0).(*User), args.Error(1)
// }

// // bcrypt returns a different hash each time, so we need a custom matcher that
// // avoids direct hash comparison.
// func newUserMatcher(expected *User, password primitive.SensitiveString) func(arg any) bool {
// 	return func(arg any) bool {
// 		user, ok := arg.(*User)
// 		if !ok {
// 			return false
// 		}

// 		return expected.Equals(user) && user.HasPassword(password)
// 	}
// }
