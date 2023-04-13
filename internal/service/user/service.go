package user

import (
	"context"
	"crypto/rsa"
	"errors"
	"time"

	"github.com/angusgmorrison/realworld/pkg/validate"
	"github.com/google/uuid"
)

// Service describes the business API accessible to controller methods.
type Service interface {
	// Register a new user.
	Register(ctx context.Context, req *RegistrationRequest) (*AuthenticatedUser, error)
	// Authenticate a user, returning the user and token if successful.
	Authenticate(ctx context.Context, req *AuthRequest) (*AuthenticatedUser, error)
	// GetUser a user by ID.
	GetUser(ctx context.Context, id uuid.UUID) (*User, error)
	// Update a user.
	UpdateUser(ctx context.Context, req *UpdateRequest) (*User, error)
}

// Repository is a store of user data.
type Repository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email EmailAddress) (*User, error)
	CreateUser(ctx context.Context, req *RegistrationRequest) (*User, error)
	UpdateUser(ctx context.Context, req *UpdateRequest) (*User, error)
}

type service struct {
	repo          Repository
	jwtPrivateKey *rsa.PrivateKey
	jwtTTL        time.Duration
}

// NewService creates a new user service.
func NewService(repo Repository, jwtPrivateKey *rsa.PrivateKey, jwtTTL time.Duration) *service {
	return &service{
		repo:          repo,
		jwtPrivateKey: jwtPrivateKey,
		jwtTTL:        jwtTTL,
	}
}

// Register creates a new user and returns it with a signed JWT.
func (s *service) Register(ctx context.Context, req *RegistrationRequest) (*AuthenticatedUser, error) {
	if err := validate.Struct(req); err != nil {
		return nil, err
	}

	user, err := s.repo.CreateUser(ctx, req)
	if err != nil {
		return nil, err
	}

	jwt, err := newJWT(s.jwtPrivateKey, s.jwtTTL, user.ID.String())
	if err != nil {
		return nil, err
	}

	return &AuthenticatedUser{
		User:  user,
		Token: jwt,
	}, nil
}

// Authenticate looks up the user by email and compares its password hash with
// the request. If there is a match, the user is returned along with a signed
// JWT.
func (s *service) Authenticate(ctx context.Context, req *AuthRequest) (*AuthenticatedUser, error) {
	if err := validate.Struct(req); err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, &AuthError{cause: err}
		}
		return nil, err
	}

	if !user.HasPassword(req.Password) {
		return nil, &AuthError{cause: ErrPasswordMismatch}
	}

	jwt, err := newJWT(s.jwtPrivateKey, s.jwtTTL, user.ID.String())
	if err != nil {
		return nil, err
	}

	return &AuthenticatedUser{
		User:  user,
		Token: jwt,
	}, nil
}

// Get returns the user with the given ID.
func (s *service) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.GetUserByID(ctx, id)
}
