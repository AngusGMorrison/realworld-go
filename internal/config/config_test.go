// nolint:paralleltest
package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	t.Run("with all variables set, it loads correctly", func(t *testing.T) {
		env, err := newEnv()
		require.NoError(t, err)
		defer func() {
			_ = os.Remove(filepath.Join(os.TempDir(), env["REALWORLD_JWT_RSA_PRIVATE_KEY_PEM_BASENAME"]))
		}()

		for k, v := range env {
			t.Setenv(k, v)
		}

		wantEnableStackTrace, err := strconv.ParseBool(env["REALWORLD_ENABLE_STACK_TRACE"])
		require.NoError(t, err)

		cfg, err := New()
		assert.NoError(t, err)
		assert.Equal(t, env["REALWORLD_APP_NAME"], cfg.AppName)
		assert.Equal(t, env["REALWORLD_CORS_ALLOWED_ORIGINS"], cfg.CORSAllowedOrigins)
		assert.Equal(t, env["REALWORLD_DATA_MOUNT"], cfg.DataDir)
		assert.Equal(t, env["REALWORLD_DB_HOST"], cfg.DBHost)
		assert.Equal(t, env["REALWORLD_DB_PASSWORD"], cfg.DBPassword)
		assert.Equal(t, env["REALWORLD_DB_PORT"], cfg.DBPort)
		assert.Equal(t, env["REALWORLD_DB_NAME"], cfg.DBName)
		assert.Equal(t, env["REALWORLD_DB_SSL_MODE"], cfg.DbSslMode)
		assert.Equal(t, env["REALWORLD_DB_USER"], cfg.DBUser)
		assert.Equal(t, wantEnableStackTrace, cfg.EnableStackTrace)
		assert.Equal(t, env["REALWORLD_HOST"], cfg.Host)
		assert.Equal(t, env["REALWORLD_PORT"], cfg.Port)
		assert.Equal(t, env["REALWORLD_JWT_ISSUER"], cfg.JwtIssuer)
		assert.Equal(t, env["REALWORLD_JWT_RSA_PRIVATE_KEY_PEM_BASENAME"], cfg.JwtRSAPrivateKeyPEMBasename)
	})

	t.Run("missing any env var, it returns an error", func(t *testing.T) {
		testCases := []string{
			"REALWORLD_APP_NAME",
			"REALWORLD_CORS_ALLOWED_ORIGINS",
			"REALWORLD_DATA_MOUNT",
			"REALWORLD_DB_HOST",
			"REALWORLD_DB_PASSWORD",
			"REALWORLD_DB_PORT",
			"REALWORLD_DB_NAME",
			"REALWORLD_DB_SSL_MODE",
			"REALWORLD_DB_USER",
			"REALWORLD_ENABLE_STACK_TRACE",
			"REALWORLD_HOST",
			"REALWORLD_PORT",
			"REALWORLD_JWT_ISSUER",
			"REALWORLD_JWT_RSA_PRIVATE_KEY_PEM_BASENAME",
		}

		for _, tc := range testCases {
			t.Run(tc, func(t *testing.T) {
				env, err := newEnv()
				require.NoError(t, err)
				defer func() {
					_ = os.Remove(filepath.Join(os.TempDir(), env["REALWORLD_JWT_RSA_PRIVATE_KEY_PEM_BASENAME"]))
				}()

				var oldValue string
				defer func() { _ = os.Setenv(tc, oldValue) }()
				for k, v := range env {
					if k == tc {
						oldValue = os.Getenv(k)
						_ = os.Unsetenv(k)
						continue
					}

					t.Setenv(k, v)
				}

				_, err = New()
				assert.Error(t, err)
			})
		}
	})
}

func newEnv() (map[string]string, error) {
	rsaPEMFileName, err := generateTempRSAKeyPEMFile()
	if err != nil {
		return nil, fmt.Errorf("generate RSA private key PEM file: %w", err)
	}

	return map[string]string{
		"REALWORLD_APP_NAME":                         "myapp",
		"REALWORLD_CORS_ALLOWED_ORIGINS":             "http://myapp:3000",
		"REALWORLD_DATA_MOUNT":                       os.TempDir(),
		"REALWORLD_DB_HOST":                          "https://db.com",
		"REALWORLD_DB_PASSWORD":                      "secret",
		"REALWORLD_DB_PORT":                          "1111",
		"REALWORLD_DB_NAME":                          "realworld",
		"REALWORLD_DB_SSL_MODE":                      "full",
		"REALWORLD_DB_USER":                          "db_user",
		"REALWORLD_ENABLE_STACK_TRACE":               "true",
		"REALWORLD_HOST":                             "172.43.1.34",
		"REALWORLD_PORT":                             "9001",
		"REALWORLD_JWT_ISSUER":                       "https://myapp.com",
		"REALWORLD_JWT_RSA_PRIVATE_KEY_PEM_BASENAME": filepath.Base(rsaPEMFileName),
		"REALWORLD_JWT_TTL":                          "1h",
		"REALWORLD_READ_TIMEOUT":                     "10s",
		"REALWORLD_WRITE_TIMEOUT":                    "10s",
	}, nil
}

func generateTempRSAKeyPEMFile() (string, error) { // Generate a new RSA private key with 2048 bits
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", fmt.Errorf("generate RSA private key: %w", err)
	}

	// Encode the private key to the PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyFile, err := os.CreateTemp(os.TempDir(), "realworld-go-test-*.key")
	if err != nil {
		return "", fmt.Errorf("create private key file: %w", err)
	}
	defer func() { _ = privateKeyFile.Close() }()

	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return "", fmt.Errorf("encode private key to PEM: %w", err)
	}

	return privateKeyFile.Name(), nil
}

func Test_Config_ServerAddress(t *testing.T) {
	t.Parallel()

	cfg := Config{Host: "localhost", Port: "9001"}
	want := "localhost:9001"

	assert.Equal(t, want, cfg.ServerAddress())
}
