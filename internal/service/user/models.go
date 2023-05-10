package user

import (
	"crypto/rsa"
	"fmt"

	"github.com/angusgmorrison/realworld/pkg/primitive"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User is the central domain type for this package.
type User struct {
	ID           uuid.UUID
	Username     string
	Email        primitive.EmailAddress
	PasswordHash primitive.SensitiveString
	Bio          string
	ImageURL     string
}

// NewUserFromRegistrationRequest creates a new [User] from a [RegistrationRequest].
func NewUserFromRegistrationRequest(req *RegistrationRequest) (*User, error) {
	hash, err := bcryptHash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	return &User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: hash,
	}, nil
}

// HasPassword returns true if the given password matches the [User]'s password hash.
func (u *User) HasPassword(password primitive.SensitiveString) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return false
	}
	return true
}

// Equals returns true if two users are equal in all fields but their password
// hash, since direct comparison of bcrypt hashes without the input password is
// impossible by design.
func (u *User) Equals(other *User) bool {
	return u.ID == other.ID &&
		u.Username == other.Username &&
		u.Email == other.Email &&
		u.Bio == other.Bio &&
		u.ImageURL == other.ImageURL
}

// AuthenticatedUser is a User with a valid token.
type AuthenticatedUser struct {
	Token string
	User  *User
}

// Equals returns true if two authenticated users:
//   - have JWTs with the same subject claim (timestamp fields are not compared);
//   - are equal in all other fields but password hash (which can't be compared).
func (au *AuthenticatedUser) Equals(other *AuthenticatedUser, jwtPublicKey *rsa.PublicKey) bool {
	return jwtSubjectsEqual(au.Token, other.Token, jwtPublicKey) &&
		au.User.Equals(other.User)
}

// RegistrationRequest describes the data required to register a new user.
type RegistrationRequest struct {
	Username string                 `json:"username" validate:"required,max=32"`
	Email    primitive.EmailAddress `json:"email" validate:"required,email"`
	RequiredValidatingPassword
}

func (r *RegistrationRequest) String() string {
	return fmt.Sprintf("{Username:%q, Email: %q, Password: REDACTED}", r.Username, r.Email)
}

// AuthRequest describes the data required to authenticate a user.
//
// Password is not validated beyond checking its presence. For security, all
// errors related to hashing and comparing the password with the stored hash
// MUST be obfuscated using the generic AuthError.
type AuthRequest struct {
	Email    primitive.EmailAddress    `json:"email" validate:"required,email"`
	Password primitive.SensitiveString `json:"password" validate:"required"`
}

// UpdateRequest describes the data required to update a user. Since zero or
// more fields may be updated in a single request, pointer fields are required
// to distinguish the absence of a value (i.e. no change) from the zero value.
type UpdateRequest struct {
	UserID   uuid.UUID               `validate:"required"`
	Email    *primitive.EmailAddress `json:"email" validate:"omitempty,email"`
	Bio      *string                 `json:"bio"`
	ImageURL *string                 `json:"image" validate:"omitempty,url"`
	OptionalValidatingPassword
}
