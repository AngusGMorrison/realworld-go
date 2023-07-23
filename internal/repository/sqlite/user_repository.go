package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"github.com/angusgmorrison/logfusc"
	"github.com/angusgmorrison/realworld/internal/repository/sqlite/sqlc"
	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/angusgmorrison/realworld/pkg/option"
	"github.com/google/uuid"
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
			return nil, &user.NotFoundError{ID: id}
		}
	}

	return parseUserFromGetUserRow(row)
}

func parseUserFromGetUserRow(row sqlc.GetUserByIdRow) (*user.User, error) {
	id := uuid.MustParse(row.ID)

	email, err := user.ParseEmailAddress(row.Email)
	if err != nil {
		return nil, err
	}

	username, err := user.ParseUsername(row.Username)
	if err != nil {
		return nil, err
	}

	imageURL := option.None[user.URL]()
	if row.ImageUrl.Valid {
		parsedURL, err := user.ParseURL(row.ImageUrl.String)
		if err != nil {
			return nil, err
		}
		imageURL = option.Some[user.URL](parsedURL)
	}

	passwordHash := user.WrapHashedPassword(
		logfusc.NewSecret([]byte(row.PasswordHash)),
	)

	bio := option.None[user.Bio]()
	if row.Bio.Valid {
		bio = option.Some[user.Bio](user.Bio(row.Bio.String))
	}

	return user.NewUser(id, username, email, passwordHash, bio, imageURL), nil
}
