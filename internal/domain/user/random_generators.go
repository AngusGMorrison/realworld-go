// nolint:gosec
package user

import (
	"math/rand"
	"testing"

	"github.com/angusgmorrison/realworld-go/pkg/etag"
	"github.com/google/uuid"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"

	"github.com/angusgmorrison/realworld-go/pkg/option"
)

func RandomEmailAddressCandidate() string {
	return gofakeit.Email()
}

func RandomUsernameCandidate() string {
	return gofakeit.Regex(usernamePattern)
}

func RandomPasswordCandidate() string {
	length := rand.Intn(PasswordMaxLen-PasswordMinLen) + PasswordMinLen
	raw := gofakeit.Password(true, true, true, true, true, length)
	return raw
}

func RandomBio() Bio {
	paragraphs := rand.Intn(4) + 1
	sentences := rand.Intn(2) + 1
	words := rand.Intn(10) + 1
	loremIpsum := gofakeit.LoremIpsumParagraph(paragraphs, sentences, words, " ")
	return Bio(loremIpsum)
}

func RandomURLCandidate() string {
	return gofakeit.URL()
}

func RandomEmailAddress(t *testing.T) EmailAddress {
	t.Helper()

	email, err := ParseEmailAddress(RandomEmailAddressCandidate())
	require.NoError(t, err)

	return email
}

func RandomUsername(t *testing.T) Username {
	t.Helper()

	username, err := ParseUsername(RandomUsernameCandidate())
	require.NoError(t, err)

	return username
}

func RandomPasswordHash(t *testing.T) PasswordHash {
	t.Helper()

	password := RandomPasswordCandidate()
	hash, err := ParsePassword(password)
	require.NoError(t, err)

	return hash
}

func RandomURL(t *testing.T) URL {
	t.Helper()

	url, err := ParseURL(RandomURLCandidate())
	require.NoError(t, err)

	return url
}

func RandomOption[T any](t *testing.T) option.Option[T] {
	t.Helper()

	if rand.Intn(2) == 0 {
		switch any(*new(T)).(type) {
		case EmailAddress:
			email := any(RandomEmailAddress(t)).(T)
			return option.Some(email)
		case Username:
			username := any(RandomUsername(t)).(T)
			return option.Some(username)
		case PasswordHash:
			password := any(RandomPasswordHash(t)).(T)
			return option.Some(password)
		case URL:
			url := any(RandomURL(t)).(T)
			return option.Some(url)
		case Bio:
			bio := any(RandomBio()).(T)
			return option.Some(bio)
		default:
			require.FailNow(
				t,
				"Unsupported type passed to RandomOption",
				"RandomOption does not support type %T",
				any(*new(T)),
			)
		}
	}

	return option.None[T]()
}

func RandomOptionFromInstance[T any](instance T) option.Option[T] {
	if rand.Intn(2) == 0 {
		return option.Some(instance)
	}

	return option.None[T]()
}

func RandomRegistrationRequest(t *testing.T) *RegistrationRequest {
	t.Helper()

	username := RandomUsername(t)
	email := RandomEmailAddress(t)
	password := RandomPasswordHash(t)
	return NewRegistrationRequest(username, email, password)
}

func RandomAuthRequest(t *testing.T) *AuthRequest {
	t.Helper()

	email := RandomEmailAddress(t)
	passwordCandidate := RandomPasswordCandidate()
	return NewAuthRequest(email, passwordCandidate)
}

func RandomUpdateRequest(t *testing.T) *UpdateRequest {
	t.Helper()

	id := uuid.New()
	eTag := etag.Random()
	email := RandomOption[EmailAddress](t)
	password := RandomOption[PasswordHash](t)
	bio := RandomOption[Bio](t)
	image := RandomOption[URL](t)
	return NewUpdateRequest(id, eTag, email, password, bio, image)
}

func RandomUser(t *testing.T) *User {
	t.Helper()

	id := uuid.New()
	eTag := etag.Random()
	username := RandomUsername(t)
	email := RandomEmailAddress(t)
	password := RandomPasswordHash(t)
	bio := RandomOption[Bio](t)
	image := RandomOption[URL](t)
	return NewUser(id, eTag, username, email, password, bio, image)
}
