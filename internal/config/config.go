package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ReadTimeout  time.Duration `split_words:"true" default:"5s"`
	WriteTimeout time.Duration `split_words:"true" default:"5s"`
}

func New() (Config, error) {
	var cfg Config
	if err := envconfig.Process("realword", &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
