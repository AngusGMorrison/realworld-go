package middleware

import (
	"context"
	"net/http"
	"testing"

	"github.com/angusgmorrison/realworld-go/internal/testutil"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ContentTypeValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		reqContentType string
		wantErr        error
	}{
		{
			name:           "supported content type",
			reqContentType: fiber.MIMEApplicationJSON,
			wantErr:        nil,
		},
		{
			name:           "supported content type with charset",
			reqContentType: "application/json; charset=utf-8",
			wantErr:        nil,
		},
		{
			name:           "invalid content type",
			reqContentType: fiber.MIMETextPlainCharsetUTF8,
			wantErr:        assert.AnError,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			assertionHandler := func(c *fiber.Ctx) error {
				err := c.Next()
				assert.ErrorIs(t, err, tc.wantErr)
				return nil
			}
			supportedContentTypes := []string{fiber.MIMEApplicationJSON}
			unsupportedContentTypeHandler := func(unsupportedType string) error {
				return assert.AnError
			}
			app.Post("/", assertionHandler, ContentTypeValidation(supportedContentTypes, unsupportedContentTypeHandler))

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			require.NoError(t, err)
			req.Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			_, err = app.Test(req, testutil.FiberTestTimeoutMillis)
			require.NoError(t, err)
		})
	}
}
