package v0

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"github.com/angusgmorrison/realworld-go/internal/testutil"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

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
