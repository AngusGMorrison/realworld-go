package user

import (
	"errors"
	"fmt"
	neturl "net/url"
	"regexp"

	"github.com/angusgmorrison/realworld/pkg/tidy"
)

const (
	minUsernameLen  = 1
	maxUsernameLen  = 16
	usernamePattern = `^[a-zA-Z0-9_]+$`
)

var (
	// ErrUsernameTooShort is returned by the [NewUsername] constructor when the
	// raw username is too short.
	ErrUsernameTooShort = fmt.Errorf("username must be at lest %d character long", minUsernameLen)

	// ErrUsernameTooLong is returned by the [NewUsername] constructor when the raw
	// username is too long.
	ErrUsernameTooLong = fmt.Errorf("username must be at most %d characters long", maxUsernameLen)

	usernameRegex = regexp.MustCompile(usernamePattern)

	// ErrUsernameFormat is returned by the [NewUsername] constructor when the
	// raw username contains invalid characters.
	ErrUsernameFormat = errors.New(
		"username must only contain alphanumeric characters and underscores",
	)
)

// Username represents a valid Username.
type Username struct {
	raw string
}

// ParseUsername returns either a valid [Username] or an error if the raw username
// is invalid.
func ParseUsername(raw string) (Username, error) {
	if len(raw) < minUsernameLen {
		return Username{}, ErrUsernameTooShort
	}
	if len(raw) > maxUsernameLen {
		return Username{}, ErrUsernameTooLong
	}
	if !usernameRegex.MatchString(raw) {
		return Username{}, ErrUsernameFormat
	}

	return Username{raw: raw}, nil
}

func (u Username) String() string {
	return u.raw
}

// NonZero satisfies [tidy.Strict].
func (u Username) NonZero() error {
	if u == (Username{}) {
		return &tidy.ZeroValueError{ZeroStrict: u}
	}
	return nil
}

// Bio represents a user's bio. Implements [tidy.Strict].
type Bio string

// NonZero satisfies [tidy.Strict].
func (b Bio) NonZero() error {
	if b == "" {
		return &tidy.ZeroValueError{ZeroStrict: b}
	}
	return nil
}

// URL represents a valid, immutable URL.
type URL struct {
	inner *neturl.URL
}

// ParseURL parses `candidate`, returning a valid [URL] if successful, and an
// error otherwise.
func ParseURL(candidate string) (*URL, error) {
	u, err := neturl.Parse(candidate)
	if err != nil {
		return nil, fmt.Errorf("parse candidate url %q: %w", candidate, err)
	}

	return &URL{inner: u}, nil
}

func (u URL) String() string {
	return u.inner.String()
}

// NonZero satisfies [tidy.Strict].
func (u URL) NonZero() error {
	if u == (URL{}) {
		return &tidy.ZeroValueError{ZeroStrict: u}
	}
	return nil
}

// User is the central domain type for this package.
type User struct {
	id           tidy.UUIDv4
	username     Username
	email        tidy.EmailAddress
	passwordHash PasswordHash
	bio          tidy.Option[Bio]
	imageURL     tidy.Option[URL]
}

// NewUser returns a new User instance, or an error if any of its required
// constituent fields are zero values.
func NewUser(
	id tidy.UUIDv4,
	username Username,
	email tidy.EmailAddress,
	passwordHash PasswordHash,
	bio tidy.Option[Bio],
	imageURL tidy.Option[URL],
) (User, error) {
	user := User{
		id:           id,
		username:     username,
		email:        email,
		passwordHash: passwordHash,
		bio:          bio,
		imageURL:     imageURL,
	}

	if err := user.NonZero(); err != nil {
		return User{}, err
	}

	return user, nil
}

// NonZero satisfies [tidy.Strict].
func (u *User) NonZero() error {
	if err := tidy.NonZero(u.id, u.username, u.email, u.passwordHash, u.bio, u.imageURL); err != nil {
		return fmt.Errorf("User entity must not have zero-value fields: %w", err)
	}
	return nil
}

// ID returns the user's unique identifier.
func (u *User) ID() tidy.UUIDv4 {
	return u.id
}

// Username returns the user's username.
func (u *User) Username() Username {
	return u.username
}

// Email returns the user's email address.
func (u *User) Email() tidy.EmailAddress {
	return u.email
}

// PasswordHash returns the user's password hash.
func (u *User) PasswordHash() PasswordHash {
	return u.passwordHash
}

// Bio returns the user's bio as a [tidy.Option[Bio]].
func (u *User) Bio() tidy.Option[Bio] {
	return u.bio
}

