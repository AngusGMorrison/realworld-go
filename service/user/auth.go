package user

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// TODO: implement password hashing
func digest(password string) (PasswordDigest, error) {
	digestBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return "", ErrPasswordTooLong
		}
		return "", err
	}

	return PasswordDigest(digestBytes), nil
}
