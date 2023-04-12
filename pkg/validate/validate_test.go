package validate

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

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
