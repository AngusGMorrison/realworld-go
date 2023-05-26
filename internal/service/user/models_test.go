package user

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/angusgmorrison/realworld/pkg/tidy"
	"github.com/stretchr/testify/assert"
)

var (
	defaultEmail, _ = tidy.ParseEmailAddress("test@test.com")
	defaultUserID   = tidy.NewUUIDv4()
	defaultUsername = Username{raw: "johndoe"}
)

func Test_ParseUsername(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     string
		want      Username
		assertErr assert.ErrorAssertionFunc
	}{
		{
			name:  "valid username",
			input: "johndoe",
			want:  Username{raw: "johndoe"},
			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.NoError(t, err)
			},
		},
		{
			name:  "too short",
			input: strings.Repeat("a", minUsernameLen-1),
			want:  Username{},
			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return errors.Is(err, ErrUsernameTooShort)
			},
		},
		{
			name:  "too long",
			input: strings.Repeat("a", maxUsernameLen+1),
			want:  Username{},
			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return errors.Is(err, ErrUsernameTooLong)
			},
		},
		{
			name:  "invalid format",
			input: "john doe",
			want:  Username{},
			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return errors.Is(err, ErrUsernameFormat)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseUsername(tc.input)

			assert.Equal(t, tc.want, got)
			tc.assertErr(t, err)
		})
	}
}

