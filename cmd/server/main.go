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

	if err := db.Migrate(); err != nil {
		return fmt.Errorf("migrate DB: %w", err)
	}

	jwtPrivateKey, err := cfg.AuthTokenRS256PrivateKey()
	if err != nil {
		return fmt.Errorf("load JWT private key: %w", err)
	}

	userService := user.NewService(db, jwtPrivateKey, cfg.JWTTTL)

	jwtPublicKey, err := cfg.AuthTokenRS256PublicKey()
	if err != nil {
		return fmt.Errorf("load JWT public key: %w", err)
	}

	srv := rest.NewServer(
		userService,
		jwtPublicKey,
		&rest.ReadTimeoutOption{Timeout: cfg.ReadTimeout},
		&rest.WriteTimeoutOption{Timeout: cfg.WriteTimeout},
	)
	if err := srv.Listen(cfg.Addr); err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.Addr, err)
	}

	return nil
}
