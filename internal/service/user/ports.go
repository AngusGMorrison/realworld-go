package user

import (
	"context"
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

	// UpdateUser updates an existing user.
	UpdateUser(ctx context.Context, req *UpdateRequest) (*User, error)
}

// Repository is a store of user data.
//   - MUST return [ValidationError] if a database constraint is violated.
type Repository interface {
	// GetUserByID retrieves the [User] with `ID`.
	// 	- MUST return [NotFoundError] if no such User exists.
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)

	// GetUserByEmail returns a user by email.
	//  - MUST return [NotFoundError] if no such user exists.
	GetUserByEmail(ctx context.Context, email EmailAddress) (*User, error)

	// CreateUser persists a new user.
	CreateUser(ctx context.Context, req *RegistrationRequest) (*User, error)

	// UpdateUser updates an existing user.
	UpdateUser(ctx context.Context, req *UpdateRequest) (*User, error)
}
