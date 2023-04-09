package user

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// RequiredValidatingPassword tags required passwords for validation. It should
// be embedded wherever a password MUST be present and MUST be validated before
// use.
//
// Changes to password validation tags MUST be kept in sync with
// OptionalValidatingPassword.
type RequiredValidatingPassword struct {
	Password string `json:"password" validate:"required,pw_min,pw_max"`
}

// Hash returns the hashed password.
func (rvp RequiredValidatingPassword) Hash() (string, error) {
	return bcryptHash(rvp.Password)
}

// OptionalValidatingPassword tags optional passwords for validation. It should
// be embedded wherever a password MUST be validated IFF present.
//
// Changes to password validation tags MUST be kept in sync with
// RequiredValidatingPassword.
type OptionalValidatingPassword struct {
	Password string `json:"password" validate:"omitempty,pw_min,pw_max"`
}

// Hash returns the hashed password.
func (ovp OptionalValidatingPassword) Hash() (string, error) {
	return bcryptHash(ovp.Password)
}

func bcryptHash(password string) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return string(hashBytes), nil
}
