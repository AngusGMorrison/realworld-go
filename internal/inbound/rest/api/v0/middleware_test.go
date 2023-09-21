package v0

import (
	"context"
	"net/http"
	"testing"

	"github.com/angusgmorrison/realworld-go/internal/inbound/rest/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ContentTypeValidation(t *testing.T) {
	t.Parallel()

	requestID := uuid.New().String()
	deterministicRequestIDSetter := func(c *fiber.Ctx) error {
		c.Locals(middleware.RequestIDKey, requestID)
		return c.Next()
	}

	testCases := []struct {
		name            string
		contentType     string
		requestIDSetter func(c *fiber.Ctx) error
		wantErr         error
	}{
		{
			name:            "supported content type",
			contentType:     fiber.MIMEApplicationJSON,
			requestIDSetter: deterministicRequestIDSetter,
			wantErr:         nil,
		},
		{
			name:            "supported content type with suffix",
			contentType:     fiber.MIMEApplicationJSONCharsetUTF8,
			requestIDSetter: deterministicRequestIDSetter,
			wantErr:         nil,
		},
		{
			name:            "unsupported content type",
			contentType:     fiber.MIMETextPlain,
			requestIDSetter: deterministicRequestIDSetter,
			wantErr:         NewUnsupportedContentTypeError(requestID, fiber.MIMETextPlain, supportedContentTypes),
		},
		{
			name:        "unsupported content type with missing request ID on context",
			contentType: fiber.MIMETextPlain,
			requestIDSetter: func(c *fiber.Ctx) error {
				return c.Next()
			},
			wantErr: middleware.ErrMissingRequestID,
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
			app.Use(tc.requestIDSetter, assertionHandler, ContentTypeValidation)
			app.Get("/", func(c *fiber.Ctx) error {
				return nil
			})

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			require.NoError(t, err)

			req.Header.Set(fiber.HeaderContentType, tc.contentType)

			_, err = app.Test(req)
			require.NoError(t, err)
		})
	}
}
