package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/angusgmorrison/realworld-go/internal/config"

	"github.com/angusgmorrison/realworld-go/internal/domain/user"
	"github.com/angusgmorrison/realworld-go/internal/outbound/postgres/sqlc"
	"github.com/angusgmorrison/realworld-go/pkg/option"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Client_GetUserByID(t *testing.T) {
	t.Parallel()

	t.Run("user exists", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.New()
		require.NoError(t, err)

		db, err := New(NewURL(cfg))
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		existingUser := user.RandomUser(t)
		queries := sqlc.New(db.db)

		_, err = queries.CreateUser(context.Background(), newCreateUserParamsFromUser(existingUser))
		require.NoError(t, err)

		gotUser, gotErr := db.GetUserByID(context.Background(), existingUser.ID())
		assert.NoError(t, gotErr)
		// Password hashes cannot be compared, so compare other fields piecewise.
		assert.Equal(t, existingUser.ID(), gotUser.ID())
		assert.Equal(t, existingUser.Username(), gotUser.Username())
		assert.Equal(t, existingUser.Email(), gotUser.Email())
		assert.Equal(t, existingUser.Bio(), gotUser.Bio())
		assert.Equal(t, existingUser.ImageURL(), gotUser.ImageURL())
		assert.NotEmpty(t, gotUser.PasswordHash())
	})

	t.Run("user does not exist", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.New()
		require.NoError(t, err)

		db, err := New(NewURL(cfg))
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		nonExistentUserID := uuid.New()

		gotUser, gotErr := db.GetUserByID(context.Background(), nonExistentUserID)
		assert.ErrorIs(t, gotErr, user.NewNotFoundByIDError(nonExistentUserID))
		assert.Nil(t, gotUser)
	})

	t.Run("database error", func(t *testing.T) {
		t.Parallel()

		queries := &mockQueries{}
		db := &Client{
			queries: queries,
		}
		userID := uuid.New()
		wantErr := errors.New("some error")

		queries.On(
			"GetUserById",
			context.Background(),
			userID.String(),
		).Return(sqlc.GetUserByIdRow{}, wantErr)

		gotUser, gotErr := db.GetUserByID(context.Background(), userID)
		assert.ErrorIs(t, gotErr, wantErr)
		assert.Nil(t, gotUser)
	})
}

func Test_Client_GetUserByEmail(t *testing.T) {
	t.Parallel()

	t.Run("user exists", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.New()
		require.NoError(t, err)

		db, err := New(NewURL(cfg))
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		existingUser := user.RandomUser(t)
		queries := sqlc.New(db.db)

		_, err = queries.CreateUser(context.Background(), newCreateUserParamsFromUser(existingUser))
		require.NoError(t, err)

		gotUser, gotErr := db.GetUserByEmail(context.Background(), existingUser.Email())
		assert.NoError(t, gotErr)
		// Password hashes cannot be compared, so compare the other fields piecewise.
		assert.Equal(t, existingUser.ID(), gotUser.ID())
		assert.Equal(t, existingUser.Username(), gotUser.Username())
		assert.Equal(t, existingUser.Email(), gotUser.Email())
		assert.Equal(t, existingUser.Bio(), gotUser.Bio())
		assert.Equal(t, existingUser.ImageURL(), gotUser.ImageURL())
		assert.NotEmpty(t, gotUser.PasswordHash())
	})

	t.Run("user does not exist", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.New()
		require.NoError(t, err)

		db, err := New(NewURL(cfg))
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		nonExistentUserEmail := user.RandomEmailAddress(t)

		gotUser, gotErr := db.GetUserByEmail(context.Background(), nonExistentUserEmail)
		assert.ErrorIs(t, gotErr, user.NewNotFoundByEmailError(nonExistentUserEmail))
		assert.Nil(t, gotUser)
	})

	t.Run("database error", func(t *testing.T) {
		t.Parallel()

		queries := &mockQueries{}
		db := &Client{
			queries: queries,
		}
		email := user.RandomEmailAddress(t)
		wantErr := errors.New("some error")

		queries.On(
			"GetUserByEmail",
			context.Background(),
			email.String(),
		).Return(sqlc.GetUserByEmailRow{}, wantErr)

		gotUser, gotErr := db.GetUserByEmail(context.Background(), email)
		assert.ErrorIs(t, gotErr, wantErr)
		assert.Nil(t, gotUser)
	})
}

