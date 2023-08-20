package user

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/angusgmorrison/logfusc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/angusgmorrison/realworld-go/pkg/option"
)

func Test_ParseEmailAddress(t *testing.T) {
	t.Parallel()

	validEmailCandidate := RandomEmailAddressCandidate()

	testCases := []struct {
		name             string
		candidate        string
		wantEmailAddress EmailAddress
		wantErr          error
	}{
		{
			name:             "valid email address",
			candidate:        validEmailCandidate,
			wantEmailAddress: EmailAddress{raw: validEmailCandidate},
			wantErr:          nil,
		},
		{
			name:             "invalid email address",
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

	email := RandomEmailAddress(t)

	assert.Equal(t, email.raw, email.String())
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

	username := RandomUsername(t)

	assert.Equal(t, username.raw, username.String())
}

func Test_parsePassword(t *testing.T) {
	t.Parallel()

	anyError := errors.New("any error")

	assertEmptyPasswordHash := func(t *testing.T, hash PasswordHash, candidate logfusc.Secret[[]byte]) {
		t.Helper()
		assert.Empty(t, hash.Expose())
	}

	testCases := []struct {
		name               string
		candidate          logfusc.Secret[[]byte]
		hasher             passwordHasher
		assertPasswordHash func(t *testing.T, hash PasswordHash, candidate logfusc.Secret[[]byte])
		wantErr            error
	}{
		{
			name:      "valid password",
			candidate: RandomPasswordCandidate(),
			hasher:    bcryptHasher,
			assertPasswordHash: func(t *testing.T, hash PasswordHash, candidate logfusc.Secret[[]byte]) {
				t.Helper()
				assert.NoError(t, bcrypt.CompareHashAndPassword(hash.Expose(), candidate.Expose()))
			},
			wantErr: nil,
		},
		{
			name:               "password too short",
			candidate:          logfusc.NewSecret([]byte(strings.Repeat("a", PasswordMinLen-1))),
			hasher:             bcryptHasher,
			assertPasswordHash: assertEmptyPasswordHash,
			wantErr:            NewPasswordTooShortError(),
		},
		{
			name:               "password too long",
			candidate:          logfusc.NewSecret([]byte(strings.Repeat("a", PasswordMaxLen+1))),
			hasher:             bcryptHasher,
			assertPasswordHash: assertEmptyPasswordHash,
			wantErr:            NewPasswordTooLongError(),
		},
		{
			name:      "hasher returns any error",
			candidate: RandomPasswordCandidate(),
			hasher: func(secret logfusc.Secret[[]byte]) (PasswordHash, error) {
				return PasswordHash{}, anyError
			},
			assertPasswordHash: assertEmptyPasswordHash,
			wantErr:            anyError,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotPasswordHash, gotErr := parsePassword(tc.candidate, tc.hasher)

			tc.assertPasswordHash(t, gotPasswordHash, tc.candidate)
			assert.ErrorIs(t, gotErr, tc.wantErr)
		})
	}
}

func Test_NewPasswordHashFromTrustedSource(t *testing.T) {
	t.Parallel()

	want := RandomPasswordHash(t)

	got := NewPasswordHashFromTrustedSource(want.inner)

	assert.Equal(t, want, got)
}

func Test_PasswordHash_Expose(t *testing.T) {
	t.Parallel()

	hash := RandomPasswordHash(t)

	assert.Equal(t, hash.inner.Expose(), hash.Expose())
}

func Test_PasswordHash_StringMethods(t *testing.T) {
	t.Parallel()

	hash := RandomPasswordHash(t)
	want := fmt.Sprintf("PasswordHash{inner:%s}", hash.inner.GoString())

	assert.Equal(t, want, hash.GoString())
	assert.Equal(t, want, hash.String())
}

func Test_ParseURL(t *testing.T) {
	t.Parallel()

	validURLCandidate := RandomURLCandidate()
	netURL, err := url.Parse(validURLCandidate)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		candidate string
		wantURL   URL
		wantErr   error
	}{
		{
			name:      "valid url",
			candidate: validURLCandidate,
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

	domainURL := RandomURL(t)

	assert.Equal(t, domainURL.inner.String(), domainURL.String())
}

func Test_NewUser(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	username := RandomUsername(t)
	email := RandomEmailAddress(t)
	passwordHash := RandomPasswordHash(t)
	bio := RandomOption[Bio](t)
	imageURL := RandomOption[URL](t)
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
	username := RandomUsername(t)
	email := RandomEmailAddress(t)
	passwordHash := RandomPasswordHash(t)
	bio := RandomOption[Bio](t)
	imageURL := RandomOption[URL](t)
	user := NewUser(id, username, email, passwordHash, bio, imageURL)
	want := fmt.Sprintf("User{id:%v, username:%q, email:%q, passwordHash:%s, bio:%q, imageURL:%q}",
		id, username, email, passwordHash, bio.UnwrapOrZero(), imageURL.UnwrapOrZero())

	assert.Equal(t, want, user.GoString())
	assert.Equal(t, want, user.String())
}

func Test_NewRegistrationRequest(t *testing.T) {
	t.Parallel()

	username := RandomUsername(t)
	email := RandomEmailAddress(t)
	passwordHash := RandomPasswordHash(t)
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

	validUsernameCandidate := RandomUsernameCandidate()
	validEmailCandidate := RandomEmailAddressCandidate()
	validPasswordCandidate := RandomPasswordCandidate()

	testCases := []struct {
		name                      string
		usernameCandidate         string
		emailCandidate            string
		passwordCandidate         logfusc.Secret[[]byte]
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

				gotPasswordComparisonErr := bcrypt.CompareHashAndPassword(got.passwordHash.Expose(), validPasswordCandidate.Expose())
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
			passwordCandidate:         logfusc.NewSecret([]byte{}),
			assertRegistrationRequest: assert.Nil,
			wantErr: ValidationErrors{
				NewPasswordTooShortError().(*ValidationError),
			},
		},
		{
			name:                      "multiple invalid inputs",
			usernameCandidate:         "",
			emailCandidate:            "",
			passwordCandidate:         logfusc.NewSecret([]byte{}),
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

	username := RandomUsername(t)
	email := RandomEmailAddress(t)
	passwordHash := RandomPasswordHash(t)
	registrationRequest := NewRegistrationRequest(username, email, passwordHash)
	want := fmt.Sprintf("RegistrationRequest{username:%q, email:%q, passwordHash:%s}",
		username, email, passwordHash,
	)

	assert.Equal(t, want, registrationRequest.GoString())
	assert.Equal(t, want, registrationRequest.String())
}

func Test_NewAuthRequest(t *testing.T) {
	t.Parallel()

	email := RandomEmailAddress(t)
	passwordCandidate := RandomPasswordCandidate()
	want := &AuthRequest{
		email:             email,
		passwordCandidate: passwordCandidate,
	}

	got := NewAuthRequest(email, passwordCandidate)

	assert.Equal(t, want, got)
}

func Test_ParseAuthRequest(t *testing.T) {
	t.Parallel()

	validEmailCandidate := RandomEmailAddressCandidate()
	validPasswordCandidate := RandomPasswordCandidate()

	testCases := []struct {
		name              string
		emailCandidate    string
		passwordCandidate logfusc.Secret[[]byte]
		wantAuthRequest   *AuthRequest
		wantErr           error
	}{
		{
			name:              "valid inputs",
			emailCandidate:    validEmailCandidate,
			passwordCandidate: validPasswordCandidate,
			wantAuthRequest: &AuthRequest{
				email:             EmailAddress{raw: validEmailCandidate},
				passwordCandidate: validPasswordCandidate,
			},
			wantErr: nil,
		},
		{
			name:              "invalid email",
			emailCandidate:    "",
			passwordCandidate: validPasswordCandidate,
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

	email := RandomEmailAddress(t)
	passwordCandidate := RandomPasswordCandidate()
	authRequest := NewAuthRequest(email, passwordCandidate)
	want := fmt.Sprintf("AuthRequest{email:%q, passwordCandidate:%s}", email, passwordCandidate)

	assert.Equal(t, want, authRequest.GoString())
	assert.Equal(t, want, authRequest.String())
}

func Test_NewUpdateRequest(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	email := RandomOption[EmailAddress](t)
	passwordHash := RandomOption[PasswordHash](t)
	bioOpt := RandomOption[Bio](t)
	urlOpt := RandomOption[URL](t)
	want := &UpdateRequest{
		userID:       userID,
		email:        email,
		passwordHash: passwordHash,
		bio:          bioOpt,
		imageURL:     urlOpt,
	}

	got := NewUpdateRequest(userID, email, passwordHash, bioOpt, urlOpt)

	assert.Equal(t, want, got)
}

func Test_ParseUpdateRequest(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	validEmailCandidate := RandomEmailAddressCandidate()
	validPasswordCandidate := RandomPasswordCandidate()
	bio := RandomBio()
	validURLCandidate := RandomURLCandidate()
	email, err := ParseEmailAddress(validEmailCandidate)
	require.NoError(t, err)
	passwordHash, err := ParsePassword(validPasswordCandidate)
	require.NoError(t, err)
	imageURL, err := ParseURL(validURLCandidate)
	require.NoError(t, err)

	assertEqualUpdateRequests := func(t *testing.T, want, got *UpdateRequest) {
		t.Helper()

		if want == nil && got == nil {
			return
		}

		assert.Equal(t, want.userID, got.userID)
		assert.Equal(t, want.email, got.email)
		assert.Equal(t, want.bio, got.bio)
		assert.Equal(t, want.imageURL, got.imageURL)

		if !want.passwordHash.IsSome() {
			assert.True(t,
				!got.passwordHash.IsSome(),
				"passwordHash should be an empty Option, but value %v was found",
				got.passwordHash.UnwrapOrZero(),
			)
		} else {
			err := bcrypt.CompareHashAndPassword(
				got.passwordHash.UnwrapOrZero().Expose(),
				validPasswordCandidate.Expose(),
			)
			assert.NoError(t, err)
		}
	}

	testCases := []struct {
		name              string
		userID            uuid.UUID
		emailCandidate    option.Option[string]
		passwordCandidate option.Option[logfusc.Secret[[]byte]]
		bio               option.Option[string]
		urlCandidate      option.Option[string]
		wantUpdateRequest *UpdateRequest
		wantErr           error
	}{
		{
			name:              "valid inputs, optional inputs present",
			userID:            userID,
			emailCandidate:    option.Some(validEmailCandidate),
			passwordCandidate: option.Some(validPasswordCandidate),
			bio:               option.Some(string(bio)),
			urlCandidate:      option.Some(validURLCandidate),
			wantUpdateRequest: &UpdateRequest{
				userID:       userID,
				email:        option.Some(email),
				passwordHash: option.Some(passwordHash),
				bio:          option.Some(bio),
				imageURL:     option.Some(imageURL),
			},
			wantErr: nil,
		},
		{
			name:              "valid inputs, optional inputs absent",
			userID:            userID,
			emailCandidate:    option.None[string](),
			passwordCandidate: option.None[logfusc.Secret[[]byte]](),
			bio:               option.None[string](),
			urlCandidate:      option.None[string](),
			wantUpdateRequest: &UpdateRequest{
				userID:       userID,
				email:        option.None[EmailAddress](),
				passwordHash: option.None[PasswordHash](),
				bio:          option.None[Bio](),
				imageURL:     option.None[URL](),
			},
			wantErr: nil,
		},
		{
			name:              "invalid email",
			userID:            uuid.New(),
			emailCandidate:    option.Some(""),
			passwordCandidate: option.Some(validPasswordCandidate),
			bio:               option.Some(string(bio)),
			urlCandidate:      option.Some(validURLCandidate),
			wantUpdateRequest: nil,
			wantErr: ValidationErrors{
				NewEmailAddressFormatError("").(*ValidationError),
			},
		},
		{
			name:              "invalid password",
			userID:            uuid.New(),
			emailCandidate:    option.Some(validEmailCandidate),
			passwordCandidate: option.Some(logfusc.NewSecret([]byte{})),
			bio:               option.Some(string(bio)),
			urlCandidate:      option.Some(validURLCandidate),
			wantUpdateRequest: nil,
			wantErr: ValidationErrors{
				NewPasswordTooShortError().(*ValidationError),
			},
		},
		{
			name:              "invalid imageURL",
			userID:            uuid.New(),
			emailCandidate:    option.Some(validEmailCandidate),
			passwordCandidate: option.Some(validPasswordCandidate),
			bio:               option.Some(string(bio)),
			urlCandidate:      option.Some(string([]byte{0x07})),
			wantUpdateRequest: nil,
			wantErr: ValidationErrors{
				NewInvalidURLError().(*ValidationError),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotUpdateRequest, gotErr := ParseUpdateRequest(
				tc.userID,
				tc.emailCandidate,
				tc.passwordCandidate,
				tc.bio,
				tc.urlCandidate,
			)

			assertEqualUpdateRequests(t, tc.wantUpdateRequest, gotUpdateRequest)
			assert.Equal(t, tc.wantErr, gotErr)
		})
	}
}

func Test_UpdateRequest_StringMethods(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	email := RandomOption[EmailAddress](t)
	passwordHash := RandomOption[PasswordHash](t)
	bio := RandomOption[Bio](t)
	imageUrl := RandomOption[URL](t)
	updateRequest := NewUpdateRequest(userID, email, passwordHash, bio, imageUrl)
	want := fmt.Sprintf("UpdateRequest{userID:%q, email:%q, passwordHash:%s, bio:%q, imageURL:%q}",
		userID, email, passwordHash, bio, imageUrl)

	assert.Equal(t, want, updateRequest.GoString())
	assert.Equal(t, want, updateRequest.String())
}
