package option

import (
	"errors"
)

// Option represents an optional type.
type Option[T any] struct {
	some  bool
	value T
}

// Some returns an Option[T] populated with the given type.
func Some[T any](value T) Option[T] {
	return Option[T]{
		some:  true,
		value: value,
	}
}

// None returns an empty Option[T] whose zero-value is considered valid by
// [Option.NonZero] and [NonZero].
func None[S any]() Option[S] {
	return Option[S]{}
}

// Some returns true if the Option is populated with a non-zero value.
func (o Option[T]) Some() bool {
	return o.some
}

// ErrEmptyOption is returned by [Option.Value] when attempting to retrieve a
// value from an empty Option.
var ErrEmptyOption = errors.New("Option value is empty")

// Value returns the value of the [Option], or [ErrEmptyOption] error if the
// Option is empty.
func (o Option[T]) Value() (T, error) {
	if !o.Some() {
		return *new(T), ErrEmptyOption
	}
	return o.value, nil
}

// ValueOrZero returns the value of the [Option], which is the zero-value of T
// if the Option is None.
func (o Option[T]) ValueOrZero() T {
	return o.value
}
