package v0

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/angusgmorrison/realworld-go/internal/testutil"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewRS256JWTAuthMiddleware(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	t.Run("sets the subject on the request context when the token is valid", func(t *testing.T) {
		t.Parallel()

		provider := NewJWTProvider(privateKey, 15*time.Minute, "test")
		subject := uuid.New()
		token, err := provider.TokenFor(subject)
		require.NoError(t, err)

		authMiddleware := NewRS256JWTAuthMiddleware(&privateKey.PublicKey)
		app := fiber.New()
		app.Get("/", authMiddleware, func(c *fiber.Ctx) error {
			gotSubject, ok := currentUserIDFromContext(c)
			require.True(t, ok, "current user was not set on the request context")
			assert.Equal(t, subject, gotSubject)
			return nil
		})

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
		require.NoError(t, err)

		req.Header.Set("Authorization", "Token "+string(token))
		_, err = app.Test(req)
		require.NoError(t, err)
	})

	t.Run("returns an unauthorized error when the token is invalid", func(t *testing.T) {
		t.Parallel()

		app := fiber.New(fiber.Config{
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				var userFacingErr *UserFacingError
				require.ErrorAs(t, err, &userFacingErr)
				assert.Equal(t, http.StatusUnauthorized, userFacingErr.StatusCode)
				return nil
			},
		})
		app.Get("/", NewRS256JWTAuthMiddleware(&privateKey.PublicKey))

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
		require.NoError(t, err)

		req.Header.Set("Authorization", "Token invalid")
		_, err = app.Test(req)
		require.NoError(t, err)
	})

	t.Run("returns an unauthorized error when the token is missing", func(t *testing.T) {
		t.Parallel()

		app := fiber.New(fiber.Config{
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				var userFacingErr *UserFacingError
				require.ErrorAs(t, err, &userFacingErr)
				assert.Equal(t, http.StatusUnauthorized, userFacingErr.StatusCode)
				return nil
			},
		})
		app.Get("/", NewRS256JWTAuthMiddleware(&privateKey.PublicKey))

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
		require.NoError(t, err)

		_, err = app.Test(req)
		require.NoError(t, err)
	})
}

func Test_NewJWTProvider(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	ttl := 15 * time.Minute
	issuer := "test"
	wantProvider := &jwtProvider{
		privateKey: privateKey,
		ttl:        ttl,
		issuer:     issuer,
		timeSource: testutil.StdTimeSource{},
	}

	gotProvider := NewJWTProvider(privateKey, ttl, issuer)
	assert.Equal(t, wantProvider, gotProvider)
}

func Test_jwtProvider_tokenFor(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	ttl := 15 * time.Minute
	issuer := "test"
	subject := uuid.New()
	now := time.Now()
	provider := &jwtProvider{
		privateKey: privateKey,
		ttl:        ttl,
		issuer:     issuer,
		timeSource: testutil.FixedTimeSource{Time: now},
	}
	wantClaims := jwt.MapClaims{
		"exp": jwt.NewNumericDate(now.Add(ttl)),
		"iat": jwt.NewNumericDate(now).Unix(),
		"iss": issuer,
		"nbf": jwt.NewNumericDate(now).Unix(),
		"sub": subject.String(),
	}

	// Due to the reduction in time precision that occurs when JWTs are signed,
	// we need to compare the JSON representations of the claims rather than
	// the claims themselves.
	wantClaimsJson, err := json.Marshal(wantClaims)
	require.NoError(t, err)

	gotToken, err := provider.TokenFor(subject)
	require.NoError(t, err)

	parsedToken, err := jwt.Parse(string(gotToken), func(token *jwt.Token) (interface{}, error) {
		return privateKey.Public(), nil
	})
	require.NoError(t, err)

	gotClaims := parsedToken.Claims.(jwt.MapClaims)
	gotClaimsJson, err := json.Marshal(gotClaims)
	require.NoError(t, err)
	assert.JSONEq(t, string(wantClaimsJson), string(gotClaimsJson))
}
