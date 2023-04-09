package validate

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func Test_Validations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		invalidExample  any
		wantTranslation string
	}{
		{
			name: "email",
			invalidExample: struct {
				Email string `validate:"email"`
			}{
				Email: "invalid@",
			},
			wantTranslation: `"invalid@" is not a valid email address`,
		},
		{
			name: "pw_max",
			invalidExample: struct {
				Password string `validate:"pw_max"`
			}{
				Password: strings.Repeat("a", passwordMaxLen+1),
			},
			wantTranslation: fmt.Sprintf("must be at most %d bytes", passwordMaxLen),
		},
		{
			name: "pw_min",
			invalidExample: struct {
				Password string `validate:"pw_min"`
			}{
				Password: "1234567",
			},
			wantTranslation: fmt.Sprintf("must be at least %d bytes", passwordMinLen),
		},
		{
			name: "required",
			invalidExample: struct {
				Required string `validate:"required"`
			}{
				Required: "",
			},
			wantTranslation: "is required",
		},
		{
			name: "url",
			invalidExample: struct {
				URL string `validate:"url"`
			}{
				URL: "invalid URL",
			},
			wantTranslation: `"invalid URL" is not a valid URL`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotErrs := Struct(tc.invalidExample).(validator.ValidationErrors)

			assert.Equal(
				t,
				tc.wantTranslation,
				Translate(gotErrs[0]),
			)
		})
	}
}

func Test_TagNames(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		invalidExample any
		wantFieldName  string
	}{
		{
			name: "when no overriding tag is present it falls back to the name of the field",
			invalidExample: struct {
				Required string `validate:"required"`
			}{
				Required: "",
			},
			wantFieldName: "Required",
		},
		{
			name: "when overriding tag is present it uses that instead of the name of the field",
			invalidExample: struct {
				Required string `json:"name_override" validate:"required"`
			}{
				Required: "",
			},
			wantFieldName: "name_override",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotErrs := Struct(tc.invalidExample).(validator.ValidationErrors)

			assert.Equal(
				t,
				tc.wantFieldName,
				gotErrs[0].Field(),
			)
		})
	}
}
