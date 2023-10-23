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
	UserExists(ctx context.Context, id string) (bool, error)
}

// Client is a Postgres client.
type Client struct {
	db      *sql.DB
	queries queries
}

// New returns a [Client] instance connected to the server at `url`, running any
// migrations that have not yet been applied.
func New(url URL) (*Client, error) {
	db, err := sql.Open("postgres", url.Expose())
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

// URL is a Postgres connection URL.
type URL struct {
	host     string
	port     string
	dbName   string
	user     string
	password string
	sslMode  string
}

func NewURL(cfg config.Config) URL {
	return URL{
		host:     cfg.DBHost,
		port:     cfg.DBPort,
		dbName:   cfg.DBName,
		user:     cfg.DBUser,
		password: cfg.DBPassword,
		sslMode:  cfg.DbSslMode,
	}
}

// GoString returns a Go-syntax representation of the URL, with the password
// redacted.
func (u URL) GoString() string {
	return fmt.Sprintf(
		"postgres.URL{host:%q, port:%q, dbName:%q, user:%q, password:REDACTED, sslMode:%q}",
		u.host,
		u.port,
		u.dbName,
		u.user,
		u.sslMode,
	)
}

// String returns a connection string with the password redacted.
func (u URL) String() string {
	return fmt.Sprintf(
		"postgres://%s:REDACTED@%s:%s/%s?sslmode=%s",
		u.user,
		u.host,
		u.port,
		u.dbName,
		u.sslMode,
	)
}

// Expose returns a connection string with the password exposed.
func (u URL) Expose() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		u.user,
		u.password,
		u.host,
		u.port,
		u.dbName,
		u.sslMode,
	)
}
