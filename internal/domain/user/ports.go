package user

import (
	"context"
	"github.com/google/uuid"
)

// Service describes the business API accessible to inbound methods.
type Service interface {
	// Register a new user.
	Register(ctx context.Context, req *RegistrationRequest) (*User, error)

	// Authenticate a user, returning the authenticated [User] if successful.
	//
	// # Errors
	//	- [AuthError].
	Authenticate(ctx context.Context, req *AuthRequest) (*User, error)

	// GetUser a user by ID.
	GetUser(ctx context.Context, id uuid.UUID) (*User, error)

	// UpdateUser updates an existing user.
	UpdateUser(ctx context.Context, req *UpdateRequest) (*User, error)
}

// Repository is a store of user data.
//   - MUST return [ValidationError] if a database constraint is violated.
type Repository interface {
	// GetUserByID retrieves the [User] with `ID`.
	//
	// # Errors
	// 	- [NotFoundError] if no such User exists.
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)

	// GetUserByEmail returns a user by email.
	//
	// # Errors
	// 	- [NotFoundError] if no such User exists.
	GetUserByEmail(ctx context.Context, email EmailAddress) (*User, error)

	// CreateUser persists a new user.
	//
	// # Errors
	// 	- [ValidationError] if either email or username is already taken.
	CreateUser(ctx context.Context, req *RegistrationRequest) (*User, error)

	// UpdateUser updates an existing user.
	//
	// # Errors
	// 	- [NotFoundError] if no such User exists.
	// 	- [ValidationError] if either email or username is already taken.
	UpdateUser(ctx context.Context, req *UpdateRequest) (*User, error)
}
