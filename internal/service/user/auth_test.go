package user

//
//import (
//	"strings"
//	"testing"
//
//	"github.com/angusgmorrison/logfusc"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//	"golang.org/x/crypto/bcrypt"
//)
//
//func Test_NewPasswordCandidate(t *testing.T) {
//	t.Parallel()
//
//	password := "password"
//	expectedPC := PasswordCandidate{Secret: logfusc.NewSecret(password)}
//
//	gotPC := NewPasswordCandidate(password)
//
//	assert.Equal(t, expectedPC, gotPC)
//}
//
//func Test_PasswordCandidate_NonZero(t *testing.T) {
//	t.Parallel()
//
//	t.Run("when candidate is empty, it returns nil", func(t *testing.T) {
//		t.Parallel()
//
//		pc := PasswordCandidate{}
//
//		assert.Nil(t, pc.NonZero())
//	})
//
//	t.Run("when candidate is non-empty, it returns nil", func(t *testing.T) {
//		t.Parallel()
//
//		pc := NewPasswordCandidate("password")
//
//		assert.Nil(t, pc.NonZero())
//	})
//}
//
//func Test_ParsePassword(t *testing.T) {
//	t.Parallel()
//
//	t.Run("when password is valid, it returns a PasswordHash", func(t *testing.T) {
//		t.Parallel()
//
//		password := NewPasswordCandidate("password")
//
//		hash, err := ParsePassword(password)
//
//		assert.NoError(t, err)
//		assert.NoError(t, bcrypt.CompareHashAndPassword(hash.Expose(), []byte(password.Expose())))
//	})
//
//	t.Run("when password candidate is invalid", func(t *testing.T) {
//		t.Parallel()
//
//		testCases := []struct {
//			name     string
//			password PasswordCandidate
//			err      error
//		}{
//			{
//				name:     "when too too short, it returns ErrPasswordTooShort",
//				password: NewPasswordCandidate("short"),
//				err:      ErrPasswordTooShort,
//			},
//			{
//				name:     "when too too long, it returns ErrPasswordTooLong",
//				password: NewPasswordCandidate(strings.Repeat("a", PasswordMaxLen+1)),
//				err:      ErrPasswordTooLong,
//			},
//		}
//
//		for _, tc := range testCases {
//			tc := tc
//
//			t.Run(tc.name, func(t *testing.T) {
//				t.Parallel()
//
//				hash, err := ParsePassword(tc.password)
//
//				assert.Empty(t, hash)
//				assert.ErrorIs(t, err, tc.err)
//			})
//		}
//	})
//}
//
//func Test_PasswordHash_NonZero(t *testing.T) {
//	t.Parallel()
//
//	t.Run("when password hash is nil, it returns a ZeroValueError", func(t *testing.T) {
//		t.Parallel()
//
//		ph := PasswordHash{}
//
//		assertErrorAsZeroValueError(t, ph.NonZero())
//	})
//
//	t.Run("when password hash has len 0, it returns a ZeroValueError", func(t *testing.T) {
//		t.Parallel()
//
//		var b []byte
//		ph := PasswordHash{inner: logfusc.NewSecret(b)}
//
//		assertErrorAsZeroValueError(t, ph.NonZero())
//	})
//
//	t.Run("when password hash has len > 0, it returns nil", func(t *testing.T) {
//		t.Parallel()
//
//		ph := PasswordHash{inner: logfusc.NewSecret([]byte("password"))}
//
//		assert.NoError(t, ph.NonZero())
//	})
//}
//
//func Test_tryAuthenticate(t *testing.T) {
//	t.Parallel()
//
//	t.Run("when password is valid, it returns nil", func(t *testing.T) {
//		t.Parallel()
//
//		candidate := NewPasswordCandidate("password")
//		hash, err := ParsePassword(candidate)
//		require.NoError(t, err)
//
//		user := User{passwordHash: hash}
//
//		gotErr := tryAuthenticate(&user, candidate)
//
//		assert.Nil(t, gotErr)
//	})
//
//	t.Run("when password is invalid, it returns an AuthError", func(t *testing.T) {
//		t.Parallel()
//
//		hash, err := ParsePassword(NewPasswordCandidate("password"))
//		require.NoError(t, err)
//
//		user := User{passwordHash: hash}
//
//		gotErr := tryAuthenticate(&user, NewPasswordCandidate("wrong"))
//
//		var authErr *AuthError
//		assert.ErrorAs(t, gotErr, &authErr)
//	})
//}
