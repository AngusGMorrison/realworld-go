package testutil

import (
	"fmt"
	"github.com/angusgmorrison/logfusc"
	"github.com/angusgmorrison/realworld-go/internal/domain/user"
)

func NewRegistrationRequestMatcher(want *user.RegistrationRequest, wantPassword logfusc.Secret[string]) func(*user.RegistrationRequest) bool {
	return func(got *user.RegistrationRequest) bool {
		if err := user.CompareHashAndPassword(got.PasswordHash(), wantPassword); err != nil {
			fmt.Println(err)
			return false
		}

		return got.Username() == want.Username() &&
			got.Email() == want.Email()
	}
}
