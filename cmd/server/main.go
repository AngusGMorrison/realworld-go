// Bootstrap the realworld web server.
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/angusgmorrison/realworld-go/internal/outbound/postgres"

	"github.com/angusgmorrison/realworld-go/internal/inbound/rest/server"

	"github.com/angusgmorrison/realworld-go/internal/config"
	"github.com/angusgmorrison/realworld-go/internal/domain/user"
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

	db, err := postgres.New(postgres.NewURL(cfg))
	if err != nil {
		return fmt.Errorf("create Postgres client: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	userService := user.NewService(db)

	serverConfig := server.Config{
		AppName:          cfg.AppName,
		ReadTimeout:      cfg.ReadTimeout,
		WriteTimeout:     cfg.WriteTimeout,
		EnableStackTrace: cfg.EnableStackTrace,
		JwtCfg: server.JWTConfig{
			RS265PrivateKey: cfg.JWTPrivateKey(),
			TTL:             cfg.JwtTTL,
			Issuer:          cfg.JwtIssuer,
		},
		AllowOrigins: cfg.CORSAllowedOrigins,
	}

	srv := server.New(serverConfig, userService)

	if err = srv.Listen(cfg.ServerAddress()); err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.ServerAddress(), err)
	}

	return nil
}
