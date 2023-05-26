package user

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/angusgmorrison/logfusc"
	"github.com/angusgmorrison/realworld/pkg/tidy"
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
// Embeds [logfusc.Secret], allowing it to be deserialized directly from JSON
// and logged safely.
type PasswordCandidate struct {
	logfusc.Secret[string]
}

// NewPasswordCandidate wraps an unvalidated password in a [PasswordCandidate].
func NewPasswordCandidate(password string) PasswordCandidate {
	return PasswordCandidate{Secret: logfusc.NewSecret(password)}
}

// NonZero returns nil in all cases, since it is either the parsing of a
// PasswordCandidate to a [PasswordHash] or the comparison of a
// PasswordCandidate with a [PasswordHash] that determines whether the password
// is valid.
func (pc PasswordCandidate) NonZero() error {
	return nil
}

// PasswordHash represents a validated and hashed password.
type PasswordHash struct {
	inner logfusc.Secret[[]byte]
}

// NonZero satisfies [tidy.Strict].
func (ph PasswordHash) NonZero() error {
	if ph.inner.Expose() == nil {
		return &tidy.ZeroValueError{ZeroStrict: ph}
	}
	return nil
}

// ErrPasswordTooShort is returned when a password is too short.
var ErrPasswordTooShort = fmt.Errorf("password must be at least %d bytes long", passwordMinLen)

// ErrPasswordTooLong is returned when a password exceeds bcrypt's maximum
// hashable length.
var ErrPasswordTooLong = fmt.Errorf("password must be at most %d bytes long", passwordMaxLen)

// ParsePassword returns a validated and hashed password from `candidate`, or
// an error if the password is invalid or unhashable.
func ParsePassword(candidate PasswordCandidate) (PasswordHash, error) {
	exposedPassword := candidate.Expose()
	if len(exposedPassword) < passwordMinLen {
		return PasswordHash{}, ErrPasswordTooShort
	}
	if len(exposedPassword) > passwordMaxLen {
		return PasswordHash{}, ErrPasswordTooLong
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
func tryAuthenticate(usr *User, password PasswordCandidate) error {
	if err := bcrypt.CompareHashAndPassword(usr.passwordHash.Expose(), []byte(password.Expose())); err != nil {
		return &AuthError{cause: err}
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

func (src *jwtSource) newWithSubject(sub tidy.UUIDv4) (JWT, error) {
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

func (src *jwtSource) subjectOf(token JWT) (tidy.UUIDv4, error) {
	parsedToken, err := src.parser.Parse(token.Expose(), src.keyFunc)
	if err != nil {
		return tidy.UUIDv4{}, fmt.Errorf("parse JWT: %w", err)
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return tidy.UUIDv4{}, fmt.Errorf("expected token claims to be jwt.MapClaims, got %T", parsedToken.Claims)
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return tidy.UUIDv4{}, fmt.Errorf("expected token sub field to be string, got %T", claims["sub"])
	}

	uuid, err := tidy.ParseUUIDv4(sub)
	if err != nil {
		return tidy.UUIDv4{}, fmt.Errorf("parse token sub field into UUIDv4: %w", err)
	}

	return uuid, nil
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
