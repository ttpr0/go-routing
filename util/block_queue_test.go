package util

import (
	"testing"
	"time"
)

func TestBlockQueue(t *testing.T) {
	queue := NewBlockQueue[int]()

	val1 := 0
	val2 := 0
	go func(queue *BlockQueue[int], val *int) {
		for i := 10; i < 100; i++ {
			time.Sleep(1000)
			queue.Push(i)
		}
		*val = 1
	}(queue, &val1)
	go func(queue *BlockQueue[int], val *int) {
		for i := 10; i < 100; i++ {
			time.Sleep(1000)
			queue.Push(i)
		}
		*val = 1
	}(queue, &val2)

	for i := 0; i < 180; i++ {
		v := queue.Pop()
		if v < 10 || v > 99 {
			t.Errorf("queue should return [10, 99], but got %d", v)
		}
	}
	if val1 == 0 {
		t.Errorf("expected 1, but got %d", val1)
	}
	if val2 == 0 {
		t.Errorf("expected 1, but got %d", val2)
	}
}
