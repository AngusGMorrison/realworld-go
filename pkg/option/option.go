package option

import (
	"encoding/json"
	"errors"
	"fmt"
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

// IsSome returns true if the Option is populated with a non-zero value.
func (o Option[T]) IsSome() bool {
	return o.some
}

// ErrEmptyOption is returned by [Option.Unwrap] when attempting to retrieve a
// value from an empty Option.
var ErrEmptyOption = errors.New("expected Option value was empty")

// Unwrap returns the value of the [Option], or [ErrEmptyOption] error if the
// Option is empty.
func (o Option[T]) Unwrap() (T, error) {
	if !o.IsSome() {
		return *new(T), ErrEmptyOption
	}
	return o.value, nil
}

// UnwrapOrZero returns the value of the [Option], which is the zero-value of T
// if the Option is None.
func (o Option[T]) UnwrapOrZero() T {
	return o.value
}

// Conversion is a function that converts a T into a U.
type Conversion[T any, U any] func(T) (U, error)

// Map transforms an Option[T] into an Option[U] by applying the given conversion to T.
// None[T] is returned as None[U].
//
// # Errors
//   - Any error returned by the conversion.
func Map[T any, U any](opt Option[T], convert Conversion[T, U]) (Option[U], error) {
	if !opt.IsSome() {
		return None[U](), nil
	}

	u, err := convert(opt.UnwrapOrZero())
	if err != nil {
		return None[U](), fmt.Errorf("convert %#v to %T: %w", opt.value, *new(U), err)
	}

	return Some(u), nil
}

func (o Option[T]) GoString() string {
	return fmt.Sprintf("option.Option[%[1]T]{some:%[2]t, value:%#[1]v}", o.value, o.some)
}

func (o Option[T]) String() string {
	if o.some {
		return fmt.Sprintf("Some[%[1]T]{%[1]v}", o.value)
	}
	return fmt.Sprintf("None[%T]", o.value)
}

func (o *Option[T]) UnmarshalJSON(bytes []byte) error {
	// By the fact that UnmarshalJSON has been called, we know that the Option is
	// Some. None corresponds to omitted fields, which do not trigger UnmarshalJSON.
	o.some = true

	if len(bytes) == 0 {
		o.value = *new(T)
		return nil
	}

	if err := json.Unmarshal(bytes, &o.value); err != nil {
		return err // nolint:wrapcheck
	}

	return nil
}
