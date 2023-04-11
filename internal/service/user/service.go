package user

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/angusgmorrison/realworld/pkg/validate"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// Service describes the business API accessible to ingress methods.
type Service interface {
	// Register a new user.
	Register(ctx context.Context, req *RegistrationRequest) (*AuthenticatedUser, error)
	// Authenticate a user, returning the user and token if successful.
	Authenticate(ctx context.Context, req *AuthRequest) (*AuthenticatedUser, error)
	// Get a user by ID.
	Get(ctx context.Context, id uuid.UUID) (*User, error)
	// Update a user.
	Update(ctx context.Context, req *UpdateRequest) (*User, error)
}

// Repository is a store of user data.
type Repository interface {
	CreateUser(ctx context.Context, req *RegistrationRequest) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email EmailAddress) (*User, error)
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

	jwt, err := s.newJWT(user.ID)
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

	reqPasswordHash, err := req.PasswordHash()
	if err != nil {
		// The only way this operation can fail is if the input password exceeds
		// bycrypt's maxiumum hashable length. We return an AuthError to
		// obfuscate the cause.
		return nil, &AuthError{cause: err}
	}

	if user.PasswordHash != reqPasswordHash {
		return nil, &AuthError{cause: ErrPasswordMismatch}
	}

	jwt, err := s.newJWT(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthenticatedUser{
		User:  user,
		Token: jwt,
	}, nil
}

// Get returns the user with the given ID.
func (s *service) Get(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *service) newJWT(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(s.jwtTTL).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(s.jwtPrivateKey)
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}

	return signedToken, nil
}
