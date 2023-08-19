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
// `cause` may contain supplementary information about the cause of the
// error that is not safe to expose to users (e.g. for instrumentation).
type UserFacingError struct {
	StatusCode int
	cause      error
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

func (e *UserFacingError) Cause() error {
	return e.cause
}

func NewBadRequestError() error {
	return &UserFacingError{
		StatusCode: http.StatusBadRequest,
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

func NewUnauthorizedError(detail string) error {
	return &UserFacingError{
		StatusCode: http.StatusUnauthorized,
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
