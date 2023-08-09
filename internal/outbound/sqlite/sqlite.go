package sqlite

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/angusgmorrison/realworld/internal/outbound/sqlite/sqlc"
	"github.com/golang-migrate/migrate/v4"
	migratesqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/golang-migrate/migrate/v4/source/iofs" // register the iofs source
	"os"
	"path/filepath"
)

//go:embed migrations/*.sql
var migrations embed.FS

// SQLite is an SQLite3 database with an open connection.
type SQLite struct {
	innerDB *sql.DB
	queries *sqlc.Queries
}

// New opens or creates an SQLite database at `dbPath` and runs all migrations,
// returning the DB instance.
func New(dbPath string) (*SQLite, error) {
	sanitizedPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("sanitize path %q: %w", dbPath, err)
	}

	if err := createFileIfNotExists(sanitizedPath); err != nil {
		return nil, fmt.Errorf("create DB: %w", err)
	}

	db, err := sql.Open("sqlite3", sanitizedPath)
	if err != nil {
		return nil, fmt.Errorf("open DB at %q: %w", sanitizedPath, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping DB: %w", err)
	}

	sqlite := &SQLite{
		innerDB: db,
		queries: sqlc.New(db),
	}

	if err := sqlite.migrate(); err != nil {
		return nil, err
	}

	return sqlite, nil
}

// Close closes the database connection.
func (db *SQLite) Close() error {
	if err := db.innerDB.Close(); err != nil {
		return fmt.Errorf("close underlying DB: %w", err)
	}
	return nil
}

func createFileIfNotExists(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat file at %q: %w", path, err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file at %q: %w", path, err)
	}

	if err = file.Close(); err != nil {
		return fmt.Errorf("close newly created file at %q: %w", path, err)
	}

	return nil
}

// Migrate runs all up migrations.
func (db *SQLite) migrate() error {
	migrator, err := newMigrator(db.innerDB)
	if err != nil {
		return err
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate DB: %w", err)
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
