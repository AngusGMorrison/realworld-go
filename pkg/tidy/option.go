package tidy

import (
	"errors"
	"fmt"
)

// Option represents an optional [Strict] type. If the option is empty, the
// zero-value of the wrapped type will be treated as as non-zero for the
// purposes of validation with [Option.NonZero] and [NonZero]. If the option is
// populated, non-zero constraints on the wrapped value will be enforced as
// normal.
//
// It is recommended for domains to parameterize options by non-pointer types,
// since retained references to pointer types may be manipulated from outside
// the domain, violating the domain's invariants.
type Option[S Strict] struct {
	some   bool
	strict S
}

// Some returns an Option[T] populated with the given strict type, or an error
// if the strict type is zero-valued.
func Some[S Strict](strict S) (Option[S], error) {
	if err := strict.NonZero(); err != nil {
		return Option[S]{}, fmt.Errorf("cannot populate Option with a zero-value: %w", err)
	}

	return Option[S]{
		some:   true,
		strict: strict,
	}, nil
}

// None returns an empty Option[T] whose zero-value is considered valid by
// [Option.NonZero] and [NonZero].
func None[S Strict]() Option[S] {
	return Option[S]{}
}

// Some returns true if the Option is populated with a non-zero value.
func (o Option[_]) Some() bool {
	return o.some
}

// NonZero returns nil if the Option is empty, or the result of calling
// T.NonZero otherwise.
func (o Option[_]) NonZero() error {
	if !o.Some() {
		return nil
	}

	//nolint:wrapcheck
	return o.strict.NonZero()
}

// ErrEmptyOption is returned by [Option.Value] when attempting to retrieve a
// value from an empty Option.
var ErrEmptyOption = errors.New("Option value is empty")

// Value returns the value of the Option, or [ErrEmptyOption] error if the if
// Option is empty.
func (o Option[S]) Value() (S, error) {
	if !o.Some() {
		return *new(S), ErrEmptyOption
	}
	return o.strict, nil
}