func Test_Client_CreateUser(t *testing.T) {
	t.Parallel()

	t.Run("user does not exist", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.New()
		require.NoError(t, err)

		db, err := New(NewURL(cfg))
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		req := user.RandomRegistrationRequest(t)

		gotUser, gotErr := db.CreateUser(context.Background(), req)
		assert.NoError(t, gotErr)
		assert.NotEmpty(t, gotUser.ID())
		assert.Equal(t, req.Username(), gotUser.Username())
		assert.Equal(t, req.Email(), gotUser.Email())
		assert.Equal(t, option.None[user.Bio](), gotUser.Bio())
		assert.Equal(t, option.None[user.URL](), gotUser.ImageURL())
	})

	t.Run("user with same username already exists", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.New()
		require.NoError(t, err)

		db, err := New(NewURL(cfg))
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		existingUser := user.RandomUser(t)
		queries := sqlc.New(db.db)

		_, err = queries.CreateUser(context.Background(), newCreateUserParamsFromUser(existingUser))
		require.NoError(t, err)

		req, err := user.ParseRegistrationRequest(
			existingUser.Username().String(),
			user.RandomEmailAddressCandidate(),
			user.RandomPasswordCandidate(),
		)
		require.NoError(t, err)

		gotUser, gotErr := db.CreateUser(context.Background(), req)
		assert.ErrorIs(t, gotErr, user.NewDuplicateUsernameError(req.Username()))
		assert.Nil(t, gotUser)
	})

	t.Run("user with same email address already exists", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.New()
		require.NoError(t, err)

		db, err := New(NewURL(cfg))
		require.NoError(t, err)
		defer func() { _ = db.Close() }()

		existingUser := user.RandomUser(t)
		queries := sqlc.New(db.db)

		_, err = queries.CreateUser(context.Background(), newCreateUserParamsFromUser(existingUser))
		require.NoError(t, err)

		req, err := user.ParseRegistrationRequest(
			user.RandomUsernameCandidate(),
			existingUser.Email().String(),
			user.RandomPasswordCandidate(),
		)
		require.NoError(t, err)

		gotUser, gotErr := db.CreateUser(context.Background(), req)
		assert.ErrorIs(t, gotErr, user.NewDuplicateEmailError(req.Email()))
		assert.Nil(t, gotUser)
	})

	t.Run("database error", func(t *testing.T) {
		t.Parallel()

		queries := &mockQueries{}
		db := &Client{
			queries: queries,
		}
		wantErr := errors.New("some error")

		queries.On(
			"CreateUser",
			context.Background(),
			mock.AnythingOfType("sqlc.CreateUserParams"),
		).Return(sqlc.User{}, wantErr)

		gotUser, gotErr := db.CreateUser(context.Background(), user.RandomRegistrationRequest(t))
		assert.ErrorIs(t, gotErr, wantErr)
		assert.Nil(t, gotUser)
	})
}

