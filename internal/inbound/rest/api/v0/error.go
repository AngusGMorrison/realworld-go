package v0

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// UserFacingError is an error that will be rendered in responses to the client.
//
// `body` must contain only information that is safe to expose to
// users.
//
// `Cause` may contain supplementary information about the Cause of the
// error that is not safe to expose to users (e.g. for instrumentation).
type UserFacingError struct {
	StatusCode int
	Cause      error
	body       fiber.Map
}

func (e *UserFacingError) Error() string {
	return fmt.Sprintf("%d %s: %v", e.StatusCode, http.StatusText(e.StatusCode), e.body)
}

func (e *UserFacingError) Body() fiber.Map {
	return fiber.Map{
		"errors": e.body,
	}
}

func (e *UserFacingError) Unwrap() error {
	return e.Cause
}

func NewBadRequestError(cause error) error {
	return &UserFacingError{
		StatusCode: http.StatusBadRequest,
		Cause:      cause,
		body: fiber.Map{
			"bad_request": []string{"request body was invalid JSON or contained unknown fields"},
		},
	}
}

func NewNotFoundError(resourceName string, detail string) error {
	return &UserFacingError{
		StatusCode: http.StatusNotFound,
		body: fiber.Map{
			resourceName: []string{detail},
		},
	}
}

func NewUnauthorizedError(detail string, cause error) error {
	return &UserFacingError{
		StatusCode: http.StatusUnauthorized,
		Cause:      cause,
		body: fiber.Map{
			"unauthorized": []string{detail},
		},
	}
}

type userFacingValidationErrorMessages map[string][]string

func (e userFacingValidationErrorMessages) toFiberMap() fiber.Map {
	fiberMap := make(fiber.Map, len(e))
	for k, v := range e {
		fiberMap[k] = v
	}
	return fiberMap
}

func NewUnprocessableEntityError(messages userFacingValidationErrorMessages) error {
	return &UserFacingError{
		StatusCode: http.StatusUnprocessableEntity,
		body:       messages.toFiberMap(),
	}
}
