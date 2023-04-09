package validate

import (
	"reflect"
	"strings"
)

func registerTagNameFuncs() {
	validate.RegisterTagNameFunc(jsonTagName)
}

func jsonTagName(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}

	return name
}
