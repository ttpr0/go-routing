package util

import (
	"container/heap"
)

type PriorityQueueItem[T any] struct {
	value    T
	priority float32
	index    int
}

type PriorityQueue[T any] []*PriorityQueueItem[T]

func (self PriorityQueue[T]) Len() int { return len(self) }

func (self PriorityQueue[T]) Less(i, j int) bool {
	return self[i].priority < self[j].priority
}

func (self PriorityQueue[T]) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
	self[i].index = i
	self[j].index = j
}

func (self *PriorityQueue[T]) Push(x any) {
	n := len(*self)
	item := x.(*PriorityQueueItem[T])
	item.index = n
	*self = append(*self, item)
}

func (self *PriorityQueue[T]) Pop() any {
	old := *self
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*self = old[0 : n-1]
	return item
}

func (self *PriorityQueue[T]) Update(item *PriorityQueueItem[T], value T, priority float32) {
	item.value = value
	item.priority = priority
	heap.Fix(self, item.index)
}

func (self *PriorityQueue[T]) Enqueue(value T, priority float32) {
	heap.Push(self, &PriorityQueueItem[T]{value: value, priority: priority})
}

func (self *PriorityQueue[T]) Dequeue() (T, bool) {
	if len(*self) == 0 {
		var result T
		return result, false
	}
	item := heap.Pop(self).(*PriorityQueueItem[T])
	return item.value, true
}

func NewPriorityQueue[T any](length int) PriorityQueue[T] {
	pq := make(PriorityQueue[T], 0, length)
	heap.Init(&pq)
	return pq
}
