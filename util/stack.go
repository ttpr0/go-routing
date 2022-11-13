package util

import "sync"

type _StackItem[T any] struct {
	value T
	prev  *_StackItem[T]
}

type Stack[T any] struct {
	tail   *_StackItem[T]
	length int
}

func NewStack[T any]() Stack[T] {
	return Stack[T]{}
}

var stack_lock sync.Mutex

func (self *Stack[T]) Push(value T) {
	stack_lock.Lock()
	defer stack_lock.Unlock()
	if self.tail == nil {
		self.tail = &_StackItem[T]{
			value: value,
		}
		self.length = 1
	} else {
		item := &_StackItem[T]{
			value: value,
		}
		item.prev = self.tail
		self.tail = item
		self.length += 1
	}
}
func (self *Stack[T]) Pop() (T, bool) {
	if self.length <= 0 {
		var i T
		return i, false
	}

	stack_lock.Lock()
	defer stack_lock.Unlock()

	value := self.tail.value
	self.tail = self.tail.prev
	self.length -= 1
	return value, true
}
func (self *Stack[T]) Size() int {
	return self.length
}
