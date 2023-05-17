package valueobj

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEmailAddress(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		wantEmail   EmailAddress
		assertError assert.ErrorAssertionFunc
	}{
		{
			name:        "valid email address",
			input:       "test@example.com",
			wantEmail:   EmailAddress{raw: "test@example.com"},
			assertError: assert.NoError,
		},
		{
			name:        "valid email address with single-value domain",
			input:       "test@example",
			wantEmail:   EmailAddress{raw: "test@example"},
			assertError: assert.NoError,
		},
		{
			name:      "invalid email address",
			input:     "invalid",
			wantEmail: EmailAddress{},
			assertError: func(t assert.TestingT, err error, _ ...interface{}) bool {
				var parseErr *ParseEmailAddressError
				return assert.ErrorAs(t, err, &parseErr)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			email, err := ParseEmailAddress(tc.input)

			assert.Equal(t, tc.wantEmail, email)
			tc.assertError(t, err)
		})
	}
}
