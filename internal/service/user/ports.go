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
type Repository interface {
	// GetUserByID retrieves the [User] with `ID`.
	// 	- MUST return [ErrUserNotFound] if no such User exists.
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)

	// GetUserByEmail returns a user by email. MUST return [ErrUserNotFound] if
	// no such user exists.
	GetUserByEmail(ctx context.Context, email EmailAddress) (*User, error)

	// CreateUser persists a new user.
	// 	- MUST return [ErrEmailRegistered] if the email address is in use.
	// 	- MUST return [ErrUsernameTaken] if the username is in use.
	CreateUser(ctx context.Context, req *RegistrationRequest) (*User, error)

	// UpdateUser updates an existing user.
	// 	- MUST return [ErrEmailRegistered] on an attempt to update the email
	//	  address to one that is already in use.
	// 	- MUST return [ErrUsernameTaken] on an attempt to update the username to
	//	  one that is already in use.
	UpdateUser(ctx context.Context, req *UpdateRequest) (*User, error)
}
