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
	// The name of the application.
	AppName string `split_words:"true" default:"realworld"`

	// The path to the directory containing runtime data, such as the DB file
	// and encryption keys.
	DataDir string `envconfig:"REALWORLD_DATA_MOUNT" split_words:"true" required:"true"`

	// The name of the SQLite DB file.
	DBBasename string `split_words:"true" default:"realworld.db"`

	// Enables stack tracing by panic recovery middleware.
	EnableStackTrace bool `split_words:"true" default:"false"`

	// The server host.
	Host string `split_words:"true" default:"0.0.0.0"`

	// The port to listen on.
	Port string `split_words:"true" default:"8080"`

	// The URL to be used as the issuer of JWTs signed by the server.
	JwtIssuer string `split_words:"true" required:"true"`

	// The name of the RSA private key PEM file used to generate JWTs.
	JwtRSAPrivateKeyPEMBasename string `envconfig:"REALWORLD_JWT_RSA_PRIVATE_KEY_PEM_BASENAME" split_words:"true" required:"true"`

	// The lifetime of JWTs issued by the server.
	JwtTTL time.Duration `envconfig:"REALWORLD_JWT_TTL" default:"24h"`

	// The read timeout for incoming requests to the server.
	ReadTimeout time.Duration `split_words:"true" default:"5s"`

	// The write timeout for outgoing responses from the server.
	WriteTimeout time.Duration `split_words:"true" default:"5s"`

	jwtRSAPrivateKey *rsa.PrivateKey
}

// New attempts to parse a `Config` object from the environment.
func New() (Config, error) {
	var cfg Config
	if err := envconfig.Process(envVarPrefix, &cfg); err != nil {
		return Config{}, fmt.Errorf("read config variables with prefix %q from the environment: %w", envVarPrefix, err)
	}

	privateKeyPath := filepath.Join(cfg.DataDir, cfg.JwtRSAPrivateKeyPEMBasename)
	privateKey, err := parseRSAPrivateKeyPEM(privateKeyPath)
	if err != nil {
		return Config{}, err
	}

	cfg.jwtRSAPrivateKey = privateKey

	return cfg, nil
}

// DBPath returns the absolute path to the database file.
func (c *Config) DBPath() string {
	return filepath.Join(c.DataDir, c.DBBasename)
}

// JWTPrivateKey parses the RSA private key PEM loaded from the environment into
// a private key object.
func (c *Config) JWTPrivateKey() *rsa.PrivateKey {
	return c.jwtRSAPrivateKey
}

func parseRSAPrivateKeyPEM(path string) (*rsa.PrivateKey, error) {
	pemBytes, err := os.ReadFile(path) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("read RSA private key PEM from %q: %w", path, err)
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("parse RSA private key PEM: %w", err)
	}

	return key, nil
}

// ServerAddress returns the address the server should listen on.
func (c *Config) ServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
