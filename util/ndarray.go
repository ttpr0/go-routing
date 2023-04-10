package util

import "unsafe"

type NDArray[T any] struct {
	data    []T
	shape   []int
	offset  int
	strides []int
}

// Returns the total number of elements in the NDArray.
func (self *NDArray[T]) Length() int {
	return len(self.data)
}

// Returns the shape of the NDArray.
func (self *NDArray[T]) Shape() []int {
	return self.shape
}

// Returns a new NDArray with the given shape.
// Returned array shares the underlying data with self.
func (self *NDArray[T]) Reshape(shape ...int) (NDArray[T], bool) {
	new_length := 1
	for _, size := range shape {
		new_length *= size
	}
	if new_length != len(self.data) {
		var t NDArray[T]
		return t, false
	}
	strides := _StridesFromShape(shape)

	arr := NDArray[T]{
		data:    self.data,
		shape:   shape,
		strides: strides,
	}
	return arr, true
}

// Returns the element at indices.
func (self *NDArray[T]) Get(indices ...int) T {
	if len(indices) != len(self.shape) {
		var t T
		return t
	}
	index := 0
	for i := 0; i < len(indices); i++ {
		index += indices[i] * self.strides[i]
	}
	return self.data[index]
}

// Sets the element at indices.
func (self *NDArray[T]) Set(value T, indices ...int) {
	if len(indices) != len(self.shape) {
		return
	}
	index := 0
	for i := 0; i < len(indices); i++ {
		index += indices[i] * self.strides[i]
	}
	self.data[index] = value
}

// Creates and Returns a new Array with shape.
func NewNDArray[T any](shape ...int) NDArray[T] {
	length := 1
	for _, size := range shape {
		length *= size
	}
	data := make([]T, length)

	strides := _StridesFromShape(shape)

	arr := NDArray[T]{
		data:    data,
		shape:   shape,
		offset:  0,
		strides: strides,
	}
	return arr
}

// Creates and Returns a new Array view on the underlying data at ptr.
func NewUnsafeNDArray[T any](ptr *T, shape []int) NDArray[T] {
	length := 1
	for _, size := range shape {
		length *= size
	}

	data := unsafe.Slice(ptr, length)

	strides := _StridesFromShape(shape)

	arr := NDArray[T]{
		data:    data,
		shape:   shape,
		offset:  0,
		strides: strides,
	}
	return arr
}

func _StridesFromShape(shape []int) []int {
	strides := make([]int, len(shape))
	strides[len(strides)-1] = 1
	for i := len(shape) - 1; i > 0; i-- {
		strides[i-1] = strides[i] * shape[i]
	}
	return strides
}
