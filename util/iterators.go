package util

func NewListIterator[T any](list List[T]) IIterator[T] {
	return &ListIterator[T]{
		state: 0,
		list:  list,
	}
}

type ListIterator[T any] struct {
	state int
	list  List[T]
}

func (self *ListIterator[T]) Next() (T, bool) {
	if self.state == len(self.list) {
		var t T
		return t, false
	}
	ref := self.list[self.state]
	self.state += 1
	return ref, true
}
