package validate

import "github.com/go-playground/validator/v10"

func registerValidations() {
	validate.RegisterValidation("pw_min", validatePasswordMinLen)
	validate.RegisterValidation("pw_max", validatePasswordMaxLen)
}

const (
	bcryptMaxLen   = 72
	passwordMinLen = 8
	passwordMaxLen = bcryptMaxLen
)

// Implements validator.Func.
func validatePasswordMinLen(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return len(password) >= passwordMinLen
}

// Implements validator.Func.
func validatePasswordMaxLen(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return len(password) <= passwordMaxLen
}
