package user

import (
	"crypto/rsa"
	"errors"
	"fmt"
	neturl "net/url"
	"regexp"

	"github.com/angusgmorrison/realworld/pkg/valueobj"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

const (
	minUsernameLen  = 1
	maxUsernameLen  = 32
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

// username represents a valid username.
type username struct {
	raw string
}

// ParseUsername returns either a valid [Username] or an error if the raw username
// is invalid.
func ParseUsername(raw string) (username, error) {
	if len(raw) < minUsernameLen {
		return username{}, ErrUsernameTooShort
	}
	if len(raw) > maxUsernameLen {
		return username{}, ErrUsernameTooLong
	}
	if !usernameRegex.MatchString(raw) {
		return username{}, ErrUsernameFormat
	}

	return username{raw: raw}, nil
}

func (u username) String() string {
	return u.raw
}

// User is the central domain type for this package.
type user struct {
	id           uuid.UUID
	username     username
	email        valueobj.EmailAddress
	passwordHash passwordHash
	bio          string
	imageURL     url
}

func NewUser(
	id uuid.UUID,
	username username,
	email valueobj.EmailAddress,
	passwordHash passwordHash,
	bio string,
	imageURL url,
) *user {
	return &user{
		id:           id,
		username:     username,
		email:        email,
		passwordHash: passwordHash,
		bio:          bio,
		imageURL:     imageURL,
	}
}

// ID returns the user's unique identifier.
func (user *user) ID() uuid.UUID {
	return user.id
}

// Username returns the user's username.
func (user *user) Username() username {
	return user.username
}

// Email returns the user's email address.
func (user *user) Email() valueobj.EmailAddress {
	return user.email
}

// passwordHash returns the user's password hash.
func (user *user) PasswordHash() passwordHash {
	return user.passwordHash
}

// Bio returns the user's bio, which may be empty.
func (user *user) Bio() string {
	return user.bio
}

// ImageURL returns the user's image URL, which may be nil.
func (user *user) ImageURL() url {
	return user.imageURL
}

// Equal returns true if two users are equal in all fields but their password
// hash, since direct comparison of bcrypt hashes without the input password is
// impossible by design.
func (u *user) Equal(other *user) bool {
	return u.id == other.id &&
		u.username == other.username &&
		u.email == other.email &&
		u.bio == other.bio &&
		u.imageURL == other.imageURL
}

// authenticatedUser is a User with a valid JWT.
type authenticatedUser struct {
	token JWT
	user  *user
}

// Token returns the user's JWT.
func (au *authenticatedUser) Token() JWT {
	return au.token
}

// User returns the user.
func (au *authenticatedUser) User() *user {
	return au.user
}

// Equals returns true if two authenticated users:
//   - have JWTs with the same subject claim (timestamp fields are not compared);
//   - are equal in all other fields but password hash (which can't be compared).
func (au *authenticatedUser) Equal(other *authenticatedUser, jwtPublicKey *rsa.PublicKey) bool {
	return jwtSubjectsEqual(au.token, other.token, jwtPublicKey) &&
		cmp.Equal(au.User, other.User)
}

// registrationRequest carries validated data required to register a new user.
type registrationRequest struct {
	username     username
	email        valueobj.EmailAddress
	passwordHash passwordHash
}

// NewRegistrationRequest parses each of its inputs and returns a valid
// [registrationRequest]. If any one of the inputs can't be parsed, the
// corresponding error is returned instead.
func NewRegistrationRequest(
	username username, email valueobj.EmailAddress, passwordHash passwordHash,
) (*registrationRequest, error) {
	return &registrationRequest{
		username:     username,
		email:        email,
		passwordHash: passwordHash,
	}, nil
}

// Username returns the username.
func (r *registrationRequest) Username() username {
	return r.username
}

// EmailAddress returns the email address.
func (r *registrationRequest) EmailAddress() valueobj.EmailAddress {
	return r.email
}

// passwordHash returns the password hash.
func (r *registrationRequest) PasswordHash() passwordHash {
	return r.passwordHash
}

// authRequest describes the data required to authenticate a user.
type authRequest struct {
	email             valueobj.EmailAddress
	passwordCandidate PasswordCandidate
}

// NewAuthRequest parses the candidate email address and returns a valid [authRequest].
//
// Failure to parse the email address results in a generic [AuthError] to avoid
// leaking information about the nature of the authentication failure to the end
// user.
func NewAuthRequest(email valueobj.EmailAddress, passwordCandidate PasswordCandidate) (*authRequest, error) {
	return &authRequest{
		email:             email,
		passwordCandidate: passwordCandidate,
	}, nil
}

// Email returns the email address.
func (ar *authRequest) Email() valueobj.EmailAddress {
	return ar.email
}

// PasswordCandidate returns the password candidate.
func (ar *authRequest) PasswordCandidate() PasswordCandidate {
	return ar.passwordCandidate
}

// url represents a valid, immutable URL.
type url struct {
	inner *neturl.URL
}

// NewURL parses `candidate`, returning a valid [url] if successful, and an
// error otherwise.
func NewURL(candidate string) (*url, error) {
	u, err := neturl.Parse(candidate)
	if err != nil {
		return nil, fmt.Errorf("parse candidate url %q: %w", candidate, err)
	}

	return &url{inner: u}, nil
}

func (u url) String() string {
	return u.inner.String()
}

// UpdateRequest describes the data required to update a user. Since zero or
// more fields may be updated in a single request, pointer fields are required
// to distinguish the absence of a value (i.e. no change) from the zero value.
type updateRequest struct {
	userID   uuid.UUID
	email    *valueobj.EmailAddress
	bio      *string
	imageURL *url
	pwHash   *passwordHash
}

// ErrUserIDEmpty is returned by the domain whenever an empty user UUID is provided.
var ErrUserIDEmpty = errors.New("user ID must not be empty")

// NewUpdateRequest invokes each of its optional setters in turn, returning a
// valid [updateRequest] if all succeed, or a composition of all validation errors.
func NewUpdateRequest(
	userID uuid.UUID,
	email *valueobj.EmailAddress,
	bio *string,
	imageURL *url,
	passwordHash *passwordHash,
) (*updateRequest, error) {
	if (userID == uuid.UUID{}) {
		return nil, ErrUserIDEmpty
	}

	return &updateRequest{
		userID:   userID,
		email:    email,
		bio:      bio,
		imageURL: imageURL,
		pwHash:   passwordHash,
	}, nil
}

func (req *updateRequest) UserID() uuid.UUID {
	return req.userID
}

func (req *updateRequest) Email() valueobj.EmailAddress {
	if req.email == nil {
		return valueobj.EmailAddress{}
	}
	return *req.email
}

func (req *updateRequest) Bio() string {
	if req.bio == nil {
		return ""
	}
	return *req.bio
}

func (req *updateRequest) ImageURL() url {
	if req.imageURL == nil {
		return url{}
	}
	return *req.imageURL
}

func (req *updateRequest) PasswordHash() passwordHash {
	if req.pwHash == nil {
		return passwordHash{}
	}
	return *req.pwHash
}
