package user

import (
	"fmt"
	"github.com/google/uuid"
)

// Field identifies the field of a user referenced in an error.
type Field int

const (
	IDField Field = iota + 1
	UsernameField
	EmailField
	PasswordField
)

var fieldNames = [4]string{"id", "username", "email", "password"}

func (f Field) String() string {
	return fieldNames[f-1]
}

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
	IDField Field
	ID      string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("user with %s %q not found", e.IDField, e.ID)
}

func NewNotFoundByIDError(id uuid.UUID) error {
	return &NotFoundError{
		IDField: IDField,
		ID:      id.String(),
	}
}

func NewNotFoundByEmailError(email EmailAddress) error {
	return &NotFoundError{
		IDField: EmailField,
		ID:      email.String(),
	}
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

func NewDuplicateEmailError(email EmailAddress) error {
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
