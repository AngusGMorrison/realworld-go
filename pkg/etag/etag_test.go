package etag

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_ParseETag(t *testing.T) {
	t.Parallel()

	t.Run("when the raw ETag is valid", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		timestamp := time.Now()
		raw := fmt.Sprintf(`"%s%s%s"`, id, eTagSeparator, timestamp.Format(time.RFC3339Nano))

		got, err := Parse(raw)
		assert.NoError(t, err)
		assert.Equal(t, id, got.id)
		assert.Truef(
			t,
			timestamp.Equal(got.updatedAt),
			"expected equal timestamps\n\twant:\t%s\n\tgot:\t%s\n",
			timestamp.Format(time.RFC3339Nano),
			got.updatedAt.Format(time.RFC3339Nano),
		)
	})

	t.Run("errors", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name string
			raw  string
		}{
			{
				name: "not a double-quoted string",
				raw:  fmt.Sprintf("%s%s%s", uuid.New(), eTagSeparator, time.Now().Format(time.RFC3339Nano)),
			},
			{
				name: "has < 2 components",
				raw:  fmt.Sprintf(`"%s"`, uuid.New()),
			},
			{
				name: "has > 2 components",
				raw: fmt.Sprintf(
					`"%s%s%s%s"`,
					uuid.New(),
					eTagSeparator,
					time.Now().Format(time.RFC3339Nano),
					eTagSeparator,
				),
			},
			{
				name: "has invalid UUID",
				raw: fmt.Sprintf(
					`"%s%s%s"`,
					"not a UUID",
					eTagSeparator,
					time.Now().Format(time.RFC3339Nano),
				),
			},
			{
				name: "has invalid timestamp",
				raw: fmt.Sprintf(
					`"%s%s%s"`,
					uuid.New(),
					eTagSeparator,
					"not a timestamp",
				),
			},
		}

		for _, tc := range testCases {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				tag, err := Parse(tc.raw)

				var parseErr *ParseETagError
				assert.ErrorAs(t, err, &parseErr)
				assert.Empty(t, tag)
			})
		}
	})
}

func Test_ETag_String(t *testing.T) {
	t.Parallel()

	eTag := ETag{
		id:        uuid.New(),
		updatedAt: time.Now(),
	}
	expected := fmt.Sprintf(`"%s%s%s"`, eTag.id, eTagSeparator, eTag.updatedAt.Format(time.RFC3339Nano))

	got := eTag.String()
	assert.Equal(t, expected, got)
}