func Test_Username_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		u    Username
		want string
	}{
		{
			name: "valid username",
			u:    defaultUsername,
			want: "johndoe",
		},
		{
			name: "empty username",
			u:    Username{},
			want: "",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.u.String()

			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_Username_NonZero(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		u    Username
		want error
	}{
		{
			name: "valid username",
			u:    defaultUsername,
			want: nil,
		},
		{
			name: "empty username",
			u:    Username{},
			want: &tidy.ZeroValueError{ZeroStrict: Username{}},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.u.NonZero()

			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_Bio_NonZero(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		bio  Bio
		want error
	}{
		{
			name: "valid Bio",
			bio:  "some bio",
			want: nil,
		},
		{
			name: "empty Bio",
			bio:  "",
			want: &tidy.ZeroValueError{ZeroStrict: Bio("")},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.bio.NonZero()

			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_URL_NonZero(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		u    URL
		want error
	}{
		{
			name: "valid URL",
			u:    URL{inner: &url.URL{Scheme: "https", Host: "example.com"}},
			want: nil,
		},
		{
			name: "empty URL",
			u:    URL{},
			want: &tidy.ZeroValueError{ZeroStrict: URL{}},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.u.NonZero()

			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_NewUser(t *testing.T) {
	t.Parallel()

	id := tidy.NewUUIDv4()
	passwordHash, _ := ParsePassword(NewPasswordCandidate("password"))
	bioSome, _ := tidy.Some(Bio("some bio"))
	bioNone := tidy.None[Bio]()
	imageURLSome, _ := tidy.Some(URL{inner: &url.URL{Scheme: "https", Host: "example.com"}})
	imageURLNone := tidy.None[URL]()

	t.Run("when all required arguments are present", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name         string
			id           tidy.UUIDv4
			username     Username
			email        tidy.EmailAddress
			passwordHash PasswordHash
			bio          tidy.Option[Bio]
			imageURL     tidy.Option[URL]
		}{
			{
				name:         "when all optional arguments are present, it returns a User",
				id:           id,
				username:     defaultUsername,
				email:        defaultEmail,
				passwordHash: passwordHash,
				bio:          bioSome,
				imageURL:     imageURLSome,
			},
			{
				name:         "when bio is None, it returns a User",
				id:           id,
				username:     defaultUsername,
				email:        defaultEmail,
				passwordHash: passwordHash,
				bio:          bioNone,
				imageURL:     imageURLSome,
			},
			{
				name:         "when imageURL is None, it returns a User",
				id:           id,
				username:     defaultUsername,
				email:        defaultEmail,
				passwordHash: passwordHash,
				bio:          bioSome,
				imageURL:     imageURLNone,
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				expectedUser := User{
					id:           tc.id,
					username:     tc.username,
					email:        tc.email,
					passwordHash: tc.passwordHash,
					bio:          tc.bio,
					imageURL:     tc.imageURL,
				}

				gotUser, err := NewUser(tc.id, tc.username, tc.email, tc.passwordHash, tc.bio, tc.imageURL)

				assert.NoError(t, err)
				assert.Equal(t, expectedUser, gotUser)
			})
		}
	})

	t.Run("when required arguments are empty", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name         string
			id           tidy.UUIDv4
			username     Username
			email        tidy.EmailAddress
			passwordHash PasswordHash
		}{
			{
				name:         "when id is empty, it returns a ZeroValueError",
				id:           tidy.UUIDv4{},
				username:     defaultUsername,
				email:        defaultEmail,
				passwordHash: passwordHash,
			},
			{
				name:         "when username is empty, it returns a ZeroValueError",
				id:           id,
				username:     Username{},
				email:        defaultEmail,
				passwordHash: passwordHash,
			},
			{
				name:         "when email is empty, it returns a ZeroValueError",
				id:           id,
				username:     defaultUsername,
				email:        tidy.EmailAddress{},
				passwordHash: passwordHash,
			},
			{
				name:         "when passwordHash is empty, it returns a ZeroValueError",
				id:           id,
				username:     defaultUsername,
				email:        defaultEmail,
				passwordHash: PasswordHash{},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				gotUser, gotErr := NewUser(tc.id, tc.username, tc.email, tc.passwordHash, tidy.None[Bio](), tidy.None[URL]())

				var zeroValueErr *tidy.ZeroValueError
				assert.ErrorAs(t, gotErr, &zeroValueErr)
				assert.Empty(t, gotUser)
			})
		}
	})
}

func Test_NewRegistrationRequest(t *testing.T) {
	t.Parallel()

	passwordHash, _ := ParsePassword(NewPasswordCandidate("password"))

	t.Run("when all arguments are non-zero, it returns a RegistrationRequest", func(t *testing.T) {
		t.Parallel()

		expectedReq := RegistrationRequest{
			username:     defaultUsername,
			email:        defaultEmail,
			passwordHash: passwordHash,
		}

		gotReq, gotErr := NewRegistrationRequest(defaultUsername, defaultEmail, passwordHash)

		assert.NoError(t, gotErr)
		assert.Equal(t, expectedReq, gotReq)
	})

	t.Run("when arguments are empty", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name         string
			username     Username
			email        tidy.EmailAddress
			passwordHash PasswordHash
		}{
			{
				name:         "when username is empty, it returns a ZeroValueError",
				username:     Username{},
				email:        defaultEmail,
				passwordHash: passwordHash,
			},
			{
				name:         "when email is empty, it returns a ZeroValueError",
				username:     defaultUsername,
				email:        tidy.EmailAddress{},
				passwordHash: passwordHash,
			},
			{
				name:         "when passwordHash is empty, it returns a ZeroValueError",
				username:     defaultUsername,
				email:        defaultEmail,
				passwordHash: PasswordHash{},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				gotReq, gotErr := NewRegistrationRequest(tc.username, tc.email, tc.passwordHash)

				var zeroValueErr *tidy.ZeroValueError
				assert.ErrorAs(t, gotErr, &zeroValueErr)
				assert.Empty(t, gotReq)
			})
		}
	})
}

func Test_RegistrationRequest_NonZero(t *testing.T) {
	t.Parallel()

	passwordHash, _ := ParsePassword(NewPasswordCandidate("password"))

	testCases := []struct {
		name      string
		req       *RegistrationRequest
		assertErr assert.ErrorAssertionFunc
	}{
		{
			name:      "when req is nil, it returns a ZeroValueError",
			req:       nil,
			assertErr: assertErrorAsZeroValueError,
		},
		{
			name: "when username is empty, it returns a ZeroValueError",
			req: &RegistrationRequest{
				username:     Username{},
				email:        defaultEmail,
				passwordHash: passwordHash,
			},
			assertErr: assertErrorAsZeroValueError,
		},
		{
			name: "when email is empty, it returns a ZeroValueError",
			req: &RegistrationRequest{
				username:     defaultUsername,
				email:        tidy.EmailAddress{},
				passwordHash: passwordHash,
			},
			assertErr: assertErrorAsZeroValueError,
		},
		{
			name: "when passwordHash is empty, it returns a ZeroValueError",
			req: &RegistrationRequest{
				username:     defaultUsername,
				email:        defaultEmail,
				passwordHash: PasswordHash{},
			},
			assertErr: assertErrorAsZeroValueError,
		},
		{
			name: "when all fields are non-zero, it returns nil",
			req: &RegistrationRequest{
				username:     defaultUsername,
				email:        defaultEmail,
				passwordHash: passwordHash,
			},
			assertErr: assert.NoError,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotErr := tc.req.NonZero()

			tc.assertErr(t, gotErr)
		})
	}
}

