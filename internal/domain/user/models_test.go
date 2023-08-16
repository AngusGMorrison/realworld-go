package user

import (
	"fmt"
	"github.com/angusgmorrison/logfusc"
	"github.com/angusgmorrison/realworld-go/pkg/option"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"net/url"
	"strings"
	"testing"
)

func Test_ParseEmailAddress(t *testing.T) {
	t.Parallel()

	validEmail := RandomEmailAddressCandidate()

	testCases := []struct {
		name             string
		candidate        string
		wantEmailAddress EmailAddress
		wantErr          error
	}{
		{
			name:             "valid validEmail address",
			candidate:        validEmail,
			wantEmailAddress: EmailAddress{raw: validEmail},
			wantErr:          nil,
		},
		{
			name:             "invalid validEmail address",
			candidate:        "test",
			wantEmailAddress: EmailAddress{},
			wantErr:          NewEmailAddressFormatError("test"),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotEmailAddress, gotErr := ParseEmailAddress(tc.candidate)

			assert.Equal(t, tc.wantEmailAddress, gotEmailAddress)
			assert.Equal(t, tc.wantErr, gotErr)
		})
	}
}

func Test_EmailAddress_String(t *testing.T) {
	t.Parallel()

	rawEmail := RandomEmailAddressCandidate()

	email := EmailAddress{raw: rawEmail}

	assert.Equal(t, rawEmail, email.String())
}

func Test_ParseUsername(t *testing.T) {
	t.Parallel()

	validUsernameCandidate := RandomUsernameCandidate()

	testCases := []struct {
		name         string
		candidate    string
		wantUsername Username
		wantErr      error
	}{
		{
			name:         "valid username",
			candidate:    validUsernameCandidate,
			wantUsername: Username{raw: validUsernameCandidate},
			wantErr:      nil,
		},
		{
			name:         "username too short",
			candidate:    strings.Repeat("a", UsernameMinLen-1),
			wantUsername: Username{},
			wantErr:      NewUsernameTooShortError(),
		},
		{
			name:         "username too long",
			candidate:    strings.Repeat("a", UsernameMaxLen+1),
			wantUsername: Username{},
			wantErr:      NewUsernameTooLongError(),
		},
		{
			name:         "username contains invalid characters",
			candidate:    "test!",
			wantUsername: Username{},
			wantErr:      NewUsernameFormatError(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotUsername, gotErr := ParseUsername(tc.candidate)

			assert.Equal(t, tc.wantUsername, gotUsername)
			assert.Equal(t, tc.wantErr, gotErr)
		})
	}
}

func Test_Username_String(t *testing.T) {
	t.Parallel()

	username := Username{raw: "test"}

	assert.Equal(t, "test", username.String())
}

func Test_ParsePassword(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		candidate          string
		assertPasswordHash func(t *testing.T, hash PasswordHash, candidate string)
		wantErr            error
	}{
		{
			name:      "valid password",
			candidate: strings.Repeat("a", PasswordMinLen),
			assertPasswordHash: func(t *testing.T, hash PasswordHash, candidate string) {
				t.Helper()
				assert.NoError(t, bcrypt.CompareHashAndPassword(hash.Expose(), []byte(candidate)))
			},
			wantErr: nil,
		},
		{
			name:      "password too short",
			candidate: strings.Repeat("a", PasswordMinLen-1),
			assertPasswordHash: func(t *testing.T, hash PasswordHash, candidate string) {
				t.Helper()
				assert.Empty(t, hash.Expose())
			},
			wantErr: NewPasswordTooShortError(),
		},
		{
			name:      "password too long",
			candidate: strings.Repeat("a", PasswordMaxLen+1),
			assertPasswordHash: func(t *testing.T, hash PasswordHash, candidate string) {
				t.Helper()
				assert.Empty(t, hash.Expose())
			},
			wantErr: NewPasswordTooLongError(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotPasswordHash, gotErr := ParsePassword(logfusc.NewSecret(tc.candidate))

			tc.assertPasswordHash(t, gotPasswordHash, tc.candidate)
			assert.Equal(t, tc.wantErr, gotErr)
		})
	}
}

func Test_NewPasswordHashFromTrustedSource(t *testing.T) {
	t.Parallel()

	hashBytes := logfusc.NewSecret([]byte{1, 2, 3})
	want := PasswordHash{inner: hashBytes}

	got := NewPasswordHashFromTrustedSource(hashBytes)

	assert.Equal(t, want, got)
}

func Test_PasswordHash_Expose(t *testing.T) {
	t.Parallel()

	hashBytes := []byte{1, 2, 3}
	hash := PasswordHash{inner: logfusc.NewSecret(hashBytes)}

	assert.Equal(t, hashBytes, hash.Expose())
}

