package util

import (
	"sync"
)

type _QueueItem[T any] struct {
	value T
	next  *_QueueItem[T]
}

type Queue[T any] struct {
	head   *_QueueItem[T]
	tail   *_QueueItem[T]
	length int
}

// Creates and returns a new Queue
func NewQueue[T any]() Queue[T] {
	return Queue[T]{}
}

var queue_lock sync.Mutex

// Adds a new item to the queue
func (self *Queue[T]) Push(value T) {
	queue_lock.Lock()
	defer queue_lock.Unlock()
	if self.head == nil {
		self.head = &_QueueItem[T]{
			value: value,
		}
		self.tail = self.head
		self.length = 1
	} else {
		item := &_QueueItem[T]{
			value: value,
		}
		self.tail.next = item
		self.tail = item
		self.length += 1
	}
}

// Returns and removes the currently earliest added item from the queue and a bool indicating success.
// If Queue is empty, returned bool is false
func (self *Queue[T]) Pop() (T, bool) {
	if self.length <= 0 {
		var i T
		return i, false
	}

	queue_lock.Lock()
	defer queue_lock.Unlock()

	value := self.head.value
	self.head = self.head.next
	self.length -= 1
	return value, true
}

// Returns the number of items in the queue
func (self *Queue[T]) Size() int {
	return self.length
}
