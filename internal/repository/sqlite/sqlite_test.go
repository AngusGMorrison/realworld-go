package sqlite

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	code, err := setUpAndTearDown(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}

	os.Exit(code)
}

func setUpAndTearDown(m *testing.M) (int, error) {
	db, err := New("testdata/test.db")
	if err != nil {
		return 1, err
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		return 1, err
	}

	code := m.Run()

	tables := []string{"users"}
	var errs []error
	for _, table := range tables {
		_, err := db.innerDB.Exec(fmt.Sprintf("DELETE FROM %s;", table))
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return 1, fmt.Errorf("truncate tables: %v", errs)
	}

	return code, nil
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

		db, err := New("testdata/test.db")
		require.NoError(t, err, "open DB connection")
		defer db.Close()

		tx, err := db.innerDB.Begin()
		require.NoError(t, err, "begin transaction")
		defer tx.Rollback()

		expectedUser, err = insertUser(context.Background(), tx, expectedUser)
		require.NoError(t, err, "insert user")

		gotUser, err := getUserByID(context.Background(), tx, expectedUser.ID)
		require.NoError(t, err, "get user by ID")

		assert.Equal(t, expectedUser, gotUser)
	})

	t.Run("when user does not exist it returns ErrUserNotFound", func(t *testing.T) {
		t.Parallel()

		db, err := New("testdata/test.db")
		require.NoError(t, err, "open DB connection")
		defer db.Close()

		tx, err := db.innerDB.Begin()
		require.NoError(t, err, "begin transaction")
		defer tx.Rollback()

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

		db, err := New("testdata/test.db")
		require.NoError(t, err, "open DB connection")
		defer db.Close()

		tx, err := db.innerDB.Begin()
		require.NoError(t, err, "begin transaction")
		defer tx.Rollback()

		expectedUser, err = insertUser(context.Background(), tx, expectedUser)
		require.NoError(t, err, "insert user")

		gotUser, err := getUserByEmail(context.Background(), tx, expectedUser.Email)
		require.NoError(t, err, "get user by email")

		assert.Equal(t, expectedUser, gotUser)
	})

	t.Run("when user does not exist it returns ErrUserNotFound", func(t *testing.T) {
		t.Parallel()

		db, err := New("testdata/test.db")
		require.NoError(t, err, "open DB connection")
		defer db.Close()

		tx, err := db.innerDB.Begin()
		require.NoError(t, err, "begin transaction")
		defer tx.Rollback()

		gotUser, err := getUserByEmail(context.Background(), tx, "missing@test.com")
		require.ErrorIs(t, err, user.ErrUserNotFound)

		assert.Nil(t, gotUser)
	})
}

func Test_insertUser(t *testing.T) {
	t.Parallel()

	usr := &user.User{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		Bio:          bio,
		ImageURL:     imageURL,
	}

	db, err := New("testdata/test.db")
	require.NoError(t, err, "open DB connection")
	defer db.Close()

	tx, err := db.innerDB.Begin()
	require.NoError(t, err, "begin transaction")
	defer tx.Rollback()

	usr, err = insertUser(context.Background(), tx, usr)
	require.NoError(t, err, "insert user")

	usr, err = getUserByID(context.Background(), tx, usr.ID)
	require.NoError(t, err, "get inserted user")
}

func Test_updateUser(t *testing.T) {
	t.Parallel()

	originalUser := &user.User{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		Bio:          bio,
		ImageURL:     imageURL,
	}

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

				db, err := New("testdata/test.db")
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

		db, err := New("testdata/test.db")
		require.NoError(t, err, "open DB connection")
		defer db.Close()

		tx, err := db.innerDB.Begin()
		require.NoError(t, err, "begin transaction")
		defer tx.Rollback()

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

}
