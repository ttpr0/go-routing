package util

import "fmt"

type Optional[T any] struct {
	Value      T
	_has_value bool
}

// Returns if the optional value is set.
func (self *Optional[T]) HasValue() bool {
	return self._has_value
}

func (self Optional[T]) String() string {
	return fmt.Sprintf("%v", self.Value)
}

// Creates an optional with set value.
func Some[T any](value T) Optional[T] {
	return Optional[T]{
		Value:      value,
		_has_value: true,
	}
}

// Creates an unset optional.
func None[T any]() Optional[T] {
	var value T
	return Optional[T]{
		Value:      value,
		_has_value: false,
	}
}
