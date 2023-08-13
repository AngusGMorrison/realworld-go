package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// service satisfies the inbound Service interface.
type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// Register creates a new user and returns it with a signed JWT.
func (s *service) Register(ctx context.Context, req *RegistrationRequest) (*User, error) {
	user, err := s.repo.CreateUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create user from %#v: %w", req, err)
	}

	return user, nil
}

// Authenticate looks up the user by email and compares its password hash with
// the request. If there is a match, the user is returned along with a signed
// JWT.
func (s *service) Authenticate(ctx context.Context, req *AuthRequest) (*User, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.email)
	if err != nil {
		var notFoundErr *NotFoundError
		if errors.As(err, &notFoundErr) {
			return nil, &AuthError{Cause: err}
		}
		return nil, fmt.Errorf("get user from %#v: %w", req, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.passwordHash.Expose(), []byte(req.passwordCandidate.Expose())); err != nil {
		return nil, &AuthError{Cause: err}
	}

	return user, nil
}

// GetUser returns the user with the given IDFieldValue.
func (s *service) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user with IDFieldValue %s: %w", id, err)
	}

	return user, nil
}

// UpdateUser updates the user with the given IDFieldValue.
func (s *service) UpdateUser(ctx context.Context, req *UpdateRequest) (*User, error) {
	user, err := s.repo.UpdateUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update user with IDFieldValue %s: %w", req.UserID(), err)
	}

	return user, nil
}
