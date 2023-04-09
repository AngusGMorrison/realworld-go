package user

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	// Register a new user.
	Register(ctx context.Context, req *RegistrationRequest) (*AuthenticatedUser, error)
	// Authenticate a user, returning the user and token if successful.
	Authenticate(ctx context.Context, req *AuthRequest) (*AuthenticatedUser, error)
	// Get a user by ID.
	Get(ctx context.Context, id uuid.UUID) (*AuthenticatedUser, error)
	// Update a user.
	Update(ctx context.Context, req *UpdateRequest) (*AuthenticatedUser, error)
}
