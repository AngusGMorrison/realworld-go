package middleware

import (
	"crypto/rsa"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type userIDKeyT int

const UserIDKey userIDKeyT = 0

// NewRS256Auth returns middleware wrapping Fiber's JWT middleware that parses
// the current user ID from the JWT claims and sets it on the request context.
// To simplify debugging, this middleware is best used in conjunction with the
// request-scoped logging middleware provided in this package.
func NewRS256Auth(key *rsa.PublicKey) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey:     key,
		SigningMethod:  "RS256",
		ErrorHandler:   DefaultRS256AuthFailureHandler,
		SuccessHandler: DefaultRS256AuthSuccessHandler,
	})
}

// DefaultRS256AuthFailureHandler is the default failure handler for the JWT
// middleware. It is exported as a convenience for testing.
func DefaultRS256AuthFailureHandler(c *fiber.Ctx, err error) error {
	return fiber.NewError(fiber.StatusUnauthorized)
}

// DefaultRS256AuthSuccessHandler is the default success handler for the JWT
// middleware. It is exported as a convenience for testing.
func DefaultRS256AuthSuccessHandler(c *fiber.Ctx) error {
	logger := GetLogger(c)
	// JWT and MapClaim are guaranteed to be present by jwtware; if not, panic is appropriate.
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)

	sub, ok := claims["sub"]
	if !ok {
		if logger != nil {
			logger.Printf("JWT claims missing \"sub\" field: %+v\n", claims)
		}
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	userID, err := uuid.Parse(sub.(string))
	if err != nil {
		if logger != nil {
			logger.Printf("Failed to parse user UUID from JWT claims: %v\n", err)
		}
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	c.Locals(UserIDKey, userID)

	return c.Next()
}

// CurrentUser returns the user ID from the request context. If no user ID is
// set (e.g. because the route is not authenticated), an empty UUID is returned.
func CurrentUser(c *fiber.Ctx) uuid.UUID {
	userID, _ := c.Locals(UserIDKey).(uuid.UUID)
	return userID
}
