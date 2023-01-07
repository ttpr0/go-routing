package util

type List[T any] []T

func (self *List[T]) Length() int {
	return len(*self)
}
func (self *List[T]) Clear() {
	*self = (*self)[:0]
}
func (self *List[T]) Get(index int) T {
	return (*self)[index]
}
func (self *List[T]) Set(index int, value T) {
	(*self)[index] = value
}
func (self *List[T]) Add(value T) {
	*self = append(*self, value)
}
func (self *List[T]) Remove(index int) {
	l := self.Length()
	if index >= l || index < 0 {
		return
	}
	copy((*self)[index:l-1], (*self)[index+1:l])
	*self = (*self)[:l-1]
}
func (self *List[T]) Slice(start int, end int) List[T] {
	return (*self)[start:end]
}
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

func NewList[T any](cap int) List[T] {
	return make([]T, 0, cap)
}
