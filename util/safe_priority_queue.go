package util

import (
	"container/heap"
	"sync"
)

type SafePriorityQueue[T any, P number] struct {
	items []_PriorityQueueItem[T, P]
	lock  *sync.Mutex
}

func (self *SafePriorityQueue[T, P]) Len() int { return len(self.items) }
func (self *SafePriorityQueue[T, P]) Less(i, j int) bool {
	return self.items[i].priority < self.items[j].priority
}
func (self *SafePriorityQueue[T, P]) Swap(i, j int) {
	self.items[i], self.items[j] = self.items[j], self.items[i]
}
func (self *SafePriorityQueue[T, P]) Push(x any) {
	item := x.(_PriorityQueueItem[T, P])
	self.items = append(self.items, item)
}
func (self *SafePriorityQueue[T, P]) Pop() any {
	old := self.items
	n := len(old)
	item := old[n-1]
	self.items = old[0 : n-1]
	return item
}

// Updates the value and priority of an item in the PriorityQueue.
func (self *SafePriorityQueue[T, P]) Update(item int, value T, priority P) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.items[item].value = value
	self.items[item].priority = priority
	heap.Fix(self, item)
}

// Adds a new item to the PriorityQueue with a given value and priority.
func (self *SafePriorityQueue[T, P]) Enqueue(value T, priority P) {
	self.lock.Lock()
	defer self.lock.Unlock()
	heap.Push(self, _PriorityQueueItem[T, P]{value: value, priority: priority})
}

// Removes and returns the item with the lowest priority from the PriorityQueue and a bool indicating success.
// If the PriorityQueue is empty false will be returned.
func (self *SafePriorityQueue[T, P]) Dequeue() (T, bool) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if len(self.items) == 0 {
		var result T
		return result, false
	}
	item := heap.Pop(self).(_PriorityQueueItem[T, P])
	return item.value, true
}

// Clears the PriorityQueue.
func (self *SafePriorityQueue[T, P]) Clear() {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.items = self.items[:0]
}

// Returns the item with the lowest priority from the PriorityQueue without removing.
func (self *SafePriorityQueue[T, P]) Peek() (T, bool) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if self.Len() == 0 {
		var t T
		return t, false
	}
	item := self.items[0]
	return item.value, true
}

// Creates and returns a new PriorityQueue with value T and Priority P.
//
// length specifies initial capacity.
func NewSafePriorityQueue[T any, P number](length int) SafePriorityQueue[T, P] {
	pq := SafePriorityQueue[T, P]{
		items: make([]_PriorityQueueItem[T, P], 0, length),
		lock:  &sync.Mutex{},
	}
	heap.Init(&pq)
	return pq
}
