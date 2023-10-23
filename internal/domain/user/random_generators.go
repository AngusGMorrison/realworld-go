// nolint:gosec
package testutil

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/angusgmorrison/realworld-go/internal/domain/user"

	"github.com/google/uuid"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"

	"github.com/angusgmorrison/realworld-go/pkg/option"
)

func RandomEmailAddressCandidate() string {
	return gofakeit.Email()
}

var usernamePattern = fmt.Sprintf(user.UsernamePatternTemplate, user.UsernameMinLen, user.UsernameMaxLen)

func RandomUsernameCandidate() string {
	return gofakeit.Regex(usernamePattern)
}

func RandomPasswordCandidate() string {
	length := rand.Intn(user.PasswordMaxLen-user.PasswordMinLen) + user.PasswordMinLen
	raw := gofakeit.Password(true, true, true, true, true, length)
	return raw
}

func RandomBio() user.Bio {
	paragraphs := rand.Intn(4) + 1
	sentences := rand.Intn(2) + 1
	words := rand.Intn(10) + 1
	loremIpsum := gofakeit.LoremIpsumParagraph(paragraphs, sentences, words, " ")
	return user.Bio(loremIpsum)
}

func RandomURLCandidate() string {
	return gofakeit.URL()
}

func RandomEmailAddress(t *testing.T) user.EmailAddress {
	t.Helper()

	email, err := user.ParseEmailAddress(RandomEmailAddressCandidate())
	require.NoError(t, err)

	return email
}

func RandomUsername(t *testing.T) user.Username {
	t.Helper()

	username, err := user.ParseUsername(RandomUsernameCandidate())
	require.NoError(t, err)

	return username
}

func RandomPasswordHash(t *testing.T) user.PasswordHash {
	t.Helper()

	password := RandomPasswordCandidate()
	hash, err := user.ParsePassword(password)
	require.NoError(t, err)

	return hash
}

func RandomURL(t *testing.T) user.URL {
	t.Helper()

	url, err := user.ParseURL(RandomURLCandidate())
	require.NoError(t, err)

	return url
}

type ETagGenerator func() user.ETag

func RandomETag(generator ...ETagGenerator) user.ETag {
	if len(generator) > 0 {
		return generator[0]()
	}
	return user.ETag(gofakeit.UUID())
}

func RandomOption[T any](t *testing.T) option.Option[T] {
	t.Helper()

	if rand.Intn(2) == 0 {
		switch any(*new(T)).(type) {
		case user.EmailAddress:
			email := any(RandomEmailAddress(t)).(T)
			return option.Some(email)
		case user.Username:
			username := any(RandomUsername(t)).(T)
			return option.Some(username)
		case user.PasswordHash:
			password := any(RandomPasswordHash(t)).(T)
			return option.Some(password)
		case user.URL:
			url := any(RandomURL(t)).(T)
			return option.Some(url)
		case user.Bio:
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

func RandomRegistrationRequest(t *testing.T) *user.RegistrationRequest {
	t.Helper()

	username := RandomUsername(t)
	email := RandomEmailAddress(t)
	password := RandomPasswordHash(t)
	return user.NewRegistrationRequest(username, email, password)
}

func RandomAuthRequest(t *testing.T) *user.AuthRequest {
	t.Helper()

	email := RandomEmailAddress(t)
	passwordCandidate := RandomPasswordCandidate()
	return user.NewAuthRequest(email, passwordCandidate)
}

func RandomUpdateRequest(t *testing.T) *user.UpdateRequest {
	t.Helper()

	id := uuid.New()
	eTag := RandomETag()
	email := RandomOption[user.EmailAddress](t)
	password := RandomOption[user.PasswordHash](t)
	bio := RandomOption[user.Bio](t)
	image := RandomOption[user.URL](t)
	return user.NewUpdateRequest(id, eTag, email, password, bio, image)
}

func RandomUser(t *testing.T) *user.User {
	t.Helper()

	id := uuid.New()
	eTag := RandomETag()
	username := RandomUsername(t)
	email := RandomEmailAddress(t)
	password := RandomPasswordHash(t)
	bio := RandomOption[user.Bio](t)
	image := RandomOption[user.URL](t)
	return user.NewUser(id, eTag, username, email, password, bio, image)
}
