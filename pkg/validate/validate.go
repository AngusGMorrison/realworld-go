package validate

import (
	"github.com/go-playground/validator/v10"
)

var (
	// validate is a singleton exposing thread-safe struct validation methods.
	validate *validator.Validate
)

func init() {
	validate = validator.New()
}

// Struct validates the given struct.
func Struct(s any) error {
	return validate.Struct(s)
}
