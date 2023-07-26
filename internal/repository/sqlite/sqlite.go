package sqlite

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/angusgmorrison/realworld/internal/repository/sqlite/sqlc"
	"github.com/golang-migrate/migrate/v4"
	migratesqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/golang-migrate/migrate/v4/source/iofs" // register the iofs source
)

//go:embed migrations/*.sql
var migrations embed.FS

// SQLite is an SQLite3 database with an open connection.
type SQLite struct {
	innerDB *sql.DB
	queries *sqlc.Queries
}

// New creates a new SQLite database, opens a connection and pings the DB.
func New(dbPath string) (*SQLite, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open DB at %s: %w", dbPath, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping DB: %w", err)
	}

	return &SQLite{
		innerDB: db,
		queries: sqlc.New(db),
	}, nil
}

// Close closes the database connection.
func (db *SQLite) Close() error {
	if err := db.innerDB.Close(); err != nil {
		return fmt.Errorf("close underlying DB: %w", err)
	}
	return nil
}

// Migrate runs all up migrations.
func (db *SQLite) Migrate() error {
	migrator, err := newMigrator(db.innerDB)
	if err != nil {
		return err
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}

	return nil
}

// Rollback rolls back the last migration.
func (db *SQLite) Rollback() error {
	migrator, err := newMigrator(db.innerDB)
	if err != nil {
		return err
	}

	if err := migrator.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate one step down: %w", err)
	}

	return nil
}

func newMigrator(db *sql.DB) (*migrate.Migrate, error) {
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("create io.FS migration source: %w", err)
	}

	dbInstance, err := migratesqlite3.WithInstance(db, &migratesqlite3.Config{})
	if err != nil {
		return nil, fmt.Errorf("create SQLite migration driver from DB instance: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "sqlite3", dbInstance)
	if err != nil {
		return nil, fmt.Errorf("create migrator from io.FS source and SQLite driver: %w", err)
	}

	return m, nil
}
