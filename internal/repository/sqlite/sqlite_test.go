package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// It is essential that each package uses a distinct test database to avoid
// unexpected drops by parallel tests.
const dbPath = "testdata/sqlite_test.db"

func TestMain(m *testing.M) {
	code, err := setUpAndTearDown(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}

	os.Exit(code)
}

func setUpAndTearDown(m *testing.M) (int, error) {
	db, err := New(dbPath)
	if err != nil {
		return 1, err
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		return 1, err
	}

	code := m.Run()

	if err := os.Remove(dbPath); err != nil {
		return 1, err
	}

	return code, nil
}

func newTx(t *testing.T) (tx *sql.Tx, rollback func()) {
	db, err := New(dbPath)
	require.NoError(t, err, "open DB connection")

	tx, err = db.innerDB.Begin()
	if err != nil {
		db.Close()
		require.NoError(t, err, "begin transaction")
	}

	return tx, func() {
		tx.Rollback()
		db.Close()
	}
}

const (
	email        = "test@test.com"
	username     = "testuser"
	password     = "password"
	passwordHash = "abc123"
	bio          = "test bio"
	imageURL     = "https://test.com/image.png"
)

func Test_getUserByID(t *testing.T) {
	t.Parallel()

	t.Run("when user exists it returns the user", func(t *testing.T) {
		t.Parallel()

		expectedUser := &user.User{
			Email:        email,
			Username:     username,
			PasswordHash: passwordHash,
			Bio:          bio,
			ImageURL:     imageURL,
		}

		tx, rollback := newTx(t)
		defer rollback()

		expectedUser, err := insertUser(context.Background(), tx, expectedUser)
		require.NoError(t, err, "insert user")

		gotUser, err := getUserByID(context.Background(), tx, expectedUser.ID)
		require.NoError(t, err, "get user by ID")

		assert.Equal(t, expectedUser, gotUser)
	})

	t.Run("when user does not exist it returns ErrUserNotFound", func(t *testing.T) {
		t.Parallel()

		tx, rollback := newTx(t)
		defer rollback()

		gotUser, err := getUserByID(context.Background(), tx, uuid.New())
		require.ErrorIs(t, err, user.ErrUserNotFound)

		assert.Nil(t, gotUser)
	})
}

func Test_getUserByEmail(t *testing.T) {
	t.Parallel()

	t.Run("when user exists it returns the user", func(t *testing.T) {
		t.Parallel()

		expectedUser := &user.User{
			Email:        email,
			Username:     username,
			PasswordHash: passwordHash,
			Bio:          bio,
			ImageURL:     imageURL,
		}

		tx, rollback := newTx(t)
		defer rollback()

		expectedUser, err := insertUser(context.Background(), tx, expectedUser)
		require.NoError(t, err, "insert user")

		gotUser, err := getUserByEmail(context.Background(), tx, expectedUser.Email)
		require.NoError(t, err, "get user by email")

		assert.Equal(t, expectedUser, gotUser)
	})

	t.Run("when user does not exist it returns ErrUserNotFound", func(t *testing.T) {
		t.Parallel()

		tx, rollback := newTx(t)
		defer rollback()

		gotUser, err := getUserByEmail(context.Background(), tx, "missing@test.com")
		require.ErrorIs(t, err, user.ErrUserNotFound)

		assert.Nil(t, gotUser)
	})
}

func Test_insertUser(t *testing.T) {
	t.Parallel()

	t.Run("when constraints are met it inserts the user", func(t *testing.T) {
		usr := &user.User{
			Email:        email,
			Username:     username,
			PasswordHash: passwordHash,
			Bio:          bio,
			ImageURL:     imageURL,
		}

		tx, rollback := newTx(t)
		defer rollback()

		usr, err := insertUser(context.Background(), tx, usr)
		require.NoError(t, err, "insert user")

		usr, err = getUserByID(context.Background(), tx, usr.ID)
		require.NoError(t, err, "get inserted user")
	})

	// TODO: Return different errors for email and username.
	t.Run("when the email is not unique it returns ErrUserExists", func(t *testing.T) {
		originalUser := &user.User{
			Email:        email,
			Username:     username,
			PasswordHash: passwordHash,
			Bio:          bio,
			ImageURL:     imageURL,
		}
		dup := *originalUser
		dup.Username = "unique username"

		tx, rollback := newTx(t)
		defer rollback()

		_, err := insertUser(context.Background(), tx, originalUser)
		require.NoError(t, err, "insert user")

		gotUser, err := insertUser(context.Background(), tx, &dup)

		assert.ErrorIs(t, err, user.ErrEmailRegistered)
		assert.Nil(t, gotUser)
	})

	t.Run("when the username is not unique it returns ErrUserExists", func(t *testing.T) {
		originalUser := &user.User{
			Email:        email,
			Username:     username,
			PasswordHash: passwordHash,
			Bio:          bio,
			ImageURL:     imageURL,
		}
		dup := *originalUser
		dup.Email = "unique@test.com"

		tx, rollback := newTx(t)
		defer rollback()

		_, err := insertUser(context.Background(), tx, originalUser)
		require.NoError(t, err, "insert user")

		gotUser, err := insertUser(context.Background(), tx, &dup)

		assert.ErrorIs(t, err, user.ErrUsernameTaken)
		assert.Nil(t, gotUser)
	})
}

