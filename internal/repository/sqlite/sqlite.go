package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/angusgmorrison/realworld/internal/service/user"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
	db *sql.DB
}

func New(db *sql.DB) *SQLite {
	return &SQLite{db}
}

func (r *SQLite) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	query := `SELECT id, email, username, bio, password_hash, image_url FROM users WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var usr user.User
	err := row.Scan(&usr.ID, &usr.Email, &usr.Username, &usr.Bio, &usr.PasswordHash, &usr.ImageURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return &usr, nil
}

func (r *SQLite) GetUserByEmail(ctx context.Context, email user.EmailAddress) (*user.User, error) {
	query := `SELECT id, email, username, bio, password_hash, image_url FROM users WHERE email = ?`
	row := r.db.QueryRowContext(ctx, query, email)
	return newUserFromRow(row)
}

func (r *SQLite) CreateUser(ctx context.Context, req *user.RegistrationRequest) (*user.User, error) {
	passwordHash, err := req.HashPassword()
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO users (id, email, username, password_hash) VALUES (?, ?, ?, ?, ?, ?)`
	id := uuid.New()
	_, err = r.db.ExecContext(ctx, query, uuid.New(), req.Email, req.Username, passwordHash)
	if err != nil {
		return nil, err
	}

	return &user.User{
		ID:           id,
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: passwordHash,
	}, nil
}

func (r *SQLite) UpdateUser(ctx context.Context, req *user.UpdateRequest) (*user.User, error) {
	var queryBuilder strings.Builder
	args := make([]interface{}, 0)

	queryBuilder.WriteString("UPDATE users SET ")
	if req.Email != nil {
		queryBuilder.WriteString("email = ?, ")
		args = append(args, *req.Email)
	}
	if req.Bio != nil {
		queryBuilder.WriteString("bio = ?, ")
		args = append(args, *req.Bio)
	}
	if req.Password != nil {
		passwordHash, err := req.HashPassword()
		if err != nil {
			return nil, err
		}
		queryBuilder.WriteString("password_hash = ?, ")
		args = append(args, passwordHash)
	}
	if req.ImageURL != nil {
		queryBuilder.WriteString("image_url = ? ")
		args = append(args, *req.ImageURL)
	}
	queryBuilder.WriteString("WHERE id = ? ")
	queryBuilder.WriteString("RETURNING (id, email, username, bio, password_hash, image_url)")

	if len(args) == 0 {
		return r.GetUserByID(ctx, req.UserID)
	}

	args = append(args, req.UserID)

	row := r.db.QueryRowContext(ctx, queryBuilder.String(), args...)
	return newUserFromRow(row)
}

func newUserFromRow(row *sql.Row) (*user.User, error) {
	var usr user.User
	err := row.Scan(&usr.ID, &usr.Email, &usr.Username, &usr.Bio, &usr.PasswordHash, &usr.ImageURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return &usr, nil
}
