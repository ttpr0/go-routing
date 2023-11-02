package util

import (
	"math/rand"
	"testing"
)

func create_random_data(n1, n2 int) Array[Tuple[int32, int32]] {
	arr := NewArray[Tuple[int32, int32]](n1)
	for i := 0; i < n1; i++ {
		arr[i] = MakeTuple(rand.Int31(), rand.Int31n(int32(n2)))
	}
	return arr
}

const N = 10000

func Benchmark_FlagsGet(b *testing.B) {
	items := create_random_data(N, 1000)
	arr := NewFlags[int32](N, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < N; j++ {
			item := items[j]
			val := arr.Get(item.B)
			*val = item.A
		}
	}
}

func Benchmark_FlagsGetCompare(b *testing.B) {
	items := create_random_data(N, 1000)
	arr := make([]int32, N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < N; j++ {
			item := items[j]
			arr[item.B] = item.A
		}
	}
}
