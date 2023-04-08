package user

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// ValidationError is a wrapper for zero or more errors that caused a model to
// fail validation. Combining these errors under a single umbrella type makes it
// easy for ingress handlers to detect errors caused by invalid input and
// respond accordingly.
type ValidationError struct {
	fieldErrors []error
}

// Error composes a string representation of the underlying field errors.
func (e *ValidationError) Error() string {
	if len(e.fieldErrors) == 0 {
		return "{}"
	}

	var builder strings.Builder
	builder.WriteString("{\n")
	for _, err := range e.fieldErrors {
		fmt.Fprintf(&builder, "\t%s,\n", err)
	}
	builder.WriteByte('}')
	return builder.String()
}

// FieldErrors returns the underlying errors that caused validation to fail.
func (e *ValidationError) FieldErrors() []error {
	return e.fieldErrors
}

func (e *ValidationError) Push(errs ...error) {
	e.fieldErrors = append(e.fieldErrors, errs...)
}

func (e *ValidationError) Any() bool {
	return len(e.fieldErrors) > 0
}

// AuthError is a wrapper for authentication errors, which may include errors
// that would otherwise be considered validation errors. This reinforces the
// security convention that an end user should not receive the specifics of why
// an authentication request failed.
type AuthError struct {
	cause error
}

func (e *AuthError) Error() string {
	return "unauthorized"
}

func (e *AuthError) Unwrap() error {
	return e.cause
}

// Image URL errors.
var ErrImageURLUnparseable = errors.New("image URL could not be parsed")

// Email address errors.
var ErrEmailAddressUnparseable = errors.New("email address is not RFC 5322-compliant")

// Password errors.
var ErrPasswordTooLong = bcrypt.ErrPasswordTooLong
var ErrPasswordMismatch = errors.New("password did not match stored hash")

// User errors.
var ErrUserNotFound = errors.New("user not found")
