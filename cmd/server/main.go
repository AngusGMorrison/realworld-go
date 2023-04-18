package main

import (
	"fmt"
	"log"
	"os"

	"github.com/angusgmorrison/realworld/internal/config"
	"github.com/angusgmorrison/realworld/internal/controller/rest"
	"github.com/angusgmorrison/realworld/internal/repository/sqlite"
	"github.com/angusgmorrison/realworld/internal/service/user"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	dbFile, err := os.Create(cfg.DBPath())
	if err != nil {
		return fmt.Errorf("create DB file at %q: %w", cfg.DBPath(), err)
	}
	dbFile.Close()

	db, err := sqlite.New(cfg.DBPath())
	if err != nil {
		return fmt.Errorf("open DB at %q: %w", cfg.DBPath(), err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		return fmt.Errorf("migrate DB: %w", err)
	}

	jwtPrivateKey, err := cfg.AuthTokenRS256PrivateKey()
	if err != nil {
		return fmt.Errorf("load JWT private key: %w", err)
	}

	userService := user.NewService(db, jwtPrivateKey, cfg.AuthTokenTTL)

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