func Test_PasswordHash_StringMethods(t *testing.T) {
	t.Parallel()

	hashBytes := logfusc.NewSecret([]byte{1, 2, 3})
	hash := PasswordHash{inner: logfusc.NewSecret([]byte{1, 2, 3})}
	want := fmt.Sprintf("PasswordHash{inner:%s}", hashBytes.GoString())

	assert.Equal(t, want, hash.GoString())
	assert.Equal(t, want, hash.String())
}

func Test_ParseBio(t *testing.T) {
	t.Parallel()

	bio := "some bio"
	want := Bio(bio)

	got, err := ParseBio(bio)

	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func Test_ParseURL(t *testing.T) {
	t.Parallel()

	netURL, err := url.Parse("https://example.com")
	require.NoError(t, err)

	testCases := []struct {
		name      string
		candidate string
		wantURL   URL
		wantErr   error
	}{
		{
			name:      "valid url",
			candidate: "https://example.com",
			wantURL:   URL{inner: netURL},
			wantErr:   nil,
		},
		{
			name:      "invalid url",
			candidate: string([]byte{0x7f}), // ASCII control characters are not allowed
			wantURL:   URL{},
			wantErr:   NewInvalidURLError(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotURL, gotErr := ParseURL(tc.candidate)

			assert.Equal(t, tc.wantURL, gotURL)
			assert.Equal(t, tc.wantErr, gotErr)
		})
	}
}

func Test_URL_String(t *testing.T) {
	t.Parallel()

	domainURL := URL{inner: &url.URL{Scheme: "https", Host: "example.com"}}

	assert.Equal(t, domainURL.inner.String(), domainURL.String())
}

func Test_NewUser(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	username := Username{raw: "test"}
	email := EmailAddress{raw: "test@test.com"}
	passwordHash := NewPasswordHashFromTrustedSource(logfusc.NewSecret([]byte{1, 2, 3}))
	bio := option.Some(Bio("some bio"))
	imageURL := option.Some(URL{inner: &url.URL{Scheme: "https", Host: "example.com"}})
	wantUser := &User{
		id:           id,
		username:     username,
		email:        email,
		passwordHash: passwordHash,
		bio:          bio,
		imageURL:     imageURL,
	}

	gotUser := NewUser(id, username, email, passwordHash, bio, imageURL)

	assert.Equal(t, wantUser, gotUser)
}

func Test_User_StringMethods(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	username := Username{raw: "test"}
	email := EmailAddress{raw: "test@test.com"}
	passwordHash := NewPasswordHashFromTrustedSource(logfusc.NewSecret([]byte{1, 2, 3}))
	bio := option.Some(Bio("some bio"))
	imageURL := option.Some(URL{inner: &url.URL{Scheme: "https", Host: "example.com"}})
	user := NewUser(id, username, email, passwordHash, bio, imageURL)
	want := fmt.Sprintf("User{id:%v, username:%q, email:%q, passwordHash:%s, bio:%q, imageURL:%q}",
		id, username, email, passwordHash, bio.ValueOrZero(), imageURL.ValueOrZero())

	assert.Equal(t, want, user.GoString())
	assert.Equal(t, want, user.String())
}

func Test_NewRegistrationRequest(t *testing.T) {
	t.Parallel()

	username := Username{raw: "test"}
	email := EmailAddress{raw: "test@test.com"}
	passwordHash := NewPasswordHashFromTrustedSource(logfusc.NewSecret([]byte{1, 2, 3}))
	want := &RegistrationRequest{
		username:     username,
		email:        email,
		passwordHash: passwordHash,
	}

	got := NewRegistrationRequest(username, email, passwordHash)

	assert.Equal(t, want, got)
}

func Test_ParseRegistrationRequest(t *testing.T) {
	t.Parallel()

	validUsernameCandidate := "test"
	validEmailCandidate := "test@test.com"
	validPasswordCandidate := logfusc.NewSecret("testPassword")

	testCases := []struct {
		name                      string
		usernameCandidate         string
		emailCandidate            string
		passwordCandidate         logfusc.Secret[string]
		assertRegistrationRequest assert.ValueAssertionFunc
		wantErr                   error
	}{
		{
			name:              "valid inputs",
			usernameCandidate: validUsernameCandidate,
			emailCandidate:    validEmailCandidate,
			passwordCandidate: validPasswordCandidate,
			assertRegistrationRequest: func(t assert.TestingT, v interface{}, msgAndArgs ...interface{}) bool {
				wantUsername := Username{raw: validUsernameCandidate}
				wantEmail := EmailAddress{raw: validEmailCandidate}

				got := v.(*RegistrationRequest)

				if pass := assert.Equal(t, wantUsername, got.username); !pass {
					return pass
				}

				if pass := assert.Equal(t, wantEmail, got.email); !pass {
					return pass
				}

				gotPasswordComparisonErr := bcrypt.CompareHashAndPassword(got.passwordHash.Expose(), []byte(validPasswordCandidate.Expose()))
				if pass := assert.NoError(t, gotPasswordComparisonErr); !pass {
					return pass
				}

				return true
			},
			wantErr: nil,
		},
		{
			name:                      "invalid username",
			usernameCandidate:         "",
			emailCandidate:            validEmailCandidate,
			passwordCandidate:         validPasswordCandidate,
			assertRegistrationRequest: assert.Nil,
			wantErr: ValidationErrors{
				NewUsernameTooShortError().(*ValidationError),
			},
		},
		{
			name:                      "invalid email",
			usernameCandidate:         validUsernameCandidate,
			emailCandidate:            "",
			passwordCandidate:         validPasswordCandidate,
			assertRegistrationRequest: assert.Nil,
			wantErr: ValidationErrors{
				NewEmailAddressFormatError("").(*ValidationError),
			},
		},
		{
			name:                      "invalid password",
			usernameCandidate:         validUsernameCandidate,
			emailCandidate:            validEmailCandidate,
			passwordCandidate:         logfusc.NewSecret(""),
			assertRegistrationRequest: assert.Nil,
			wantErr: ValidationErrors{
				NewPasswordTooShortError().(*ValidationError),
			},
		},
		{
			name:                      "multiple invalid inputs",
			usernameCandidate:         "",
			emailCandidate:            "",
			passwordCandidate:         logfusc.NewSecret(""),
			assertRegistrationRequest: assert.Nil,
			wantErr: ValidationErrors{
				NewUsernameTooShortError().(*ValidationError),
				NewEmailAddressFormatError("").(*ValidationError),
				NewPasswordTooShortError().(*ValidationError),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotRegistrationRequest, gotErr := ParseRegistrationRequest(tc.usernameCandidate, tc.emailCandidate, tc.passwordCandidate)

			tc.assertRegistrationRequest(t, gotRegistrationRequest)
			assert.Equal(t, tc.wantErr, gotErr)
		})
	}
}

func Test_RegistrationRequest_StringMethods(t *testing.T) {
	t.Parallel()

	username := Username{raw: "test"}
	email := EmailAddress{raw: "test@test.com"}
	passwordHash := NewPasswordHashFromTrustedSource(logfusc.NewSecret([]byte{1, 2, 3}))
	registrationRequest := NewRegistrationRequest(username, email, passwordHash)
	want := fmt.Sprintf("RegistrationRequest{username:%q, email:%q, passwordHash:%s}",
		username, email, passwordHash,
	)

	assert.Equal(t, want, registrationRequest.GoString())
	assert.Equal(t, want, registrationRequest.String())
}

func Test_NewAuthRequest(t *testing.T) {
	t.Parallel()

	email := EmailAddress{raw: "test@test.com"}
	passwordCandidate := logfusc.NewSecret("testPassword")
	want := &AuthRequest{
		email:             email,
		passwordCandidate: passwordCandidate,
	}

	got := NewAuthRequest(email, passwordCandidate)

	assert.Equal(t, want, got)
}

func Test_ParseAuthRequest(t *testing.T) {
	t.Parallel()

	validEmailCandidate := "test@test.com"
	validPassword := logfusc.NewSecret("testPassword")

	testCases := []struct {
		name              string
		emailCandidate    string
		passwordCandidate logfusc.Secret[string]
		wantAuthRequest   *AuthRequest
		wantErr           error
	}{
		{
			name:              "valid inputs",
			emailCandidate:    validEmailCandidate,
			passwordCandidate: validPassword,
			wantAuthRequest: &AuthRequest{
				email:             EmailAddress{raw: validEmailCandidate},
				passwordCandidate: validPassword,
			},
			wantErr: nil,
		},
		{
			name:              "invalid email",
			emailCandidate:    "",
			passwordCandidate: validPassword,
			wantAuthRequest:   nil,
			wantErr:           NewEmailAddressFormatError(""),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotAuthRequest, gotErr := ParseAuthRequest(tc.emailCandidate, tc.passwordCandidate)

			assert.Equal(t, tc.wantAuthRequest, gotAuthRequest)
			assert.Equal(t, tc.wantErr, gotErr)
		})
	}
}

func Test_AuthRequest_StringMethods(t *testing.T) {
	t.Parallel()

	email := EmailAddress{raw: "test@test.com"}
	passwordCandidate := logfusc.NewSecret("testPassword")
	authRequest := NewAuthRequest(email, passwordCandidate)
	want := fmt.Sprintf("AuthRequest{email:%q, passwordCandidate:%s}", email, passwordCandidate)

	assert.Equal(t, want, authRequest.GoString())
	assert.Equal(t, want, authRequest.String())
}

func Test_NewUpdateRequest(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	email := option.Some(RandomEmailAddress(t))
	passwordHash := option.Some(RandomPasswordHash(t))
	bio := option.Some(RandomBio(t))
	url := option.Some(RandomURL(t))
	want := &UpdateRequest{
		userID:       userID,
		email:        email,
		passwordHash: passwordHash,
		bio:          bio,
		imageURL:     url,
	}

	got := NewUpdateRequest(userID, email, passwordHash, bio, url)

	assert.Equal(t, want, got)
}
