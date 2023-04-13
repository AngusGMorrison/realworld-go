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

// ErrPasswordMismatch may be given as the cause of an AuthError when the
// password provided with an AuthRequest does not match the password stored for
// the user.
var ErrPasswordMismatch = errors.New("password mismatch")
var ErrUserNotFound = errors.New("user not found")
var ErrEmailRegistered = errors.New("email address is already registered")
var ErrUsernameTaken = errors.New("username is taken")
