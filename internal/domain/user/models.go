package user

import (
	"fmt"
	"github.com/angusgmorrison/logfusc"
	"github.com/angusgmorrison/realworld/pkg/option"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

const (
	PasswordMinLen = 8
	PasswordMaxLen = 72
)

// PasswordHash represents a validated and hashed password.
type PasswordHash struct {
	inner logfusc.Secret[[]byte]
}

func ParsePassword(candidate logfusc.Secret[string]) (PasswordHash, error) {
	if err := validatePasswordCandidate(candidate); err != nil {
		return PasswordHash{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(candidate.Expose()), bcrypt.DefaultCost)
	if err != nil {
		return PasswordHash{}, fmt.Errorf("hash password: %w", err)
	}

	return PasswordHash{inner: logfusc.NewSecret(hash)}, nil
}

func validatePasswordCandidate(candidate logfusc.Secret[string]) error {
	exposedPassword := candidate.Expose()
	if len(exposedPassword) < PasswordMinLen {
		return NewPasswordTooShortError()
	}
	if len(exposedPassword) > PasswordMaxLen {
		return NewPasswordTooLongError()
	}
	return nil
}

// NewPasswordHashFromTrustedSource wraps a hashed password in a [PasswordHash].
func NewPasswordHashFromTrustedSource(raw logfusc.Secret[[]byte]) PasswordHash {
	return PasswordHash{inner: raw}
}

// Expose returns the hashed password as a byte slice.
func (ph PasswordHash) Expose() []byte {
	return ph.inner.Expose()
}

// GoString satisfies [fmt.GoStringer]. Must be implemented manually to ensure
// that the inner password hash is not printed. The fmt package uses reflection to
// print unexported fields without invoking their String or GoString methods.
func (ph PasswordHash) GoString() string {
	return fmt.Sprintf("PasswordHash{inner:%q}", ph.inner)
}

func (ph PasswordHash) String() string {
	return ph.GoString()
}

// Bio represents a user's bio.
type Bio string

// ParseBio is a helper function compatible with [option.Map].
// It performs a simple type conversion from string to [Bio].
//
// # Errors
//   - nil
func ParseBio(raw string) (Bio, error) {
	return Bio(raw), nil
}

// URL represents a valid, immutable URL.
type URL struct {
	inner *neturl.URL
}

// ParseURL returns a valid [URL] if successful, and an error otherwise.
func ParseURL(candidate string) (URL, error) {
	u, err := neturl.Parse(candidate)
	if err != nil {
		return URL{}, NewInvalidURLError()
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

// GoString satisfies [fmt.GoStringer]. Must be implemented manually to ensure
// that User's password hash is not printed. The fmt package uses reflection to
// print unexported fields without invoking their String or GoString methods.
func (u User) GoString() string {
	return fmt.Sprintf("User{id:%q, username:%q, email:%q, passwordHash:%q, bio:%q, imageURL:%q}",
		u.id, u.username, u.email, u.passwordHash, u.bio.ValueOrZero(), u.imageURL.ValueOrZero())
}

func (u User) String() string {
	return u.GoString()
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

// ParseRegistrationRequest returns a new [RegistrationRequest] from raw inputs.
//
// # Errors
//   - [ValidationErrors], if one or more inputs are invalid.
//   - Unexpected internal response.
func ParseRegistrationRequest(
	rawUsername string,
	rawEmail string,
	rawPassword logfusc.Secret[string],
) (*RegistrationRequest, error) {
	var validationErrs ValidationErrors
	username, err := ParseUsername(rawUsername)
	if pushErr := validationErrs.PushValidationError(err); pushErr != nil {
		return nil, pushErr
	}

	email, err := ParseEmailAddress(rawEmail)
	if pushErr := validationErrs.PushValidationError(err); pushErr != nil {
		return nil, pushErr
	}

	passwordHash, err := ParsePassword(rawPassword)
	if pushErr := validationErrs.PushValidationError(err); pushErr != nil {
		return nil, pushErr
	}

	if validationErrs.Any() {
		return nil, validationErrs
	}

	return NewRegistrationRequest(username, email, passwordHash), nil
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

// GoString satisfies [fmt.GoStringer]. Must be implemented manually to ensure
// that the RegistrationRequest's password hash is not printed. The fmt package
// uses reflection to print unexported fields without invoking their String or
// GoString methods.
func (r RegistrationRequest) GoString() string {
	return fmt.Sprintf("RegistrationRequest{username:%q, email:%q, passwordHash:%q}",
		r.username, r.email, r.passwordHash)
}

func (r RegistrationRequest) String() string {
	return r.GoString()
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

// ParseAuthRequest returns a new [AuthRequest] from raw inputs.
//
// # Errors
//   - [ValidationError], if `rawEmail` is invalid.
func ParseAuthRequest(rawEmail string, passwordCandidate logfusc.Secret[string]) (*AuthRequest, error) {
	email, err := ParseEmailAddress(rawEmail)
	if err != nil {
		return nil, err
	}

	return NewAuthRequest(email, passwordCandidate), nil
}

func (ar *AuthRequest) Email() EmailAddress {
	return ar.email
}

func (ar *AuthRequest) PasswordCandidate() logfusc.Secret[string] {
	return ar.passwordCandidate
}

// GoString satisfies [fmt.GoStringer]. Must be implemented manually to ensure
// that the AuthRequest's password hash is not printed. The fmt package
// uses reflection to print unexported fields without invoking their String or
// GoString methods.
func (ar AuthRequest) GoString() string {
	return fmt.Sprintf("AuthRequest{email:%q, passwordCandidate:%q}",
		ar.email, ar.passwordCandidate)
}

func (ar AuthRequest) String() string {
	return ar.GoString()
}

// UpdateRequest describes the data required to update a user. All fields but
// userID are optional. To unset a FieldType, pass [Option] containing the zero
// value of the FieldType's type.
type UpdateRequest struct {
	userID       uuid.UUID
	email        option.Option[EmailAddress]
	bio          option.Option[Bio]
	imageURL     option.Option[URL]
	passwordHash option.Option[PasswordHash]
}

func NewUpdateRequest(
	userID uuid.UUID,
	email option.Option[EmailAddress],
	passwordHash option.Option[PasswordHash],
	bio option.Option[Bio],
	imageURL option.Option[URL],
) *UpdateRequest {
	return &UpdateRequest{
		userID:       userID,
		email:        email,
		passwordHash: passwordHash,
		bio:          bio,
		imageURL:     imageURL,
	}
}

// ParseUpdateRequest returns a new [UpdateRequest] from raw inputs.
//
// # Errors
//   - [ValidationErrors], if one or more inputs are invalid.
//   - Unexpected internal response.
func ParseUpdateRequest(
	userID uuid.UUID,
	rawEmail option.Option[string],
	rawPassword option.Option[logfusc.Secret[string]],
	rawBio option.Option[string],
	rawImageURL option.Option[string],
) (*UpdateRequest, error) {
	var validationErrs ValidationErrors
	email, err := option.Map(rawEmail, ParseEmailAddress)
	if pushErr := validationErrs.PushValidationError(err); pushErr != nil {
		return nil, pushErr
	}
	passwordHash, err := option.Map(rawPassword, ParsePassword)
	if pushErr := validationErrs.PushValidationError(err); pushErr != nil {
		return nil, pushErr
	}
	bio, err := option.Map(rawBio, ParseBio)
	if pushErr := validationErrs.PushValidationError(err); pushErr != nil {
		return nil, pushErr
	}
	imageURL, err := option.Map(rawImageURL, ParseURL)
	if pushErr := validationErrs.PushValidationError(err); pushErr != nil {
		return nil, pushErr
	}

	if validationErrs.Any() {
		return nil, validationErrs
	}

	return NewUpdateRequest(userID, email, passwordHash, bio, imageURL), nil
}

func (ur *UpdateRequest) UserID() uuid.UUID {
	return ur.userID
}

func (ur *UpdateRequest) Email() option.Option[EmailAddress] {
	return ur.email
}

func (ur *UpdateRequest) Bio() option.Option[Bio] {
	return ur.bio
}

func (ur *UpdateRequest) ImageURL() option.Option[URL] {
	return ur.imageURL
}

func (ur *UpdateRequest) PasswordHash() option.Option[PasswordHash] {
	return ur.passwordHash
}

// GoString satisfies [fmt.GoStringer]. Must be implemented manually to ensure
// that the UpdateRequest's password hash is not printed. The fmt package uses
// reflection to print unexported fields without invoking their String or
// GoString methods.
func (ur UpdateRequest) GoString() string {
	return fmt.Sprintf("UpdateRequest{userID:%q, email:%q, passwordHash:%q, imageURL:%q, bio:%q}",
		ur.userID, ur.email.ValueOrZero(), ur.passwordHash.ValueOrZero(), ur.bio.ValueOrZero(), ur.imageURL.ValueOrZero())
}

func (ur UpdateRequest) String() string {
	return ur.GoString()
}
