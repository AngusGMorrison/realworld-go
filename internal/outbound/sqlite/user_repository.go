package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/angusgmorrison/logfusc"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"

	"github.com/angusgmorrison/realworld-go/internal/domain/user"
	"github.com/angusgmorrison/realworld-go/internal/outbound/sqlite/sqlc"
	"github.com/angusgmorrison/realworld-go/pkg/option"
)

var _ user.Repository = (*SQLite)(nil)

// GetUserByID returns the [user.User] with the given ID, or
// [user.NotFoundError] if no such user exists.
func (db *SQLite) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return getUserById(ctx, db.queries, id)
}

func getUserById(ctx context.Context, q queries, id uuid.UUID) (*user.User, error) {
	row, err := q.GetUserById(ctx, id.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.NewNotFoundByIDError(id)
		}
		return nil, fmt.Errorf("SQLite error: %w", err)
	}

	return parseUser(row.ID, row.Username, row.Email, row.PasswordHash, row.Bio, row.ImageUrl)
}

// GetUserByEmail returns the [user.User] with the given email, or
// [user.NotFoundError] if no user exists with that email.
func (db *SQLite) GetUserByEmail(ctx context.Context, email user.EmailAddress) (*user.User, error) {
	return getUserByEmail(ctx, db.queries, email)
}

func getUserByEmail(ctx context.Context, q queries, email user.EmailAddress) (*user.User, error) {
	row, err := q.GetUserByEmail(ctx, email.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.NewNotFoundByEmailError(email)
		}
		return nil, fmt.Errorf("SQLite error: %w", err)
	}

	return parseUser(row.ID, row.Username, row.Email, row.PasswordHash, row.Bio, row.ImageUrl)
}

// CreateUser creates a new user record from the given
// [user.RegistrationRequest] and returns the created [user.User].
//
// Returns [user.ValidationError] if database constraints are violated.
func (db *SQLite) CreateUser(ctx context.Context, req *user.RegistrationRequest) (*user.User, error) {
	return createUser(ctx, db.queries, req)
}

func createUser(ctx context.Context, q queries, req *user.RegistrationRequest) (*user.User, error) {
	row, err := q.CreateUser(ctx, newCreateUserParamsFromRegistrationRequest(req))
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			return nil, createUserErrToDomain(sqliteErr, req)
		}
		return nil, fmt.Errorf("create user record from request %#v: %w", req, err)
	}

	return parseUser(row.ID, row.Username, row.Email, row.PasswordHash, row.Bio, row.ImageUrl)
}

func newCreateUserParamsFromRegistrationRequest(req *user.RegistrationRequest) sqlc.CreateUserParams {
	return sqlc.CreateUserParams{
		ID:           uuid.New().String(),
		Email:        req.Email().String(),
		Username:     req.Username().String(),
		PasswordHash: string(req.PasswordHash().Expose()),
	}
}

func createUserErrToDomain(err sqlite3.Error, req *user.RegistrationRequest) error {
	if err.ExtendedCode == sqlite3.ErrConstraintUnique {
		msg := err.Error()
		if strings.Contains(msg, "users.") {
			if strings.Contains(msg, ".email") {
				return user.NewDuplicateEmailError(req.Email())
			}
			if strings.Contains(msg, ".username") {
				return user.NewDuplicateUsernameError(req.Username())
			}
		}
	}

	// Default to the original error if unhandled.
	return err
}

// UpdateUser updates the user record specified by `req` and returns the updated
// [user.User].
//
// Returns [user.ValidationError] if database constraints are violated.
func (db *SQLite) UpdateUser(ctx context.Context, req *user.UpdateRequest) (*user.User, error) {
	return updateUser(ctx, db.queries, req)
}

func updateUser(ctx context.Context, q queries, req *user.UpdateRequest) (*user.User, error) {
	row, err := q.UpdateUser(ctx, newUpdateUserParamsFromDomain(req))
	if err != nil {
		return nil, updateUserErrorToDomain(err, req)
	}

	return parseUser(row.ID, row.Username, row.Email, row.PasswordHash, row.Bio, row.ImageUrl)
}

func newUpdateUserParamsFromDomain(req *user.UpdateRequest) sqlc.UpdateUserParams {
	email := sql.NullString{
		String: req.Email().UnwrapOrZero().String(),
		Valid:  req.Email().IsSome(),
	}
	bio := sql.NullString{
		String: string(req.Bio().UnwrapOrZero()),
		Valid:  req.Bio().IsSome(),
	}
	imageURL := sql.NullString{
		String: req.ImageURL().UnwrapOrZero().String(),
		Valid:  req.ImageURL().IsSome(),
	}
	passwordHash := sql.NullString{
		String: string(req.PasswordHash().UnwrapOrZero().Expose()),
		Valid:  req.PasswordHash().IsSome(),
	}

	return sqlc.UpdateUserParams{
		ID:           req.UserID().String(),
		Email:        email,
		Bio:          bio,
		ImageUrl:     imageURL,
		PasswordHash: passwordHash,
	}
}

func updateUserErrorToDomain(err error, req *user.UpdateRequest) error {
	if errors.Is(err, sql.ErrNoRows) {
		return user.NewNotFoundByIDError(req.UserID())
	}

	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
		msg := err.Error()
		if strings.Contains(msg, "users.email") {
			return user.NewDuplicateEmailError(req.Email().UnwrapOrZero())
		}
	}

	return fmt.Errorf("database error: %w", err)
}

func parseUser(
	id string,
	username string,
	email string,
	passwordHash string,
	bio sql.NullString,
	imageURL sql.NullString,
) (*user.User, error) {
	parsedID := uuid.MustParse(id)

	parsedEmail, err := user.ParseEmailAddress(email)
	if err != nil {
		return nil, err
	}

	parsedUsername, err := user.ParseUsername(username)
	if err != nil {
		return nil, err
	}

	parsedImageURL := option.None[user.URL]()
	if imageURL.Valid {
		parsedURL, err := user.ParseURL(imageURL.String)
		if err != nil {
			return nil, err
		}
		parsedImageURL = option.Some[user.URL](parsedURL)
	}

	parsedPasswordHash := user.NewPasswordHashFromTrustedSource(
		logfusc.NewSecret([]byte(passwordHash)),
	)

	parsedBio := option.None[user.Bio]()
	if bio.Valid {
		parsedBio = option.Some[user.Bio](user.Bio(bio.String))
	}

	return user.NewUser(parsedID, parsedUsername, parsedEmail, parsedPasswordHash, parsedBio, parsedImageURL), nil
}
