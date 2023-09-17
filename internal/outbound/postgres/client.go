package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/angusgmorrison/realworld-go/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/golang-migrate/migrate/v4/source/iofs" // register the iofs source

	"github.com/angusgmorrison/realworld-go/internal/outbound/postgres/sqlc"
)

const migrationsPath = "migrations"

//go:embed migrations/*.sql
var migrations embed.FS

// queries describes all the queries that can be run against the database. It
// mirrors the generated sqlc code, allowing database errors to be mocked.
type queries interface {
	CreateUser(ctx context.Context, params sqlc.CreateUserParams) (sqlc.User, error)
	DeleteUser(ctx context.Context, id string) error
	GetUserByEmail(ctx context.Context, email string) (sqlc.GetUserByEmailRow, error)
	GetUserById(ctx context.Context, id string) (sqlc.GetUserByIdRow, error)
	UpdateUser(ctx context.Context, params sqlc.UpdateUserParams) (sqlc.User, error)
}

// Client is a Postgres client.
type Client struct {
	db      *sql.DB
	queries queries
}

// New returns a [Client] instance connected to the server at `url`, running any
// migrations that have not yet been applied.
func New(cfg config.Config) (*Client, error) {
	url := postgresURL(cfg)
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("open Postgres DB at %q: %w", url, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping DB at %q: %w", url, err)
	}

	client := &Client{
		db:      db,
		queries: sqlc.New(db),
	}

	if err := client.migrate(); err != nil {
		return nil, err
	}

	return client, nil
}

// Close closes the database connection.
func (c *Client) Close() error {
	if err := c.db.Close(); err != nil {
		return fmt.Errorf("close DB: %w", err)
	}
	return nil
}

// Migrate runs all up migrations.
func (c *Client) migrate() error {
	migrator, err := newMigrator(c.db)
	if err != nil {
		return err
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate DB: %w", err)
	}

	return nil
}

func newMigrator(db *sql.DB) (*migrate.Migrate, error) {
	source, err := iofs.New(migrations, migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("create io.FS migration source: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("create Client migration driver from DB instance: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("create migrator from io.FS source and Client driver: %w", err)
	}

	return m, nil
}

func postgresURL(cfg config.Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DbSslMode,
	)
}
