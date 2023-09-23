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
	AppName string `split_words:"true" required:"true"`

	// A comma-separated list of allow origins for CORS requests.
	CORSAllowedOrigins string `envconfig:"REALWORLD_CORS_ALLOWED_ORIGINS" split_words:"true" required:"true"`

	// The path to the directory containing runtime data, such as the DB file
	// and encryption keys.
	DataDir string `envconfig:"REALWORLD_DATA_MOUNT" split_words:"true" required:"true"`

	// The database server hostname.
	DBHost string `split_words:"true" required:"true"`

	// The database password.
	DBPassword string `split_words:"true" required:"true"`

	// The database server port.
	DBPort string `split_words:"true" required:"true"`

	// The database to connect to.
	DBName string `split_words:"true" required:"true"`

	// The SSL mode for database connections.
	DbSslMode string `split_words:"true" required:"true"`

	// The database username.
	DBUser string `split_words:"true" required:"true"`

	// Enables stack tracing by panic recovery middleware.
	EnableStackTrace bool `split_words:"true" required:"true"`

	// The server host.
	Host string `split_words:"true" required:"true"`

	// The port to listen on.
	Port string `split_words:"true" required:"true"`

	// The URL to be used as the issuer of JWTs signed by the server.
	JwtIssuer string `split_words:"true" required:"true"`

	// The name of the RSA private key PEM file used to generate JWTs.
	JwtRSAPrivateKeyPEMBasename string `envconfig:"REALWORLD_JWT_RSA_PRIVATE_KEY_PEM_BASENAME" split_words:"true" required:"true"`

	// The lifetime of JWTs issued by the server.
	JwtTTL time.Duration `envconfig:"REALWORLD_JWT_TTL" required:"true"`

	// The read timeout for incoming requests to the server.
	ReadTimeout time.Duration `split_words:"true" required:"true"`

	// The write timeout for outgoing responses from the server.
	WriteTimeout time.Duration `split_words:"true" required:"true"`

	jwtRSAPrivateKey *rsa.PrivateKey
}

// New attempts to parse a `Config` object from the environment.
func New() (Config, error) {
	var cfg Config
	if err := envconfig.Process(envVarPrefix, &cfg); err != nil {
		return Config{}, fmt.Errorf("read config variables with prefix %q from the environment: %w", envVarPrefix, err)
	}

	privateKeyPath := filepath.Join(cfg.DataDir, cfg.JwtRSAPrivateKeyPEMBasename)
	privateKey, err := parseRSAKey(privateKeyPath)
	if err != nil {
		return Config{}, err
	}

	cfg.jwtRSAPrivateKey = privateKey

	return cfg, nil
}

// JWTPrivateKey parses the RSA private key PEM loaded from the environment into
// a private key object.
func (c *Config) JWTPrivateKey() *rsa.PrivateKey {
	return c.jwtRSAPrivateKey
}

func parseRSAKey(path string) (*rsa.PrivateKey, error) {
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
