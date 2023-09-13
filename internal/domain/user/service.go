package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// service satisfies the inbound Service interface.
type service struct {
	repo               Repository
	passwordComparator passwordComparator
}

func NewService(repo Repository) Service {
	return &service{
		repo:               repo,
		passwordComparator: bcryptCompare,
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

	if err := s.passwordComparator(user.passwordHash, req.passwordCandidate); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser returns the user with the given ID.
func (s *service) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user with ID %s: %w", id, err)
	}

	return user, nil
}

// UpdateUser updates the user with the given ID.
func (s *service) UpdateUser(ctx context.Context, req *UpdateRequest) (*User, error) {
	user, err := s.repo.UpdateUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update user with ID %s: %w", req.UserID(), err)
	}

	return user, nil
}
