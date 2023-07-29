package user

import (
	"fmt"
	"github.com/angusgmorrison/logfusc"
	"github.com/angusgmorrison/realworld/pkg/option"
	"github.com/google/uuid"
	"net/mail"
	neturl "net/url"
	"regexp"
)

// EmailAddress is a dedicated string type for valid email addresses. New
// instances are validated for RFC5332 compliance.
type EmailAddress struct {
	raw string
}

// ParseEmailAddress returns a new email address from `candidate`, validating that
// the email address conforms to RFC5332 standards (with the minor
// divergences introduce by the Go standard library, documented in [net/mail]).
//
// The formats that constitute a valid email address may surprise you. For
// example, single-value domain names like `angus@com` are valid.
func ParseEmailAddress(candidate string) (EmailAddress, error) {
	if _, err := mail.ParseAddress(candidate); err != nil {
		return EmailAddress{}, NewEmailAddressFormatError(candidate)
	}

	return EmailAddress{raw: candidate}, nil
}

// String returns the raw email address.
func (ea EmailAddress) String() string {
	return ea.raw
}

const (
	// UsernameMinLen is the minimum length of a username.
	UsernameMinLen = 1

	// UsernameMaxLen is the maximum length of a username.
	UsernameMaxLen = 16

	UsernamePattern = `^[a-zA-Z0-9_]+$`
)

var (
	usernameRegex = regexp.MustCompile(UsernamePattern)
)

// Username represents a valid Username.
type Username struct {
	raw string
}

// ParseUsername returns either a valid [Username] or an error indicating why
// the raw username was invalid.
func ParseUsername(raw string) (Username, error) {
	if len(raw) < UsernameMinLen {
		return Username{}, NewUsernameTooShortError()
	}
	if len(raw) > UsernameMaxLen {
		return Username{}, NewUsernameTooLongError()
	}
	if !usernameRegex.MatchString(raw) {
		return Username{}, NewUsernameFormatError()
	}

	return Username{raw: raw}, nil
}

func (u Username) String() string {
	return u.raw
}

// Bio represents a user's bio.
type Bio string

// URL represents a valid, immutable URL.
type URL struct {
	inner *neturl.URL
}

// ParseURL returns a valid [URL] if successful, and an error otherwise.
func ParseURL(candidate string) (URL, error) {
	u, err := neturl.Parse(candidate)
	if err != nil {
		return URL{}, fmt.Errorf("parse candidate url %q: %w", candidate, err)
	}

	return URL{inner: u}, nil
}

func (u URL) String() string {
	if u.inner == nil {
		return ""
	}
	return u.inner.String()
}

// User is the central domain type for this package.
type User struct {
	id           uuid.UUID
	username     Username
	email        EmailAddress
	passwordHash PasswordHash
	bio          option.Option[Bio]
	imageURL     option.Option[URL]
}

func NewUser(
	id uuid.UUID,
	username Username,
	email EmailAddress,
	passwordHash PasswordHash,
	bio option.Option[Bio],
	imageURL option.Option[URL],
) *User {
	return &User{
		id:           id,
		username:     username,
		email:        email,
		passwordHash: passwordHash,
		bio:          bio,
		imageURL:     imageURL,
	}
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Username() Username {
	return u.username
}

func (u *User) Email() EmailAddress {
	return u.email
}

func (u *User) PasswordHash() PasswordHash {
	return u.passwordHash
}

func (u *User) Bio() option.Option[Bio] {
	return u.bio
}

func (u *User) ImageURL() option.Option[URL] {
	return u.imageURL
}

// Equals returns true if two users are equal in all fields but their password
// hash, since direct comparison of bcrypt hashes without the input password is
// impossible by design.
func (u *User) Equals(other *User) bool {
	return u.id == other.id &&
		u.username == other.username &&
		u.email == other.email &&
		u.bio == other.bio &&
		u.imageURL == other.imageURL
}

// AuthenticatedUser is a User with a valid JWT.
type AuthenticatedUser struct {
	token JWT
	user  *User
}

func (au *AuthenticatedUser) Token() JWT {
	return au.token
}

func (au *AuthenticatedUser) User() *User {
	return au.user
}

// Equals returns true if two authenticated users are equal in their user Field.
// The JWT is not considered to be part of the user's identity.
func (au *AuthenticatedUser) Equals(other *AuthenticatedUser) bool {
	return au.User().Equals(other.User())
}

// RegistrationRequest carries validated data required to register a new user.
type RegistrationRequest struct {
	username     Username
	email        EmailAddress
	passwordHash PasswordHash
}

func NewRegistrationRequest(
	username Username, email EmailAddress, passwordHash PasswordHash,
) *RegistrationRequest {
	return &RegistrationRequest{
		username:     username,
		email:        email,
		passwordHash: passwordHash,
	}
}

func (r *RegistrationRequest) Username() Username {
	return r.username
}

func (r *RegistrationRequest) EmailAddress() EmailAddress {
	return r.email
}

func (r *RegistrationRequest) PasswordHash() PasswordHash {
	return r.passwordHash
}

// AuthRequest describes the data required to authenticate a user.
type AuthRequest struct {
	email             EmailAddress
	passwordCandidate logfusc.Secret[string]
}

func NewAuthRequest(email EmailAddress, passwordCandidate logfusc.Secret[string]) *AuthRequest {
	return &AuthRequest{
		email:             email,
		passwordCandidate: passwordCandidate,
	}
}

func (ar *AuthRequest) Email() EmailAddress {
	return ar.email
}

func (ar *AuthRequest) PasswordCandidate() logfusc.Secret[string] {
	return ar.passwordCandidate
}

// UpdateRequest describes the data required to update a user. All fields but
// userID are optional. To unset a Field, pass [Option] containing the zero
// value of the Field's type.
type UpdateRequest struct {
	userID   uuid.UUID
	email    option.Option[EmailAddress]
	bio      option.Option[Bio]
	imageURL option.Option[URL]
	pwHash   option.Option[PasswordHash]
}

func NewUpdateRequest(
	userID uuid.UUID,
	email option.Option[EmailAddress],
	passwordHash option.Option[PasswordHash],
	bio option.Option[Bio],
	imageURL option.Option[URL],
) *UpdateRequest {
	return &UpdateRequest{
		userID:   userID,
		email:    email,
		bio:      bio,
		imageURL: imageURL,
		pwHash:   passwordHash,
	}
}

func (req *UpdateRequest) UserID() uuid.UUID {
	return req.userID
}

func (req *UpdateRequest) Email() option.Option[EmailAddress] {
	return req.email
}

func (req *UpdateRequest) Bio() option.Option[Bio] {
	return req.bio
}

func (req *UpdateRequest) ImageURL() option.Option[URL] {
	return req.imageURL
}

func (req *UpdateRequest) PasswordHash() option.Option[PasswordHash] {
	return req.pwHash
}
