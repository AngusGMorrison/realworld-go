// Package config generates an application config object from the environment.
package config

import (
	"crypto/rsa"
	"fmt"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/kelseyhightower/envconfig"
)

const envVarPrefix = "REALWORLD"

// Config represents the complete configuration settings for the application.
type Config struct {
	Addr                string        `split_words:"true" default:":8080"`
	ReadTimeout         time.Duration `split_words:"true" default:"5s"`
	WriteTimeout        time.Duration `split_words:"true" default:"5s"`
	JWTRSAPrivateKeyPEM string        `envconfig:"REALWORLD_JWT_RSA_PRIVATE_KEY_PEM" required:"true"`
	JWTTTL              time.Duration `envconfig:"REALWORLD_JWT_TTL" default:"24h"`
	DBBasename          string        `split_words:"true" default:"realworld.db"`
	VolumeMountPath     string        `split_words:"true" required:"true"`
}

// New attempts to parse a `Config` object from the environment.
func New() (Config, error) {
	var cfg Config
	if err := envconfig.Process(envVarPrefix, &cfg); err != nil {
		return cfg, fmt.Errorf("read config variables with prefix %q from the environment: %w", envVarPrefix, err)
	}

	return cfg, nil
}

// AuthTokenRS256PrivateKey parses the RSA private key PEM loaded from the
// environment into a private key object.
func (c *Config) AuthTokenRS256PrivateKey() (*rsa.PrivateKey, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(c.JWTRSAPrivateKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("parse REALWORLD_JWT_RSA_PRIVATE_KEY_PEM: %w", err)
	}
	return key, nil
}

// AuthTokenRS256PublicKey extracts a public counterpart of the private key
// parsed from the environment PEM string.
func (c *Config) AuthTokenRS256PublicKey() (*rsa.PublicKey, error) {
	key, err := c.AuthTokenRS256PrivateKey()
	if err != nil {
		return nil, err
	}

	publicKey, ok := key.Public().(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("expected RSA public key but got %T", key.Public())
	}

	return publicKey, nil
}

// DBPath returns the absolute path to the database file.
func (c *Config) DBPath() string {
	return filepath.Join(c.VolumeMountPath, c.DBBasename)
}
