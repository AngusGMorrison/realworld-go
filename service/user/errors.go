package user

import "fmt"

// ValidationError specifies an error that returns a message describing how its
// subject failed validation. The message MUST be appropriate for display to an
// end user.
type ValidationError interface {
	ValidationError() string
}

// EmailAddressValidation error indicates a raw email address that fails to
// conform to RFC 5322 standards for email addresses.
type EmailAddressValidationError struct {
	rawEmailAddress string
	cause           error
}

func NewEmailAddressValidationError(rawEmail string, cause error) *EmailAddressValidationError {
	return &EmailAddressValidationError{
		rawEmailAddress: rawEmail,
		cause:           cause,
	}
}

func (e *EmailAddressValidationError) Error() string {
	return fmt.Sprintf("invalid email address %q: %s", e.rawEmailAddress, e.cause)
}

// ValidationError satisfies the ValidationError interface. This is deliberately
// decoupled from (*EmailAddressValidationError).Error() to allow Error() to
// change without affecting the message deemdd suitable for display to an end
// user.
func (e *EmailAddressValidationError) ValidationError() string {
	return fmt.Sprintf("invalid email address: %s", e.rawEmailAddress)
}

// AvatarURLValidationError indicates a raw avatar URL that could not be parsed
// by net/url.
type AvatarURLValidationError struct {
	rawAvatarURL string
	cause        error
}

func NewAvatarURLValidationError(rawAvatarURL string, cause error) *AvatarURLValidationError {
	return &AvatarURLValidationError{
		rawAvatarURL: rawAvatarURL,
		cause:        cause,
	}
}

func (e *AvatarURLValidationError) Error() string {
	return fmt.Sprintf("invalid avatar URL %q: %s", e.rawAvatarURL, e.cause)
}

// ValidationError satisfies the ValidationError interface.
func (e *AvatarURLValidationError) ValidationError() string {
	return fmt.Sprintf("invalid avatar URL: %s", e.rawAvatarURL)
}
