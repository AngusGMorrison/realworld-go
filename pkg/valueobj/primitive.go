package valueobj

import (
	"fmt"
	"net/mail"
)

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

// String returns the raw email address.
func (e EmailAddress) String() string {
	return e.raw
}
