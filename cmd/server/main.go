// Bootstrap the realworld web server.
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/angusgmorrison/realworld-go/internal/config"
	"github.com/angusgmorrison/realworld-go/internal/domain/user"
	"github.com/angusgmorrison/realworld-go/internal/inbound/rest"
	"github.com/angusgmorrison/realworld-go/internal/outbound/sqlite"
)

func main() {
	// Delegate to `run`, allowing us to exit the program from a single location,
	// ensuring all deferred functions are executed.
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
		if closeErr := db.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	userService := user.NewService(db)

	serverConfig := rest.Config{
		AppName:          cfg.AppName,
		ReadTimeout:      cfg.ReadTimeout,
		WriteTimeout:     cfg.WriteTimeout,
		EnableStackTrace: cfg.EnableStackTrace,
		JwtCfg: rest.JWTConfig{
			RS265PrivateKey: cfg.JWTPrivateKey(),
			TTL:             cfg.JwtTTL,
			Issuer:          cfg.JwtIssuer,
		},
	}

	server := rest.NewServer(serverConfig, userService)

	if err = server.Listen(cfg.ServerAddress()); err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.ServerAddress(), err)
	}

	return nil
}
