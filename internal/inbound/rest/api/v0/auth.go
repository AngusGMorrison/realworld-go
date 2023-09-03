package v0

import (
	"crypto/rsa"
	"fmt"
	"github.com/angusgmorrison/realworld-go/internal/testutil"
	"time"

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

// userIDKey is the context key under which the current user ID, is any, is stored.
const userIDKey userIDKeyT = 0

// NewRS256JWTAuthMiddleware wraps Fiber's JWT middleware, parsing the current
// user ID from the JWT claims and setting it on the request context.
func NewRS256JWTAuthMiddleware(publicKey *rsa.PublicKey) fiber.Handler {
	return jwtware.New(jwtware.Config{
		AuthScheme:     "Token", // required by the RealWorld spec
		ContextKey:     requestJWTKey,
		ErrorHandler:   handleError,
		SigningKey:     jwtware.SigningKey{JWTAlg: jwtware.RS256, Key: publicKey},
		SuccessHandler: setSubjectOnContext,
	})
}

func handleError(_ *fiber.Ctx, err error) error {
	return NewUnauthorizedError("Invalid or missing authentication token", err)
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

// currentUserIDFromContext attempts to retrieve the current user ID from the request
// context. The boolean value is true if it is set, and false otherwise.
func currentUserIDFromContext(c *fiber.Ctx) (uuid.UUID, bool) {
	userID, _ := c.Locals(userIDKey).(uuid.UUID)
	if userID == uuid.Nil {
		return uuid.Nil, false
	}

	return userID, true
}

// mustGetCurrentUserIDFromContext retrieves the current user ID from the request context,
// panicking if it is not set.
func mustGetCurrentUserIDFromContext(c *fiber.Ctx) uuid.UUID {
	userID, ok := currentUserIDFromContext(c)
	if !ok {
		panic("current user ID not set on request context")
	}
	return userID
}

// currentJWTFromContext attempts to retrieve the current JWT from the request context.
// The boolean value is true if it is set, and false otherwise.
func currentJWTFromContext(c *fiber.Ctx) (JWT, bool) {
	token, ok := c.Locals(requestJWTKey).(*jwt.Token)
	if !ok {
		return JWT(""), false
	}

	return JWT(token.Raw), true
}

// mustGetCurrentJWTFromContext retrieves the current JWT from the request context,
// panicking if it is not set.
func mustGetCurrentJWTFromContext(c *fiber.Ctx) JWT {
	token, ok := currentJWTFromContext(c)
	if !ok {
		panic("current JWT not set on request context")
	}
	return token
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
