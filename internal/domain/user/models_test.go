package user

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/angusgmorrison/realworld-go/pkg/etag"

	"github.com/angusgmorrison/realworld-go/pkg/option"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	assertEmptyPasswordHash := func(t *testing.T, hash PasswordHash, candidate string) {
		t.Helper()
		assert.Empty(t, hash)
	}

	testCases := []struct {
		name               string
		candidate          string
		hasher             passwordHasher
		assertPasswordHash func(t *testing.T, hash PasswordHash, candidate string)
		wantErr            error
	}{
		{
			name:      "valid password",
			candidate: RandomPasswordCandidate(),
			hasher:    bcryptHash,
			assertPasswordHash: func(t *testing.T, hash PasswordHash, candidate string) {
				t.Helper()
				assert.NoError(t, bcryptCompare(hash, candidate))
			},
			wantErr: nil,
		},
		{
			name:               "password too short",
			candidate:          strings.Repeat("a", PasswordMinLen-1),
			hasher:             bcryptHash,
			assertPasswordHash: assertEmptyPasswordHash,
			wantErr:            NewPasswordTooShortError(),
		},
		{
			name:               "password too long",
			candidate:          strings.Repeat("a", PasswordMaxLen+1),
			hasher:             bcryptHash,
			assertPasswordHash: assertEmptyPasswordHash,
			wantErr:            NewPasswordTooLongError(),
		},
		{
			name:      "hasher returns any error",
			candidate: RandomPasswordCandidate(),
			hasher: func(secret string) (PasswordHash, error) {
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

	got := NewPasswordHashFromTrustedSource(want.bytes)

	assert.Equal(t, want, got)
}

func Test_PasswordHash_GoString(t *testing.T) {
	t.Parallel()

	hash := RandomPasswordHash(t)

	assert.Equal(t, "PasswordHash{bytes:REDACTED}", hash.GoString())
}

func Test_PasswordHash_String(t *testing.T) {
	t.Parallel()

	hash := RandomPasswordHash(t)

	assert.Equal(t, "{REDACTED}", hash.String())
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

func Test_URL_Equal(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		url1 URL
		url2 URL
		want bool
	}{
		{
			name: "zero values",
			url1: URL{},
			url2: URL{},
			want: true,
		},
		{
			name: "equal values",
			url1: URL{inner: &url.URL{Scheme: "https", Host: "example.com"}},
			url2: URL{inner: &url.URL{Scheme: "https", Host: "example.com"}},
			want: true,
		},
		{
			name: "unequal values",
			url1: URL{inner: &url.URL{Scheme: "https", Host: "example.com"}},
			url2: URL{inner: &url.URL{Scheme: "http", Host: "example.com"}},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.url1.Equal(tc.url2)

			assert.Equal(t, tc.want, got)
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
	eTag := etag.Random()
	username := RandomUsername(t)
	email := RandomEmailAddress(t)
	passwordHash := RandomPasswordHash(t)
	bio := RandomOption[Bio](t)
	imageURL := RandomOption[URL](t)
	wantUser := &User{
		id:           id,
		eTag:         eTag,
		username:     username,
		email:        email,
		passwordHash: passwordHash,
		bio:          bio,
		imageURL:     imageURL,
	}

	gotUser := NewUser(id, eTag, username, email, passwordHash, bio, imageURL)

	assert.Equal(t, wantUser, gotUser)
}

func Test_User_GoString(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	eTag := etag.Random()
	username := RandomUsername(t)
	email := RandomEmailAddress(t)
	passwordHash := RandomPasswordHash(t)
	bio := RandomOption[Bio](t)
	imageURL := RandomOption[URL](t)
	user := NewUser(id, eTag, username, email, passwordHash, bio, imageURL)
	want := fmt.Sprintf(
		"User{id:%#v, eTag:%#v, username:%#v, email:%#v, passwordHash:PasswordHash{bytes:REDACTED}, bio:%#v, imageURL:%#v}",
		id, eTag, username, email, bio, imageURL)

	assert.Equal(t, want, user.GoString())
}

func Test_User_String(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	eTag := etag.Random()
	username := RandomUsername(t)
	email := RandomEmailAddress(t)
	passwordHash := RandomPasswordHash(t)
	bio := RandomOption[Bio](t)
	imageURL := RandomOption[URL](t)
	user := NewUser(id, eTag, username, email, passwordHash, bio, imageURL)
	want := fmt.Sprintf("{%s %s %s %s {REDACTED} %s %s}", id, eTag, username, email, bio, imageURL)

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
		passwordCandidate         string
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
					return false
				}

				if pass := assert.Equal(t, wantEmail, got.email); !pass {
					return false
				}

				gotPasswordComparisonErr := bcryptCompare(got.PasswordHash(), validPasswordCandidate)
				if pass := assert.NoError(t, gotPasswordComparisonErr); !pass {
					return false
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
			passwordCandidate:         "",
			assertRegistrationRequest: assert.Nil,
			wantErr: ValidationErrors{
				NewPasswordTooShortError().(*ValidationError),
			},
		},
		{
			name:                      "multiple invalid inputs",
			usernameCandidate:         "",
			emailCandidate:            "",
			passwordCandidate:         "",
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

func Test_RegistrationRequest_Equal(t *testing.T) {
	t.Parallel()

	email1 := RandomEmailAddress(t)
	email2 := RandomEmailAddress(t)
	username1 := RandomUsername(t)
	username2 := RandomUsername(t)
	password1 := RandomPasswordCandidate()
	password2 := RandomPasswordCandidate()
	hash1, err := bcryptHash(password1)
	require.NoError(t, err)
	hash2, err := bcryptHash(password2)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		req1     *RegistrationRequest
		req2     *RegistrationRequest
		password string
		want     bool
	}{
		{
			name:     "zero values",
			req1:     &RegistrationRequest{},
			req2:     &RegistrationRequest{},
			password: "",
			want:     true,
		},
		{
			name:     "equal values",
			req1:     NewRegistrationRequest(username1, email1, hash1),
			req2:     NewRegistrationRequest(username1, email1, hash1),
			password: password1,
			want:     true,
		},
		{
			name:     "unequal usernames",
			req1:     NewRegistrationRequest(username1, email1, hash1),
			req2:     NewRegistrationRequest(username2, email1, hash1),
			password: password1,
			want:     false,
		},
		{
			name:     "unequal emails",
			req1:     NewRegistrationRequest(username1, email1, hash1),
			req2:     NewRegistrationRequest(username1, email2, hash1),
			password: password1,
			want:     false,
		},
		{
			name:     "unequal passwords",
			req1:     NewRegistrationRequest(username1, email1, hash1),
			req2:     NewRegistrationRequest(username1, email1, hash2),
			password: password1,
			want:     false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.req1.Equal(tc.req2, tc.password)

			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_RegistrationRequest_GoString(t *testing.T) {
	t.Parallel()

	username := RandomUsername(t)
	email := RandomEmailAddress(t)
	passwordHash := RandomPasswordHash(t)
	registrationRequest := NewRegistrationRequest(username, email, passwordHash)
	want := fmt.Sprintf("RegistrationRequest{username:%#v, email:%#v, passwordHash:PasswordHash{bytes:REDACTED}}",
		username, email)

	assert.Equal(t, want, registrationRequest.GoString())
}

func Test_RegistrationRequest_String(t *testing.T) {
	t.Parallel()

	username := RandomUsername(t)
	email := RandomEmailAddress(t)
	passwordHash := RandomPasswordHash(t)
	registrationRequest := NewRegistrationRequest(username, email, passwordHash)
	want := fmt.Sprintf("{%s %s {REDACTED}}", username, email)

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
		passwordCandidate string
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
			wantErr:           ValidationErrors{NewEmailAddressFormatError("").(*ValidationError)},
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

func Test_AuthRequest_GoString(t *testing.T) {
	t.Parallel()

	email := RandomEmailAddress(t)
	passwordCandidate := RandomPasswordCandidate()
	authRequest := NewAuthRequest(email, passwordCandidate)
	want := fmt.Sprintf("AuthRequest{email:%#v, passwordCandidate:REDACTED}", email)

	assert.Equal(t, want, authRequest.GoString())
}

func Test_AuthRequest_String(t *testing.T) {
	t.Parallel()

	email := RandomEmailAddress(t)
	passwordCandidate := RandomPasswordCandidate()
	authRequest := NewAuthRequest(email, passwordCandidate)
	want := fmt.Sprintf("{%s REDACTED}", email)

	assert.Equal(t, want, authRequest.String())
}

func Test_NewUpdateRequest(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	eTag := etag.Random()
	email := RandomOption[EmailAddress](t)
	passwordHash := RandomOption[PasswordHash](t)
	bioOpt := RandomOption[Bio](t)
	urlOpt := RandomOption[URL](t)
	want := &UpdateRequest{
		userID:       userID,
		eTag:         eTag,
		email:        email,
		passwordHash: passwordHash,
		bio:          bioOpt,
		imageURL:     urlOpt,
	}

	got := NewUpdateRequest(userID, eTag, email, passwordHash, bioOpt, urlOpt)

	assert.Equal(t, want, got)
}

func Test_ParseUpdateRequest(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	eTag := etag.Random()
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
			err := bcryptCompare(got.PasswordHash().UnwrapOrZero(), validPasswordCandidate)
			assert.NoError(t, err)
		}
	}

	testCases := []struct {
		name              string
		emailCandidate    option.Option[string]
		passwordCandidate option.Option[string]
		bio               option.Option[string]
		urlCandidate      option.Option[string]
		wantUpdateRequest *UpdateRequest
		wantErr           error
	}{
		{
			name:              "valid inputs, optional inputs present",
			emailCandidate:    option.Some(validEmailCandidate),
			passwordCandidate: option.Some(validPasswordCandidate),
			bio:               option.Some(string(bio)),
			urlCandidate:      option.Some(validURLCandidate),
			wantUpdateRequest: &UpdateRequest{
				userID:       userID,
				eTag:         eTag,
				email:        option.Some(email),
				passwordHash: option.Some(passwordHash),
				bio:          option.Some(bio),
				imageURL:     option.Some(imageURL),
			},
			wantErr: nil,
		},
		{
			name:              "valid inputs, optional inputs absent",
			emailCandidate:    option.None[string](),
			passwordCandidate: option.None[string](),
			bio:               option.None[string](),
			urlCandidate:      option.None[string](),
			wantUpdateRequest: &UpdateRequest{
				userID:       userID,
				eTag:         eTag,
				email:        option.None[EmailAddress](),
				passwordHash: option.None[PasswordHash](),
				bio:          option.None[Bio](),
				imageURL:     option.None[URL](),
			},
			wantErr: nil,
		},
		{
			name:              "invalid email",
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
			emailCandidate:    option.Some(validEmailCandidate),
			passwordCandidate: option.Some(""),
			bio:               option.Some(string(bio)),
			urlCandidate:      option.Some(validURLCandidate),
			wantUpdateRequest: nil,
			wantErr: ValidationErrors{
				NewPasswordTooShortError().(*ValidationError),
			},
		},
		{
			name:              "invalid imageURL",
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
				userID,
				eTag,
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

func Test_UpdateRequest_GoString(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	eTag := etag.Random()
	email := RandomOption[EmailAddress](t)
	passwordHash := option.Some(RandomPasswordHash(t)) // None[PasswordHash] is always safe to print
	bio := RandomOption[Bio](t)
	imageUrl := RandomOption[URL](t)
	updateRequest := NewUpdateRequest(userID, eTag, email, passwordHash, bio, imageUrl)
	want := fmt.Sprintf(
		"UpdateRequest{userID:%#v, eTag:%#v, email:%#v, passwordHash:option.Option[user.PasswordHash]{some:true, value:PasswordHash{bytes:REDACTED}}, bio:%#v, imageURL:%#v}",
		userID, eTag, email, bio, imageUrl)

	assert.Equal(t, want, updateRequest.GoString())
}

func Test_UpdateRequest_String(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	eTag := etag.Random()
	email := RandomOption[EmailAddress](t)
	passwordHash := option.Some(RandomPasswordHash(t)) // None[PasswordHash] is always safe to print
	bio := RandomOption[Bio](t)
	imageUrl := RandomOption[URL](t)
	updateRequest := NewUpdateRequest(userID, eTag, email, passwordHash, bio, imageUrl)
	want := fmt.Sprintf("{%s %s %s Some[user.PasswordHash]{{REDACTED}} %s %s}", userID, eTag, email, bio, imageUrl)

	assert.Equal(t, want, updateRequest.String())
}

func Test_UpdateRequest_Equal(t *testing.T) {
	t.Parallel()

	userID1 := uuid.New()
	userID2 := uuid.New()
	email1 := RandomEmailAddress(t)
	email2 := RandomEmailAddress(t)
	password1 := RandomPasswordCandidate()
	password2 := RandomPasswordCandidate()
	hash1, err := bcryptHash(password1)
	require.NoError(t, err)
	hash2, err := bcryptHash(password2)
	require.NoError(t, err)
	bio1 := RandomBio()
	bio2 := RandomBio()
	url1 := RandomURL(t)
	url2 := RandomURL(t)

	testCases := []struct {
		name     string
		req1     *UpdateRequest
		req2     *UpdateRequest
		password option.Option[string]
		want     bool
	}{
		{
			name:     "zero values",
			req1:     &UpdateRequest{},
			req2:     &UpdateRequest{},
			password: option.None[string](),
			want:     true,
		},
		{
			name: "equal values (some)",
			req1: &UpdateRequest{
				userID:       userID1,
				email:        option.Some(email1),
				passwordHash: option.Some(hash1),
				bio:          option.Some(bio1),
				imageURL:     option.Some(url1),
			},
			req2: &UpdateRequest{
				userID:       userID1,
				email:        option.Some(email1),
				passwordHash: option.Some(hash1),
				bio:          option.Some(bio1),
				imageURL:     option.Some(url1),
			},
			password: option.Some(password1),
			want:     true,
		},
		{
			name: "equal values (none)",
			req1: &UpdateRequest{
				userID:       userID1,
				email:        option.None[EmailAddress](),
				passwordHash: option.None[PasswordHash](),
				bio:          option.None[Bio](),
				imageURL:     option.None[URL](),
			},
			req2: &UpdateRequest{
				userID:       userID1,
				email:        option.None[EmailAddress](),
				passwordHash: option.None[PasswordHash](),
				bio:          option.None[Bio](),
				imageURL:     option.None[URL](),
			},
			password: option.None[string](),
			want:     true,
		},
		{
			name: "unequal userIDs",
			req1: &UpdateRequest{
				userID:       userID1,
				email:        option.None[EmailAddress](),
				passwordHash: option.None[PasswordHash](),
				bio:          option.None[Bio](),
				imageURL:     option.None[URL](),
			},
			req2: &UpdateRequest{
				userID:       userID2,
				email:        option.None[EmailAddress](),
				passwordHash: option.None[PasswordHash](),
				bio:          option.None[Bio](),
				imageURL:     option.None[URL](),
			},
			password: option.None[string](),
			want:     false,
		},
		{
			name: "unequal emails (some)",
			req1: &UpdateRequest{
				userID:       userID1,
				email:        option.Some(email1),
				passwordHash: option.None[PasswordHash](),
				bio:          option.None[Bio](),
				imageURL:     option.None[URL](),
			},
			req2: &UpdateRequest{
				userID:       userID1,
				email:        option.Some(email2),
				passwordHash: option.None[PasswordHash](),
				bio:          option.None[Bio](),
				imageURL:     option.None[URL](),
			},
			password: option.None[string](),
			want:     false,
		},
		{
			name: "unequal emails (none)",
			req1: &UpdateRequest{
				userID:       userID1,
				email:        option.None[EmailAddress](),
				passwordHash: option.None[PasswordHash](),
				bio:          option.None[Bio](),
				imageURL:     option.None[URL](),
			},
			req2: &UpdateRequest{
				userID:       userID1,
				email:        option.Some(email2),
				passwordHash: option.None[PasswordHash](),
				bio:          option.None[Bio](),
				imageURL:     option.None[URL](),
			},
			password: option.None[string](),
			want:     false,
		},
		{
			name: "unequal passwords (some)",
			req1: &UpdateRequest{
				userID:       userID1,
				email:        option.None[EmailAddress](),
				passwordHash: option.Some(hash1),
				bio:          option.None[Bio](),
				imageURL:     option.None[URL](),
			},
			req2: &UpdateRequest{
				userID:       userID1,
				email:        option.None[EmailAddress](),
				passwordHash: option.Some(hash2),
				bio:          option.None[Bio](),
				imageURL:     option.None[URL](),
			},
			password: option.Some(password1),
			want:     false,
		},
		{
			name: "unequal passwords (none)",
			req1: &UpdateRequest{
				userID:       userID1,
				email:        option.None[EmailAddress](),
				passwordHash: option.None[PasswordHash](),
				bio:          option.None[Bio](),
				imageURL:     option.None[URL](),
			},
			req2: &UpdateRequest{
				userID:       userID1,
				email:        option.None[EmailAddress](),
				passwordHash: option.Some(hash2),
				bio:          option.None[Bio](),
				imageURL:     option.None[URL](),
			},
			password: option.None[string](),
			want:     false,
		},
		{
			name: "unequal bios (some)",
			req1: &UpdateRequest{
				userID:       userID1,
				email:        option.None[EmailAddress](),
				passwordHash: option.None[PasswordHash](),
				bio:          option.Some(bio1),
				imageURL:     option.None[URL](),
			},
			req2: &UpdateRequest{
				userID:       userID1,
				email:        option.None[EmailAddress](),
				passwordHash: option.None[PasswordHash](),
				bio:          option.Some(bio2),
				imageURL:     option.None[URL](),
			},
			password: option.None[string](),
			want:     false,
		},
		{
			name: "unequal bios (none)",
			req1: &UpdateRequest{
				userID:       userID1,
				email:        option.None[EmailAddress](),
				passwordHash: option.None[PasswordHash](),
				bio:          option.None[Bio](),
				imageURL:     option.None[URL](),
			},
			req2: &UpdateRequest{
				userID:       userID1,
				email:        option.None[EmailAddress](),
				passwordHash: option.None[PasswordHash](),
				bio:          option.Some(bio2),
				imageURL:     option.None[URL](),
			},
			password: option.None[string](),
			want:     false,
		},
		{
			name: "unequal image URLs (some)",
			req1: &UpdateRequest{
				userID:       userID1,
				email:        option.None[EmailAddress](),
				passwordHash: option.None[PasswordHash](),
				bio:          option.None[Bio](),
				imageURL:     option.Some(url1),
			},
			req2: &UpdateRequest{
				userID:       userID1,
				email:        option.None[EmailAddress](),
				passwordHash: option.None[PasswordHash](),
				bio:          option.None[Bio](),
				imageURL:     option.Some(url2),
			},
			password: option.None[string](),
			want:     false,
		},
		{
			name: "unequal image URLs (none)",
			req1: &UpdateRequest{
				userID:       userID1,
				email:        option.None[EmailAddress](),
				passwordHash: option.None[PasswordHash](),
				bio:          option.None[Bio](),
				imageURL:     option.None[URL](),
			},
			req2: &UpdateRequest{
				userID:       userID1,
				email:        option.None[EmailAddress](),
				passwordHash: option.None[PasswordHash](),
				bio:          option.None[Bio](),
				imageURL:     option.Some(url2),
			},
			password: option.None[string](),
			want:     false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.req1.Equal(tc.req2, tc.password)

			assert.Equal(t, tc.want, got)
		})
	}
}
