package util

import (
	"fmt"
	"sync"
)

type BlockQueue[T any] struct {
	queue      Queue[T]
	block      *Block
	read_lock  sync.Mutex
	write_lock sync.Mutex
}

// Adds a new item to the queue
func (self *BlockQueue[T]) Push(value T) {
	self.write_lock.Lock()
	defer self.write_lock.Unlock()
	self.queue.Push(value)
	if self.queue.Size() == 1 {
		self.block.Release()
	}
}

// Returns and removes the currently earliest added item from the queue.
// If Queue is empty operation blocks until a new item is added
func (self *BlockQueue[T]) Pop() T {
	self.read_lock.Lock()
	defer self.read_lock.Unlock()
	if self.queue.length <= 0 {
		self.block.Take()
	}
	if self.queue.length == 0 {
		fmt.Println(self.queue.Size())
	}
	value, _ := self.queue.Pop()
	return value
}

// Returns the number of items in the queue
func (self *BlockQueue[T]) Size() int {
	return self.queue.Size()
}

func NewBlockQueue[T any]() *BlockQueue[T] {
	queue := &BlockQueue[T]{
		queue:      NewQueue[T](),
		block:      NewBlock(),
		read_lock:  sync.Mutex{},
		write_lock: sync.Mutex{},
	}
	return queue
}
