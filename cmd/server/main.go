// Bootstrap the realworld web server.
package main

import (
	"fmt"
	"log"

	"github.com/angusgmorrison/realworld/internal/config"
	"github.com/angusgmorrison/realworld/internal/controller/rest"
	"github.com/angusgmorrison/realworld/internal/repository/sqlite"
	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/hashicorp/go-multierror"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() (err error) {
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := sqlite.New(cfg.DBPath())
	if err != nil {
		return fmt.Errorf("open DB at %q: %w", cfg.DBPath(), err)
	}
	defer func() {
		if dbErr := db.Close(); dbErr != nil {
			err = multierror.Append(err, dbErr)
		}
	}()

	jwtPrivateKey, err := cfg.JWTPrivateKey()
	if err != nil {
		return fmt.Errorf("load JWT private key: %w", err)
	}

	userService := user.NewService(db, jwtPrivateKey, cfg.JwtTtl)

	jwtPublicKey, err := cfg.JWTPublicKey()
	if err != nil {
		return fmt.Errorf("load JWT public key: %w", err)
	}

	srv := rest.NewServer(
		userService,
		jwtPublicKey,
		&rest.ReadTimeoutOption{Timeout: cfg.ReadTimeout},
		&rest.WriteTimeoutOption{Timeout: cfg.WriteTimeout},
	)
	if err := srv.Listen(cfg.ServerAddress()); err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.ServerAddress(), err)
	}

	return nil
}
