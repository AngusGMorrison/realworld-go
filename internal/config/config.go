package config

import (
	"crypto/rsa"
	"fmt"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr                string        `split_words:"true" default:":8080"`
	ReadTimeout         time.Duration `split_words:"true" default:"5s"`
	WriteTimeout        time.Duration `split_words:"true" default:"5s"`
	JWTRSAPrivateKeyPEM string        `envconfig:"REALWORLD_JWT_RSA_PRIVATE_KEY_PEM" required:"true"`
	JWTTTL              time.Duration `envconfig:"REALWORLD_JWT_TTL" default:"24h"`
	DBBasename          string        `split_words:"true" default:"realworld.db"`
	VolumeMountPath     string        `split_words:"true" required:"true"`
	jwtPrivateKey       *rsa.PrivateKey
}

func New() (Config, error) {
	var cfg Config
	if err := envconfig.Process("realworld", &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (c *Config) AuthTokenRS256PrivateKey() (*rsa.PrivateKey, error) {
	return jwt.ParseRSAPrivateKeyFromPEM([]byte(c.JWTRSAPrivateKeyPEM))
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
