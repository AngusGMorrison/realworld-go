package postgres

import (
	"context"

	"github.com/angusgmorrison/realworld-go/internal/outbound/postgres/sqlc"
	"github.com/stretchr/testify/mock"
)

type mockQueries struct {
	mock.Mock
}

func (m *mockQueries) CreateUser(ctx context.Context, params sqlc.CreateUserParams) (sqlc.User, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *mockQueries) DeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockQueries) GetUserByEmail(ctx context.Context, email string) (sqlc.GetUserByEmailRow, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(sqlc.GetUserByEmailRow), args.Error(1)
}

func (m *mockQueries) GetUserById(ctx context.Context, id string) (sqlc.GetUserByIdRow, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(sqlc.GetUserByIdRow), args.Error(1)
}

func (m *mockQueries) UpdateUser(ctx context.Context, params sqlc.UpdateUserParams) (sqlc.User, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(sqlc.User), args.Error(1)
}

func (m *mockQueries) UserExists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}
