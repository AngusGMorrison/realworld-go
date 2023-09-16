package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) Register(ctx context.Context, req *RegistrationRequest) (*User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) Authenticate(ctx context.Context, req *AuthRequest) (*User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) UpdateUser(ctx context.Context, req *UpdateRequest) (*User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*User), args.Error(1)
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email EmailAddress) (*User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) CreateUser(ctx context.Context, req *RegistrationRequest) (*User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) UpdateUser(ctx context.Context, req *UpdateRequest) (*User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*User), args.Error(1)
}
