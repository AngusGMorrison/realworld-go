package users

import (
	"crypto/rsa"
	"fmt"
	"github.com/angusgmorrison/logfusc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"time"
)

// JWT represents a JSON Web Token.
//
// As an alias for [logfusc.Secret], it can be deserialized directly from JSON.
type JWT = logfusc.Secret[string]

// JWTProvider is a source of signed JSON Web Tokens.
type JWTProvider interface {
	FromSubject(sub uuid.UUID) (JWT, error)
}

type jwtProvider struct {
	privateKey *rsa.PrivateKey
	ttl        time.Duration
}

func NewJWTProvider(privateKey *rsa.PrivateKey, ttl time.Duration) JWTProvider {
	return &jwtProvider{
		privateKey: privateKey,
		ttl:        ttl,
	}
}

func (src *jwtProvider) FromSubject(sub uuid.UUID) (JWT, error) {
	claims := jwt.MapClaims{
		"sub": sub,
		"exp": time.Now().Add(src.ttl).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(src.privateKey)
	if err != nil {
		return JWT{}, fmt.Errorf("sign JWT: %w", err)
	}

	return logfusc.NewSecret(signedToken), nil
}
