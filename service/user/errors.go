package user

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrEmailAddressUnparseable = errors.New("email address is not RFC 5322-compliant")
var ErrAvatarURLUnparseable = errors.New("avatar URL could not be parsed")
var ErrPasswordTooLong = bcrypt.ErrPasswordTooLong
