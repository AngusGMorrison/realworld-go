package user

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/angusgmorrison/realworld/pkg/logfusc"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	passwordMinLen = 8
	passwordMaxLen = 72
)

// PasswordCandidate represents a password that has not yet been validated or
// hashed.
//
// As an alias for [logfusc.Secret], it can be deserialized directly from JSON.
type PasswordCandidate = logfusc.Secret[string]

// passwordHash represents a validated and hashed password.
type passwordHash struct {
	inner logfusc.Secret[[]byte]
}

// ErrPasswordTooShort is returned when a password is too short.
var ErrPasswordTooShort = fmt.Errorf("password must be at least %d bytes long", passwordMinLen)

// ErrPasswordTooLong is returned when a password exceeds bcrypt's maximum
// hashable length.
var ErrPasswordTooLong = fmt.Errorf("password must be at most %d bytes long", passwordMaxLen)

// ParsepasswordHash returns a validated and hashed password from `candidate`, or
// an error if the password is invalid or unhashable.
func ParsePassword(candidate PasswordCandidate) (passwordHash, error) {
	exposedPassword := candidate.Expose()
	if len(exposedPassword) < passwordMinLen {
		return passwordHash{}, ErrPasswordTooShort
	}
	if len(exposedPassword) > passwordMaxLen {
		return passwordHash{}, ErrPasswordTooLong
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(exposedPassword), bcrypt.DefaultCost)
	if err != nil {
		return passwordHash{}, fmt.Errorf("hash password: %w", err)
	}

	return passwordHash{inner: logfusc.NewSecret(hashBytes)}, nil
}

// NewPasswordHash wraps a hashed password in a [passwordHash].
func NewPasswordHash(hash logfusc.Secret[[]byte]) passwordHash {
	return passwordHash{inner: hash}
}

// Expose returns the hashed password as a byte slice.
func (p passwordHash) Expose() []byte {
	return p.inner.Expose()
}

// tryAuthenticate compares `password` to the [User]'s hashed password,
// returning nil in the event of a match and [AuthError] otherwise.
func tryAuthenticate(usr *user, password PasswordCandidate) error {
	if err := bcrypt.CompareHashAndPassword(usr.passwordHash.Expose(), []byte(password.Expose())); err != nil {
		return &AuthError{cause: err}
	}
	return nil
}

// JWT represents a JSON Web Token.
//
// As an alias for [logfusc.Secret], it can be deserialized directly from JSON.
type JWT = logfusc.Secret[string]

func newJWT(key *rsa.PrivateKey, ttl time.Duration, sub string) (JWT, error) {
	claims := jwt.MapClaims{
		"sub": sub,
		"exp": time.Now().Add(ttl).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(key)
	if err != nil {
		return logfusc.Secret[string]{}, fmt.Errorf("sign JWT: %w", err)
	}

	return logfusc.NewSecret(signedToken), nil
}

// TODO: Replace with wrapped jwt object.
func jwtSubjectsEqual(first, second JWT, publicKey *rsa.PublicKey) bool {
	jwtParser := jwt.NewParser()
	keyFunc := func(_ *jwt.Token) (any, error) {
		return publicKey, nil
	}
	parsedFirstToken, err := jwtParser.Parse(first.Expose(), keyFunc)
	if err != nil {
		return false
	}
	firstClaims, ok := parsedFirstToken.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	parsedOtherToken, err := jwtParser.Parse(second.Expose(), keyFunc)
	if err != nil {
		return false
	}
	secondClaims, ok := parsedOtherToken.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	return firstClaims["sub"] == secondClaims["sub"]
}