func Test_NewAuthRequest(t *testing.T) {
	t.Parallel()

	t.Run("when email is non-zero, it returns an AuthRequest", func(t *testing.T) {
		t.Parallel()

		expectedReq := AuthRequest{
			email: defaultEmail,
		}

		gotReq, gotErr := NewAuthRequest(defaultEmail, PasswordCandidate{})

		assert.NoError(t, gotErr)
		assert.Equal(t, expectedReq, gotReq)
	})

	t.Run("when email is empty, it returns a ZeroValueError", func(t *testing.T) {
		t.Parallel()

		gotReq, gotErr := NewAuthRequest(tidy.EmailAddress{}, PasswordCandidate{})

		var zeroValueErr *tidy.ZeroValueError
		assert.ErrorAs(t, gotErr, &zeroValueErr)
		assert.Empty(t, gotReq)
	})
}

func Test_AuthRequest_NonZero(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		req       *AuthRequest
		assertErr assert.ErrorAssertionFunc
	}{
		{
			name:      "when req is nil, it returns a ZeroValueError",
			req:       nil,
			assertErr: assertErrorAsZeroValueError,
		},
		{
			name: "when email is empty, it returns a ZeroValueError",
			req: &AuthRequest{
				email: tidy.EmailAddress{},
			},
			assertErr: assertErrorAsZeroValueError,
		},
		{
			name: "when email is non-zero, it returns nil",
			req: &AuthRequest{
				email: defaultEmail,
			},
			assertErr: assert.NoError,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotErr := tc.req.NonZero()

			tc.assertErr(t, gotErr)
		})
	}
}

func Test_NewUpdateRequest(t *testing.T) {
	t.Parallel()

	t.Run("when userID is non-zero, it returns an UpdateRequest", func(t *testing.T) {
		t.Parallel()

		expectedReq := UpdateRequest{
			userID:   defaultUserID,
			email:    tidy.None[tidy.EmailAddress](),
			bio:      tidy.None[Bio](),
			imageURL: tidy.None[URL](),
			pwHash:   tidy.None[PasswordHash](),
		}

		gotReq, gotErr := NewUpdateRequest(
			defaultUserID,
			tidy.None[tidy.EmailAddress](),
			tidy.None[Bio](),
			tidy.None[URL](),
			tidy.None[PasswordHash](),
		)

		assert.NoError(t, gotErr)
		assert.Equal(t, expectedReq, gotReq)
	})

	t.Run("when userID is empty, it returns a ZeroValueError", func(t *testing.T) {
		t.Parallel()

		gotReq, gotErr := NewUpdateRequest(
			tidy.UUIDv4{},
			tidy.None[tidy.EmailAddress](),
			tidy.None[Bio](),
			tidy.None[URL](),
			tidy.None[PasswordHash](),
		)

		var zeroValueErr *tidy.ZeroValueError
		assert.ErrorAs(t, gotErr, &zeroValueErr)
		assert.Empty(t, gotReq)
	})
}

func Test_UpdateRequest_NonZero(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		req       *UpdateRequest
		assertErr assert.ErrorAssertionFunc
	}{
		{
			name:      "when req is nil, it returns a ZeroValueError",
			req:       nil,
			assertErr: assertErrorAsZeroValueError,
		},
		{
			name: "when userID is empty, it returns a ZeroValueError",
			req: &UpdateRequest{
				userID: tidy.UUIDv4{},
			},
			assertErr: assertErrorAsZeroValueError,
		},
		{
			name: "when userID is non-zero, it returns nil",
			req: &UpdateRequest{
				userID: defaultUserID,
			},
			assertErr: assert.NoError,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotErr := tc.req.NonZero()

			tc.assertErr(t, gotErr)
		})
	}
}

func assertErrorAsZeroValueError(t assert.TestingT, err error, _ ...any) bool {
	var zeroValueErr *tidy.ZeroValueError
	return assert.ErrorAs(t, err, &zeroValueErr)
}
