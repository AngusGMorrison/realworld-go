package user

import (
	"net/mail"

	"github.com/google/uuid"
)

type EmailAddress string

// NewEmailAddressFromString returns a new EmailAddress from a raw string. If
// the raw string is not a valid email address according to RFC 5322 standards,
// an EmailAddressValidationError is returned.
//
// Note that RFC 5322 permits some seemingly bad email addresses such as
// mail@localdestination, since an email addresses don't necessarily need a
// public domain.
func NewEmailAddressFromString(raw string) (EmailAddress, error) {
	if _, err := mail.ParseAddress(raw); err != nil {
		return "", ErrEmailAddressUnparseable
	}

	return EmailAddress(raw), nil
}

// User is the central domain type for this package.
type User struct {
	username       string
	email          EmailAddress
	passwordDigest string
	bio            string
	imageURL       string
}

func NewUser(
	username string,
	email EmailAddress,
	digest string,
	bio string,
	imageURL string,
) *User {
	return &User{
		username:       username,
		email:          email,
		passwordDigest: digest,
		bio:            bio,
		imageURL:       imageURL,
	}
}

// Username returns the user's username.
func (u *User) Username() string {
	return u.username
}

// Email returns the user's email address.
func (u *User) Email() EmailAddress {
	return u.email
}

// Digest returns the user's password digest.
func (u *User) Digest() string {
	return u.passwordDigest
}

// Bio returns the user's bio.
func (u *User) Bio() string {
	return u.bio
}

// ImageURL returns the user's image URL.
func (u *User) ImageURL() string {
	return u.imageURL
}

// AuthenticatedUser is a User with a valid token.
type AuthenticatedUser struct {
	token string
	*User
}

func (au *AuthenticatedUser) Token() string {
	return au.token
}

// RegisterRequest describes the data required to register a new user.
type RegisterRequest struct {
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
