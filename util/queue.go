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

func NewQueue[T any]() Queue[T] {
	return Queue[T]{}
}

var queue_lock sync.Mutex

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
func (self *Queue[T]) Size() int {
	return self.length
}
