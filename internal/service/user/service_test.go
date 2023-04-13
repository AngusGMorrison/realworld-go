package user

import (
	"context"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

const (
	email    = "test@test.com"
	username = "testuser"
	password = "password"
	bio      = "test bio"
	imageURL = "https://test.com/image.png"
)

func Test_service_Register(t *testing.T) {
	t.Parallel()

	t.Run("when request validation fails it returns validation errors", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name string
			req  *RegistrationRequest
		}{
			{
				name: "missing email",
				req: &RegistrationRequest{
					Email:    "",
					Username: username,
					RequiredValidatingPassword: RequiredValidatingPassword{
						Password: password,
					},
				},
			},
			{
				name: "invalid email",
				req: &RegistrationRequest{
					Email:    "invalid",
					Username: username,
					RequiredValidatingPassword: RequiredValidatingPassword{
						Password: password,
					},
				},
			},
			{
				name: "missing password",
				req: &RegistrationRequest{
					Email:    email,
					Username: username,
					RequiredValidatingPassword: RequiredValidatingPassword{
						Password: "1",
					},
				},
			},
			{
				name: "password too short",
				req: &RegistrationRequest{
					Email:    email,
					Username: username,
					RequiredValidatingPassword: RequiredValidatingPassword{
						Password: "1",
					},
				},
			},
			{
				name: "password too long",
				req: &RegistrationRequest{
					Email:    email,
					Username: username,
					RequiredValidatingPassword: RequiredValidatingPassword{
						Password: strings.Repeat("1", 73),
					},
				},
			},
			{
				name: "missing username",
				req: &RegistrationRequest{
					Email:    email,
					Username: "",
					RequiredValidatingPassword: RequiredValidatingPassword{
						Password: password,
					},
				},
			},
			{
				name: "username too long",
				req: &RegistrationRequest{
					Email:    email,
					Username: strings.Repeat("a", 33),
					RequiredValidatingPassword: RequiredValidatingPassword{
						Password: password,
					},
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				s := NewService(nil, nil, 0)

				usr, err := s.Register(context.Background(), tc.req)

				var errs validator.ValidationErrors
				assert.ErrorAs(t, err, &errs)
				assert.Nil(t, usr)
			})
		}
	})

	t.Run("when the repository returns an error it returns the error", func(t *testing.T) {
		t.Parallel()

		repo := mockRepository{}
		repo.On("CreateUser", mock.AnythingOfType("context.Context"), ).Return(nil, errors.New("error"))
		s := NewService(nil, nil, 0)

		usr, err := s.Register(context.Background(), &RegistrationRequest{
			Email:    email,
			Username: username,
			RequiredValidatingPassword: RequiredValidatingPassword{
				Password: password,
			},
		})

		assert.Error(t, err)
		assert.Nil(t, usr)
	}
}

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*User), args.Error(1)
}

func (m *mockRepository) GetUserByEmail(ctx context.Context, email EmailAddress) (*User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*User), args.Error(1)
}

func (m *mockRepository) CreateUser(ctx context.Context, req *RegistrationRequest) (*User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*User), args.Error(1)
}

func (m *mockRepository) UpdateUser(ctx context.Context, req *UpdateRequest) (*User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*User), args.Error(1)
}
