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
	registerTagNameFuncs()
}

// Struct validates the given struct.
func Struct(s any) error {
	return validate.Struct(s)
}

// Fields validates the given struct fields.
func Fields(s any, fields ...string) error {
	return validate.StructPartial(s, fields...)
}
