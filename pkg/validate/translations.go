package validate

import (
	"strconv"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func registerTranslations() {
	validate.RegisterTranslation("email", translator, registerEmailTranslation, translateEmailError)
	validate.RegisterTranslation("pw_min", translator, registerPWMinTranslation, translatePWMinError)
	validate.RegisterTranslation("pw_max", translator, registerPWMaxTranslation, translatePWMaxError)
	validate.RegisterTranslation("required", translator, registerRequiredTranslation, translateRequiredError)
	validate.RegisterTranslation("url", translator, registerURLTranslation, translateURLError)
}

func registerEmailTranslation(translator ut.Translator) error {
	return translator.Add("email", "\"{0}\" is not a valid email address", false)
}

func translateEmailError(translator ut.Translator, fieldError validator.FieldError) string {
	res, err := translator.T(fieldError.Tag(), fieldError.Value().(string))
	if err != nil {
		return fieldError.Error()
	}

	return res
}

func registerRequiredTranslation(translator ut.Translator) error {
	return translator.Add("required", "is required", false)
}

func translateRequiredError(translator ut.Translator, fieldError validator.FieldError) string {
	res, err := translator.T(fieldError.Tag())
	if err != nil {
		return fieldError.Error()
	}

	return res
}

func registerURLTranslation(translator ut.Translator) error {
	return translator.Add("url", "\"{0}\" is not a valid URL", false)
}

func translateURLError(translator ut.Translator, fieldError validator.FieldError) string {
	res, err := translator.T(fieldError.Tag(), fieldError.Value().(string))
	if err != nil {
		return fieldError.Error()
	}

	return res
}

func registerPWMinTranslation(translator ut.Translator) error {
	return translator.Add("pw_min", "must be at least {0} bytes", false)
}

func translatePWMinError(translator ut.Translator, fieldError validator.FieldError) string {
	res, err := translator.T(fieldError.Tag(), strconv.Itoa(passwordMinLen))
	if err != nil {
		return fieldError.Error()
	}

	return res
}

func registerPWMaxTranslation(translator ut.Translator) error {
	return translator.Add("pw_max", "must be at most {0} bytes", false)
}

func translatePWMaxError(translator ut.Translator, fieldError validator.FieldError) string {
	res, err := translator.T(fieldError.Tag(), strconv.Itoa(passwordMaxLen))
	if err != nil {
		return fieldError.Error()
	}

	return res
}
