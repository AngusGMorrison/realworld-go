package testutil

import (
	"github.com/angusgmorrison/logfusc"
	"github.com/angusgmorrison/realworld-go/internal/domain/user"
	"github.com/angusgmorrison/realworld-go/pkg/option"
	"github.com/stretchr/testify/assert"
	"testing"
)

func NewUserRegistrationRequestMatcher(
	t *testing.T,
	want *user.RegistrationRequest,
	wantPassword logfusc.Secret[string],
) func(*user.RegistrationRequest) bool {
	t.Helper()

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

func NewUserUpdateRequestMatcher(
	t *testing.T,
	want *user.UpdateRequest,
	wantPassword option.Option[logfusc.Secret[string]],
) func(*user.UpdateRequest) bool {
	t.Helper()

	return func(got *user.UpdateRequest) bool {
		t.Helper()

		if want.PasswordHash().IsSome() {
			err := user.CompareHashAndPassword(got.PasswordHash().UnwrapOrZero(), wantPassword.UnwrapOrZero())
			if pass := assert.NoError(t, err, "the received PasswordHash did not match the expected candidate"); !pass {
				return false
			}
		} else {
			if pass := assert.Truef(t, !got.PasswordHash().IsSome(), "want None[PasswordHash], but got Some"); !pass {
				return false
			}
		}

		if pass := assert.Equal(t, want.Email(), got.Email()); !pass {
			return false
		}

		if pass := assert.Equal(t, want.Bio(), got.Bio()); !pass {
			return false
		}

		if pass := assert.Equal(t, want.ImageURL(), got.ImageURL()); !pass {
			return false
		}

		return true
	}
}
