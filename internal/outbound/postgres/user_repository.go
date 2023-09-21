package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/angusgmorrison/realworld-go/internal/domain/user"
	"github.com/angusgmorrison/realworld-go/internal/outbound/postgres/sqlc"
	"github.com/angusgmorrison/realworld-go/pkg/option"
	"github.com/google/uuid"
)

var _ user.Repository = (*Client)(nil)

// GetUserByID returns the [user.User] with the given RequestID, or
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
		return nil, fmt.Errorf("Client error: %w", err)
	}

	return parseUser(row.ID, row.Username, row.Email, row.PasswordHash, row.Bio, row.ImageUrl)
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
		return nil, fmt.Errorf("Client error: %w", err)
	}

	return parseUser(row.ID, row.Username, row.Email, row.PasswordHash, row.Bio, row.ImageUrl)
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

	return parseUser(row.ID, row.Username, row.Email, row.PasswordHash, row.Bio, row.ImageUrl)
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
	return updateUser(ctx, c.queries, req)
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
		String: string(req.PasswordHash().UnwrapOrZero().Bytes()),
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

	var postgresErr *pq.Error
	if errors.As(err, &postgresErr) && postgresErr.Code.Name() == "unique_violation" {
		if strings.Contains(postgresErr.Constraint, "email") {
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
		[]byte(passwordHash),
	)

	parsedBio := option.None[user.Bio]()
	if bio.Valid {
		parsedBio = option.Some[user.Bio](user.Bio(bio.String))
	}

	return user.NewUser(parsedID, parsedUsername, parsedEmail, parsedPasswordHash, parsedBio, parsedImageURL), nil
}