// ImageURL returns the user's image URL as a [tidy.Option[URL]].
func (u *User) ImageURL() tidy.Option[URL] {
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

// Token returns the user's JWT.
func (au *AuthenticatedUser) Token() JWT {
	return au.token
}

// User returns the user.
func (au *AuthenticatedUser) User() *User {
	return au.user
}

// Equals returns true if two authenticated users are equal in their user field.
// The JWT is not considered to be part of the user's identity.
func (au *AuthenticatedUser) Equals(other *AuthenticatedUser) bool {
	return au.User().Equals(other.User())
}

// RegistrationRequest carries validated data required to register a new user.
type RegistrationRequest struct {
	username     Username
	email        tidy.EmailAddress
	passwordHash PasswordHash
}

// NewRegistrationRequest parses each of its inputs and returns a valid
// [registrationRequest]. If any one of the inputs can't be parsed, the
// corresponding error is returned instead.
func NewRegistrationRequest(
	username Username, email tidy.EmailAddress, passwordHash PasswordHash,
) (RegistrationRequest, error) {
	req := RegistrationRequest{
		username:     username,
		email:        email,
		passwordHash: passwordHash,
	}

	if err := req.NonZero(); err != nil {
		return RegistrationRequest{}, err
	}

	return req, nil
}

// NonZero satisfies [tidy.Strict].
func (r *RegistrationRequest) NonZero() error {
	if r == nil {
		return &tidy.ZeroValueError{ZeroStrict: r}
	}

	if err := tidy.NonZero(r.username, r.email, r.passwordHash); err != nil {
		return fmt.Errorf("RegistrationRequest strict type must not have zero-value fields: %w", err)
	}

	return nil
}

// Username returns the username.
func (r *RegistrationRequest) Username() Username {
	return r.username
}

// EmailAddress returns the email address.
func (r *RegistrationRequest) EmailAddress() tidy.EmailAddress {
	return r.email
}

// PasswordHash returns the password hash.
func (r *RegistrationRequest) PasswordHash() PasswordHash {
	return r.passwordHash
}

// AuthRequest describes the data required to authenticate a user.
type AuthRequest struct {
	email             tidy.EmailAddress
	passwordCandidate PasswordCandidate
}

// NewAuthRequest parses the candidate email address and returns a valid [authRequest].
//
// Failure to parse the email address results in a generic [AuthError] to avoid
// leaking information about the nature of the authentication failure to the end
// user.
func NewAuthRequest(email tidy.EmailAddress, passwordCandidate PasswordCandidate) (AuthRequest, error) {
	req := AuthRequest{
		email:             email,
		passwordCandidate: passwordCandidate,
	}

	if err := req.NonZero(); err != nil {
		return AuthRequest{}, err
	}

	return req, nil
}

// NonZero satisfies [tidy.Strict]. Any password candidate is considered
// non-zero, since it is the comparison with a user's password hash that
// determines its validity.
func (ar *AuthRequest) NonZero() error {
	if ar == nil {
		return &tidy.ZeroValueError{ZeroStrict: ar}
	}

	if err := tidy.NonZero(ar.Email()); err != nil {
		return fmt.Errorf("AuthRequest strict type must not have zero-value fields: %w", err)
	}

	return nil
}

// Email returns the email address.
func (ar *AuthRequest) Email() tidy.EmailAddress {
	return ar.email
}

// PasswordCandidate returns the password candidate.
func (ar *AuthRequest) PasswordCandidate() PasswordCandidate {
	return ar.passwordCandidate
}

// UpdateRequest describes the data required to update a user. Since zero or
// more fields may be updated in a single request, pointer fields are required
// to distinguish the absence of a value (i.e. no change) from the zero value.
type UpdateRequest struct {
	userID   tidy.UUIDv4
	email    tidy.Option[tidy.EmailAddress]
	bio      tidy.Option[Bio]
	imageURL tidy.Option[URL]
	pwHash   tidy.Option[PasswordHash]
}

// NewUpdateRequest invokes each of its optional setters in turn, returning a
// valid [updateRequest] if all succeed, or a composition of all validation errors.
func NewUpdateRequest(
	userID tidy.UUIDv4,
	email tidy.Option[tidy.EmailAddress],
	bio tidy.Option[Bio],
	imageURL tidy.Option[URL],
	passwordHash tidy.Option[PasswordHash],
) (UpdateRequest, error) {
	req := UpdateRequest{
		userID:   userID,
		email:    email,
		bio:      bio,
		imageURL: imageURL,
		pwHash:   passwordHash,
	}

	if err := req.NonZero(); err != nil {
		return UpdateRequest{}, err
	}

	return req, nil
}

// NonZero satisfies [tidy.Strict].
func (req *UpdateRequest) NonZero() error {
	if req == nil {
		return &tidy.ZeroValueError{ZeroStrict: req}
	}

	if err := tidy.NonZero(req.userID); err != nil {
		return fmt.Errorf("UpdateRequest strict type must not have zero-value fields: %w", err)
	}

	return nil
}

// UserID returns the user's ID.
func (req *UpdateRequest) UserID() tidy.UUIDv4 {
	return req.userID
}

// Email returns the user's new email address, or the zero value if no change is
// requested.
func (req *UpdateRequest) Email() tidy.Option[tidy.EmailAddress] {
	return req.email
}

// Bio returns the user's new bio, or the empty string if no change is requested.
func (req *UpdateRequest) Bio() tidy.Option[Bio] {
	return req.bio
}

// ImageURL returns the user's new image URL, or the zero value if no change is
// requested.
func (req *UpdateRequest) ImageURL() tidy.Option[URL] {
	return req.imageURL
}

// PasswordHash returns the user's new password hash, or the zero value if no change
// is requested.
func (req *UpdateRequest) PasswordHash() tidy.Option[PasswordHash] {
	return req.pwHash
}
