package testutil

import (
	"github.com/angusgmorrison/realworld-go/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"testing"
)

func NewUserRegistrationRequestMatcher(
	t *testing.T,
	want *user.RegistrationRequest,
	wantPassword string,
) func(*user.RegistrationRequest) bool {
	t.Helper()

	return func(got *user.RegistrationRequest) bool {
		t.Helper()

		if pass := assert.True(t, want.Equal(got, wantPassword), "want %#v, got %#v", want, got); !pass {
			return false
		}

		return true
	}
}

func NewUserUpdateRequestMatcher(
	t *testing.T,
	want *user.UpdateRequest,
	wantPassword string,
) func(*user.UpdateRequest) bool {
	t.Helper()

	return func(got *user.UpdateRequest) bool {
		t.Helper()

		if pass := assert.True(t, want.Equal(got, wantPassword), "want %#v, got %#v", want, got); !pass {
			return false
		}

		return true
	}
}
