package user

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"time"
)

// service satisfies the inbound Service interface.
type service struct {
	repo      Repository
	jwtSource *jwtSource
}

func NewService(repo Repository, jwtPrivateKey *rsa.PrivateKey, jwtTTL time.Duration) Service {
	return &service{
		repo: repo,
		jwtSource: &jwtSource{
			privateKey: jwtPrivateKey,
			ttl:        jwtTTL,
		},
	}
}

// Register creates a new user and returns it with a signed JWT.
func (s *service) Register(ctx context.Context, req *RegistrationRequest) (*AuthenticatedUser, error) {
	user, err := s.repo.CreateUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create user from %#v: %w", req, err)
	}

	jwt, err := s.jwtSource.newWithSubject(user.ID())
	if err != nil {
		return nil, err
	}

	return &AuthenticatedUser{
		user:  user,
		token: jwt,
	}, nil
}

// Authenticate looks up the user by email and compares its password hash with
// the request. If there is a match, the user is returned along with a signed
// JWT.
func (s *service) Authenticate(ctx context.Context, req *AuthRequest) (*AuthenticatedUser, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.email)
	if err != nil {
		var notFoundErr *NotFoundError
		if errors.As(err, &notFoundErr) {
			return nil, &AuthError{Cause: err}
		}
		return nil, fmt.Errorf("get user from %#v: %w", req, err)
	}

	if err := tryAuthenticate(user.passwordHash, req.passwordCandidate); err != nil {
		return nil, err
	}

	jwt, err := s.jwtSource.newWithSubject(user.ID())
	if err != nil {
		return nil, err
	}

	return &AuthenticatedUser{
		user:  user,
		token: jwt,
	}, nil
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