func Test_updateUser(t *testing.T) {
	t.Parallel()

	var (
		newEmail    = user.EmailAddress("newemail@test.com")
		newBio      = "A new bio."
		newPassword = "newpassword"
		newImageURL = "https://test.com/new.jpg"
	)

	t.Run("when a non-password field is updated it updates the user", func(t *testing.T) {
		testCases := []struct {
			name         string
			updateReq    *user.UpdateRequest
			expectedUser *user.User
		}{
			{
				name: "it updates the email field",
				updateReq: &user.UpdateRequest{
					Email: &newEmail,
				},
				expectedUser: &user.User{
					Email:        newEmail,
					Username:     username,
					PasswordHash: passwordHash,
					Bio:          bio,
					ImageURL:     imageURL,
				},
			},
			{
				name: "it updates the bio field",
				updateReq: &user.UpdateRequest{
					Bio: &newBio,
				},
				expectedUser: &user.User{
					Email:        email,
					Username:     username,
					PasswordHash: passwordHash,
					Bio:          newBio,
					ImageURL:     imageURL,
				},
			},
			{
				name: "it updates the image_url field",
				updateReq: &user.UpdateRequest{
					ImageURL: &newImageURL,
				},
				expectedUser: &user.User{
					Email:        email,
					Username:     username,
					PasswordHash: passwordHash,
					Bio:          bio,
					ImageURL:     newImageURL,
				},
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				originalUser := &user.User{
					Email:        email,
					Username:     username,
					PasswordHash: passwordHash,
					Bio:          bio,
					ImageURL:     imageURL,
				}

				db, err := New(dbPath)
				require.NoError(t, err, "open DB connection")
				defer db.Close()

				tx, err := db.innerDB.Begin()
				require.NoError(t, err, "begin transaction")
				defer tx.Rollback()

				insertedUser, err := insertUser(context.Background(), tx, originalUser)
				require.NoError(t, err, "insert test user")

				tc.updateReq.UserID = insertedUser.ID
				tc.expectedUser.ID = insertedUser.ID

				updatedUser, err := updateUser(context.Background(), tx, tc.updateReq)
				require.NoError(t, err, "update user")

				assert.Equal(t, tc.expectedUser, updatedUser)
			})
		}
	})

	t.Run("when the password_hash field is updated it updates the user", func(t *testing.T) {
		t.Parallel()

		originalUser := &user.User{
			Email:        email,
			Username:     username,
			PasswordHash: passwordHash,
			Bio:          bio,
			ImageURL:     imageURL,
		}

		tx, rollback := newTx(t)
		defer rollback()

		insertedUser, err := insertUser(context.Background(), tx, originalUser)
		require.NoError(t, err, "insert test user")

		updateReq := &user.UpdateRequest{
			UserID: insertedUser.ID,
			OptionalValidatingPassword: user.OptionalValidatingPassword{
				Password: &newPassword,
			},
		}
		updatedUser, err := updateUser(context.Background(), tx, updateReq)
		require.NoError(t, err, "update user")

		assert.True(t, updatedUser.HasPassword(newPassword), "password mismatch")
	})

	t.Run("when the email unique constraint is violated it returns ErrUserExists", func(t *testing.T) {
		t.Parallel()

		targetUser := &user.User{
			Email:        email,
			Username:     username,
			PasswordHash: passwordHash,
			Bio:          bio,
			ImageURL:     imageURL,
		}
		targetUserCopy := *targetUser
		targetUserCopy.Email = newEmail
		targetUserCopy.Username = "unique username"
		userWithDesiredEmail := &targetUserCopy

		tx, rollback := newTx(t)
		defer rollback()

		targetUser, err := insertUser(context.Background(), tx, targetUser)
		require.NoError(t, err, "insert targetUser")

		_, err = insertUser(context.Background(), tx, userWithDesiredEmail)
		require.NoError(t, err, "insert userWithDesiredEmail")

		updateReq := &user.UpdateRequest{
			UserID: targetUser.ID,
			Email:  &newEmail,
		}

		updatedUser, err := updateUser(context.Background(), tx, updateReq)

		assert.ErrorIs(t, err, user.ErrEmailRegistered)
		assert.Nil(t, updatedUser)
	})

}
