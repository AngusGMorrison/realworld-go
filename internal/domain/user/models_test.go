package user

//
//import (
//	"errors"
//	"github.com/angusgmorrison/realworld/pkg/tidy/option"
//	"github.com/google/uuid"
//	"net/url"
//	"strings"
//	"testing"
//
//	"github.com/angusgmorrison/realworld/pkg/tidy"
//	"github.com/stretchr/testify/assert"
//)
//
//var (
//	defaultEmail, _ = tidy.ParseEmailAddress("test@test.com")
//	defaultUserID   = tidy.NewUUIDv4()
//	defaultUsername = Username{raw: "johndoe"}
//)
//
//func Test_ParseUsername(t *testing.T) {
//	t.Parallel()
//
//	testCases := []struct {
//		name      string
//		input     string
//		want      Username
//		assertErr assert.ErrorAssertionFunc
//	}{
//		{
//			name:  "valid username",
//			input: "johndoe",
//			want:  Username{raw: "johndoe"},
//			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
//				return assert.NoError(t, err)
//			},
//		},
//		{
//			name:  "too short",
//			input: strings.Repeat("a", UsernameMinLen-1),
//			want:  Username{},
//			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
//				return errors.Is(err, ErrUsernameTooShort)
//			},
//		},
//		{
//			name:  "too long",
//			input: strings.Repeat("a", UsernameMaxLen+1),
//			want:  Username{},
//			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
//				return errors.Is(err, ErrUsernameTooLong)
//			},
//		},
//		{
//			name:  "invalid format",
//			input: "john doe",
//			want:  Username{},
//			assertErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
//				return errors.Is(err, ErrUsernameFormat)
//			},
//		},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//
//		t.Run(tc.name, func(t *testing.T) {
//			t.Parallel()
//
//			got, err := ParseUsername(tc.input)
//
//			assert.Equal(t, tc.want, got)
//			tc.assertErr(t, err)
//		})
//	}
//}
//
//func Test_Username_String(t *testing.T) {
//	t.Parallel()
//
//	testCases := []struct {
//		name string
//		u    Username
//		want string
//	}{
//		{
//			name: "valid username",
//			u:    defaultUsername,
//			want: "johndoe",
//		},
//		{
//			name: "empty username",
//			u:    Username{},
//			want: "",
//		},
//	}
//
//	for _, tc := range testCases {
//		tc := tc
//
//		t.Run(tc.name, func(t *testing.T) {
//			t.Parallel()
//
//			got := tc.u.String()
//
//			assert.Equal(t, tc.want, got)
//		})
//	}
//}
//
//func Test_NewUser(t *testing.T) {
//	t.Parallel()
//
//	id := tidy.NewUUIDv4()
//	passwordHash, _ := ParsePassword(NewPasswordCandidate("password"))
//	bioSome, _ := option.Some(Bio("some bio"))
//	bioNone := option.None[Bio]()
//	imageURLSome, _ := option.Some(URL{inner: &url.URL{Scheme: "https", Host: "example.com"}})
//	imageURLNone := option.None[URL]()
//
//	t.Run("when all required arguments are present", func(t *testing.T) {
//		t.Parallel()
//
//		testCases := []struct {
//			name         string
//			id           uuid.UUID
//			username     Username
//			email        tidy.EmailAddress
//			passwordHash PasswordHash
//			bio          option.Option[Bio]
//			imageURL     option.Option[URL]
//		}{
//			{
//				name:         "when all optional arguments are present, it returns a User",
//				id:           id,
//				username:     defaultUsername,
//				email:        defaultEmail,
//				passwordHash: passwordHash,
//				bio:          bioSome,
//				imageURL:     imageURLSome,
//			},
//			{
//				name:         "when bio is None, it returns a User",
//				id:           id,
//				username:     defaultUsername,
//				email:        defaultEmail,
//				passwordHash: passwordHash,
//				bio:          bioNone,
//				imageURL:     imageURLSome,
//			},
//			{
//				name:         "when imageURL is None, it returns a User",
//				id:           id,
//				username:     defaultUsername,
//				email:        defaultEmail,
//				passwordHash: passwordHash,
//				bio:          bioSome,
//				imageURL:     imageURLNone,
//			},
//		}
//
//		for _, tc := range testCases {
//			tc := tc
//
//			t.Run(tc.name, func(t *testing.T) {
//				t.Parallel()
//
//				expectedUser := User{
//					id:           tc.id,
//					username:     tc.username,
//					email:        tc.email,
//					passwordHash: tc.passwordHash,
//					bio:          tc.bio,
//					imageURL:     tc.imageURL,
//				}
//
//				gotUser, err := NewUser(tc.id, tc.username, tc.email, tc.passwordHash, tc.bio, tc.imageURL)
//
//				assert.NoError(t, err)
//				assert.Equal(t, expectedUser, gotUser)
//			})
//		}
//	})
//
//	t.Run("when required arguments are empty", func(t *testing.T) {
//		t.Parallel()
//
//		testCases := []struct {
//			name         string
//			id           uuid.UUID
//			username     Username
//			email        tidy.EmailAddress
//			passwordHash PasswordHash
//		}{
//			{
//				name:         "when ID is empty, it returns a ZeroValueError",
//				id:           uuid.UUID{},
//				username:     defaultUsername,
//				email:        defaultEmail,
//				passwordHash: passwordHash,
//			},
//			{
//				name:         "when username is empty, it returns a ZeroValueError",
//				id:           id,
//				username:     Username{},
//				email:        defaultEmail,
//				passwordHash: passwordHash,
//			},
//			{
//				name:         "when email is empty, it returns a ZeroValueError",
//				id:           id,
//				username:     defaultUsername,
//				email:        tidy.EmailAddress{},
//				passwordHash: passwordHash,
//			},
//			{
//				name:         "when passwordHash is empty, it returns a ZeroValueError",
//				id:           id,
//				username:     defaultUsername,
//				email:        defaultEmail,
//				passwordHash: PasswordHash{},
//			},
//		}
//
//		for _, tc := range testCases {
//			tc := tc
//
//			t.Run(tc.name, func(t *testing.T) {
//				t.Parallel()
//
//				gotUser, gotErr := NewUser(tc.id, tc.username, tc.email, tc.passwordHash, option.None[Bio](), option.None[URL]())
//
//				var zeroValueErr *tidy.ZeroValueError
//				assert.ErrorAs(t, gotErr, &zeroValueErr)
//				assert.Empty(t, gotUser)
//			})
//		}
//	})
//}
//
//func Test_NewRegistrationRequest(t *testing.T) {
//	t.Parallel()
//
//	passwordHash, _ := ParsePassword(NewPasswordCandidate("password"))
//
//	t.Run("when all arguments are non-zero, it returns a RegistrationRequest", func(t *testing.T) {
//		t.Parallel()
//
//		expectedReq := RegistrationRequest{
//			username:     defaultUsername,
//			email:        defaultEmail,
//			passwordHash: passwordHash,
//		}
//
//		gotReq, gotErr := NewRegistrationRequest(defaultUsername, defaultEmail, passwordHash)
//
//		assert.NoError(t, gotErr)
//		assert.Equal(t, expectedReq, gotReq)
//	})
//
//	t.Run("when arguments are empty", func(t *testing.T) {
//		t.Parallel()
//
//		testCases := []struct {
//			name         string
//			username     Username
//			email        tidy.EmailAddress
//			passwordHash PasswordHash
//		}{
//			{
//				name:         "when username is empty, it returns a ZeroValueError",
//				username:     Username{},
//				email:        defaultEmail,
//				passwordHash: passwordHash,
//			},
//			{
//				name:         "when email is empty, it returns a ZeroValueError",
//				username:     defaultUsername,
//				email:        tidy.EmailAddress{},
//				passwordHash: passwordHash,
//			},
//			{
//				name:         "when passwordHash is empty, it returns a ZeroValueError",
//				username:     defaultUsername,
//				email:        defaultEmail,
//				passwordHash: PasswordHash{},
//			},
//		}
//
//		for _, tc := range testCases {
//			tc := tc
//
//			t.Run(tc.name, func(t *testing.T) {
//				t.Parallel()
//
//				gotReq, gotErr := NewRegistrationRequest(tc.username, tc.email, tc.passwordHash)
//
//				var zeroValueErr *tidy.ZeroValueError
//				assert.ErrorAs(t, gotErr, &zeroValueErr)
//				assert.Empty(t, gotReq)
//			})
//		}
//	})
//}
//
//func Test_NewAuthRequest(t *testing.T) {
//	t.Parallel()
//
//	t.Run("when email is non-zero, it returns an AuthRequest", func(t *testing.T) {
//		t.Parallel()
//
//		expectedReq := AuthRequest{
//			email: defaultEmail,
//		}
//
//		gotReq, gotErr := NewAuthRequest(defaultEmail, PasswordCandidate{})
//
//		assert.NoError(t, gotErr)
//		assert.Equal(t, expectedReq, gotReq)
//	})
//
//	t.Run("when email is empty, it returns a ZeroValueError", func(t *testing.T) {
//		t.Parallel()
//
//		gotReq, gotErr := NewAuthRequest(tidy.EmailAddress{}, PasswordCandidate{})
//
//		var zeroValueErr *tidy.ZeroValueError
//		assert.ErrorAs(t, gotErr, &zeroValueErr)
//		assert.Empty(t, gotReq)
//	})
//}
//
//func Test_NewUpdateRequest(t *testing.T) {
//	t.Parallel()
//
//	t.Run("when userID is non-zero, it returns an UpdateRequest", func(t *testing.T) {
//		t.Parallel()
//
//		expectedReq := UpdateRequest{
//			userID:   defaultUserID,
//			email:    option.None[tidy.EmailAddress](),
//			bio:      option.None[Bio](),
//			imageURL: option.None[URL](),
//			passwordHash:   option.None[PasswordHash](),
//		}
//
//		gotReq, gotErr := NewUpdateRequest(
//			defaultUserID,
//			option.None[tidy.EmailAddress](),
//			option.None[Bio](),
//			option.None[URL](),
//			option.None[PasswordHash](),
//		)
//
//		assert.NoError(t, gotErr)
//		assert.Equal(t, expectedReq, gotReq)
//	})
//
//	t.Run("when userID is empty, it returns a ZeroValueError", func(t *testing.T) {
//		t.Parallel()
//
//		gotReq, gotErr := NewUpdateRequest(
//			uuid.UUID{},
//			option.None[tidy.EmailAddress](),
//			option.None[Bio](),
//			option.None[URL](),
//			option.None[PasswordHash](),
//		)
//
//		var zeroValueErr *tidy.ZeroValueError
//		assert.ErrorAs(t, gotErr, &zeroValueErr)
//		assert.Empty(t, gotReq)
//	})
//}
