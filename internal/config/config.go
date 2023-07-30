// Package config generates an application config object from the environment.
package config

import (
	"crypto/rsa"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/kelseyhightower/envconfig"
)

const envVarPrefix = "REALWORLD"

// Config represents the complete configuration settings for the application.
type Config struct {
	// The port to listen on.
	Port string `split_words:"true" default:"8080"`

	// The server host.
	Host string `split_words:"true" default:"0.0.0.0"`

	// The read timeout for incoming requests to the server.
	ReadTimeout time.Duration `split_words:"true" default:"5s"`

	// The write timeout for outgoing responses from the server.
	WriteTimeout time.Duration `split_words:"true" default:"5s"`

	// The lifetime of JWTs issued by the server.
	JwtTtl time.Duration `envconfig:"REALWORLD_JWT_TTL" default:"24h"`

	// The path to the directory containing runtime data, such as the DB file
	// and encryption keys.
	DataDir string `split_words:"true" required:"true"`

	// The name of the SQLite DB file.
	DBBasename string `split_words:"true" default:"realworld.db"`

	// The name of the RS256 private key PEM file, used to generate JWTs.
	JWTRSAPrivateKeyPEMBasename string `envconfig:"REALWORLD_JWT_RSA_PRIVATE_KEY_PEM_BASENAME" split_words:"true" required:"true"`
}

// New attempts to parse a `Config` object from the environment.
func New() (Config, error) {
	var cfg Config
	if err := envconfig.Process(envVarPrefix, &cfg); err != nil {
		return cfg, fmt.Errorf("read config variables with prefix %q from the environment: %w", envVarPrefix, err)
	}

	return cfg, nil
}

func (c *Config) ServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// JWTPrivateKey parses the RSA private key PEM loaded from the environment into
// a private key object.
func (c *Config) JWTPrivateKey() (*rsa.PrivateKey, error) {
	pemBytes, err := os.ReadFile(c.JWTRSAPrivateKeyPEMPath())
	if err != nil {
		return nil, fmt.Errorf("read JWT private key PEM from %q: %w", c.JWTRSAPrivateKeyPEMPath(), err)
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("parse REALWORLD_JWT_RSA_PRIVATE_KEY_PEM: %w", err)
	}
	return key, nil
}

// JWTPublicKey extracts a public counterpart of the private key parsed from the
// environment PEM string.
func (c *Config) JWTPublicKey() (*rsa.PublicKey, error) {
	key, err := c.JWTPrivateKey()
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
	return filepath.Join(c.DataDir, c.DBBasename)
}

// JWTRSAPrivateKeyPEMPath returns the absolute path to the JWT RSA private key
// PEM file.
func (c *Config) JWTRSAPrivateKeyPEMPath() string {
	return filepath.Join(c.DataDir, c.JWTRSAPrivateKeyPEMBasename)
}
