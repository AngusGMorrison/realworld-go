package server

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_JWTConfig_PublicKey(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	cfg := JWTConfig{RS265PrivateKey: privateKey}

	got := cfg.PublicKey()
	assert.Equal(t, &privateKey.PublicKey, got)
}
