package util

import (
	"math/rand"
	"testing"
)

func create_random_index(n1, n2 int) Array[Tuple[int32, int]] {
	arr := NewArray[Tuple[int32, int]](n1)
	rand.Seed(100)
	for i := 0; i < n1; i++ {
		arr[i] = MakeTuple(rand.Int31(), int(rand.Int31n(int32(n2))))
	}
	return arr
}

func Benchmark_ArraySet(b *testing.B) {
	const N = 100
	items := create_random_index(N, 1000)
	arr := NewArray[int32](1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < N; j++ {
			item := items[j]
			arr.Set(item.B, item.A)
		}
	}
}

func Benchmark_SliceSet(b *testing.B) {
	const N = 100
	items := create_random_index(N, 1000)
	arr := make([]int32, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < N; j++ {
			item := items[j]
			arr[item.B] = item.A
		}
	}
}

func Benchmark_ArrayGet(b *testing.B) {
	const N = 100
	items := create_random_index(N, 1000)
	arr := NewArray[int32](1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < N; j++ {
			item := items[j]
			item.A = arr.Get(item.B)
		}
	}
}

func Benchmark_SliceGet(b *testing.B) {
	const N = 100
	items := create_random_index(N, 1000)
	arr := make([]int32, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < N; j++ {
			item := items[j]
			item.A = arr[item.B]
		}
	}
}