func Test_Client_UpdateUser(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name         string
			email        option.Option[user.EmailAddress]
			passwordHash option.Option[user.PasswordHash]
			bio          option.Option[user.Bio]
			imageURL     option.Option[user.URL]
			assertUpdate func(t *testing.T, req *user.UpdateRequest, originalUser, updatedUser *user.User)
		}{
			{
				name:         "update all fields",
				email:        option.Some(user.RandomEmailAddress(t)),
				passwordHash: option.Some(user.RandomPasswordHash(t)),
				bio:          option.Some(user.RandomBio()),
				imageURL:     option.Some(user.RandomURL(t)),
				assertUpdate: func(t *testing.T, req *user.UpdateRequest, originalUser, updatedUser *user.User) {
					t.Helper()
					assert.Equal(t, req.Email().UnwrapOrZero(), updatedUser.Email())
					assert.Equal(t, req.PasswordHash().UnwrapOrZero(), updatedUser.PasswordHash())
					assert.Equal(t, req.Bio(), updatedUser.Bio())
					assert.Equal(t, req.ImageURL(), updatedUser.ImageURL())
				},
			},
			{
				name:         "update email address",
				email:        option.Some(user.RandomEmailAddress(t)),
				passwordHash: option.None[user.PasswordHash](),
				bio:          option.None[user.Bio](),
				imageURL:     option.None[user.URL](),
				assertUpdate: func(t *testing.T, req *user.UpdateRequest, originalUser, updatedUser *user.User) {
					t.Helper()
					assert.Equal(t, req.Email().UnwrapOrZero(), updatedUser.Email())
					assert.Equal(t, originalUser.PasswordHash(), updatedUser.PasswordHash())
					assert.Equal(t, originalUser.Bio(), updatedUser.Bio())
					assert.Equal(t, originalUser.ImageURL(), updatedUser.ImageURL())
				},
			},
			{
				name:         "update password hash",
				email:        option.None[user.EmailAddress](),
				passwordHash: option.Some(user.RandomPasswordHash(t)),
				bio:          option.None[user.Bio](),
				imageURL:     option.None[user.URL](),
				assertUpdate: func(t *testing.T, req *user.UpdateRequest, originalUser, updatedUser *user.User) {
					t.Helper()
					assert.Equal(t, originalUser.Email(), originalUser.Email())
					assert.Equal(t, req.PasswordHash().UnwrapOrZero(), updatedUser.PasswordHash())
					assert.Equal(t, originalUser.Bio(), updatedUser.Bio())
					assert.Equal(t, originalUser.ImageURL(), updatedUser.ImageURL())
				},
			},
			{
				name:         "update bio",
				email:        option.None[user.EmailAddress](),
				passwordHash: option.None[user.PasswordHash](),
				bio:          option.Some(user.RandomBio()),
				imageURL:     option.None[user.URL](),
				assertUpdate: func(t *testing.T, req *user.UpdateRequest, originalUser, updatedUser *user.User) {
					t.Helper()
					assert.Equal(t, originalUser.Email(), updatedUser.Email())
					assert.Equal(t, originalUser.PasswordHash(), updatedUser.PasswordHash())
					assert.Equal(t, req.Bio(), updatedUser.Bio())
					assert.Equal(t, originalUser.ImageURL(), updatedUser.ImageURL())
				},
			},
			{
				name:         "update image url",
				email:        option.None[user.EmailAddress](),
				passwordHash: option.None[user.PasswordHash](),
				bio:          option.None[user.Bio](),
				imageURL:     option.Some(user.RandomURL(t)),
				assertUpdate: func(t *testing.T, req *user.UpdateRequest, originalUser, updatedUser *user.User) {
					t.Helper()
					assert.Equal(t, originalUser.Email(), updatedUser.Email())
					assert.Equal(t, originalUser.PasswordHash(), updatedUser.PasswordHash())
					assert.Equal(t, originalUser.Bio(), updatedUser.Bio())
					assert.Equal(t, req.ImageURL(), updatedUser.ImageURL())
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				cfg, err := config.New()
				require.NoError(t, err)

				db, err := New(NewURL(cfg))
				require.NoError(t, err)

				row, err := db.queries.CreateUser(
					context.Background(),
					newCreateUserParamsFromUser(user.RandomUser(t)),
				)
				require.NoError(t, err)

				originalUser, err := parseUser(row.ID, row.Username, row.Email, row.PasswordHash, row.Bio, row.ImageUrl)
				require.NoError(t, err)

				req := user.NewUpdateRequest(originalUser.ID(), tc.email, tc.passwordHash, tc.bio, tc.imageURL)

				updatedUser, err := db.UpdateUser(context.Background(), req)
				assert.NoError(t, err)
				tc.assertUpdate(t, req, originalUser, updatedUser)
			})
		}
	})

	t.Run("user does not exist", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.New()
		require.NoError(t, err)

		db, err := New(NewURL(cfg))
		require.NoError(t, err)

		req := user.RandomUpdateRequest(t)

		gotUser, gotErr := db.UpdateUser(context.Background(), req)
		assert.ErrorIs(t, gotErr, user.NewNotFoundByIDError(req.UserID()))
		assert.Nil(t, gotUser)
	})

	t.Run("email address already exists", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.New()
		require.NoError(t, err)

		db, err := New(NewURL(cfg))
		require.NoError(t, err)

		existingUserWithTargetEmail := user.RandomUser(t)
		_, err = db.queries.CreateUser(
			context.Background(),
			newCreateUserParamsFromUser(existingUserWithTargetEmail),
		)
		require.NoError(t, err)

		userToUpdate := user.RandomUser(t)
		_, err = db.queries.CreateUser(
			context.Background(),
			newCreateUserParamsFromUser(userToUpdate),
		)
		require.NoError(t, err)

		req := user.NewUpdateRequest(
			userToUpdate.ID(),
			option.Some(existingUserWithTargetEmail.Email()),
			option.None[user.PasswordHash](),
			option.None[user.Bio](),
			option.None[user.URL](),
		)

		gotUser, gotErr := db.UpdateUser(context.Background(), req)
		assert.ErrorIs(t, gotErr, user.NewDuplicateEmailError(existingUserWithTargetEmail.Email()))
		assert.Nil(t, gotUser)
	})

	t.Run("database error", func(t *testing.T) {
		t.Parallel()

		queries := &mockQueries{}
		db := &Client{
			queries: queries,
		}
		req := user.RandomUpdateRequest(t)
		params := newUpdateUserParamsFromDomain(req)
		wantErr := errors.New("some error")

		queries.On("UpdateUser", context.Background(), params).Return(sqlc.User{}, wantErr)

		gotUser, gotErr := db.UpdateUser(context.Background(), req)
		assert.ErrorIs(t, gotErr, wantErr)
		assert.Nil(t, gotUser)
	})
}

func newCreateUserParamsFromUser(usr *user.User) sqlc.CreateUserParams {
	return sqlc.CreateUserParams{
		ID:           usr.ID().String(),
		Username:     usr.Username().String(),
		Email:        usr.Email().String(),
		PasswordHash: usr.PasswordHash().String(),
		Bio: sql.NullString{
			String: string(usr.Bio().UnwrapOrZero()),
			Valid:  usr.Bio().IsSome(),
		},
		ImageUrl: sql.NullString{
			String: usr.ImageURL().UnwrapOrZero().String(),
			Valid:  usr.ImageURL().IsSome(),
		},
	}
}
