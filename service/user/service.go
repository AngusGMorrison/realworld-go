package user

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	// Register a new user.
	Register(ctx context.Context, req *RegisterRequest) (*User, error)
	// Authenticate a user.
	Authenticate(ctx context.Context, req *AuthRequest) (*User, error)
	// Get a user by ID.
	Get(ctx context.Context, id uuid.UUID) (*User, error)
	// Update a user.
	Update(ctx context.Context, req *UpdateRequest) (*User, error)
}
