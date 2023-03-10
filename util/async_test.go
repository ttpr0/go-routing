package util

import (
	"math/rand"
	"testing"
	"time"
)

func TestAwait(t *testing.T) {
	prom1 := Async(func() int {
		return 111
	})
	prom2 := Async(func() string {
		return "test"
	})
	prom3 := Async(func() Tuple[int, float64] {
		return MakeTuple(1, 1.5)
	})

	val1 := Await(prom1)
	if val1 != 111 {
		t.Errorf("expected 111, got %v", val1)
	}
	val2 := Await(prom2)
	if val2 != "test" {
		t.Errorf("expected 'test', got %v", val2)
	}
	val3 := Await(prom3)
	if val3 != MakeTuple(1, 1.5) {
		t.Errorf("expected [1, 1.5], got %v", val3)
	}
}

func TestAwaitList(t *testing.T) {
	promises := NewList[Promise[int]](10)
	for i := 0; i < 100; i++ {
		index := 0
		promises.Add(Async(func() int {
			time.Sleep(time.Duration(rand.Intn(10) * 1000))
			return index
		}))
	}

	iter := AwaitList(promises)
	c := 0
	for {
		val, ok := iter.Next()
		if !ok {
			break
		}
		c += 1
		if val < 0 || val > 99 {
			t.Errorf("expected value between [0, 9], got %v", val)
		}
	}
	if c != 100 {
		t.Errorf("expected 10 values, got %v", c)
	}
}
