package user

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_AuthError_Error(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		err  *AuthError
		want string
	}{
		{
			name: "no cause",
			err:  &AuthError{},
			want: "unauthorized",
		},
		{
			name: "with cause",
			err:  &AuthError{Cause: errors.New("cause")},
			want: "unauthorized",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.err.Error())
		})
	}
}

func Test_AuthError_Unwrap(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		err  *AuthError
		want error
	}{
		{
			name: "no cause",
			err:  &AuthError{},
			want: nil,
		},
		{
			name: "with cause",
			err:  &AuthError{Cause: errors.New("cause")},
			want: errors.New("cause"),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.err.Unwrap())
		})
	}
}

func Test_NotFoundError_Error(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		id        string
		fieldType FieldType
		want      string
	}{
		{
			name:      "id field",
			id:        "id",
			fieldType: IDFieldType,
			want:      "user with id \"id\" not found",
		},
		{
			name:      "email field",
			id:        "test@test.com",
			fieldType: EmailFieldType,
			want:      "user with email \"test@test.com\" not found",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := &NotFoundError{
				IDFieldValue: tc.id,
				IDFieldType:  tc.fieldType,
			}

			assert.Equal(t, tc.want, err.Error())
		})
	}
}

func Test_NewNotFoundByIDError(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	want := &NotFoundError{
		IDFieldValue: id.String(),
		IDFieldType:  IDFieldType,
	}

	got := NewNotFoundByIDError(id)

	assert.Equal(t, want, got)
}

func Test_NewNotFoundByEmailError(t *testing.T) {
	t.Parallel()

	email, err := ParseEmailAddress("test@test.com")
	require.NoError(t, err)

	want := &NotFoundError{
		IDFieldValue: email.String(),
		IDFieldType:  EmailFieldType,
	}

	got := NewNotFoundByEmailError(email)

	assert.Equal(t, want, got)
}

func Test_ValidationErrors_PushValidationError(t *testing.T) {
	t.Parallel()

	validationErr := &ValidationError{
		FieldType: IDFieldType,
		Message:   "reason",
	}
	nonValidationErr := errors.New("error")

	testCases := []struct {
		name               string
		err                error
		wantValidationErrs ValidationErrors
		wantErr            error
	}{
		{
			name: "push ValidationError",
			err:  validationErr,
			wantValidationErrs: ValidationErrors{
				validationErr,
			},
			wantErr: nil,
		},
		{
			name:               "push nil",
			err:                nil,
			wantValidationErrs: nil,
			wantErr:            nil,
		},
		{
			name:               "push non-ValidationError",
			err:                nonValidationErr,
			wantValidationErrs: nil,
			wantErr:            nonValidationErr,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var validationErrs ValidationErrors
			err := validationErrs.PushValidationError(tc.err)

			assert.Equal(t, tc.wantValidationErrs, validationErrs)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func Test_ValidationErrors_Any(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		validationErrs ValidationErrors
		want           bool
	}{
		{
			name:           "empty",
			validationErrs: ValidationErrors{},
			want:           false,
		},
		{
			name:           "nil",
			validationErrs: nil,
			want:           false,
		},
		{
			name: "non-empty",
			validationErrs: ValidationErrors{
				&ValidationError{},
			},
			want: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.validationErrs.Any())
		})
	}
}

func Test_ValidationErrors_Error(t *testing.T) {
	t.Parallel()

	validationErr := &ValidationError{
		FieldType: IDFieldType,
		Message:   "reason",
	}

	testCases := []struct {
		name           string
		validationErrs ValidationErrors
		want           string
	}{
		{
			name:           "empty",
			validationErrs: ValidationErrors{},
			want:           "validation errors:\n",
		},
		{
			name:           "nil",
			validationErrs: nil,
			want:           "validation errors:\n",
		},
		{
			name: "non-empty",
			validationErrs: ValidationErrors{
				validationErr,
			},
			want: fmt.Sprintf("validation errors:\n\t- %s\n", validationErr),
		},
		{
			name: "multiple",
			validationErrs: ValidationErrors{
				validationErr,
				validationErr,
			},
			want: fmt.Sprintf("validation errors:\n\t- %[1]s\n\t- %[1]s\n", validationErr),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.validationErrs.Error())
		})
	}
}

func Test_ValidationError_Error(t *testing.T) {
	t.Parallel()

	err := &ValidationError{
		FieldType: IDFieldType,
		Message:   "reason",
	}
	want := "id: reason"

	assert.Equal(t, want, err.Error())
}
