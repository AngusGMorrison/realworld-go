package tidy

import (
	"fmt"
	"net/mail"

	"github.com/google/uuid"
)

// UUIDv4 provides a minimal UUIDv4 [Strict] type.
type UUIDv4 struct {
	inner uuid.UUID
}

// NewUUIDv4 generates a new UUIDv4 or panics.
func NewUUIDv4() UUIDv4 {
	return UUIDv4{inner: uuid.New()}
}

// ParseUUIDv4 parses a UUIDv4 from `candidate`, returning an error if the
// candidate is not a valid UUIDv4.
func ParseUUIDv4(candidate string) (UUIDv4, error) {
	u, err := uuid.Parse(candidate)
	if err != nil {
		return UUIDv4{}, fmt.Errorf("parse candidate UUID %q: %w", candidate, err)
	}
	if u.Version() != uuid.Version(4) {
		return UUIDv4{}, fmt.Errorf("parse candidate UUID %q: expected a UUIDv4, got v%d", candidate, u.Version())
	}

	return UUIDv4{inner: u}, nil
}

// MustParseUUIDv4 attempts to parse a UUIDv4, and panics on failure.
func MustParseUUIDv4(candidate string) UUIDv4 {
	u, err := ParseUUIDv4(candidate)
	if err != nil {
		panic(err)
	}
	return u
}

func (u UUIDv4) String() string {
	return u.inner.String()
}

// NonZero satisfies [Strict] type.
func (u UUIDv4) NonZero() error {
	if u == (UUIDv4{}) {
		return &ZeroValueError{ZeroStrict: u}
	}
	return nil
}

// EmailAddress is a dedicated string type for valid email addresses. New instances
// are validated for RFC5332 compliance.
type EmailAddress struct {
	raw string
}

// ParseEmailAddressError is returned by [ParseEmailAddress] when the candidate
// email address cannot be parsed.
type ParseEmailAddressError struct {
	candidate string
	cause     error
}

func (e *ParseEmailAddressError) Error() string {
	return fmt.Sprintf("parse email address %q: %v", e.candidate, e.cause)
}

func (e *ParseEmailAddressError) Unwrap() error {
	return e.cause
}

// ParseEmailAddress returns a new email address from `candidate`, validating that
// the email address conforms to RFC5332 standards (with the minor
// divergences introduce by the Go standard library, documented in [net/mail]).
//
// The formats that constitue a valid email address may surprise you. For
// example, single-value domain names like `angus@com` are valid.
func ParseEmailAddress(candidate string) (EmailAddress, error) {
	if _, err := mail.ParseAddress(candidate); err != nil {
		return EmailAddress{}, &ParseEmailAddressError{candidate: candidate, cause: err}
	}

	return EmailAddress{raw: candidate}, nil
}

// NonZero returns [ZeroValueError] if the email address is empty.
func (ea EmailAddress) NonZero() error {
	if ea == (EmailAddress{}) {
		return &ZeroValueError{ZeroStrict: ea}
	}
	return nil
}

// String returns the raw email address.
func (ea EmailAddress) String() string {
	return ea.raw
}
