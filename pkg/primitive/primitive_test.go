package primitive

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSensitiveString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		format   string
		input    SensitiveString
		expected string
	}{
		{
			format:   "%s",
			input:    "secret",
			expected: "REDACTED",
		},
		{
			format:   "%q",
			input:    "secret",
			expected: "REDACTED",
		},
		{
			format:   "%v",
			input:    "secret",
			expected: "REDACTED",
		},
		{
			format:   "%+v",
			input:    "secret",
			expected: "REDACTED",
		},
		{
			format:   "%#v",
			input:    "secret",
			expected: "REDACTED",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(fmt.Sprintf("format %s", tc.format), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.input.String())
		})
	}
}

func TestEmailAddress(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    EmailAddress
		expected bool
	}{
		{
			name:     "valid email address",
			input:    EmailAddress("test@example.com"),
			expected: true,
		},
		{
			name:     "invalid email address",
			input:    EmailAddress("invalid"),
			expected: false,
		},
		{
			name:     "single-value domain",
			input:    EmailAddress("test@example"),
			expected: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, tc.input.IsRFC5322Valid())
		})
	}
}
