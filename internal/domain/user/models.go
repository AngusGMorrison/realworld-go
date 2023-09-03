package user

import (
	"fmt"
	"net/mail"
	neturl "net/url"
	"regexp"

	"github.com/angusgmorrison/logfusc"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/angusgmorrison/realworld-go/pkg/option"
)

// EmailAddress is a dedicated usernameCandidate type for valid email addresses. New
// instances are validated for RFC5332 compliance.
type EmailAddress struct {
	raw string
}

// ParseEmailAddress returns a new email address from `candidate`, validating
// that the email address conforms to RFC5332 standards (with the minor
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
	// UsernameMinLen is the minimum length of a username in bytes.
	UsernameMinLen = 3

	// UsernameMaxLen is the maximum length of a username in bytes.
	UsernameMaxLen = 16

	UsernamePatternTemplate = "^[a-zA-Z0-9_]{%d,%d}$"
)

var (
	usernamePattern = fmt.Sprintf(UsernamePatternTemplate, UsernameMinLen, UsernameMaxLen)
	usernameRegex   = regexp.MustCompile(fmt.Sprintf(UsernamePatternTemplate, UsernameMinLen, UsernameMaxLen))
)

// Username represents a valid Username.
type Username struct {
	raw string
}

// ParseUsername returns either a valid [Username] or an error indicating why
// the raw username was invalid.
func ParseUsername(candidate string) (Username, error) {
	if len(candidate) < UsernameMinLen {
		return Username{}, NewUsernameTooShortError()
	}
	if len(candidate) > UsernameMaxLen {
		return Username{}, NewUsernameTooLongError()
	}
	if !usernameRegex.MatchString(candidate) {
		return Username{}, NewUsernameFormatError()
	}

	return Username{raw: candidate}, nil
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
	return parsePassword(candidate, bcryptHasher)
}

func parsePassword(candidate logfusc.Secret[string], hasher passwordHasher) (PasswordHash, error) {
	if err := validatePasswordCandidate(candidate); err != nil {
		return PasswordHash{}, err
	}

	hash, err := hasher(candidate)
	if err != nil {
		return PasswordHash{}, err
	}

	return hash, nil
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
	return fmt.Sprintf("PasswordHash{inner:%s}", ph.inner)
}

func (ph PasswordHash) String() string {
	return ph.GoString()
}

type passwordHasher func(candidate logfusc.Secret[string]) (PasswordHash, error)

func bcryptHasher(candidate logfusc.Secret[string]) (PasswordHash, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(candidate.Expose()), bcrypt.DefaultCost)
	if err != nil {
		return PasswordHash{}, fmt.Errorf("hash password: %w", err)
	}
	return NewPasswordHashFromTrustedSource(logfusc.NewSecret(hash)), nil
}

type passwordComparator func(hash PasswordHash, candidate logfusc.Secret[string]) *AuthError

// CompareHashAndPassword compares a hashed password with its possible plaintext
// equivalent. Returns nil on success, or an [AuthError] on failure.
//
// This function is exposed as a test utility, since it is often necessary to
// compare two domain models, but direct comparison of password hashes is
// impossible. CompareHashAndPassword allows dependent packages to remain
// ignorant of the hashing algorithm.
func CompareHashAndPassword(hash PasswordHash, candidate logfusc.Secret[string]) error {
	// An explicit nil check is required to return a nil error interface, as opposed
	// to non-nil error containing nil *AuthError.
	if authErr := bcryptComparator(hash, candidate); authErr != nil {
		return authErr
	}
	return nil
}

func bcryptComparator(hash PasswordHash, candidate logfusc.Secret[string]) *AuthError {
	if err := bcrypt.CompareHashAndPassword(hash.Expose(), []byte(candidate.Expose())); err != nil {
		return &AuthError{Cause: err}
	}
	return nil
}

// Bio represents a user's bio.
type Bio string

// ParseBio is a convenience function for use with option.Map.
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

// GoString satisfies [fmt.GoStringer]. Must be implemented manually to ensure
// that User's password hash is not printed. The fmt package uses reflection to
// print unexported fields without invoking their String or GoString methods.
func (u User) GoString() string {
	return fmt.Sprintf("User{id:%v, username:%q, email:%q, passwordHash:%s, bio:%q, imageURL:%q}",
		u.id, u.username, u.email, u.passwordHash, u.bio.UnwrapOrZero(), u.imageURL.UnwrapOrZero())
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
	usernameCandidate string,
	emailCandidate string,
	passwordCandidate logfusc.Secret[string],
) (*RegistrationRequest, error) {
	var validationErrs ValidationErrors
	username, err := ParseUsername(usernameCandidate)
	if pushErr := validationErrs.PushValidationError(err); pushErr != nil {
		return nil, pushErr
	}

	email, err := ParseEmailAddress(emailCandidate)
	if pushErr := validationErrs.PushValidationError(err); pushErr != nil {
		return nil, pushErr
	}

	passwordHash, err := ParsePassword(passwordCandidate)
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

func (r *RegistrationRequest) Email() EmailAddress {
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
	return fmt.Sprintf("RegistrationRequest{username:%q, email:%q, passwordHash:%s}",
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
//   - [ValidationErrors], if `emailCandidate` is invalid.
func ParseAuthRequest(emailCandidate string, passwordCandidate logfusc.Secret[string]) (*AuthRequest, error) {
	var validationErrs ValidationErrors
	email, err := ParseEmailAddress(emailCandidate)
	if pushErr := validationErrs.PushValidationError(err); pushErr != nil {
		return nil, pushErr
	}

	if validationErrs.Any() {
		return nil, validationErrs
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
	return fmt.Sprintf("AuthRequest{email:%q, passwordCandidate:%s}",
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
	emailCandidate option.Option[string],
	passwordCandidate option.Option[logfusc.Secret[string]],
	rawBio option.Option[string],
	imageURLCandidate option.Option[string],
) (*UpdateRequest, error) {
	var validationErrs ValidationErrors
	email, err := option.Map(emailCandidate, ParseEmailAddress)
	if pushErr := validationErrs.PushValidationError(err); pushErr != nil {
		return nil, pushErr
	}
	passwordHash, err := option.Map(passwordCandidate, ParsePassword)
	if pushErr := validationErrs.PushValidationError(err); pushErr != nil {
		return nil, pushErr
	}
	bio, err := option.Map(rawBio, ParseBio)
	if pushErr := validationErrs.PushValidationError(err); pushErr != nil {
		return nil, pushErr
	}
	imageURL, err := option.Map(imageURLCandidate, ParseURL)
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
	return fmt.Sprintf("UpdateRequest{userID:%q, email:%q, passwordHash:%s, bio:%q, imageURL:%q}",
		ur.userID, ur.email, ur.passwordHash, ur.bio, ur.imageURL)
}

func (ur UpdateRequest) String() string {
	return ur.GoString()
}
