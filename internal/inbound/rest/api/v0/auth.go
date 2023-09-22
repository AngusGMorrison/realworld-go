package v0

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/angusgmorrison/realworld-go/internal/inbound/rest/middleware"

	"github.com/angusgmorrison/realworld-go/internal/testutil"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWT represents a JSON Web Token.
type JWT string

// JWTConfig is the minimal configuration required by [JWTProvider]s and JWT
// middleware.
type JWTConfig struct {
	RS256PrivateKey *rsa.PrivateKey
	TTL             time.Duration
	Issuer          string
}

func (cfg JWTConfig) PublicKey() *rsa.PublicKey {
	return &cfg.RS256PrivateKey.PublicKey
}

// Claims is an extensible set of JWT claims that includes all RFC 7519
// Registered Claims.
type Claims struct {
	jwt.RegisteredClaims
}

const requestJWTKey = "requestJWT"

type userIDKeyT int

// userIDKey is the context key under which the current user ID, if any, is stored.
const userIDKey userIDKeyT = 0

// Authentication wraps Fiber's JWT middleware, parsing the current
// user ID from the JWT claims and setting it on the request context.
//
// A request ID is expected to be present on the request context.
func Authentication(publicKey *rsa.PublicKey) fiber.Handler {
	return jwtware.New(jwtware.Config{
		AuthScheme:     "Token", // required by the RealWorld spec
		ContextKey:     requestJWTKey,
		ErrorHandler:   handleError,
		SigningKey:     jwtware.SigningKey{JWTAlg: jwtware.RS256, Key: publicKey},
		SuccessHandler: setSubjectOnContext,
	})
}

func handleError(c *fiber.Ctx, err error) error {
	requestID, ok := c.Locals(middleware.RequestIDKey).(string)
	if !ok {
		return fmt.Errorf("unhandled auth middleware error: request ID not set on context: %w", err)
	}

	return NewUnauthorizedError(requestID, err)
}

func setSubjectOnContext(c *fiber.Ctx) error {
	token := c.Locals(requestJWTKey).(*jwt.Token)

	sub, err := token.Claims.GetSubject()
	if err != nil {
		return fmt.Errorf("get subject from JWT claims:\n\tError: %w\n\tClaims: %#v", err, token.Claims)
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return fmt.Errorf("parse user ID string %q as UUID.\n\tError: %v\n\tClaims: %#v", sub, err, token.Claims)
	}

	c.Locals(userIDKey, userID)

	return c.Next()
}

var errMissingCurrentUserID = errors.New("current user ID not set on request context")

// currentUserIDFrom attempts to retrieve the current user ID from the request
// context.
func currentUserIDFrom(c *fiber.Ctx) (uuid.UUID, error) {
	userID, ok := c.Locals(userIDKey).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return uuid.Nil, errMissingCurrentUserID
	}

	return userID, nil
}

var errMissingCurrentJWT = errors.New("current JWT not set on request context")

// currentJWTFrom attempts to retrieve the current JWT from the request
// context.
func currentJWTFrom(c *fiber.Ctx) (JWT, error) {
	token, ok := c.Locals(requestJWTKey).(*jwt.Token)
	if !ok {
		return JWT(""), errMissingCurrentJWT
	}

	return JWT(token.Raw), nil
}

// JWTProvider is a source of signed JSON Web Tokens.
type JWTProvider interface {
	// TokenFor returns a signed JWT for the given subject.
	//
	// # Errors
	//	- Internal errors related to signing the token.
	TokenFor(subject uuid.UUID) (JWT, error)
}

type jwtProvider struct {
	privateKey *rsa.PrivateKey
	ttl        time.Duration
	issuer     string
	timeSource testutil.TimeSource
}

// NewJWTProvider returns the default JWT provider.
func NewJWTProvider(privateKey *rsa.PrivateKey, ttl time.Duration, issuer string) JWTProvider {
	return &jwtProvider{
		privateKey: privateKey,
		ttl:        ttl,
		issuer:     issuer,
		timeSource: testutil.StdTimeSource{},
	}
}

// TokenFor returns a signed JWT for the given subject.
func (jp *jwtProvider) TokenFor(subject uuid.UUID) (JWT, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jp.exp(),
			IssuedAt:  jp.iat(),
			Issuer:    jp.issuer,
			NotBefore: jp.nbf(),
			Subject:   subject.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(jp.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}

	return JWT(signedToken), nil
}

func (jp *jwtProvider) exp() *jwt.NumericDate {
	return jwt.NewNumericDate(jp.timeSource.Now().Add(jp.ttl))
}

func (jp *jwtProvider) iat() *jwt.NumericDate {
	return jwt.NewNumericDate(jp.timeSource.Now())
}

func (jp *jwtProvider) nbf() *jwt.NumericDate {
	return jwt.NewNumericDate(jp.timeSource.Now())
}
