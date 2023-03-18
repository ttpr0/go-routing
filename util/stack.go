package util

import "sync"

type _StackItem[T any] struct {
	value T
	prev  *_StackItem[T]
}

type Stack[T any] struct {
	tail   *_StackItem[T]
	length int
	lock   sync.Mutex
}

// Creates and returns a new Stack
func NewStack[T any]() Stack[T] {
	return Stack[T]{}
}

// Adds a new item to the stack
func (self *Stack[T]) Push(value T) {
	self.lock.Lock()
	defer self.lock.Unlock()
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

// Returns and removes the last added item from the stack and a bool indicating success.
// If the stack is empty then the returned bool is false.
func (self *Stack[T]) Pop() (T, bool) {
	if self.length <= 0 {
		var i T
		return i, false
	}

	self.lock.Lock()
	defer self.lock.Unlock()

	value := self.tail.value
	self.tail = self.tail.prev
	self.length -= 1
	return value, true
}

// Returns the number of elements in the stack
func (self *Stack[T]) Size() int {
	return self.length
}
