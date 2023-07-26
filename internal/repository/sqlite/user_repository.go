package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/angusgmorrison/logfusc"
	"github.com/angusgmorrison/realworld/internal/repository/sqlite/sqlc"
	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/angusgmorrison/realworld/pkg/option"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"strings"
)

var _ user.Repository = (*SQLite)(nil)

// GetUserByID returns the [user.User] with the given ID, or
// [user.NotFoundError] if no such user exists.
func (db *SQLite) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return getUserById(ctx, id, db.innerDB)
}

func getUserById(ctx context.Context, id uuid.UUID, tx sqlc.DBTX) (*user.User, error) {
	queries := sqlc.New(tx)
	row, err := queries.GetUserById(ctx, id.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.NewNotFoundByIDError(id)
		}
	}

	return parseUser(row.ID, row.Email, row.Username, row.Bio, row.PasswordHash, row.ImageUrl)
}

// GetUserByEmail returns the [user.User] with the given email, or
// [user.NotFoundError] if no user exists with that email.
func (db *SQLite) GetUserByEmail(ctx context.Context, email user.EmailAddress) (*user.User, error) {
	return getUserByEmail(ctx, email, db.innerDB)
}

func getUserByEmail(ctx context.Context, email user.EmailAddress, tx sqlc.DBTX) (*user.User, error) {
	queries := sqlc.New(tx)
	row, err := queries.GetUserByEmail(ctx, email.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.NewNotFoundByEmailError(email)
		}
	}

	return parseUser(row.ID, row.Email, row.Username, row.Bio, row.PasswordHash, row.ImageUrl)
}

// CreateUser creates a new user record from the given
// [user.RegistrationRequest] and returns the created [user.User].
//
// Returns [user.ValidationError] if database constraints are violated.
func (db *SQLite) CreateUser(ctx context.Context, req *user.RegistrationRequest) (*user.User, error) {
	return createUser(ctx, req, db.innerDB)
}

func createUser(ctx context.Context, req *user.RegistrationRequest, tx sqlc.DBTX) (*user.User, error) {
	queries := sqlc.New(tx)
	row, err := queries.CreateUser(ctx, newCreateUserParamsFromDomain(req))
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			return nil, createUserErrToDomain(sqliteErr, req)
		}
		return nil, fmt.Errorf("create user record from request %#v: %w", req, err)
	}

	return parseUser(row.ID, row.Email, row.Username, row.Bio, row.PasswordHash, row.ImageUrl)
}

func newCreateUserParamsFromDomain(req *user.RegistrationRequest) sqlc.CreateUserParams {
	return sqlc.CreateUserParams{
		ID:           uuid.New().String(),
		Email:        req.EmailAddress().String(),
		Username:     req.Username().String(),
		PasswordHash: string(req.PasswordHash().Expose()),
	}
}

func createUserErrToDomain(err sqlite3.Error, req *user.RegistrationRequest) error {
	if err.ExtendedCode == sqlite3.ErrConstraintUnique {
		msg := err.Error()
		if strings.Contains(msg, "users.") {
			if strings.Contains(msg, ".email") {
				return user.NewDuplicateEmailError(req.EmailAddress())
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
	return updateUser(ctx, req, db.innerDB)
}

func updateUser(ctx context.Context, req *user.UpdateRequest, tx sqlc.DBTX) (*user.User, error) {
	queries := sqlc.New(tx)
	row, err := queries.UpdateUser(ctx, newUpdateUserParamsFromDomain(req))
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			return nil, updateUserErrToDomain(sqliteErr, req)
		}
		return nil, fmt.Errorf("update user record from request %#v: %w", req, err)
	}

	return parseUser(row.ID, row.Email, row.Username, row.Bio, row.PasswordHash, row.ImageUrl)
}

func newUpdateUserParamsFromDomain(req *user.UpdateRequest) sqlc.UpdateUserParams {
	email := sql.NullString{
		String: req.Email().ValueOrZero().String(),
		Valid:  req.Email().Some(),
	}
	bio := sql.NullString{
		String: string(req.Bio().ValueOrZero()),
		Valid:  req.Bio().Some(),
	}
	imageURL := sql.NullString{
		String: req.ImageURL().ValueOrZero().String(),
		Valid:  req.ImageURL().Some(),
	}
	passwordHash := sql.NullString{
		String: string(req.PasswordHash().ValueOrZero().Expose()),
		Valid:  req.PasswordHash().Some(),
	}

	return sqlc.UpdateUserParams{
		ID:           req.UserID().String(),
		Email:        email,
		Bio:          bio,
		ImageUrl:     imageURL,
		PasswordHash: passwordHash,
	}
}

func updateUserErrToDomain(err sqlite3.Error, req *user.UpdateRequest) error {
	if err.ExtendedCode == sqlite3.ErrConstraintUnique {
		msg := err.Error()
		if strings.Contains(msg, "users.") {
			if strings.Contains(msg, ".email") {
				return user.NewDuplicateEmailError(req.Email().ValueOrZero())
			}
		}
	}

	// Default to the original error if unhandled.
	return err
}

func parseUser(
	id string,
	email string,
	username string,
	bio sql.NullString,
	passwordHash string,
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

	parsedPasswordHash := user.WrapHashedPassword(
		logfusc.NewSecret([]byte(passwordHash)),
	)

	parsedBio := option.None[user.Bio]()
	if bio.Valid {
		parsedBio = option.Some[user.Bio](user.Bio(bio.String))
	}

	return user.NewUser(parsedID, parsedUsername, parsedEmail, parsedPasswordHash, parsedBio, parsedImageURL), nil
}
