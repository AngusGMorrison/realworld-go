package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/angusgmorrison/realworld/pkg/primitive"

	"github.com/golang-migrate/migrate/v4"
	migratesqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/golang-migrate/migrate/v4/source/iofs" // register the iofs source
	"github.com/google/uuid"
	sqlite3 "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrations embed.FS

// SQLite is an SQLite3 database with an open connection.
type SQLite struct {
	innerDB *sql.DB
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

	return &SQLite{db}, nil
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

// executor allows us to use either a *sql.DB or *sql.Tx interchangeably.
type executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// GetUserByID returns the [user.User] with the given ID, or
// [user.ErrUserNotFound] if no user exists with that ID.
func (db *SQLite) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return getUserByID(ctx, db.innerDB, id)
}

func getUserByID(ctx context.Context, ex executor, id uuid.UUID) (*user.User, error) {
	query := `SELECT id, email, username, bio, password_hash, image_url FROM users WHERE id = ?`
	row := ex.QueryRowContext(ctx, query, id)

	var usr user.User
	err := row.Scan(&usr.ID, &usr.Email, &usr.Username, &usr.Bio, &usr.PasswordHash, &usr.ImageURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("scan rows for user ID %s: %w", id, err)
	}
	return &usr, nil
}

// GetUserByEmail returns the [user.User] with the given email, or
// [user.ErrUserNotFound] if no user exists with that email.
func (db *SQLite) GetUserByEmail(ctx context.Context, email primitive.EmailAddress) (*user.User, error) {
	return getUserByEmail(ctx, db.innerDB, email)
}

func getUserByEmail(ctx context.Context, ex executor, email primitive.EmailAddress) (*user.User, error) {
	query := `SELECT id, email, username, bio, password_hash, image_url FROM users WHERE email = ?`
	row := ex.QueryRowContext(ctx, query, email)
	return newUserFromRow(row)
}

// CreateUser inserts a new [user.User] into the database, modifying the user
// in-place to includes its row ID. The modified user is returned.
//
// Returns either [user.ErrEmailRegistered] or [user.ErrUsernameTaken] if the
// the respective unique constraints are violated.
func (db *SQLite) CreateUser(ctx context.Context, usr *user.User) (*user.User, error) {
	return insertUser(ctx, db.innerDB, usr)
}

func insertUser(ctx context.Context, ex executor, usr *user.User) (*user.User, error) {
	id := uuid.New()
	query := `INSERT INTO users (id, email, username, password_hash, bio, image_url) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := ex.ExecContext(ctx, query, id, usr.Email, usr.Username, usr.PasswordHash, usr.Bio, usr.ImageURL)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			return nil, sqliteErrToDomain(sqliteErr)
		}
		return nil, fmt.Errorf("execute INSERT query for user %#v: %w", usr, err)
	}

	usr.ID = id
	return usr, nil
}

// UpdateUser updates the given [user.User] in the database, returning the
// updated user.
//
// Returns [user.ErrEmailRegistered] if the new email is already registered.
func (db *SQLite) UpdateUser(ctx context.Context, req *user.UpdateRequest) (*user.User, error) {
	return updateUser(ctx, db.innerDB, req)
}

func updateUser(ctx context.Context, ex executor, req *user.UpdateRequest) (*user.User, error) {
	fields := make([]string, 0, 4)
	args := make([]interface{}, 0)
	if req.Email != nil {
		fields = append(fields, "email = ?")
		args = append(args, *req.Email)
	}
	if req.Bio != nil {
		fields = append(fields, "bio = ?")
		args = append(args, *req.Bio)
	}
	if req.Password != nil {
		passwordHash, err := req.HashPassword()
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		fields = append(fields, "password_hash = ?")
		args = append(args, passwordHash)
	}
	if req.ImageURL != nil {
		fields = append(fields, "image_url = ?")
		args = append(args, *req.ImageURL)
	}

	var queryBuilder strings.Builder
	queryBuilder.WriteString("UPDATE users SET ")
	queryBuilder.WriteString(strings.Join(fields, ", "))
	queryBuilder.WriteString(" WHERE id = ? ")
	queryBuilder.WriteString("RETURNING id, email, username, bio, password_hash, image_url")

	if len(args) == 0 {
		return getUserByID(ctx, ex, req.UserID)
	}

	args = append(args, req.UserID)

	row := ex.QueryRowContext(ctx, queryBuilder.String(), args...)
	return newUserFromRow(row)
}

func newUserFromRow(row *sql.Row) (*user.User, error) {
	var usr user.User
	err := row.Scan(&usr.ID, &usr.Email, &usr.Username, &usr.Bio, &usr.PasswordHash, &usr.ImageURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}

		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			return nil, sqliteErrToDomain(sqliteErr)
		}
		return nil, fmt.Errorf("scan row  into User: %w", err)
	}

	return &usr, nil
}

func sqliteErrToDomain(err sqlite3.Error) error {
	if err.ExtendedCode == sqlite3.ErrConstraintUnique {
		msg := err.Error()
		if strings.Contains(msg, "users.") {
			if strings.Contains(msg, ".email") {
				return user.ErrEmailRegistered
			}
			if strings.Contains(msg, ".username") {
				return user.ErrUsernameTaken
			}
		}
	}

	// Default to the original error if unhandled.
	return err
}
