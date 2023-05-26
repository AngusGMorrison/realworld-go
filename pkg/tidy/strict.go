package tidy

import (
	"errors"
	"fmt"
)

// Strict is utilized in TDD to validate that consumers of domain APIs cannot
// pass zero-value objects to the domain, bypassing validations in the
// type's constructor.
//
// A Strict type is considered "zero" if any of its required fields are missing.
// [Option] fields MUST NOT affect the zero-value status of Strict objects.
type Strict interface {
	// NonZero MUST return [ZeroValueError] if ANY of the receiver's
	// required fields have their zero-values, or nil otherwise.
	NonZero() error
}

// ZeroValueError is returned whenever an empty strict object is passed to a
// domain function that requires a non-zero value.
type ZeroValueError struct {
	ZeroStrict Strict
}

func (e *ZeroValueError) Error() string {
	return fmt.Sprintf("strict type %T must not have zero-value", e.ZeroStrict)
}

var errNilStrict = errors.New("Strict interface value must not be nil")

// NonZero validates that all [Strict] objects passed are non-zero. All
// [Strict] objects are evaluated, and the results are aggregated into a single
// error.
func NonZero(stricts ...Strict) error {
	var errs []error
	for _, strict := range stricts {
		if strict == nil {
			errs = append(errs, errNilStrict)
			continue
		}

		if err := strict.NonZero(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
