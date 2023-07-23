package user

import (
	"fmt"
	"github.com/angusgmorrison/realworld/pkg/tidy"
	"github.com/google/uuid"
)

// AuthError is a wrapper for authentication errors, which may include errors
// that would otherwise be considered validation errors. This reinforces the
// security convention that an end user should not receive the specifics of why
// an authentication request failed.
type AuthError struct {
	Cause error
}

func (e *AuthError) Error() string {
	return "unauthorized"
}

func (e *AuthError) Unwrap() error {
	return e.Cause
}

// NotFoundError should be returned by a [Repository] when the specified user
// does not exist.
type NotFoundError struct {
	ID uuid.UUID
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("user %s not found", e.ID)
}

type Field int

const (
	UsernameField Field = iota + 1
	EmailField
	PasswordField
)

var fieldNames = [3]string{"username", "email", "password"}

func (f Field) String() string {
	return fieldNames[f-1]
}

type ValidationError struct {
	Field   Field
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func NewUsernameTooShortError() error {
	return &ValidationError{
		Field:   UsernameField,
		Message: fmt.Sprintf("must be at least %d characters long", UsernameMinLen),
	}
}

func NewUsernameTooLongError() error {
	return &ValidationError{
		Field:   UsernameField,
		Message: fmt.Sprintf("must be at most %d characters long", UsernameMaxLen),
	}
}

func NewUsernameFormatError() error {
	return &ValidationError{
		Field:   UsernameField,
		Message: fmt.Sprintf("must match %q", UsernamePattern),
	}
}

func NewDuplicateUsernameError(username Username) error {
	return &ValidationError{
		Field:   UsernameField,
		Message: fmt.Sprintf("%q is already registered", username),
	}
}

func NewEmailAddressFormatError(candidate string) error {
	return &ValidationError{
		Field:   EmailField,
		Message: fmt.Sprintf("%q is not a valid email address", candidate),
	}
}

func NewDuplicateEmailError(email tidy.EmailAddress) error {
	return &ValidationError{
		Field:   EmailField,
		Message: fmt.Sprintf("%q is already registered", email),
	}
}

func NewPasswordTooShortError() error {
	return &ValidationError{
		Field:   PasswordField,
		Message: fmt.Sprintf("must be at least %d bytes long", PasswordMinLen),
	}
}

func NewPasswordTooLongError() error {
	return &ValidationError{
		Field:   PasswordField,
		Message: fmt.Sprintf("must be at most %d bytes long", PasswordMaxLen),
	}
}
