package presenter

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/angusgmorrison/realworld/internal/service/user"
	"github.com/angusgmorrison/realworld/pkg/validate"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	email        = "test@test.com"
	username     = "testuser"
	password     = "password"
	passwordHash = "abc123"
	bio          = "test bio"
	imageURL     = "https://test.com/image.png"
	token        = "test-token"
)

func Test_Fiber_ShowBadRequest(t *testing.T) {
	t.Parallel()

	presenter := NewFiberPresenter()
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return presenter.ShowBadRequest(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
	wantResBody, err := json.Marshal(fiber.Map{
		"error": "request body is not a valid JSON string",
	})
	require.NoError(t, err, "marshal wantResBody")

	res, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, res.StatusCode)

	gotResBody, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err, "read response body")
	assert.JSONEq(t, string(wantResBody), string(gotResBody))
}

func Test_Fiber_ShowRegister(t *testing.T) {
	t.Parallel()

	user := &user.User{
		ID:       uuid.New(),
		Email:    email,
		Username: username,
		Bio:      bio,
		ImageURL: imageURL,
	}

	presenter := NewFiberPresenter()
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return presenter.ShowRegister(c, user, token)
	})

	req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
	wantResBody, err := json.Marshal(fiber.Map{
		"user": fiber.Map{
			"email":    email,
			"username": username,
			"token":    token,
			"bio":      bio,
			"image":    imageURL,
		},
	})
	require.NoError(t, err, "marshal wantResBody")

	res, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, res.StatusCode)

	gotResBody, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err, "read response body")
	assert.JSONEq(t, string(wantResBody), string(gotResBody))
}

func Test_Fiber_ShowLogin(t *testing.T) {
	t.Parallel()

	user := &user.User{
		ID:       uuid.New(),
		Email:    email,
		Username: username,
		Bio:      bio,
		ImageURL: imageURL,
	}

	presenter := NewFiberPresenter()
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return presenter.ShowLogin(c, user, token)
	})

	req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
	wantResBody, err := json.Marshal(fiber.Map{
		"user": fiber.Map{
			"email":    email,
			"username": username,
			"token":    token,
			"bio":      bio,
			"image":    imageURL,
		},
	})
	require.NoError(t, err, "marshal wantResBody")

	res, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, res.StatusCode)

	gotResBody, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err, "read response body")
	assert.JSONEq(t, string(wantResBody), string(gotResBody))
}

func Test_Fiber_ShowGetCurrentUser(t *testing.T) {
	t.Parallel()

	user := &user.User{
		ID:       uuid.New(),
		Email:    email,
		Username: username,
		Bio:      bio,
		ImageURL: imageURL,
	}

	presenter := NewFiberPresenter()
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return presenter.ShowLogin(c, user, token)
	})

	req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
	wantResBody, err := json.Marshal(fiber.Map{
		"user": fiber.Map{
			"email":    email,
			"username": username,
			"token":    token,
			"bio":      bio,
			"image":    imageURL,
		},
	})
	require.NoError(t, err, "marshal wantResBody")

	res, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, res.StatusCode)

	gotResBody, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err, "read response body")
	assert.JSONEq(t, string(wantResBody), string(gotResBody))
}

func Test_Fiber_ShowUpdateCurrentUser(t *testing.T) {
	t.Parallel()

	user := &user.User{
		ID:       uuid.New(),
		Email:    email,
		Username: username,
		Bio:      bio,
		ImageURL: imageURL,
	}

	presenter := NewFiberPresenter()
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return presenter.ShowLogin(c, user, token)
	})

	req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
	wantResBody, err := json.Marshal(fiber.Map{
		"user": fiber.Map{
			"email":    email,
			"username": username,
			"token":    token,
			"bio":      bio,
			"image":    imageURL,
		},
	})
	require.NoError(t, err, "marshal wantResBody")

	res, err := app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, res.StatusCode)

	gotResBody, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err, "read response body")
	assert.JSONEq(t, string(wantResBody), string(gotResBody))
}

func Test_Fiber_ShowUserError(t *testing.T) {
	t.Parallel()

	invalidStruct := struct {
		Email         string `validate:"email"`
		Required      string `validate:"required"`
		ShortPassword string `validate:"min=2"`
		LongPassword  string `validate:"max=2"`
	}{
		Email:         "invalid",
		Required:      "",
		ShortPassword: "1",
		LongPassword:  "123",
	}
	validationErrs := validate.Struct(invalidStruct).(validator.ValidationErrors)

	testCases := []struct {
		name        string
		err         error
		wantStatus  int
		wantResBody fiber.Map
	}{
		{
			name:        "user.AuthError",
			err:         &user.AuthError{},
			wantStatus:  fiber.StatusUnauthorized,
			wantResBody: nil,
		},
		{
			name:       "user.ErrUserNotFound",
			err:        user.ErrUserNotFound,
			wantStatus: fiber.StatusNotFound,
			wantResBody: fiber.Map{
				"errors": fiber.Map{
					"email": []string{"user not found"},
				},
			},
		},
		{
			name:       "user.ErrEmailRegistered",
			err:        user.ErrEmailRegistered,
			wantStatus: fiber.StatusUnprocessableEntity,
			wantResBody: fiber.Map{
				"errors": fiber.Map{
					"email": []string{"is already registered"},
				},
			},
		},
		{
			name:       "user.ErrUsernameTaken",
			err:        user.ErrUsernameTaken,
			wantStatus: fiber.StatusUnprocessableEntity,
			wantResBody: fiber.Map{
				"errors": fiber.Map{
					"username": []string{"is taken"},
				},
			},
		},
		{
			name:       "validator.ValidationErrors",
			err:        validationErrs,
			wantStatus: fiber.StatusUnprocessableEntity,
			wantResBody: fiber.Map{
				"errors": fiber.Map{
					"Email":         []string{"is invalid"},
					"Required":      []string{"is required"},
					"ShortPassword": []string{"must be at least 2 characters"},
					"LongPassword":  []string{"must be at most 2 bytes"},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			presenter := NewFiberPresenter()
			app := fiber.New()
			app.Get("/", func(c *fiber.Ctx) error {
				return presenter.ShowUserError(c, tc.err)
			})

			req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)

			res, err := app.Test(req)

			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, res.StatusCode)

			if tc.wantResBody != nil {
				gotResBody, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err, "read response body")
				wantResBody, err := json.Marshal(tc.wantResBody)
				require.NoError(t, err, "marshal wantResBody")
				assert.JSONEq(t, string(wantResBody), string(gotResBody))
			}
		})
	}
}
