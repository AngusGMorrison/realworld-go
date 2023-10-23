package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/angusgmorrison/realworld-go/pkg/etag"

	"github.com/lib/pq"

	"github.com/angusgmorrison/realworld-go/internal/domain/user"
	"github.com/angusgmorrison/realworld-go/internal/outbound/postgres/sqlc"
	"github.com/angusgmorrison/realworld-go/pkg/option"
	"github.com/google/uuid"
)

var _ user.Repository = (*Client)(nil)

// GetUserByID returns the [user.User] with the given ID, or
// [user.NotFoundError] if no such user exists.
func (c *Client) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return getUserById(ctx, c.queries, id)
}

func getUserById(ctx context.Context, q queries, id uuid.UUID) (*user.User, error) {
	row, err := q.GetUserById(ctx, id.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.NewNotFoundByIDError(id)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return parseUser(row.ID, row.Username, row.Email, row.PasswordHash, row.Bio, row.ImageUrl, row.UpdatedAt)
}

// GetUserByEmail returns the [user.User] with the given email, or
// [user.NotFoundError] if no user exists with that email.
func (c *Client) GetUserByEmail(ctx context.Context, email user.EmailAddress) (*user.User, error) {
	return getUserByEmail(ctx, c.queries, email)
}

func getUserByEmail(ctx context.Context, q queries, email user.EmailAddress) (*user.User, error) {
	row, err := q.GetUserByEmail(ctx, email.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.NewNotFoundByEmailError(email)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return parseUser(row.ID, row.Username, row.Email, row.PasswordHash, row.Bio, row.ImageUrl, row.UpdatedAt)
}

// CreateUser creates a new user record from the given
// [user.RegistrationRequest] and returns the created [user.User].
//
// Returns [user.ValidationError] if database constraints are violated.
func (c *Client) CreateUser(ctx context.Context, req *user.RegistrationRequest) (*user.User, error) {
	return createUser(ctx, c.queries, req)
}

func createUser(ctx context.Context, q queries, req *user.RegistrationRequest) (*user.User, error) {
	row, err := q.CreateUser(ctx, newCreateUserParamsFromRegistrationRequest(req))
	if err != nil {
		var postgresErr *pq.Error
		if errors.As(err, &postgresErr) {
			return nil, createUserErrToDomain(postgresErr, req)
		}
		return nil, fmt.Errorf("create user record from request %#v: %w", req, err)
	}

	return parseUser(row.ID, row.Username, row.Email, row.PasswordHash, row.Bio, row.ImageUrl, row.UpdatedAt)
}

func newCreateUserParamsFromRegistrationRequest(req *user.RegistrationRequest) sqlc.CreateUserParams {
	return sqlc.CreateUserParams{
		ID:           uuid.New().String(),
		Email:        req.Email().String(),
		Username:     req.Username().String(),
		PasswordHash: string(req.PasswordHash().Bytes()),
	}
}

func createUserErrToDomain(err *pq.Error, req *user.RegistrationRequest) error {
	if err.Code.Name() == "unique_violation" {
		if strings.Contains(err.Constraint, "email") {
			return user.NewDuplicateEmailError(req.Email())
		}
		if strings.Contains(err.Constraint, "username") {
			return user.NewDuplicateUsernameError(req.Username())
		}
	}

	// Default to the original error if unhandled.
	return err
}

// UpdateUser updates the user record specified by `req` and returns the updated
// [user.User].
//
// Returns [user.ValidationError] if database constraints are violated.
func (c *Client) UpdateUser(ctx context.Context, req *user.UpdateRequest) (*user.User, error) {
	tx, err := c.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin UpdateUser transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	queries := sqlc.New(tx)
	usr, err := updateUser(ctx, queries, req)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		// TODO: Handle serialization failures.
		return nil, fmt.Errorf("commit UpdateUser transaction: %w", err)
	}

	return usr, nil
}

func updateUser(ctx context.Context, q queries, req *user.UpdateRequest) (*user.User, error) {
	ok, err := q.UserExists(ctx, req.UserID().String())
	if err != nil {
		return nil, fmt.Errorf("query existence of user with ID %q: %w", req.UserID(), err)
	}
	if !ok {
		return nil, user.NewNotFoundByIDError(req.UserID())
	}

	params := parseUpdateUserParams(req)
	row, err := q.UpdateUser(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &user.ConcurrentModificationError{
				ID:   req.UserID(),
				ETag: req.ETag(),
			}
		}

		var postgresErr *pq.Error
		if errors.As(err, &postgresErr) && postgresErr.Code.Name() == "unique_violation" {
			if strings.Contains(postgresErr.Constraint, "email") {
				return nil, user.NewDuplicateEmailError(req.Email().UnwrapOrZero())
			}
		}

		return nil, fmt.Errorf("database error: %w", err)
	}

	return parseUser(row.ID, row.Username, row.Email, row.PasswordHash, row.Bio, row.ImageUrl, row.UpdatedAt)
}

func parseUpdateUserParams(req *user.UpdateRequest) sqlc.UpdateUserParams {
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
		String: string(req.PasswordHash().UnwrapOrZero().Bytes()),
		Valid:  req.PasswordHash().IsSome(),
	}

	return sqlc.UpdateUserParams{
		ID:           req.UserID().String(),
		UpdatedAt:    req.ETag().UpdatedAt(),
		Email:        email,
		Bio:          bio,
		ImageUrl:     imageURL,
		PasswordHash: passwordHash,
	}
}

func parseUser(
	id string,
	username string,
	email string,
	passwordHash string,
	bio sql.NullString,
	imageURL sql.NullString,
	updatedAt time.Time,
) (*user.User, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("retrieved unparsable user ID %q from database: %w", id, err)
	}

	eTag := etag.New(parsedID, updatedAt)

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
		[]byte(passwordHash),
	)

	parsedBio := option.None[user.Bio]()
	if bio.Valid {
		parsedBio = option.Some[user.Bio](user.Bio(bio.String))
	}

	return user.NewUser(parsedID, eTag, parsedUsername, parsedEmail, parsedPasswordHash, parsedBio, parsedImageURL), nil
}
