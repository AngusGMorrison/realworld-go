package validate

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

var (
	// validate is a singleton exposing thread-safe struct validation methods.
	validate   *validator.Validate
	translator ut.Translator
)

func init() {
	validate = validator.New()
	translator = ut.New(en.New()).GetFallback()

	registerTagNameFuncs()
	registerTranslations()
	registerValidations()
}

// Struct validates the given struct.
func Struct(s any) error {
	return validate.Struct(s)
}

func Translate(err validator.FieldError) string {
	return err.Translate(translator)
}
