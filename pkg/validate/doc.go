// Package validate provides thread-safe validation functions for structs tagged
// according to the validator package.
//
// In addition to the standard validations, its validator singleton is
// initialized with the following custom validations:
//   - pw_min: password minimum length;
//   - pw_max: password maximum length.
//
// Field names in returned errors are derived from json tags as a first restort,
// falling back to the underlying struct field name if no json tag is present.
//
// The registered translator provides error messages via (FieldError).Translate
// that are suitable for display to end users. The untranslated error messages
// remain available via (FieldError).Error.
package validate
