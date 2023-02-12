package util

type List[T any] []T

// Returns the number of elements in the List.
func (self *List[T]) Length() int {
	return len(*self)
}

// Clears the List.
func (self *List[T]) Clear() {
	*self = (*self)[:0]
}

// Returns the element at index.
func (self *List[T]) Get(index int) T {
	return (*self)[index]
}

// Sets the element at index.
func (self *List[T]) Set(index int, value T) {
	(*self)[index] = value
}

// Adds an elements to the end of the List.
func (self *List[T]) Add(value T) {
	*self = append(*self, value)
}

// Removes the element at index. Index of all following elements is reduced by one.
func (self *List[T]) Remove(index int) {
	l := self.Length()
	if index >= l || index < 0 {
		return
	}
	copy((*self)[index:l-1], (*self)[index+1:l])
	*self = (*self)[:l-1]
}

// Slices the List from start to end and returns the new List.
// Returns a view to the original List.
func (self *List[T]) Slice(start int, end int) List[T] {
	return (*self)[start:end]
}

// Returns an IIterator for all elements of the List.
func (self *List[T]) Values() IIterator[T] {
	return _ListIterator[T]{self, 0}
}

type _ListIterator[T any] struct {
	list *List[T]
	pos  int
}

func (self _ListIterator[T]) Next() (T, bool) {
	self.pos += 1
	if self.pos == self.list.Length() {
		var i T
		return i, false
	}
	return self.list.Get(self.pos), true
}

// Creates and Returns a new List with initial capacity cap.
func NewList[T any](cap int) List[T] {
	return make([]T, 0, cap)
}

// Returns the index of the first occurrence in the List.
// Returns -1 if the value is not in the List.
func GetIndexOf[T comparable](list List[T], value T) int {
	for i, v := range list {
		if v == value {
			return i
		}
	}
	return -1
}

// Checks if the List contains the value specified.
func Contains[T comparable](list List[T], value T) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
