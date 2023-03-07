package util

import (
	"container/heap"
)

type _PriorityQueueItem[T any, P number] struct {
	value    T
	priority P
}

type PriorityQueue[T any, P number] []_PriorityQueueItem[T, P]

func (self *PriorityQueue[T, P]) Len() int { return len(*self) }
func (self *PriorityQueue[T, P]) Less(i, j int) bool {
	return (*self)[i].priority < (*self)[j].priority
}
func (self *PriorityQueue[T, P]) Swap(i, j int) {
	(*self)[i], (*self)[j] = (*self)[j], (*self)[i]
}
func (self *PriorityQueue[T, P]) Push(x any) {
	item := x.(_PriorityQueueItem[T, P])
	*self = append(*self, item)
}
func (self *PriorityQueue[T, P]) Pop() any {
	old := *self
	n := len(old)
	item := old[n-1]
	*self = old[0 : n-1]
	return item
}

// Updates the value and priority of an item in the PriorityQueue.
func (self *PriorityQueue[T, P]) Update(item int, value T, priority P) {
	(*self)[item].value = value
	(*self)[item].priority = priority
	heap.Fix(self, item)
}

// Adds a new item to the PriorityQueue with a given value and priority.
func (self *PriorityQueue[T, P]) Enqueue(value T, priority P) {
	heap.Push(self, _PriorityQueueItem[T, P]{value: value, priority: priority})
}

// Removes and returns the item with the lowest priority from the PriorityQueue and a bool indicating success.
// If the PriorityQueue is empty false will be returned.
func (self *PriorityQueue[T, P]) Dequeue() (T, bool) {
	if len(*self) == 0 {
		var result T
		return result, false
	}
	item := heap.Pop(self).(_PriorityQueueItem[T, P])
	return item.value, true
}

// Clears the PriorityQueue.
func (self *PriorityQueue[T, P]) Clear() {
	*self = (*self)[:0]
}

// Returns the item with the lowest priority from the PriorityQueue without removing.
func (self *PriorityQueue[T, P]) Peek() (T, bool) {
	if self.Len() == 0 {
		var t T
		return t, false
	}
	item := (*self)[0]
	return item.value, true
}

// Creates and returns a new PriorityQueue with value T and Priority P.
//
// length specifies initial capacity.
func NewPriorityQueue[T any, P number](length int) PriorityQueue[T, P] {
	pq := make(PriorityQueue[T, P], 0, length)
	heap.Init(&pq)
	return pq
}
