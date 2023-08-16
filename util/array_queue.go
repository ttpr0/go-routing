package util

import (
	"sync"
)

type ArrayQueue[T any] struct {
	data   []T
	start  int
	length int
	lock   sync.Mutex
}

// Creates and returns a new Queue
func NewArrayQueue[T any](cap int) ArrayQueue[T] {
	return ArrayQueue[T]{
		data:   make([]T, 0, cap),
		start:  0,
		length: 0,
		lock:   sync.Mutex{},
	}
}

// Adds a new item to the queue
func (self *ArrayQueue[T]) Push(value T) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if self.start+self.length < len(self.data) {
		self.data = append(self.data, value)
		self.length += 1
	} else {
		if self.start > 0 {
			copy(self.data[0:self.length], self.data[self.start:self.start+self.length])
			self.data = self.data[0:self.length]
			self.start = 0
		}
		self.data = append(self.data, value)
		self.length += 1
	}
}

// Returns and removes the currently earliest added item from the queue and a bool indicating success.
// If Queue is empty, returned bool is false
func (self *ArrayQueue[T]) Pop() (T, bool) {
	if self.length <= 0 {
		var i T
		return i, false
	}

	self.lock.Lock()
	defer self.lock.Unlock()

	value := self.data[self.start]
	self.start += 1
	self.length -= 1
	return value, true
}

// Returns the number of items in the queue
func (self *ArrayQueue[T]) Size() int {
	return self.length
}

// copies the entire queue
func (self *ArrayQueue[T]) Copy() ArrayQueue[T] {
	new_data := make([]T, len(self.data))
	copy(new_data, self.data)
	return ArrayQueue[T]{
		data:   new_data,
		start:  self.start,
		length: self.length,
		lock:   sync.Mutex{},
	}
}
