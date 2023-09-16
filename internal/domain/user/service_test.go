package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var anyError = errors.New("any error")

func Test_service_Register(t *testing.T) {
	t.Parallel()

	req := RandomRegistrationRequest(t)

	testCases := []struct {
		name     string
		wantUser *User
		wantErr  error
	}{
		{
			name:     "repo call succeeds",
			wantUser: &User{},
			wantErr:  nil,
		},
		{
			name:     "repo returns any error",
			wantUser: nil,
			wantErr:  anyError,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &MockRepository{}
			svc := NewService(repo)

			repo.On("CreateUser", context.Background(), req).Return(tc.wantUser, tc.wantErr)

			gotUser, gotErr := svc.Register(context.Background(), req)

			assert.Equal(t, tc.wantUser, gotUser)
			assert.ErrorIs(t, gotErr, tc.wantErr)
			repo.AssertExpectations(t)
		})
	}
}

func Test_service_Authenticate(t *testing.T) {
	t.Parallel()

	req := RandomAuthRequest(t)

	testCases := []struct {
		name               string
		repoUser           *User
		repoErr            error
		passwordComparator passwordComparator
		wantUser           *User
		wantErr            error
	}{
		{
			name:     "success",
			repoUser: &User{},
			repoErr:  nil,
			passwordComparator: func(hash PasswordHash, candidate string) error {
				return nil
			},
			wantUser: &User{},
			wantErr:  nil,
		},
		{
			name:               "repo returns NotFoundError",
			repoUser:           nil,
			repoErr:            &NotFoundError{},
			passwordComparator: nil,
			wantUser:           nil,
			wantErr: &AuthError{
				Cause: &NotFoundError{},
			},
		},
		{
			name:               "repo returns any other error",
			repoUser:           nil,
			repoErr:            anyError,
			passwordComparator: nil,
			wantUser:           nil,
			wantErr:            anyError,
		},
		{
			name:     "passwordComparator returns an error",
			repoUser: &User{},
			repoErr:  nil,
			passwordComparator: func(hash PasswordHash, candidate string) error {
				return &AuthError{}
			},
			wantUser: nil,
			wantErr:  &AuthError{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &MockRepository{}
			svc := &service{
				repo:               repo,
				passwordComparator: tc.passwordComparator,
			}

			repo.On("GetUserByEmail", context.Background(), req.Email()).Return(tc.wantUser, tc.wantErr)

			gotUser, gotErr := svc.Authenticate(context.Background(), req)

			assert.Equal(t, tc.wantUser, gotUser)
			assert.ErrorIs(t, gotErr, tc.wantErr)
			repo.AssertExpectations(t)
		})
	}
}

func Test_service_GetUser(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		wantUser *User
		wantErr  error
	}{
		{
			name:     "repo call succeeds",
			wantUser: &User{},
			wantErr:  nil,
		},
		{
			name:     "repo returns any error",
			wantUser: nil,
			wantErr:  anyError,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &MockRepository{}
			svc := NewService(repo)
			userID := uuid.New()

			repo.On("GetUserByID", context.Background(), userID).Return(tc.wantUser, tc.wantErr)

			gotUser, gotErr := svc.GetUser(context.Background(), userID)

			assert.Equal(t, tc.wantUser, gotUser)
			assert.ErrorIs(t, gotErr, tc.wantErr)
			repo.AssertExpectations(t)
		})
	}
}

func Test_service_UpdateUser(t *testing.T) {
	t.Parallel()

	req := RandomUpdateRequest(t)

	testCases := []struct {
		name     string
		wantUser *User
		wantErr  error
	}{
		{
			name:     "repo call succeeds",
			wantUser: &User{},
			wantErr:  nil,
		},
		{
			name:     "repo returns any error",
			wantUser: nil,
			wantErr:  anyError,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &MockRepository{}
			svc := NewService(repo)

			repo.On("UpdateUser", context.Background(), req).Return(tc.wantUser, tc.wantErr)

			gotUser, gotErr := svc.UpdateUser(context.Background(), req)

			assert.Equal(t, tc.wantUser, gotUser)
			assert.ErrorIs(t, gotErr, tc.wantErr)
			repo.AssertExpectations(t)
		})
	}
}
