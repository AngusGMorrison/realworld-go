package user

import (
	"github.com/angusgmorrison/logfusc"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func RandomEmailAddressCandidate() string {
	return gofakeit.Email()
}

func RandomUsernameCandidate() string {
	return gofakeit.Regex(usernamePattern)
}

func RandomPasswordCandidate() logfusc.Secret[string] {
	length := rand.Intn(PasswordMaxLen-PasswordMinLen) + PasswordMinLen
	raw := gofakeit.Password(true, true, true, true, true, length)
	return logfusc.NewSecret(raw)
}

func RandomBioCandidate() string {
	paragraphs := rand.Intn(4) + 1
	sentences := rand.Intn(2) + 1
	words := rand.Intn(10) + 1
	return gofakeit.LoremIpsumParagraph(paragraphs, sentences, words, " ")
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

func RandomBio(t *testing.T) Bio {
	t.Helper()

	return Bio(RandomBioCandidate())
}
