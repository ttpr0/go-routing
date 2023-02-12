package util

import (
	"math/rand"
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	queue := NewPriorityQueue[int32, int32](10)

	for i := 0; i < 100; i++ {
		v := rand.Int31n(100)
		queue.Enqueue(v, v)
	}

	prev := int32(-1)
	for {
		item, ok := queue.Dequeue()
		if !ok {
			break
		}
		if item < prev {
			t.Errorf("wrong priority ordering")
		}
		prev = item
	}
}
