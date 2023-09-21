package v0

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"

	"github.com/angusgmorrison/realworld-go/internal/domain/user"

	"github.com/gofiber/fiber/v2"
)

// JSONError maps an application error to its JSON representation.
type JSONError struct {
	RequestID string              `json:"id"`
	Status    int                 `json:"status"`
	Message   string              `json:"message"`
	Errors    map[string][]string `json:"errors,omitempty"`
}

func (e *JSONError) Error() string {
	bytes, err := json.Marshal(e)
	if err != nil {
		panic(fmt.Errorf("marshal JSONError: %w", err)) // never
	}

	return string(bytes)
}

func (e *JSONError) Is(other error) bool {
	var otherJSONErr *JSONError
	return errors.As(other, &otherJSONErr) &&
		e.RequestID == otherJSONErr.RequestID &&
		e.Status == otherJSONErr.Status &&
		e.Message == otherJSONErr.Message &&
		maps.EqualFunc(e.Errors, otherJSONErr.Errors, func(v1 []string, v2 []string) bool {
			return slices.Equal(v1, v2)
		})
}

// NewBadRequestError should be used in responses to syntactically invalid
// requests.
func NewBadRequestError(requestID string) error {
	return &JSONError{
		RequestID: requestID,
		Status:    fiber.StatusBadRequest,
		Message:   http.StatusText(fiber.StatusBadRequest),
	}
}

type missingResource struct {
	name   string
	idType string
	id     string
}

// NewNotFoundError should be used in responses to requests for resources that
// don't exist.
func NewNotFoundError(requestID string, resource missingResource) error {
	return &JSONError{
		RequestID: requestID,
		Status:    fiber.StatusNotFound,
		Message:   fmt.Sprintf("%s with %s %q not found.", resource.name, resource.idType, resource.id),
	}
}

func NewUnauthorizedError(requestID string) error {
	return &JSONError{
		RequestID: requestID,
		Status:    fiber.StatusUnauthorized,
		Message:   http.StatusText(fiber.StatusUnauthorized),
	}
}

type validationErrorMessages map[string][]string

func NewUnprocessableEntityError(requestID string, validationErrs user.ValidationErrors) error {
	errorMessages := make(validationErrorMessages)
	if err := formatValidationErrors(validationErrs, errorMessages); err != nil {
		return err
	}

	return &JSONError{
		RequestID: requestID,
		Status:    fiber.StatusUnprocessableEntity,
		Message:   "Request contained invalid fields.",
		Errors:    errorMessages,
	}
}

func formatValidationErrors(errs user.ValidationErrors, messages validationErrorMessages) error {
	for _, err := range errs {
		switch err.Field {
		case user.EmailFieldType:
			messages["email"] = append(messages["email"], err.Message)
		case user.PasswordFieldType:
			messages["password"] = append(messages["password"], err.Message)
		case user.UsernameFieldType:
			messages["username"] = append(messages["username"], err.Message)
		case user.URLFieldType:
			messages["image"] = append(messages["image"], err.Message)
		default:
			return fmt.Errorf(
				"unhandled validation error with field type %[1]q (enum value %[1]d): %w",
				err.Field,
				err,
			)
		}
	}

	return nil
}

func NewUnsupportedContentTypeError(requestID string, contentType string, supportedContentTypes []string) error {
	message := fmt.Sprintf(
		"Content type %q is not supported. Supported content types are: %s.",
		contentType,
		strings.Join(supportedContentTypes, ", "),
	)
	return &JSONError{
		RequestID: requestID,
		Status:    fiber.StatusUnsupportedMediaType,
		Message:   message,
	}
}

func NewInternalServerError(requestID string) error {
	return &JSONError{
		RequestID: requestID,
		Status:    fiber.StatusInternalServerError,
		Message:   http.StatusText(fiber.StatusInternalServerError),
	}
}