package user

import (
	"crypto/rsa"
	"fmt"
	"github.com/google/uuid"
	"time"

	"github.com/angusgmorrison/logfusc"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	PasswordMinLen = 8
	PasswordMaxLen = 72
)

// PasswordHash represents a validated and hashed password.
type PasswordHash struct {
	inner logfusc.Secret[[]byte]
}

// ParsePassword returns a validated and hashed password from `candidate`, or
// an error if the password is invalid or unhashable.
func ParsePassword(candidate logfusc.Secret[string]) (PasswordHash, error) {
	exposedPassword := candidate.Expose()
	if len(exposedPassword) < PasswordMinLen {
		return PasswordHash{}, NewPasswordTooShortError()
	}
	if len(exposedPassword) > PasswordMaxLen {
		return PasswordHash{}, NewPasswordTooLongError()
	}

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(exposedPassword), bcrypt.DefaultCost)
	if err != nil {
		return PasswordHash{}, fmt.Errorf("hash password: %w", err)
	}

	return PasswordHash{inner: logfusc.NewSecret(hashBytes)}, nil
}

// NewPasswordHash wraps a hashed password in a [passwordHash].
func NewPasswordHash(hash logfusc.Secret[[]byte]) PasswordHash {
	return PasswordHash{inner: hash}
}

// Expose returns the hashed password as a byte slice.
func (ph PasswordHash) Expose() []byte {
	return ph.inner.Expose()
}

// tryAuthenticate compares `password` to the [User]'s hashed password,
// returning nil in the event of a match and [AuthError] otherwise.
func tryAuthenticate(hash PasswordHash, candidate logfusc.Secret[string]) error {
	if err := bcrypt.CompareHashAndPassword(hash.Expose(), []byte(candidate.Expose())); err != nil {
		return &AuthError{Cause: err}
	}
	return nil
}

// JWT represents a JSON Web Token.
//
// As an alias for [logfusc.Secret], it can be deserialized directly from JSON.
type JWT = logfusc.Secret[string]

type jwtSource struct {
	privateKey *rsa.PrivateKey
	parser     *jwt.Parser
	keyFunc    jwt.Keyfunc
	ttl        time.Duration
}

func (src *jwtSource) newWithSubject(sub uuid.UUID) (JWT, error) {
	claims := jwt.MapClaims{
		"sub": sub,
		"exp": time.Now().Add(src.ttl).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(src.privateKey)
	if err != nil {
		return logfusc.Secret[string]{}, fmt.Errorf("sign JWT: %w", err)
	}

	return logfusc.NewSecret(signedToken), nil
}

func (src *jwtSource) subjectOf(token JWT) (uuid.UUID, error) {
	parsedToken, err := src.parser.Parse(token.Expose(), src.keyFunc)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse JWT: %w", err)
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("expected token claims to be jwt.MapClaims, got %T", parsedToken.Claims)
	}

	rawSub, ok := claims["sub"].(string)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("expected token sub Field to be string, got %T", claims["sub"])
	}

	sub, err := uuid.Parse(rawSub)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse token sub Field into UUIDv4: %w", err)
	}

	return sub, nil
}

func (src *jwtSource) subjectsEqual(first, second JWT) bool {
	firstSub, err := src.subjectOf(first)
	if err != nil {
		return false
	}

	secondSub, err := src.subjectOf(second)
	if err != nil {
		return false
	}

	return firstSub == secondSub
}
