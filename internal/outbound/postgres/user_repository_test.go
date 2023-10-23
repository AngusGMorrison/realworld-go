package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/angusgmorrison/realworld-go/pkg/etag"

	"github.com/angusgmorrison/realworld-go/pkg/option"

	"github.com/stretchr/testify/mock"

	"github.com/angusgmorrison/realworld-go/internal/config"

	"github.com/angusgmorrison/realworld-go/internal/domain/user"
	"github.com/angusgmorrison/realworld-go/internal/outbound/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

		existingUser := saveRandomUser(t, db)

		gotUser, gotErr := db.GetUserByID(context.Background(), existingUser.ID())
		assert.NoError(t, gotErr)
		// Password hashes cannot be compared, so compare other fields piecewise.
		assert.Equal(t, existingUser.ID(), gotUser.ID())
		assert.Equal(t, existingUser.ETag(), gotUser.ETag())
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

		existingUser := saveRandomUser(t, db)

		gotUser, gotErr := db.GetUserByEmail(context.Background(), existingUser.Email())
		assert.NoError(t, gotErr)
		// Password hashes cannot be compared, so compare the other fields piecewise.
		assert.Equal(t, existingUser.ID(), gotUser.ID())
		assert.Equal(t, existingUser.ETag(), gotUser.ETag())
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
		assertETagPresent(t, gotUser)
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
					assertETagChanged(t, originalUser.ETag(), updatedUser.ETag())
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
					assertETagChanged(t, originalUser.ETag(), updatedUser.ETag())
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
					assertETagChanged(t, originalUser.ETag(), updatedUser.ETag())
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
					assertETagChanged(t, originalUser.ETag(), updatedUser.ETag())
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
					assertETagChanged(t, originalUser.ETag(), updatedUser.ETag())
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

				originalUser := saveRandomUser(t, db)
				req := user.NewUpdateRequest(
					originalUser.ID(),
					originalUser.ETag(),
					tc.email,
					tc.passwordHash,
					tc.bio,
					tc.imageURL,
				)

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

		randomUser := user.RandomUser(t)
		registrationReq := user.NewRegistrationRequest(
			randomUser.Username(),
			randomUser.Email(),
			randomUser.PasswordHash(),
		)
		existingUserWithTargetEmail, err := db.CreateUser(context.Background(), registrationReq)
		require.NoError(t, err)

		randomUser = user.RandomUser(t)
		registrationReq = user.NewRegistrationRequest(
			randomUser.Username(),
			randomUser.Email(),
			randomUser.PasswordHash(),
		)
		userToUpdate, err := db.CreateUser(context.Background(), registrationReq)
		require.NoError(t, err)

		updateReq := user.NewUpdateRequest(
			userToUpdate.ID(),
			userToUpdate.ETag(),
			option.Some(existingUserWithTargetEmail.Email()),
			option.None[user.PasswordHash](),
			option.None[user.Bio](),
			option.None[user.URL](),
		)

		gotUser, gotErr := db.UpdateUser(context.Background(), updateReq)
		assert.ErrorIs(t, gotErr, user.NewDuplicateEmailError(existingUserWithTargetEmail.Email()))
		assert.Nil(t, gotUser)
	})

	t.Run("database error", func(t *testing.T) {
		t.Parallel()

		queries := &mockQueries{}
		req := user.RandomUpdateRequest(t)
		params := parseUpdateUserParams(req)
		wantErr := errors.New("some error")

		queries.On(
			"UserExists",
			context.Background(),
			req.UserID().String(),
		).Return(true, nil)
		queries.On("UpdateUser", context.Background(), params).Return(sqlc.User{}, wantErr)

		gotUser, gotErr := updateUser(context.Background(), queries, req)
		assert.ErrorIs(t, gotErr, wantErr)
		assert.Nil(t, gotUser)
	})

	t.Run("concurrent modification error", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.New()
		require.NoError(t, err)

		db, err := New(NewURL(cfg))
		require.NoError(t, err)

		randomUser := user.RandomUser(t)
		registrationReq := user.NewRegistrationRequest(
			randomUser.Username(),
			randomUser.Email(),
			randomUser.PasswordHash(),
		)
		userToUpdate, err := db.CreateUser(context.Background(), registrationReq)
		require.NoError(t, err)

		staleETag := etag.New(
			userToUpdate.ID(),
			userToUpdate.ETag().UpdatedAt().Add(-1*time.Second),
		)

		updateReq := user.NewUpdateRequest(
			userToUpdate.ID(),
			staleETag,
			option.None[user.EmailAddress](),
			option.None[user.PasswordHash](),
			option.None[user.Bio](),
			option.None[user.URL](),
		)

		gotUser, gotErr := db.UpdateUser(context.Background(), updateReq)

		var concurrentModificationErr *user.ConcurrentModificationError
		assert.ErrorAs(t, gotErr, &concurrentModificationErr)
		assert.Nil(t, gotUser)
	})
}

func newCreateUserParamsFromUser(usr *user.User) sqlc.CreateUserParams {
	return sqlc.CreateUserParams{
		ID:           usr.ID().String(),
		Username:     usr.Username().String(),
		Email:        usr.Email().String(),
		PasswordHash: string(usr.PasswordHash().Bytes()),
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

func assertETagPresent(t *testing.T, usr *user.User) {
	t.Helper()

	eTag := usr.ETag()
	assert.Equal(t, usr.ID(), eTag.ID())
	assert.NotZerof(t, eTag.UpdatedAt(), "expected ETag to have non-zero timestamp")
}

func assertETagChanged(t *testing.T, original, updated etag.ETag) {
	t.Helper()

	assert.Equal(t, original.ID(), updated.ID())
	assert.Truef(
		t,
		updated.UpdatedAt().After(original.UpdatedAt()),
		"expected updated ETag %q to have timestamp after original ETag %q",
		updated,
		original,
	)
}

func saveRandomUser(t *testing.T, db *Client) *user.User {
	t.Helper()

	req := user.RandomRegistrationRequest(t)
	createdUser, err := db.CreateUser(context.Background(), req)
	require.NoError(t, err)

	return createdUser
}
