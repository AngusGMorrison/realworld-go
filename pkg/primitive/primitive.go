package primitive

import "net/mail"

// SensitiveString represents a string that should be not be printed (e.g. in
// logs) for security reasons.
type SensitiveString string

func (p SensitiveString) String() string {
	return "REDACTED"
}

// GoString satisfies `fmt.GoStringer`, which controls formatting in response to
// the `%#v` directive.
func (p SensitiveString) GoString() string {
	return p.String()
}

// EmailAddress is a dedicated string type for email addresses. It enhances the
// benefits of the type system, but performs no validation. The contained email
// address is not guaranteed to be valid.
type EmailAddress string

// IsRFC5322Valid validates that the email address conforms to RFC5332 standards
// apart from the minor divergences introduce by the Go standard library,
// documented in [net/mail].
//
// The formats that constitue a valid email address may surprise you. For
// example, single-value domain names like `angus@com` are valid.
func (ea EmailAddress) IsRFC5322Valid() bool {
	if _, err := mail.ParseAddress(string(ea)); err != nil {
		return false
	}
	return true
}
