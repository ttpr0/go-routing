package util

import (
	"math/rand"
	"testing"
)

func create_random(n int) Array[int32] {
	arr := NewArray[int32](n)
	rand.Seed(100)
	for i := 0; i < n; i++ {
		arr[i] = rand.Int31()
	}
	return arr
}

func Benchmark_Queue(b *testing.B) {
	const N = 100
	items := create_random(N)
	queue := NewQueue[int32]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < N; j++ {
			queue.Push(items[j])
		}
		for j := 0; j < N; j++ {
			queue.Pop()
		}
	}
}

func Benchmark_ArrayQueue(b *testing.B) {
	const N = 100
	items := create_random(N)
	queue := NewArrayQueue[int32](N / 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < N; j++ {
			queue.Push(items[j])
		}
		for j := 0; j < N; j++ {
			queue.Pop()
		}
	}
}

func create_random_op(n int) Array[Tuple[int32, bool]] {
	arr := NewArray[Tuple[int32, bool]](n)
	rand.Seed(100)
	for i := 0; i < n; i++ {
		arr[i] = MakeTuple(rand.Int31(), rand.Int31n(100) > 50)
	}
	return arr
}

func Benchmark_QueueRandomOp(b *testing.B) {
	const N = 100
	items := create_random_op(N)
	queue := NewQueue[int32]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < N; j++ {
			item := items[j]
			if item.B {
				queue.Push(item.A)
			} else {
				queue.Pop()
			}
		}
	}
}

func Benchmark_ArrayQueueRandomOp(b *testing.B) {
	const N = 100
	items := create_random_op(N)
	queue := NewArrayQueue[int32](N / 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < N; j++ {
			item := items[j]
			if item.B {
				queue.Push(item.A)
			} else {
				queue.Pop()
			}
		}
	}
}
