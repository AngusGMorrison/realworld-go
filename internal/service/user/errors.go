package user

import (
	"errors"
)

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

var ErrUserNotFound = errors.New("user not found")
var ErrUserExists = errors.New("user already exists")
