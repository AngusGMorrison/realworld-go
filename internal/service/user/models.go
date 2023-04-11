package user

import (
	"github.com/google/uuid"
)

type EmailAddress string

// User is the central domain type for this package.
type User struct {
	ID           uuid.UUID
	Username     string
	Email        EmailAddress
	PasswordHash string
	Bio          string
	ImageURL     string
}

// AuthenticatedUser is a User with a valid token.
type AuthenticatedUser struct {
	Token string
	User  *User
}

// RegistrationRequest describes the data required to register a new user.
type RegistrationRequest struct {
	Username string       `json:"username" validate:"required"`
	Email    EmailAddress `json:"email" validate:"required,email"`
	RequiredValidatingPassword
}

// AuthRequest describes the data required to authenticate a user.
//
// Password is not validated beyond checking its presence. For security, all
// errors related to hashing and comparing the password with the stored hash
// MUST be obfuscated using the generic AuthError.
type AuthRequest struct {
	Email    EmailAddress `json:"email" validate:"required,email"`
	Password string       `json:"password" validate:"required"`
}

func (ar *AuthRequest) PasswordHash() (string, error) {
	return bcryptHash(ar.Password)
}

// UpdateRequest describes the data required to update a user. Since zero or
// more fields may be updated in a single request, pointer fields are required
// to distinguish the absence of a value (i.e. no change) from the zero value.
type UpdateRequest struct {
	UserID   uuid.UUID     `validate:"required"`
	Email    *EmailAddress `validate:"omitempty,email"`
	Bio      *string
	ImageURL *string `validate:"omitempty,url"`
	OptionalValidatingPassword
}
