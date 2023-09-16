package testutil

import (
	"context"

	"github.com/angusgmorrison/realworld-go/internal/domain/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, req *user.RegistrationRequest) (*user.User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) Authenticate(ctx context.Context, req *user.AuthRequest) (*user.User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, id uuid.UUID) (*user.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, req *user.UpdateRequest) (*user.User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*user.User), args.Error(1)
}
