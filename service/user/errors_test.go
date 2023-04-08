package user

import (
	"testing"
)

func Test_ValidationError_Error(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		fieldErrors []error
		want        string
	}{
		{
			name:        "No field errors",
			fieldErrors: nil,
			want:        "{}",
		},
		{
			name:        "One field error",
			fieldErrors: []error{ErrEmailAddressUnparseable},
			want:        "{\n\temail address is not RFC 5322-compliant,\n}",
		},
		{
			name:        "Multiple field errors",
			fieldErrors: []error{ErrEmailAddressUnparseable, ErrImageURLUnparseable},
			want:        "{\n\temail address is not RFC 5322-compliant,\n\timage URL could not be parsed,\n}",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validationError := ValidationError{
				fieldErrors: tc.fieldErrors,
			}

			if got := validationError.Error(); got != tc.want {
				t.Errorf("Expected output to be %s, but got %s", tc.want, got)
			}
		})
	}
}
