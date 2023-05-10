package user

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/angusgmorrison/realworld/pkg/primitive"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// RequiredValidatingPassword tags required passwords for validation. It should
// be embedded wherever a password MUST be present and MUST be validated before
// use.
//
// Changes to password validation tags MUST be kept in sync with
// OptionalValidatingPassword.
type RequiredValidatingPassword struct {
	Password primitive.SensitiveString `json:"password" validate:"required,min=8,max=72"` // bcrypt max password length is 72 bytes
}

// HashPassword returns the hashed password.
func (rvp RequiredValidatingPassword) HashPassword() (primitive.SensitiveString, error) {
	return bcryptHash(rvp.Password)
}

// OptionalValidatingPassword tags optional passwords for validation. It should
// be embedded wherever a password MUST be validated IFF present.
//
// Changes to password validation tags MUST be kept in sync with
// RequiredValidatingPassword.
type OptionalValidatingPassword struct {
	Password *primitive.SensitiveString `json:"password" validate:"omitempty,min=8,max=72"`
}

// HashPassword returns the hashed password.
func (ovp OptionalValidatingPassword) HashPassword() (primitive.SensitiveString, error) {
	return bcryptHash(*ovp.Password)
}

func bcryptHash(password primitive.SensitiveString) (primitive.SensitiveString, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return primitive.SensitiveString(hashBytes), nil
}

func newJWT(key *rsa.PrivateKey, ttl time.Duration, sub string) (string, error) {
	claims := jwt.MapClaims{
		"sub": sub,
		"exp": time.Now().Add(ttl).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}

	return signedToken, nil
}

func jwtSubjectsEqual(first, second string, publicKey *rsa.PublicKey) bool {
	jwtParser := jwt.NewParser()
	keyFunc := func(_ *jwt.Token) (any, error) {
		return publicKey, nil
	}
	parsedFirstToken, err := jwtParser.Parse(first, keyFunc)
	if err != nil {
		return false
	}
	firstClaims, ok := parsedFirstToken.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	parsedOtherToken, err := jwtParser.Parse(second, keyFunc)
	if err != nil {
		return false
	}
	secondClaims, ok := parsedOtherToken.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	return firstClaims["sub"] == secondClaims["sub"]
}
