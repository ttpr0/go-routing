package util

import (
	"testing"
)

func TestArrayQueueSize(t *testing.T) {
	queue := NewArrayQueue[int](10)

	for i := 0; i < 100; i++ {
		queue.Push(i)
	}
	if queue.Size() != 100 {
		t.Errorf("queue size should be 100, but got %d", queue.Size())
	}

	for i := 0; i < 10; i++ {
		queue.Pop()
	}
	if queue.Size() != 90 {
		t.Errorf("queue size should be 90, but got %d", queue.Size())
	}

	for i := 0; i < 90; i++ {
		queue.Pop()
	}
	_, ok := queue.Pop()
	if ok {
		t.Errorf("should return false on empty queue, but got %v", ok)
	}
}

func TestArrayQueue(t *testing.T) {
	queue := NewArrayQueue[int](10)

	for i := 0; i < 100; i++ {
		queue.Push(i)
	}

	val := 0
	for {
		v, ok := queue.Pop()
		if !ok {
			break
		}
		if v != val {
			t.Errorf("queue should return %d, but got %d", val, v)
		}
		val += 1
	}
}

func TestArrayQueueCopy(t *testing.T) {
	init_queue := NewArrayQueue[int](10)

	for i := 0; i < 100; i++ {
		init_queue.Push(i)
	}

	init_queue.Pop()
	init_queue.Pop()
	init_queue.Pop()

	queue := init_queue.Copy()

	val := 3
	for {
		v, ok := queue.Pop()
		if !ok {
			break
		}
		if v != val {
			t.Errorf("queue should return %d, but got %d", val, v)
		}
		val += 1
	}
}
