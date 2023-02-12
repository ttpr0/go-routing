package util

import (
	"fmt"
	"sync"
)

type _DequeItem[T any] struct {
	value T
	next  *_DequeItem[T]
	prev  *_DequeItem[T]
}

type Deque[T any] struct {
	head   *_DequeItem[T]
	tail   *_DequeItem[T]
	length int
}

// creates and returns a new Deque
func NewDeque[T any]() Deque[T] {
	return Deque[T]{}
}

var deque_lock sync.Mutex

// push the value to the back of the deque
func (self *Deque[T]) PushBack(value T) {
	deque_lock.Lock()
	defer deque_lock.Unlock()
	if self.length == 0 {
		item := &_DequeItem[T]{
			value: value,
		}
		self.head = item
		self.tail = item
		self.length += 1
	} else {
		item := &_DequeItem[T]{
			value: value,
		}
		self.tail.next = item
		item.prev = self.tail
		self.tail = item
		self.length += 1
	}
}

// push the value to the front of the deque
func (self *Deque[T]) PushFront(value T) {
	deque_lock.Lock()
	defer deque_lock.Unlock()
	if self.length == 0 {
		item := &_DequeItem[T]{
			value: value,
		}
		self.head = item
		self.tail = item
		self.length += 1
	} else {
		item := &_DequeItem[T]{
			value: value,
		}
		self.head.prev = item
		item.next = self.head
		self.head = item
		self.length += 1
	}
}

// Returns and removes the last element from the deque and a bool indicating success.
//
// If the deque is empty the returned bool will be false else true.
func (self *Deque[T]) PopBack() (T, bool) {
	if self.length <= 0 {
		var i T
		return i, false
	}

	deque_lock.Lock()
	defer deque_lock.Unlock()

	if self.length == 1 {
		value := self.tail.value
		self.head = nil
		self.tail = nil
		self.length -= 1
		return value, true
	} else {
		value := self.tail.value
		self.tail = self.tail.prev
		self.tail.next = nil
		self.length -= 1
		return value, true
	}
}

// Returns and removes the first element from the deque and a bool indicating success.
//
// If the deque is empty the returned bool will be false else true.
func (self *Deque[T]) PopFront() (T, bool) {
	if self.length <= 0 {
		var i T
		return i, false
	}

	deque_lock.Lock()
	defer deque_lock.Unlock()

	if self.length == 1 {
		value := self.head.value
		self.tail = nil
		self.head = nil
		self.length -= 1
		return value, true
	} else {
		value := self.head.value
		self.head = self.head.next
		self.head.prev = nil
		self.length -= 1
		return value, true
	}
}

// Returns the element at the given index from the deque and a bool indicating success.
//
// If the element does not exist the returned bool will be false else true.
func (self *Deque[T]) GetAt(index int) (T, bool) {
	if index < 0 || index >= self.length {
		var i T
		return i, false
	}
	curr := self.head
	for i := 0; i < index; i++ {
		curr = curr.next
	}
	return curr.value, true
}

// Sets the element at the given index from the deque and a bool indicating success.
func (self *Deque[T]) SetAt(index int, value T) bool {
	if index < 0 || index >= self.length {
		return false
	}
	curr := self.head
	for i := 0; i < index; i++ {
		curr = curr.next
	}
	curr.value = value
	return true
}

// Adds an element to the deque at the given index. Elements after this index will be moved back by one.
func (self *Deque[T]) AddAt(index int, value T) bool {
	if index < 0 || index > self.length {
		return false
	}
	if index == 0 {
		self.PushFront(value)
		return true
	}
	if index == self.length {
		self.PushBack(value)
		return true
	}

	deque_lock.Lock()
	defer deque_lock.Unlock()

	curr := self.head
	for i := 0; i < index; i++ {
		curr = curr.next
	}
	prev := curr.prev
	item := &_DequeItem[T]{
		value: value,
	}
	prev.next = item
	item.prev = prev
	item.next = curr
	curr.prev = item
	self.length += 1
	return true
}

// Removes an element from the deque at the given index. Index of Elements after this index will be reduced by one.
func (self *Deque[T]) RemoveAt(index int) bool {
	if index < 0 || index >= self.length {
		return false
	}
	if index == 0 {
		_, ok := self.PopFront()
		return ok
	}
	if index == self.length-1 {
		_, ok := self.PopBack()
		return ok
	}

	deque_lock.Lock()
	defer deque_lock.Unlock()

	curr := self.head
	for i := 0; i < index; i++ {
		curr = curr.next
	}
	prev := curr.prev
	next := curr.next
	prev.next = next
	next.prev = prev
	self.length -= 1
	return true
}

// Returns the number of elements in the deque
func (self *Deque[T]) Size() int {
	return self.length
}

// Returns a string representation of the deque
func (self Deque[T]) String() string {
	s := "{"
	curr := self.head
	for i := 0; i < self.length; i++ {
		s = string(fmt.Append([]byte(s), curr.value))
		s += ", "
		curr = curr.next
	}
	s += "}"
	return s
}
