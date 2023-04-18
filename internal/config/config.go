package config

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"path/filepath"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr                string        `split_words:"true" default:":8080"`
	ReadTimeout         time.Duration `split_words:"true" default:"5s"`
	WriteTimeout        time.Duration `split_words:"true" default:"5s"`
	AuthTokenRS256Pem   string        `envconfig:"REALWORLD_AUTH_TOKEN_RS256_PEM" required:"true"`
	AuthTokenTTL        time.Duration `split_words:"true" default:"24h"`
	DBBasename          string        `split_words:"true" default:"realworld.db"`
	VolumeMountPath     string        `split_words:"true" required:"true"`
	authTokenPrivateKey *rsa.PrivateKey
}

func New() (Config, error) {
	var cfg Config
	if err := envconfig.Process("realworld", &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (c *Config) AuthTokenRS256PrivateKey() (*rsa.PrivateKey, error) {
	if c.authTokenPrivateKey != nil {
		return c.authTokenPrivateKey, nil
	}

	key, err := x509.ParsePKCS1PrivateKey([]byte(c.AuthTokenRS256Pem))
	if err != nil {
		return nil, err
	}

	c.authTokenPrivateKey = key

	return key, nil
}

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

func (c *Config) DBPath() string {
	return filepath.Join(c.VolumeMountPath, c.DBBasename)
}
