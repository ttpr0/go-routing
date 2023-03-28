package util

type Array[T any] []T

// Returns the length of the Array.
func (self *Array[T]) Length() int {
	return len(*self)
}

// Returns the element at index.
func (self *Array[T]) Get(index int) T {
	return (*self)[index]
}

// Sets the element at index.
func (self *Array[T]) Set(index int, value T) {
	(*self)[index] = value
}

// Slices the Array from start to end and returns the sliced List.
// Returns a view to the original Array.
func (self *Array[T]) Slice(start int, end int) List[T] {
	return List[T]((*self)[start:end])
}

// Creates and Returns a new Array with capacity cap.
func NewArray[T any](cap int) Array[T] {
	return make([]T, cap)
}
