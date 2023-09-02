package testutil

import (
	"github.com/angusgmorrison/logfusc"
	"github.com/angusgmorrison/realworld-go/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"testing"
)

func NewRegistrationRequestMatcher(t *testing.T, want *user.RegistrationRequest, wantPassword logfusc.Secret[string]) func(*user.RegistrationRequest) bool {
	return func(got *user.RegistrationRequest) bool {
		t.Helper()

		err := user.CompareHashAndPassword(got.PasswordHash(), wantPassword)
		if pass := assert.NoError(t, err); !pass {
			return false
		}

		if pass := assert.Equal(t, want.Username(), got.Username()); !pass {
			return false
		}

		if pass := assert.Equal(t, want.Email(), got.Email()); !pass {
			return false
		}

		return true
	}
}
