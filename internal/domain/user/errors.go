package user

import (
	"errors"
	"fmt"
	"strings"

	"github.com/angusgmorrison/realworld-go/pkg/etag"

	"github.com/google/uuid"
)

// FieldType identifies the type of field implicated in an error.
type FieldType int

const (
	UUIDFieldType FieldType = iota + 1
	UsernameFieldType
	EmailFieldType
	PasswordFieldType
	URLFieldType
)

var fieldNames = [5]string{"id", "username", "email", "password", "imageURL"}

func (f FieldType) String() string {
	if int(f) > len(fieldNames) {
		return "unknown"
	}
	return fieldNames[f-1]
}

// AuthError is a wrapper for an authentication error result, which may include
// errors that would be considered validation errors by other endpoints. This
// reinforces the security convention that an end user should not receive the
// specifics of why an authentication request failed.
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
	IDType  FieldType
	IDValue string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("user with %s %q not found", e.IDType, e.IDValue)
}

func (e *NotFoundError) Is(target error) bool {
	var otherNotFoundError *NotFoundError
	if !errors.As(target, &otherNotFoundError) {
		return false
	}

	return e.IDType == otherNotFoundError.IDType && e.IDValue == otherNotFoundError.IDValue
}

func NewNotFoundByIDError(id uuid.UUID) error {
	return &NotFoundError{
		IDType:  UUIDFieldType,
		IDValue: id.String(),
	}
}

func NewNotFoundByEmailError(email EmailAddress) error {
	return &NotFoundError{
		IDType:  EmailFieldType,
		IDValue: email.String(),
	}
}

// ValidationErrors is a collection of [ValidationError], allowing us to report
// on the validation of several inputs rather than returning the first validation
// error encountered.
type ValidationErrors []*ValidationError

// PushValidationError adds a [ValidationError] to the collection. If the error
// is not a ValidationError (including nil response), it is returned as-is. The
// supports the following pattern for successive validations:
//
//		var validationErrs ValidationErrors
//		field1, err := parseField1()
//		if err := validationErrs.PushValidationError(err); err != nil {
//	     	return nil, err // not a ValidationError
//		}
//	 	field2, err := parseField2()
//	 	...
func (e *ValidationErrors) PushValidationError(err error) error {
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		*e = append(*e, validationErr)
		return nil
	}
	return err
}

// Any returns true if the collection contains any response, and false otherwise.
func (e ValidationErrors) Any() bool {
	return len(e) > 0
}

func (e ValidationErrors) Error() string {
	var builder strings.Builder
	builder.WriteString("validation errors:\n")
	for _, err := range e {
		builder.WriteString(fmt.Sprintf("\t- %s\n", err.Error()))
	}
	return builder.String()
}

// ValidationError represents a single error encountered when validation a field
// of the given [FieldType].
type ValidationError struct {
	Field   FieldType
	Message string
}

func (e *ValidationError) Is(target error) bool {
	var otherValidationErr *ValidationError
	if !errors.As(target, &otherValidationErr) {
		return false
	}

	return e.Field == otherValidationErr.Field && e.Message == otherValidationErr.Message
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("{Field: %q, Message: %q}", e.Field, e.Message)
}

func NewUsernameTooShortError() error {
	return &ValidationError{
		Field:   UsernameFieldType,
		Message: fmt.Sprintf("must be at least %d characters long", UsernameMinLen),
	}
}

func NewUsernameTooLongError() error {
	return &ValidationError{
		Field:   UsernameFieldType,
		Message: fmt.Sprintf("must be at most %d characters long", UsernameMaxLen),
	}
}

func NewUsernameFormatError() error {
	return &ValidationError{
		Field:   UsernameFieldType,
		Message: fmt.Sprintf("must match %q", usernamePattern),
	}
}

func NewDuplicateUsernameError(username Username) error {
	return &ValidationError{
		Field:   UsernameFieldType,
		Message: fmt.Sprintf("%q is already registered", username),
	}
}

func NewEmailAddressFormatError(candidate string) error {
	return &ValidationError{
		Field:   EmailFieldType,
		Message: fmt.Sprintf("%q is not a valid email address", candidate),
	}
}

func NewDuplicateEmailError(email EmailAddress) error {
	return &ValidationError{
		Field:   EmailFieldType,
		Message: fmt.Sprintf("%q is already registered", email),
	}
}

func NewPasswordTooShortError() error {
	return &ValidationError{
		Field:   PasswordFieldType,
		Message: fmt.Sprintf("must be at least %d bytes long", PasswordMinLen),
	}
}

func NewPasswordTooLongError() error {
	return &ValidationError{
		Field:   PasswordFieldType,
		Message: fmt.Sprintf("must be at most %d bytes long", PasswordMaxLen),
	}
}

func NewInvalidURLError() error {
	return &ValidationError{
		Field:   URLFieldType,
		Message: "must be a valid URL",
	}
}

// ConcurrentModificationError indicates that a resource was modified prior to
// processing the current request, making the current request stale.
type ConcurrentModificationError struct {
	ID   uuid.UUID
	ETag etag.ETag
}

func (e *ConcurrentModificationError) Error() string {
	return fmt.Sprintf("user %q with ETag %q was modified since last read", e.ID, e.ETag)
}
